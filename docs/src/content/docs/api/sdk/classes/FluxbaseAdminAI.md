---
editUrl: false
next: false
prev: false
title: "FluxbaseAdminAI"
---

Admin AI manager for managing AI chatbots and providers
Provides create, update, delete, sync, and monitoring operations

## Constructors

### Constructor

> **new FluxbaseAdminAI**(`fetch`): `FluxbaseAdminAI`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseAdminAI`

## Methods

### clearEmbeddingProvider()

> **clearEmbeddingProvider**(`id`): `Promise`\<\{ `data`: \{ `use_for_embeddings`: `boolean`; \} \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Clear the embedding provider assignment for a provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: \{ `use_for_embeddings`: `boolean`; \} \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple

***

### createIntegration()

> **createIntegration**(`request`): `Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Create a new tool integration. api_key in config is encrypted at the
storage layer; never appears in plaintext in API responses.

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `request` | [`CreateToolIntegrationRequest`](/api/sdk/interfaces/createtoolintegrationrequest/) |

#### Returns

`Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### createProvider()

> **createProvider**(`params`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Create a new AI provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `params` | \{ `config?`: `Record`\<`string`, `unknown`\>; `display_name?`: `string`; `is_default?`: `boolean`; `name`: `string`; `provider_type`: `string`; \} | Provider configuration including name, provider_type, and optional config |
| `params.config?` | `Record`\<`string`, `unknown`\> | - |
| `params.display_name?` | `string` | - |
| `params.is_default?` | `boolean` | - |
| `params.name` | `string` | - |
| `params.provider_type` | `string` | - |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with created provider

***

### createSyncConfig()

> **createSyncConfig**(`kbId`, `request`): `Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Create a table export sync configuration for a knowledge base

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `request` | [`CreateSyncConfigRequest`](/api/sdk/interfaces/createsyncconfigrequest/) |

#### Returns

`Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### deleteChatbot()

> **deleteChatbot**(`id`): `Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Delete a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.deleteChatbot('uuid')
```

***

### deleteIntegration()

> **deleteIntegration**(`id`): `Promise`\<\{ `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Delete a tool integration. Refuses to delete read-only integrations
configured via env/YAML.

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `id` | `string` |

#### Returns

