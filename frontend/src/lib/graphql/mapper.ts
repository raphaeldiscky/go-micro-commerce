import type { User } from '@/types/__generated__/graphql'
import type { LoginMutation, RegisterMutation } from './auth.generated'

type GraphQLUser =
  | LoginMutation['login']['user']
  | RegisterMutation['register']['user']

export function mapGraphQLUserToApiUser(graphqlUser: GraphQLUser): User {
  return {
    __typename: 'User',
    conversations: [],
    createdAt: graphqlUser.createdAt,
    email: graphqlUser.email,
    emailVerified: graphqlUser.emailVerified,
    firstName: graphqlUser.firstName,
    id: graphqlUser.id,
    isActive: graphqlUser.isActive,
    lastName: graphqlUser.lastName,
    roles: graphqlUser.roles,
    notifications: [],
    unreadCount: 0,
    updatedAt: graphqlUser.updatedAt,
  }
}
