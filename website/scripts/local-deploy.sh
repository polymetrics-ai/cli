#!/usr/bin/env bash
# Build with Podman, load into minikube, deploy to k8s, port-forward.
set -euo pipefail

CLUSTER_NAME="polymetrics-website"
IMAGE_NAME="polymetrics-website:latest"
NAMESPACE="polymetrics-website"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEBSITE_DIR="$(dirname "$SCRIPT_DIR")"

# ── 1. build ──────────────────────────────────────────────────────────────────
echo "→ Building image with Podman..."
podman build -t "$IMAGE_NAME" "$WEBSITE_DIR"

# ── 2. cluster ────────────────────────────────────────────────────────────────
MINIKUBE_STATUS=$(minikube status --profile "$CLUSTER_NAME" -o json 2>/dev/null \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('Host','Stopped'))" \
  2>/dev/null || echo "Stopped")

if [ "$MINIKUBE_STATUS" = "Running" ]; then
  echo "→ Minikube profile '$CLUSTER_NAME' already running"
else
  echo "→ Starting minikube (driver=podman, profile=$CLUSTER_NAME)..."
  minikube start \
    --driver=podman \
    --profile="$CLUSTER_NAME" \
    --cpus=2 \
    --memory=2g \
    --container-runtime=containerd
fi

# Switch kubectl context
kubectl config use-context "$CLUSTER_NAME"

# ── 3. load image ─────────────────────────────────────────────────────────────
echo "→ Loading image into minikube..."
# Save from Podman, load into minikube
TMP_TAR=$(mktemp /tmp/polymetrics-website.XXXXXX.tar)
trap 'rm -f "$TMP_TAR"' EXIT
podman save "$IMAGE_NAME" -o "$TMP_TAR"
minikube image load "$TMP_TAR" --profile "$CLUSTER_NAME"

# ── 4. apply manifests ────────────────────────────────────────────────────────
echo "→ Applying Kubernetes manifests..."
kubectl apply -f "$WEBSITE_DIR/deploy/namespace.yaml"
kubectl apply -f "$WEBSITE_DIR/deploy/deployment.yaml"
kubectl apply -f "$WEBSITE_DIR/deploy/service.yaml"

# ── 5. rollout ────────────────────────────────────────────────────────────────
echo "→ Waiting for rollout..."
kubectl rollout status deployment/polymetrics-website \
  -n "$NAMESPACE" --timeout=120s

echo ""
echo "✓ Deployed! Port-forwarding → http://localhost:3000"
echo "  (Ctrl-C to stop)"
kubectl port-forward -n "$NAMESPACE" svc/polymetrics-website 3000:3000
