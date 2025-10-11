/**
 * Notification types matching the database schema
 */

export type NotificationType =
  | 'NEW_MESSAGE'
  | 'NEW_PRODUCT'
  | 'ORDER_UPDATE'
  | 'PAYMENT_SUCCESS'
  | 'SYSTEM_ALERT'
  | 'ORDER_CONFIRMED'
  | 'ORDER_SHIPPED'
  | 'ORDER_DELIVERED'
  | 'ORDER_CANCELLED'

export type NotificationTab = 'all' | 'unread' | 'read'

export interface Notification {
  id: string
  userId: string
  type: NotificationType
  title: string
  message: string
  metadata: Record<string, unknown> | null
  isRead: boolean
  readAt: string | null
  createdAt: string
  updatedAt: string
}

export interface NotificationState {
  notifications: Array<Notification>
  isDrawerOpen: boolean
  activeTab: NotificationTab
  localReadIds: Set<string>
}

export interface NotificationActions {
  setNotifications: (notifications: Array<Notification>) => void
  addNotification: (notification: Notification) => void
  markAsRead: (id: string) => void
  markAllAsRead: () => void
  toggleDrawer: (open?: boolean) => void
  setActiveTab: (tab: NotificationTab) => void
  getFilteredNotifications: (tab: NotificationTab) => Array<Notification>
  getUnreadCount: () => number
  getTabCount: (tab: NotificationTab) => number
  resetLocalReadState: () => void
}

export type NotificationStore = NotificationState & NotificationActions
