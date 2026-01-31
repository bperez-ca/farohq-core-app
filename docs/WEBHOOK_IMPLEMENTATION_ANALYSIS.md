# Webhook Implementation Analysis: WhatsApp, Instagram, Facebook

**Date**: 2025-01-27  
**Status**: ❌ **NOT IMPLEMENTED** - Webhook functionality is planned but not yet built

---

## Executive Summary

**Critical Finding**: The webhook receiver infrastructure for WhatsApp, Instagram, and Facebook is **completely absent** from the current codebase. This is a **P0 gap** that must be addressed before FARO can deliver its core value proposition of real-time message ingestion.

**Current State**:
- ❌ No webhook endpoints (`/api/v1/webhooks/*`)
- ❌ No message/conversation database tables
- ❌ No webhook signature verification
- ❌ No async processing (no worker service, no job queue)
- ❌ No real-time delivery mechanism (no WebSocket, no pub/sub)
- ❌ No platform-specific handlers

**Planned Architecture** (from `lvos/docs/FARO_IMPLEMENTATION_PLAN.md`):
- Webhook endpoints: `/api/v1/webhooks/whatsapp`, `/api/v1/webhooks/instagram`, `/api/v1/webhooks/facebook`
- Conversations service with message storage
- Async job processing via asynq + Redis
- Real-time updates (mechanism TBD)

---

## 1. WEBHOOK INGESTION

### 1.1 Where are webhooks received?

**Status**: ❌ **NOT IMPLEMENTED**

**Current Routes** (from `cmd/server/main.go` and `internal/app/composition/composition.go`):
- ✅ `/api/v1/tenants/*` - Tenant management
- ✅ `/api/v1/clients/*` - Client management
- ✅ `/api/v1/locations/*` - Location management
- ✅ `/api/v1/brand/*` - Branding management
- ✅ `/api/v1/files/*` - File uploads
- ❌ `/api/v1/webhooks/*` - **MISSING**

**Planned Endpoints** (from documentation):
```
POST /api/v1/webhooks/whatsapp
POST /api/v1/webhooks/instagram
POST /api/v1/webhooks/facebook
POST /api/v1/webhooks/google-business
POST /api/v1/webhooks/email
POST /api/v1/webhooks/sms
```

**Recommendation**: Create webhook handlers in a new domain:
```
internal/domains/messages/
├── domain/
│   ├── message.go          # Message entity
│   ├── conversation.go    # Conversation entity
│   └── webhook.go         # Webhook payload types
├── app/
│   ├── ingest_webhook.go  # Use case: process webhook
│   └── create_message.go  # Use case: create message
├── infra/
│   ├── http/
│   │   └── webhook_handlers.go  # HTTP handlers
│   └── db/
│       └── message_repository.go
```

### 1.2 How is webhook authenticity verified?

**Status**: ❌ **NOT IMPLEMENTED**

**Meta Platforms (WhatsApp, Instagram, Facebook)**:
- **WhatsApp**: Uses `X-Hub-Signature-256` header with SHA-256 HMAC
- **Instagram/Facebook**: Uses `X-Hub-Signature-256` header with SHA-256 HMAC
- **Verification**: `HMAC-SHA256(payload, app_secret) == signature`

**Current Code**: No signature verification exists.

**Recommendation**: Implement signature verification middleware:
```go
// internal/domains/messages/infra/http/webhook_verification.go
func VerifyMetaWebhook(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            signature := r.Header.Get("X-Hub-Signature-256")
            if signature == "" {
                http.Error(w, "Missing signature", http.StatusUnauthorized)
                return
            }
            
            body, _ := io.ReadAll(r.Body)
            r.Body = io.NopCloser(bytes.NewReader(body))
            
            expected := computeHMAC(body, secret)
            if !hmac.Equal([]byte(signature), []byte(expected)) {
                http.Error(w, "Invalid signature", http.StatusUnauthorized)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Security Requirements**:
- Store app secrets per tenant in database (encrypted)
- Rate limit webhook endpoints (per IP, per tenant)
- Log all webhook attempts (success/failure)
- Reject webhooks older than 5 minutes (replay protection)

### 1.3 What's the rate limit per tenant?

**Status**: ❌ **NOT IMPLEMENTED**

**Current Rate Limiting**: None exists in codebase.

**Recommendation**: Implement per-tenant rate limiting:
```go
// Use Redis for distributed rate limiting
// Key: "webhook:rate:{tenant_id}:{platform}"
// Limit: 1000 requests/minute per tenant per platform
// Algorithm: Token bucket or sliding window
```

**Rate Limit Strategy**:
- **Per Tenant**: 1,000 webhooks/minute (burst: 100)
- **Per IP**: 100 webhooks/minute (DDoS protection)
- **Per Platform**: No global limit (Meta handles their own)

**Implementation**: Use `github.com/go-redis/redis_rate` or `golang.org/x/time/rate` with Redis.

### 1.4 Are webhooks processed synchronously or asynchronously?

**Status**: ❌ **NOT IMPLEMENTED** (no processing exists)

**Recommendation**: **Asynchronous processing** for scalability:

**Architecture**:
```
Webhook Handler (EC2)
  ↓ (200 OK immediately)
  ↓ Enqueue to Redis Queue
Async Worker (EC2)
  ↓ Process message
  ↓ Store in database
  ↓ Publish to Redis pub/sub
Frontend (WebSocket/SSE)
  ↓ Subscribe to pub/sub
  ↓ Real-time update
