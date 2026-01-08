# E2E Testing Setup

## Quick Start

### Option 1: Using Make (Recommended)

```bash
# Start E2E environment (PostgreSQL + migrations)
make e2e-start

# Run E2E tests
make e2e-test

# Stop E2E environment
make e2e-stop
```

### Option 2: Direct Script Execution

If Make targets don't work, you can run scripts directly:

```bash
# Start E2E environment
./scripts/start-e2e.sh

# Run E2E tests
./scripts/run-e2e-tests.sh

# Stop E2E environment
./scripts/stop-e2e.sh
```

## What It Does

1. **Starts PostgreSQL** in Docker (port 5433)
2. **Runs migrations** to set up the database schema
3. **Waits for PostgreSQL** to be ready
4. **Provides instructions** for starting core-app and portal

## Environment

The E2E environment uses:
- **Database**: `postgres://postgres:password@localhost:5433/localvisibilityos?sslmode=disable`
- **Port**: `5433` (to avoid conflicts with local PostgreSQL on 5432)

## Full E2E Workflow

```bash
# 1. Start E2E infrastructure
make e2e-start

# 2. In another terminal, start the core-app
export DATABASE_URL="postgres://postgres:password@localhost:5433/localvisibilityos?sslmode=disable"
export CLERK_JWKS_URL="https://real-pegasus-21.clerk.accounts.dev/.well-known/jwks.json"
make dev

# 3. In another terminal, start the portal
cd ../farohq-portal
npm run dev

# 4. Run E2E tests (optional)
make e2e-test

# 5. When done, stop everything
make e2e-stop
```

## Troubleshooting

### "make: Nothing to be done for `e2e-start'"

If Make says nothing to be done, run the script directly:

```bash
./scripts/start-e2e.sh
```

### PostgreSQL Already Running

If you get port conflicts, either:
1. Stop existing PostgreSQL: `docker stop farohq-core-app-postgres-e2e`
2. Or use a different port in `docker-compose.e2e.yml`

### Migrations Fail

Ensure the migrate tool is installed:

```bash
make install-migrate
# or
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Clean Up

To completely remove the E2E environment:

```bash
make e2e-stop
docker volume rm farohq-core-app_postgres_e2e_data
```
