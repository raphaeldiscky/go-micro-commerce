import { ScrollArea } from '@/components/ui/scroll-area'
import type { Notification } from '@/types/notification'
import { BellOff, CheckCircle2, Inbox } from 'lucide-react'
import { NotificationRow } from './NotificationRow'

interface NotificationListProps {
  notifications: Array<Notification>
  onNotificationClick: (id: string) => void
  emptyMessage?: string
  emptyIcon?: 'inbox' | 'bell' | 'check'
}

function EmptyState({
  message,
  icon,
}: {
  message: string
  icon: 'inbox' | 'bell' | 'check'
}) {
  const Icon =
    icon === 'inbox' ? Inbox : icon === 'bell' ? BellOff : CheckCircle2

  return (
    <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
      <div className="rounded-full bg-muted p-4 mb-4">
        <Icon className="h-8 w-8 text-muted-foreground" />
      </div>
      <p className="text-sm text-muted-foreground">{message}</p>
    </div>
  )
}

export function NotificationList({
  notifications,
  onNotificationClick,
  emptyMessage = 'No notifications yet',
  emptyIcon = 'inbox',
}: NotificationListProps) {
  if (notifications.length === 0) {
    return <EmptyState icon={emptyIcon} message={emptyMessage} />
  }

  return (
    <ScrollArea className="h-full w-full">
      <div className="space-y-0">
        {notifications.map((notification) => (
          <NotificationRow
            key={notification.id}
            notification={notification}
            onClick={onNotificationClick}
          />
        ))}
      </div>
    </ScrollArea>
  )
}
