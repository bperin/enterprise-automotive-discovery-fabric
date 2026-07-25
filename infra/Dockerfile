# Multi-stage build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/service ./cmd/service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
CMD ["./service"]