`Promise`\<\{ `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### deleteProvider()

> **deleteProvider**(`id`): `Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Delete an AI provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple

***

### deleteSyncConfig()

> **deleteSyncConfig**(`kbId`, `syncId`): `Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Delete a table export sync configuration

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `syncId` | `string` |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### exportTableToKB()

> **exportTableToKB**(`schema`, `table`, `request`): `Promise`\<\{ `data`: [`ExportTableToKBResponse`](/api/sdk/interfaces/exporttabletokbresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Export a database table into a knowledge base as a document

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `schema` | `string` |
| `table` | `string` |
| `request` | [`ExportTableToKBRequest`](/api/sdk/interfaces/exporttabletokbrequest/) |

#### Returns

`Promise`\<\{ `data`: [`ExportTableToKBResponse`](/api/sdk/interfaces/exporttabletokbresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### getAuditLog()

> **getAuditLog**(`query?`): `Promise`\<\{ `data`: [`AuditLogResponse`](/api/sdk/interfaces/auditlogresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get the AI SQL audit log (admin)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `query` | [`AuditLogQuery`](/api/sdk/interfaces/auditlogquery/) |

#### Returns

`Promise`\<\{ `data`: [`AuditLogResponse`](/api/sdk/interfaces/auditlogresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### getChatbot()

> **getChatbot**(`id`): `Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get details of a specific chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |

#### Returns

`Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with chatbot details

#### Example

```typescript
const { data, error } = await client.admin.ai.getChatbot('uuid')
if (data) {
  console.log('Chatbot:', data.name)
}
```

***

### getConversationMessages()

> **getConversationMessages**(`conversationId`, `query?`): `Promise`\<\{ `data`: [`AdminMessagesResponse`](/api/sdk/interfaces/adminmessagesresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get all messages of a conversation (admin)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `conversationId` | `string` |
| `query` | \{ `limit?`: `number`; `offset?`: `number`; \} |
| `query.limit?` | `number` |
| `query.offset?` | `number` |

#### Returns

`Promise`\<\{ `data`: [`AdminMessagesResponse`](/api/sdk/interfaces/adminmessagesresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### getIntegration()

> **getIntegration**(`id`): `Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get a single tool integration by ID.

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `id` | `string` |

#### Returns

`Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### getMetrics()

> **getMetrics**(): `Promise`\<\{ `data`: [`AIMetrics`](/api/sdk/interfaces/aimetrics/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get aggregated AI metrics (requests, tokens, per-chatbot/provider stats)

#### Returns

`Promise`\<\{ `data`: [`AIMetrics`](/api/sdk/interfaces/aimetrics/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

#### Example

```typescript
const { data } = await client.admin.ai.getMetrics()
```

***

### getProvider()

> **getProvider**(`id`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get details of a specific AI provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with provider details

***

### getTableDetails()

> **getTableDetails**(`schema`, `table`): `Promise`\<\{ `data`: [`TableDetails`](/api/sdk/interfaces/tabledetails/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Get detailed table information including columns

Use this to discover available columns before exporting.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `schema` | `string` | Schema name (e.g., 'public') |
| `table` | `string` | Table name |

#### Returns

`Promise`\<\{ `data`: [`TableDetails`](/api/sdk/interfaces/tabledetails/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with table details

#### Example

```typescript
const { data, error } = await client.admin.ai.getTableDetails('public', 'users')
if (data) {
  console.log('Columns:', data.columns.map(c => c.name))
  console.log('Primary key:', data.primary_key)
}
```

***

### grantDocumentPermission()

> **grantDocumentPermission**(`kbId`, `documentId`, `request`): `Promise`\<\{ `data`: [`DocumentPermissionGrant`](/api/sdk/interfaces/documentpermissiongrant/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Grant a user permission on a document (admin)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `documentId` | `string` |
| `request` | [`GrantDocumentPermissionRequest`](/api/sdk/interfaces/grantdocumentpermissionrequest/) |

#### Returns

`Promise`\<\{ `data`: [`DocumentPermissionGrant`](/api/sdk/interfaces/documentpermissiongrant/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### linkKnowledgeBase()

> **linkKnowledgeBase**(`chatbotId`, `request`): `Promise`\<\{ `data`: [`ChatbotKnowledgeBaseLink`](/api/sdk/interfaces/chatbotknowledgebaselink/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Link a knowledge base to a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |
| `request` | [`LinkKnowledgeBaseRequest`](/api/sdk/interfaces/linkknowledgebaserequest/) | Link configuration |

#### Returns

`Promise`\<\{ `data`: [`ChatbotKnowledgeBaseLink`](/api/sdk/interfaces/chatbotknowledgebaselink/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with link details

#### Example

```typescript
const { data, error } = await client.admin.ai.linkKnowledgeBase('chatbot-uuid', {
  knowledge_base_id: 'kb-uuid',
  priority: 1,
  max_chunks: 5,
  similarity_threshold: 0.7,
})
```

***

### listChatbotKnowledgeBases()

> **listChatbotKnowledgeBases**(`chatbotId`): `Promise`\<\{ `data`: [`ChatbotKnowledgeBaseLink`](/api/sdk/interfaces/chatbotknowledgebaselink/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List knowledge bases linked to a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |

#### Returns

`Promise`\<\{ `data`: [`ChatbotKnowledgeBaseLink`](/api/sdk/interfaces/chatbotknowledgebaselink/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with linked knowledge bases

#### Example

```typescript
const { data, error } = await client.admin.ai.listChatbotKnowledgeBases('chatbot-uuid')
if (data) {
  console.log('Linked KBs:', data.map(l => l.knowledge_base_id))
}
```

***

### listChatbots()

> **listChatbots**(`namespace?`): `Promise`\<\{ `data`: [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List all chatbots (admin view)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace?` | `string` | Optional namespace filter |

#### Returns

`Promise`\<\{ `data`: [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with array of chatbot summaries

#### Example

```typescript
const { data, error } = await client.admin.ai.listChatbots()
if (data) {
  console.log('Chatbots:', data.map(c => c.name))
}
```

***

### listConversations()

> **listConversations**(`query?`): `Promise`\<\{ `data`: [`AdminConversationListResponse`](/api/sdk/interfaces/adminconversationlistresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List AI conversations across all chatbots (admin)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `query` | [`AdminConversationQuery`](/api/sdk/interfaces/adminconversationquery/) |

#### Returns

`Promise`\<\{ `data`: [`AdminConversationListResponse`](/api/sdk/interfaces/adminconversationlistresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### listDocumentPermissions()

> **listDocumentPermissions**(`kbId`, `documentId`): `Promise`\<\{ `data`: [`DocumentPermissionGrant`](/api/sdk/interfaces/documentpermissiongrant/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List permissions granted on a document (admin)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `documentId` | `string` |

#### Returns

`Promise`\<\{ `data`: [`DocumentPermissionGrant`](/api/sdk/interfaces/documentpermissiongrant/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### listIntegrations()

> **listIntegrations**(`params?`): `Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List tool integrations, optionally filtered by integration_type.

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `params?` | \{ `integration_type?`: [`IntegrationType`](/api/sdk/type-aliases/integrationtype/); `signal?`: `AbortSignal` \| `null`; \} |
| `params.integration_type?` | [`IntegrationType`](/api/sdk/type-aliases/integrationtype/) |
| `params.signal?` | `AbortSignal` \| `null` |

#### Returns

`Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

#### Example

```ts
const { data, error } = await client.admin.ai.listIntegrations({ integration_type: "web_search" })
```

***

### listProviders()

> **listProviders**(`options?`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List all AI providers

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options` | \{ `signal?`: `AbortSignal` \| `null`; \} | Optional request options, including an AbortSignal so callers (e.g. TanStack Query) can cancel the request |
| `options.signal?` | `AbortSignal` \| `null` | - |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with array of providers

#### Example

```typescript
const { data, error } = await client.admin.ai.listProviders()
if (data) {
  console.log('Providers:', data.map(p => p.name))
}
```

***

### listSyncConfigs()

> **listSyncConfigs**(`kbId`): `Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List table export sync configurations for a knowledge base

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |

#### Returns

`Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### listTables()

> **listTables**(): `Promise`\<\{ `data`: [`ExportableTable`](/api/sdk/interfaces/exportabletable/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

List tables available for knowledge base export

#### Returns

`Promise`\<\{ `data`: [`ExportableTable`](/api/sdk/interfaces/exportabletable/)[] \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### revokeDocumentPermission()

> **revokeDocumentPermission**(`kbId`, `documentId`, `userId`): `Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Revoke a document permission (admin)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `documentId` | `string` |
| `userId` | `string` |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### setDefaultIntegration()

> **setDefaultIntegration**(`id`): `Promise`\<\{ `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Mark an integration as the default for its integration_type within
the current tenant. The server clears any prior default in the same
transaction.

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `id` | `string` |

#### Returns

`Promise`\<\{ `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### setDefaultProvider()

> **setDefaultProvider**(`id`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Set a provider as the default provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with updated provider

***

### setEmbeddingProvider()

> **setEmbeddingProvider**(`id`): `Promise`\<\{ `data`: \{ `use_for_embeddings`: `boolean`; \} \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Set a provider as the embedding provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: \{ `use_for_embeddings`: `boolean`; \} \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with updated provider

***

### sync()

> **sync**(`options?`): `Promise`\<\{ `data`: [`SyncChatbotsResult`](/api/sdk/interfaces/syncchatbotsresult/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Sync chatbots from filesystem or API payload

Can sync from:
1. Filesystem (if no chatbots provided) - loads from configured chatbots directory
2. API payload (if chatbots array provided) - syncs provided chatbot specifications

Requires service_role or admin authentication.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options?` | [`SyncChatbotsOptions`](/api/sdk/interfaces/syncchatbotsoptions/) | Sync options including namespace and optional chatbots array |

#### Returns

`Promise`\<\{ `data`: [`SyncChatbotsResult`](/api/sdk/interfaces/syncchatbotsresult/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with sync results

#### Example

```typescript
// Sync from filesystem
const { data, error } = await client.admin.ai.sync()

// Sync with provided chatbot code
const { data, error } = await client.admin.ai.sync({
  namespace: 'default',
  chatbots: [{
    name: 'sql-assistant',
    code: myChatbotCode,
  }],
  options: {
    delete_missing: false, // Don't remove chatbots not in this sync
    dry_run: false,        // Preview changes without applying
  }
})

if (data) {
  console.log(`Synced: ${data.summary.created} created, ${data.summary.updated} updated`)
}
```

***

### testIntegration()

> **testIntegration**(`id`): `Promise`\<\{ `data`: [`TestToolIntegrationResult`](/api/sdk/interfaces/testtoolintegrationresult/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Test an integration by running a real "hello world" call against
its provider. Stores last_tested_at + last_test_status + error
on the integration row so the admin UI can display health.

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `id` | `string` |

#### Returns

`Promise`\<\{ `data`: [`TestToolIntegrationResult`](/api/sdk/interfaces/testtoolintegrationresult/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

#### Example

```ts
const { data, error } = await client.admin.ai.testIntegration(integrationId)
if (data?.status === "ok") { console.log("Tavily credentials verified") }
```

***

### toggleChatbot()

> **toggleChatbot**(`id`, `enabled`): `Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Enable or disable a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |
| `enabled` | `boolean` | Whether to enable or disable |

#### Returns

`Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with updated chatbot

#### Example

```typescript
const { data, error } = await client.admin.ai.toggleChatbot('uuid', true)
```

***

### triggerSync()

> **triggerSync**(`kbId`, `syncId`): `Promise`\<\{ `data`: [`ExportTableToKBResponse`](/api/sdk/interfaces/exporttabletokbresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Trigger a table export sync immediately

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `syncId` | `string` |

#### Returns

`Promise`\<\{ `data`: [`ExportTableToKBResponse`](/api/sdk/interfaces/exporttabletokbresponse/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### unlinkKnowledgeBase()

> **unlinkKnowledgeBase**(`chatbotId`, `knowledgeBaseId`): `Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Unlink a knowledge base from a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |
| `knowledgeBaseId` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.unlinkKnowledgeBase('chatbot-uuid', 'kb-uuid')
```

***

### updateChatbot()

> **updateChatbot**(`id`, `updates`): `Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Update a chatbot's configuration (partial update).

Only the provided fields are changed; omitted fields are left untouched.
For limit/budget fields, 0 means "unlimited". Changes take effect on the
next request (limits are read fresh from the database per message).

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |
| `updates` | \{ `allow_unauthenticated?`: `boolean`; `conversation_ttl_hours?`: `number`; `daily_request_limit?`: `number`; `daily_token_budget?`: `number`; `description?`: `string`; `enabled?`: `boolean`; `is_public?`: `boolean`; `max_conversation_turns?`: `number`; `max_tokens?`: `number`; `persist_conversations?`: `boolean`; `provider_id?`: `string` \| `null`; `rate_limit_per_minute?`: `number`; `temperature?`: `number`; \} | Fields to update (all optional) |
| `updates.allow_unauthenticated?` | `boolean` | - |
| `updates.conversation_ttl_hours?` | `number` | - |
| `updates.daily_request_limit?` | `number` | - |
| `updates.daily_token_budget?` | `number` | - |
| `updates.description?` | `string` | - |
| `updates.enabled?` | `boolean` | - |
| `updates.is_public?` | `boolean` | - |
| `updates.max_conversation_turns?` | `number` | - |
| `updates.max_tokens?` | `number` | - |
| `updates.persist_conversations?` | `boolean` | - |
| `updates.provider_id?` | `string` \| `null` | - |
| `updates.rate_limit_per_minute?` | `number` | - |
| `updates.temperature?` | `number` | - |

#### Returns

`Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with the updated chatbot

#### Example

```typescript
const { data, error } = await client.admin.ai.updateChatbot('uuid', {
  daily_request_limit: 500,   // 0 = unlimited
  daily_token_budget: 200000, // 0 = unlimited
})
```

***

### updateChatbotKnowledgeBase()

> **updateChatbotKnowledgeBase**(`chatbotId`, `knowledgeBaseId`, `updates`): `Promise`\<\{ `data`: [`ChatbotKnowledgeBaseLink`](/api/sdk/interfaces/chatbotknowledgebaselink/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Update a chatbot-knowledge base link

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `updates` | [`UpdateChatbotKnowledgeBaseRequest`](/api/sdk/interfaces/updatechatbotknowledgebaserequest/) | Fields to update |

#### Returns

`Promise`\<\{ `data`: [`ChatbotKnowledgeBaseLink`](/api/sdk/interfaces/chatbotknowledgebaselink/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with updated link

#### Example

```typescript
const { data, error } = await client.admin.ai.updateChatbotKnowledgeBase(
  'chatbot-uuid',
  'kb-uuid',
  { max_chunks: 10, enabled: true }
)
```

***

### updateIntegration()

> **updateIntegration**(`id`, `request`): `Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Update an existing tool integration. Passing config.api_key =
"***masked***" preserves the existing encrypted value (useful when
the admin changes only the name and not the API key).

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `id` | `string` |
| `request` | [`UpdateToolIntegrationRequest`](/api/sdk/interfaces/updatetoolintegrationrequest/) |

#### Returns

`Promise`\<\{ `data`: [`ToolIntegration`](/api/sdk/interfaces/toolintegration/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

***

### updateProvider()

> **updateProvider**(`id`, `updates`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Update an existing AI provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |
| `updates` | \{ `config?`: `Record`\<`string`, `unknown`\>; `display_name?`: `string`; `embedding_model?`: `string` \| `null`; `enabled?`: `boolean`; \} | Fields to update |
| `updates.config?` | `Record`\<`string`, `unknown`\> | - |
| `updates.display_name?` | `string` | - |
| `updates.embedding_model?` | `string` \| `null` | - |
| `updates.enabled?` | `boolean` | - |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Promise resolving to { data, error } tuple with updated provider

***

### updateSyncConfig()

> **updateSyncConfig**(`kbId`, `syncId`, `request`): `Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>

Update a table export sync configuration

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `kbId` | `string` |
| `syncId` | `string` |
| `request` | [`UpdateSyncConfigRequest`](/api/sdk/interfaces/updatesyncconfigrequest/) |

#### Returns

`Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: [`FluxbaseError`](/api/sdk/interfaces/fluxbaseerror/) \| `null`; \}\>