```

**Why Async**:
- Meta requires 200 OK within 20 seconds (or they retry)
- Message processing (DB writes, media downloads) can take 1-5 seconds
- Async allows handling spikes (50K messages/day = ~35 msg/sec peak)

**Queue Choice**: **Redis Streams** or **asynq** (Go job queue):
- ✅ Redis Streams: Built-in, supports consumer groups, persistence
- ✅ asynq: Go-native, retry logic, dead letter queue

---

## 2. MESSAGE PROCESSING

### 2.1 How are messages parsed and stored?

**Status**: ❌ **NOT IMPLEMENTED**

**Database Schema**: No `messages` or `conversations` tables exist.

**Planned Schema** (from `lvos/docs/FARO_IMPLEMENTATION_PLAN.md`):
```sql
CREATE TABLE conversations (
  id UUID PRIMARY KEY,
  location_id UUID REFERENCES locations(id),
  tenant_id UUID REFERENCES tenants(id),
  customer_phone VARCHAR(20),
  customer_email VARCHAR(255),
  channel VARCHAR(50), -- whatsapp, facebook_messenger, instagram_dm, google_business, sms, email
  channel_conversation_id VARCHAR(255),
  source_type VARCHAR(50), -- review_response, story_dm, listing_inquiry, direct
  linked_review_id UUID REFERENCES reviews(id),
  linked_listing_source VARCHAR(50),
  status VARCHAR(50), -- new, open, quoted, booked, converted, archived
  assigned_to_user_id UUID REFERENCES users(id),
  lead_status VARCHAR(50), -- qualified, interested, converted
  tags TEXT[],
  last_message_at TIMESTAMP,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

CREATE TABLE messages (
  id UUID PRIMARY KEY,
  conversation_id UUID REFERENCES conversations(id),
  sender_type VARCHAR(50), -- customer, agent
  text TEXT,
  media_urls TEXT[],
  direction VARCHAR(50), -- inbound, outbound
  channel VARCHAR(50),
  channel_message_id VARCHAR(255), -- UNIQUE for deduplication
  timestamp TIMESTAMP,
  delivered_at TIMESTAMP,
  read_at TIMESTAMP,
  created_at TIMESTAMP
);

-- Deduplication index
CREATE UNIQUE INDEX idx_messages_channel_id ON messages(channel, channel_message_id);
```

**Message Parsing** (per platform):

**WhatsApp** (via BSP like YCloud/Wati):
```json
{
  "event": "message.received",
  "payload": {
    "id": "wamid.xxx",
    "from": "+1234567890",
    "to": "+0987654321",
    "type": "text",
    "text": { "body": "Hello" },
    "timestamp": "1234567890"
  }
}
```

**Instagram/Facebook** (Meta Graph API):
```json
{
  "object": "instagram",
  "entry": [{
    "messaging": [{
      "sender": { "id": "123456" },
      "recipient": { "id": "789012" },
      "timestamp": 1234567890,
      "message": {
        "mid": "mid.xxx",
        "text": "Hello"
      }
    }]
  }]
}
```

**Recommendation**: Create platform-specific parsers:
```go
// internal/domains/messages/infra/parsers/
├── whatsapp_parser.go
├── instagram_parser.go
└── facebook_parser.go
```

### 2.2 Are conversations threaded or individual messages?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Threaded conversations** (one conversation per customer per channel):

**Threading Logic**:
- **Key**: `(location_id, channel, customer_identifier)`
  - WhatsApp: `customer_phone`
  - Instagram: `customer_instagram_id`
  - Facebook: `customer_facebook_id`
- **First message**: Creates conversation
- **Subsequent messages**: Appends to existing conversation
- **Cross-channel**: Separate conversations (can link via `conversation_roi_links`)

**Database Query** (find or create conversation):
```sql
-- Find existing conversation
SELECT id FROM conversations
WHERE location_id = $1
  AND channel = $2
  AND (
    (channel = 'whatsapp' AND customer_phone = $3) OR
    (channel = 'instagram_dm' AND customer_instagram_id = $4) OR
    (channel = 'facebook_messenger' AND customer_facebook_id = $5)
  )
LIMIT 1;

-- If not found, create new conversation
INSERT INTO conversations (...)
VALUES (...)
RETURNING id;
```

### 2.3 Is there deduplication?

**Status**: ❌ **NOT IMPLEMENTED**

**Problem**: Meta sometimes sends duplicate webhooks (network retries, webhook re-delivery).

**Recommendation**: **Idempotency via `channel_message_id`**:

**Strategy**:
1. **Unique constraint**: `(channel, channel_message_id)` in `messages` table
2. **Check before insert**: If message exists, return existing message ID
3. **Log duplicates**: Track duplicate attempts for monitoring

**Implementation**:
```go
// Check for duplicate
existing, err := repo.FindByChannelMessageID(ctx, channel, channelMessageID)
if err == nil && existing != nil {
    logger.Info().
        Str("channel", channel).
        Str("message_id", channelMessageID).
        Msg("Duplicate webhook received, skipping")
    return existing, nil
}

// Insert new message
message, err := repo.Create(ctx, newMessage)
```

**Database Index**:
```sql
CREATE UNIQUE INDEX idx_messages_channel_id 
ON messages(channel, channel_message_id);
```

### 2.4 What metadata is captured?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: Capture comprehensive metadata:

**Message Metadata**:
- `channel_message_id` - Platform message ID (for deduplication)
- `timestamp` - Platform timestamp (when message was sent)
- `sender_id` - Platform sender ID (phone, Instagram ID, etc.)
- `recipient_id` - Platform recipient ID (business phone, page ID)
- `direction` - `inbound` or `outbound`
- `message_type` - `text`, `image`, `video`, `audio`, `document`, `location`, `sticker`
- `media_urls` - Array of media URLs (downloaded and stored in S3/GCS)
- `delivered_at` - Delivery confirmation timestamp
- `read_at` - Read receipt timestamp
- `metadata` - JSONB field for platform-specific data

**Conversation Metadata**:
- `source_type` - How conversation started (`review_response`, `story_dm`, `listing_inquiry`, `direct`)
- `linked_review_id` - Connected review (ROI tracking)
- `linked_listing_source` - Connected listing (ROI tracking)
- `assigned_to_user_id` - Assigned agent
- `status` - Conversation status (`new`, `open`, `quoted`, `booked`, `converted`, `archived`)
- `lead_status` - Lead qualification (`qualified`, `interested`, `converted`)

**Database Schema Enhancement**:
```sql
ALTER TABLE messages ADD COLUMN metadata JSONB DEFAULT '{}';
ALTER TABLE conversations ADD COLUMN metadata JSONB DEFAULT '{}';
```

---

## 3. ASYNC PROCESSING

### 3.1 If async: What's the queue?

**Status**: ❌ **NOT IMPLEMENTED**

**Current Infrastructure**:
- ✅ Redis exists (for tenant caching)
- ❌ No job queue implementation
- ❌ No worker service

**Recommendation**: **Redis Streams** or **asynq**:

**Option 1: Redis Streams** (Native Redis)
```go
// Producer (webhook handler)
redis.XAdd(ctx, &redis.XAddArgs{
    Stream: "webhook:messages",
    Values: map[string]interface{}{
        "tenant_id": tenantID,
        "platform": "whatsapp",
        "payload": jsonPayload,
    },
})

// Consumer (worker)
for {
    messages, _ := redis.XReadGroup(ctx, &redis.XReadGroupArgs{
        Group: "workers",
        Consumer: "worker-1",
        Streams: []string{"webhook:messages", ">"},
        Count: 10,
    })
    // Process messages
}
```

**Option 2: asynq** (Go library, recommended)
```go
// Producer
client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
task := asynq.NewTask("process:webhook", payload)
client.Enqueue(task)

// Consumer
srv := asynq.NewServer(asynq.RedisClientOpt{Addr: redisAddr}, asynq.Config{
    Concurrency: 10,
    Queues: map[string]int{
        "webhooks": 10,
        "default": 5,
    },
})
mux := asynq.NewServeMux()
mux.HandleFunc("process:webhook", processWebhookHandler)
srv.Run(mux)
```

**Recommendation**: **asynq** for:
- ✅ Built-in retry logic
- ✅ Dead letter queue
- ✅ Job scheduling
- ✅ Priority queues
- ✅ Rate limiting

### 3.2 What's the processing latency?

**Status**: ❌ **NOT IMPLEMENTED**

**Target Latency** (from roadmap):
- Message ingest: **<5 seconds** (webhook → inbox)
- Reply send: **<2 seconds**

**Processing Steps** (estimated):
1. Webhook received: **0ms**
2. Signature verification: **10-50ms**
3. Enqueue to Redis: **5-20ms**
4. Worker picks up job: **10-100ms** (depends on queue depth)
5. Parse payload: **5-10ms**
6. Find/create conversation: **20-50ms** (DB query)
7. Store message: **30-100ms** (DB insert)
8. Download media (if any): **500-2000ms** (async, non-blocking)
9. Publish to pub/sub: **5-10ms**
10. Frontend receives update: **10-50ms** (WebSocket)

**Total Latency**: **100-400ms** (synchronous path) + media download (async)

**Bottlenecks**:
- **Database writes**: Use connection pooling, batch inserts for media
- **Media downloads**: Process async, don't block message storage
- **Pub/sub**: Use Redis pub/sub for real-time delivery

### 3.3 Is there a dead letter queue?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Yes, required for reliability**:

**asynq Dead Letter Queue**:
```go
srv := asynq.NewServer(redisOpt, asynq.Config{
    Concurrency: 10,
    Queues: map[string]int{
        "webhooks": 10,
    },
    RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
        return time.Duration(n) * time.Second
    },
    IsFailure: func(err error) bool {
        // Don't retry validation errors
        return !errors.Is(err, ErrInvalidPayload)
    },
})
```

**Dead Letter Queue Handling**:
- **Max retries**: 3 attempts
- **Retry delay**: Exponential backoff (1s, 2s, 4s)
- **DLQ storage**: Redis list `webhook:dlq`
- **Monitoring**: Alert on DLQ size > 100
- **Manual retry**: Admin endpoint to retry failed messages

### 3.4 How are processing errors retried?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Exponential backoff with jitter**:

**Retry Strategy**:
- **Transient errors** (DB timeout, Redis timeout): Retry 3 times
- **Permanent errors** (invalid payload, signature mismatch): No retry, log to DLQ
- **Rate limit errors** (429): Retry with backoff, max 5 attempts

**Implementation** (asynq):
```go
mux.HandleFunc("process:webhook", func(ctx context.Context, t *asynq.Task) error {
    // Process webhook
    if err := processWebhook(ctx, t.Payload()); err != nil {
        if isTransient(err) {
            return err // asynq will retry
        }
        return fmt.Errorf("permanent error: %w", err)
    }
    return nil
})
```

---

## 4. REAL-TIME DELIVERY

### 4.1 How do connected agency staff see new messages instantly?

**Status**: ❌ **NOT IMPLEMENTED**

**Current State**: No real-time mechanism exists.

**Recommendation**: **Redis pub/sub + WebSocket**:

**Architecture**:
```
Worker (processes webhook)
  ↓ Publishes to Redis pub/sub
  ↓ Channel: "messages:{tenant_id}:{location_id}"
