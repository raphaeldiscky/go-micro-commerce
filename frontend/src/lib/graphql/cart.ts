import { gql } from 'graphql-request'

/**
 * Query to fetch the authenticated user's active cart with all items
 */
export const GET_MY_CART_QUERY = gql`
  query GetMyCart {
    getMyCart {
      id
      customerId
      status
      items {
        id
        cartId
        productId
        quantity
        selectedForCheckout
        createdAt
        updatedAt
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Query to fetch a checkout session by ID
 */
export const GET_CHECKOUT_SESSION_QUERY = gql`
  query GetCheckoutSession($id: UUID!) {
    getCheckoutSession(id: $id) {
      id
      idempotencyKey
      customerId
      addressId
      carrierId
      status
      paymentGateway
      paymentMethod
      currency
      items {
        id
        productId
        quantity
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to add an item to the user's active cart
 */
export const ADD_ITEM_TO_CART_MUTATION = gql`
  mutation AddItemToCart($input: AddCartItemInput!) {
    addItemToCart(input: $input) {
      id
      customerId
      status
      items {
        id
        cartId
        productId
        quantity
        selectedForCheckout
        createdAt
        updatedAt
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to remove an item from the user's cart
 */
export const REMOVE_ITEM_FROM_CART_MUTATION = gql`
  mutation RemoveItemFromCart($itemId: UUID!) {
    removeItemFromCart(itemId: $itemId) {
      id
      customerId
      status
      items {
        id
        cartId
        productId
        quantity
        selectedForCheckout
        createdAt
        updatedAt
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to update the quantity of a cart item
 */
export const UPDATE_ITEM_QUANTITY_MUTATION = gql`
  mutation UpdateItemQuantity(
    $itemId: UUID!
    $input: UpdateCartItemQuantityInput!
  ) {
    updateItemQuantity(itemId: $itemId, input: $input) {
      id
      customerId
      status
      items {
        id
        cartId
        productId
        quantity
        selectedForCheckout
        createdAt
        updatedAt
      }
      createdAt
      updatedAt
    }
  }
`

/**
 * Mutation to select or deselect an item for checkout
 */
export const SELECT_ITEM_FOR_CHECKOUT_MUTATION = gql`
  mutation SelectItemForCheckout(
    $itemId: UUID!
    $input: SelectItemForCheckoutInput!
  ) {
    selectItemForCheckout(itemId: $itemId, input: $input) {
      id
      customerId
      status
      items {
        id
        cartId
        productId
        quantity
        selectedForCheckout
        createdAt
        updatedAt
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
      addressId
      carrierId
      status
      paymentGateway
      paymentMethod
      currency
      items {
        id
        productId
        quantity
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
      addressId
      carrierId
      status
      paymentGateway
      paymentMethod
      currency
      items {
        id
        productId
        quantity
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
      addressId
      carrierId
      status
      paymentGateway
      paymentMethod
      currency
      items {
        id
        productId
        quantity
      }
      createdAt
      updatedAt
    }
  }
`
