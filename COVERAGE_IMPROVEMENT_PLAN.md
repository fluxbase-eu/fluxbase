# Fluxbase Code Coverage Improvement Plan

## Target: 90%+ Code Coverage

**Current Coverage:** 23% (according to Codecov)
**Target Coverage:** 90%+ overall, 95%+ on critical modules

---

## Current State Analysis

**Codebase Metrics:**
- Source files: 294 files (131,664 lines)
- Test files: 166 files (71,955 lines)
- Test-to-source ratio: 55% (substantial test code exists)

**Key Finding:** The low coverage percentage despite substantial test code indicates:
1. Existing tests don't cover enough code paths within tested files
2. Many large critical files lack tests entirely
3. Error handling and edge cases are undertested

**Coverage Gap:** To reach 90%, we need to cover ~88,000 additional lines of code.

---

## Coverage by Module

### Well-Tested Modules (have test files, but need deeper coverage)
| Module | Source Files | Test Files | Current Est. | Target |
|--------|--------------|------------|--------------|--------|
| auth | 40 | 40 | ~60% | 95% |
| middleware | 16 | 17 | ~55% | 95% |
| database | 5 | 5 | ~50% | 90% |
| logging | 5 | 5 | ~50% | 90% |
| secrets | 2 | 2 | ~45% | 90% |
| settings | 2 | 2 | ~45% | 90% |
| observability | 2 | 2 | ~40% | 90% |
| webhook | 2 | 3 | ~50% | 90% |
| crypto | 1 | 1 | ~60% | 95% |
| query | 1 | 1 | ~50% | 90% |

### Modules Needing New Tests
| Module | Source Files | Test Files | Gap | Target |
|--------|--------------|------------|-----|--------|
| **api** | 62 | 28 | 34 files | 90% |
| **ai** | 36 | 9 | 27 files | 85% |
| branching | 7 | 2 | 5 files | 90% |
| rpc | 8 | 3 | 5 files | 90% |
| jobs | 7 | 3 | 4 files | 90% |
| pubsub | 5 | 2 | 3 files | 90% |
| config | 4 | 1 | 3 files | 90% |
| storage | 12 | 6 | 6 files | 90% |
| realtime | 10 | 8 | 2 files | 90% |
| functions | 6 | 4 | 2 files | 90% |
| mcp | 6 | 3 | 3 files | 90% |
| extensions | 3 | 1 | 2 files | 85% |
| migrations | 3 | 1 | 2 files | 85% |

---

## Priority 1: API Module (34 untested files)

The `api` module is the largest (62 source files, 52,614 lines) and most critical for coverage improvement.

### Untested Files - Must Add Tests

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

### Existing Test Files - Must Deepen Coverage

These files have tests but likely need additional test cases for 90%+ coverage:
- `auth_handler_test.go` - Add edge cases, error paths
- `query_parser_test.go` - Add malformed input tests
- `rest_crud_test.go` - Add transaction edge cases
- `storage_*_test.go` - Add streaming/chunking edge cases
- `graphql_types_test.go` - Add schema edge cases

---

## Priority 2: AI Module (27 untested files)

The `ai` module (36 source files, 19,385 lines) handles AI/ML features.

### Untested Files - Must Add Tests

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

## Priority 3: Auth Module - Deepen Existing Coverage

The auth module has test files for all source files, but needs deeper coverage to reach 95%.

### Files Requiring Additional Test Cases
1. `service.go` - Add concurrent session tests, token refresh edge cases
2. `jwt.go` - Add expiry edge cases, malformed token tests
3. `oauth.go` - Add provider error handling, callback edge cases
4. `saml.go` - Add assertion validation edge cases
5. `totp.go` - Add timing attack tests, rate limit tests
6. `password.go` - Add hash migration tests, strength edge cases
7. `session.go` - Add concurrent access tests, cleanup tests
8. `scopes.go` - Add permission boundary tests

---

## Priority 4: Infrastructure Modules

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

### Storage Module (6 untested files)
1. `provider_s3.go` - S3 provider
2. `provider_local.go` - Local filesystem
3. `policy.go` - Access policies
4. `streaming.go` - Stream handling
5. `thumbnails.go` - Image processing
6. `metadata.go` - File metadata

