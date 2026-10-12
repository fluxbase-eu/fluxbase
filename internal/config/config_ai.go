package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// AIConfig contains AI chatbot settings
type AIConfig struct {
	Enabled              bool          `mapstructure:"enabled"`                // Enable AI chatbot functionality
	ChatbotsDir          string        `mapstructure:"chatbots_dir"`           // Directory for chatbot definitions
	AutoLoadOnBoot       bool          `mapstructure:"auto_load_on_boot"`      // Load chatbots from filesystem at boot
	DefaultMaxTokens     int           `mapstructure:"default_max_tokens"`     // Default max tokens per request
	DefaultModel         string        `mapstructure:"default_model"`          // Default AI model
	QueryTimeout         time.Duration `mapstructure:"query_timeout"`          // SQL query execution timeout
	MaxRowsPerQuery      int           `mapstructure:"max_rows_per_query"`     // Max rows returned per query
	ConversationCacheTTL time.Duration `mapstructure:"conversation_cache_ttl"` // TTL for conversation cache
	MaxConversationTurns int           `mapstructure:"max_conversation_turns"` // Max turns per conversation
	SyncAllowedIPRanges  []string      `mapstructure:"sync_allowed_ip_ranges"` // IP CIDR ranges allowed to sync chatbots

	// Provider Configuration (read-only in dashboard when set)
	// If ProviderType is set, a config-based provider will be added to the list
	ProviderType  string `mapstructure:"provider_type"`  // Provider type: openai, azure, ollama, anthropic
	ProviderName  string `mapstructure:"provider_name"`  // Display name for config provider
	ProviderModel string `mapstructure:"provider_model"` // Default model for config provider

	// Embedding Configuration (for vector search)
	EmbeddingEnabled  bool   `mapstructure:"embedding_enabled"`  // Enable embedding generation for vector search
	EmbeddingProvider string `mapstructure:"embedding_provider"` // Embedding provider: openai, azure, ollama (defaults to ProviderType)
	EmbeddingModel    string `mapstructure:"embedding_model"`    // Embedding model: text-embedding-3-small, text-embedding-3-large, etc.

	// OpenAI Settings
	OpenAIAPIKey         string `mapstructure:"openai_api_key"`
	OpenAIOrganizationID string `mapstructure:"openai_organization_id"`
	OpenAIBaseURL        string `mapstructure:"openai_base_url"`

	// Azure Settings
	AzureAPIKey         string `mapstructure:"azure_api_key"`
	AzureEndpoint       string `mapstructure:"azure_endpoint"`
	AzureDeploymentName string `mapstructure:"azure_deployment_name"`
	AzureAPIVersion     string `mapstructure:"azure_api_version"`

	// Azure Embedding Settings (optional, falls back to Azure Settings)
	AzureEmbeddingDeploymentName string `mapstructure:"azure_embedding_deployment_name"` // Separate deployment for embeddings

	// Ollama Settings
	OllamaEndpoint string `mapstructure:"ollama_endpoint"`
	OllamaModel    string `mapstructure:"ollama_model"`

	// Tavily websearch Settings (for Web Agent specialist)
	// When TavilyAPIKey is set, a read-only synthetic FROM_CONFIG tool
	// integration row is exposed so instance-level deployments work the
	// same as multi-tenant dashboard deployments. Mirror the OpenAI/
	// Ollama env-config pattern.
	TavilyAPIKey       string `mapstructure:"tavily_api_key"`
	TavilyDefaultDepth string `mapstructure:"tavily_default_depth"` // "basic" (default) or "advanced"
	TavilyBaseURL      string `mapstructure:"tavily_base_url"`      // optional override

	// OCR Configuration (for image-based PDF extraction in knowledge bases)
	OCREnabled   bool     `mapstructure:"ocr_enabled"`   // Enable OCR for image-based PDFs
	OCRProvider  string   `mapstructure:"ocr_provider"`  // OCR provider: tesseract
	OCRLanguages []string `mapstructure:"ocr_languages"` // Default languages for OCR (e.g., ["eng", "deu"])

	// RAG Configuration (for retrieval-augmented generation)
	RAGGraphBoostWeight float64 `mapstructure:"rag_graph_boost_weight"` // How much to weight entity matches vs vector similarity (0.0-1.0, default 0)

	// Entity Extraction Configuration (knowledge graph quality)
	// Strategy: rules (regex, default), llm (model-backed with rule fallback),
	// hybrid (both merged). Per-KB override via entity_extraction_strategy.
	EntityExtractionStrategy     string        `mapstructure:"entity_extraction_strategy"`       // rules | llm | hybrid
	EntityExtractionMaxTokens    int           `mapstructure:"entity_extraction_max_tokens"`     // Max tokens per LLM extraction request (default 2000)
	EntityExtractionMinTextChars int           `mapstructure:"entity_extraction_min_text_chars"` // Skip the LLM below this text length (default 200)
	EntityExtractionTimeout      time.Duration `mapstructure:"entity_extraction_timeout"`        // Per-extraction-request timeout (default 60s)

	// Rerank Configuration (RAG retrieval quality). When enabled, retrieved
	// chunks are re-ordered by a relevance model before being used. Provider:
	// llm (uses the default chat provider, no extra vendor) or cohere (Cohere
	// Rerank API). Per-KB opt-out via rerank_enabled.
	RerankEnabled  bool          `mapstructure:"rerank_enabled"`  // Enable reranking (default false)
	RerankProvider string        `mapstructure:"rerank_provider"` // llm | cohere (default llm)
	RerankModel    string        `mapstructure:"rerank_model"`    // Model override (cohere default: rerank-v3.5; llm: provider default)
	RerankAPIKey   string        `mapstructure:"rerank_api_key"`  // API key for cohere provider
	RerankBaseURL  string        `mapstructure:"rerank_base_url"` // Optional base URL override (cohere)
	RerankTopN     int           `mapstructure:"rerank_top_n"`    // Max results kept after reranking (default 10)
	RerankTimeout  time.Duration `mapstructure:"rerank_timeout"`  // Per-rerank-request timeout (default 10s)

	// Response Cache Configuration (AI chatbot cost/latency optimization).
	// Opt-in per chatbot via @fluxbase:response-cache; entries expire after
	// this TTL and are invalidated when linked KB documents change. Set to 0s
	// to disable globally.
	ResponseCacheTTL time.Duration `mapstructure:"response_cache_ttl"` // default 1h
}

