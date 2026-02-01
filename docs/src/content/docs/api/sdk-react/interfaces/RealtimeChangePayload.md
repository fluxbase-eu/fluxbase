---
editUrl: false
next: false
prev: false
title: "RealtimeChangePayload"
---

:::caution[Deprecated]
Use RealtimePostgresChangesPayload instead
:::

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="new_record"></a> ~~`new_record?`~~ | `Record`\<`string`, `unknown`\> | :::caution[Deprecated] Use 'new' instead ::: |
| <a id="old_record"></a> ~~`old_record?`~~ | `Record`\<`string`, `unknown`\> | :::caution[Deprecated] Use 'old' instead ::: |
| <a id="schema"></a> ~~`schema`~~ | `string` | - |
| <a id="table"></a> ~~`table`~~ | `string` | - |
| <a id="timestamp"></a> ~~`timestamp`~~ | `string` | :::caution[Deprecated] Use commit_timestamp instead ::: |
| <a id="type"></a> ~~`type`~~ | `"INSERT"` \| `"UPDATE"` \| `"DELETE"` | :::caution[Deprecated] Use eventType instead ::: |
