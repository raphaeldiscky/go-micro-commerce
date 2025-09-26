import { apiRequest } from './client'
import type { ApiPaginatedResponse, ApiSuccessResponse } from './types'
import { ApiError } from './types'

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

export interface Conversation {
  created_at: string
  description?: string
  id: string
  last_message?: {
    content: string
    sender_name: string
    timestamp: string
  }
  name: string
  participant_count?: number
  priority?: number
  status?: string
  subject: string
  type?: 'channel' | 'direct' | 'group'
  unread_count: number
  updated_at: string
}

export interface ConversationDetails {
  created_at: string
  created_by: string
  description?: string
  id: string
  name: string
  settings: {
    allow_member_invite: boolean
    is_public: boolean
    message_retention_days?: number
  }
  type: 'channel' | 'direct' | 'group'
  updated_at: string
}

export interface JoinConversationResponse {
  data: {
    conversation_id: string
    joined_at: string
  }
  message: string
}

export interface Message {
  content: string
  conversation_id: string
  created_at: string
  delivery_status: 'delivered' | 'read' | 'sent'
  id: string
  message_type: 'file' | 'image' | 'system' | 'text'
  reply_to_id?: string
  sender_id: string
  sender_name?: string
  updated_at: string
}

export interface MessageReceipt {
  message_id: string
  receipt_type: 'delivered' | 'read'
  timestamp: string
  user_id: string
}

export interface Participant {
  conversation_id: string
  id: string
  is_active: boolean
  joined_at: string
  role: string
  user_id: string
  user_type: string
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

export async function checkChatHealth(): Promise<{
  status: string
  timestamp: string
}> {
  const response = await apiRequest<{ status: string; timestamp: string }>(
    '/chats/v1/nodes/health',
  )
  return response
}

export async function createConversation(data: {
  description?: string
  name: string
  participant_ids?: Array<string>
  type: 'channel' | 'direct' | 'group'
}): Promise<ConversationDetails> {
  const response = await apiRequest<ApiSuccessResponse<ConversationDetails>>(
    '/chats/v1/conversations',
    {
      body: JSON.stringify(data),
      method: 'POST',
    },
  )
  return response.data
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

export async function getConversationDetails(
  conversationId: string,
): Promise<ConversationDetails> {
  const response = await apiRequest<ApiSuccessResponse<ConversationDetails>>(
    `/chats/v1/${conversationId}`,
  )
  return response.data
}

export async function getConversationMessages(
  conversationId: string,
  page: number = 1,
  limit: number = 50,
): Promise<{ messages: Array<Message>; hasMore: boolean; totalPages: number }> {
  const response = await apiRequest<ApiPaginatedResponse<Message>>(
    `/chats/v1/${conversationId}/messages?page=${page}&size=${limit}`,
  )

  return {
    messages: response.data,
    hasMore: response.pagination.page < response.pagination.total_page,
    totalPages: response.pagination.total_page,
  }
}

export async function getConversationParticipants(
  conversationId: string,
): Promise<Array<Participant>> {
  const response = await apiRequest<ApiSuccessResponse<Array<Participant>>>(
    `/chats/v1/${conversationId}/participants`,
  )
  return response.data
}

export async function getConversations(): Promise<Array<Conversation>> {
  const response = await apiRequest<ApiSuccessResponse<Array<Conversation>>>(
    '/chats/v1/conversations',
  )
  return response.data
}

export async function getOnlineUsers(): Promise<Array<string>> {
  const response = await apiRequest<ApiSuccessResponse<Array<string>>>(
    '/chats/v1/users/online',
  )
  return response.data
}

export async function joinConversation(
  conversationId: string,
): Promise<{ conversation_id: string; joined_at: string }> {
  const response = await apiRequest<JoinConversationResponse>(
    `/chats/v1/${conversationId}/join`,
    {
      method: 'POST',
    },
  )
  return response.data
}

export async function sendDeliveryReceipt(
  conversationId: string,
  messageId: string,
): Promise<void> {
  await apiRequest(`/chats/v1/${conversationId}/delivery-receipt`, {
    body: JSON.stringify({ message_id: messageId }),
    method: 'POST',
  })
}

export async function sendMessage(
  conversationId: string,
  message: SendMessageRequest,
): Promise<Message> {
  const response = await apiRequest<ApiSuccessResponse<Message>>(
    `/chats/v1/${conversationId}/messages`,
    {
      body: JSON.stringify(message),
      method: 'POST',
    },
  )
  return response.data
}

export async function sendReadReceipt(
  conversationId: string,
  messageId: string,
): Promise<void> {
  await apiRequest(`/chats/v1/${conversationId}/read-receipt`, {
    body: JSON.stringify({ message_id: messageId }),
    method: 'POST',
  })
}

export async function sendTypingIndicator(
  conversationId: string,
  isTyping: boolean,
): Promise<void> {
  await apiRequest(`/chats/v1/${conversationId}/typing`, {
    body: JSON.stringify({ is_typing: isTyping }),
    method: 'POST',
  })
}

export async function updatePresence(presence: PresenceUpdate): Promise<void> {
  await apiRequest('/chats/v1/presence', {
    body: JSON.stringify(presence),
    method: 'PUT',
  })
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
