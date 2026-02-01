---
editUrl: false
next: false
prev: false
title: "FluxbaseStorage"
---

## Constructors

### Constructor

> **new FluxbaseStorage**(`fetch`): `FluxbaseStorage`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`FluxbaseStorage`

## Methods

### createBucket()

> **createBucket**(`bucketName`): `Promise`\<\{ `data`: \{ `name`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Create a new bucket

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `bucketName` | `string` | The name of the bucket to create |

#### Returns

`Promise`\<\{ `data`: \{ `name`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

***

### deleteBucket()

> **deleteBucket**(`bucketName`): `Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Delete a bucket

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `bucketName` | `string` | The name of the bucket to delete |

#### Returns

`Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

***

### emptyBucket()

> **emptyBucket**(`bucketName`): `Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Empty a bucket (delete all files)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `bucketName` | `string` | The name of the bucket to empty |

#### Returns

`Promise`\<\{ `data`: \{ `message`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

***

### from()

> **from**(`bucketName`): [`StorageBucket`](/api/sdk/classes/storagebucket/)

Get a reference to a storage bucket

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `bucketName` | `string` | The name of the bucket |

#### Returns

[`StorageBucket`](/api/sdk/classes/storagebucket/)

***

### getBucket()

> **getBucket**(`bucketName`): `Promise`\<\{ `data`: `Bucket` \| `null`; `error`: `Error` \| `null`; \}\>

Get bucket details

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `bucketName` | `string` | The name of the bucket |

#### Returns

`Promise`\<\{ `data`: `Bucket` \| `null`; `error`: `Error` \| `null`; \}\>

***

### listBuckets()

> **listBuckets**(): `Promise`\<\{ `data`: `object`[] \| `null`; `error`: `Error` \| `null`; \}\>

List all buckets

#### Returns

`Promise`\<\{ `data`: `object`[] \| `null`; `error`: `Error` \| `null`; \}\>

***

### updateBucketSettings()

> **updateBucketSettings**(`bucketName`, `settings`): `Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>

Update bucket settings (RLS - requires admin or service key)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `bucketName` | `string` | The name of the bucket |
| `settings` | `BucketSettings` | Bucket settings to update |

#### Returns

`Promise`\<\{ `data`: `null`; `error`: `Error` \| `null`; \}\>
