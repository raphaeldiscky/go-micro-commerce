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
 * queryKey: queryKeys.chat.messages(conversationId)
 *
 * // Invalidate all chat queries
 * queryClient.invalidateQueries({ queryKey: queryKeys.chat.all })
 *
 * // Invalidate specific messages
 * queryClient.invalidateQueries({ queryKey: queryKeys.chat.messages(conversationId) })
 */

export const queryKeys = {
  /**
   * Authentication query keys
   */
  auth: {
    all: ['auth'] as const,
    currentUser: () => [...queryKeys.auth.all, 'currentUser'] as const,
  },

  /**
   * Product service query keys
   */
  products: {
    all: ['products'] as const,
    lists: () => [...queryKeys.products.all, 'list'] as const,
    list: (limit?: string, cursor?: string) =>
      [...queryKeys.products.lists(), { limit, cursor }] as const,
    details: () => [...queryKeys.products.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.products.details(), id] as const,
    health: () => [...queryKeys.products.all, 'health'] as const,
  },

  /**
   * Chat service query keys
   */
  chat: {
    all: ['chat'] as const,
    conversations: () => [...queryKeys.chat.all, 'conversations'] as const,
    messages: (conversationId: string) =>
      [...queryKeys.chat.all, 'messages', conversationId] as const,
    conversationDetails: (conversationId: string) =>
      [...queryKeys.chat.all, 'conversation-details', conversationId] as const,
    conversationParticipants: (conversationId: string) =>
      [
        ...queryKeys.chat.all,
        'conversation-participants',
        conversationId,
      ] as const,
    onlineUsers: () => [...queryKeys.chat.all, 'online-users'] as const,
    ticket: (userId: string) =>
      [...queryKeys.chat.all, 'ticket', userId] as const,
  },
} as const
