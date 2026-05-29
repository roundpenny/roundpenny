FROM golang:1.26-alpine AS builder
WORKDIR /app/services/notification
COPY services/notification/go.mod ./
COPY services/notification/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /notification-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /notification-server /app/notification-server
COPY --from=builder /app/services/notification/migrations /app/migrations
EXPOSE 8091
CMD ["/app/notification-server"]
