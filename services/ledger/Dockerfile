FROM golang:1.26-alpine AS builder
WORKDIR /app/services/ledger
COPY services/ledger/go.mod ./
COPY services/ledger/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ledger-server ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /ledger-server /app/ledger-server
COPY --from=builder /app/services/ledger/migrations /app/migrations
USER appuser:appgroup
EXPOSE 8087
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8087/v1/health || exit 1
CMD ["/app/ledger-server"]
