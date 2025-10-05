/**
 * WebSocket and Real-time Event Types
 * These are used for WebSocket connections and real-time features
 * Historical data queries use GraphQL (see @/lib/graphql/chat.ts)
 *
 * NOTE: WebSocket connection uses JWT directly (not ticket-based)
 * - Request node address via GraphQL: requestChatConnection
 * - Connect with: ws://${nodeAddress}/v1/ws?token=${jwt}&conversation_id=${id}
 */

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

export interface ApiErrorResponse {
  errors?: Array<{ field: string; message: string }>
  message: string
}

export class ApiError extends Error {
  public readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}
