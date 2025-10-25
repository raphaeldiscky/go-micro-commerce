// @TODO: change later from fulfillment-service
export interface ShippingCarrier {
  id: string
  name: string
  type: 'standard' | 'express' | 'overnight'
  estimatedDays: {
    min: number
    max: number
  }
}

export interface CreateCheckoutSessionRequest {
  paymentId: string
}

export interface CreateCheckoutSessionResponse {
  sessionId: string
  checkoutUrl: string
}

export interface PaymentDetails {
  paymentId: string
  orderId: string
  amount: number
  currency: string
  paymentStatus: 'pending' | 'processing' | 'completed' | 'failed' | 'timeout'
  paymentGateway: string
  paymentDeadline: string
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
