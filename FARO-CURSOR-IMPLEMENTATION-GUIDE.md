# FARO HQ - Complete Cursor Implementation Guide
## Unified Strategic Documentation for AI-Assisted Development

**Version:** 2.0 (Consolidated)  
**Date:** January 24, 2026  
**Purpose:** Single source of truth for Cursor IDE implementation  
**Target:** Full-stack development team using AI-assisted coding

---

## TABLE OF CONTENTS

1. [Executive Summary](#1-executive-summary)
2. [Product Vision & Market](#2-product-vision--market)
3. [Technical Architecture](#3-technical-architecture)
4. [Feature Specifications](#4-feature-specifications)
5. [Epic Mapping & Dependencies](#5-epic-mapping--dependencies)
6. [Implementation Phases](#6-implementation-phases)
7. [Database Schema](#7-database-schema)
8. [API Specifications](#8-api-specifications)
9. [Frontend Components](#9-frontend-components)
10. [Testing & Quality](#10-testing--quality)
11. [Deployment & DevOps](#11-deployment--devops)
12. [Success Metrics](#12-success-metrics)

---

## 1. EXECUTIVE SUMMARY

### What is FARO?

**FARO** is a unified customer messaging platform that enables agencies to manage all customer conversations (WhatsApp, Instagram, Facebook, Email) in a single inbox with AI-powered automation and ROI attribution.

### Key Value Propositions

| Stakeholder | Value |
|-------------|-------|
| **Agencies** | Manage all client messaging in one place, prove ROI |
| **Team Members** | No context switching, unified inbox |
| **Business Owners** | Track revenue from conversations |

### Business Model

```
Freemium Tiers:
├── Free: 100 msg/mo, 1 user, WhatsApp only
├── Pro: $99/mo, 10K msg/mo, 5 users, all channels
└── Enterprise: Custom pricing, unlimited, SSO, dedicated support
```

### Key Dates

| Milestone | Date | Target |
|-----------|------|--------|
| MVP Launch | Feb 7, 2026 | 10 beta customers |
| Phase 2 Complete | Jun 30, 2026 | 1,000 customers, $50K MRR |
| Phase 3 Complete | Sep 30, 2026 | 3,000 customers, $200K MRR |
| Phase 4 Complete | Dec 31, 2026 | 5,000 customers, $500K MRR |

---

## 2. PRODUCT VISION & MARKET

### Problem Statement

Agencies manage customer conversations across 4-5 platforms (WhatsApp, Instagram, Facebook, Email, SMS). Team members waste 3-5 hours/day context-switching. No unified inbox = lost messages, slow responses, poor customer experience.

### Target Market

- **Primary:** Digital marketing agencies (50-500 employees)
- **Secondary:** SMBs with customer service teams
- **Geography:** Initial focus on LATAM (WhatsApp-heavy), then US/EU

### Competitive Positioning

| Competitor | Price | Our Advantage |
|------------|-------|---------------|
| Intercom | $500/mo | 5x cheaper, agency-focused |
| Zendesk | $300/mo | Multi-channel from day 1 |
| Freshdesk | $200/mo | AI built-in, not add-on |
| HubSpot | $800/mo | ROI attribution unique |

### Success Metrics (North Stars)

1. **Activation:** 80% of signups send first message within 24h
2. **Engagement:** Average 50+ conversations/day per workspace
3. **Retention:** <5% monthly churn
4. **Revenue:** $500K MRR by Dec 31, 2026

---

## 3. TECHNICAL ARCHITECTURE

### Stack Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         FARO ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────┤
│  FRONTEND                                                        │
│  ├── Next.js 14 (App Router, Server Components)                 │
│  ├── TypeScript (strict mode)                                   │
│  ├── Tailwind CSS + shadcn/ui                                   │
│  ├── Zustand (client state)                                     │
│  ├── TanStack Query (server state)                              │
│  └── Socket.io (real-time)                                      │
├─────────────────────────────────────────────────────────────────┤
│  BACKEND                                                         │
│  ├── Go 1.21+ (Gin framework)                                   │
│  ├── PostgreSQL 15 (Supabase)                                   │
│  ├── Redis (caching, queues)                                    │
│  ├── WebSocket server (real-time messaging)                     │
│  └── Background workers (Asynq)                                 │
├─────────────────────────────────────────────────────────────────┤
│  INTEGRATIONS                                                    │
│  ├── Twilio (WhatsApp Business API)                             │
│  ├── Meta Graph API (Instagram, Facebook)                       │
│  ├── Google APIs (Gmail, Business Messages)                     │
│  ├── OpenAI (GPT-4, Whisper)                                    │
│  └── Stripe (billing)                                           │
├─────────────────────────────────────────────────────────────────┤
│  INFRASTRUCTURE                                                  │
│  ├── Vercel (frontend hosting)                                  │
│  ├── Railway / Render (backend hosting)                         │
│  ├── Supabase (database, auth, storage)                         │
│  ├── Upstash (Redis)                                            │
│  └── Cloudflare (CDN, custom domains)                           │
├─────────────────────────────────────────────────────────────────┤
│  MONITORING                                                      │
│  ├── Datadog (APM, metrics)                                     │
│  ├── Sentry (error tracking)                                    │
│  ├── PostHog (product analytics)                                │
│  └── Slack (alerts)                                             │
└─────────────────────────────────────────────────────────────────┘
```

### Multi-Tenancy Model

```
Three-Level Hierarchy:
├── Workspace (Tenant)
│   ├── Team Members (Users)
│   ├── Channels (WhatsApp, IG, FB, Email)
│   ├── Contacts (Customers)
│   └── Conversations (Message Threads)
```

**Row-Level Security (RLS) Pattern:**

```sql
-- Every query automatically filtered by workspace_id
CREATE POLICY workspace_isolation ON conversations
  FOR ALL TO authenticated
  USING (workspace_id = current_setting('app.workspace_id')::uuid)
  WITH CHECK (workspace_id = current_setting('app.workspace_id')::uuid);
```

### Data Flow

```
Customer Message Flow:
1. Customer sends WhatsApp message
2. Twilio receives → webhook to /api/webhooks/whatsapp
3. Backend validates, creates message record
4. WebSocket broadcasts to connected team members
5. Message appears in inbox in <1 second

Team Reply Flow:
1. Team member types reply in inbox
2. POST /api/conversations/:id/messages
3. Backend sends to Twilio API
4. Twilio delivers to customer
5. Delivery status webhook updates UI
```

---

## 4. FEATURE SPECIFICATIONS

### MVP Features (8 Total)

| ID | Feature | Owner | Days | Priority |
|----|---------|-------|------|----------|
| F-001 | Authentication & Workspace | Backend | 5 | P0 |
| F-002 | WhatsApp Integration | Backend | 10 | P0 |
| F-003 | Inbox & Message Threading | Frontend | 5 | P0 |
| F-004 | Contact Management | Frontend | 5 | P0 |
| F-005 | Team Collaboration | Frontend | 4 | P0 |
| F-006 | Admin Team Invites | Full Stack | 3 | P0 |
| F-007 | Security & Compliance | DevOps | 5 | P0 |
| F-008 | Monitoring & Alerting | DevOps | 3 | P0 |

---

### F-001: Authentication & Workspace Setup

**Business Goal:** Enable secure multi-team access with proper identity verification

**User Stories:**
```
US-001: As an agency owner, I want to sign up with Google OAuth
        so I don't need to remember another password
        Acceptance: OAuth works, user created, workspace generated

US-002: As an agency owner, I want my workspace created automatically
        so I can start using the product immediately
        Acceptance: Workspace created with unique slug, user is owner

US-003: As a team member, I want to join via invite link
        so I can access my team's workspace
        Acceptance: Click link → OAuth → added to workspace → see inbox
```

**Technical Specification:**

```typescript
// Frontend: Auth Flow
// File: app/auth/page.tsx

'use client';
import { useAuth } from '@/hooks/useAuth';

export default function AuthPage() {
  const { signInWithGoogle, isLoading } = useAuth();
  
  return (
    <div className="min-h-screen flex items-center justify-center">
      <Card className="w-[400px]">
        <CardHeader>
          <CardTitle>Welcome to FARO</CardTitle>
          <CardDescription>Sign in to manage your conversations</CardDescription>
        </CardHeader>
        <CardContent>
          <Button 
            onClick={signInWithGoogle} 
            disabled={isLoading}
            className="w-full"
          >
            <GoogleIcon className="mr-2 h-4 w-4" />
            Continue with Google
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
```

```go
// Backend: Auth Endpoints
// File: internal/api/auth.go

// POST /api/auth/google
type GoogleAuthRequest struct {
    Code string `json:"code" binding:"required"`
}

type AuthResponse struct {
    Token       string    `json:"token"`
    User        User      `json:"user"`
    Workspace   Workspace `json:"workspace"`
    ExpiresAt   time.Time `json:"expires_at"`
}

func (h *AuthHandler) GoogleAuth(c *gin.Context) {
    var req GoogleAuthRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Exchange code for tokens
    token, err := h.googleOAuth.Exchange(c, req.Code)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid OAuth code"})
        return
    }
    
    // Get user info from Google
    userInfo, err := h.googleOAuth.GetUserInfo(token)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to get user info"})
        return
    }
    
    // Find or create user
    user, created, err := h.userService.FindOrCreate(c, userInfo)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to create user"})
        return
    }
    
    // If new user, create workspace
    var workspace *Workspace
    if created {
        workspace, err = h.workspaceService.Create(c, user.ID, userInfo.Name+" Workspace")
        if err != nil {
            c.JSON(500, gin.H{"error": "Failed to create workspace"})
            return
        }
    } else {
        workspace, err = h.workspaceService.GetByUser(c, user.ID)
    }
    
    // Generate JWT
    jwt, expiresAt, err := h.tokenService.Generate(user.ID, workspace.ID)
    
    c.JSON(200, AuthResponse{
        Token:     jwt,
        User:      *user,
        Workspace: *workspace,
        ExpiresAt: expiresAt,
    })
}
```

**Database Schema:**

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    google_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workspaces table
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workspace members (many-to-many)
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- owner, admin, member, viewer
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY workspace_members_policy ON workspace_members
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT workspace_id FROM workspace_members 
        WHERE user_id = current_setting('app.user_id')::uuid
    ));
```

**Acceptance Criteria:**

```gherkin
Feature: User Authentication

  Scenario: New user signs up with Google
    Given I am on the login page
    When I click "Continue with Google"
    And I complete Google OAuth flow
    Then I should be redirected to the inbox
    And a new workspace should be created for me
    And I should be the workspace owner

  Scenario: Existing user signs in
    Given I have an existing account
    When I sign in with Google
    Then I should be redirected to my workspace inbox
    And I should see my existing conversations

  Scenario: User signs in from invite link
    Given I received an invite link via email
    When I click the invite link
    And I complete Google OAuth flow
    Then I should be added to the workspace
    And I should have the role specified in the invite
```

---

### F-002: WhatsApp Integration (Twilio)

**Business Goal:** Enable two-way WhatsApp messaging through the unified inbox

**User Stories:**
```
US-004: As a team member, I want to receive WhatsApp messages in the inbox
        so I can respond to customers without switching apps
        Acceptance: Message appears in <1 second

US-005: As a team member, I want to send WhatsApp replies from the inbox
        so I can help customers efficiently
        Acceptance: Reply delivered, status shows (sent/delivered/read)

US-006: As a team member, I want to see message delivery status
        so I know if the customer received my message
        Acceptance: Status updates in real-time (pending → sent → delivered → read)
```

**Technical Specification:**

```go
// Backend: WhatsApp Webhook Handler
// File: internal/api/webhooks/whatsapp.go

// POST /api/webhooks/whatsapp
func (h *WhatsAppHandler) HandleWebhook(c *gin.Context) {
    // Validate Twilio signature
    signature := c.GetHeader("X-Twilio-Signature")
    body, _ := c.GetRawData()
    
    if !h.twilio.ValidateSignature(signature, body) {
        c.JSON(401, gin.H{"error": "Invalid signature"})
        return
    }
    
    var payload TwilioWebhookPayload
    if err := json.Unmarshal(body, &payload); err != nil {
        c.JSON(400, gin.H{"error": "Invalid payload"})
        return
    }
    
    switch payload.Type {
    case "message":
        h.handleIncomingMessage(c, payload)
    case "status":
        h.handleStatusUpdate(c, payload)
    }
}

func (h *WhatsAppHandler) handleIncomingMessage(c *gin.Context, payload TwilioWebhookPayload) {
    // Find or create contact
    contact, err := h.contactService.FindOrCreateByPhone(c, payload.From)
    if err != nil {
        log.Error("Failed to find/create contact", "error", err)
        c.JSON(500, gin.H{"error": "Internal error"})
        return
    }
    
    // Find or create conversation
    conversation, err := h.conversationService.FindOrCreateByContact(c, contact.ID, "whatsapp")
    if err != nil {
        log.Error("Failed to find/create conversation", "error", err)
        c.JSON(500, gin.H{"error": "Internal error"})
        return
    }
    
    // Create message
    message := &Message{
        ConversationID: conversation.ID,
        ContactID:      contact.ID,
        Direction:      "inbound",
        Channel:        "whatsapp",
        Content:        payload.Body,
        MediaURL:       payload.MediaURL,
        MediaType:      payload.MediaType,
        ExternalID:     payload.MessageSid,
        Status:         "received",
    }
    
    if err := h.messageService.Create(c, message); err != nil {
        log.Error("Failed to create message", "error", err)
        c.JSON(500, gin.H{"error": "Internal error"})
        return
    }
    
    // Broadcast via WebSocket
    h.wsHub.BroadcastToWorkspace(conversation.WorkspaceID, WSEvent{
        Type: "message.new",
        Data: message,
    })
    
    c.JSON(200, gin.H{"status": "ok"})
}

// Send outbound message
// POST /api/conversations/:id/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
    conversationID := c.Param("id")
    
    var req SendMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Get conversation
    conversation, err := h.conversationService.GetByID(c, conversationID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Conversation not found"})
        return
    }
    
    // Create message record (pending status)
    message := &Message{
        ConversationID: conversation.ID,
        ContactID:      conversation.ContactID,
        Direction:      "outbound",
        Channel:        conversation.Channel,
        Content:        req.Content,
        Status:         "pending",
        SenderID:       c.GetString("user_id"),
    }
    
    if err := h.messageService.Create(c, message); err != nil {
        c.JSON(500, gin.H{"error": "Failed to create message"})
        return
    }
    
    // Send via Twilio (async)
    go func() {
        result, err := h.twilio.SendWhatsApp(conversation.Contact.Phone, req.Content)
        if err != nil {
            h.messageService.UpdateStatus(context.Background(), message.ID, "failed")
            return
        }
        h.messageService.UpdateExternalID(context.Background(), message.ID, result.Sid)
        h.messageService.UpdateStatus(context.Background(), message.ID, "sent")
    }()
    
    c.JSON(201, message)
}
```

```typescript
// Frontend: Message Input Component
// File: components/inbox/MessageInput.tsx

'use client';

import { useState, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Send, Paperclip, Smile } from 'lucide-react';

interface MessageInputProps {
  conversationId: string;
}

export function MessageInput({ conversationId }: MessageInputProps) {
  const [content, setContent] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const queryClient = useQueryClient();
  
  const sendMessage = useMutation({
    mutationFn: async (content: string) => {
      const res = await fetch(`/api/conversations/${conversationId}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content }),
      });
      if (!res.ok) throw new Error('Failed to send');
      return res.json();
    },
    onSuccess: () => {
      setContent('');
      queryClient.invalidateQueries({ queryKey: ['messages', conversationId] });
    },
  });
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;
    sendMessage.mutate(content);
  };
  
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="p-4 border-t">
      <div className="flex items-end gap-2">
        <Button type="button" variant="ghost" size="icon">
          <Paperclip className="h-5 w-5" />
        </Button>
        <Button type="button" variant="ghost" size="icon">
          <Smile className="h-5 w-5" />
        </Button>
        <Textarea
          ref={textareaRef}
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message..."
          className="min-h-[40px] max-h-[120px] resize-none"
          rows={1}
        />
        <Button 
          type="submit" 
          size="icon"
          disabled={!content.trim() || sendMessage.isPending}
        >
          <Send className="h-5 w-5" />
        </Button>
      </div>
    </form>
  );
}
```

**Database Schema:**

```sql
-- Channels configuration
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- whatsapp, instagram, facebook, email, gbm
    name VARCHAR(100) NOT NULL,
    config JSONB NOT NULL, -- encrypted credentials
    status VARCHAR(20) DEFAULT 'active', -- active, paused, error
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, type)
);

