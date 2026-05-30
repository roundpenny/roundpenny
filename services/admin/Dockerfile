FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
COPY go.work ./
COPY pkg/ ./pkg/
COPY services/admin/ ./services/admin/
RUN cd services/admin && go build -o /admin-server ./cmd/server

FROM alpine:3.20
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /admin-server /usr/local/bin/admin-server
COPY services/admin/migrations/ /migrations/
COPY services/admin/web/ /web/
USER appuser:appgroup
EXPOSE 8095
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8095/v1/health || exit 1
CMD ["admin-server"]
