module github.com/roundup-platform/services/payment-gateway

go 1.26.3

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/roundup-platform/pkg/config v0.0.0
	github.com/roundup-platform/pkg/cors v0.0.0
	github.com/roundup-platform/pkg/db v0.0.0
	github.com/roundup-platform/pkg/idempotency v0.0.0
	github.com/roundup-platform/pkg/monitoring v0.0.0
	github.com/roundup-platform/pkg/stripe v0.0.0
	github.com/roundup-platform/pkg/tls v0.0.0
	github.com/roundup-platform/pkg/tracing v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/redis/go-redis/v9 v9.7.1 // indirect
	github.com/stripe/stripe-go/v79 v79.12.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/grpc v1.81.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/roundup-platform/pkg/config => ../../pkg/config
	github.com/roundup-platform/pkg/cors => ../../pkg/cors
	github.com/roundup-platform/pkg/db => ../../pkg/db
	github.com/roundup-platform/pkg/event => ../../pkg/event
	github.com/roundup-platform/pkg/idempotency => ../../pkg/idempotency
	github.com/roundup-platform/pkg/kafka => ../../pkg/kafka
	github.com/roundup-platform/pkg/monitoring => ../../pkg/monitoring
	github.com/roundup-platform/pkg/stripe => ../../pkg/stripe
	github.com/roundup-platform/pkg/tls => ../../pkg/tls
	github.com/roundup-platform/pkg/tracing => ../../pkg/tracing
)
