#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if [[ -n "$(go env GOWORK)" ]]; then
  echo "FAIL: a go.work file would couple the standalone Shepherd module" >&2
  exit 1
fi

if go list ./... | grep -q 'agent-runtime/shepherd'; then
  echo "FAIL: root package graph includes Shepherd" >&2
  exit 1
fi

if go list -deps ./cmd/pm | grep -q 'agent-runtime/shepherd'; then
  echo "FAIL: pm dependency graph includes Shepherd" >&2
  exit 1
fi

echo "PASS: Shepherd is excluded from the root module and pm binary"

