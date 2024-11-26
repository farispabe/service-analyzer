# Stage 1: Build Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o core-service

# Stage 2: Run the Go binary in a minimal container
FROM alpine:latest

WORKDIR /app

# Install required tools: PostgreSQL client, git, bash, and wait-for-it.sh
RUN apk add --no-cache bash postgresql-client git wget && \
    wget -O /wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x /wait-for-it.sh

# Copy files from the builder stage
COPY --from=builder /app/core-service /app/core-service
COPY ./migrations /app/migrations
COPY ./entrypoint.sh /app/entrypoint.sh
COPY ./.pgpass /app/.pgpass

RUN chmod +x /app/entrypoint.sh
RUN chmod 600 /app/.pgpass

EXPOSE 7012

ENTRYPOINT ["bash", "/app/entrypoint.sh"]
