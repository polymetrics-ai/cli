#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
assembler="$repo_root/scripts/assemble-release-assets.sh"
verifier="$repo_root/scripts/verify-release-assets.sh"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

version=0.0.0-production-layout
mkdir -p "$tmp/binaries/darwin_arm64"
printf 'installed-pm\n' > "$tmp/binaries/darwin_arm64/pm"
chmod 0755 "$tmp/binaries/darwin_arm64/pm"

SOURCE_DATE_EPOCH=1 "$assembler" \
  --version "$version" \
  --binaries "$tmp/binaries" \
  --out "$tmp/dist" \
  --targets darwin/arm64

archive="$tmp/dist/pm_${version}_darwin_arm64.tar.gz"
contents=$(tar -tzf "$archive" | LC_ALL=C sort)
expected=$(printf '%s\n' LICENSE NOTICE README.md pm | LC_ALL=C sort)
if [[ "$contents" != "$expected" ]]; then
  printf 'assembler did not produce the canonical root archive layout\n' >&2
  diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$contents") || true
  exit 1
fi

output=$("$verifier" "$tmp/dist" --targets darwin/arm64)
if [[ "$output" != *"release-size-report kind=archive subject=$archive"* ]] ||
  [[ "$output" != *"release-size-report kind=installed_binary subject=$archive!pm"* ]] ||
  [[ "$output" != *"verified 1 release assets in $tmp/dist"* ]]; then
  printf 'production archive verifier did not validate the assembled target and its size budget\n' >&2
  printf '%s\n' "$output" >&2
  exit 1
fi

write_impostor_archive() {
  local impostor=$1 digest
  python3 - "$archive" "$impostor" <<'PY'
import io
import sys
import tarfile

archive, impostor = sys.argv[1:]
entries = (
    ("LICENSE", b"license\n"),
    ("NOTICE", b"notice\n"),
    ("README.md", b"readme\n"),
    ("pm", b"installed-pm\n"),
    (impostor, b"impostor\n"),
)
with tarfile.open(archive, "w:gz") as tf:
    for name, data in entries:
        info = tarfile.TarInfo(name)
        info.size = len(data)
        info.mode = 0o755 if name.endswith("pm") else 0o644
        tf.addfile(info, io.BytesIO(data))
PY
  if command -v sha256sum >/dev/null 2>&1; then
    digest=$(sha256sum "$archive" | awk '{ print $1 }')
  else
    digest=$(shasum -a 256 "$archive" | awk '{ print $1 }')
  fi
  printf '%s  %s\n' "$digest" "$(basename "$archive")" > "$tmp/dist/checksums.txt"
}

for impostor in nested/pm ../pm; do
  write_impostor_archive "$impostor"
  if rejected=$("$verifier" "$tmp/dist" --targets darwin/arm64 2>&1); then
    printf 'release verifier accepted %s alongside the root binary\n' "$impostor" >&2
    exit 1
  fi
  if [[ "$rejected" != *'unexpected archive contents'* ]]; then
    printf 'release verifier did not identify the %s archive impostor: %s\n' "$impostor" "$rejected" >&2
    exit 1
  fi
done

printf '%s\n' 'release production layout passed'
