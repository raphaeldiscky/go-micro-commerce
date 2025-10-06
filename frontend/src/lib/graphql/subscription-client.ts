import { env } from '@/env'
import { getAccessToken } from '@/lib/api/client'
import { createClient } from 'graphql-ws'
import type { Client } from 'graphql-ws'

let subscriptionClient: Client | null = null

/**
 * Create or get GraphQL WebSocket subscription client
 * Uses graphql-transport-ws protocol for real-time subscriptions
 */
export function getSubscriptionClient(): Client {
  if (!subscriptionClient) {
    subscriptionClient = createClient({
      url: env.VITE_GRAPHQL_SUBSCRIPTION_URL,
      connectionParams: () => {
        const token = getAccessToken()
        return {
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        }
      },
      retryAttempts: 5,
      shouldRetry: () => true,
      keepAlive: 10000, // 10 seconds
    })
  }

  return subscriptionClient
}

/**
 * Close the subscription client connection
 */
export function closeSubscriptionClient(): void {
  if (subscriptionClient) {
    subscriptionClient.dispose()
    subscriptionClient = null
  }
}
