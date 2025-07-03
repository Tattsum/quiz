#!/bin/bash

# テストデータベース環境でのテスト実行スクリプト

set -e

echo "🚀 Starting test environment..."

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up test environment..."
    docker-compose -f docker-compose.test.yml down -v --remove-orphans
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Start test databases
echo "📊 Starting test databases..."
docker-compose -f docker-compose.test.yml up -d

# Wait for databases to be ready
echo "⏳ Waiting for databases to be ready..."
timeout 60 bash -c '
    until docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U quiz_user -d quiz_db_test; do
        echo "Waiting for PostgreSQL..."
        sleep 2
    done
'

timeout 30 bash -c '
    until docker-compose -f docker-compose.test.yml exec -T redis-test redis-cli ping; do
        echo "Waiting for Redis..."
        sleep 2
    done
'

echo "✅ Test databases are ready!"

# Set test environment variables
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=quiz_user
export DB_PASSWORD=quiz_password
export DB_NAME=quiz_db_test
export REDIS_HOST=localhost
export REDIS_PORT=6380
export TEST_ENV=true

# Run tests with coverage
echo "🧪 Running tests with coverage..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Generate coverage report
echo "📈 Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out | tee coverage.txt

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "📊 Total coverage: ${COVERAGE}%"

if (( $(echo "$COVERAGE < 70" | bc -l) )); then
    echo "❌ Coverage ${COVERAGE}% is below threshold of 70%"
    exit 1
else
    echo "✅ Coverage ${COVERAGE}% meets threshold!"
fi

echo "🎉 All tests completed successfully!"