---
editUrl: false
next: false
prev: false
title: "FluxbaseAdminFunctions"
---

Admin Functions manager for managing edge functions
Provides create, update, delete, and bulk sync operations

## Constructors

### Constructor

> **new FluxbaseAdminFunctions**(`fetch`): `FluxbaseAdminFunctions`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseAdminFunctions`

## Methods

### create()

> **create**(`request`): `Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/) \| `null`; `error`: `Error` \| `null`; \}\>

Create a new edge function

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`CreateFunctionRequest`](/api/sdk/interfaces/createfunctionrequest/) | Function configuration and code |

#### Returns

`Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with created function metadata

#### Example

```typescript
const { data, error } = await client.admin.functions.create({
  name: 'my-function',
  code: 'export default async function handler(req) { return { hello: "world" } }',
  enabled: true
})
```

***

### delete()

> **delete**(`name`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Delete an edge function

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Function name |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple

#### Example

```typescript
const { data, error } = await client.admin.functions.delete('my-function')
```

***

### get()

> **get**(`name`): `Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/) \| `null`; `error`: `Error` \| `null`; \}\>

Get details of a specific edge function

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Function name |

#### Returns

`Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with function metadata

#### Example

```typescript
const { data, error } = await client.admin.functions.get('my-function')
if (data) {
  console.log('Function version:', data.version)
}
```

***

### getExecutions()

> **getExecutions**(`name`, `limit?`): `Promise`\<\{ `data`: [`EdgeFunctionExecution`](/api/sdk/interfaces/edgefunctionexecution/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Get execution history for an edge function

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Function name |
| `limit?` | `number` | Maximum number of executions to return (optional) |

#### Returns

`Promise`\<\{ `data`: [`EdgeFunctionExecution`](/api/sdk/interfaces/edgefunctionexecution/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with execution records

#### Example

```typescript
const { data, error } = await client.admin.functions.getExecutions('my-function', 10)
if (data) {
  data.forEach(exec => {
    console.log(`${exec.executed_at}: ${exec.status} (${exec.duration_ms}ms)`)
  })
}
```

***

### list()

> **list**(`namespace?`): `Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List all edge functions (admin view)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `namespace?` | `string` | Optional namespace filter (if not provided, lists all public functions) |

#### Returns

`Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/)[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of functions

#### Example

```typescript
// List all public functions
const { data, error } = await client.admin.functions.list()

// List functions in a specific namespace
const { data, error } = await client.admin.functions.list('my-namespace')
if (data) {
  console.log('Functions:', data.map(f => f.name))
}
```

***

### listNamespaces()

> **listNamespaces**(): `Promise`\<\{ `data`: `string`[] \| `null`; `error`: `Error` \| `null`; \}\>

List all namespaces that have edge functions

#### Returns

`Promise`\<\{ `data`: `string`[] \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with array of namespace strings

#### Example

```typescript
const { data, error } = await client.admin.functions.listNamespaces()
if (data) {
  console.log('Available namespaces:', data)
}
```

***

### sync()

> **sync**(`options`): `Promise`\<\{ `data`: [`SyncFunctionsResult`](/api/sdk/interfaces/syncfunctionsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Sync multiple functions to a namespace

Bulk create/update/delete functions in a specific namespace. This is useful for
deploying functions from your application to Fluxbase in Kubernetes or other
container environments.

Requires service_role or admin authentication.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options` | [`SyncFunctionsOptions`](/api/sdk/interfaces/syncfunctionsoptions/) | Sync configuration including namespace, functions, and options |

#### Returns

`Promise`\<\{ `data`: [`SyncFunctionsResult`](/api/sdk/interfaces/syncfunctionsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with sync results

#### Example

```typescript
// Sync functions to "payment-service" namespace
const { data, error } = await client.admin.functions.sync({
  namespace: 'payment-service',
  functions: [
    {
      name: 'process-payment',
      code: 'export default async function handler(req) { ... }',
      enabled: true,
      allow_net: true
    },
    {
      name: 'refund-payment',
      code: 'export default async function handler(req) { ... }',
      enabled: true
    }
  ],
  options: {
    delete_missing: true  // Remove functions not in this list
  }
})

if (data) {
  console.log(`Synced: ${data.summary.created} created, ${data.summary.updated} updated`)
}

// Dry run to preview changes
const { data, error } = await client.admin.functions.sync({
  namespace: 'myapp',
  functions: [...],
  options: { dry_run: true }
})
```

***

### syncWithBundling()

> **syncWithBundling**(`options`, `bundleOptions?`): `Promise`\<\{ `data`: [`SyncFunctionsResult`](/api/sdk/interfaces/syncfunctionsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Sync edge functions with automatic client-side bundling

This is a convenience method that bundles all function code using esbuild
before sending to the server. Requires esbuild as a peer dependency.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options` | [`SyncFunctionsOptions`](/api/sdk/interfaces/syncfunctionsoptions/) | Sync options including namespace and functions array |
| `bundleOptions?` | `Partial`\<[`BundleOptions`](/api/sdk/interfaces/bundleoptions/)\> | Optional bundling configuration |

#### Returns

`Promise`\<\{ `data`: [`SyncFunctionsResult`](/api/sdk/interfaces/syncfunctionsresult/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with sync results

#### Example

```typescript
const { data, error } = await client.admin.functions.syncWithBundling({
  namespace: 'default',
  functions: [
    { name: 'hello', code: helloCode },
    { name: 'goodbye', code: goodbyeCode },
  ],
  options: { delete_missing: true }
})
```

***

### update()

> **update**(`name`, `updates`): `Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/) \| `null`; `error`: `Error` \| `null`; \}\>

Update an existing edge function

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Function name |
| `updates` | [`UpdateFunctionRequest`](/api/sdk/interfaces/updatefunctionrequest/) | Fields to update |

#### Returns

`Promise`\<\{ `data`: [`EdgeFunction`](/api/sdk/interfaces/edgefunction/) \| `null`; `error`: `Error` \| `null`; \}\>

Promise resolving to { data, error } tuple with updated function metadata

#### Example

```typescript
const { data, error } = await client.admin.functions.update('my-function', {
  enabled: false,
  description: 'Updated description'
})
```

***

### bundleCode()

> `static` **bundleCode**(`options`): `Promise`\<[`BundleResult`](/api/sdk/interfaces/bundleresult/)\>

Bundle function code using esbuild (client-side)

Transforms and bundles TypeScript/JavaScript code into a single file
that can be executed by the Fluxbase edge functions runtime.

Requires esbuild as a peer dependency.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options` | [`BundleOptions`](/api/sdk/interfaces/bundleoptions/) | Bundle options including source code |

#### Returns

`Promise`\<[`BundleResult`](/api/sdk/interfaces/bundleresult/)\>

Promise resolving to bundled code

#### Throws

Error if esbuild is not available

#### Example

```typescript
const bundled = await FluxbaseAdminFunctions.bundleCode({
  code: `
    import { helper } from './utils'
    export default async function handler(req) {
      return helper(req.body)
    }
  `,
  minify: true,
})

// Use bundled code in sync
await client.admin.functions.sync({
  namespace: 'default',
  functions: [{
    name: 'my-function',
    code: bundled.code,
    is_pre_bundled: true,
  }]
})
```
