#!/bin/bash
# Copyright (c) 2026 RoundPenny. All rights reserved.
# Usage: ./scripts/manage-secrets.sh encrypt|decrypt <input> <output>
# Uses AES-256-CBC via OpenSSL

set -euo pipefail

cmd="${1:-}"
input="${2:-}"
output="${3:-}"

case "$cmd" in
    encrypt)
        openssl enc -aes-256-cbc -salt -pbkdf2 -in "$input" -out "$output"
        echo "Encrypted: $input -> $output"
        ;;
    decrypt)
        openssl enc -d -aes-256-cbc -pbkdf2 -in "$input" -out "$output"
        echo "Decrypted: $input -> $output"
        ;;
    *)
        echo "Usage: $0 {encrypt|decrypt} <input> <output>"
        exit 1
        ;;
esac
