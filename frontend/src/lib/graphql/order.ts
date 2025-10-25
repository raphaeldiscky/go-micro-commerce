import { gql } from 'graphql-request'

/**
 * List authenticated user's orders with cursor pagination
 */
export const LIST_MY_ORDERS_QUERY = gql`
  query ListMyOrders($limit: Int!, $cursor: String) {
    listMyOrders(limit: $limit, cursor: $cursor) {
      edges {
        node {
          id
          idempotencyKey
          customerId
          status
          currency
          paymentGateway
          shippingCost
          subtotal
          totalPrice
          totalTax
          totalDiscount
          items {
            id
            orderId
            productId
            quantity
            unitPrice
            totalPrice
            taxRate
            totalTax
            totalDiscount
            createdAt
            updatedAt
          }
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
 * Create a new order using saga workflow
 */
export const CREATE_ORDER_MUTATION = gql`
  mutation CreateOrder($input: CreateOrderInput!) {
    createOrder(input: $input) {
      id
      idempotencyKey
      customerId
      status
      currency
      paymentGateway
      shippingCost
      subtotal
      totalPrice
      totalTax
      totalDiscount
      items {
        id
        orderId
        productId
        quantity
        unitPrice
        totalPrice
        taxRate
        totalTax
        totalDiscount
        createdAt
        updatedAt
      }
      createdAt
      updatedAt
    }
  }
`
