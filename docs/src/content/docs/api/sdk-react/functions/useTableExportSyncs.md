---
editUrl: false
next: false
prev: false
title: "useTableExportSyncs"
---

> **useTableExportSyncs**(`knowledgeBaseId`, `options?`): [`UseTableExportSyncsReturn`](/api/sdk-react/interfaces/usetableexportsyncsreturn/)

Hook for listing table export sync configurations

## Parameters

| Parameter | Type |
| ------ | ------ |
| `knowledgeBaseId` | `string` |
| `options` | [`UseTableExportSyncsOptions`](/api/sdk-react/interfaces/usetableexportsyncsoptions/) |

## Returns

[`UseTableExportSyncsReturn`](/api/sdk-react/interfaces/usetableexportsyncsreturn/)

## Example

```tsx
function SyncConfigsList({ kbId }: { kbId: string }) {
  const { configs, isLoading, error } = useTableExportSyncs(kbId)

  if (isLoading) return <div>Loading...</div>

  return (
    <ul>
      {configs.map(config => (
        <li key={config.id}>
          {config.schema_name}.{config.table_name} ({config.sync_mode})
        </li>
      ))}
    </ul>
  )
}
```
