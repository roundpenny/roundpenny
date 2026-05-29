FROM golang:1.26-alpine AS builder
WORKDIR /app/services/analytics
COPY services/analytics/go.mod ./
COPY services/analytics/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /analytics-server /app/analytics-server
COPY --from=builder /app/services/analytics/migrations /app/migrations
EXPOSE 8093
CMD ["/app/analytics-server"]
