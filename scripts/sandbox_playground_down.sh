#!/usr/bin/env bash
# Stop and remove the Docker playground container.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.sandbox.yml"

if ! command -v docker >/dev/null 2>&1; then
  echo "Error: docker command not found" >&2
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "Error: docker compose plugin not available" >&2
  exit 1
fi

cd "$PROJECT_ROOT"
docker compose -f "$COMPOSE_FILE" --profile playground stop sandbox-playground
docker compose -f "$COMPOSE_FILE" --profile playground rm -f sandbox-playground
