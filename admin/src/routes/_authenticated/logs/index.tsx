import { createFileRoute } from '@tanstack/react-router'
import { ScrollText } from 'lucide-react'
import { LogViewer } from '@/features/logs/components/log-viewer'

export const Route = createFileRoute('/_authenticated/logs/')({
  component: LogsPage,
})

function LogsPage() {
  return (
    <div className='flex h-full flex-col'>
      <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
        <div className='flex items-center gap-3'>
          <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
            <ScrollText className='text-primary h-5 w-5' />
          </div>
          <div>
            <h1 className='text-xl font-semibold'>Log Stream</h1>
            <p className='text-muted-foreground text-sm'>
              Real-time application logs
            </p>
          </div>
        </div>
      </div>

      <div className='min-h-0 flex-1 p-6'>
        <LogViewer />
      </div>
    </div>
  )
}
