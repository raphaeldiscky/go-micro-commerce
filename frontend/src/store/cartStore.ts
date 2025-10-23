import { productApi } from '@/lib/api/product'
import {
  ADD_ITEM_TO_CART_MUTATION,
  GET_MY_CART_QUERY,
  REMOVE_ITEM_FROM_CART_MUTATION,
  SELECT_ITEM_FOR_CHECKOUT_MUTATION,
  UPDATE_ITEM_QUANTITY_MUTATION,
} from '@/lib/graphql/cart'
import type {
  AddItemToCartMutation,
  GetMyCartQuery,
  RemoveItemFromCartMutation,
  SelectItemForCheckoutMutation,
  UpdateItemQuantityMutation,
} from '@/lib/graphql/cart.generated'
import { graphClient } from '@/lib/graphql/client'
import { BatchGetProductsByIDsRequestSchema } from '@/proto/product/v1/product_pb'
import { create as createProto } from '@bufbuild/protobuf'
import type { Timestamp } from '@bufbuild/protobuf/wkt'
import { toast } from 'sonner'
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'
import { useShallow } from 'zustand/react/shallow'

// Extract types from GraphQL schema
type Cart = NonNullable<GetMyCartQuery['getMyCart']>
type CartItem = Cart['items'][number]

type Product = {
  id: string
  name: string
  price: number
  quantity: number
  version: number
  reservedQuantity: number
  createdAt?: Timestamp
  updatedAt?: Timestamp
}

// Enriched cart item with product data from Connect-RPC
export type EnrichedCartItem = CartItem & {
  product?: Product
}

const EMPTY_CART_ITEMS: Array<CartItem> = []

// Cart store state interface
export interface CartState {
  cart: Cart | null
  productsMap: Map<string, Product>
  isLoading: boolean
  isDrawerOpen: boolean
  lastFetchedAt: number | null
}

// Cart store actions interface
export interface CartActions {
  // Cart and item management
  fetchCart: (force?: boolean) => Promise<void>
  addItem: (productId: string, quantity?: number) => Promise<void>
  removeItem: (itemId: string) => Promise<void>
  updateQuantity: (itemId: string, quantity: number) => Promise<void>
  toggleSelection: (itemId: string) => Promise<void>
  selectAll: () => Promise<void>
  deselectAll: () => Promise<void>

  // Product management
  addProductToMap: (product: Product) => void
  getEnrichedCartItems: () => Array<EnrichedCartItem>

  // Drawer management
  openDrawer: () => void
  closeDrawer: () => void
  toggleDrawer: () => void

  // Utility selectors
  getTotalItemCount: () => number
  getSelectedItems: () => Array<EnrichedCartItem>
  getSelectedTotal: () => number
}

export type CartStore = CartState & CartActions

