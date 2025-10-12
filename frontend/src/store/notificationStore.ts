import { create } from 'zustand'

/**
 * Zustand store for notification UI state
 * Data management is handled by TanStack Query via hooks
 */

interface NotificationUIState {
  isDrawerOpen: boolean
  activeTab: 'all' | 'unread' | 'read'
}

interface NotificationUIActions {
  toggleDrawer: (open?: boolean) => void
  setActiveTab: (tab: 'all' | 'unread' | 'read') => void
}

type NotificationUIStore = NotificationUIState & NotificationUIActions

export const useNotificationStore = create<NotificationUIStore>((set) => ({
  // UI State
  isDrawerOpen: false,
  activeTab: 'all',

  // UI Actions
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
