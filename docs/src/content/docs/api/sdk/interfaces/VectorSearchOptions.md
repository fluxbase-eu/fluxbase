---
editUrl: false
next: false
prev: false
title: "VectorSearchOptions"
---

Options for vector search via the convenience endpoint

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="column"></a> `column` | `string` | Vector column to search |
| <a id="filters"></a> `filters?` | [`QueryFilter`](/api/sdk/interfaces/queryfilter/)[] | Additional filters to apply |
| <a id="match_count"></a> `match_count?` | `number` | Maximum number of results |
| <a id="match_threshold"></a> `match_threshold?` | `number` | Minimum similarity threshold (0-1 for cosine, varies for others) |
| <a id="metric"></a> `metric?` | [`VectorMetric`](/api/sdk/type-aliases/vectormetric/) | Distance metric to use |
| <a id="query"></a> `query?` | `string` | Text query to search for (will be auto-embedded) |
| <a id="select"></a> `select?` | `string` | Columns to select (default: all) |
| <a id="table"></a> `table` | `string` | Table to search in |
| <a id="vector"></a> `vector?` | `number`[] | Direct vector input (alternative to text query) |
