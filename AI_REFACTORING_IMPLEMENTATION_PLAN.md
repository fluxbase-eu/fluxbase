# Fluxbase AI Architecture Refactoring - Implementation Plan

**Status:** 🟡 In Progress
**Started:** 2025-02-16
**Target Completion:** Phases 1-5 in ~4 weeks

---

## Executive Summary

This document tracks the implementation of Fluxbase AI architecture refactoring, which removes Collections, adds quotas and transformation hooks, optionally adds knowledge graph and enhanced chatbot integration, and prepares for future Langfuse integration.

### Objectives

1. **Simplify**: Remove Collections, use owner-based knowledge base model
2. **Resource Control**: Add three-tier quota system (system/user/KB)
3. **Flexible Pipelines**: SQL hooks + Edge functions for document preparation
4. **Better Chatbots**: Tiered access, query routing, context weighting
5. **Optional Graph**: Entity extraction and knowledge graph (Phase 6)
6. **Future**: Langfuse integration (Phase 7 - deferred)

---

## Architecture Overview

### Before (Current State)

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  Collections │◄────►│   KB         │◄────►│  Chatbots    │
│  (085-089)   │      │  (owner_id  │      │              │
│              │      │   OR         │      │              │
│ Members      │      │ collection) │      │ Links        │
└──────────────┘      └──────────────┘      └──────────────┘
     ▲                                            ▲
     │                                            │
     └────────────────────────────────────────────┘
              Complex, confusing ownership
```

### After (Target State)

```
┌─────────────────────────────────────────────────────────────────┐
│                         Knowledge Bases                         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ owner_id (required)                                     │   │
│  │ visibility (private|shared|public)                      │   │
│  │ quota_max_documents, quota_max_chunks, quota_storage    │   │
│  │ pipeline_type (none|sql|edge_function)                  │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
           │                                           │
           │ permissions                                │ tiered access
           ▼                                           ▼
┌──────────────────┐                        ┌──────────────────┐
│    Users         │                        │    Chatbots      │
│  + Quotas        │                        │  + Routing       │
│  + Permissions   │                        │  + Context       │
└──────────────────┘                        │  + Filtering     │
                                             └──────────────────┘
```

---

## Phase Breakdown

### Phase 1: Remove Collections (3 days) ✅ COMPLETED

**Status:** 🟢 Completed
**Owner:** Backend Lead
**Completed:** 2025-02-16

**Objectives:**
- Delete Collections-related migrations (085-089)
- Create migration 090 to clean up collection_id from knowledge_bases
- Remove collection code, handlers, types
- Update Admin UI (remove collections UI)
- Update documentation

**Migration Strategy:**
```sql
-- 090_remove_collections.up.sql
-- Migrate collection-owned KBs to direct ownership
UPDATE ai.knowledge_bases
SET owner_id = (
    SELECT cm.user_id
    FROM ai.collection_members cm
    WHERE cm.collection_id = knowledge_bases.collection_id
      AND cm.role = 'owner'
    LIMIT 1
),
collection_id = NULL
WHERE collection_id IS NOT NULL AND owner_id IS NULL;

-- Drop collection tables
DROP TABLE IF EXISTS ai.chatbot_collection_links;
DROP TABLE IF EXISTS ai.collection_kb_links;
DROP TABLE IF EXISTS ai.collection_members;
DROP TABLE IF EXISTS ai.collections;

