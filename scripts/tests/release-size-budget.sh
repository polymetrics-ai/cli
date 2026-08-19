#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
guard="$repo_root/scripts/verify-release-size-budget.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/archive"
printf 'license\n' > "$tmp/archive/LICENSE"
printf 'notice\n' > "$tmp/archive/NOTICE"
printf 'readme\n' > "$tmp/archive/README.md"
dd if=/dev/zero of="$tmp/archive/pm" bs=1 count=12 status=none
tar -C "$tmp/archive" -czf "$tmp/pm.tar.gz" LICENSE NOTICE README.md pm

output=$("$guard" \
  --archive "$tmp/pm.tar.gz" \
  --binary pm \
  --max-archive-bytes 4096 \
  --max-installed-binary-bytes 12)
expected=$(printf '%s\n' \
  "release-size-report kind=archive subject=$tmp/pm.tar.gz bytes=$(wc -c < "$tmp/pm.tar.gz" | tr -d '[:space:]') budget=4096" \
  "release-size-report kind=installed_binary subject=$tmp/pm.tar.gz!pm bytes=12 budget=12")
if [[ "$output" != "$expected" ]]; then
  printf 'unexpected deterministic size report\n' >&2
  diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$output") || true
  exit 1
fi

quiet_output=$("$guard" \
  --archive "$tmp/pm.tar.gz" \
  --binary pm \
  --max-archive-bytes 4096 \
  --max-installed-binary-bytes 12 \
  --quiet)
if [[ -n "$quiet_output" ]]; then
  printf 'quiet size guard output = %s, want empty\n' "$quiet_output" >&2
  exit 1
fi

if failure=$("$guard" \
  --archive "$tmp/pm.tar.gz" \
  --binary pm \
  --max-archive-bytes 4096 \
  --max-installed-binary-bytes 11 2>&1); then
  printf '%s\n' 'size guard accepted an oversized installed binary' >&2
  exit 1
fi
if [[ "$failure" != *'release size budget exceeded: kind=installed_binary'* ]]; then
  printf 'unexpected size guard failure: %s\n' "$failure" >&2
  exit 1
fi

if missing=$("$guard" --archive 2>&1); then
  printf '%s\n' 'size guard accepted a missing option value' >&2
  exit 1
fi
if [[ "$missing" != *'--archive requires a value'* ]]; then
  printf 'unexpected missing-value failure: %s\n' "$missing" >&2
  exit 1
fi

printf '%s\n' 'release size budget guard passed'
