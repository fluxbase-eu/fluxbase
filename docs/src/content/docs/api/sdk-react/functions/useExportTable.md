---
editUrl: false
next: false
prev: false
title: "useExportTable"
---

> **useExportTable**(`knowledgeBaseId`): [`UseExportTableReturn`](/api/sdk-react/interfaces/useexporttablereturn/)

Hook for exporting a table to a knowledge base

## Parameters

| Parameter | Type |
| ------ | ------ |
| `knowledgeBaseId` | `string` |

## Returns

[`UseExportTableReturn`](/api/sdk-react/interfaces/useexporttablereturn/)

## Example

```tsx
function ExportTableButton({ kbId, schema, table }: Props) {
  const { exportTable, isLoading, error } = useExportTable(kbId)

  const handleExport = async () => {
    const result = await exportTable({
      schema,
      table,
      columns: ['id', 'name', 'email'],
      include_foreign_keys: true,
    })
    if (result) {
      console.log('Exported document:', result.document_id)
    }
  }

  return (
    <button onClick={handleExport} disabled={isLoading}>
      {isLoading ? 'Exporting...' : 'Export Table'}
    </button>
  )
}
```
