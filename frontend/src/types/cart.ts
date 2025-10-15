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
  type: 'credit_card' | 'debit_card' | 'bank_transfer' | 'ewallet' | 'cod'
  icon?: string
  isActive: boolean
  description?: string
}

// Order summary for checkout
export interface OrderSummary {
  subtotal: number
  shipping: number
  discount: number
  total: number
}

// Checkout form data
export interface CheckoutFormData {
  orderNote?: string
  shippingMethod?: string
  paymentMethod?: string
}

// Cart store state interface
export interface CartState {
  cart: Cart | null
  items: Array<CartItem>
  checkoutSession: CheckoutSession | null
  isLoading: boolean
  isDrawerOpen: boolean
  isCheckoutLoading: boolean
  checkoutData: CheckoutFormData
  selectedShippingOption: ShippingOption | null
  selectedPaymentMethod: PaymentMethod | null
}

// Cart store actions interface
export interface CartActions {
  // Cart and item management
  initializeCart: (customerId: string) => void
  addItem: (product: MockProduct, quantity?: number) => void
  removeItem: (itemId: string) => void
  updateQuantity: (itemId: string, quantity: number) => void
  toggleSelection: (itemId: string) => void
  selectAll: () => void
  deselectAll: () => void
  clearCart: () => void

  // Drawer management
  openDrawer: () => void
  closeDrawer: () => void
  toggleDrawer: () => void

  // Checkout flow
  startCheckout: (navigateToCheckout?: (checkoutId: string) => void) => void
  setShippingMethod: (method: ShippingOption) => void
  setPaymentMethod: (method: PaymentMethod) => void
  setOrderNote: (note: string) => void
  placeOrder: () => Promise<{
    success: boolean
    orderId?: string
    error?: string
  }>

  // Utility selectors
  getTotalItemCount: () => number
  getSelectedItems: () => Array<CartItem>
  getSelectedTotal: () => number
  getSubtotal: () => number
  getOrderSummary: () => OrderSummary
}

// Cart store type (state + actions)
export type CartStore = CartState & CartActions

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