### Realtime Module (2 untested files)
1. `subscription.go` - Subscription management
2. `broadcaster.go` - Message broadcasting

### Functions Module (2 untested files)
1. `handler.go` - Function HTTP handler
2. `scheduler.go` - Function scheduling

### MCP Module (3 untested files)
1. `server.go` - MCP server
2. `handler.go` - MCP HTTP handler
3. `tools/` - Tool implementations

---

## Implementation Strategy

### Phase 1: Foundation & Quick Wins (Target: 35%)
**Effort: 2 weeks**

Add tests for smaller, well-isolated files and deepen existing test coverage:

**New Tests:**
- [ ] `internal/api/validation_helpers.go`
- [ ] `internal/api/rest_query.go`
- [ ] `internal/api/rest_geojson.go`
- [ ] `internal/api/storage_utils.go`
- [ ] `internal/config/branching.go`
- [ ] `internal/config/graphql.go`
- [ ] `internal/config/mcp.go`
- [ ] `internal/pubsub/interface.go`

**Deepen Existing:**
- [ ] Add error path tests to `internal/auth/*_test.go`
- [ ] Add edge cases to `internal/middleware/*_test.go`
- [ ] Add boundary tests to `internal/database/*_test.go`

### Phase 2: Critical API Handlers (Target: 50%)
**Effort: 3 weeks**

**New Tests for High-Impact Handlers:**
- [ ] `internal/api/branch_handler.go`
- [ ] `internal/api/policy_handler.go`
- [ ] `internal/api/servicekey_handler.go`
- [ ] `internal/api/webhook_handler.go`
- [ ] `internal/api/invitation_handler.go`
- [ ] `internal/api/user_management_handler.go`
- [ ] `internal/api/ddl_handler.go`
- [ ] `internal/api/oauth_provider_handler.go`
- [ ] `internal/api/saml_provider_handler.go`

**Deepen Existing:**
- [ ] `internal/api/auth_handler_test.go` - Add 20+ new test cases
- [ ] `internal/api/query_parser_test.go` - Add malformed input tests
- [ ] `internal/api/rest_crud_test.go` - Add transaction tests

### Phase 3: AI Module (Target: 60%)
**Effort: 3 weeks**

**New Tests:**
- [ ] `internal/ai/handler.go` (2,002 lines - critical)
- [ ] `internal/ai/storage.go` (1,501 lines)
- [ ] `internal/ai/knowledge_base_handler.go`
- [ ] `internal/ai/knowledge_base_storage.go`
- [ ] `internal/ai/conversation.go`
- [ ] `internal/ai/document_processor.go`
- [ ] `internal/ai/schema_builder.go`
- [ ] `internal/ai/rag_service.go`
- [ ] `internal/ai/audit.go`

**Provider Mocks:**
- [ ] Create mock providers for OpenAI, Azure, Ollama
- [ ] Add provider error handling tests
- [ ] Add rate limiting tests

### Phase 4: GraphQL & Advanced API (Target: 70%)
**Effort: 2 weeks**

**New Tests:**
- [ ] `internal/api/graphql_handler.go`
- [ ] `internal/api/graphql_resolvers.go`
- [ ] `internal/api/graphql_schema.go`
- [ ] `internal/api/vector_handler.go`
- [ ] `internal/api/vector_manager.go`
- [ ] All settings handlers (8 files)

**Deepen Existing:**
- [ ] `internal/api/graphql_types_test.go` - Add schema edge cases

### Phase 5: Infrastructure (Target: 80%)
**Effort: 3 weeks**

**Branching:**
- [ ] `internal/branching/manager.go`
- [ ] `internal/branching/router.go`
- [ ] `internal/branching/scheduler.go`
- [ ] `internal/branching/seeder.go`
- [ ] `internal/branching/storage.go`

**RPC:**
- [ ] `internal/rpc/executor.go`
- [ ] `internal/rpc/handler.go`
- [ ] `internal/rpc/loader.go`
- [ ] `internal/rpc/scheduler.go`
- [ ] `internal/rpc/storage.go`

