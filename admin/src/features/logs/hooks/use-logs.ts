import { useQuery } from '@tanstack/react-query'
import {
  logsApi,
  type LogQueryOptionsAPI,
  type LogQueryResultAPI,
  type LogStatsAPI,
} from '@/lib/api'

/**
 * Hook for fetching logs with filters and pagination
 */
export function useLogs(
  options: LogQueryOptionsAPI,
  enabled = true,
  refetchInterval?: number | false
) {
  return useQuery<LogQueryResultAPI>({
    queryKey: ['logs', options],
    queryFn: () =>
      logsApi.query({
        ...options,
        levels: options.levels?.length ? options.levels : undefined,
      }),
    staleTime: 10000, // 10 seconds
    refetchOnWindowFocus: false,
    enabled,
    refetchInterval,
  })
}

/**
 * Hook for fetching log statistics
 */
export function useLogStats() {
  return useQuery<LogStatsAPI>({
    queryKey: ['log-stats'],
    queryFn: async () => {
      try {
        return await logsApi.getStats()
      } catch (err) {
        // If logging API is not available (404), return default stats
        // instead of throwing an error
        const error = err as { response?: { status?: number } }
        if (error.response?.status === 404) {
          return {
            total_entries: 0,
            entries_by_category: {},
            entries_by_level: {},
          }
        }
        throw err
      }
    },
    staleTime: 30000, // 30 seconds
    refetchOnWindowFocus: false,
    retry: 1, // Only retry once on failure
  })
}
