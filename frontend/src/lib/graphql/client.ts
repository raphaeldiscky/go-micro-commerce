import { env } from '@/env'
import { getAccessToken } from '@/lib/api/client'
import { GraphQLClient } from 'graphql-request'

// GraphQL client instance
export const graphqlClient = new GraphQLClient(env.VITE_GRAPHQL_GATEWAY_URL, {
  credentials: 'include', // Send HTTP-only cookies with requests
  timeout: 30000, // 30 second timeout to prevent hanging requests
  headers: () => {
    const token = getAccessToken()

    return {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    }
  },
})
