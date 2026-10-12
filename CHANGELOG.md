# Changelog

All notable changes to Fluxbase are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
for SDK packages; the server itself is versioned `YYYY.M.D` (see `VERSION`).

## Unreleased

### Added — AI retrieval quality

- **LLM-based entity extraction** for knowledge graphs: new
  `ai.entity_extraction_strategy` config (`rules`, `llm`, `hybrid`) with a
  per-KB override (`entity_extraction_strategy` on the KB). The LLM extractor
  falls back to rules on provider failure so ingestion never breaks.
- **Reciprocal rank fusion** for hybrid knowledge-base search: weighted RRF
  (k=60, normalized 0–1) replaces weighted linear score fusion. The similarity
  threshold now filters on raw cosine similarity; `KeywordBoost` is expressed
  as exact-match-first rank ordering.
- **Pluggable reranker** (`ai.rerank_enabled`, off by default) with two
  providers: `llm` (listwise rerank via the default chat provider) and
  `cohere` (Cohere Rerank v2 API). Per-KB opt-out via `rerank_enabled`.
  Rerank latency is recorded in `retrieval_log.rerank_duration_ms` and
  `fluxbase_ai_rerank_*` metrics.
- **Semantic response cache** for chatbots: opt-in per chatbot via the
  `@fluxbase:response-cache` annotation. Single-turn questions are served on
  an exact hash match or cosine ≥ 0.95 embedding match; entries expire after
  `ai.response_cache_ttl` (default 1h) and are invalidated when a linked KB
  document changes. New `fluxbase_ai_response_cache_{hits,misses}_total`
  metrics.

### Added — MCP

- **Server-initiated notifications** over the SSE GET channel: long-lived
  `GET /mcp` streams (previously 501) now receive
  `notifications/tools/list_changed` and `notifications/resources/list_changed`
  when custom tools/resources change, with 30s keep-alive pings.

### Fixed — SDK / API contract

- **User knowledge bases can now be updated and deleted**: added
  `PATCH`/`DELETE /api/v1/ai/knowledge-bases/{id}` (editor/owner gated).
  Previously only admin routes existed, so the SDK's `update`/`delete`
  methods 404'd.
- **`admin.ai.toggleChatbot`** now sends `POST`, matching the backend route
  (the previous `PUT` returned 405).
- **User KB search request shape** aligned with the backend binding: the
  endpoint accepts `query`, `limit`, `mode`, `semantic_weight`, `threshold`,
  and `rerank` (the SDK previously sent `max_chunks`/`threshold`, which the
  handler silently dropped).

### Added — SDKs

- Core SDK: `admin.ai.getMetrics`, admin conversations + messages, audit log,
  table list/export, table-export sync-config CRUD + trigger, document
  permission grant/list/revoke; `knowledgeBase.share/listPermissions/
  revokePermission/debugSearch/listChatbots`; `vector.getCapabilities`.
- React hooks: new `use-admin-ai` module (metrics, conversations, audit,
  table export, sync configs, document permissions, list queries), plus
  `useUsage`, `useUpdateConversation`, `useSearchEntities`,
  `useEntityRelationships`, `useUpdateDocument`,
  `useDeleteDocumentsByFilter`, `useKBDebugSearch`, `useShareKB`,
  `useRevokeKBPermission`, `useKBLinkedChatbots`, `useVectorCapabilities`.
- Svelte stores: AI chatbots/conversations, knowledge-base list/search/
  upload/delete, `vectorCapabilities`.
- Vue composables: `useFluxbaseAIChatbots`, `useFluxbaseAIConversations`,
  `useFluxbaseAIUsage`, `useFluxbaseKnowledgeBases`, `useFluxbaseKBSearch`,
  `useFluxbaseVectorSearch`, `useFluxbaseVectorCapabilities`.
- Kotlin: new `ai` module (chatbots, conversations, usage) and
  `knowledgeBase` module (CRUD, sharing, documents, hybrid search, knowledge
  graph); `vector.search` gains `filters`; `vector.getCapabilities` added.

### Changed — API surface

- The user knowledge-base routes and the admin AI surface are now part of the
  hand-curated OpenAPI spec and guarded by the `contract-drift` CI job
  (`sdk/src/types-generated.ts` regenerated).

## 2026.9.10

Release for version 2026.9.10 (see `VERSION`); history before this file was
introduced lives in the git log (`git log v2026.9.10`).
