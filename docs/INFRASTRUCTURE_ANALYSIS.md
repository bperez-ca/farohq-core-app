# FARO Production Infrastructure Analysis

**Date**: 2025-01-27  
**Project**: FaroHQ Core App + Portal  
**Target**: AWS ECS (EC2-backed) + Amplify + RDS + ElastiCache  
**Current State**: Development/Staging (Docker Compose + Cloud Run docs)

---

## Executive Summary

**Current Deployment Status**: ⚠️ **Not Production-Ready**

The application is currently configured for:
- **Local Development**: Docker Compose with PostgreSQL, Dragonfly (Redis-compatible), Mailhog
- **Documented Target**: Google Cloud Run (serverless, not ECS)
- **Frontend**: AWS Amplify (already deployed, ~$40-50/mo)

**Key Findings**:
- ✅ Well-structured Dockerfile (multistage, distroless, <50MB image)
- ✅ Environment-based configuration (supports dev/staging/prod)
- ⚠️ **No production deployment** currently active
- ⚠️ **No CI/CD pipeline** configured
- ⚠️ **No orchestration** (K8s/ECS) configured
- ⚠️ **No monitoring/logging** infrastructure
- ⚠️ **No backup strategy** documented
- ✅ Secrets via environment variables (needs AWS Secrets Manager for prod)

**Migration Effort to AWS ECS**: **Medium-High** (2-3 weeks)
- Dockerfile is ECS-ready
- Need: ECS task definitions, ALB, CloudWatch, Secrets Manager integration
- Database migration from Cloud SQL to RDS
- Redis migration to ElastiCache

---

## 1. DEPLOYMENT TARGETS

### Current State

**Status**: ❌ **No Production Deployment Active**

#### Local Development
- **Platform**: Docker Compose
- **Services**:
  - PostgreSQL 15 (alpine) on port 5432
  - Dragonfly (Redis-compatible) on port 6379
  - Mailhog (email testing) on ports 1025/8025
  - Core App (Go) on port 8080

#### Documented Target (Not Active)
- **Platform**: Google Cloud Run (serverless)
- **Documentation**: `README.md` contains Cloud Run deployment instructions
- **Status**: Instructions exist but no active deployment confirmed

#### Frontend
- **Platform**: AWS Amplify ✅ (Active)
- **Cost**: ~$40-50/month
- **Status**: Deployed and working
- **Analysis**: See `VERCEL_ANALYSIS.md` for detailed assessment

### Dockerfile Analysis

```1:40:Dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o farohq-core-app cmd/server/main.go

# Final stage - use distroless for minimal image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy the binary from builder stage
COPY --from=builder /app/farohq-core-app .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose port (Cloud Run will set PORT env var)
EXPOSE 8080

# Run as non-root user (distroless provides nonroot user)
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["./farohq-core-app"]
```

**Assessment**:
- ✅ **Multistage build**: Reduces final image size
- ✅ **Distroless base**: Minimal attack surface (~20-30MB final image)
- ✅ **Non-root user**: Security best practice
- ✅ **Static binary**: No CGO dependencies (portable)
- ✅ **Migrations included**: Database migrations bundled

**Estimated Image Size**: **~25-35MB** (well under 500MB target)

### Environment-Specific Configurations

**Current Setup**:
- Environment variables via `internal/platform/config/config.go`
- Supports: `development`, `staging`, `production` (via `ENVIRONMENT` env var)
- Configuration loaded from environment variables (no config files)

**Environment Variables**:
```go
// Database
DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD, DB_SSLMODE
// Or single: DATABASE_URL

// Auth
CLERK_JWKS_URL

// AWS
AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, S3_BUCKET_NAME

// Redis
REDIS_URL

// Server
PORT (default: 8080)
ENVIRONMENT (development/staging/production)
```

**Missing**:
- ❌ No separate config files for dev/staging/prod
- ❌ No AWS Secrets Manager integration (secrets in env vars)
- ❌ No parameter store integration

### Secrets Management

**Current**: Environment variables only
- Secrets stored as plain env vars
- No rotation mechanism
- No centralized secrets store

**Production Requirements**:
- ✅ **AWS Secrets Manager** for database credentials
- ✅ **AWS Secrets Manager** for API keys (Data4SEO, Meta, OpenAI)
- ✅ **IAM roles** for ECS tasks (no hardcoded AWS keys)
- ⚠️ **Rotation policy**: Not implemented

**Risk**: Secrets could leak via:
- Logs (mitigated: structured logging, no PII)
- Environment variable exposure
- Container image layers (mitigated: distroless base)

