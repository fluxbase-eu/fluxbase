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
| `events?` | (`"DELETE"` \| `"INSERT"` \| `"UPDATE"`)[] | Events to track (default: ['INSERT', 'UPDATE', 'DELETE']) |
| `exclude?` | `string`[] | Columns to exclude from notifications |
| `schema` | `string` | Schema name (default: 'public') |
| `table` | `string` | Table name |
