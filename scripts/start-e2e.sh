#!/bin/bash

set -e

echo "🚀 Starting E2E environment..."

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "📦 Installing golang-migrate..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Ensure go mod is tidy
echo "📦 Ensuring dependencies are up to date..."
cd "$(dirname "$0")/.."
go mod tidy

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker-compose -f docker-compose.e2e.yml up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
timeout=30
counter=0
until docker exec farohq-core-app-postgres-e2e pg_isready -U postgres > /dev/null 2>&1; do
    sleep 1
    counter=$((counter + 1))
    if [ $counter -ge $timeout ]; then
        echo "❌ PostgreSQL failed to start within $timeout seconds"
        exit 1
    fi
done

echo "✅ PostgreSQL is ready"

# Run migrations
echo "📊 Running migrations..."
export DATABASE_URL="postgres://postgres:password@localhost:5433/localvisibilityos?sslmode=disable"
migrate -path migrations -database "$DATABASE_URL" up

echo "✅ E2E environment is ready!"
echo ""
echo "Next steps:"
echo "1. Start the core-app: make dev"
echo "2. Start the portal: cd ../farohq-portal && npm run dev"
echo ""
echo "To stop: make e2e-stop"
