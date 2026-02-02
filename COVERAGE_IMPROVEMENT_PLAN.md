# Fluxbase Code Coverage Improvement Plan

## Current State Analysis

**Reported Coverage:** 23% (according to Codecov)

**Codebase Metrics:**
- Source files: 294 files (131,664 lines)
- Test files: 166 files (71,955 lines)
- Test-to-source ratio: 55% (good amount of test code exists)

**Key Finding:** The low coverage percentage despite substantial test code indicates that existing tests don't cover enough code paths, and many large critical files lack tests entirely.

---

## Coverage by Module

### Well-Tested Modules (ratio >= 1.0)
| Module | Source Files | Test Files | Status |
|--------|--------------|------------|--------|
| auth | 40 | 40 | Excellent |
| middleware | 16 | 17 | Excellent |
| database | 5 | 5 | Good |
| logging | 5 | 5 | Good |
| secrets | 2 | 2 | Good |
| settings | 2 | 2 | Good |
| observability | 2 | 2 | Good |
| webhook | 2 | 3 | Good |
| crypto | 1 | 1 | Good |
| query | 1 | 1 | Good |

### Modules Needing Improvement

| Module | Source Files | Test Files | Gap | Priority |
|--------|--------------|------------|-----|----------|
| **api** | 62 | 28 | 34 files | **CRITICAL** |
| **ai** | 36 | 9 | 27 files | **HIGH** |
| branching | 7 | 2 | 5 files | HIGH |
| rpc | 8 | 3 | 5 files | HIGH |
| jobs | 7 | 3 | 4 files | MEDIUM |
| pubsub | 5 | 2 | 3 files | MEDIUM |
| config | 4 | 1 | 3 files | MEDIUM |
| storage | 12 | 6 | 6 files | MEDIUM |
| realtime | 10 | 8 | 2 files | LOW |
| scaling | 1 | 0 | 1 file | LOW (excluded) |

---

## Priority 1: API Module (34 untested files)

The `api` module is the largest (62 source files, 52,614 lines) and most critical for coverage improvement. Adding tests here will have the biggest impact on overall coverage.

### Untested Files (sorted by estimated impact)

**High-Impact Handlers (>700 lines each):**
1. `branch_handler.go` (972 lines) - Database branching API
2. `oauth_provider_handler.go` (938 lines) - OAuth provider management
3. `saml_provider_handler.go` (857 lines) - SAML SSO configuration
4. `servicekey_handler.go` (793 lines) - Service key management
5. `policy_handler.go` (786 lines) - RLS policy management
6. `vector_handler.go` (739 lines) - Vector/embedding operations

**Medium-Impact Handlers (300-700 lines):**
7. `graphql_handler.go` - GraphQL endpoint
8. `graphql_resolvers.go` (890 lines) - GraphQL resolvers
9. `graphql_schema.go` - Schema generation
10. `webhook_handler.go` - Webhook management
11. `ddl_handler.go` - DDL operations
12. `invitation_handler.go` - User invitations
13. `user_management_handler.go` - User CRUD
14. `admin_auth_handler.go` - Admin authentication
15. `admin_session_handler.go` - Admin sessions

**Settings/Config Handlers:**
16. `app_settings_handler.go`
17. `captcha_settings_handler.go`
18. `custom_settings_handler.go`
19. `email_settings_handler.go`
20. `email_template_handler.go`
21. `settings_handler.go`
22. `system_settings_handler.go`
23. `user_settings_handler.go`

**Storage-Related:**
24. `storage_chunked.go` - Chunked uploads
25. `storage_multipart.go` - Multipart uploads
26. `storage_sharing.go` - File sharing
27. `storage_streaming.go` - Streaming downloads
28. `storage_utils.go` - Storage utilities

**Other:**
29. `auth_middleware.go` - Auth middleware helpers
30. `auth_saml.go` - SAML authentication
31. `clientkey_handler.go` - Client key management
32. `custom_mcp_handler.go` - Custom MCP endpoints
33. `github_webhook_handler.go` - GitHub integration
34. `logging_handlers.go` - Log management
35. `realtime_admin_handler.go` - Realtime admin
36. `rest_geojson.go` - GeoJSON support
37. `rest_query.go` - Query helpers
38. `schema_handler.go` - Schema inspection
39. `validation_helpers.go` - Validation utilities
40. `vector_manager.go` - Vector management

---

## Priority 2: AI Module (27 untested files)

The `ai` module (36 source files, 19,385 lines) handles AI/ML features including embeddings, chat, and knowledge bases.

### Untested Files

**Core Services (High Priority):**
1. `handler.go` (2,002 lines) - Main AI HTTP handler
2. `storage.go` (1,501 lines) - AI data storage
3. `knowledge_base_storage.go` (1,286 lines) - KB storage
4. `knowledge_base_handler.go` (1,086 lines) - KB HTTP handler
5. `schema_builder.go` (676 lines) - Schema generation
6. `conversation.go` (559 lines) - Conversation management
7. `document_processor.go` (528 lines) - Document processing

**Provider Implementations:**
8. `provider_openai.go` (519 lines)
9. `provider_ollama.go` (508 lines)
10. `provider_azure.go` (418 lines)
11. `provider.go` - Provider interface

**Embedding Services:**
12. `embedding_service.go` - Main embedding service
13. `embedding_openai.go` - OpenAI embeddings
14. `embedding_ollama.go` - Ollama embeddings
15. `embedding_azure.go` - Azure embeddings

**Other:**
16. `audit.go` (366 lines) - AI audit logging
17. `knowledge_base.go` - KB core logic
18. `loader.go` - AI model loading
19. `executor.go` - AI execution
20. `mcp_executor.go` - MCP execution
21. `rag_service.go` (464 lines) - RAG implementation
22. `http_tool.go` - HTTP tool for AI
23. `settings_resolver.go` - AI settings
24. `ocr_provider.go` - OCR abstraction
25. `ocr_service.go` - OCR service

