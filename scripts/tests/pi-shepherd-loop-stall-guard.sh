#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/pi-shepherd-stall-guard.XXXXXX")"
trap 'rm -rf "$tmp_dir"' EXIT

AUTO_LOOP_STATE_DIR="$tmp_dir/state" \
SHEPHERD_STALL_GUARD_SELF_TEST=1 \
STALL_MINUTES=1 \
"$repo_root/scripts/pi-shepherd-loop.sh"
