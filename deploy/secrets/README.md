# RoundPenny Secrets Management

## Overview

RoundPenny supports two modes for secret management:

- **Local Development** — Environment variables (EnvManager)
- **Production** — HashiCorp Vault KV v2 (VaultManager)

## Modes

### 1. Environment Variables (Local Dev)

Set secrets directly in `.env` at the project root. The `EnvManager` reads from `os.Getenv()`.

```bash
cp .env.example .env
# edit .env with your local values
```

### 2. Vault (Production)

The `VaultManager` authenticates to Vault with a token and reads secrets from a KV v2 engine.

**Setup:**

```bash
# Start Vault in dev mode (not for production)
vault server -dev

# Enable KV v2
vault secrets enable -path=secret kv-v2

# Write secrets
vault kv put secret/db_password value=...
vault kv put secret/jwt_secret value=...
vault kv put secret/stripe_api_key value=...
vault kv put secret/sendgrid_api_key value=...
vault kv put secret/onfido_api_token value=...
vault kv put secret/redis_password value=...
```

**Application Configuration:**

Set the following environment variables for the services:

| Variable              | Description                        |
| --------------------- | ---------------------------------- |
| `VAULT_ADDR`          | Vault server URL                   |
| `VAULT_TOKEN`         | Vault authentication token         |
| `VAULT_SECRET_PATH`   | KV v2 mount path (default: secret) |

## Managing Encrypted .env Files

Use `scripts/manage-secrets.sh` to encrypt/decrypt `.env` files for secure storage.

```bash
# Encrypt
./scripts/manage-secrets.sh encrypt .env .env.encrypted

# Decrypt
./scripts/manage-secrets.sh decrypt .env.encrypted .env
```

You will be prompted for a password. Use a strong, shared password managed through a team password manager.

## Secret Inventory

The following environment variables contain sensitive values:

| Variable            | Used By                          |
| ------------------- | -------------------------------- |
| `DB_PASSWORD`      | All services with data storage   |
| `JWT_SECRET`       | Token signing / validation       |
| `STRIPE_API_KEY`   | Payment processing               |
| `SENDGRID_API_KEY` | Email notifications              |
| `ONFIDO_API_TOKEN` | KYC identity verification        |
| `REDIS_PASSWORD`   | Cache and rate limiting          |

## Secret Rotation Policy

- **JWT_SECRET**: Rotate every 90 days or immediately upon compromise.
- **STRIPE_API_KEY**: Rotate every 180 days. Use Stripe's rolling keys feature.
- **SENDGRID_API_KEY**: Rotate every 180 days.
- **ONFIDO_API_TOKEN**: Rotate every 180 days.
- **DB_PASSWORD**: Rotate every 90 days.
- **REDIS_PASSWORD**: Rotate every 180 days.

**Rotation procedure:**

1. Write the new secret to Vault at the same path.
2. Restart the affected services (they will re-read on restart).
3. For zero-downtime rotation, use blue/green deployment or sidecar refresh logic.

### Emergency Rotation

If a secret is compromised:

1. Immediately write the new value to Vault.
2. Restart all affected services.
3. Revoke any compromised tokens/keys at the provider (Stripe, SendGrid, Onfido).
4. Audit access logs to determine scope of exposure.
