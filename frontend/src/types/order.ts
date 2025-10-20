// Order types for checkout and order management

export interface PlaceOrderRequest {
  idempotencyKey: string
  items: Array<{
    productId: string
    quantity: number
  }>
  paymentMethod: string // e.g. "card", "bank_transfer", "ewallet"
  paymentGateway: string // e.g. "stripe", "midtrans", "xendit"
  currency: string // e.g. "IDR", "USD"
  shipping: {
    fromAddress: {
      city: string
      state: string
      postalCode: string
      country: string
    }
    toAddress: {
      city: string
      state: string
      postalCode: string
      country: string
    }
    dimensions: {
      width: number
      height: number
      length: number
      unit: string
    }
    weightKg: number
    carrierId: string
  }
}

export interface PlaceOrderResponse {
  orderId: string
  paymentId: string
  status: OrderStatus
  paymentStatus: 'pending' | 'processing' | 'completed' | 'failed' | 'timeout'
  paymentGateway: string
  totalAmount: number
  currency: string
  paymentDeadline: string // ISO 8601 date string
  createdAt: string
}

export interface CreateCheckoutSessionRequest {
  paymentId: string
}

export interface CreateCheckoutSessionResponse {
  sessionId: string
  checkoutUrl: string // Stripe hosted checkout URL
}

export interface PaymentDetails {
  paymentId: string
  orderId: string
  amount: number
  currency: string
  paymentStatus: 'pending' | 'processing' | 'completed' | 'failed' | 'timeout'
  paymentMethod: string
  paymentGateway: string
  paymentDeadline: string // ISO 8601 date string
  order: {
    orderId: string
    items: Array<{
      productId: string
      productName: string
      quantity: number
      price: number
    }>
    subtotal: number
    shippingCost: number
    total: number
    shippingAddress: {
      receiverName: string
      addressLine1: string
      addressLine2?: string
      city: string
      state: string
      postalCode: string
      countryCode: string
    }
  }
  createdAt: string
  updatedAt: string
}

export type OrderStatus =
  | 'pending'
  | 'processing'
  | 'payment_pending'
  | 'payment_expired'
  | 'paid'
  | 'delivered'
  | 'completed'
  | 'failed'
  | 'canceled'

export interface OrderDetails {
  orderId: string
  customerId: string
  items: Array<{
    productId: string
    productName: string
    quantity: number
    price: number
  }>
  subtotal: number
  shippingCost: number
  total: number
  currency: string
  status: OrderStatus
  paymentMethod: string
  paymentGateway: string
  shippingAddress: {
    receiverName: string
    addressLine1: string
    addressLine2?: string
    city: string
    state: string
    postalCode: string
    countryCode: string
  }
  createdAt: string
  updatedAt: string
}

// Shipping carrier information
export interface ShippingCarrier {
  id: string
  name: string
  type: 'standard' | 'express' | 'overnight'
  estimatedDays: {
    min: number
    max: number
  }
}
