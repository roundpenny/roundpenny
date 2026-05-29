FROM golang:1.26-alpine AS builder
WORKDIR /app/services/auth
COPY services/auth/go.mod ./
COPY services/auth/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /auth-server /app/auth-server
COPY --from=builder /app/services/auth/migrations /app/migrations
EXPOSE 8081
CMD ["/app/auth-server"]
