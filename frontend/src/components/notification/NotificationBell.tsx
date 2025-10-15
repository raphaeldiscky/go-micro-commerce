import { Badge } from '@/components/ui/badge'
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
        aria-label={`Notifications with ${unreadCount} unread items`}
        className={cn(
          'relative h-10 w-10 rounded-full transition-all duration-200 hover:scale-105 active:scale-95'
        )}
        onClick={handleClick}
        size="icon"
        variant="ghost"
      >
        <Bell className="h-5 w-5" />

        {/* Badge showing unread notification count */}
        {unreadCount > 0 && (
          <Badge
            className={cn(
              'absolute -right-1 -top-1 z-10',
              // Shape and alignment
              'flex h-5 w-5 items-center justify-center rounded-full',
              // Remove default padding and line-height issues
              'p-0 leading-none',
              // Visual styles
              'bg-red-500 text-white text-[10px] font-bold',
              // Optional animation
              'animate-pulse',
            )}
          >
            {unreadCount > 99 ? '99+' : unreadCount}
          </Badge>
        )}
      </Button>
      <NotificationDrawer />
    </>
  )
}
