---
editUrl: false
next: false
prev: false
title: "FluxbaseAdminMigrations"
---

Admin Migrations manager for database migration operations
Provides create, update, delete, apply, rollback, and smart sync operations

## Constructors

### Constructor

> **new FluxbaseAdminMigrations**(`fetch`): `FluxbaseAdminMigrations`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseAdminMigrations`

## Methods

### apply()

> **apply**(`name`, `namespace`): `Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Apply a specific migration

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `name` | `string` | `undefined` | Migration name |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |

#### Returns

`Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with result message

#### Example

```typescript
const { data, error } = await client.admin.migrations.apply('001_create_users', 'myapp')
if (data) {
  console.log(data.message) // "Migration applied successfully"
}
```

***

### applyPending()

> **applyPending**(`namespace`): `Promise`\<\{ `data`: \{ `applied`: `string`[]; `failed`: `string`[]; `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Apply all pending migrations in order

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |

#### Returns

`Promise`\<\{ `data`: \{ `applied`: `string`[]; `failed`: `string`[]; `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with applied/failed counts

#### Example

```typescript
const { data, error } = await client.admin.migrations.applyPending('myapp')
if (data) {
  console.log(`Applied: ${data.applied.length}, Failed: ${data.failed.length}`)
}
```

***

### create()

> **create**(`request`): `Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/) \| `null`; `error`: `Error` \| `null`; \}\>

Create a new migration

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`CreateMigrationRequest`](/api/sdk/interfaces/createmigrationrequest/) | Migration configuration |

#### Returns

`Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with created migration

#### Example

```typescript
const { data, error } = await client.admin.migrations.create({
  namespace: 'myapp',
  name: '001_create_users',
  up_sql: 'CREATE TABLE app.users (id UUID PRIMARY KEY, email TEXT)',
  down_sql: 'DROP TABLE app.users',
  description: 'Create users table'
})
```

***

### delete()

> **delete**(`name`, `namespace`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete a migration (only if status is pending)

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `name` | `string` | `undefined` | Migration name |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.migrations.delete('001_create_users', 'myapp')
```

***

### get()

> **get**(`name`, `namespace`): `Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/) \| `null`; `error`: `Error` \| `null`; \}\>

Get details of a specific migration

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `name` | `string` | `undefined` | Migration name |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |

#### Returns

`Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with migration details

#### Example

```typescript
const { data, error } = await client.admin.migrations.get('001_create_users', 'myapp')
```

***

### getExecutions()

> **getExecutions**(`name`, `namespace`, `limit`): `Promise`\<\{ `data`: [`MigrationExecution`](/api/sdk/interfaces/migrationexecution/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Get execution history for a migration

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `name` | `string` | `undefined` | Migration name |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |
| `limit` | `number` | `50` | Maximum number of executions to return (default: 50, max: 100) |

#### Returns

`Promise`\<\{ `data`: [`MigrationExecution`](/api/sdk/interfaces/migrationexecution/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with execution records

#### Example

```typescript
const { data, error } = await client.admin.migrations.getExecutions(
  '001_create_users',
  'myapp',
  10
)
if (data) {
  data.forEach(exec => {
    console.log(`${exec.executed_at}: ${exec.action} - ${exec.status}`)
  })
}
```

***

### list()

> **list**(`namespace`, `status?`): `Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List migrations in a namespace

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |
| `status?` | `"pending"` \| `"failed"` \| `"applied"` \| `"rolled_back"` | `undefined` | Filter by status: 'pending', 'applied', 'failed', 'rolled_back' |

#### Returns

`Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with migrations array

#### Example

```typescript
// List all migrations
const { data, error } = await client.admin.migrations.list('myapp')

// List only pending migrations
const { data, error } = await client.admin.migrations.list('myapp', 'pending')
```

***

### register()

> **register**(`migration`): `object`

Register a migration locally for smart sync

Call this method to register migrations in your application code.
When you call sync(), only new or changed migrations will be sent to the server.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `migration` | [`CreateMigrationRequest`](/api/sdk/interfaces/createmigrationrequest/) | Migration definition |

#### Returns

`object`

tuple (always succeeds unless validation fails)

| Name | Type |
| ------ | ------ |
| `error` | `Error` \| `null` |

#### Example

```typescript
// In your app initialization
const { error: err1 } = client.admin.migrations.register({
  name: '001_create_users_table',
  namespace: 'myapp',
  up_sql: 'CREATE TABLE app.users (...)',
  down_sql: 'DROP TABLE app.users',
  description: 'Initial users table'
})

const { error: err2 } = client.admin.migrations.register({
  name: '002_add_posts_table',
  namespace: 'myapp',
  up_sql: 'CREATE TABLE app.posts (...)',
  down_sql: 'DROP TABLE app.posts'
})

// Sync all registered migrations
await client.admin.migrations.sync()
```

***

### rollback()

> **rollback**(`name`, `namespace`): `Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Rollback a specific migration

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `name` | `string` | `undefined` | Migration name |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |

#### Returns

`Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with result message

#### Example

```typescript
const { data, error } = await client.admin.migrations.rollback('001_create_users', 'myapp')
```

***

### sync()

> **sync**(`options`): `Promise`\<\{ `data`: [`SyncMigrationsResult`](/api/sdk/interfaces/syncmigrationsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Smart sync all registered migrations

Automatically determines which migrations need to be created or updated by:
1. Fetching existing migrations from the server
2. Comparing content hashes to detect changes
3. Only sending new or changed migrations

After successful sync, can optionally auto-apply new migrations and refresh
the server's schema cache.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options` | `Partial`\<[`SyncMigrationsOptions`](/api/sdk/interfaces/syncmigrationsoptions/)\> | Sync options |

#### Returns

`Promise`\<\{ `data`: [`SyncMigrationsResult`](/api/sdk/interfaces/syncmigrationsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with sync results

#### Example

```typescript
// Basic sync (idempotent - safe to call on every app startup)
const { data, error } = await client.admin.migrations.sync()
if (data) {
  console.log(`Created: ${data.summary.created}, Updated: ${data.summary.updated}`)
}

// Sync with auto-apply (applies new migrations automatically)
const { data, error } = await client.admin.migrations.sync({
  auto_apply: true
})

// Dry run to preview changes without applying
const { data, error } = await client.admin.migrations.sync({
  dry_run: true
})
```

***

### update()

> **update**(`name`, `updates`, `namespace`): `Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/) \| `null`; `error`: `Error` \| `null`; \}\>

Update a migration (only if status is pending)

#### Parameters

| Parameter | Type | Default value | Description |
| ------ | ------ | ------ | ------ |
| `name` | `string` | `undefined` | Migration name |
| `updates` | [`UpdateMigrationRequest`](/api/sdk/interfaces/updatemigrationrequest/) | `undefined` | Fields to update |
| `namespace` | `string` | `"default"` | Migration namespace (default: 'default') |

#### Returns

`Promise`\<\{ `data`: [`Migration`](/api/sdk/interfaces/migration/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated migration

#### Example

```typescript
const { data, error } = await client.admin.migrations.update(
  '001_create_users',
  { description: 'Updated description' },
  'myapp'
)
```
