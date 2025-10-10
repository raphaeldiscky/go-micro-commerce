import { QUERY_KEY } from '@/constants/query-key'
import { useUser } from '@/hooks/auth/useAuth'
import { ONLINE_USERS_QUERY, graphClient } from '@/lib/graphql'
import type { User } from '@/types/__generated__/graphql'
import { PresenceStatus } from '@/types/__generated__/graphql'
import { useMutation, useQuery } from '@tanstack/react-query'
import { gql } from 'graphql-request'
import { useCallback, useEffect, useState } from 'react'

interface OnlineUsersQueryResponse {
  onlineUsers: Array<User>
}

const UPDATE_PRESENCE_MUTATION = gql`
  mutation UpdatePresence($status: PresenceStatus!) {
    updatePresence(status: $status) {
      userId
      status
      lastSeen
    }
  }
`

/**
 * Hook for managing user presence via GraphQL
 */
export function usePresence() {
  const [currentStatus, setCurrentStatus] = useState<PresenceStatus>(
    PresenceStatus.Online,
  )
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set())
  const user = useUser()

  // Mutation for updating presence status
  const updatePresenceMutation = useMutation({
    mutationFn: async (status: PresenceStatus) => {
      return graphClient.request(UPDATE_PRESENCE_MUTATION, { status })
    },
    onSuccess: (_, status) => {
      setCurrentStatus(status)
    },
  })

  // Query for online users
  const { data: onlineUsersList, refetch: refetchOnlineUsers } = useQuery({
    enabled: !!user, // Only fetch when user is authenticated
    queryFn: async () => {
      const data =
        await graphClient.request<OnlineUsersQueryResponse>(ONLINE_USERS_QUERY)
      return data.onlineUsers
    },
    queryKey: QUERY_KEY.chat.onlineUsers(),
    refetchInterval: user ? 30 * 1000 : false,
    staleTime: 15 * 1000, // Consider stale after 15 seconds
  })

  /**
   * Update user's presence status
   */
  const setPresenceStatus = useCallback(
    (status: PresenceStatus) => {
      if (!user) return

      updatePresenceMutation.mutate(status)
    },
    // updatePresenceMutation is stable from useMutation, no need to include in deps
    [user],
  )

  /**
   * Check if a user is online
   */
  const isUserOnline = useCallback(
    (userId: string) => {
      return onlineUsers.has(userId)
    },
    [onlineUsers],
  )

  /**
   * Update online users from WebSocket events
   */
  const updateOnlineUsers = useCallback((users: Array<string>) => {
    setOnlineUsers(new Set(users))
  }, [])

  /**
   * Add online user
   */
  const addOnlineUser = useCallback((userId: string) => {
    setOnlineUsers((prev) => new Set([userId, ...prev]))
  }, [])

  /**
   * Remove online user
   */
  const removeOnlineUser = useCallback((userId: string) => {
    setOnlineUsers((prev) => {
      const newSet = new Set(prev)
      newSet.delete(userId)
      return newSet
    })
  }, [])

  // Update online users when query data changes
  useEffect(() => {
    if (onlineUsersList && Array.isArray(onlineUsersList)) {
      // Extract user IDs from User objects
      const userIds = onlineUsersList.map((u) => u.id)
      setOnlineUsers(new Set(userIds))
    } else {
      // Handle case where API might return wrapped data or different format
      console.warn('Online users data is not an array:', onlineUsersList)
      setOnlineUsers(new Set())
    }
  }, [onlineUsersList])

  // Set user as online when component mounts
  useEffect(() => {
    // Skip if not authenticated
    if (!user) {
      return
    }

    // Initial setup - set user as online
    setPresenceStatus(PresenceStatus.Online)

    // Set user as away when window loses focus
    const handleVisibilityChange = () => {
      if (document.hidden) {
        setPresenceStatus(PresenceStatus.Away)
      } else {
        setPresenceStatus(PresenceStatus.Online)
      }
    }

    // Set user as offline when window is closed
    const handleBeforeUnload = () => {
      setPresenceStatus(PresenceStatus.Offline)
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('beforeunload', handleBeforeUnload)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
      window.removeEventListener('beforeunload', handleBeforeUnload)
      setPresenceStatus(PresenceStatus.Offline)
    }
  }, [user, setPresenceStatus])

  return {
    addOnlineUser,
    currentStatus,
    isUserOnline,
    onlineUsers: Array.from(onlineUsers),
    refetchOnlineUsers,
    removeOnlineUser,
    setPresenceStatus,
    updateOnlineUsers,
  }
}