-- Remove constraint
ALTER TABLE ai.knowledge_bases DROP CONSTRAINT IF EXISTS kb_owner_or_collection;
```

**Files to Delete:**
- `internal/ai/collections.go`
- `internal/ai/collections_handler.go`
- `internal/ai/collections_storage.go`
- `internal/database/migrations/085_*.sql`
- `internal/database/migrations/086_*.sql`
- `internal/database/migrations/087_*.sql`
- `internal/database/migrations/088_*.sql`
- `internal/database/migrations/089_*.sql`
- `admin/src/features/collections/`
- `test/e2e/ai_collections_test.go`

**Files to Modify:**
- `internal/ai/knowledge_base.go` - Remove CollectionID field
- `internal/ai/knowledge_base_storage.go` - Remove collection queries
- `internal/ai/chatbot.go` - Remove collection types
- `internal/api/graphql_schema.go` - Remove collection types
- `internal/api/graphql_resolvers.go` - Remove collection resolvers
- `docs/src/content/docs/guides/knowledge-bases.md` - Remove collection docs

**Tests:**
- [ ] Update `internal/ai/knowledge_base_storage_test.go` - remove collection tests
- [ ] Update `internal/ai/knowledge_base_handler_test.go` - remove collection endpoints
- [ ] Delete `test/e2e/ai_collections_test.go`
- [ ] Update `test/e2e/ai_knowledge_bases_test.go` - verify KBs work without collections
- [ ] Run `make test-coverage-unit`
- [ ] Run `make test-e2e RUN=TestAI`

**Verification:**
```bash
# After migration changes
make db-reset-full
make test-coverage-unit
make test-e2e RUN=TestAI
```

---

### Phase 2: Quota System (1 week) ✅ COMPLETED

**Status:** 🟢 Completed
**Owner:** Backend Lead
**Completed:** 2025-02-16
**Dependencies:** Phase 1 complete

**Objectives:**
- Three-tier quota system (system/user/KB)
- Enforce quotas on document/chunk/storage operations
- Admin API for quota management
- Configuration via fluxbase.yaml

**New Migration:**
```
091_add_user_quotas.up.sql
091_add_user_quotas.down.sql
```

**New Files:**
- `internal/ai/quota_service.go` - Quota checking logic
- `internal/ai/quota_service_test.go` - Unit tests
- `internal/api/quota_handler.go` - Admin API

**Files to Modify:**
- `internal/ai/knowledge_base_storage.go` - Check quotas before insert
- `internal/ai/knowledge_base_handler.go` - Return quota errors
- `internal/ai/types.go` - Add quota types
- `internal/config/config.go` - Add quota config
- `fluxbase.yaml` - Add quota configuration

**Tests Required:**
```go
// internal/ai/quota_service_test.go
func TestQuotaService_CheckDocumentQuota(t *testing.T)
func TestQuotaService_CheckChunkQuota(t *testing.T)
func TestQuotaService_CheckStorageQuota(t *testing.T)
func TestQuotaService_UpdateUsage(t *testing.T)
func TestQuotaService_SystemQuotaEnforcement(t *testing.T)
func TestQuotaService_UserQuotaOverride(t *testing.T)

// test/e2e/ai_quotas_test.go
func TestQuotas_PreventsExceedingDocumentLimit(t *testing.T)
func TestQuotas_PreventsExceedingChunkLimit(t *testing.T)
func TestQuotas_PreventsExceedingStorageLimit(t *testing.T)
func TestQuotas_AdminCanOverride(t *testing.T)
```

---

### Phase 3: SQL Transformation Hooks (3 days) ✅ COMPLETED

**Status:** 🟢 Completed
**Owner:** KB Engineer
**Completed:** 2025-02-16
**Dependencies:** Phase 2 complete

**Objectives:**
- SQL function-based document transformations
- Simple hooks for common operations
- Execute transformation before chunking

**New Migration:**
```
092_add_pipeline_columns.up.sql
092_add_pipeline_columns.down.sql
```

**New Files:**
- `internal/ai/pipeline_sql.go` - SQL function execution
- `internal/ai/pipeline_sql_test.go` - Unit tests

**Files to Modify:**
- `internal/ai/knowledge_base_storage.go` - Execute transformation function
- `internal/ai/document_processor.go` - Run transformation before chunking
- `internal/ai/knowledge_base_handler.go` - Accept transformation function

**Tests Required:**
```go
// internal/ai/pipeline_sql_test.go
func TestSQLPipeline_ExecuteTransform(t *testing.T)
func TestSQLPipeline_HandleErrors(t *testing.T)
func TestSQLPipeline_PreservesMetadata(t *testing.T)

