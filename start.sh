#!/bin/bash

# Quick Start Script for Sales API
# This script helps you get started quickly with the Sales API

set -e

echo "🚀 Sales API Quick Start"
echo "========================"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ .env file created"
    echo "⚠️  Please edit .env file with your actual configuration before proceeding"
    echo ""
    read -p "Press Enter to continue after editing .env..."
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

echo "🐳 Starting Docker containers..."
docker-compose up -d

echo ""
echo "⏳ Waiting for database to be ready..."
sleep 10

echo ""
echo "📊 Running database migrations..."
docker-compose run --rm migrate up

echo ""
echo "✅ Sales API is ready!"
echo ""
echo "📍 API is available at: http://localhost:4000"
echo "📊 Metrics available at: http://localhost:4000/v1/metrics"
echo ""
echo "📝 Useful commands:"
echo "  - View logs: make docker/logs"
echo "  - Stop API: make docker/down"
echo "  - Restart API: make docker/restart"
echo "  - Run tests: make test"
echo ""
echo "🔗 Next steps:"
echo "  1. Register a user: curl -X POST http://localhost:4000/v1/users ..."
echo "  2. Check the README.md for full API documentation"
echo "  3. View logs: make docker/logs"
echo ""
