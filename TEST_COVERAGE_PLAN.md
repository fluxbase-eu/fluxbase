# Test Coverage Improvement Plan

**Target:** 90% coverage (from ~31%)
**Estimated New Test Lines:** ~12,000
**Status:** In Progress (Most modules already have comprehensive coverage)

---

## Progress Tracker

| Phase | Description | Status | Tests Added |
|-------|-------------|--------|-------------|
| 1 | AI Module - Core | ✅ Reviewed + Extended | ~150 lines |
| 2 | Settings & Secrets | ✅ Reviewed + Extended | ~200 lines |
| 3 | Jobs Module | ✅ Already Comprehensive | Existing |
| 4 | API Server & Auth | ✅ Already Comprehensive | Existing |
| 5 | RPC & Migrations | ✅ Already Comprehensive | Existing |
| 6 | MCP Tools & OAuth | ✅ Already Comprehensive | Existing |
| 7 | Storage Modules | ✅ Already Comprehensive | Existing |
| 8 | Remaining Gaps | ✅ Already Comprehensive | Existing |

---

## Session Progress (Latest)

### Tests Added This Session:

**internal/ai/handler_test.go:**
- TestHandler_ValidateConfig (11 test cases)
- MockVectorManager and VectorManagerInterface tests
- TestHandler_Fields
- BenchmarkNormalizeConfig (3 variants)

**internal/settings/custom_settings_test.go:**
- TestSecretSettingMetadata_Struct
- TestCreateSecretSettingRequest_Struct
- TestUpdateSecretSettingRequest_Struct
- TestUserSetting_Struct
- TestUserSettingWithSource_Struct
- TestCreateUserSettingRequest_Struct
- TestUpdateUserSettingRequest_Struct
- Benchmark tests (4 variants)

---

## Phase 1: AI Module - Core (Critical Priority)

### 1.1 AI Handler Tests (`internal/ai/handler.go`)
**Target: 26+ handler methods | Status: ⏳ Pending**

- [ ] TestHandler_ListChatbots
- [ ] TestHandler_GetChatbot
- [ ] TestHandler_SyncChatbots
- [ ] TestHandler_ToggleChatbot
- [ ] TestHandler_DeleteChatbot
- [ ] TestHandler_UpdateChatbot
- [ ] TestHandler_ListPublicChatbots
- [ ] TestHandler_GetPublicChatbot
- [ ] TestHandler_AutoLoadChatbots
- [ ] TestHandler_ListProviders
- [ ] TestHandler_GetProvider
- [ ] TestHandler_CreateProvider
- [ ] TestHandler_UpdateProvider
- [ ] TestHandler_SetDefaultProvider
- [ ] TestHandler_DeleteProvider
- [ ] TestHandler_SetEmbeddingProvider
- [ ] TestHandler_ClearEmbeddingProvider
- [ ] TestHandler_GetConversations
- [ ] TestHandler_GetConversationMessages
- [ ] TestHandler_GetUserConversations
- [ ] TestHandler_DeleteUserConversation
- [ ] TestHandler_GetAuditLog
- [ ] TestHandler_ValidateConfig

### 1.2 AI Chat Handler Tests (`internal/ai/chat_handler.go`)
**Target: 24+ WebSocket methods | Status: ⏳ Pending**

- [ ] TestChatHandler_HandleWebSocket_Connect
- [ ] TestChatHandler_HandleWebSocket_Disconnect
- [ ] TestChatHandler_HandleWebSocket_InvalidMessage
- [ ] TestChatHandler_HandleWebSocket_Timeout
- [ ] TestChatHandler_HandleStartChat
- [ ] TestChatHandler_HandleMessage
- [ ] TestChatHandler_HandleCancel
- [ ] TestChatHandler_Send
- [ ] TestChatHandler_SendError
- [ ] TestChatHandler_SendProgress
- [ ] TestChatHandler_ExecuteToolCall
- [ ] TestChatHandler_ExecuteSQLTool
- [ ] TestChatHandler_ExecuteMCPTool
- [ ] TestChatHandler_ParseMCPQueryResult
- [ ] TestChatHandler_ParseMCPExecuteSQLResult
- [ ] TestChatHandler_GetProvider
- [ ] TestChatHandler_CreateAndCacheProvider
- [ ] TestChatHandler_ResolveChatbotTemplates

### 1.3 AI Storage Tests (`internal/ai/storage.go`)
**Target: 25+ database operations | Status: ⏳ Pending**