**Jobs:**
- [ ] `internal/jobs/manager.go`
- [ ] `internal/jobs/scheduler.go`
- [ ] `internal/jobs/storage.go`
- [ ] `internal/jobs/worker.go`

**PubSub:**
- [ ] `internal/pubsub/postgres.go`
- [ ] `internal/pubsub/redis.go`

### Phase 6: Storage & Streaming (Target: 85%)
**Effort: 2 weeks**

**New Tests:**
- [ ] `internal/api/storage_chunked.go`
- [ ] `internal/api/storage_multipart.go`
- [ ] `internal/api/storage_sharing.go`
- [ ] `internal/api/storage_streaming.go`
- [ ] `internal/storage/provider_s3.go`
- [ ] `internal/storage/provider_local.go`
- [ ] `internal/storage/policy.go`
- [ ] `internal/storage/streaming.go`

**Deepen Existing:**
- [ ] Add large file tests
- [ ] Add concurrent upload tests
- [ ] Add resumable upload tests

### Phase 7: Realtime, Functions & MCP (Target: 88%)
**Effort: 2 weeks**

**Realtime:**
- [ ] `internal/realtime/subscription.go`
- [ ] `internal/realtime/broadcaster.go`
- [ ] Deepen `internal/realtime/*_test.go`

**Functions:**
- [ ] `internal/functions/handler.go`
- [ ] `internal/functions/scheduler.go`
- [ ] Deepen loader/bundler tests

**MCP:**
- [ ] `internal/mcp/server.go`
- [ ] `internal/mcp/handler.go`
- [ ] `internal/mcp/tools/*.go`

### Phase 8: Auth Deep Dive & Polish (Target: 90%+)
**Effort: 2 weeks**

**Deepen Auth Coverage to 95%:**
- [ ] Add concurrent session tests
- [ ] Add token refresh race conditions
- [ ] Add OAuth callback edge cases
- [ ] Add SAML assertion edge cases
- [ ] Add TOTP timing tests
- [ ] Add password migration tests
- [ ] Add impersonation security tests
- [ ] Add MFA recovery tests

**Remaining API Files:**
- [ ] `internal/api/auth_middleware.go`
- [ ] `internal/api/auth_saml.go`
- [ ] `internal/api/clientkey_handler.go`
- [ ] `internal/api/custom_mcp_handler.go`
- [ ] `internal/api/github_webhook_handler.go`
- [ ] `internal/api/logging_handlers.go`
- [ ] `internal/api/realtime_admin_handler.go`
- [ ] `internal/api/schema_handler.go`
- [ ] `internal/api/admin_auth_handler.go`
- [ ] `internal/api/admin_session_handler.go`

**AI Provider Coverage:**
- [ ] `internal/ai/provider_openai.go`
- [ ] `internal/ai/provider_ollama.go`
- [ ] `internal/ai/provider_azure.go`
- [ ] `internal/ai/embedding_*.go`

---

## Testing Requirements for 90%+ Coverage

### Per-File Requirements
Each file must have tests covering:
1. **All public functions** - Every exported function must have at least one test
2. **Happy paths** - Normal operation for all major flows
3. **Error conditions** - All error return paths
4. **Edge cases** - Empty inputs, nil values, boundary conditions
5. **Validation logic** - All input validation branches
6. **Concurrency** - Race conditions where applicable

### Branch Coverage Requirements
- **Critical modules (auth, api core):** 95% branch coverage
- **Standard modules:** 90% branch coverage
- **Infrastructure modules:** 85% branch coverage

### Test Quality Standards
```go
// Good: Specific, descriptive test name
func TestCreateBranch_ExceedsUserLimit_ReturnsLimitError(t *testing.T)

// Good: Table-driven tests for comprehensive coverage
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"missing @", "userexample.com", true},
        {"empty string", "", true},
        {"unicode domain", "user@例え.jp", false},
        // ... 20+ cases
    }
}

// Good: Test error paths explicitly
func TestGetUser_NotFound_ReturnsErrNotFound(t *testing.T)
func TestGetUser_DatabaseError_ReturnsErrInternal(t *testing.T)
```

### Mocking Strategy
- Use `internal/testutil/` for shared mocks
- Create interface mocks for all external dependencies
- Use `httptest` for HTTP handler tests
- Use `pgxmock` for database tests
- Create mock providers for AI services

