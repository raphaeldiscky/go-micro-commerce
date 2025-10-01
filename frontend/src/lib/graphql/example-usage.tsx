/**
 * Example: How to use GraphQL with TanStack Query (Type-Safe)
 *
 * 1. First, run codegen to generate types:
 *    pnpm codegen
 *
 * 2. Write GraphQL queries with gql tag in your component
 * 3. Generated types will be created next to this file as `example-usage.generated.ts`
 * 4. Import types from the generated file
 * 5. Use typed graphqlClient.request<TData, TVariables>()
 */

import { useMutation, useQuery } from '@tanstack/react-query'
import { gql } from 'graphql-request'
import { graphqlClient } from './client'
import type {
  GetUserQuery,
  GetUserQueryVariables,
  LoginMutation,
  LoginMutationVariables,
} from './example-usage.generated'

// Example GraphQL query
const GET_USER = gql`
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      email
      firstName
      lastName
    }
  }
`

// Example GraphQL mutation
const LOGIN = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      refreshToken
      user {
        id
        email
        firstName
        lastName
      }
    }
  }
`

// Example component using GraphQL with TanStack Query
export function UserProfile({ userId }: { userId: string }) {
  // Fetch user data with FULL type safety
  const { data, isLoading, error } = useQuery({
    queryKey: ['user', userId],
    queryFn: async () =>
      graphqlClient.request<GetUserQuery, GetUserQueryVariables>(GET_USER, {
        id: userId,
      }),
  })

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return (
    <div>
      <h1>
        {data?.user?.firstName} {data?.user?.lastName}
      </h1>
      <p>{data?.user?.email}</p>
    </div>
  )
}

// Example login form using mutation
export function LoginForm() {
  const loginMutation = useMutation({
    mutationFn: async (input: LoginMutationVariables['input']) =>
      graphqlClient.request<LoginMutation, LoginMutationVariables>(LOGIN, {
        input,
      }),
    onSuccess: (data) => {
      localStorage.setItem('token', data.login.token)
      console.log('Logged in user:', data.login.user.email)
    },
  })

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    loginMutation.mutate({
      email: formData.get('email') as string,
      password: formData.get('password') as string,
    })
  }

  return (
    <form onSubmit={handleSubmit}>
      <input name="email" type="email" required />
      <input name="password" type="password" required />
      <button type="submit" disabled={loginMutation.isPending}>
        {loginMutation.isPending ? 'Logging in...' : 'Login'}
      </button>
      {loginMutation.error && <div>Error: {loginMutation.error.message}</div>}
    </form>
  )
}
