import { generateMockNotifications } from '@/lib/mock/notifications'
import type { NotificationStore, NotificationTab } from '@/types/notification'
import { create } from 'zustand'

export const useNotificationStore = create<NotificationStore>((set, get) => ({
  // State
  notifications: generateMockNotifications(),
  isDrawerOpen: false,
  activeTab: 'all',
  localReadIds: new Set<string>(),

  // Actions
  setNotifications: (notifications) => set({ notifications }),

  addNotification: (notification) =>
    set((state) => ({
      notifications: [notification, ...state.notifications],
    })),

  markAsRead: (id) =>
    set((state) => {
      const updatedNotifications = state.notifications.map((notif) =>
        notif.id === id
          ? {
              ...notif,
              isRead: true,
              readAt: new Date().toISOString(),
            }
          : notif,
      )

      const newLocalReadIds = new Set(state.localReadIds)
      newLocalReadIds.add(id)

      return {
        notifications: updatedNotifications,
        localReadIds: newLocalReadIds,
      }
    }),

  markAllAsRead: () =>
    set((state) => {
      const now = new Date().toISOString()
      const updatedNotifications = state.notifications.map((notif) =>
        !notif.isRead
          ? {
              ...notif,
              isRead: true,
              readAt: now,
            }
          : notif,
      )

      const newLocalReadIds = new Set(state.localReadIds)
      state.notifications
        .filter((n) => !n.isRead)
        .forEach((n) => newLocalReadIds.add(n.id))

      return {
        notifications: updatedNotifications,
        localReadIds: newLocalReadIds,
      }
    }),

  toggleDrawer: (open) =>
    set((state) => {
      const newIsOpen = open !== undefined ? open : !state.isDrawerOpen

      // Reset local read state when closing
      if (!newIsOpen) {
        return {
          isDrawerOpen: newIsOpen,
          localReadIds: new Set<string>(),
        }
      }

      return { isDrawerOpen: newIsOpen }
    }),

  setActiveTab: (tab) => set({ activeTab: tab }),

  getFilteredNotifications: (tab: NotificationTab) => {
    const state = get()

    switch (tab) {
      case 'all':
        return state.notifications
      case 'unread':
        return state.notifications.filter((n) => !n.isRead)
      case 'read':
        return state.notifications.filter((n) => n.isRead)
      default:
        return state.notifications
    }
  },

  getUnreadCount: () => {
    const state = get()
    return state.notifications.filter((n) => !n.isRead).length
  },

  getTabCount: (tab: NotificationTab) => {
    const state = get()

    switch (tab) {
      case 'all':
        return state.notifications.length
      case 'unread': {
        // Count unread, excluding locally marked as read
        return state.notifications.filter(
          (n) => !n.isRead && !state.localReadIds.has(n.id),
        ).length
      }
      case 'read': {
        // Count read plus locally marked as read
        return (
          state.notifications.filter((n) => n.isRead).length +
          state.localReadIds.size
        )
      }
      default:
        return 0
    }
  },

  resetLocalReadState: () => set({ localReadIds: new Set<string>() }),
}))

// Selectors for convenience
export const useNotifications = () =>
  useNotificationStore((state) => state.notifications)
export const useIsDrawerOpen = () =>
  useNotificationStore((state) => state.isDrawerOpen)
export const useActiveTab = () =>
  useNotificationStore((state) => state.activeTab)
export const useUnreadCount = () =>
  useNotificationStore((state) => state.getUnreadCount())
export const useLocalReadIds = () =>
  useNotificationStore((state) => state.localReadIds)
