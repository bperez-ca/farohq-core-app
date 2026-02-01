# FARO Cursor Implementation Tasks
## Complete Task Breakdown for AI-Assisted Development

---

## HOW TO USE THIS DOCUMENT

When working with Cursor, copy the relevant task section and paste it as context. Each task includes:

1. **Context** - What this task is about
2. **Files to Create/Modify** - Specific file paths
3. **Implementation Details** - Code specifications
4. **Acceptance Criteria** - How to verify completion
5. **Dependencies** - What must be done first

---

## MVP TASKS (Week 1-2)

### TASK MVP-001: Project Setup

**Context:** Initialize the FARO monorepo with Next.js frontend and Go backend.

**Cursor Prompt:**
```
Create a monorepo for FARO with:
- apps/web: Next.js 14 with App Router, TypeScript, Tailwind
- apps/api: Go with Gin framework
- Use Turborepo for build orchestration

Include:
- ESLint and Prettier config
- TypeScript strict mode
- Environment variable templates
- Docker Compose for local development
```

**Expected Output:**
```
faro/
├── apps/
│   ├── web/                    # Next.js frontend
│   │   ├── src/
│   │   │   ├── app/           # App router pages
│   │   │   ├── components/    # React components
│   │   │   ├── hooks/         # Custom hooks
│   │   │   ├── lib/           # Utilities
│   │   │   └── types/         # TypeScript types
│   │   ├── package.json
│   │   └── tsconfig.json
│   │
│   └── api/                    # Go backend
│       ├── cmd/api/           # Entry point
│       ├── internal/
│       │   ├── api/           # HTTP handlers
│       │   ├── config/        # Configuration
│       │   ├── db/            # Database
│       │   ├── middleware/    # Middleware
│       │   ├── models/        # Data models
│       │   ├── services/      # Business logic
│       │   └── websocket/     # WebSocket hub
│       ├── go.mod
│       └── go.sum
│
├── packages/
│   └── shared/                 # Shared types/utils
│
├── docker-compose.yml
├── turbo.json
└── package.json
```

---

### TASK MVP-002: Database Schema

**Context:** Create the complete PostgreSQL schema for MVP features.

**Cursor Prompt:**
```
Create SQL migrations for FARO MVP with these tables:
- users (id, email, name, avatar_url, google_id)
- workspaces (id, name, slug, owner_id, settings, plan)
- workspace_members (workspace_id, user_id, role)
- workspace_invites (id, workspace_id, email, role, token, status, expires_at)
- channels (id, workspace_id, type, name, config, status)
- contacts (id, workspace_id, phone, email, name, tags, notes, message_count)
- conversations (id, workspace_id, channel_id, contact_id, status, assigned_to, last_message_at)
- messages (id, conversation_id, workspace_id, direction, channel, content, status)
- conversation_notes (id, conversation_id, author_id, content)
- notifications (id, user_id, type, title, body, read)
- audit_logs (id, workspace_id, user_id, action, resource_type, resource_id)

Include:
- All foreign keys and constraints
- Indexes for common queries
- Row-Level Security policies
- Triggers for updated_at and message counts

Use Supabase/PostgreSQL syntax.
```

**File:** `migrations/001_initial_schema.sql`

---

### TASK MVP-003: Go Backend Structure

**Context:** Set up the Go backend with proper project structure.

**Cursor Prompt:**
```
Create Go backend structure for FARO:

1. cmd/api/main.go - Entry point with:
   - Config loading from environment
   - Database connection pool
   - Router setup with Gin
   - Graceful shutdown

2. internal/config/config.go - Configuration struct

3. internal/db/db.go - Database connection with pgx

4. internal/middleware/
   - auth.go - JWT validation middleware
   - cors.go - CORS configuration
   - logger.go - Request logging
   - rls.go - Set RLS context (workspace_id, user_id)

5. internal/models/ - All data structs with JSON tags

6. internal/services/ - Business logic layer

7. internal/api/ - HTTP handlers

Follow these patterns:
- Dependency injection via constructors
- Context propagation for cancellation
- Structured logging with zap
- Error wrapping with stack traces
```

