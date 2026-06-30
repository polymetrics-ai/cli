#!/usr/bin/env sh
set -eu

fail() {
  printf '%s\n' "deploy-image: $*" >&2
  exit 1
}

require_name() {
  name="$1"
  value="$2"

  case "$value" in
    ''|*[!abcdefghijklmnopqrstuvwxyz0123456789-]*|-*|*-)
      fail "$name must be a Kubernetes DNS label"
      ;;
  esac
}

require_image() {
  case "$1" in
    ''|*[!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._:/@+-]*)
      fail "WEBSITE_IMAGE contains unsupported characters"
      ;;
  esac
}

require_timeout() {
  case "$1" in
    ''|*[!0123456789smh]*)
      fail "WEBSITE_ROLLOUT_TIMEOUT must use kubectl duration units"
      ;;
  esac
}

command -v kubectl >/dev/null 2>&1 || fail "kubectl is required on the deploy runner"

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

IMAGE="${1:-${WEBSITE_IMAGE:-}}"
NAMESPACE="${WEBSITE_NAMESPACE:-polymetrics-website}"
DEPLOYMENT="${WEBSITE_DEPLOYMENT:-polymetrics-website}"
CONTAINER="${WEBSITE_CONTAINER:-website}"
ROLLOUT_TIMEOUT="${WEBSITE_ROLLOUT_TIMEOUT:-120s}"
IMAGE_PULL_SECRET="${WEBSITE_IMAGE_PULL_SECRET:-}"

require_image "$IMAGE"
require_name "WEBSITE_NAMESPACE" "$NAMESPACE"
require_name "WEBSITE_DEPLOYMENT" "$DEPLOYMENT"
require_name "WEBSITE_CONTAINER" "$CONTAINER"
require_timeout "$ROLLOUT_TIMEOUT"

if [ -n "${IMAGE_PULL_SECRET:-}" ]; then
  require_name "WEBSITE_IMAGE_PULL_SECRET" "$IMAGE_PULL_SECRET"
fi

if [ -n "${WEBSITE_KUBE_CONTEXT:-}" ]; then
  kubectl config use-context "$WEBSITE_KUBE_CONTEXT"
fi

kubectl apply -f "$SCRIPT_DIR/namespace.yaml"
kubectl apply -f "$SCRIPT_DIR/deployment.yaml"
kubectl apply -f "$SCRIPT_DIR/service.yaml"

kubectl -n "$NAMESPACE" set image "deployment/$DEPLOYMENT" "$CONTAINER=$IMAGE"

policy_patch=$(printf '{"spec":{"template":{"spec":{"containers":[{"name":"%s","imagePullPolicy":"IfNotPresent"}]}}}}' "$CONTAINER")
kubectl -n "$NAMESPACE" patch deployment "$DEPLOYMENT" --type strategic --patch "$policy_patch"

if [ -n "${IMAGE_PULL_SECRET:-}" ]; then
  pull_patch=$(printf '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"%s"}]}}}}' "$IMAGE_PULL_SECRET")
  kubectl -n "$NAMESPACE" patch deployment "$DEPLOYMENT" --type strategic --patch "$pull_patch"
fi

kubectl -n "$NAMESPACE" rollout status "deployment/$DEPLOYMENT" --timeout="$ROLLOUT_TIMEOUT"
