#!/usr/bin/env bash
set -euo pipefail
export LC_ALL=C
sql() {
  mysql --protocol=TCP --ssl-mode=DISABLED -h "$MYSQL_HOST" -u "$MYSQL_USER" \
    --default-character-set=utf8mb4 --batch --skip-column-names "$MYSQL_DATABASE" "$@"
}
sql -e 'CREATE TABLE IF NOT EXISTS casino_schema_migrations (
  version VARCHAR(160) PRIMARY KEY, checksum CHAR(64) NOT NULL,
  dirty BOOLEAN NOT NULL DEFAULT TRUE, applied_at DATETIME(3) NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;'
apply() {
  local version="$1" file="$2" checksum state
  checksum=$(sha256sum "$file" | cut -d ' ' -f 1)
  state=$(sql -e "SELECT CONCAT(checksum, ':', dirty) FROM casino_schema_migrations WHERE version='$version'")
  if [[ -n "$state" ]]; then
    if [[ "$state" != "$checksum:0" ]]; then
      echo "Migration $version is dirty or changed. Restore/repair it before restarting; startup has been stopped." >&2
      exit 1
    fi
    echo "Already applied: $version"
    return
  fi
  echo "Applying: $version"
  sql -e "INSERT INTO casino_schema_migrations(version,checksum,dirty) VALUES('$version','$checksum',1)"
  sql < "$file"
  sql -e "UPDATE casino_schema_migrations SET dirty=0,applied_at=NOW(3) WHERE version='$version'"
}
apply 00000000_base.sql /migrations/base.sql
while IFS= read -r version || [[ -n "$version" ]]; do
  [[ -z "$version" || "$version" == \#* ]] && continue
  [[ "$version" =~ ^[0-9]{8}_[a-z0-9_]+\.sql$ ]] || { echo "Invalid migration name: $version" >&2; exit 1; }
  apply "$version" "/migrations/sql/$version"
done < /migrations/manifest.txt
echo 'Database migrations complete.'
