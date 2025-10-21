import { gql } from 'graphql-request'

/**
 * List notifications with cursor pagination
 */
export const LIST_NOTIFICATIONS_QUERY = gql`
  query ListNotifications($limit: Int!, $cursor: String) {
    listNotifications(limit: $limit, cursor: $cursor) {
      edges {
        node {
          id
          userId
          type
          title
          message
          metadata
          isRead
          readAt
          createdAt
          updatedAt
        }
        cursor
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
 * List unread notifications with cursor pagination
 */
export const LIST_UNREAD_NOTIFICATIONS_QUERY = gql`
  query ListUnreadNotifications($limit: Int!, $cursor: String) {
    listUnreadNotifications(limit: $limit, cursor: $cursor) {
      edges {
        node {
          id
          userId
          type
          title
          message
          metadata
          isRead
          readAt
          createdAt
          updatedAt
        }
        cursor
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
 * Get unread notification count
 */
export const GET_UNREAD_COUNT_QUERY = gql`
  query GetUnreadCount {
    getUnreadCount {
      count
    }
  }
`

/**
 * Get notification counts for all tabs (all, unread, read)
 */
export const GET_TAB_COUNTS_QUERY = gql`
  query GetTabCounts {
    getTabCounts {
      all
      unread
      read
    }
  }
`

/**
 * Mark a notification as read
 */
export const MARK_AS_READ_MUTATION = gql`
  mutation MarkAsRead($id: UUID!) {
    markAsRead(id: $id) {
      id
      userId
      type
      title
      message
      metadata
      isRead
      readAt
      createdAt
      updatedAt
    }
  }
`

/**
 * Mark all notifications as read
 */
export const MARK_ALL_AS_READ_MUTATION = gql`
  mutation MarkAllAsRead {
    markAllAsRead
  }
`

/**
 * Subscribe to notification events (new, read, deleted)
 */
export const NOTIFICATION_EVENTS_SUBSCRIPTION = gql`
  subscription NotificationEvents {
    notificationEvents {
      __typename
      ... on NewNotification {
        __typename
        id
        userId
        type
        title
        message
        metadata
        isRead
        createdAt
      }
      ... on NotificationRead {
        __typename
        id
        userId
        readAt
      }
      ... on NotificationDeleted {
        __typename
        id
        userId
      }
    }
  }
`
