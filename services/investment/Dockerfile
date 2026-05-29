FROM golang:1.26-alpine AS builder
WORKDIR /app/services/investment
COPY services/investment/go.mod ./
COPY services/investment/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /investment-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /investment-server /app/investment-server
COPY --from=builder /app/services/investment/migrations /app/migrations
EXPOSE 8086
CMD ["/app/investment-server"]
