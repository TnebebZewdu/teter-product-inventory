# Stage 1: Build the Go application
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/api ./cmd/api


# Stage 2: Create the production image
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

COPY --from=builder --chown=appuser:appgroup /app/api .

# Run the application without root privileges
USER appuser

EXPOSE 8080

CMD ["./api"]