---

### TASK MVP-004: Authentication Flow

**Context:** Implement Google OAuth authentication with JWT tokens.

**Cursor Prompt:**
```
Implement authentication for FARO:

BACKEND (Go):
1. POST /api/auth/google
   - Accept OAuth code from frontend
   - Exchange for Google tokens
   - Get user info from Google
   - Find or create user in database
   - If new user, create workspace automatically
   - Generate JWT with user_id and workspace_id
   - Return token + user + workspace

2. POST /api/auth/refresh
   - Accept refresh token
   - Validate and generate new access token

3. Middleware to validate JWT on protected routes

FRONTEND (Next.js):
1. Login page with "Sign in with Google" button
2. OAuth callback handler
3. Token storage in httpOnly cookie
4. Auth context provider
5. Protected route wrapper

JWT payload:
{
  "sub": "user_id",
  "workspace_id": "workspace_id",
  "role": "owner|admin|member|viewer",
  "exp": timestamp
}
```

**Files:**
- `apps/api/internal/api/auth.go`
- `apps/api/internal/middleware/auth.go`
- `apps/api/internal/services/auth_service.go`
- `apps/web/src/app/auth/page.tsx`
- `apps/web/src/app/auth/callback/page.tsx`
- `apps/web/src/hooks/useAuth.ts`
- `apps/web/src/components/providers/AuthProvider.tsx`

---

### TASK MVP-005: WhatsApp Integration

**Context:** Integrate Twilio WhatsApp Business API for sending and receiving messages.

**Cursor Prompt:**
```
Implement WhatsApp messaging for FARO using Twilio:

WEBHOOK RECEIVER:
1. POST /api/webhooks/whatsapp
   - Validate Twilio signature
   - Parse incoming message payload
   - Find or create contact by phone
   - Find or create conversation
   - Create message record
   - Broadcast via WebSocket to frontend
   - Return 200 OK

2. POST /api/webhooks/twilio/status
   - Handle delivery status updates
   - Update message status (sent → delivered → read)
   - Broadcast status update via WebSocket

MESSAGE SENDING:
1. POST /api/conversations/:id/messages
   - Validate user can send to this conversation
   - Create message record with "pending" status
   - Send to Twilio API asynchronously
   - Update status on Twilio callback
   - Return message immediately (optimistic)

TWILIO SERVICE:
- SendWhatsApp(to, body, mediaUrl?) → (sid, error)
- ValidateSignature(signature, body) → bool

Include retry logic for failed sends.
```

**Files:**
- `apps/api/internal/api/webhooks.go`
- `apps/api/internal/api/messages.go`
- `apps/api/internal/services/twilio_service.go`
- `apps/api/internal/services/message_service.go`

---

### TASK MVP-006: WebSocket Real-Time Updates

**Context:** Implement WebSocket server for real-time message delivery.

**Cursor Prompt:**
```
Implement WebSocket server for FARO:

SERVER (Go):
1. WebSocket Hub
   - Track connections by workspace_id
   - Broadcast events to workspace members
   - Handle connection/disconnection
   - Ping/pong for keepalive

2. Connection Handler
   - Authenticate via JWT in query param
   - Register to Hub with workspace_id
   - Handle incoming messages (future: typing indicators)

3. Event Types:
   - message.new - New message received
   - message.status - Message status update
   - conversation.updated - Conversation changed
   - notification.new - New notification

CLIENT (Next.js):
1. useWebSocket hook
   - Connect on mount with JWT
   - Reconnect on disconnect
   - Event subscription pattern
   - Connection state tracking

2. Integration with TanStack Query
   - Invalidate queries on relevant events
   - Optimistic updates with rollback
```

**Files:**
- `apps/api/internal/websocket/hub.go`
- `apps/api/internal/websocket/client.go`
- `apps/api/internal/websocket/events.go`
- `apps/web/src/hooks/useWebSocket.ts`
- `apps/web/src/lib/websocket.ts`

---

### TASK MVP-007: Inbox UI

**Context:** Build the main inbox interface with conversation list and message thread.

