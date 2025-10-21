import { fetchMockOrders } from '@/lib/mock/orders'
import type { OrderFilters, OrderTransactionsResponse } from '@/types/order'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface OrderState {
  orders: OrderTransactionsResponse
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
  appendOrders: (response: OrderTransactionsResponse) => void
  initialize: () => void
}

type OrderStore = OrderState & OrderActions

// Default empty state
const defaultOrdersState: OrderTransactionsResponse = {
  orders: [],
  pagination: {
    hasNextPage: false,
    hasPreviousPage: false,
  },
}

const defaultFilters: OrderFilters = {}

export const useOrderStore = create<OrderStore>()(
  persist(
    (set, get) => ({
      // State
      orders: defaultOrdersState,
      filters: defaultFilters,
      isLoading: false,
      error: null,
      hasInitialized: false,

      // Actions
      fetchOrders: async (cursor?: string, filters?: OrderFilters) => {
        set({ isLoading: true, error: null })

        try {
          // Use mock data service
          const mockResponse = await fetchMockOrders(
            cursor,
            filters || get().filters,
            20,
          )

          if (cursor) {
            // Append orders for pagination (load more)
            const currentOrders = get().orders.orders
            const newOrders = mockResponse.orders.filter(
              (order) =>
                !currentOrders.some((existing) => existing.id === order.id),
            )

            set({
              orders: {
                orders: [...currentOrders, ...newOrders],
                pagination: mockResponse.pagination,
              },
              isLoading: false,
            })
          } else {
            // Replace orders for new search/filter
            set({
              orders: mockResponse,
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
        })
      },

      appendOrders: (response: OrderTransactionsResponse) => {
        const currentOrders = get().orders.orders
        const newOrders = response.orders.filter(
          (order) =>
            !currentOrders.some((existing) => existing.id === order.id),
        )

        set({
          orders: {
            orders: [...currentOrders, ...newOrders],
            pagination: response.pagination,
          },
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
export const useOrders = () => useOrderStore((state) => state.orders.orders)
export const useOrdersPagination = () =>
  useOrderStore((state) => state.orders.pagination)
export const useOrderFilters = () => useOrderStore((state) => state.filters)
export const useOrdersLoading = () => useOrderStore((state) => state.isLoading)
export const useOrdersError = () => useOrderStore((state) => state.error)
export const useOrdersInitialized = () =>
  useOrderStore((state) => state.hasInitialized)
