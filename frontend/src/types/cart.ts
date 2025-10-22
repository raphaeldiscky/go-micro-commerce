// Matches 'carts' table
export interface Cart {
  id: string
  customer_id: string
  created_at: string
  updated_at: string
}

// Matches 'cart_items' table
export interface CartItem {
  id: string
  cart_id: string
  product_id: string
  quantity: number
  selected_for_checkout: boolean
  added_at: string
  product: MockProduct // JOIN with products data
}

// Matches 'checkout_sessions' table
export interface CheckoutSession {
  id: string
  customer_id: string
  cart_id: string
  status: 'pending' | 'placed' | 'failed' | 'expired'
  created_at: string
  updated_at: string
}

// Mock product data structure (extends existing Product type)
export interface MockProduct {
  id: string
  name: string
  description?: string
  price: number
  quantity: number
  reservedQuantity: number
  image?: string
  category?: string
  sku?: string
  version: number
  createdAt?: Date | null
  updatedAt?: Date | null
}

// Shipping options
export interface ShippingOption {
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

// Payment methods
export interface PaymentMethod {
  id: string
  name: string
  type: 'card' | 'bank_transfer' | 'ewallet' | 'cod'
  icon?: string
  isActive: boolean
  description?: string
  supportedGateways: Array<PaymentGateway>
}

export interface PaymentGateway {
  id: string
  name: string
  type: 'stripe' | 'paypal'
}

// Order summary for checkout
export interface OrderSummary {
  subtotal: number
  shipping: number
  discount: number
  total: number
}

// Pagination for cart items
export interface CartItemsPagination {
  cursor?: string
  hasMore: boolean
  totalCount: number
}

// Add to cart request
export interface AddToCartRequest {
  productId: string
  quantity: number
}

// Cart error types
export type CartError =
  | 'OUT_OF_STOCK'
  | 'INVALID_QUANTITY'
  | 'ITEM_NOT_FOUND'
  | 'SHIPPING_UNAVAILABLE'
  | 'PAYMENT_FAILED'
  | 'CHECKOUT_FAILED'
