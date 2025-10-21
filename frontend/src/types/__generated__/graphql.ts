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

export type Address = {
  __typename?: 'Address'
  addressLine1: Scalars['String']['output']
  addressLine2?: Maybe<Scalars['String']['output']>
  city: Scalars['String']['output']
  countryCode: Scalars['String']['output']
  createdAt: Scalars['Time']['output']
  fullAddress: Scalars['String']['output']
  id: Scalars['ID']['output']
  isDefault: Scalars['Boolean']['output']
  latitude?: Maybe<Scalars['Float']['output']>
  longitude?: Maybe<Scalars['Float']['output']>
  note?: Maybe<Scalars['String']['output']>
  postalCode: Scalars['String']['output']
  receiverName: Scalars['String']['output']
  state?: Maybe<Scalars['String']['output']>
  updatedAt: Scalars['Time']['output']
  userId: Scalars['ID']['output']
}

export type AddressConnection = {
  __typename?: 'AddressConnection'
  edges: Array<AddressEdge>
  pageInfo: PageInfo
}

export type AddressEdge = {
  __typename?: 'AddressEdge'
  cursor: Scalars['String']['output']
  node: Address
}

export type AuthPayload = {
  __typename?: 'AuthPayload'
  refreshToken: Scalars['String']['output']
  token: Scalars['String']['output']
  user: User
}

