# Notification Service Architecture

## Table of Contents

1. [Overview](#overview)
2. [High-Level Architecture](#high-level-architecture)
3. [Core Components](#core-components)
4. [Data Flow](#data-flow)
5. [SSE and Redis Pub/Sub Integration](#sse-and-redis-pubsub-integration)
6. [Scalability Strategy](#scalability-strategy)
7. [Channel Strategy Comparison](#channel-strategy-comparison)
8. [Database Schema](#database-schema)
9. [Delivery Guarantees](#delivery-guarantees)
10. [API Endpoints](#api-endpoints)
11. [Configuration](#configuration)
12. [Monitoring and Observability](#monitoring-and-observability)
13. [Possible Improvements](#possible-improvements)

---

## Overview

The Notification Service is a horizontally scalable, event-driven microservice responsible for:

- **Consuming events** from other services via Kafka
- **Processing notifications** with exactly-once delivery guarantees using the inbox pattern
- **Storing notifications** persistently in PostgreSQL
- **Delivering real-time push notifications** via Server-Sent Events (SSE)
- **Enabling cross-instance fan-out** using Redis Pub/Sub for multi-instance scalability
- **Sending email notifications** via SMTP for various business events

### Technology Stack

- **Language**: Go 1.23+
- **Database**: PostgreSQL (notification storage, inbox pattern)
- **Message Broker**: Apache Kafka (event ingestion)
- **Cache & Pub/Sub**: Redis Cluster (cross-instance messaging)
- **Real-time Protocol**: Server-Sent Events (SSE)
- **HTTP Framework**: Echo v4
- **Email**: SMTP

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          NOTIFICATION SERVICE                               │
│                                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                    │
│  │   Instance   │  │   Instance   │  │   Instance   │                    │
│  │      #1      │  │      #2      │  │      #N      │                    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                    │
│         │                 │                 │                             │
│         │                 │                 │                             │
│  ┌──────▼─────────────────▼─────────────────▼───────┐                    │
│  │         Redis Pub/Sub Event Bus                   │                    │
│  │    (Cross-Instance Fan-out & Coordination)        │                    │
│  └────────────────────────────────────────────────────┘                    │
│                                                                             │
│  Each Instance Contains:                                                   │
│  ┌───────────────────────────────────────────────────────────┐            │
│  │  ┌───────────────┐  ┌───────────────┐  ┌─────────────┐   │            │
│  │  │ Kafka Consumer│  │ Inbox Processor│  │ HTTP Server │   │            │
│  │  │   Worker      │  │    Worker      │  │   Worker    │   │            │
│  │  └───────┬───────┘  └───────┬───────┘  └──────┬──────┘   │            │
│  │          │                   │                 │          │            │
│  │          │                   │                 │          │            │
│  │  ┌───────▼───────────────────▼─────────────────▼──────┐   │            │
│  │  │          Notification Service Layer              │   │            │
│  │  │    (Business Logic & Event Processing)            │   │            │
│  │  └───────┬───────────────────┬─────────────────┬──────┘   │            │
│  │          │                   │                 │          │            │
│  │  ┌───────▼───────┐  ┌────────▼─────┐  ┌───────▼──────┐   │            │
│  │  │   SSE Hub     │  │  PostgreSQL  │  │ Email Service│   │            │
│  │  │  (Real-time)  │  │ (Persistent) │  │    (SMTP)    │   │            │
│  │  └───────┬───────┘  └──────────────┘  └──────────────┘   │            │
│  │          │                                                │            │
│  │          │ Static Subscription                            │            │
│  │          │ notification:shard:{0-255}                     │            │
│  │          │                                                │            │
│  │  ┌───────▼───────────────────────────────────────────┐   │            │
│  │  │         Redis Pub/Sub (Shard Channels)            │   │            │
│  │  └───────────────────────────────────────────────────┘   │            │
│  └────────────────────────────────────────────────────────────┘            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                  │
                                  │ SSE Connections
                                  │
                        ┌─────────▼──────────┐
                        │   Client Browsers  │
                        │  (Real-time Push)  │
                        └────────────────────┘
```

### Event Flow

```
┌──────────────┐
│ Order Service│────┐
│ Auth Service │    │
│ Other Svcs   │    │ Publish Events
└──────────────┘    │
                    │
                    ▼
            ┌────────────┐
            │   Kafka    │
            │  (Topics)  │
            └──────┬─────┘
                   │
                   │ Consume
                   │
    ┌──────────────▼───────────────┐
    │    Kafka Consumer Worker     │
    │  (Transactional Inbox Write) │
    └──────────────┬───────────────┘
                   │
                   │ Store in inbox_events
                   │
            ┌──────▼──────┐
            │ PostgreSQL  │
            │ inbox_events│
            └──────┬──────┘
                   │
                   │ Poll (every 5s)
                   │
      ┌────────────▼──────────────┐
      │   Inbox Processor Worker  │
      │ (Process + Retry Logic)   │
      └────────────┬──────────────┘
                   │
                   │ Route by event_type
                   │
    ┌──────────────▼────────────────┐
    │ NotificationEventService      │
    │ - ProcessNotificationRequest  │
    │ - ProcessEmailVerification    │
    │ - ProcessUserVerified         │
    └────────────┬──────────────────┘
                 │
        ┌────────┴────────┐
        │                 │
        ▼                 ▼
 ┌──────────┐      ┌──────────┐
 │  Email   │      │   Push   │
 │ (SMTP)   │      │  (SSE)   │
 └──────────┘      └─────┬────┘
                         │
                         │ 1. Save to DB
                         │ 2. Broadcast local SSE
                         │ 3. Publish to Redis
                         │
                  ┌──────▼────────┐
                  │ notifications │
                  │    table      │
                  └───────────────┘
                         │
            ┌────────────┴────────────┐
            │                         │
            ▼                         ▼
    ┌──────────────┐         ┌──────────────┐
    │  SSE Hub     │         │ Redis Pub/Sub│
    │  (Instance)  │◄────────│notification: │
    │  Broadcast   │         │shard:{0-255} │
    └──────┬───────┘         └──────────────┘
           │                         │
           │                         │ Cross-instance
           │                         │ fan-out
           │                         │
           │                 ┌───────▼────────┐
           │                 │ Other Instances│
           │                 │   SSE Hubs     │
           │                 └───────┬────────┘
           │                         │
           └─────────┬───────────────┘
                     │
                     │ SSE Stream
                     │
              ┌──────▼──────┐
              │   Clients   │
              │ (Browsers)  │
              └─────────────┘
```

---

## Core Components

### 1. Kafka Consumer Worker

**Responsibility**: Consumes events from Kafka topics and stores them transactionally in the inbox.

**Key Features**:

- Subscribes to multiple Kafka topics (e.g., `order.lifecycle`, `user.verification`)
- Implements the **Transactional Inbox Pattern** for exactly-once delivery
- Stores events in `inbox_events` table with deduplication by `message_id`
- Handles message deserialization and validation
- Supports graceful shutdown with offset management

**Topics Consumed**:

- `order.lifecycle`: Order-related events (created, updated, shipped, delivered, etc.)
- `user.verification`: User verification events (email verification, user verified)
- `notification.requests`: Direct notification requests from other services

**Implementation**: `cmd/worker/kafka_consumer.go`

---

### 2. Inbox Processor Worker

**Responsibility**: Polls and processes pending events from the inbox with retry logic and exponential backoff.

**Key Features**:

- **Polling Mechanism**: Checks for pending events every 5 seconds (configurable)
- **Idempotent Processing**: Processes each event exactly once
- **Retry Strategy**:
  - Exponential backoff: `2^attempts × base_backoff`
  - Configurable max retry attempts (default: 5)
  - Schedules failed events for future retry
- **Error Handling**: Marks permanently failed events after max retries
- **Cleanup Loop**: Removes processed events older than retention period (default: 7 days)

**State Transitions**:

```
pending → processing → processed (success)
                    ↓
                  retry → processing → processed (eventual success)
                                    ↓
                                  failed (permanent failure)
```

**Implementation**: `internal/worker/inbox_processor.go`

---

### 3. Notification Event Service

**Responsibility**: Handles business logic for different notification types.

**Event Routing**:

| Event Type                          | Handler                             | Notification Type         |
| ----------------------------------- | ----------------------------------- | ------------------------- |
| `notification.requested`            | `ProcessNotificationRequest()`      | Email or Push             |
| `user.email-verification-requested` | `ProcessEmailVerificationRequest()` | Email (Verification Link) |
| `user.verified`                     | `ProcessEmailUserVerified()`        | Email (Welcome)           |

**Push Notification Flow**:

1. Extract user ID from event payload
2. Create `Notification` entity and save to database
3. Create SSE `Message` with event type `notification_created`
4. **Broadcast to local SSE connections** via `SSEHub.BroadcastToUser()`
5. **Publish to Redis** shard channel `notification:shard:{N}` (calculated via `hash(userID) % shardCount`) for cross-instance fan-out
6. Other instances receive Redis event, apply application-layer filtering, and broadcast to local SSE connections if user is connected

**Email Notification Flow**:

1. Parse event payload and generate email template
2. Send email via SMTP
3. Log success/failure

**Implementation**: `internal/service/notification_event_service.go`

---

### 4. SSE Hub (Real-Time Push)

**Responsibility**: Manages Server-Sent Events connections and broadcasts real-time notifications.

**Architecture**:

- **Universal Design**: Can be used by any service with configurable channel naming and strategy
- **Connection Management**: Tracks active SSE connections per user
- **Static Subscriptions**: Subscribes to all shard channels at startup (default shard-based strategy)
- **Application-Layer Filtering**: Filters messages by checking local user connections before broadcasting
- **Heartbeat Mechanism**: Sends heartbeat comments every 2 minutes to keep connections alive
- **Graceful Shutdown**: Closes all connections cleanly on service shutdown

**Key Methods**:

- `Register(conn)`: Registers new SSE connection
- `Unregister(conn)`: Unregisters connection
- `BroadcastToUser(userID, message)`: Sends message to all connections for a user
- `SetEventBus(eventBus, instanceID, channelBuilder, shardChannelBuilder, strategy, shardCount)`: Configures Redis pub/sub integration with channel strategy

**SSE Message Format**:

```
id: {uuid}
event: {event_type}
data: {json_payload}

```

**Implementation**: `pkg/sse/hub.go`

---

### 5. Redis Pub/Sub EventBus

**Responsibility**: Enables cross-instance communication for real-time fan-out.

**Channel Strategy** (Default: Shard-Based):

- **Shard Channels**: `notification:shard:{0-255}` (256 fixed channels)
- **Static Subscriptions**: All instances subscribe to all shards at startup
- **Application-Layer Filtering**: Filter messages by checking local user connections
- **Message Enrichment**: Payload includes userID for filtering

**Why Redis Pub/Sub?**:

- **Low Latency**: Sub-millisecond message delivery
- **Fan-out**: Single publish reaches all subscribed instances
- **Ephemeral**: No persistence needed (notifications already in DB)
- **Predictable Scaling**: Fixed subscription count independent of user count

**Event Flow** (Shard-Based):

```
Instance A: Startup → SSE Hub subscribes to all 256 shard channels
Instance B: Creates notification for user → hash(userID) % 256 = shard 42
Instance B: Publishes to notification:shard:42 with userID in payload
Instance A: Receives on shard 42 → Filters by userID → Broadcasts if user connected locally
```

**Implementation**: `pkg/eventbus/redis_eventbus.go`

---

### 6. Notification Service (CRUD Operations)

**Responsibility**: Provides CRUD operations for user notifications via REST API.

**Operations**:

- `ListNotifications()`: Paginated list with cursor-based pagination
- `ListUnreadNotifications()`: Filtered list of unread notifications
- `GetNotification()`: Retrieve single notification by ID
- `GetUnreadCount()`: Count of unread notifications
- `MarkAsRead()`: Mark single notification as read
- `MarkAllAsRead()`: Mark all notifications as read for user
- `DeleteNotification()`: Soft delete notification
- `DeleteAllNotifications()`: Delete all notifications for user

**Implementation**: `internal/service/notification_service.go`

---

## Data Flow

### Inbound Flow: Kafka → Database → SSE

```
1. Event Published (Order Service)
   ↓
2. Kafka Consumer receives event
   ↓
3. Store in inbox_events table (Transactional Inbox)
   ├─ message_id: UUID (deduplication)
   ├─ event_type: notification.requested
   ├─ payload: JSON
   ├─ status: pending
   └─ scheduled_for: NOW()
   ↓
4. Inbox Processor polls (every 5s)
   ↓
5. Mark as 'processing'
   ↓
6. Route to NotificationEventService
   ├─ Email notification → SMTP send
   └─ Push notification → SSE broadcast
      ↓
7. Create notification in database
   ├─ INSERT INTO notifications
   └─ Returns saved notification entity
   ↓
8. Broadcast via SSE Hub
   ├─ Create SSE Message
   ├─ BroadcastToUser(userID, message)
   │  ├─ Send to local SSE connections
   │  └─ Publish to Redis: notification:shard:{N} (where N = hash(userID) % 256)
   └─ Done
   ↓
9. Mark inbox event as 'processed'
   ↓
10. Cleanup: Delete processed events after 7 days
```

### Cross-Instance Fan-Out: Redis Pub/Sub

```
┌─────────────────────────────────────────────────────────────┐
│                    Instance #1                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ User A connects to SSE endpoint                     │   │
│  │   ↓                                                  │   │
│  │ SSEHub.Register(connectionA)                        │   │
│  │   ├─ Add to connections map                         │   │
│  │   ├─ Add to userConns[userA]                        │   │
│  │   └─ First connection? → Subscribe to Redis        │   │
│  │      Channel: notification:user:{userA}             │   │
│  └─────────────────────────────────────────────────────┘   │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Redis Channel: notification:user:{userA}
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    Instance #2                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Notification created for User A                     │   │
│  │   ↓                                                  │   │
│  │ notificationRepo.Create(notification)               │   │
│  │   ├─ Save to PostgreSQL                             │   │
│  │   └─ Returns saved notification                     │   │
│  │   ↓                                                  │   │
│  │ sseHub.BroadcastToUser(userA, message)              │   │
│  │   ├─ No local connections for userA                 │   │
│  │   └─ Publish to Redis anyway                        │   │
│  │      ↓                                               │   │
│  │ eventBus.Publish(                                   │   │
│  │   channel: "notification:user:{userA}",             │   │
│  │   event: BaseEvent{                                 │   │
│  │     sourceInstanceID: "instance-2",                 │   │
│  │     eventType: "notification_created",              │   │
│  │     payload: {userID, message}                      │   │
│  │   }                                                  │   │
│  │ )                                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                        │
                        │ Redis broadcasts to all subscribers
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                    Instance #1                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Redis event received on notification:user:{userA}   │   │
│  │   ↓                                                  │   │
│  │ createRedisEventHandler(userA)                      │   │
│  │   ├─ Check sourceInstanceID != our instanceID      │   │
│  │   ├─ Unmarshal SSE Message from payload             │   │
│  │   └─ BroadcastToUser(userA, message)                │   │
│  │      ↓                                               │   │
│  │ Send to all local connections for userA             │   │
│  │   └─ connectionA.Send(message)                      │   │
│  │      ↓                                               │   │
│  │ SSE stream to client:                               │   │
│  │   id: {uuid}                                        │   │
│  │   event: notification_created                       │   │
│  │   data: {notification JSON}                         │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                        │
                        │ SSE Connection
                        │
                 ┌──────▼──────┐
                 │   Browser   │
                 │  (User A)   │
                 └─────────────┘
```

**Key Points** (Shard-Based Default):

- All instances subscribe to all 256 shard channels at startup
- Instance 2 creates notification for User A → calculates shard: `hash(userA) % 256 = 42`
- Instance 2 publishes to Redis shard channel `notification:shard:42` with userID in payload
- All instances receive the message on shard 42
- Each instance filters by userID: only Instance 1 (with User A connected) broadcasts to SSE
- User A receives real-time notification with zero subscription overhead

---

## SSE and Redis Pub/Sub Integration

### Universal SSE Package Design

The `pkg/sse` package is designed to be **universal** and reusable across any service:

```go
// Initialize SSE Hub
sseHub := sse.NewHub(logger)

// Configure with EventBus and channel naming strategy
sseHub.SetEventBus(
    eventBus,                      // Redis EventBus instance
    instanceID,                    // Unique instance identifier
    redis.NotificationChannel,     // Per-user channel builder
    redis.NotificationShardChannel,// Shard channel builder
    strategy,                      // Channel strategy
    shardCount,                    // Number of shards
)
```

**Channel Builder Pattern**:

```go
// Each service can define its own channel naming strategy
func NotificationChannel(userID uuid.UUID) string {
    return fmt.Sprintf("notification:user:%s", userID)
}

func ChatChannel(conversationID uuid.UUID) string {
    return fmt.Sprintf("chat:conversation:%s", conversationID)
}
```

### Dynamic Subscription Management

**Subscribe on Connect**:

```go
func (h *Hub) handleRegister(conn *Connection) {
    // ... add to maps ...

    // Subscribe to Redis if first connection for this user
    if h.eventBus != nil && len(h.userConns[userID]) == 1 {
        h.subscribeToUserChannel(userID)
    }
}
```

**Unsubscribe on Disconnect**:

```go
func (h *Hub) handleUnregister(conn *Connection) {
    // ... remove from maps ...

    // Unsubscribe from Redis if last connection
    if len(userMap) == 0 {
        if h.eventBus != nil {
            h.unsubscribeFromUserChannel(userID)
        }
    }
}
```

**Benefits**:

- **Resource Efficiency**: Only subscribe to channels with active connections
- **Memory Savings**: No unnecessary Redis subscriptions
- **Automatic Cleanup**: Subscriptions automatically cleaned up on disconnect

### Event Deduplication

**Prevent Broadcast Loops**:

```go
func (h *Hub) createRedisEventHandler(userID uuid.UUID) eventbus.EventHandler {
    return func(ctx context.Context, event eventbus.Event) error {
        // Skip events from our own instance
        if event.GetSourceInstanceID() == h.instanceID {
            return nil
        }

        // Process and broadcast to local connections
        // ...
    }
}
```

Each instance publishes with its own `instanceID`, and receivers ignore messages from their own instance.

---

## Scalability Strategy

### Horizontal Scaling

The notification service is designed to scale horizontally:

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Instance 1  │  │  Instance 2  │  │  Instance 3  │  │  Instance N  │
│  (10k users) │  │  (10k users) │  │  (10k users) │  │  (10k users) │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │                 │
       └─────────────────┴─────────────────┴─────────────────┘
                              │
                    ┌─────────▼──────────┐
                    │   Redis Cluster    │
                    │ (Pub/Sub Fan-out)  │
                    └────────────────────┘
```

**Load Distribution**:

- SSE connections distributed across instances via load balancer
- Each instance handles independent subset of users
- Redis pub/sub ensures messages reach correct instance

### Subscription Scalability

**Challenge**: With millions of users, subscribing to all channels is not feasible.

**Solution**: Dynamic, on-demand subscriptions

- **Subscribe**: Only when user connects
- **Unsubscribe**: When user disconnects
- **Per-Instance**: Each instance only subscribes to channels for connected users

**Example**:

```
Instance 1: 10,000 connected users → 10,000 Redis subscriptions
Instance 2: 10,000 connected users → 10,000 Redis subscriptions
Total: 20,000 subscriptions (not 1,000,000)
```

### Redis Pub/Sub Performance

**Redis Pub/Sub Characteristics**:

- **Low Latency**: <1ms message delivery
- **High Throughput**: 100k+ messages/second per channel
- **Memory Efficient**: Ephemeral messages (no persistence overhead)
- **Scalable**: Handles millions of channels

**Channel Isolation**:

- Each user has dedicated channel (`notification:user:{userID}`)
- No cross-user interference
- Efficient message routing

### Database Scalability

**Read Replicas**:

```
┌─────────────┐
│   Primary   │ ← Writes (Create, Update, Delete)
│  PostgreSQL │
└──────┬──────┘
       │ Replication
       ├──────────────┬──────────────┐
       │              │              │
 ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐
 │ Replica 1│  │ Replica 2│  │ Replica N│
 └──────────┘  └──────────┘  └──────────┘
       ↑              ↑              ↑
       └──────────────┴──────────────┘
         Reads (List, Get, Count)
```

**Sharding Strategy** (Future):

- Shard by `user_id` hash
- Each shard handles subset of users
- Maintains data locality

### Kafka Consumer Scalability

**Consumer Group Partitioning**:

```
Topic: order.lifecycle (10 partitions)

┌─────────────────────────────────────────┐
│     Kafka Consumer Group                │
│  (notification-service-consumer)        │
│                                         │
│  ┌──────────┐  ┌──────────┐  ┌────┐   │
│  │Instance 1│  │Instance 2│  │... │   │
│  │ Part 0-2 │  │ Part 3-5 │  │... │   │
│  └──────────┘  └──────────┘  └────┘   │
└─────────────────────────────────────────┘
```

**Benefits**:

- Parallel consumption across instances
- Automatic partition rebalancing
- Each partition processed by single instance (ordering guarantee)

---

## Channel Strategy Comparison

The notification service supports two Redis pub/sub channel strategies: **per-user channels** and **shard-based channels**. The choice depends on your scale, user behavior, and infrastructure constraints.

### Per-User Channel Strategy

**Design**: One Redis channel per user (`notification:user:{userID}`)

**Subscription Model**: Dynamic subscriptions

- Subscribe when user connects (first SSE connection)
- Unsubscribe when user disconnects (last SSE connection)
- Each instance subscribes only to channels for locally connected users

**Architecture**:

```
┌────────────────────────────────────────────────┐
│ Instance 1 (10,000 connected users)            │
│                                                │
│ Redis Subscriptions:                           │
│  notification:user:abc-123                     │
│  notification:user:def-456                     │
│  ... (10,000 subscriptions)                    │
└────────────────────────────────────────────────┘

┌────────────────────────────────────────────────┐
│ Instance 2 (5,000 connected users)             │
│                                                │
│ Redis Subscriptions:                           │
│  notification:user:ghi-789                     │
│  notification:user:jkl-012                     │
│  ... (5,000 subscriptions)                     │
└────────────────────────────────────────────────┘

Total Redis Subscriptions: 15,000
(Only for connected users, not all users)
```

**Message Publishing**:

```go
// Per-user strategy: Direct channel per user
channelName := redis.NotificationChannel(userID)
// Result: "notification:user:{userID}"

event := &notification.CreatedEvent{
    UserID:  userID,
    Message: sseMsg,
}

eventBus.Publish(ctx, channelName, event)
```

**Characteristics**:

| Metric                  | Per-User Channel                          |
| ----------------------- | ----------------------------------------- |
| Subscriptions per instance | O(N) where N = connected users         |
| Connection churn        | High (SUBSCRIBE/UNSUBSCRIBE on every connect/disconnect) |
| Redis memory            | O(N) subscription records per instance    |
| Messages received       | Only messages for connected users         |
| Bandwidth efficiency    | High (no wasted bandwidth)                |
| Latency                 | Low (<1ms delivery)                       |
| Best for                | < 10K concurrent users per instance       |

**Pros**:

- ✅ **Efficient bandwidth**: Instances only receive messages for connected users
- ✅ **No message filtering**: All received messages are relevant
- ✅ **Simple message structure**: Direct user-to-channel mapping
- ✅ **Low memory at scale**: Only subscribed channels consume memory

**Cons**:

- ❌ **Subscription churn**: Frequent SUBSCRIBE/UNSUBSCRIBE operations
- ❌ **Not viable at massive scale**: 100K+ concurrent users = 100K+ subscriptions per instance
- ❌ **Connection overhead**: Redis maintains state for each subscription

### Shard-Based Channel Strategy

**Design**: Fixed number of shard channels (`notification:shard:0` to `notification:shard:N`)

**Subscription Model**: Static subscriptions

- All instances subscribe to ALL shards at startup (one-time operation)
- No dynamic subscriptions (zero SUBSCRIBE/UNSUBSCRIBE churn)
- Client-side filtering: Only broadcast to locally connected users

**Architecture**:

```
┌────────────────────────────────────────────────┐
│ Instance 1 (any number of connected users)     │
│                                                │
│ Redis Subscriptions (static, at startup):      │
│  notification:shard:0                          │
│  notification:shard:1                          │
│  ... notification:shard:255                    │
│                                                │
│ Total: 256 subscriptions (fixed)               │
└────────────────────────────────────────────────┘

┌────────────────────────────────────────────────┐
│ Instance 2 (any number of connected users)     │
│                                                │
│ Redis Subscriptions (static, at startup):      │
│  notification:shard:0                          │
│  notification:shard:1                          │
│  ... notification:shard:255                    │
│                                                │
│ Total: 256 subscriptions (fixed)               │
└────────────────────────────────────────────────┘

Total Redis Subscriptions: 256 × N instances
(Independent of user count)
```

**User Distribution**:

```
User Distribution (Consistent Hashing):
- hash(userID) % shardCount = shardID

Example with 256 shards:
  userID: abc-123 → hash % 256 = 42  → notification:shard:42
  userID: def-456 → hash % 256 = 117 → notification:shard:117
  userID: ghi-789 → hash % 256 = 3   → notification:shard:3
```

**Message Publishing**:

```go
// Shard-based strategy: Calculate shard for user
shardID := redis.GetUserShard(userID, shardCount)
channelName := redis.NotificationShardChannel(shardID)
// Result: "notification:shard:42"

// Include userID in payload for filtering
event := struct {
    UserID  uuid.UUID    `json:"user_id"`
    Message *sse.Message `json:"message"`
}{
    UserID:  userID,
    Message: sseMsg,
}

eventBus.Publish(ctx, channelName, event)
```

**Message Filtering**:

```go
// SSE Hub receives message from shard channel
func (h *Hub) createShardEventHandler(shardID int) eventbus.EventHandler {
    return func(ctx context.Context, event eventbus.Event) error {
        // Unmarshal includes userID
        var msg struct {
            UserID  uuid.UUID `json:"user_id"`
            Message Message   `json:"message"`
        }
        event.UnmarshalPayload(&msg)

        // Filter: Only broadcast if user is locally connected
        if _, hasLocalConnections := h.userConns[msg.UserID]; hasLocalConnections {
            h.BroadcastToUser(msg.UserID, &msg.Message)
        }
        // Else: Ignore message (user not connected to this instance)

        return nil
    }
}
```

**Characteristics**:

| Metric                  | Shard-Based Channel                       |
| ----------------------- | ----------------------------------------- |
| Subscriptions per instance | O(S) where S = shard count (typically 256) |
| Connection churn        | Zero (subscriptions are static)           |
| Redis memory            | O(S) subscription records per instance    |
| Messages received       | All messages across all users             |
| Bandwidth efficiency    | Lower (receives messages for non-connected users) |
| Latency                 | Low (<1ms delivery)                       |
| Best for                | > 10K concurrent users per instance       |

**Pros**:

- ✅ **Zero subscription churn**: Subscriptions never change after startup
- ✅ **Scales to millions**: Subscription count independent of user count
- ✅ **Predictable resource usage**: Fixed O(256) subscriptions per instance
- ✅ **Simple operations**: No dynamic subscribe/unsubscribe logic

**Cons**:

- ❌ **Bandwidth waste**: Instances receive messages for non-connected users
- ❌ **Requires filtering**: Must check if user has local connections
- ❌ **Larger message payload**: Must include userID in every message
- ❌ **More complex debugging**: Messages fanout to all instances

### Configuration

**Code Configuration** (`notification-service/internal/provider/init.go`):

```go
// Shard-based channel strategy (default and recommended)
shardCount := pkgconstant.SSEShardCount // 256

sseHub.SetEventBus(
    eventBus,
    instanceID,
    redis.NotificationShardChannel, // Shard channel builder
    shardCount,
    eventHandler,                   // Service-specific event handler
)
```

**Why Shard-Based is Default**:

- ✅ Concurrent users per instance: **Scales to millions**
- ✅ Connection pattern: **Handles both long-lived and short-lived**
- ✅ Message volume: **High throughput**
- ✅ Priority: **Scalability and predictability**

### Performance Comparison

**Scenario**: 50,000 concurrent users, 5 instances, 10,000 users per instance

| Metric                  | Per-User                | Shard-Based (256 shards) |
| ----------------------- | ----------------------- | ------------------------ |
| Subscriptions per instance | 10,000                  | 256                      |
| Total subscriptions     | 50,000                  | 1,280                    |
| Subscribe operations    | High (every connect)    | None (static)            |
| Unsubscribe operations  | High (every disconnect) | None (static)            |
| Messages received per notification | 1 (target instance only) | 5 (all instances) |
| Bandwidth per notification | 1× message size         | 5× message size          |
| CPU for filtering       | None                    | Minimal (map lookup)     |

**Bandwidth Calculation Example**:

- Notification rate: 100 notifications/second
- Message size: 1 KB

**Per-User**:

- Messages received: 100/sec × 1 KB = 100 KB/sec per instance
- Total: 100 KB/sec × 5 instances = 500 KB/sec

**Shard-Based**:

- Messages received: 100/sec × 5 instances × 1 KB = 500 KB/sec per instance
- Total: 500 KB/sec × 5 instances = 2.5 MB/sec (5× bandwidth)

**Verdict**: Shard-based uses 5× bandwidth but eliminates subscription churn entirely. At massive scale (100K+ users), the trade-off favors shard-based.

### Implementation Details

**Shard Count Selection**:

- **256 shards** (default): Good balance for most use cases
- **1,024 shards**: Better distribution, still manageable subscriptions
- **4,096 shards**: Maximum distribution, approaching per-user granularity

**Hash Function**:

```go
func GetUserShard(userID uuid.UUID, shardCount int) int {
    hash := uint32(0)
    uuidBytes := [16]byte(userID)
    for _, b := range uuidBytes {
        hash = hash*31 + uint32(b)
    }
    return int(hash % uint32(shardCount))
}
```

**Properties**:

- Deterministic: Same userID always maps to same shard
- Uniform distribution: Users evenly distributed across shards
- Fast: O(1) computation

### Migration Path

The notification service supports BOTH strategies via configuration, allowing:

1. **A/B Testing**: Run instances with different strategies
2. **Gradual Migration**: Switch strategy without code changes
3. **Environment-Specific**: Use per-user in development, shard-based in production

**Migration Steps**:

```bash
# Phase 1: Deploy code supporting both strategies
task deploy SERVICE=notification-service

# Phase 2: Update configuration (rolling deployment)
NOTIFICATION_CHANNEL_STRATEGY=shard-based
NOTIFICATION_SHARD_COUNT=256

# Phase 3: Monitor metrics
- Check sse_subscription_count (should drop to 256)
- Check redis_pubsub_messages_received (should increase)
- Check sse_messages_sent_total (should remain same)

# Phase 4: Validate (both strategies coexist during rollout)
# No disruption to SSE connections
# No dropped notifications
```

---

## Database Schema

### Table: `inbox_events`

**Purpose**: Transactional inbox for exactly-once event processing.

```sql
CREATE TABLE inbox_events (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL UNIQUE,      -- Deduplication key
    aggregate_type TEXT NOT NULL,         -- 'order', 'user', etc.
    aggregate_id UUID NOT NULL,           -- ID from source service
    event_type TEXT NOT NULL,             -- 'notification.requested', etc.
    topic TEXT NOT NULL,                  -- Kafka topic
    source_service TEXT NOT NULL,         -- 'order-service', etc.
    payload JSONB NOT NULL,               -- Full event payload
    status TEXT NOT NULL DEFAULT 'pending', -- State machine
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attempts INTEGER DEFAULT 0,
    last_error TEXT,
    correlation_id UUID,
    causation_id UUID
);
```

**Status Values**:

- `pending`: Ready for processing
- `processing`: Currently being processed
- `processed`: Successfully completed
- `retry`: Scheduled for retry after failure
- `failed`: Permanently failed after max retries
- `duplicate`: Duplicate message (already processed)

**Indexes**:

```sql
CREATE UNIQUE INDEX idx_inbox_message_id ON inbox_events(message_id);
CREATE INDEX idx_inbox_status_scheduled ON inbox_events(status, scheduled_for);
CREATE INDEX idx_inbox_aggregate_type_id ON inbox_events(aggregate_type, aggregate_id);
```

---

### Table: `notifications`

**Purpose**: Persistent storage of user notifications.

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type TEXT NOT NULL,                   -- 'order.confirmed', 'user.verified', etc.
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,                       -- Additional structured data
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Indexes**:

```sql
-- List notifications for user
CREATE INDEX idx_notifications_user_created
    ON notifications (user_id, created_at DESC);

-- List unread notifications (partial index)
CREATE INDEX idx_notifications_user_unread
    ON notifications (user_id, is_read)
    WHERE is_read = false;

-- General user lookup
CREATE INDEX idx_notifications_user_id
    ON notifications(user_id);
```

**Why These Indexes?**:

- `user_id + created_at DESC`: Optimizes paginated queries (`ListNotifications`)
- Partial index on unread: Smaller, faster index for common query
- Composite index reduces index size vs. separate indexes

---

## Delivery Guarantees

### At-Least-Once from Kafka

**Kafka Consumer Configuration**:

```yaml
enable.auto.commit: false
isolation.level: read_committed
```

- Manual offset commit after successful inbox write
- Consumer may re-deliver messages on crash/restart
- Inbox pattern handles deduplication

### Exactly-Once via Inbox Pattern

**Write Path**:

```
1. Begin transaction
2. Check if message_id exists in inbox_events
   ├─ If exists: Mark as 'duplicate', commit, return (idempotent)
   └─ If not exists: Insert into inbox_events
3. Commit transaction
4. Commit Kafka offset
```

**Process Path**:

```
1. Begin transaction
2. UPDATE inbox_events SET status = 'processing' WHERE id = ?
3. Commit transaction
4. Process business logic (create notification, send email, broadcast SSE)
   ├─ Success: UPDATE inbox_events SET status = 'processed'
   └─ Failure: UPDATE inbox_events SET status = 'retry', attempts++
```

**Deduplication Guarantees**:

- `message_id` from Kafka metadata is globally unique
- UNIQUE constraint on `message_id` prevents duplicates
- Idempotent: Processing same message multiple times has same result

### Retry Strategy

**Exponential Backoff**:

```
Attempt 1: Immediate processing
Attempt 2: 2s delay  (2^1 × 1s base backoff)
Attempt 3: 4s delay  (2^2 × 1s)
Attempt 4: 8s delay  (2^3 × 1s)
Attempt 5: 16s delay (2^4 × 1s)
Attempt 6+: Mark as 'failed'
```

**Configurable Parameters**:

- `max_retry_attempts`: 5 (default)
- `retry_backoff`: 1 second (default)
- Backoff is exponential to avoid overwhelming downstream services

### Cleanup Strategy

**Retention Policy**:

- Processed events: Deleted after 7 days (configurable)
- Failed events: Retained indefinitely for manual investigation
- Pending/Retry events: Not deleted (active processing)

**Cleanup Query**:

```sql
DELETE FROM inbox_events
WHERE status = 'processed'
  AND processed_at < NOW() - INTERVAL '7 days';
```

---

## API Endpoints

### SSE Endpoint

**`GET /api/v1/notifications/stream`**

Establishes Server-Sent Events connection for real-time push notifications.

**Headers**:

```
Authorization: Bearer {jwt_token}
Accept: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
```

**Response Stream**:

```
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

event: connected
data: {"status":"connected"}

: heartbeat

event: notification_created
id: 123e4567-e89b-12d3-a456-426614174000
data: {"id":"...","type":"order.confirmed","title":"Order Confirmed","message":"Your order #1234 has been confirmed.","created_at":"2025-10-11T10:30:00Z"}

: heartbeat
```

**Events**:

- `connected`: Initial connection established
- `notification_created`: New notification created
- `notification_read`: Notification marked as read (future)
- `notification_deleted`: Notification deleted (future)

---

### REST Endpoints

#### List Notifications

**`GET /api/v1/notifications`**

Query Parameters:

- `limit`: Number of results (default: 20, max: 100)
- `cursor`: Pagination cursor from previous response

Response:

```json
{
  "data": [
    {
      "id": "uuid",
      "type": "order.confirmed",
      "title": "Order Confirmed",
      "message": "Your order #1234 has been confirmed.",
      "metadata": {},
      "is_read": false,
      "created_at": "2025-10-11T10:30:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "encoded_cursor",
    "has_next": true
  }
}
```

---

#### List Unread Notifications

**`GET /api/v1/notifications/unread`**

Same structure as List Notifications, filtered to `is_read = false`.

---

#### Get Unread Count

**`GET /api/v1/notifications/unread/count`**

Response:

```json
{
  "count": 5
}
```

---

#### Get Single Notification

**`GET /api/v1/notifications/:notificationID`**

Response:

```json
{
  "id": "uuid",
  "type": "order.shipped",
  "title": "Order Shipped",
  "message": "Your order #1234 has been shipped. Tracking: ABC123",
  "metadata": {
    "order_id": "uuid",
    "tracking_number": "ABC123"
  },
  "is_read": false,
  "created_at": "2025-10-11T10:30:00Z",
  "updated_at": "2025-10-11T10:30:00Z"
}
```

---

#### Mark as Read

**`PUT /api/v1/notifications/:notificationID/read`**

Response: `204 No Content`

---

#### Mark All as Read

**`PUT /api/v1/notifications/read-all`**

Response: `204 No Content`

---

#### Delete Notification

**`DELETE /api/v1/notifications/:notificationID`**

Response: `204 No Content`

---

#### Delete All Notifications

**`DELETE /api/v1/notifications/all`**

Response: `204 No Content`

---

## Configuration

### Environment Variables

```yaml
# Application
APP_ENV: production
APP_PORT: 8080
APP_TIMEOUT_SHUTDOWN: 30s

# PostgreSQL
POSTGRES_HOST: localhost
POSTGRES_PORT: 5432
POSTGRES_USER: notification_user
POSTGRES_PASSWORD: secret
POSTGRES_DB: notification_db
POSTGRES_MAX_OPEN_CONNS: 25
POSTGRES_MAX_IDLE_CONNS: 5

# Redis Cluster
REDIS_ADDRS: redis-1:6379,redis-2:6379,redis-3:6379
REDIS_PASSWORD: secret
REDIS_MAX_ACTIVE_CONN: 100
REDIS_MAX_IDLE_CONN: 20

# Kafka
KAFKA_BROKERS: kafka-1:9092,kafka-2:9092,kafka-3:9092
KAFKA_CONSUMER_GROUP: notification-service-consumer
KAFKA_TOPICS: order.lifecycle,user.verification,notification.requests

# Inbox Processor
INBOX_POLL_INTERVAL: 5s
INBOX_BATCH_SIZE: 100
INBOX_MAX_RETRY_ATTEMPTS: 5
INBOX_RETRY_BACKOFF: 1s
INBOX_RETENTION_PERIOD: 168h # 7 days
INBOX_CLEANUP_INTERVAL: 1h

# SMTP
SMTP_HOST: smtp.gmail.com
SMTP_PORT: 587
SMTP_EMAIL: notifications@example.com
SMTP_PASSWORD: app_password

# Notification Channel Strategy
NOTIFICATION_CHANNEL_STRATEGY: shard-based # or "per-user"
NOTIFICATION_SHARD_COUNT: 256 # Only used for shard-based strategy
```

---

## Monitoring and Observability

### Key Metrics

**SSE Metrics**:

- `sse_active_connections`: Number of active SSE connections
- `sse_unique_users`: Number of unique users connected
- `sse_messages_sent_total`: Total SSE messages sent
- `sse_subscription_count`: Active Redis subscriptions

**Inbox Metrics**:

- `inbox_events_pending`: Events waiting for processing
- `inbox_events_processing`: Events currently being processed
- `inbox_events_processed_total`: Total successful events
- `inbox_events_failed_total`: Total failed events
- `inbox_processing_duration_seconds`: Processing time histogram

**Redis Pub/Sub Metrics**:

- `redis_pubsub_channels`: Number of subscribed channels
- `redis_pubsub_messages_received`: Messages received from Redis
- `redis_pubsub_messages_published`: Messages published to Redis

**Notification Metrics**:

- `notifications_created_total`: Total notifications created
- `notifications_sent_email_total`: Total email notifications sent
- `notifications_sent_push_total`: Total push notifications sent

### Logging

**Structured Logging**:

```json
{
  "level": "info",
  "ts": "2025-10-11T10:30:00Z",
  "msg": "SSE connection registered",
  "connection_id": "uuid",
  "user_id": "uuid",
  "instance_id": "instance-1",
  "total_connections": 1523
}
```

**Key Log Events**:

- SSE connection registered/unregistered
- Redis channel subscribed/unsubscribed
- Inbox event processing started/completed/failed
- Notification created and broadcasted
- Email sent successfully/failed

### Distributed Tracing

**OpenTelemetry Integration**:

- Trace ID propagation through Kafka metadata
- Spans for inbox processing, notification creation, SSE broadcast
- Redis pub/sub message tracing

---

## Possible Improvements

### 1. WebSocket Support

**Current**: SSE (Server-Sent Events) - unidirectional
**Improvement**: Add WebSocket support for bidirectional communication

**Benefits**:

- Client can send read receipts via WebSocket
- Real-time typing indicators
- Presence detection (online/offline status)

**Implementation**:

```go
// pkg/websocket package (similar to pkg/sse)
type Hub struct {
    // ... similar to SSE Hub ...
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, message Message)
func (h *Hub) ReceiveFromUser(userID uuid.UUID) chan Message
```

---

### 2. FCM/APNS Mobile Push

**Current**: Web-only push via SSE
**Improvement**: Add Firebase Cloud Messaging (FCM) and Apple Push Notification Service (APNS)

**Architecture**:

```
┌─────────────────────────────────────────┐
│   Notification Event Service            │
│                                         │
│   switch notificationType {            │
│     case Push:                          │
│       ├─ Send via SSE (web)             │
│       ├─ Send via FCM (Android)         │
│       └─ Send via APNS (iOS)            │
│   }                                     │
└─────────────────────────────────────────┘
```

**Device Token Management**:

```sql
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    platform TEXT NOT NULL, -- 'web', 'android', 'ios'
    token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ
);
```

---

### 3. Notification Preferences

**User-Configurable Delivery Channels**:

```sql
CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY,
    order_confirmed_push BOOLEAN DEFAULT TRUE,
    order_confirmed_email BOOLEAN DEFAULT TRUE,
    order_shipped_push BOOLEAN DEFAULT TRUE,
    order_shipped_email BOOLEAN DEFAULT FALSE,
    -- ... more preferences ...
);
```

**Service Layer**:

```go
func (s *notificationEventService) sendNotification(
    userID uuid.UUID,
    notifType string,
    message Message,
) {
    prefs := s.getPreferences(userID)

    if prefs.ShouldSendPush(notifType) {
        s.sendPush(userID, message)
    }

    if prefs.ShouldSendEmail(notifType) {
        s.sendEmail(userID, message)
    }
}
```

---

### 4. Read Receipt Broadcasting

**Current**: Read status updates via REST API
**Improvement**: Broadcast read receipts to all user's connected devices

**Flow**:

```
Device A: Marks notification as read
   ↓
REST API: PUT /notifications/:id/read
   ↓
Update database
   ↓
Publish to Redis: notification:user:{userID}
   ↓
Device B (same user): Receives read event via SSE
   ↓
Update UI: Show notification as read
```

**Event Types**:

- `notification_created`
- `notification_read` (NEW)
- `notification_deleted` (NEW)

---

### 5. Rate Limiting

**Problem**: Prevent notification flooding (e.g., 100 notifications in 1 second)

**Solution**: Implement rate limiting per user

**Implementation**:

```go
// Redis-based rate limiter
func (s *notificationEventService) checkRateLimit(userID uuid.UUID) error {
    key := fmt.Sprintf("rate_limit:user:%s", userID)
    count, _ := redis.Incr(key)

    if count == 1 {
        redis.Expire(key, 60 * time.Second) // 1 minute window
    }

    if count > 100 { // Max 100 notifications per minute
        return ErrRateLimitExceeded
    }

    return nil
}
```

**Alternatives**:

- Token bucket algorithm
- Sliding window rate limiting
- Priority-based throttling (critical notifications bypass rate limit)

---

### 6. Notification Batching

**Problem**: Too many individual notifications for related events

**Example**:

```
Instead of:
- "John liked your post"
- "Jane liked your post"
- "Bob liked your post"

Batch as:
- "John, Jane, and Bob liked your post"
```

**Implementation**:

```go
type NotificationBatcher struct {
    window time.Duration
    batch  map[string][]*Notification
}

func (b *NotificationBatcher) Add(notif *Notification) {
    key := fmt.Sprintf("%s:%s", notif.UserID, notif.Type)
    b.batch[key] = append(b.batch[key], notif)

    // Flush after window expires
    time.AfterFunc(b.window, func() {
        b.Flush(key)
    })
}
```

---

### 7. Notification Templates

**Current**: Message formatting in service layer
**Improvement**: Centralized template management

```sql
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL UNIQUE,
    title_template TEXT NOT NULL,
    message_template TEXT NOT NULL,
    variables JSONB, -- Required template variables
    created_at TIMESTAMPTZ
);
```

**Template Rendering**:

```go
func (s *notificationEventService) renderTemplate(
    templateType string,
    variables map[string]interface{},
) (*Notification, error) {
    template := s.getTemplate(templateType)

    title := s.render(template.TitleTemplate, variables)
    message := s.render(template.MessageTemplate, variables)

    return &Notification{
        Type:    templateType,
        Title:   title,
        Message: message,
    }, nil
}
```

---

### 8. Dead Letter Queue (DLQ)

**Current**: Failed inbox events marked as 'failed'
**Improvement**: Move permanently failed events to DLQ for manual review

**Implementation**:

```sql
CREATE TABLE notification_dlq (
    id UUID PRIMARY KEY,
    original_event_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    failure_reason TEXT NOT NULL,
    attempts INTEGER NOT NULL,
    failed_at TIMESTAMPTZ NOT NULL,
    investigated BOOLEAN DEFAULT FALSE
);
```

**Admin UI**:

- View all DLQ entries
- Retry individual event
- Mark as investigated
- Export for debugging

---

### 9. Notification Analytics

**Track User Engagement**:

```sql
CREATE TABLE notification_analytics (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL,
    user_id UUID NOT NULL,
    delivered_at TIMESTAMPTZ,
    opened_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ
);
```

**Metrics**:

- Delivery rate
- Open rate
- Click-through rate
- Time-to-read distribution

---

### 10. Notification Archiving

**Current**: Notifications stored indefinitely
**Improvement**: Archive old notifications to cold storage

**Strategy**:

```
Hot Storage (PostgreSQL): Last 30 days
Warm Storage (S3 + Athena): 30-365 days
Cold Storage (Glacier): >365 days
```

**Implementation**:

```go
func (s *notificationService) ArchiveOldNotifications(ctx context.Context) {
    threshold := time.Now().AddDate(0, 0, -30) // 30 days ago

    notifications := s.repo.FindOlderThan(ctx, threshold)

    // Export to S3
    s3.Upload(notifications)

    // Delete from PostgreSQL
    s.repo.DeleteOlderThan(ctx, threshold)
}
```

---

### 11. Multi-Tenancy Support

**Isolate Notifications by Tenant**:

```sql
ALTER TABLE notifications ADD COLUMN tenant_id UUID NOT NULL;
CREATE INDEX idx_notifications_tenant_user ON notifications(tenant_id, user_id);
```

**Redis Channels**:

```
notification:tenant:{tenantID}:user:{userID}
```

---

### 12. Notification Grouping/Threading

**Group Related Notifications**:

```sql
ALTER TABLE notifications ADD COLUMN thread_id UUID;
CREATE INDEX idx_notifications_thread ON notifications(user_id, thread_id);
```

**Example**:

```
Thread: "Order #1234"
├─ Order Confirmed
├─ Order Shipped
├─ Out for Delivery
└─ Order Delivered
```

---

### 13. Smart Notification Scheduling

**Respect User Timezone and Preferences**:

```sql
CREATE TABLE notification_schedules (
    user_id UUID PRIMARY KEY,
    timezone TEXT NOT NULL,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    weekends_enabled BOOLEAN
);
```

**Delay Non-Critical Notifications**:

- Don't send marketing notifications during quiet hours
- Queue for next available time window
- Critical notifications (security alerts) bypass schedule

---

### 14. A/B Testing for Notifications

**Test Different Message Variants**:

```sql
CREATE TABLE notification_experiments (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL,
    variant_a TEXT NOT NULL,
    variant_b TEXT NOT NULL,
    conversion_metric TEXT, -- 'click', 'read', 'action_taken'
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ
);
```

**Track Results**:

- Assign users to variant A or B
- Measure conversion rates
- Automatically select winning variant

---

### 15. Notification Actions

**Interactive Notifications**:

```json
{
  "id": "uuid",
  "title": "Order Shipped",
  "message": "Your order #1234 has shipped",
  "actions": [
    {
      "id": "track",
      "label": "Track Package",
      "url": "/orders/1234/track"
    },
    {
      "id": "dismiss",
      "label": "Dismiss",
      "action": "dismiss"
    }
  ]
}
```

**Click Tracking**:

```sql
CREATE TABLE notification_actions (
    id UUID PRIMARY KEY,
    notification_id UUID NOT NULL,
    action_id TEXT NOT NULL,
    clicked_at TIMESTAMPTZ NOT NULL
);
```

---

## Conclusion

The Notification Service implements a robust, scalable architecture for real-time push notifications with:

✅ **Exactly-Once Delivery** via Inbox Pattern
✅ **Horizontal Scalability** with Redis Pub/Sub fan-out and shard-based channels
✅ **Real-Time Streaming** via Server-Sent Events
✅ **Persistent Storage** for offline retrieval
✅ **Retry Logic** with exponential backoff
✅ **Flexible Subscription Strategies** supporting both per-user and shard-based channels
✅ **Application-Layer Filtering** for efficient message routing

The **shard-based channel strategy** is the default configuration, providing:
- Zero subscription churn (static subscriptions at startup)
- Fixed O(256) subscriptions per instance (independent of user count)
- Predictable scaling to millions of concurrent users
- Application-layer filtering to minimize bandwidth waste

The universal SSE package (`pkg/sse`) is truly service-agnostic with:
- Configurable channel builders for both per-user and shard-based strategies
- No hardcoded service-specific logic
- Reusable across any service (notifications, chat, orders, etc.)

This makes it a powerful, flexible building block for real-time features throughout the microservices architecture.
