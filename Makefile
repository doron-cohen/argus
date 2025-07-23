TOOLS_BIN := $(GOPATH)/bin
oapi-codegen := $(TOOLS_BIN)/oapi-codegen
OPENAPI_SPEC := backend/api/openapi.yaml
API_OUT := backend/api/api.gen.go
CLIENT_OUT := backend/api/client/client.gen.go

.PHONY: all install-tools backend/gen-all backend/go-mod-tidy

all: backend/gen-all backend/go-mod-tidy

install-tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

backend/gen-api-server: install-tools
	$(oapi-codegen) -generate types,chi-server,spec -package api -o $(API_OUT) $(OPENAPI_SPEC)

backend/gen-api-client: install-tools
	$(oapi-codegen) -generate types,client -package client -o $(CLIENT_OUT) $(OPENAPI_SPEC)

backend/gen-all: backend/gen-api-server backend/gen-api-client
	true

backend/go-mod-tidy:
	cd backend && go mod tidy
	cd backend/api/client && go mod tidy