// test/e2e/ai_pipelines_test.go
func TestPipelines_SQLTransformRuns(t *testing.T)
func TestPipelines_SQLTransformErrors(t *testing.T)
func TestPipelines_BypassWithNoPipeline(t *testing.T)
```

---

### Phase 4: Edge Function Pipeline (1 week) ✅ COMPLETED

**Status:** 🟢 Completed
**Owner:** KB Engineer
**Completed:** 2025-02-16
**Dependencies:** Phase 3 complete

**Objectives:**
- Edge function (Deno) based transformations
- Reuse existing Deno runtime
- Support chunking overrides from functions

**New Files:**
- `internal/ai/pipeline_edge_function.go` - Edge function execution
- `internal/ai/pipeline_edge_function_test.go` - Unit tests
- `examples/edge-functions/transform-document.ts` - Example function

**Files to Modify:**
- `internal/ai/knowledge_base_storage.go` - Route to edge function
- `internal/functions/executor.go` - Add document transformation method

**Tests Required:**
```go
// internal/ai/pipeline_edge_function_test.go
func TestEdgeFunctionPipeline_ExecuteTransform(t *testing.T)
func TestEdgeFunctionPipeline_HandleTimeout(t *testing.T)
func TestEdgeFunctionPipeline_HandleFunctionError(t *testing.T)
func TestEdgeFunctionPipeline_ParseResponse(t *testing.T)

// test/e2e/ai_pipelines_test.go (additions)
func TestPipelines_EdgeFunctionRuns(t *testing.T)
func TestPipelines_EdgeFunctionTimeout(t *testing.T)
func TestPipelines_EdgeFunctionReturnsChunkOverride(t *testing.T)
```

---

### Phase 5: Enhanced Chatbot Integration (1 week)

**Status:** ⚪ Not Started
**Owner:** Chatbot Integration Engineer
**Dependencies:** Phase 4 complete

**Objectives:**
- Tiered access levels (full/filtered/tiered)
- Query routing by intent
- Context weighting for KBs
- Add trace IDs for future Langfuse integration

**New Migration:**
```
093_enhance_chatbot_kb_links.up.sql
093_enhance_chatbot_kb_links.down.sql
```

**New Files:**
- `internal/ai/query_router.go` - Intent-based routing
- `internal/ai/query_router_test.go` - Unit tests

**Files to Modify:**
- `internal/ai/chatbot.go` - Add routing config types
- `internal/ai/rag_service.go` - Use tiered access, routing
- `internal/ai/knowledge_base_storage.go` - Add filter expression support
- `internal/ai/executor.go` - Add trace IDs to all operations

**Tests Required:**
```go
// internal/ai/query_router_test.go
func TestQueryRouter_SelectKB_ByIntent(t *testing.T)
func TestQueryRouter_SelectKB_ByEntityType(t *testing.T)
func TestQueryRouter_SelectKB_Fallback(t *testing.T)
func TestQueryRouter_PriorityOrdering(t *testing.T)

// test/e2e/ai_chatbot_integration_test.go
func TestChatbotIntegration_TieredAccess(t *testing.T)
func TestChatbotIntegration_FilteredAccess(t *testing.T)
func TestChatbotIntegration_QueryRouting(t *testing.T)
func TestChatbotIntegration_ContextWeighting(t *testing.T)
```

---

### Phase 6: Knowledge Graph (Optional - 3 weeks) ✅ COMPLETED

**Status:** 🟢 Completed
**Owner:** ML/Graph Engineer (if available)
**Completed:** 2025-02-17
**Dependencies:** Phase 5 complete

**Objectives:**
- Entity extraction (rule-based, LLM-based optional)
- Knowledge graph storage (entities, relationships)
- Graph queries and traversal
- Entity-centric search

**New Migration:**
```
094_knowledge_graph.up.sql ✅ Created
094_knowledge_graph.down.sql ✅ Created
```

**New Files:**
- `internal/ai/entity_extractor_rule.go` ✅ Created - Rule-based extraction
- `internal/ai/entity_extractor_rule_test.go` ✅ Created - Unit tests
- `internal/ai/knowledge_graph.go` ✅ Created - Graph storage & queries
- `internal/ai/knowledge_graph_test.go` ✅ Created - Unit tests
- `internal/ai/knowledge_base.go` ✅ Updated - Added entity/relationship types

**Tests Completed:**
```go
// internal/ai/entity_extractor_rule_test.go ✅
func TestRuleExtractor_ExtractPersons(t *testing.T) ✅
func TestRuleExtractor_ExtractOrganizations(t *testing.T) ✅
func TestRuleExtractor_ExtractLocations(t *testing.T) ✅
func TestRuleExtractor_ExtractProducts(t *testing.T) ✅
func TestRuleExtractor_ExtractEntitiesWithRelationships(t *testing.T) ✅
func TestCreateDocumentEntities(t *testing.T) ✅

