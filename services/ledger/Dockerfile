FROM golang:1.26-alpine AS builder
WORKDIR /app/services/ledger
COPY services/ledger/go.mod ./
COPY services/ledger/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /ledger-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /ledger-server /app/ledger-server
COPY --from=builder /app/services/ledger/migrations /app/migrations
EXPOSE 8087
CMD ["/app/ledger-server"]
