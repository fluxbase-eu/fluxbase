import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import {
  Database,
  FileCode,
  Shield,
  Settings,
  ArrowRight,
  Sparkles,
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

interface QuickAction {
  title: string
  description: string
  icon: React.ReactNode
  href: string
  priority?: 'high' | 'normal'
}

export function SmartQuickActions() {
  // Fetch data to determine what actions to show
  const { data: tablesData, isLoading: isLoadingTables } = useQuery({
    queryKey: ['dashboard', 'table-count'],
    queryFn: async () => {
      // Placeholder - would be actual API call
      return { count: 0 }
    },
  })

  const { isLoading: isLoadingUsers } = useQuery({
    queryKey: ['dashboard', 'user-count'],
    queryFn: async () => {
      // Placeholder
      return { count: 0 }
    },
  })

  const { data: functionsData, isLoading: isLoadingFunctions } = useQuery({
    queryKey: ['dashboard', 'function-count'],
    queryFn: async () => {
      // Placeholder
      return { count: 0 }
    },
  })

  // Track whether tables/functions exist for smart action prioritization
  const hasTables = (tablesData?.count ?? 0) > 0
  const hasFunctions = (functionsData?.count ?? 0) > 0

  // Determine quick actions based on current state
  const quickActions: QuickAction[] = []

  // High priority actions for empty states
  if (!hasTables) {
    quickActions.push({
      title: 'Create a Database Table',
      description: 'Define your data schema',
      icon: <Database className='h-4 w-4' />,
      href: '/tables',
      priority: 'high',
    })
  }

  if (!hasFunctions) {
    quickActions.push({
      title: 'Deploy an Edge Function',
      description: 'Add serverless logic to your app',
      icon: <FileCode className='h-4 w-4' />,
      href: '/functions',
      priority: 'high',
    })
  }

  // Always-available actions
  quickActions.push({
    title: 'Configure Security',
    description: 'Set up RLS policies and authentication',
    icon: <Shield className='h-4 w-4' />,
    href: '/policies',
  })

  quickActions.push({
    title: 'System Settings',
    description: 'Manage features and configuration',
    icon: <Settings className='h-4 w-4' />,
    href: '/features',
  })

  const highPriorityActions = quickActions.filter((a) => a.priority === 'high')

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <Sparkles className='h-5 w-5 text-primary' />
          Quick Actions
        </CardTitle>
        <CardDescription>
          {highPriorityActions.length > 0
            ? `Complete setup (${highPriorityActions.length} remaining)`
            : 'Common administrative tasks'}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className='space-y-2'>
          {isLoadingTables || isLoadingUsers || isLoadingFunctions ? (
            <div className='space-y-3'>
              {[1, 2, 3].map((i) => (
                <div key={i} className='flex items-center gap-3'>
                  <Skeleton className='h-9 w-9 rounded-lg' />
                  <div className='flex-1 space-y-1'>
                    <Skeleton className='h-4 w-32' />
                    <Skeleton className='h-3 w-48' />
                  </div>
                </div>
              ))}
            </div>
          ) : (
            quickActions.slice(0, 4).map((action) => (
              <Button
                key={action.href}
                variant={action.priority === 'high' ? 'default' : 'ghost'}
                className='w-full justify-start gap-3'
                asChild
              >
                <Link to={action.href}>
                  <div
                    className={`flex items-center justify-center ${
                      action.priority === 'high' ? 'bg-primary-foreground text-primary' : 'bg-muted'
                    } h-9 w-9 rounded-lg`}
                  >
                    {action.icon}
                  </div>
                  <div className='text-start'>
                    <div className='font-medium'>{action.title}</div>
                    <div className='text-muted-foreground text-xs'>{action.description}</div>
                  </div>
                  <ArrowRight className='ml-auto h-4 w-4 opacity-50' />
                </Link>
              </Button>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}
