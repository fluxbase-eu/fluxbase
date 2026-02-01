---
editUrl: false
next: false
prev: false
title: "SyncChatbotsResult"
---

Result of a chatbot sync operation

## Properties

| Property | Type |
| ------ | ------ |
| <a id="details"></a> `details` | `object` |
| `details.created` | `string`[] |
| `details.deleted` | `string`[] |
| `details.unchanged` | `string`[] |
| `details.updated` | `string`[] |
| <a id="dry_run"></a> `dry_run` | `boolean` |
| <a id="errors"></a> `errors` | [`SyncError`](/api/sdk/interfaces/syncerror/)[] |
| <a id="message"></a> `message` | `string` |
| <a id="namespace"></a> `namespace` | `string` |
| <a id="summary"></a> `summary` | `object` |
| `summary.created` | `number` |
| `summary.deleted` | `number` |
| `summary.errors` | `number` |
| `summary.unchanged` | `number` |
| `summary.updated` | `number` |
