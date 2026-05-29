FROM golang:1.26-alpine AS builder
WORKDIR /app/services/user
COPY services/user/go.mod ./
COPY services/user/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /user-server /app/user-server
COPY --from=builder /app/services/user/migrations /app/migrations
EXPOSE 8088
CMD ["/app/user-server"]
