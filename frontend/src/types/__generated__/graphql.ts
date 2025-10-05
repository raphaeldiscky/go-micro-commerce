/* eslint-disable */
export type Maybe<T> = T | null
export type InputMaybe<T> = Maybe<T>
export type Exact<T extends { [key: string]: unknown }> = {
  [K in keyof T]: T[K]
}
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]?: Maybe<T[SubKey]>
}
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & {
  [SubKey in K]: Maybe<T[SubKey]>
}
export type MakeEmpty<
  T extends { [key: string]: unknown },
  K extends keyof T,
> = { [_ in K]?: never }
export type Incremental<T> =
  | T
  | {
      [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never
    }
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string }
  String: { input: string; output: string }
  Boolean: { input: boolean; output: boolean }
  Int: { input: number; output: number }
  Float: { input: number; output: number }
  Time: { input: any; output: any }
  join__FieldSet: { input: any; output: any }
  link__Import: { input: any; output: any }
}

export type AuthPayload = {
  __typename?: 'AuthPayload'
  refreshToken: Scalars['String']['output']
  token: Scalars['String']['output']
  user: User
}

export type Conversation = {
  __typename?: 'Conversation'
  createdAt: Scalars['Time']['output']
  endedAt?: Maybe<Scalars['Time']['output']>
  id: Scalars['ID']['output']
  messages: MessageConnection
  participantCount: Scalars['Int']['output']
  participants: Array<Participant>
  priority: Scalars['Int']['output']
  status: ConversationStatus
  subject?: Maybe<Scalars['String']['output']>
  updatedAt: Scalars['Time']['output']
}

export type ConversationMessagesArgs = {
  limit?: InputMaybe<Scalars['Int']['input']>
  offset?: InputMaybe<Scalars['Int']['input']>
}

export enum ConversationStatus {
  Active = 'ACTIVE',
  Ended = 'ENDED',
  Waiting = 'WAITING',
}

export type CreateConversationInput = {
  initialMessage?: InputMaybe<Scalars['String']['input']>
  priority: Scalars['Int']['input']
  subject?: InputMaybe<Scalars['String']['input']>
}

export type JoinConversationInput = {
  conversationId: Scalars['ID']['input']
  role: ParticipantRole
}

export type LoginInput = {
  email: Scalars['String']['input']
  password: Scalars['String']['input']
}

export type Message = {
  __typename?: 'Message'
  content: Scalars['String']['output']
  conversation: Conversation
  conversationId: Scalars['ID']['output']
  createdAt: Scalars['Time']['output']
  id: Scalars['ID']['output']
  isSystem: Scalars['Boolean']['output']
  messageType: MessageType
  sender?: Maybe<User>
  senderId?: Maybe<Scalars['ID']['output']>
}

export type MessageConnection = {
  __typename?: 'MessageConnection'
  edges: Array<MessageEdge>
  pageInfo: PageInfo
}

export type MessageEdge = {
  __typename?: 'MessageEdge'
  cursor: Scalars['String']['output']
  node: Message
}

export enum MessageType {
  File = 'FILE',
  Image = 'IMAGE',
  System = 'SYSTEM',
  Text = 'TEXT',
}

export type Mutation = {
  __typename?: 'Mutation'
  assignConversationToAdmin: Conversation
  createConversation: Conversation
  endConversation: Conversation
  joinConversation: Participant
  leaveConversation: Scalars['Boolean']['output']
  login: AuthPayload
  logout: Scalars['Boolean']['output']
  refreshToken: AuthPayload
  register: AuthPayload
}

export type MutationAssignConversationToAdminArgs = {
  adminId: Scalars['ID']['input']
  conversationId: Scalars['ID']['input']
}

export type MutationCreateConversationArgs = {
  input: CreateConversationInput
}

export type MutationEndConversationArgs = {
  conversationId: Scalars['ID']['input']
}

export type MutationJoinConversationArgs = {
  input: JoinConversationInput
}

export type MutationLeaveConversationArgs = {
  conversationId: Scalars['ID']['input']
}

export type MutationLoginArgs = {
  input: LoginInput
}

export type MutationRegisterArgs = {
  input: RegisterUserInput
}

export type OnlineStatus = {
  __typename?: 'OnlineStatus'
  isOnline: Scalars['Boolean']['output']
  lastSeen?: Maybe<Scalars['Time']['output']>
}

export type PageInfo = {
  __typename?: 'PageInfo'
  endCursor?: Maybe<Scalars['String']['output']>
  hasNextPage: Scalars['Boolean']['output']
  hasPreviousPage: Scalars['Boolean']['output']
  startCursor?: Maybe<Scalars['String']['output']>
}

export type Participant = {
  __typename?: 'Participant'
  conversation: Conversation
  conversationId: Scalars['ID']['output']
  id: Scalars['ID']['output']
  isActive: Scalars['Boolean']['output']
  joinedAt: Scalars['Time']['output']
  leftAt?: Maybe<Scalars['Time']['output']>
  role: ParticipantRole
  user: User
  userId: Scalars['ID']['output']
  userType: UserType
}

export enum ParticipantRole {
  Member = 'MEMBER',
  Moderator = 'MODERATOR',
  Owner = 'OWNER',
}

export type Query = {
  __typename?: 'Query'
  conversation?: Maybe<Conversation>
  conversationMessages: MessageConnection
  conversationParticipants: Array<Participant>
  conversations: Array<Conversation>
  me?: Maybe<User>
  onlineUsers: Array<User>
  user?: Maybe<User>
  waitingConversations: Array<Conversation>
}

export type QueryConversationArgs = {
  id: Scalars['ID']['input']
}

export type QueryConversationMessagesArgs = {
  after?: InputMaybe<Scalars['String']['input']>
  before?: InputMaybe<Scalars['String']['input']>
  conversationId: Scalars['ID']['input']
  first?: InputMaybe<Scalars['Int']['input']>
  last?: InputMaybe<Scalars['Int']['input']>
}

export type QueryConversationParticipantsArgs = {
  conversationId: Scalars['ID']['input']
}

export type QueryUserArgs = {
  id: Scalars['ID']['input']
}

export type RegisterUserInput = {
  email: Scalars['String']['input']
  firstName: Scalars['String']['input']
  lastName: Scalars['String']['input']
  password: Scalars['String']['input']
  username: Scalars['String']['input']
}

export type User = {
  __typename?: 'User'
  conversations: Array<Conversation>
  createdAt: Scalars['Time']['output']
  email: Scalars['String']['output']
  emailVerified: Scalars['Boolean']['output']
  firstName: Scalars['String']['output']
  id: Scalars['ID']['output']
  isActive: Scalars['Boolean']['output']
  lastName: Scalars['String']['output']
  onlineStatus?: Maybe<OnlineStatus>
  updatedAt: Scalars['Time']['output']
}

export enum UserType {
  Admin = 'ADMIN',
  User = 'USER',
}

export enum Join__Graph {
  AuthService = 'AUTH_SERVICE',
  ChatService = 'CHAT_SERVICE',
}

export enum Link__Purpose {
  /** `EXECUTION` features provide metadata necessary for operation execution. */
  Execution = 'EXECUTION',
  /** `SECURITY` features provide metadata necessary to securely resolve fields. */
  Security = 'SECURITY',
}
