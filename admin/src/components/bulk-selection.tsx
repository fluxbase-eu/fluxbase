import { useState } from 'react'
import { FileText, Trash2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { bulkOperationsApi, dataExportApi } from '@/lib/api'
import { Button } from '@/components/ui/button'

interface BulkSelectionProps {
  items: Array<{
    id: string
    name: string
    type: string
  }>
  table: string
  onRefresh?: () => void
}

interface BulkAction {
  label: string
  icon: React.ReactNode
  onClick: () => void | Promise<void>
  variant?: 'default' | 'destructive'
  disabled?: boolean
}

export function BulkSelection({ items, table, onRefresh }: BulkSelectionProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [isProcessing, setIsProcessing] = useState(false)

  const selectedCount = selectedIds.size

  const handleSelectAll = () => {
    setSelectedIds(new Set(items.map((item) => item.id)))
  }

  const handleDeselectAll = () => {
    setSelectedIds(new Set())
  }

  const handleExport = async () => {
    if (selectedCount === 0) return

    setIsProcessing(true)
    try {
      const targets = Array.from(selectedIds)
      await dataExportApi.export(table, targets, 'json')
      toast.success(`Exported ${selectedCount} item(s)`)
    } catch (error) {
      toast.error(
        `Failed to export: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    } finally {
      setIsProcessing(false)
    }
  }

  const handleDelete = async () => {
    if (selectedCount === 0) return

    setIsProcessing(true)
    try {
      const targets = Array.from(selectedIds)
      const result = await bulkOperationsApi.delete(table, targets)
      toast.success(result.message || `Deleted ${selectedCount} item(s)`)
      setSelectedIds(new Set())
      onRefresh?.()
    } catch (error) {
      toast.error(
        `Failed to delete: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    } finally {
      setIsProcessing(false)
    }
  }

  const bulkActions: BulkAction[] = [
    {
      label: 'Export Selected',
      icon: isProcessing ? null : <FileText className='h-4 w-4' />,
      onClick: handleExport,
      disabled: isProcessing || selectedCount === 0,
    },
    {
      label: 'Delete Selected',
      icon: isProcessing ? null : <Trash2 className='h-4 w-4' />,
      onClick: handleDelete,
      variant: 'destructive',
      disabled: isProcessing || selectedCount === 0,
    },
  ]

  return (
    <div className='flex items-center justify-between border-b px-4 py-2'>
      <div className='flex items-center gap-2'>
        <Button
          variant='ghost'
          size='sm'
          onClick={handleSelectAll}
          disabled={selectedCount === items.length || isProcessing}
        >
          Select All ({selectedCount}/{items.length})
        </Button>

        <Button
          variant='default'
          size='sm'
          onClick={handleDeselectAll}
          disabled={selectedCount === 0 || isProcessing}
        >
          Deselect All
        </Button>
      </div>

      <div className='flex h-8 w-full flex-1 items-center gap-3 py-1'>
        {bulkActions.map((action) => (
          <Button
            key={action.label}
            variant={action.variant || 'default'}
            size='sm'
            onClick={action.onClick}
            disabled={action.disabled}
            className='flex-1 items-center gap-1'
          >
            {isProcessing && action.label.includes('Delete') ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : isProcessing && action.label.includes('Export') ? (
              <Loader2 className='h-4 w-4 animate-spin' />
            ) : null}
            {action.icon}
            <span className='ml-2'>{action.label}</span>
          </Button>
        ))}
      </div>
    </div>
  )
}
