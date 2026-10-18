package api

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/ai/integrations"
	"github.com/nimbleflux/fluxbase/internal/api/vector"
	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/logging"
	"github.com/nimbleflux/fluxbase/internal/observability"
	"github.com/nimbleflux/fluxbase/internal/settings"
	"github.com/nimbleflux/fluxbase/internal/storage"
)

type AIModule struct {
	Handlers *AIHandlers
	Quota    *QuotaHandlers
}

func (m *AIModule) Name() string { return "ai" }

func (m *AIModule) Init(ctx context.Context, registry *ServiceRegistry) error {
	cfg := registry.Config
	db := registry.DB
	storageSvc := GetService[*storage.Service](registry)
	loggingSvc := GetService[*logging.Service](registry)
	secretsSvc := GetService[*settings.SecretsService](registry)
	userMgmtSvc := GetService[*auth.UserManagementService](registry)

	aiStorage := ai.NewStorage(db)
	aiStorage.SetConfig(&cfg.AI)
	// Master key for encrypting provider config secrets (api_key etc.) at
	// rest. May be nil — encryption stays off in that case (legacy behavior).
	aiStorage.SetEncryptionKey(cfg.EncryptionKeyBytes)

	vectorManager := vector.NewManager(&cfg.AI, aiStorage, db.Inspector(), db)

	// Tool integrations (Web Agent specialist). Construct the EnvConfig
	// from YAML/env vars so instance-level Tavily config produces a
	// synthetic read-only FROM_CONFIG row — same pattern as AI providers.
	integrationsEnvConfig := &integrations.EnvConfig{
		TavilyAPIKey:       cfg.AI.TavilyAPIKey,
		TavilyDefaultDepth: cfg.AI.TavilyDefaultDepth,
		TavilyBaseURL:      cfg.AI.TavilyBaseURL,
	}
	integrationsStorage := integrations.NewStorage(db, cfg.EncryptionKeyBytes, integrationsEnvConfig)
	integrationsHandler := integrations.NewHandler(integrationsStorage)

	var vectorHandler *vector.Handler
	vectorHandler, err := vector.NewHandler(vectorManager, db.Inspector(), db, cfg)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize vector handler")
	} else if vectorHandler.IsEmbeddingConfigured() {
		provider := cfg.AI.EmbeddingProvider
		if provider == "" {
			provider = cfg.AI.ProviderType
		}
		model := ""
		if vectorHandler.GetEmbeddingService() != nil {
			model = vectorHandler.GetEmbeddingService().DefaultModel()
		}
		log.Info().
			Str("provider", provider).
			Str("model", model).
			Bool("explicit_config", cfg.AI.EmbeddingEnabled).
			Msg("Vector handler initialized with embedding support")
	} else {
		log.Info().Msg("Vector handler initialized (embedding not available)")
	}

	var aiHandler *ai.Handler
	var aiChatHandler *ai.ChatHandler
	var aiConversations *ai.ConversationManager
	var aiMetrics *observability.Metrics
	var knowledgeBaseHandler *ai.KnowledgeBaseHandler
	var reranker ai.Reranker
	var kbStorage *ai.KnowledgeBaseStorage
	var knowledgeGraph *ai.KnowledgeGraph
	var docProcessor *ai.DocumentProcessor
	var tableExportSyncService *ai.TableExportSyncService
	var ocrService *ai.OCRService
	var quotaHandler *QuotaHandler
	var internalAIHandler *InternalAIHandler

	if cfg.AI.Enabled {
		aiMetrics = observability.NewMetrics()

		aiLoader := ai.NewLoader(cfg.AI.ChatbotsDir)
		aiConversations = ai.NewConversationManager(db, cfg.AI.ConversationCacheTTL, cfg.AI.MaxConversationTurns)
		aiConversations.SetMetrics(aiMetrics)
		aiHandler = ai.NewHandler(aiStorage, aiLoader, &cfg.AI, vectorManager)

		var embeddingService *ai.EmbeddingService
		if vectorHandler != nil {
			embeddingService = vectorHandler.GetEmbeddingService()
		}

		// Semantic response cache store: opt-in per chatbot via the
		// @fluxbase:response-cache annotation; entries expire after
		// ai.response_cache_ttl and are invalidated when a linked KB
		// document changes. Shared by the chat path and the document
		// processor invalidation hook below.
		var responseCacheStore *ai.ResponseCacheStore
		if cfg.AI.ResponseCacheTTL > 0 && embeddingService != nil {
			responseCacheStore = ai.NewResponseCacheStore(db)
			responseCacheStore.StartPurgeSweeper(context.Background(), time.Hour)
			log.Info().Dur("ttl", cfg.AI.ResponseCacheTTL).Msg("Response cache enabled")
		}

		// Reranker: optional relevance re-ordering of retrieved chunks.
		// Provider "llm" reuses the default chat provider; "cohere" calls the
		// Cohere Rerank API. Per-KB opt-out via kb.rerank_enabled.
		if cfg.AI.RerankEnabled {
			switch strings.ToLower(strings.TrimSpace(cfg.AI.RerankProvider)) {
			case "cohere":
				reranker = ai.NewCohereReranker(ai.CohereRerankerConfig{
					APIKey:  cfg.AI.RerankAPIKey,
					Model:   cfg.AI.RerankModel,
					BaseURL: cfg.AI.RerankBaseURL,
					Timeout: cfg.AI.RerankTimeout,
				})
			default: // "llm"
				reranker = ai.NewLLMReranker(ai.LLMRerankerConfig{
					Resolve: func(ctx context.Context) (ai.Provider, error) {
						record, err := aiStorage.GetEffectiveDefaultProvider(ctx)
						if err != nil {
							return nil, err
						}
						if record == nil {
							return nil, nil
						}
						return ai.NewProvider(ai.ProviderConfig{
							Name:        record.Name,
							DisplayName: record.DisplayName,
							Type:        ai.ProviderType(record.ProviderType),
							Config:      record.Config,
						})
					},
					Model: cfg.AI.RerankModel,
				})
			}
			log.Info().Str("provider", reranker.Provider()).Int("top_n", cfg.AI.RerankTopN).Msg("Reranker enabled")
		}

		aiChatHandler = ai.NewChatHandler(db, aiStorage, aiConversations, aiMetrics, &cfg.AI, embeddingService, loggingSvc)
		if reranker != nil {
			aiChatHandler.SetReranker(reranker, cfg.AI.RerankTopN)
		}
		if responseCacheStore != nil {
			aiChatHandler.SetResponseCache(responseCacheStore)
		}

		// Wire integrations storage into the chat handler so the supervisor's
		// Web Agent can resolve Tavily credentials at turn start.
		aiChatHandler.SetIntegrationsStorage(integrationsStorage)

		// Wire chat handler back into the admin handler so provider mutations
		// invalidate the chat path's provider cache.
		aiHandler.SetChatHandler(aiChatHandler)

		if secretsSvc != nil {
			settingsResolver := ai.NewSettingsResolver(secretsSvc, 5*time.Minute)
			aiChatHandler.SetSettingsResolver(settingsResolver)
		}

		log.Info().
			Str("chatbots_dir", cfg.AI.ChatbotsDir).
			Bool("auto_load", cfg.AI.AutoLoadOnBoot).
			Str("provider_type", cfg.AI.ProviderType).
			Str("provider_name", cfg.AI.ProviderName).
			Str("provider_model", cfg.AI.ProviderModel).
			Bool("rag_enabled", embeddingService != nil).
			Msg("AI components initialized")

		if cfg.AI.OCREnabled {
			var ocrErr error
			ocrService, ocrErr = ai.NewOCRService(ai.OCRServiceConfig{
				Enabled:          cfg.AI.OCREnabled,
				ProviderType:     ai.OCRProviderType(cfg.AI.OCRProvider),
				DefaultLanguages: cfg.AI.OCRLanguages,
			})
			if ocrErr != nil {
				log.Warn().Err(ocrErr).Msg("Failed to initialize OCR service, OCR will be disabled")
			} else if ocrService.IsEnabled() {
				log.Info().
					Str("provider", cfg.AI.OCRProvider).
					Strs("languages", cfg.AI.OCRLanguages).
					Msg("OCR service initialized")
			}
		}

		kbStorage = ai.NewKnowledgeBaseStorage(db)
		knowledgeGraph = ai.NewKnowledgeGraph(kbStorage)
		log.Info().Msg("Knowledge graph initialized")

		quotaService := ai.NewQuotaService(kbStorage)

		// Entity extraction: strategy dispatch (rules/llm/hybrid) per KB.
		// The LLM extractor resolves the effective default provider at
		// extraction time so admin provider changes apply without a restart.
		ruleExtractor := ai.NewRuleBasedExtractor()
		var entityExtractor ai.EntityExtractor = ruleExtractor
		strategy := strings.ToLower(strings.TrimSpace(cfg.AI.EntityExtractionStrategy))
		if strategy == "" {
			strategy = ai.ExtractionStrategyRules
		}
		extractionLLM := ai.NewLLMExtractor(ai.LLMExtractorConfig{
			Resolve: func(ctx context.Context) (ai.Provider, error) {
				record, err := aiStorage.GetEffectiveDefaultProvider(ctx)
				if err != nil {
					return nil, err
				}
				if record == nil {
					return nil, nil
				}
				return ai.NewProvider(ai.ProviderConfig{
					Name:        record.Name,
					DisplayName: record.DisplayName,
					Type:        ai.ProviderType(record.ProviderType),
					Config:      record.Config,
				})
			},
			MaxTokens:    cfg.AI.EntityExtractionMaxTokens,
			MinTextChars: cfg.AI.EntityExtractionMinTextChars,
			Timeout:      cfg.AI.EntityExtractionTimeout,
		})
		entityExtractor = ai.NewStrategyEntityExtractor(ruleExtractor, extractionLLM, kbStorage, strategy)
		log.Info().Str("default_strategy", strategy).Msg("Entity extractor initialized")

		// The chat handler's RAG path reuses the strategy-aware extractor.
		aiChatHandler.SetEntityExtractor(entityExtractor)

		if vectorHandler != nil && vectorHandler.GetEmbeddingService() != nil {
			docProcessor = ai.NewDocumentProcessor(kbStorage, vectorHandler.GetEmbeddingService(), entityExtractor, knowledgeGraph)
		}

		if docProcessor != nil {
			// KB quota enforcement for the document add/upload paths.
			docProcessor.SetQuotaService(quotaService)

			// Document processing recovery: periodically requeue documents
			// stuck in 'processing' (e.g. after a crash) and drain pending.
			docProcessor.StartProcessingSweeper(context.Background(), 5*time.Minute)
		}

		// Invalidate response-cache entries when a linked KB's content
		// changes, so cached grounded answers never go stale.
		if responseCacheStore != nil && docProcessor != nil {
			docProcessor.SetDocumentProcessedHook(func(kbID string) {
				invalidateCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := responseCacheStore.InvalidateForKnowledgeBase(invalidateCtx, kbID); err != nil {
					log.Warn().Err(err).Str("kb_id", kbID).Msg("Failed to invalidate response cache for KB")
				}
			})
		}

		if ocrService != nil && ocrService.IsEnabled() {
			knowledgeBaseHandler = ai.NewKnowledgeBaseHandlerWithOCR(kbStorage, docProcessor, ocrService)
		} else {
			knowledgeBaseHandler = ai.NewKnowledgeBaseHandler(kbStorage, docProcessor)
		}
		if storageSvc != nil {
			knowledgeBaseHandler.SetStorageService(storageSvc)
		}
		if reranker != nil {
			knowledgeBaseHandler.SetReranker(reranker)
		}

		tableExporter := ai.NewTableExporter(db, docProcessor, knowledgeGraph, kbStorage)
		knowledgeBaseHandler.SetTableExporter(tableExporter)
		knowledgeBaseHandler.SetKnowledgeGraph(knowledgeGraph)
		log.Info().Msg("Table exporter initialized")

		tableExportSyncService = ai.NewTableExportSyncService(db, tableExporter, kbStorage)
		knowledgeBaseHandler.SetSyncService(tableExportSyncService)
		log.Info().Msg("Table export sync service initialized")

		aiHandler.SetKnowledgeBaseStorage(kbStorage)
		log.Info().Msg("AI handler configured with knowledge base storage")

		log.Info().
			Bool("processing_enabled", docProcessor != nil).
			Bool("ocr_enabled", ocrService != nil && ocrService.IsEnabled()).
			Bool("entity_extraction_enabled", true).
			Str("entity_extraction_strategy", strategy).
			Bool("table_export_enabled", true).
			Bool("sync_enabled", true).
			Msg("Knowledge base handler initialized")

		quotaHandler = NewQuotaHandler(quotaService, userMgmtSvc)
		log.Info().Msg("Quota service and handler initialized")

		var embeddingSvc *ai.EmbeddingService
		if vectorHandler != nil {
			embeddingSvc = vectorHandler.GetEmbeddingService()
		}
		internalAIHandler = NewInternalAIHandler(aiStorage, embeddingSvc, cfg.AI.ProviderName)
		log.Info().
			Str("default_provider", cfg.AI.ProviderName).
			Bool("embedding_enabled", embeddingSvc != nil).
			Msg("Internal AI handler initialized for MCP tools/functions/jobs")
	}

	m.Handlers = &AIHandlers{
		Handler:         aiHandler,
		Chat:            aiChatHandler,
		Conversations:   aiConversations,
		Metrics:         aiMetrics,
		KnowledgeBase:   knowledgeBaseHandler,
		KBStorage:       kbStorage,
		KnowledgeGraph:  knowledgeGraph,
		DocProcessor:    docProcessor,
		TableExportSync: tableExportSyncService,
		VectorManager:   vectorManager,
		VectorHandler:   vectorHandler,
		Internal:        internalAIHandler,
		Integrations:    integrationsHandler,
		Reranker:        reranker,
	}
	m.Quota = &QuotaHandlers{Handler: quotaHandler}

	if cfg.AI.Enabled && cfg.AI.AutoLoadOnBoot && aiHandler != nil {
		if err := aiHandler.AutoLoadChatbots(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to auto-load AI chatbots")
		} else {
			log.Info().Msg("AI chatbots auto-loaded successfully")
		}
	}

	if vectorHandler != nil {
		registry.Register(vectorHandler)
	}
	if aiChatHandler != nil {
		registry.Register(aiChatHandler)
	}
	if kbStorage != nil {
		registry.Register(kbStorage)
	}
	return nil
}

func (m *AIModule) Shutdown(ctx context.Context) error {
	if m.Handlers != nil && m.Handlers.Conversations != nil {
		m.Handlers.Conversations.Close()
	}
	return nil
}
