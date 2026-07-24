#!/usr/bin/env bash
set -euo pipefail

DIST_DIR=${1:-dist}

if [[ ! -d "$DIST_DIR" ]]; then
  printf 'release asset directory not found: %s\n' "$DIST_DIR" >&2
  exit 1
fi
if [[ ! -f "$DIST_DIR/checksums.txt" ]]; then
  printf 'release checksum manifest not found: %s/checksums.txt\n' "$DIST_DIR" >&2
  exit 1
fi

# Keep this list aligned with .goreleaser.yaml. The release keeps the existing
# Windows targets while guaranteeing both supported macOS and Linux arches.
targets=(
  "darwin amd64 tar.gz pm"
  "darwin arm64 tar.gz pm"
  "linux amd64 tar.gz pm"
  "linux arm64 tar.gz pm"
  "windows amd64 zip pm.exe"
  "windows arm64 zip pm.exe"
)

shopt -s nullglob
assets=()
for target in "${targets[@]}"; do
  read -r goos goarch extension binary_name <<<"$target"
  matches=("$DIST_DIR"/pm_*_"${goos}"_"${goarch}"."${extension}")
  if [[ ${#matches[@]} -ne 1 ]]; then
    printf 'expected one %s/%s release asset, found %d\n' "$goos" "$goarch" "${#matches[@]}" >&2
    printf 'pattern: %s/pm_*_%s_%s.%s\n' "$DIST_DIR" "$goos" "$goarch" "$extension" >&2
    exit 1
  fi

  asset=${matches[0]}
  assets+=("$(basename "$asset")")

  if [[ "$extension" == "zip" ]]; then
    contents=$(unzip -Z1 "$asset" | LC_ALL=C sort)
  else
    contents=$(tar -tzf "$asset" | LC_ALL=C sort)
  fi
  expected=$(printf '%s\n' LICENSE NOTICE README.md "$binary_name" | LC_ALL=C sort)
  if [[ "$contents" != "$expected" ]]; then
    printf 'unexpected archive contents: %s\n' "$asset" >&2
    diff -u <(printf '%s\n' "$expected") <(printf '%s\n' "$contents") || true
    exit 1
  fi
done

expected_names=$(printf '%s\n' "${assets[@]}" | LC_ALL=C sort)
manifest_names=$(
  awk '
    NF != 2 || $1 !~ /^[0-9a-fA-F]{64}$/ || $2 ~ /^\// || $2 ~ /\.\./ || $2 ~ /\\/ { exit 2 }
    { print $2 }
  ' "$DIST_DIR/checksums.txt" | LC_ALL=C sort
) || {
  printf 'invalid checksum manifest format: %s/checksums.txt\n' "$DIST_DIR" >&2
  exit 1
}
if [[ "$manifest_names" != "$expected_names" ]]; then
  printf 'checksum manifest does not cover exactly the expected release assets\n' >&2
  diff -u <(printf '%s\n' "$expected_names") <(printf '%s\n' "$manifest_names") || true
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  (cd "$DIST_DIR" && sha256sum --check checksums.txt)
elif command -v shasum >/dev/null 2>&1; then
  (cd "$DIST_DIR" && shasum -a 256 --check checksums.txt)
else
  printf 'sha256sum or shasum is required to verify release assets\n' >&2
  exit 1
fi

printf 'verified %d release assets in %s\n' "${#assets[@]}" "$DIST_DIR"
