#!/usr/bin/env bash
# server-setup.sh — one-time server provisioning for Link Shortener.
#
# Run as root or via sudo:
#   sudo bash server-setup.sh
#
# This script is idempotent: safe to re-run without side effects.
# Existing files (especially the environment file) are never overwritten.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

APP_USER=linkshortener
APP_DIR=/opt/linkshortener
ENV_DIR=/etc/linkshortener
ENV_FILE="${ENV_DIR}/environment"
SERVICE_FILE=/etc/systemd/system/linkshortener.service

# ---------- 1. System user --------------------------------------------------
echo "==> Creating system user '${APP_USER}' (skipped if already exists) ..."
if id -u "${APP_USER}" &>/dev/null; then
    echo "    User '${APP_USER}' already exists — skipping."
else
    useradd --system --no-create-home --shell /usr/sbin/nologin "${APP_USER}"
    echo "    User '${APP_USER}' created."
fi

# ---------- 2. Application directories -------------------------------------
echo "==> Creating application directory ${APP_DIR} ..."
mkdir -p "${APP_DIR}/frontend"
chown -R "${APP_USER}:${APP_USER}" "${APP_DIR}"
chmod 750 "${APP_DIR}"

# ---------- 3. Configuration directory + environment file ------------------
echo "==> Creating configuration directory ${ENV_DIR} ..."
mkdir -p "${ENV_DIR}"

if [ -f "${ENV_FILE}" ]; then
    echo "    ${ENV_FILE} already exists — skipping copy to preserve existing settings."
else
    echo "    Copying environment template to ${ENV_FILE} ..."
    cp "${SCRIPT_DIR}/environment.template" "${ENV_FILE}"
    chown root:"${APP_USER}" "${ENV_FILE}"
    chmod 640 "${ENV_FILE}"
    echo "    *** Edit ${ENV_FILE} and replace every CHANGE_ME value before starting the service. ***"
fi

# ---------- 4. Systemd service ---------------------------------------------
echo "==> Installing systemd service ..."
cp "${SCRIPT_DIR}/linkshortener.service" "${SERVICE_FILE}"
systemctl daemon-reload
systemctl enable linkshortener
echo "    Service installed and enabled (not yet started)."

# ---------- Summary --------------------------------------------------------
echo ""
echo "============================================================"
echo "  Server setup complete."
echo "============================================================"
echo ""
echo "NEXT STEPS (complete in order before starting the service):"
echo ""
echo "  1. Fill in the environment file:"
echo "       sudo nano ${ENV_FILE}"
echo "     Replace every CHANGE_ME value."
echo "     Generate secrets with:  openssl rand -hex 32"
echo ""
echo "  2. Apply the database schema (run from your local machine):"
echo "       cat sql/schema.sql | ssh \$SERVER_USER@\$SERVER_HOST \"sudo -u postgres psql linkshortener\""
echo ""
echo "  3. Install the Caddyfile (edit domain name first, then from your local machine):"
echo "       scp deploy/Caddyfile \$SERVER_USER@\$SERVER_HOST:/tmp/Caddyfile"
echo "       ssh \$SERVER_USER@\$SERVER_HOST \"sudo cp /tmp/Caddyfile /etc/caddy/Caddyfile && sudo systemctl reload caddy\""
echo ""
echo "  4. Deploy the application (from your local machine):"
echo "       make deploy SSH_HOST=\$SERVER_USER@\$SERVER_HOST APP_BASE_URL=https://yourdomain.com"
echo ""
echo "  5. Verify the service is running:"
echo "       ssh \$SERVER_USER@\$SERVER_HOST \"sudo systemctl status linkshortener\""
echo ""
