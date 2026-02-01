---
editUrl: false
next: false
prev: false
title: "useStorageSignedUrl"
---

> **useStorageSignedUrl**(`bucket`, `path`, `expiresIn?`): `UseQueryResult`\<`string` \| `null`, `Error`\>

Hook to create a signed URL

:::caution[Deprecated]
Use useStorageSignedUrlWithOptions for more control including transforms
:::

## Parameters

| Parameter | Type |
| ------ | ------ |
| `bucket` | `string` |
| `path` | `string` \| `null` |
| `expiresIn?` | `number` |

## Returns

`UseQueryResult`\<`string` \| `null`, `Error`\>