**Cursor Prompt:**
```
Build the inbox UI for FARO:

LAYOUT (3-column on desktop):
1. Left: Conversation List (320px fixed)
   - Search input
   - Filter tabs (All, Unread, Mine)
   - Virtualized list of ConversationCard
   - Click to select conversation

2. Center: Message Thread (flexible)
   - ThreadHeader with contact info + actions
   - Message list (scrollable, auto-scroll to bottom)
   - MessageBubble component (inbound vs outbound styling)
   - MessageInput with send button

3. Right: Contact Panel (280px, hidden on mobile)
   - Contact avatar and name
   - Contact details (phone, email)
   - Tags (editable)
   - Notes section
   - Quick actions

MOBILE:
- Single column, navigate between views
- Bottom navigation

COMPONENTS:
- ConversationList.tsx
- ConversationCard.tsx
- ConversationThread.tsx
- MessageBubble.tsx
- MessageInput.tsx
- ThreadHeader.tsx
- ContactPanel.tsx

Use shadcn/ui components (Avatar, Badge, Button, Input, Textarea).
Use TanStack Query for data fetching.
Use WebSocket for real-time updates.
```

**Files:**
- `apps/web/src/app/(dashboard)/inbox/page.tsx`
- `apps/web/src/components/inbox/*.tsx`

---

### TASK MVP-008: Conversations API

**Context:** Build REST API for conversation management.

**Cursor Prompt:**
```
Build Conversations API for FARO:

ENDPOINTS:

GET /api/conversations
  Query: status, assigned_to, search, limit, offset
  Response: { conversations: [], total: number }
  - Join with contact for name/avatar
  - Include last_message preview
  - Sort by last_message_at DESC

GET /api/conversations/:id
  Response: { conversation, contact, channel }
  - Full conversation details
  - Contact profile
  - Channel info

PATCH /api/conversations/:id
  Body: { status?, assigned_to? }
  Response: { conversation }
  - Update conversation
  - Log assignment change
  - Send notification if assigned

GET /api/conversations/:id/messages
  Query: before, limit
  Response: { messages: [], has_more: boolean }
  - Paginated, newest first
  - Include sender info for outbound

POST /api/conversations/:id/messages
  Body: { content, media_url?, media_type? }
  Response: { message }
  - Create message with pending status
  - Trigger async send to channel
  - Broadcast via WebSocket

All endpoints must:
- Validate JWT
- Set RLS context (workspace_id)
- Return proper error responses
```

**Files:**
- `apps/api/internal/api/conversations.go`
- `apps/api/internal/services/conversation_service.go`

---

### TASK MVP-009: Contacts Management

**Context:** Build contact directory and profile management.

**Cursor Prompt:**
```
Build Contacts feature for FARO:

BACKEND:

GET /api/contacts
  Query: search, status, tags, limit, offset
  Response: { contacts: [], total }
  - Full-text search on name, email, phone
  - Filter by status and tags
  - Include message_count, last_message_at

GET /api/contacts/:id
  Response: { contact, conversations: [] }
  - Full contact profile
  - Recent conversations (last 5)

PATCH /api/contacts/:id
  Body: { name?, email?, tags?, notes?, status? }
  Response: { contact }
  - Update contact fields
  - Log changes

POST /api/contacts/:id/block
  Response: 204
  - Set status to "blocked"
  - Close all conversations

FRONTEND:

ContactsPage:
- Search input with debounce
- Filter dropdowns (status, tags)
- Virtualized contact list
- Click to open ContactProfile sheet

ContactProfile:
- Avatar, name, phone, email
- Editable tags (chip input)
- Notes (rich text editor)
- Conversation history list
- Block/unblock action
```

**Files:**
- `apps/api/internal/api/contacts.go`
- `apps/api/internal/services/contact_service.go`
- `apps/web/src/app/(dashboard)/contacts/page.tsx`
- `apps/web/src/components/contacts/*.tsx`

---

### TASK MVP-010: Team Collaboration

**Context:** Implement conversation assignment and team notes.

