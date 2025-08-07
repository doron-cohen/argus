TOOLS_BIN := $(shell go env GOPATH)/bin
oapi-codegen := $(TOOLS_BIN)/oapi-codegen
OPENAPI_SPEC := backend/api/openapi.yaml
API_OUT := backend/api/api.gen.go
CLIENT_OUT := backend/api/client/client.gen.go
SYNC_OPENAPI_SPEC := backend/sync/api/openapi.yaml
SYNC_API_OUT := backend/sync/api/api.gen.go
SYNC_CLIENT_OUT := backend/sync/api/client/client.gen.go
REPORTS_OPENAPI_SPEC := backend/reports/api/openapi.yaml
REPORTS_API_OUT := backend/reports/api/api.gen.go
REPORTS_CLIENT_OUT := backend/reports/api/client/client.gen.go

# Check if Volta is available
VOLTA_AVAILABLE := $(shell command -v volta 2> /dev/null)

.PHONY: all install-tools backend/gen-all backend/go-mod-tidy backend/lint backend/test backend/build backend/ci frontend/test frontend/build frontend/lint

all: backend/gen-all backend/go-mod-tidy frontend/build

install-tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

backend/gen-api-server: install-tools
	$(oapi-codegen) -generate types,chi-server,spec -package api -o $(API_OUT) $(OPENAPI_SPEC)

backend/gen-api-client: install-tools
	$(oapi-codegen) -generate types,client -package client -o $(CLIENT_OUT) $(OPENAPI_SPEC)

backend/gen-sync-api-server: install-tools
	$(oapi-codegen) -generate types,chi-server,spec -package api -o $(SYNC_API_OUT) $(SYNC_OPENAPI_SPEC)

backend/gen-sync-api-client: install-tools
	$(oapi-codegen) -generate types,client -package client -o $(SYNC_CLIENT_OUT) $(SYNC_OPENAPI_SPEC)

backend/gen-reports-api-server: install-tools
	$(oapi-codegen) -generate types,chi-server,spec -package api -o $(REPORTS_API_OUT) $(REPORTS_OPENAPI_SPEC)

backend/gen-reports-api-client: install-tools
	$(oapi-codegen) -generate types,client -package client -o $(REPORTS_CLIENT_OUT) $(REPORTS_OPENAPI_SPEC)

backend/gen-all: backend/gen-api-server backend/gen-api-client backend/gen-sync-api-server backend/gen-sync-api-client backend/gen-reports-api-server backend/gen-reports-api-client
	true

backend/go-mod-tidy:
	cd backend && go mod tidy
	cd backend/api/client && go mod tidy
	cd backend/sync/api/client && go mod tidy
	cd backend/reports/api/client && go mod tidy

backend/lint:
	cd backend && golangci-lint run --timeout=5m

backend/test: frontend/build
	cd backend && go test -v $(if $(filter 1,$(CGO_ENABLED)),-race,) -coverprofile=coverage.out ./...

backend/build: frontend/build
	cd backend && go build -ldflags="-w -s" -o bin/argus ./cmd/main.go

backend/build-with-deps: backend/go-mod-tidy frontend/build
	cd backend && go mod download
	cd backend && go build -ldflags="-w -s" -o bin/argus ./cmd/main.go

backend/clean:
	cd backend && rm -f coverage.out bin/argus

backend/gen-all-and-diff: backend/gen-all backend/go-mod-tidy
	@echo "Checking for uncommitted changes in go.sum or generated files..."
	@if [ -n "$$(git status --porcelain backend/*/go.mod backend/*/go.sum)" ]; then \
		echo "Error: Uncommitted changes detected in go.mod or go.sum files"; \
		echo "Please run 'make backend/go-mod-tidy' and commit the changes"; \
		git diff backend/*/go.mod backend/*/go.sum; \
		exit 1; \
	fi
	@echo "All go.mod and go.sum files are clean"

# Frontend tasks with Volta support (fallback to direct commands in CI)
frontend/install:
	cd frontend && bun install

frontend/dev:
	cd frontend && bun run dev

frontend/build:
	cd frontend && bun run build:prod

frontend/build-dev:
	cd frontend && bun run build:dev

