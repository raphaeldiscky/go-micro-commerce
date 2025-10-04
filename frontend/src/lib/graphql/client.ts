import { env } from '@/env'
import { getAccessToken } from '@/lib/api/client'
import { GraphQLClient } from 'graphql-request'

// GraphQL client instance
export const graphqlClient = new GraphQLClient(env.VITE_GRAPHQL_GATEWAY_URL, {
  credentials: 'include', // Send HTTP-only cookies with requests
  headers: () => {
    const token = getAccessToken()

    return {
      ...(token ? { authorization: `Bearer ${token}` } : {}),
    }
  },
})
