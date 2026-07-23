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
schema = record.get("schema_version", record.get("schemaVersion"))
if schema is None:
    # Missing schema is detected read-only legacy input. Its historical empty kind represented
    # final parent readiness; current producers must never write this shape.
    if not kind:
        print("human_ready")
    else:
        print("blocked_human_decision")
elif schema != "canonical_v2":
    # Any explicit unsupported schema stops safely, regardless of a familiar kind.
    print("blocked_human_decision")
elif kind == "parent_ready":
    print("human_ready")
else:
    # correction_cap_exceeded and every missing/unknown current kind require a human decision.
    print("blocked_human_decision")
PY