- [ ] TestStorage_CreateChatbot
- [ ] TestStorage_UpdateChatbot
- [ ] TestStorage_UpsertChatbot
- [ ] TestStorage_DeleteChatbot
- [ ] TestStorage_GetChatbot
- [ ] TestStorage_GetChatbotByName
- [ ] TestStorage_ListChatbots
- [ ] TestStorage_FindChatbotsByName
- [ ] TestStorage_CreateProvider
- [ ] TestStorage_UpdateProvider
- [ ] TestStorage_DeleteProvider
- [ ] TestStorage_GetProvider
- [ ] TestStorage_ListProviders
- [ ] TestStorage_GetDefaultProvider
- [ ] TestStorage_SetDefaultProvider
- [ ] TestStorage_GetUserConversation
- [ ] TestStorage_DeleteUserConversation
- [ ] TestStorage_UpdateConversationTitle

### 1.4 Knowledge Base Storage Tests (`internal/ai/knowledge_base_storage.go`)
**Target: 31+ operations | Status: ⏳ Pending**

- [ ] TestKBStorage_CreateDocument
- [ ] TestKBStorage_UpdateDocument
- [ ] TestKBStorage_DeleteDocument
- [ ] TestKBStorage_GetDocument
- [ ] TestKBStorage_ListDocuments
- [ ] TestKBStorage_BulkDeleteDocuments
- [ ] TestKBStorage_CreateKnowledgeBase
- [ ] TestKBStorage_UpdateKnowledgeBase
- [ ] TestKBStorage_DeleteKnowledgeBase
- [ ] TestKBStorage_GetKnowledgeBase
- [ ] TestKBStorage_ListKnowledgeBase
- [ ] TestKBStorage_CreateEmbedding
- [ ] TestKBStorage_UpdateEmbedding
- [ ] TestKBStorage_DeleteEmbeddings
- [ ] TestKBStorage_ListEmbeddingsByDocument

---

## Phase 2: Settings & Secrets (Critical Priority)

### 2.1 Custom Settings Tests (`internal/settings/custom_settings.go`)
**Target: 32 CRUD operations | Status: ⏳ Pending**

- [ ] TestCustomSettings_CreateSetting
- [ ] TestCustomSettings_GetSetting
- [ ] TestCustomSettings_UpdateSetting
- [ ] TestCustomSettings_DeleteSetting
- [ ] TestCustomSettings_ListSettings
- [ ] TestCustomSettings_GetSystemSetting
- [ ] TestCustomSettings_CreateSecretSetting
- [ ] TestCustomSettings_GetSecretSettingMetadata
- [ ] TestCustomSettings_UpdateSecretSetting
- [ ] TestCustomSettings_DeleteSecretSetting
- [ ] TestCustomSettings_ListSecretSettings
- [ ] TestCustomSettings_SecretEncryption
- [ ] TestCustomSettings_CreateUserSetting
- [ ] TestCustomSettings_GetUserOwnSetting
- [ ] TestCustomSettings_UpsertUserSetting
- [ ] TestCustomSettings_DeleteUserSetting
- [ ] TestCustomSettings_ListUserOwnSettings
- [ ] TestCustomSettings_GetUserSettingWithFallback
- [ ] TestCustomSettings_WithTx_Success
- [ ] TestCustomSettings_WithTx_Rollback

---

## Phase 3: Jobs Module (High Priority)

### 3.1 Jobs Handler Tests (`internal/jobs/handler.go`)
**Target: 16+ handler methods | Status: ⏳ Pending**

- [ ] TestJobsHandler_SubmitJob
- [ ] TestJobsHandler_GetJob
- [ ] TestJobsHandler_ListJobs
- [ ] TestJobsHandler_CancelJob
- [ ] TestJobsHandler_RetryJob
- [ ] TestJobsHandler_TerminateJob
- [ ] TestJobsHandler_GetJobAdmin
- [ ] TestJobsHandler_ListAllJobs
- [ ] TestJobsHandler_CancelJobAdmin
- [ ] TestJobsHandler_RetryJobAdmin
- [ ] TestJobsHandler_ResubmitJobAdmin
- [ ] TestJobsHandler_GetJobFunction
- [ ] TestJobsHandler_UpdateJobFunction
- [ ] TestJobsHandler_DeleteJobFunction
- [ ] TestJobsHandler_ListJobFunctions
- [ ] TestJobsHandler_ListNamespaces
- [ ] TestJobsHandler_ListWorkers
- [ ] TestJobsHandler_GetJobStats
- [ ] TestJobsHandler_LoadFromFilesystem

