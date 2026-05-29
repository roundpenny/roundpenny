FROM golang:1.26-alpine AS builder
WORKDIR /app/services/transaction
COPY services/transaction/go.mod ./
COPY services/transaction/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /transaction-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /transaction-server /app/transaction-server
COPY --from=builder /app/services/transaction/migrations /app/migrations
EXPOSE 8083
CMD ["/app/transaction-server"]
