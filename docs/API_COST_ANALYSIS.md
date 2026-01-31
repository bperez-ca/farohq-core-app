# Paid API Integration Cost Analysis & Optimization Strategy

**Date**: 2025-01-27  
**Status**: ⚠️ **CRITICAL FINDING - APIs NOT YET IMPLEMENTED**  
**Current Spend**: $0 (no integrations active)  
**Target**: Prevent 20%+ cost leaks when integrations go live

---

## Executive Summary

**Finding**: Data4SEO, Meta Cloud API, and OpenAI integrations are **not currently implemented** in the codebase. However, they are planned in the roadmap and will significantly impact unit economics once live.

**Risk**: Without proper cost controls implemented **before** these integrations go live, FARO could experience:
- **Uncontrolled API spend** (no budget caps per agency)
- **Duplicate API calls** (no deduplication)
- **Cache misses** (no response caching)
- **Retry storms** (no exponential backoff)
- **No cost visibility** (no tracking or alerts)

**Recommendation**: Implement cost optimization architecture **before** integrating these APIs to prevent cost leaks.

---

## 1. DATA4SEO INTEGRATION ANALYSIS

### Current Status
❌ **NOT IMPLEMENTED**

**Expected Endpoints** (based on roadmap):
- GBP profile snapshots: `$0.50-1.50 per snapshot`
- Local rankings data: `$0.50-1.50 per snapshot`
- Competitive intelligence: `$0.50-1.50 per snapshot`

**Expected Usage** (when implemented):
- GBP profile sync: Every location sync
- Rankings: Daily per location
- Snapshots: On-demand (paid by agency wallet)

### Identified Risks & Cost Leaks

#### ❌ Risk 1: No Caching Strategy
**Problem**: If implemented without caching, same GBP profile could be fetched multiple times per day.

**Example Cost Leak**:
- Agency has 10 locations
- User opens GBP sync dashboard 5 times/day
- Each dashboard load fetches fresh profile (no cache)
- Cost: `10 locations × 5 loads × $1.00 = $50/day = $1,500/month`

**With Proper Caching**:
- Cache GBP profiles for 15 minutes (profiles change infrequently)
- Cost: `10 locations × 4 syncs/day × $1.00 = $40/day = $1,200/month`
- **Savings: $300/month (20% reduction)**

#### ❌ Risk 2: No Rate Limiting Per Agency
**Problem**: Single agency could trigger unlimited API calls.

**Example Cost Leak**:
- Malicious user or bug triggers refresh loop
- 100 API calls/minute for 1 hour
- Cost: `100 calls/min × 60 min × $1.00 = $6,000/hour`

**Mitigation Needed**:
- Per-agency daily budget cap: `$50/day default`
- Rate limiter: `10 calls/minute per agency`
- Alert on 80% budget consumed
- Auto-pause expensive operations if budget exceeded

#### ❌ Risk 3: No Request Deduplication
**Problem**: Multiple users refreshing same location simultaneously triggers duplicate calls.

**Example Cost Leak**:
- 5 agency members refresh location dashboard simultaneously
- Each triggers GBP profile fetch
- Cost: `5 duplicate calls × $1.00 = $5 wasted`

**Mitigation Needed**:
- In-flight request deduplication (cache pending requests)
- Lock per `location_id + endpoint` for 30 seconds
- Subsequent requests wait for first to complete

#### ❌ Risk 4: Using Priority Queue Instead of Standard
**Problem**: Priority queue (1min delay) is more expensive than standard queue (45min delay).

**Current Queue System**: ❌ **NOT IMPLEMENTED** (asynq mentioned in roadmap)

**Cost Impact**:
- Standard queue: Acceptable for GBP sync (45min delay)
- Priority queue: Only needed for real-time user requests
- **Recommendation**: Use standard queue for background syncs, priority only for user-triggered syncs

### Recommended Implementation (Before Going Live)

