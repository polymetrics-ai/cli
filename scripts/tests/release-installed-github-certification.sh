#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
assembler="$repo_root/scripts/assemble-release-assets.sh"
archive=''

while [[ $# -gt 0 ]]; do
  case "$1" in
    --archive)
      if [[ $# -lt 2 || -z "${2:-}" ]]; then
        printf '%s\n' '--archive requires a value' >&2
        exit 2
      fi
      archive=$2
      shift 2
      ;;
    -h|--help)
      printf '%s\n' 'usage: release-installed-github-certification.sh [--archive <tar.gz>]' >&2
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n' "$1" >&2
      exit 2
      ;;
  esac
done

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

if [[ -z "$archive" ]]; then
  goos=$(go env GOOS)
  goarch=$(go env GOARCH)
  target="$goos/$goarch"
  version=0.0.0-installed-certification
  mkdir -p "$tmp/binaries/${goos}_${goarch}"
  go build -trimpath -ldflags '-s -w' -o "$tmp/binaries/${goos}_${goarch}/pm" ./cmd/pm
  SOURCE_DATE_EPOCH=1 "$assembler" \
    --version "$version" \
    --binaries "$tmp/binaries" \
    --out "$tmp/dist" \
    --targets "$target" >/dev/null
  archive="$tmp/dist/pm_${version}_${goos}_${goarch}.tar.gz"
fi

if [[ ! -f "$archive" ]]; then
  printf 'release archive not found: %s\n' "$archive" >&2
  exit 1
fi

installed="$tmp/installed"
project="$tmp/project"
mkdir -p "$installed" "$project"
tar -C "$installed" -xzf "$archive"
if [[ ! -x "$installed/pm" ]]; then
  printf 'release archive did not extract an executable root pm: %s\n' "$archive" >&2
  exit 1
fi
case "$project" in
  "$repo_root"/*)
    printf 'installed certification project must be outside the checkout: %s\n' "$project" >&2
    exit 1
    ;;
esac

"$installed/pm" init --root "$project" --json >/dev/null
set +e
(cd "$project" && "$installed/pm" connectors certify github --full --json) > "$tmp/report.json" 2> "$tmp/report.err"
certification_status=$?
set -e
if [[ "$certification_status" == 0 ]]; then
  printf 'credential-free installed GitHub certification unexpectedly succeeded\n' >&2
  sed -n '1,80p' "$tmp/report.err" >&2
  exit 1
fi
if [[ -s "$tmp/report.err" ]]; then
  printf 'credential-free installed GitHub certification wrote stderr\n' >&2
  sed -n '1,80p' "$tmp/report.err" >&2
  exit 1
fi

python3 - "$tmp/report.json" <<'PY'
import json
import sys

document = json.load(open(sys.argv[1], encoding="utf-8"))
report = document.get("report") or {}
graphql = (report.get("capabilities") or {}).get("graphql") or {}
expected = {
    "schema_conformant": 29,
    "live_required": 2,
    "fixture_bound": 274,
}
actual = {key: graphql.get(key) for key in expected}
if actual != expected:
    raise SystemExit(f"installed GraphQL inventory = {actual}, want {expected}")
stages = report.get("stages") or []
if not any(stage.get("name") == "graphql_schema_conformance" and stage.get("passed") is True for stage in stages):
    raise SystemExit("installed GitHub certification did not pass graphql_schema_conformance")
PY

printf '%s\n' 'installed GitHub certification archive proof passed'