export type ChatConnection = {
  __typename?: 'ChatConnection'
  nodeAddress: Scalars['String']['output']
  userId: Scalars['ID']['output']
  userType: Scalars['String']['output']
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

export type ConversationEvent =
  | DeliveryReceipt
  | NewMessage
  | ReadReceipt
  | TypingIndicator

export enum ConversationStatus {
  Active = 'ACTIVE',
  Ended = 'ENDED',
  Waiting = 'WAITING',
}

export type CreateAddressInput = {
  addressLine1: Scalars['String']['input']
  addressLine2?: InputMaybe<Scalars['String']['input']>
  city: Scalars['String']['input']
  countryCode: Scalars['String']['input']
  isDefault: Scalars['Boolean']['input']
  latitude?: InputMaybe<Scalars['Float']['input']>
  longitude?: InputMaybe<Scalars['Float']['input']>
  note?: InputMaybe<Scalars['String']['input']>
  postalCode: Scalars['String']['input']
  receiverName: Scalars['String']['input']
  state?: InputMaybe<Scalars['String']['input']>
}

export type CreateConversationInput = {
  initialMessage?: InputMaybe<Scalars['String']['input']>
  priority: Scalars['Int']['input']
  subject?: InputMaybe<Scalars['String']['input']>
}

export type DeliveryReceipt = {
  __typename?: 'DeliveryReceipt'
  conversationId: Scalars['ID']['output']
  deliveredAt: Scalars['Time']['output']
  messageId: Scalars['ID']['output']
  recipientId: Scalars['ID']['output']
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
  createAddress: Address
  createConversation: Conversation
  deleteAddress: Scalars['Boolean']['output']
  endConversation: Conversation
  joinConversation: Participant
  leaveConversation: Scalars['Boolean']['output']
  login: AuthPayload
  logout: Scalars['Boolean']['output']
  markAllAsRead: Scalars['Boolean']['output']
  markAsRead: Notification
  refreshToken: AuthPayload
  register: AuthPayload
  requestChatConnection: ChatConnection
  sendDeliveryReceipt: DeliveryReceipt
  sendMessage: Message
  sendReadReceipt: ReadReceipt
  sendTypingIndicator: TypingIndicator
  setDefaultAddress: Address
  updateAddress: Address
  updatePresence: PresenceUpdate
}

export type MutationAssignConversationToAdminArgs = {
  adminId: Scalars['ID']['input']
  conversationId: Scalars['ID']['input']
}

export type MutationCreateAddressArgs = {
  input: CreateAddressInput
}

export type MutationCreateConversationArgs = {
  input: CreateConversationInput
}

export type MutationDeleteAddressArgs = {
  id: Scalars['ID']['input']
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

export type MutationMarkAsReadArgs = {
  id: Scalars['ID']['input']
}

export type MutationRegisterArgs = {
  input: RegisterUserInput
}

export type MutationSendDeliveryReceiptArgs = {
  input: SendDeliveryReceiptInput
}

export type MutationSendMessageArgs = {
  input: SendMessageInput
}

export type MutationSendReadReceiptArgs = {
  input: SendReadReceiptInput
}

export type MutationSendTypingIndicatorArgs = {
  input: TypingIndicatorInput
}

export type MutationSetDefaultAddressArgs = {
  id: Scalars['ID']['input']
}

export type MutationUpdateAddressArgs = {
  id: Scalars['ID']['input']
  input: UpdateAddressInput
}

export type MutationUpdatePresenceArgs = {
  status: PresenceStatus
}

export type NewMessage = {
  __typename?: 'NewMessage'
  content: Scalars['String']['output']
  conversationId: Scalars['ID']['output']
  createdAt: Scalars['Time']['output']
  id: Scalars['ID']['output']
  isSystem: Scalars['Boolean']['output']
  messageType: MessageType
  senderId: Scalars['ID']['output']
}

export type NewNotification = {
  __typename?: 'NewNotification'
  createdAt: Scalars['Time']['output']
  id: Scalars['ID']['output']
  isRead: Scalars['Boolean']['output']
  message: Scalars['String']['output']
  metadata?: Maybe<Scalars['String']['output']>
  title: Scalars['String']['output']
  type: PushNotificationType
  userId: Scalars['ID']['output']
}

export type Notification = {
  __typename?: 'Notification'
  createdAt: Scalars['Time']['output']
  id: Scalars['ID']['output']
  isRead: Scalars['Boolean']['output']
  message: Scalars['String']['output']
  metadata?: Maybe<Scalars['String']['output']>
  readAt?: Maybe<Scalars['Time']['output']>
  title: Scalars['String']['output']
  type: PushNotificationType
  updatedAt: Scalars['Time']['output']
  userId: Scalars['ID']['output']
}

export type NotificationConnection = {
  __typename?: 'NotificationConnection'
  edges: Array<NotificationEdge>
  pageInfo: PageInfo
}

export type NotificationDeleted = {
  __typename?: 'NotificationDeleted'
  id: Scalars['ID']['output']
  userId: Scalars['ID']['output']
}

export type NotificationEdge = {
  __typename?: 'NotificationEdge'
  cursor: Scalars['String']['output']
  node: Notification
}

export type NotificationEvent =
  | NewNotification
  | NotificationDeleted
  | NotificationRead

export type NotificationRead = {
  __typename?: 'NotificationRead'
  id: Scalars['ID']['output']
  readAt: Scalars['Time']['output']
  userId: Scalars['ID']['output']
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

export enum PresenceStatus {
  Away = 'AWAY',
  Busy = 'BUSY',
  Offline = 'OFFLINE',
  Online = 'ONLINE',
}

export type PresenceUpdate = {
  __typename?: 'PresenceUpdate'
  lastSeen?: Maybe<Scalars['Time']['output']>
  status: PresenceStatus
  userId: Scalars['ID']['output']
}

export enum PushNotificationType {
  NewMessage = 'NEW_MESSAGE',
  NewProduct = 'NEW_PRODUCT',
  OrderCancelled = 'ORDER_CANCELLED',
  OrderConfirmed = 'ORDER_CONFIRMED',
  OrderDelivered = 'ORDER_DELIVERED',
  OrderShipped = 'ORDER_SHIPPED',
  OrderUpdate = 'ORDER_UPDATE',
  PaymentSuccess = 'PAYMENT_SUCCESS',
  SystemAlert = 'SYSTEM_ALERT',
}

export type Query = {
  __typename?: 'Query'
  conversation?: Maybe<Conversation>
  conversationMessages: MessageConnection
  conversationParticipants: Array<Participant>
  conversations: Array<Conversation>
  getAddress: Address
  getDefaultAddress: Address
  getTabCounts: TabCounts
  getUnreadCount: UnreadCount
  listAddresses: AddressConnection
  listNotifications: NotificationConnection
  listUnreadNotifications: NotificationConnection
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

export type QueryGetAddressArgs = {
  id: Scalars['ID']['input']
}

export type QueryListAddressesArgs = {
  cursor?: InputMaybe<Scalars['String']['input']>
  limit: Scalars['Int']['input']
}

export type QueryListNotificationsArgs = {
  cursor?: InputMaybe<Scalars['String']['input']>
  limit: Scalars['Int']['input']
}

export type QueryListUnreadNotificationsArgs = {
  cursor?: InputMaybe<Scalars['String']['input']>
  limit: Scalars['Int']['input']
}

export type QueryUserArgs = {
  id: Scalars['ID']['input']
}

export type ReadReceipt = {
  __typename?: 'ReadReceipt'
  conversationId: Scalars['ID']['output']
  messageId: Scalars['ID']['output']
  readAt: Scalars['Time']['output']
  readerId: Scalars['ID']['output']
}

export type RegisterUserInput = {
  email: Scalars['String']['input']
  firstName: Scalars['String']['input']
  lastName: Scalars['String']['input']
  password: Scalars['String']['input']
  username: Scalars['String']['input']
}

export enum Role {
  Admin = 'ADMIN',
  User = 'USER',
}

export type SendDeliveryReceiptInput = {
  conversationId: Scalars['ID']['input']
  messageId: Scalars['ID']['input']
}

export type SendMessageInput = {
  content: Scalars['String']['input']
  conversationId: Scalars['ID']['input']
  messageType?: InputMaybe<MessageType>
  replyToId?: InputMaybe<Scalars['ID']['input']>
}

export type SendReadReceiptInput = {
  conversationId: Scalars['ID']['input']
  messageId: Scalars['ID']['input']
}

export type Subscription = {
  __typename?: 'Subscription'
  conversationEvents: ConversationEvent
  notificationEvents: NotificationEvent
  userEvents: UserEvent
}

export type SubscriptionConversationEventsArgs = {
  conversationId: Scalars['ID']['input']
}

export type TabCounts = {
  __typename?: 'TabCounts'
  all: Scalars['Int']['output']
  read: Scalars['Int']['output']
  unread: Scalars['Int']['output']
}

export type TypingIndicator = {
  __typename?: 'TypingIndicator'
  conversationId: Scalars['ID']['output']
  isTyping: Scalars['Boolean']['output']
  timestamp: Scalars['Time']['output']
  userId: Scalars['ID']['output']
}

export type TypingIndicatorInput = {
  conversationId: Scalars['ID']['input']
  isTyping: Scalars['Boolean']['input']
}

export type UnreadCount = {
  __typename?: 'UnreadCount'
  count: Scalars['Int']['output']
}

export type UpdateAddressInput = {
  addressLine1?: InputMaybe<Scalars['String']['input']>
  addressLine2?: InputMaybe<Scalars['String']['input']>
  city?: InputMaybe<Scalars['String']['input']>
  countryCode?: InputMaybe<Scalars['String']['input']>
  latitude?: InputMaybe<Scalars['Float']['input']>
  longitude?: InputMaybe<Scalars['Float']['input']>
  note?: InputMaybe<Scalars['String']['input']>
  postalCode?: InputMaybe<Scalars['String']['input']>
  receiverName?: InputMaybe<Scalars['String']['input']>
  state?: InputMaybe<Scalars['String']['input']>
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
  notifications: Array<Notification>
  onlineStatus?: Maybe<OnlineStatus>
  roles: Array<Scalars['String']['output']>
  unreadCount: Scalars['Int']['output']
  updatedAt: Scalars['Time']['output']
}

export type UserEvent = PresenceUpdate

export enum UserType {
  Admin = 'ADMIN',
  User = 'USER',
}

export enum Join__Graph {
  AuthService = 'AUTH_SERVICE',
  ChatService = 'CHAT_SERVICE',
  NotificationService = 'NOTIFICATION_SERVICE',
}

export enum Link__Purpose {
  /** `EXECUTION` features provide metadata necessary for operation execution. */
  Execution = 'EXECUTION',
  /** `SECURITY` features provide metadata necessary to securely resolve fields. */
  Security = 'SECURITY',
}
