#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C
umask 077
sql() {
  mysql --protocol=TCP --ssl-mode=DISABLED -h "$MYSQL_HOST" -u "$MYSQL_USER" \
    --default-character-set=utf8mb4 --batch --skip-column-names "$MYSQL_DATABASE" "$@"
}
version=20261126_single_h5
checksum=$(cat /database/schema.sql /database/seed.sql /database/upgrade_legacy.sql | sha256sum | cut -d ' ' -f 1)
sql -e 'CREATE TABLE IF NOT EXISTS casino_schema_migrations (
  version VARCHAR(160) PRIMARY KEY, checksum CHAR(64) NOT NULL,
  dirty BOOLEAN NOT NULL DEFAULT TRUE, applied_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;'
state=$(sql -e "SELECT CONCAT(checksum, ':', dirty) FROM casino_schema_migrations WHERE version='$version'")
if [[ -n "$state" ]]; then
  [[ "$state" == "$checksum:0" ]] || { echo 'Database installation is incomplete or its files changed. Restore the saved backup before retrying.' >&2; exit 1; }
  echo 'Single H5 database is current.'
  exit 0
fi
legacy=$(sql -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='game_saas_sites'")
existing=$(sql -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='game_users'")
if [[ "$legacy" == 1 ]]; then
  mkdir -p /backups
  backup="/backups/pre-single-h5-$(date -u +%Y%m%dT%H%M%SZ).sql"
  mysqldump --protocol=TCP --ssl-mode=DISABLED -h "$MYSQL_HOST" -u "$MYSQL_USER" \
    --default-character-set=utf8mb4 --no-tablespaces --set-gtid-purged=OFF \
    --single-transaction --hex-blob --triggers "$MYSQL_DATABASE" > "$backup.tmp"
  [[ -s "$backup.tmp" ]] || { echo 'Database backup is empty; upgrade stopped.' >&2; exit 1; }
  mv "$backup.tmp" "$backup"
  echo "Full database backup saved: $backup"
  sql -e "INSERT INTO casino_schema_migrations(version,checksum,dirty) VALUES('$version','$checksum',1)"
  sql < /database/upgrade_legacy.sql
elif [[ "$existing" == 0 ]]; then
  sql -e "INSERT INTO casino_schema_migrations(version,checksum,dirty) VALUES('$version','$checksum',1)"
  sql < /database/schema.sql
  sql < /database/seed.sql
else
  echo 'Unrecognized database state; installation stopped without changing business data.' >&2
  exit 1
fi
remaining=$(sql -e "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND COLUMN_NAME='tenant_id'")
[[ "$remaining" == 0 ]] || { echo 'Obsolete ownership columns remain; startup stopped.' >&2; exit 1; }
sql -e "UPDATE casino_schema_migrations SET dirty=0,applied_at=NOW(3) WHERE version='$version'; DELETE FROM casino_schema_migrations WHERE version<>'$version';"
echo 'Single H5 database installation complete.'