frontend/type-check:
	cd frontend && bun run type-check

# Frontend tests
frontend/test:
	cd frontend && bun run type-check

frontend/test-unit:
	cd frontend && bun run test:unit

frontend/test-e2e: frontend/install
    cd frontend && bunx playwright install
    cd frontend && CI=true bun run test:e2e --reporter=list

frontend/test-e2e-with-seed: frontend/install
    cd frontend && bunx playwright install
    ARGUS_BASE_URL=http://localhost:8080 bun ./scripts/report-seeder.ts --only auth-service --all-statuses --reports-per-component 5
    cd frontend && CI=true bun run test:e2e --reporter=list

frontend/test-e2e-real: frontend/install
	cd frontend && bun run test:e2e

# Run E2E tests with real application (fixed shell variable scope issue)
frontend/test-e2e-app: frontend/install
    docker compose up -d --wait
    ARGUS_BASE_URL=http://localhost:8080 bun ./frontend/scripts/report-seeder.ts --only auth-service --all-statuses --reports-per-component 5
    cd frontend && CI=true bun run test:e2e --reporter=list; test_exit_code=$$?; docker compose down; exit $$test_exit_code

frontend/test-e2e-ci: frontend/install
    cd frontend && bunx playwright install
    ARGUS_BASE_URL=http://localhost:8080 bun ./frontend/scripts/report-seeder.ts --only auth-service --all-statuses --reports-per-component 5
    cd frontend && CI=true bun run test:e2e --reporter=list

frontend/test-all: frontend/test frontend/test-unit frontend/test-unit-bun frontend/test-e2e

# Frontend lint
frontend/lint:
	cd frontend && bun run type-check

frontend/clean:
	cd frontend && rm -rf dist coverage.out node_modules

# Frontend production build validation
frontend/validate-build:
	@echo "Validating frontend production build..."
	@cd frontend && test -f dist/main.js || (echo "Error: main.js not found in dist/" && exit 1)
	@cd frontend && test -f dist/styles.css || (echo "Error: styles.css not found in dist/" && exit 1)
	@echo "Frontend build validation passed"

# Combined tasks

ci: backend/ci frontend/ci

test: backend/test frontend/test-all

build: backend/build frontend/build

lint: backend/lint frontend/lint

clean: backend/clean frontend/clean

# Docker tasks
docker/build:
	docker build -t argus-backend -f backend/Dockerfile .

docker/build-backend:
	docker build -t argus-backend -f backend/Dockerfile .

docker/up:
	docker compose up -d

docker/up-build:
	docker compose up -d --build

docker/down:
	docker compose down

docker/logs:
	docker compose logs -f

docker/logs-backend:
	docker compose logs -f backend

docker/logs-postgres:
	docker compose logs -f postgres

docker/restart:
	docker compose restart

docker/restart-backend:
	docker compose restart backend

docker/clean:
	docker compose down -v
	docker system prune -f

docker/develop:
	docker compose watch

docker/status:
	docker compose ps

docker/test:
	./test-docker.sh

# Frontend CI pipeline
frontend/ci: frontend/install frontend/lint frontend/test frontend/build frontend/validate-build frontend/test-e2e-ci

# Backend CI pipeline
backend/ci: backend/gen-all backend/go-mod-tidy backend/lint backend/test backend/build

# Seed test data
seed-reports:
	bun scripts/seed-reports.js

seed-reports-help:
	bun scripts/seed-reports.js --help

test-seed:
	bun scripts/test-seed-script.js

# Clear database (requires make dev to be running)
clean-db:
	@echo "⚠️  This will clear all data in the database"
	@echo "Make sure the development server is running (make dev)"
	curl -X DELETE http://localhost:8080/api/admin/reset || echo "❌ Failed to reset database"

# Seed specific scenarios for testing
seed-test-scenarios: seed-mixed-reports

seed-mixed-reports:
	@echo "🌱 Seeding mixed scenario: some components with reports, others without"
	bun scripts/seed-reports.js --exclude user-service --reports-per-component 3

seed-comprehensive:
	@echo "🌱 Seeding comprehensive test data with all status types"
	bun scripts/seed-reports.js --all-statuses --reports-per-component 7
