import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiKeyService } from '@/services'
import { CreateAPIKeyRequest } from '@/types'

export const useAPIKeys = () => {
  const queryClient = useQueryClient()

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['apiKeys'],
    queryFn: () => apiKeyService.list(),
    retry: 1,
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateAPIKeyRequest) => apiKeyService.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] })
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: number) => apiKeyService.revoke(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] })
    },
  })

  return {
    apiKeys: data?.keys || [],
    isLoading,
    error: error ? String(error) : null,
    refetch,
    createKey: createMutation.mutate,
    revokeKey: revokeMutation.mutate,
    isCreating: createMutation.isPending,
    isRevoking: revokeMutation.isPending,
    createError: createMutation.error ? String(createMutation.error) : null,
    revokeError: revokeMutation.error ? String(revokeMutation.error) : null,
  }
}
