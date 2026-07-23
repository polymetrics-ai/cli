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
if kind == "correction_cap_exceeded":
    print("blocked_human_decision")
elif kind in {"", "parent_ready", "final_parent_readiness"}:
    # Empty kind is the read-only legacy human-ready shape.
    print("human_ready")
else:
    # Unknown human gates fail closed rather than implying merge readiness.
    print("blocked_human_decision")
PY
