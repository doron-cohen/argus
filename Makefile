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

.PHONY: all install-tools backend/gen-all backend/go-mod-tidy backend/lint backend/test backend/build backend/ci

all: backend/gen-all backend/go-mod-tidy

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

backend/test:
	cd backend && go test -v $(if $(filter 1,$(CGO_ENABLED)),-race,) -coverprofile=coverage.out ./...

backend/build:
	cd backend && go build -ldflags="-w -s" -o bin/argus ./cmd/main.go

backend/gen-all-and-diff: backend/gen-all backend/go-mod-tidy
	@echo "Checking for uncommitted changes in go.sum or generated files..."
	@if [ -n "$$(git status --porcelain backend/*/go.mod backend/*/go.sum)" ]; then \
		echo "Error: Uncommitted changes detected in go.mod or go.sum files"; \
		echo "Please run 'make backend/go-mod-tidy' and commit the changes"; \
		git diff backend/*/go.mod backend/*/go.sum; \
		exit 1; \
	fi
	@echo "All go.mod and go.sum files are clean"
