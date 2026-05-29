FROM golang:1.26-alpine AS builder
WORKDIR /app/services/wallet
COPY services/wallet/go.mod ./
COPY services/wallet/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /wallet-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /wallet-server /app/wallet-server
COPY --from=builder /app/services/wallet/migrations /app/migrations
EXPOSE 8084
CMD ["/app/wallet-server"]
