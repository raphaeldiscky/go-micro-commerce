import type {
  AccountStats,
  PasswordChangeRequest,
  ProfileUpdateRequest,
} from '@/schemas/account'
import { toast } from 'sonner'
import { create } from 'zustand'

// Mock data for demonstration (stats only - addresses use GraphQL now)
const mockStats: AccountStats = {
  totalOrders: 12,
  totalSpent: 1245.99,
  averageOrderValue: 103.83,
  lastOrderDate: '2024-01-15T10:30:00Z',
  memberSince: '2023-06-15T14:22:00Z',
}

interface AccountState {
  user: unknown // From auth store
  stats: AccountStats | null
  isLoading: boolean
  isUpdating: boolean
  error: string | null
}

interface AccountActions {
  // Profile management
  updateProfile: (data: ProfileUpdateRequest) => Promise<void>
  changePassword: (data: PasswordChangeRequest) => Promise<void>

  // Stats
  loadStats: () => Promise<void>

  // State management
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

type AccountStore = AccountState & AccountActions

export const useAccountStore = create<AccountStore>()((set) => ({
  // Initial state
  user: null,
  stats: null,
  isLoading: false,
  isUpdating: false,
  error: null,

  // Profile management
  updateProfile: async (data: ProfileUpdateRequest) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => {
        setTimeout(resolve, 1000)
      })

      // In a real app, this would call the API
      console.log('Updating profile:', data)

      set({ isUpdating: false })
      toast.success('Profile updated successfully')
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to update profile'
      set({ isUpdating: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  changePassword: async (data: PasswordChangeRequest) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => {
        setTimeout(resolve, 1500)
      })

      // In a real app, this would call the API
      console.log('Changing password:', data)

      set({ isUpdating: false })
      toast.success('Password changed successfully')
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to change password'
      set({ isUpdating: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  // Stats
  loadStats: async () => {
    set({ isLoading: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => {
        setTimeout(resolve, 400)
      })

      // Use mock data for now
      set({ stats: mockStats, isLoading: false })
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : 'Failed to load account statistics'
      set({ isLoading: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  // State management
  setLoading: (loading: boolean) => set({ isLoading: loading }),
  setError: (error: string | null) => set({ error }),
}))

// Selectors for easier access
export const useAccountStats = () => useAccountStore((state) => state.stats)
export const useAccountLoading = () =>
  useAccountStore((state) => state.isLoading)
export const useAccountUpdating = () =>
  useAccountStore((state) => state.isUpdating)
export const useAccountError = () => useAccountStore((state) => state.error)