#### 1. API Client with Caching Layer
```go
// internal/domains/data4seo/infra/http/client.go
type Data4SEOClient struct {
    httpClient  *http.Client
    cache       *redis.Client
    rateLimiter *RateLimiter  // per-agency rate limiter
}

func (c *Data4SEOClient) GetGBPProfile(ctx context.Context, locationID uuid.UUID, tenantID uuid.UUID) (*GBPProfile, error) {
    // 1. Check cache first (TTL: 15 minutes)
    cacheKey := fmt.Sprintf("data4seo:gbp_profile:%s", locationID)
    if cached, err := c.cache.Get(ctx, cacheKey).Result(); err == nil {
        return deserializeProfile(cached), nil
    }
    
    // 2. Check rate limit
    if err := c.rateLimiter.Allow(ctx, tenantID, "gbp_profile", 10, time.Minute); err != nil {
        return nil, ErrRateLimitExceeded
    }
    
    // 3. Check budget cap
    if err := c.checkBudgetCap(ctx, tenantID, 1.00); err != nil {
        return nil, ErrBudgetExceeded
    }
    
    // 4. Check in-flight requests (deduplication)
    lockKey := fmt.Sprintf("data4seo:lock:%s", locationID)
    acquired, err := c.cache.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
    if !acquired {
        // Wait for in-flight request to complete
        return c.waitForInFlight(ctx, locationID)
    }
    defer c.cache.Del(ctx, lockKey)
    
    // 5. Make API call
    profile, err := c.fetchFromAPI(ctx, locationID)
    
    // 6. Track cost
    c.trackCost(ctx, tenantID, "data4seo", "gbp_profile", 1.00)
    
    // 7. Cache response
    c.cache.Set(ctx, cacheKey, serializeProfile(profile), 15*time.Minute)
    
    return profile, err
}
```

#### 2. Cache TTL Recommendations
| Endpoint | Cache TTL | Rationale |
|----------|-----------|-----------|
| GBP Profiles | 15 minutes | Profiles change infrequently |
| Rankings | 1 hour | Rankings update daily |
| Snapshots | 24 hours | Historical data, expensive |

#### 3. Rate Limiting Configuration
```go
// Per-agency rate limits
type RateLimits struct {
    GBPProfile    int   // 10 calls/minute
    Rankings      int   // 5 calls/minute
    Snapshots     int   // 2 calls/minute
    DailyBudget   float64 // $50/day default
}
```

#### 4. Queue Strategy
- **Standard Queue** (45min delay): Background GBP syncs, scheduled rankings
- **Priority Queue** (1min delay): User-triggered syncs only
- **Cost Savings**: 90% of syncs can use standard queue

### Estimated Cost Per Operation

| Operation | Cost | Frequency (per agency) | Monthly Cost (10 locations) |
|-----------|------|------------------------|----------------------------|
| GBP Profile Sync | $1.00 | 4×/day | $1,200 |
| Rankings (daily) | $1.00 | 1×/day | $300 |
| Snapshots (on-demand) | $1.00 | 10×/month | $10 |

**Without Optimization**: ~$1,510/month per agency (10 locations)  
**With Optimization**: ~$1,200/month per agency (20% reduction)

---

## 2. META CLOUD API INTEGRATION (WhatsApp/IG/FB)

### Current Status
❌ **NOT IMPLEMENTED**

**Expected Endpoints**:
- WhatsApp Business API: `$0.0015-0.045 per conversation` (varies by country)
- Instagram Messaging API: `$0.0015-0.045 per conversation`
- Facebook Messenger API: `$0.0015-0.045 per conversation`

**Expected Usage** (when implemented):
- Webhook receivers for incoming messages
- Outbound message sending
- 24-hour free messaging window optimization

### Identified Risks & Cost Leaks

#### ❌ Risk 1: Webhook Deduplication Missing
**Problem**: Meta may send duplicate webhooks, or network issues cause retries.

**Example Cost Leak**:
- Incoming WhatsApp message triggers webhook
- Webhook processed → sends reply → **$0.02 charged**
- Duplicate webhook arrives 2 seconds later
- Reply sent again → **$0.02 wasted**

