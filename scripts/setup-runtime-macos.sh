#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

have() {
  command -v "$1" >/dev/null 2>&1
}

echo "Setting up Polymetrics runtime dependencies on macOS"

if have podman; then
  echo "podman=found"
else
  if have brew; then
    echo "Installing Podman with Homebrew"
    brew install podman
  elif have docker; then
    echo "podman=missing; docker=found. Runtime scripts will use Docker fallback."
  else
    echo "Neither Podman nor Docker is installed, and Homebrew is unavailable." >&2
    echo "Install Podman Desktop or Homebrew, then rerun this script." >&2
    exit 1
  fi
fi

if have podman; then
  if ! podman compose version >/dev/null 2>&1 && ! have podman-compose; then
    if have brew; then
      echo "Installing podman-compose with Homebrew"
      brew install podman-compose
    else
      echo "podman compose frontend missing. Install podman-compose or Docker Compose." >&2
      exit 1
    fi
  fi

  if ! podman machine inspect polymetrics-runtime >/dev/null 2>&1; then
    echo "Creating Podman machine polymetrics-runtime"
    podman machine init polymetrics-runtime
  fi
  echo "Starting Podman machine polymetrics-runtime"
  podman machine start polymetrics-runtime >/dev/null 2>&1 || true
fi

"$ROOT_DIR/scripts/runtime.sh" doctor

cat <<'NEXT'

Runtime setup complete.

Start services:
  scripts/runtime.sh up

Stop services:
  scripts/runtime.sh down
NEXT

