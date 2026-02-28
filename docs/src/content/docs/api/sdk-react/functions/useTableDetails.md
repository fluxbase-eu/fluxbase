---
editUrl: false
next: false
prev: false
title: "useTableDetails"
---

> **useTableDetails**(`options`): [`UseTableDetailsReturn`](/api/sdk-react/interfaces/usetabledetailsreturn/)

Hook for fetching detailed table information including columns

## Parameters

| Parameter | Type |
| ------ | ------ |
| `options` | [`UseTableDetailsOptions`](/api/sdk-react/interfaces/usetabledetailsoptions/) |

## Returns

[`UseTableDetailsReturn`](/api/sdk-react/interfaces/usetabledetailsreturn/)

## Example

```tsx
function TableColumnsList({ schema, table }: { schema: string; table: string }) {
  const { data, isLoading, error } = useTableDetails({ schema, table })

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return (
    <ul>
      {data?.columns.map(col => (
        <li key={col.name}>
          {col.name} ({col.data_type})
          {col.is_primary_key && ' 🔑'}
        </li>
      ))}
    </ul>
  )
}
```
