# Build stage
FROM golang:alpine AS builder

# Install git. Run `apk add --no-cache git` if your dependencies are in private repositories or require git.
# RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# CGO_ENABLED=0 ensures a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests (needed for email/SMS providers)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the configuration port (default 3000)
EXPOSE 3000

# Run the binary
CMD ["./main"]
