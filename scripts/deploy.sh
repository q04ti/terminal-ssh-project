#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# AFTERDARK — Deployment Script for Linux / VPS / Hack Club Nest
# ==============================================================================

APP_USER="afterdark"
INSTALL_BIN="/usr/local/bin/afterdark"
DATA_DIR="/var/lib/afterdark/data"
SERVICE_SRC="./scripts/afterdark.service"
SERVICE_DEST="/etc/systemd/system/afterdark.service"
PORT="${PORT:-2222}"

echo "[*] Deploying AFTERDARK SSH service on port ${PORT}..."

# 1. Create dedicated system user if not exists
if ! id "${APP_USER}" &>/dev/null; then
    echo "[+] Creating system user ${APP_USER}..."
    useradd --system --shell /usr/sbin/nologin --home-dir /var/lib/afterdark "${APP_USER}" || true
fi

# 2. Create persistent data directory
echo "[+] Ensuring data directory exists at ${DATA_DIR}..."
mkdir -p "${DATA_DIR}"
chown -R "${APP_USER}:${APP_USER}" /var/lib/afterdark
chmod 700 "${DATA_DIR}"

# 3. Build production binary
echo "[+] Compiling static Go binary..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o ./afterdark ./cmd/afterdark

# 4. Install binary (keep backup for rollback)
if [[ -f "${INSTALL_BIN}" ]]; then
    echo "[+] Backing up previous binary for rollback: ${INSTALL_BIN}.bak"
    cp "${INSTALL_BIN}" "${INSTALL_BIN}.bak"
fi
cp ./afterdark "${INSTALL_BIN}"
chmod 755 "${INSTALL_BIN}"

# 5. Install systemd service
echo "[+] Configuring systemd service..."
cp "${SERVICE_SRC}" "${SERVICE_DEST}"
systemctl daemon-reload
systemctl enable afterdark
systemctl restart afterdark

# 6. Verify health / port status
echo "[+] Waiting for service startup..."
sleep 2

if systemctl is-active --quiet afterdark; then
    echo "[✓] AFTERDARK is active and running!"
    systemctl status afterdark --no-pager -l
else
    echo "[!] Deployment error: afterdark failed to start. Rolling back..."
    if [[ -f "${INSTALL_BIN}.bak" ]]; then
        cp "${INSTALL_BIN}.bak" "${INSTALL_BIN}"
        systemctl restart afterdark
    fi
    exit 1
fi

echo ""
echo "=========================================================="
echo "Deployment successful!"
echo "Connect with: ssh <SERVER_IP_OR_DOMAIN> -p ${PORT}"
echo "Inspect logs: journalctl -u afterdark -f"
echo "=========================================================="
