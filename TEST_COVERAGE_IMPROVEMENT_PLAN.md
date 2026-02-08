# Test Coverage Improvement Plan

## Executive Summary

**Current Coverage:** 30.8%
**Target Coverage:** 90%
**Gap:** 59.2 percentage points

This plan organizes the work into 8 developer tracks, each focusing on specific modules. Work is prioritized by business criticality and coverage gaps.

---

## Current State Analysis

### Module Coverage Breakdown

| Module | Coverage | Files | Low-Cov Files | Priority |
|--------|----------|-------|---------------|----------|
| internal/settings | 10.6% | 47 | 47 | P0 |
| internal/migrations | 14.3% | 28 | 28 | P1 |
| internal/branching | 19.6% | 88 | 81 | P0 |
| internal/database | 24.0% | 75 | 75 | P1 |
| internal/jobs | 24.5% | 126 | 120 | P1 |
| internal/secrets | 23.1% | 26 | 26 | P2 |
| internal/extensions | 22.2% | 18 | 18 | P2 |
| internal/webhook | 34.8% | 47 | 42 | P1 |
| internal/functions | 42.2% | 107 | 89 | P1 |
| internal/ai | 46.1% | 401 | 357 | P2 |
| internal/api | 48.1% | 684 | 497 | P1 |
| internal/auth | 51.1% | 531 | 430 | P1 |
| internal/rpc | 54.2% | 83 | 70 | P2 |
| internal/realtime | 55.7% | 143 | 124 | P2 |
| internal/email | 59.0% | 71 | 62 | P2 |
| internal/pubsub | 60.9% | 26 | 23 | P2 |
| internal/ratelimit | 68.2% | 37 | 30 | P2 |
| internal/middleware | 72.8% | 147 | 103 | P2 |
| internal/storage | 73.9% | 141 | 72 | P2 |
| internal/runtime | 74.6% | 32 | 31 | P2 |
| internal/mcp | 79.4% | 474 | 453 | P3 |
| internal/adminui | 75.0% | 4 | 4 | P3 |
| internal/config | 83.5% | 29 | 19 | P3 |
| internal/logging | 83.6% | 47 | 37 | P3 |
| internal/observability | 83.7% | 56 | 44 | P3 |
| internal/crypto | 92.0% | 6 | 3 | P3 |

*Low-Cov Files = Files with <30% coverage*

### Test Infrastructure Assessment

**Strengths:**
- Well-established test utilities in `test/` package
- Mock implementations for most external dependencies
- Shared database context with singleton pattern
- E2E test framework with test tables (products, tasks)
- Two database users for RLS testing (fluxbase_app, fluxbase_rls_test)

**Weaknesses:**
- No tests in internal/testutil or internal/testcontext (excluded from coverage)
- Limited integration tests for complex workflows
- Some modules have tests but poor coverage

---

## Test Patterns Reference

