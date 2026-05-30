# Copyright (c) 2026 RoundPenny. All rights reserved.

FROM golang:1.26-alpine AS builder
WORKDIR /app/services/wallet
COPY services/wallet/go.mod ./
COPY services/wallet/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /wallet-server ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /wallet-server /app/wallet-server
COPY --from=builder /app/services/wallet/migrations /app/migrations
USER appuser:appgroup
EXPOSE 8084
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8084/v1/health || exit 1
CMD ["/app/wallet-server"]
