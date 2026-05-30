# Copyright (c) 2026 RoundPenny. All rights reserved.

FROM golang:1.26-alpine AS builder
WORKDIR /app/services/notification
COPY services/notification/go.mod ./
COPY services/notification/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /notification-server ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /notification-server /app/notification-server
COPY --from=builder /app/services/notification/migrations /app/migrations
USER appuser:appgroup
EXPOSE 8091
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8091/v1/health || exit 1
CMD ["/app/notification-server"]
