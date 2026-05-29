FROM golang:1.26-alpine AS builder
WORKDIR /app/services/merchant
COPY services/merchant/go.mod ./
COPY services/merchant/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /merchant-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /merchant-server /app/merchant-server
COPY --from=builder /app/services/merchant/migrations /app/migrations
EXPOSE 8092
CMD ["/app/merchant-server"]