**Monthly Impact**: 
- 1,000 conversations/day
- 2% duplicate rate
- Cost: `1,000 × 0.02 × 30 days = $12/month wasted`

#### ❌ Risk 2: Storing Raw Webhook Responses in DB
**Problem**: Full webhook payloads stored in database increases storage costs.

**Example Cost Leak**:
- Each webhook: ~2KB JSON
- 10,000 messages/month
- Storage: `10,000 × 2KB = 20MB/month`
- At $0.023/GB: `$0.46/month` (small, but compounds)

**Mitigation**: Store only essential fields, archive raw payloads to S3 (cheaper).

#### ❌ Risk 3: Missing 24-Hour Window Optimization
**Problem**: Meta allows free replies within 24 hours of user message. Missing window = paid message.

**Example Cost Leak**:
- User sends message at 10:00 AM
- Agent replies at 10:30 AM next day (25 hours later)
- Cost: `$0.02` instead of free

**Monthly Impact**:
- 5% of replies miss window
- 1,000 conversations/day
- Cost: `1,000 × 0.05 × 30 × $0.02 = $30/month wasted`

#### ❌ Risk 4: No Message Queue Prioritization
**Problem**: All messages sent immediately, even non-urgent ones.

**Cost Impact**:
- Batch non-urgent replies within 24-hour window
- Use template messages (cheaper) when possible
- **Savings**: 10-15% reduction in message costs

#### ❌ Risk 5: No Cost Per Agency Tracking
**Problem**: Cannot identify power users or set per-agency budgets.

**Impact**:
- No visibility into which agencies drive costs
- Cannot set budget caps per agency
- Cannot invoice agencies accurately

### Recommended Implementation (Before Going Live)

#### 1. Webhook Handler with Deduplication
```go
// internal/domains/conversations/infra/http/webhook_handler.go
func (h *WebhookHandler) HandleWhatsAppWebhook(ctx context.Context, payload []byte) error {
    // 1. Extract message ID (deduplication key)
    var webhook WhatsAppWebhook
    if err := json.Unmarshal(payload, &webhook); err != nil {
        return err
    }
    
    messageID := webhook.Entry[0].Changes[0].Value.Messages[0].ID
    
    // 2. Check if already processed (deduplication)
    dedupKey := fmt.Sprintf("meta:processed:%s", messageID)
    alreadyProcessed, err := h.cache.SetNX(ctx, dedupKey, "1", 24*time.Hour).Result()
    if !alreadyProcessed {
        h.logger.Info().Str("message_id", messageID).Msg("Duplicate webhook ignored")
        return nil // Already processed, ignore
    }
    
    // 3. Process message
    conversation, err := h.processMessage(ctx, webhook)
    
    // 4. Track cost (incoming is usually free, but track anyway)
    h.trackCost(ctx, conversation.TenantID, "meta", "whatsapp_incoming", 0.00)
    
    return nil
}
```

#### 2. 24-Hour Window Tracker
```go
// Track last user message timestamp per conversation
type Conversation struct {
    ID                  uuid.UUID
    TenantID            uuid.UUID
    Channel             string
    LastUserMessageAt   time.Time  // Critical for free window
    LastAgentMessageAt  time.Time
}

func (s *MessageService) SendReply(ctx context.Context, convID uuid.UUID, text string) error {
    conv, err := s.repo.Get(ctx, convID)
    if err != nil {
        return err
    }
    
    // Check if within 24-hour free window
    withinFreeWindow := time.Since(conv.LastUserMessageAt) < 24*time.Hour
    
    cost := 0.00
    if !withinFreeWindow {
        cost = 0.02 // Paid message outside window
        s.logger.Warn().
            Str("conversation_id", convID.String()).
            Dur("hours_since_user_message", time.Since(conv.LastUserMessageAt)).
            Msg("Sending paid message outside 24-hour window")
    }
    
    // Send message via Meta API
    err = s.metaClient.SendMessage(ctx, conv.ChannelID, text)
    
    // Track cost
    s.trackCost(ctx, conv.TenantID, "meta", conv.Channel+"_outbound", cost)
    
    return err
}
```

