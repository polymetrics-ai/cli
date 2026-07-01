#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\n' "deploy-podman-quadlet: $*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

valid_image_ref() {
  case "$1" in
    ''|*[!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._:/@+-]*)
      return 1
      ;;
  esac
}

valid_unit_name() {
  case "$1" in
    ''|*/*|*[!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.@-]*)
      return 1
      ;;
  esac
}

image_repo() {
  ref="${1%@*}"
  case "$ref" in
    *:*) printf '%s\n' "${ref%:*}" ;;
    *) printf '%s\n' "$ref" ;;
  esac
}

resolve_digest() {
  image="$1"
  repo="$(image_repo "$image")"
  podman image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$image" |
    awk -v prefix="$repo@sha256:" '
      index($0, prefix) == 1 { print; found = 1; exit }
      !first && /@sha256:/ { first = $0 }
      END { if (!found && first) print first }
    '
}

set_user_systemd_env() {
  uid="$(id -u)"
  export XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$uid}"
  if [ -S "$XDG_RUNTIME_DIR/bus" ]; then
    export DBUS_SESSION_BUS_ADDRESS="${DBUS_SESSION_BUS_ADDRESS:-unix:path=$XDG_RUNTIME_DIR/bus}"
  fi
}

update_quadlet_image() {
  quadlet="$1"
  digest="$2"
  tmp="$(mktemp)"
  if ! awk -v image="$digest" '
    BEGIN { replaced = 0 }
    /^Image=/ && replaced == 0 {
      print "Image=" image
      replaced = 1
      next
    }
    { print }
    END {
      if (replaced != 1) {
        exit 42
      }
    }
  ' "$quadlet" > "$tmp"; then
    rm -f "$tmp"
    fail "could not replace Image= in $quadlet"
  fi
  install -m 0644 "$tmp" "$quadlet"
  rm -f "$tmp"
}

wait_for_http() {
  url="$1"
  timeout="${2:-120s}"
  seconds="${timeout%s}"
  case "$seconds" in
    ''|*[!0-9]*) seconds=120 ;;
  esac

  attempts=$((seconds / 3))
  [ "$attempts" -gt 0 ] || attempts=1

  i=1
  while [ "$i" -le "$attempts" ]; do
    if curl -fsS --max-time 5 "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 3
    i=$((i + 1))
  done

  return 1
}

need_cmd awk
need_cmd curl
need_cmd install
need_cmd podman
need_cmd systemctl

set_user_systemd_env

IMAGE="${1:-${WEBSITE_IMAGE:-}}"
REGISTRY_HOST="${REGISTRY:-ghcr.io}"
REGISTRY_USERNAME="${REGISTRY_USERNAME:-${GITHUB_ACTOR:-}}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-${GITHUB_TOKEN:-}}"
SERVICE="${WEBSITE_SERVICE:-cli-polymetrics}"
QUADLET="${WEBSITE_QUADLET:-$HOME/.config/containers/systemd/cli-polymetrics.container}"
HEALTH_URL="${WEBSITE_HEALTH_URL:-http://127.0.0.1:18081/}"
PUBLIC_URL="${WEBSITE_PUBLIC_URL:-https://cli.polymetrics.ai/}"
ROLLOUT_TIMEOUT="${WEBSITE_ROLLOUT_TIMEOUT:-120s}"

valid_image_ref "$IMAGE" || fail "WEBSITE_IMAGE must be a container image tag or digest"
valid_unit_name "$SERVICE" || fail "WEBSITE_SERVICE is not a safe systemd unit name"
[ -f "$QUADLET" ] || fail "Quadlet file not found: $QUADLET"
[ -w "$QUADLET" ] || fail "Quadlet file is not writable: $QUADLET"

if [ -n "$REGISTRY_USERNAME" ] && [ -n "$REGISTRY_PASSWORD" ]; then
  printf '%s' "$REGISTRY_PASSWORD" | podman login "$REGISTRY_HOST" \
    --username "$REGISTRY_USERNAME" \
    --password-stdin >/dev/null
fi

podman pull "$IMAGE"
DIGEST="$(resolve_digest "$IMAGE")"
case "$DIGEST" in
  *@sha256:*) ;;
  *) fail "could not resolve a digest for $IMAGE" ;;
esac

backup="$QUADLET.backup.$(date -u +%Y%m%d%H%M%S)"
cp "$QUADLET" "$backup"
update_quadlet_image "$QUADLET" "$DIGEST"

systemctl --user daemon-reload
systemctl --user restart "$SERVICE"

if ! wait_for_http "$HEALTH_URL" "$ROLLOUT_TIMEOUT"; then
  systemctl --user status "$SERVICE" --no-pager || true
  cp "$backup" "$QUADLET"
  systemctl --user daemon-reload
  systemctl --user restart "$SERVICE" || true
  fail "health check failed for $HEALTH_URL; restored previous Quadlet image"
fi

if [ -n "$PUBLIC_URL" ]; then
  curl -fsSI --max-time 20 "$PUBLIC_URL" >/dev/null
fi

printf '%s\n' "deployed $DIGEST to $SERVICE"