---

## Priority 3: Other Modules

### Branching Module (5 untested files)
1. `manager.go` - Branch creation/deletion
2. `router.go` - Connection routing
3. `scheduler.go` - Branch cleanup scheduler
4. `seeder.go` - Data seeding
5. `storage.go` - Branch metadata storage

### RPC Module (5 untested files)
1. `executor.go` - Procedure execution
2. `handler.go` - HTTP handler
3. `loader.go` - Procedure loading
4. `scheduler.go` - RPC scheduling
5. `storage.go` - RPC metadata

### Jobs Module (4 untested files)
1. `manager.go` - Job orchestration
2. `scheduler.go` - Cron scheduling
3. `storage.go` - Job storage
4. `worker.go` - Job execution

### PubSub Module (3 untested files)
1. `postgres.go` - PostgreSQL backend
2. `redis.go` - Redis backend
3. `interface.go` - Interface definitions

### Config Module (3 untested files)
1. `branching.go` - Branching config
2. `graphql.go` - GraphQL config
3. `mcp.go` - MCP config

---

## Implementation Strategy

### Phase 1: Quick Wins (Target: +5% coverage)
**Estimated effort: 1-2 weeks**

Focus on smaller, well-isolated files:
- [ ] `internal/api/validation_helpers.go`
- [ ] `internal/api/rest_query.go`
- [ ] `internal/api/rest_geojson.go`
- [ ] `internal/api/storage_utils.go`
- [ ] `internal/config/branching.go`
- [ ] `internal/config/graphql.go`
- [ ] `internal/config/mcp.go`
- [ ] `internal/pubsub/interface.go`

### Phase 2: Critical Handlers (Target: +10% coverage)
**Estimated effort: 2-3 weeks**

Focus on high-impact API handlers:
- [ ] `internal/api/branch_handler.go`
- [ ] `internal/api/policy_handler.go`
- [ ] `internal/api/servicekey_handler.go`
- [ ] `internal/api/webhook_handler.go`
- [ ] `internal/api/invitation_handler.go`
- [ ] `internal/api/user_management_handler.go`
- [ ] `internal/api/ddl_handler.go`

### Phase 3: AI Module (Target: +5% coverage)
**Estimated effort: 2-3 weeks**

Focus on AI core functionality:
- [ ] `internal/ai/handler.go`
- [ ] `internal/ai/storage.go`
- [ ] `internal/ai/knowledge_base_handler.go`
- [ ] `internal/ai/knowledge_base_storage.go`
- [ ] `internal/ai/conversation.go`
- [ ] `internal/ai/document_processor.go`

### Phase 4: Infrastructure (Target: +5% coverage)
**Estimated effort: 2 weeks**

- [ ] `internal/branching/*` (all 5 files)
- [ ] `internal/rpc/*` (all 5 files)
- [ ] `internal/jobs/*` (all 4 files)
- [ ] `internal/pubsub/postgres.go`
- [ ] `internal/pubsub/redis.go`

### Phase 5: Remaining Handlers (Target: +5% coverage)
**Estimated effort: 2-3 weeks**

- [ ] All settings handlers
- [ ] GraphQL handlers
- [ ] Storage advanced features
- [ ] OAuth/SAML providers

---

## Testing Guidelines

### Unit Test Requirements
1. **Happy path coverage** - Test normal operation
2. **Error conditions** - Test error handling
3. **Edge cases** - Empty inputs, nil values, boundaries
4. **Validation** - Input validation logic

### Test File Naming
- Place test files alongside source: `handler.go` → `handler_test.go`
- Use descriptive test names: `TestCreateBranch_ExceedsUserLimit_ReturnsError`

### Mocking Strategy
- Use `internal/testutil/` helpers for common mocks
- Mock database with `testutil.MockDB`
- Mock HTTP with `httptest.NewRecorder()`
- Use interface-based mocking for services

### Coverage Exclusions
The following are already excluded in `.testcoverage.yml`:
- Pure type definitions (`**/types.go`)
- Interface-only files
- Generated files (`openapi.go`)
- Infrastructure requiring external deps
- CLI commands (tested via integration)
- Test utilities

---

## Success Metrics

| Milestone | Target Coverage | Timeline |
|-----------|-----------------|----------|
| Current | 23% | - |
| Phase 1 Complete | 28% | Week 2 |
| Phase 2 Complete | 38% | Week 5 |
| Phase 3 Complete | 43% | Week 8 |
| Phase 4 Complete | 48% | Week 10 |
| Phase 5 Complete | 53% | Week 13 |

**Ultimate Target:** 50%+ overall coverage with 70%+ on critical modules (auth, api core, database)

---

## Verification Commands

```bash
# Run tests with coverage
make test-coverage

# Check coverage thresholds
make test-coverage-check

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# View coverage for specific package
go test -coverprofile=cover.out ./internal/api/...
go tool cover -func=cover.out | grep -v "100.0%"
```

---

## Notes

1. **Why 23% seems low:** The codebase has ~72k lines of test code, but much of it may be concentrated in well-tested modules (auth, middleware) while large critical modules (api, ai) have significant gaps.

2. **Excluded files:** Some files are intentionally excluded from coverage (types, interfaces, generated code). Ensure `.testcoverage.yml` exclusions are appropriate.

3. **E2E tests:** Many handlers may have some coverage through E2E tests in `test/e2e/`. Consider whether unit tests or integration tests are more appropriate for each file.

4. **Test quality:** Focus on meaningful tests that verify behavior, not just line coverage. A test that exercises code without asserting correctness provides limited value.