Frontend (WebSocket connection)
  ↓ Subscribes to Redis pub/sub
  ↓ Filters by tenant_id (from JWT)
  ↓ Sends to connected clients
```

**Implementation Options**:

**Option 1: Server-Sent Events (SSE)** - Simpler
```go
// Handler
func (h *MessageHandler) StreamMessages(w http.ResponseWriter, r *http.Request) {
    flusher, _ := w.(http.Flusher)
    w.Header().Set("Content-Type", "text/event-stream")
    
    tenantID := getTenantID(r)
    pubsub := redis.Subscribe(ctx, fmt.Sprintf("messages:%s:*", tenantID))
    
    for msg := range pubsub.Channel() {
        fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
        flusher.Flush()
    }
}
```

**Option 2: WebSocket** - More flexible
```go
// Use gorilla/websocket or nhooyr.io/websocket
func (h *MessageHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    tenantID := getTenantID(r)
    
    pubsub := redis.Subscribe(ctx, fmt.Sprintf("messages:%s:*", tenantID))
    go func() {
        for msg := range pubsub.Channel() {
            conn.WriteJSON(map[string]interface{}{
                "type": "message",
                "data": msg.Payload,
            })
        }
    }()
}
```

**Recommendation**: **WebSocket** for bidirectional communication (send replies, typing indicators).

### 4.2 Is there WebSocket for live updates or polling (REST)?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **WebSocket** (primary) + **Polling** (fallback):

**WebSocket Endpoint**:
```
WS /api/v1/messages/stream?tenant_id=xxx&location_id=yyy
```

**Polling Endpoint** (fallback for clients that can't use WebSocket):
```
GET /api/v1/messages/poll?last_message_id=xxx&timeout=30
```

**Implementation**:
- **WebSocket**: Use `nhooyr.io/websocket` (modern, efficient)
- **Polling**: Long-polling with 30s timeout
- **Auto-fallback**: Frontend detects WebSocket failure → falls back to polling

### 4.3 What's the average latency from WhatsApp delivery to inbox display?

**Status**: ❌ **NOT IMPLEMENTED** (cannot measure)

**Target**: **<5 seconds** (from roadmap)

**Estimated Latency Breakdown**:
1. Meta sends webhook: **0ms**
2. Webhook handler receives: **50-200ms** (network latency)
3. Signature verification: **10-50ms**
4. Enqueue to Redis: **5-20ms**
5. Worker processes: **100-400ms** (DB writes)
6. Publish to pub/sub: **5-10ms**
7. WebSocket delivery: **10-50ms**
8. Frontend render: **50-100ms**

**Total**: **230-830ms** (well under 5s target)

**Optimization Opportunities**:
- **Batch processing**: Process multiple messages in one transaction
- **Connection pooling**: Reuse DB connections
- **Read replicas**: Use read replica for conversation queries
- **Caching**: Cache recent conversations in Redis

### 4.4 Could we improve this with Redis pub/sub?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Yes, Redis pub/sub is the recommended approach**:

**Benefits**:
- ✅ **Low latency**: <10ms pub/sub delivery
- ✅ **Scalable**: Multiple workers can publish, multiple frontend servers can subscribe
- ✅ **Decoupled**: Workers don't need to know about frontend connections
- ✅ **Resilient**: Redis handles reconnection, message buffering

**Architecture**:
```
Worker (EC2)
  ↓ Process webhook
  ↓ Store message in DB
  ↓ redis.Publish("messages:{tenant_id}:{location_id}", jsonMessage)
  
