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
  useMarkAllAsRead,
  useMarkAsRead,
  useNotifications,
  useTabCounts,
  useUnreadNotifications,
} from '@/hooks/notifications'
import {
  useActiveTab,
  useIsDrawerOpen,
  useNotificationStore,
} from '@/store/notificationStore'
import { formatDistanceToNow } from 'date-fns'
import { CheckCheck, Loader2 } from 'lucide-react'
import { useMemo } from 'react'
import { NotificationList } from './NotificationList'

export function NotificationDrawer() {
  const isOpen = useIsDrawerOpen()
  const activeTab = useActiveTab()
  const toggleDrawer = useNotificationStore((state) => state.toggleDrawer)
  const setActiveTab = useNotificationStore((state) => state.setActiveTab)

  // Fetch notifications data
  const { data: allNotificationsData, isLoading: isLoadingAll } =
    useNotifications(20)
  const { data: unreadNotificationsData, isLoading: isLoadingUnread } =
    useUnreadNotifications(20)
  const { data: tabCounts } = useTabCounts()

  // Mutations
  const markAsReadMutation = useMarkAsRead()
  const markAllAsReadMutation = useMarkAllAsRead()

  // Extract notifications from infinite query pages
  const allNotifications = useMemo(() => {
    if (!allNotificationsData) return []
    return allNotificationsData.pages.flatMap((page) =>
      page.listNotifications.edges.map((edge) => edge.node),
    )
  }, [allNotificationsData])

  const unreadNotifications = useMemo(() => {
    if (!unreadNotificationsData) return []
    return unreadNotificationsData.pages.flatMap((page) =>
      page.listUnreadNotifications.edges.map((edge) => edge.node),
    )
  }, [unreadNotificationsData])

  const readNotifications = useMemo(() => {
    return allNotifications.filter((n) => n.isRead)
  }, [allNotifications])

  const handleOpenChange = (open: boolean) => {
    toggleDrawer(open)
  }

  const handleTabChange = (value: string) => {
    setActiveTab(value as 'all' | 'unread' | 'read')
  }

  const handleNotificationClick = (id: string) => {
    markAsReadMutation.mutate(id)
  }

  const handleMarkAllAsRead = () => {
    markAllAsReadMutation.mutate()
  }

  const lastUpdated =
    allNotifications.length > 0
      ? formatDistanceToNow(new Date(allNotifications[0].createdAt), {
          addSuffix: true,
        })
      : 'Never'

  const isLoading = isLoadingAll || isLoadingUnread

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
                disabled={
                  (tabCounts?.unread ?? 0) === 0 ||
                  markAllAsReadMutation.isPending
                }
                onClick={handleMarkAllAsRead}
                size="sm"
                variant="ghost"
              >
                {markAllAsReadMutation.isPending ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <CheckCheck className="h-4 w-4 mr-2" />
                )}
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
          {isLoading ? (
            <div className="flex items-center justify-center h-full">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <Tabs
              className="flex flex-col h-full"
              onValueChange={handleTabChange}
              value={activeTab}
            >
              <div className="px-6">
                <TabsList className="grid w-full grid-cols-3 mt-4 flex-shrink-0">
                  <TabsTrigger
                    className="flex items-center gap-1.5"
                    value="all"
                  >
                    All
                    {(tabCounts?.all ?? 0) > 0 && (
                      <div className="h-5 w-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium">
                        {tabCounts?.all}
                      </div>
                    )}
                  </TabsTrigger>
                  <TabsTrigger
                    className="flex items-center gap-1.5"
                    value="unread"
                  >
                    Unread
                    {(tabCounts?.unread ?? 0) > 0 && (
                      <div className="h-5 w-5 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-medium">
                        {tabCounts?.unread}
                      </div>
                    )}
                  </TabsTrigger>
                  <TabsTrigger
                    className="flex items-center gap-1.5"
                    value="read"
                  >
                    Read
                    {(tabCounts?.read ?? 0) > 0 && (
                      <div className="h-5 w-5 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium">
                        {tabCounts?.read}
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

              <TabsContent
                className="mt-4 flex-1 overflow-hidden"
                value="unread"
              >
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
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
