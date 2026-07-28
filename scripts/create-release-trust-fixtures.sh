#!/usr/bin/env bash
set -euo pipefail

DIST_DIR=${1:-dist}

subjects=$(./scripts/verify-release-assets.sh "$DIST_DIR" --print-subjects)
if [[ -z "$subjects" ]]; then
  printf 'no release trust subjects found in %s\n' "$DIST_DIR" >&2
  exit 1
fi

while IFS= read -r subject; do
  [[ -n "$subject" ]] || continue
  if [[ ! -f "$subject" ]]; then
    printf 'release trust subject not found: %s\n' "$subject" >&2
    exit 1
  fi
  digest=$(
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "$subject" | awk '{ print $1 }'
    elif command -v shasum >/dev/null 2>&1; then
      shasum -a 256 "$subject" | awk '{ print $1 }'
    else
      printf 'sha256sum or shasum is required to create release trust fixtures\n' >&2
      exit 1
    fi
  )
  python3 - "$subject" "$digest" >"$subject.sigstore.json" <<'PY'
import json
import os
import sys

subject, digest = sys.argv[1:3]
json.dump(
    {
        "_pm_unsigned_fixture": True,
        "warning": "UNSIGNED offline snapshot fixture for release verification tests only; not a Sigstore signature.",
        "subject": {
            "name": os.path.basename(subject),
            "sha256": digest,
        },
    },
    sys.stdout,
    indent=2,
    sort_keys=True,
)
sys.stdout.write("\n")
PY
done <<<"$subjects"

printf 'created offline trust-evidence fixtures for %d subjects in %s\n' "$(printf '%s\n' "$subjects" | sed '/^$/d' | wc -l | tr -d ' ')" "$DIST_DIR"
