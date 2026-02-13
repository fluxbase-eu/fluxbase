import { useQuery } from '@tanstack/react-query'
import { useFluxbaseClient } from '@fluxbase/sdk-react'
import { CheckCircle2, XCircle, AlertTriangle, Loader2 } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface ServiceHealth {
  name: string
  status: 'healthy' | 'degraded' | 'down'
  details?: string
}

export function HealthIndicators() {
  const client = useFluxbaseClient()

  const { data: health, isLoading } = useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const { data, error } = await client.admin.getHealth()
      if (error) throw error
      return data
    },
    refetchInterval: 15000, // Refresh every 15 seconds
  })

  const services: ServiceHealth[] = [
    {
      name: 'API',
      status: health?.status === 'ok' ? 'healthy' : 'down',
      details: health?.status === 'ok' ? 'All endpoints operational' : 'Service unavailable',
    },
    {
      name: 'Database',
      status: health?.services.database ? 'healthy' : 'down',
      details: health?.services.database ? 'Connected' : 'Connection failed',
    },
    {
      name: 'Realtime',
      status: health?.services.realtime ? 'healthy' : 'degraded',
      details: health?.services.realtime ? 'WebSocket connected' : 'WebSocket disabled',
    },
    {
      name: 'Jobs',
      status: 'healthy', // TODO: Get actual job processor status
      details: 'Processing queue',
    },
  ]

  const allHealthy = services.every((s) => s.status === 'healthy')

  return (
    <Card>
      <CardContent className='px-4 py-2'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            {isLoading ? (
              <Loader2 className='h-4 w-4 animate-spin text-muted-foreground' />
            ) : allHealthy ? (
              <CheckCircle2 className='h-4 w-4 text-green-500' />
            ) : (
              <AlertTriangle className='h-4 w-4 text-yellow-500' />
            )}
            <span className='text-sm font-medium'>
              {isLoading ? 'Checking...' : allHealthy ? 'All systems operational' : 'Some systems degraded'}
            </span>
          </div>
          <div className='flex gap-3'>
            {services.map((service) => (
              <div
                key={service.name}
                className='flex items-center gap-1.5'
                title={service.details}
              >
                {isLoading ? (
                  <Skeleton className='h-4 w-4' />
                ) : (
                  <>
                    {service.status === 'healthy' && (
                      <CheckCircle2 className='h-3.5 w-3.5 text-green-500' />
                    )}
                    {service.status === 'degraded' && (
                      <AlertTriangle className='h-3.5 w-3.5 text-yellow-500' />
                    )}
                    {service.status === 'down' && (
                      <XCircle className='h-3.5 w-3.5 text-red-500' />
                    )}
                    <span className='text-xs text-muted-foreground'>{service.name}</span>
                  </>
                )}
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
