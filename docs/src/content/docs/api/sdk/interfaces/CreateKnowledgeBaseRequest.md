---
editUrl: false
next: false
prev: false
title: "CreateKnowledgeBaseRequest"
---

Request to create a knowledge base

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="chunk_overlap"></a> `chunk_overlap?` | `number` | - |
| <a id="chunk_size"></a> `chunk_size?` | `number` | - |
| <a id="chunk_strategy"></a> `chunk_strategy?` | `string` | - |
| <a id="description"></a> `description?` | `string` | - |
| <a id="embedding_dimensions"></a> `embedding_dimensions?` | `number` | - |
| <a id="embedding_model"></a> `embedding_model?` | `string` | - |
| <a id="entity_extraction_enabled"></a> `entity_extraction_enabled?` | `boolean` | Toggle entity extraction for this KB (default true) |
| <a id="entity_extraction_strategy"></a> `entity_extraction_strategy?` | `""` \| `"rules"` \| `"llm"` \| `"hybrid"` | Entity extraction strategy: "rules", "llm", or "hybrid" (default inherits ai.entity_extraction_strategy) |
| <a id="initial_permissions"></a> `initial_permissions?` | `object`[] | Permissions granted to users on creation |
| <a id="name"></a> `name` | `string` | - |
| <a id="namespace"></a> `namespace?` | `string` | - |
| <a id="rerank_enabled"></a> `rerank_enabled?` | `boolean` | Opt this KB's results into reranking when a reranker is configured (default true) |
| <a id="visibility"></a> `visibility?` | `"private"` \| `"shared"` \| `"public"` | - |
