# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o farohq-core-app cmd/server/main.go

# Final stage - use distroless for minimal image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy the binary from builder stage
COPY --from=builder /app/farohq-core-app .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose port (Cloud Run will set PORT env var)
EXPOSE 8080

# Run as non-root user (distroless provides nonroot user)
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["./farohq-core-app"]