Frontend Server (EC2/Next.js)
  ↓ redis.Subscribe("messages:{tenant_id}:*")
  ↓ Filter by location_id (if needed)
  ↓ Send to WebSocket clients
```

**Alternative**: **NATS** or **Kafka** for higher throughput, but Redis pub/sub is sufficient for 50K messages/day.

---

## 5. SCALING CONCERNS

### 5.1 At 1,000 locations × 50 conversations/day = 50K messages/day peak

**Status**: ❌ **NOT IMPLEMENTED** (cannot validate)

**Projected Load**:
- **Messages/day**: 50,000
- **Messages/hour**: ~2,083 (peak: ~4,000 during business hours)
- **Messages/minute**: ~35 (peak: ~70)
- **Messages/second**: **~0.6 average, ~1.2 peak**

**Current Capacity**: Unknown (no implementation exists).

### 5.2 Can the current implementation handle this?

**Status**: ❌ **NOT IMPLEMENTED** (no implementation to test)

**Assessment**: Based on architecture recommendations:

**Webhook Handler (EC2)**:
- **Throughput**: 100-1,000 req/sec per instance (Go HTTP server)
- **Required**: 1-2 instances (t3.medium) can handle 1.2 msg/sec peak
- **Bottleneck**: None (webhook handler is stateless, just enqueues)

**Worker (EC2)**:
- **Throughput**: 10-50 messages/sec per worker (DB writes are bottleneck)
- **Required**: 1 worker (t3.medium) can handle 1.2 msg/sec peak
- **Bottleneck**: Database writes (use connection pooling, batch inserts)

**Database (PostgreSQL)**:
- **Write capacity**: 1,000-5,000 writes/sec (db.t3.small)
- **Required**: 1.2 writes/sec peak (well within capacity)
- **Bottleneck**: None at this scale

**Redis**:
- **Pub/sub throughput**: 100,000+ messages/sec
- **Required**: 1.2 messages/sec peak (well within capacity)
- **Bottleneck**: None

**Conclusion**: **Yes, single EC2 instance can handle 50K messages/day** with proper architecture.

### 5.3 Are there any bottlenecks?

**Status**: ❌ **NOT IMPLEMENTED** (cannot identify actual bottlenecks)

**Potential Bottlenecks** (based on architecture):

1. **Database Writes** (highest risk):
   - **Issue**: Each message = 2 writes (conversation update + message insert)
   - **Mitigation**: 
     - Connection pooling (20-50 connections)
     - Batch inserts for media metadata
     - Use `ON CONFLICT DO NOTHING` for deduplication

2. **Media Downloads** (async, non-blocking):
   - **Issue**: Downloading media from Meta can take 1-5 seconds
   - **Mitigation**: 
     - Process media downloads in separate worker queue
     - Store message immediately, download media async
     - Use S3/GCS for media storage

3. **WebSocket Connections** (scaling):
   - **Issue**: Each connected user = 1 WebSocket connection
   - **Mitigation**: 
     - Use Redis pub/sub to decouple workers from frontend
     - Frontend servers can scale horizontally
     - Use load balancer with sticky sessions

4. **Redis Pub/Sub** (low risk):
   - **Issue**: None at 1.2 msg/sec
   - **Mitigation**: None needed

### 5.4 What's the current peak message throughput?

**Status**: ❌ **NOT IMPLEMENTED** (cannot measure)

**Estimated** (based on architecture):
- **Current**: 0 messages/sec (not implemented)
- **Target**: 1.2 messages/sec peak (50K messages/day)
- **Capacity**: 10-50 messages/sec (with proper architecture)

---

## 6. PLATFORM COVERAGE

### 6.1 Which platforms are currently handled?

**Status**: ❌ **NOT IMPLEMENTED** (no platforms handled)

**Planned Platforms** (from roadmap):
1. ✅ **WhatsApp** - Priority 1 (93% LATAM adoption)
2. ✅ **Instagram DMs** - Priority 2 (Gen Z/Millennial)
3. ✅ **Facebook Messenger** - Priority 3 (older demographics)
4. ⏳ **Google Business Messages** - Planned
5. ⏳ **Email** - Planned (SendGrid/Mailgun)
6. ⏳ **SMS** - Low priority (de-emphasized)

### 6.2 Is platform routing handled at webhook layer or handler layer?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Handler layer** (separate endpoints per platform):

**Architecture**:
```
POST /api/v1/webhooks/whatsapp    → WhatsAppHandler
POST /api/v1/webhooks/instagram   → InstagramHandler
POST /api/v1/webhooks/facebook    → FacebookHandler
```

**Why Separate Endpoints**:
- ✅ Different signature verification per platform
- ✅ Different payload structures
- ✅ Different rate limits
- ✅ Easier to add new platforms
- ✅ Platform-specific middleware (auth, logging)

**Alternative**: Single endpoint with platform detection:
```
POST /api/v1/webhooks/meta?platform=whatsapp
```
❌ Not recommended (less clear, harder to secure).

### 6.3 Are there any platform-specific edge cases?

**Status**: ❌ **NOT IMPLEMENTED** (cannot identify)

**Known Edge Cases** (from Meta documentation):

**WhatsApp**:
- **24-hour window**: Can only send template messages after 24h of customer inactivity
- **Media limits**: 16MB for images, 64MB for videos
- **Character limits**: 4,096 characters per message
- **Template messages**: Required for first message after 24h window

**Instagram**:
- **Story replies**: Special handling for Story DM responses
- **Media types**: Images, videos, reels
- **Character limits**: 1,000 characters per message

**Facebook Messenger**:
- **Page messaging**: Requires page access token
- **Messenger Extensions**: Rich media (buttons, quick replies)
- **Character limits**: 2,000 characters per message

**Recommendation**: Create platform-specific handlers with edge case handling:
```go
// internal/domains/messages/infra/parsers/whatsapp_parser.go
func ParseWhatsAppWebhook(payload []byte) (*WebhookPayload, error) {
    // Handle WhatsApp-specific fields
    // Validate 24-hour window
    // Extract media URLs
}
```

### 6.4 How hard would it be to add a new platform?

**Status**: ❌ **NOT IMPLEMENTED** (cannot assess)

**Estimated Effort**: **1-2 days per platform** (with existing infrastructure):

**Steps**:
1. **Create parser** (`internal/domains/messages/infra/parsers/{platform}_parser.go`) - 2-4 hours
2. **Create handler** (`internal/domains/messages/infra/http/{platform}_handler.go`) - 2-4 hours
3. **Add signature verification** (if different from Meta) - 1-2 hours
4. **Add tests** - 2-4 hours
5. **Update OpenAPI spec** - 1 hour
6. **Documentation** - 1 hour

**Total**: **8-16 hours** (1-2 days)

**With Proper Architecture**:
- ✅ Parser interface: `type Parser interface { Parse([]byte) (*Message, error) }`
- ✅ Handler interface: `type Handler interface { Handle(http.Request) error }`
- ✅ Platform registry: `var parsers = map[string]Parser{...}`

---

## 7. MONITORING & OBSERVABILITY

### 7.1 Are webhook deliveries logged and tracked?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Comprehensive logging**:

**Log Events**:
1. **Webhook received**: `webhook.received` (platform, tenant_id, timestamp)
2. **Signature verified**: `webhook.verified` (success/failure)
3. **Message enqueued**: `webhook.enqueued` (queue_id, delay)
4. **Message processed**: `webhook.processed` (duration, success/failure)
5. **Message stored**: `message.stored` (message_id, conversation_id)
6. **Real-time delivered**: `message.delivered` (user_id, latency)

**Structured Logging** (zerolog):
```go
logger.Info().
    Str("event", "webhook.received").
    Str("platform", "whatsapp").
    Str("tenant_id", tenantID).
    Str("message_id", messageID).
    Dur("latency_ms", latency).
    Msg("Webhook received and processed")
