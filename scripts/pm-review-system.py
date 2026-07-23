#!/usr/bin/env python3
"""PM review-system compiler and evaluator.

RED-baseline checkpoint: treatment intentionally delegates to the historical
presence/direct-file behavior. Semantic treatment is implemented only after the
failing corpus is frozen and captured.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
import time
from pathlib import Path
from typing import Any

SCHEMA_VERSION = "polymetrics.ai/pm-review-observation/v1"
CANONICAL_DISPOSITIONS = {
    "accepted",
    "accepted_with_modification",
    "declined",
    "duplicate",
    "deferred",
    "needs_human",
}


def load_case(path: Path, case_id: str) -> dict[str, Any]:
    document = json.loads(path.read_text())
    for case in document.get("cases", []):
        if case.get("case_id") == case_id:
            return case
    raise SystemExit(f"unknown case id: {case_id}")


def baseline_findings(case: dict[str, Any]) -> list[dict[str, str]]:
    """Model the pre-hardening marker/direct-file checks for RED measurement."""
    family = case.get("family")
    data = case.get("input", {})

    if family == "dependency_consistency":
        return [] if data.get("prose", {}).get("required_gates") else [
            {"category": family, "claim": "dependency marker missing from prose"}
        ]

    if family == "reference_closure":
        roots = set(data.get("roots", []))
        for node in data.get("nodes", []):
            if node.get("path") in roots and node.get("prohibited"):
                return [{"category": family, "claim": "direct root is prohibited"}]
        return []

    if family in {"lineage_monotonicity", "lineage_events"}:
        budget = data.get("correction_budget", {})
        if budget.get("max_correction_rounds", 0) < 1 or not isinstance(
            budget.get("rounds_by_range"), dict
        ):
            return [{"category": family, "claim": "budget shape is invalid"}]
        return []

    if family == "terminal_kind":
        kind = data.get("human_gate_kind")
        if kind in {"parent_ready", "final_parent_readiness", "correction_cap_exceeded"}:
            return []
        return [{"category": family, "claim": "human gate kind is unknown"}]

    if family == "disposition_rows":
        for row in data.get("rows", []):
            if not re.match(r"^(?:F|N|R)[A-Za-z0-9-]*$", str(row.get("id", ""))):
                continue
            if row.get("disposition") not in CANONICAL_DISPOSITIONS:
                return [{"category": family, "claim": "known-prefix disposition is invalid"}]
        return []

    if family == "stale_evidence":
        return [] if data.get("packet", {}).get("exact_head_sha") else [
            {"category": family, "claim": "packet head is absent"}
        ]

    if family in {
        "path_safety",
        "packet_coverage",
        "packet_overflow",
        "cap_transition",
        "schema_kind",
        "missing_target",
    }:
        return []

    return []


def baseline_decision(case: dict[str, Any]) -> str | None:
    if case.get("family") != "packet_threshold":
        return None
    return "combined"


def detect(case: dict[str, Any], mode: str) -> dict[str, Any]:
    started = time.perf_counter_ns()
    # RED: treatment intentionally reproduces baseline false negatives.
    findings = baseline_findings(case)
    decision = baseline_decision(case)
    elapsed_ms = (time.perf_counter_ns() - started) / 1_000_000
    return {
        "schema_version": SCHEMA_VERSION,
        "case_id": case["case_id"],
        "suite": case["suite"],
        "mode": mode,
        "findings": findings,
        "decision": decision,
        "latency_ms": round(elapsed_ms, 6),
        "tokens": {"status": "unavailable", "reason": "deterministic detector; no model call"},
        "cost": {"status": "unavailable", "reason": "deterministic detector; no model call"},
    }


def command_detect(args: argparse.Namespace) -> int:
    case = load_case(Path(args.input), args.case_id)
    json.dump(detect(case, args.mode), sys.stdout, sort_keys=True)
    sys.stdout.write("\n")
    return 0


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description="Compile and measure canonical PM review inputs")
    subparsers = result.add_subparsers(dest="command", required=True)
    detect_parser = subparsers.add_parser("detect", help="run one blinded fixture without oracle")
    detect_parser.add_argument("--mode", choices=("baseline", "treatment"), required=True)
    detect_parser.add_argument("--input", required=True)
    detect_parser.add_argument("--case-id", required=True)
    detect_parser.set_defaults(func=command_detect)
    return result


def main() -> int:
    args = parser().parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