-- Conversations
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'open', -- open, closed, archived
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    last_message_at TIMESTAMPTZ,
    unread_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Messages
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL, -- denormalized for RLS
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL, -- null for inbound
    direction VARCHAR(10) NOT NULL, -- inbound, outbound
    channel VARCHAR(20) NOT NULL,
    content TEXT,
    media_url VARCHAR(500),
    media_type VARCHAR(50),
    external_id VARCHAR(255), -- Twilio message SID
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, delivered, read, failed
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_workspace ON messages(workspace_id, created_at DESC);
CREATE INDEX idx_conversations_workspace ON conversations(workspace_id, last_message_at DESC);
CREATE INDEX idx_conversations_assigned ON conversations(assigned_to, status);

-- Enable RLS
ALTER TABLE channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;

-- RLS Policies
CREATE POLICY channels_workspace ON channels
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY conversations_workspace ON conversations
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY messages_workspace ON messages
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);
```

**Acceptance Criteria:**

```gherkin
Feature: WhatsApp Messaging

  Scenario: Receive incoming WhatsApp message
    Given the Twilio webhook is configured
    When a customer sends a WhatsApp message
    Then the message should appear in the inbox within 1 second
    And a notification should be sent via WebSocket
    And the conversation should be created if new

  Scenario: Send outbound WhatsApp message
    Given I have an open conversation
    When I type a message and click Send
    Then the message should show "Sending..." status
    And the message should be delivered to the customer
    And the status should update to "Sent" then "Delivered"

  Scenario: Message delivery failure
    Given I send a message to an invalid number
    When Twilio reports a delivery failure
    Then the message status should show "Failed"
    And I should see an error indicator