// Validate validates AI configuration
func (ac *AIConfig) Validate() error {
	// Validate chatbots directory
	if ac.ChatbotsDir == "" {
		return fmt.Errorf("chatbots_dir cannot be empty")
	}

	// Validate token settings
	if ac.DefaultMaxTokens <= 0 {
		return fmt.Errorf("default_max_tokens must be positive, got: %d", ac.DefaultMaxTokens)
	}

	// Validate query timeout
	if ac.QueryTimeout <= 0 {
		return fmt.Errorf("query_timeout must be positive, got: %v", ac.QueryTimeout)
	}

	// Validate max rows per query
	if ac.MaxRowsPerQuery <= 0 {
		return fmt.Errorf("max_rows_per_query must be positive, got: %d", ac.MaxRowsPerQuery)
	}

	// Validate conversation settings
	if ac.ConversationCacheTTL <= 0 {
		return fmt.Errorf("conversation_cache_ttl must be positive, got: %v", ac.ConversationCacheTTL)
	}
	if ac.MaxConversationTurns <= 0 {
		return fmt.Errorf("max_conversation_turns must be positive, got: %d", ac.MaxConversationTurns)
	}

	// Warn if max rows is very high
	if ac.MaxRowsPerQuery > 10000 {
		log.Warn().Int("max_rows_per_query", ac.MaxRowsPerQuery).Msg("max_rows_per_query is over 10000 - large result sets may impact performance")
	}

	// Validate entity extraction strategy
	switch strings.ToLower(strings.TrimSpace(ac.EntityExtractionStrategy)) {
	case "", "rules", "llm", "hybrid":
	default:
		return fmt.Errorf("entity_extraction_strategy must be one of rules, llm, hybrid, got: %q", ac.EntityExtractionStrategy)
	}

	// Warn when an LLM strategy is selected without a chat provider source
	if ac.EntityExtractionStrategy != "" && ac.EntityExtractionStrategy != "rules" && ac.ProviderType == "" {
		log.Warn().Str("strategy", ac.EntityExtractionStrategy).Msg("entity_extraction_strategy uses the LLM but no ai.provider_type is configured; extraction will fall back to rules until a default provider is set (admin UI/CLI or config)")
	}

	// Validate rerank configuration
	if ac.RerankEnabled {
		switch strings.ToLower(strings.TrimSpace(ac.RerankProvider)) {
		case "", "llm":
		case "cohere":
			if ac.RerankAPIKey == "" && ac.RerankBaseURL == "" {
				return fmt.Errorf("rerank_provider cohere requires rerank_api_key (or rerank_base_url for a self-hosted compatible API)")
			}
		default:
			return fmt.Errorf("rerank_provider must be llm or cohere, got: %q", ac.RerankProvider)
		}
	}

	return nil
}
