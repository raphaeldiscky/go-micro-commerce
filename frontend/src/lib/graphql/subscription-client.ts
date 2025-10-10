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
    // Build WebSocket URL with token as query parameter
    // Browser WebSocket API doesn't support custom headers, so we use query params
    const baseUrl = env.VITE_GRAPHQL_SUBSCRIPTION_URL
    const token = getAccessToken()
    const url = token ? `${baseUrl}?token=${token}` : baseUrl

    subscriptionClient = createClient({
      url,
      connectionParams: () => {
        // Still send in connectionParams for graphql-transport-ws protocol compatibility
        const currentToken = getAccessToken()
        return {
          ...(currentToken ? { Authorization: `Bearer ${currentToken}` } : {}),
        }
      },
      retryAttempts: 5,
      shouldRetry: () => true,
      keepAlive: 10000, // 10 seconds
      on: {
        connected: () => {
          console.log('✅ GraphQL WebSocket Connected', {
            url,
            timestamp: new Date().toISOString(),
          })
        },
        closed: (event) => {
          console.log('❌ GraphQL WebSocket Closed', {
            code: (event as CloseEvent).code,
            reason: (event as CloseEvent).reason,
            timestamp: new Date().toISOString(),
          })
        },
        error: (error) => {
          console.error('❌ GraphQL WebSocket Error:', {
            error,
            timestamp: new Date().toISOString(),
          })
        },
      },
    })
  }

  return subscriptionClient
}

/**
 * Close the subscription client connection
 * Call this when logging out or when token changes
 */
export function closeSubscriptionClient(): void {
  if (subscriptionClient) {
    subscriptionClient.dispose()
    subscriptionClient = null
  }
}

/**
 * Reset the subscription client
 * Useful when token changes (e.g., after refresh)
 */
export function resetSubscriptionClient(): void {
  closeSubscriptionClient()
  // Next call to getSubscriptionClient() will create a new client with updated token
}
