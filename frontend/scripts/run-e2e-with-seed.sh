#!/bin/bash

# Script to run E2E tests with seed data
set -e

echo "🌱 Running seed script to prepare test data..."
cd .. && bun scripts/seed-reports.js --only auth-service --all-statuses --reports-per-component 5

echo "🧪 Running E2E tests..."
cd frontend && CI=true bun run test:e2e --reporter=list 