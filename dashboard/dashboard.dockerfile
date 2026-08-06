FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashboard-api ./dashboard

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/dashboard-api .

EXPOSE 8081

ENTRYPOINT ["./dashboard-api"]
