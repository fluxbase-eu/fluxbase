---
editUrl: false
next: false
prev: false
title: "FluxbaseAdminRPC"
---

Admin RPC manager for managing RPC procedures
Provides sync, CRUD, and execution monitoring operations

## Constructors

### Constructor

> **new FluxbaseAdminRPC**(`fetch`): `FluxbaseAdminRPC`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseAdminRPC`

## Methods

### cancelExecution()

> **cancelExecution**(`executionId`): `Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Cancel a running execution

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `executionId` | `string` | Execution ID |

#### Returns

`Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated execution

#### Example

```typescript
const { data, error } = await client.admin.rpc.cancelExecution('execution-uuid')
```

***

### delete()

> **delete**(`namespace`, `name`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete an RPC procedure

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace` | `string` | Procedure namespace |
| `name` | `string` | Procedure name |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.rpc.delete('default', 'get-user-orders')
```

***

### get()

> **get**(`namespace`, `name`): `Promise`\<\{ `data`: [`RPCProcedure`](/api/sdk/interfaces/rpcprocedure/) \| `null`; `error`: `Error` \| `null`; \}\>

Get details of a specific RPC procedure

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace` | `string` | Procedure namespace |
| `name` | `string` | Procedure name |

#### Returns

`Promise`\<\{ `data`: [`RPCProcedure`](/api/sdk/interfaces/rpcprocedure/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with procedure details

#### Example

```typescript
const { data, error } = await client.admin.rpc.get('default', 'get-user-orders')
if (data) {
  console.log('Procedure:', data.name)
  console.log('SQL:', data.sql_query)
}
```

***

### getExecution()

> **getExecution**(`executionId`): `Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Get details of a specific execution

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `executionId` | `string` | Execution ID |

#### Returns

`Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with execution details

#### Example

```typescript
const { data, error } = await client.admin.rpc.getExecution('execution-uuid')
if (data) {
  console.log('Status:', data.status)
  console.log('Duration:', data.duration_ms, 'ms')
}
```

***

### getExecutionLogs()

> **getExecutionLogs**(`executionId`, `afterLine?`): `Promise`\<\{ `data`: [`ExecutionLog`](/api/sdk/interfaces/executionlog/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Get execution logs for a specific execution

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `executionId` | `string` | Execution ID |
| `afterLine?` | `number` | Optional line number to get logs after (for polling) |

#### Returns

`Promise`\<\{ `data`: [`ExecutionLog`](/api/sdk/interfaces/executionlog/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with execution logs

#### Example

```typescript
const { data, error } = await client.admin.rpc.getExecutionLogs('execution-uuid')
if (data) {
  for (const log of data) {
    console.log(`[${log.level}] ${log.message}`)
  }
}
```

***

### list()

> **list**(`namespace?`): `Promise`\<\{ `data`: [`RPCProcedureSummary`](/api/sdk/interfaces/rpcproceduresummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List all RPC procedures (admin view)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace?` | `string` | Optional namespace filter |

#### Returns

`Promise`\<\{ `data`: [`RPCProcedureSummary`](/api/sdk/interfaces/rpcproceduresummary/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of procedure summaries

#### Example

```typescript
const { data, error } = await client.admin.rpc.list()
if (data) {
  console.log('Procedures:', data.map(p => p.name))
}
```

***

### listExecutions()

> **listExecutions**(`filters?`): `Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List RPC executions with optional filters

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `filters?` | [`RPCExecutionFilters`](/api/sdk/interfaces/rpcexecutionfilters/) | Optional filters for namespace, procedure, status, user |

#### Returns

`Promise`\<\{ `data`: [`RPCExecution`](/api/sdk/interfaces/rpcexecution/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of executions

#### Example

```typescript
// List all executions
const { data, error } = await client.admin.rpc.listExecutions()

// List failed executions for a specific procedure
const { data, error } = await client.admin.rpc.listExecutions({
  namespace: 'default',
  procedure: 'get-user-orders',
  status: 'failed',
})
```

***

### listNamespaces()

> **listNamespaces**(): `Promise`\<\{ `data`: `string`[] \| `null`; `error`: `Error` \| `null`; \}\>

List all namespaces

#### Returns

`Promise`\<\{ `data`: `string`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of namespace names

#### Example

```typescript
const { data, error } = await client.admin.rpc.listNamespaces()
if (data) {
  console.log('Namespaces:', data)
}
```

***

### sync()

> **sync**(`options?`): `Promise`\<\{ `data`: [`SyncRPCResult`](/api/sdk/interfaces/syncrpcresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Sync RPC procedures from filesystem or API payload

Can sync from:
1. Filesystem (if no procedures provided) - loads from configured procedures directory
2. API payload (if procedures array provided) - syncs provided procedure specifications

Requires service_role or admin authentication.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options?` | [`SyncRPCOptions`](/api/sdk/interfaces/syncrpcoptions/) | Sync options including namespace and optional procedures array |

#### Returns

`Promise`\<\{ `data`: [`SyncRPCResult`](/api/sdk/interfaces/syncrpcresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with sync results

#### Example

```typescript
// Sync from filesystem
const { data, error } = await client.admin.rpc.sync()

// Sync with provided procedure code
const { data, error } = await client.admin.rpc.sync({
  namespace: 'default',
  procedures: [{
    name: 'get-user-orders',
    code: myProcedureSQL,
  }],
  options: {
    delete_missing: false, // Don't remove procedures not in this sync
    dry_run: false,        // Preview changes without applying
  }
})

if (data) {
  console.log(`Synced: ${data.summary.created} created, ${data.summary.updated} updated`)
}
```

***

### toggle()

> **toggle**(`namespace`, `name`, `enabled`): `Promise`\<\{ `data`: [`RPCProcedure`](/api/sdk/interfaces/rpcprocedure/) \| `null`; `error`: `Error` \| `null`; \}\>

Enable or disable an RPC procedure

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace` | `string` | Procedure namespace |
| `name` | `string` | Procedure name |
| `enabled` | `boolean` | Whether to enable or disable |

#### Returns

`Promise`\<\{ `data`: [`RPCProcedure`](/api/sdk/interfaces/rpcprocedure/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated procedure

#### Example

```typescript
const { data, error } = await client.admin.rpc.toggle('default', 'get-user-orders', true)
```

***

### update()

> **update**(`namespace`, `name`, `updates`): `Promise`\<\{ `data`: [`RPCProcedure`](/api/sdk/interfaces/rpcprocedure/) \| `null`; `error`: `Error` \| `null`; \}\>

Update an RPC procedure

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace` | `string` | Procedure namespace |
| `name` | `string` | Procedure name |
| `updates` | [`UpdateRPCProcedureRequest`](/api/sdk/interfaces/updaterpcprocedurerequest/) | Fields to update |

#### Returns

`Promise`\<\{ `data`: [`RPCProcedure`](/api/sdk/interfaces/rpcprocedure/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated procedure

#### Example

```typescript
const { data, error } = await client.admin.rpc.update('default', 'get-user-orders', {
  enabled: false,
  max_execution_time_seconds: 60,
})
```