```

---

### F-003: Inbox & Message Threading

**Business Goal:** Provide a unified, real-time inbox for all conversations

**User Stories:**
```
US-007: As a team member, I want to see all conversations in one list
        so I can quickly find and respond to customers
        Acceptance: List loads in <500ms with 1,000 conversations

US-008: As a team member, I want to see conversation messages in a thread
        so I understand the full context
        Acceptance: Thread shows all messages with timestamps

US-009: As a team member, I want real-time updates
        so I see new messages without refreshing
        Acceptance: New messages appear within 1 second via WebSocket
```

**Technical Specification:**

```typescript
// Frontend: Inbox Layout
// File: app/(dashboard)/inbox/page.tsx

import { Suspense } from 'react';
import { ConversationList } from '@/components/inbox/ConversationList';
import { ConversationThread } from '@/components/inbox/ConversationThread';
import { ContactPanel } from '@/components/inbox/ContactPanel';

export default function InboxPage({
  searchParams,
}: {
  searchParams: { conversation?: string };
}) {
  const selectedId = searchParams.conversation;
  
  return (
    <div className="flex h-[calc(100vh-64px)]">
      {/* Conversation List - 320px fixed width */}
      <div className="w-80 border-r flex-shrink-0">
        <Suspense fallback={<ConversationListSkeleton />}>
          <ConversationList selectedId={selectedId} />
        </Suspense>
      </div>
      
      {/* Message Thread - flexible width */}
      <div className="flex-1 flex flex-col">
        {selectedId ? (
          <Suspense fallback={<ThreadSkeleton />}>
            <ConversationThread conversationId={selectedId} />
          </Suspense>
        ) : (
          <EmptyState />
        )}
      </div>
      
      {/* Contact Panel - 280px fixed width */}
      {selectedId && (
        <div className="w-72 border-l flex-shrink-0 hidden lg:block">
          <Suspense fallback={<ContactPanelSkeleton />}>
            <ContactPanel conversationId={selectedId} />
          </Suspense>
        </div>
      )}
    </div>
  );
}
```

```typescript
// Frontend: Conversation List Component
// File: components/inbox/ConversationList.tsx

'use client';

import { useQuery } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { useWebSocket } from '@/hooks/useWebSocket';
import { cn } from '@/lib/utils';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { formatRelativeTime } from '@/lib/date';

interface Conversation {
  id: string;
  contact: {
    id: string;
    name: string;
    avatar_url?: string;
    phone: string;
  };
  channel: string;
  status: string;
  last_message: {
    content: string;
    created_at: string;
  };
  unread_count: number;
  assigned_to?: {
    id: string;
    name: string;
  };
}

