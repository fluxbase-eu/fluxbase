---
editUrl: false
next: false
prev: false
title: "EnableRealtimeRequest"
---

Request to enable realtime on a table

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="events"></a> `events?` | (`"DELETE"` \| `"INSERT"` \| `"UPDATE"`)[] | Events to track (default: ['INSERT', 'UPDATE', 'DELETE']) |
| <a id="exclude"></a> `exclude?` | `string`[] | Columns to exclude from notifications |
| <a id="schema"></a> `schema` | `string` | Schema name (default: 'public') |
| <a id="table"></a> `table` | `string` | Table name |
