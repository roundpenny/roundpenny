#!/bin/bash
# Copyright (c) 2026 RoundPenny. All rights reserved.
set -euo pipefail

BACKUP_FILE="${1:?Usage: $0 <backup-file>}"
DB_HOST="${2:-localhost}"
DB_PORT="${3:-5432}"
DB_USER="${4:-roundup}"
DB_NAME="${5:-roundup}"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "=== RoundPenny Database Restore ==="
echo "Backup: $BACKUP_FILE"
echo "Host: $DB_HOST:$DB_PORT"
echo "Database: $DB_NAME"
echo ""
echo "WARNING: This will DROP and recreate the database!"
read -p "Are you sure? (yes/N): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Restore cancelled."
    exit 1
fi

# Drop existing connections and recreate database
PGPASSWORD="${PGPASSWORD:-}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres <<SQL
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}';
DROP DATABASE IF EXISTS "${DB_NAME}";
CREATE DATABASE "${DB_NAME}";
SQL

# Restore from backup
gunzip -c "$BACKUP_FILE" | pg_restore \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --no-owner \
    --no-acl \
    --verbose

echo "Restore complete."