**Cursor Prompt:**
```
Build Team Collaboration for FARO:

ASSIGNMENT:

PATCH /api/conversations/:id/assign
  Body: { assigned_to: uuid | null }
  Response: { conversation }
  - Update assignment
  - Create notification for assignee
  - Broadcast update

Frontend:
- AssigneeDropdown in ThreadHeader
- Filter conversations by "Assigned to me"
- Assignment notification badge

TEAM NOTES:

GET /api/conversations/:id/notes
  Response: { notes: [] }
  - List notes for conversation
  - Include author info

POST /api/conversations/:id/notes
  Body: { content, mentioned_user_ids? }
  Response: { note }
  - Create private note
  - Create notifications for mentioned users

DELETE /api/notes/:id
  Response: 204
  - Only author can delete

Frontend:
- NotesSection in ContactPanel
- Note composer with @mention autocomplete
- Note list with author/timestamp
- Delete button (if author)

NOTIFICATIONS:

GET /api/notifications
  Query: unread_only, limit
  Response: { notifications: [], unread_count }

POST /api/notifications/:id/read
  Response: 204

POST /api/notifications/read-all
  Response: 204

Frontend:
- NotificationBell in header
- Dropdown with notification list
- Click to navigate to resource
- Mark as read on click
```

**Files:**
- `apps/api/internal/api/assignments.go`
- `apps/api/internal/api/notes.go`
- `apps/api/internal/api/notifications.go`
- `apps/web/src/components/inbox/AssigneeDropdown.tsx`
- `apps/web/src/components/inbox/NotesSection.tsx`
- `apps/web/src/components/shared/NotificationBell.tsx`

---

### TASK MVP-011: Team Management

**Context:** Build workspace team invite and management.

**Cursor Prompt:**
```
Build Team Management for FARO:

INVITES:

POST /api/workspaces/:id/invites
  Body: { email, role }
  Response: { invite }
  - Only owner/admin can invite
  - Generate secure token
  - Send email with invite link
  - Expires in 7 days

POST /api/invites/:token/accept
  Response: { workspace }
  - Validate token not expired
  - Add user to workspace
  - Mark invite as accepted

TEAM CRUD:

GET /api/workspaces/:id/team
  Response: { members: [] }
  - List all members with role
  - Include pending invites

PATCH /api/workspaces/:id/team/:member_id
  Body: { role }
  Response: { member }
  - Only owner can change roles
  - Cannot change own role

DELETE /api/workspaces/:id/team/:member_id
  Response: 204
  - Only owner can remove
  - Cannot remove self

Frontend:
- TeamPage in settings
- Member list with role dropdown
- Invite modal with email + role
- Remove confirmation dialog
- Pending invites section
```

**Files:**
- `apps/api/internal/api/team.go`
- `apps/api/internal/api/invites.go`
- `apps/api/internal/services/invite_service.go`
- `apps/web/src/app/(dashboard)/settings/team/page.tsx`
- `apps/web/src/components/settings/InviteModal.tsx`
- `apps/web/src/components/settings/TeamMemberCard.tsx`

---

### TASK MVP-012: Security & Audit

**Context:** Implement audit logging and security baseline.

**Cursor Prompt:**
```
Build Security features for FARO:

AUDIT LOGGING:

Middleware to log all mutations:
- Who (user_id)
- What (action, resource_type, resource_id)
- When (timestamp)
- Where (IP address, user agent)
- Result (success/failure)

Actions to log:
- auth.login, auth.logout
- message.send
- conversation.assign
- contact.update, contact.block
- team.invite, team.remove
- workspace.update

GET /api/workspaces/:id/audit-logs
  Query: user_id, action, from_date, to_date, limit, offset
  Response: { logs: [], total }
  - Only owner/admin can view
  - Mask IP addresses (hide last octet)

GET /api/workspaces/:id/audit-logs/export
  Response: CSV file

Frontend:
- AuditLogsPage in security settings
- Filterable log table
- Export button

DATA ENCRYPTION:
- Encrypt sensitive config (API keys) with KMS
- TLS for all connections
- RLS for data isolation
```