---

## Test Infrastructure Additions

### Required Mock Implementations
```
internal/testutil/
├── mock_db.go           # Database mock (exists, enhance)
├── mock_storage.go      # Storage provider mock
├── mock_ai_provider.go  # AI provider mock (new)
├── mock_email.go        # Email service mock
├── mock_pubsub.go       # PubSub mock (new)
├── mock_redis.go        # Redis client mock (new)
├── http_helpers.go      # HTTP test helpers (exists, enhance)
└── fixtures/            # Test data fixtures (new)
    ├── users.go
    ├── branches.go
    ├── storage.go
    └── ai_responses.go
```

### Integration Test Additions
```
test/e2e/
├── api_branch_test.go      # Branch API e2e (new)
├── api_graphql_test.go     # GraphQL e2e (new)
├── ai_chat_test.go         # AI chat e2e (new)
├── storage_upload_test.go  # Storage e2e (enhance)
└── realtime_test.go        # Realtime e2e (new)
```

---

## Success Metrics

| Milestone | Target Coverage | Timeline |
|-----------|-----------------|----------|
| Current | 23% | - |
| Phase 1 Complete | 35% | Week 2 |
| Phase 2 Complete | 50% | Week 5 |
| Phase 3 Complete | 60% | Week 8 |
| Phase 4 Complete | 70% | Week 10 |
| Phase 5 Complete | 80% | Week 13 |
| Phase 6 Complete | 85% | Week 15 |
| Phase 7 Complete | 88% | Week 17 |
| Phase 8 Complete | **90%+** | Week 19 |

### Per-Module Targets at Completion
| Module | Target | Priority |
|--------|--------|----------|
| auth | 95% | Critical |
| api | 90% | Critical |
| middleware | 95% | Critical |
| database | 90% | High |
| ai | 85% | High |
| storage | 90% | High |
| jobs | 90% | Medium |
| branching | 90% | Medium |
| rpc | 90% | Medium |
| realtime | 90% | Medium |
| functions | 90% | Medium |
| pubsub | 90% | Medium |
| mcp | 90% | Medium |
| logging | 90% | Low |
| config | 90% | Low |

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

# Check coverage percentage
go tool cover -func=cover.out | grep total

# Find uncovered lines in a file
go tool cover -func=cover.out | grep -E "0.0%|[0-9]\.[0-9]%"

# Run with race detector (important for concurrency tests)
go test -race -coverprofile=cover.out ./...
```

---

## Coverage Exclusions

The following should be excluded in `.testcoverage.yml` (verify current config):
- Pure type definitions (`**/types.go`)
- Interface-only files (`**/interfaces.go`)
- Generated files (`openapi.go`, `*.gen.go`)
- Infrastructure requiring external deps that can't be mocked
- CLI commands (tested via integration tests)
- Test utilities (`internal/testutil/`)
- Main entry points (`cmd/*/main.go`)

---

## Risk Factors & Mitigations

### Risk: External Dependencies
Some code requires external services (PostgreSQL, Redis, S3, AI providers).
**Mitigation:** Create comprehensive mocks, use Docker for integration tests.

### Risk: Flaky Tests
High coverage can lead to flaky tests if not carefully designed.
**Mitigation:** Avoid time-dependent tests, use deterministic test data, proper test isolation.

### Risk: Test Maintenance Burden
90%+ coverage requires significant test maintenance.
**Mitigation:** Use table-driven tests, shared fixtures, good test abstractions.

### Risk: Coverage Without Quality
High line coverage doesn't guarantee bug-free code.
**Mitigation:** Focus on meaningful assertions, test business logic not just execution paths.

---

## Notes

1. **Realistic Timeline:** 19 weeks is aggressive. Adjust based on team size and other priorities.

2. **Incremental PRs:** Each phase should be broken into multiple PRs (~5-10 files per PR).

3. **CI Integration:** Update CI to fail builds if coverage drops below current threshold.

4. **Coverage Ratchet:** After each phase, increase minimum coverage threshold in CI.

5. **Test Review:** All test PRs should be reviewed for test quality, not just coverage increase.
