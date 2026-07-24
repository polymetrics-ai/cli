#!/usr/bin/env python3
"""Compile, validate, synthesize, and measure canonical PM review packets.

The tool is intentionally dependency-free. It emits JSON to stdout and never
writes project files. The parent orchestrator remains the sole lifecycle owner.
"""

from __future__ import annotations

import argparse
import ast
import fnmatch
import hashlib
import json
import math
import os
import posixpath
import re
import subprocess
import sys
import tempfile
import time
from collections import defaultdict, deque
from pathlib import Path, PurePosixPath
from typing import Any, Iterable

OBSERVATION_SCHEMA = "polymetrics.ai/pm-review-observation/v1"
CONFIG_SCHEMA = "polymetrics.ai/pm-review-system/v3"
SCOPE_SCHEMA = "polymetrics.ai/pm-review-scope/v1"
COMPILE_SCHEMA = "polymetrics.ai/pm-review-compile/v3"
IMPACT_SCHEMA = "polymetrics.ai/pm-review-impact-graph/v2"
PACKET_SCHEMA = "polymetrics.ai/pm-review-packet/v3"
PACKET_RESPONSE_SCHEMA = "polymetrics.ai/pm-review-packet-response/v3"
SYNTHESIS_SCHEMA = "polymetrics.ai/pm-review-synthesis/v3"
LAB_EVIDENCE_SCHEMA = "polymetrics.ai/pm-review-lab-evidence/v2"
MEASUREMENT_SCHEMA = "polymetrics.ai/pm-review-measurement/v1"
CANONICAL_SCHEMA = "canonical_v2"
CANONICAL_GATE_KINDS = {"parent_ready", "correction_cap_exceeded"}
CANONICAL_DISPOSITIONS = {
    "accepted",
    "accepted_with_modification",
    "declined",
    "duplicate",
    "deferred",
    "needs_human",
}
HEX_SHA = re.compile(r"^[0-9a-f]{40}$")
CONTROL = re.compile(r"[\x00-\x1f\x7f]")
REFERENCE_SUFFIXES = (".md", ".json", ".yaml", ".yml", ".sh", ".py", ".go")
INDEX_SUFFIXES = set(REFERENCE_SUFFIXES)
REPO_REFERENCE = re.compile(
    r"(?<![A-Za-z0-9])((?:(?:\.agents|\.pi|scripts|\.planning|cmd|internal)/"
    r"[A-Za-z0-9_.\-/]+\.(?:md|json|ya?ml|sh|py|go)|(?:AGENTS|CLAUDE)\.md))"
)
MARKDOWN_LINK = re.compile(r"\[[^\]]*\]\(([^)]+)\)")


class ReviewSystemError(ValueError):
    """Expected invalid review input."""


