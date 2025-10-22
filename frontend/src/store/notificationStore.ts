import { create } from 'zustand'

interface NotificationState {
  isDrawerOpen: boolean
  activeTab: 'all' | 'unread' | 'read'
}

interface NotificationActions {
  toggleDrawer: (open?: boolean) => void
  setActiveTab: (tab: 'all' | 'unread' | 'read') => void
}

type NotificationStore = NotificationState & NotificationActions

export const useNotificationStore = create<NotificationStore>((set) => ({
  isDrawerOpen: false,
  activeTab: 'all',

  toggleDrawer: (open) =>
    set((state) => ({
      isDrawerOpen: open !== undefined ? open : !state.isDrawerOpen,
    })),

  setActiveTab: (tab) => set({ activeTab: tab }),
}))

// Selectors for convenience
export const useIsDrawerOpen = () =>
  useNotificationStore((state) => state.isDrawerOpen)
export const useActiveTab = () =>
  useNotificationStore((state) => state.activeTab)
