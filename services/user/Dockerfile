FROM golang:1.26-alpine AS builder
WORKDIR /app/services/user
COPY services/user/go.mod ./
COPY services/user/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-server ./cmd/server

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /user-server /app/user-server
COPY --from=builder /app/services/user/migrations /app/migrations
USER appuser:appgroup
EXPOSE 8088
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget --no-verbose --tries=1 --spider http://localhost:8088/v1/health || exit 1
CMD ["/app/user-server"]