export function ConversationList({ selectedId }: { selectedId?: string }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  
  const { data, isLoading, refetch } = useQuery({
    queryKey: ['conversations'],
    queryFn: async () => {
      const res = await fetch('/api/conversations');
      if (!res.ok) throw new Error('Failed to fetch');
      return res.json() as Promise<{ conversations: Conversation[] }>;
    },
    refetchInterval: 30000, // Refetch every 30s as fallback
  });
  
  // Real-time updates via WebSocket
  useWebSocket('message.new', (event) => {
    refetch();
  });
  
  useWebSocket('conversation.updated', (event) => {
    refetch();
  });
  
  const selectConversation = (id: string) => {
    const params = new URLSearchParams(searchParams);
    params.set('conversation', id);
    router.push(`/inbox?${params.toString()}`);
  };
  
  if (isLoading) return <ConversationListSkeleton />;
  
  return (
    <div className="flex flex-col h-full">
      {/* Search & Filters */}
      <div className="p-4 border-b">
        <Input 
          placeholder="Search conversations..." 
          className="w-full"
        />
        <div className="flex gap-2 mt-2">
          <Badge variant="secondary">All</Badge>
          <Badge variant="outline">Unread</Badge>
          <Badge variant="outline">Assigned to me</Badge>
        </div>
      </div>
      
      {/* Conversation List */}
      <div className="flex-1 overflow-y-auto">
        {data?.conversations.map((conversation) => (
          <button
            key={conversation.id}
            onClick={() => selectConversation(conversation.id)}
            className={cn(
              "w-full p-4 flex items-start gap-3 hover:bg-muted/50 transition-colors border-b",
              selectedId === conversation.id && "bg-muted"
            )}
          >
            <Avatar>
              <AvatarImage src={conversation.contact.avatar_url} />
              <AvatarFallback>
                {conversation.contact.name.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            
            <div className="flex-1 min-w-0 text-left">
              <div className="flex items-center justify-between">
                <span className="font-medium truncate">
                  {conversation.contact.name}
                </span>
                <span className="text-xs text-muted-foreground">
                  {formatRelativeTime(conversation.last_message.created_at)}
                </span>
              </div>
              
              <p className="text-sm text-muted-foreground truncate">
                {conversation.last_message.content}
              </p>
              
              <div className="flex items-center gap-2 mt-1">
                <ChannelIcon channel={conversation.channel} />
                {conversation.unread_count > 0 && (
                  <Badge variant="default" className="h-5 px-1.5">
                    {conversation.unread_count}
                  </Badge>
                )}
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
```

```typescript
// Frontend: Message Thread Component
// File: components/inbox/ConversationThread.tsx

'use client';

import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useRef } from 'react';
import { useWebSocket } from '@/hooks/useWebSocket';
import { MessageBubble } from './MessageBubble';
import { MessageInput } from './MessageInput';
import { ThreadHeader } from './ThreadHeader';

interface Message {
  id: string;
  direction: 'inbound' | 'outbound';
  content: string;
  media_url?: string;
  media_type?: string;
  status: string;
  sender?: { id: string; name: string };
  created_at: string;
}

export function ConversationThread({ conversationId }: { conversationId: string }) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();
  
  const { data, isLoading } = useQuery({
    queryKey: ['messages', conversationId],
    queryFn: async () => {
      const res = await fetch(`/api/conversations/${conversationId}/messages`);
      if (!res.ok) throw new Error('Failed to fetch');
      return res.json() as Promise<{ messages: Message[]; conversation: any }>;
    },
  });
  
  // Real-time message updates
  useWebSocket('message.new', (event) => {
    if (event.data.conversation_id === conversationId) {
      queryClient.setQueryData(['messages', conversationId], (old: any) => ({
        ...old,
        messages: [...(old?.messages || []), event.data],
      }));
    }
  });
  
  useWebSocket('message.status', (event) => {
    if (event.data.conversation_id === conversationId) {
      queryClient.setQueryData(['messages', conversationId], (old: any) => ({
        ...old,
        messages: old?.messages?.map((m: Message) =>
          m.id === event.data.message_id
            ? { ...m, status: event.data.status }
            : m
        ),
      }));
    }
  });
  
  // Auto-scroll to bottom on new messages
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [data?.messages]);
  
  if (isLoading) return <ThreadSkeleton />;
  
  return (
    <div className="flex flex-col h-full">
      <ThreadHeader conversation={data?.conversation} />
      
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {data?.messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
        ))}
        <div ref={messagesEndRef} />
      </div>
      
      <MessageInput conversationId={conversationId} />
    </div>
  );
}
```

**Acceptance Criteria:**

```gherkin
Feature: Inbox & Threading

  Scenario: Load conversation list
    Given I am logged in
    When I navigate to the inbox
    Then I should see all my workspace conversations
    And the list should load in under 500ms
    And conversations should be sorted by last message time

  Scenario: Real-time message updates
    Given I have the inbox open
    When a new message arrives via WebSocket
    Then the message should appear in the thread
    And the conversation should move to the top of the list
    And the unread count should update

  Scenario: View conversation thread
    Given I click on a conversation
    When the thread loads
    Then I should see all messages in chronological order
    And I should see message delivery status
    And I should be able to send a reply
```

---

### F-004: Contact Management

**Business Goal:** Organize and search customer contacts efficiently

**Technical Specification:**

```sql
-- Contacts table
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    phone VARCHAR(20),
    email VARCHAR(255),
    name VARCHAR(100),
    avatar_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, blocked
    sources TEXT[] DEFAULT ARRAY[]::TEXT[], -- channels they've messaged from
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    notes TEXT,
    custom_fields JSONB DEFAULT '{}',
    message_count INTEGER DEFAULT 0,
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, phone)
);

-- Full-text search index
CREATE INDEX idx_contacts_search ON contacts 
    USING GIN (to_tsvector('english', coalesce(name, '') || ' ' || coalesce(email, '') || ' ' || coalesce(phone, '')));

-- Enable RLS
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;

CREATE POLICY contacts_workspace ON contacts
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);
```

```go
// Backend: Contact API
// File: internal/api/contacts.go

// GET /api/contacts
func (h *ContactHandler) List(c *gin.Context) {
    workspaceID := c.GetString("workspace_id")
    
    query := c.DefaultQuery("q", "")
    status := c.DefaultQuery("status", "")
    limit := c.DefaultQuery("limit", "50")
    offset := c.DefaultQuery("offset", "0")
    
    contacts, total, err := h.contactService.List(c, workspaceID, ListContactsParams{
        Query:  query,
        Status: status,
        Limit:  parseInt(limit),
        Offset: parseInt(offset),
    })
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch contacts"})
        return
    }
    
    c.JSON(200, gin.H{
        "contacts": contacts,
        "total":    total,
        "limit":    parseInt(limit),
        "offset":   parseInt(offset),
    })
}

// GET /api/contacts/:id
func (h *ContactHandler) Get(c *gin.Context) {
    contactID := c.Param("id")
    
    contact, err := h.contactService.GetByID(c, contactID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Contact not found"})
        return
    }
    
    // Get recent conversations
    conversations, _ := h.conversationService.GetByContact(c, contactID, 5)
    
    c.JSON(200, gin.H{
        "contact":       contact,
        "conversations": conversations,
    })
}

