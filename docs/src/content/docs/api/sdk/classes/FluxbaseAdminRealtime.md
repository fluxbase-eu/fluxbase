---
editUrl: false
next: false
prev: false
title: "FluxbaseAdminRealtime"
---

Realtime Admin Manager

Provides methods for enabling and managing realtime subscriptions on database tables.
When enabled, changes to a table (INSERT, UPDATE, DELETE) are automatically broadcast
to WebSocket subscribers.

## Example

```typescript
const realtime = client.admin.realtime

// Enable realtime on a table
await realtime.enableRealtime('products')

// Enable with options
await realtime.enableRealtime('orders', {
  events: ['INSERT', 'UPDATE'],
  exclude: ['internal_notes', 'raw_data']
})

// List all realtime-enabled tables
const { tables } = await realtime.listTables()

// Check status of a specific table
const status = await realtime.getStatus('public', 'products')

// Disable realtime
await realtime.disableRealtime('public', 'products')
```

## Constructors

### new FluxbaseAdminRealtime()

> **new FluxbaseAdminRealtime**(`fetch`): [`FluxbaseAdminRealtime`](/api/sdk/classes/fluxbaseadminrealtime/)

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

[`FluxbaseAdminRealtime`](/api/sdk/classes/fluxbaseadminrealtime/)

## Methods

### disableRealtime()

> **disableRealtime**(`schema`, `table`): `Promise`\<`object`\>

Disable realtime on a table

Removes the realtime trigger from a table. Existing subscribers will stop
receiving updates for this table.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `schema` | `string` | Schema name |
| `table` | `string` | Table name |

#### Returns

`Promise`\<`object`\>

Promise resolving to success message

| Name | Type |
| ------ | ------ |
| `message` | `string` |
| `success` | `boolean` |

#### Example

```typescript
await client.admin.realtime.disableRealtime('public', 'products')
console.log('Realtime disabled')
```

***

### enableRealtime()

> **enableRealtime**(`table`, `options`?): `Promise`\<[`EnableRealtimeResponse`](/api/sdk/interfaces/enablerealtimeresponse/)\>

Enable realtime on a table

Creates the necessary database triggers to broadcast changes to WebSocket subscribers.
Also sets REPLICA IDENTITY FULL to include old values in UPDATE/DELETE events.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `table` | `string` | Table name to enable realtime on |
| `options`? | `object` | Optional configuration |
| `options.events`? | (`"DELETE"` \| `"INSERT"` \| `"UPDATE"`)[] | - |
| `options.exclude`? | `string`[] | - |
| `options.schema`? | `string` | - |

#### Returns

`Promise`\<[`EnableRealtimeResponse`](/api/sdk/interfaces/enablerealtimeresponse/)\>

Promise resolving to EnableRealtimeResponse

#### Example

```typescript
// Enable realtime on products table (all events)
await client.admin.realtime.enableRealtime('products')

// Enable on a specific schema
await client.admin.realtime.enableRealtime('orders', {
  schema: 'sales'
})

// Enable specific events only
await client.admin.realtime.enableRealtime('audit_log', {
  events: ['INSERT'] // Only broadcast inserts
})

// Exclude large columns from notifications
await client.admin.realtime.enableRealtime('posts', {
  exclude: ['content', 'raw_html'] // Skip these in payload
})
```

***

### getStatus()

> **getStatus**(`schema`, `table`): `Promise`\<[`RealtimeTableStatus`](/api/sdk/interfaces/realtimetablestatus/)\>

Get realtime status for a specific table

Returns the realtime configuration for a table, including whether it's enabled,
which events are tracked, and which columns are excluded.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `schema` | `string` | Schema name |
| `table` | `string` | Table name |

#### Returns

`Promise`\<[`RealtimeTableStatus`](/api/sdk/interfaces/realtimetablestatus/)\>

Promise resolving to RealtimeTableStatus

#### Example

```typescript
const status = await client.admin.realtime.getStatus('public', 'products')

if (status.realtime_enabled) {
  console.log('Events:', status.events.join(', '))
  console.log('Excluded:', status.excluded_columns?.join(', ') || 'none')
} else {
  console.log('Realtime not enabled')
}
```

***

### listTables()

> **listTables**(`options`?): `Promise`\<[`ListRealtimeTablesResponse`](/api/sdk/interfaces/listrealtimetablesresponse/)\>

List all realtime-enabled tables

Returns a list of all tables that have realtime enabled, along with their
configuration (events, excluded columns, etc.).

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options`? | `object` | Optional filter options |
| `options.includeDisabled`? | `boolean` | - |

#### Returns

`Promise`\<[`ListRealtimeTablesResponse`](/api/sdk/interfaces/listrealtimetablesresponse/)\>

Promise resolving to ListRealtimeTablesResponse

#### Example

```typescript
// List all enabled tables
const { tables, count } = await client.admin.realtime.listTables()
console.log(`${count} tables have realtime enabled`)

tables.forEach(t => {
  console.log(`${t.schema}.${t.table}: ${t.events.join(', ')}`)
})

// Include disabled tables
const all = await client.admin.realtime.listTables({ includeDisabled: true })
```

***

### updateConfig()

> **updateConfig**(`schema`, `table`, `config`): `Promise`\<`object`\>

Update realtime configuration for a table

Modifies the events or excluded columns for a realtime-enabled table
without recreating the trigger.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `schema` | `string` | Schema name |
| `table` | `string` | Table name |
| `config` | [`UpdateRealtimeConfigRequest`](/api/sdk/interfaces/updaterealtimeconfigrequest/) | New configuration |

#### Returns

`Promise`\<`object`\>

Promise resolving to success message

| Name | Type |
| ------ | ------ |
| `message` | `string` |
| `success` | `boolean` |

#### Example

```typescript
// Change which events are tracked
await client.admin.realtime.updateConfig('public', 'products', {
  events: ['INSERT', 'UPDATE'] // Stop tracking deletes
})

// Update excluded columns
await client.admin.realtime.updateConfig('public', 'posts', {
  exclude: ['raw_content', 'search_vector']
})

// Clear excluded columns
await client.admin.realtime.updateConfig('public', 'posts', {
  exclude: [] // Include all columns again
})
```
