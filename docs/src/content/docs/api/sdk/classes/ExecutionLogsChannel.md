---
editUrl: false
next: false
prev: false
title: "ExecutionLogsChannel"
---

Specialized channel for execution log subscriptions
Provides a cleaner API than the generic RealtimeChannel

## Constructors

### Constructor

> **new ExecutionLogsChannel**(`url`, `executionId`, `type`, `token`, `tokenRefreshCallback`): `ExecutionLogsChannel`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `url` | `string` |
| `executionId` | `string` |
| `type` | [`ExecutionType`](/api/sdk/type-aliases/executiontype/) |
| `token` | `string` \| `null` |
| `tokenRefreshCallback` | (() => `Promise`\<`string` \| `null`\>) \| `null` |

#### Returns

`ExecutionLogsChannel`

## Methods

### onLog()

> **onLog**(`callback`): `this`

Register a callback for log events

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `callback` | [`ExecutionLogCallback`](/api/sdk/type-aliases/executionlogcallback/) | Function to call when log entries are received |

#### Returns

`this`

This channel for chaining

#### Example

```typescript
channel.onLog((log) => {
  console.log(`[${log.level}] Line ${log.line_number}: ${log.message}`)
})
```

***

### subscribe()

> **subscribe**(`callback?`, `timeout?`): `this`

Subscribe to execution logs

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `callback?` | [`RealtimeSubscribeStatusCallback`](/api/sdk/type-aliases/realtimesubscribestatuscallback/) | Optional status callback (Supabase-compatible). Fires SUBSCRIBED when the server acknowledges the subscribe_logs message, TIMED_OUT when the acknowledgment does not arrive within the timeout, and CHANNEL_ERROR when the connection fails before the acknowledgment. |
| `timeout?` | `number` | - |

#### Returns

`this`

Promise that resolves when subscribed

#### Example

```typescript
await channel.subscribe()
```

***

### unsubscribe()

> **unsubscribe**(): `Promise`\<`"error"` \| `"ok"` \| `"timed out"`\>

Unsubscribe from execution logs

#### Returns

`Promise`\<`"error"` \| `"ok"` \| `"timed out"`\>

Promise resolving to status

#### Example

```typescript
await channel.unsubscribe()
```
