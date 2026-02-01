---
editUrl: false
next: false
prev: false
title: "FunctionSpec"
---

Function specification for bulk sync operations

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="allow_env"></a> `allow_env?` | `boolean` | - |
| <a id="allow_net"></a> `allow_net?` | `boolean` | - |
| <a id="allow_read"></a> `allow_read?` | `boolean` | - |
| <a id="allow_unauthenticated"></a> `allow_unauthenticated?` | `boolean` | - |
| <a id="allow_write"></a> `allow_write?` | `boolean` | - |
| <a id="code"></a> `code` | `string` | - |
| <a id="cron_schedule"></a> `cron_schedule?` | `string` | - |
| <a id="description"></a> `description?` | `string` | - |
| <a id="enabled"></a> `enabled?` | `boolean` | - |
| <a id="is_pre_bundled"></a> `is_pre_bundled?` | `boolean` | If true, code is already bundled and server will skip bundling |
| <a id="is_public"></a> `is_public?` | `boolean` | - |
| <a id="memory_limit_mb"></a> `memory_limit_mb?` | `number` | - |
| <a id="name"></a> `name` | `string` | - |
| <a id="nodepaths"></a> `nodePaths?` | `string`[] | Additional paths to search for node_modules during bundling (used by syncWithBundling) |
| <a id="original_code"></a> `original_code?` | `string` | Original source code (for debugging when pre-bundled) |
| <a id="sourcedir"></a> `sourceDir?` | `string` | Source directory for resolving relative imports during bundling (used by syncWithBundling) |
| <a id="timeout_seconds"></a> `timeout_seconds?` | `number` | - |