**Files:**
- `apps/api/internal/middleware/audit.go`
- `apps/api/internal/api/audit.go`
- `apps/web/src/app/(dashboard)/settings/security/page.tsx`

---

### TASK MVP-013: Monitoring & Health

**Context:** Set up monitoring, metrics, and alerting.

**Cursor Prompt:**
```
Build Monitoring for FARO:

HEALTH CHECK:

GET /health
  Response: {
    status: "healthy" | "degraded",
    checks: {
      database: "ok" | "error",
      redis: "ok" | "error",
      websocket: "ok" | "error"
    },
    timestamp: ISO date
  }

METRICS (Datadog/Prometheus):

Track:
- api.request.count (by endpoint, status)
- api.request.duration (p50, p95, p99)
- message.count (by channel, direction)
- websocket.connections (by workspace)
- error.count (by type)

ALERTING:

Slack alerts for:
- Error rate > 0.5%
- API latency p95 > 1000ms
- Webhook delivery < 95%
- Database connections > 90%

Implementation:
- Datadog agent integration
- Sentry for error tracking
- Slack webhook for alerts
```

**Files:**
- `apps/api/internal/api/health.go`
- `apps/api/internal/monitoring/metrics.go`
- `apps/api/internal/monitoring/alerts.go`

---

## PHASE 2 TASKS (Apr - Jun 2026)

### TASK P2-001: Instagram Integration

**Context:** Add Instagram DM support via Meta Graph API.

**Cursor Prompt:**
```
Implement Instagram DM integration:

1. OAuth flow for Instagram Business accounts
2. Webhook receiver for incoming DMs
3. Send DM via Graph API
4. Media support (images, videos)
5. Story replies
6. Channel-specific message formatting

Follow same patterns as WhatsApp integration.
Store Instagram page ID and access token in channel config.
```

---

### TASK P2-002: Facebook Messenger Integration

**Context:** Add Facebook Messenger support via Meta Graph API.

**Cursor Prompt:**
```
Implement Facebook Messenger integration:

1. Page subscription via webhook
2. Receive messages (text, attachments, quick_replies)
3. Send messages with buttons and templates
4. Sender actions (typing indicators)
5. Handover protocol for bot → human

Reuse Meta API client from Instagram integration.
```

---

### TASK P2-003: Email Integration

**Context:** Add email channel support via SendGrid.

**Cursor Prompt:**
```
Implement Email integration:

1. Configure workspace email address
2. Inbound webhook from SendGrid
3. Parse email content and attachments
4. Send replies via SendGrid API
5. Thread matching by subject/references
6. HTML to text conversion for inbox preview
```

---

### TASK P2-004: AI Message Suggestions

**Context:** Implement AI-powered reply suggestions.

**Cursor Prompt:**
```
Implement AI suggestions using OpenAI:

1. Generate 3 reply suggestions based on:
   - Conversation history (last 10 messages)
   - Contact profile (name, tags)
   - Business context (workspace settings)

2. API endpoint:
   POST /api/conversations/:id/suggestions
   Response: { suggestions: [string, string, string] }

3. Frontend:
   - Suggestion chips above input
   - Click to insert into input
   - Keyboard shortcuts (Cmd+1, Cmd+2, Cmd+3)
   - Track usage for analytics

4. Rate limiting and caching
5. No PII sent to OpenAI (anonymize)
```

---

### TASK P2-005: Billing Integration

**Context:** Implement Stripe billing for Pro tier.

**Cursor Prompt:**
```
Implement Stripe billing:

1. Subscription management:
   - Create checkout session
   - Handle webhook events
   - Manage subscription lifecycle

2. Usage tracking:
   - Message count per workspace
   - Track against plan limits
   - Overage alerts

3. Billing portal:
   - View current plan
   - Upgrade/downgrade
   - View invoices
   - Update payment method

4. Feature gating based on plan
```

---

## TESTING TASKS

### TASK TEST-001: Unit Tests

