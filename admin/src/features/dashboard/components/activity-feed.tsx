import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, XCircle, AlertTriangle, Info, Clock } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'

type ActivityType = 'success' | 'error' | 'warning' | 'info'

interface ActivityItem {
  id: string
  type: ActivityType
  message: string
  timestamp: string
  source?: string
}

const formatRelativeTime = (date: string): string => {
  const now = new Date()
  const then = new Date(date)
  const diffMs = now.getTime() - then.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  return then.toLocaleDateString()
}

export function ActivityFeed() {
  // For now, use mock data since we don't have a real activity log endpoint
  // TODO: Replace with actual API call when available
  const { data: activities, isLoading } = useQuery({
    queryKey: ['dashboard', 'activities'],
    queryFn: async (): Promise<ActivityItem[]> => {
      // Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 500))

      // Return mock data for now
      return [
        {
          id: '1',
          type: 'success',
          message: 'New user registered',
          timestamp: new Date(Date.now() - 2 * 60000).toISOString(),
          source: 'auth',
        },
        {
          id: '2',
          type: 'info',
          message: 'Function "process-payment" executed successfully',
          timestamp: new Date(Date.now() - 5 * 60000).toISOString(),
          source: 'functions',
        },
        {
          id: '3',
          type: 'warning',
          message: 'Database connections at 75% capacity',
          timestamp: new Date(Date.now() - 15 * 60000).toISOString(),
          source: 'database',
        },
        {
          id: '4',
          type: 'success',
          message: 'Storage bucket "uploads" created',
          timestamp: new Date(Date.now() - 30 * 60000).toISOString(),
          source: 'storage',
        },
        {
          id: '5',
          type: 'error',
          message: 'Function "webhook-handler" failed: timeout',
          timestamp: new Date(Date.now() - 45 * 60000).toISOString(),
          source: 'functions',
        },
      ]
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  })

  const getActivityIcon = (type: ActivityType) => {
    switch (type) {
      case 'success':
        return <CheckCircle2 className='h-4 w-4 text-green-500' />
      case 'error':
        return <XCircle className='h-4 w-4 text-red-500' />
      case 'warning':
        return <AlertTriangle className='h-4 w-4 text-yellow-500' />
      case 'info':
        return <Info className='h-4 w-4 text-blue-500' />
    }
  }

  const getActivityBadgeVariant = (type: ActivityType) => {
    switch (type) {
      case 'success':
        return 'secondary'
      case 'error':
        return 'destructive'
      case 'warning':
        return 'secondary'
      case 'info':
        return 'outline'
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <Clock className='h-4 w-4' />
          Recent Activity
        </CardTitle>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className='space-y-3'>
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className='flex gap-3'>
                <Skeleton className='h-4 w-4 rounded-full' />
                <div className='flex-1 space-y-1'>
                  <Skeleton className='h-4 w-48' />
                  <Skeleton className='h-3 w-24' />
                </div>
              </div>
            ))}
          </div>
        ) : activities && activities.length > 0 ? (
          <div className='space-y-3'>
            {activities.slice(0, 10).map((activity) => (
              <div
                key={activity.id}
                className='group flex gap-3 rounded-lg p-2 transition-colors hover:bg-muted/50'
              >
                <div className='mt-0.5 shrink-0'>{getActivityIcon(activity.type)}</div>
                <div className='min-w-0 flex-1'>
                  <p className='text-sm leading-tight'>{activity.message}</p>
                  <div className='mt-1 flex items-center gap-2'>
                    <span className='text-muted-foreground text-xs'>
                      {formatRelativeTime(activity.timestamp)}
                    </span>
                    {activity.source && (
                      <>
                        <span className='text-muted-foreground/50'>•</span>
                        <Badge variant={getActivityBadgeVariant(activity.type)} className='h-4 px-1 py-0 text-[10px]'>
                          {activity.source}
                        </Badge>
                      </>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className='py-8 text-center text-sm text-muted-foreground'>
            No recent activity
          </div>
        )}
      </CardContent>
    </Card>
  )
}