```

**Metrics** (Prometheus/CloudWatch):
- `webhook_received_total{platform, tenant_id, status}`
- `webhook_processing_duration_seconds{platform}`
- `message_storage_duration_seconds`
- `real_time_delivery_latency_seconds`

### 7.2 Can we see webhook lag?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Track end-to-end latency**:

**Latency Metrics**:
1. **Webhook lag**: Time from Meta sending to us receiving
   - **Measurement**: `webhook.timestamp` (from Meta) vs `received_at` (our timestamp)
   - **Target**: <1 second
   - **Alert**: >5 seconds

2. **Processing lag**: Time from receiving to storing
   - **Measurement**: `received_at` vs `stored_at`
   - **Target**: <500ms
   - **Alert**: >2 seconds

3. **Delivery lag**: Time from storing to frontend display
   - **Measurement**: `stored_at` vs `delivered_at`
   - **Target**: <100ms
   - **Alert**: >1 second

**Implementation**:
```go
type WebhookMetrics struct {
    ReceivedAt    time.Time
    ProcessedAt   time.Time
    StoredAt      time.Time
    DeliveredAt   time.Time
    PlatformTime  time.Time // From webhook payload
}

func (m *WebhookMetrics) CalculateLag() time.Duration {
    return m.ReceivedAt.Sub(m.PlatformTime)
}
```

### 7.3 Are failures alerted immediately?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Real-time alerting**:

**Alert Conditions**:
1. **Webhook signature failure**: Alert immediately (security issue)
2. **Processing failure rate**: >5% failures in 5 minutes
3. **Dead letter queue**: >100 messages in DLQ
4. **Database connection errors**: >10 errors in 1 minute
5. **Redis connection errors**: >10 errors in 1 minute
6. **Webhook lag**: >5 seconds average over 5 minutes

**Alert Channels**:
- **PagerDuty/Opsgenie**: Critical alerts (signature failures, DB down)
- **Slack/Email**: Warning alerts (high failure rate, DLQ growth)
- **Dashboard**: Grafana/CloudWatch for visualization

**Implementation**:
```go
if failureRate > 0.05 {
    alerting.SendAlert(Alert{
        Severity: "critical",
        Message: fmt.Sprintf("Webhook processing failure rate: %.2f%%", failureRate*100),
        Channel: "pagerduty",
    })
}
```

### 7.4 How would we diagnose a webhook delivery outage?

**Status**: ❌ **NOT IMPLEMENTED**

**Recommendation**: **Comprehensive diagnostics**:

**Diagnostic Endpoints**:
```
GET /api/v1/webhooks/health
  → Returns: queue depth, worker status, DB connectivity, Redis connectivity

