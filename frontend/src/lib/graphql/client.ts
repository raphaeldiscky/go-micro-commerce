import { env } from '@/env'
import { getAccessToken } from '@/lib/api/client'
import { GraphQLClient } from 'graphql-request'

/**
 * Single GraphQL client for all operations (public and authenticated)
 *
 * Authorization flow:
 * - Public operations (login, register): Work without token
 * - Protected operations (me, chat, etc): Require token
 * - Authorization enforced by @requiresAuth directives in subgraphs
 * - Client conditionally sends JWT when available
 */
export const graphClient = new GraphQLClient(env.VITE_GRAPHQL_GATEWAY_URL, {
  credentials: 'include', // Send HTTP-only cookies with requests
  headers: () => {
    const token = getAccessToken()

    return {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  },
})

// Alias for backward compatibility
export const graphAuthClient = graphClient
