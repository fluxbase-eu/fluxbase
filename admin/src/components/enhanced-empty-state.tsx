import { Button } from '@/components/ui/button'

interface EmptyStateProps {
  message: string
  title?: string
  actionLabel?: string
  action?: () => void
  icon?: React.ReactNode
}

export function EnhancedEmptyState({ message, title, actionLabel, action, icon }: EmptyStateProps) {
  return (
    <div className='flex flex-col items-center gap-4 py-8'>
      {icon && (
        <div className='bg-blue-100 rounded-lg p-4'>
          <div className='h-12 w-12 flex items-center justify-center'>
            {icon}
          </div>
        </div>
      )}
      <div className='flex flex-col gap-2 text-center'>
        {title && <div className='text-sm font-medium'>{title}</div>}
        <div className='text-xs text-muted-foreground'>{message}</div>
        {actionLabel && action && (
          <Button variant='ghost' size='sm' onClick={action}>
            {actionLabel}
          </Button>
        )}
      </div>
    </div>
  )
}