// PATCH /api/contacts/:id
func (h *ContactHandler) Update(c *gin.Context) {
    contactID := c.Param("id")
    
    var req UpdateContactRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    contact, err := h.contactService.Update(c, contactID, req)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to update contact"})
        return
    }
    
    c.JSON(200, gin.H{"contact": contact})
}
```

---

### F-005: Team Collaboration & Assignment

**Business Goal:** Enable team coordination by assigning conversations

**Technical Specification:**

```sql
-- Add assignment fields to conversations
ALTER TABLE conversations
ADD COLUMN assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN assigned_at TIMESTAMPTZ;

-- Team notes (private, not sent to customer)
CREATE TABLE conversation_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    mentioned_user_ids UUID[] DEFAULT ARRAY[]::UUID[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- assignment, mention, message
    title VARCHAR(255) NOT NULL,
    body TEXT,
    resource_type VARCHAR(50),
    resource_id UUID,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE conversation_notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;

CREATE POLICY notes_workspace ON conversation_notes
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY notifications_user ON notifications
    FOR ALL TO authenticated
    USING (user_id = current_setting('app.user_id')::uuid);
```

```go
// Backend: Assignment API
// File: internal/api/assignments.go

// PATCH /api/conversations/:id/assign
func (h *ConversationHandler) Assign(c *gin.Context) {
    conversationID := c.Param("id")
    userID := c.GetString("user_id")
    
    var req AssignRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    conversation, err := h.conversationService.Assign(c, conversationID, req.AssignedTo)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to assign"})
        return
    }
    
    // Send notification to assignee
    if req.AssignedTo != nil && *req.AssignedTo != userID {
        h.notificationService.Create(c, &Notification{
            UserID:       *req.AssignedTo,
            Type:         "assignment",
            Title:        "Conversation assigned to you",
            Body:         fmt.Sprintf("%s assigned you a conversation", c.GetString("user_name")),
            ResourceType: "conversation",
            ResourceID:   conversationID,
        })
    }
    
    // Broadcast update
    h.wsHub.BroadcastToWorkspace(conversation.WorkspaceID, WSEvent{
        Type: "conversation.updated",
        Data: conversation,
    })
    
    c.JSON(200, gin.H{"conversation": conversation})
}
```

---

### F-006: Admin Team Invites

**Business Goal:** Allow workspace owners to manage team member access

**Technical Specification:**

```sql
-- Workspace invites
CREATE TABLE workspace_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    token VARCHAR(64) UNIQUE NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, declined, expired
    expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '7 days',
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, email)
);

CREATE INDEX idx_invites_token ON workspace_invites(token);
```

```go
// Backend: Team Invite API
// File: internal/api/team.go

// POST /api/workspaces/:id/invites
func (h *TeamHandler) SendInvite(c *gin.Context) {
    workspaceID := c.Param("id")
    userID := c.GetString("user_id")
    
    // Check if user is owner/admin
    member, _ := h.workspaceService.GetMember(c, workspaceID, userID)
    if member.Role != "owner" && member.Role != "admin" {
        c.JSON(403, gin.H{"error": "Only owners and admins can invite"})
        return
    }
    
    var req InviteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    // Generate secure token
    token := generateSecureToken(32)
    
    invite := &WorkspaceInvite{
        WorkspaceID: workspaceID,
        Email:       req.Email,
        Role:        req.Role,
        Token:       token,
        InvitedBy:   userID,
    }
    
    if err := h.inviteService.Create(c, invite); err != nil {
        c.JSON(500, gin.H{"error": "Failed to create invite"})
        return
    }
    
    // Send email
    go h.emailService.SendInvite(invite)
    
    c.JSON(201, gin.H{"invite": invite})
}

// POST /api/invites/:token/accept
func (h *TeamHandler) AcceptInvite(c *gin.Context) {
    token := c.Param("token")
    userID := c.GetString("user_id")
    
    invite, err := h.inviteService.GetByToken(c, token)
    if err != nil {
        c.JSON(404, gin.H{"error": "Invite not found"})
        return
    }
    
    if invite.Status != "pending" {
        c.JSON(400, gin.H{"error": "Invite already used"})
        return
    }
    
    if time.Now().After(invite.ExpiresAt) {
        c.JSON(400, gin.H{"error": "Invite expired"})
        return
    }
    
    // Add user to workspace
    err = h.workspaceService.AddMember(c, invite.WorkspaceID, userID, invite.Role)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to add to workspace"})
        return
    }
    
    // Mark invite as accepted
    h.inviteService.MarkAccepted(c, invite.ID)
    
    workspace, _ := h.workspaceService.GetByID(c, invite.WorkspaceID)
    
    c.JSON(200, gin.H{"workspace": workspace})
}
```

---

### F-007: Security Foundation & Compliance

**Business Goal:** Establish baseline security and compliance framework

**Technical Specification:**

```sql
-- Audit logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    result VARCHAR(20) DEFAULT 'success', -- success, failed
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_workspace ON audit_logs(workspace_id, created_at DESC);
CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at DESC);

-- Enable RLS
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_workspace ON audit_logs
    FOR SELECT TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);
```

```go
// Backend: Audit Logging Middleware
// File: internal/middleware/audit.go

func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Process request
        c.Next()
        
        // Skip non-mutating requests
        if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
            return
        }
        
        // Log audit event
        event := AuditLog{
            WorkspaceID:  c.GetString("workspace_id"),
            UserID:       c.GetString("user_id"),
            Action:       getAction(c),
            ResourceType: getResourceType(c),
            ResourceID:   getResourceID(c),
            IPAddress:    c.ClientIP(),
            UserAgent:    c.Request.UserAgent(),
            Result:       getResult(c),
        }
        
        go db.Create(&event)
    }
}

func getAction(c *gin.Context) string {
    method := c.Request.Method
    path := c.FullPath()
    
    actions := map[string]map[string]string{
        "POST": {
            "/api/conversations/:id/messages": "message.send",
            "/api/workspaces/:id/invites":     "team.invite",
        },
        "PATCH": {
            "/api/conversations/:id/assign": "conversation.assign",
            "/api/contacts/:id":             "contact.update",
        },
        "DELETE": {
            "/api/workspaces/:id/team/:member_id": "team.remove",
        },
    }
    
    if methodActions, ok := actions[method]; ok {
        if action, ok := methodActions[path]; ok {
            return action
        }
    }
    
    return fmt.Sprintf("%s.%s", strings.ToLower(method), path)
}
```

---

### F-008: Monitoring & Alerting

**Business Goal:** Detect and alert on critical system issues

**Technical Specification:**

```go
// Backend: Metrics & Health
// File: internal/monitoring/metrics.go

import (
    "github.com/DataDog/datadog-go/v5/statsd"
    "github.com/getsentry/sentry-go"
)

var client *statsd.Client

func Init() {
    var err error
    client, err = statsd.New("127.0.0.1:8125")
    if err != nil {
        log.Fatal(err)
    }
    
    // Initialize Sentry
    sentry.Init(sentry.ClientOptions{
        Dsn:              os.Getenv("SENTRY_DSN"),
        Environment:      os.Getenv("ENVIRONMENT"),
        TracesSampleRate: 0.1,
    })
}

