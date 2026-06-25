#!/usr/bin/env bash
# Run on the remote host: ./deploy.sh [branch]
set -euo pipefail

BRANCH="${1:-main}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Pulling latest from branch: $BRANCH"
cd "$SERVER_DIR/.."
git fetch origin
git checkout "$BRANCH"
git pull origin "$BRANCH"

echo "==> Stopping server"
cd "$SERVER_DIR"
docker compose down

echo "==> Syncing mods"
docker compose run --rm mod-sync

echo "==> Starting server"
docker compose up -d minecraft

echo "==> Server started. Logs:"
docker compose logs -f --tail=30 minecraft
