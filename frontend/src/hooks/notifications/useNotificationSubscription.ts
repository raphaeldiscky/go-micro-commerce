import { QUERY_KEY } from '@/constants/query-key'
import { NOTIFICATION_EVENTS_SUBSCRIPTION } from '@/lib/graphql'
import { getSseSubscriptionClient } from '@/lib/graphql/subscription-client'
import type { NotificationEvent } from '@/types/__generated__/graphql'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { toast } from 'sonner'

interface SubscriptionData {
  notificationEvents: NotificationEvent
}

/**
 * Hook for subscribing to real-time notification events via SSE
 * Automatically handles new notifications, read events, and deleted events
 *
 * NOTE: In development with React StrictMode enabled, you may see the subscription
 * start, stop, and restart. This is expected behavior - StrictMode intentionally
 * double-mounts components to help detect side effects. The subscription will
 * work correctly in production.
 *
 * @param enabled - Whether the subscription is active (default: true)
 */
export function useNotificationSubscription(enabled = true) {
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)
  const isSubscribedRef = useRef(false)

  useEffect(() => {
    if (!enabled) return

    // Prevent duplicate subscriptions during React StrictMode double-mounting
    if (isSubscribedRef.current) {
      console.log('⚠️ Subscription already active, skipping duplicate mount')
      return
    }

    const client = getSseSubscriptionClient()

    console.log('🔌 Starting notification subscription...', {
      timestamp: new Date().toISOString(),
    })

    // Mark as subscribed before creating the subscription
    isSubscribedRef.current = true

    // Subscribe to notification events
    unsubscribeRef.current = client.subscribe<SubscriptionData>(
      {
        query: NOTIFICATION_EVENTS_SUBSCRIPTION,
      },
      {
        next: (result) => {
          console.log('📬 Notification subscription - received data:', {
            hasData: !!result.data,
            hasEvent: !!result.data?.notificationEvents,
            timestamp: new Date().toISOString(),
          })

          const event = result.data?.notificationEvents
          if (!event) {
            console.warn('⚠️ Received subscription data without event')
            return
          }

          console.log('📬 Notification Event Received:', event)

          // Handle different event types
          switch (event.__typename) {
            case 'NewNotification': {
              // Show toast for new notification
              toast.info(event.title, {
                description: event.message,
                descriptionClassName: '!text-secondary-foreground',
              })

              // Invalidate notification queries to refresh data
              queryClient.invalidateQueries({
                queryKey: QUERY_KEY.notifications.all,
              })
              break
            }

            case 'NotificationRead': {
              // Update read status in cache
              queryClient.invalidateQueries({
                queryKey: QUERY_KEY.notifications.all,
              })
              break
            }

            case 'NotificationDeleted': {
              // Remove notification from cache
              queryClient.invalidateQueries({
                queryKey: QUERY_KEY.notifications.all,
              })
              break
            }

            default:
              console.warn('Unknown notification event type:', event)
          }
        },
        error: (error) => {
          console.error('❌ Notification subscription error:', error)
          // Reset subscription flag on error to allow reconnection
          isSubscribedRef.current = false
        },
        complete: () => {
          console.log('✅ Notification subscription completed normally')
          // Reset subscription flag on completion
          isSubscribedRef.current = false
        },
      },
    )

    console.log('🔔 Notification subscription started')

    // Cleanup on unmount or when enabled changes
    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
        isSubscribedRef.current = false
        console.log('🔕 Notification subscription cleanup (component unmount)')
      }
    }
    // Only depend on 'enabled', not 'queryClient'
    // queryClient is stable from useQueryClient() and doesn't need to be a dependency
  }, [enabled])
}