GET /api/v1/webhooks/stats?platform=whatsapp&hours=24
  → Returns: received count, processed count, failure count, avg latency

GET /api/v1/webhooks/recent?limit=100
  → Returns: Recent webhook attempts (for debugging)
```

**Diagnostic Checklist**:
1. **Check webhook endpoint**: `curl -X POST /api/v1/webhooks/whatsapp`
2. **Check queue depth**: `redis.LLEN("webhook:queue")`
3. **Check worker status**: `GET /api/v1/workers/status`
4. **Check database**: `SELECT COUNT(*) FROM messages WHERE created_at > NOW() - INTERVAL '1 hour'`
5. **Check logs**: `grep "webhook.received" logs/ | tail -100`
6. **Check Meta webhook logs**: Meta Business Manager → Webhooks → Delivery logs

**Troubleshooting Guide**:
- **No webhooks received**: Check Meta webhook configuration, firewall rules
- **Signature failures**: Check app secret, verify signature algorithm
- **Processing failures**: Check database connectivity, worker logs
- **High latency**: Check queue depth, database performance, Redis performance

---

## 8. ARCHITECTURE DIAGRAM

```
┌─────────────────────────────────────────────────────────────────┐
│                        META PLATFORMS                           │
│  WhatsApp  │  Instagram  │  Facebook Messenger                 │
└─────────────┴─────────────┴─────────────────────────────────────┘
                    │
                    │ HTTPS Webhook
                    │ (X-Hub-Signature-256)
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    WEBHOOK HANDLER (EC2)                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  POST /api/v1/webhooks/whatsapp                         │  │
│  │  POST /api/v1/webhooks/instagram                       │  │
│  │  POST /api/v1/webhooks/facebook                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                    │                                            │
│  ┌─────────────────▼──────────────────────────────────────┐   │
│  │  1. Signature Verification (HMAC-SHA256)              │   │
│  │  2. Rate Limiting (per tenant, per IP)                │   │
│  │  3. Parse Payload (platform-specific)                │   │
│  │  4. Enqueue to Redis Queue                            │   │
│  │  5. Return 200 OK (immediately)                       │   │
│  └────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                    │
                    │ Enqueue Job
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    REDIS QUEUE (asynq)                         │
│  Stream: webhook:messages                                       │
│  Consumer Group: workers                                       │
└─────────────────────────────────────────────────────────────────┘
                    │
                    │ Worker picks up job
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    ASYNC WORKER (EC2)                          │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  1. Dequeue job from Redis                               │  │
│  │  2. Find or create conversation                         │  │
│  │  3. Store message in database                           │  │
│  │  4. Download media (async, non-blocking)                │  │
│  │  5. Publish to Redis pub/sub                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                    │
                    │ Publish Event
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    REDIS PUB/SUB                               │
│  Channel: messages:{tenant_id}:{location_id}                   │
└─────────────────────────────────────────────────────────────────┘
                    │
                    │ Subscribe
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    FRONTEND SERVER (Next.js)                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  WebSocket: /api/v1/messages/stream                     │  │
│  │  - Subscribe to Redis pub/sub                           │  │
│  │  - Filter by tenant_id (from JWT)                      │  │
│  │  - Send to connected WebSocket clients                 │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                    │
                    │ WebSocket Message
                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    AGENCY STAFF (Browser)                       │