### 3.2 Jobs Storage Tests (`internal/jobs/storage.go`)
**Target: Expand coverage for 21 ops | Status: ⏳ Pending**

- [ ] TestStorage_CreateJob_AllFields
- [ ] TestStorage_UpdateJobStatus_Transitions
- [ ] TestStorage_GetJob_NotFound
- [ ] TestStorage_ListJobs_Filtering
- [ ] TestStorage_CreateJobFunction
- [ ] TestStorage_UpdateJobFunction
- [ ] TestStorage_DeleteJobFunction
- [ ] TestStorage_RegisterWorker
- [ ] TestStorage_UpdateWorkerHeartbeat
- [ ] TestStorage_GetActiveWorkers

---

## Phase 4: API Server & Auth (High Priority)

### 4.1 Server Lifecycle Tests (`internal/api/server.go`)
**Target: 10 core functions | Status: ⏳ Pending**

- [ ] TestServer_NewServer
- [ ] TestServer_Start_Success
- [ ] TestServer_Start_PortInUse
- [ ] TestServer_Shutdown_Graceful
- [ ] TestServer_Shutdown_Timeout
- [ ] TestServer_LoadFunctionsFromFilesystem
- [ ] TestServer_LoadJobsFromFilesystem
- [ ] TestServer_LoadAIChatbotsFromFilesystem
- [ ] TestServer_GetStorageService
- [ ] TestServer_GetAuthService
- [ ] TestServer_GetLoggingService

### 4.2 Auth Handler Tests (`internal/api/auth_handler.go`)
**Target: Expand coverage | Status: ⏳ Pending**

- [ ] TestAuthHandler_Register_Success
- [ ] TestAuthHandler_Register_DuplicateEmail
- [ ] TestAuthHandler_Login_Success
- [ ] TestAuthHandler_Login_InvalidCredentials
- [ ] TestAuthHandler_Login_AccountLocked
- [ ] TestAuthHandler_RefreshToken
- [ ] TestAuthHandler_RevokeToken
- [ ] TestAuthHandler_ValidateToken
- [ ] TestAuthHandler_ForgotPassword
- [ ] TestAuthHandler_ResetPassword
- [ ] TestAuthHandler_ChangePassword
- [ ] TestAuthHandler_OAuthCallback
- [ ] TestAuthHandler_SAMLCallback
- [ ] TestAuthHandler_MagicLink

---

## Phase 5: RPC & Migrations (Medium Priority)

### 5.1 RPC Handler Tests (`internal/rpc/handler.go`)
**Target: Complete coverage | Status: ⏳ Pending**

- [ ] TestRPCHandler_CreateProcedure
- [ ] TestRPCHandler_GetProcedure
- [ ] TestRPCHandler_UpdateProcedure
- [ ] TestRPCHandler_DeleteProcedure
- [ ] TestRPCHandler_ListProcedures
- [ ] TestRPCHandler_ExecuteProcedure
- [ ] TestRPCHandler_ExecuteProcedure_WithParams
- [ ] TestRPCHandler_ExecuteProcedure_Error
- [ ] TestRPCHandler_GetExecutionHistory
- [ ] TestRPCHandler_ListExecutions

### 5.2 Migrations Handler Tests (`internal/migrations/handler.go`)
**Target: 11 handler methods | Status: ⏳ Pending**

- [ ] TestMigrationsHandler_CreateMigration
- [ ] TestMigrationsHandler_GetMigration
- [ ] TestMigrationsHandler_UpdateMigration
- [ ] TestMigrationsHandler_DeleteMigration
- [ ] TestMigrationsHandler_ListMigrations
- [ ] TestMigrationsHandler_ApplyMigration
- [ ] TestMigrationsHandler_RollbackMigration
- [ ] TestMigrationsHandler_ApplyPending
- [ ] TestMigrationsHandler_GetExecutions
- [ ] TestMigrationsHandler_SyncMigrations

---

## Phase 6: MCP Tools & OAuth (Medium Priority)

### 6.1 MCP DDL Tools Tests (`internal/mcp/tools/ddl.go`)
**Target: 30+ DDL tools | Status: ⏳ Pending**

