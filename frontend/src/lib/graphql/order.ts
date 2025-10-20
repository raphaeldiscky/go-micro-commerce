import { gql } from 'graphql-request'

/**
 * Create order mutation
 */
export const CREATE_ORDER_MUTATION = gql`
  mutation CreateOrder($input: PlaceOrderInput!) {
    createOrder(input: $input) {
      orderId
      paymentId
      status
      paymentStatus
      paymentGateway
      totalAmount
      currency
      paymentDeadline
      createdAt
    }
  }
`

/**
 * Get order details query
 */
export const GET_ORDER_QUERY = gql`
  query GetOrder($orderId: ID!) {
    getOrder(orderId: $orderId) {
      orderId
      customerId
      items {
        productId
        productName
        quantity
        price
      }
      subtotal
      shippingCost
      total
      currency
      status
      paymentMethod
      paymentGateway
      shippingAddress {
        receiverName
        addressLine1
        addressLine2
        city
        state
        postalCode
        countryCode
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Create checkout session mutation
 */
export const CREATE_CHECKOUT_SESSION_MUTATION = gql`
  mutation CreateCheckoutSession($input: CreateCheckoutSessionInput!) {
    createCheckoutSession(input: $input) {
      sessionId
      checkoutUrl
    }
  }
`

/**
 * Get payment details query
 */
export const GET_PAYMENT_DETAILS_QUERY = gql`
  query GetPaymentDetails($paymentId: ID!) {
    getPaymentDetails(paymentId: $paymentId) {
      paymentId
      orderId
      amount
      currency
      status
      paymentMethod
      paymentGateway
      paymentDeadline
      createdAt
      updatedAt
    }
  }
`
