#!/usr/bin/env bash
# Assemble release archives, Linux packages and the checksum manifest from
# binaries that were each built natively on their own runner.
#
# Why this exists rather than one GoReleaser invocation: pm embeds DuckDB, so
# every target is a cgo build. cgo cannot be cross-compiled from a single host
# without a toolchain that can link go-duckdb's prebuilt C++ archives, and
# GoReleaser OSS cannot assemble archives from binaries built elsewhere (its
# `prebuilt` builder is Pro-only). So the build fans out to native runners and
# this script does the assembly step GoReleaser would otherwise do.
#
# It deliberately produces byte-identical output for a given commit: archives
# are written with a fixed mtime, sorted entries and numeric root ownership, so
# SOURCE_DATE_EPOCH keeps releases reproducible exactly as GoReleaser's
# mod_timestamp did.
#
# The file names here are a contract with scripts/verify-release-assets.sh,
# which independently re-derives the expected set and fails on any mismatch.
set -euo pipefail

usage() {
  cat >&2 <<'MSG'
usage: assemble-release-assets.sh --version <x.y.z> --binaries <dir> --out <dir>
                                  [--targets <goos/goarch,...>]

  --version   release version without a leading v
  --binaries  directory holding <goos>_<goarch>/pm[.exe] from the build matrix
  --out       directory to write archives, packages and checksums.txt into
  --targets   assemble only these targets instead of the full release set.
              Pull requests use this to exercise the Linux packaging path on one
              cheap runner; a release always assembles every target.
MSG
  exit 2
}

VERSION=""
BINARIES=""
OUT=""
TARGETS=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="${2:-}"; shift 2 ;;
    --binaries) BINARIES="${2:-}"; shift 2 ;;
    --out) OUT="${2:-}"; shift 2 ;;
    --targets) TARGETS="${2:-}"; shift 2 ;;
    -h|--help) usage ;;
    *) printf 'unknown argument: %s\n' "$1" >&2; usage ;;
  esac
done
[[ -n "$VERSION" && -n "$BINARIES" && -n "$OUT" ]] || usage

if [[ ! "$VERSION" =~ ^[A-Za-z0-9][A-Za-z0-9._+-]*$ ]]; then
  printf 'invalid release version: %s\n' "$VERSION" >&2
  exit 2
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$OUT"

# A fixed timestamp keeps archives reproducible. The release workflow exports
# SOURCE_DATE_EPOCH from the commit date; fall back to it here so a local run
# behaves the same way.
: "${SOURCE_DATE_EPOCH:=$(cd "$REPO_ROOT" && git log -1 --format=%ct)}"
export SOURCE_DATE_EPOCH

# Windows is absent on purpose, both architectures. windows/arm64 was never
# buildable — go-duckdb ships no library for it — and windows/amd64 was dropped
# because pm has no Windows customer; it comes back on a customer ask, from git
# history. Keep this list in step with archive_targets in
# scripts/verify-release-assets.sh and the goos/goarch matrix in release.yml;
# scripts/tests/release-target-parity.sh asserts they agree.
ARCHIVE_TARGETS=(
  "darwin amd64 tar.gz pm"
  "darwin arm64 tar.gz pm"
  "linux amd64 tar.gz pm"
  "linux arm64 tar.gz pm"
)

# deb keeps Go's arch names; rpm uses its own. Both are what
# verify-release-assets.sh expects to find.
PACKAGE_TARGETS=(
  "deb amd64 amd64"
  "deb arm64 arm64"
  "rpm amd64 x86_64"
  "rpm arm64 aarch64"
)

# selected reports whether a target is in --targets, and is always true when no
# filter was given.
selected() {
  local goos=$1 goarch=$2 entry
  [[ -z "$TARGETS" ]] && return 0
  IFS=',' read -ra entries <<<"$TARGETS"
  for entry in "${entries[@]}"; do
    [[ "$entry" == "$goos/$goarch" ]] && return 0
  done
  return 1
}

# Reproducible archives need GNU tar's --sort/--mtime/--numeric-owner. bsdtar,
# which is `tar` on macOS, has none of them. Releases assemble on Linux where
# `tar` is GNU tar; this also finds `gtar` so a local dry run on macOS produces
# byte-identical output rather than quietly producing a different archive.
resolve_gnu_tar() {
  local candidate
  for candidate in tar gtar gnutar; do
    if command -v "$candidate" >/dev/null && "$candidate" --version 2>/dev/null | head -1 | grep -q 'GNU tar'; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done
  return 1
}

stage_dir="$(mktemp -d)"
trap 'rm -rf "$stage_dir"' EXIT

printf 'assembling pm %s from %s\n' "$VERSION" "$BINARIES"