---

## 2. CONTAINERIZATION

### Base Image

**Current**: `gcr.io/distroless/static-debian12:nonroot`
- **Size**: ~20MB base
- **Security**: Minimal attack surface (no shell, no package manager)
- **Compatibility**: Works on ECS (Debian-based)

**Alternative Considered**: `alpine` (smaller but distroless is more secure)

### Image Size Analysis

**Estimated Final Image**: **~25-35MB**
- Base: ~20MB (distroless)
- Binary: ~10-15MB (Go static binary)
- Migrations: <1MB (SQL files)

**Target**: <500MB ✅ **Well under target**

### Layer Optimization

**Current Dockerfile**:
1. Builder stage: golang:1.23-alpine (~300MB, discarded)
2. Final stage: distroless (~20MB)
3. Binary copy: +10-15MB
4. Migrations copy: +<1MB

**Optimization Opportunities**:
- ✅ Already optimized (minimal layers)
- ✅ Dependencies cached separately (go.mod/go.sum copied first)
- ✅ No unnecessary files (`.dockerignore` present)

### .dockerignore

```1:40:.dockerignore
# Git
.git
.gitignore

# Documentation
*.md
docs/

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# Build artifacts
bin/
dist/
*.exe
*.test

# Dependencies
vendor/

# Environment files
.env
.env.local
.env.*.local

# Test files
*_test.go
coverage.out

# CI/CD
.github/

# OS
.DS_Store
Thumbs.db
```

**Assessment**: ✅ **Comprehensive** - Excludes dev files, tests, docs

---

## 3. ORCHESTRATION

### Current State

**Status**: ❌ **No Orchestration Configured**

- No Kubernetes manifests
- No ECS task definitions
- No Docker Swarm
- Only Docker Compose for local development

### Docker Compose (Local Only)

**Services**:
- `postgres`: PostgreSQL 15-alpine
- `dragonfly`: Redis-compatible cache
- `mailhog`: Email testing
- Core app runs via `make dev` (not in compose)

**Purpose**: Local development only, not production

### Replicas/Instances

**Current**: 1 instance (local dev)
**Production Target**: 
- ECS: 2-3 tasks (for high availability)
- Auto-scaling: 2-10 tasks based on CPU/memory

### Auto-Scaling

**Status**: ❌ **Not Configured**

**ECS Auto-Scaling Requirements**:
- Target: CPU 70%, Memory 80%
- Min: 2 tasks
- Max: 10 tasks
- Scale-up: +1 task per 30s
- Scale-down: -1 task per 5min (cooldown)

### Rolling Updates

**Status**: ❌ **Not Configured**

**ECS Rolling Update Strategy**:
- **Type**: Rolling update with health checks
- **Minimum healthy**: 50% (1 task during update)
- **Maximum**: 200% (2 tasks during update)
- **Health check**: `/healthz` and `/readyz` endpoints exist ✅

**Zero-Downtime Deploy**:
1. Deploy new task definition
2. ECS starts new tasks (health check passes)
3. ALB routes traffic to new tasks
4. Old tasks drained (30s grace period)
5. Old tasks terminated

---

## 4. CI/CD PIPELINE

### Current State

**Status**: ❌ **No CI/CD Pipeline Configured**

- No GitHub Actions workflows found
- No GitLab CI
- No CircleCI
- No deployment automation

### Required Pipeline

**Recommended**: GitHub Actions (if using GitHub)

**Pipeline Stages**:
1. **Lint & Test** (2-3 min)
   - `make lint`
   - `make test`
   - Go vet, fmt check

2. **Build Docker Image** (3-5 min)
   - Build multi-stage Dockerfile
   - Push to ECR (Elastic Container Registry)

3. **Security Scan** (2-3 min)
   - Trivy/Clair scan for vulnerabilities
   - Check for secrets in image

4. **Deploy to Staging** (5-10 min)
   - Update ECS service (staging)
   - Run migrations
   - Smoke tests

5. **Deploy to Production** (manual approval)
   - Update ECS service (production)
   - Run migrations
   - Health check verification

**Total Time**: ~15-25 minutes (from push to staging)

### Rollback Strategy

**Current**: ❌ **Not Configured**

**ECS Rollback**:
1. Revert to previous task definition (ECS keeps last N definitions)
2. Update service to use previous revision
3. ECS performs rolling update (zero downtime)
4. Verify health checks pass

**Database Rollback**:
- Migrations: `make migrate-down` (manual)
- No automated rollback (risky - requires testing)

---

