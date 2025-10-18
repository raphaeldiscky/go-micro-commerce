import { useNotificationSubscription } from '@/hooks/notifications'
import {
  closeSubscriptionClient,
  resetSubscriptionClient,
} from '@/lib/graphql/subscription-client'
import { useIsAuthenticated, useUser } from '@/store/authStore'
import { useEffect, useRef } from 'react'

/**
 * NotificationProvider
 *
 * Manages real-time notification subscriptions at the application level.
 * This provider handles the complete subscription lifecycle including:
 * - Starting subscriptions when user is authenticated
 * - Stopping subscriptions when user logs out
 * - Resetting subscription clients when token refreshes
 * - Cleaning up connections on unmount
 *
 * This ensures notifications work reliably across route navigation
 * and handles authentication state changes properly.
 */
export function NotificationProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const isAuthenticated = useIsAuthenticated()
  const user = useUser()
  const previousUserIdRef = useRef<string | null>(null)

  // Subscribe to notifications when authenticated
  // The hook internally handles connection and reconnection logic
  useNotificationSubscription(isAuthenticated)

  // Handle authentication state changes
  useEffect(() => {
    if (!isAuthenticated) {
      console.log('🔕 NotificationProvider - User logged out, closing clients')
      closeSubscriptionClient()
      previousUserIdRef.current = null
    }
  }, [isAuthenticated])

  // Handle token refresh detection via user ID changes
  // When the token refreshes, authStore refetches the user
  // We detect this and reset subscription clients to reconnect with the new token
  useEffect(() => {
    const currentUserId = user?.id ?? null

    // If user exists and user ID changed (not initial mount)
    if (currentUserId && previousUserIdRef.current !== currentUserId) {
      // Only reset if this isn't the first time we're seeing a user
      if (previousUserIdRef.current !== null) {
        console.log(
          '🔄 NotificationProvider - User changed, resetting subscription clients',
          {
            from: previousUserIdRef.current,
            to: currentUserId,
          },
        )
        resetSubscriptionClient()
      }

      previousUserIdRef.current = currentUserId
    }
  }, [user?.id])

  // Cleanup on unmount (app closing)
  useEffect(() => {
    return () => {
      console.log('🧹 NotificationProvider - Unmounting, cleaning up clients')
      closeSubscriptionClient()
    }
  }, [])

  // Provider doesn't render any UI, just manages infrastructure
  return <>{children}</>
}
