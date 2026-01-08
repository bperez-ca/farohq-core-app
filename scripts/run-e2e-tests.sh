#!/bin/bash

set -e

echo "🧪 Running E2E tests..."

cd "$(dirname "$0")/.."

# Ensure E2E environment is running
if ! docker ps | grep -q farohq-core-app-postgres-e2e; then
    echo "⚠️  E2E environment not running. Starting it now..."
    ./scripts/start-e2e.sh
fi

# Set test database URL
export DATABASE_URL="postgres://postgres:password@localhost:5433/localvisibilityos?sslmode=disable"

# Run E2E tests
echo "🔬 Running tests with E2E database..."
go test -v ./... -tags=e2e

echo "✅ E2E tests completed"
