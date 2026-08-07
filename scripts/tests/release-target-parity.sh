#!/usr/bin/env bash
# The release target list now lives in three places that must agree:
#
#   scripts/assemble-release-assets.sh   what gets built into archives/packages
#   scripts/verify-release-assets.sh     what a release is required to contain
#   .github/workflows/release.yml        which runners build which target
#
# They are separate on purpose — the verifier has to be able to disagree with
# the assembler, or it is not an independent check. That makes drift between
# them the real risk: a target silently dropped from the matrix would produce a
# release that is missing a platform, and one silently added would produce an
# asset nothing verifies. This asserts they describe the same set, without
# building anything.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

extract_bash_array() {
  local file=$1 name=$2
  awk -v name="$name" '
    $0 ~ "^" name "=\\(" { capture = 1; next }
    capture && /^\)/ { capture = 0 }
    capture { gsub(/^[ \t]*"|"[ \t]*$/, ""); if (length($0)) print }
  ' "$file"
}

# goos/goarch pairs the assembler will archive.
assemble_targets=$(
  extract_bash_array scripts/assemble-release-assets.sh ARCHIVE_TARGETS |
    awk '{print $1 "/" $2}' | LC_ALL=C sort
)

# goos/goarch pairs a release is required to contain.
verify_targets=$(
  extract_bash_array scripts/verify-release-assets.sh archive_targets |
    awk '{print $1 "/" $2}' | LC_ALL=C sort
)

# goos/goarch pairs the release build matrix produces.
matrix_targets=$(
  awk '
    /^  release-binaries:/ { injob = 1; next }
    # Any other top-level job key ends the block.
    injob && /^  [a-z][a-z0-9-]*:[ ]*$/ { injob = 0 }
    injob && /^[ ]*-[ ]*goos:[ ]*[a-z0-9]+[ ]*$/ {
      goos = $0
      sub(/^[ ]*-[ ]*goos:[ ]*/, "", goos)
      sub(/[ ]*$/, "", goos)
      next
    }
    injob && /^[ ]*goarch:[ ]*[a-z0-9]+[ ]*$/ && goos != "" {
      goarch = $0
      sub(/^[ ]*goarch:[ ]*/, "", goarch)
      sub(/[ ]*$/, "", goarch)
      print goos "/" goarch
      goos = ""
      next
    }
  ' .github/workflows/release.yml | LC_ALL=C sort -u
)

fail=0
report() {
  printf 'release target parity: %s\n' "$1" >&2
  fail=1
}

if [[ -z "$assemble_targets" ]]; then report "no targets parsed from scripts/assemble-release-assets.sh"; fi
if [[ -z "$verify_targets" ]]; then report "no targets parsed from scripts/verify-release-assets.sh"; fi
if [[ -z "$matrix_targets" ]]; then report "no targets parsed from the release.yml build matrix"; fi

if [[ "$assemble_targets" != "$verify_targets" ]]; then
  report "assembler and verifier disagree"
  diff <(printf '%s\n' "$assemble_targets") <(printf '%s\n' "$verify_targets") >&2 || true
fi

if [[ "$assemble_targets" != "$matrix_targets" ]]; then
  report "assembler and the release build matrix disagree"
  diff <(printf '%s\n' "$assemble_targets") <(printf '%s\n' "$matrix_targets") >&2 || true
fi

# Windows must stay absent everywhere, both architectures. arm64 was never
# buildable (go-duckdb ships no library for it) and amd64 was dropped for having
# no customer. Asserting the absence is what stops either creeping back in
# without the runner, the MSI path and the packaging tests coming back with it.
if printf '%s\n' "$assemble_targets" "$verify_targets" "$matrix_targets" | grep -q '^windows/'; then
  report "a windows target is present; Windows was dropped from the release and CI and returns only on a customer ask"
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

printf 'release target parity: assembler, verifier and build matrix agree on:\n'
printf '  %s\n' $assemble_targets