for target in "${ARCHIVE_TARGETS[@]}"; do
  read -r goos goarch extension binary <<<"$target"
  selected "$goos" "$goarch" || continue
  source_binary="$BINARIES/${goos}_${goarch}/${binary}"
  if [[ ! -f "$source_binary" ]]; then
    printf 'missing %s binary for %s/%s at %s\n' "$binary" "$goos" "$goarch" "$source_binary" >&2
    printf 'every target must be built before assembly; nothing is substituted\n' >&2
    exit 1
  fi

  work="$stage_dir/${goos}_${goarch}"
  mkdir -p "$work"
  install -m 0755 "$source_binary" "$work/$binary"
  install -m 0644 "$REPO_ROOT/LICENSE" "$work/LICENSE"
  install -m 0644 "$REPO_ROOT/NOTICE" "$work/NOTICE"
  install -m 0644 "$REPO_ROOT/README.md" "$work/README.md"
  touch -d "@$SOURCE_DATE_EPOCH" "$work"/* 2>/dev/null || \
    TZ=UTC touch -t "$(TZ=UTC date -r "$SOURCE_DATE_EPOCH" +%Y%m%d%H%M.%S)" "$work"/*

  archive="$OUT/pm_${VERSION}_${goos}_${goarch}.${extension}"
  rm -f "$archive"
  # A failure mid-write must not leave a short archive behind that a later step
  # could checksum and publish.
  trap 'rm -f "$archive"; rm -rf "$stage_dir"' EXIT
  case "$extension" in
    tar.gz)
      # --sort, numeric root ownership and a fixed mtime are what make the
      # archive byte-identical across runs; gzip -n drops its own timestamp.
      if ! gnu_tar="$(resolve_gnu_tar)"; then
        printf 'GNU tar is required for reproducible archives and was not found.\n' >&2
        printf 'On macOS: brew install gnu-tar. Releases assemble on Linux, where tar is GNU tar.\n' >&2
        exit 1
      fi
      "$gnu_tar" --sort=name \
          --owner=0 --group=0 --numeric-owner \
          --mtime="@$SOURCE_DATE_EPOCH" \
          -C "$work" -cf - . | gzip -9n > "$archive"
      ;;
    *)
      printf 'unsupported archive format: %s\n' "$extension" >&2
      exit 1
      ;;
  esac
  trap 'rm -rf "$stage_dir"' EXIT
  printf '  archive %s\n' "$(basename "$archive")"
done

if [[ -n "$TARGETS" ]] && ! selected linux amd64 && ! selected linux arm64; then
  printf 'no Linux targets selected; skipping package assembly\n'
elif ! command -v nfpm >/dev/null; then
  printf 'nfpm was not found on PATH; install it before assembling Linux packages.\n' >&2
  printf '  go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.43.0\n' >&2
  printf '  export PATH="$(go env GOPATH)/bin:$PATH"\n' >&2
  printf 'The PATH line matters: go install writes to GOPATH/bin, which is not on\n' >&2
  printf 'PATH by default in this repository CI jobs. GoReleaser used to provide\n' >&2
  printf 'nfpm implicitly, so nothing had to install it before.\n' >&2
  exit 1
fi

for target in "${PACKAGE_TARGETS[@]}"; do
  read -r format goarch package_arch <<<"$target"
  selected linux "$goarch" || continue
  source_binary="$BINARIES/linux_${goarch}/pm"
  if [[ ! -f "$source_binary" ]]; then
    printf 'missing linux/%s binary at %s\n' "$goarch" "$source_binary" >&2
    exit 1
  fi
  package="$OUT/pm_${VERSION}_linux_${package_arch}.${format}"
  rm -f "$package"
  # The config is rendered here rather than relying on nfpm expanding
  # environment variables itself: it does not expand them inside a `src` glob,
  # which fails as `glob failed: ${PM_PACKAGE_BINARY}` — a packaging step that
  # errors rather than silently shipping an empty package, but still a failure.
  rendered="$stage_dir/nfpm_${format}_${goarch}.yaml"
  sed -e "s|\${PM_PACKAGE_ARCH}|$goarch|g" \
      -e "s|\${PM_PACKAGE_VERSION}|$VERSION|g" \
      -e "s|\${PM_PACKAGE_BINARY}|$source_binary|g" \
      "$REPO_ROOT/packaging/linux/nfpm.yaml" > "$rendered"
  if grep -q '\${PM_PACKAGE_' "$rendered"; then
    printf 'unsubstituted placeholder left in %s\n' "$rendered" >&2
    grep -n '\${PM_PACKAGE_' "$rendered" >&2
    exit 1
  fi
  nfpm package --config "$rendered" --packager "$format" --target "$package"
  printf '  package %s\n' "$(basename "$package")"
done

# One manifest over every asset, in the format verify-release-assets.sh checks
# with `sha256sum --check`.
(
  cd "$OUT"
  rm -f checksums.txt
  shopt -s nullglob
  assets=(pm_*.tar.gz pm_*.zip pm_*.deb pm_*.rpm)
  shopt -u nullglob
  if [[ ${#assets[@]} -eq 0 ]]; then
    printf 'no assets were assembled; refusing to write an empty manifest\n' >&2
    exit 1
  fi
  if command -v sha256sum >/dev/null; then
    sha256sum "${assets[@]}" | LC_ALL=C sort -k2 > checksums.txt
  else
    shasum -a 256 "${assets[@]}" | LC_ALL=C sort -k2 > checksums.txt
  fi
)
printf '  manifest checksums.txt (%s entries)\n' "$(wc -l < "$OUT/checksums.txt" | tr -d ' ')"
