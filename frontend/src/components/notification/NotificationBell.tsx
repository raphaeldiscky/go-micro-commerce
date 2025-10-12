import { Button } from '@/components/ui/button'
import {
  useNotificationSubscription,
  useUnreadCount,
} from '@/hooks/notifications'
import { cn } from '@/lib/utils'
import { useNotificationStore } from '@/store/notificationStore'
import { Bell } from 'lucide-react'
import { NotificationDrawer } from './NotificationDrawer'

export function NotificationBell() {
  const { data: unreadCount = 0 } = useUnreadCount()
  const toggleDrawer = useNotificationStore((state) => state.toggleDrawer)

  // Subscribe to real-time notification events
  useNotificationSubscription()

  const handleClick = () => {
    toggleDrawer(true)
  }

  return (
    <>
      <Button
        aria-label={`Notifications${unreadCount > 0 ? `, ${unreadCount} unread` : ''}`}
        className="relative"
        onClick={handleClick}
        size="sm"
        variant="ghost"
      >
        <Bell className="h-5 w-5" />
        {unreadCount > 0 && (
          <div
            className={cn(
              'absolute -top-1.5 -right-1.5',
              'h-5 w-5 min-w-[20px] rounded-full',
              'bg-destructive text-destructive-foreground',
              'flex items-center justify-center',
              'text-xs font-semibold',
              'animate-in fade-in zoom-in duration-200',
              'shadow-sm',
            )}
          >
            {unreadCount > 99 ? '99+' : unreadCount}
          </div>
        )}
      </Button>
      <NotificationDrawer />
    </>
  )
}
