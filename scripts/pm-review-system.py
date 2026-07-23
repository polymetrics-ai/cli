#!/usr/bin/env python3
"""Compile, validate, synthesize, and measure canonical PM review packets.

The tool is intentionally dependency-free. It emits JSON to stdout and never
writes project files. The parent orchestrator remains the sole lifecycle owner.
"""

from __future__ import annotations

import argparse
import fnmatch
import json
import os
import re
import subprocess
import sys
import time
from collections import deque
from pathlib import Path, PurePosixPath
from typing import Any, Iterable

OBSERVATION_SCHEMA = "polymetrics.ai/pm-review-observation/v1"
COMPILE_SCHEMA = "polymetrics.ai/pm-review-compile/v1"
PACKET_SCHEMA = "polymetrics.ai/pm-review-packet/v1"
SYNTHESIS_SCHEMA = "polymetrics.ai/pm-review-synthesis/v1"
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
REFERENCE_SUFFIXES = (".md", ".json", ".yaml", ".yml", ".sh", ".py")
REPO_REFERENCE = re.compile(
    r"(?<![A-Za-z0-9])((?:\.agents|\.pi|scripts|\.planning)/"
    r"[A-Za-z0-9_.\-/]+\.(?:md|json|ya?ml|sh|py))"
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


def treatment_findings(case: dict[str, Any]) -> list[dict[str, Any]]:
    family = case.get("family")
    data = case.get("input", {})
    result: list[dict[str, Any]] = []

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
    if value.startswith((".agents/", ".pi/", "scripts/", ".planning/")):
        return validate_relative_path(value, "active reference")
    source_parent = PurePosixPath(source).parent
    combined = (source_parent / value).as_posix()
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
            if not target.startswith(allowed_prefixes):
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
    return inventory, sorted(files), findings


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


def build_packets(
    base: str,
    head: str,
    files: list[str],
    line_counts: dict[str, int],
    domains: dict[str, str],
    closure_files: list[str],
    authority_files: list[str],
    config: dict[str, Any],
) -> tuple[str, list[dict[str, Any]], list[dict[str, Any]]]:
    thresholds = config["thresholds"]
    domain_values = sorted(set(domains.values()))
    changed_lines = sum(line_counts.values())
    combined = (
        len(files) <= thresholds["combined_max_files"]
        and changed_lines <= thresholds["combined_max_non_generated_lines"]
        and len(domain_values) <= thresholds["combined_max_domains"]
        and len(closure_files) <= thresholds["packet_max_context_files"]
        and len(authority_files) <= thresholds["packet_max_context_files"]
    )
    selection = "combined" if combined else "split"
    packets: list[dict[str, Any]] = []
    problems: list[dict[str, Any]] = []

    def append_packet(role: str, changed: list[str], closure: list[str], authority: list[str]) -> None:
        packet_number = 1 + sum(1 for packet in packets if packet["role"] == role)
        estimated = (
            sum(line_counts.get(path, 0) for path in changed) * thresholds["estimated_tokens_per_changed_line"]
            + (len(closure) + len(authority)) * thresholds["estimated_tokens_per_context_file"]
        )
        overflow = estimated > thresholds["packet_target_tokens"]
        if overflow:
            problems.append(finding("packet_overflow", f"{role} packet estimate {estimated} exceeds target {thresholds['packet_target_tokens']}"))
        packets.append(
            {
                "schema_version": PACKET_SCHEMA,
                "packet_id": f"{role}-{packet_number:02d}",
                "role": role,
                "exact_base_sha": base,
                "exact_head_sha": head,
                "changed_files": changed,
                "closure_files": closure,
                "authority_files": authority,
                "invariants": config["packet_invariants"].get(role, config["packet_invariants"]["combined"]),
                "context": {
                    "target_tokens": thresholds["packet_target_tokens"],
                    "estimated_tokens": estimated,
                    "overflow": overflow,
                    "truncated": False,
                },
                "required_response_fields": [
                    "reviewed_files",
                    "closure_files",
                    "authority_files",
                    "invariants",
                    "unreviewed_files",
                    "findings",
                    "context",
                ],
            }
        )

    if combined:
        append_packet("combined", files, closure_files, authority_files)
        return selection, packets, problems

    by_role: dict[str, list[str]] = {}
    for path in files:
        by_role.setdefault(domains[path], []).append(path)
    for role in ("architecture_reference", "authority_workflow_state", "implementation_test"):
        for part in chunks(sorted(by_role.get(role, [])), thresholds["packet_max_changed_files"]):
            append_packet(role, part, [], [])

    def attach_context(role: str, key: str, values: list[str]) -> None:
        parts = chunks(values, thresholds["packet_max_context_files"])
        role_packets = [packet for packet in packets if packet["role"] == role]
        for index, part in enumerate(parts):
            if index < len(role_packets):
                role_packets[index][key] = part
                role_packets[index]["context"]["estimated_tokens"] += len(part) * thresholds["estimated_tokens_per_context_file"]
                if role_packets[index]["context"]["estimated_tokens"] > thresholds["packet_target_tokens"]:
                    role_packets[index]["context"]["overflow"] = True
                    problems.append(finding("packet_overflow", f"{role_packets[index]['packet_id']} exceeds target after context assignment"))
            else:
                append_packet(role, [], part if key == "closure_files" else [], part if key == "authority_files" else [])

    attach_context("architecture_reference", "closure_files", closure_files)
    attach_context("authority_workflow_state", "authority_files", authority_files)
    if any(packet["context"]["overflow"] for packet in packets):
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
        base = validate_sha(args.base, "exact base")
        head = validate_sha(args.head, "exact head")
        run_git(root, ["cat-file", "-e", f"{base}^{{commit}}"])
        run_git(root, ["cat-file", "-e", f"{head}^{{commit}}"])
        if not args.allow_non_head:
            current = run_git(root, ["rev-parse", "HEAD"]).strip()
            if current != head:
                raise ReviewSystemError(f"worktree HEAD drift: {current} != {head}")
        files, line_counts = changed_files(root, base, head)
        findings: list[dict[str, Any]] = []
        for path in files:
            if path_matches(path, config.get("forbidden_changed_paths", [])):
                findings.append(finding("changed_path_scope", "changed path is forbidden", path))
            elif not path_matches(path, config.get("allowed_changed_paths", [])):
                findings.append(finding("changed_path_scope", "changed path is outside the positive allowlist", path))
        closure, edges, closure_findings = compile_closure(root, config)
        findings.extend(closure_findings)
        authority, authority_files, authority_findings = authority_inventory(root, config)
        findings.extend(authority_findings)
        domains = {path: classify_domain(path, config) for path in files}
        closure_context = sorted(set(closure) - set(files))
        authority_context = sorted(set(authority_files) - set(files) - set(closure_context))
        selection, packets, packet_findings = build_packets(
            base,
            head,
            files,
            line_counts,
            domains,
            closure_context,
            authority_context,
            config,
        )
        findings.extend(packet_findings)
        status = "blocked" if findings or selection == "blocked" else "ready"
        emit_json(
            {
                "schema_version": COMPILE_SCHEMA,
                "status": status,
                "owner": config.get("owner"),
                "exact_base_sha": base,
                "exact_head_sha": head,
                "changed_files": files,
                "changed_lines": sum(line_counts.values()),
                "domains": domains,
                "reference_closure": {"files": closure, "edges": edges},
                "authority_inventory": authority,
                "selection": selection,
                "packets": packets,
                "findings": findings,
                "coverage_manifest": {
                    "changed_files": files,
                    "closure_files": closure_context,
                    "authority_files": authority_context,
                    "packet_ids": [packet["packet_id"] for packet in packets],
                },
                "content_policy": "paths and metadata only; no file contents or environment values",
            }
        )
        return 1 if status == "blocked" else 0
    except ReviewSystemError as exc:
        emit_json({"schema_version": COMPILE_SCHEMA, "status": "blocked", "findings": [finding("compile_input", str(exc))]})
        return 2


def response_invariant_ids(response: dict[str, Any]) -> set[str]:
    result: set[str] = set()
    for item in response.get("invariants", []):
        if isinstance(item, dict) and item.get("status") == "pass" and isinstance(item.get("id"), str):
            result.add(item["id"])
    return result


def command_synthesize(args: argparse.Namespace) -> int:
    try:
        root = Path(args.repo_root).resolve(strict=True)
        manifest = load_json(resolve_safe(root, args.manifest))
        responses_dir = resolve_safe(root, args.responses_dir, must_exist=False)
        root_resolved = root.resolve(strict=True)
        if Path(os.path.commonpath((str(root_resolved), str(responses_dir)))) != root_resolved:
            raise ReviewSystemError("responses directory escapes repository")
        findings: list[dict[str, Any]] = []
        blockers: list[dict[str, Any]] = []
        response_records: list[dict[str, Any]] = []
        for packet in manifest.get("packets", []):
            response_path = responses_dir / f"{packet['packet_id']}.json"
            if not response_path.is_file():
                blockers.append(finding("packet_response", f"missing packet response {packet['packet_id']}"))
                continue
            response = load_json(response_path)
            response_records.append(response)
            if response.get("schema_version") != "polymetrics.ai/pm-review-packet-response/v1":
                blockers.append(finding("packet_response", f"invalid response schema for {packet['packet_id']}"))
            if response.get("packet_id") != packet["packet_id"]:
                blockers.append(finding("packet_response", f"packet id mismatch for {packet['packet_id']}"))
            for field in ("exact_base_sha", "exact_head_sha"):
                if response.get(field) != manifest.get(field):
                    blockers.append(finding("stale_evidence", f"{packet['packet_id']} {field} is stale"))
            expected_sets = {
                "reviewed_files": set(packet.get("changed_files", [])),
                "closure_files": set(packet.get("closure_files", [])),
                "authority_files": set(packet.get("authority_files", [])),
            }
            for field, expected in expected_sets.items():
                actual = set(response.get(field, []))
                if not expected.issubset(actual):
                    blockers.append(finding("packet_coverage", f"{packet['packet_id']} omits {field}: {sorted(expected - actual)}"))
            missing_invariants = set(packet.get("invariants", [])) - response_invariant_ids(response)
            if missing_invariants:
                blockers.append(finding("packet_coverage", f"{packet['packet_id']} omits passing invariants: {sorted(missing_invariants)}"))
            if response.get("unreviewed_files"):
                blockers.append(finding("packet_coverage", f"{packet['packet_id']} declares unreviewed files"))
            context = response.get("context", {})
            if context.get("overflow") or context.get("truncated"):
                blockers.append(finding("packet_overflow", f"{packet['packet_id']} overflowed or truncated"))
            if response.get("status") not in {"clean", "findings", "blocked"}:
                blockers.append(finding("packet_response", f"{packet['packet_id']} has invalid status"))
            if response.get("status") == "blocked":
                blockers.append(finding("packet_response", f"{packet['packet_id']} reviewer is blocked"))
            for item in response.get("findings", []):
                findings.append({"packet_id": packet["packet_id"], **item})
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
    except ReviewSystemError as exc:
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
    compile_parser.add_argument("--base", required=True)
    compile_parser.add_argument("--head", required=True)
    compile_parser.add_argument("--allow-non-head", action="store_true")
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