#### 3. Message Cost Tracking
```go
// Track cost per message type
type MessageCost struct {
    TenantID    uuid.UUID
    Channel     string  // "whatsapp", "instagram", "facebook"
    Type        string  // "incoming", "outbound", "template"
    Cost        float64
    Timestamp   time.Time
}

// Store in database for invoicing
// Aggregate per agency for dashboard
```

#### 4. Budget Caps Per Agency
```go
// Check budget before sending expensive messages
func (s *MessageService) checkBudget(ctx context.Context, tenantID uuid.UUID, cost float64) error {
    dailySpend, err := s.costTracker.GetDailySpend(ctx, tenantID, "meta")
    if err != nil {
        return err
    }
    
    budget := s.getAgencyBudget(ctx, tenantID) // Default: $100/day
    
    if dailySpend + cost > budget {
        return ErrBudgetExceeded
    }
    
    return nil
}
```

### Estimated Cost Per Operation

| Operation | Cost | Frequency (per agency) | Monthly Cost (1,000 conversations) |
|-----------|------|------------------------|-----------------------------------|
| WhatsApp Incoming | Free | 500/day | $0 |
| WhatsApp Outbound (24h window) | Free | 400/day | $0 |
| WhatsApp Outbound (paid) | $0.02 | 100/day | $60 |
| Instagram DMs | $0.015 | 200/day | $90 |
| Facebook Messenger | $0.01 | 100/day | $30 |

**Without Optimization**: ~$180/month per agency  
**With Optimization**: ~$150/month per agency (17% reduction via window optimization)

---

## 3. OPENAI INTEGRATION ANALYSIS

### Current Status
❌ **NOT IMPLEMENTED** (mock implementation in `faro-mocks/lib/reviews/aiMock.ts`)

**Expected Endpoints**:
- Review response drafts: `~$0.01 per draft` (token-based)
- Sentiment analysis: `~$0.001 per review`
- Voice transcription: `~$0.006 per minute` (Whisper API)

**Expected Usage** (when implemented):
- Generate 3 reply suggestions per review
- Sentiment tagging on review ingestion
- WhatsApp voice note transcription (LATAM)

### Identified Risks & Cost Leaks

#### ❌ Risk 1: No Draft Caching
**Problem**: Same review text could generate drafts multiple times.

**Example Cost Leak**:
- User clicks "Generate Draft" → `$0.01`
- User edits review, clicks again → `$0.01` (same text)
- 3 team members each generate draft → `$0.03` (same text)

**Monthly Impact**:
- 100 reviews/month per agency
- 2 duplicate generations per review
- Cost: `100 × 2 × $0.01 = $2/month wasted`

#### ❌ Risk 2: Using GPT-4 Instead of GPT-3.5
**Problem**: GPT-4 is 15× more expensive than GPT-3.5 for simple tasks.

**Cost Comparison**:
- GPT-3.5-turbo: `$0.0015/1K input tokens, $0.002/1K output tokens`
- GPT-4: `$0.03/1K input tokens, $0.06/1K output tokens`

**Example Cost Leak**:
- Review reply draft: ~500 tokens
- GPT-4 cost: `(500 × $0.03/1000) + (200 × $0.06/1000) = $0.027`
- GPT-3.5 cost: `(500 × $0.0015/1000) + (200 × $0.002/1000) = $0.00115`
- **15× more expensive!**

**Monthly Impact** (using GPT-4):
- 100 reviews/month per agency
- 3 drafts per review
- Cost: `100 × 3 × $0.027 = $8.10/month`

**With GPT-3.5**:
- Cost: `100 × 3 × $0.00115 = $0.35/month`
- **Savings: $7.75/month (96% reduction)**