// Track API request
func TrackRequest(endpoint string, duration time.Duration, status int) {
    tags := []string{
        "endpoint:" + endpoint,
        "status:" + strconv.Itoa(status),
    }
    
    client.Timing("api.request.duration", duration, tags, 1)
    client.Incr("api.request.count", tags, 1)
    
    if status >= 500 {
        client.Incr("api.request.error", tags, 1)
    }
}

// Track message
func TrackMessage(direction string, channel string) {
    tags := []string{
        "direction:" + direction,
        "channel:" + channel,
    }
    
    client.Incr("message.count", tags, 1)
}

// Report error
func ReportError(err error, ctx map[string]interface{}) {
    sentry.WithScope(func(scope *sentry.Scope) {
        for k, v := range ctx {
            scope.SetExtra(k, v)
        }
        sentry.CaptureException(err)
    })
}
```

```go
// Backend: Health Check Endpoint
// File: internal/api/health.go

// GET /health
func (h *HealthHandler) Check(c *gin.Context) {
    checks := map[string]string{
        "database":  "ok",
        "redis":     "ok",
        "twilio":    "ok",
        "websocket": "ok",
    }
    
    // Check database
    if err := h.db.Exec("SELECT 1").Error; err != nil {
        checks["database"] = "error"
    }
    
    // Check Redis
    if _, err := h.redis.Ping(c).Result(); err != nil {
        checks["redis"] = "error"
    }
    
    // Determine overall status
    status := "healthy"
    for _, v := range checks {
        if v != "ok" {
            status = "degraded"
            break
        }
    }
    
    c.JSON(200, gin.H{
        "status":    status,
        "checks":    checks,
        "timestamp": time.Now().UTC(),
    })
}
```

---

## 5. EPIC MAPPING & DEPENDENCIES

### Epic Structure

```
FARO Product Epics
│
├── MVP (Feb 2026)
│   ├── MVP-001: Core Messaging & Inbox
│   │   ├── F-002: WhatsApp Integration
│   │   ├── F-003: Inbox & Threading
│   │   └── F-004: Contact Management
│   │
│   ├── MVP-002: Authentication & Workspace
│   │   ├── F-001: Authentication
│   │   ├── F-006: Admin Team Invites
│   │   └── F-007: Security Foundation
│   │
│   └── MVP-003: Team Collaboration
│       ├── F-005: Team Collaboration
│       └── F-008: Monitoring
│
├── Phase 2 (Apr-Jun 2026)
│   ├── P2-001: Multi-Channel Messaging
│   │   ├── F-009: Instagram DMs
│   │   ├── F-010: Facebook Messenger
│   │   ├── F-011: Email Integration
│   │   └── F-012: Google Business Messages
│   │
│   ├── P2-002: AI & Smart Features
│   │   ├── F-013: AI Message Suggestions
│   │   ├── F-014: Response Templates
│   │   └── F-015: Basic Analytics
│   │
│   └── P2-003: Monetization
│       ├── F-016: Lead Capture
│       ├── F-017: Billing & Subscriptions
│       └── F-018: Usage Tracking
│
├── Phase 3 (Jul-Sep 2026)
│   ├── P3-001: Conversation Automation
│   │   ├── F-019: Booking Automation
│   │   ├── F-020: AI Chatbot
│   │   └── F-021: Rules Engine
│   │
│   └── P3-002: Analytics & Attribution
│       ├── F-022: Source Attribution
│       ├── F-023: ROI Dashboard
│       └── F-024: Compliance Reports
│
└── Phase 4 (Oct-Dec 2026)
    ├── P4-001: E-Commerce
    │   ├── F-028: Shopify Integration
    │   ├── F-029: Payment Links
    │   └── F-030: Order Tracking
    │
    └── P4-002: Platform & APIs
        ├── F-031: Portuguese Localization
        ├── F-034: REST API v1
        ├── F-035: Webhooks
        └── F-036: Third-Party Integrations
```

### Dependency Graph

```
MVP Dependencies:
F-001 (Auth) ──┬── F-003 (Inbox)
               ├── F-004 (Contacts)
               ├── F-005 (Team Collab)
               ├── F-006 (Invites)
               └── F-007 (Security)

F-002 (WhatsApp) ── F-003 (Inbox)

F-003 (Inbox) ── F-005 (Team Collab)

F-008 (Monitoring) ── Independent

Phase 2 Dependencies:
MVP Complete ──┬── F-009 (Instagram)
               ├── F-010 (Facebook)
               ├── F-011 (Email)
               └── F-012 (GBM)

F-003 (Inbox) ── F-013 (AI Suggestions)

F-001 (Auth) ── F-017 (Billing)
```

---

## 6. IMPLEMENTATION PHASES

### MVP Timeline (Jan 22 - Feb 7, 2026)

```
Week 1 (Jan 22-26): Foundation
├── Day 1-2: Project setup, CI/CD, database schema
├── Day 3-5: F-001 Auth + F-007 Security (parallel)
└── Day 3-5: F-002 WhatsApp webhook setup

Week 2 (Jan 27-31): Core Features
├── Day 6-8: F-003 Inbox UI + F-002 WhatsApp complete
├── Day 7-9: F-004 Contacts + F-005 Team Collab
├── Day 9-10: F-006 Invites + F-008 Monitoring
└── Day 10: Integration testing

Week 3 (Feb 1-7): Launch
├── Feb 1: Soft launch (10 beta customers)
├── Feb 2-6: Bug fixes, performance optimization
└── Feb 7: Public launch
```

### Phase 2 Timeline (Apr 1 - Jun 30, 2026)

```
Sprint 1-2 (Apr 1-14): Multi-Channel Foundation
├── F-009: Instagram DM Integration
├── F-010: Facebook Messenger Integration
└── Meta API authentication, unified webhook handling

Sprint 3-4 (Apr 15-28): Email & GBM
├── F-011: Email Integration (SendGrid)
├── F-012: Google Business Messages
└── Channel-specific message formatting

Sprint 5-6 (May 1-14): AI Features
├── F-013: AI Message Suggestions (OpenAI)
├── F-014: Response Templates
└── Context-aware suggestion engine

Sprint 7-8 (May 15-31): Analytics
├── F-015: Basic Analytics Dashboard
└── Message volume, response time, sentiment

Sprint 9-10 (Jun 1-14): Leads & Billing
├── F-016: Lead Capture & Qualification
├── F-017: Billing & Subscriptions (Stripe)
└── F-018: Usage Tracking & Metering

Sprint 11-13 (Jun 15-30): Polish & Launch
├── Performance optimization
├── Enterprise features prep
└── Phase 2 public launch
```

---

## 7. DATABASE SCHEMA

### Complete Schema (MVP)

```sql
-- ============================================
-- FARO DATABASE SCHEMA
-- Version: 1.0 (MVP)
-- ============================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================
-- CORE TABLES
-- ============================================

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    google_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workspaces
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    settings JSONB DEFAULT '{}',
    plan VARCHAR(20) DEFAULT 'free', -- free, pro, enterprise
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Workspace Members
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- owner, admin, member, viewer
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

-- Workspace Invites
CREATE TABLE workspace_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    token VARCHAR(64) UNIQUE NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    expires_at TIMESTAMPTZ DEFAULT NOW() + INTERVAL '7 days',
    accepted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, email)
);

