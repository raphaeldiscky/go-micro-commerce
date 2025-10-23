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
