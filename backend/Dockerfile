# in backend/Dockerfile

# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the unified binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/code-judge main.go

# --- Final Stage ---
FROM alpine:latest

RUN apk add --no-cache curl

ARG DOCKER_GID=1001

WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/code-judge /app/code-judge

# Copy the config directory
COPY config /app/config

RUN addgroup -S -g ${DOCKER_GID} docker

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -G docker

RUN mkdir -p /tmp/runner && chown -R appuser:appgroup /tmp/runner

USER appuser

# Expose the port for the web server
EXPOSE 8080

# The CMD will be provided by docker-compose