-- ============================================
-- MESSAGING TABLES
-- ============================================

-- Channels
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- whatsapp, instagram, facebook, email, gbm
    name VARCHAR(100) NOT NULL,
    config JSONB NOT NULL, -- encrypted credentials
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, type)
);

-- Contacts
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    phone VARCHAR(20),
    email VARCHAR(255),
    name VARCHAR(100),
    avatar_url VARCHAR(500),
    status VARCHAR(20) DEFAULT 'active',
    sources TEXT[] DEFAULT ARRAY[]::TEXT[],
    tags TEXT[] DEFAULT ARRAY[]::TEXT[],
    notes TEXT,
    custom_fields JSONB DEFAULT '{}',
    message_count INTEGER DEFAULT 0,
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, phone)
);

-- Conversations
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'open',
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ,
    last_message_at TIMESTAMPTZ,
    last_message_preview TEXT,
    unread_count INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Messages
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,
    contact_id UUID REFERENCES contacts(id) ON DELETE SET NULL,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    direction VARCHAR(10) NOT NULL, -- inbound, outbound
    channel VARCHAR(20) NOT NULL,
    content TEXT,
    media_url VARCHAR(500),
    media_type VARCHAR(50),
    external_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- COLLABORATION TABLES
-- ============================================

-- Conversation Notes
CREATE TABLE conversation_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    workspace_id UUID NOT NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    mentioned_user_ids UUID[] DEFAULT ARRAY[]::UUID[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    resource_type VARCHAR(50),
    resource_id UUID,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- SECURITY TABLES
-- ============================================

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}',
    result VARCHAR(20) DEFAULT 'success',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);

-- Workspaces
CREATE INDEX idx_workspaces_slug ON workspaces(slug);
CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);

-- Workspace Members
CREATE INDEX idx_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_members_user ON workspace_members(user_id);

-- Invites
CREATE INDEX idx_invites_token ON workspace_invites(token);
CREATE INDEX idx_invites_email ON workspace_invites(email);

-- Contacts
CREATE INDEX idx_contacts_workspace ON contacts(workspace_id);
CREATE INDEX idx_contacts_phone ON contacts(workspace_id, phone);
CREATE INDEX idx_contacts_search ON contacts USING GIN (
    to_tsvector('english', coalesce(name, '') || ' ' || coalesce(email, '') || ' ' || coalesce(phone, ''))
);

-- Conversations
CREATE INDEX idx_conversations_workspace ON conversations(workspace_id, last_message_at DESC);
CREATE INDEX idx_conversations_contact ON conversations(contact_id);
CREATE INDEX idx_conversations_assigned ON conversations(assigned_to, status);

-- Messages
CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_workspace ON messages(workspace_id, created_at DESC);
CREATE INDEX idx_messages_external ON messages(external_id);

-- Notifications
CREATE INDEX idx_notifications_user ON notifications(user_id, read, created_at DESC);

-- Audit Logs
CREATE INDEX idx_audit_workspace ON audit_logs(workspace_id, created_at DESC);
CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at DESC);

-- ============================================
-- ROW LEVEL SECURITY
-- ============================================

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE workspace_invites ENABLE ROW LEVEL SECURITY;
ALTER TABLE channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE conversation_notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- Workspace-based policies
CREATE POLICY workspace_members_policy ON workspace_members
    FOR ALL TO authenticated
    USING (workspace_id IN (
        SELECT workspace_id FROM workspace_members 
        WHERE user_id = current_setting('app.user_id')::uuid
    ));

CREATE POLICY channels_policy ON channels
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY contacts_policy ON contacts
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY conversations_policy ON conversations
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY messages_policy ON messages
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY notes_policy ON conversation_notes
    FOR ALL TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

CREATE POLICY notifications_policy ON notifications
    FOR SELECT TO authenticated
    USING (user_id = current_setting('app.user_id')::uuid);

CREATE POLICY audit_policy ON audit_logs
    FOR SELECT TO authenticated
    USING (workspace_id = current_setting('app.workspace_id')::uuid);

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_workspaces_updated_at BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_contacts_updated_at BEFORE UPDATE ON contacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Update conversation on new message
CREATE OR REPLACE FUNCTION update_conversation_on_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations SET
        last_message_at = NEW.created_at,
        last_message_preview = LEFT(NEW.content, 100),
        unread_count = CASE 
            WHEN NEW.direction = 'inbound' THEN unread_count + 1 
            ELSE unread_count 
        END,
        updated_at = NOW()
    WHERE id = NEW.conversation_id;
    
    -- Update contact message count
    UPDATE contacts SET
        message_count = message_count + 1,
        last_message_at = NEW.created_at
    WHERE id = NEW.contact_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_conversation_on_message AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION update_conversation_on_message();
```

---

## 8. API SPECIFICATIONS

### API Overview

```
Base URL: https://api.faro.app/v1
Authentication: Bearer token (JWT)
Content-Type: application/json

Rate Limits:
- Free: 100 requests/minute
- Pro: 1,000 requests/minute
- Enterprise: 10,000 requests/minute
```

### Endpoints Summary

```
Authentication:
POST   /auth/google              - Google OAuth
POST   /auth/refresh             - Refresh token
POST   /auth/logout              - Logout

Workspaces:
GET    /workspaces               - List user's workspaces
GET    /workspaces/:id           - Get workspace details
PATCH  /workspaces/:id           - Update workspace
POST   /workspaces/:id/invites   - Send invite
GET    /workspaces/:id/team      - List team members
PATCH  /workspaces/:id/team/:id  - Update member role
DELETE /workspaces/:id/team/:id  - Remove member

Conversations:
GET    /conversations            - List conversations
GET    /conversations/:id        - Get conversation
PATCH  /conversations/:id        - Update conversation
PATCH  /conversations/:id/assign - Assign conversation
POST   /conversations/:id/close  - Close conversation

Messages:
GET    /conversations/:id/messages  - List messages
POST   /conversations/:id/messages  - Send message

Contacts:
GET    /contacts                 - List contacts
GET    /contacts/:id             - Get contact
PATCH  /contacts/:id             - Update contact
POST   /contacts/:id/block       - Block contact

Notes:
GET    /conversations/:id/notes  - List notes
POST   /conversations/:id/notes  - Create note
DELETE /notes/:id                - Delete note

Notifications:
GET    /notifications            - List notifications
POST   /notifications/:id/read   - Mark as read
POST   /notifications/read-all   - Mark all as read

Webhooks:
POST   /webhooks/whatsapp        - Twilio WhatsApp webhook
POST   /webhooks/twilio/status   - Twilio status webhook

