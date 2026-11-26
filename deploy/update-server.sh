#!/usr/bin/env bash
# Run on the deployment server. Never resets the checkout or restores live DB data.
set -Eeuo pipefail
umask 077
export GIT_TERMINAL_PROMPT=0

DEPLOY_DIR=${CASINO_PROJECT_DIR:-/opt/casino}
DEPLOY_STATE=${CASINO_DEPLOY_STATE_DIR:-/var/lib/casino-update}
DEPLOY_PROJECT=${CASINO_COMPOSE_PROJECT:-casino}
DEPLOY_RETRY=${CASINO_DEPLOY_RETRY_SECONDS:-300}
DEPLOY_TIMEOUT=${CASINO_DEPLOY_HEALTH_TIMEOUT:-600}
[[ "$DEPLOY_RETRY" =~ ^[0-9]+$ && "$DEPLOY_TIMEOUT" =~ ^[0-9]+$ ]] || exit 2
mkdir -p "$DEPLOY_STATE/backups"
chmod 700 "$DEPLOY_STATE" "$DEPLOY_STATE/backups"
exec 9>"$DEPLOY_STATE/update.lock"
flock -n 9 || exit 0
exec 3>&1

deploy_note() { printf '%s\n' "$*" >&3; }
compose() { docker compose --project-directory "$DEPLOY_DIR" -p "$DEPLOY_PROJECT" "$@"; }
DEPLOY_TARGET=
DEPLOY_OLD_API=
DEPLOY_OLD_WEB=
DEPLOY_STARTED=0

deploy_failed() {
  local result=$?
  trap - ERR
  set +e
  if [[ -n "$DEPLOY_TARGET" ]]; then
    printf '%s\n' "$DEPLOY_TARGET" > "$DEPLOY_STATE/last-failed"
    date +%s > "$DEPLOY_STATE/last-attempt"
  fi
  # Restore application image tags. Database snapshots require deliberate recovery
  # so automatic rollback can never erase transactions received during deployment.
  [[ -z "$DEPLOY_OLD_API" ]] || docker image tag "$DEPLOY_OLD_API" casino-api:local >> "$DEPLOY_STATE/deploy.log" 2>&1
  [[ -z "$DEPLOY_OLD_WEB" ]] || docker image tag "$DEPLOY_OLD_WEB" casino-web:local >> "$DEPLOY_STATE/deploy.log" 2>&1
  if [[ "$DEPLOY_STARTED" == 1 && -n "$DEPLOY_OLD_API" && -n "$DEPLOY_OLD_WEB" ]]; then
    if compose -f "$DEPLOY_STATE/compose.before.yaml" up -d --no-build --no-deps --wait --wait-timeout 120 api web >> "$DEPLOY_STATE/deploy.log" 2>&1; then
      deploy_note 'Update failed; previous application images are running. Review the private deployment log.'
    else
      deploy_note 'Update and application rollback failed. Review the private deployment log and database migration state.'
    fi
  else
    deploy_note 'Update stopped before changing live services. Review the private deployment log.'
  fi
  exit "$result"
}
trap deploy_failed ERR

cd "$DEPLOY_DIR"
[[ $(git rev-parse --abbrev-ref HEAD) == main ]] || { deploy_note 'Automatic deployment requires the main branch.'; exit 1; }
[[ -z $(git status --porcelain) ]] || { deploy_note 'Checkout has local changes; automatic deployment skipped.'; exit 1; }
if ! git fetch --quiet origin main > "$DEPLOY_STATE/fetch.log" 2>&1; then
  deploy_note 'Git fetch failed. Existing services were left running; check the private fetch log.'
  exit 1
fi
DEPLOY_TARGET=$(git rev-parse FETCH_HEAD)
DEPLOY_CURRENT=$(git rev-parse HEAD)
DEPLOY_SUCCESS=$(cat "$DEPLOY_STATE/last-success" 2>/dev/null || true)
if [[ "$DEPLOY_TARGET" == "$DEPLOY_SUCCESS" && "$DEPLOY_CURRENT" == "$DEPLOY_TARGET" && ${CASINO_DEPLOY_FORCE:-0} != 1 ]]; then
  exit 0
fi
if [[ $(cat "$DEPLOY_STATE/last-failed" 2>/dev/null || true) == "$DEPLOY_TARGET" && ${CASINO_DEPLOY_FORCE:-0} != 1 ]]; then
  DEPLOY_LAST=$(cat "$DEPLOY_STATE/last-attempt" 2>/dev/null || printf '0')
  if [[ "$DEPLOY_LAST" =~ ^[0-9]+$ ]] && (( $(date +%s) - DEPLOY_LAST < DEPLOY_RETRY )); then exit 0; fi
fi
git merge-base --is-ancestor "$DEPLOY_CURRENT" "$DEPLOY_TARGET" || { deploy_note 'Main diverged from origin; refusing to overwrite local history.'; exit 1; }

printf 'Target: %s\n' "$DEPLOY_TARGET" > "$DEPLOY_STATE/deploy.log"
compose config > "$DEPLOY_STATE/compose.before.yaml"
DEPLOY_OLD_API=$(docker image inspect casino-api:local --format '{{.Id}}' 2>/dev/null || true)
DEPLOY_OLD_WEB=$(docker image inspect casino-web:local --format '{{.Id}}' 2>/dev/null || true)
deploy_note "Deploying main ${DEPLOY_TARGET:0:12}; build output is in the private deployment log."
git merge --ff-only "$DEPLOY_TARGET" >> "$DEPLOY_STATE/deploy.log" 2>&1
compose build api web >> "$DEPLOY_STATE/deploy.log" 2>&1
compose run --rm --no-deps api /app/check-config -c /app/configs/casino.yml >> "$DEPLOY_STATE/deploy.log" 2>&1

if [[ -n $(compose ps --status running -q mysql) ]]; then
  DEPLOY_BACKUP="$DEPLOY_STATE/backups/$(date -u +%Y%m%dT%H%M%SZ)-${DEPLOY_TARGET:0:12}.sql.gz"
  compose exec -T mysql sh -c 'MYSQL_PWD="$MYSQL_PASSWORD" mysqldump -u casino --single-transaction --no-tablespaces game_casino' | gzip > "$DEPLOY_BACKUP"
  deploy_note 'Saved a private database snapshot before deployment.'
fi

DEPLOY_STARTED=1
compose up -d --no-build --wait --wait-timeout "$DEPLOY_TIMEOUT" >> "$DEPLOY_STATE/deploy.log" 2>&1
DEPLOY_H5_MAPPING=$(compose port web 8080 | head -n 1)
DEPLOY_ADMIN_MAPPING=$(compose port web 8081 | head -n 1)
H5_PORT="${DEPLOY_H5_MAPPING##*:}" ADMIN_PORT="${DEPLOY_ADMIN_MAPPING##*:}" python3 scripts/smoke-test.py >> "$DEPLOY_STATE/deploy.log" 2>&1
printf '%s\n' "$DEPLOY_TARGET" > "$DEPLOY_STATE/last-success.tmp"
mv "$DEPLOY_STATE/last-success.tmp" "$DEPLOY_STATE/last-success"
date -u +%FT%TZ > "$DEPLOY_STATE/last-success-at"
deploy_note "Deployment healthy: ${DEPLOY_TARGET:0:12}."
