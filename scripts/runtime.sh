#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_DIR="$ROOT_DIR/deploy/compose"
COMPOSE_FILE="$COMPOSE_DIR/polymetrics-runtime.yml"
ENV_FILE="$COMPOSE_DIR/.env"
ENV_EXAMPLE="$COMPOSE_DIR/.env.example"

usage() {
  cat <<'USAGE'
Usage: scripts/runtime.sh <command> [service]

Commands:
  doctor          validate container runtime and compose config
  up              start PostgreSQL, DragonflyDB, and Temporal
  down            stop runtime services
  reset           stop services and remove runtime volumes
  ps              show service status
  logs [service]  stream logs for all services or one service
  config          render compose config

Runtime preference:
  1. podman + podman compose
  2. podman + podman-compose
  3. docker compose
  4. docker-compose
USAGE
}

ensure_env() {
  if [ ! -f "$ENV_FILE" ]; then
    cp "$ENV_EXAMPLE" "$ENV_FILE"
  fi
}

ensure_podman_machine() {
  if [ "$(uname -s)" != "Darwin" ]; then
    return 0
  fi
  if ! podman machine inspect polymetrics-runtime >/dev/null 2>&1; then
    podman machine init polymetrics-runtime
  fi
  podman machine start polymetrics-runtime >/dev/null 2>&1 || true
}

detect_compose() {
  if command -v podman >/dev/null 2>&1; then
    ensure_podman_machine
    if podman compose version >/dev/null 2>&1; then
      COMPOSE_CMD=(podman compose)
      ENGINE="podman"
      return 0
    fi
    if command -v podman-compose >/dev/null 2>&1; then
      COMPOSE_CMD=(podman-compose)
      ENGINE="podman"
      return 0
    fi
    echo "podman is installed, but no compose frontend was found." >&2
    echo "Run scripts/setup-runtime-macos.sh or scripts/setup-runtime-linux.sh, or install podman-compose." >&2
    exit 1
  fi

  if command -v docker >/dev/null 2>&1; then
    if docker compose version >/dev/null 2>&1; then
      COMPOSE_CMD=(docker compose)
      ENGINE="docker"
      return 0
    fi
    if command -v docker-compose >/dev/null 2>&1; then
      COMPOSE_CMD=(docker-compose)
      ENGINE="docker"
      return 0
    fi
  fi

  echo "No supported container runtime found. Install Podman first, or Docker as fallback." >&2
  exit 1
}

compose() {
  ensure_env
  (cd "$COMPOSE_DIR" && "${COMPOSE_CMD[@]}" --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@")
}

main() {
  if [ $# -lt 1 ]; then
    usage
    exit 2
  fi

  detect_compose
  command="$1"
  shift || true

  case "$command" in
    doctor)
      echo "engine=$ENGINE"
      "${COMPOSE_CMD[@]}" version
      compose config >/dev/null
      echo "compose_config=ok"
      ;;
    up)
      compose up -d
      ;;
    down)
      compose down
      ;;
    reset)
      compose down -v
      ;;
    ps|status)
      compose ps
      ;;
    logs)
      compose logs -f "$@"
      ;;
    config)
      compose config
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      echo "unknown command: $command" >&2
      usage
      exit 2
      ;;
  esac
}

main "$@"

