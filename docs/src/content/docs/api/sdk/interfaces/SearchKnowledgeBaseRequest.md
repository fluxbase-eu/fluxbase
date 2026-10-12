---
editUrl: false
next: false
prev: false
title: "SearchKnowledgeBaseRequest"
---

Request to search a knowledge base

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="limit"></a> `limit?` | `number` | Maximum number of chunks to return (default 10) |
| <a id="mode"></a> `mode?` | `"hybrid"` \| `"semantic"` \| `"keyword"` | Search mode: "semantic", "keyword", or "hybrid" (default depends on endpoint) |
| <a id="query"></a> `query` | `string` | - |
| <a id="rerank"></a> `rerank?` | `boolean` | Re-order results with the configured reranker (overrides the KB's rerank_enabled) |
| <a id="semantic_weight"></a> `semantic_weight?` | `number` | Hybrid mode weighting for the semantic ranked list, 0-1 (keyword weight = 1 - semantic) |
| <a id="threshold"></a> `threshold?` | `number` | Minimum semantic similarity floor (cosine); 0 disables filtering |
