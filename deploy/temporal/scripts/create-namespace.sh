#!/bin/sh
set -eu

NAMESPACE="${DEFAULT_NAMESPACE:-default}"
TEMPORAL_ADDRESS="${TEMPORAL_ADDRESS:-temporal:7233}"
MAX_ATTEMPTS="${TEMPORAL_HEALTH_CHECK_MAX_ATTEMPTS:-30}"
SLEEP_SECONDS="${TEMPORAL_HEALTH_CHECK_SLEEP_SECONDS:-5}"
SERVER_HOST="$(printf '%s' "$TEMPORAL_ADDRESS" | cut -d: -f1)"
SERVER_PORT="$(printf '%s' "$TEMPORAL_ADDRESS" | cut -d: -f2)"

echo "Waiting for Temporal server at ${TEMPORAL_ADDRESS}"
attempt=1
while ! nc -z -w 10 "$SERVER_HOST" "$SERVER_PORT"; do
  if [ "$attempt" -ge "$MAX_ATTEMPTS" ]; then
    echo "Temporal server port did not become available"
    exit 1
  fi
  attempt=$((attempt + 1))
  sleep "$SLEEP_SECONDS"
done

attempt=1
while ! temporal operator cluster health --address "$TEMPORAL_ADDRESS"; do
  if [ "$attempt" -ge "$MAX_ATTEMPTS" ]; then
    echo "Temporal server did not become healthy"
    exit 1
  fi
  attempt=$((attempt + 1))
  sleep "$SLEEP_SECONDS"
done

if temporal operator namespace describe -n "$NAMESPACE" --address "$TEMPORAL_ADDRESS" >/dev/null 2>&1; then
  echo "Temporal namespace '$NAMESPACE' already exists"
else
  temporal operator namespace create -n "$NAMESPACE" --address "$TEMPORAL_ADDRESS"
  echo "Temporal namespace '$NAMESPACE' created"
fi