## 5. NETWORKING

### Current State

**Local Development**:
- Direct port exposure: `8080:8080`
- No load balancer
- No HTTPS (HTTP only)

**Production Target**: AWS ECS with ALB

### Backend Exposure

**ECS Architecture**:
```
Internet
  ↓
Route53 DNS
  ↓
Application Load Balancer (ALB)
  ↓ (HTTPS, port 443)
Target Group (ECS tasks on port 8080)
  ↓
ECS Service (2-10 tasks)
```

**ALB Configuration**:
- **Listener**: HTTPS (443) → HTTP (8080) to tasks
- **SSL Certificate**: ACM (AWS Certificate Manager)
- **Health Check**: `/healthz` (every 30s)
- **Sticky Sessions**: Not required (stateless app)

### HTTPS Enforcement

**Status**: ⚠️ **Not Configured** (will be via ALB)

**Requirements**:
- ALB terminates SSL/TLS
- ACM certificate (free, auto-renewal)
- HTTP → HTTPS redirect (ALB rule)
- HSTS headers (via ALB response headers)

### DNS Setup

**Current**: Not configured for production

**Custom Agency Domains**:
- **Pattern**: `{agency-slug}.portal.farohq.com` or `portal.{agency-domain}.com`
- **Resolution**: 
  1. DNS → ALB (Route53 or external DNS)
  2. ALB → ECS (host header preserved)
  3. App resolves tenant from host header
  4. Brand theme fetched by domain

**Route53 Configuration**:
- **A Record**: `api.farohq.com` → ALB (alias)
- **Wildcard**: `*.portal.farohq.com` → ALB (alias)
- **CNAME**: Agency custom domains → `api.farohq.com`

### Network Security

**Status**: ❌ **Not Configured**

**Required**:
- **Security Groups**:
  - ALB: Inbound 443 (HTTPS) from 0.0.0.0/0
  - ECS Tasks: Inbound 8080 from ALB security group only
  - RDS: Inbound 5432 from ECS security group only
  - ElastiCache: Inbound 6379 from ECS security group only

- **VPC Configuration**:
  - Public subnets: ALB
  - Private subnets: ECS tasks, RDS, ElastiCache
  - NAT Gateway: For ECS tasks to reach internet (S3, external APIs)

---

## 6. DATABASES & PERSISTENCE

### Current Database

**Local Development**: PostgreSQL 15 (alpine) via Docker Compose

**Production Target**: PostgreSQL RDS

### Database Type

**Current**: PostgreSQL 15 ✅
- **Version**: 15-alpine (local)
- **Target**: RDS PostgreSQL 15.x
- **Compatibility**: ✅ Fully compatible

### Managed vs Self-Hosted

**Current**: Self-hosted (Docker Compose, local only)

**Production Target**: **RDS PostgreSQL** (managed)

**RDS Configuration**:
- **Instance Class**: `db.t3.medium` (2 vCPU, 4GB RAM) for Month 12
- **Multi-AZ**: Yes (high availability)
- **Storage**: 100GB gp3 (auto-scaling to 1TB)
- **Backup**: Automated daily backups, 7-day retention
- **Encryption**: At-rest (KMS) and in-transit (SSL)

### Backups

**Status**: ❌ **Not Configured**

**RDS Automated Backups**:
- **Frequency**: Daily (during maintenance window)
- **Retention**: 7 days (can extend to 35 days)
- **Point-in-Time Recovery**: Enabled (5-minute granularity)
- **Cross-Region**: Optional (for disaster recovery)

**Manual Backup Strategy**:
```bash
# Full database backup (for migrations)
pg_dump -h $RDS_HOST -U $DB_USER -d $DB_NAME -F c -f backup_$(date +%Y%m%d).dump

# Single tenant export (using RLS)
SET LOCAL lv.tenant_id = 'agency-uuid';
pg_dump --data-only --table=clients --table=locations ... > tenant_backup.sql
```

**Testing**: ❌ **Not documented** - Need backup restore testing procedure

### Read Replicas

**Status**: ❌ **Not Configured**

**Recommendation for Month 12**:
- **Primary**: `db.t3.medium` (Multi-AZ)
- **Read Replica**: `db.t3.medium` (single AZ, same region)
- **Purpose**: Offload read queries (reports, analytics)
- **Lag**: <1 second (synchronous replication)

**When to Add**:
- Read:Write ratio > 2:1
- Database CPU > 60% on primary
- Read query latency > 100ms

---

## 7. CACHING

### Current Cache

**Local Development**: Dragonfly (Redis-compatible) via Docker Compose

