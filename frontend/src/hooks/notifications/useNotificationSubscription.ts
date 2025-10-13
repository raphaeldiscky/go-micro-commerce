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
 * @param enabled - Whether the subscription is active (default: true)
 */
export function useNotificationSubscription(enabled = true) {
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)

  useEffect(() => {
    if (!enabled) return

    const client = getSseSubscriptionClient()

    console.log('🔌 Starting notification subscription...', {
      timestamp: new Date().toISOString(),
    })

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
        },
        complete: () => {
          console.log('✅ Notification subscription completed')
        },
      },
    )

    console.log('🔔 Notification subscription started')

    // Cleanup on unmount
    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
        console.log('🔕 Notification subscription stopped')
      }
    }
  }, [enabled, queryClient])
}
