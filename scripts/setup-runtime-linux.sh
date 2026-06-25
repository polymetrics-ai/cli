#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

have() {
  command -v "$1" >/dev/null 2>&1
}

run_sudo() {
  if [ "$(id -u)" -eq 0 ]; then
    "$@"
  else
    sudo "$@"
  fi
}

install_podman() {
  if have podman; then
    echo "podman=found"
    return 0
  fi

  echo "podman=missing; attempting package-manager install"

  if have apt-get; then
    run_sudo apt-get update
    run_sudo apt-get install -y podman podman-compose
  elif have dnf; then
    run_sudo dnf install -y podman podman-compose
  elif have yum; then
    run_sudo yum install -y podman podman-compose
  elif have zypper; then
    run_sudo zypper --non-interactive install podman podman-compose
  elif have pacman; then
    run_sudo pacman -Syu --noconfirm podman podman-compose
  elif have apk; then
    run_sudo apk add podman podman-compose
  else
    echo "No supported package manager found for automatic Podman install." >&2
    return 1
  fi
}

echo "Setting up Polymetrics runtime dependencies on Linux"

if ! install_podman; then
  if have docker; then
    echo "podman install unavailable; docker=found. Runtime scripts will use Docker fallback."
  else
    echo "Neither Podman nor Docker is available." >&2
    echo "Install Podman first, or Docker as fallback, then rerun this script." >&2
    exit 1
  fi
fi

if have podman; then
  if ! podman compose version >/dev/null 2>&1 && ! have podman-compose; then
    echo "podman is installed, but no compose frontend was found." >&2
    echo "Install podman-compose or a Podman package that provides 'podman compose'." >&2
    exit 1
  fi
fi

"$ROOT_DIR/scripts/runtime.sh" doctor

cat <<'NEXT'

Runtime setup complete.

Start services:
  scripts/runtime.sh up

Stop services:
  scripts/runtime.sh down
NEXT

