---
editUrl: false
next: false
prev: false
title: "RealtimePostgresChangesPayload"
---

Realtime postgres_changes payload structure
Compatible with Supabase realtime payloads

## Type Parameters

| Type Parameter | Default type |
| ------ | ------ |
| `T` | `unknown` |

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="commit_timestamp"></a> `commit_timestamp` | `string` | Commit timestamp (Supabase-compatible field name) |
| <a id="errors"></a> `errors` | `string` \| `null` | Error message if any |
| <a id="eventtype"></a> `eventType` | `"DELETE"` \| `"INSERT"` \| `"UPDATE"` \| `"*"` | Event type (Supabase-compatible field name) |
| <a id="new"></a> `new` | `T` | New record data (Supabase-compatible field name) |
| <a id="old"></a> `old` | `T` | Old record data (Supabase-compatible field name) |
| <a id="schema"></a> `schema` | `string` | Database schema |
| <a id="table"></a> `table` | `string` | Table name |
