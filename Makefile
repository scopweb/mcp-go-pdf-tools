BINARY_SERVER = mcp-server
BINARY_MCP = mcp-server-mcp
BINARY_CLI = mcp-cli

all: build-server build-cli

build-server:
	go build -o bin/$(BINARY_SERVER) ./cmd/server

build-mcp:
	go build -o bin/$(BINARY_MCP) ./cmd/mcp-server

build-cli:
	go build -o bin/$(BINARY_CLI) ./cmd/cli

docker-build:
	docker build -t mcp-pdf-server:local .

clean:
	rm -rf bin/*

run-server:
	go run ./cmd/server

run-mcp:
	go run ./cmd/mcp-server

run-cli:
	go run ./cmd/cli --help