export const useCartStore = create<CartStore>()(
  devtools(
    persist(
      (set, get) => ({
        // Initial state
        cart: null,
        productsMap: new Map(),
        isLoading: false,
        isDrawerOpen: false,
        lastFetchedAt: null,

        // Cart and item management
        fetchCart: async (force = false) => {
          const state = get()

          // Smart caching: skip fetch if data is fresh (< 60 seconds old)
          // Can be bypassed with force parameter
          if (!force && state.lastFetchedAt) {
            const timeSinceLastFetch = Date.now() - state.lastFetchedAt
            if (timeSinceLastFetch < 60000) {
              // Data is fresh, skip fetch
              return
            }
          }

          set({ isLoading: true })

          try {
            const data =
              await graphClient.request<GetMyCartQuery>(GET_MY_CART_QUERY)

            // Extract product IDs from cart items
            const productIds =
              data.getMyCart?.items.map((item) => item.productId) || []

            // Fetch products if cart has items
            if (productIds.length > 0) {
              try {
                const request = createProto(
                  BatchGetProductsByIDsRequestSchema,
                  {
                    ids: productIds,
                  },
                )
                const productsResponse =
                  await productApi.batchGetProductsByIDs(request)

                const productsMap = new Map<string, Product>()
                for (const product of productsResponse.products) {
                  const processedProduct: Product = {
                    id: product.id,
                    name: product.name,
                    price: Number(product.price),
                    quantity: Number(product.quantity),
                    version: Number(product.version),
                    reservedQuantity: Number(product.reservedQuantity),
                  }
                  productsMap.set(product.id, processedProduct)
                }

                set({
                  cart: data.getMyCart,
                  productsMap,
                  isLoading: false,
                  lastFetchedAt: Date.now(),
                })
              } catch (productError) {
                console.error('Failed to fetch products:', productError)
                // Still set cart even if product fetch fails
                set({
                  cart: data.getMyCart,
                  isLoading: false,
                  lastFetchedAt: Date.now(),
                })
              }
            } else {
              set({
                cart: data.getMyCart,
                isLoading: false,
                lastFetchedAt: Date.now(),
              })
            }
          } catch (error) {
            set({ isLoading: false })
            console.error('Failed to fetch cart:', error)

            // If no cart exists, create an empty one locally
            set({ cart: null })
          }
        },

        addItem: async (productId: string, quantity = 1) => {
          set({ isLoading: true })

          try {
            const data = await graphClient.request<AddItemToCartMutation>(
              ADD_ITEM_TO_CART_MUTATION,
              {
                input: {
                  productId,
                  quantity,
                },
              },
            )

            set({ cart: data.addItemToCart, isLoading: false })
            toast.success('Added to cart')
          } catch (error) {
            set({ isLoading: false })
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to add item to cart'
            toast.error(errorMessage)
            throw error
          }
        },

        removeItem: async (itemId: string) => {
          set({ isLoading: true })

          try {
            const data = await graphClient.request<RemoveItemFromCartMutation>(
              REMOVE_ITEM_FROM_CART_MUTATION,
              { itemId },
            )

            set({ cart: data.removeItemFromCart, isLoading: false })
            toast.success('Removed from cart')
          } catch (error) {
            set({ isLoading: false })
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to remove item from cart'
            toast.error(errorMessage)
            throw error
          }
        },

        updateQuantity: async (itemId: string, quantity: number) => {
          if (quantity <= 0) {
            await get().removeItem(itemId)
            return
          }

          set({ isLoading: true })

          try {
            const data = await graphClient.request<UpdateItemQuantityMutation>(
              UPDATE_ITEM_QUANTITY_MUTATION,
              {
                itemId,
                input: { quantity },
              },
            )

            set({ cart: data.updateItemQuantity, isLoading: false })
          } catch (error) {
            set({ isLoading: false })
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to update quantity'
            toast.error(errorMessage)
            throw error
          }
        },

        toggleSelection: async (itemId: string) => {
          const state = get()
          const item = state.cart?.items.find((x) => x.id === itemId)

          if (!item) return

          try {
            const data =
              await graphClient.request<SelectItemForCheckoutMutation>(
                SELECT_ITEM_FOR_CHECKOUT_MUTATION,
                {
                  itemId,
                  input: { selected: !item.selectedForCheckout },
                },
              )

            set({ cart: data.selectItemForCheckout })
          } catch (error) {
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to update selection'
            toast.error(errorMessage)
            throw error
          }
        },

        selectAll: async () => {
          const state = get()

          if (!state.cart?.items) return

          try {
            // Update all items to selected
            for (const item of state.cart.items) {
              if (!item.selectedForCheckout) {
                const data =
                  await graphClient.request<SelectItemForCheckoutMutation>(
                    SELECT_ITEM_FOR_CHECKOUT_MUTATION,
                    {
                      itemId: item.id,
                      input: { selected: true },
                    },
                  )

                set({ cart: data.selectItemForCheckout })
              }
            }
          } catch (error) {
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to select all items'
            toast.error(errorMessage)
            throw error
          }
        },

        deselectAll: async () => {
          const state = get()

          if (!state.cart?.items) return

          try {
            // Update all items to deselected
            for (const item of state.cart.items) {
              if (item.selectedForCheckout) {
                const data =
                  await graphClient.request<SelectItemForCheckoutMutation>(
                    SELECT_ITEM_FOR_CHECKOUT_MUTATION,
                    {
                      itemId: item.id,
                      input: { selected: false },
                    },
                  )

                set({ cart: data.selectItemForCheckout })
              }
            }
          } catch (error) {
            const errorMessage =
              error instanceof Error
                ? error.message
                : 'Failed to deselect all items'
            toast.error(errorMessage)
            throw error
          }
        },

        // Product management
        addProductToMap: (product: Product) => {
          const currentMap = get().productsMap
          const existingProduct = currentMap.get(product.id)

          // Only update if product doesn't exist or has changed (different reference)
          if (!existingProduct || existingProduct !== product) {
            const updatedProductsMap = new Map(currentMap)
            updatedProductsMap.set(product.id, product)
            set({ productsMap: updatedProductsMap })
          }
        },

        getEnrichedCartItems: () => {
          const state = get()
          if (!state.cart?.items) return []

          return state.cart.items.map((item) => ({
            ...item,
            product: state.productsMap.get(item.productId),
          }))
        },

        // Drawer management
        openDrawer: () => set({ isDrawerOpen: true }),
        closeDrawer: () => set({ isDrawerOpen: false }),
        toggleDrawer: () =>
          set((state) => ({ isDrawerOpen: !state.isDrawerOpen })),

        // Utility selectors
        getTotalItemCount: () => {
          const state = get()
          return (state.cart?.items || []).reduce(
            (total, item) => total + item.quantity,
            0,
          )
        },

        getSelectedItems: () => {
          const enrichedItems = get().getEnrichedCartItems()
          return enrichedItems.filter((item) => item.selectedForCheckout)
        },

        getSelectedTotal: () => {
          const selectedItems = get().getSelectedItems()
          return selectedItems.reduce((total, item) => {
            if (item.product) {
              return total + item.product.price * item.quantity
            }
            return total
          }, 0)
        },
      }),
      {
        name: 'cart-store',
        partialize: (state) => ({
          cart: state.cart,
          productsMap: Array.from(state.productsMap.entries()),
          lastFetchedAt: state.lastFetchedAt,
        }),
        merge: (persistedState, currentState) => {
          // The persisted state has productsMap as an array of entries, not a Map
          const persisted = persistedState as {
            cart: Cart | null
            productsMap: Array<[string, Product]>
            lastFetchedAt: number | null
          }
          return {
            ...currentState,
            cart: persisted.cart ?? null,
            productsMap: new Map(persisted.productsMap),
            lastFetchedAt: persisted.lastFetchedAt ?? null,
          }
        },
      },
    ),
  ),
)

// Selectors for easier access
export const useCart = () => useCartStore((state) => state.cart)

// Simple selectors - return raw state, no transformations
export const useCartItems = () =>
  useCartStore((state) => state.cart?.items ?? EMPTY_CART_ITEMS)

export const useProductsMap = () => useCartStore((state) => state.productsMap)

// Multi-property selector with shallow comparison
export const useCartData = () =>
  useCartStore(
    useShallow((state) => ({
      items: state.cart?.items ?? EMPTY_CART_ITEMS,
      productsMap: state.productsMap,
    })),
  )

export const useCartItemCount = () =>
  useCartStore((state) =>
    (state.cart?.items ?? EMPTY_CART_ITEMS).reduce(
      (total, item) => total + item.quantity,
      0,
    ),
  )

export const useIsCartDrawerOpen = () =>
  useCartStore((state) => state.isDrawerOpen)
export const useIsCartLoading = () => useCartStore((state) => state.isLoading)