#### ❌ Risk 3: No Token Limit Per Draft
**Problem**: Runaway costs if model generates very long responses.

**Mitigation Needed**:
- Set `max_tokens: 200` for reply drafts (sufficient for review replies)
- Default limit prevents cost spikes

#### ❌ Risk 4: Generating Drafts on Every Review Ingestion
**Problem**: Generating drafts for all reviews, even ones users never reply to.

**Example Cost Leak**:
- 1,000 reviews ingested/month
- Generate 3 drafts per review automatically
- Cost: `1,000 × 3 × $0.01 = $30/month`
- But only 200 reviews get replies
- **Waste: $24/month (80% of drafts unused)**

**Better Strategy**: Generate drafts only when user clicks "Generate Draft" button.

### Recommended Implementation (Before Going Live)

#### 1. OpenAI Client with Caching
```go
// internal/domains/reviews/infra/openai/client.go
type OpenAIClient struct {
    client      *openai.Client
    cache       *redis.Client
    model       string  // "gpt-3.5-turbo" for drafts, "gpt-4" only if needed
}

func (c *OpenAIClient) GenerateReviewDraft(ctx context.Context, reviewText string, rating int) ([]string, error) {
    // 1. Create cache key from review text hash
    cacheKey := fmt.Sprintf("openai:draft:%s:%d", hashText(reviewText), rating)
    
    // 2. Check cache (TTL: 7 days - review replies rarely change)
    if cached, err := c.cache.Get(ctx, cacheKey).Result(); err == nil {
        return deserializeDrafts(cached), nil
    }
    
    // 3. Check budget
    tenantID := getTenantID(ctx)
    if err := c.checkBudget(ctx, tenantID, 0.01); err != nil {
        return nil, ErrBudgetExceeded
    }
    
    // 4. Generate drafts (use GPT-3.5 for cost efficiency)
    drafts, err := c.generate(ctx, reviewText, rating, "gpt-3.5-turbo", 200) // max 200 tokens
    
    // 5. Track cost
    tokensUsed := countTokens(reviewText) + countTokens(joinDrafts(drafts))
    cost := c.calculateCost(tokensUsed, "gpt-3.5-turbo")
    c.trackCost(ctx, tenantID, "openai", "review_draft", cost)
    
    // 6. Cache drafts
    c.cache.Set(ctx, cacheKey, serializeDrafts(drafts), 7*24*time.Hour)
    
    return drafts, err
}
```

#### 2. Draft Generation Strategy
- **On-Demand Only**: Generate drafts when user clicks "Generate Draft" (not on ingestion)
- **Cache Results**: Same review text = cached draft (7-day TTL)
- **Model Selection**: Use GPT-3.5 for drafts, GPT-4 only for complex tasks
- **Token Limits**: `max_tokens: 200` for reply drafts

#### 3. Cost Per Operation

| Operation | Model | Cost | Frequency | Monthly Cost (100 reviews) |
|-----------|-------|------|-----------|----------------------------|
| Review Draft (GPT-3.5) | gpt-3.5-turbo | $0.00115 | 200 drafts | $0.23 |
| Review Draft (GPT-4) | gpt-4 | $0.027 | 200 drafts | $5.40 |
| Sentiment Analysis | gpt-3.5-turbo | $0.001 | 100 reviews | $0.10 |
| Voice Transcription | whisper-1 | $0.006/min | 50 minutes | $0.30 |

**Without Optimization (GPT-4)**: ~$5.80/month per agency  
**With Optimization (GPT-3.5 + caching)**: ~$0.63/month per agency (**89% reduction**)

---

## 4. COST VISIBILITY & TRACKING

### Current Status
❌ **NOT IMPLEMENTED** (A8 feature in roadmap: P0 Critical)

### Required Features

