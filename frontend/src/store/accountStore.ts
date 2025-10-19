import type {
  AccountStats,
  AccountStore,
  AddressRequest,
  CustomerAddress,
  PasswordChangeRequest,
  ProfileUpdateRequest,
} from '@/types/account'
import { toast } from 'sonner'
import { create } from 'zustand'

const generateId = () =>
  `address-${Date.now()}-${Math.random().toString(36).substring(7)}`

// Mock data for demonstration
const mockStats: AccountStats = {
  totalOrders: 12,
  totalSpent: 1245.99,
  averageOrderValue: 103.83,
  lastOrderDate: '2024-01-15T10:30:00Z',
  memberSince: '2023-06-15T14:22:00Z',
}

const mockAddresses: Array<CustomerAddress> = [
  {
    id: generateId(),
    userId: 'user-123',
    isDefault: true,
    receiverName: 'John Doe',
    addressLine1: '123 Main Street',
    addressLine2: 'Apt 4B',
    city: 'New York',
    state: 'NY',
    postalCode: '10001',
    countryCode: 'US',
    note: 'Ring doorbell twice',
    createdAt: '2023-06-15T14:22:00Z',
    updatedAt: '2023-06-15T14:22:00Z',
  },
  {
    id: generateId(),
    userId: 'user-123',
    isDefault: false,
    receiverName: 'John Doe',
    addressLine1: '456 Business Ave',
    addressLine2: '',
    city: 'New York',
    state: 'NY',
    postalCode: '10002',
    countryCode: 'US',
    note: '',
    createdAt: '2023-08-20T09:15:00Z',
    updatedAt: '2023-08-20T09:15:00Z',
  },
]

export const useAccountStore = create<AccountStore>()((set, get) => ({
  // Initial state
  user: null,
  addresses: [],
  stats: null,
  isLoading: false,
  isUpdating: false,
  error: null,

  // Profile management
  updateProfile: async (data: ProfileUpdateRequest) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000))

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
      await new Promise((resolve) => setTimeout(resolve, 1500))

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

  // Address management
  loadAddresses: async () => {
    set({ isLoading: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 500))

      // Use mock data for now
      set({ addresses: mockAddresses, isLoading: false })
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to load addresses'
      set({ isLoading: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  addAddress: async (data: AddressRequest) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 800))

      const newAddress: CustomerAddress = {
        id: generateId(),
        userId: 'user-123', // Would come from auth store
        ...data,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }

      // If this is set as default, unset other defaults
      const currentAddresses = get().addresses
      let updatedAddresses = currentAddresses

      if (data.isDefault) {
        updatedAddresses = currentAddresses.map((addr) => ({
          ...addr,
          isDefault: false,
        }))
      }

      updatedAddresses = [...updatedAddresses, newAddress]

      set({ addresses: updatedAddresses, isUpdating: false })
      toast.success('Address added successfully')
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to add address'
      set({ isUpdating: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  updateAddress: async (id: string, data: AddressRequest) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 800))

      const currentAddresses = get().addresses

      // If this is set as default, unset other defaults
      let updatedAddresses = currentAddresses
      if (data.isDefault) {
        updatedAddresses = currentAddresses.map((addr) => ({
          ...addr,
          isDefault: false,
        }))
      }

      updatedAddresses = updatedAddresses.map((addr) =>
        addr.id === id
          ? {
              ...addr,
              ...data,
              updatedAt: new Date().toISOString(),
            }
          : addr,
      )

      set({ addresses: updatedAddresses, isUpdating: false })
      toast.success('Address updated successfully')
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to update address'
      set({ isUpdating: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  deleteAddress: async (id: string) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 500))

      const currentAddresses = get().addresses
      const addressToDelete = currentAddresses.find((addr) => addr.id === id)

      if (addressToDelete?.isDefault) {
        throw new Error('Cannot delete default address')
      }

      const updatedAddresses = currentAddresses.filter((addr) => addr.id !== id)

      set({ addresses: updatedAddresses, isUpdating: false })
      toast.success('Address deleted successfully')
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to delete address'
      set({ isUpdating: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  setDefaultAddress: async (id: string) => {
    set({ isUpdating: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 600))

      const currentAddresses = get().addresses
      const updatedAddresses = currentAddresses.map((addr) => ({
        ...addr,
        isDefault: addr.id === id,
        updatedAt: new Date().toISOString(),
      }))

      set({ addresses: updatedAddresses, isUpdating: false })
      toast.success('Default address updated')
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Failed to set default address'
      set({ isUpdating: false, error: errorMessage })
      toast.error(errorMessage)
    }
  },

  // Stats
  loadStats: async () => {
    set({ isLoading: true, error: null })

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 400))

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
export const useAccountAddresses = () =>
  useAccountStore((state) => state.addresses)
export const useAccountLoading = () =>
  useAccountStore((state) => state.isLoading)
export const useAccountUpdating = () =>
  useAccountStore((state) => state.isUpdating)
export const useAccountError = () => useAccountStore((state) => state.error)
