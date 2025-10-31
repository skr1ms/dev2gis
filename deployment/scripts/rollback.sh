#!/usr/bin/env bash
set -Eeuo pipefail

# rollback.sh --env prod|dev
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="$BASE_DIR/deployment/logs"
LAST="$LOG_DIR/last-good.txt"

ENV_NAME=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --env) ENV_NAME="$2"; shift 2;;
    *) echo "Unknown arg $1"; exit 1;;
  esac
done
[[ -z "$ENV_NAME" ]] && { echo "Need --env prod|dev"; exit 2; }
[[ -f "$LAST" ]] || { echo "No last-good.txt"; exit 3; }

COMPOSE_FILE_PROD="$BASE_DIR/deployment/docker/docker-compose.yml"
COMPOSE_FILE_DEV="$BASE_DIR/deployment/docker/docker-compose.dev.yml"
ENV_FILE="$BASE_DIR/.env.$ENV_NAME"
[[ -f "$ENV_FILE" ]] || { echo "Missing env file $ENV_FILE"; exit 4; }

COMPOSE_FILE="$COMPOSE_FILE_PROD"
[[ "$ENV_NAME" == "dev" ]] && COMPOSE_FILE="$COMPOSE_FILE_DEV"

ROLLBACK_COMMIT=$(grep '^commit=' "$LAST" | cut -d= -f2)
echo "[RB] commit=$ROLLBACK_COMMIT"
export IMAGE_TAG="$ROLLBACK_COMMIT"
set -a; source "$ENV_FILE"; set +a

SERVICES=($(grep -v '^commit=' "$LAST" | cut -d= -f1))
echo "[RB] services=(${SERVICES[*]})"

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull "${SERVICES[@]}" || true
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d "${SERVICES[@]}"

echo "[RB] Done."