def emit_json(value: Any) -> None:
    json.dump(value, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")


def finding(category: str, claim: str, path: str | None = None) -> dict[str, Any]:
    result: dict[str, Any] = {"category": category, "claim": claim}
    if path is not None:
        result["path"] = path
    return result


def validate_sha(value: str, label: str) -> str:
    if CONTROL.search(value) or value.startswith("-") or not HEX_SHA.fullmatch(value):
        raise ReviewSystemError(f"{label} must be exactly 40 lowercase hexadecimal characters")
    return value


def validate_relative_path(value: str, label: str = "path") -> str:
    if not isinstance(value, str) or not value:
        raise ReviewSystemError(f"{label} must be a non-empty string")
    if CONTROL.search(value):
        raise ReviewSystemError(f"{label} contains a control character")
    if value.startswith("-"):
        raise ReviewSystemError(f"{label} must not be option-like")
    if "\\" in value:
        raise ReviewSystemError(f"{label} must use repository POSIX separators")
    path = PurePosixPath(value)
    if path.is_absolute() or ".." in path.parts:
        raise ReviewSystemError(f"{label} must stay repository-relative")
    normalized = path.as_posix()
    if normalized in {"", "."}:
        raise ReviewSystemError(f"{label} must name a repository file")
    return normalized


def resolve_safe(root: Path, relative: str, *, must_exist: bool = True) -> Path:
    normalized = validate_relative_path(relative)
    root_resolved = root.resolve(strict=True)
    candidate = root_resolved / normalized
    try:
        resolved = candidate.resolve(strict=must_exist)
    except OSError as exc:
        raise ReviewSystemError(f"cannot resolve {normalized}: {exc}") from exc
    try:
        common = Path(os.path.commonpath((str(root_resolved), str(resolved))))
    except ValueError as exc:
        raise ReviewSystemError(f"path escapes repository: {normalized}") from exc
    if common != root_resolved:
        raise ReviewSystemError(f"path escapes repository: {normalized}")
    if must_exist and not resolved.is_file():
        raise ReviewSystemError(f"required file does not exist: {normalized}")
    return resolved


def path_matches(path: str, patterns: Iterable[str]) -> bool:
    for pattern in patterns:
        if pattern.endswith("/**") and (path == pattern[:-3] or path.startswith(pattern[:-2])):
            return True
        if fnmatch.fnmatchcase(path, pattern):
            return True
    return False


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise ReviewSystemError(f"cannot read JSON {path}: {exc}") from exc


def load_case(path: Path, case_id: str) -> dict[str, Any]:
    document = load_json(path)
    for case in document.get("cases", []):
        if case.get("case_id") == case_id:
            return case
    raise ReviewSystemError(f"unknown case id: {case_id}")


def baseline_findings(case: dict[str, Any]) -> list[dict[str, Any]]:
    """Model the pre-hardening marker/direct-file checks for comparison."""
    family = case.get("family")
    data = case.get("input", {})

    if family == "dependency_consistency":
        return [] if data.get("prose", {}).get("required_gates") else [
            finding(family, "dependency marker missing from prose")
        ]

    if family == "reference_closure":
        roots = set(data.get("roots", []))
        for node in data.get("nodes", []):
            if node.get("path") in roots and node.get("prohibited"):
                return [finding(family, "direct root is prohibited")]
        return []

    if family in {"lineage_monotonicity", "lineage_events"}:
        budget = data.get("correction_budget", {})
        if budget.get("max_correction_rounds", 0) < 1 or not isinstance(
            budget.get("rounds_by_range"), dict
        ):
            return [finding(family, "budget shape is invalid")]
        return []

    if family == "terminal_kind":
        kind = data.get("human_gate_kind")
        if kind in {"parent_ready", "final_parent_readiness", "correction_cap_exceeded"}:
            return []
        return [finding(family, "human gate kind is unknown")]

    if family == "disposition_rows":
        for row in data.get("rows", []):
            if not re.match(r"^(?:F|N|R)[A-Za-z0-9-]*$", str(row.get("id", ""))):
                continue
            if row.get("disposition") not in CANONICAL_DISPOSITIONS:
                return [finding(family, "known-prefix disposition is invalid")]
        return []

    if family == "stale_evidence":
        return [] if data.get("packet", {}).get("exact_head_sha") else [
            finding(family, "packet head is absent")
        ]

    return []


def graph_findings(data: dict[str, Any], category: str) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    nodes = {node.get("path"): node for node in data.get("nodes", []) if node.get("path")}
    pending = deque(data.get("roots", []))
    seen: set[str] = set()
    while pending:
        current = pending.popleft()
        if current in seen:
            continue
        seen.add(current)
        node = nodes.get(current)
        if node is None:
            result.append(finding(category, f"active target is missing: {current}", current))
            continue
        if node.get("prohibited"):
            result.append(finding(category, f"active closure reaches prohibited target: {current}", current))
        for edge in node.get("references", []):
            target = edge.get("target", "")
            try:
                validate_relative_path(target, "active reference")
            except ReviewSystemError as exc:
                result.append(finding(category, str(exc), current))
                continue
            if target not in nodes:
                result.append(finding(category, f"active target is missing: {target}", current))
                continue
            pending.append(target)
    return result


def impact_contract_findings(data: dict[str, Any]) -> list[dict[str, Any]]:
    nodes = set(data.get("nodes", []))
    edges = data.get("edges", [])
    forward: dict[str, list[dict[str, Any]]] = defaultdict(list)
    reverse: dict[str, list[dict[str, Any]]] = defaultdict(list)
    result: list[dict[str, Any]] = []
    for edge in edges:
        source = edge.get("source")
        target = edge.get("target")
        certainty = edge.get("certainty")
        if not source or not target or certainty not in {"active", "inactive", "unknown"}:
            result.append(finding("impact_graph_contract", "edge lacks typed source/target/certainty"))
            continue
        if certainty != "inactive" and target not in nodes:
            result.append(finding("impact_graph_contract", f"{certainty} impact target is absent: {target}"))
        if certainty != "inactive":
            forward[source].append(edge)
            reverse[target].append(edge)
    pending = deque((seed, 0) for seed in sorted(set(data.get("seeds", []))))
    seen: set[str] = set()
    bound_hit = False
    maximum = data.get("max_depth")
    while pending:
        current, depth = pending.popleft()
        if current in seen:
            continue
        seen.add(current)
        neighbors = [edge["target"] for edge in forward.get(current, [])]
        neighbors.extend(edge["source"] for edge in reverse.get(current, []))
        if isinstance(maximum, int) and depth >= maximum:
            if any(neighbor not in seen for neighbor in neighbors):
                bound_hit = True
            continue
        for neighbor in sorted(set(neighbors)):
            if neighbor not in seen:
                pending.append((neighbor, depth + 1))
    reported = set(data.get("reported_impact", []))
    missing = sorted(seen - reported)
    if missing:
        result.append(finding("impact_graph_contract", f"reported impact omits bidirectional nodes: {missing}"))
    if bound_hit and data.get("reported_status") != "blocked":
        result.append(finding("impact_graph_contract", "continuing graph frontier was not blocked"))
    return result


def lab_contract_findings(data: dict[str, Any]) -> list[dict[str, Any]]:
    result: list[dict[str, Any]] = []
    if data.get("no_experiment_reason"):
        return result
    if data.get("change_scope") != "lab":
        result.append(finding("hypothesis_lab_contract", "temporary change escaped the private lab"))
    if data.get("command_category") not in {"targeted_test", "compiler", "linter", "parser", "fixture_generation", "read_only_history"}:
        result.append(finding("hypothesis_lab_contract", "forbidden command/effect category was accepted"))
    if data.get("secret_material_present"):
        result.append(finding("hypothesis_lab_contract", "secret material reached lab evidence"))
    if data.get("limit_hit") and data.get("status") == "clean":
        result.append(finding("hypothesis_lab_contract", "resource-limited experiment was represented as clean"))
    if data.get("candidate_unchanged") is not True:
        result.append(finding("hypothesis_lab_contract", "candidate identity was not preserved"))
    if data.get("cleanup_verified") is not True:
        result.append(finding("hypothesis_lab_contract", "lab cleanup was not verified"))
    if data.get("supports") == "inconclusive" and data.get("status") == "clean":
        result.append(finding("hypothesis_lab_contract", "inconclusive experiment was represented as clean proof"))
    return result


def treatment_findings(case: dict[str, Any]) -> list[dict[str, Any]]:
    family = case.get("family")
    data = case.get("input", {})
    result: list[dict[str, Any]] = []

    if family == "impact_graph_contract":
        return impact_contract_findings(data)

    if family == "hypothesis_lab_contract":
        return lab_contract_findings(data)

    if family == "review_contract_version":
        if data.get("schema_version") != data.get("required_schema_version") and data.get("reported_status") == "clean":
            result.append(finding(family, "incompatible review contract requires explicit migration"))
        return result

    if family == "dependency_consistency":
        authority = data.get("authority", {})
        ready = data.get("ready_item", {})
        gate = authority.get("gate_id")
        gates = ready.get("integration_gates", [])
        blocked = authority.get("blocks_worker_start") and not authority.get("integrated")
        if gate not in gates:
            result.append(finding(family, f"authoritative dependency gate is absent from ready item: {gate}"))
        if blocked and ready.get("decision") != "not_spawned_dependency_blocked":
            result.append(finding(family, "blocked authoritative dependency is dispatchable"))
        return result

    if family in {"reference_closure", "missing_target"}:
        return graph_findings(data, family)

    if family == "lineage_monotonicity":
        base = data.get("exact_base_sha", "")
        lineage = data.get("candidate_lineage", "")
        expected_key = f"{base}...{lineage}"
        rounds = data.get("correction_budget", {}).get("rounds_by_range", {})
        if set(rounds) != {expected_key}:
            result.append(finding(family, "rounds are not keyed only by exact base and stable candidate lineage"))
        heads = data.get("replacement_heads", [])
        if len(heads) != len(set(heads)):
            result.append(finding(family, "replacement head history is not append-only unique"))
        return result

    if family == "lineage_events":
        previous = -1
        max_seen = 0
        migrated_legacy = False
        for event in data.get("events", []):
            rounds = event.get("rounds")
            if not isinstance(rounds, int) or rounds < previous:
                result.append(finding(family, "replacement, resume, or migration reduced correction rounds"))
                break
            previous = rounds
            max_seen = max(max_seen, rounds)
            if event.get("event") == "migrate_legacy":
                migrated_legacy = True
            elif migrated_legacy and event.get("event") == "write_legacy":
                result.append(finding(family, "canonical state wrote the legacy counter after migration"))
        prior_heads = data.get("prior_head_history", [])
        current_heads = data.get("head_history", prior_heads)
        if current_heads[: len(prior_heads)] != prior_heads or len(current_heads) != len(set(current_heads)):
            result.append(finding(family, "candidate head history is not append-only"))
        base = data.get("exact_base_sha", "")
        lineage = data.get("candidate_lineage", "")
        stable = data.get("correction_budget", {}).get("rounds_by_range", {}).get(f"{base}...{lineage}")
        if not isinstance(stable, int) or stable < max_seen:
            result.append(finding(family, "persisted stable-lineage count is below observed event history"))
        return result

    if family == "terminal_kind":
        if data.get("schema_version") == CANONICAL_SCHEMA and data.get("terminal") == "human_gate":
            if data.get("human_gate_kind") not in CANONICAL_GATE_KINDS:
                result.append(finding(family, "current-schema human gate kind is noncanonical"))
        return result

    if family == "disposition_rows":
        for row in data.get("rows", []):
            identifier = row.get("id")
            if not isinstance(identifier, str) or not identifier.strip():
                result.append(finding(family, "finding row has an empty identifier"))
            if row.get("disposition") not in CANONICAL_DISPOSITIONS:
                result.append(finding(family, f"finding {identifier} uses a noncanonical disposition"))
        return result

    if family == "stale_evidence":
        candidate = data.get("candidate", {})
        packet = data.get("packet", {})
        for field in ("exact_base_sha", "exact_head_sha"):
            if candidate.get(field) != packet.get(field):
                result.append(finding(family, f"packet {field} does not match current candidate"))
        return result

    if family == "schema_kind":
        if data.get("schema_version") != CANONICAL_SCHEMA:
            result.append(finding(family, "explicit unsupported schema version must stop safely"))
        if data.get("terminal") == "human_gate" and data.get("human_gate_kind") not in CANONICAL_GATE_KINDS:
            result.append(finding(family, "unknown current human gate kind must stop safely"))
        return result

    if family == "path_safety":
        for raw in data.get("references", []):
            try:
                validate_relative_path(raw, "active reference")
            except ReviewSystemError as exc:
                result.append(finding(family, str(exc), data.get("root")))
        return result

    if family == "packet_coverage":
        changed = set(data.get("changed_files", []))
        packet = data.get("packet", {})
        reviewed = set(packet.get("reviewed_files", []))
        missing = sorted(changed - reviewed)
        if missing:
            result.append(finding(family, f"changed files are not reviewed: {missing}"))
        if packet.get("unreviewed_files"):
            result.append(finding(family, "packet declares unreviewed files"))
        return result

    if family == "packet_overflow":
        packet = data.get("packet", {})
        context = packet.get("context", {})
        if context.get("overflow") or context.get("truncated"):
            result.append(finding(family, "overflowed or truncated packet cannot return clean"))
        return result

    if family == "cap_transition":
        rounds = data.get("rounds", 0)
        maximum = data.get("max_correction_rounds", 0)
        if rounds > maximum:
            if data.get("terminal") != "human_gate" or data.get("human_gate_kind") != "correction_cap_exceeded":
                result.append(finding(family, "cap exceed did not enter correction-cap human gate"))
            if data.get("review_status") != "blocked" or not data.get("outstanding_findings"):
                result.append(finding(family, "cap exceed did not preserve blocked outstanding findings"))
        return result

    return result


def threshold_decision(data: dict[str, Any], mode: str) -> str | None:
    if mode == "baseline":
        return "combined"
    if not data.get("partitionable", True):
        return "blocked"
    if (
        data.get("review_files", 0) <= 20
        and data.get("changed_lines", 0) <= 600
        and data.get("domains", 0) <= 1
    ):
        return "combined"
    return "split"


def detect(case: dict[str, Any], mode: str) -> dict[str, Any]:
    started = time.perf_counter_ns()
    findings = baseline_findings(case) if mode == "baseline" else treatment_findings(case)
    decision = threshold_decision(case.get("input", {}), mode) if case.get("family") == "packet_threshold" else None
    elapsed_ms = (time.perf_counter_ns() - started) / 1_000_000
    return {
        "schema_version": OBSERVATION_SCHEMA,
        "case_id": case["case_id"],
        "suite": case["suite"],
        "mode": mode,
        "findings": findings,
        "decision": decision,
        "latency_ms": round(elapsed_ms, 6),
        "tokens": {"status": "unavailable", "reason": "deterministic detector; no model call"},
        "cost": {"status": "unavailable", "reason": "deterministic detector; no model call"},
    }


def run_git(root: Path, args: list[str]) -> str:
    proc = subprocess.run(
        ["git", "-C", str(root), *args],
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise ReviewSystemError(f"git {' '.join(args[:2])} failed: {detail}")
    return proc.stdout


def normalize_reference(source: str, raw: str) -> str | None:
    value = raw.strip().strip("`'\"")
    value = value.split("#", 1)[0].split("?", 1)[0]
    if not value or value.startswith(("http://", "https://", "mailto:", "#")):
        return None
    if any(marker in value for marker in ("<", ">", "{", "}", "$")):
        return None
    if not value.endswith(REFERENCE_SUFFIXES):
        return None
    if value in {"AGENTS.md", "CLAUDE.md"} or value.startswith(
        (".agents/", ".pi/", "scripts/", ".planning/", "cmd/", "internal/")
    ):
        return validate_relative_path(value, "active reference")
    source_parent = PurePosixPath(source).parent
    combined = posixpath.normpath((source_parent / value).as_posix())
    return validate_relative_path(combined, "active reference")


def extract_references(relative: str, text: str) -> list[dict[str, Any]]:
    """Extract active references, not every historical/test mention of a path."""
    candidates: list[tuple[str, str, int]] = []
    suffix = PurePosixPath(relative).suffix
    lines = text.splitlines()
    inactive_words = ("historical", "legacy", "deprecated", "prohibited", "forbid", "reject")

    if relative.endswith("pm-review-system.json"):
        document = json.loads(text)
        for raw in document.get("canonical_roots", []):
            candidates.append((raw, "config_canonical_root", 0))
        for authority in document.get("authorities", []):
            candidates.append((authority.get("authoritative_path", ""), "config_authority", 0))
            for field in ("writers", "readers", "mirrors"):
                for raw in authority.get(field, []):
                    candidates.append((raw, f"config_authority_{field}", 0))
    elif suffix in {".sh", ".py"}:
        active_command = re.compile(r"(?:^|\s)(?:source|bash|python3|exec|runpy\.run_path)\b|\$repo_root/")
        for number, line_text in enumerate(lines, 1):
            if not active_command.search(line_text):
                continue
            for match in REPO_REFERENCE.finditer(line_text):
                candidates.append((match.group(1), "script_execution_path", number))
    else:
        for match in MARKDOWN_LINK.finditer(text):
            line = text.count("\n", 0, match.start()) + 1
            line_text = lines[line - 1].lower() if line <= len(lines) else ""
            if not any(word in line_text for word in inactive_words):
                candidates.append((match.group(1), "markdown_link", line))
        for number, line_text in enumerate(lines, 1):
            lowered = line_text.lower()
            if any(word in lowered for word in inactive_words):
                continue
            if suffix in {".yaml", ".yml"}:
                key = line_text.split(":", 1)[0].lower() if ":" in line_text else ""
                if not any(word in key for word in ("path", "prompt", "contract", "workflow", "schema", "template", "required", "source")):
                    continue
                reason = "yaml_or_frontmatter_path"
            else:
                reason = "inline_required_path"
            for match in REPO_REFERENCE.finditer(line_text):
                candidates.append((match.group(1), reason, number))

    result: list[dict[str, Any]] = []
    seen: set[tuple[str, str, int]] = set()
    for raw, reason, line in candidates:
        target = normalize_reference(relative, raw)
        if target is None:
            continue
        key = (target, reason, line)
        if key not in seen:
            seen.add(key)
            result.append({"source": relative, "target": target, "reason": reason, "line": line})
    return result


def compile_closure(root: Path, config: dict[str, Any]) -> tuple[list[str], list[dict[str, Any]], list[dict[str, Any]]]:
    pending = deque(config["canonical_roots"])
    seen: set[str] = set()
    edges: list[dict[str, Any]] = []
    findings: list[dict[str, Any]] = []
    prohibited = set(config.get("prohibited_active_targets", []))
    ignored = tuple(config.get("ignored_reference_prefixes", []))
    allowed_prefixes = tuple(config.get("reference_prefixes", []))
    explicit_files = set(config.get("explicit_reference_files", []))

    while pending:
        relative = validate_relative_path(pending.popleft(), "canonical root")
        if relative in seen:
            continue
        try:
            path = resolve_safe(root, relative)
        except ReviewSystemError as exc:
            findings.append(finding("reference_closure", str(exc), relative))
            continue
        seen.add(relative)
        if relative in prohibited:
            findings.append(finding("reference_closure", "active closure reaches prohibited target", relative))
        try:
            text = path.read_text()
        except (OSError, UnicodeDecodeError) as exc:
            findings.append(finding("reference_closure", f"cannot read active target: {exc}", relative))
            continue
        try:
            discovered = extract_references(relative, text)
        except ReviewSystemError as exc:
            findings.append(finding("path_safety", str(exc), relative))
            continue
        for edge in discovered:
            target = edge["target"]
            if target.startswith(ignored):
                continue
            if target not in explicit_files and not target.startswith(allowed_prefixes):
                continue
            edges.append(edge)
            try:
                resolve_safe(root, target)
            except ReviewSystemError as exc:
                findings.append(finding("reference_closure", str(exc), relative))
                continue
            if target not in seen:
                pending.append(target)
    return sorted(seen), sorted(edges, key=lambda item: (item["source"], item["target"], item["line"])), findings


def authority_inventory(root: Path, config: dict[str, Any]) -> tuple[list[dict[str, Any]], list[str], list[dict[str, Any]]]:
    inventory: list[dict[str, Any]] = []
    files: set[str] = set()
    findings: list[dict[str, Any]] = []
    for record in config.get("authorities", []):
        item = {key: record.get(key) for key in ("id", "authoritative_path", "writers", "readers", "mirrors", "invariants")}
        inventory.append(item)
        paths = [record.get("authoritative_path", "")]
        for field in ("writers", "readers", "mirrors"):
            paths.extend(record.get(field, []))
        for relative in paths:
            try:
                resolve_safe(root, relative)
                files.add(relative)
            except ReviewSystemError as exc:
                findings.append(finding("authority_inventory", str(exc), relative))
        expected_head = record.get("expected_current_head")
        if expected_head:
            try:
                state = load_json(resolve_safe(root, record["authoritative_path"]))
                if state.get("currentHead") != expected_head:
                    findings.append(
                        finding(
                            "authoritative_state_consistency",
                            f"authoritative currentHead {state.get('currentHead')!r} != expected {expected_head!r}",
                            record["authoritative_path"],
                        )
                    )
            except ReviewSystemError as exc:
                findings.append(finding("authoritative_state_consistency", str(exc), record["authoritative_path"]))
        live = record.get("live_dispatch_check")
        if live:
            try:
                state = load_json(resolve_safe(root, record["authoritative_path"]))
            except ReviewSystemError as exc:
                findings.append(finding("dependency_consistency", str(exc), record["authoritative_path"]))
                continue
            gate = state.get(live["gate_object"], {})
            queue = state.get(live["queue_object"], [])
            ready = next((row for row in queue if row.get(live["issue_field"]) == live["issue"]), None)
            if ready is None:
                findings.append(finding("dependency_consistency", f"authoritative ready item {live['issue']} is missing"))
                continue
            blocked = gate.get(live["blocks_field"]) is True and gate.get("status") not in live["integrated_statuses"]
            if live["gate_id"] not in ready.get(live["integration_gates_field"], []):
                findings.append(finding("dependency_consistency", f"ready item {live['issue']} omits authoritative gate {live['gate_id']}"))
            if blocked and ready.get(live["decision_field"]) != live["blocked_decision"]:
                findings.append(finding("dependency_consistency", f"ready item {live['issue']} is dispatchable before authoritative gate"))
    for relationship in config.get("configured_relationships", []):
        if relationship.get("validator") != "pm_review_phase_v2":
            continue
        try:
            schema = load_json(resolve_safe(root, relationship["source"]))
            state = load_json(resolve_safe(root, relationship["target"]))
        except ReviewSystemError as exc:
            findings.append(finding("authority_inventory", str(exc), relationship.get("target")))
            continue
        required = set(schema.get("required", []))
        missing = required - set(state)
        expected_version = schema.get("properties", {}).get("schemaVersion", {}).get("const")
        budget = state.get("correctionBudget")
        if missing or state.get("schemaVersion") != expected_version or not isinstance(budget, dict):
            findings.append(
                finding(
                    "authoritative_state_consistency",
                    f"dedicated PM phase state does not validate: missing={sorted(missing)} expected_schema={expected_version!r}",
                    relationship["target"],
                )
            )
    return inventory, sorted(files), findings


def stable_edge(
    source: str,
    target: str,
    relation: str,
    parser_name: str,
    reason: str,
    certainty: str,
    line: int = 0,
) -> dict[str, Any]:
    identity = "\0".join((source, target, relation, parser_name, reason, certainty, str(line)))
    return {
        "id": "edge-" + hashlib.sha256(identity.encode()).hexdigest()[:20],
        "source": source,
        "target": target,
        "relation": relation,
        "direction": "forward",
        "parser": parser_name,
        "reason": reason,
        "line": line,
        "certainty": certainty,
    }


def line_certainty(text: str) -> str:
    lowered = text.lower()
    # Negative requirements containing "reject", "forbid", or "prohibited" are live contracts,
    # not historical edges. Only explicit lifecycle markers establish inactivity.
    if any(word in lowered for word in ("historical", "legacy-only", "deprecated", "superseded", "retired")):
        return "inactive"
    if any(word in lowered for word in ("${{", " optional", "condition", "when:")):
        return "unknown"
    return "active"


def reference_relation(text: str, matched_path: str = "") -> str:
    lowered = text.lower().replace(matched_path.lower(), "") if matched_path else text.lower()
    if any(
        word in lowered
        for word in (
            "require",
            "must read",
            "must use",
            "load ",
            "loads ",
            "template",
            "contract",
            "workflow",
            "schema",
            "prompt",
            "consume",
            "writes ",
            "reads ",
            "run ",
            "source ",
        )
    ):
        return "required_reference"
    return "descriptive_reference"


def script_invocation_line(text: str) -> bool:
    return bool(
        re.search(
            r"(?:^|[\s;|&({=])(?:source|exec|bash|sh|python3)(?:\s|$)|runpy\.run_path",
            text,
        )
    )


def structural_inactive_lines(relative: str, lines: list[str]) -> set[int]:
    inactive: set[int] = set()
    if relative.endswith("TDD-LEDGER.md"):
        return set(range(1, len(lines) + 1))
    section_inactive = False
    heredoc: str | None = None
    fixture_continuation = False
    for number, line in enumerate(lines, 1):
        stripped = line.strip()
        if PurePosixPath(relative).suffix == ".md" and stripped.startswith("#"):
            heading = stripped.lstrip("#").strip().lower()
            section_inactive = any(word in heading for word in ("historical", "forbidden", "excluded", "deprecated"))
        if section_inactive:
            inactive.add(number)
        if PurePosixPath(relative).suffix == ".sh":
            if heredoc is not None:
                inactive.add(number)
                if stripped == heredoc:
                    heredoc = None
                continue
            marker = re.search(r"<<-?['\"]?([A-Za-z_][A-Za-z0-9_]*)['\"]?", line)
            if marker:
                heredoc = marker.group(1)
                inactive.add(number)
            if fixture_continuation or any(value in line for value in ("$impact_repo", "$test_tmp", "$fixture_root")):
                inactive.add(number)
                fixture_continuation = line.rstrip().endswith("\\")
            elif fixture_continuation:
                fixture_continuation = line.rstrip().endswith("\\")
    return inactive


def typed_file_edges(
    relative: str,
    text: str,
    universe: set[str],
) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    # The review-system policy and per-run scope are interpreted structurally. Their pattern arrays
    # are declarations, not active repository references.
    if relative.endswith(("pm-review-system.json", "REVIEW-SCOPE.json")) or relative.startswith(
        "scripts/tests/fixtures/"
    ):
        return [], []
    result: list[dict[str, Any]] = []
    findings: list[dict[str, Any]] = []
    seen: set[str] = set()
    lines = text.splitlines()
    suffix = PurePosixPath(relative).suffix
    structurally_inactive = structural_inactive_lines(relative, lines)

    for number, line_text in enumerate(lines, 1):
        certainty = "inactive" if number in structurally_inactive else line_certainty(line_text)
        if certainty != "inactive" and "fixture" in line_text.lower() and any(
            marker in relative for marker in ("TDD-LEDGER.md", "VERIFICATION.md", "REVIEW-R1-DISPOSITION.md")
        ):
            certainty = "inactive"
        for match in MARKDOWN_LINK.finditer(line_text):
            try:
                target = normalize_reference(relative, match.group(1))
            except ReviewSystemError as exc:
                findings.append(finding("impact_path", str(exc), relative))
                continue
            if target:
                edge = stable_edge(
                    relative,
                    target,
                    reference_relation(line_text, match.group(1)),
                    "markdown",
                    "markdown_link",
                    certainty,
                    number,
                )
                if edge["id"] not in seen:
                    seen.add(edge["id"])
                    result.append(edge)
        for match in REPO_REFERENCE.finditer(line_text):
            if match.start() >= 2 and line_text[match.start() - 2 : match.start()] == "~/":
                continue
            target = match.group(1)
            relation = reference_relation(line_text, target)
            parser_name = "structured_path" if suffix in {".json", ".yaml", ".yml"} else "text_path"
            reason = "inline_repository_path"
            if suffix in {".sh", ".py"} and script_invocation_line(line_text):
                relation = "script_invokes"
                parser_name = "script"
                reason = "script_execution_path"
            edge = stable_edge(relative, target, relation, parser_name, reason, certainty, number)
            if edge["id"] not in seen:
                seen.add(edge["id"])
                result.append(edge)

    if suffix == ".py":
        try:
            tree = ast.parse(text, filename=relative)
        except SyntaxError as exc:
            findings.append(finding("impact_parser", f"Python syntax prevents authoritative import parsing: {exc}", relative))
            return sorted(result, key=lambda item: item["id"]), findings
        source_parent = PurePosixPath(relative).parent
        for node in ast.walk(tree):
            modules: list[str] = []
            if isinstance(node, ast.Import):
                modules.extend(alias.name for alias in node.names)
            elif isinstance(node, ast.ImportFrom) and node.module:
                modules.append(node.module)
            for module in modules:
                candidates = [module.replace(".", "/") + ".py"]
                candidates.append((source_parent / (module.rsplit(".", 1)[-1] + ".py")).as_posix())
                target = next((candidate for candidate in candidates if candidate in universe), None)
                if target:
                    edge = stable_edge(
                        relative,
                        target,
                        "python_import",
                        "python_ast",
                        f"import:{module}",
                        "active",
                        getattr(node, "lineno", 0),
                    )
                    if edge["id"] not in seen:
                        seen.add(edge["id"])
                        result.append(edge)
    return sorted(result, key=lambda item: item["id"]), findings


def decode_json_stream(text: str) -> list[dict[str, Any]]:
    decoder = json.JSONDecoder()
    position = 0
    result: list[dict[str, Any]] = []
    while position < len(text):
        while position < len(text) and text[position].isspace():
            position += 1
        if position >= len(text):
            break
        item, position = decoder.raw_decode(text, position)
        if isinstance(item, dict):
            result.append(item)
    return result


def go_impact_edges(
    root: Path,
    timeout_seconds: float,
) -> tuple[set[str], list[dict[str, Any]], dict[str, Any], list[dict[str, Any]]]:
    root = root.resolve(strict=True)
    nodes: set[str] = set()
    edges: list[dict[str, Any]] = []
    findings: list[dict[str, Any]] = []
    with tempfile.TemporaryDirectory(prefix="pm-review-go-") as temporary:
        temp = Path(temporary)
        (temp / "home").mkdir()
        (temp / "cache").mkdir()
        module_cache_value = os.environ.get("GOMODCACHE", "")
        if module_cache_value:
            module_cache = Path(module_cache_value).resolve(strict=True)
        else:
            probe = subprocess.run(
                ["go", "env", "GOMODCACHE"],
                check=False,
                capture_output=True,
                text=True,
                env={"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "HOME": str(Path.home()), "GOENV": "off"},
            )
            if probe.returncode != 0 or not probe.stdout.strip():
                findings.append(finding("impact_go", "authoritative go list cannot locate a pre-populated module cache"))
                return nodes, edges, {"status": "blocked", "reason": "module cache unavailable"}, findings
            try:
                module_cache = Path(probe.stdout.strip()).resolve(strict=True)
            except OSError:
                module_cache = temp / "modcache"
                module_cache.mkdir()
        if not module_cache.is_dir():
            module_cache = temp / "modcache"
            module_cache.mkdir(exist_ok=True)
        env = {
            "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
            "HOME": str(temp / "home"),
            "GOCACHE": str(temp / "cache"),
            "GOMODCACHE": str(module_cache),
            "GOPROXY": "off",
            "GOSUMDB": "off",
            "GONOSUMDB": "*",
            "GOTOOLCHAIN": "local",
            "GOWORK": "off",
            "GOENV": "off",
            "GIT_CONFIG_GLOBAL": os.devnull,
            "GIT_CONFIG_SYSTEM": os.devnull,
            "GIT_TERMINAL_PROMPT": "0",
        }
        try:
            proc = subprocess.run(
                ["go", "list", "-mod=mod", "-json", "-deps", "-test", "./..."],
                cwd=root,
                env=env,
                check=False,
                capture_output=True,
                text=True,
                timeout=timeout_seconds,
            )
        except (OSError, subprocess.TimeoutExpired) as exc:
            findings.append(finding("impact_go", f"authoritative go list failed or timed out: {exc}"))
            return nodes, edges, {"status": "blocked", "reason": str(exc)}, findings
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        findings.append(finding("impact_go", f"authoritative go list failed: {detail[:1000]}"))
        return nodes, edges, {"status": "blocked", "reason": detail[:1000]}, findings

    packages = decode_json_stream(proc.stdout)
    local: dict[str, dict[str, Any]] = {}
    for package in packages:
        directory = package.get("Dir", "")
        if not isinstance(directory, str) or not directory.startswith(str(root)) or package.get("Standard"):
            continue
        import_path = str(package.get("ImportPath", "")).split(" [", 1)[0]
        if not import_path or import_path.endswith(".test"):
            continue
        record = local.setdefault(import_path, {"Dir": directory, "files": defaultdict(set), "imports": set()})
        for field in ("GoFiles", "CgoFiles", "TestGoFiles", "XTestGoFiles", "IgnoredGoFiles"):
            record["files"][field].update(package.get(field, []) or [])
        record["imports"].update(package.get("Imports", []) or [])
        record["imports"].update(package.get("TestImports", []) or [])
        record["imports"].update(package.get("XTestImports", []) or [])

    for import_path, package in sorted(local.items()):
        package_node = "go-package:" + import_path
        nodes.add(package_node)
        fields = (
            ("GoFiles", "go_contains", "active"),
            ("CgoFiles", "go_contains", "active"),
            ("TestGoFiles", "go_test", "active"),
            ("XTestGoFiles", "go_test", "active"),
            ("IgnoredGoFiles", "platform_variant", "unknown"),
        )
        for field, relation, certainty in fields:
            for name in sorted(package["files"].get(field, set())):
                absolute = Path(package["Dir"]) / name
                try:
                    relative = absolute.relative_to(root).as_posix()
                except ValueError:
                    continue
                validate_relative_path(relative, "go list file")
                nodes.add(relative)
                edges.append(stable_edge(package_node, relative, relation, "go_list", field, certainty))
                edges.append(stable_edge(relative, package_node, "go_member_of", "go_list", field, certainty))
        for dependency in sorted(package["imports"]):
            normalized = str(dependency).split(" [", 1)[0]
            if normalized in local:
                edges.append(stable_edge(package_node, "go-package:" + normalized, "go_imports", "go_list", "Imports", "active"))
    context = {
        "status": "complete",
        "command": ["go", "list", "-mod=mod", "-json", "-deps", "-test", "./..."],
        "go_version": subprocess.run(["go", "version"], check=False, capture_output=True, text=True).stdout.strip(),
        "goos": subprocess.run(["go", "env", "GOOS"], env={"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "GOENV": "off"}, check=False, capture_output=True, text=True).stdout.strip(),
        "goarch": subprocess.run(["go", "env", "GOARCH"], env={"PATH": os.environ.get("PATH", "/usr/bin:/bin"), "GOENV": "off"}, check=False, capture_output=True, text=True).stdout.strip(),
        "network_policy": "GOPROXY=off; GOSUMDB=off; GOTOOLCHAIN=local; scrubbed temporary HOME/GOCACHE; pre-populated module source cache",
        "module_cache_policy": "existing extracted module sources may be read while go cache metadata stays local/offline; -mod=mod writes only inside a disposable exact-commit snapshot",
    }
    unique = {edge["id"]: edge for edge in edges}
    return nodes, sorted(unique.values(), key=lambda item: item["id"]), context, findings


def clone_commit_snapshot(root: Path, commit: str) -> tempfile.TemporaryDirectory[str]:
    temporary = tempfile.TemporaryDirectory(prefix="pm-review-commit-")
    snapshot = Path(temporary.name) / "repo"
    proc = subprocess.run(
        [
            "git",
            "-c",
            "protocol.file.allow=always",
            "clone",
            "--no-hardlinks",
            "--no-local",
            "--quiet",
            "--no-checkout",
            str(root),
            str(snapshot),
        ],
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        temporary.cleanup()
        raise ReviewSystemError(f"exact-commit snapshot clone failed: {(proc.stderr or proc.stdout).strip()[:500]}")
    proc = subprocess.run(
        ["git", "-C", str(snapshot), "checkout", "--quiet", "--detach", commit],
        check=False,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        temporary.cleanup()
        raise ReviewSystemError(f"exact-commit snapshot checkout failed: {(proc.stderr or proc.stdout).strip()[:500]}")
    return temporary


def build_impact_index(
    root: Path,
    base: str,
    head: str,
    changed: list[str],
    config: dict[str, Any],
) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    settings = config["impact_graph"]
    findings: list[dict[str, Any]] = []
    tracked_raw = run_git(root, ["ls-tree", "-r", "--name-only", "-z", head, "--"])
    tracked = {path for path in tracked_raw.split("\0") if path}
    prefixes = tuple(settings.get("index_prefixes", []))
    go_prefixes = tuple(settings.get("go_index_prefixes", []))
    canonical = set(config.get("canonical_roots", []))
    explicit = set(config.get("explicit_reference_files", []))
    configured_endpoints: set[str] = set()
    for record in config.get("authorities", []):
        configured_endpoints.add(record.get("authoritative_path", ""))
        for field in ("writers", "readers", "mirrors"):
            configured_endpoints.update(record.get(field, []))
    for record in config.get("configured_relationships", []):
        configured_endpoints.update((record.get("source", ""), record.get("target", "")))
    configured_endpoints.discard("")

    include_go = any(path.endswith(".go") for path in changed)
    selected = {
        path
        for path in tracked
        if (
            PurePosixPath(path).suffix in INDEX_SUFFIXES
            and (path.startswith(prefixes) or path in canonical or path in explicit or path in configured_endpoints or path in changed)
        )
        or (include_go and path.endswith(".go") and path.startswith(go_prefixes))
    }
    selected.update(path for path in changed if PurePosixPath(path).suffix in INDEX_SUFFIXES)
    selected.update(path for path in configured_endpoints if path in tracked)
    selected.update(path for path in explicit if path in tracked)

    max_index_files = int(settings["max_index_files"])
    if len(selected) > max_index_files:
        findings.append(finding("impact_graph_bound", f"index file bound exceeded before reads: {len(selected)} > {max_index_files}"))
        selected = set(sorted(selected)[:max_index_files])
    max_nodes = int(settings["max_nodes"])
    if len(selected) > max_nodes:
        findings.append(finding("impact_graph_bound", f"graph node bound exceeded before reads: {len(selected)} > {max_nodes}"))
        selected = set(sorted(selected)[:max_nodes])

    nodes: set[str] = set(selected)
    node_types: dict[str, str] = {path: "file" for path in selected}
    file_bytes: dict[str, int] = {}
    edges: list[dict[str, Any]] = []
    edge_ids: set[str] = set()
    total_bytes = 0
    parsed: set[str] = set()
    pending = deque(sorted(selected))
    ignored = tuple(config.get("ignored_reference_prefixes", []))
    prohibited = set(config.get("prohibited_active_targets", []))
    max_edges = int(settings["max_edges"])
    index_stopped = False

    def include_node(relative: str, source: str) -> bool:
        if relative in nodes:
            return True
        if len(nodes) >= max_nodes or len(nodes) >= max_index_files:
            findings.append(finding("impact_graph_bound", f"index expansion stopped before {relative}: node/file bound", source))
            return False
        nodes.add(relative)
        selected.add(relative)
        node_types[relative] = "file"
        pending.append(relative)
        return True

    def append_edge(edge: dict[str, Any]) -> bool:
        nonlocal index_stopped
        if edge["id"] in edge_ids:
            return True
        if len(edges) >= max_edges:
            findings.append(finding("impact_graph_bound", f"graph edge bound exceeded during construction: > {max_edges}"))
            index_stopped = True
            return False
        edge_ids.add(edge["id"])
        edges.append(edge)
        return True

    while pending and not index_stopped:
        relative = pending.popleft()
        if relative in parsed:
            continue
        parsed.add(relative)
        if relative not in tracked:
            node_types[relative] = "deleted_file"
            continue
        try:
            path = resolve_safe(root, relative)
            size = path.stat().st_size
            if total_bytes + size > int(settings["max_index_bytes"]):
                findings.append(
                    finding(
                        "impact_graph_bound",
                        f"index byte bound would be exceeded before reading {relative}: {total_bytes + size} > {settings['max_index_bytes']}",
                        relative,
                    )
                )
                index_stopped = True
                break
            text = path.read_text()
        except (ReviewSystemError, OSError, UnicodeDecodeError) as exc:
            findings.append(finding("impact_index", f"cannot index review-relevant file: {exc}", relative))
            continue
        total_bytes += size
        file_bytes[relative] = size
        parsed_edges, parser_findings = typed_file_edges(relative, text, tracked | nodes)
        findings.extend(parser_findings)
        for edge in parsed_edges:
            target = edge["target"]
            if target.startswith(ignored):
                continue
            if edge["certainty"] != "inactive":
                if target not in tracked:
                    findings.append(
                        finding(
                            "impact_graph",
                            f"{edge['certainty']} impact target is unresolved: {target}",
                            relative,
                        )
                    )
                elif not include_node(target, relative):
                    continue
                if target in prohibited:
                    findings.append(
                        finding(
                            "impact_graph",
                            f"{edge['certainty']} impact edge reaches prohibited target {target} ({edge['id']})",
                            relative,
                        )
                    )
            if not append_edge(edge):
                break

    configured_edges: list[dict[str, Any]] = []
    for record in config.get("authorities", []):
        state = record.get("authoritative_path", "")
        for writer in record.get("writers", []):
            configured_edges.append(stable_edge(writer, state, "authority_writes", "config", f"authority:{record.get('id')}:writer", "active"))
        for reader in record.get("readers", []):
            configured_edges.append(stable_edge(reader, state, "authority_reads", "config", f"authority:{record.get('id')}:reader", "active"))
        for mirror in record.get("mirrors", []):
            configured_edges.append(stable_edge(state, mirror, "authority_mirror", "config", f"authority:{record.get('id')}:mirror", "active"))
    for record in config.get("configured_relationships", []):
        configured_edges.append(
            stable_edge(
                record.get("source", ""),
                record.get("target", ""),
                record.get("relation", "configured_relation"),
                "config",
                record.get("reason", "configured_relationship"),
                record.get("certainty", "unknown"),
            )
        )
    for edge in configured_edges:
        for endpoint in (edge["source"], edge["target"]):
            if endpoint not in tracked:
                findings.append(finding("impact_graph", f"configured {edge['certainty']} endpoint is unresolved: {endpoint}", edge["source"]))
            else:
                include_node(endpoint, edge["source"])
        if edge["target"] in prohibited and edge["certainty"] != "inactive":
            findings.append(finding("impact_graph", f"configured edge reaches prohibited target {edge['target']} ({edge['id']})", edge["source"]))
        append_edge(edge)

    go_context: dict[str, Any] = {"status": "not_needed", "reason": "no changed Go file"}
    if any(path.endswith(".go") for path in changed):
        try:
            head_snapshot_tmp = clone_commit_snapshot(root, head)
            with head_snapshot_tmp:
                head_snapshot = Path(head_snapshot_tmp.name) / "repo"
                go_nodes, go_edges, go_context, go_findings = go_impact_edges(
                    head_snapshot, settings["go_command_timeout_seconds"]
                )
        except ReviewSystemError as exc:
            go_nodes, go_edges = set(), []
            go_context = {"status": "blocked", "reason": str(exc)}
            go_findings = [finding("impact_go", str(exc))]
        findings.extend(go_findings)
        deleted_go = sorted(path for path in changed if path.endswith(".go") and path not in tracked)
        if deleted_go:
            try:
                snapshot_tmp = clone_commit_snapshot(root, base)
                with snapshot_tmp:
                    snapshot = Path(snapshot_tmp.name) / "repo"
                    base_nodes, base_edges, base_context, base_findings = go_impact_edges(
                        snapshot, settings["go_command_timeout_seconds"]
                    )
                go_nodes.update(base_nodes)
                for edge in base_edges:
                    base_edge = stable_edge(
                        edge["source"],
                        edge["target"],
                        edge["relation"],
                        edge["parser"],
                        "base_deleted_context:" + edge["reason"],
                        edge["certainty"],
                        edge.get("line", 0),
                    )
                    go_edges.append(base_edge)
                findings.extend(base_findings)
                go_context = {**go_context, "base_deleted_file_context": base_context, "deleted_go_files": deleted_go}
            except ReviewSystemError as exc:
                findings.append(finding("impact_go", f"deleted Go base context failed: {exc}"))
        for node in sorted(go_nodes):
            if node not in nodes and len(nodes) >= max_nodes:
                findings.append(finding("impact_graph_bound", f"Go node bound exceeded during construction: > {max_nodes}"))
                break
            nodes.add(node)
            node_types[node] = "go_package" if node.startswith("go-package:") else ("deleted_file" if node not in tracked else "file")
        for edge in sorted(go_edges, key=lambda item: item["id"]):
            append_edge(edge)

    unique = {edge["id"]: edge for edge in edges}
    indexed_edges = sorted(unique.values(), key=lambda item: item["id"])
    for edge in indexed_edges:
        for field in ("source", "target"):
            endpoint = edge[field]
            try:
                if not endpoint.startswith("go-package:"):
                    validate_relative_path(endpoint, f"impact edge {field}")
            except ReviewSystemError as exc:
                findings.append(finding("impact_path", str(exc), edge.get("source")))
        if edge["certainty"] != "inactive":
            target = edge["target"]
            if target not in nodes:
                findings.append(finding("impact_graph", f"{edge['certainty']} indexed target is absent: {target}", edge["source"]))
            elif node_types.get(target) == "deleted_file" and not edge.get("reason", "").startswith("base_deleted_context:"):
                findings.append(finding("impact_graph", f"{edge['certainty']} indexed target was deleted: {target}", edge["source"]))

    return {
        "nodes": nodes,
        "node_types": node_types,
        "edges": indexed_edges,
        "tracked": tracked,
        "index_files": sorted(parsed),
        "index_bytes": total_bytes,
        "file_bytes": file_bytes,
        "go_context": go_context,
    }, findings

def traversal_direction(edge: dict[str, Any], reverse: bool) -> str:
    relation = edge["relation"]
    if relation in {"authority_mirror", "platform_variant", "go_test", "fixture", "sibling_variant"}:
        return "lateral"
    if relation.startswith("temporal_") or relation in {"migration", "restart_resume", "version_invalidation"}:
        return "temporal"
    return "upstream" if reverse else "downstream"


def compile_impact_graph(
    root: Path,
    base: str,
    head: str,
    changed: list[str],
    config: dict[str, Any],
) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    index, findings = build_impact_index(root, base, head, changed, config)
    settings = config["impact_graph"]
    nodes: set[str] = index["nodes"]
    forward: dict[str, list[dict[str, Any]]] = defaultdict(list)
    reverse: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for edge in index["edges"]:
        forward[edge["source"]].append(edge)
        reverse[edge["target"]].append(edge)
    for mapping in (forward, reverse):
        for key in mapping:
            mapping[key].sort(key=lambda item: item["id"])

    seeds = sorted(set(config.get("canonical_roots", [])) | set(changed))
    for seed in seeds:
        if seed not in nodes:
            nodes.add(seed)
            index["node_types"][seed] = "deleted_file" if seed in changed else "missing_seed"
            if seed not in changed:
                findings.append(finding("impact_graph", "canonical seed is absent from indexed universe", seed))

    policy = settings.get("relation_policy", {})
    default_policy = settings.get(
        "default_relation_policy",
        {direction: settings["max_depth"] for direction in ("upstream", "downstream", "lateral", "temporal")},
    )
    policy_keys = sorted(
        {
            (edge["relation"], direction)
            for edge in index["edges"]
            for direction in ("upstream", "downstream", "lateral", "temporal")
            if int(policy.get(edge["relation"], default_policy).get(direction, 0)) > 0
        }
    )
    key_index = {key: position for position, key in enumerate(policy_keys)}
    zero_state = tuple(0 for _ in policy_keys)
    pending = deque((seed, 0, zero_state) for seed in seeds)
    accepted_states: dict[str, list[tuple[int, ...]]] = defaultdict(list)
    seen_nodes: dict[str, int] = {}
    traversed: dict[str, dict[str, Any]] = {}
    states = 0
    bound_reasons: list[str] = []

    def dominated(node: str, candidate: tuple[int, ...]) -> bool:
        existing = accepted_states[node]
        if any(all(left <= right for left, right in zip(state, candidate)) for state in existing):
            return True
        accepted_states[node] = [
            state for state in existing if not all(left <= right for left, right in zip(candidate, state))
        ]
        accepted_states[node].append(candidate)
        return False

    while pending:
        current, depth, state = pending.popleft()
        states += 1
        if states > settings["max_traversal_states"]:
            bound_reasons.append(f"traversal states exceed {settings['max_traversal_states']}")
            break
        if dominated(current, state):
            continue
        seen_nodes[current] = min(depth, seen_nodes.get(current, depth))
        candidates = [(edge, False, edge["target"]) for edge in forward.get(current, [])]
        candidates.extend((edge, True, edge["source"]) for edge in reverse.get(current, []))
        for edge, reversed_edge, neighbor in sorted(candidates, key=lambda item: (item[0]["id"], item[1])):
            if edge["certainty"] == "inactive":
                continue
            direction = traversal_direction(edge, reversed_edge)
            relation_policy = policy.get(edge["relation"], default_policy)
            relation_limit = int(relation_policy.get(direction, 0))
            key = (edge["relation"], direction)
            position = key_index.get(key)
            used = state[position] if position is not None else 0
            if relation_limit <= 0 or used >= relation_limit:
                continue
            if depth >= settings["max_depth"]:
                bound_reasons.append(f"depth frontier at {current} via {edge['id']} exceeds {settings['max_depth']}")
                continue
            if neighbor not in nodes:
                findings.append(finding("impact_graph", f"{edge['certainty']} impact target is unresolved: {neighbor}", current))
                continue
            next_state_values = list(state)
            if position is not None:
                next_state_values[position] += 1
            next_state = tuple(next_state_values)
            next_depth = depth + 1
            record = traversed.setdefault(
                edge["id"],
                {**edge, "traversal_directions": set(), "minimum_depth": next_depth},
            )
            record["traversal_directions"].add(direction)
            record["minimum_depth"] = min(record["minimum_depth"], next_depth)
            pending.append((neighbor, next_depth, next_state))
        if len(seen_nodes) > settings["max_impact_files"]:
            bound_reasons.append(f"impact file/node count exceeds {settings['max_impact_files']}")
            break
        if len(traversed) > settings["max_impact_edges"]:
            bound_reasons.append(f"impact edge count exceeds {settings['max_impact_edges']}")
            break

    impact_edges: list[dict[str, Any]] = []
    for edge in traversed.values():
        edge["traversal_directions"] = sorted(edge["traversal_directions"])
        impact_edges.append(edge)
    impact_edges.sort(key=lambda item: item["id"])
    impact_files = sorted(node for node in seen_nodes if not node.startswith("go-package:"))
    virtual_nodes = sorted(node for node in seen_nodes if node.startswith("go-package:"))
    if bound_reasons:
        for reason in sorted(set(bound_reasons)):
            findings.append(finding("impact_graph_bound", reason))
    certainty_counts: dict[str, int] = defaultdict(int)
    relation_counts: dict[str, int] = defaultdict(int)
    for edge in index["edges"]:
        certainty_counts[edge["certainty"]] += 1
        relation_counts[edge["relation"]] += 1
    result = {
        "schema_version": IMPACT_SCHEMA,
        "status": "blocked" if findings else "complete",
        "seed_files": seeds,
        "universe": {
            "index_files": index["index_files"],
            "index_file_count": len(index["index_files"]),
            "index_bytes": index["index_bytes"],
            "file_bytes": index["file_bytes"],
            "excluded_policy": "tracked supported artifacts outside configured prefixes unless changed/canonical/explicit/configured/referenced",
        },
        "indexed_edge_counts": {
            "total": len(index["edges"]),
            "by_certainty": dict(sorted(certainty_counts.items())),
            "by_relation": dict(sorted(relation_counts.items())),
        },
        "indexed_edges": index["edges"],
        "files": impact_files,
        "virtual_nodes": virtual_nodes,
        "edges": impact_edges,
        "go_context": index["go_context"],
        "bounds": {"hit": bool(bound_reasons), "reasons": sorted(set(bound_reasons)), "limits": settings},
        "algorithm": "materialized forward/reverse adjacency; deterministic multi-source relation/direction-budget BFS with dominance-pruned policy state",
        "precision_scope": "practical file/package impact; no symbol-level call/data-flow claim",
    }
    return result, findings

def classify_domain(path: str, config: dict[str, Any]) -> str:
    for rule in config.get("domain_rules", []):
        if path_matches(path, rule.get("patterns", [])):
            return rule["domain"]
    return "implementation_test"


def changed_files(root: Path, base: str, head: str) -> tuple[list[str], dict[str, int]]:
    names = run_git(root, ["diff", "--no-renames", "--name-only", "-z", f"{base}...{head}", "--"])
    files = [name for name in names.split("\0") if name]
    line_counts: dict[str, int] = {name: 0 for name in files}
    numstat = run_git(root, ["diff", "--no-renames", "--numstat", f"{base}...{head}", "--"])
    for line in numstat.splitlines():
        parts = line.split("\t", 2)
        if len(parts) != 3:
            continue
        additions, deletions, name = parts
        if additions == "-" or deletions == "-":
            line_counts[name] = 0
        else:
            line_counts[name] = int(additions) + int(deletions)
    for path in files:
        validate_relative_path(path, "changed path")
    return sorted(files), line_counts


def chunks(values: list[str], size: int) -> list[list[str]]:
    return [values[index : index + size] for index in range(0, len(values), size)]


def exact_blob_sizes(root: Path, base: str, head: str, paths: Iterable[str]) -> dict[str, int]:
    sizes: dict[str, int] = {}
    for relative in sorted(set(paths)):
        if relative.startswith("go-package:"):
            continue
        validate_relative_path(relative, "packet context path")
        for commit in (head, base):
            proc = subprocess.run(
                ["git", "-C", str(root), "cat-file", "-s", f"{commit}:{relative}"],
                check=False,
                capture_output=True,
                text=True,
            )
            if proc.returncode == 0 and proc.stdout.strip().isdigit():
                sizes[relative] = int(proc.stdout.strip())
                break
        else:
            raise ReviewSystemError(f"packet context has no exact base/head blob: {relative}")
    return sizes


def build_packets(
    base: str,
    head: str,
    head_tree: str,
    files: list[str],
    line_counts: dict[str, int],
    domains: dict[str, str],
    closure_files: list[str],
    authority_files: list[str],
    impact_files: list[str],
    impact_edges: list[dict[str, Any]],
    blob_sizes: dict[str, int],
    config: dict[str, Any],
) -> tuple[str, list[dict[str, Any]], list[dict[str, Any]]]:
    thresholds = config["thresholds"]
    graph_limits = config["impact_graph"]
    domain_values = sorted(set(domains.values()))
    changed_lines = sum(line_counts.values())
    combined = (
        len(files) <= thresholds["combined_max_files"]
        and changed_lines <= thresholds["combined_max_non_generated_lines"]
        and len(domain_values) <= thresholds["combined_max_domains"]
    )
    selection = "combined" if combined else "split"
    packets: list[dict[str, Any]] = []
    problems: list[dict[str, Any]] = []
    impact_edge_by_id = {edge["id"]: edge for edge in impact_edges}
    target_tokens = int(thresholds["packet_target_tokens"])

    def file_tokens(paths: Iterable[str]) -> int:
        return sum(max(1, math.ceil(blob_sizes[path] / 4)) for path in sorted(set(paths)))

    def edge_context_tokens(paths: Iterable[str]) -> int:
        maximum = int(graph_limits.get("edge_context_max_bytes_per_file", 8192))
        return sum(max(1, math.ceil(min(blob_sizes[path], maximum) / 4)) for path in sorted(set(paths)))

    def estimate(
        changed: list[str],
        closure: list[str],
        authority: list[str],
        impact: list[str],
        edge_context: list[str],
        edge_ids: list[str],
    ) -> int:
        changed_cost = sum(line_counts.get(path, 0) for path in changed) * int(
            thresholds["estimated_tokens_per_changed_line"]
        )
        full_context = set(closure) | set(authority) | set(impact)
        context_cost = file_tokens(full_context) + edge_context_tokens(set(edge_context) - full_context)
        metadata_bytes = len(
            json.dumps([impact_edge_by_id[edge_id] for edge_id in edge_ids], sort_keys=True, separators=(",", ":")).encode()
        )
        return changed_cost + context_cost + math.ceil(metadata_bytes / 4)

    def append_packet(
        role: str,
        changed: list[str],
        closure: list[str],
        authority: list[str],
        impact: list[str] | None = None,
        impact_edge_ids: list[str] | None = None,
        edge_context_files: list[str] | None = None,
    ) -> None:
        impact = sorted(impact or [])
        impact_edge_ids = sorted(impact_edge_ids or [])
        edge_context_files = sorted(edge_context_files or [])
        packet_number = 1 + sum(1 for packet in packets if packet["role"] == role)
        estimated = estimate(changed, closure, authority, impact, edge_context_files, impact_edge_ids)
        overflow = estimated > target_tokens
        if overflow:
            problems.append(
                finding(
                    "packet_overflow",
                    f"{role} packet estimate {estimated} exceeds target {target_tokens}",
                )
            )
        packets.append(
            {
                "schema_version": PACKET_SCHEMA,
                "packet_id": f"{role}-{packet_number:02d}",
                "role": role,
                "exact_base_sha": base,
                "exact_head_sha": head,
                "exact_head_tree": head_tree,
                "changed_files": sorted(changed),
                "closure_files": sorted(closure),
                "authority_files": sorted(authority),
                "impact_files": impact,
                "impact_edge_ids": impact_edge_ids,
                "impact_edges": [impact_edge_by_id[edge_id] for edge_id in impact_edge_ids],
                "edge_context_files": edge_context_files,
                "invariants": config["packet_invariants"].get(role, config["packet_invariants"]["combined"]),
                "context": {
                    "target_tokens": target_tokens,
                    "estimated_tokens": estimated,
                    "estimation": "exact_head_bytes_conservative",
                    "bytes_per_token_upper_bound": 4,
                    "edge_context_mode": "provenance line plus bounded exact-blob excerpts",
                    "edge_context_max_bytes_per_file": int(graph_limits.get("edge_context_max_bytes_per_file", 8192)),
                    "overflow": overflow,
                    "truncated": False,
                },
                "required_response_fields": [
                    "reviewed_files",
                    "closure_files",
                    "authority_files",
                    "impact_files",
                    "impact_edge_ids",
                    "edge_context_files",
                    "invariants",
                    "review_behaviors",
                    "experiments",
                    "no_experiment_reason",
                    "unreviewed_files",
                    "findings",
                    "residual_risk",
                    "context",
                    "wall_clock_ms",
                ],
            }
        )

    def greedy_file_groups(values: list[str], maximum_files: int) -> list[list[str]]:
        groups: list[list[str]] = []
        current: list[str] = []
        for value in sorted(values):
            proposed = [*current, value]
            if current and (len(proposed) > maximum_files or file_tokens(proposed) > target_tokens):
                groups.append(current)
                current = [value]
            else:
                current = proposed
            if file_tokens(current) > target_tokens:
                problems.append(finding("packet_overflow", f"single context file cannot fit packet target: {value}"))
        if current:
            groups.append(current)
        return groups

    if combined:
        append_packet("combined", files, [], [])
    else:
        by_role: dict[str, list[str]] = {}
        for path in files:
            by_role.setdefault(domains[path], []).append(path)
        for role in ("architecture_reference", "authority_workflow_state", "implementation_test"):
            for part in chunks(sorted(by_role.get(role, [])), thresholds["packet_max_changed_files"]):
                append_packet(role, part, [], [])

    for part in greedy_file_groups(closure_files, int(thresholds["packet_max_context_files"])):
        append_packet("architecture_reference", [], part, [])
    for part in greedy_file_groups(authority_files, int(thresholds["packet_max_context_files"])):
        append_packet("authority_workflow_state", [], [], part)

    edge_groups: list[list[dict[str, Any]]] = []
    current_edges: list[dict[str, Any]] = []
    for edge in sorted(impact_edges, key=lambda item: item["id"]):
        proposed = [*current_edges, edge]
        proposed_ids = [item["id"] for item in proposed]
        endpoints = sorted(
            {
                endpoint
                for item in proposed
                for endpoint in (item["source"], item["target"])
                if not endpoint.startswith("go-package:")
            }
        )
        proposed_estimate = estimate([], [], [], [], endpoints, proposed_ids)
        if current_edges and (
            len(proposed) > int(graph_limits["packet_max_impact_edges"]) or proposed_estimate > target_tokens
        ):
            edge_groups.append(current_edges)
            current_edges = [edge]
        else:
            current_edges = proposed
        single_ids = [item["id"] for item in current_edges]
        single_endpoints = sorted(
            {
                endpoint
                for item in current_edges
                for endpoint in (item["source"], item["target"])
                if not endpoint.startswith("go-package:")
            }
        )
        if estimate([], [], [], [], single_endpoints, single_ids) > target_tokens:
            problems.append(finding("packet_overflow", f"atomic impact edge neighborhood cannot fit: {edge['id']}"))
    if current_edges:
        edge_groups.append(current_edges)

    for group in edge_groups:
        edge_ids = [edge["id"] for edge in group]
        endpoints = sorted(
            {
                endpoint
                for edge in group
                for endpoint in (edge["source"], edge["target"])
                if not endpoint.startswith("go-package:")
            }
        )
        append_packet("impact_graph", [], [], [], [], edge_ids, endpoints)

    for part in greedy_file_groups(sorted(impact_files), int(graph_limits["packet_max_impact_files"])):
        append_packet("impact_graph", [], [], [], part, [], [])

    maximum_packets = int(graph_limits.get("max_packets", 64))
    if len(packets) > maximum_packets:
        problems.append(finding("packet_overflow", f"packet count {len(packets)} exceeds {maximum_packets}"))
    if any(packet["context"]["overflow"] for packet in packets) or problems:
        selection = "blocked"
    return selection, packets, problems

def command_detect(args: argparse.Namespace) -> int:
    try:
        case = load_case(Path(args.input), args.case_id)
        emit_json(detect(case, args.mode))
        return 0
    except ReviewSystemError as exc:
        print(f"pm review detect error: {exc}", file=sys.stderr)
        return 2


def command_observe(args: argparse.Namespace) -> int:
    try:
        document = load_json(Path(args.input))
        started = time.perf_counter_ns()
        observations = [detect(case, args.mode) for case in document.get("cases", [])]
        emit_json(
            {
                "schema_version": "polymetrics.ai/pm-review-observations/v1",
                "mode": args.mode,
                "input": str(args.input),
                "observations": observations,
                "wall_clock_ms": round((time.perf_counter_ns() - started) / 1_000_000, 6),
            }
        )
        return 0
    except ReviewSystemError as exc:
        print(f"pm review observe error: {exc}", file=sys.stderr)
        return 2


def metric_ratio(numerator: int, denominator: int) -> float | None:
    return round(numerator / denominator, 6) if denominator else None


def command_score(args: argparse.Namespace) -> int:
    try:
        observed = load_json(Path(args.observations))
        oracle = load_json(Path(args.oracle)).get("cases", {})
        rows = observed.get("observations", [])
        tp = fp = fn = tn = 0
        decision_total = decision_correct = 0
        suites: dict[str, dict[str, int]] = {}
        categories: dict[str, dict[str, int]] = {}
        for row in rows:
            expected = oracle.get(row["case_id"])
            if expected is None:
                raise ReviewSystemError(f"oracle missing case {row['case_id']}")
            suite = row.get("suite", "unknown")
            suites.setdefault(suite, {"tp": 0, "fp": 0, "fn": 0, "tn": 0})
            category = expected.get("category", "threshold")
            categories.setdefault(category, {"tp": 0, "fp": 0, "fn": 0, "tn": 0})
            if expected["expected"] == "decision":
                decision_total += 1
                if row.get("decision") == expected["decision"]:
                    decision_correct += 1
                continue
            detected = bool(row.get("findings"))
            if expected["expected"] == "finding" and detected:
                tp += 1
                suites[suite]["tp"] += 1
                categories[category]["tp"] += 1
            elif expected["expected"] == "finding":
                fn += 1
                suites[suite]["fn"] += 1
                categories[category]["fn"] += 1
            elif detected:
                fp += 1
                suites[suite]["fp"] += 1
                categories[category]["fp"] += 1
            else:
                tn += 1
                suites[suite]["tn"] += 1
                categories[category]["tn"] += 1
        emit_json(
            {
                "schema_version": MEASUREMENT_SCHEMA,
                "mode": observed.get("mode"),
                "counts": {"true_positive": tp, "false_positive": fp, "false_negative": fn, "true_negative": tn},
                "first_round_recall": metric_ratio(tp, tp + fn),
                "first_round_precision": metric_ratio(tp, tp + fp),
                "escaped_defects": fn,
                "escaped_defect_rate": metric_ratio(fn, tp + fn),
                "false_positive_rate": metric_ratio(fp, fp + tn),
                "threshold_decisions": {"correct": decision_correct, "total": decision_total, "accuracy": metric_ratio(decision_correct, decision_total)},
                "suites": suites,
                "categories": categories,
                "exact_version_invalidations_detected": sum(
                    1 for row in rows for item in row.get("findings", []) if item.get("category") == "stale_evidence"
                ),
                "context_overflow_cases_detected": sum(
                    1 for row in rows for item in row.get("findings", []) if item.get("category") == "packet_overflow"
                ),
                "review_rounds": {"status": "unavailable", "reason": "fixture detector does not run correction loops"},
                "wall_clock_ms": observed.get("wall_clock_ms"),
                "tokens": {"status": "unavailable", "reason": "deterministic detector; no model call"},
                "cost": {"status": "unavailable", "reason": "deterministic detector; no model call"},
                "claim_scope": "deterministic fixture preflight only; not hosted-model or prospective production recall",
            }
        )
        return 0
    except ReviewSystemError as exc:
        print(f"pm review score error: {exc}", file=sys.stderr)
        return 2


def command_compile(args: argparse.Namespace) -> int:
    try:
        root = Path(args.repo_root).resolve(strict=True)
        if not (root / ".git").exists() and not run_git(root, ["rev-parse", "--git-dir"]).strip():
            raise ReviewSystemError("repo root is not a Git worktree")
        config_path = resolve_safe(root, args.config)
        config = load_json(config_path)
        if config.get("schema_version") != CONFIG_SCHEMA:
            raise ReviewSystemError(
                f"review-system config migration required: {config.get('schema_version')!r} != {CONFIG_SCHEMA!r}"
            )
        scope_path = resolve_safe(root, args.scope)
        scope = load_json(scope_path)
        if scope.get("schema_version") != SCOPE_SCHEMA:
            raise ReviewSystemError(
                f"review scope migration required: {scope.get('schema_version')!r} != {SCOPE_SCHEMA!r}"
            )
        for field in ("allowed_changed_paths", "forbidden_changed_paths"):
            if not isinstance(scope.get(field), list) or not all(isinstance(item, str) and item for item in scope[field]):
                raise ReviewSystemError(f"review scope {field} must be a non-empty string list")
        base = validate_sha(args.base, "exact base")
        head = validate_sha(args.head, "exact head")
        run_git(root, ["cat-file", "-e", f"{base}^{{commit}}"])
        run_git(root, ["cat-file", "-e", f"{head}^{{commit}}"])
        head_tree = run_git(root, ["rev-parse", f"{head}^{{tree}}"]).strip()
        merge_base = run_git(root, ["merge-base", base, head]).strip()
        if merge_base != base:
            raise ReviewSystemError(f"exact base is not the candidate merge base: {merge_base} != {base}")
        current = run_git(root, ["rev-parse", "HEAD"]).strip()
        if current != head:
            raise ReviewSystemError(f"worktree HEAD drift: {current} != {head}")
        worktree_drift = run_git(root, ["--no-optional-locks", "status", "--porcelain", "--untracked-files=all"]).strip()
        if worktree_drift:
            raise ReviewSystemError("worktree differs from exact reviewed head, including untracked files")
        files, line_counts = changed_files(root, base, head)
        findings: list[dict[str, Any]] = []
        for path in files:
            if path_matches(path, scope.get("forbidden_changed_paths", [])):
                findings.append(finding("changed_path_scope", "changed path is forbidden", path))
            elif not path_matches(path, scope.get("allowed_changed_paths", [])):
                findings.append(finding("changed_path_scope", "changed path is outside the positive allowlist", path))
        closure, edges, closure_findings = compile_closure(root, config)
        findings.extend(closure_findings)
        authority, authority_files, authority_findings = authority_inventory(root, config)
        findings.extend(authority_findings)
        impact_graph, impact_findings = compile_impact_graph(root, base, head, files, config)
        findings.extend(impact_findings)
        domains = {path: classify_domain(path, config) for path in files}
        closure_context = sorted(set(closure) - set(files))
        authority_context = sorted(set(authority_files) - set(files))
        packet_context_paths = set(files) | set(closure_context) | set(authority_context) | set(impact_graph["files"])
        for edge in impact_graph["edges"]:
            packet_context_paths.update(
                endpoint for endpoint in (edge["source"], edge["target"]) if not endpoint.startswith("go-package:")
            )
        blob_sizes = exact_blob_sizes(root, base, head, packet_context_paths)
        selection, packets, packet_findings = build_packets(
            base,
            head,
            head_tree,
            files,
            line_counts,
            domains,
            closure_context,
            authority_context,
            impact_graph["files"],
            impact_graph["edges"],
            blob_sizes,
            config,
        )
        findings.extend(packet_findings)
        status = "blocked" if findings or selection == "blocked" else "ready"
        emit_json(
            {
                "schema_version": COMPILE_SCHEMA,
                "status": status,
                "owner": config.get("owner"),
                "scope": {
                    "schema_version": SCOPE_SCHEMA,
                    "path": args.scope,
                    "sha256": hashlib.sha256(scope_path.read_bytes()).hexdigest(),
                    "issue": scope.get("issue"),
                    "candidate_lineage": scope.get("candidate_lineage"),
                },
                "exact_base_sha": base,
                "exact_head_sha": head,
                "exact_head_tree": head_tree,
                "changed_files": files,
                "changed_lines": sum(line_counts.values()),
                "domains": domains,
                "reference_closure": {"files": closure, "edges": edges},
                "authority_inventory": authority,
                "impact_graph": impact_graph,
                "selection": selection,
                "packets": packets,
                "findings": findings,
                "coverage_manifest": {
                    "changed_files": files,
                    "closure_files": closure_context,
                    "authority_files": authority_context,
                    "impact_files": impact_graph["files"],
                    "impact_edge_ids": [edge["id"] for edge in impact_graph["edges"]],
                    "edge_context_files": sorted(
                        {path for packet in packets for path in packet.get("edge_context_files", [])}
                    ),
                    "packet_ids": [packet["packet_id"] for packet in packets],
                },
                "content_policy": "paths and metadata only; no file contents or environment values",
            }
        )
        return 1 if status == "blocked" else 0
    except ReviewSystemError as exc:
        emit_json({"schema_version": COMPILE_SCHEMA, "status": "blocked", "findings": [finding("compile_input", str(exc))]})
        return 2


def response_invariants(
    packet_id: str,
    expected: set[str],
    response: dict[str, Any],
) -> tuple[list[dict[str, Any]], set[str]]:
    blockers: list[dict[str, Any]] = []
    items = response.get("invariants")
    if not isinstance(items, list):
        return [finding("packet_response", f"{packet_id} invariants must be a list")], set()
    seen: set[str] = set()
    failed: set[str] = set()
    for item in items:
        if not isinstance(item, dict) or set(item) != {"id", "status", "evidence_paths"}:
            blockers.append(finding("packet_response", f"{packet_id} invariant entry is malformed"))
            continue
        identifier = item.get("id")
        status = item.get("status")
        evidence_paths = item.get("evidence_paths")
        if not isinstance(identifier, str) or identifier in seen or identifier not in expected:
            blockers.append(finding("packet_coverage", f"{packet_id} invariant id is duplicate or unassigned: {identifier!r}"))
            continue
        if status not in {"pass", "fail", "blocked"}:
            blockers.append(finding("packet_response", f"{packet_id} invariant {identifier} status is invalid"))
        if not isinstance(evidence_paths, list) or not all(isinstance(path, str) for path in evidence_paths):
            blockers.append(finding("packet_response", f"{packet_id} invariant {identifier} evidence_paths is malformed"))
        seen.add(identifier)
        if status == "fail":
            failed.add(identifier)
        elif status == "blocked":
            blockers.append(finding("packet_response", f"{packet_id} invariant {identifier} is blocked"))
    if seen != expected:
        blockers.append(finding("packet_coverage", f"{packet_id} invariant assignment differs: missing={sorted(expected-seen)} extra={sorted(seen-expected)}"))
    return blockers, failed


def validate_review_behaviors(packet_id: str, response: dict[str, Any]) -> list[dict[str, Any]]:
    blockers: list[dict[str, Any]] = []
    behaviors = response.get("review_behaviors")
    if not isinstance(behaviors, dict) or behaviors.get("impact_model_built_first") is not True:
        blockers.append(finding("hypothesis_evidence", f"{packet_id} did not build the impact model before line judgment"))
        return blockers
    directions = behaviors.get("directions_traced")
    if not isinstance(directions, list) or set(directions) != {"upstream", "downstream", "lateral", "temporal"}:
        blockers.append(finding("hypothesis_evidence", f"{packet_id} must trace exactly all four impact directions"))
    for field in ("history_inspected", "sibling_paths_compared"):
        value = behaviors.get(field)
        if not isinstance(value, dict) or set(value) != {"status", "reason"} or value.get("status") not in {"inspected", "not_needed"} or not isinstance(value.get("reason"), str) or not value["reason"].strip():
            blockers.append(finding("hypothesis_evidence", f"{packet_id} lacks reasoned {field}"))
    hypotheses = behaviors.get("hypotheses")
    if not isinstance(hypotheses, list) or not hypotheses or not all(isinstance(value, str) and value.strip() for value in hypotheses):
        blockers.append(finding("hypothesis_evidence", f"{packet_id} lacks explicit falsifiable hypotheses and alternatives"))
    if not isinstance(behaviors.get("disconfirming_evidence"), str) or not behaviors["disconfirming_evidence"].strip():
        blockers.append(finding("hypothesis_evidence", f"{packet_id} lacks disconfirming evidence"))
    return blockers


def validate_experiments(
    root: Path,
    manifest: dict[str, Any],
    packet: dict[str, Any],
    response: dict[str, Any],
) -> list[dict[str, Any]]:
    packet_id = packet["packet_id"]
    blockers: list[dict[str, Any]] = []
    experiments = response.get("experiments")
    if not isinstance(experiments, list):
        return [finding("hypothesis_evidence", f"{packet_id} experiments must be a list")]
    if not experiments:
        reason = response.get("no_experiment_reason")
        if not isinstance(reason, str) or not reason.strip():
            blockers.append(finding("hypothesis_evidence", f"{packet_id} has no experiment and no decisive-static-evidence reason"))
        return blockers
    if response.get("no_experiment_reason") is not None:
        blockers.append(finding("hypothesis_evidence", f"{packet_id} cannot declare both experiments and no-experiment reason"))
    required = {
        "hypothesis_id", "claim", "alternative", "impact_edges_examined", "temporary_change",
        "command", "expected_discriminator", "observed", "supports", "candidate_unchanged",
        "lab_cleanup_verified", "lab_evidence_path", "lab_evidence_sha256",
    }
    assigned_edges = set(packet.get("impact_edge_ids", []))
    for experiment in experiments:
        if not isinstance(experiment, dict) or not required.issubset(experiment):
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment lacks required counterfactual evidence"))
            continue
        for field in ("hypothesis_id", "claim", "alternative", "temporary_change", "lab_evidence_path", "lab_evidence_sha256"):
            if not isinstance(experiment.get(field), str) or not experiment[field].strip():
                blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment {field} is malformed"))
        examined = experiment.get("impact_edges_examined")
        if not isinstance(examined, list) or not all(isinstance(item, str) for item in examined) or not set(examined).issubset(assigned_edges):
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment cites unassigned impact edges"))
        if not isinstance(experiment.get("command"), list) or not experiment["command"] or not all(isinstance(item, str) and item for item in experiment["command"]):
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment command is malformed"))
        if not isinstance(experiment.get("expected_discriminator"), dict) or not isinstance(experiment.get("observed"), dict):
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment discriminator evidence is malformed"))
        if experiment.get("supports") not in {"claim", "alternative"}:
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment is inconclusive and cannot prove clean"))
        if experiment.get("candidate_unchanged") is not True or experiment.get("lab_cleanup_verified") is not True:
            blockers.append(finding("lab_safety", f"{packet_id} experiment lacks candidate/cleanup proof"))
        try:
            evidence_path = resolve_safe(root, experiment.get("lab_evidence_path", ""))
            evidence_bytes = evidence_path.read_bytes()
            if hashlib.sha256(evidence_bytes).hexdigest() != experiment.get("lab_evidence_sha256"):
                raise ReviewSystemError("lab evidence hash differs")
            evidence = json.loads(evidence_bytes)
            if evidence.get("schema_version") != LAB_EVIDENCE_SCHEMA or evidence.get("status") != "evidence":
                raise ReviewSystemError("lab evidence schema/status is not clean evidence")
            for field in ("exact_base_sha", "exact_head_sha", "exact_head_tree"):
                if evidence.get(field) != manifest.get(field):
                    raise ReviewSystemError(f"lab evidence {field} is stale")
            if evidence.get("packet_id") != packet_id or evidence.get("candidate_unchanged") is not True or evidence.get("lab_cleanup_verified") is not True:
                raise ReviewSystemError("lab evidence packet, candidate, or cleanup proof differs")
            observed = evidence.get("experiment", {}).get("observed", {})
            expected = evidence.get("experiment", {}).get("expected_discriminator", {})
            if not isinstance(expected, dict) or expected.get("exit_code") != observed.get("exit_code") or observed.get("limit_hit") is not None or observed.get("processes_remaining") != 0:
                raise ReviewSystemError("lab discriminator, limit, or residue proof is incomplete")
        except (ReviewSystemError, OSError, json.JSONDecodeError) as exc:
            blockers.append(finding("lab_safety", f"{packet_id} lab evidence is invalid: {exc}"))
    return blockers


def validate_response_shape(packet_id: str, response: Any) -> list[dict[str, Any]]:
    if not isinstance(response, dict):
        return [finding("packet_response", f"{packet_id} response must be an object")]
    required = {
        "schema_version", "packet_id", "exact_base_sha", "exact_head_sha", "exact_head_tree",
        "status", "reviewed_files", "closure_files", "authority_files", "impact_files",
        "impact_edge_ids", "edge_context_files", "invariants", "unreviewed_files",
        "review_behaviors", "experiments", "no_experiment_reason", "findings", "residual_risk",
        "context", "wall_clock_ms",
    }
    blockers: list[dict[str, Any]] = []
    missing = required - set(response)
    if missing:
        blockers.append(finding("packet_response", f"{packet_id} lacks required fields: {sorted(missing)}"))
    for field in ("reviewed_files", "closure_files", "authority_files", "impact_files", "impact_edge_ids", "edge_context_files", "unreviewed_files", "residual_risk"):
        value = response.get(field)
        if not isinstance(value, list) or not all(isinstance(item, str) for item in value) or len(value) != len(set(value)):
            blockers.append(finding("packet_response", f"{packet_id} {field} must be a duplicate-free string list"))
    context = response.get("context")
    if not isinstance(context, dict) or set(context) != {"input_tokens", "output_tokens", "cost", "overflow", "truncated"}:
        blockers.append(finding("packet_response", f"{packet_id} context shape is malformed"))
    elif not isinstance(context.get("overflow"), bool) or not isinstance(context.get("truncated"), bool):
        blockers.append(finding("packet_response", f"{packet_id} context flags are malformed"))
    findings_value = response.get("findings")
    if not isinstance(findings_value, list):
        blockers.append(finding("packet_response", f"{packet_id} findings must be a list"))
    else:
        finding_fields = {"severity", "category", "path", "line", "evidence", "impact", "smallest_safe_correction"}
        for item in findings_value:
            if not isinstance(item, dict) or not finding_fields.issubset(item) or item.get("severity") not in {"critical", "high", "medium", "low"}:
                blockers.append(finding("packet_response", f"{packet_id} finding shape is malformed"))
                continue
            if not all(isinstance(item.get(field), str) and item[field].strip() for field in finding_fields - {"severity"}):
                blockers.append(finding("packet_response", f"{packet_id} finding fields must be non-empty strings"))
    wall = response.get("wall_clock_ms")
    if wall is not None and (not isinstance(wall, (int, float)) or wall < 0):
        blockers.append(finding("packet_response", f"{packet_id} wall_clock_ms is malformed"))
    return blockers


def verify_manifest_candidate(root: Path, manifest: dict[str, Any]) -> list[dict[str, Any]]:
    blockers: list[dict[str, Any]] = []
    if manifest.get("status") != "ready" or manifest.get("selection") == "blocked" or manifest.get("findings"):
        blockers.append(finding("compile_manifest", "compile manifest is not ready and finding-free"))
    impact = manifest.get("impact_graph")
    if not isinstance(impact, dict) or impact.get("status") != "complete" or impact.get("bounds", {}).get("hit") is not False:
        blockers.append(finding("compile_manifest", "compile manifest impact graph is incomplete or bounded"))
    packets = manifest.get("packets")
    if not isinstance(packets, list) or not packets:
        blockers.append(finding("compile_manifest", "compile manifest has no review packets"))
    for field in ("exact_base_sha", "exact_head_sha", "exact_head_tree"):
        try:
            validate_sha(manifest.get(field), f"manifest {field}")
        except ReviewSystemError as exc:
            blockers.append(finding("stale_evidence", str(exc)))
    if blockers:
        return blockers
    current = run_git(root, ["rev-parse", "HEAD"]).strip()
    tree = run_git(root, ["rev-parse", "HEAD^{tree}"]).strip()
    status = run_git(root, ["--no-optional-locks", "status", "--porcelain", "--untracked-files=all"]).strip()
    if current != manifest["exact_head_sha"] or tree != manifest["exact_head_tree"] or status:
        blockers.append(finding("stale_evidence", "current clean HEAD/tree does not match the compile manifest"))
    merge_base = run_git(root, ["merge-base", manifest["exact_base_sha"], manifest["exact_head_sha"]]).strip()
    if merge_base != manifest["exact_base_sha"]:
        blockers.append(finding("stale_evidence", "manifest exact base is not the candidate merge base"))
    scope = manifest.get("scope")
    if not isinstance(scope, dict) or scope.get("schema_version") != SCOPE_SCHEMA:
        blockers.append(finding("compile_manifest", "manifest scope binding is absent or incompatible"))
    else:
        try:
            scope_path = resolve_safe(root, scope.get("path", ""))
            if hashlib.sha256(scope_path.read_bytes()).hexdigest() != scope.get("sha256"):
                blockers.append(finding("stale_evidence", "manifest scope hash differs from current exact scope"))
        except ReviewSystemError as exc:
            blockers.append(finding("compile_manifest", str(exc)))
    return blockers


def command_synthesize(args: argparse.Namespace) -> int:
    try:
        root = Path(args.repo_root).resolve(strict=True)
        manifest = load_json(resolve_safe(root, args.manifest))
        if manifest.get("schema_version") != COMPILE_SCHEMA:
            raise ReviewSystemError(
                f"compile manifest migration required: {manifest.get('schema_version')!r} != {COMPILE_SCHEMA!r}"
            )
        responses_dir = resolve_safe(root, args.responses_dir, must_exist=False)
        root_resolved = root.resolve(strict=True)
        if Path(os.path.commonpath((str(root_resolved), str(responses_dir)))) != root_resolved:
            raise ReviewSystemError("responses directory escapes repository")
        findings: list[dict[str, Any]] = []
        blockers = verify_manifest_candidate(root, manifest)
        response_records: list[dict[str, Any]] = []
        for packet in manifest.get("packets", []):
            response_path = responses_dir / f"{packet['packet_id']}.json"
            if not response_path.is_file():
                blockers.append(finding("packet_response", f"missing packet response {packet['packet_id']}"))
                continue
            response = load_json(response_path)
            response_records.append(response)
            blockers.extend(validate_response_shape(packet["packet_id"], response))
            if response.get("schema_version") != PACKET_RESPONSE_SCHEMA:
                blockers.append(finding("packet_response", f"response migration required for {packet['packet_id']}: {response.get('schema_version')!r} != {PACKET_RESPONSE_SCHEMA!r}"))
            if response.get("packet_id") != packet["packet_id"]:
                blockers.append(finding("packet_response", f"packet id mismatch for {packet['packet_id']}"))
            for field in ("exact_base_sha", "exact_head_sha", "exact_head_tree"):
                if response.get(field) != manifest.get(field):
                    blockers.append(finding("stale_evidence", f"{packet['packet_id']} {field} is stale"))
            expected_sets = {
                "reviewed_files": packet.get("changed_files", []),
                "closure_files": packet.get("closure_files", []),
                "authority_files": packet.get("authority_files", []),
                "impact_files": packet.get("impact_files", []),
                "impact_edge_ids": packet.get("impact_edge_ids", []),
                "edge_context_files": packet.get("edge_context_files", []),
            }
            for field, expected_values in expected_sets.items():
                actual = response.get(field)
                if not isinstance(actual, list) or actual != expected_values:
                    blockers.append(finding("packet_coverage", f"{packet['packet_id']} {field} differs from exact assignment"))
            invariant_blockers, failed_invariants = response_invariants(
                packet["packet_id"], set(packet.get("invariants", [])), response
            )
            blockers.extend(invariant_blockers)
            if response.get("unreviewed_files"):
                blockers.append(finding("packet_coverage", f"{packet['packet_id']} declares unreviewed files"))
            blockers.extend(validate_review_behaviors(packet["packet_id"], response))
            blockers.extend(validate_experiments(root, manifest, packet, response))
            context = response.get("context") if isinstance(response.get("context"), dict) else {}
            if context.get("overflow") or context.get("truncated"):
                blockers.append(finding("packet_overflow", f"{packet['packet_id']} overflowed or truncated"))
            packet_findings = response.get("findings") if isinstance(response.get("findings"), list) else []
            status = response.get("status")
            if status == "clean" and packet_findings:
                blockers.append(finding("packet_response", f"{packet['packet_id']} clean status contains findings"))
            elif status == "findings" and not packet_findings:
                blockers.append(finding("packet_response", f"{packet['packet_id']} findings status has no findings"))
            elif status == "blocked":
                blockers.append(finding("packet_response", f"{packet['packet_id']} reviewer is blocked"))
            elif status not in {"clean", "findings", "blocked"}:
                blockers.append(finding("packet_response", f"{packet['packet_id']} has invalid status"))
            if failed_invariants and not packet_findings:
                blockers.append(finding("packet_response", f"{packet['packet_id']} failed invariants lack paired findings: {sorted(failed_invariants)}"))
            for item in packet_findings:
                if isinstance(item, dict):
                    findings.append({"packet_id": packet["packet_id"], **item})
        blockers.extend(verify_manifest_candidate(root, manifest))
        if blockers:
            status = "blocked"
        elif findings:
            status = "findings_correction_required"
        else:
            status = "clean"
        emit_json(
            {
                "schema_version": SYNTHESIS_SCHEMA,
                "owner": "parent_orchestrator",
                "exact_base_sha": manifest.get("exact_base_sha"),
                "exact_head_sha": manifest.get("exact_head_sha"),
                "exact_head_tree": manifest.get("exact_head_tree"),
                "status": status,
                "packet_count": len(manifest.get("packets", [])),
                "response_count": len(response_records),
                "findings": findings,
                "blockers": blockers,
                "shepherd": {"status": "pending", "rule": "run independently only after clean synthesis"},
                "human_merge_authority": True,
            }
        )
        return 0 if status == "clean" else 1
    except (ReviewSystemError, TypeError, KeyError, json.JSONDecodeError) as exc:
        emit_json({"schema_version": SYNTHESIS_SCHEMA, "status": "blocked", "blockers": [finding("synthesis_input", str(exc))]})
        return 2

def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description="Compile and measure canonical PM review inputs")
    subparsers = result.add_subparsers(dest="command", required=True)

    detect_parser = subparsers.add_parser("detect", help="run one fixture without oracle")
    detect_parser.add_argument("--mode", choices=("baseline", "treatment"), required=True)
    detect_parser.add_argument("--input", required=True)
    detect_parser.add_argument("--case-id", required=True)
    detect_parser.set_defaults(func=command_detect)

    observe_parser = subparsers.add_parser("observe", help="run all detector-visible fixtures without oracle")
    observe_parser.add_argument("--mode", choices=("baseline", "treatment"), required=True)
    observe_parser.add_argument("--input", required=True)
    observe_parser.set_defaults(func=command_observe)

    score_parser = subparsers.add_parser("score", help="score completed observations against separate oracle")
    score_parser.add_argument("--observations", required=True)
    score_parser.add_argument("--oracle", required=True)
    score_parser.set_defaults(func=command_score)

    compile_parser = subparsers.add_parser("compile", help="compile exact-head bounded review packets")
    compile_parser.add_argument("--repo-root", default=".")
    compile_parser.add_argument("--config", default=".agents/agentic-delivery/contracts/pm-review-system.json")
    compile_parser.add_argument("--scope", required=True)
    compile_parser.add_argument("--base", required=True)
    compile_parser.add_argument("--head", required=True)
    compile_parser.set_defaults(func=command_compile)

    synthesize_parser = subparsers.add_parser("synthesize", help="synthesize raw packet responses into one PM verdict")
    synthesize_parser.add_argument("--repo-root", default=".")
    synthesize_parser.add_argument("--manifest", required=True)
    synthesize_parser.add_argument("--responses-dir", required=True)
    synthesize_parser.set_defaults(func=command_synthesize)
    return result


def main() -> int:
    try:
        args = parser().parse_args()
        return args.func(args)
    except ReviewSystemError as exc:
        print(f"pm review system error: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
