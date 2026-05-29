.PHONY: all build auth roundup-engine transaction wallet fee investment ledger k8s docker-compose

all: build

build: auth roundup-engine transaction wallet fee investment ledger

auth:
	cd services/auth && CGO_ENABLED=0 go build -o bin/auth-server ./cmd/server

roundup-engine:
	cd services/roundup-engine && CGO_ENABLED=0 go build -o bin/roundup-engine ./cmd/server

transaction:
	cd services/transaction && CGO_ENABLED=0 go build -o bin/transaction-server ./cmd/server

wallet:
	cd services/wallet && CGO_ENABLED=0 go build -o bin/wallet-server ./cmd/server

fee:
	cd services/fee && CGO_ENABLED=0 go build -o bin/fee-server ./cmd/server

investment:
	cd services/investment && CGO_ENABLED=0 go build -o bin/investment-server ./cmd/server

ledger:
	cd services/ledger && CGO_ENABLED=0 go build -o bin/ledger-server ./cmd/server

docker-auth:
	docker build -t roundup/auth-service:latest services/auth

docker-roundup-engine:
	docker build -t roundup/roundup-engine:latest services/roundup-engine

docker-transaction:
	docker build -t roundup/transaction-service:latest services/transaction

docker-wallet:
	docker build -t roundup/wallet-service:latest services/wallet

docker-fee:
	docker build -t roundup/fee-service:latest services/fee

docker-investment:
	docker build -t roundup/investment-service:latest services/investment

docker-all: docker-auth docker-roundup-engine docker-transaction docker-wallet docker-fee docker-investment

k8s-apply:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/config.yaml
	kubectl apply -f deploy/k8s/postgres.yaml
	kubectl apply -f deploy/k8s/kafka.yaml
	kubectl apply -f deploy/k8s/auth-service.yaml
	kubectl apply -f deploy/k8s/roundup-engine.yaml
	kubectl apply -f deploy/k8s/transaction-service.yaml
	kubectl apply -f deploy/k8s/wallet-service.yaml
	kubectl apply -f deploy/k8s/fee-service.yaml
	kubectl apply -f deploy/k8s/investment-service.yaml
	kubectl apply -f deploy/k8s/kong-gateway.yaml

k8s-delete:
	kubectl delete namespace roundup-platform

clean:
	rm -rf services/*/bin
	rm -rf services/*/vendor
