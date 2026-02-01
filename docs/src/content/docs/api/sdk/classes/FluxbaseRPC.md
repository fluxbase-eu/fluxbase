---
editUrl: false
next: false
prev: false
title: "FluxbaseRPC"
---

FluxbaseRPC provides methods for invoking RPC procedures

## Example

```typescript
// Invoke a procedure synchronously
const { data, error } = await fluxbase.rpc.invoke('get-user-orders', {
  user_id: '123',
  limit: 10
});

// Invoke asynchronously
const { data: asyncResult } = await fluxbase.rpc.invoke('long-running-report', {
  start_date: '2024-01-01'
}, { async: true });

// Poll for status
const { data: status } = await fluxbase.rpc.getStatus(asyncResult.execution_id);
```

## Constructors

### Constructor

> **new FluxbaseRPC**(`fetch`): `FluxbaseRPC`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | `RPCFetch` |

#### Returns

`FluxbaseRPC`

## Methods

### getLogs()

> **getLogs**(`executionId`, `afterLine?`): `Promise`\<\{ `data`: [`ExecutionLog`](/api/sdk/interfaces/executionlog/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Get execution logs (for debugging and monitoring)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `executionId` | `string` | The execution ID |
| `afterLine?` | `number` | Optional line number to get logs after (for polling) |

#### Returns

`Promise`\<\{ `data`: [`ExecutionLog`](/api/sdk/interfaces/executionlog/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with execution logs

#### Example

```typescript
const { data: logs } = await fluxbase.rpc.getLogs('execution-uuid');
for (const log of logs) {
  console.log(`[${log.level}] ${log.message}`);
}
```

***

### getStatus()

> **getStatus**(`executionId`): `Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Get execution status (for async invocations or checking history)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `executionId` | `string` | The execution ID returned from async invoke |

#### Returns

`Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with execution details

#### Example

```typescript
const { data, error } = await fluxbase.rpc.getStatus('execution-uuid');
if (data.status === 'completed') {
  console.log('Result:', data.result);
} else if (data.status === 'running') {
  console.log('Still running...');
}
```

***

### invoke()

> **invoke**\<`T`\>(`name`, `params?`, `options?`): `Promise`\<\{ `data`: [`RPCInvokeResponse`](/api/sdk/interfaces/rpcinvokeresponse/)\<`T`\> \| `null`; `error`: `Error` \| `null`; \}\>

Invoke an RPC procedure

#### Type Parameters

| Type Parameter | Default type |
| ------ | ------ |
| `T` | `unknown` |

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Procedure name |
| `params?` | `Record`\<`string`, `unknown`\> | Optional parameters to pass to the procedure |
| `options?` | `RPCInvokeOptions` | Optional invocation options |

#### Returns

`Promise`\<\{ `data`: [`RPCInvokeResponse`](/api/sdk/interfaces/rpcinvokeresponse/)\<`T`\> \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with invocation response

#### Example

```typescript
// Synchronous invocation
const { data, error } = await fluxbase.rpc.invoke('get-user-orders', {
  user_id: '123',
  limit: 10
});
console.log(data.result); // Query results

// Asynchronous invocation
const { data: asyncData } = await fluxbase.rpc.invoke('generate-report', {
  year: 2024
}, { async: true });
console.log(asyncData.execution_id); // Use to poll status
```

***

### list()

> **list**(`namespace?`): `Promise`\<\{ `data`: [`RPCProcedureSummary`](/api/sdk/interfaces/rpcproceduresummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List available RPC procedures (public, enabled)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace?` | `string` | Optional namespace filter |

#### Returns

`Promise`\<\{ `data`: [`RPCProcedureSummary`](/api/sdk/interfaces/rpcproceduresummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of procedure summaries

***

### waitForCompletion()

> **waitForCompletion**(`executionId`, `options?`): `Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Poll for execution completion with exponential backoff

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `executionId` | `string` | The execution ID to poll |
| `options?` | \{ `initialIntervalMs?`: `number`; `maxIntervalMs?`: `number`; `maxWaitMs?`: `number`; `onProgress?`: (`execution`) => `void`; \} | Polling options |
| `options.initialIntervalMs?` | `number` | Initial polling interval in milliseconds (default: 500) |
| `options.maxIntervalMs?` | `number` | Maximum polling interval in milliseconds (default: 5000) |
| `options.maxWaitMs?` | `number` | Maximum time to wait in milliseconds (default: 30000) |
| `options.onProgress?` | (`execution`) => `void` | Callback for progress updates |

#### Returns

`Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to final execution state

#### Example

```typescript
const { data: result } = await fluxbase.rpc.invoke('long-task', {}, { async: true });
const { data: final } = await fluxbase.rpc.waitForCompletion(result.execution_id, {
  maxWaitMs: 60000, // Wait up to 1 minute
  onProgress: (exec) => console.log(`Status: ${exec.status}`)
});
console.log('Final result:', final.result);
```
