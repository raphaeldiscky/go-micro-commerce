import type { PresenceUpdate } from '@/lib/api'
import { getOnlineUsers, updatePresence } from '@/lib/api'
import { useMutation, useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useState } from 'react'

/**
 * Hook for managing user presence
 */
export function usePresence() {
  const [currentStatus, setCurrentStatus] =
    useState<PresenceUpdate['status']>('online')
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set())

  // Query for online users
  const { data: onlineUsersList, refetch: refetchOnlineUsers } = useQuery({
    queryFn: getOnlineUsers,
    queryKey: ['online-users'],
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
    staleTime: 15 * 1000, // Consider stale after 15 seconds
  })

  // Mutation for updating presence
  const updatePresenceMutation = useMutation({
    mutationFn: (presence: PresenceUpdate) => updatePresence(presence),
    onError: (error) => {
      console.error('Failed to update presence:', error)
    },
    onSuccess: (_, variables) => {
      setCurrentStatus(variables.status)
    },
  })

  /**
   * Update user's presence status
   */
  const setPresenceStatus = useCallback(
    (status: PresenceUpdate['status']) => {
      const presenceUpdate: PresenceUpdate = {
        last_seen: status === 'offline' ? new Date().toISOString() : undefined,
        status,
      }

      updatePresenceMutation.mutate(presenceUpdate)
    },
    [updatePresenceMutation],
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
      setOnlineUsers(new Set(onlineUsersList))
    } else {
      // Handle case where API might return wrapped data or different format
      console.warn('Online users data is not an array:', onlineUsersList)
      setOnlineUsers(new Set())
    }
  }, [onlineUsersList])

  // Set user as online when component mounts - no dependency needed for one-time setup
  useEffect(() => {
    // Initial setup
    setPresenceStatus('online')

    // Set user as away when window loses focus
    const handleVisibilityChange = () => {
      if (document.hidden) {
        setPresenceStatus('away')
      } else {
        setPresenceStatus('online')
      }
    }

    // Set user as offline when window is closed
    const handleBeforeUnload = () => {
      // Use navigator.sendBeacon for reliable offline status
      const presenceData = JSON.stringify({
        last_seen: new Date().toISOString(),
        status: 'offline',
      })
      navigator.sendBeacon('/api/chats/v1/presence', presenceData)
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('beforeunload', handleBeforeUnload)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
      window.removeEventListener('beforeunload', handleBeforeUnload)
      // Set offline on unmount
      setPresenceStatus('offline')
    }
    // Only run on mount/unmount, not when setPresenceStatus changes
  }, [])

  return {
    addOnlineUser,
    currentStatus,
    isUpdating: updatePresenceMutation.isPending,
    isUserOnline,
    onlineUsers: Array.from(onlineUsers),
    refetchOnlineUsers,
    removeOnlineUser,
    setPresenceStatus,
    updateOnlineUsers,
  }
}
