import { QUERY_KEY } from '@/constants/query-key'
import { NOTIFICATION_EVENTS_SUBSCRIPTION } from '@/lib/graphql'
import { getSseSubscriptionClient } from '@/lib/graphql/subscription-client'
import type {
  NotificationEvent,
  PaymentStatus,
} from '@/types/__generated__/graphql'
import { PushNotificationType } from '@/types/__generated__/graphql'
import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'

interface SubscriptionData {
  notificationEvents: NotificationEvent
}

interface PaymentNotificationMetadata {
  orderId: string
  paymentId: string
  status: PaymentStatus
  amount?: string
  currency?: string
}

/**
 * Hook for subscribing to real-time payment status updates via SSE
 * Listens to payment-related notifications (PAYMENT_SUCCESS, PAYMENT_FAILED, PAYMENT_TIMEOUT)
 * and invalidates payment queries to trigger refetch
 *
 * @param orderId - The order ID to listen for payment updates
 * @param options - Subscription options
 */
export function usePaymentStatusSubscription(
  orderId: string,
  options?: {
    enabled?: boolean
    onPaymentSuccess?: () => void
    onPaymentFailed?: (error?: string) => void
    onPaymentTimeout?: () => void
  },
) {
  const queryClient = useQueryClient()
  const unsubscribeRef = useRef<(() => void) | null>(null)
  const isSubscribedRef = useRef(false)

  useEffect(() => {
    if (!options?.enabled) return

    // Prevent duplicate subscriptions
    if (isSubscribedRef.current) {
      console.log(
        '⚠️ Payment subscription already active, skipping duplicate mount',
      )
      return
    }

    const client = getSseSubscriptionClient()

    console.log('🔌 Starting payment status subscription...', {
      orderId,
      timestamp: new Date().toISOString(),
    })

    // Mark as subscribed before creating the subscription
    isSubscribedRef.current = true

    // Subscribe to notification events and filter for payment events
    unsubscribeRef.current = client.subscribe<SubscriptionData>(
      {
        query: NOTIFICATION_EVENTS_SUBSCRIPTION,
      },
      {
        next: (result) => {
          const event = result.data?.notificationEvents
          if (!event) {
            console.warn('⚠️ Received subscription data without event')
            return
          }

          // Only process NewNotification events with payment types
          if (event.__typename !== 'NewNotification') {
            return
          }

          // Filter for payment-related notification types
          const PAYMENT_NOTIFICATION_TYPES = new Set<string>([
            PushNotificationType.PaymentSuccess,
            PushNotificationType.PaymentFailed,
            PushNotificationType.PaymentTimeout,
          ])

          if (!PAYMENT_NOTIFICATION_TYPES.has(event.type)) {
            return
          }

          // Parse metadata to get order ID and payment details
          let metadata: PaymentNotificationMetadata | null = null
          try {
            metadata = event.metadata ? JSON.parse(event.metadata) : null
          } catch (error) {
            console.error(
              'Failed to parse payment notification metadata:',
              error,
            )
            return
          }

          // Only process notifications for the specific order
          if (!metadata || metadata.orderId !== orderId) {
            return
          }

          console.log('💳 Payment Status Update:', {
            type: event.type,
            orderId: metadata.orderId,
            paymentId: metadata.paymentId,
            status: metadata.status,
          })

          // Invalidate payment query to trigger refetch
          queryClient.invalidateQueries({
            queryKey: QUERY_KEY.order.paymentByOrderId(orderId),
          })

          // Handle different payment status types
          switch (event.type) {
            case PushNotificationType.PaymentSuccess:
              console.log('✅ Payment completed successfully')
              options.onPaymentSuccess?.()
              break

            case PushNotificationType.PaymentFailed:
              console.log('❌ Payment failed')
              options.onPaymentFailed?.(event.message)
              break

            case PushNotificationType.PaymentTimeout:
              console.log('⏱️ Payment timed out')
              options.onPaymentTimeout?.()
              break
          }
        },
        error: (error) => {
          console.error('❌ Payment subscription error:', error)
          isSubscribedRef.current = false
        },
        complete: () => {
          console.log('✅ Payment subscription completed normally')
          isSubscribedRef.current = false
        },
      },
    )

    console.log('🔔 Payment status subscription started for order:', orderId)

    // Cleanup on unmount or when dependencies change
    return () => {
      if (unsubscribeRef.current) {
        unsubscribeRef.current()
        unsubscribeRef.current = null
        isSubscribedRef.current = false
        console.log('🔕 Payment subscription cleanup (component unmount)')
      }
    }
  }, [orderId, options?.enabled, queryClient, options])
}
