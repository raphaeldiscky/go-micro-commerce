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
          checkoutSessionId
          customerId
          status
          currency
          paymentGateway
          courier {
            courierId
          }
          destination {
            city
            state
            postalCode
            countryCode
          }
          origin {
            city
            state
            postalCode
            countryCode
          }
          package {
            weightKg
            length
            height
            width
            unit
          }
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
      checkoutSessionId
      customerId
      status
      currency
      paymentGateway
      courier {
        courierId
      }
      destination {
        city
        state
        postalCode
        countryCode
      }
      origin {
        city
        state
        postalCode
        countryCode
      }
      package {
        weightKg
        length
        height
        width
        unit
      }
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

/**
 * Place an order from a checkout session with payment intent creation
 */
export const PLACE_ORDER_MUTATION = gql`
  mutation PlaceOrder($input: PlaceOrderInput!) {
    placeOrder(input: $input) {
      order {
        id
        idempotencyKey
        checkoutSessionId
        customerId
        status
        currency
        paymentGateway
        courier {
          courierId
        }
        package {
          weightKg
          length
          height
          width
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
      paymentMetadata {
        paymentId
        paymentGateway
        gatewayTransactionId
        gatewayMetadata
        amount
        currency
      }
    }
  }
`
