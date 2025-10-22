import { z } from 'zod'

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
    .max(200, 'Street address is too long')
    .optional(),
  city: z
    .string()
    .min(2, 'City must be at least 2 characters')
    .max(100, 'City name is too long'),
  state: z
    .string()
    .min(1, 'State/Province is required')
    .max(50, 'State name is too long')
    .optional(),
  postalCode: z
    .string()
    .min(1, 'Postal code is required')
    .max(20, 'Postal code is too long'),
  countryCode: z
    .string()
    .min(1, 'Country code is required')
    .max(2, 'Country code name is too long'),
  latitude: z.number().min(-90).max(90).optional(),
  longitude: z.number().min(-180).max(180).optional(),
  note: z
    .string()
    .max(500, 'Delivery instructions must be less than 500 characters')
    .optional(),
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

// Interface types for form data
export interface ProfileUpdateRequest extends ProfileFormValues {}
export interface PasswordChangeRequest extends PasswordFormValues {}
