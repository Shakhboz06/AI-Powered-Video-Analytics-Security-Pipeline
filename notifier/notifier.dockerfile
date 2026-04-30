# 1) Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy module files
COPY go.mod go.sum ./
RUN go mod download

# Copy everything (so notifier/ and config/ are here)
COPY . .

# Build the aggregator binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notifier ./notifier

# 2) Runtime stage
FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/notifier .

ENTRYPOINT ["./notifier"]
