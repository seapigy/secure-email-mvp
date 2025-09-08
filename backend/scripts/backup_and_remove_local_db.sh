#!/usr/bin/env bash
set -euo pipefail

# DO NOT RUN AGAINST PRODUCTION WITHOUT CONFIRMATION.
# Scans repo for local sqlite/db files, backs them up, and removes them.
# If a file looks like it contains production connection strings, aborts.

BACKUP_DIR="/opt/secure-email/backups"
DEV_BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

mkdir -p "$BACKUP_DIR" || true
mkdir -p "$DEV_BACKUP_DIR" || true

echo "Searching for local DB files..."
FOUND=0
shopt -s nullglob
for f in $(git ls-files | grep -E "\.db$|sqlite|secure-email.*\.db" || true); do
  FOUND=1
  echo "Found DB file tracked by git: $f"
  echo "Backing up to $DEV_BACKUP_DIR"
  cp "$f" "$DEV_BACKUP_DIR/$(basename "$f").$TIMESTAMP"
done

# Also scan working directory for obvious DB files (not necessarily in git)
for f in ./*.db ./*sqlite* 2>/dev/null; do
  if [ -f "$f" ]; then
    FOUND=1
    echo "Found local DB file: $f"
    echo "Creating gzip backup to $BACKUP_DIR"
    gzip -c "$f" > "$BACKUP_DIR/$(basename "$f").$TIMESTAMP.db.gz"
    echo "Backup created: $BACKUP_DIR/$(basename "$f").$TIMESTAMP.db.gz"
    # Basic check for strings that look like production connection URLs
    if grep -E -q "postgresql|oracle|mysql|DATABASE_URL|host=|user=|password=" "$f" 2>/dev/null || grep -E -q "https?://" "$f" 2>/dev/null; then
      echo "WARNING: file $f may contain production connection strings. Manual review required. Exiting."
      exit 2
    fi
    echo "Removing local DB file: $f"
    rm -f "$f"
  fi
done

if [ $FOUND -eq 0 ]; then
  echo "No local DB files found."
else
  echo "Local DB files backed up and removed (if any)."
fi

echo "Done."
