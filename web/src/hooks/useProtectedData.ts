import { useQuery } from '@tanstack/react-query'
import { protectedService } from '@/services'

export const useProtectedData = () => {
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['protectedData'],
    queryFn: () => protectedService.getData(),
    retry: 1,
    enabled: false, // Don't fetch automatically
  })

  return {
    data: data?.data,
    message: data?.message,
    isLoading,
    error: error ? String(error) : null,
    fetchData: refetch,
  }
}
