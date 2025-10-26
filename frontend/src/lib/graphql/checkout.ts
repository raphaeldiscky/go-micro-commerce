import { gql } from 'graphql-request'

/**
 * Query to fetch a checkout session by ID
 */
export const GET_CHECKOUT_SESSION_QUERY = gql`
  query GetCheckoutSession($id: UUID!) {
    getCheckoutSession(id: $id) {
      id
      idempotencyKey
      customerId
      courier {
        courierId
      }
      package {
        weightKg
        width
        height
        length
        unit
      }
      origin {
        city
        state
        postalCode
        countryCode
      }
      destination {
        city
        state
        postalCode
        countryCode
      }
      status
      paymentGateway
      currency
      items {
        id
        productId
        productName
        quantity
        unitPrice
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to create a checkout session from the user's cart
 */
export const CREATE_CHECKOUT_SESSION_MUTATION = gql`
  mutation CreateCheckoutSession($input: CreateCheckoutSessionInput!) {
    createCheckoutSession(input: $input) {
      id
      idempotencyKey
      customerId
      courier {
        courierId
      }
      package {
        weightKg
        width
        height
        length
        unit
      }
      origin {
        city
        state
        postalCode
        countryCode
      }
      destination {
        city
        state
        postalCode
        countryCode
      }
      status
      paymentGateway
      currency
      items {
        id
        productId
        productName
        quantity
        unitPrice
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to update a checkout session with address, carrier, or payment gateway
 */
export const UPDATE_CHECKOUT_SESSION_MUTATION = gql`
  mutation UpdateCheckoutSession(
    $sessionId: UUID!
    $input: UpdateCheckoutSessionInput!
  ) {
    updateCheckoutSession(sessionId: $sessionId, input: $input) {
      id
      idempotencyKey
      customerId
      courier {
        courierId
      }
      package {
        weightKg
        width
        height
        length
        unit
      }
      origin {
        city
        state
        postalCode
        countryCode
      }
      destination {
        city
        state
        postalCode
        countryCode
      }
      status
      paymentGateway
      currency
      items {
        id
        productId
        productName
        quantity
        unitPrice
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to place an order from a checkout session
 */
export const PLACE_ORDER_MUTATION = gql`
  mutation PlaceOrder($sessionId: UUID!, $input: PlaceOrderInput!) {
    placeOrder(sessionId: $sessionId, input: $input) {
      id
      idempotencyKey
      customerId
      courier {
        courierId
      }
      package {
        weightKg
        width
        height
        length
        unit
      }
      origin {
        city
        state
        postalCode
        countryCode
      }
      destination {
        city
        state
        postalCode
        countryCode
      }
      status
      paymentGateway
      currency
      items {
        id
        productId
        productName
        quantity
        unitPrice
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to cancel a checkout session
 */
export const CANCEL_CHECKOUT_SESSION_MUTATION = gql`
  mutation CancelCheckoutSession($sessionId: UUID!) {
    cancelCheckoutSession(sessionId: $sessionId) {
      id
      idempotencyKey
      customerId
      courier {
        courierId
      }
      package {
        weightKg
        width
        height
        length
        unit
      }
      origin {
        city
        state
        postalCode
        countryCode
      }
      destination {
        city
        state
        postalCode
        countryCode
      }
      status
      paymentGateway
      currency
      items {
        id
        productId
        productName
        quantity
        unitPrice
      }
      createdAt
      updatedAt
    }
  }
`
