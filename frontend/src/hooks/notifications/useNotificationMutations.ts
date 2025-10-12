import { QUERY_KEY } from '@/constants/query-key'
import {
  MARK_ALL_AS_READ_MUTATION,
  MARK_AS_READ_MUTATION,
  graphClient,
} from '@/lib/graphql'
import type { Notification } from '@/types/__generated__/graphql'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

interface MarkAsReadMutationResponse {
  markAsRead: Notification
}

interface MarkAllAsReadMutationResponse {
  markAllAsRead: boolean
}

/**
 * Hook for marking a single notification as read
 */
export function useMarkAsRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const data = await graphClient.request<MarkAsReadMutationResponse>(
        MARK_AS_READ_MUTATION,
        { id },
      )
      return data.markAsRead
    },
    onError: (error) => {
      console.error('Failed to mark notification as read:', error)
      toast.error('Failed to mark notification as read')
    },
    onSuccess: () => {
      // Invalidate all notification queries to refresh data
      queryClient.invalidateQueries({
        queryKey: QUERY_KEY.notifications.all,
      })
    },
  })
}

/**
 * Hook for marking all notifications as read
 */
export function useMarkAllAsRead() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      const data = await graphClient.request<MarkAllAsReadMutationResponse>(
        MARK_ALL_AS_READ_MUTATION,
      )
      return data.markAllAsRead
    },
    onError: (error) => {
      console.error('Failed to mark all notifications as read:', error)
      toast.error('Failed to mark all as read')
    },
    onSuccess: () => {
      // Invalidate all notification queries to refresh data
      queryClient.invalidateQueries({
        queryKey: QUERY_KEY.notifications.all,
      })
      toast.success('All notifications marked as read')
    },
  })
}
