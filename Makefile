.PHONY: all build test vet lint docker-all clean

all: build

# Build all services
build: auth user merchant transaction roundup-engine wallet fee investment ledger payment-gateway fraud analytics notification admin subscription

# Test all services
test:
	@for dir in services/*/; do \
		echo "=== Testing $$dir ==="; \
		(cd "$$dir" && go test ./... -count=1) || exit 1; \
	done
	@for dir in pkg/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "=== Testing $$dir ==="; \
			(cd "$$dir" && go test ./... -count=1) || exit 1; \
		fi; \
	done

# Vet all services
vet:
	@for dir in services/*/; do \
		echo "=== Vetting $$dir ==="; \
		(cd "$$dir" && go vet ./...) || exit 1; \
	done

# Build individual services
auth:
	cd services/auth && CGO_ENABLED=0 go build -o bin/auth-server ./cmd/server

user:
	cd services/user && CGO_ENABLED=0 go build -o bin/user-server ./cmd/server

merchant:
	cd services/merchant && CGO_ENABLED=0 go build -o bin/merchant-server ./cmd/server

transaction:
	cd services/transaction && CGO_ENABLED=0 go build -o bin/transaction-server ./cmd/server

roundup-engine:
	cd services/roundup-engine && CGO_ENABLED=0 go build -o bin/roundup-engine ./cmd/server

wallet:
	cd services/wallet && CGO_ENABLED=0 go build -o bin/wallet-server ./cmd/server

fee:
	cd services/fee && CGO_ENABLED=0 go build -o bin/fee-server ./cmd/server

investment:
	cd services/investment && CGO_ENABLED=0 go build -o bin/investment-server ./cmd/server

ledger:
	cd services/ledger && CGO_ENABLED=0 go build -o bin/ledger-server ./cmd/server

payment-gateway:
	cd services/payment-gateway && CGO_ENABLED=0 go build -o bin/payment-gateway-server ./cmd/server

fraud:
	cd services/fraud && CGO_ENABLED=0 go build -o bin/fraud-server ./cmd/server

analytics:
	cd services/analytics && CGO_ENABLED=0 go build -o bin/analytics-server ./cmd/server

notification:
	cd services/notification && CGO_ENABLED=0 go build -o bin/notification-server ./cmd/server

admin:
	cd services/admin && CGO_ENABLED=0 go build -o bin/admin-server ./cmd/server

subscription:
	cd services/subscription && CGO_ENABLED=0 go build -o bin/subscription-server ./cmd/server

# Docker images
docker-auth:
	docker build -t roundpenny/auth-service:latest -f services/auth/Dockerfile .

docker-user:
	docker build -t roundpenny/user-service:latest -f services/user/Dockerfile .

docker-merchant:
	docker build -t roundpenny/merchant-service:latest -f services/merchant/Dockerfile .

docker-transaction:
	docker build -t roundpenny/transaction-service:latest -f services/transaction/Dockerfile .

docker-roundup-engine:
	docker build -t roundpenny/roundup-engine:latest -f services/roundup-engine/Dockerfile .

docker-wallet:
	docker build -t roundpenny/wallet-service:latest -f services/wallet/Dockerfile .

docker-fee:
	docker build -t roundpenny/fee-service:latest -f services/fee/Dockerfile .

docker-investment:
	docker build -t roundpenny/investment-service:latest -f services/investment/Dockerfile .

docker-ledger:
	docker build -t roundpenny/ledger-service:latest -f services/ledger/Dockerfile .

docker-payment-gateway:
	docker build -t roundpenny/payment-gateway-service:latest -f services/payment-gateway/Dockerfile .

docker-fraud:
	docker build -t roundpenny/fraud-service:latest -f services/fraud/Dockerfile .

docker-analytics:
	docker build -t roundpenny/analytics-service:latest -f services/analytics/Dockerfile .

docker-notification:
	docker build -t roundpenny/notification-service:latest -f services/notification/Dockerfile .

docker-admin:
	docker build -t roundpenny/admin-service:latest -f services/admin/Dockerfile .

docker-subscription:
	docker build -t roundpenny/subscription-service:latest -f services/subscription/Dockerfile .

docker-all: docker-auth docker-user docker-merchant docker-transaction docker-roundup-engine docker-wallet docker-fee docker-investment docker-ledger docker-payment-gateway docker-fraud docker-analytics docker-notification docker-admin docker-subscription

docker-compose-up:
	docker compose up -d --build

docker-compose-down:
	docker compose down -v

# Misc
clean:
	rm -rf services/*/bin
	rm -rf services/*/vendor
