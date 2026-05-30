#!/bin/bash
# Copyright (c) 2026 RoundPenny. All rights reserved.
# Initial Let's Encrypt certificate setup
# Usage: ./init-letsencrypt.sh <domain> <email>
set -euo pipefail
DOMAIN="${1:?Domain required}"
EMAIL="${2:?Email required}"
docker compose -f deploy/ssl/certbot-compose.yml run --rm certbot certonly --webroot -w /var/www/certbot -d "$DOMAIN" --email "$EMAIL" --agree-tos --no-eff-email
echo "Certificates obtained for $DOMAIN"