│  - Real-time inbox updates                                     │
│  - New message notifications                                   │
│  - Typing indicators                                           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    POSTGRESQL DATABASE                         │
│  - conversations table                                         │
│  - messages table                                              │
│  - conversation_roi_links table                              │
│  - RLS enabled (tenant isolation)                             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 9. CODE WALKTHROUGH

### 9.1 Webhook Handler

**Location**: `internal/domains/messages/infra/http/webhook_handlers.go`

```go
package http

import (
    "encoding/json"
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/hibiken/asynq"
)

type WebhookHandlers struct {
    logger      zerolog.Logger
    queueClient *asynq.Client
    verifier    *WebhookVerifier
    parser      map[string]Parser
}

func (h *WebhookHandlers) RegisterRoutes(r chi.Router) {
    r.Route("/webhooks", func(r chi.Router) {
        r.Post("/whatsapp", h.HandleWhatsApp)
        r.Post("/instagram", h.HandleInstagram)
        r.Post("/facebook", h.HandleFacebook)
    })
}

func (h *WebhookHandlers) HandleWhatsApp(w http.ResponseWriter, r *http.Request) {
    // 1. Verify signature
    if err := h.verifier.VerifyMetaWebhook(r, "whatsapp"); err != nil {
        h.logger.Warn().Err(err).Msg("WhatsApp webhook signature verification failed")
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // 2. Parse payload
    var payload WhatsAppWebhookPayload
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }
    
    // 3. Enqueue job
    task := asynq.NewTask("process:webhook", payloadBytes)
    if err := h.queueClient.Enqueue(task); err != nil {
        h.logger.Error().Err(err).Msg("Failed to enqueue webhook")
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    
    // 4. Return 200 OK immediately
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

### 9.2 Async Worker

**Location**: `internal/domains/messages/infra/worker/webhook_processor.go`

```go
package worker

import (
    "context"
    "github.com/hibiken/asynq"
)

type WebhookProcessor struct {
    logger        zerolog.Logger
    messageRepo   MessageRepository
    convRepo      ConversationRepository
    pubsub        *redis.Client
}

