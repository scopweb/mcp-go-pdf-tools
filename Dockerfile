# Multi-stage build for the MCP PDF tools server
FROM golang:1.24-alpine AS build
WORKDIR /src

# Cache modules
COPY go.mod go.sum ./
RUN apk add --no-cache git ca-certificates && go env -w GOPROXY=https://proxy.golang.org
RUN go mod download

# Copy sources
COPY . .

# Build the server
WORKDIR /src/cmd/server
RUN go build -o /usr/local/bin/mcp-pdf-server .

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /usr/local/bin/mcp-pdf-server /usr/local/bin/mcp-pdf-server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mcp-pdf-server"]
