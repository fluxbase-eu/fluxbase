---
editUrl: false
next: false
prev: false
title: "useCreateTableExportSync"
---

> **useCreateTableExportSync**(`knowledgeBaseId`): [`UseCreateTableExportSyncReturn`](/api/sdk-react/interfaces/usecreatetableexportsyncreturn/)

Hook for creating a table export sync configuration

## Parameters

| Parameter | Type |
| ------ | ------ |
| `knowledgeBaseId` | `string` |

## Returns

[`UseCreateTableExportSyncReturn`](/api/sdk-react/interfaces/usecreatetableexportsyncreturn/)

## Example

```tsx
function CreateSyncForm({ kbId }: { kbId: string }) {
  const { createSync, isLoading, error } = useCreateTableExportSync(kbId)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const config = await createSync({
      schema_name: 'public',
      table_name: 'users',
      columns: ['id', 'name', 'email'],
      sync_mode: 'automatic',
      sync_on_insert: true,
      sync_on_update: true,
    })
    if (config) {
      console.log('Created sync config:', config.id)
    }
  }

  return <form onSubmit={handleSubmit}>...</form>
}
```
