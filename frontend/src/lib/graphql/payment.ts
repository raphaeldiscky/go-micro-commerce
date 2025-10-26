import { gql } from 'graphql-request'

// Query to fetch payment details by order ID
export const GET_PAYMENT_BY_ORDER_ID = gql`
  query GetPaymentByOrderId($orderId: UUID!) {
    getPaymentByOrderId(orderId: $orderId) {
      id
      orderId
      amount
      currency
      status
      paymentGateway
      clientSecret
      expiresAt
      createdAt
      updatedAt
      completedAt
      failedAt
    }
  }
`
