import { env } from '@/env'
import { GraphQLClient } from 'graphql-request'

// GraphQL client instance
export const graphqlClient = new GraphQLClient(env.VITE_GRAPHQL_GATEWAY_URL, {
  headers: () => {
    // Get token from localStorage or your auth store
    const token = localStorage.getItem('token')

    return {
      ...(token ? { authorization: `Bearer ${token}` } : {}),
    }
  },
})
