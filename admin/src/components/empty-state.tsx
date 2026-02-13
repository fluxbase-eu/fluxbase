import { LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  actions?: Array<{
    label: string
    onClick: () => void
    variant?: 'default' | 'outline' | 'ghost'
    icon?: React.ReactNode
  }>
  templates?: Array<{
    label: string
    onClick: () => void
  }>
  className?: string
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  actions,
  templates,
  className,
}: EmptyStateProps) {
  return (
    <Card className={cn('border-dashed', className)}>
      <CardContent className='p-12 text-center'>
        <div className='mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted'>
          <Icon className='h-8 w-8 text-muted-foreground' />
        </div>
        <h3 className='mb-2 text-lg font-semibold'>{title}</h3>
        <p className='text-muted-foreground mb-6 text-sm'>{description}</p>

        {templates && templates.length > 0 && (
          <div className='mb-6'>
            <p className='text-muted-foreground mb-3 text-sm font-medium'>
              Quick Templates:
            </p>
            <div className='flex flex-wrap gap-2 justify-center'>
              {templates.map((template, index) => (
                <Badge
                  key={index}
                  variant='outline'
                  className='cursor-pointer hover:bg-accent'
                  onClick={template.onClick}
                >
                  {template.label}
                </Badge>
              ))}
            </div>
          </div>
        )}

        {actions && actions.length > 0 && (
          <div className='flex flex-col gap-3 sm:flex-row sm:justify-center'>
            {actions.map((action, index) => (
              <Button
                key={index}
                variant={action.variant || 'default'}
                onClick={action.onClick}
                className={cn('gap-2', action.variant === 'ghost' && 'border')}
              >
                {action.icon}
                {action.label}
              </Button>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
