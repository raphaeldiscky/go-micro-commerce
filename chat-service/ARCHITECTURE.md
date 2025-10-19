# Chat Service Architecture

## Table of Contents

1. [Overview](#overview)
2. [Architecture Layers](#architecture-layers)
3. [WebSocket Infrastructure](#websocket-infrastructure)
4. [Chat-Specific Implementation](#chat-specific-implementation)
5. [Redis Pub/Sub Strategy](#redis-pubsub-strategy)
6. [Scalability Patterns](#scalability-patterns)
7. [Connection Lifecycle](#connection-lifecycle)
8. [Message Flow](#message-flow)
9. [Cross-Instance Communication](#cross-instance-communication)
10. [Auto-Join Strategy](#auto-join-strategy)
11. [Database Persistence](#database-persistence)
12. [GraphQL Integration](#graphql-integration)
13. [Possible Improvements](#possible-improvements)

---

## Overview

The chat service implements a **highly scalable, real-time messaging system** using WebSockets for client communication and Redis pub/sub for cross-instance coordination. The architecture supports:

- **Multi-instance horizontal scaling** with automatic cross-instance message delivery
- **Dynamic subscription management** to optimize Redis resource usage
- **Universal WebSocket patterns** that can be extended for other real-time features
- **Graceful degradation** when Redis is unavailable (single-instance mode)
- **Auto-join conversations** for seamless user experience

### System Components

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │◄───WS───►│   Instance   │◄───────►│   Database  │
│ (Browser)   │          │   (Chat)     │         │ (Postgres)  │
└─────────────┘          └──────────────┘         └─────────────┘
                                │
                         Redis Pub/Sub
                                │
                         ┌──────┴──────┐
                    ┌────▼───┐    ┌────▼───┐
                    │Instance│    │Instance│
                    │   2    │    │   3    │
                    └────────┘    └────────┘
```

---

## Architecture Layers

The chat service follows a **layered architecture** with clear separation of concerns:

### 1. Universal WebSocket Layer (`pkg/websocket`)

- **Responsibility**: Generic WebSocket infrastructure reusable across services
- **Components**: BaseHub, BaseConnection, Message envelope
- **Features**: Connection management, broadcasting, heartbeat, cleanup

### 2. EventBus Abstraction Layer (`pkg/rediseventbus`)

- **Responsibility**: Pub/sub abstraction for cross-instance communication
- **Implementation**: Redis-based event bus with handler routing
- **Features**: Subscribe/unsubscribe, event filtering, instance ID tracking

### 3. Redis Infrastructure Layer (`pkg/redis`)

- **Responsibility**: Low-level Redis pub/sub operations
- **Components**: Publisher, Subscriber, Message, ChannelBuilder
- **Features**: Retry logic, channel patterns, message serialization

### 4. Chat-Specific Layer (`chat-service/internal/websocket`)

- **Responsibility**: Chat domain logic and message handling
- **Components**: ChatHub, ChatConnection, ChatEventHandler
- **Features**: Conversation management, user types, message types

### 5. Service Layer (`chat-service/internal/service`)

- **Responsibility**: Business logic and database operations
- **Components**: ChatService, ConversationAccess
- **Features**: CRUD operations, access control, pagination

### 6. Handler Layer (`chat-service/internal/handler`)

- **Responsibility**: HTTP/WebSocket request handling
- **Components**: WebSocketHandler, ChatHandler
- **Features**: Authentication, connection upgrade, REST APIs

---

## WebSocket Infrastructure

### BaseHub (`pkg/websocket/hub.go`)

The **BaseHub** is the central nervous system of the WebSocket infrastructure.

#### Responsibilities

- **Connection Registry**: Track all active connections by ID and user ID
- **Channel Management**: Organize connections into logical channels (rooms)
- **Broadcasting**: Send messages to filtered sets of connections
- **Lifecycle Management**: Register/unregister connections, cleanup stale connections
- **External Subscriptions**: Bridge to GraphQL subscriptions

#### Key Data Structures

```go
type BaseHub struct {
    connections        map[uuid.UUID]Connection               // All connections by ID
    userConns          map[uuid.UUID]map[uuid.UUID]Connection // Connections by user
    channels           map[string]map[uuid.UUID]Connection    // Connections by channel
    channelSubscribers map[string]map[string]chan<- *Message  // External subscribers
    register           chan Connection                         // Registration queue
    unregister         chan Connection                         // Unregistration queue
    broadcast          chan *BroadcastRequest                  // Broadcast queue
    mutex              sync.RWMutex                           // Protects all maps
    closeOnce          sync.Once                              // Prevents double close
    done               chan struct{}                          // Shutdown signal
}
```

#### Broadcasting Strategies

**1. Broadcast to All**

```go
hub.Broadcast(message, nil) // No filter
```

**2. Broadcast to Channel**

```go
hub.BroadcastToChannel("chat:conversation:{uuid}", message)
```

**3. Broadcast with Filter**

```go
filter := func(conn Connection) bool {
    return conn.UserID() != excludeUserID
}
hub.Broadcast(message, filter)
```

**4. Broadcast to User**

```go
hub.BroadcastToUser(userID, message)
```

#### Cleanup Strategy

The hub runs a **periodic cleanup ticker** (every 2 minutes) to remove stale connections:

```go
staleThreshold := time.Now().Add(-2 * time.Minute)
for _, conn := range h.connections {
    if !conn.IsActive() || conn.GetLastHeartbeat().Before(staleThreshold) {
        conn.Close() // Triggers unregister
    }
}
```

---

### BaseConnection (`pkg/websocket/connection.go`)

The **BaseConnection** manages individual WebSocket connections.

#### Responsibilities

- **Read Pump**: Receive messages from client
- **Write Pump**: Send messages to client with ping/pong heartbeat
- **Heartbeat Management**: Track connection health
- **Graceful Shutdown**: Handle cleanup on disconnect

#### Connection Configuration

```go
type ConnectionConfig struct {
    ReadBufferSize  int           // 4096 bytes (default)
    WriteBufferSize int           // 4096 bytes (default)
    MaxMessageSize  int64         // 512 KB (default)
    PongWait        time.Duration // 60 seconds (must receive pong)
    GracePeriod     time.Duration // 30 seconds (initial connection grace)
    PingPeriod      time.Duration // 45 seconds (send ping interval)
    WriteWait       time.Duration // 10 seconds (write timeout)
    SendBufferSize  int           // 256 (channel buffer)
}
```

**Key Insight**: `PingPeriod < PongWait` ensures the server sends pings before the read deadline expires.

#### Read Pump Flow

```
1. Set initial read deadline with grace period (30s)
2. Register pong handler (extends deadline by PongWait on pong)
3. Loop:
   a. Wait for message from client
   b. Update heartbeat timestamp
   c. Parse message and call handler
   d. Handle errors (close on unexpected errors)
4. On exit: Unregister connection, call OnDisconnect
```

#### Write Pump Flow

```
1. Create ticker for ping interval (45s)
2. Loop:
   a. Wait for message from send channel OR ticker
   b. If message:
      - Check if connection is active
      - Set write deadline (10s)
      - Write JSON message to WebSocket
   c. If ticker:
      - Send ping message
      - Client responds with pong (handled by read pump)
3. On exit: Send close message, clean up
```

#### Self-Reference Pattern

BaseConnection uses a **self-reference pattern** to support type-safe handlers:

```go
type BaseConnection struct {
    self Connection // Actual connection instance (e.g., ChatConnection)
}

func (c *BaseConnection) Start(ctx context.Context) {
    c.handler.OnConnect(c.self) // Pass actual type, not base
}
```

This allows handlers to receive the correct connection type (e.g., `*ChatConnection` instead of `*BaseConnection`).

---

### Message Envelope (`pkg/websocket/message.go`)

The **Message** is a universal envelope for WebSocket communication.

#### Structure

```go
type Message struct {
    ID        uuid.UUID       `json:"id"`                  // Unique message ID
    Type      MessageType     `json:"type"`                // Message type
    Channel   *string         `json:"channel,omitempty"`   // Target channel/room
    SenderID  *uuid.UUID      `json:"sender_id,omitempty"` // Sender identification
    Content   json.RawMessage `json:"content"`             // Flexible payload
    Timestamp time.Time       `json:"timestamp"`           // Creation time
}
```

#### Standard Message Types

```go
const (
    MessageTypeHeartbeat MessageType = "heartbeat"
    MessageTypeError     MessageType = "error"
    MessageTypeSystem    MessageType = "system"
)
```

#### Chat-Specific Message Types

```go
const (
    ChatMessageTypeChat            MessageType = "chat"
    ChatMessageTypeTyping          MessageType = "typing"
    ChatMessageTypePresence        MessageType = "presence"
    ChatMessageTypeDeliveryReceipt MessageType = "delivery_receipt"
    ChatMessageTypeReadReceipt     MessageType = "read_receipt"
)
```

---

## Chat-Specific Implementation

### ChatHub (`chat-service/internal/websocket/chat_hub.go`)

The **ChatHub** extends BaseHub with chat-specific features and Redis integration.

#### Additional Responsibilities

- **Redis Integration**: Publish events to EventBus for cross-instance delivery
- **Dynamic Subscription**: Subscribe/unsubscribe from Redis channels based on active connections
- **Conversation Management**: Join/leave conversation channels
- **User Type Filtering**: Broadcast to specific user types (admin, user)

#### Key Data Structures

```go
type ChatHub struct {
    *pkgwebsocket.BaseHub

    eventBus       eventbus.EventBus         // Redis pub/sub integration
    eventHandler   *event.ChatEventHandler   // Event routing
    instanceID     string                    // Unique instance identifier
    activeChannels map[string]int            // Channel → connection count
    channelMutex   sync.RWMutex              // Protects activeChannels
}
```

#### Dynamic Subscription Strategy

**Problem**: Subscribing to all conversation channels wastes Redis resources.

**Solution**: Only subscribe when connections exist, unsubscribe when empty.

**JoinConversation Flow**:

```go
func (h *ChatHub) JoinConversation(conn *ChatConnection, conversationID uuid.UUID) {
    channelName := "chat:conversation:{conversationID}"

    // Subscribe to Redis if first connection
    if h.activeChannels[channelName] == 0 {
        h.eventBus.Subscribe(channelName, h.eventHandler.HandleEvent)
    }
    h.activeChannels[channelName]++

    // Join local hub channel
    h.JoinChannel(conn, channelName)
    conn.JoinConversation(conversationID)
}
```

**LeaveConversation Flow**:

```go
func (h *ChatHub) LeaveConversation(conn *ChatConnection, conversationID uuid.UUID) {
    channelName := "chat:conversation:{conversationID}"

    // Unsubscribe from Redis if no more connections
    h.activeChannels[channelName]--
    if h.activeChannels[channelName] <= 0 {
        h.eventBus.Unsubscribe(channelName)
        delete(h.activeChannels, channelName)
    }

    // Leave local hub channel
    h.LeaveChannel(conn, channelName)
    conn.LeaveConversation()
}
```

**Benefits**:

- Reduces Redis memory usage
- Decreases message processing overhead
- Scales to thousands of conversations

---

### ChatConnection (`chat-service/internal/websocket/chat_connection.go`)

The **ChatConnection** extends BaseConnection with chat-specific state and behavior.

#### Additional Fields

```go
type ChatConnection struct {
    *pkgwebsocket.BaseConnection

    userType       constant.UserType    // admin, user, etc.
    conversationID *uuid.UUID           // Current conversation (if any)
    connectionRepo repository.ConnectionRepository
    logger         logger.Logger
}
```

#### Auto-Join Conversations

On connection, the user is **automatically joined** to their active conversations:

```go
func (h *ChatConnectionHandler) autoJoinUserConversations(chatConn *ChatConnection) error {
    conversations, err := h.conversationGetter(ctx, chatConn.UserID(), chatConn.UserType())
    if err != nil {
        return err
    }

    // Limit to 50 conversations to prevent performance issues
    const maxAutoJoin = 50
    if len(conversations) > maxAutoJoin {
        conversations = conversations[:maxAutoJoin]
    }

    for _, conv := range conversations {
        h.hub.JoinConversation(chatConn, conv.ID)
    }

    return nil
}
```

**Why 50?**

- Balance between UX (immediate message delivery) and performance
- Heavy users with 100+ conversations would cause excessive Redis subscriptions
- Users can manually join additional conversations later

---

### Message Handlers

The **ChatConnectionHandler** implements the `ConnectionHandler` interface.

#### OnConnect

- Log connection establishment
- Auto-join user conversations (up to 50)
- Send welcome message (optional)

#### OnDisconnect

- Mark connection as inactive in database
- Log disconnection
- Cleanup happens automatically via Unregister

#### OnMessage

- Route to specific handler based on message type
- Update heartbeat in database
- Handle errors gracefully

#### OnError

- Log unexpected errors
- Ignore normal closure errors
- Optionally notify monitoring system

#### Message Type Routing

```go
switch message.Type {
case ChatMessageTypeChat:
    return h.handleChatMessage(conn, message)
case ChatMessageTypeTyping:
    return h.handleTypingMessage(conn, message)
case ChatMessageTypePresence:
    return h.handlePresenceMessage(conn, message)
case ChatMessageTypeDeliveryReceipt:
    return h.handleDeliveryReceiptMessage(conn, message)
case ChatMessageTypeReadReceipt:
    return h.handleReadReceiptMessage(conn, message)
}
```

---

## Redis Pub/Sub Strategy

### EventBus (`pkg/rediseventbus/eventbus.go`)

The **EventBus** provides a high-level abstraction over Redis pub/sub.

#### Architecture

```
ChatHub → EventBus → Redis Publisher/Subscriber → Redis Server
```

#### Key Features

**1. Subscribe with Handler**

```go
func (b *redisEventBus) Subscribe(channel string, handler EventHandler) error {
    // Add handler to local registry
    b.subscriptions[channel] = append(b.subscriptions[channel], handler)

    // Subscribe to Redis if first handler
    if len(b.subscriptions[channel]) == 1 {
        b.subscribeToRedis(channel)
    }
}
```

**2. Publish Event**

```go
func (b *redisEventBus) Publish(ctx context.Context, channel string, event Event) error {
    data, _ := event.Marshal()
    redisMsg := redis.NewMessage(metadata, data)
    return b.publisher.Publish(ctx, channel, redisMsg)
}
```

**3. Instance ID Filtering**

To prevent message loops in multi-instance deployments, each instance skips its own events:

```go
func (b *redisEventBus) handleRedisMessage(ctx context.Context, channel string, redisMsg *redis.Message) error {
    event, _ := Unmarshal(redisMsg.Payload)

    // Skip events from our own instance
    if event.SourceInstanceID == b.instanceID {
        return nil
    }

    // Call all registered handlers
    for _, handler := range b.subscriptions[channel] {
        handler(ctx, event)
    }
}
```

**Without filtering**:

```
Instance A → Redis → Instance A,B,C (all receive)
Instance A processes its own event → infinite loop
```

**With filtering**:

```
Instance A → Redis → Instance A,B,C (all receive)
Instance A skips (same instance ID)
Instance B,C process event
```

---

### Channel Naming Convention (`pkg/redis/channel.go`)

Channels follow a **hierarchical naming pattern**:

```
{service}:{entity}:{id}:{event}
```

#### Examples

**Conversation Channel**:

```go
chat:conversation:550e8400-e29b-41d4-a716-446655440000
```

**User Presence Channel**:

```go
chat:presence:550e8400-e29b-41d4-a716-446655440000
```

**Order Event Channel**:

```go
order:order:550e8400-e29b-41d4-a716-446655440000:created
```

#### Channel Builder

Fluent API for constructing channel names:

```go
channel := NewChannelBuilder().
    Service("chat").
    Entity("conversation").
    ID(conversationID).
    Build()
// Result: "chat:conversation:{uuid}"
```

#### Pattern Matching

Subscribe to multiple channels using patterns:

```go
// All chat channels
pattern := NewChannelBuilder().Service("chat").BuildPattern()
// Result: "chat:*"

// All conversation channels
pattern := NewChannelBuilder().Service("chat").Entity("conversation").BuildPattern()
// Result: "chat:conversation:*"
```

---

### Redis Message Format (`pkg/redis/message.go`)

Redis messages use a standardized envelope:

```go
type Message struct {
    MessageID string                 `json:"message_id"` // Unique ID
    Metadata  *MessageMetadata        `json:"metadata"`   // Tracing, source, etc.
    Payload   json.RawMessage         `json:"payload"`    // Actual data
}

type MessageMetadata struct {
    Source    string    `json:"source"`     // Source service
    Timestamp time.Time `json:"timestamp"`  // Creation time
    TraceID   string    `json:"trace_id"`   // Distributed tracing
}
```

---

## Scalability Patterns

### 1. Multi-Instance Deployment

The chat service supports **horizontal scaling** by running multiple instances behind a load balancer.

```
                    ┌───────────────┐
                    │ Load Balancer │
                    └───────┬───────┘
                            │
            ┌───────────────┼───────────────┐
            │               │               │
     ┌──────▼──────┐ ┌──────▼──────┐ ┌──────▼──────┐
     │  Instance 1 │ │  Instance 2 │ │  Instance 3 │
     └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
            │               │               │
            └───────────────┼───────────────┘
                            │
                    ┌───────▼───────┐
                    │  Redis Pub/Sub │
                    └───────────────┘
```

**Key Characteristics**:

- Each instance has its own in-memory connection registry
- Redis acts as the message broker between instances
- Clients maintain WebSocket to one instance only
- Messages are delivered across all instances automatically

---

### 2. Dynamic Subscription Management

**Problem**: Subscribing to all channels wastes resources.

**Solution**: Subscribe only when connections exist.

**Before Optimization**:

```
Instance A subscribes to 10,000 conversation channels
→ High Redis memory usage
→ Processes events for empty conversations
```

**After Optimization**:

```
Instance A subscribes to ~50 active conversation channels
→ Low Redis memory usage
→ Only processes relevant events
```

**Implementation**: See [JoinConversation/LeaveConversation](#dynamic-subscription-strategy) above.

---

### 3. Connection Limiting

#### Auto-Join Limit

**Why**: Users with 100+ conversations would cause:

- Excessive Redis subscriptions (100 channels per connection)
- Slow connection time (100 database queries)
- High memory usage (100 channel maps)

**Solution**: Limit auto-join to 50 most recent conversations.

```go
const maxAutoJoin = 50
if len(conversations) > maxAutoJoin {
    conversations = conversations[:maxAutoJoin]
}
```

Users can manually join older conversations via the API.

#### Send Buffer Limit

Each connection has a buffered channel for outgoing messages:

```go
send chan *Message // Buffer size: 256
```

**Why**: Prevents slow clients from blocking the hub.

**Behavior**:

- If buffer is full, connection is closed
- Fast clients can handle bursts without drops
- Slow clients are disconnected to free resources

---

### 4. Stale Connection Cleanup

**Problem**: Connections may not close cleanly (network issues, crashes, etc.)

**Solution**: Periodic cleanup ticker removes stale connections.

```go
cleanupTicker := time.NewTicker(2 * time.Minute)

for {
    case <-cleanupTicker.C:
        staleThreshold := time.Now().Add(-2 * time.Minute)
        for _, conn := range h.connections {
            if !conn.IsActive() || conn.GetLastHeartbeat().Before(staleThreshold) {
                conn.Close()
            }
        }
}
```

**Criteria for stale**:

- No heartbeat in last 2 minutes
- Connection marked as inactive

---

### 5. Graceful Shutdown

On shutdown, the hub:

1. Signals all goroutines via context cancellation
2. Closes all active connections
3. Waits up to 30 seconds for clean shutdown
4. Force-closes remaining connections

```go
func (h *ChatHub) Shutdown(ctx context.Context) error {
    close(h.done) // Signal shutdown

    shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    done := make(chan struct{})
    go func() {
        h.closeAllConnections()
        close(done)
    }()

    select {
    case <-done:
        // Clean shutdown
    case <-shutdownCtx.Done():
        // Timeout, force close
    }
}
```

---

## Connection Lifecycle

### Step-by-Step Flow

#### 1. WebSocket Upgrade

```
Client → GET /ws?token={jwt}
Server → HTTP 101 Switching Protocols
```

**Authentication**:

- Token is validated with auth service
- Claims extracted: user ID, roles, email
- User type determined from roles (admin, user)

```go
claims, err := h.connectionService.ValidateAuthToken(ctx, token)
userType := h.determineUserTypeFromRoles(claims.Roles)
```

#### 2. Connection Creation

```go
chatConn := NewChatConnection(
    userID,
    userType,
    websocketConn,
    hub,
    connectionRepo,
    messageRepo,
    conversationGetter,
    logger,
)
```

#### 3. Hub Registration

```go
hub.Register(chatConn)
```

**Hub actions**:

- Add to `connections` map
- Add to `userConns` map
- Log connection event

#### 4. Database Persistence

```go
connEntity := &entity.Connection{
    ID:            chatConn.ID(),
    UserID:        chatConn.UserID(),
    ConnectedAt:   time.Now(),
    LastHeartbeat: time.Now(),
    IsActive:      true,
}
connectionRepo.Create(ctx, connEntity)
```

#### 5. Auto-Join Conversations

```go
conversations := conversationGetter(ctx, userID, userType)
for _, conv := range conversations[:50] { // Limit 50
    hub.JoinConversation(chatConn, conv.ID)
}
```

**For each conversation**:

- Check `activeChannels[channel]`
- If 0, subscribe to Redis channel
- Increment `activeChannels[channel]`
- Add connection to local hub channel
- Set `chatConn.conversationID`

#### 6. Start Read/Write Pumps

```go
chatConn.Start(ctx)
```

**Read pump**: Waits for messages from client
**Write pump**: Sends messages to client, sends pings

#### 7. Message Processing

```
Client sends message →
Read pump receives →
Handler processes →
Database saves message →
Hub broadcasts to conversation →
Redis publishes for other instances →
Write pump sends to local connections
```

#### 8. Heartbeat Updates

Every message received:

```go
conn.UpdateHeartbeat() // In-memory timestamp
connectionRepo.UpdateHeartbeat(ctx, conn.ID()) // Database update
```

#### 9. Connection Close

Triggered by:

- Client disconnect
- Network error
- Ping/pong timeout
- Server shutdown

**Close sequence**:

```go
1. Set isActive = false
2. Close send channel (stops write pump)
3. Close WebSocket connection (stops read pump)
4. Read pump defers: Unregister → OnDisconnect
```

#### 10. Unregister from Hub

```go
hub.Unregister(chatConn)
```

**Hub actions**:

- Remove from `connections` map
- Remove from `userConns` map
- Remove from all `channels` maps

**ChatHub override**:

```go
// If connection was in a conversation
if chatConn.conversationID != nil {
    channel := "chat:conversation:{id}"
    activeChannels[channel]--

    // Unsubscribe from Redis if no more connections
    if activeChannels[channel] <= 0 {
        eventBus.Unsubscribe(channel)
        delete(activeChannels, channel)
    }
}

// Call parent unregister
h.BaseHub.Unregister(conn)
```

#### 11. Database Cleanup

```go
connectionRepo.MarkAsInactive(ctx, conn.ID())
```

---

## Message Flow

### Local Broadcasting (Single Instance)

```
1. User A sends chat message
2. ChatConnection.OnMessage receives message
3. Validate user is participant in conversation
4. Save message to database
5. Hub.BroadcastToConversation(conversationID, message, excludeUserA)
6. Filter connections: return conn.ConversationID == conversationID && conn.UserID != userA
7. For each matching connection: conn.Send(message)
8. Write pump sends message to WebSocket client
```

---

### Cross-Instance Broadcasting (Multi-Instance)

```
Instance A:
1. User A (connected to Instance A) sends message
2. Save to database
3. Broadcast to local connections (see above)
4. Publish to Redis: eventBus.Publish("chat:conversation:{id}", event)

Redis:
5. Receive message from Instance A
6. Forward to all subscribers (Instance A, B, C)

Instance B:
7. EventBus.handleRedisMessage receives event
8. Check event.SourceInstanceID != instanceID (true, continue)
9. Call registered handler: ChatHub.handleChatMessageEvent
10. ChatHub.broadcastToLocalConversation (no Redis publish)
11. Filter local connections in conversation
12. Send message to matched connections

Instance A:
7. EventBus.handleRedisMessage receives event
8. Check event.SourceInstanceID == instanceID (false, skip)
9. Event not processed (already broadcast locally)
```

**Key Insight**: Instance A broadcasts locally first, then publishes to Redis. When Instance A receives its own event from Redis, it skips processing to avoid duplication.

---

### Message Types and Their Flow

#### 1. Chat Message

```
Flow: Save → Broadcast to conversation → Send delivery receipt to sender
Content: { conversationID, text, messageType }
```

#### 2. Typing Indicator

```
Flow: Broadcast to conversation (exclude sender) → No database save
Content: { conversationID, isTyping }
```

#### 3. Presence Update

```
Flow: Broadcast to all connections → No database save
Content: { userID, status, event }
```

#### 4. Delivery Receipt

```
Flow: Broadcast to conversation (exclude sender) → No database save
Content: { messageID, conversationID, recipientID, timestamp }
```

#### 5. Read Receipt

```
Flow: Broadcast to conversation (exclude sender) → Update database
Content: { messageID, conversationID, readerID, timestamp }
```

---

## Cross-Instance Communication

### Problem Statement

**Scenario**:

- User A connects to Instance 1
- User B connects to Instance 2
- Both in same conversation
- User A sends message
- User B must receive message

**Challenge**: Instances don't share memory.

---

### Solution: Redis Pub/Sub as Message Broker

#### Architecture

```
┌─────────────┐                    ┌─────────────┐
│ Instance 1  │                    │ Instance 2  │
│             │                    │             │
│ User A ────►│                    │◄──── User B │
│             │                    │             │
│ 1. Local ◄──┤                    ├──► 4. Local │
│ Broadcast   │                    │   Broadcast │
│             │                    │             │
│ 2. Publish  │                    │ 3. Receive  │
│      │      │                    │      ▲      │
└──────┼──────┘                    └──────┼──────┘
       │                                  │
       └──────────► Redis Pub/Sub ────────┘
                   (chat:conversation:{id})
```

#### Detailed Flow

**Step 1: Local Broadcast (Instance 1)**

```go
// User A sends message
messageEntity := entity.NewMessage(conversationID, userA, "Hello")
messageRepo.Create(ctx, messageEntity) // Save to DB

// Broadcast to local connections
hub.BroadcastToConversation(conversationID, message, userA)
// → User B not connected to Instance 1, doesn't receive yet
```

**Step 2: Publish to Redis (Instance 1)**

```go
if h.eventBus != nil {
    chatEvent := &event.ChatMessageEvent{
        ConversationID: conversationID,
        Message:        message,
        ExcludeUserID:  &userA,
    }
    h.publishEvent(ctx, channelName, event.TypeChatMessage, chatEvent)
}
```

**Step 3: Receive from Redis (Instance 2)**

```go
// EventBus receives message from Redis
func (b *redisEventBus) handleRedisMessage(ctx context.Context, channel string, redisMsg *redis.Message) error {
    event := Unmarshal(redisMsg.Payload)

    // Check instance ID
    if event.SourceInstanceID == b.instanceID {
        return nil // Skip own events
    }

    // Route to handler
    for _, handler := range b.subscriptions[channel] {
        handler(ctx, event) // Calls ChatHub.handleChatMessageEvent
    }
}
```

**Step 4: Local Broadcast (Instance 2)**

```go
func (h *ChatHub) handleChatMessageEvent(ctx context.Context, e *event.ChatMessageEvent) error {
    return h.broadcastToLocalConversation(
        e.ConversationID,
        e.Message,
        e.ExcludeUserID, // userA
    )
}

func (h *ChatHub) broadcastToLocalConversation(conversationID, message, excludeUserID) {
    // Broadcast to local connections ONLY (no Redis publish)
    channelName := "chat:conversation:{conversationID}"
    connections := h.GetChannelConnections(channelName) // User B is here!

    for _, conn := range connections {
        if conn.UserID() != excludeUserID {
            conn.Send(message) // User B receives message!
        }
    }
}
```

---

### Preventing Message Loops

**Without Instance ID Filtering**:

```
Instance 1 → Publish to Redis
Redis → Broadcast to Instance 1, 2, 3
Instance 1 → Process event → Broadcast locally → Publish to Redis
Redis → Broadcast to Instance 1, 2, 3
Instance 1 → Process event → ...INFINITE LOOP
```

**With Instance ID Filtering**:

```
Instance 1 → Publish to Redis with sourceInstanceID = "instance-1"
Redis → Broadcast to Instance 1, 2, 3
Instance 1 → Check sourceInstanceID == "instance-1" → SKIP
Instance 2 → Check sourceInstanceID == "instance-2" → PROCESS
Instance 3 → Check sourceInstanceID == "instance-3" → PROCESS
```

**Implementation**:

```go
type BaseEvent struct {
    SourceInstanceID string      `json:"source_instance_id"`
    EventType        string      `json:"event_type"`
    Payload          interface{} `json:"payload"`
}

func (b *redisEventBus) handleRedisMessage(ctx context.Context, channel string, redisMsg *redis.Message) error {
    event := Unmarshal(redisMsg.Payload)

    // Critical check to prevent loops
    if event.SourceInstanceID == b.instanceID {
        b.logger.Debug("Skipping event from own instance")
        return nil
    }

    // Process event
    // ...
}
```

---

## Auto-Join Strategy

### Design Goals

1. **Immediate Message Delivery**: Users receive messages without manual subscription
2. **Performance**: Don't overload system with 1000s of subscriptions
3. **User Experience**: Seamless chat experience on connection

---

### Implementation

#### When User Connects

```go
func (h *ChatConnectionHandler) OnConnect(conn pkgwebsocket.Connection) error {
    chatConn := conn.(*ChatConnection)

    // Auto-join conversations
    if err := h.autoJoinUserConversations(chatConn); err != nil {
        h.logger.Error("Failed to auto-join conversations", "error", err)
        // Don't fail connection
    }

    return nil
}
```

#### Auto-Join Logic

```go
func (h *ChatConnectionHandler) autoJoinUserConversations(chatConn *ChatConnection) error {
    // Get user's conversations from database
    conversations, err := h.conversationGetter(
        ctx,
        chatConn.UserID(),
        chatConn.UserType(),
    )
    if err != nil {
        return err
    }

    // Limit to 50 to prevent performance issues
    const maxAutoJoin = 50
    if len(conversations) > maxAutoJoin {
        h.logger.Info("Limiting auto-join for heavy user",
            "user_id", chatConn.UserID(),
            "total_conversations", len(conversations),
            "auto_join_limit", maxAutoJoin)
        conversations = conversations[:maxAutoJoin]
    }

    // Join each conversation
    for _, conv := range conversations {
        h.hub.JoinConversation(chatConn, conv.ID)
    }

    return nil
}
```

#### Why Limit to 50?

**User with 10 conversations**:

- 10 Redis subscriptions (10 \* 1 connection)
- Minimal overhead

**User with 100 conversations**:

- 100 Redis subscriptions (100 \* 1 connection)
- 100 channel maps in memory
- 100 database queries on connect
- Significant overhead

**User with 1000 conversations**:

- 1000 Redis subscriptions
- Would overwhelm Redis and instance
- Slow connection time (1000 queries)

**Solution**: Cap at 50, prioritize recent conversations.

#### Manual Join for Additional Conversations

Users can join older conversations via REST API:

```
POST /api/v1/conversations/{id}/join
```

This calls `ChatHub.JoinConversation` directly.

---

### Conversation Selection Strategy

**Current**: First 50 conversations (database order).

**Possible Improvements**:

1. **Last Activity**: Join 50 most recently active conversations
2. **Unread Messages**: Prioritize conversations with unread messages
3. **Favorites**: User-marked important conversations
4. **Hybrid**: 25 recent + 25 unread

---

## Database Persistence

### Connection Tracking

Each WebSocket connection is persisted in the database:

```sql
CREATE TABLE connections (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL,
    connection_id VARCHAR(255) NOT NULL,
    socket_id     VARCHAR(255) NOT NULL,
    connected_at  TIMESTAMP NOT NULL,
    disconnected_at TIMESTAMP,
    last_heartbeat TIMESTAMP NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT true,
    user_agent    TEXT,
    ip_address    INET
);
```

**Why persist connections?**

- **Debugging**: Track connection issues
- **Analytics**: Active user counts, session duration
- **Billing**: Usage-based pricing
- **Presence**: Who's currently online

---

### Message Persistence

All chat messages are stored in the database:

```sql
CREATE TABLE messages (
    id              UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    sender_id       UUID NOT NULL,
    message_type    VARCHAR(50) NOT NULL,
    text            TEXT,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    CONSTRAINT fk_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
```

**Message Flow**:

1. Receive from WebSocket
2. Validate user participation
3. Save to database (source of truth)
4. Broadcast to online participants
5. Offline participants receive via REST API later

---

### Conversation Participants

```sql
CREATE TABLE participants (
    id              UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    user_id         UUID NOT NULL,
    user_type       VARCHAR(50) NOT NULL,
    role            VARCHAR(50) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    joined_at       TIMESTAMP NOT NULL,
    left_at         TIMESTAMP
);
```

**Access Control**: User can only send messages to conversations they're a participant in.

```go
func (h *ChatConnectionHandler) validateUserParticipation(
    ctx context.Context,
    userID uuid.UUID,
    userType constant.UserType,
    conversationID uuid.UUID,
) error {
    conversations, err := h.conversationGetter(ctx, userID, userType)
    if err != nil {
        return err
    }

    for _, conv := range conversations {
        if conv.ID == conversationID {
            return nil // User is participant
        }
    }

    return ErrNotParticipant
}
```

---

## GraphQL Integration

### External Channel Subscribers

The BaseHub supports **external subscribers** for bridging WebSocket events to GraphQL subscriptions.

#### Use Case

```graphql
subscription OnNewMessage($conversationID: ID!) {
  onNewMessage(conversationID: $conversationID) {
    id
    text
    sender {
      id
      name
    }
    timestamp
  }
}
```

#### Implementation

**1. Subscribe to Channel**

```go
messageChan := make(chan *websocket.Message, 256)

unsubscribe := hub.SubscribeToChannel(
    "chat:conversation:{conversationID}",
    messageChan,
)
defer unsubscribe()
```

**2. Forward Messages to GraphQL Client**

```go
for message := range messageChan {
    // Convert to GraphQL response
    gqlMessage := toGraphQLMessage(message)

    // Send to GraphQL subscription
    subscription.SendNext(gqlMessage)
}
```

**3. Broadcasting**

When a message is broadcast to a channel:

```go
func (h *BaseHub) handleBroadcast(request *BroadcastRequest) {
    // Send to WebSocket connections
    for _, conn := range h.connections {
        if filter(conn) {
            conn.Send(request.Message)
        }
    }

    // Forward to external subscribers (GraphQL)
    if request.Message.Channel != nil {
        channelName := *request.Message.Channel
        if subscribers, exists := h.channelSubscribers[channelName]; exists {
            for _, subChan := range subscribers {
                select {
                case subChan <- request.Message:
                    // Sent successfully
                default:
                    // Subscriber is slow, drop message
                }
            }
        }
    }
}
```

**Benefits**:

- Unified message delivery (WebSocket + GraphQL)
- No code duplication
- Same message ordering guarantees

---

## Possible Improvements

### 1. Performance Optimizations

#### Message Batching

**Current**: Send each message individually
**Improvement**: Batch multiple messages into single WebSocket frame

```go
type MessageBatch struct {
    Messages []*Message `json:"messages"`
}

// Accumulate messages for 10ms, then send batch
batchTicker := time.NewTicker(10 * time.Millisecond)
var batch []*Message

for {
    select {
    case msg := <-c.send:
        batch = append(batch, msg)
    case <-batchTicker.C:
        if len(batch) > 0 {
            conn.WriteJSON(MessageBatch{Messages: batch})
            batch = batch[:0]
        }
    }
}
```

**Benefits**:

- Reduce WebSocket frame overhead
- Higher throughput for busy conversations
- Lower CPU usage

---

#### Connection Pooling for Database

**Current**: Each handler creates new DB connection
**Improvement**: Use connection pool with prepared statements

```go
// Prepared statement
var updateHeartbeatStmt *sql.Stmt

func init() {
    updateHeartbeatStmt = db.Prepare(
        "UPDATE connections SET last_heartbeat = $1 WHERE id = $2",
    )
}

func (r *connectionRepo) UpdateHeartbeat(ctx context.Context, connID string) error {
    _, err := updateHeartbeatStmt.ExecContext(ctx, time.Now(), connID)
    return err
}
```

---

#### Redis Pipeline

**Current**: Publish messages one at a time
**Improvement**: Use Redis pipeline for bulk publishing

```go
pipe := redisClient.Pipeline()
for _, event := range events {
    pipe.Publish(ctx, channel, event)
}
pipe.Exec(ctx)
```

---

### 2. Monitoring and Observability

#### Metrics to Track

```go
// Connection metrics
connectionCount := prometheus.NewGauge(...)
connectionDuration := prometheus.NewHistogram(...)
connectionErrors := prometheus.NewCounter(...)

// Message metrics
messagesReceived := prometheus.NewCounter(...)
messagesSent := prometheus.NewCounter(...)
messageLatency := prometheus.NewHistogram(...)

// Redis metrics
redisPublishLatency := prometheus.NewHistogram(...)
redisSubscriptionCount := prometheus.NewGauge(...)
```

#### Distributed Tracing

```go
import "go.opentelemetry.io/otel"

func (h *ChatConnectionHandler) handleChatMessage(
    conn pkgwebsocket.Connection,
    message *pkgwebsocket.Message,
) error {
    ctx, span := otel.Tracer("chat").Start(ctx, "handleChatMessage")
    defer span.End()

    span.SetAttributes(
        attribute.String("user_id", conn.UserID().String()),
        attribute.String("conversation_id", conversationID.String()),
    )

    // Process message...
}
```

#### Health Checks

```go
GET /health
{
    "status": "healthy",
    "websocket": {
        "connections": 1234,
        "subscriptions": 567
    },
    "redis": {
        "connected": true,
        "latency_ms": 2.5
    },
    "database": {
        "connected": true,
        "pool_active": 15,
        "pool_idle": 5
    }
}
```

---

### 3. Advanced Features

#### Message History on Connect

**Current**: Auto-join conversations, no history
**Improvement**: Send last N messages on join

```go
func (h *ChatHub) JoinConversation(conn *ChatConnection, conversationID uuid.UUID) {
    // Join channel
    h.JoinChannel(conn, channelName)

    // Send message history
    messages, _ := h.messageRepo.FindByConversationID(
        ctx,
        conversationID,
        50, // Last 50 messages
        0,
    )

    for _, msg := range messages {
        wsMsg := toWebSocketMessage(msg)
        conn.Send(wsMsg)
    }
}
```

---

#### Offline Message Queue

**Current**: Offline users miss messages
**Improvement**: Queue messages for offline users, deliver on reconnect

```go
type OfflineMessageQueue struct {
    userID   uuid.UUID
    messages []*Message
    mutex    sync.RWMutex
}

func (q *OfflineMessageQueue) Enqueue(message *Message) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    q.messages = append(q.messages, message)

    // Persist to database
    offlineMessageRepo.Create(ctx, q.userID, message)
}

func (q *OfflineMessageQueue) Drain(conn Connection) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    for _, msg := range q.messages {
        conn.Send(msg)
    }

    q.messages = q.messages[:0]

    // Mark as delivered
    offlineMessageRepo.MarkDelivered(ctx, q.userID)
}
```

---

#### Presence System

**Current**: No presence tracking
**Improvement**: Track online/offline/away status

```go
type PresenceManager struct {
    userStatus map[uuid.UUID]PresenceStatus
    mutex      sync.RWMutex
}

type PresenceStatus string

const (
    PresenceOnline  PresenceStatus = "online"
    PresenceAway    PresenceStatus = "away"
    PresenceOffline PresenceStatus = "offline"
)

func (pm *PresenceManager) SetStatus(userID uuid.UUID, status PresenceStatus) {
    pm.mutex.Lock()
    defer pm.mutex.Unlock()

    pm.userStatus[userID] = status

    // Broadcast presence update
    hub.BroadcastPresenceUpdate(userID, status)
}

// Auto-detect away
func (pm *PresenceManager) StartAwayDetection(conn Connection) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            lastHeartbeat := conn.GetLastHeartbeat()
            if time.Since(lastHeartbeat) > 5*time.Minute {
                pm.SetStatus(conn.UserID(), PresenceAway)
            }
        }
    }
}
```

---

#### Read Receipts

**Current**: Delivery receipts only
**Improvement**: Track when messages are read

```go
type ReadReceipt struct {
    MessageID      uuid.UUID `json:"message_id"`
    ConversationID uuid.UUID `json:"conversation_id"`
    ReaderID       uuid.UUID `json:"reader_id"`
    ReadAt         int64     `json:"read_at"`
}

// Client sends read receipt
func (h *ChatConnectionHandler) handleReadReceiptMessage(
    conn pkgwebsocket.Connection,
    message *pkgwebsocket.Message,
) error {
    var receipt ReadReceipt
    message.ParseContent(&receipt)

    // Update database
    messageRepo.MarkAsRead(ctx, receipt.MessageID, receipt.ReaderID, time.Unix(receipt.ReadAt, 0))

    // Broadcast to sender
    hub.BroadcastReadReceipt(receipt.ConversationID, message, conn.UserID())

    return nil
}
```

---

#### Message Editing

**Current**: Messages are immutable
**Improvement**: Allow editing with history

```go
type MessageEdit struct {
    ID         uuid.UUID `json:"id"`
    MessageID  uuid.UUID `json:"message_id"`
    OldText    string    `json:"old_text"`
    NewText    string    `json:"new_text"`
    EditedAt   time.Time `json:"edited_at"`
    EditedByID uuid.UUID `json:"edited_by_id"`
}

func (r *messageRepo) Edit(ctx context.Context, messageID uuid.UUID, newText string) error {
    // Save edit history
    tx, _ := r.db.BeginTx(ctx, nil)

    // Get current message
    var oldText string
    tx.QueryRow("SELECT text FROM messages WHERE id = $1", messageID).Scan(&oldText)

    // Update message
    tx.Exec("UPDATE messages SET text = $1, updated_at = $2 WHERE id = $3",
        newText, time.Now(), messageID)

    // Insert edit record
    tx.Exec("INSERT INTO message_edits (message_id, old_text, new_text, edited_at) VALUES ($1, $2, $3, $4)",
        messageID, oldText, newText, time.Now())

    tx.Commit()
}
```

---

#### Rate Limiting

**Current**: No rate limiting
**Improvement**: Prevent spam and abuse

```go
type RateLimiter struct {
    limits map[uuid.UUID]*rate.Limiter
    mutex  sync.RWMutex
}

func (rl *RateLimiter) Allow(userID uuid.UUID) bool {
    rl.mutex.RLock()
    limiter, exists := rl.limits[userID]
    rl.mutex.RUnlock()

    if !exists {
        rl.mutex.Lock()
        limiter = rate.NewLimiter(rate.Every(time.Second), 10) // 10 messages/sec
        rl.limits[userID] = limiter
        rl.mutex.Unlock()
    }

    return limiter.Allow()
}

func (h *ChatConnectionHandler) handleChatMessage(
    conn pkgwebsocket.Connection,
    message *pkgwebsocket.Message,
) error {
    if !rateLimiter.Allow(conn.UserID()) {
        return errors.New("rate limit exceeded")
    }

    // Process message...
}
```

---

### 4. Security Hardening

#### Message Validation

```go
func validateMessage(message *Message) error {
    // Check length
    if len(message.Text) > 10000 {
        return errors.New("message too long")
    }

    // Check content
    if containsProfanity(message.Text) {
        return errors.New("inappropriate content")
    }

    // Check rate
    if !rateLimiter.Allow(message.SenderID) {
        return errors.New("rate limit exceeded")
    }

    return nil
}
```

#### XSS Prevention

```go
import "html"

func sanitizeMessage(text string) string {
    // HTML escape
    text = html.EscapeString(text)

    // Remove JavaScript protocols
    text = strings.ReplaceAll(text, "javascript:", "")

    return text
}
```

#### Token Refresh

```go
// Refresh token before expiry
func (c *ChatConnection) startTokenRefresh(refreshInterval time.Duration) {
    ticker := time.NewTicker(refreshInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            newToken, err := c.authClient.RefreshToken(c.token)
            if err != nil {
                c.Close()
                return
            }
            c.token = newToken
        }
    }
}
```

---

## Conclusion

The chat service implements a **production-ready, scalable real-time messaging system** with:

✅ **Multi-instance horizontal scaling** via Redis pub/sub
✅ **Dynamic subscription management** for efficient resource usage
✅ **Graceful degradation** when Redis is unavailable
✅ **Auto-join conversations** for seamless user experience
✅ **Comprehensive database persistence** for reliability
✅ **Universal WebSocket patterns** for code reusability
✅ **Clean architecture** with clear separation of concerns

The system is designed to handle **thousands of concurrent connections** per instance and scale horizontally to support **millions of users**.

---

## References

- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [Redis Pub/Sub](https://redis.io/topics/pubsub)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)

---

**Last Updated**: 2025-10-11
**Version**: 1.0.0
**Author**: Chat Service Team