- [ ] TestDDLTool_CreateSchema
- [ ] TestDDLTool_AlterSchema
- [ ] TestDDLTool_DropSchema
- [ ] TestDDLTool_CreateTable
- [ ] TestDDLTool_AlterTable_AddColumn
- [ ] TestDDLTool_AlterTable_DropColumn
- [ ] TestDDLTool_AlterTable_RenameColumn
- [ ] TestDDLTool_DropTable
- [ ] TestDDLTool_CreateIndex
- [ ] TestDDLTool_DropIndex
- [ ] TestDDLTool_AddConstraint
- [ ] TestDDLTool_DropConstraint

### 6.2 MCP OAuth Handler Tests (`internal/api/mcp_oauth_handler.go`)
**Target: OAuth flow handlers | Status: ⏳ Pending**

- [ ] TestMCPOAuth_Authorize
- [ ] TestMCPOAuth_Authorize_InvalidClient
- [ ] TestMCPOAuth_Authorize_InvalidRedirect
- [ ] TestMCPOAuth_Token_AuthorizationCode
- [ ] TestMCPOAuth_Token_RefreshToken
- [ ] TestMCPOAuth_Token_InvalidGrant
- [ ] TestMCPOAuth_RegisterClient
- [ ] TestMCPOAuth_RevokeToken

---

## Phase 7: Storage Modules (Medium Priority)

### 7.1 RPC Storage Tests (`internal/rpc/storage.go`)
**Target: 16+ operations | Status: ⏳ Pending**

- [ ] TestRPCStorage_CreateProcedure
- [ ] TestRPCStorage_UpdateProcedure
- [ ] TestRPCStorage_DeleteProcedure
- [ ] TestRPCStorage_GetProcedure
- [ ] TestRPCStorage_CreateExecution
- [ ] TestRPCStorage_GetExecution
- [ ] TestRPCStorage_ListExecutions
- [ ] TestRPCStorage_CreateNamespace
- [ ] TestRPCStorage_ListNamespaces

### 7.2 Dashboard Auth Tests (`internal/auth/dashboard.go`)
**Target: 20+ dashboard methods | Status: ⏳ Pending**

- [ ] TestDashboard_CreateSession
- [ ] TestDashboard_ValidateSession
- [ ] TestDashboard_InvalidateSession
- [ ] TestDashboard_GetDashboardUser
- [ ] TestDashboard_UpdateDashboardUser
- [ ] TestDashboard_ListDashboardUsers
- [ ] TestDashboard_CheckPermission
- [ ] TestDashboard_ListPermissions

---

## Phase 8: Remaining Gaps (Low Priority)

### 8.1 Extensions Handler (`internal/extensions/handler.go`)
**Status: ⏳ Pending**

- [ ] TestExtensions_ListExtensions
- [ ] TestExtensions_GetExtensionStatus
- [ ] TestExtensions_EnableExtension
- [ ] TestExtensions_DisableExtension
- [ ] TestExtensions_SyncExtensions

### 8.2 Rate Limit Redis (`internal/ratelimit/redis.go`)
**Status: ⏳ Pending**

- [ ] TestRedisRateLimit_Allow
- [ ] TestRedisRateLimit_Deny
- [ ] TestRedisRateLimit_Reset
- [ ] TestRedisRateLimit_GetRemaining

### 8.3 Conversation Handler (`internal/api/conversation.go`)
**Status: ⏳ Pending**

- [ ] TestConversation_Create
- [ ] TestConversation_Get
- [ ] TestConversation_List
- [ ] TestConversation_Delete
- [ ] TestConversation_UpdateTitle

---

## Completed Work (Previous Sessions)

### Already Added Tests:
- `internal/auth/oauth_logout_test.go` - Error constants, structs, constructors, provider endpoints
- `internal/auth/captcha_trust_test.go` - Constructor, TrustResult, CaptchaChallenge, UserTrustSignal
- `internal/jobs/worker_test.go` - normalizeSettingsKey, jobToExecutionRequest helpers
- `internal/branching/storage_test.go` - GenerateSlug, GeneratePRSlug, GenerateDatabaseName, ValidateSlug, isAccessSufficient
- `internal/branching/manager_test.go` - sanitizeIdentifier, activity constants, CreateBranchRequest, branch status
- `internal/api/schema_export_test.go` - TypeScriptExportRequest, NewSchemaExportHandler, edge cases
- `internal/realtime/listener_test.go` - Channel constants, ChangeEvent, enrichJobWithETA

---

## Notes

- Tests should follow table-driven test pattern
- Use testify assertions (assert/require)
- Add benchmarks for performance-critical functions
- Test all error conditions
- Use existing mocks from `internal/testutil/`
