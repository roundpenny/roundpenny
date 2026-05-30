# Copyright (c) 2026 RoundPenny. All rights reserved.

FROM golang:1.26-alpine AS builder
WORKDIR /app/services/fraud
COPY services/fraud/go.mod ./
COPY services/fraud/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /fraud-server ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /fraud-server /app/fraud-server
COPY --from=builder /app/services/fraud/migrations /app/migrations
USER appuser:appgroup
EXPOSE 8094
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8094/v1/health || exit 1
CMD ["/app/fraud-server"]
