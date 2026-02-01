---
editUrl: false
next: false
prev: false
title: "SyncMigrationsResult"
---

Result of a migration sync operation

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="details"></a> `details` | `object` | Detailed results |
| `details.applied` | `string`[] | - |
| `details.created` | `string`[] | - |
| `details.errors` | `string`[] | - |
| `details.skipped` | `string`[] | - |
| `details.unchanged` | `string`[] | - |
| `details.updated` | `string`[] | - |
| <a id="dry_run"></a> `dry_run` | `boolean` | Whether this was a dry run |
| <a id="message"></a> `message` | `string` | Status message |
| <a id="namespace"></a> `namespace` | `string` | Namespace that was synced |
| <a id="summary"></a> `summary` | `object` | Summary counts |
| `summary.applied` | `number` | - |
| `summary.created` | `number` | - |
| `summary.errors` | `number` | - |
| `summary.skipped` | `number` | - |
| `summary.unchanged` | `number` | - |
| `summary.updated` | `number` | - |
| <a id="warnings"></a> `warnings?` | `string`[] | Warning messages |
