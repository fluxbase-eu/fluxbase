import { getRouteApi, Link } from '@tanstack/react-router'
import { LayoutDashboard } from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { Main } from '@/components/layout/main'
import { FluxbaseStats } from './components/fluxbase-stats'
import { SecuritySummary } from './components/security-summary'

const route = getRouteApi('/_authenticated/')

export function Dashboard() {
  const search = route.useSearch()
  const navigate = route.useNavigate()
  return (
    <Main>
        <div className='bg-background flex items-center justify-between border-b px-6 py-4'>
          <div className='flex items-center gap-3'>
            <div className='bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg'>
              <LayoutDashboard className='text-primary h-5 w-5' />
            </div>
            <div>
              <h1 className='text-xl font-semibold'>Dashboard</h1>
              <p className='text-muted-foreground text-sm'>
                Monitor your Backend as a Service
              </p>
            </div>
          </div>
        </div>

        <div className='p-6'>
        <Tabs
          orientation='vertical'
          value={search.tab || 'overview'}
          onValueChange={(tab) => navigate({ search: { tab } })}
          className='space-y-4'
        >
          <TabsContent value='overview' className='space-y-4'>
            {/* Fluxbase System Stats */}
            <FluxbaseStats />

            {/* Security Summary */}
            <SecuritySummary />

            {/* Quick Actions */}
            <Card>
              <CardHeader>
                <CardTitle>Quick Actions</CardTitle>
                <CardDescription>Common administrative tasks</CardDescription>
              </CardHeader>
              <CardContent>
                <div className='grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Database</p>
                    <Link
                      to='/tables'
                      className='text-primary hover:underline'
                    >
                      Browse database tables →
                    </Link>
                  </div>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Users</p>
                    <Link
                      to='/users'
                      className='text-primary hover:underline'
                    >
                      Manage user accounts →
                    </Link>
                  </div>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Functions</p>
                    <Link
                      to='/functions'
                      className='text-primary hover:underline'
                    >
                      Manage Edge Functions →
                    </Link>
                  </div>
                  <div className='text-sm'>
                    <p className='text-muted-foreground mb-1'>Settings</p>
                    <Link
                      to='/settings'
                      className='text-primary hover:underline'
                    >
                      Configure system settings →
                    </Link>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
        </div>
    </Main>
  )
}