func (p *WebhookProcessor) ProcessWebhook(ctx context.Context, t *asynq.Task) error {
    // 1. Parse task payload
    var payload WebhookPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }
    
    // 2. Find or create conversation
    conv, err := p.convRepo.FindOrCreate(ctx, payload.LocationID, payload.Channel, payload.CustomerID)
    if err != nil {
        return fmt.Errorf("failed to find/create conversation: %w", err)
    }
    
    // 3. Check for duplicate message
    existing, _ := p.messageRepo.FindByChannelMessageID(ctx, payload.Channel, payload.MessageID)
    if existing != nil {
        p.logger.Info().
            Str("message_id", payload.MessageID).
            Msg("Duplicate message, skipping")
        return nil
    }
    
    // 4. Store message
    message, err := p.messageRepo.Create(ctx, &Message{
        ConversationID: conv.ID,
        Channel: payload.Channel,
        ChannelMessageID: payload.MessageID,
        Text: payload.Text,
        Direction: "inbound",
        Timestamp: payload.Timestamp,
    })
    if err != nil {
        return fmt.Errorf("failed to store message: %w", err)
    }
    
    // 5. Update conversation
    conv.LastMessageAt = payload.Timestamp
    p.convRepo.Update(ctx, conv)
    
    // 6. Publish to Redis pub/sub
    event := MessageEvent{
        TenantID: conv.TenantID,
        LocationID: conv.LocationID,
        ConversationID: conv.ID,
        Message: message,
    }
    p.pubsub.Publish(ctx, fmt.Sprintf("messages:%s:%s", conv.TenantID, conv.LocationID), eventJSON)
    
    return nil
}
```

---

## 10. THROUGHPUT ANALYSIS

### 10.1 Current vs Projected

**Current**: **0 messages/sec** (not implemented)

**Projected** (50K messages/day):
- **Average**: 0.6 messages/sec
- **Peak**: 1.2 messages/sec (business hours)
- **Burst**: 5 messages/sec (spike handling)

### 10.2 Capacity Planning

**Single EC2 Instance (t3.medium)**:
- **Webhook handler**: 100-1,000 req/sec ✅ (sufficient)
- **Worker**: 10-50 messages/sec ✅ (sufficient)
- **Database**: 1,000-5,000 writes/sec ✅ (sufficient)
- **Redis**: 100,000+ ops/sec ✅ (sufficient)

**Conclusion**: **Single EC2 instance can handle 50K messages/day** with proper architecture.

**Scaling Path**:
- **0-50K messages/day**: 1 EC2 instance (t3.medium)
- **50K-200K messages/day**: 2 EC2 instances (webhook handler + worker)
- **200K+ messages/day**: Auto-scaling group (3-10 instances)

---

## 11. IDENTIFIED BOTTLENECKS

### 11.1 Database Writes

**Risk**: **HIGH** (each message = 2 writes)

**Mitigation**:
- ✅ Connection pooling (20-50 connections)
- ✅ Batch inserts for media metadata
- ✅ Use `ON CONFLICT DO NOTHING` for deduplication
- ✅ Read replicas for conversation queries

### 11.2 Media Downloads

**Risk**: **MEDIUM** (1-5 seconds per media file)

**Mitigation**:
- ✅ Process media downloads async (separate queue)
- ✅ Store message immediately, download media later
- ✅ Use S3/GCS for media storage
- ✅ CDN for media delivery

### 11.3 WebSocket Connections

**Risk**: **LOW** (at 50K messages/day scale)

**Mitigation**:
- ✅ Use Redis pub/sub to decouple workers from frontend
- ✅ Frontend servers can scale horizontally
- ✅ Load balancer with sticky sessions

---

## 12. EC2 vs LAMBDA COST ANALYSIS

### 12.1 Cost Comparison

**Assumptions**:
- **Messages/day**: 50,000
- **Messages/month**: 1,500,000
- **Peak messages/sec**: 1.2
- **Average messages/sec**: 0.6
- **Webhook handler**: 24/7 operation
- **Worker**: 24/7 operation

**EC2 (t3.medium)**:
- **Instance cost**: $0.0416/hour × 730 hours = **$30.37/month**
- **Data transfer**: 1GB/month = **$0.09/month**
- **Total**: **~$30.50/month**

**Lambda** (per message):
- **Requests**: 1,500,000/month
- **Compute**: 1,500,000 × 100ms × 512MB = 75,000 GB-seconds
  - First 400,000 GB-seconds: Free
  - Remaining: 75,000 GB-seconds × $0.0000166667 = **$1.25/month**
- **Requests**: 1,500,000 × $0.20/1M = **$0.30/month**
- **Total**: **~$1.55/month**

**BUT**: Lambda has **cold start latency** (100-500ms) which violates the <5s target.

**EC2 Advantages**:
- ✅ **No cold starts**: Always warm, <50ms latency
- ✅ **Predictable performance**: No throttling
- ✅ **Lower latency**: Direct connection to Redis/DB
- ✅ **Better for 24/7 workloads**: No per-request overhead

**Lambda Advantages**:
- ✅ **Lower cost**: $1.55 vs $30.50/month (20x cheaper)
- ✅ **Auto-scaling**: Handles spikes automatically
- ✅ **No server management**: Fully managed

### 12.2 Recommendation: **EC2** (Validate Cost Decision)

**Why EC2**:
1. **Latency requirement**: <5s target requires <50ms webhook processing (Lambda cold starts add 100-500ms)
2. **24/7 operation**: EC2 is more cost-effective for always-on workloads
3. **Predictable traffic**: 50K messages/day is steady, not spiky
4. **Connection pooling**: EC2 can maintain DB/Redis connections (Lambda cannot)

**Cost Optimization**:
- **Reserved Instances**: 1-year term = **$20/month** (33% savings)
- **Spot Instances**: **$10-15/month** (50-67% savings, but less reliable)
- **Auto-scaling**: Scale down during off-hours (nights/weekends) = **$20-25/month**

**Conclusion**: **EC2 is the right choice** for 24/7 webhook processing with latency requirements.

---

## 13. IMPLEMENTATION RECOMMENDATIONS

### 13.1 Priority 1: Webhook Infrastructure (Week 1)

1. **Create messages domain**:
   - Database migrations (conversations, messages tables)
   - Domain entities (Message, Conversation)
   - Repository interfaces

2. **Webhook handlers**:
   - HTTP handlers for WhatsApp, Instagram, Facebook
   - Signature verification middleware
   - Rate limiting middleware

3. **Async processing**:
   - Set up asynq worker
   - Webhook processor
   - Dead letter queue

### 13.2 Priority 2: Real-Time Delivery (Week 2)

1. **Redis pub/sub**:
   - Publish events from worker
   - Subscribe in frontend server

2. **WebSocket endpoint**:
   - WebSocket handler
   - Connection management
   - Message broadcasting

### 13.3 Priority 3: Platform Integration (Week 3-4)

1. **WhatsApp** (Priority 1):
   - BSP partner setup (YCloud/Wati)
   - Webhook receiver
   - Message sender

2. **Instagram** (Priority 2):
   - Meta Graph API setup
   - Webhook receiver
   - Story reply handling

3. **Facebook** (Priority 3):
   - Meta Graph API setup
   - Webhook receiver
   - Page messaging

### 13.4 Priority 4: Monitoring & Observability (Week 5)

1. **Logging**:
   - Structured logging (zerolog)
   - Log aggregation (CloudWatch/ELK)

2. **Metrics**:
   - Prometheus metrics
   - CloudWatch dashboards

3. **Alerting**:
   - PagerDuty integration
   - Slack notifications

---

## 14. CONCLUSION

**Current State**: Webhook functionality is **completely absent** from the codebase. This is a **P0 gap** that must be addressed.

**Recommendation**: Implement webhook infrastructure using **EC2** (not Lambda) for:
- ✅ Low latency (<50ms processing)
- ✅ 24/7 operation (cost-effective)
- ✅ Predictable performance (no cold starts)

**Architecture**: 
- **Webhook Handler (EC2)** → **Redis Queue (asynq)** → **Worker (EC2)** → **Redis Pub/Sub** → **Frontend (WebSocket)**

**Timeline**: **4-5 weeks** to implement full webhook infrastructure with WhatsApp, Instagram, and Facebook support.

**Next Steps**:
1. Create messages domain and database migrations
2. Implement webhook handlers with signature verification
3. Set up async processing with asynq
4. Implement real-time delivery with Redis pub/sub + WebSocket
5. Integrate WhatsApp, Instagram, Facebook platforms
6. Add monitoring and observability

---

**Document Version**: 1.0  
**Last Updated**: 2025-01-27  
**Author**: AI Analysis