Health:
GET    /health                   - Health check
```

---

## 9. FRONTEND COMPONENTS

### Component Library (shadcn/ui based)

```
components/
├── ui/                          # Base components (shadcn)
│   ├── button.tsx
│   ├── input.tsx
│   ├── textarea.tsx
│   ├── avatar.tsx
│   ├── badge.tsx
│   ├── card.tsx
│   ├── dialog.tsx
│   ├── dropdown-menu.tsx
│   ├── skeleton.tsx
│   ├── toast.tsx
│   └── ...
│
├── inbox/                       # Inbox-specific components
│   ├── ConversationList.tsx
│   ├── ConversationCard.tsx
│   ├── ConversationThread.tsx
│   ├── MessageBubble.tsx
│   ├── MessageInput.tsx
│   ├── ThreadHeader.tsx
│   ├── ContactPanel.tsx
│   ├── ChannelIcon.tsx
│   └── MessageStatus.tsx
│
├── contacts/                    # Contact components
│   ├── ContactList.tsx
│   ├── ContactCard.tsx
│   ├── ContactProfile.tsx
│   ├── ContactSearch.tsx
│   └── ContactTags.tsx
│
├── team/                        # Team management
│   ├── TeamList.tsx
│   ├── TeamMemberCard.tsx
│   ├── InviteModal.tsx
│   └── RoleSelector.tsx
│
├── settings/                    # Settings components
│   ├── WorkspaceSettings.tsx
│   ├── ChannelSettings.tsx
│   ├── NotificationSettings.tsx
│   └── SecuritySettings.tsx
│
└── shared/                      # Shared components
    ├── Layout.tsx
    ├── Sidebar.tsx
    ├── Header.tsx
    ├── EmptyState.tsx
    ├── LoadingState.tsx
    └── ErrorBoundary.tsx
```

### Page Structure

```
app/
├── (auth)/
│   ├── login/page.tsx
│   ├── signup/page.tsx
│   └── invite/[token]/page.tsx
│
├── (dashboard)/
│   ├── layout.tsx               # Dashboard layout with sidebar
│   ├── inbox/
│   │   └── page.tsx            # Inbox page
│   ├── contacts/
│   │   ├── page.tsx            # Contact list
│   │   └── [id]/page.tsx       # Contact detail
│   ├── settings/
│   │   ├── page.tsx            # General settings
│   │   ├── team/page.tsx       # Team management
│   │   ├── channels/page.tsx   # Channel config
│   │   └── security/page.tsx   # Security settings
│   └── analytics/
│       └── page.tsx            # Analytics dashboard (Phase 2)
│
├── api/                         # API routes
│   ├── auth/
│   │   └── [...nextauth]/route.ts
│   └── webhooks/
│       └── whatsapp/route.ts
│
└── layout.tsx                   # Root layout
```

---

## 10. TESTING & QUALITY

### Testing Strategy

```
Test Types:
├── Unit Tests (80% coverage target)
│   ├── Backend: Go test
│   └── Frontend: Jest + Testing Library
│
├── Integration Tests
│   ├── API endpoints
│   └── Database operations
│
├── E2E Tests (Playwright)
│   ├── Auth flows
│   ├── Messaging flows
│   └── Settings flows
│
└── Performance Tests
    ├── Load testing (k6)
    └── Lighthouse CI
```

### Quality Gates

```
Before Merge:
- [ ] All tests pass
- [ ] Code coverage >= 80%
- [ ] No TypeScript errors
- [ ] No ESLint warnings
- [ ] PR approved by 2 reviewers

Before Deploy:
- [ ] Staging tests pass
- [ ] Security scan clean
- [ ] Performance benchmarks met
- [ ] Rollback plan ready
```

---

## 11. DEPLOYMENT & DEVOPS

### Infrastructure

```
Production:
├── Vercel (Frontend)
│   ├── Edge functions
│   ├── CDN
│   └── Preview deployments
│
├── Railway (Backend)
│   ├── Go API server
│   ├── WebSocket server
│   └── Background workers
│
├── Supabase (Database)
│   ├── PostgreSQL
│   ├── Auth (optional)
│   └── Storage
│
├── Upstash (Redis)
│   ├── Caching
│   └── Rate limiting
│
└── Cloudflare
    ├── DNS
    ├── SSL
    └── Custom domains
```

### CI/CD Pipeline

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run tests
        run: |
          go test ./...
          npm test

  deploy-backend:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy to Railway
        uses: railway/deploy-action@v1
        with:
          token: ${{ secrets.RAILWAY_TOKEN }}

  deploy-frontend:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Deploy to Vercel
        uses: vercel/action@v1
        with:
          token: ${{ secrets.VERCEL_TOKEN }}
```

---

## 12. SUCCESS METRICS

### MVP Success Criteria (Feb 7, 2026)

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Uptime | 99.5% | Datadog |
| API Response (p95) | <200ms | Datadog APM |
| Message Delivery | <1s | End-to-end tracking |
| Error Rate | <0.1% | Sentry |
| Beta Customers | 10+ | Auth records |
| NPS | >50 | Survey |

### Phase 2 Success Criteria (Jun 30, 2026)

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Paying Customers | 1,000+ | Stripe |
| MRR | $50K+ | Stripe |
| Channels Active | 4 | Feature flags |
| User Retention | 80%+ | Cohort analysis |
| AI Suggestion Usage | 40%+ | Event tracking |

### Year 1 Success Criteria (Dec 31, 2026)

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Paying Customers | 5,000+ | Stripe |
| MRR | $500K+ | Stripe |
| Team Size | 10+ FTE | HR records |
| Net Retention | 120%+ | Revenue cohorts |
| Series A | $2-3M | Closed funding |

---

## APPENDIX A: CURSOR PROMPTS

### Feature Development Prompt

```
You are building FARO, a unified customer messaging platform.

Context:
- Stack: Next.js 14, Go, PostgreSQL, Supabase, Twilio
- Current Phase: MVP
- Feature: [FEATURE_NAME]

Requirements:
[PASTE FEATURE SPEC FROM THIS DOCUMENT]

Please implement:
1. Database schema changes
2. Backend API endpoints
3. Frontend components
4. Tests

Follow these patterns:
- Use shadcn/ui components
- Use TanStack Query for data fetching
- Use Zustand for client state
- Follow RLS patterns for multi-tenancy
```

### Bug Fix Prompt

```
You are debugging FARO.

Issue: [DESCRIBE BUG]
Expected: [EXPECTED BEHAVIOR]
Actual: [ACTUAL BEHAVIOR]

Relevant code:
[PASTE CODE]

Please:
1. Identify the root cause
2. Propose a fix
3. Add tests to prevent regression
```

### Code Review Prompt

```
Review this code for FARO:

[PASTE CODE]

Check for:
1. Security (RLS, input validation, auth)
2. Performance (N+1 queries, indexes)
3. Error handling
4. TypeScript types
5. Test coverage
```

---

## APPENDIX B: GLOSSARY

| Term | Definition |
|------|------------|
| **Workspace** | A tenant (organization/agency) using FARO |
| **Channel** | A messaging platform (WhatsApp, Instagram, etc.) |
| **Contact** | A customer who messages through any channel |
| **Conversation** | A thread of messages with a contact |
| **RLS** | Row-Level Security (Postgres feature for multi-tenancy) |
| **MRR** | Monthly Recurring Revenue |
| **NPS** | Net Promoter Score |

---

**Document Version:** 2.0  
**Last Updated:** January 24, 2026  
**Maintained By:** CTO/CPO  
**Distribution:** Engineering Team (via Cursor)
