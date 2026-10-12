---
editUrl: false
next: false
prev: false
title: "ExportTableToKBRequest"
---

Request to export a table into a knowledge base as a document

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="columns"></a> `columns?` | `string`[] | - |
| <a id="include_foreign_keys"></a> `include_foreign_keys?` | `boolean` | - |
| <a id="include_indexes"></a> `include_indexes?` | `boolean` | - |
| <a id="include_sample_rows"></a> `include_sample_rows?` | `boolean` | - |
| <a id="knowledge_base_id"></a> `knowledge_base_id` | `string` | - |
| <a id="sample_row_count"></a> `sample_row_count?` | `number` | - |
| <a id="user_id_column"></a> `user_id_column?` | `string` | Overrides the column marking per-user data (default "user_id" convention) |