#### 1. Cost Tracking Table
```sql
CREATE TABLE api_costs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    service VARCHAR(50) NOT NULL,  -- "data4seo", "meta", "openai"
    operation VARCHAR(100) NOT NULL,  -- "gbp_profile", "whatsapp_outbound", "review_draft"
    cost DECIMAL(10, 4) NOT NULL,
    metadata JSONB,  -- {location_id, conversation_id, review_id, etc.}
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_costs_tenant_date ON api_costs(tenant_id, created_at);
CREATE INDEX idx_api_costs_service ON api_costs(service, created_at);
```

#### 2. Daily Cost Aggregation
```go
// Track daily spend per agency per service
type DailyCost struct {
    TenantID    uuid.UUID
    Service     string  // "data4seo", "meta", "openai"
    Date        time.Time
    TotalCost   float64
    Operations  int
}
```

#### 3. Cost Dashboard API
```go
// GET /api/v1/usage/costs?tenant_id=...&start_date=...&end_date=...
type CostBreakdown struct {
    Service      string
    Operation    string
    Count        int
    TotalCost    float64
    AvgCost      float64
}

type UsageResponse struct {
    TenantID      uuid.UUID
    Period        string  // "2025-01"
    TotalCost     float64
    Breakdown     []CostBreakdown
    Budget        float64
    BudgetRemaining float64
}
```

#### 4. Budget Alerts
```go
// Alert when 80% of daily budget consumed
if dailySpend > budget * 0.8 {
    sendAlert(ctx, tenantID, "Budget warning: 80% consumed")
}

// Pause expensive operations at 100%
if dailySpend >= budget {
    return ErrBudgetExceeded
}
```

### Implementation Priority
1. ✅ **Cost tracking table** (store all API costs)
2. ✅ **Daily aggregation** (sum costs per agency per day)
3. ✅ **Dashboard API** (expose costs to frontend)
4. ✅ **Budget alerts** (notify on high spend)

---

## 5. ERROR HANDLING & RETRIES

### Current Status
⚠️ **PARTIAL** (basic retry in auth middleware, no exponential backoff)

### Identified Risks

#### ❌ Risk 1: Retry Storms on Expensive Operations
**Problem**: Fixed retry interval could cause duplicate charges.

**Example Cost Leak**:
- Data4SEO API call fails (rate limit)
- System retries immediately 3 times
- All 3 retries succeed after rate limit clears
- Cost: `3 × $1.00 = $3.00` instead of `$1.00`

**Mitigation**: Use exponential backoff + idempotency keys.

#### ❌ Risk 2: Retrying Non-Idempotent Operations
**Problem**: Some operations should not be retried (e.g., sending message twice).

**Example Cost Leak**:
- WhatsApp message send fails (network error)
- System retries immediately
- Both attempts succeed
- User receives duplicate message → **wasted cost + bad UX**

**Mitigation**: Use idempotency keys for all paid operations.

### Recommended Retry Strategy

```go
// Exponential backoff with jitter
func retryWithBackoff(ctx context.Context, fn func() error, maxRetries int) error {
    baseDelay := 1 * time.Second
    
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        // Exponential backoff: 1s, 2s, 4s, 8s
        delay := baseDelay * time.Duration(math.Pow(2, float64(i)))
        
        // Add jitter (±20%)
        jitter := time.Duration(rand.Float64() * 0.4 * float64(delay))
        delay = delay - (delay / 5) + jitter
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            continue
        }
    }
    
    return ErrMaxRetriesExceeded
}

// Idempotency for expensive operations
func (c *Data4SEOClient) GetGBPProfile(ctx context.Context, locationID uuid.UUID, idempotencyKey string) (*GBPProfile, error) {
    // Check if already processed with this key
    resultKey := fmt.Sprintf("data4seo:result:%s", idempotencyKey)
    if cached, err := c.cache.Get(ctx, resultKey).Result(); err == nil {
        return deserializeProfile(cached), nil
    }
    
    // Make API call
    profile, err := c.fetchFromAPI(ctx, locationID)
    
    // Cache result with idempotency key (24 hour TTL)
    c.cache.Set(ctx, resultKey, serializeProfile(profile), 24*time.Hour)
    
    return profile, err
}
```

### Retry Rules by Service

