# Copyright (c) 2026 RoundPenny. All rights reserved.

FROM golang:1.26-alpine AS builder
WORKDIR /app/services/subscription
COPY services/subscription/go.mod ./
COPY services/subscription/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /subscription-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /subscription-server /app/subscription-server
COPY --from=builder /app/services/subscription/migrations /app/migrations
EXPOSE 8096
CMD ["/app/subscription-server"]
