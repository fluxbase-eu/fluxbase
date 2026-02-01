---
editUrl: false
next: false
prev: false
title: "useRealtime"
---

> **useRealtime**(`options`): `object`

Hook to subscribe to realtime changes for a channel

NOTE: The callback and invalidateKey are stored in refs to prevent
subscription recreation on every render when inline functions/arrays are used.

## Parameters

| Parameter | Type |
| ------ | ------ |
| `options` | `UseRealtimeOptions` |

## Returns

`object`

| Name | Type | Default value |
| ------ | ------ | ------ |
| `channel` | `RealtimeChannel` \| `null` | `channelRef.current` |
