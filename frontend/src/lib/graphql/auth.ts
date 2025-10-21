import { gql } from 'graphql-request'

/**
 * Get current user query
 */
export const ME_QUERY = gql`
  query Me {
    me {
      id
      email
      firstName
      lastName
      isActive
      roles
      emailVerified
      createdAt
      updatedAt
    }
  }
`

/**
 * Login mutation
 * Authenticates a user with email and password
 */
export const LOGIN_MUTATION = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      refreshToken
      user {
        id
        email
        firstName
        lastName
        isActive
        roles
        emailVerified
        createdAt
        updatedAt
      }
    }
  }
`

/**
 * Register mutation
 * Creates a new user account
 */
export const REGISTER_MUTATION = gql`
  mutation Register($input: RegisterUserInput!) {
    register(input: $input) {
      token
      refreshToken
      user {
        id
        email
        firstName
        lastName
        isActive
        roles
        emailVerified
        createdAt
        updatedAt
      }
    }
  }
`

/**
 * Logout mutation
 * Logs out the current user
 */
export const LOGOUT_MUTATION = gql`
  mutation Logout {
    logout
  }
`

/**
 * Refresh token mutation
 * Gets a new access token using the refresh token from HTTP-only cookie
 */
export const REFRESH_TOKEN_MUTATION = gql`
  mutation RefreshToken {
    refreshToken {
      token
      refreshToken
    }
  }
`
