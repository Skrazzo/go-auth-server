FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o auth-server

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/auth-server .
COPY --from=builder /app/templates ./templates

# Run the binary
CMD ["./auth-server"]