**Production Target**: **Redis ElastiCache**

### Redis Deployment

**Current**: Dragonfly (local, port 6379)

**ElastiCache Configuration**:
- **Engine**: Redis 7.x
- **Node Type**: `cache.t3.micro` (Month 1-3) → `cache.t3.small` (Month 12)
- **Cluster Mode**: Disabled (single node, simple setup)
- **Multi-AZ**: Yes (automatic failover)
- **Encryption**: At-rest and in-transit

### Clustering

**Status**: ❌ **Not Configured** (single node)

**Cluster Mode** (Future):
- **When**: >100GB cache size, >100K ops/sec
- **Configuration**: 3 shards, 2 replicas each (6 nodes total)
- **Cost**: ~$300-500/month

### Persistence

**Current**: Not configured (in-memory only)

**ElastiCache Persistence**:
- **RDB Snapshots**: Every 6 hours (default)
- **AOF (Append-Only File)**: Disabled (for performance)
- **Backup Retention**: 1 day (can extend to 35 days)

**Recommendation**: Enable RDB snapshots for cache warm-up after restart

### Memory & Eviction

**Current**: Not configured

**ElastiCache Configuration**:
- **Memory Limit**: 1.3GB (cache.t3.small)
- **Eviction Policy**: `allkeys-lru` (Least Recently Used)
- **Max Memory Policy**: `allkeys-lru` (evict least used keys when full)

**Cache Usage**:
- **Tenant Resolution**: 5-minute TTL
- **Brand Themes**: 15-minute TTL (configurable)
- **User Roles**: 5-minute TTL

**Estimated Cache Size (Month 12)**:
- 30 agencies × 10KB (brand theme) = 300KB
- 1,000 locations × 1KB (metadata) = 1MB
- **Total**: ~2-5MB (well under 1.3GB limit)

---

## 8. MONITORING & LOGGING

### Current Monitoring

**Status**: ⚠️ **Basic Logging Only**

**Logging**:
- **Library**: `zerolog` (structured JSON logging)
- **Output**: `stdout` (container logs)
- **Fields**: `service`, `timestamp`, custom fields
- **Destination**: Not configured (logs go to container stdout)

**Monitoring**: ❌ **Not Configured**

### Monitoring Tools

**Production Target**: **AWS CloudWatch**

**Required Dashboards**:
1. **Application Metrics**:
   - Request rate (requests/sec)
   - Error rate (4xx, 5xx)
   - Latency (p50, p95, p99)
   - Active connections

2. **Infrastructure Metrics**:
   - CPU utilization (ECS tasks)
   - Memory utilization (ECS tasks)
   - Database connections (RDS)
   - Cache hit rate (ElastiCache)

3. **Business Metrics**:
   - Active tenants
   - API calls per tenant
   - Storage usage (S3)

### Logging Infrastructure

**Current**: stdout only

**CloudWatch Logs Configuration**:
- **Log Group**: `/ecs/farohq-core-app`
- **Retention**: 30 days (can extend to 1 year)
- **Streams**: One per ECS task
- **Filtering**: By tenant_id, request_id, log level

**Log Format** (zerolog):
```json
{
  "level": "info",
  "service": "farohq-core-app",
  "time": "2025-01-27T10:00:00Z",
  "tenant_id": "agency-uuid",
  "request_id": "req-123",
  "message": "Request processed"
}
```

### Debugging Production Issues

**Current**: ❌ **No Tooling**

**Required**:
1. **CloudWatch Logs Insights**: Query logs by tenant, error type, time range
2. **X-Ray Tracing**: Distributed tracing (optional, for complex debugging)
3. **CloudWatch Alarms**: Alert on error rate > 1%, latency > 500ms
4. **SNS Notifications**: Email/Slack alerts on critical errors

**Example Query** (CloudWatch Logs Insights):
```
fields @timestamp, @message, tenant_id, request_id
| filter level = "error"
| sort @timestamp desc
| limit 100
```

---

## 9. SECRETS MANAGEMENT

### Current Secrets

**Status**: ⚠️ **Environment Variables Only**

**Secrets Stored**:
- Database credentials: `DB_PASSWORD`
- AWS keys: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
- Clerk: `CLERK_JWKS_URL` (public, but URL contains instance name)
- API keys: Not yet configured (Data4SEO, Meta, OpenAI)

### Secrets Storage

**Current**: Plain environment variables

**Production Target**: **AWS Secrets Manager**

