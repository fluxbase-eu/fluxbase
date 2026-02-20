import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface ExportFormat {
  id: string
  name: string
  icon: string
}

const EXPORT_FORMATS: ExportFormat[] = [
  { id: 'csv', name: 'CSV', icon: 'FileText' },
  { id: 'json', name: 'JSON', icon: '{ }' },
]

interface DataExportDialogProps {
  open: boolean
  onClose: () => void
}

export function DataExportDialog({ open, onClose }: DataExportDialogProps) {
  const [_exportFormat,] = useState('csv')
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set())

  const handleToggle = (itemId: string) => {
    setSelectedItems((prev) => {
      const newSet = new Set(prev)
      if (newSet.has(itemId)) {
        newSet.delete(itemId)
      } else {
        newSet.add(itemId)
      }
      return newSet
    })
  }

  const handleExport = () => {
    onClose()
    // In real implementation, this would call the export API
    // Exporting data in ${exportFormat} format with ${selectedItems.size} items
  }

  return (
    <Dialog open={open}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Export Data</DialogTitle>
          <DialogDescription>
            Select the data you want to export and choose your preferred format.
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          {EXPORT_FORMATS.map((format) => (
            <Button
              key={format.id}
              variant={selectedItems.has(format.id) ? 'default' : 'ghost'}
              className='w-full justify-start'
              onClick={() => handleToggle(format.id)}
            >
              <div className='flex items-center gap-2'>
                <Checkbox checked={selectedItems.has(format.id)} />
                <span>{format.name}</span>
                <span className='text-xs text-muted-foreground ml-2'>{format.icon}</span>
              </div>
            </Button>
          ))}
        </div>

        <DialogFooter className='mt-4'>
          <Button onClick={onClose} variant='outline'>Cancel</Button>
          <Button onClick={handleExport} variant='default' disabled={selectedItems.size === 0}>
            Export Selected
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
