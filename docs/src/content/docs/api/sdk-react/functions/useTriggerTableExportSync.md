---
editUrl: false
next: false
prev: false
title: "useTriggerTableExportSync"
---

> **useTriggerTableExportSync**(`knowledgeBaseId`): [`UseTriggerTableExportSyncReturn`](/api/sdk-react/interfaces/usetriggertableexportsyncreturn/)

Hook for manually triggering a table export sync

## Parameters

| Parameter | Type |
| ------ | ------ |
| `knowledgeBaseId` | `string` |

## Returns

[`UseTriggerTableExportSyncReturn`](/api/sdk-react/interfaces/usetriggertableexportsyncreturn/)

## Example

```tsx
function TriggerSyncButton({ kbId, syncId }: Props) {
  const { triggerSync, isLoading, error } = useTriggerTableExportSync(kbId)

  const handleTrigger = async () => {
    const result = await triggerSync(syncId)
    if (result) {
      console.log('Sync completed:', result.document_id)
    }
  }

  return (
    <button onClick={handleTrigger} disabled={isLoading}>
      {isLoading ? 'Syncing...' : 'Sync Now'}
    </button>
  )
}
```
