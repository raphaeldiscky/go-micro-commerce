import { z } from 'zod'

// Customer address interface
export interface CustomerAddress {
  id: string
  customerId: string
  isDefault: boolean
  recipientName: string
  street: string
  apartment?: string // Unit, apartment, suite number
  city: string
  state: string
  postalCode: string
  country: string
  phone?: string
  instructions?: string // Delivery instructions
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
  email: z.string().email('Please enter a valid email address'),
  phone: z
    .string()
    .regex(/^\+?[\d\s\-()]*$/, 'Please enter a valid phone number')
    .optional(),
  avatar: z.string().optional(),
})

export const addressSchema = z.object({
  recipientName: z
    .string()
    .min(2, 'Recipient name must be at least 2 characters')
    .max(100, 'Recipient name must be less than 100 characters'),
  street: z
    .string()
    .min(5, 'Please enter a complete street address')
    .max(200, 'Street address is too long'),
  apartment: z.string().max(50, 'Apartment number is too long').optional(),
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
  country: z
    .string()
    .min(1, 'Country is required')
    .max(50, 'Country name is too long'),
  phone: z
    .string()
    .regex(/^\+?[\d\s\-()]*$/, 'Please enter a valid phone number')
    .max(20, 'Phone number is too long')
    .optional(),
  instructions: z
    .string()
    .max(500, 'Delivery instructions must be less than 500 characters')
    .optional(),
  isDefault: z.boolean().default(false),
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
