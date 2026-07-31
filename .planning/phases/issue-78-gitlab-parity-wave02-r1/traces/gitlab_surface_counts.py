#!/usr/bin/env python3
"""Local GitLab generated surface count check for issue #78."""
from __future__ import annotations

import json
import sys
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[4]
DEF_DIR = ROOT / "internal/connectors/defs/gitlab"
EXPECTED = {
    "gitlab.etl_read": 308,
    "gitlab.reverse_etl_write": 498,
    "gitlab.direct_read_query_search": 6,
    "gitlab.binary_file": 298,
    "gitlab.cdc_changefeed": 34,
    "gitlab.excluded_not_applicable": 2,
}


def main() -> int:
    api = json.loads((DEF_DIR / "api_surface.json").read_text())
    operations = json.loads((DEF_DIR / "operations.json").read_text())["operations"]
    cli = json.loads((DEF_DIR / "cli_surface.json").read_text())
    counts = Counter(op["audit_event"] for op in operations)
    problems: list[str] = []
    if len(api.get("endpoints", [])) != 1146:
        problems.append(f"api_surface endpoints={len(api.get('endpoints', []))}, want 1146")
    if len(operations) != 1146:
        problems.append(f"operations rows={len(operations)}, want 1146")
    if len(cli.get("commands", [])) != 1146:
        problems.append(f"cli commands={len(cli.get('commands', []))}, want 1146")
    for key, want in EXPECTED.items():
        got = counts[key]
        if got != want:
            problems.append(f"{key}={got}, want {want}")
    covered_streams = sorted(
        row["covered_by"]["stream"]
        for row in api.get("endpoints", [])
        if row.get("covered_by", {}).get("stream")
    )
    if covered_streams != ["groups", "issues", "projects", "users"]:
        problems.append(f"covered_streams={covered_streams!r}")
    if problems:
        print("FAIL GitLab generated surface counts")
        for problem in problems:
            print(f"- {problem}")
        return 1
    print("PASS GitLab generated surface counts")
    print(json.dumps({"api_surface": 1146, "operations": 1146, "cli_commands": 1146, "counts": dict(counts)}, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