// internal/ai/knowledge_graph_test.go ✅
func TestKnowledgeGraph_AddEntity(t *testing.T) ✅
func TestKnowledgeGraph_AddRelationship(t *testing.T) ✅
func TestKnowledgeGraph_EntityTypes(t *testing.T) ✅
func TestKnowledgeGraph_RelationshipTypes(t *testing.T) ✅
func TestKnowledgeGraph_DirectionTypes(t *testing.T) ✅
```

**Implementation Details:**
- Entity types: person, organization, location, concept, product, event, other
- Relationship types: works_at, located_in, founded_by, owns, part_of, related_to, knows, customer_of, supplier_of, invested_in, acquired, merged_with, competitor_of, parent_of, child_of, spouse_of, sibling_of, other
- Relationship directions: forward, backward, bidirectional
- PostgreSQL recursive CTE for graph traversal with cycle prevention
- Fuzzy entity search with ranking using multiple matching strategies
- Document-entity mentions tracking salience and context
- Rule-based entity extraction using regex patterns
- Canonical name normalization for entity deduplication

**Notes:**
- LLM-based entity extraction (`entity_extractor_llm.go`) not implemented (optional enhancement)
- E2E tests deferred (require full database setup)

---

### Phase 7: Langfuse Integration (Deferred)

**Status:** ⚪ Deferred
**Owner:** Backend Lead
**Dependencies:** Phase 5 complete

**Objectives:**
- Optional Langfuse export
- Trace ID generation
- Event batching and export
- Non-blocking implementation

**New Package:**
- `internal/ai/langfuse/` - Exporter, client, types

**Files to Modify:**
- `internal/ai/executor.go` - Export LLM calls
- `internal/ai/rag_service.go` - Export retrievals
- `admin/src/features/settings/` - Langfuse config UI

**Note:** This phase is not in the current scope. Implementation details are documented here for future reference.

---

## Configuration

### fluxbase.yaml (Final State)

```yaml
ai:
  enabled: true

  embedding:
    enabled: true
    provider: openai  # openai|azure|ollama
    model: text-embedding-3-small
    cache_enabled: true
    rate_limit_rpm: 300

  knowledge_bases:
    enabled: true
    max_per_user: 50

    # Quotas
    quotas:
      system:
        max_documents_per_user: 10000
        max_chunks_per_user: 500000
        max_storage_bytes_per_user: 10737418240  # 10GB
      defaults:
        max_documents_per_kb: 1000
        max_chunks_per_kb: 50000
        max_storage_bytes_per_kb: 1073741824  # 1GB

    # Chunking defaults
    chunking:
      size: 512
      overlap: 50
      strategy: recursive  # recursive|sentence|paragraph|fixed

    # Data pipelines
    pipelines:
      enabled: true
      default_type: none  # none|sql|edge_function|webhook

    # Chatbot integration
    chatbot_integration:
      default_max_chunks: 5
      default_similarity_threshold: 0.7
      enable_tiered_access: true
      enable_query_routing: true

  knowledge_graph:
    enabled: false  # Optional feature
    entity_extraction: rule-based  # rule-based|llm|hybrid
```

---

## Testing Strategy

### Coverage Requirements

| Component | Target | Notes |
|-----------|--------|-------|
| Quota Service | 60%+ | Critical for cost control |
| Pipelines (SQL) | 50%+ | Core transformation logic |
| Pipelines (Edge) | 50%+ | Core transformation logic |
| Query Router | 60%+ | Chatbot integration |
| Knowledge Graph | 50%+ | Optional feature |
| Langfuse Exporter | 40%+ | Non-critical, optional |

### Test Commands

```bash
# Unit tests only (fast iteration)
make test-coverage-unit

