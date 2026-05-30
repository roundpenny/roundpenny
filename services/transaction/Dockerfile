# Copyright (c) 2026 RoundPenny. All rights reserved.

FROM golang:1.26-alpine AS builder
WORKDIR /app/services/transaction
COPY services/transaction/go.mod ./
COPY services/transaction/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /transaction-server ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /transaction-server /app/transaction-server
COPY --from=builder /app/services/transaction/migrations /app/migrations
USER appuser:appgroup
EXPOSE 8083
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8083/v1/health || exit 1
CMD ["/app/transaction-server"]
