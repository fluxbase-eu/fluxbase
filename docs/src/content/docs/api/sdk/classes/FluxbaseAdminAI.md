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

### addDocument()

> **addDocument**(`knowledgeBaseId`, `request`): `Promise`\<\{ `data`: `AddDocumentResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Add a document to a knowledge base

Document will be chunked and embedded asynchronously.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `request` | `AddDocumentRequest` | Document content and metadata |

#### Returns

`Promise`\<\{ `data`: `AddDocumentResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with document ID

#### Example

```typescript
const { data, error } = await client.admin.ai.addDocument('kb-uuid', {
  title: 'Getting Started Guide',
  content: 'This is the content of the document...',
  metadata: { category: 'guides' },
})
if (data) {
  console.log('Document ID:', data.document_id)
}
```

***

### clearEmbeddingProvider()

> **clearEmbeddingProvider**(`id`): `Promise`\<\{ `data`: \{ `use_for_embeddings`: `boolean`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Clear explicit embedding provider preference (revert to default)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID to clear |

#### Returns

`Promise`\<\{ `data`: \{ `use_for_embeddings`: `boolean`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.clearEmbeddingProvider('uuid')
```

***

### createKnowledgeBase()

> **createKnowledgeBase**(`request`): `Promise`\<\{ `data`: `KnowledgeBase` \| `null`; `error`: `Error` \| `null`; \}\>

Create a new knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | `CreateKnowledgeBaseRequest` | Knowledge base configuration |

#### Returns

`Promise`\<\{ `data`: `KnowledgeBase` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with created knowledge base

#### Example

```typescript
const { data, error } = await client.admin.ai.createKnowledgeBase({
  name: 'product-docs',
  description: 'Product documentation',
  chunk_size: 512,
  chunk_overlap: 50,
})
```

***

### createProvider()

> **createProvider**(`request`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Create a new AI provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`CreateAIProviderRequest`](/api/sdk/interfaces/createaiproviderrequest/) | Provider configuration |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with created provider

#### Example

```typescript
const { data, error } = await client.admin.ai.createProvider({
  name: 'openai-main',
  display_name: 'OpenAI (Main)',
  provider_type: 'openai',
  is_default: true,
  config: {
    api_key: 'sk-...',
    model: 'gpt-4-turbo',
  }
})
```

***

### createTableExportSync()

> **createTableExportSync**(`knowledgeBaseId`, `config`): `Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: `Error` \| `null`; \}\>

Create a table export preset

Saves a table export configuration for easy re-export. Use triggerTableExportSync
to re-export when the schema changes.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `config` | [`CreateTableExportSyncConfig`](/api/sdk/interfaces/createtableexportsyncconfig/) | Export preset configuration |

#### Returns

`Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with created preset

#### Example

```typescript
const { data, error } = await client.admin.ai.createTableExportSync('kb-uuid', {
  schema_name: 'public',
  table_name: 'products',
  columns: ['id', 'name', 'description', 'price'],
  include_foreign_keys: true,
  export_now: true, // Trigger initial export
})
```

***

### deleteChatbot()

> **deleteChatbot**(`id`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.deleteChatbot('uuid')
```

***

### deleteDocument()

> **deleteDocument**(`knowledgeBaseId`, `documentId`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete a document from a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `documentId` | `string` | Document ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.deleteDocument('kb-uuid', 'doc-uuid')
```

***

### deleteDocumentsByFilter()

> **deleteDocumentsByFilter**(`knowledgeBaseId`, `filter`): `Promise`\<\{ `data`: `DeleteDocumentsByFilterResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Delete documents from a knowledge base by filter

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `filter` | `DeleteDocumentsByFilterRequest` | Filter criteria for deletion |

#### Returns

`Promise`\<\{ `data`: `DeleteDocumentsByFilterResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with deletion count

#### Example

```typescript
// Delete by tags
const { data, error } = await client.admin.ai.deleteDocumentsByFilter('kb-uuid', {
  tags: ['deprecated', 'archive'],
})

// Delete by metadata
const { data, error } = await client.admin.ai.deleteDocumentsByFilter('kb-uuid', {
  metadata: { source: 'legacy-system' },
})

if (data) {
  console.log(`Deleted ${data.deleted_count} documents`)
}
```

***

### deleteKnowledgeBase()

> **deleteKnowledgeBase**(`id`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.deleteKnowledgeBase('uuid')
```

***

### deleteProvider()

> **deleteProvider**(`id`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete a provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.deleteProvider('uuid')
```

***

### deleteTableExportSync()

> **deleteTableExportSync**(`knowledgeBaseId`, `syncId`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete a table export sync configuration

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `syncId` | `string` | Sync config ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.deleteTableExportSync('kb-uuid', 'sync-id')
```

***

### exportTable()

> **exportTable**(`knowledgeBaseId`, `options`): `Promise`\<\{ `data`: [`ExportTableResult`](/api/sdk/interfaces/exporttableresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Export a database table to a knowledge base

The table schema will be exported as a markdown document and indexed.
Optionally filter which columns to export for security or relevance.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `options` | [`ExportTableOptions`](/api/sdk/interfaces/exporttableoptions/) | Export options including column selection |

#### Returns

`Promise`\<\{ `data`: [`ExportTableResult`](/api/sdk/interfaces/exporttableresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with export result

#### Example

```typescript
// Export all columns
const { data, error } = await client.admin.ai.exportTable('kb-uuid', {
  schema: 'public',
  table: 'users',
  include_foreign_keys: true,
})

// Export specific columns (recommended for sensitive data)
const { data, error } = await client.admin.ai.exportTable('kb-uuid', {
  schema: 'public',
  table: 'users',
  columns: ['id', 'name', 'email', 'created_at'],
})
```

***

### getCapabilities()

> **getCapabilities**(): `Promise`\<\{ `data`: `KnowledgeBaseCapabilities` \| `null`; `error`: `Error` \| `null`; \}\>

Get knowledge base system capabilities

Returns information about OCR support, supported file types, etc.

#### Returns

`Promise`\<\{ `data`: `KnowledgeBaseCapabilities` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with capabilities

#### Example

```typescript
const { data, error } = await client.admin.ai.getCapabilities()
if (data) {
  console.log('OCR available:', data.ocr_available)
  console.log('Supported types:', data.supported_file_types)
}
```

***

### getChatbot()

> **getChatbot**(`id`): `Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: `Error` \| `null`; \}\>

Get details of a specific chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |

#### Returns

`Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with chatbot details

#### Example

```typescript
const { data, error } = await client.admin.ai.getChatbot('uuid')
if (data) {
  console.log('Chatbot:', data.name)
}
```

***

### getDocument()

> **getDocument**(`knowledgeBaseId`, `documentId`): `Promise`\<\{ `data`: `KnowledgeBaseDocument` \| `null`; `error`: `Error` \| `null`; \}\>

Get a specific document

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `documentId` | `string` | Document ID |

#### Returns

`Promise`\<\{ `data`: `KnowledgeBaseDocument` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with document details

#### Example

```typescript
const { data, error } = await client.admin.ai.getDocument('kb-uuid', 'doc-uuid')
```

***

### getEntityRelationships()

> **getEntityRelationships**(`knowledgeBaseId`, `entityId`): `Promise`\<\{ `data`: `EntityRelationship`[] \| `null`; `error`: `Error` \| `null`; \}\>

Get relationships for a specific entity

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `entityId` | `string` | Entity ID |

#### Returns

`Promise`\<\{ `data`: `EntityRelationship`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with entity relationships

#### Example

```typescript
const { data, error } = await client.admin.ai.getEntityRelationships('kb-uuid', 'entity-uuid')
if (data) {
  console.log('Relationships:', data.map(r => `${r.relationship_type} -> ${r.target_entity?.name}`))
}
```

***

### getKnowledgeBase()

> **getKnowledgeBase**(`id`): `Promise`\<\{ `data`: `KnowledgeBase` \| `null`; `error`: `Error` \| `null`; \}\>

Get a specific knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: `KnowledgeBase` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with knowledge base details

#### Example

```typescript
const { data, error } = await client.admin.ai.getKnowledgeBase('uuid')
if (data) {
  console.log('Knowledge base:', data.name)
}
```

***

### getKnowledgeGraph()

> **getKnowledgeGraph**(`knowledgeBaseId`): `Promise`\<\{ `data`: `KnowledgeGraphData` \| `null`; `error`: `Error` \| `null`; \}\>

Get the knowledge graph for a knowledge base

Returns all entities and relationships for visualization.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: `KnowledgeGraphData` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with graph data

#### Example

```typescript
const { data, error } = await client.admin.ai.getKnowledgeGraph('kb-uuid')
if (data) {
  console.log('Graph:', data.entity_count, 'entities,', data.relationship_count, 'relationships')
  // Use with visualization libraries like D3.js, Cytoscape.js, etc.
}
```

***

### getProvider()

> **getProvider**(`id`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Get details of a specific provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with provider details

#### Example

```typescript
const { data, error } = await client.admin.ai.getProvider('uuid')
if (data) {
  console.log('Provider:', data.display_name)
}
```

***

### getTableDetails()

> **getTableDetails**(`schema`, `table`): `Promise`\<\{ `data`: [`TableDetails`](/api/sdk/interfaces/tabledetails/) \| `null`; `error`: `Error` \| `null`; \}\>

Get detailed table information including columns

Use this to discover available columns before exporting.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `schema` | `string` | Schema name (e.g., 'public') |
| `table` | `string` | Table name |

#### Returns

`Promise`\<\{ `data`: [`TableDetails`](/api/sdk/interfaces/tabledetails/) \| `null`; `error`: `Error` \| `null`; \}\>

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

### linkKnowledgeBase()

> **linkKnowledgeBase**(`chatbotId`, `request`): `Promise`\<\{ `data`: `ChatbotKnowledgeBaseLink` \| `null`; `error`: `Error` \| `null`; \}\>

Link a knowledge base to a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |
| `request` | `LinkKnowledgeBaseRequest` | Link configuration |

#### Returns

`Promise`\<\{ `data`: `ChatbotKnowledgeBaseLink` \| `null`; `error`: `Error` \| `null`; \}\>

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

> **listChatbotKnowledgeBases**(`chatbotId`): `Promise`\<\{ `data`: `ChatbotKnowledgeBaseLink`[] \| `null`; `error`: `Error` \| `null`; \}\>

List knowledge bases linked to a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |

#### Returns

`Promise`\<\{ `data`: `ChatbotKnowledgeBaseLink`[] \| `null`; `error`: `Error` \| `null`; \}\>

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

> **listChatbots**(`namespace?`): `Promise`\<\{ `data`: [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List all chatbots (admin view)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace?` | `string` | Optional namespace filter |

#### Returns

`Promise`\<\{ `data`: [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of chatbot summaries

#### Example

```typescript
const { data, error } = await client.admin.ai.listChatbots()
if (data) {
  console.log('Chatbots:', data.map(c => c.name))
}
```

***

### listChatbotsUsingKB()

> **listChatbotsUsingKB**(`knowledgeBaseId`): `Promise`\<\{ `data`: [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List all chatbots that use a specific knowledge base

Reverse lookup to find which chatbots are linked to a knowledge base.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: [`AIChatbotSummary`](/api/sdk/interfaces/aichatbotsummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of chatbot summaries

#### Example

```typescript
const { data, error } = await client.admin.ai.listChatbotsUsingKB('kb-uuid')
if (data) {
  console.log('Used by chatbots:', data.map(c => c.name))
}
```

***

### listDocuments()

> **listDocuments**(`knowledgeBaseId`): `Promise`\<\{ `data`: `KnowledgeBaseDocument`[] \| `null`; `error`: `Error` \| `null`; \}\>

List documents in a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: `KnowledgeBaseDocument`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of documents

#### Example

```typescript
const { data, error } = await client.admin.ai.listDocuments('kb-uuid')
if (data) {
  console.log('Documents:', data.map(d => d.title))
}
```

***

### listEntities()

> **listEntities**(`knowledgeBaseId`, `entityType?`): `Promise`\<\{ `data`: `Entity`[] \| `null`; `error`: `Error` \| `null`; \}\>

List entities in a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `entityType?` | `string` | Optional entity type filter |

#### Returns

`Promise`\<\{ `data`: `Entity`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of entities

#### Example

```typescript
// List all entities
const { data, error } = await client.admin.ai.listEntities('kb-uuid')

// Filter by type
const { data, error } = await client.admin.ai.listEntities('kb-uuid', 'person')

if (data) {
  console.log('Entities:', data.map(e => e.name))
}
```

***

### listKnowledgeBases()

> **listKnowledgeBases**(): `Promise`\<\{ `data`: `KnowledgeBaseSummary`[] \| `null`; `error`: `Error` \| `null`; \}\>

List all knowledge bases

#### Returns

`Promise`\<\{ `data`: `KnowledgeBaseSummary`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of knowledge base summaries

#### Example

```typescript
const { data, error } = await client.admin.ai.listKnowledgeBases()
if (data) {
  console.log('Knowledge bases:', data.map(kb => kb.name))
}
```

***

### listProviders()

> **listProviders**(): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List all AI providers

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of providers

#### Example

```typescript
const { data, error } = await client.admin.ai.listProviders()
if (data) {
  console.log('Providers:', data.map(p => p.name))
}
```

***

### listTableExportSyncs()

> **listTableExportSyncs**(`knowledgeBaseId`): `Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List table export presets for a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of presets

#### Example

```typescript
const { data, error } = await client.admin.ai.listTableExportSyncs('kb-uuid')
if (data) {
  data.forEach(config => {
    console.log(`${config.schema_name}.${config.table_name}`)
  })
}
```

***

### searchEntities()

> **searchEntities**(`knowledgeBaseId`, `query`, `types?`): `Promise`\<\{ `data`: `Entity`[] \| `null`; `error`: `Error` \| `null`; \}\>

Search for entities in a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `query` | `string` | Search query |
| `types?` | `string`[] | Optional entity type filters |

#### Returns

`Promise`\<\{ `data`: `Entity`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with matching entities

#### Example

```typescript
// Search all entity types
const { data, error } = await client.admin.ai.searchEntities('kb-uuid', 'John')

// Search specific types
const { data, error } = await client.admin.ai.searchEntities('kb-uuid', 'Apple', ['organization', 'product'])

if (data) {
  console.log('Found entities:', data.map(e => `${e.name} (${e.entity_type})`))
}
```

***

### searchKnowledgeBase()

> **searchKnowledgeBase**(`knowledgeBaseId`, `query`, `options?`): `Promise`\<\{ `data`: `SearchKnowledgeBaseResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Search a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `query` | `string` | Search query |
| `options?` | \{ `max_chunks?`: `number`; `threshold?`: `number`; \} | Search options |
| `options.max_chunks?` | `number` | - |
| `options.threshold?` | `number` | - |

#### Returns

`Promise`\<\{ `data`: `SearchKnowledgeBaseResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with search results

#### Example

```typescript
const { data, error } = await client.admin.ai.searchKnowledgeBase('kb-uuid', 'how to reset password', {
  max_chunks: 5,
  threshold: 0.7,
})
if (data) {
  console.log('Results:', data.results.map(r => r.content))
}
```

***

### setDefaultProvider()

> **setDefaultProvider**(`id`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Set a provider as the default

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated provider

#### Example

```typescript
const { data, error } = await client.admin.ai.setDefaultProvider('uuid')
```

***

### setEmbeddingProvider()

> **setEmbeddingProvider**(`id`): `Promise`\<\{ `data`: \{ `id`: `string`; `use_for_embeddings`: `boolean`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Set a provider as the embedding provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |

#### Returns

`Promise`\<\{ `data`: \{ `id`: `string`; `use_for_embeddings`: `boolean`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.setEmbeddingProvider('uuid')
```

***

### sync()

> **sync**(`options?`): `Promise`\<\{ `data`: [`SyncChatbotsResult`](/api/sdk/interfaces/syncchatbotsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

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

`Promise`\<\{ `data`: [`SyncChatbotsResult`](/api/sdk/interfaces/syncchatbotsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

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

### toggleChatbot()

> **toggleChatbot**(`id`, `enabled`): `Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: `Error` \| `null`; \}\>

Enable or disable a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Chatbot ID |
| `enabled` | `boolean` | Whether to enable or disable |

#### Returns

`Promise`\<\{ `data`: [`AIChatbot`](/api/sdk/interfaces/aichatbot/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated chatbot

#### Example

```typescript
const { data, error } = await client.admin.ai.toggleChatbot('uuid', true)
```

***

### triggerTableExportSync()

> **triggerTableExportSync**(`knowledgeBaseId`, `syncId`): `Promise`\<\{ `data`: [`ExportTableResult`](/api/sdk/interfaces/exporttableresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Manually trigger a table export sync

Immediately re-exports the table to the knowledge base,
regardless of the sync mode.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `syncId` | `string` | Sync config ID |

#### Returns

`Promise`\<\{ `data`: [`ExportTableResult`](/api/sdk/interfaces/exporttableresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with export result

#### Example

```typescript
const { data, error } = await client.admin.ai.triggerTableExportSync('kb-uuid', 'sync-id')
if (data) {
  console.log('Exported document:', data.document_id)
}
```

***

### unlinkKnowledgeBase()

> **unlinkKnowledgeBase**(`chatbotId`, `knowledgeBaseId`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Unlink a knowledge base from a chatbot

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |
| `knowledgeBaseId` | `string` | Knowledge base ID |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.ai.unlinkKnowledgeBase('chatbot-uuid', 'kb-uuid')
```

***

### updateChatbotKnowledgeBase()

> **updateChatbotKnowledgeBase**(`chatbotId`, `knowledgeBaseId`, `updates`): `Promise`\<\{ `data`: `ChatbotKnowledgeBaseLink` \| `null`; `error`: `Error` \| `null`; \}\>

Update a chatbot-knowledge base link

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `chatbotId` | `string` | Chatbot ID |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `updates` | `UpdateChatbotKnowledgeBaseRequest` | Fields to update |

#### Returns

`Promise`\<\{ `data`: `ChatbotKnowledgeBaseLink` \| `null`; `error`: `Error` \| `null`; \}\>

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

### updateDocument()

> **updateDocument**(`knowledgeBaseId`, `documentId`, `updates`): `Promise`\<\{ `data`: `KnowledgeBaseDocument` \| `null`; `error`: `Error` \| `null`; \}\>

Update a document in a knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `documentId` | `string` | Document ID |
| `updates` | `UpdateDocumentRequest` | Fields to update |

#### Returns

`Promise`\<\{ `data`: `KnowledgeBaseDocument` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated document

#### Example

```typescript
const { data, error } = await client.admin.ai.updateDocument('kb-uuid', 'doc-uuid', {
  title: 'Updated Title',
  tags: ['updated', 'tag'],
  metadata: { category: 'updated' },
})
```

***

### updateKnowledgeBase()

> **updateKnowledgeBase**(`id`, `updates`): `Promise`\<\{ `data`: `KnowledgeBase` \| `null`; `error`: `Error` \| `null`; \}\>

Update an existing knowledge base

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Knowledge base ID |
| `updates` | `UpdateKnowledgeBaseRequest` | Fields to update |

#### Returns

`Promise`\<\{ `data`: `KnowledgeBase` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated knowledge base

#### Example

```typescript
const { data, error } = await client.admin.ai.updateKnowledgeBase('uuid', {
  description: 'Updated description',
  enabled: true,
})
```

***

### updateProvider()

> **updateProvider**(`id`, `updates`): `Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Update an existing AI provider

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Provider ID |
| `updates` | [`UpdateAIProviderRequest`](/api/sdk/interfaces/updateaiproviderrequest/) | Fields to update |

#### Returns

`Promise`\<\{ `data`: [`AIProvider`](/api/sdk/interfaces/aiprovider/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated provider

#### Example

```typescript
const { data, error } = await client.admin.ai.updateProvider('uuid', {
  display_name: 'Updated Name',
  config: {
    api_key: 'new-key',
    model: 'gpt-4-turbo',
  },
  enabled: true,
})
```

***

### updateTableExportSync()

> **updateTableExportSync**(`knowledgeBaseId`, `syncId`, `updates`): `Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: `Error` \| `null`; \}\>

Update a table export preset

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `syncId` | `string` | Preset ID |
| `updates` | [`UpdateTableExportSyncConfig`](/api/sdk/interfaces/updatetableexportsyncconfig/) | Fields to update |

#### Returns

`Promise`\<\{ `data`: [`TableExportSyncConfig`](/api/sdk/interfaces/tableexportsyncconfig/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated preset

#### Example

```typescript
const { data, error } = await client.admin.ai.updateTableExportSync('kb-uuid', 'sync-id', {
  columns: ['id', 'name', 'email', 'updated_at'],
})
```

***

### uploadDocument()

> **uploadDocument**(`knowledgeBaseId`, `file`, `title?`): `Promise`\<\{ `data`: `UploadDocumentResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Upload a document file to a knowledge base

Supported file types: PDF, TXT, MD, HTML, CSV, DOCX, XLSX, RTF, EPUB, JSON
Maximum file size: 50MB

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `knowledgeBaseId` | `string` | Knowledge base ID |
| `file` | `Blob` \| `File` | File to upload (File or Blob) |
| `title?` | `string` | Optional document title (defaults to filename without extension) |

#### Returns

`Promise`\<\{ `data`: `UploadDocumentResponse` \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with upload result

#### Example

```typescript
// Browser
const fileInput = document.getElementById('file') as HTMLInputElement
const file = fileInput.files?.[0]
if (file) {
  const { data, error } = await client.admin.ai.uploadDocument('kb-uuid', file)
  if (data) {
    console.log('Document ID:', data.document_id)
    console.log('Extracted length:', data.extracted_length)
  }
}

// Node.js (with node-fetch or similar)
import { Blob } from 'buffer'
const content = await fs.readFile('document.pdf')
const blob = new Blob([content], { type: 'application/pdf' })
const { data, error } = await client.admin.ai.uploadDocument('kb-uuid', blob, 'My Document')
```
