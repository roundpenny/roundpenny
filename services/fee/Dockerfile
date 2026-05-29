FROM golang:1.26-alpine AS builder
WORKDIR /app/services/fee
COPY services/fee/go.mod ./
COPY services/fee/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /fee-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /fee-server /app/fee-server
COPY --from=builder /app/services/fee/migrations /app/migrations
EXPOSE 8085
CMD ["/app/fee-server"]
