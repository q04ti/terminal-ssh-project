#!/usr/bin/env bash
# Quick health check for AFTERDARK daemon
PORT="${1:-2222}"
HOST="${2:-127.0.0.1}"

if nc -z -w3 "$HOST" "$PORT" 2>/dev/null || (echo > /dev/tcp/"$HOST"/"$PORT") 2>/dev/null; then
    echo "[✓] AFTERDARK SSH port $PORT is reachable on $HOST."
    exit 0
else
    echo "[✗] AFTERDARK SSH port $PORT is NOT responding on $HOST."
    exit 1
fi
