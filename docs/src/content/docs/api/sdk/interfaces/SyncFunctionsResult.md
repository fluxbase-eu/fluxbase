---
editUrl: false
next: false
prev: false
title: "SyncFunctionsResult"
---

Result of a function sync operation

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="details"></a> `details` | `object` | Detailed results |
| `details.created` | `string`[] | - |
| `details.deleted` | `string`[] | - |
| `details.unchanged` | `string`[] | - |
| `details.updated` | `string`[] | - |
| <a id="dry_run"></a> `dry_run` | `boolean` | Whether this was a dry run |
| <a id="errors"></a> `errors` | [`SyncError`](/api/sdk/interfaces/syncerror/)[] | Errors encountered |
| <a id="message"></a> `message` | `string` | Status message |
| <a id="namespace"></a> `namespace` | `string` | Namespace that was synced |
| <a id="summary"></a> `summary` | `object` | Summary counts |
| `summary.created` | `number` | - |
| `summary.deleted` | `number` | - |
| `summary.errors` | `number` | - |
| `summary.unchanged` | `number` | - |
| `summary.updated` | `number` | - |
