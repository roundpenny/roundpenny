FROM golang:1.26-alpine AS builder
WORKDIR /app/services/fraud
COPY services/fraud/go.mod ./
COPY services/fraud/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /fraud-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /fraud-server /app/fraud-server
COPY --from=builder /app/services/fraud/migrations /app/migrations
EXPOSE 8094
CMD ["/app/fraud-server"]