**Cursor Prompt:**
```
Write unit tests for FARO:

BACKEND (Go):
- Service layer tests with mocked DB
- Handler tests with mocked services
- Middleware tests
- 80% coverage target

FRONTEND (Jest + Testing Library):
- Component tests
- Hook tests
- Utility function tests
- 80% coverage target

Use table-driven tests in Go.
Use describe/it pattern in Jest.
```

---

### TASK TEST-002: Integration Tests

**Cursor Prompt:**
```
Write integration tests for FARO:

1. API endpoint tests:
   - Real database (test container)
   - Full request/response cycle
   - RLS policy verification

2. WebSocket tests:
   - Connection lifecycle
   - Event broadcasting
   - Multi-client scenarios

3. Webhook tests:
   - Twilio signature validation
   - Message processing pipeline
```

---

### TASK TEST-003: E2E Tests

**Cursor Prompt:**
```
Write E2E tests with Playwright:

1. Auth flow:
   - Google OAuth (mocked)
   - Workspace creation
   - Team invite accept

2. Messaging flow:
   - Receive message
   - Send reply
   - Status updates

3. Settings flow:
   - Team management
   - Channel configuration
```

---

## DEPLOYMENT TASKS

### TASK DEPLOY-001: CI/CD Pipeline

**Cursor Prompt:**
```
Create GitHub Actions workflow:

1. On PR:
   - Run linters
   - Run tests
   - Build check

2. On merge to main:
   - Run tests
   - Build images
   - Deploy to staging
   - Run smoke tests
   - Deploy to production

3. Rollback on failure
```

---

### TASK DEPLOY-002: Infrastructure

**Cursor Prompt:**
```
Set up production infrastructure:

1. Vercel project for frontend
   - Environment variables
   - Custom domain
   - Preview deployments

2. Railway/Render for backend
   - Dockerfile
   - Health checks
   - Auto-scaling

3. Supabase project
   - Production instance
   - Connection pooling
   - Backups

4. Upstash Redis
   - Caching
   - Rate limiting

5. Cloudflare
   - DNS
   - SSL
   - WAF
```

---

## QUICK REFERENCE: CURSOR COMMANDS

### Generate Component
```
Create a React component for [COMPONENT_NAME] that:
- Uses shadcn/ui components
- Is fully typed with TypeScript
- Includes loading and error states
- Is responsive (mobile-first)
- Follows FARO design patterns
```

### Generate API Endpoint
```
Create a Go API endpoint for [ENDPOINT_PATH] that:
- Validates JWT authentication
- Sets RLS context
- Follows REST conventions
- Returns proper error responses
- Logs to audit trail
```

### Generate Test
```
Write tests for [FILE_PATH] that:
- Cover happy path and error cases
- Use mocks for external dependencies
- Follow project testing patterns
- Achieve 80%+ coverage
```

### Debug Issue
```
Debug this issue in FARO:
- Error: [ERROR_MESSAGE]
- Expected: [EXPECTED_BEHAVIOR]
- Actual: [ACTUAL_BEHAVIOR]
- Relevant code: [CODE_SNIPPET]

Please identify the root cause and provide a fix.
```

---

## APPENDIX: FILE TEMPLATES

### Go Handler Template
```go
package api

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
)

type [Name]Handler struct {
    service *services.[Name]Service
}

func New[Name]Handler(service *services.[Name]Service) *[Name]Handler {
    return &[Name]Handler{service: service}
}

func (h *[Name]Handler) List(c *gin.Context) {
    workspaceID := c.GetString("workspace_id")
    
    items, err := h.service.List(c, workspaceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"items": items})
}
```

### React Component Template
```tsx
'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';

interface [Name]Props {
  id: string;
}

export function [Name]({ id }: [Name]Props) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['[name]', id],
    queryFn: () => fetch(`/api/[name]/${id}`).then(res => res.json()),
  });
  
  if (isLoading) return <[Name]Skeleton />;
  if (error) return <ErrorState error={error} />;
  
  return (
    <div>
      {/* Component content */}
    </div>
  );
}
```

---

**Document Version:** 1.0  
**Last Updated:** January 24, 2026  
**For:** Cursor IDE Implementation
