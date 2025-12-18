# Build stage
FROM golang:alpine AS builder

# Install git (required for fetching dependencies)
RUN apk add --no-cache git

# Install oapi-codegen
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1

WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
# We assume the build context is the repository root
COPY service/go.mod service/go.sum ./service/

# Download dependencies
WORKDIR /app/service
RUN go mod download

# Copy the rest of the source code
WORKDIR /app
COPY openapi.yaml .
COPY service/ ./service/

# Generate code
WORKDIR /app/service
RUN go generate ./...

# Build the application
# CGO_ENABLED=0 ensures a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests (needed for email/SMS providers)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/service/main .

# Expose the configuration port (default 3000)
EXPOSE 3000

# Run the binary
CMD ["./main"]
