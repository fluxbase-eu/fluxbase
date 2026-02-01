---
editUrl: false
next: false
prev: false
title: "SyncFunctionsOptions"
---

Options for syncing functions

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="functions"></a> `functions` | [`FunctionSpec`](/api/sdk/interfaces/functionspec/)[] | Functions to sync |
| <a id="namespace"></a> `namespace?` | `string` | Namespace to sync functions to (defaults to "default") |
| <a id="options"></a> `options?` | `object` | Options for sync operation |
| `options.delete_missing?` | `boolean` | Delete functions in namespace that are not in the sync payload |
| `options.dry_run?` | `boolean` | Preview changes without applying them |
