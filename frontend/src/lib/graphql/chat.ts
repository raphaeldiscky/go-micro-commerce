import { gql } from 'graphql-request'

/**
 * Get user's conversations
 */
export const CONVERSATIONS_QUERY = gql`
  query Conversations {
    conversations {
      id
      subject
      status
      priority
      participantCount
      createdAt
      updatedAt
      endedAt
    }
  }
`

/**
 * Get single conversation by ID
 */
export const CONVERSATION_QUERY = gql`
  query Conversation($id: ID!) {
    conversation(id: $id) {
      id
      subject
      status
      priority
      createdAt
      updatedAt
      endedAt
    }
  }
`

/**
 * Get conversation messages with cursor pagination
 */
export const CONVERSATION_MESSAGES_QUERY = gql`
  query ConversationMessages(
    $conversationId: ID!
    $first: Int
    $after: String
    $last: Int
    $before: String
  ) {
    conversationMessages(
      conversationId: $conversationId
      first: $first
      after: $after
      last: $last
      before: $before
    ) {
      edges {
        cursor
        node {
          id
          conversationId
          senderId
          content
          messageType
          isSystem
          createdAt
        }
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
    }
  }
`

/**
 * Get conversation participants
 */
export const CONVERSATION_PARTICIPANTS_QUERY = gql`
  query ConversationParticipants($conversationId: ID!) {
    conversationParticipants(conversationId: $conversationId) {
      id
      conversationId
      userId
      userType
      role
      joinedAt
      leftAt
      isActive
    }
  }
`

/**
 * Get online users
 */
export const ONLINE_USERS_QUERY = gql`
  query OnlineUsers {
    onlineUsers {
      id
      email
      firstName
      lastName
      isActive
      onlineStatus {
        isOnline
        lastSeen
      }
    }
  }
`

/**
 * Create conversation mutation
 */
export const CREATE_CONVERSATION_MUTATION = gql`
  mutation CreateConversation($input: CreateConversationInput!) {
    createConversation(input: $input) {
      id
      subject
      status
      priority
      createdAt
      updatedAt
      endedAt
    }
  }
`

/**
 * Join conversation mutation
 */
export const JOIN_CONVERSATION_MUTATION = gql`
  mutation JoinConversation($input: JoinConversationInput!) {
    joinConversation(input: $input) {
      id
      conversationId
      userId
      userType
      role
      joinedAt
      leftAt
      isActive
    }
  }
`

/**
 * Request chat connection (returns WebSocket node address)
 */
export const REQUEST_CHAT_CONNECTION_MUTATION = gql`
  mutation RequestChatConnection {
    requestChatConnection {
      nodeAddress
      userId
      userType
    }
  }
`

/**
 * Subscribe to conversation events (messages, typing, receipts)
 */
export const CONVERSATION_EVENTS_SUBSCRIPTION = gql`
  subscription ConversationEvents($conversationId: ID!) {
    conversationEvents(conversationId: $conversationId) {
      __typename
      ... on NewMessage {
        __typename
        id
        conversationId
        senderId
        content
        messageType
        isSystem
        createdAt
      }
      ... on TypingIndicator {
        __typename
        userId
        conversationId
        isTyping
        timestamp
      }
      ... on DeliveryReceipt {
        __typename
        messageId
        conversationId
        recipientId
        deliveredAt
      }
      ... on ReadReceipt {
        __typename
        messageId
        conversationId
        readerId
        readAt
      }
    }
  }
`

/**
 * Subscribe to user events (presence updates)
 */
export const USER_EVENTS_SUBSCRIPTION = gql`
  subscription UserEvents {
    userEvents {
      __typename
      ... on PresenceUpdate {
        __typename
        userId
        status
        lastSeen
      }
    }
  }
`

/**
 * Send a message to a conversation
 */
export const SEND_MESSAGE_MUTATION = gql`
  mutation SendMessage($input: SendMessageInput!) {
    sendMessage(input: $input) {
      id
      conversationId
      senderId
      content
      messageType
      isSystem
      createdAt
    }
  }
`

/**
 * Send delivery receipt for a message
 */
export const SEND_DELIVERY_RECEIPT_MUTATION = gql`
  mutation SendDeliveryReceipt($input: SendDeliveryReceiptInput!) {
    sendDeliveryReceipt(input: $input) {
      messageId
      conversationId
      recipientId
      deliveredAt
    }
  }
`

/**
 * Send read receipt for a message
 */
export const SEND_READ_RECEIPT_MUTATION = gql`
  mutation SendReadReceipt($input: SendReadReceiptInput!) {
    sendReadReceipt(input: $input) {
      messageId
      conversationId
      readerId
      readAt
    }
  }
`
