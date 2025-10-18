import { env } from '@/env'
import { getAccessToken } from '@/lib/api/client'
import { createClient as createSseClient } from 'graphql-sse'
import type { ClientOptions, Client as SseClient } from 'graphql-sse'
import { createClient as createWsClient } from 'graphql-ws'
import type { Client as WsClient } from 'graphql-ws'

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
        console.log('🔑 SSE Client - Getting auth token', {
          hasToken: !!currentToken,
          tokenLength: currentToken?.length || 0,
          timestamp: new Date().toISOString(),
        })
        return {
          ...(currentToken ? { Authorization: `Bearer ${currentToken}` } : {}),
        }
      },
      retryAttempts: 5,
      retry: async (retries) => {
        console.log('🔄 SSE Client - Retrying connection', {
          attempt: retries + 1,
          maxAttempts: 5,
          timestamp: new Date().toISOString(),
        })
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
  console.log('🔌 Closing subscription clients...', {
    hasWsClient: wsSubscriptionClient !== null,
    hasSseClient: sseSubscriptionClient !== null,
    timestamp: new Date().toISOString(),
  })

  if (wsSubscriptionClient) {
    wsSubscriptionClient.dispose()
    wsSubscriptionClient = null
    console.log('✅ WebSocket client disposed')
  }
  if (sseSubscriptionClient) {
    sseSubscriptionClient.dispose()
    sseSubscriptionClient = null
    console.log('✅ SSE client disposed')
  }
}

/**
 * Reset all subscription clients
 * Useful when token changes (e.g., after refresh)
 */
export function resetSubscriptionClient(): void {
  console.log('🔄 Resetting subscription clients...', {
    hadWsClient: wsSubscriptionClient !== null,
    hadSseClient: sseSubscriptionClient !== null,
    timestamp: new Date().toISOString(),
  })
  closeSubscriptionClient()
  console.log('✅ Subscription clients reset complete')
  // Next call to get*SubscriptionClient() will create new clients with updated token
}

/**
 * Check if any subscription clients are currently active
 * Useful for debugging and monitoring connection state
 */
export function hasActiveClients(): boolean {
  return wsSubscriptionClient !== null || sseSubscriptionClient !== null
}
