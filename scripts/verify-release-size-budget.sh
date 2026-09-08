#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
usage: verify-release-size-budget.sh --archive <path> --binary <name> \
  [--quiet]

Reports archive and installed binary byte counts without a size ceiling.
Missing, corrupt, or ambiguous binary artifacts still fail validation.
EOF
}

archive=''
binary=''
quiet=0

require_value() {
  if [[ $# -lt 2 || -z "${2:-}" ]]; then
    printf '%s requires a value\n' "$1" >&2
    exit 2
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --archive)
      require_value "$@"
      archive=$2
      shift 2
      ;;
    --binary)
      require_value "$@"
      binary=$2
      shift 2
      ;;
    --quiet)
      quiet=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n' "$1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$archive" || -z "$binary" ]]; then
  printf '%s\n' '--archive and --binary are required' >&2
  usage >&2
  exit 2
fi
if [[ ! -f "$archive" ]]; then
  printf 'release archive not found: %s\n' "$archive" >&2
  exit 1
fi
if [[ "$binary" == */* || "$binary" == . || "$binary" == .. ]]; then
  printf 'release binary name must be one archive entry: %s\n' "$binary" >&2
  exit 2
fi
file_size_bytes() {
  wc -c < "$1" | tr -d '[:space:]'
}

report_size() {
  local kind=$1 subject=$2 bytes=$3
  if [[ "$quiet" != 1 ]]; then
    printf 'release-size-report kind=%s subject=%s bytes=%s\n' "$kind" "$subject" "$bytes"
  fi
}

# The caller already verifies the full archive layout. This script independently
# checks that the named installed binary occurs exactly once before streaming it
# from the archive, so no checkout or temporary extraction path participates in
# the installed-size measurement.
binary_entries=$(tar -tzf "$archive" | awk -v binary="$binary" '$0 == binary { count++ } END { print count + 0 }')
if [[ "$binary_entries" != 1 ]]; then
  printf 'release archive %s contains %s entries named %s, want exactly one\n' "$archive" "$binary_entries" "$binary" >&2
  exit 1
fi

archive_bytes=$(file_size_bytes "$archive")
installed_binary_bytes=$(tar -xOzf "$archive" "$binary" | wc -c | tr -d '[:space:]')

report_size archive "$archive" "$archive_bytes"
report_size installed_binary "${archive}!${binary}" "$installed_binary_bytes"
