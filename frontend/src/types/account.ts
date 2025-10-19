import { z } from 'zod'

// Customer address interface
export interface CustomerAddress {
  id: string
  userId: string
  isDefault: boolean
  receiverName: string
  addressLine1: string
  addressLine2: string
  city: string
  state: string
  postalCode: string
  countryCode: string
  latitude: number
  longitude: number
  note: string
  createdAt: string
  updatedAt: string
}
export const profileSchema = z.object({
  firstName: z
    .string()
    .min(2, 'First name must be at least 2 characters')
    .max(50, 'First name must be less than 50 characters'),
  lastName: z
    .string()
    .min(2, 'Last name must be at least 2 characters')
    .max(50, 'Last name must be less than 50 characters'),
  email: z.email('Please enter a valid email address'),
  phone: z
    .string()
    .regex(/^\+?[\d\s\-()]*$/, 'Please enter a valid phone number')
    .optional(),
  avatar: z.string().optional(),
})

export const addressSchema = z.object({
  receiverName: z
    .string()
    .min(2, 'Receiver name must be at least 2 characters')
    .max(100, 'Receiver name must be less than 100 characters'),
  addressLine1: z
    .string()
    .min(5, 'Please enter a complete street address')
    .max(200, 'Street address is too long'),
  addressLine2: z
    .string()
    .min(5, 'Please enter a complete street address')
    .max(200, 'Street address is too long'),
  city: z
    .string()
    .min(2, 'City must be at least 2 characters')
    .max(100, 'City name is too long'),
  state: z
    .string()
    .min(1, 'State/Province is required')
    .max(50, 'State name is too long'),
  postalCode: z
    .string()
    .min(1, 'Postal code is required')
    .max(20, 'Postal code is too long'),
  countryCode: z
    .string()
    .min(1, 'Country code is required')
    .max(2, 'Country code name is too long'),
  note: z
    .string()
    .max(500, 'Delivery instructions must be less than 500 characters'),
  isDefault: z.boolean(),
})

export const passwordSchema = z
  .object({
    currentPassword: z
      .string()
      .min(1, 'Current password is required')
      .max(128, 'Current password is too long'),
    newPassword: z
      .string()
      .min(8, 'Password must be at least 8 characters')
      .max(128, 'Password is too long')
      .regex(
        /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
        'Password must contain at least one uppercase letter, one lowercase letter, and one number',
      ),
    confirmPassword: z
      .string()
      .min(1, 'Please confirm your new password')
      .max(128, 'Password is too long'),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  })

// Form value types inferred from schemas
export type ProfileFormValues = z.infer<typeof profileSchema>
export type AddressFormValues = z.infer<typeof addressSchema>
export type PasswordFormValues = z.infer<typeof passwordSchema>

// Interface types for store
export interface ProfileUpdateRequest extends ProfileFormValues {}
export interface PasswordChangeRequest extends PasswordFormValues {}
export interface AddressRequest extends AddressFormValues {}

// Account statistics
export interface AccountStats {
  totalOrders: number
  totalSpent: number
  averageOrderValue: number
  lastOrderDate?: string
  memberSince: string
}

// Account store state
export interface AccountState {
  user: any // From auth store
  addresses: Array<CustomerAddress>
  stats: AccountStats | null
  isLoading: boolean
  isUpdating: boolean
  error: string | null
}

// Account store actions
export interface AccountActions {
  // Profile management
  updateProfile: (data: ProfileUpdateRequest) => Promise<void>
  changePassword: (data: PasswordChangeRequest) => Promise<void>

  // Address management
  loadAddresses: () => Promise<void>
  addAddress: (data: AddressRequest) => Promise<void>
  updateAddress: (id: string, data: AddressRequest) => Promise<void>
  deleteAddress: (id: string) => Promise<void>
  setDefaultAddress: (id: string) => Promise<void>

  // Stats
  loadStats: () => Promise<void>

  // State management
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
}

// Account store type
export type AccountStore = AccountState & AccountActions
