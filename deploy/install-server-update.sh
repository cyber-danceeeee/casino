#!/usr/bin/env bash
# Install on the Linux deployment server, after cloning into /opt/casino.
set -Eeuo pipefail
umask 077
[[ $(id -u) == 0 ]] || { echo 'Run this installer as root on the deployment server.' >&2; exit 1; }
for dependency in git docker python3 flock systemctl gzip; do
  command -v "$dependency" >/dev/null || { echo "Missing dependency: $dependency" >&2; exit 1; }
done
[[ -d /opt/casino/.git && -f /opt/casino/.env ]] || { echo 'Prepare /opt/casino and its private .env first.' >&2; exit 1; }
DEPLOY_INSTALL_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
install -d -m 700 /var/lib/casino-update
install -m 644 "$DEPLOY_INSTALL_DIR/systemd/casino-update.service" /etc/systemd/system/casino-update.service
install -m 644 "$DEPLOY_INSTALL_DIR/systemd/casino-update.timer" /etc/systemd/system/casino-update.timer
if [[ ! -f /etc/casino-update.env ]]; then
  cat > /etc/casino-update.env <<'ENV'
CASINO_PROJECT_DIR=/opt/casino
CASINO_DEPLOY_STATE_DIR=/var/lib/casino-update
CASINO_COMPOSE_PROJECT=casino
CASINO_DEPLOY_RETRY_SECONDS=300
CASINO_DEPLOY_HEALTH_TIMEOUT=600
ENV
  chmod 600 /etc/casino-update.env
fi
systemctl daemon-reload
systemctl enable --now casino-update.timer
echo 'Automatic updates enabled. Push to main; the server checks every minute.'
echo 'Status: systemctl status casino-update.timer'
echo 'History: journalctl -u casino-update.service -n 50 --no-pager'
echo 'Private build log: /var/lib/casino-update/deploy.log'
