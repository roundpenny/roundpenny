FROM golang:1.26-alpine AS builder
WORKDIR /app/services/roundup-engine
COPY services/roundup-engine/go.mod ./
COPY services/roundup-engine/ ./
COPY pkg /app/pkg
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /roundup-engine ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /roundup-engine /app/roundup-engine
EXPOSE 8082
CMD ["/app/roundup-engine"]
