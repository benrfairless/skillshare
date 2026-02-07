#!/usr/bin/env bash
# Enter the running Docker playground or execute a command inside it.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.sandbox.yml"
SERVICE="sandbox-playground"

if ! command -v docker >/dev/null 2>&1; then
  echo "Error: docker command not found" >&2
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "Error: docker compose plugin not available" >&2
  exit 1
fi

cd "$PROJECT_ROOT"

if [[ -z "$(docker compose -f "$COMPOSE_FILE" --profile playground ps -q "$SERVICE")" ]]; then
  echo "Playground is not running. Start it first:"
  echo "  ./scripts/sandbox_playground_up.sh"
  exit 1
fi

if [[ $# -gt 0 ]]; then
  docker compose -f "$COMPOSE_FILE" --profile playground exec --user "$(id -u):$(id -g)" "$SERVICE" bash -c "$*"
else
  docker compose -f "$COMPOSE_FILE" --profile playground exec --user "$(id -u):$(id -g)" "$SERVICE" bash
fi