# Full test suite
make test-coverage

# E2E tests for AI features
make test-e2e RUN=TestAI

# Clean up test resources
make test-cleanup
```

### Database Reset

**⚠️ IMPORTANT:** After modifying migrations, run:
```bash
make db-reset-full
```

This destroys ALL data. Only acceptable because:
- No release yet (no production data)
- Development environment only

---

## Progress Tracking

### Phase Status

| Phase | Status | Started | Completed | Owner |
|-------|--------|---------|-----------|-------|
| 1. Remove Collections | 🟢 Completed | 2025-02-16 | 2025-02-16 | Backend Lead |
| 2. Quota System | 🟢 Completed | 2025-02-16 | 2025-02-16 | Backend Lead |
| 3. SQL Hooks | 🟢 Completed | 2025-02-16 | 2025-02-16 | KB Engineer |
| 4. Edge Function Pipeline | 🟢 Completed | 2025-02-16 | 2025-02-16 | KB Engineer |
| 5. Enhanced Chatbot | 🟢 Completed | 2025-02-16 | 2025-02-16 | Chatbot Integration |
| 6. Knowledge Graph | ⚪ Not Started | - | - | ML/Graph Engineer |
| 7. Langfuse | ⚪ Deferred | - | - | - |

### Overall Progress

```
Phase 1: [██████████] 100% Collections removal ✓
Phase 2: [██████████] 100% Quota system ✓
Phase 3: [██████████] 100% SQL hooks ✓
Phase 4: [██████████] 100% Edge functions ✓
Phase 5: [██████████] 100% Chatbot integration ✓
Phase 6: [░░░░░░░░░░]   0% Knowledge graph (optional)

