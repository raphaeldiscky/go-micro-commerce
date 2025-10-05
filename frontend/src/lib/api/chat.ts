import { apiRequest } from './client'
import type { ApiSuccessResponse } from './types'
import { ApiError } from './types'

/**
 * WebSocket and Real-time Event Types
 * These are used for WebSocket connections and real-time features
 * Historical data queries use GraphQL (see @/lib/graphql/chat.ts)
 */

export interface ChatTicketData {
  expires_at: string
  node_address: string
  ticket: string
  user_id: string
  user_type: string
}

export interface ChatTicketResponse {
  data: ChatTicketData
  message: string
}

export interface PresenceUpdate {
  last_seen?: string
  status: 'away' | 'busy' | 'offline' | 'online'
}

export interface SendMessageRequest {
  content: string
  message_type?: 'file' | 'image' | 'text'
  reply_to_id?: string
}

export interface TypingIndicator {
  is_typing: boolean
  timestamp: string
  user_id: string
  username: string
}

// REST Message type (for WebSocket sendMessage response)
export interface MessageResponse {
  content: string
  conversation_id: string
  created_at: string
  id: string
  message_type: 'file' | 'image' | 'system' | 'text'
  sender_id: string
  updated_at: string
}

/**
 * WebSocket and Real-time API Functions
 * For historical data queries, use GraphQL hooks from @/hooks/chat/*
 */

export async function checkChatHealth(): Promise<{
  status: string
  timestamp: string
}> {
  const response = await apiRequest<{ status: string; timestamp: string }>(
    '/chats/v1/nodes/health',
  )
  return response
}

export async function getChatTicket(userId: string): Promise<ChatTicketData> {
  const response = await apiRequest<ApiSuccessResponse<ChatTicketData>>(
    '/chats/v1/connect',
    {
      body: JSON.stringify({
        user_id: userId,
      }),
      method: 'POST',
    },
  )

  if (!response.data.ticket) {
    throw new ApiError('Invalid response: missing ticket')
  }

  if (!response.data.node_address) {
    throw new ApiError('Invalid response: missing node_address')
  }

  return response.data
}

// Note: Real-time events (typing, receipts, presence) are now sent via WebSocket
// See ChatWebSocketContext and related hooks for WebSocket message sending

export async function sendMessage(
  conversationId: string,
  message: SendMessageRequest,
): Promise<MessageResponse> {
  const response = await apiRequest<ApiSuccessResponse<MessageResponse>>(
    `/chats/v1/${conversationId}/messages`,
    {
      body: JSON.stringify(message),
      method: 'POST',
    },
  )
  return response.data
}

export async function validateTicket(ticket: string): Promise<boolean> {
  try {
    await apiRequest('/chats/v1/validate-ticket', {
      body: JSON.stringify({ ticket }),
      method: 'POST',
    })
    return true
  } catch {
    return false
  }
}
