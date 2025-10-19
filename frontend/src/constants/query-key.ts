/**
 * Centralized TanStack Query Keys
 *
 * Benefits:
 * - Single source of truth for all query keys
 * - Type-safe with TypeScript `as const`
 * - Hierarchical structure for easy invalidation
 * - Autocomplete support in IDE
 *
 * Usage:
 * @example
 * // In hooks
 * queryKey: QUERY_KEY.chat.messages(conversationId)
 *
 * // Invalidate all chat queries
 * queryClient.invalidateQueries({ queryKey: QUERY_KEY.chat.all })
 *
 * // Invalidate specific messages
 * queryClient.invalidateQueries({ queryKey: QUERY_KEY.chat.messages(conversationId) })
 */

export const QUERY_KEY = {
  /**
   * Authentication query keys
   */
  auth: {
    all: ['auth'] as const,
    currentUser: () => [...QUERY_KEY.auth.all, 'currentUser'] as const,
  },

  /**
   * Product service query keys
   */
  products: {
    all: ['products'] as const,
    lists: () => [...QUERY_KEY.products.all, 'list'] as const,
    list: (limit?: string, cursor?: string) =>
      [...QUERY_KEY.products.lists(), { limit, cursor }] as const,
    details: () => [...QUERY_KEY.products.all, 'detail'] as const,
    detail: (id: string) => [...QUERY_KEY.products.details(), id] as const,
    health: () => [...QUERY_KEY.products.all, 'health'] as const,
  },

  /**
   * Chat service query keys
   */
  chat: {
    all: ['chat'] as const,
    conversations: () => [...QUERY_KEY.chat.all, 'conversations'] as const,
    messages: (conversationId: string) =>
      [...QUERY_KEY.chat.all, 'messages', conversationId] as const,
    conversationDetails: (conversationId: string) =>
      [...QUERY_KEY.chat.all, 'conversation-details', conversationId] as const,
    conversationParticipants: (conversationId: string) =>
      [
        ...QUERY_KEY.chat.all,
        'conversation-participants',
        conversationId,
      ] as const,
    onlineUsers: () => [...QUERY_KEY.chat.all, 'online-users'] as const,
    ticket: (userId: string) =>
      [...QUERY_KEY.chat.all, 'ticket', userId] as const,
    typingIndicator: (conversationId: string, userId: string) =>
      [
        ...QUERY_KEY.chat.all,
        'typing-indicator',
        conversationId,
        userId,
      ] as const,
  },

  /**
   * Notification service query keys
   */
  notifications: {
    all: ['notifications'] as const,
    lists: () => [...QUERY_KEY.notifications.all, 'list'] as const,
    list: (limit: number, cursor?: string) =>
      [...QUERY_KEY.notifications.lists(), { limit, cursor }] as const,
    unreadLists: () => [...QUERY_KEY.notifications.all, 'unread-list'] as const,
    unreadList: (limit: number, cursor?: string) =>
      [...QUERY_KEY.notifications.unreadLists(), { limit, cursor }] as const,
    unreadCount: () =>
      [...QUERY_KEY.notifications.all, 'unread-count'] as const,
    tabCounts: () => [...QUERY_KEY.notifications.all, 'tab-counts'] as const,
  },

  /**
   * Cart and checkout query keys
   */
  cart: {
    all: ['cart'] as const,
    current: () => [...QUERY_KEY.cart.all, 'current'] as const,
    items: () => [...QUERY_KEY.cart.all, 'items'] as const,
    item: (itemId: string) => [...QUERY_KEY.cart.items(), itemId] as const,
    addItem: () => [...QUERY_KEY.cart.all, 'add-item'] as const,
    removeItem: () => [...QUERY_KEY.cart.all, 'remove-item'] as const,
    updateQuantity: () => [...QUERY_KEY.cart.all, 'update-quantity'] as const,
    clearCart: () => [...QUERY_KEY.cart.all, 'clear-cart'] as const,
  },

  /**
   * Checkout query keys
   */
  checkout: {
    all: ['checkout'] as const,
    session: () => [...QUERY_KEY.checkout.all, 'session'] as const,
    placeOrder: () => [...QUERY_KEY.checkout.all, 'place-order'] as const,
    orderHistory: () => [...QUERY_KEY.checkout.all, 'order-history'] as const,
    orderDetails: (orderId: string) =>
      [...QUERY_KEY.checkout.all, 'order-details', orderId] as const,
    shipping: {
      all: () => [...QUERY_KEY.checkout.all, 'shipping'] as const,
      options: () => [...QUERY_KEY.checkout.shipping.all(), 'options'] as const,
      calculate: () =>
        [...QUERY_KEY.checkout.shipping.all(), 'calculate'] as const,
    },
    payment: {
      all: () => [...QUERY_KEY.checkout.all, 'payment'] as const,
      methods: () => [...QUERY_KEY.checkout.payment.all(), 'methods'] as const,
      process: () => [...QUERY_KEY.checkout.payment.all(), 'process'] as const,
    },
  },

  /**
   * Address query keys
   */
  address: {
    all: ['address'] as const,
    lists: () => [...QUERY_KEY.address.all, 'list'] as const,
    list: (limit: number, cursor?: string) =>
      [...QUERY_KEY.address.lists(), { limit, cursor }] as const,
    details: () => [...QUERY_KEY.address.all, 'detail'] as const,
    detail: (id: string) => [...QUERY_KEY.address.details(), id] as const,
    default: () => [...QUERY_KEY.address.all, 'default'] as const,
  },
} as const
