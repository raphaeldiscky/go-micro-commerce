import { env } from '@/env'
import { getAccessToken } from '@/lib/api/client'
import type { ClientOptions, Client as SseClient } from 'graphql-sse'
import { createClient as createSseClient } from 'graphql-sse'
import type { Client as WsClient } from 'graphql-ws'
import { createClient as createWsClient } from 'graphql-ws'

let wsSubscriptionClient: WsClient | null = null
let sseSubscriptionClient: SseClient | null = null

/**
 * Create or get GraphQL WebSocket subscription client
 * Uses graphql-transport-ws protocol for real-time subscriptions
 */
export function getWsSubscriptionClient(): WsClient {
  if (!wsSubscriptionClient) {
    // Build WebSocket URL with token as query parameter
    // Browser WebSocket API doesn't support custom headers, so we use query params
    const baseUrl = env.VITE_GRAPHQL_SUBSCRIPTION_WS_URL
    const token = getAccessToken()
    const url = token ? `${baseUrl}?token=${token}` : baseUrl

    wsSubscriptionClient = createWsClient({
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

  return wsSubscriptionClient
}

/**
 * Create or get GraphQL SSE subscription client
 * Uses graphql-transport-ws protocol over SSE for real-time subscriptions
 * SSE is more reliable than WebSocket for unidirectional server->client streaming
 */
export function getSseSubscriptionClient(): SseClient {
  if (!sseSubscriptionClient) {
    const baseUrl = env.VITE_GRAPHQL_SUBSCRIPTION_SSE_URL
    const token = getAccessToken()

    const options: ClientOptions = {
      url: baseUrl,
      headers: () => {
        const currentToken = getAccessToken()
        return {
          ...(currentToken ? { Authorization: `Bearer ${currentToken}` } : {}),
        }
      },
      retryAttempts: 5,
      retry: async (retries) => {
        // Exponential backoff: 1s, 2s, 4s, 8s, 16s
        await new Promise((resolve) =>
          setTimeout(resolve, 1000 * Math.pow(2, retries)),
        )
      },
      onMessage: (message) => {
        console.log('📨 SSE Message Received:', {
          event: message.event,
          data: message.data,
          timestamp: new Date().toISOString(),
        })
      },
    }

    sseSubscriptionClient = createSseClient(options)

    console.log('✅ GraphQL SSE Client Created', {
      url: baseUrl,
      hasToken: !!token,
      timestamp: new Date().toISOString(),
    })
  }

  return sseSubscriptionClient
}

/**
 * Close all subscription client connections
 * Call this when logging out or when token changes
 */
export function closeSubscriptionClient(): void {
  if (wsSubscriptionClient) {
    wsSubscriptionClient.dispose()
    wsSubscriptionClient = null
  }
  if (sseSubscriptionClient) {
    sseSubscriptionClient.dispose()
    sseSubscriptionClient = null
  }
}

/**
 * Reset all subscription clients
 * Useful when token changes (e.g., after refresh)
 */
export function resetSubscriptionClient(): void {
  closeSubscriptionClient()
  // Next call to get*SubscriptionClient() will create new clients with updated token
}
