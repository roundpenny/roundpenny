# Copyright (c) 2026 RoundPenny. All rights reserved.

# RoundPenny Vault Configuration — Production

storage "file" {
  path = "/vault/data"
}

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_disable   = false
  tls_cert_file = "/vault/tls/tls.crt"
  tls_key_file  = "/vault/tls/tls.key"
}

api_addr     = "https://vault.roundpenny.com:8200"
cluster_addr = "https://vault.roundpenny.com:8201"

ui = true

# Audit logging
audit "file" {
  type = "file"
  path = "/vault/logs/audit.log"
  format = "json"
}

# Secrets engine — KV v2
secrets "kv-v2" {
  path        = "secret"
  description = "RoundPenny application secrets"

  config {
    max_versions = 10
    cas_required = false
    delete_version_after = "0s"
  }
}

# Auth method — AppRole for service-to-vault auth
auth "approle" {
  path        = "approle"
  description = "AppRole auth for RoundPenny services"
}

# AppRole configuration
write "auth/approle/role/roundpenny" {
  secret_id_num_uses  = 0
  secret_id_ttl       = "0"
  token_policies      = ["roundpenny-secrets"]
  token_ttl           = "1h"
  token_max_ttl       = "4h"
  bind_secret_id      = true
}

# Policy for RoundPenny services
policy "roundpenny-secrets" {
  path "secret/data/*" {
    capabilities = ["read", "list"]
  }

  path "secret/metadata/*" {
    capabilities = ["read", "list"]
  }
}

# Admin policy (CI/CD pipelines, manual operations)
policy "roundpenny-admin" {
  path "secret/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
  }

  path "auth/approle/*" {
    capabilities = ["create", "read", "update", "delete", "list"]
  }

  path "sys/policy/*" {
    capabilities = ["read", "list"]
  }
}
