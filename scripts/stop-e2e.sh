#!/bin/bash

set -e

echo "🛑 Stopping E2E environment..."

cd "$(dirname "$0")/.."

# Stop and remove containers
echo "🐘 Stopping PostgreSQL..."
docker-compose -f docker-compose.e2e.yml down

echo "✅ E2E environment stopped"
