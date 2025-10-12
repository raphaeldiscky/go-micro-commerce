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
} as const
