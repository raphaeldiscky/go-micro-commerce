import { cn } from '@/lib/utils'
import type {
  Notification,
  NotificationType,
} from '@/types/__generated__/graphql'
import { formatDistanceToNow } from 'date-fns'
import {
  Bell,
  Check,
  MessageSquare,
  Package,
  ShoppingBag,
  Truck,
  X,
} from 'lucide-react'

interface NotificationRowProps {
  notification: Notification
  onClick: (id: string) => void
}

function getNotificationIcon(type: NotificationType) {
  switch (type) {
    case 'NEW_MESSAGE':
      return MessageSquare
    case 'NEW_PRODUCT':
      return Package
    case 'ORDER_CONFIRMED':
    case 'ORDER_UPDATE':
      return ShoppingBag
    case 'ORDER_SHIPPED':
    case 'ORDER_DELIVERED':
      return Truck
    case 'ORDER_CANCELLED':
      return X
    case 'PAYMENT_SUCCESS':
      return Check
    case 'SYSTEM_ALERT':
      return Bell
    default:
      return Bell
  }
}

function getIconColor(type: NotificationType): string {
  switch (type) {
    case 'NEW_MESSAGE':
      return 'text-blue-500'
    case 'NEW_PRODUCT':
      return 'text-purple-500'
    case 'ORDER_CONFIRMED':
    case 'ORDER_UPDATE':
      return 'text-orange-500'
    case 'ORDER_SHIPPED':
    case 'ORDER_DELIVERED':
      return 'text-green-500'
    case 'ORDER_CANCELLED':
      return 'text-red-500'
    case 'PAYMENT_SUCCESS':
      return 'text-emerald-500'
    case 'SYSTEM_ALERT':
      return 'text-yellow-500'
    default:
      return 'text-gray-500'
  }
}

export function NotificationRow({
  notification,
  onClick,
}: NotificationRowProps) {
  const Icon = getNotificationIcon(notification.type)
  const iconColor = getIconColor(notification.type)

  const isUnread = !notification.isRead

  const handleClick = () => {
    if (!notification.isRead) {
      onClick(notification.id)
    }
  }

  return (
    <button
      aria-label={`${notification.title}. ${notification.isRead ? 'Read' : 'Unread'}`}
      className={cn(
        'w-full text-left transition-all duration-200 overflow-hidden',
        'hover:bg-accent cursor-pointer border-b border-border',
        'focus:outline-none focus:ring-2 focus:ring-ring',
        isUnread &&
          'bg-blue-50/50 dark:bg-blue-950/20 border-l-4 border-l-blue-500',
        notification.isRead && 'bg-transparent',
      )}
      onClick={handleClick}
      type="button"
    >
      <div className="flex gap-3 px-6 py-3">
        {/* Icon */}
        <div className="flex-shrink-0 mt-0.5">
          <div
            className={cn(
              'h-10 w-10 rounded-full flex items-center justify-center',
              'bg-background border-2',
            )}
          >
            <Icon className={cn('h-5 w-5', iconColor)} />
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2 mb-1">
            <h4 className="font-semibold text-sm leading-tight">
              {notification.title}
            </h4>
            {isUnread && (
              <div className="flex-shrink-0 w-2 h-2 rounded-full bg-blue-500 mt-1.5" />
            )}
          </div>

          <p className="text-sm text-muted-foreground line-clamp-2 mb-1.5 leading-relaxed">
            {notification.message}
          </p>

          <span className="text-xs text-muted-foreground">
            {formatDistanceToNow(new Date(notification.createdAt), {
              addSuffix: true,
            })}
          </span>
        </div>
      </div>
    </button>
  )
}
