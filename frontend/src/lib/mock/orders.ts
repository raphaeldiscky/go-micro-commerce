import type {
  CursorPaginationInfo,
  OrderFilters,
  OrderStatus,
  OrderTransaction,
  OrderTransactionsResponse,
} from '@/types/order'

// Sample customer IDs
const CUSTOMER_IDS = [
  'e57dfd25-3c4f-4c41-8567-d8e6a413b62b',
  'a2b3c4d5-e6f7-4a8b-9c0d-1e2f3a4b5c6d',
  'f8e7d6c5-b4a3-4f2e-1d0c-9b8a7f6e5d4c',
  'c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f',
  '7f6e5d4c-3b2a-4f1e-d0c9-b8a7f6e5d4c3',
]

// Payment gateways
const PAYMENT_GATEWAYS = ['stripe', 'paypal', 'square', 'braintree', 'adyen']

// Payment methods
const PAYMENT_METHODS = [
  'credit_card',
  'debit_card',
  'paypal',
  'apple_pay',
  'google_pay',
  'bank_transfer',
]

// Currencies
const CURRENCIES = ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD']

// Order statuses distribution (weighted for realistic data)
const STATUS_WEIGHTS: Record<OrderStatus, number> = {
  pending: 15,
  processing: 20,
  payment_pending: 10,
  payment_expired: 5,
  paid: 25,
  shipped: 15,
  delivered: 10,
  completed: 30,
  failed: 5,
  canceled: 8,
}

function getRandomItem<T>(items: Array<T>): T {
  return items[Math.floor(Math.random() * items.length)]
}

function getRandomStatus(): OrderStatus {
  const totalWeight = Object.values(STATUS_WEIGHTS).reduce((a, b) => a + b, 0)
  let random = Math.random() * totalWeight

  for (const [status, weight] of Object.entries(STATUS_WEIGHTS)) {
    random -= weight
    if (random <= 0) {
      return status as OrderStatus
    }
  }

  return 'pending'
}

function generateRandomOrder(index: number): OrderTransaction {
  const now = new Date()
  const daysAgo = Math.floor(Math.random() * 90) // Orders from last 90 days
  const createdAt = new Date(now.getTime() - daysAgo * 24 * 60 * 60 * 1000)

  const subtotal = Math.floor(Math.random() * 500) + 50 // $50-$550
  const shippingCost = Math.floor(Math.random() * 30) + 10 // $10-$40
  const totalTax = Math.floor(subtotal * 0.08) // ~8% tax
  const totalDiscount = Math.floor(Math.random() * 50) // $0-$50 discount
  const totalPrice = subtotal + shippingCost + totalTax - totalDiscount

  return {
    id: `order-${index.toString().padStart(4, '0')}-${Math.random()
      .toString(36)
      .substr(2, 9)}`,
    idempotencyKey: `ik-${Math.random().toString(36).substr(2, 16)}`,
    customerId: getRandomItem(CUSTOMER_IDS),
    status: getRandomStatus(),
    paymentGateway: getRandomItem(PAYMENT_GATEWAYS),
    paymentMethod: getRandomItem(PAYMENT_METHODS),
    currency: getRandomItem(CURRENCIES),
    shippingCost,
    subtotal,
    totalTax,
    totalDiscount,
    totalPrice: Math.max(totalPrice, 0), // Ensure non-negative
    createdAt: createdAt.toISOString(),
    updatedAt: createdAt.toISOString(),
  }
}

function filterOrders(
  orders: Array<OrderTransaction>,
  filters: OrderFilters,
): Array<OrderTransaction> {
  return orders.filter((order) => {
    // Status filter
    if (filters.status && order.status !== filters.status) {
      return false
    }

    // Search filter (for future implementation)
    if (filters.search) {
      const searchLower = filters.search.toLowerCase()
      const searchableText =
        `${order.id} ${order.customerId} ${order.paymentGateway}`.toLowerCase()
      if (!searchableText.includes(searchLower)) {
        return false
      }
    }

    // Date range filter (for future implementation)
    if (filters.dateFrom) {
      const fromDate = new Date(filters.dateFrom)
      if (new Date(order.createdAt) < fromDate) {
        return false
      }
    }

    if (filters.dateTo) {
      const toDate = new Date(filters.dateTo)
      if (new Date(order.createdAt) > toDate) {
        return false
      }
    }

    // Amount range filter (for future implementation)
    if (filters.minAmount && order.totalPrice < filters.minAmount) {
      return false
    }

    if (filters.maxAmount && order.totalPrice > filters.maxAmount) {
      return false
    }

    return true
  })
}

function getCursorForOrder(order: OrderTransaction): string {
  return btoa(`${order.id}|${order.createdAt}`)
}

function getOrderFromCursor(
  cursor: string,
): { id: string; createdAt: string } | null {
  try {
    const decoded = atob(cursor)
    const [id, createdAt] = decoded.split('|')
    return { id, createdAt }
  } catch {
    return null
  }
}

function sortOrdersByDate(
  orders: Array<OrderTransaction>,
): Array<OrderTransaction> {
  return [...orders].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  )
}

export async function fetchMockOrders(
  cursor?: string,
  filters: OrderFilters = {},
  limit: number = 20,
): Promise<OrderTransactionsResponse> {
  // Simulate API delay
  await new Promise((resolve) => setTimeout(resolve, 300 + Math.random() * 500))

  // Generate mock data (cache it for consistency)
  if (!mockDataCache) {
    mockDataCache = Array.from({ length: 150 }, (_, i) =>
      generateRandomOrder(i + 1),
    )
  }

  const allOrders = sortOrdersByDate(mockDataCache)

  // Apply filters
  const filteredOrders = filterOrders(allOrders, filters)

  // Find starting point based on cursor
  let startIndex = 0
  if (cursor) {
    const cursorInfo = getOrderFromCursor(cursor)
    if (cursorInfo) {
      startIndex = filteredOrders.findIndex(
        (order) =>
          order.id === cursorInfo.id &&
          order.createdAt === cursorInfo.createdAt,
      )
      if (startIndex !== -1) {
        startIndex += 1 // Start after the cursor
      } else {
        startIndex = 0 // Invalid cursor, start from beginning
      }
    }
  }

  // Get page of orders
  const endIndex = Math.min(startIndex + limit, filteredOrders.length)
  const pageOrders = filteredOrders.slice(startIndex, endIndex)

  // Build pagination info
  const pagination: CursorPaginationInfo = {
    hasNextPage: endIndex < filteredOrders.length,
    hasPreviousPage: startIndex > 0,
    startCursor:
      pageOrders.length > 0 ? getCursorForOrder(pageOrders[0]) : undefined,
    endCursor:
      pageOrders.length > 0
        ? getCursorForOrder(pageOrders[pageOrders.length - 1])
        : undefined,
    nextCursor:
      endIndex < filteredOrders.length
        ? getCursorForOrder(filteredOrders[endIndex])
        : undefined,
  }

  return {
    orders: pageOrders,
    pagination,
  }
}

// Mock data cache
let mockDataCache: Array<OrderTransaction> | undefined

export function generateMockOrder(count: number = 1): Array<OrderTransaction> {
  return Array.from({ length: count }, (_, i) =>
    generateRandomOrder(Date.now() + i),
  )
}
