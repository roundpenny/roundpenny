#!/bin/bash
# Copyright (c) 2026 RoundPenny. All rights reserved.
set -euo pipefail

DB_HOST="${1:-localhost}"
DB_PORT="${2:-5432}"
DB_USER="${3:-roundup}"
DB_NAME="${4:-roundup}"
BACKUP_DIR="${5:-./backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/roundpenny_${TIMESTAMP}.sql.gz"
RETENTION_DAYS="${RETENTION_DAYS:-30}"

mkdir -p "$BACKUP_DIR"

echo "=== RoundPenny Database Backup ==="
echo "Host: $DB_HOST:$DB_PORT"
echo "Database: $DB_NAME"
echo "Output: $BACKUP_FILE"

# Full database dump
PGPASSWORD="${PGPASSWORD:-}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --no-owner \
    --no-acl \
    --format=custom \
    --verbose \
    | gzip > "$BACKUP_FILE"

echo "Backup complete: $(ls -lh "$BACKUP_FILE")"

# Rotate old backups
find "$BACKUP_DIR" -name "roundpenny_*.sql.gz" -mtime +$RETENTION_DAYS -delete
echo "Removed backups older than $RETENTION_DAYS days"

# Test backup integrity
echo "Testing backup integrity..."
gzip -t "$BACKUP_FILE" && echo "Backup integrity check: PASSED" || echo "Backup integrity check: FAILED"
