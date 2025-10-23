// NOTE: Cart, CartItem, and CheckoutSession types are generated from GraphQL schema
// Import them from: @/lib/graphql/cart.generated.ts
//
// The types below are UI-SPECIFIC types used for display and mock data only.
// They are not fetched from the backend GraphQL API.
//
// GraphQL provides PaymentMethod and PaymentGateway as enums (CARD, STRIPE, MOCK),
// but the UI needs rich objects with display information (names, icons, descriptions).

// Shipping options - UI-specific type for displaying shipping methods
export interface ShippingOptionUI {
  id: string
  name: string
  description?: string
  price: number
  estimatedDays: {
    min: number
    max: number
  }
  isActive: boolean
}

// Payment methods - UI-specific type for displaying payment options
export interface PaymentMethodUI {
  id: string
  name: string
  type: 'card' | 'bank_transfer' | 'ewallet' | 'cod'
  icon?: string
  isActive: boolean
  description?: string
  supportedGateways: Array<PaymentGatewayUI>
}

export interface PaymentGatewayUI {
  id: string
  name: string
  type: 'stripe' | 'paypal'
}

// Order summary for checkout - calculated on frontend
export interface OrderSummary {
  subtotal: number
  shipping: number
  discount: number
  total: number
}