### 1. Table-Driven Tests (Most Common)
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {"scenario 1", input1, expected1, false},
        {"scenario 2", input2, expected2, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Integration Tests with Database
```go
func TestWithDatabase(t *testing.T) {
    ctx := testutil.NewIntegrationTestContext(t)
    defer ctx.CleanupTestData()

    // Create test data
    userID := ctx.CreateTestUser("test@example.com")

    // Test your code
    result, err := YourFunction(ctx.DB, userID)

    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "expected", result)
}
```

### 3. Using Mocks
```go
func TestWithMocks(t *testing.T) {
    mockStorage := testutil.NewMockStorageProvider()
    mockPubSub := testutil.NewMockPubSub()

    // Configure mock behavior
    mockStorage.OnUpload = func(...) error {
        return nil // or specific test behavior
    }

    // Inject mocks into your service
    service := NewService(mockStorage, mockPubSub)

    // Test
    err := service.DoSomething(...)
    assert.NoError(t, err)
}
```

### 4. HTTP Handler Tests
```go
func TestHandler(t *testing.T) {
    app := fiber.New()
    app.Post("/endpoint", YourHandler)

    req := httptest.NewRequest("POST", "/endpoint", strings.NewReader(`{"key":"value"}`))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    require.NoError(t, err)

    assert.Equal(t, 200, resp.StatusCode)
}
```

### 5. Test Isolation Best Practices

**Always:**
- Use `defer ctx.CleanupTestData()` for integration tests
- Use unique test emails (e.g., `test-{uuid}@example.com`)
- Truncate tables between tests when needed
- Reset global state with `test.ResetGlobalTestState()`

**Never:**
- Call `Close()` on shared test contexts (singleton manages lifecycle)
- Hardcode test data that could conflict
- Assume tests run in any order

---

## Developer Team Structure

### Developer 1: Core Auth & Settings (P0 - Critical Path)
**Modules:** `internal/settings`, `internal/auth`

**Goal:** Settings (10.6% → 80%), Auth (51.1% → 85%)

**Priority Files:**
- internal/settings/service.go
- internal/auth/service.go (core auth methods)
- internal/auth/session.go (session management)
- internal/auth/user.go (user CRUD)
- internal/auth/mfa.go (MFA functionality)

**Test Strategy:**
1. Settings service with cache mocking
2. User registration, login, password reset
3. Session lifecycle (create, validate, expire)
4. MFA setup and verification
5. Integration tests for full auth flows

**Estimated Impact:** +800 new tests, +6% overall coverage

---

### Developer 2: Database & Branching (P0 - Data Layer)
**Modules:** `internal/database`, `internal/branching`

**Goal:** Database (24% → 75%), Branching (19.6% → 80%)

**Priority Files:**
- internal/database/schema_inspector.go
- internal/database/executor.go
- internal/branching/manager.go (CREATE/DROP DATABASE)
- internal/branching/router.go (connection routing)
- internal/branching/storage.go (branch metadata)

**Test Strategy:**
1. Schema introspection with mock schemas
2. Query execution (SELECT, INSERT, UPDATE, DELETE)
3. Branch creation and deletion
4. Connection routing per branch
5. Branch metadata CRUD

**Estimated Impact:** +600 new tests, +5% overall coverage

---

### Developer 3: API Handlers - Core REST (P1 - HTTP Layer)
**Modules:** `internal/api` (REST handlers)

**Goal:** 48.1% → 85%

**Priority Files:**
- internal/api/rest_crud.go (CRUD operations)
- internal/api/query_parser.go (query parsing)
- internal/api/query_builder.go (SQL building)
- internal/api/rest_batch.go (batch operations)
- internal/api/schema_handler.go (schema endpoints)

**Test Strategy:**
1. CRUD operations (GET, POST, PATCH, DELETE)
2. Query parsing (filters, ordering, pagination)
3. Query builder (SQL generation)
4. Batch operations
5. Schema introspection endpoints

**Estimated Impact:** +500 new tests, +8% overall coverage

---

### Developer 4: Jobs, Webhooks & Functions (P1 - Async)
**Modules:** `internal/jobs`, `internal/webhook`, `internal/functions`

**Goal:** Jobs (24.5% → 75%), Webhooks (34.8% → 80%), Functions (42.2% → 80%)

**Priority Files:**
- internal/jobs/manager.go (job orchestration)
- internal/jobs/worker.go (job execution)
- internal/webhook/service.go (webhook delivery)
- internal/functions/handler.go (function HTTP handler)
- internal/functions/loader.go (function loading)

**Test Strategy:**
1. Job scheduling and execution
2. Job progress tracking
3. Webhook delivery with retry logic
4. Function execution (Deno runtime mocking)
5. Function loading and bundling

**Estimated Impact:** +700 new tests, +7% overall coverage

---

### Developer 5: API Handlers - Auth & Storage (P1 - Feature APIs)
**Modules:** `internal/api` (auth, storage handlers)

**Goal:** Specific handler files to 80%+

**Priority Files:**
- internal/api/auth_handler.go (auth endpoints)
- internal/api/storage_handler.go (file upload/download)
- internal/api/storage_buckets.go (bucket management)
- internal/api/oauth_handler.go (OAuth flows)
- internal/api/ddl_handler.go (DDL operations)

**Test Strategy:**
1. Authentication endpoints (signup, login, logout)
2. File upload/download/streaming
3. Bucket CRUD operations
4. OAuth authorization flow
5. DDL operations (CREATE TABLE, ALTER, etc.)

**Estimated Impact:** +400 new tests, +5% overall coverage

---

### Developer 6: Realtime, RPC & Email (P2 - Communication)
**Modules:** `internal/realtime`, `internal/rpc`, `internal/email`

**Goal:** Realtime (55.7% → 80%), RPC (54.2% → 80%), Email (59% → 85%)

**Priority Files:**
- internal/realtime/hub.go (WebSocket hub)
- internal/realtime/manager.go (subscription manager)
- internal/rpc/service.go (RPC execution)
- internal/email/service.go (email sending)
- internal/email/smtp.go (SMTP provider)

**Test Strategy:**
1. WebSocket connection lifecycle
2. Subscription management
3. RPC procedure execution
4. Email sending (with MailHog integration)
5. SMTP provider implementation

**Estimated Impact:** +500 new tests, +5% overall coverage

---

### Developer 7: AI, PubSub & Extensions (P2 - Advanced Features)
**Modules:** `internal/ai`, `internal/pubsub`, `internal/extensions`

**Goal:** AI (46.1% → 75%), PubSub (60.9% → 85%), Extensions (22.2% → 75%)

**Priority Files:**
- internal/ai/chatbot.go (AI chat interface)
- internal/ai/embedding.go (vector embeddings)
- internal/pubsub/local.go (local pubsub)
- internal/pubsub/postgres.go (PostgreSQL pubsub)
- internal/extensions/manager.go (extension management)

**Test Strategy:**
1. Chat interface with mock AI providers
2. Embedding generation (OpenAI, Ollama, Azure)
3. PubSub publish/subscribe
4. PostgreSQL LISTEN/NOTIFY
5. Extension installation/removal

**Estimated Impact:** +400 new tests, +4% overall coverage

---

### Developer 8: Middleware, Secrets & Rate Limiting (P2 - Infrastructure)
**Modules:** `internal/middleware`, `internal/secrets`, `internal/ratelimit`

**Goal:** Middleware (72.8% → 90%), Secrets (23.1% → 80%), Ratelimit (68.2% → 90%)

**Priority Files:**
- internal/middleware/auth.go (authentication middleware)
- internal/secrets/handler.go (secret encryption)
- internal/ratelimit/store.go (rate limiting logic)
- internal/middleware/cors.go (CORS handling)
- internal/middleware/ratelimit.go (rate limiting middleware)

**Test Strategy:**
1. Authentication middleware (JWT validation)
2. Secret encryption/decryption
3. Rate limiting (in-memory, PostgreSQL, Redis)
4. CORS handling
5. Request logging

**Estimated Impact:** +300 new tests, +3% overall coverage

---

## Execution Timeline

### Phase 1: Foundation (Week 1-2)
- All developers: Set up test environments, run existing tests, understand patterns
- Developer 1: Settings service tests (quick wins)
- Developer 2: Database schema inspector tests
- Developer 3: Query parser tests (isolated, high value)

**Target:** 35% → 40% overall coverage

### Phase 2: Core Features (Week 3-4)
- Developer 1: Auth service integration tests
- Developer 2: Branch manager tests
- Developer 3: CRUD handler tests
- Developer 4: Job manager tests
- Developer 5: Auth handler tests
- Developer 6: Realtime hub tests
- Developer 7: AI embedding tests
- Developer 8: Middleware auth tests

**Target:** 40% → 55% overall coverage

### Phase 3: Advanced Features (Week 5-6)
- All developers: Focus on complex integration tests
- Cross-module testing (e.g., auth + API + database)
- Edge case and error path testing

**Target:** 55% → 75% overall coverage

### Phase 4: Polish & Edge Cases (Week 7-8)
- Focus on remaining low-coverage files
- Add tests for edge cases and error paths
- Performance and race condition testing
- Documentation of test patterns

**Target:** 75% → 90% overall coverage

---

## Test Coverage Enforcement

### Update .testcoverage.yml
Incrementally increase thresholds:

```yaml
# Phase 1 (Week 2)
threshold:
  file: 10
  package: 15
  total: 35

# Phase 2 (Week 4)
threshold:
  file: 20
  package: 30
  total: 50

# Phase 3 (Week 6)
threshold:
  file: 40
  package: 50
  total: 70

# Target (Week 8)
threshold:
  file: 60
  package: 70
  total: 85
```

### CI/CD Integration
- Block PRs that decrease coverage
- Require coverage reports on all PRs
- Add coverage badge to README

---

## Testing Guidelines

### When to Write Tests
1. **New features:** Tests must accompany the PR
2. **Bug fixes:** Add regression test
3. **Refactoring:** Ensure existing tests pass
4. **Code review:** Request tests for uncovered paths

### Test File Organization
```
internal/
  module/
    module.go          # Production code
    module_test.go     # Table-driven tests
    module_integration_test.go  # Integration tests (build tag: integration)
```

### Naming Conventions
- Test function: `TestFunctionName_Scenario_ExpectedBehavior`
- Test cases: Descriptive names like `"returns error when user not found"`
- Test data: `validInput`, `emptyInput`, `malformedInput`

### Coverage Goals by File Type
- Business logic: 80%+
- HTTP handlers: 70%+
- Utility functions: 90%+
- Type definitions: 0% (excluded)
- Error-only files: 0% (excluded)

---

## Success Metrics

### Week 2 Checkpoint
- [ ] 35% overall coverage
- [ ] Settings at 40%+
- [ ] Database schema inspector at 50%+
- [ ] Query parser at 60%+

### Week 4 Checkpoint
- [ ] 50% overall coverage
- [ ] Auth service at 70%+
- [ ] Branch manager at 60%+
- [ ] REST CRUD handlers at 70%+
- [ ] Job manager at 50%+

### Week 6 Checkpoint
- [ ] 70% overall coverage
- [ ] All core modules at 60%+
- [ ] Integration tests for critical paths
- [ ] Zero regressions in existing tests

### Week 8 (Target)
- [ ] 90% overall coverage
- [ ] All modules at 75%+ (except excluded files)
- [ ] CI coverage enforcement active
- [ ] Test patterns documented

---

## References

- Test utilities: `internal/testutil/`
- E2E framework: `test/e2e/`
- Test context: `test/e2e_helpers.go`
- Coverage config: `.testcoverage.yml`
- CLAUDE.md: Project documentation and architecture
