---
editUrl: false
next: false
prev: false
title: "KBDebugSearchResponse"
---

Debug search response: results plus the effective search options

## Properties

| Property | Type |
| ------ | ------ |
| <a id="debug_info"></a> `debug_info?` | `Record`\<`string`, `unknown`\> |
| <a id="knowledge_base"></a> `knowledge_base?` | `object` |
| `knowledge_base.id` | `string` |
| `knowledge_base.name` | `string` |
| <a id="query"></a> `query` | `string` |
| <a id="result_count"></a> `result_count` | `number` |
| <a id="results"></a> `results` | [`KnowledgeBaseSearchResult`](/api/sdk/interfaces/knowledgebasesearchresult/)[] |
| <a id="search_options"></a> `search_options?` | `Record`\<`string`, `unknown`\> |
