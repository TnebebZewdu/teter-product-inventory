# Stage 1: Build the Go application
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the project source code
COPY . .

# Build the API executable
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api

# Stage 2: Run the application
FROM alpine:latest

WORKDIR /app

# Install certificates for secure connections
RUN apk --no-cache add ca-certificates

# Copy the executable from the builder stage
COPY --from=builder /app/api .

# Expose the API port
EXPOSE 8080

# Start the API
CMD ["./api"]