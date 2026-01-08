# FaroHQ Core App

A multi-tenant, white-label SaaS backend built with Go, following Hexagonal Architecture (Ports & Adapters) with Vertical Slicing and Screaming Architecture.

## Architecture

This application is structured using:
- **Hexagonal Architecture**: Clear separation between business logic (domain), application services (app), and infrastructure (infra)
- **Vertical Slicing**: Each domain owns its complete slice (domain, app, infra)
- **Screaming Architecture**: Directory structure immediately conveys business domains

### Domain Structure

```
internal/domains/
├── tenants/     # Tenant (agency), client, location, and member management
├── brand/       # White-label branding (themes, logos, favicons)
├── files/       # File upload/download via S3 pre-signed URLs
└── auth/        # Authentication (Clerk JWT integration)
```

Each domain follows this structure:
```
<domain>/
├── domain/          # Pure business logic (entities, VOs, errors)
│   ├── model/       # Domain entities
│   ├── services/    # Domain services
│   └── ports/       # Inbound/outbound interfaces
├── app/             # Use cases (application services)
│   └── usecases/    # One file per use case
└── infra/           # Adapters (database, HTTP, S3, etc.)
    ├── db/          # PostgreSQL repositories
    ├── http/        # HTTP handlers and routers
    └── s3/          # S3 storage adapter (for files domain)
```

## Prerequisites

- Go 1.23+
- PostgreSQL 15+
- AWS S3 (or LocalStack for local development)
- Clerk account (for authentication)

## Environment Variables

Create a `.env` file in the root directory:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=localvisibilityos
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable
# Or use DATABASE_URL directly:
# DATABASE_URL=postgres://user:password@localhost:5432/localvisibilityos?sslmode=disable

# Clerk
CLERK_JWKS_URL=https://your-clerk-instance.clerk.accounts.dev/.well-known/jwks.json

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
S3_BUCKET_NAME=lvos-files

# Local development (optional - for LocalStack)
AWS_ENDPOINT_URL=http://localhost:4566

# Server
PORT=8080
WEB_URL=http://localhost:3000
```

## Local Development

### 1. Start Infrastructure

```bash
# Start PostgreSQL (using Docker Compose if available)
docker-compose up -d postgres

# Or use your local PostgreSQL instance
```

### 2. Run Migrations

```bash
# Install golang-migrate if not already installed
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
make migrate-up
```

### 3. Run the Application

```bash
# Development mode
make dev

# Or directly
go run cmd/server/main.go
```

The server will start on `http://localhost:8080` (or the port specified in `PORT` env var).

## API Endpoints

### Health Checks
- `GET /healthz` - Liveness probe
- `GET /readyz` - Readiness probe (checks database connectivity)

### Authentication
- `GET /api/v1/auth/me` - Get current user info (requires auth)

### Tenants
- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants/{id}` - Get tenant
- `PUT /api/v1/tenants/{id}` - Update tenant
- `POST /api/v1/tenants/{id}/invites` - Invite member
- `GET /api/v1/tenants/{id}/members` - List members
- `DELETE /api/v1/tenants/{id}/members/{user_id}` - Remove member
- `GET /api/v1/tenants/{id}/roles` - List available roles
- `GET /api/v1/tenants/{id}/seat-usage` - Get seat usage
- `POST /api/v1/tenants/{id}/clients` - Create client
- `GET /api/v1/tenants/{id}/clients` - List clients

### Clients
- `GET /api/v1/clients/{id}` - Get client
- `PUT /api/v1/clients/{id}` - Update client
- `POST /api/v1/clients/{id}/members` - Add client member
- `GET /api/v1/clients/{id}/members` - List client members
- `DELETE /api/v1/clients/{id}/members/{memberId}` - Remove client member
- `POST /api/v1/clients/{id}/locations` - Create location
- `GET /api/v1/clients/{id}/locations` - List locations

### Locations
- `PUT /api/v1/locations/{id}` - Update location

### Brand
- `GET /api/v1/brand/by-domain?domain=example.com` - Get branding by domain
- `GET /api/v1/brand/by-host?host=example.com` - Get branding by host
- `GET /api/v1/brands` - List brands (requires auth)
- `POST /api/v1/brands` - Create/update brand (requires auth)
- `GET /api/v1/brands/{brandId}` - Get brand (requires auth)
- `PUT /api/v1/brands/{brandId}` - Update brand (requires auth)
- `DELETE /api/v1/brands/{brandId}` - Delete brand (requires auth)

### Files
- `POST /api/v1/files/sign` - Generate pre-signed URL for upload
- `DELETE /api/v1/files/{key}` - Delete file

## Database Migrations

Migrations are managed using `golang-migrate/migrate`.

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status

# Create a new migration
make migrate-create NAME=add_new_table
```

Migrations are located in `migrations/` directory.

## Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Run Docker container
make docker-run
```

## Testing

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose
```

## Linting

```bash
# Run linters
make lint

# Format code
make fmt
```

## Deployment to Google Cloud Run

### Prerequisites
- Google Cloud SDK installed
- Docker installed
- Cloud Run API enabled
- Cloud SQL instance created

### Steps

1. **Build and push Docker image:**
```bash
# Set your project ID
export PROJECT_ID=your-project-id
export REGION=us-central1
export SERVICE_NAME=farohq-core-app

# Build and push
gcloud builds submit --tag gcr.io/$PROJECT_ID/$SERVICE_NAME
```

2. **Deploy to Cloud Run:**
```bash
gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars "PORT=8080" \
  --set-env-vars "DATABASE_URL=/cloudsql/project:region:instance" \
  --set-env-vars "CLERK_JWKS_URL=..." \
  --set-env-vars "AWS_REGION=..." \
  --set-env-vars "S3_BUCKET_NAME=..." \
  --add-cloudsql-instances project:region:instance \
  --service-account service-account@project.iam.gserviceaccount.com
```

3. **Run migrations:**
```bash
# Connect to Cloud SQL and run migrations
# Or use Cloud SQL Proxy locally
cloud_sql_proxy -instances=project:region:instance=tcp:5432
make migrate-up
```

## Multi-Tenancy

The application uses Row Level Security (RLS) in PostgreSQL to enforce tenant isolation. The `TenantResolution` middleware:
1. Resolves tenant from request host/domain
2. Sets `lv.tenant_id` in PostgreSQL session for RLS
3. All database queries are automatically filtered by tenant

## Security

- **Authentication**: Clerk JWT tokens validated via JWKS
- **Authorization**: Role-based access control (owner, admin, staff, viewer)
- **Tenant Isolation**: Enforced via RLS at database level
- **File Uploads**: Pre-signed URLs with expiration (10 minutes)
- **Secrets**: Never logged, stored in environment variables only

## License

Proprietary - FaroHQ

