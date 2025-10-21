import { graphClient } from '@/lib/graphql/client'
import { LIST_MY_ORDERS_QUERY } from '@/lib/graphql/order'
import type { ListMyOrdersQuery } from '@/lib/graphql/order.generated'
import { parseDecimal } from '@/lib/utils/decimal'
import type { Order, PageInfo } from '@/types/__generated__/graphql'
import type { OrderDisplay, OrderFilters } from '@/types/order'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface OrderState {
  orders: Array<OrderDisplay>
  pagination: PageInfo
  filters: OrderFilters
  isLoading: boolean
  error: string | null
  hasInitialized: boolean
}

interface OrderActions {
  fetchOrders: (cursor?: string, filters?: OrderFilters) => Promise<void>
  setFilters: (filters: OrderFilters) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  resetOrders: () => void
  initialize: () => void
}

type OrderStore = OrderState & OrderActions

/**
 * Convert GraphQL Order (with Decimal strings) to OrderDisplay (with Decimal instances)
 */
function mapOrderToDisplay(order: Order): OrderDisplay {
  return {
    ...order,
    shippingCost: parseDecimal(order.shippingCost),
    subtotal: parseDecimal(order.subtotal),
    totalPrice: parseDecimal(order.totalPrice),
    totalTax: parseDecimal(order.totalTax),
    totalDiscount: parseDecimal(order.totalDiscount),
  }
}

// Default empty state
const defaultOrdersState: Array<OrderDisplay> = []

const defaultPagination: PageInfo = {
  hasNextPage: false,
  hasPreviousPage: false,
  startCursor: null,
  endCursor: null,
}

const defaultFilters: OrderFilters = {}

export const useOrderStore = create<OrderStore>()(
  persist(
    (set, get) => ({
      // State
      orders: defaultOrdersState,
      pagination: defaultPagination,
      filters: defaultFilters,
      isLoading: false,
      error: null,
      hasInitialized: false,

      // Actions
      fetchOrders: async (cursor?: string, filters?: OrderFilters) => {
        set({ isLoading: true, error: null })

        try {
          // Fetch orders from GraphQL
          const data = await graphClient.request<ListMyOrdersQuery>(
            LIST_MY_ORDERS_QUERY,
            { limit: 20, cursor },
          )

          const { edges, pageInfo } = data.listMyOrders

          // Map GraphQL orders to display format with Decimal instances
          const orders = edges.map((edge) => mapOrderToDisplay(edge.node))

          if (cursor) {
            // Append orders for pagination (load more)
            const currentOrders = get().orders
            const newOrders = orders.filter(
              (order) =>
                !currentOrders.some((existing) => existing.id === order.id),
            )

            set({
              orders: [...currentOrders, ...newOrders],
              pagination: pageInfo,
              isLoading: false,
            })
          } else {
            // Replace orders for new search/filter
            set({
              orders,
              pagination: pageInfo,
              isLoading: false,
            })
          }
        } catch (error) {
          set({
            error:
              error instanceof Error ? error.message : 'Failed to fetch orders',
            isLoading: false,
          })
        }
      },

      setFilters: (filters: OrderFilters) => {
        const currentFilters = get().filters
        const newFilters = { ...currentFilters, ...filters }
        set({ filters: newFilters })

        // Refetch orders with new filters (reset pagination)
        get().fetchOrders(undefined, newFilters)
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading })
      },

      setError: (error: string | null) => {
        set({ error })
      },

      clearError: () => {
        set({ error: null })
      },

      resetOrders: () => {
        set({
          orders: defaultOrdersState,
          pagination: defaultPagination,
        })
      },

      initialize: () => {
        if (!get().hasInitialized) {
          set({ hasInitialized: true })
          // Load initial orders
          get().fetchOrders(undefined, get().filters)
        }
      },
    }),
    {
      name: 'order-store',
      partialize: (state) => ({
        // Only persist filters, not the orders data or loading states
        filters: state.filters,
      }),
    },
  ),
)

// Selectors for easier access
export const useOrders = () => useOrderStore((state) => state.orders)
export const useOrdersPagination = () =>
  useOrderStore((state) => state.pagination)
export const useOrderFilters = () => useOrderStore((state) => state.filters)
export const useOrdersLoading = () => useOrderStore((state) => state.isLoading)
export const useOrdersError = () => useOrderStore((state) => state.error)
export const useOrdersInitialized = () =>
  useOrderStore((state) => state.hasInitialized)
