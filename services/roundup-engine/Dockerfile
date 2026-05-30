# Copyright (c) 2026 RoundPenny. All rights reserved.

FROM golang:1.26-alpine AS builder
WORKDIR /app/services/roundup-engine
COPY services/roundup-engine/go.mod ./
COPY services/roundup-engine/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /roundup-engine ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /roundup-engine /app/roundup-engine
USER appuser:appgroup
EXPOSE 8082
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8082/v1/health || exit 1
CMD ["/app/roundup-engine"]
