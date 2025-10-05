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
