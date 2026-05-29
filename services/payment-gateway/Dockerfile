FROM golang:1.26-alpine AS builder
WORKDIR /app/services/payment-gateway
COPY services/payment-gateway/go.mod ./
COPY services/payment-gateway/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /payment-server /app/payment-server
COPY --from=builder /app/services/payment-gateway/migrations /app/migrations
EXPOSE 8090
CMD ["/app/payment-server"]
