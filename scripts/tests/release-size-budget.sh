#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
guard="$repo_root/scripts/verify-release-size-budget.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

make_archive() {
  local archive=$1 kind=$2
  python3 - "$archive" "$kind" <<'PY'
import io
import os
import sys
import tarfile

archive, kind = sys.argv[1:]
entries = [
    ("LICENSE", b"license\n"),
    ("NOTICE", b"notice\n"),
    ("README.md", b"readme\n"),
]
pm = b"012345678901"
if kind == "canonical":
    entries.append(("pm", pm))
elif kind == "oversized-archive":
    entries.extend((("pm", pm), ("padding", os.urandom(8192))))
elif kind == "missing-pm":
    pass
elif kind == "duplicate-pm":
    entries.extend((("pm", pm), ("pm", pm)))
elif kind == "nested-impostor":
    entries.append(("nested/pm", pm))
elif kind == "traversal-impostor":
    entries.append(("../pm", pm))
else:
    raise SystemExit(f"unknown archive fixture {kind}")

with tarfile.open(archive, "w:gz") as tf:
    for name, data in entries:
        info = tarfile.TarInfo(name)
        info.size = len(data)
        info.mode = 0o755 if name.endswith("pm") else 0o644
        tf.addfile(info, io.BytesIO(data))
PY
}

expect_failure() {
  local label=$1 want=$2 output
  shift 2
  if output=$("$@" 2>&1); then
    printf '%s\n' "$label unexpectedly succeeded" >&2
    exit 1
  fi
  if [[ "$output" != *"$want"* ]]; then
    printf '%s unexpected failure: %s\n' "$label" "$output" >&2
    exit 1
  fi
}

canonical="$tmp/pm.tar.gz"
make_archive "$canonical" canonical

output=$("$guard" \
  --archive "$canonical" \
  --binary pm \
  --max-archive-bytes 4096 \
  --max-installed-binary-bytes 12)
expected=$(printf '%s\n' \
  "release-size-report kind=archive subject=$canonical bytes=$(wc -c < "$canonical" | tr -d '[:space:]') budget=4096" \
  "release-size-report kind=installed_binary subject=$canonical!pm bytes=12 budget=12")
if [[ "$output" != "$expected" ]]; then
  printf 'unexpected deterministic size report\n' >&2
  diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$output") || true
  exit 1
fi

quiet_output=$("$guard" \
  --archive "$canonical" \
  --binary pm \
  --max-archive-bytes 4096 \
  --max-installed-binary-bytes 12 \
  --quiet)
if [[ -n "$quiet_output" ]]; then
  printf 'quiet size guard output = %s, want empty\n' "$quiet_output" >&2
  exit 1
fi

expect_failure 'oversized installed binary' 'release size budget exceeded: kind=installed_binary' \
  "$guard" --archive "$canonical" --binary pm --max-archive-bytes 4096 --max-installed-binary-bytes 11
expect_failure 'quiet oversized installed binary' 'release size budget exceeded: kind=installed_binary' \
  "$guard" --archive "$canonical" --binary pm --max-archive-bytes 4096 --max-installed-binary-bytes 11 --quiet

oversized_archive="$tmp/oversized-archive.tar.gz"
make_archive "$oversized_archive" oversized-archive
expect_failure 'oversized archive' 'release size budget exceeded: kind=archive' \
  "$guard" --archive "$oversized_archive" --binary pm --max-archive-bytes 1 --max-installed-binary-bytes 12

for case_name in missing-pm duplicate-pm nested-impostor traversal-impostor; do
  archive="$tmp/$case_name.tar.gz"
  make_archive "$archive" "$case_name"
  case "$case_name" in
    duplicate-pm) expected_count=2 ;;
    *) expected_count=0 ;;
  esac
  expect_failure "$case_name" "contains $expected_count entries named pm, want exactly one" \
    "$guard" --archive "$archive" --binary pm --max-archive-bytes 4096 --max-installed-binary-bytes 12
done

expect_failure 'missing option value' '--archive requires a value' "$guard" --archive

printf '%s\n' 'release size budget guard passed'
