---
editUrl: false
next: false
prev: false
title: "VectorSearchResult"
---

Result from vector search

## Type Parameters

| Type Parameter | Default type |
| ------ | ------ |
| `T` | `Record`\<`string`, `unknown`\> |

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="data"></a> `data` | `T`[] | Matched records |
| <a id="distances"></a> `distances` | `number`[] | Distance scores for each result |
| <a id="model"></a> `model?` | `string` | Embedding model used (if query text was embedded) |
