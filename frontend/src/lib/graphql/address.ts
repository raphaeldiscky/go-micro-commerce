import { gql } from 'graphql-request'

/**
 * List addresses query with cursor pagination
 */
export const LIST_ADDRESSES_QUERY = gql`
  query ListAddresses($limit: Int!, $cursor: String) {
    listAddresses(limit: $limit, cursor: $cursor) {
      edges {
        node {
          id
          userId
          receiverName
          addressLine1
          addressLine2
          city
          state
          postalCode
          countryCode
          latitude
          longitude
          isDefault
          note
          fullAddress
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
 * Get single address by ID
 */
export const GET_ADDRESS_QUERY = gql`
  query GetAddress($id: UUID!) {
    getAddress(id: $id) {
      id
      userId
      receiverName
      addressLine1
      addressLine2
      city
      state
      postalCode
      countryCode
      latitude
      longitude
      isDefault
      note
      fullAddress
      createdAt
      updatedAt
    }
  }
`

/**
 * Get default address
 */
export const GET_DEFAULT_ADDRESS_QUERY = gql`
  query GetDefaultAddress {
    getDefaultAddress {
      id
      userId
      receiverName
      addressLine1
      addressLine2
      city
      state
      postalCode
      countryCode
      latitude
      longitude
      isDefault
      note
      fullAddress
      createdAt
      updatedAt
    }
  }
`

/**
 * Create new address mutation
 */
export const CREATE_ADDRESS_MUTATION = gql`
  mutation CreateAddress($input: CreateAddressInput!) {
    createAddress(input: $input) {
      id
      userId
      receiverName
      addressLine1
      addressLine2
      city
      state
      postalCode
      countryCode
      latitude
      longitude
      isDefault
      note
      fullAddress
      createdAt
      updatedAt
    }
  }
`

/**
 * Update existing address mutation
 */
export const UPDATE_ADDRESS_MUTATION = gql`
  mutation UpdateAddress($id: UUID!, $input: UpdateAddressInput!) {
    updateAddress(id: $id, input: $input) {
      id
      userId
      receiverName
      addressLine1
      addressLine2
      city
      state
      postalCode
      countryCode
      latitude
      longitude
      isDefault
      note
      fullAddress
      createdAt
      updatedAt
    }
  }
`

/**
 * Delete address mutation
 */
export const DELETE_ADDRESS_MUTATION = gql`
  mutation DeleteAddress($id: UUID!) {
    deleteAddress(id: $id)
  }
`

/**
 * Set address as default mutation
 */
export const SET_DEFAULT_ADDRESS_MUTATION = gql`
  mutation SetDefaultAddress($id: UUID!) {
    setDefaultAddress(id: $id) {
      id
      userId
      receiverName
      addressLine1
      addressLine2
      city
      state
      postalCode
      countryCode
      latitude
      longitude
      isDefault
      note
      fullAddress
      createdAt
      updatedAt
    }
  }
`
