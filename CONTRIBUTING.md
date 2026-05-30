# Contributing to RoundPenny

## Getting Started

```bash
git clone https://github.com/tmuhammet36-cmd/roundpenny.git
cd roundpenny
docker compose up -d   # start dependencies
make test              # run all tests
make build             # build all services
```

## Project Structure

- `services/` — 15 Go microservices
- `pkg/` — 20+ shared packages
- `deploy/` — Kong, Helm, Terraform, monitoring configs
- `docs/` — API spec, legal docs, runbook

## Before Submitting

- Run `make test` — all tests pass
- Run `make vet` — no vet warnings
- Run `go mod tidy` in the service you changed
- Follow existing code patterns (see any service for reference)

## Code Style

- No external test frameworks (stdlib `testing` only)
- Hand-rolled mocks for repository interfaces
- Structured logging (`log/slog`)
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Idempotency key support on mutation endpoints

## Questions

Open a GitHub issue or reach out to `partner@roundpenny.com`.
