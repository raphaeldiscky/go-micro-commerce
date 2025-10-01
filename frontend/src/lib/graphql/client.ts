import { env } from '@/env'
import { GraphQLClient } from 'graphql-request'

// GraphQL client instance
export const graphqlClient = new GraphQLClient(env.VITE_GRAPHQL_GATEWAY_URL, {
  credentials: 'include',
  headers: () => {
    const token = localStorage.getItem('access_token')

    return {
      ...(token ? { authorization: `Bearer ${token}` } : {}),
    }
  },
})
