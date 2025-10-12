import { QUERY_KEY } from '@/constants/query-key'
import {
  GET_TAB_COUNTS_QUERY,
  GET_UNREAD_COUNT_QUERY,
  LIST_NOTIFICATIONS_QUERY,
  LIST_UNREAD_NOTIFICATIONS_QUERY,
  graphClient,
} from '@/lib/graphql'
import type {
  NotificationConnection,
  TabCounts,
  UnreadCount,
} from '@/types/__generated__/graphql'
import { useInfiniteQuery, useQuery } from '@tanstack/react-query'

interface ListNotificationsQueryResponse {
  listNotifications: NotificationConnection
}

interface ListUnreadNotificationsQueryResponse {
  listUnreadNotifications: NotificationConnection
}

interface GetUnreadCountQueryResponse {
  getUnreadCount: UnreadCount
}

interface GetTabCountsQueryResponse {
  getTabCounts: TabCounts
}

/**
 * Hook for fetching notifications with cursor pagination
 * Supports infinite scroll
 */
export function useNotifications(limit = 20) {
  return useInfiniteQuery({
    gcTime: 5 * 60 * 1000, // 5 minutes
    getNextPageParam: (lastPage: ListNotificationsQueryResponse) => {
      const { pageInfo } = lastPage.listNotifications
      return pageInfo.hasNextPage ? pageInfo.endCursor : undefined
    },
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam }) => {
      const data = await graphClient.request<ListNotificationsQueryResponse>(
        LIST_NOTIFICATIONS_QUERY,
        {
          limit,
          cursor: pageParam,
        },
      )
      return data
    },
    queryKey: QUERY_KEY.notifications.list(limit),
    refetchOnWindowFocus: false,
    retry: 3,
    staleTime: 30 * 1000, // 30 seconds
  })
}

/**
 * Hook for fetching unread notifications with cursor pagination
 * Supports infinite scroll
 */
export function useUnreadNotifications(limit = 20) {
  return useInfiniteQuery({
    gcTime: 5 * 60 * 1000,
    getNextPageParam: (lastPage: ListUnreadNotificationsQueryResponse) => {
      const { pageInfo } = lastPage.listUnreadNotifications
      return pageInfo.hasNextPage ? pageInfo.endCursor : undefined
    },
    initialPageParam: undefined as string | undefined,
    queryFn: async ({ pageParam }) => {
      const data =
        await graphClient.request<ListUnreadNotificationsQueryResponse>(
          LIST_UNREAD_NOTIFICATIONS_QUERY,
          {
            limit,
            cursor: pageParam,
          },
        )
      return data
    },
    queryKey: QUERY_KEY.notifications.unreadList(limit),
    refetchOnWindowFocus: false,
    retry: 3,
    staleTime: 30 * 1000,
  })
}

/**
 * Hook for fetching unread notification count
 * Useful for notification badge
 */
export function useUnreadCount() {
  return useQuery({
    gcTime: 5 * 60 * 1000,
    queryFn: async () => {
      const data = await graphClient.request<GetUnreadCountQueryResponse>(
        GET_UNREAD_COUNT_QUERY,
      )
      return data.getUnreadCount.count
    },
    queryKey: QUERY_KEY.notifications.unreadCount(),
    refetchInterval: 60 * 1000, // Refetch every minute
    refetchOnWindowFocus: true,
    retry: 3,
    staleTime: 30 * 1000,
  })
}

/**
 * Hook for fetching tab counts (all, unread, read)
 * Useful for notification drawer tabs
 */
export function useTabCounts() {
  return useQuery({
    gcTime: 5 * 60 * 1000,
    queryFn: async () => {
      const data =
        await graphClient.request<GetTabCountsQueryResponse>(
          GET_TAB_COUNTS_QUERY,
        )
      return data.getTabCounts
    },
    queryKey: QUERY_KEY.notifications.tabCounts(),
    refetchOnWindowFocus: true,
    retry: 3,
    staleTime: 30 * 1000,
  })
}