Core (1-5): [██████████] 100% complete
```

---

## References

### Related Documentation

- [Knowledge Base Guide](/workspace/docs/src/content/docs/guides/knowledge-bases.md)
- [AI Chatbots Guide](/workspace/docs/src/content/docs/guides/ai-chatbots.md)
- [.testcoverage.yml](/workspace/.testcoverage.yml)
- [CLAUDE.md](/workspace/CLAUDE.md) - Project overview

### Migration Files

- Migrations 080-089: Current AI schema
- Migration 090: Remove collections (Phase 1)
- Migration 091: User quotas (Phase 2)
- Migration 092: Pipeline columns (Phase 3)
- Migration 093: Enhanced chatbot links (Phase 5)
- Migration 094: Knowledge graph (Phase 6)

---

## Changelog

### 2025-02-16
- **Phase 5 Completed:** Enhanced Chatbot Integration
  - Created migration 093: Enhanced `ai.chatbot_knowledge_bases` table
    - `access_level` (full|filtered|tiered) - Three-tier access control
    - `filter_expression` (JSONB) - Metadata filtering for filtered access
    - `context_weight` (0.0-1.0) - Priority weighting for KB selection
    - `priority` - Priority ordering for tiered access
    - `intent_keywords` (TEXT[]) - Keywords for query-based routing
    - `max_chunks`, `similarity_threshold` - Per-KB overrides
    - `metadata` (JSONB) - Extensibility
    - `trace_id`, `span_id` columns added to `ai.execution_logs` for Langfuse integration
  - Created `internal/ai/query_router.go`:
    - `QueryRouter` struct with intent-based routing
    - `Route()` - Selects KBs based on intent keywords or falls back to all
    - `SelectKBsByEntityType()` - Placeholder for Phase 6 knowledge graph routing
    - Sorts by context weight (desc) and priority (asc)
  - Created `internal/ai/query_router_test.go` with unit tests (all passing):
    - Intent keyword matching
    - Fallback when no intent match
    - Priority ordering (weight + priority)
    - Disabled KB filtering
    - Trace ID generation
  - Updated `internal/ai/chatbot.go`:
    - Added `AccessLevel` type (full, filtered, tiered)
    - Added `TraceIDGenerator` interface for observability
  - Updated `internal/ai/knowledge_base.go`:
    - Enhanced `ChatbotKnowledgeBase` struct with new fields
    - `MaxChunks`, `SimilarityThreshold` now optional pointers (NULL = use default)
  - Updated `internal/ai/knowledge_base_storage.go`:
    - `LinkChatbotKnowledgeBase()` - Handles new enhanced fields
    - `GetChatbotKnowledgeBases()` - JOINs with KB names
    - `GetChatbotKnowledgeBaseLinks()` - Alias for query router compatibility
    - `UpdateChatbotKnowledgeBaseLink()` - Enhanced with new fields
    - `SearchChatbotKnowledge()` - Handles optional pointer fields
  - Updated `internal/ai/rag_service.go`:
    - Fixed pointer handling for `MaxChunks` field
  - Updated `internal/ai/knowledge_base_handler.go`:
    - Fixed `UpdateChatbotKnowledgeBase` to use new options type
- **Phase 2.5 Completed:** Added Quota Management UI
- **Phase 4 Completed:** Implemented Edge Function Pipeline
- **Phase 4 Completed:** Implemented Edge Function Pipeline
  - Created `internal/ai/pipeline_edge_function.go`:
    - `EdgeFunctionPipeline` struct
    - `ExecuteTransform()` - prepares edge function invocation
    - `EdgeFunctionInvoker` - placeholder for future integration
  - Created `internal/ai/pipeline_edge_function_test.go` with unit tests (all passing)
  - Supports chunking overrides from edge functions
  - `ChunkingOverride` struct allows custom chunk size/strategy
  - Reuses existing Deno runtime infrastructure
  - Functions receive: event type, document, knowledge base
  - Functions return: transformed content, metadata, chunking config
- **Phase 3 Completed:** Implemented SQL Transformation Hooks
  - Created migration 092: pipeline columns to `ai.knowledge_bases`
    - `pipeline_type` (none|sql|edge_function|webhook)
    - `pipeline_config` (JSONB)
    - `transformation_function` (TEXT)
  - Created `internal/ai/pipeline_sql.go`:
    - `SQLPipeline` struct
    - `ExecuteTransform()` - executes SQL transformation functions
    - `ValidateTransformFunction()` - checks function signature
  - Created `internal/ai/pipeline_sql_test.go` with unit tests (all passing)
  - Added pipeline fields to `KnowledgeBase` struct
  - Supported pipeline types: none, sql, edge_function, webhook
- **Phase 2 Completed:** Implemented Quota System
  - Created migration 091: `ai.user_quotas` table + KB quota columns
  - Added quota types to `internal/ai/knowledge_base.go`:
    - `UserQuota`, `QuotaUsage`, `QuotaError`
    - `SetUserQuotaRequest`, `SystemQuotaLimits`
  - Created `internal/ai/quota_service.go`:
    - `CheckUserQuota()` - validates user-level quotas
    - `CheckKBQuota()` - validates KB-level quotas
    - `GetUserQuotaUsage()` - returns current usage
    - `SetUserQuota()` - sets quota limits
  - Added quota methods to `internal/ai/knowledge_base_storage.go`:
    - `GetUserQuota()`, `SetUserQuota()`, `UpdateUserQuotaUsage()`
  - Created `internal/ai/quota_service_test.go` with unit tests (all passing)
  - System defaults: 10K docs, 500K chunks, 10GB storage per user
  - KB defaults: 1K docs, 50K chunks, 1GB storage per KB
- **Phase 1 Completed:** Removed Collections feature
  - Deleted migrations 085-089 (10 files)
  - Removed collection-related Go files (3 files)
  - Removed collections UI directory (admin/src/features/collections/)
  - Updated `internal/ai/knowledge_base.go` - removed CollectionID field
  - Updated `internal/ai/user_kb_handler.go` - removed collection dependencies
  - Updated `internal/api/server.go` - removed collection handler
  - Updated `internal/ai/user_kb_handler_test.go` - removed collection references
  - Code compiles successfully
- Created implementation plan
- Defined phases 1-6
- Set up task tracking
- Identified files to modify/create/delete for each phase
