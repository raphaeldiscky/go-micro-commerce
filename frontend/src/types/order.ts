/**
 * Frontend-specific order types
 *
 * Note: Most order types are generated from GraphQL schema in @/types/__generated__/graphql
 * This file only contains frontend-specific types that don't exist in GraphQL
 */

import type { Order, OrderStatus } from '@/types/__generated__/graphql'
import type { Decimal } from 'decimal.js'

/**
 * Order filters for frontend UI
 * Used by OrderStore and OrderFilters component
 */
export interface OrderFilters {
  status?: OrderStatus
  search?: string // For future search functionality
  dateFrom?: string // For future date filtering
  dateTo?: string // For future date filtering
  minAmount?: number // For future amount filtering
  maxAmount?: number // For future amount filtering
}

/**
 * Order display model for frontend UI
 * Maps GraphQL Order type with Decimal instances for precision calculations
 */
export interface OrderDisplay
  extends Omit<
    Order,
    'shippingCost' | 'subtotal' | 'totalPrice' | 'totalTax' | 'totalDiscount'
  > {
  shippingCost: Decimal
  subtotal: Decimal
  totalPrice: Decimal
  totalTax: Decimal
  totalDiscount: Decimal
}

/**
 * Shipping carrier information (frontend-specific)
 */
export interface ShippingCarrier {
  id: string
  name: string
  type: 'standard' | 'express' | 'overnight'
  estimatedDays: {
    min: number
    max: number
  }
}

/**
 * Checkout session types (temporary until payment service GraphQL is ready)
 */
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
  paymentMethod: string
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
