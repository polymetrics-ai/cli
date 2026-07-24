#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: scripts/pm-terminal-classifier.sh <RUN.json>" >&2
  exit 2
fi

python3 - "$1" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
try:
    record = json.loads(path.read_text())
except (OSError, json.JSONDecodeError) as exc:
    raise SystemExit(f"cannot classify PM terminal state: {exc}")

if record.get("terminal") != "human_gate":
    print("not_human_gate")
    raise SystemExit(0)

kind = record.get("human_gate_kind", "")
keys = [key for key in ("schema_version", "schemaVersion") if key in record]
if not keys:
    # Missing schema keys identify read-only legacy input. An explicitly null/malformed schema is
    # current malformed state and must never inherit historical readiness semantics.
    if not kind:
        print("human_ready")
    else:
        print("blocked_human_decision")
    raise SystemExit(0)
values = [record[key] for key in keys]
if len(values) != 1 or not isinstance(values[0], str) or not values[0] or len(set(values)) != 1:
    print("blocked_human_decision")
    raise SystemExit(0)
schema = values[0]
if schema != "canonical_v2":
    # Any explicit unsupported schema stops safely, regardless of a familiar kind.
    print("blocked_human_decision")
elif kind == "parent_ready":
    print("human_ready")
else:
    # correction_cap_exceeded and every missing/unknown current kind require a human decision.
    print("blocked_human_decision")
PY
