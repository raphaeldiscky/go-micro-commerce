import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  useActiveTab,
  useIsDrawerOpen,
  useNotificationStore,
} from '@/store/notificationStore'
import type { NotificationTab } from '@/types/notification'
import { formatDistanceToNow } from 'date-fns'
import { CheckCheck } from 'lucide-react'
import { NotificationList } from './NotificationList'

export function NotificationDrawer() {
  const isOpen = useIsDrawerOpen()
  const activeTab = useActiveTab()
  const toggleDrawer = useNotificationStore((state) => state.toggleDrawer)
  const setActiveTab = useNotificationStore((state) => state.setActiveTab)
  const markAsRead = useNotificationStore((state) => state.markAsRead)
  const markAllAsRead = useNotificationStore((state) => state.markAllAsRead)
  const getFilteredNotifications = useNotificationStore(
    (state) => state.getFilteredNotifications,
  )
  const getTabCount = useNotificationStore((state) => state.getTabCount)
  const getUnreadCount = useNotificationStore((state) => state.getUnreadCount)
  const notifications = useNotificationStore((state) => state.notifications)

  const allNotifications = getFilteredNotifications('all')
  const unreadNotifications = getFilteredNotifications('unread')
  const readNotifications = getFilteredNotifications('read')

  const allCount = getTabCount('all')
  const unreadCount = getTabCount('unread')
  const readCount = getTabCount('read')

  const handleOpenChange = (open: boolean) => {
    toggleDrawer(open)
  }

  const handleTabChange = (value: string) => {
    setActiveTab(value as NotificationTab)
  }

  const handleNotificationClick = (id: string) => {
    markAsRead(id)
  }

  const handleMarkAllAsRead = () => {
    markAllAsRead()
  }

  const lastUpdated =
    notifications.length > 0
      ? formatDistanceToNow(new Date(notifications[0].createdAt), {
          addSuffix: true,
        })
      : 'Never'

  return (
    <Sheet onOpenChange={handleOpenChange} open={isOpen}>
      <SheetContent
        className="w-full sm:max-w-lg p-0 flex flex-col"
        onOpenAutoFocus={(e) => e.preventDefault()}
        side="right"
      >
        {/* Header Section - Fixed */}
        <div className="flex-shrink-0 border-b px-6 py-4">
          <SheetHeader className="space-y-3">
            <div className="flex items-center justify-between">
              <SheetTitle>Notifications</SheetTitle>
              <Button
                disabled={getUnreadCount() === 0}
                onClick={handleMarkAllAsRead}
                size="sm"
                variant="ghost"
              >
                <CheckCheck className="h-4 w-4 mr-2" />
                Mark all as read
              </Button>
            </div>
            <SheetDescription className="text-xs text-muted-foreground">
              Last updated {lastUpdated}
            </SheetDescription>
          </SheetHeader>
        </div>

        {/* Tabs Section - Scrollable */}
        <div className="flex-1 overflow-hidden">
          <Tabs
            className="flex flex-col h-full"
            onValueChange={handleTabChange}
            value={activeTab}
          >
            <div className="px-6">
              <TabsList className="grid w-full grid-cols-3 mt-4 flex-shrink-0">
                <TabsTrigger className="flex items-center gap-1.5" value="all">
                  All
                  {allCount > 0 && (
                    <div className="h-5 w-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium">
                      {allCount}
                    </div>
                  )}
                </TabsTrigger>
                <TabsTrigger className="flex items-center gap-1.5" value="unread">
                  Unread
                  {unreadCount > 0 && (
                    <div className="h-5 w-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-medium">
                      {unreadCount}
                    </div>
                  )}
                </TabsTrigger>
                <TabsTrigger className="flex items-center gap-1.5" value="read">
                  Read
                  {readCount > 0 && (
                    <div className="h-5 w-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium">
                      {readCount}
                    </div>
                  )}
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent className="mt-4 flex-1 overflow-hidden" value="all">
              <NotificationList
                emptyIcon="inbox"
                emptyMessage="No notifications yet"
                notifications={allNotifications}
                onNotificationClick={handleNotificationClick}
              />
            </TabsContent>

            <TabsContent className="mt-4 flex-1 overflow-hidden" value="unread">
              <NotificationList
                emptyIcon="bell"
                emptyMessage="No unread notifications"
                notifications={unreadNotifications}
                onNotificationClick={handleNotificationClick}
              />
            </TabsContent>

            <TabsContent className="mt-4 flex-1 overflow-hidden" value="read">
              <NotificationList
                emptyIcon="check"
                emptyMessage="No read notifications"
                notifications={readNotifications}
                onNotificationClick={handleNotificationClick}
              />
            </TabsContent>
          </Tabs>
        </div>
      </SheetContent>
    </Sheet>
  )
}