| Service | Operation | Retry? | Max Retries | Backoff |
|---------|-----------|--------|-------------|---------|
| Data4SEO | GBP Profile | Yes | 3 | Exponential |
| Data4SEO | Snapshot | Yes | 3 | Exponential |
| Meta | Send Message | No | 0 | N/A (use idempotency) |
| Meta | Webhook | No | 0 | N/A (deduplicate) |
| OpenAI | Generate Draft | Yes | 2 | Exponential |

---

## 6. BILLING & WALLET MANAGEMENT

### Current Status
❌ **NOT IMPLEMENTED** (A8 feature: P0 Critical)

### Required Features

#### 1. Wallet Balance Table
```sql
CREATE TABLE agency_wallets (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id),
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    auto_reload BOOLEAN DEFAULT false,
    reload_threshold DECIMAL(10, 2) DEFAULT 10.00,
    reload_amount DECIMAL(10, 2) DEFAULT 100.00,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES agency_wallets(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    type VARCHAR(20) NOT NULL,  -- "deposit", "charge", "refund"
    amount DECIMAL(10, 2) NOT NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### 2. Pre-Charge Check
```go
func (w *WalletService) Charge(ctx context.Context, tenantID uuid.UUID, amount float64, description string) error {
    // 1. Get wallet balance
    wallet, err := w.repo.GetByTenantID(ctx, tenantID)
    if err != nil {
        return err
    }
    
    // 2. Check sufficient balance
    if wallet.Balance < amount {
        return ErrInsufficientBalance
    }
    
    // 3. Deduct balance (atomic)
    newBalance, err := w.repo.DeductBalance(ctx, wallet.ID, amount)
    if err != nil {
        return err
    }
    
    // 4. Record transaction
    w.repo.CreateTransaction(ctx, wallet.ID, tenantID, "charge", amount, description)
    
    // 5. Check auto-reload threshold
    if wallet.AutoReload && newBalance < wallet.ReloadThreshold {
        go w.triggerAutoReload(ctx, wallet) // Async
    }
    
    return nil
}
```

#### 3. Auto-Reload Logic
```go
func (w *WalletService) triggerAutoReload(ctx context.Context, wallet *Wallet) error {
    // 1. Charge agency credit card (via Stripe/etc)
    payment, err := w.paymentProcessor.Charge(ctx, wallet.TenantID, wallet.ReloadAmount)
    if err != nil {
        w.logger.Error().Err(err).Msg("Auto-reload failed")
        // Send alert to agency owner
        w.sendAlert(ctx, wallet.TenantID, "Auto-reload failed. Please top up manually.")
        return err
    }
    
    // 2. Credit wallet
    _, err = w.repo.AddBalance(ctx, wallet.ID, wallet.ReloadAmount)
    
    // 3. Record transaction
    w.repo.CreateTransaction(ctx, wallet.ID, wallet.TenantID, "deposit", wallet.ReloadAmount, "Auto-reload")
    
    // 4. Send receipt
    w.sendReceipt(ctx, wallet.TenantID, payment)
    
    return nil
}
```

#### 4. Pause Expensive Operations
```go
func (w *WalletService) CanPerformOperation(ctx context.Context, tenantID uuid.UUID, cost float64) error {
    wallet, err := w.repo.GetByTenantID(ctx, tenantID)
    if err != nil {
        return err
    }
    
    // Check if balance covers cost
    if wallet.Balance < cost {
        return ErrInsufficientBalance
    }
    
    // Check if balance is critically low (pause expensive ops)
    if wallet.Balance < 5.00 && cost > 1.00 {
        return ErrBalanceTooLowForOperation
    }
    
    return nil
}
```

---

## 7. RECOMMENDED COST OPTIMIZATION IMPLEMENTATION PLAN

### Phase 1: Foundation (Week 1) - **CRITICAL BEFORE GOING LIVE**

1. ✅ **Cost Tracking Infrastructure**
   - Create `api_costs` table
   - Create `agency_wallets` table
   - Implement cost tracking service
   - **Impact**: Enables cost visibility

2. ✅ **Budget & Rate Limiting**
   - Per-agency daily budget caps
   - Rate limiters per service
   - Budget alerts at 80%
   - **Impact**: Prevents runaway costs

3. ✅ **Cache Layer**
   - Redis caching for all API responses
   - TTL configuration per endpoint
   - Cache invalidation strategy
   - **Impact**: 20-30% cost reduction

### Phase 2: Deduplication & Retries (Week 2)

4. ✅ **Request Deduplication**
   - In-flight request locking
   - Webhook deduplication
   - Idempotency keys
   - **Impact**: Prevents duplicate charges

5. ✅ **Retry Strategy**
   - Exponential backoff
   - Service-specific retry rules
   - Idempotent operations only
   - **Impact**: Prevents retry storms

### Phase 3: Optimization (Week 3)

6. ✅ **Model Selection**
   - Use GPT-3.5 for review drafts (not GPT-4)
   - Use standard queue for background jobs
   - **Impact**: 89% reduction in OpenAI costs

7. ✅ **24-Hour Window Optimization**
   - Track last user message timestamp
   - Prefer free window replies
   - **Impact**: 17% reduction in Meta costs

8. ✅ **On-Demand Draft Generation**
   - Generate drafts only when requested
   - Cache draft results
   - **Impact**: 80% reduction in unused drafts

### Phase 4: Monitoring & Alerts (Week 4)

9. ✅ **Cost Dashboard**
   - API endpoint for cost breakdown
   - Frontend dashboard
   - **Impact**: Cost visibility for agencies

10. ✅ **Auto-Reload & Pausing**
    - Wallet balance checks
    - Auto-reload on threshold
    - Pause expensive operations if balance low
    - **Impact**: Prevents service disruption

---

## 8. ESTIMATED COST SAVINGS SUMMARY

### Per Agency Monthly Costs (10 locations, 1,000 conversations)

| Service | Without Optimization | With Optimization | Savings |
|---------|---------------------|-------------------|---------|
| Data4SEO | $1,510 | $1,200 | $310 (20%) |
| Meta Cloud API | $180 | $150 | $30 (17%) |
| OpenAI | $5.80 | $0.63 | $5.17 (89%) |
| **Total** | **$1,695.80** | **$1,350.63** | **$345.17 (20%)** |

### Annual Savings (30 agencies)
- **Without optimization**: $610,488/year
- **With optimization**: $486,227/year
- **Total savings**: **$124,261/year (20% reduction)**

---

## 9. IMMEDIATE ACTION ITEMS

### Before Implementing Any Paid API Integrations

1. ✅ **Implement cost tracking infrastructure** (A8 feature)
2. ✅ **Set up Redis caching layer**
3. ✅ **Implement budget caps per agency**
4. ✅ **Add rate limiting per service**
5. ✅ **Implement request deduplication**
6. ✅ **Set up wallet management system**

### When Implementing Each Service

**Data4SEO**:
- ✅ Cache responses (15min for profiles, 1hr for rankings)
- ✅ Use standard queue for background syncs
- ✅ Per-agency daily budget cap ($50/day)

**Meta Cloud API**:
- ✅ Deduplicate webhooks by message ID
- ✅ Track 24-hour free window per conversation
- ✅ Use idempotency keys for message sends

**OpenAI**:
- ✅ Use GPT-3.5 (not GPT-4) for review drafts
- ✅ Generate drafts on-demand only (not on ingestion)
- ✅ Cache draft results (7-day TTL)
- ✅ Set token limits (200 tokens max)

---

## Conclusion

**Critical Finding**: None of the paid API integrations are currently implemented, but they will significantly impact unit economics once live.

**Recommendation**: Implement cost optimization architecture **before** integrating these APIs to prevent 20%+ cost leaks.

**Priority**: Implement Phase 1 (cost tracking + budget caps + caching) **before** any paid API integrations go live.