**Secrets to Migrate**:
1. **Database Password**: `farohq-db-password`
2. **AWS Credentials**: Use IAM roles instead (no secrets needed)
3. **Clerk JWKS URL**: Can remain as env var (not sensitive)
4. **Data4SEO API Key**: `farohq-data4seo-key`
5. **Meta Cloud API Token**: `farohq-meta-token`
6. **OpenAI API Key**: `farohq-openai-key`
7. **Postmark API Token**: `farohq-postmark-token`

### Rotation Policy

**Status**: ❌ **Not Implemented**

**AWS Secrets Manager Rotation**:
- **Database Password**: Rotate every 90 days (automatic)
- **API Keys**: Manual rotation (update secret, restart ECS tasks)
- **Rotation Lambda**: Required for database (updates RDS password)

### Secrets Leak Prevention

**Current Risks**:
- ✅ **Logs**: Structured logging, no secrets logged
- ⚠️ **Environment Variables**: Exposed to ECS task (mitigated by Secrets Manager)
- ✅ **Container Image**: Distroless base, no secrets in layers
- ⚠️ **Debug Endpoints**: No `/debug` endpoints (good)

**Best Practices**:
- ✅ Never log secrets (zerolog doesn't log env vars)
- ✅ Use IAM roles (no AWS keys in secrets)
- ✅ Secrets Manager (encrypted at rest, in transit)
- ✅ Least privilege (ECS task role only has required permissions)

---

## 10. SCALING ASSUMPTIONS

### Current Load

**Status**: ⚠️ **Unknown** (no production deployment)

**Assumptions** (Development):
- **Users**: <10 developers
- **Requests/sec**: <1 (local testing)
- **Data Size**: <100MB (test data)

### Projected Load (Month 12)

**Target Metrics**:
- **Agencies**: 30
- **Locations**: 1,000
- **Users**: ~150 (5 users/agency average)
- **Requests/sec**: ~10-20 (estimated)
- **Data Size**: ~5-10GB (locations, reviews, posts)

**Breakdown**:
- **Agencies**: 30 × 50KB = 1.5MB
- **Clients**: 300 × 20KB = 6MB
- **Locations**: 1,000 × 10KB = 10MB
- **Reviews**: 10,000 × 2KB = 20MB
- **Posts**: 5,000 × 5KB = 25MB
- **Brand Themes**: 30 × 10KB = 300KB
- **Total**: ~60MB (well under RDS 100GB limit)

### 3x Traffic Capacity

**Peak Events**: 30-60 requests/sec (3x normal)

**Current Infrastructure (Month 12)**:
- **ECS**: 2-3 tasks (can scale to 10)
- **RDS**: db.t3.medium (can handle 100+ req/sec)
- **ElastiCache**: cache.t3.small (can handle 10K+ ops/sec)
- **ALB**: Can handle 1,000+ req/sec

**Bottleneck Analysis**:
1. **Database**: ✅ Not a bottleneck (db.t3.medium handles 100+ req/sec)
2. **Cache**: ✅ Not a bottleneck (Redis handles 10K+ ops/sec)
3. **Application**: ⚠️ **Potential bottleneck** (Go app, depends on query complexity)
4. **Network**: ✅ Not a bottleneck (ALB handles 1,000+ req/sec)

**Recommendation**: 
- Start with 2 ECS tasks
- Auto-scale to 5 tasks at 70% CPU
- Monitor p95 latency (target: <200ms)

### First Bottlenecks

**Expected Order** (if scaling beyond Month 12):
1. **Application CPU** (ECS tasks) - Scale horizontally
2. **Database Connections** (RDS) - Connection pooling, read replicas
3. **Database Write Throughput** (RDS) - Larger instance, read replicas
4. **Cache Memory** (ElastiCache) - Larger node, cluster mode

---

## ARCHITECTURE DIAGRAM

### Current Architecture (Local Development)

```
┌─────────────────────────────────────────┐
│         Developer Machine               │
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │  Go App      │  │  Next.js     │   │
│  │  :8080       │  │  Portal      │   │
│  └──────┬───────┘  └──────┬───────┘   │
│         │                  │           │
│  ┌──────▼──────────────────▼──────┐  │
│  │     Docker Compose              │  │
│  │  ┌──────────┐  ┌──────────┐   │  │
│  │  │PostgreSQL │  │ Dragonfly│   │  │
│  │  │  :5432    │  │  :6379   │   │  │
│  │  └──────────┘  └──────────┘   │  │
│  └─────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Target Architecture (AWS Production)

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                             │
└────────────────────┬──────────────────────────────────────┘
                     │
                     │ HTTPS (443)
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Route53 DNS                                     │
│  api.farohq.com, *.portal.farohq.com                         │
└────────────────────┬──────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│         Application Load Balancer (ALB)                      │
│         - HTTPS termination (ACM)                           │
│         - Health checks (/healthz)                           │
│         - SSL/TLS 1.2+                                      │
└────────────────────┬──────────────────────────────────────┘
                     │
                     │ HTTP (8080)
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              ECS Cluster (EC2-backed)                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  ECS Service: farohq-core-app                        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐            │  │
│  │  │ Task 1   │  │ Task 2   │  │ Task N   │            │  │
│  │  │ :8080    │  │ :8080    │  │ :8080    │            │  │
│  │  └──────────┘  └──────────┘  └──────────┘            │  │
│  │  Auto-scaling: 2-10 tasks                           │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬──────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   RDS       │ │ ElastiCache  │ │     S3      │
│ PostgreSQL  │ │    Redis     │ │   Files     │
│ db.t3.medium│ │ cache.t3.small│ │  Bucket    │
│ Multi-AZ    │ │ Multi-AZ     │ │             │
└─────────────┘ └─────────────┘ └─────────────┘
```

### Frontend Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Internet                                  │
└────────────────────┬──────────────────────────────────────┘
                     │
                     │ HTTPS
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Route53 DNS                                      │
│  portal.farohq.com, *.portal.farohq.com                     │
└────────────────────┬──────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│           AWS Amplify (CloudFront CDN)                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Next.js App (SSR/ISR)                                 │  │
│  │  - Clerk Auth                                          │  │
│  │  - Brand Theme Provider                                │  │
│  │  - API Routes (proxy to backend)                       │  │
│  └────────────────────┬───────────────────────────────────┘  │
└───────────────────────┼──────────────────────────────────────┘
                        │
                        │ API Calls (HTTPS)
                        ▼
┌─────────────────────────────────────────────────────────────┐
│         Application Load Balancer (Backend)                 │
│         (Same ALB as backend, different target group)       │
└─────────────────────────────────────────────────────────────┘
```

---

## DEPLOYMENT PROCESS WALKTHROUGH

### Current Process (Local)

```bash
# 1. Start infrastructure
make up                    # Docker Compose: postgres, dragonfly, mailhog

# 2. Wait for DB
make db-wait               # Health check: pg_isready

# 3. Run migrations
make migrate-up            # golang-migrate: apply migrations

# 4. Start application
make dev                   # go run cmd/server/main.go
```

**Time**: ~30 seconds (local)

### Target Process (Production)

#### Initial Setup (One-time)

```bash
# 1. Create ECR repository
aws ecr create-repository --repository-name farohq-core-app

# 2. Create RDS instance
aws rds create-db-instance \
  --db-instance-identifier farohq-prod \
  --db-instance-class db.t3.medium \
  --engine postgres \
  --master-username faro_admin \
  --master-user-password <from-secrets-manager> \
  --allocated-storage 100 \
  --multi-az

# 3. Create ElastiCache cluster
aws elasticache create-cache-cluster \
  --cache-cluster-id farohq-redis \
  --cache-node-type cache.t3.small \
  --engine redis \
  --num-cache-nodes 1

# 4. Create ECS cluster
aws ecs create-cluster --cluster-name farohq-prod

# 5. Create ALB
aws elbv2 create-load-balancer \
  --name farohq-alb \
  --subnets <public-subnet-ids> \
  --security-groups <alb-sg-id>
```

#### Deployment (CI/CD)

```bash
# 1. Build & Push Docker Image (GitHub Actions)
docker build -t farohq-core-app:latest .
docker tag farohq-core-app:latest $ECR_REGISTRY/farohq-core-app:$GIT_SHA
docker push $ECR_REGISTRY/farohq-core-app:$GIT_SHA

# 2. Update ECS Task Definition
aws ecs register-task-definition \
  --family farohq-core-app \
  --container-definitions file://task-definition.json

# 3. Update ECS Service
aws ecs update-service \
  --cluster farohq-prod \
  --service farohq-core-app \
  --task-definition farohq-core-app:$REVISION \
  --force-new-deployment

# 4. Run Migrations (separate ECS task)
aws ecs run-task \
  --cluster farohq-prod \
  --task-definition farohq-migrate \
  --launch-type FARGATE

# 5. Verify Health
curl https://api.farohq.com/healthz
```

**Time**: ~10-15 minutes (from push to live)

---

## CAPACITY ANALYSIS

### Current vs Month 12

| Metric | Current (Dev) | Month 12 (Target) | Capacity (Infra) |
|--------|---------------|-------------------|------------------|
| **Agencies** | 0 | 30 | Unlimited (app limit) |
| **Locations** | <10 | 1,000 | Unlimited (DB limit) |
| **Users** | <5 | 150 | Unlimited (app limit) |
| **Requests/sec** | <1 | 10-20 | 100+ (ECS scale) |
| **Data Size** | <100MB | 5-10GB | 100GB (RDS) |
| **Cache Size** | <1MB | 2-5MB | 1.3GB (ElastiCache) |

### Resource Utilization (Month 12)

| Resource | Utilization | Capacity | Headroom |
|----------|-------------|----------|----------|
| **ECS CPU** | ~20% (2 tasks) | 100% (10 tasks) | 5x |
| **ECS Memory** | ~30% (2 tasks) | 100% (10 tasks) | 3x |
| **RDS CPU** | ~15% | 100% | 6x |
| **RDS Storage** | ~10GB | 100GB | 10x |
| **RDS Connections** | ~20 | 100 | 5x |
| **ElastiCache Memory** | ~5MB | 1.3GB | 260x |
| **ElastiCache Ops/sec** | ~50 | 10,000+ | 200x |

**Conclusion**: ✅ **Infrastructure has 3-10x headroom** for Month 12 load

### Cost Estimate (Month 12)

| Service | Configuration | Monthly Cost |
|---------|---------------|--------------|
| **ECS (EC2)** | 2x t3.medium instances | $60 |
| **RDS** | db.t3.medium Multi-AZ | $150 |
| **ElastiCache** | cache.t3.small Multi-AZ | $30 |
| **ALB** | Standard ALB | $20 |
| **S3** | 10GB storage + requests | $5 |
| **CloudWatch** | Logs + Metrics | $10 |
| **Data Transfer** | 100GB out | $10 |
| **Amplify** | Frontend hosting | $50 |
| **Route53** | DNS hosting | $1 |
| **ACM** | SSL certificates | $0 (free) |
| **Total** | | **~$336/month** |

**Comparison**:
- Current (Cloud Run): ~$50-100/month (serverless, pay-per-use)
- Target (ECS): ~$336/month (fixed cost, predictable)

---

## OPTIMIZATION RECOMMENDATIONS

### 1. Implement Connection Pooling

**Current**: Default pgxpool settings (may be inefficient)

**Optimization**:
```go
// internal/platform/db/pool.go
pool, err := pgxpool.New(ctx, dsn, pgxpool.Config{
    MaxConns: 25,              // Per ECS task
    MinConns: 5,               // Keep warm connections
    MaxConnLifetime: 30 * time.Minute,
    MaxConnIdleTime: 5 * time.Minute,
})
```

**Impact**: 
- Reduces connection churn
- Improves query latency by 10-20%
- Prevents RDS connection exhaustion

**Effort**: Low (1-2 hours)

### 2. Add Read Replicas for Analytics

**Current**: Single RDS instance (all queries hit primary)

**Optimization**:
- Create read replica (db.t3.medium)
- Route read-only queries to replica (reports, analytics)
- Keep writes on primary

**Impact**:
- Reduces primary load by 30-50%
- Enables parallel analytics queries
- Cost: +$75/month

**Effort**: Medium (4-6 hours)

### 3. Implement Query Result Caching

**Current**: Only tenant/brand caching

**Optimization**:
- Cache frequently accessed data (locations list, client list)
- TTL: 5-15 minutes (configurable per endpoint)
- Invalidate on write (cache tags)

**Impact**:
- Reduces database load by 40-60%
- Improves response time by 50-100ms
- Reduces RDS costs (smaller instance possible)

**Effort**: Medium (1-2 days)

### 4. Enable CloudWatch Logs Insights Queries

**Current**: Logs to CloudWatch, no querying

**Optimization**:
- Pre-configured Insights queries:
  - Error rate by tenant
  - Slow queries (latency > 500ms)
  - API usage by endpoint
- Dashboard with key metrics

**Impact**:
- Faster debugging (5 min → 30 sec)
- Proactive issue detection
- Cost: ~$5/month (queries)

**Effort**: Low (2-3 hours)

### 5. Implement Database Query Monitoring

**Current**: No query performance tracking

**Optimization**:
- Enable RDS Performance Insights
- Track slow queries (>100ms)
- Identify N+1 query patterns

**Impact**:
- Identify bottlenecks before they cause issues
- Optimize queries proactively
- Cost: ~$20/month (Performance Insights)

**Effort**: Low (1 hour setup, ongoing optimization)

---

## RECOMMENDATION: KEEP CURRENT vs AWS ECS REFACTOR

### Option 1: Stay on Cloud Run (Current Target)

**Pros**:
- ✅ Already documented
- ✅ Serverless (no infrastructure management)
- ✅ Auto-scaling (0 to N instances)
- ✅ Lower cost at low traffic (~$50-100/month)
- ✅ Faster deployment (2-3 minutes)

**Cons**:
- ❌ Vendor lock-in (GCP)
- ❌ Cold starts (100-500ms first request)
- ❌ Less control (can't SSH, limited debugging)
- ❌ Higher cost at scale (>$500/month at high traffic)

**Effort**: Low (already documented, just deploy)

**Cost**: $50-100/month (Month 1-6), $200-500/month (Month 12+)

### Option 2: Migrate to AWS ECS (Recommended for Production)

**Pros**:
- ✅ Full control (SSH, debugging, custom configs)
- ✅ Predictable costs (fixed EC2 instances)
- ✅ Better for long-running processes
- ✅ AWS ecosystem integration (RDS, ElastiCache, Secrets Manager)
- ✅ No cold starts (always warm)

**Cons**:
- ❌ Higher fixed cost (~$336/month)
- ❌ Infrastructure management (ECS, ALB, security groups)
- ❌ Slower deployment (10-15 minutes)
- ❌ More complex setup (requires Terraform/CloudFormation)

**Effort**: Medium-High (2-3 weeks)
- Week 1: Infrastructure setup (ECS, RDS, ElastiCache, ALB)
- Week 2: CI/CD pipeline, secrets management, monitoring
- Week 3: Testing, migration, cutover

**Cost**: ~$336/month (fixed, predictable)

### Option 3: Hybrid Approach

**Strategy**:
- **Staging**: Cloud Run (low cost, fast iteration)
- **Production**: AWS ECS (predictable, full control)

**Pros**:
- Best of both worlds
- Lower staging costs
- Production reliability

**Cons**:
- Two platforms to manage
- More complex CI/CD

**Effort**: Medium (same as Option 2, plus Cloud Run staging)

---

## FINAL RECOMMENDATION

### ✅ **Migrate to AWS ECS** (Option 2)

**Rationale**:
1. **Production Requirements**: ECS provides better control, debugging, and reliability
2. **Cost Predictability**: Fixed costs ($336/month) vs variable (Cloud Run can spike)
3. **AWS Ecosystem**: Deep integration with RDS, ElastiCache, Secrets Manager
4. **Scalability**: Better for long-term growth (can scale to 100+ tasks)
5. **Team Expertise**: If team is AWS-focused, ECS is better fit

**Migration Plan**:
1. **Phase 1** (Week 1): Infrastructure setup
   - ECS cluster, RDS, ElastiCache, ALB
   - Security groups, VPC configuration
   - Secrets Manager integration

2. **Phase 2** (Week 2): CI/CD & Monitoring
   - GitHub Actions pipeline
   - CloudWatch dashboards
   - Logging configuration

3. **Phase 3** (Week 3): Testing & Migration
   - Staging deployment
   - Load testing
   - Production cutover

**Risk**: Low-Medium (well-documented process, Dockerfile is ECS-ready)

**Cost Impact**: +$236-286/month (from Cloud Run to ECS), but predictable and scalable

---

## APPENDIX: Infrastructure Checklist

### Pre-Production Checklist

- [ ] ECS cluster created (EC2-backed)
- [ ] RDS PostgreSQL instance (Multi-AZ)
- [ ] ElastiCache Redis cluster (Multi-AZ)
- [ ] ALB configured (HTTPS, health checks)
- [ ] Security groups configured (least privilege)
- [ ] Secrets Manager integration (all secrets)
- [ ] CloudWatch dashboards (metrics, logs)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Database backups (automated, tested)
- [ ] Monitoring alerts (error rate, latency)
- [ ] Load testing (validate capacity)
- [ ] Disaster recovery plan (backup restore tested)

### Post-Deployment Checklist

- [ ] Health checks passing (`/healthz`, `/readyz`)
- [ ] SSL certificates valid (ACM)
- [ ] Database migrations applied
- [ ] Cache connectivity verified
- [ ] Logs flowing to CloudWatch
- [ ] Metrics visible in dashboards
- [ ] Auto-scaling configured (tested)
- [ ] Rollback procedure tested

---

**Report Generated**: 2025-01-27  
**Next Review**: After ECS migration (validate assumptions)
