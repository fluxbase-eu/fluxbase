---
editUrl: false
next: false
prev: false
title: "RealtimeTableStatus"
---

Status of realtime for a table

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| `created_at?` | `string` | When realtime was enabled |
| `events` | `string`[] | Events being tracked |
| `excluded_columns?` | `string`[] | Columns excluded from notifications |
| `id?` | `number` | Registry ID |
| `realtime_enabled` | `boolean` | Whether realtime is enabled |
| `schema` | `string` | Schema name |
| `table` | `string` | Table name |
| `updated_at?` | `string` | When configuration was last updated |
