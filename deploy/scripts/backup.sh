#!/bin/bash
#
# Fluxbase Backup Script
# Backs up PostgreSQL database and storage to local or S3
#
# Usage: ./backup.sh [local|s3]
#
# Optional encryption: set AGE_RECIPIENT to an age public key (age1...) and
# have the `age` binary installed — every artifact (database dump, config and
# storage tarballs) is then encrypted with `age -r` before it is written, so
# no plaintext backup ever hits the backup disk. Restore the dump with:
#   age -d -i <identity-file> db_<date>.dump.age > db_<date>.dump
#

set -euo pipefail

# Backups contain credentials and user data — keep them private by default.
umask 077

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/var/backups/fluxbase}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_MODE="${1:-local}"

# Database configuration
PGHOST="${FLUXBASE_DATABASE_HOST:-postgres}"
PGPORT="${FLUXBASE_DATABASE_PORT:-5432}"
PGUSER="${FLUXBASE_DATABASE_USER:-fluxbase}"
PGDATABASE="${FLUXBASE_DATABASE_DATABASE:-fluxbase}"

# S3 configuration (optional)
S3_BUCKET="${S3_BACKUP_BUCKET:-}"
S3_PREFIX="${S3_BACKUP_PREFIX:-fluxbase}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Optional age encryption: active when AGE_RECIPIENT is set and age is installed.
USE_AGE=false
if [ -n "${AGE_RECIPIENT:-}" ]; then
    if command -v age &> /dev/null; then
        USE_AGE=true
    else
        log_warn "AGE_RECIPIENT is set but 'age' is not installed — writing UNENCRYPTED backups"
    fi
fi

# Encrypt a file in place: writes <file>.age and removes the plaintext.
encrypt_file() {
    local plain_file="$1"
    local encrypted_file="${plain_file}.age"
    age -r "$AGE_RECIPIENT" -o "$encrypted_file" "$plain_file"
    rm -f "$plain_file"
}

# Create backup directory
mkdir -p "$BACKUP_DIR"

log_info "Starting Fluxbase backup - $DATE"

# 1. Backup PostgreSQL Database
log_info "Backing up PostgreSQL database..."
DB_BACKUP_FILE="$BACKUP_DIR/db_$DATE.dump"

if command -v pg_dump &> /dev/null; then
    # Dump to a temp path so that, with encryption enabled, plaintext never
    # exists under the final backup name.
    pg_dump -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -Fc "$PGDATABASE" > "$DB_BACKUP_FILE.tmp"
    if [ "$USE_AGE" = true ]; then
        encrypt_file "$DB_BACKUP_FILE.tmp"
        DB_BACKUP_FILE="$DB_BACKUP_FILE.age"
    else
        mv "$DB_BACKUP_FILE.tmp" "$DB_BACKUP_FILE"
    fi
    log_info "Database backup completed: $DB_BACKUP_FILE"
else
    log_error "pg_dump not found. Install postgresql-client"
    exit 1
fi

# 2. Backup Configuration
log_info "Backing up configuration..."
CONFIG_BACKUP_FILE="$BACKUP_DIR/config_$DATE.tar.gz"

if [ -d "/app/config" ]; then
    tar -czf "$CONFIG_BACKUP_FILE.tmp" -C /app config/ 2>/dev/null || true
elif [ -d "/etc/fluxbase" ]; then
    tar -czf "$CONFIG_BACKUP_FILE.tmp" /etc/fluxbase/ 2>/dev/null || true
fi

if [ -f "$CONFIG_BACKUP_FILE.tmp" ]; then
    if [ "$USE_AGE" = true ]; then
        encrypt_file "$CONFIG_BACKUP_FILE.tmp"
        CONFIG_BACKUP_FILE="$CONFIG_BACKUP_FILE.age"
    else
        mv "$CONFIG_BACKUP_FILE.tmp" "$CONFIG_BACKUP_FILE"
    fi
    log_info "Config backup completed: $CONFIG_BACKUP_FILE"
fi

# 3. Backup Storage (if using local storage)
STORAGE_DIR="${FLUXBASE_STORAGE_LOCAL_PATH:-/app/storage}"
if [ -d "$STORAGE_DIR" ] && [ "$(ls -A $STORAGE_DIR 2>/dev/null)" ]; then
    log_info "Backing up local storage..."
    STORAGE_BACKUP_FILE="$BACKUP_DIR/storage_$DATE.tar.gz"
    tar -czf "$STORAGE_BACKUP_FILE.tmp" -C "$(dirname $STORAGE_DIR)" "$(basename $STORAGE_DIR)"
    if [ "$USE_AGE" = true ]; then
        encrypt_file "$STORAGE_BACKUP_FILE.tmp"
        STORAGE_BACKUP_FILE="$STORAGE_BACKUP_FILE.age"
    else
        mv "$STORAGE_BACKUP_FILE.tmp" "$STORAGE_BACKUP_FILE"
    fi
    log_info "Storage backup completed: $STORAGE_BACKUP_FILE"
fi

# 4. Generate backup manifest
MANIFEST_FILE="$BACKUP_DIR/manifest_$DATE.txt"
cat > "$MANIFEST_FILE" << EOF
Fluxbase Backup Manifest
========================
Date: $DATE
Database: $PGDATABASE
Host: $PGHOST

Files:
$(ls -lh $BACKUP_DIR/*_$DATE.*)
EOF

log_info "Backup manifest: $MANIFEST_FILE"

# 5. Upload to S3 (if configured)
if [ "$BACKUP_MODE" = "s3" ] && [ -n "$S3_BUCKET" ]; then
    log_info "Uploading backups to S3..."

    if command -v aws &> /dev/null; then
        for file in $BACKUP_DIR/*_$DATE.*; do
            aws s3 cp "$file" "s3://$S3_BUCKET/$S3_PREFIX/$(basename $file)"
        done
        log_info "S3 upload completed"
    else
        log_warn "AWS CLI not found. Skipping S3 upload"
    fi
fi

# 6. Clean old backups (retention policy)
log_info "Cleaning old backups (retention: $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete
log_info "Old backups cleaned"

# 7. Calculate backup size
TOTAL_SIZE=$(du -sh "$BACKUP_DIR" | cut -f1)
log_info "Total backup size: $TOTAL_SIZE"

log_info "Backup completed successfully!"

# Optional: Send notification (webhook, email, Slack, etc.)
if [ -n "${BACKUP_WEBHOOK_URL:-}" ]; then
    curl -X POST "$BACKUP_WEBHOOK_URL" \
        -H "Content-Type: application/json" \
        -d "{\"status\":\"success\",\"date\":\"$DATE\",\"size\":\"$TOTAL_SIZE\"}" \
        2>/dev/null || true
fi

exit 0
