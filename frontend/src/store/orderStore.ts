import { graphClient } from '@/lib/graphql/client'
import { LIST_MY_ORDERS_QUERY } from '@/lib/graphql/order'
import type { ListMyOrdersQuery } from '@/lib/graphql/order.generated'
import type {
  Order,
  OrderStatus,
  PageInfo,
} from '@/types/__generated__/graphql'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface OrderFilters {
  status?: OrderStatus
  search?: string
}

interface OrderState {
  orders: Array<Order>
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

const defaultOrdersState: Array<Order> = []

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
      fetchOrders: async (cursor?: string) => {
        set({ isLoading: true, error: null })

        try {
          // Fetch orders from GraphQL
          const data = await graphClient.request<ListMyOrdersQuery>(
            LIST_MY_ORDERS_QUERY,
            { limit: 20, cursor },
          )

          const { edges, pageInfo } = data.listMyOrders

          if (cursor) {
            const currentOrders = get().orders
            const newOrders = edges
              .filter(
                (order) =>
                  !currentOrders.some(
                    (existing) => existing.id === order.node.id,
                  ),
              )
              .map(({ node }) => node)

            set({
              orders: [...currentOrders, ...newOrders],
              pagination: pageInfo,
              isLoading: false,
            })
          } else {
            set({
              orders: edges.map(({ node }) => node),
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
