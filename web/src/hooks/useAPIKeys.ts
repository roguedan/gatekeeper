import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { apiKeyService } from '@/services'
import { CreateAPIKeyRequest } from '@/types'
import { useToastContext } from '@/contexts'

export const useAPIKeys = () => {
  const queryClient = useQueryClient()
  const { showPending, updateToast } = useToastContext()

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['apiKeys'],
    queryFn: () => apiKeyService.list(),
    retry: 1,
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateAPIKeyRequest) => apiKeyService.create(data),
    onMutate: () => {
      const toastId = showPending('Creating API key...', 'Please wait while we generate your new key')
      return { toastId }
    },
    onSuccess: (_, __, context) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] })
      updateToast(context.toastId, {
        variant: 'success',
        title: 'API key created!',
        description: 'Your new API key has been generated successfully. Save it securely!'
      })
    },
    onError: (error, _, context) => {
      if (context?.toastId) {
        updateToast(context.toastId, {
          variant: 'error',
          title: 'Failed to create API key',
          description: error instanceof Error ? error.message : 'An error occurred'
        })
      }
    },
  })

  const revokeMutation = useMutation({
    mutationFn: (id: number) => apiKeyService.revoke(id),
    onMutate: () => {
      const toastId = showPending('Revoking API key...', 'Removing key access')
      return { toastId }
    },
    onSuccess: (_, __, context) => {
      queryClient.invalidateQueries({ queryKey: ['apiKeys'] })
      updateToast(context.toastId, {
        variant: 'success',
        title: 'API key revoked',
        description: 'The key has been successfully revoked and can no longer be used'
      })
    },
    onError: (error, _, context) => {
      if (context?.toastId) {
        updateToast(context.toastId, {
          variant: 'error',
          title: 'Failed to revoke API key',
          description: error instanceof Error ? error.message : 'An error occurred'
        })
      }
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
