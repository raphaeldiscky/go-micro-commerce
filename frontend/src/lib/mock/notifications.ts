import type { Notification, NotificationType } from '@/types/notification'
import { subDays, subHours, subMinutes } from 'date-fns'

const USER_ID = '123e4567-e89b-12d3-a456-426614174000'

interface MockNotificationTemplate {
  type: NotificationType
  title: string
  message: string
  metadata?: Record<string, unknown>
}

const notificationTemplates: Array<MockNotificationTemplate> = [
  {
    type: 'NEW_MESSAGE',
    title: 'You have a new message',
    message: 'John Doe sent you a message',
    metadata: { conversationId: 'conv-001' },
  },
  {
    type: 'NEW_MESSAGE',
    title: 'You have a new message',
    message: 'Sarah Connor replied to your message',
    metadata: { conversationId: 'conv-002' },
  },
  {
    type: 'NEW_MESSAGE',
    title: 'New conversation',
    message: 'Mike Johnson started a conversation with you',
    metadata: { conversationId: 'conv-003' },
  },
  {
    type: 'NEW_PRODUCT',
    title: 'New Product Released',
    message: 'Check out our new wireless earbuds!',
    metadata: { productId: 'prod-789' },
  },
  {
    type: 'NEW_PRODUCT',
    title: 'New Arrival',
    message: 'Premium laptop now available in store',
    metadata: { productId: 'prod-456' },
  },
  {
    type: 'ORDER_CONFIRMED',
    title: 'Order Confirmed',
    message: 'Your order #ORD-1234 has been confirmed',
    metadata: { orderId: 'ORD-1234', totalPrice: 299.99 },
  },
  {
    type: 'ORDER_SHIPPED',
    title: 'Order Shipped',
    message: 'Your order #ORD-5678 has been shipped',
    metadata: { orderId: 'ORD-5678', trackingNumber: 'TRACK123456' },
  },
  {
    type: 'ORDER_DELIVERED',
    title: 'Order Delivered',
    message: 'Your order #ORD-9012 has been delivered',
    metadata: { orderId: 'ORD-9012' },
  },
  {
    type: 'ORDER_CANCELLED',
    title: 'Order Cancelled',
    message: 'Your order #ORD-3456 has been cancelled',
    metadata: { orderId: 'ORD-3456', reason: 'Customer request' },
  },
  {
    type: 'PAYMENT_SUCCESS',
    title: 'Payment Received',
    message: 'Payment of $299.99 received successfully',
    metadata: { amount: 299.99, paymentId: 'PAY-001' },
  },
  {
    type: 'PAYMENT_SUCCESS',
    title: 'Payment Confirmed',
    message: 'Your payment of $149.50 has been processed',
    metadata: { amount: 149.5, paymentId: 'PAY-002' },
  },
  {
    type: 'SYSTEM_ALERT',
    title: 'System Maintenance',
    message: 'Scheduled maintenance on Sunday at 2:00 AM',
    metadata: { maintenanceDate: '2025-10-13T02:00:00Z' },
  },
  {
    type: 'SYSTEM_ALERT',
    title: 'Security Update',
    message: 'Please update your password for enhanced security',
    metadata: { priority: 'high' },
  },
  {
    type: 'ORDER_UPDATE',
    title: 'Order Processing',
    message: 'Your order #ORD-7890 is being processed',
    metadata: { orderId: 'ORD-7890', status: 'processing' },
  },
  {
    type: 'ORDER_UPDATE',
    title: 'Order Ready',
    message: 'Your order #ORD-2468 is ready for pickup',
    metadata: { orderId: 'ORD-2468', pickupLocation: 'Store #5' },
  },
]

function generateRandomDate(): string {
  const now = new Date()
  const randomMinutesAgo = Math.floor(Math.random() * 10080) // Up to 7 days ago in minutes

  let date: Date
  if (randomMinutesAgo < 60) {
    date = subMinutes(now, randomMinutesAgo)
  } else if (randomMinutesAgo < 1440) {
    date = subHours(now, Math.floor(randomMinutesAgo / 60))
  } else {
    date = subDays(now, Math.floor(randomMinutesAgo / 1440))
  }

  return date.toISOString()
}

function generateId(): string {
  return `${Math.random().toString(36).slice(2, 11)}-${Math.random().toString(36).slice(2, 11)}`
}

export function generateMockNotifications(): Array<Notification> {
  const notifications: Array<Notification> = []

  // Generate 15-20 notifications
  const count = Math.floor(Math.random() * 6) + 15

  for (let i = 0; i < count; i++) {
    const template =
      notificationTemplates[
        Math.floor(Math.random() * notificationTemplates.length)
      ]
    const createdAt = generateRandomDate()
    const isRead = Math.random() > 0.6 // 40% read, 60% unread

    const notification: Notification = {
      id: generateId(),
      userId: USER_ID,
      type: template.type,
      title: template.title,
      message: template.message,
      metadata: template.metadata || null,
      isRead,
      readAt: isRead ? new Date(createdAt).toISOString() : null,
      createdAt,
      updatedAt: createdAt,
    }

    notifications.push(notification)
  }

  // Sort by createdAt (newest first)
  return notifications.sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  )
}
