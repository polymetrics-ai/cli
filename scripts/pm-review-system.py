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
CONFIG_SCHEMA = "polymetrics.ai/pm-review-system/v4"
SCOPE_SCHEMA = "polymetrics.ai/pm-review-scope/v1"
COMPILE_SCHEMA = "polymetrics.ai/pm-review-compile/v4"
IMPACT_SCHEMA = "polymetrics.ai/pm-review-impact-graph/v3"
PACKET_SCHEMA = "polymetrics.ai/pm-review-packet/v4"
PACKET_RESPONSE_SCHEMA = "polymetrics.ai/pm-review-packet-response/v4"
SYNTHESIS_SCHEMA = "polymetrics.ai/pm-review-synthesis/v4"
LAB_EVIDENCE_SCHEMA = "polymetrics.ai/pm-review-lab-evidence/v3"
MEASUREMENT_SCHEMA = "polymetrics.ai/pm-review-measurement/v2"
FINDING_OBSERVATION_SCHEMA = "polymetrics.ai/pm-review-finding-observation/v1"
OCCURRENCE_SCHEMA = "polymetrics.ai/pm-review-occurrence/v1"
ROOT_CAUSE_SCHEMA = "polymetrics.ai/pm-review-root-cause/v1"
DEDUP_SCHEMA = "polymetrics.ai/pm-review-dedup/v1"
DEDUP_DECISIONS_SCHEMA = "polymetrics.ai/pm-review-dedup-decisions/v1"
DEDUP_HISTORY_SCHEMA = "polymetrics.ai/pm-review-dedup-history/v1"
TRUST_BUNDLE_SCHEMA = "polymetrics.ai/pm-review-trust-bundle/v1"
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
HEX_DIGEST = re.compile(r"^[0-9a-f]{64}$")
SAFE_ID = re.compile(r"^[A-Za-z0-9][A-Za-z0-9_.-]{0,79}$")
CONTROL = re.compile(r"[\x00-\x1f\x7f]")
REFERENCE_SUFFIXES = (".md", ".json", ".yaml", ".yml", ".sh", ".py", ".go")
INDEX_SUFFIXES = set(REFERENCE_SUFFIXES)
PATH_REFERENCE = re.compile(
    r"(?<![A-Za-z0-9])((?:(?:\.{0,2}/)?(?:[A-Za-z0-9_.-]+/)+"
    r"[A-Za-z0-9_.-]+\.(?:md|json|ya?ml|sh|py|go)|(?:AGENTS|CLAUDE)\.md))"
)
REPO_REFERENCE = PATH_REFERENCE
MARKDOWN_LINK = re.compile(r"\[[^\]]*\]\(([^)]+)\)")
CERTAINTIES = {"active", "inactive", "unknown"}
DIRECTIONS = {"upstream", "downstream", "lateral", "temporal"}
RELATIONS = {
    "required_reference", "descriptive_reference", "script_invokes", "python_import",
    "authority_writes", "authority_reads", "authority_mirror", "generates",
    "generated_consumer", "go_member_of", "go_contains", "go_imports", "go_test",
    "platform_variant", "temporal_phase_state_instance", "temporal_state_mirror",
    "migration", "restart_resume", "version_invalidation", "fixture", "sibling_variant",
}
PHASE_STATUSES = {
    "planned", "red", "green", "verifying", "review_pending",
    "findings_correction_required", "blocked", "human_ready", "complete",
    "local_codex_round_2_findings_systemic_correction_planned",
}
LOCAL_CODEX_STATUSES = {"pending", "findings_correction_required", "clean", "comments_addressed", "blocked"}
SHEPHERD_STATUSES = {"pending", "proceed", "retry", "revert", "halt", "blocked"}
SLICE_FIELDS = {
    "path", "revision", "revision_kind", "blob_sha256", "start_line", "end_line",
    "start_byte", "end_byte", "bytes", "sha256",
}
HYPOTHESIS_FIELDS = {"id", "claim", "strongest_alternative", "falsifier", "evidence_paths"}
EXPERIMENT_BINDING_FIELDS = {
    "hypothesis_id", "claim", "alternative", "impact_edges_examined", "temporary_change",
    "command", "expected_discriminator", "observed",
}
PROMPT_HEADER = b"PM exact-head packet review v4\n"
PACKET_MARKER = b"\nCOMPILED PACKET:\n"
PAYLOAD_MARKER = b"\nEXACT SLICE PAYLOADS (canonical descriptor order):\n"
SLICE_SEPARATOR = b"\n<<PM_SLICE>>\n"
MAX_EXACT_DIFF_BYTES = 64 * 1024 * 1024
MAX_RESPONSE_BYTES = 16 * 1024 * 1024
DEDUP_KEY_VERSIONS = {
    "payload_hash": "payloadHash/v1",
    "invariant": "invariant/v1",
    "category_family": "categoryFamily/v1",
    "source_anchor": "sourceAnchor/v1",
    "mechanism": "mechanism/v1",
    "correction_surface": "correctionSurface/v1",
    "claim_shingles": "claimShingles/v1",
}
DEDUP_STOP_WORDS = {
    "a", "an", "and", "are", "as", "at", "be", "by", "can", "for", "from", "has",
    "in", "into", "is", "it", "of", "on", "or", "that", "the", "this", "to", "was",
    "when", "with", "without", "must", "should", "may", "not",
}
CATEGORY_FAMILY_RULES = (
    (re.compile(r"(?:packet|context).*(?:bound|budget|overflow|slice)"), "packet_bounds"),
    (re.compile(r"(?:impact|graph).*(?:parser|relation|edge|certainty|coverage|provenance|recall|completeness)"), "impact_graph"),
    (re.compile(r"(?:lab|experiment).*(?:evidence|binding|safety|resource|denial|discriminator)"), "lab_evidence"),
    (re.compile(r"(?:manifest|exact|version|blob|stale).*(?:binding|integrity|authentication|evidence)?"), "exact_identity"),
    (re.compile(r"(?:phase|authority|state|schema|mirror)"), "authority_state"),
    (re.compile(r"(?:workflow|route|gate|shepherd|orchestration)"), "workflow_gate"),
    (re.compile(r"(?:secret|credential|trace|path_safety)"), "security_containment"),
)


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


def canonical_json_bytes(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=True).encode()


def sha256_json(value: Any) -> str:
    return hashlib.sha256(canonical_json_bytes(value)).hexdigest()


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise ReviewSystemError(f"cannot read JSON {path}: {exc}") from exc


def require_object(value: Any, label: str) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise ReviewSystemError(f"{label} must be a JSON object")
    return value


def git_blob(root: Path, commit: str, relative: str, *, maximum: int | None = None) -> bytes:
    validate_sha(commit, "blob revision")
    relative = validate_relative_path(relative, "blob path")
    proc = subprocess.Popen(
        ["git", "-C", str(root), "show", f"{commit}:{relative}"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    assert proc.stdout is not None and proc.stderr is not None
    limit = maximum if maximum is not None else 64 * 1024 * 1024
    data = proc.stdout.read(limit + 1)
    if len(data) > limit:
        proc.kill()
        proc.wait()
        raise ReviewSystemError(f"exact blob exceeds {limit} bytes: {relative}")
    stderr = proc.stderr.read(4096)
    return_code = proc.wait()
    if return_code != 0:
        detail = stderr.decode(errors="replace").strip()
        raise ReviewSystemError(f"exact blob is absent at {commit[:12]}: {relative}: {detail}")
    return data


def git_blob_optional(root: Path, commit: str, relative: str, *, maximum: int | None = None) -> bytes | None:
    try:
        return git_blob(root, commit, relative, maximum=maximum)
    except ReviewSystemError:
        return None


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
        if certainty != "inactive" and source not in nodes:
            result.append(finding("impact_graph_contract", f"{certainty} impact source is absent: {source}"))
        if certainty != "inactive" and target not in nodes:
            result.append(finding("impact_graph_contract", f"{certainty} impact target is absent: {target}"))
        if certainty != "inactive" and source in nodes and target in nodes:
            forward[source].append(edge)
            reverse[target].append(edge)
    seeds = sorted(set(data.get("seeds", [])))
    for seed in seeds:
        if seed not in nodes:
            result.append(finding("impact_graph_contract", f"impact seed is absent: {seed}"))
    pending = deque((seed, 0) for seed in seeds if seed in nodes)
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


def run_git_bytes_bounded(root: Path, args: list[str], maximum: int) -> bytes:
    if not isinstance(maximum, int) or maximum <= 0:
        raise ReviewSystemError("bounded Git output maximum must be a positive integer")
    with tempfile.TemporaryFile() as stderr_file:
        proc = subprocess.Popen(
            ["git", "-C", str(root), *args],
            stdout=subprocess.PIPE,
            stderr=stderr_file,
        )
        assert proc.stdout is not None
        payload = proc.stdout.read(maximum + 1)
        if len(payload) > maximum:
            proc.kill()
            proc.wait()
            raise ReviewSystemError(f"bounded Git output exceeds {maximum} bytes: {' '.join(args[:2])}")
        return_code = proc.wait()
        if return_code != 0:
            stderr_file.seek(0)
            detail = stderr_file.read(4096).decode(errors="replace").strip()
            raise ReviewSystemError(f"git {' '.join(args[:2])} failed: {detail}")
        return payload


def read_file_bounded(path: Path, maximum: int) -> bytes:
    with path.open("rb") as handle:
        payload = handle.read(maximum + 1)
    if len(payload) > maximum:
        raise ReviewSystemError(f"file exceeds {maximum} bytes: {path.name}")
    return payload


def normalize_reference(source: str, raw: str) -> str | None:
    value = raw.strip().strip("`'\"(),:;")
    value = value.split("#", 1)[0].split("?", 1)[0]
    if not value or value.startswith(("http://", "https://", "mailto:", "#")):
        return None
    if any(marker in value for marker in ("<", ">", "{", "}", "$")):
        return None
    if not value.endswith(REFERENCE_SUFFIXES):
        return None
    repository_roots = (".agents/", ".pi/", ".gsd/", "scripts/", ".planning/", "cmd/", "internal/", "docs/", "website/")
    if value in {"AGENTS.md", "CLAUDE.md"} or value.startswith(repository_roots):
        return validate_relative_path(value, "active reference")
    source_parent = PurePosixPath(source).parent
    combined = posixpath.normpath((source_parent / value).as_posix())
    return validate_relative_path(combined, "active reference")


def reference_candidates(text: str) -> list[str]:
    values = [match.group(1) for match in MARKDOWN_LINK.finditer(text)]
    values.extend(match.group(1) for match in PATH_REFERENCE.finditer(text))
    return list(dict.fromkeys(values))


def extract_references(relative: str, text: str) -> list[dict[str, Any]]:
    """Extract structurally active closure references with safe relative resolution."""
    candidates: list[tuple[str, str, int]] = []
    suffix = PurePosixPath(relative).suffix
    lines = text.splitlines()
    inactive_section = False
    markdown_heading = ""
    yaml_stack: list[tuple[int, str]] = []

    if relative.endswith("pm-review-system.json"):
        document = require_object(json.loads(text), "review policy")
        for raw in document.get("canonical_roots", []):
            candidates.append((raw, "config_canonical_root", 0))
        for authority in document.get("authorities", []):
            candidates.append((authority.get("authoritative_path", ""), "config_authority", 0))
            for field in ("writers", "readers", "mirrors"):
                for raw in authority.get(field, []):
                    candidates.append((raw, f"config_authority_{field}", 0))
    else:
        for number, line_text in enumerate(lines, 1):
            stripped = line_text.strip()
            if suffix == ".md" and stripped.startswith("#"):
                markdown_heading = stripped.lstrip("#").strip().lower()
                inactive_section = bool(re.search(r"\b(historical|deprecated|retired|forbidden examples?)\b", markdown_heading))
            if inactive_section or line_certainty(line_text) != "active":
                continue
            reason = "inline_required_path"
            relevant = True
            if suffix in {".sh", ".py"}:
                relevant = script_invocation_line(line_text) or bool(re.search(r"\b(?:Path|open|run_path)\s*\(", line_text))
                reason = "script_execution_path"
            elif suffix in {".yaml", ".yml"}:
                indent = len(line_text) - len(line_text.lstrip(" "))
                while yaml_stack and yaml_stack[-1][0] >= indent:
                    yaml_stack.pop()
                key_match = re.match(r"\s*(?:-\s*)?([A-Za-z_][A-Za-z0-9_-]*)\s*:", line_text)
                if key_match:
                    yaml_stack.append((indent, key_match.group(1).lower()))
                context = {key for _, key in yaml_stack}
                relevant = bool(context & {"path", "paths", "prompt", "contract", "workflow", "schema", "template", "required", "requires", "source", "inputs", "dependencies"})
                reason = "yaml_structural_path"
            if not relevant:
                continue
            for raw in reference_candidates(line_text):
                if suffix == ".sh":
                    raw = re.sub(r"^(?:repo_root|REPO_ROOT|ROOT_DIR)/", "", raw)
                    if re.match(r"^[A-Z][A-Z0-9_]*/", raw):
                        continue
                structural_keys = {key for _, key in yaml_stack} if suffix in {".yaml", ".yml"} else set()
                relation = "script_invokes" if suffix == ".sh" and script_invocation_line(line_text) else reference_relation(line_text, raw, structural_keys)
                heading_requires = suffix == ".md" and bool(re.search(r"\b(required|must read|inputs?|dependencies)\b", markdown_heading))
                if relation not in {"required_reference", "script_invokes"} and not heading_requires:
                    continue
                candidates.append((raw, reason, number))

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


def json_schema_findings(value: Any, schema: Any, location: str = "$") -> list[str]:
    """Validate the dependency-free JSON-Schema subset used by PM phase state."""
    if not isinstance(schema, dict):
        return [f"{location}: schema node is not an object"]
    errors: list[str] = []
    expected_type = schema.get("type")
    type_checks = {
        "object": lambda item: isinstance(item, dict),
        "array": lambda item: isinstance(item, list),
        "string": lambda item: isinstance(item, str),
        "boolean": lambda item: isinstance(item, bool),
        "integer": lambda item: isinstance(item, int) and not isinstance(item, bool),
        "number": lambda item: isinstance(item, (int, float)) and not isinstance(item, bool),
        "null": lambda item: item is None,
    }
    if isinstance(expected_type, list):
        type_ok = any(type_checks.get(item, lambda _: False)(value) for item in expected_type)
    else:
        type_ok = type_checks.get(expected_type, lambda _: True)(value)
    if not type_ok:
        return [f"{location}: expected {expected_type}"]
    if "const" in schema and value != schema["const"]:
        errors.append(f"{location}: value differs from const")
    if "enum" in schema and value not in schema["enum"]:
        errors.append(f"{location}: value is outside enum")
    if isinstance(value, str):
        if len(value) < int(schema.get("minLength", 0)):
            errors.append(f"{location}: string is too short")
        if schema.get("pattern") and re.fullmatch(schema["pattern"], value) is None:
            errors.append(f"{location}: string does not match pattern")
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        if "minimum" in schema and value < schema["minimum"]:
            errors.append(f"{location}: number is below minimum")
        if "maximum" in schema and value > schema["maximum"]:
            errors.append(f"{location}: number is above maximum")
    if isinstance(value, dict):
        required = schema.get("required", [])
        if not isinstance(required, list):
            errors.append(f"{location}: schema required is malformed")
            required = []
        for key in required:
            if key not in value:
                errors.append(f"{location}: missing required property {key}")
        if len(value) < int(schema.get("minProperties", 0)):
            errors.append(f"{location}: too few properties")
        properties = schema.get("properties", {})
        if not isinstance(properties, dict):
            errors.append(f"{location}: schema properties is malformed")
            properties = {}
        for key, item in value.items():
            if key in properties:
                errors.extend(json_schema_findings(item, properties[key], f"{location}.{key}"))
            elif schema.get("additionalProperties") is False:
                errors.append(f"{location}: unexpected property {key}")
            elif isinstance(schema.get("additionalProperties"), dict):
                errors.extend(json_schema_findings(item, schema["additionalProperties"], f"{location}.{key}"))
    if isinstance(value, list):
        if schema.get("uniqueItems") and len({canonical_json_bytes(item) for item in value}) != len(value):
            errors.append(f"{location}: array items are not unique")
        item_schema = schema.get("items")
        if isinstance(item_schema, dict):
            for index, item in enumerate(value):
                errors.extend(json_schema_findings(item, item_schema, f"{location}[{index}]"))
    return errors


def phase_state_semantic_findings(state: Any) -> list[str]:
    if not isinstance(state, dict):
        return ["phase state must be an object"]
    errors: list[str] = []
    budget = state.get("correctionBudget")
    if not isinstance(budget, dict) or set(budget) != {"maxCorrectionRounds", "roundsByRange", "headHistory"}:
        errors.append("correctionBudget must contain exactly maxCorrectionRounds, roundsByRange, and headHistory")
        return errors
    maximum = budget.get("maxCorrectionRounds")
    rounds = budget.get("roundsByRange")
    history = budget.get("headHistory")
    if not isinstance(maximum, int) or isinstance(maximum, bool) or maximum < 1:
        errors.append("maxCorrectionRounds must be a positive integer")
    if not isinstance(rounds, dict) or not rounds:
        errors.append("roundsByRange must be a non-empty object")
    else:
        for key, value in rounds.items():
            if not isinstance(key, str) or "..." not in key or not isinstance(value, int) or isinstance(value, bool) or value < 0:
                errors.append("roundsByRange contains a malformed lineage/count")
            elif isinstance(maximum, int) and value > maximum:
                errors.append("roundsByRange exceeds maxCorrectionRounds")
    if not isinstance(history, list) or not history or len(history) != len(set(history)) or not all(isinstance(item, str) and HEX_SHA.fullmatch(item) for item in history):
        errors.append("headHistory must be a non-empty append-only unique SHA list")
    lineage = state.get("candidateLineage")
    if isinstance(rounds, dict) and (not isinstance(lineage, str) or set(rounds) != {lineage}):
        errors.append("roundsByRange must be keyed only by the stable candidate lineage")
    guards = state.get("guards")
    if isinstance(guards, dict) and "correction_rounds" in guards:
        errors.append("canonical phase state must not write the legacy correction counter")
    status = state.get("status")
    if not isinstance(status, str) or status not in PHASE_STATUSES:
        errors.append("phase status is outside the current enum")
    local = state.get("localCodex")
    if not isinstance(local, dict) or local.get("status") not in LOCAL_CODEX_STATUSES:
        errors.append("localCodex status is absent or invalid")
    elif local.get("status") != "pending":
        if not HEX_SHA.fullmatch(str(local.get("exactHeadSHA", ""))) or not HEX_SHA.fullmatch(str(local.get("exactHeadTree", ""))):
            errors.append("non-pending localCodex state lacks exact head/tree")
        elif isinstance(history, list) and local.get("exactHeadSHA") != history[-1]:
            errors.append("localCodex exact head must equal the latest append-only headHistory entry")
    shepherd = state.get("shepherd")
    if not isinstance(shepherd, dict) or shepherd.get("status") not in SHEPHERD_STATUSES:
        errors.append("shepherd status is absent or invalid")
    elif shepherd.get("status") == "pending" and shepherd.get("verdict") is not None:
        errors.append("pending Shepherd must not carry a verdict")
    if not isinstance(state.get("verificationPassed"), bool):
        errors.append("verificationPassed must be boolean")
    if not isinstance(state.get("humanGates"), list) or not all(isinstance(item, str) and item for item in state.get("humanGates", [])):
        errors.append("humanGates must be a string list")
    return errors


def json_pointer(value: Any, pointer: str) -> Any:
    if not isinstance(pointer, str) or not pointer.startswith("/"):
        raise ReviewSystemError("mirror JSON pointer must start with /")
    current = value
    for raw in pointer.split("/")[1:]:
        key = raw.replace("~1", "/").replace("~0", "~")
        if isinstance(current, dict) and key in current:
            current = current[key]
        elif isinstance(current, list) and key.isdigit() and int(key) < len(current):
            current = current[int(key)]
        else:
            raise ReviewSystemError(f"mirror pointer is absent: {pointer}")
    return current


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
        validator = relationship.get("validator")
        if validator == "pm_review_phase_v2":
            try:
                schema = require_object(load_json(resolve_safe(root, relationship["source"])), "phase schema")
                state = load_json(resolve_safe(root, relationship["target"]))
            except ReviewSystemError as exc:
                findings.append(finding("authority_inventory", str(exc), relationship.get("target")))
                continue
            errors = json_schema_findings(state, schema)
            errors.extend(phase_state_semantic_findings(state))
            if errors:
                findings.append(
                    finding(
                        "authoritative_state_consistency",
                        "dedicated PM phase state does not validate: " + "; ".join(errors[:20]),
                        relationship["target"],
                    )
                )
        elif validator == "authority_mirror_v1":
            try:
                source = load_json(resolve_safe(root, relationship["source"]))
                target = load_json(resolve_safe(root, relationship["target"]))
                mappings = relationship.get("field_mappings")
                if not isinstance(mappings, list) or not mappings:
                    raise ReviewSystemError("authority mirror validator requires field_mappings")
                for mapping in mappings:
                    if not isinstance(mapping, dict) or set(mapping) != {"source", "target"}:
                        raise ReviewSystemError("authority mirror mapping is malformed")
                    left = json_pointer(source, mapping["source"])
                    right = json_pointer(target, mapping["target"])
                    if left != right:
                        raise ReviewSystemError(
                            f"authority mirror differs: {mapping['source']} != {mapping['target']}"
                        )
            except ReviewSystemError as exc:
                findings.append(
                    finding("authoritative_state_consistency", str(exc), relationship.get("target"))
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
    *,
    revision: str = "head",
    blob_sha256: str | None = None,
) -> dict[str, Any]:
    if relation not in RELATIONS:
        raise ReviewSystemError(f"impact relation is outside enum: {relation!r}")
    if certainty not in CERTAINTIES:
        raise ReviewSystemError(f"impact certainty is outside enum: {certainty!r}")
    identity = "\0".join((source, target, relation, parser_name, reason, certainty, str(line), revision, blob_sha256 or ""))
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
        "revision": revision,
        "source_blob_sha256": blob_sha256,
    }


def line_certainty(text: str) -> str:
    lowered = text.lower()
    if re.search(r"\b(historical|legacy-only|deprecated|superseded|retired)\b", lowered):
        return "inactive"
    if "${{" in lowered or re.search(r"\b(optional|conditionally|conditional|when|if)\b", lowered):
        return "unknown"
    return "active"


def reference_relation(text: str, matched_path: str = "", structural_keys: Iterable[str] = ()) -> str:
    lowered = text.lower()
    keys = {key.lower() for key in structural_keys}
    if keys & {"required", "requires", "inputs", "dependencies", "template", "contract", "workflow", "schema", "prompt", "source"}:
        return "required_reference"
    if re.search(r"\b(require(?:d|s)?|must\s+(?:read|use)|read|load(?:s)?|consume(?:s)?|write(?:s)?|run)\b", lowered):
        return "required_reference"
    return "descriptive_reference"


def script_invocation_line(text: str) -> bool:
    return bool(re.search(r"(?:^|[\s;|&({=])(?:source|exec|bash|sh|python3)(?:\s|$)|runpy\.run_path", text))


def structural_inactive_lines(relative: str, lines: list[str]) -> set[int]:
    inactive: set[int] = set()
    name = PurePosixPath(relative).name
    if name in {"TDD-LEDGER.md", "VERIFICATION.md"} or name.startswith("REVIEW-") or "/claude-" in relative:
        return set(range(1, len(lines) + 1))
    section_inactive = False
    fixture_continuation = False
    data_heredoc: str | None = None
    for number, line in enumerate(lines, 1):
        stripped = line.strip()
        if PurePosixPath(relative).suffix == ".md" and stripped.startswith("#"):
            heading = stripped.lstrip("#").strip().lower()
            section_inactive = bool(re.search(r"\b(historical|forbidden|excluded|deprecated|collision boundary)\b", heading))
        if section_inactive:
            inactive.add(number)
        if PurePosixPath(relative).suffix == ".sh":
            if data_heredoc is not None:
                inactive.add(number)
                if stripped == data_heredoc:
                    data_heredoc = None
                continue
            marker = re.search(r"<<-?['\"]?([A-Za-z_][A-Za-z0-9_]*)['\"]?", line)
            if marker and not re.search(r"(?:^|\s)python3?\s+-?\s*<<", line):
                data_heredoc = marker.group(1)
                inactive.add(number)
            if fixture_continuation or any(value in line for value in ("$impact_repo", "$test_tmp", "$fixture_root")):
                inactive.add(number)
                fixture_continuation = line.rstrip().endswith("\\")
            elif fixture_continuation:
                fixture_continuation = line.rstrip().endswith("\\")
    return inactive


def yaml_structural_context(lines: list[str]) -> dict[int, tuple[str, ...]]:
    result: dict[int, tuple[str, ...]] = {}
    stack: list[tuple[int, str]] = []
    for number, line in enumerate(lines, 1):
        indent = len(line) - len(line.lstrip(" "))
        while stack and stack[-1][0] >= indent:
            stack.pop()
        match = re.match(r"\s*(?:-\s*)?([A-Za-z_][A-Za-z0-9_-]*)\s*:", line)
        if match:
            stack.append((indent, match.group(1).lower()))
        result[number] = tuple(key for _, key in stack)
    return result


def json_reference_records(relative: str, text: str) -> tuple[list[tuple[str, tuple[str, ...]]], list[dict[str, Any]]]:
    try:
        document = json.loads(text)
    except json.JSONDecodeError as exc:
        return [], [finding("impact_parser", f"JSON syntax prevents structural parsing: {exc}", relative)]
    records: list[tuple[str, tuple[str, ...]]] = []
    def visit(value: Any, keys: tuple[str, ...]) -> None:
        if isinstance(value, dict):
            for key, item in value.items():
                visit(item, (*keys, str(key).lower()))
        elif isinstance(value, list):
            for item in value:
                visit(item, keys)
        elif isinstance(value, str):
            for raw in reference_candidates(value):
                records.append((raw, keys))
    visit(document, ())
    return records, []


def python_import_targets(relative: str, node: ast.AST, universe: set[str]) -> list[tuple[str, str]]:
    parent = PurePosixPath(relative).parent
    modules: list[tuple[int, str]] = []
    if isinstance(node, ast.Import):
        modules.extend((0, alias.name) for alias in node.names)
    elif isinstance(node, ast.ImportFrom):
        module = node.module or ""
        for alias in node.names:
            suffix = module or alias.name
            modules.append((node.level, suffix))
    result: list[tuple[str, str]] = []
    for level, module in modules:
        if level:
            base = parent
            for _ in range(max(0, level - 1)):
                base = base.parent
            module_path = module.replace(".", "/")
            candidates = [(base / (module_path + ".py")).as_posix(), (base / module_path / "__init__.py").as_posix()]
        else:
            module_path = module.replace(".", "/")
            candidates = [module_path + ".py", module_path + "/__init__.py", (parent / (module.rsplit(".", 1)[-1] + ".py")).as_posix()]
        target = next((candidate for candidate in candidates if candidate in universe), None)
        if target:
            result.append((target, "." * level + module))
    return result


def typed_file_edges(
    relative: str,
    text: str,
    universe: set[str],
    *,
    revision: str = "head",
) -> tuple[list[dict[str, Any]], list[dict[str, Any]]]:
    if relative.endswith(("pm-review-system.json", "REVIEW-SCOPE.json")) or relative.startswith("scripts/tests/fixtures/"):
        return [], []
    result: list[dict[str, Any]] = []
    findings: list[dict[str, Any]] = []
    seen: set[str] = set()
    lines = text.splitlines()
    suffix = PurePosixPath(relative).suffix
    source_digest = hashlib.sha256(text.encode()).hexdigest()
    structurally_inactive = structural_inactive_lines(relative, lines)
    yaml_context = yaml_structural_context(lines) if suffix in {".yaml", ".yml"} else {}

    def add(raw: str, number: int, parser_name: str, reason: str, context: Iterable[str] = (), certainty: str | None = None) -> None:
        if suffix == ".sh":
            raw = re.sub(r"^(?:repo_root|REPO_ROOT|ROOT_DIR)/", "", raw)
            if re.match(r"^[A-Z][A-Z0-9_]*/", raw):
                return
        repository_rooted = raw.startswith((".agents/", ".pi/", ".gsd/", "scripts/", ".planning/", "cmd/", "internal/", "docs/", "website/", "./", "../")) or raw in {"AGENTS.md", "CLAUDE.md"}
        source_line = lines[number - 1] if number and number <= len(lines) else ""
        if suffix == ".md" and not repository_rooted and f"`{raw}`" not in source_line and not any(match.group(1) == raw for match in MARKDOWN_LINK.finditer(source_line)):
            return
        actual_certainty = certainty or ("inactive" if number in structurally_inactive else line_certainty(source_line))
        relation = "script_invokes" if suffix == ".sh" and script_invocation_line(source_line) else reference_relation(source_line, raw, context)
        if relation == "descriptive_reference" and certainty == "active" and parser_name in {
            "shell_python_heredoc_ast", "python_ast"
        }:
            relation = "required_reference"
        elif relation == "descriptive_reference":
            actual_certainty = "inactive"
        try:
            target = normalize_reference(relative, raw)
        except ReviewSystemError as exc:
            if relation != "descriptive_reference" and actual_certainty != "inactive":
                findings.append(finding("impact_path", str(exc), relative))
            return
        if not target:
            return
        edge = stable_edge(relative, target, relation, parser_name, reason, actual_certainty, number, revision=revision, blob_sha256=source_digest)
        if edge["id"] not in seen:
            seen.add(edge["id"])
            result.append(edge)

    if suffix == ".json":
        records, parser_findings = json_reference_records(relative, text)
        findings.extend(parser_findings)
        for raw, keys in records:
            add(raw, 0, "json", "json_structural_path", keys, "unknown" if "optional" in keys else "active")
    else:
        for number, line_text in enumerate(lines, 1):
            if suffix == ".sh" and not script_invocation_line(line_text):
                continue
            context = yaml_context.get(number, ())
            parser_name = "yaml" if suffix in {".yaml", ".yml"} else ("markdown" if suffix == ".md" else "text_path")
            reason = "yaml_structural_path" if context else "inline_repository_path"
            for raw in reference_candidates(line_text):
                add(raw, number, parser_name, reason, context)

    if suffix == ".sh":
        index = 0
        while index < len(lines):
            opener = lines[index]
            marker = re.search(r"<<-?['\"]?([A-Za-z_][A-Za-z0-9_]*)['\"]?", opener)
            if marker:
                delimiter = marker.group(1)
                end = index + 1
                while end < len(lines) and lines[end].strip() != delimiter:
                    end += 1
                executed_python = bool(re.search(r"(?:^|\s)python3?\s+-?\s*<<", opener))
                if executed_python and end < len(lines):
                    body = "\n".join(lines[index + 1:end]) + "\n"
                    try:
                        tree = ast.parse(body, filename=relative + ":heredoc")
                        for node in ast.walk(tree):
                            for target, module in python_import_targets(relative, node, universe):
                                edge = stable_edge(relative, target, "python_import", "shell_python_heredoc_ast", f"heredoc_import:{module}", "active", index + 2 + getattr(node, "lineno", 0), revision=revision, blob_sha256=source_digest)
                                if edge["id"] not in seen:
                                    seen.add(edge["id"]); result.append(edge)
                            if isinstance(node, ast.Constant) and isinstance(node.value, str):
                                for raw in reference_candidates(node.value):
                                    add(raw, index + 1 + getattr(node, "lineno", 1), "shell_python_heredoc_ast", "executed_heredoc_path", (), "active")
                    except SyntaxError as exc:
                        findings.append(finding("impact_parser", f"executed Python heredoc syntax is invalid: {exc}", relative))
                index = end
            index += 1

    if suffix == ".py":
        try:
            tree = ast.parse(text, filename=relative)
        except SyntaxError as exc:
            findings.append(finding("impact_parser", f"Python syntax prevents authoritative import parsing: {exc}", relative))
            return sorted(result, key=lambda item: item["id"]), findings
        for node in ast.walk(tree):
            for target, module in python_import_targets(relative, node, universe):
                edge = stable_edge(relative, target, "python_import", "python_ast", f"import:{module}", "active", getattr(node, "lineno", 0), revision=revision, blob_sha256=source_digest)
                if edge["id"] not in seen:
                    seen.add(edge["id"]); result.append(edge)
    return sorted(result, key=lambda item: item["id"]), findings


def validate_config(config: Any) -> dict[str, Any]:
    config = require_object(config, "review-system config")
    if config.get("schema_version") != CONFIG_SCHEMA:
        raise ReviewSystemError(
            f"review-system config migration required: {config.get('schema_version')!r} != {CONFIG_SCHEMA!r}"
        )
    if config.get("owner") != "parent_orchestrator":
        raise ReviewSystemError("review-system owner must be parent_orchestrator")
    for field in ("canonical_roots", "reference_prefixes", "explicit_reference_files", "prompt_contract_files"):
        values = config.get(field)
        if values is None and field == "prompt_contract_files":
            continue
        if not isinstance(values, list) or not all(isinstance(item, str) and item for item in values):
            raise ReviewSystemError(f"review-system {field} must be a string list")
    for field in ("canonical_roots", "explicit_reference_files", "prompt_contract_files", "prohibited_active_targets"):
        for path in config.get(field, []):
            validate_relative_path(path, f"review-system {field} path")
    for field in ("reference_prefixes", "ignored_reference_prefixes"):
        for prefix in config.get(field, []):
            if CONTROL.search(prefix) or prefix.startswith(("-", "/")) or "\\" in prefix or ".." in PurePosixPath(prefix).parts:
                raise ReviewSystemError(f"review-system {field} contains an unsafe prefix")
    settings = require_object(config.get("impact_graph"), "impact_graph")
    for field in ("index_prefixes", "go_index_prefixes"):
        values = settings.get(field, [])
        if not isinstance(values, list):
            raise ReviewSystemError(f"impact_graph {field} must be a list")
        for prefix in values:
            if not isinstance(prefix, str) or CONTROL.search(prefix) or prefix.startswith(("-", "/")) or "\\" in prefix or ".." in PurePosixPath(prefix).parts:
                raise ReviewSystemError(f"impact_graph {field} contains an unsafe prefix")
    positive_limits = {
        "max_index_files", "max_index_bytes", "max_nodes", "max_edges", "max_traversal_states",
        "max_depth", "max_impact_files", "max_impact_edges", "go_command_timeout_seconds",
        "go_max_output_bytes", "go_max_packages", "packet_max_impact_files",
        "packet_max_impact_edges", "max_packets", "edge_context_max_bytes_per_file",
        "packet_max_file_slice_bytes",
    }
    for key in positive_limits:
        value = settings.get(key)
        numeric_type = (int, float) if key == "go_command_timeout_seconds" else (int,)
        if not isinstance(value, numeric_type) or isinstance(value, bool) or value <= 0:
            raise ReviewSystemError(f"impact_graph {key} must be a positive {'number' if key == 'go_command_timeout_seconds' else 'integer'}")
    if settings["max_packets"] > 64:
        raise ReviewSystemError("impact_graph max_packets exceeds the hard v4 maximum of 64")
    for policy_field in ("default_relation_policy", "relation_policy"):
        value = require_object(settings.get(policy_field), f"impact_graph {policy_field}")
        records = {"default": value} if policy_field == "default_relation_policy" else value
        for relation, policy in records.items():
            if relation != "default" and relation not in RELATIONS:
                raise ReviewSystemError(f"relation_policy key is outside enum: {relation!r}")
            if not isinstance(policy, dict) or set(policy) != DIRECTIONS:
                raise ReviewSystemError(f"relation policy {relation} must contain exactly four directions")
            if not all(isinstance(limit, int) and not isinstance(limit, bool) and limit >= 0 for limit in policy.values()):
                raise ReviewSystemError(f"relation policy {relation} limits must be non-negative integers")
    for record in config.get("configured_relationships", []):
        if not isinstance(record, dict):
            raise ReviewSystemError("configured relationship must be an object")
        for endpoint in ("source", "target"):
            validate_relative_path(record.get(endpoint), f"configured relationship {endpoint}")
        if record.get("relation") not in RELATIONS:
            raise ReviewSystemError(f"configured relationship relation is outside enum: {record.get('relation')!r}")
        if record.get("certainty") not in CERTAINTIES:
            raise ReviewSystemError(f"configured relationship certainty is outside enum: {record.get('certainty')!r}")
        if record.get("validator") not in {None, "pm_review_phase_v2", "authority_mirror_v1"}:
            raise ReviewSystemError(f"configured relationship validator is outside enum: {record.get('validator')!r}")
    for record in config.get("authorities", []):
        if not isinstance(record, dict) or not isinstance(record.get("id"), str):
            raise ReviewSystemError("authority record is malformed")
        validate_relative_path(record.get("authoritative_path"), "authority path")
        for field in ("writers", "readers", "mirrors"):
            values = record.get(field, [])
            if not isinstance(values, list):
                raise ReviewSystemError(f"authority {field} must be a list")
            for endpoint in values:
                validate_relative_path(endpoint, f"authority {field} endpoint")
    roles = {"combined", "architecture_reference", "authority_workflow_state", "implementation_test", "impact_graph"}
    for rule in config.get("domain_rules", []):
        if not isinstance(rule, dict) or rule.get("domain") not in roles or not isinstance(rule.get("patterns"), list):
            raise ReviewSystemError("domain rule has an invalid role or patterns")
        for pattern in rule["patterns"]:
            if not isinstance(pattern, str) or CONTROL.search(pattern) or pattern.startswith(("-", "/")) or "\\" in pattern or ".." in PurePosixPath(pattern).parts:
                raise ReviewSystemError("domain rule contains an unsafe pattern")
    invariants = require_object(config.get("packet_invariants"), "packet_invariants")
    if "combined" not in invariants or any(role not in roles for role in invariants):
        raise ReviewSystemError("packet invariant role is outside enum or combined is absent")
    for role, values in invariants.items():
        if not isinstance(values, list) or not values or not all(isinstance(item, str) and SAFE_ID.fullmatch(item) for item in values) or len(values) != len(set(values)):
            raise ReviewSystemError(f"packet invariants for {role} must be a non-empty unique safe-id list")
    thresholds = require_object(config.get("thresholds"), "thresholds")
    for key in (
        "combined_max_files", "combined_max_non_generated_lines", "combined_max_domains",
        "packet_max_changed_files", "packet_max_context_files", "packet_target_tokens",
        "response_reserve_tokens", "context_window_tokens",
    ):
        value = thresholds.get(key)
        if not isinstance(value, int) or isinstance(value, bool) or value <= 0:
            raise ReviewSystemError(f"threshold {key} must be a positive integer")
    if thresholds["packet_target_tokens"] > 30000:
        raise ReviewSystemError("packet_target_tokens exceeds the hard v4 rendered-input maximum of 30000")
    if thresholds["packet_target_tokens"] + thresholds["response_reserve_tokens"] > thresholds["context_window_tokens"]:
        raise ReviewSystemError("packet input target plus response reserve exceeds the context window")
    return config


def decode_json_stream(text: str, maximum_items: int | None = None) -> list[dict[str, Any]]:
    decoder = json.JSONDecoder()
    position = 0
    result: list[dict[str, Any]] = []
    while position < len(text):
        while position < len(text) and text[position].isspace():
            position += 1
        if position >= len(text):
            break
        try:
            item, position = decoder.raw_decode(text, position)
        except json.JSONDecodeError as exc:
            raise ReviewSystemError(f"bounded Go JSON stream is malformed: {exc}") from exc
        if not isinstance(item, dict):
            raise ReviewSystemError("bounded Go JSON stream contains a non-object package")
        result.append(item)
        if maximum_items is not None and len(result) > maximum_items:
            raise ReviewSystemError(f"Go package bound exceeded: > {maximum_items}")
    return result


def go_impact_edges(
    root: Path,
    timeout_seconds: float,
    maximum_output_bytes: int,
    maximum_packages: int,
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
            try:
                module_cache = Path(module_cache_value).resolve(strict=True)
            except OSError as exc:
                findings.append(finding("impact_go", f"configured module cache is unavailable: {exc}"))
                return nodes, edges, {"status": "blocked", "reason": "configured module cache unavailable"}, findings
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
        stdout_path = temp / "go-list.stdout"
        stderr_path = temp / "go-list.stderr"
        try:
            with stdout_path.open("wb") as stdout_file, stderr_path.open("wb") as stderr_file:
                proc = subprocess.Popen(
                    ["go", "list", "-mod=mod", "-json", "-deps", "-test", "./..."],
                    cwd=root,
                    env=env,
                    stdin=subprocess.DEVNULL,
                    stdout=stdout_file,
                    stderr=stderr_file,
                    start_new_session=True,
                )
                deadline = time.monotonic() + float(timeout_seconds)
                bound_hit: str | None = None
                while proc.poll() is None:
                    if time.monotonic() >= deadline:
                        bound_hit = "timeout"
                    elif stdout_path.stat().st_size + stderr_path.stat().st_size > maximum_output_bytes:
                        bound_hit = "output"
                    if bound_hit:
                        try:
                            os.killpg(proc.pid, 9)
                        except ProcessLookupError:
                            pass
                        proc.wait()
                        findings.append(finding("impact_go", f"authoritative go list hit {bound_hit} bound"))
                        return nodes, edges, {"status": "blocked", "reason": f"go list {bound_hit} bound"}, findings
                    time.sleep(0.01)
                return_code = proc.wait()
            output_bytes = stdout_path.stat().st_size + stderr_path.stat().st_size
            if output_bytes > maximum_output_bytes:
                findings.append(finding("impact_go", "authoritative go list output bound exceeded"))
                return nodes, edges, {"status": "blocked", "reason": "go list output bound"}, findings
            stdout_text = stdout_path.read_text(errors="replace")
            stderr_text = stderr_path.read_text(errors="replace")
        except OSError as exc:
            findings.append(finding("impact_go", f"authoritative go list failed: {exc}"))
            return nodes, edges, {"status": "blocked", "reason": str(exc)}, findings
    if return_code != 0:
        detail = stderr_text.strip() or stdout_text.strip()
        findings.append(finding("impact_go", f"authoritative go list failed: {detail[:1000]}"))
        return nodes, edges, {"status": "blocked", "reason": detail[:1000]}, findings

    try:
        packages = decode_json_stream(stdout_text, maximum_packages)
    except ReviewSystemError as exc:
        findings.append(finding("impact_go", str(exc)))
        return nodes, edges, {"status": "blocked", "reason": str(exc)}, findings
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
    tracked_head_raw = run_git(root, ["ls-tree", "-r", "--name-only", "-z", head, "--"])
    tracked_base_raw = run_git(root, ["ls-tree", "-r", "--name-only", "-z", base, "--"])
    tracked_head = {path for path in tracked_head_raw.split("\0") if path}
    tracked_base = {path for path in tracked_base_raw.split("\0") if path}
    tracked = tracked_head | {path for path in tracked_base if path in changed and PurePosixPath(path).suffix in INDEX_SUFFIXES}
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

    include_go = any(path.endswith((".go", "go.mod", "go.sum")) for path in changed)
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
    node_types: dict[str, str] = {
        path: ("file" if path in tracked_head else "deleted_file") for path in selected
    }
    file_bytes: dict[str, int] = {}
    edges: list[dict[str, Any]] = []
    edge_ids: set[str] = set()
    total_bytes = 0
    parsed: set[tuple[str, str]] = set()
    pending = deque((path, "head" if path in tracked_head else "base") for path in sorted(selected))
    for path in sorted(set(changed) & tracked_head & tracked_base & selected):
        if PurePosixPath(path).suffix in INDEX_SUFFIXES:
            pending.append((path, "base"))
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
        pending.append((relative, "head" if relative in tracked_head else "base"))
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
        relative, revision_name = pending.popleft()
        parsed_key = (relative, revision_name)
        if parsed_key in parsed:
            continue
        parsed.add(parsed_key)
        revision_sha = head if revision_name == "head" else base
        if (revision_name == "head" and relative not in tracked_head) or (revision_name == "base" and relative not in tracked_base):
            continue
        try:
            blob = git_blob(root, revision_sha, relative, maximum=int(settings["max_index_bytes"]))
            size = len(blob)
            if total_bytes + size > int(settings["max_index_bytes"]):
                findings.append(
                    finding(
                        "impact_graph_bound",
                        f"index byte bound would be exceeded before reading {relative}@{revision_name}: {total_bytes + size} > {settings['max_index_bytes']}",
                        relative,
                    )
                )
                index_stopped = True
                break
            text = blob.decode()
        except (ReviewSystemError, OSError, UnicodeDecodeError) as exc:
            findings.append(finding("impact_index", f"cannot index exact review-relevant blob: {exc}", relative))
            continue
        total_bytes += size
        file_bytes[relative] = max(file_bytes.get(relative, 0), size)
        parsed_edges, parser_findings = typed_file_edges(relative, text, tracked | nodes, revision=revision_name)
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
            if endpoint not in tracked_head:
                findings.append(finding("impact_graph", f"configured {edge['certainty']} endpoint is unresolved at exact head: {endpoint}", edge["source"]))
            else:
                include_node(endpoint, edge["source"])
        if edge["target"] in prohibited and edge["certainty"] != "inactive":
            findings.append(finding("impact_graph", f"configured edge reaches prohibited target {edge['target']} ({edge['id']})", edge["source"]))
        append_edge(edge)

    reachable_go_trigger = include_go or any(path.endswith(".go") for path in nodes)
    go_context: dict[str, Any] = {"status": "not_needed", "reason": "no changed or indirectly reachable Go input"}
    if reachable_go_trigger:
        try:
            head_snapshot_tmp = clone_commit_snapshot(root, head)
            with head_snapshot_tmp:
                head_snapshot = Path(head_snapshot_tmp.name) / "repo"
                go_nodes, go_edges, go_context, go_findings = go_impact_edges(
                    head_snapshot,
                    settings["go_command_timeout_seconds"],
                    int(settings["go_max_output_bytes"]),
                    int(settings["go_max_packages"]),
                )
        except ReviewSystemError as exc:
            go_nodes, go_edges = set(), []
            go_context = {"status": "blocked", "reason": str(exc)}
            go_findings = [finding("impact_go", str(exc))]
        findings.extend(go_findings)
        changed_go_in_base = sorted(path for path in changed if path.endswith(".go") and path in tracked_base)
        deleted_go = sorted(path for path in changed_go_in_base if path not in tracked_head)
        if changed_go_in_base:
            try:
                snapshot_tmp = clone_commit_snapshot(root, base)
                with snapshot_tmp:
                    snapshot = Path(snapshot_tmp.name) / "repo"
                    base_nodes, base_edges, base_context, base_findings = go_impact_edges(
                        snapshot,
                        settings["go_command_timeout_seconds"],
                        int(settings["go_max_output_bytes"]),
                        int(settings["go_max_packages"]),
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
                        revision="base",
                        blob_sha256=edge.get("source_blob_sha256"),
                    )
                    go_edges.append(base_edge)
                findings.extend(base_findings)
                go_context = {
                    **go_context,
                    "base_changed_file_context": base_context,
                    "base_changed_go_files": changed_go_in_base,
                    "deleted_go_files": deleted_go,
                }
            except ReviewSystemError as exc:
                findings.append(finding("impact_go", f"changed/deleted Go base context failed: {exc}"))
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
            source = edge["source"]
            target = edge["target"]
            if source not in nodes:
                findings.append(finding("impact_graph", f"{edge['certainty']} indexed source is absent: {source}", source))
            if target not in nodes:
                findings.append(finding("impact_graph", f"{edge['certainty']} indexed target is absent: {target}", source))
            elif node_types.get(target) == "deleted_file" and edge.get("revision") != "base" and not edge.get("reason", "").startswith("base_deleted_context:"):
                findings.append(finding("impact_graph", f"{edge['certainty']} indexed target was deleted: {target}", source))

    return {
        "nodes": nodes,
        "node_types": node_types,
        "edges": indexed_edges,
        "tracked": tracked,
        "index_files": sorted({path for path, _ in parsed}),
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


def exact_blob_metadata(
    root: Path,
    base: str,
    head: str,
    paths: Iterable[str],
    maximum_slice_bytes: int,
) -> tuple[dict[str, int], dict[str, list[dict[str, Any]]]]:
    sizes: dict[str, int] = {}
    slices: dict[str, list[dict[str, Any]]] = {}
    for relative in sorted(set(paths)):
        if relative.startswith("go-package:"):
            continue
        validate_relative_path(relative, "packet context path")
        head_blob = git_blob_optional(root, head, relative)
        base_blob = git_blob_optional(root, base, relative)
        versions: list[tuple[str, str, bytes]] = []
        if head_blob is not None:
            versions.append((head, "head", head_blob))
        if base_blob is not None and (head_blob is None or hashlib.sha256(base_blob).digest() != hashlib.sha256(head_blob).digest()):
            versions.append((base, "base_deleted" if head_blob is None else "base_changed", base_blob))
        if not versions:
            raise ReviewSystemError(f"packet context has no exact base/head blob: {relative}")
        sizes[relative] = sum(len(blob) for _, _, blob in versions)
        path_slices: list[dict[str, Any]] = []
        for revision, revision_kind, blob in versions:
            blob_digest = hashlib.sha256(blob).hexdigest()
            lines = blob.splitlines(keepends=True) or [b""]
            segments: list[tuple[int, bytes]] = []
            for line_number, line in enumerate(lines, 1):
                if not line:
                    segments.append((line_number, b""))
                else:
                    for offset in range(0, len(line), maximum_slice_bytes):
                        segments.append((line_number, line[offset:offset + maximum_slice_bytes]))
            group: list[tuple[int, bytes]] = []
            group_bytes = 0
            start_byte = 0
            def flush_group() -> None:
                nonlocal group, group_bytes, start_byte
                if not group:
                    return
                payload = b"".join(piece for _, piece in group)
                path_slices.append(
                    {
                        "path": relative,
                        "revision": revision,
                        "revision_kind": revision_kind,
                        "blob_sha256": blob_digest,
                        "start_line": group[0][0],
                        "end_line": group[-1][0],
                        "start_byte": start_byte,
                        "end_byte": start_byte + len(payload),
                        "bytes": len(payload),
                        "sha256": hashlib.sha256(payload).hexdigest(),
                    }
                )
                start_byte += len(payload)
                group = []
                group_bytes = 0
            for line_number, piece in segments:
                if group and group_bytes + len(piece) > maximum_slice_bytes:
                    flush_group()
                group.append((line_number, piece))
                group_bytes += len(piece)
            flush_group()
        slices[relative] = path_slices
    return sizes, slices


def exact_diff_metadata(
    root: Path,
    base: str,
    head: str,
    paths: Iterable[str],
    maximum_slice_bytes: int,
) -> tuple[dict[str, int], dict[str, list[dict[str, Any]]]]:
    """Return exact bounded diff slices for changed-file review packets.

    Changed packets review the patch, not duplicate full base and head blobs. Context and edge
    packets remain revision/blob-bound through exact_blob_metadata.
    """
    sizes: dict[str, int] = {}
    slices: dict[str, list[dict[str, Any]]] = {}
    for relative in sorted(set(paths)):
        validate_relative_path(relative, "changed diff path")
        payload = run_git_bytes_bounded(
            root,
            [
                "diff", "--no-ext-diff", "--no-renames", "--binary", "--full-index",
                "--unified=3", f"{base}...{head}", "--", relative,
            ],
            MAX_EXACT_DIFF_BYTES,
        )
        if not payload:
            raise ReviewSystemError(f"changed path has no exact diff payload: {relative}")
        digest = hashlib.sha256(payload).hexdigest()
        sizes[relative] = len(payload)
        lines = payload.splitlines(keepends=True) or [b""]
        path_slices: list[dict[str, Any]] = []
        current: list[bytes] = []
        current_bytes = 0
        start_line = 1
        start_byte = 0

        def flush(end_line: int) -> None:
            nonlocal current, current_bytes, start_line, start_byte
            if not current:
                return
            chunk = b"".join(current)
            path_slices.append(
                {
                    "path": relative,
                    "revision": head,
                    "revision_kind": "diff",
                    "blob_sha256": digest,
                    "start_line": start_line,
                    "end_line": end_line,
                    "start_byte": start_byte,
                    "end_byte": start_byte + len(chunk),
                    "bytes": len(chunk),
                    "sha256": hashlib.sha256(chunk).hexdigest(),
                }
            )
            start_byte += len(chunk)
            current = []
            current_bytes = 0
            start_line = end_line + 1

        for line_number, line in enumerate(lines, 1):
            pieces = [line[offset:offset + maximum_slice_bytes] for offset in range(0, len(line), maximum_slice_bytes)] or [b""]
            for piece in pieces:
                if current and current_bytes + len(piece) > maximum_slice_bytes:
                    flush(line_number - 1 if current[-1].endswith(b"\n") else line_number)
                if not current:
                    start_line = line_number
                current.append(piece)
                current_bytes += len(piece)
        flush(len(lines))
        if sum(item["bytes"] for item in path_slices) != len(payload):
            raise ReviewSystemError(f"changed diff slicing is incomplete: {relative}")
        slices[relative] = path_slices
    return sizes, slices


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
    blob_slices: dict[str, list[dict[str, Any]]],
    config: dict[str, Any],
    changed_content_sizes: dict[str, int] | None = None,
    changed_content_slices: dict[str, list[dict[str, Any]]] | None = None,
    edge_blob_slices: dict[str, list[dict[str, Any]]] | None = None,
    context_blob_slices: dict[str, list[dict[str, Any]]] | None = None,
) -> tuple[str, list[dict[str, Any]], list[dict[str, Any]]]:
    thresholds = config["thresholds"]
    changed_content_sizes = changed_content_sizes or {path: blob_sizes[path] for path in files}
    changed_content_slices = changed_content_slices or {path: blob_slices[path] for path in files}
    edge_blob_slices = edge_blob_slices or blob_slices
    context_blob_slices = context_blob_slices or blob_slices
    graph_limits = config["impact_graph"]
    domain_values = sorted(set(domains.values()))
    changed_lines = sum(line_counts.values())
    combined = (
        len(files) <= thresholds["combined_max_files"]
        and changed_lines <= thresholds["combined_max_non_generated_lines"]
        and len(domain_values) <= thresholds["combined_max_domains"]
        and sum(changed_content_sizes.get(path, 0) for path in files)
        + int(thresholds.get("rendered_prompt_fixed_bytes", 4096))
        + 4096 <= int(thresholds["packet_target_tokens"])
    )
    selection = "combined" if combined else "split"
    packets: list[dict[str, Any]] = []
    problems: list[dict[str, Any]] = []
    context_paths = set(closure_files) | set(authority_files)
    for path in sorted(set(files) | context_paths | set(impact_files)):
        source = context_blob_slices if path in context_paths else blob_slices
        assigned_slices = source.get(path)
        if not isinstance(assigned_slices, list) or not assigned_slices or sum(int(item.get("bytes", -1)) for item in assigned_slices if isinstance(item, dict)) != blob_sizes.get(path):
            problems.append(finding("packet_overflow", f"exact blob slices are absent or incomplete: {path}"))
    impact_edge_by_id = {edge["id"]: edge for edge in impact_edges}
    target_tokens = int(thresholds["packet_target_tokens"])
    response_reserve = int(thresholds["response_reserve_tokens"])
    context_window = int(thresholds["context_window_tokens"])
    fixed_prompt_bytes = int(thresholds.get("rendered_prompt_fixed_bytes", 4096))

    def file_tokens(paths: Iterable[str]) -> int:
        return sum(max(1, blob_sizes[path]) for path in sorted(set(paths)))

    def changed_tokens(paths: Iterable[str]) -> int:
        return sum(max(1, changed_content_sizes[path]) for path in sorted(set(paths)))

    def edge_context_tokens(paths: Iterable[str]) -> int:
        maximum = min(
            int(graph_limits.get("edge_context_max_bytes_per_file", 8192)),
            int(graph_limits.get("packet_max_file_slice_bytes", 8192)),
        )
        return sum(max(1, min(blob_sizes[path], maximum)) for path in sorted(set(paths)))

    def estimate(
        changed: list[str],
        closure: list[str],
        authority: list[str],
        impact: list[str],
        edge_context: list[str],
        edge_ids: list[str],
        impact_slices: list[dict[str, Any]] | None = None,
    ) -> int:
        changed_cost = changed_tokens(changed)
        full_context = set(closure) | set(authority)
        if impact_slices is None:
            full_context.update(impact)
            impact_cost = 0
        else:
            impact_cost = sum(int(item["bytes"]) for item in impact_slices)
        context_cost = file_tokens(full_context) + impact_cost + edge_context_tokens(set(edge_context) - full_context - set(impact))
        metadata_bytes = len(canonical_json_bytes([impact_edge_by_id[edge_id] for edge_id in edge_ids]))
        return fixed_prompt_bytes + changed_cost + context_cost + metadata_bytes + 4096

    def exact_edge_slices(edge_ids: Iterable[str]) -> list[dict[str, Any]]:
        selected: dict[tuple[str, str, int, int], dict[str, Any]] = {}
        for edge_id in edge_ids:
            edge = impact_edge_by_id[edge_id]
            revision_sha = base if edge.get("revision") == "base" else head
            for endpoint, line in ((edge["source"], int(edge.get("line", 0))), (edge["target"], 0)):
                if endpoint.startswith("go-package:"):
                    continue
                candidates = [item for item in edge_blob_slices.get(endpoint, []) if item["revision"] == revision_sha]
                if not candidates:
                    candidates = list(edge_blob_slices.get(endpoint, []))
                if not candidates:
                    continue
                matching = [
                    value for value in candidates
                    if line > 0 and value["start_line"] <= line <= value["end_line"]
                ]
                # A long source line may span several bounded chunks with the same line number.
                # Include every such chunk; selecting the first would silently lose provenance.
                for item in matching or candidates[:1]:
                    selected[(item["path"], item["revision"], item["start_byte"], item["end_byte"])] = item
        return sorted(selected.values(), key=lambda item: (item["path"], item["revision"], item["start_byte"]))

    def account_packet(packet: dict[str, Any]) -> None:
        all_slices = (
            packet.get("changed_file_slices", []) + packet.get("context_file_slices", [])
            + packet.get("impact_file_slices", []) + packet.get("edge_context_slices", [])
        )
        unique_slices = {
            (item["path"], item["revision"], item["start_byte"], item["end_byte"], item["sha256"]): item
            for item in all_slices
        }
        ordered_slices = sorted(
            unique_slices.values(),
            key=lambda item: (item["path"], item["revision"], item["start_byte"], item["end_byte"], item["sha256"]),
        )
        payload_bytes = sum(int(item["bytes"]) for item in ordered_slices)
        separator_bytes = len(ordered_slices) * len(SLICE_SEPARATOR)
        for _ in range(4):
            envelope_bytes = len(canonical_json_bytes(packet))
            prompt_bytes = fixed_prompt_bytes + envelope_bytes + payload_bytes + separator_bytes
            total = prompt_bytes + response_reserve
            packet["context"].update(
                {
                    "estimated_tokens": total,
                    "input_token_upper_bound": prompt_bytes,
                    "total_context_upper_bound": total,
                    "rendered_prompt_bytes": prompt_bytes,
                    "packet_envelope_bytes": envelope_bytes,
                    "slice_payload_bytes": payload_bytes,
                    "slice_separator_bytes": separator_bytes,
                    "overflow": prompt_bytes > target_tokens or total > context_window,
                }
            )

    def append_packet(
        role: str,
        changed: list[str],
        closure: list[str],
        authority: list[str],
        impact: list[str] | None = None,
        impact_edge_ids: list[str] | None = None,
        edge_context_files: list[str] | None = None,
        impact_file_slices: list[dict[str, Any]] | None = None,
        changed_file_slices_override: list[dict[str, Any]] | None = None,
        context_file_slices_override: list[dict[str, Any]] | None = None,
    ) -> None:
        impact = sorted(impact or [])
        impact_edge_ids = sorted(impact_edge_ids or [])
        edge_context_files = sorted(edge_context_files or [])
        packet_number = 1 + sum(1 for packet in packets if packet["role"] == role)
        impact_file_slices = sorted(
            impact_file_slices or [], key=lambda item: (item["path"], item["start_line"], item["end_line"])
        )
        changed_file_slices = (
            sorted(changed_file_slices_override, key=lambda item: (item["path"], item["start_byte"]))
            if changed_file_slices_override is not None
            else [item for path in sorted(changed) for item in changed_content_slices.get(path, [])]
        )
        context_file_slices = (
            sorted(
                context_file_slices_override,
                key=lambda item: (item["path"], item["revision"], item["start_byte"]),
            )
            if context_file_slices_override is not None
            else [
                item
                for path in sorted(set(closure) | set(authority))
                for item in context_blob_slices.get(path, [])
            ]
        )
        edge_context_slices = exact_edge_slices(impact_edge_ids)
        maximum_edge_bytes = int(graph_limits.get("edge_context_max_bytes_per_file", 8192))
        by_edge_path: dict[str, int] = defaultdict(int)
        for item in edge_context_slices:
            by_edge_path[item["path"]] += int(item["bytes"])
        for path, used in by_edge_path.items():
            if used > maximum_edge_bytes:
                problems.append(finding("packet_overflow", f"exact edge provenance slices exceed per-file bound for {path}: {used} > {maximum_edge_bytes}"))
        packet = {
            "schema_version": PACKET_SCHEMA,
            "packet_id": f"{role}-{packet_number:02d}",
            "role": role,
            "exact_base_sha": base,
            "exact_head_sha": head,
            "exact_head_tree": head_tree,
            "changed_files": sorted(changed),
            "changed_file_slices": changed_file_slices,
            "closure_files": sorted(closure),
            "authority_files": sorted(authority),
            "context_file_slices": context_file_slices,
            "impact_files": impact,
            "impact_edge_ids": impact_edge_ids,
            "impact_edges": [impact_edge_by_id[edge_id] for edge_id in impact_edge_ids],
            "impact_file_slices": impact_file_slices,
            "edge_context_files": edge_context_files,
            "edge_context_slices": edge_context_slices,
            "invariants": config["packet_invariants"].get(role, config["packet_invariants"]["combined"]),
            "context": {
                "target_tokens": target_tokens,
                "estimated_tokens": 0,
                "input_token_upper_bound": 0,
                "response_reserve_tokens": response_reserve,
                "context_window_tokens": context_window,
                "total_context_upper_bound": 0,
                "rendered_prompt_bytes": 0,
                "packet_envelope_bytes": 0,
                "slice_payload_bytes": 0,
                "slice_separator_bytes": 0,
                "fixed_prompt_bytes": fixed_prompt_bytes,
                "estimation": "complete_rendered_prompt_one_token_per_byte",
                "bytes_per_token_upper_bound": 1,
                "edge_context_mode": "exact revision/blob-bound line slices around assigned edge endpoints",
                "edge_context_max_bytes_per_file": maximum_edge_bytes,
                "overflow": False,
                "truncated": False,
            },
        }
        account_packet(packet)
        if packet["context"]["overflow"]:
            problems.append(
                finding(
                    "packet_overflow",
                    f"{role} rendered input {packet['context']['rendered_prompt_bytes']} exceeds input target {target_tokens} or total {packet['context']['total_context_upper_bound']} exceeds context window {context_window}",
                )
            )
        packets.append(packet)

    def slice_group_upper(paths: Iterable[str], source: dict[str, list[dict[str, Any]]]) -> int:
        assigned = [item for path in sorted(set(paths)) for item in source.get(path, [])]
        return (
            fixed_prompt_bytes + 1800 + sum(int(item["bytes"]) for item in assigned)
            + len(canonical_json_bytes(assigned))
        )

    def greedy_file_groups(values: list[str], maximum_files: int) -> list[list[str]]:
        groups: list[list[str]] = []
        current: list[str] = []
        for value in sorted(values):
            proposed = [*current, value]
            if current and (len(proposed) > maximum_files or slice_group_upper(proposed, blob_slices) > target_tokens):
                groups.append(current)
                current = [value]
            else:
                current = proposed
            if slice_group_upper(current, blob_slices) > target_tokens:
                problems.append(finding("packet_overflow", f"single context file cannot fit packet target: {value}"))
        if current:
            groups.append(current)
        return groups

    if combined:
        append_packet("combined", files, [], [])
        if packets[-1]["context"]["overflow"]:
            packets.clear()
            problems[:] = [
                item for item in problems
                if not item.get("claim", "").startswith("combined rendered input ")
            ]
            combined = False
            selection = "split"
    if not combined:
        by_role: dict[str, list[str]] = {}
        for path in files:
            by_role.setdefault(domains[path], []).append(path)
        for role in ("architecture_reference", "authority_workflow_state", "implementation_test"):
            current_paths: list[str] = []
            current_slices: list[dict[str, Any]] = []
            current_bytes = 0
            def flush_changed() -> None:
                nonlocal current_paths, current_slices, current_bytes
                if current_slices:
                    append_packet(role, sorted(set(current_paths)), [], [], changed_file_slices_override=current_slices)
                current_paths, current_slices, current_bytes = [], [], 0
            for path in sorted(by_role.get(role, [])):
                for item in changed_content_slices.get(path, []):
                    proposed_slices = [*current_slices, item]
                    proposed_upper = (
                        fixed_prompt_bytes + 1800
                        + sum(int(value["bytes"]) for value in proposed_slices)
                        + len(canonical_json_bytes(proposed_slices))
                    )
                    if current_slices and (
                        len(set(current_paths) | {path}) > int(thresholds["packet_max_changed_files"])
                        or proposed_upper > target_tokens
                    ):
                        flush_changed()
                    current_paths.append(path)
                    current_slices.append(item)
                    current_bytes += int(item["bytes"])
                    if (
                        fixed_prompt_bytes + 1800 + int(item["bytes"])
                        + len(canonical_json_bytes(item)) > target_tokens
                    ):
                        problems.append(finding("packet_overflow", f"atomic changed slice cannot fit: {path}:{item['start_line']}-{item['end_line']}"))
            flush_changed()

    def co_pack_context(
        paths: list[str], role: str, assignment_field: str,
    ) -> list[tuple[str, dict[str, Any]]]:
        remaining: list[tuple[str, dict[str, Any]]] = []
        maximum_files = int(thresholds["packet_max_context_files"])
        for path in sorted(paths):
            for context_slice in context_blob_slices.get(path, []):
                best: tuple[int, str, dict[str, Any], dict[str, Any]] | None = None
                for packet in sorted(packets, key=lambda item: item["packet_id"]):
                    assigned = set(packet["closure_files"]) | set(packet["authority_files"])
                    if path not in assigned and len(assigned) >= maximum_files:
                        continue
                    candidate = json.loads(json.dumps(packet))
                    if candidate["role"] != role:
                        candidate["role"] = "combined"
                        candidate["packet_id"] = "combined-00"
                        candidate["invariants"] = sorted(
                            set(candidate["invariants"])
                            | set(config["packet_invariants"][role])
                        )
                    candidate[assignment_field] = sorted(set(candidate[assignment_field]) | {path})
                    candidate["context_file_slices"] = sorted(
                        [*candidate["context_file_slices"], context_slice],
                        key=lambda item: (item["path"], item["revision"], item["start_byte"]),
                    )
                    account_packet(candidate)
                    if candidate["context"]["overflow"]:
                        continue
                    score = (candidate["context"]["rendered_prompt_bytes"], packet["packet_id"])
                    if best is None or score > best[:2]:
                        best = (*score, packet, candidate)
                if best is None:
                    remaining.append((path, context_slice))
                else:
                    best[2].update(best[3])
        return remaining

    def append_context_remainders(
        remaining: list[tuple[str, dict[str, Any]]], role: str, assignment_field: str,
    ) -> None:
        current_paths: set[str] = set()
        current_slices: list[dict[str, Any]] = []

        def flush() -> None:
            nonlocal current_paths, current_slices
            if not current_slices:
                return
            closure = sorted(current_paths) if assignment_field == "closure_files" else []
            authority = sorted(current_paths) if assignment_field == "authority_files" else []
            append_packet(
                role, [], closure, authority,
                context_file_slices_override=current_slices,
            )
            current_paths, current_slices = set(), []

        for path, context_slice in remaining:
            proposed_paths = current_paths | {path}
            proposed_slices = [*current_slices, context_slice]
            proposed_upper = (
                fixed_prompt_bytes + 1800
                + sum(int(item["bytes"]) for item in proposed_slices)
                + len(canonical_json_bytes(proposed_slices))
            )
            if current_slices and (
                len(proposed_paths) > int(thresholds["packet_max_context_files"])
                or proposed_upper > target_tokens
            ):
                flush()
            current_paths.add(path)
            current_slices.append(context_slice)
            if (
                fixed_prompt_bytes + 1800 + int(context_slice["bytes"])
                + len(canonical_json_bytes(context_slice)) > target_tokens
            ):
                problems.append(
                    finding(
                        "packet_overflow",
                        f"atomic context slice cannot fit: {path}:{context_slice['start_line']}-{context_slice['end_line']}",
                    )
                )
        flush()

    remaining_closure = co_pack_context(
        closure_files, "architecture_reference", "closure_files",
    )
    remaining_authority = co_pack_context(
        authority_files, "authority_workflow_state", "authority_files",
    )
    append_context_remainders(
        remaining_closure, "architecture_reference", "closure_files",
    )
    append_context_remainders(
        remaining_authority, "authority_workflow_state", "authority_files",
    )

    edge_groups: list[list[dict[str, Any]]] = []
    current_edges: list[dict[str, Any]] = []
    for edge in sorted(
        impact_edges,
        key=lambda item: (item["source"], item["target"], item["relation"], item["id"]),
    ):
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
        proposed_edge_slices = exact_edge_slices(proposed_ids)
        proposed_estimate = (
            fixed_prompt_bytes + 5000
            + len(canonical_json_bytes([impact_edge_by_id[edge_id] for edge_id in proposed_ids]))
            + len(canonical_json_bytes(proposed_edge_slices))
            + sum(int(item["bytes"]) for item in proposed_edge_slices)
        )
        proposed_by_path: dict[str, int] = defaultdict(int)
        for item in proposed_edge_slices:
            proposed_by_path[item["path"]] += int(item["bytes"])
        edge_slice_bound_hit = any(
            used > int(graph_limits.get("edge_context_max_bytes_per_file", 8192))
            for used in proposed_by_path.values()
        )
        if current_edges and (
            len(proposed) > int(graph_limits["packet_max_impact_edges"])
            or proposed_estimate > target_tokens
            or edge_slice_bound_hit
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
        single_slices = exact_edge_slices(single_ids)
        single_estimate = (
            fixed_prompt_bytes + 5000
            + len(canonical_json_bytes([impact_edge_by_id[edge_id] for edge_id in single_ids]))
            + len(canonical_json_bytes(single_slices))
            + sum(int(item["bytes"]) for item in single_slices)
        )
        single_by_path: dict[str, int] = defaultdict(int)
        for item in single_slices:
            single_by_path[item["path"]] += int(item["bytes"])
        if single_estimate > target_tokens or any(
            used > int(graph_limits.get("edge_context_max_bytes_per_file", 8192))
            for used in single_by_path.values()
        ):
            problems.append(finding("packet_overflow", f"atomic impact edge neighborhood cannot fit: {edge['id']}"))
    if current_edges:
        edge_groups.append(current_edges)

    edge_impact_assigned: set[str] = set()
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
        ordered_endpoints = [path for path in endpoints if path not in edge_impact_assigned]
        ordered_endpoints.extend(path for path in endpoints if path in edge_impact_assigned)
        covered_endpoints = ordered_endpoints[: int(graph_limits["packet_max_impact_files"])]
        edge_impact_assigned.update(covered_endpoints)
        # Edge-context slices already bind the exact content for these impact files. Do not copy
        # identical slice metadata into impact_file_slices; response coverage echoes both the file
        # assignment and the single canonical edge-context slice assignment.
        append_packet(
            "impact_graph", [], [], [], covered_endpoints, edge_ids, endpoints
        )

    # Practical impact coverage is edge-focused, not a claim that every byte of every transitive
    # file receives a second full-file review. Each impact file receives at least one exact
    # revision/blob-bound slice at a traversed edge. Isolated seeds receive one bounded anchor slice.
    assigned_impact_files = {
        path for packet in packets for path in packet.get("impact_files", [])
    }
    remaining_files = sorted(set(impact_files) - assigned_impact_files)
    current_slices: list[dict[str, Any]] = []
    current_paths: set[str] = set()
    payload_limit = max(1, target_tokens - fixed_prompt_bytes - 4096)
    for path in remaining_files:
        candidates = edge_blob_slices.get(path, [])
        if not candidates:
            problems.append(finding("packet_overflow", f"impact file lacks exact anchor slice: {path}"))
            continue
        item = next((value for value in candidates if value["revision_kind"] == "head"), candidates[0])
        # Fill spare impact-file slots in edge packets first. This preserves one coherent
        # neighborhood and avoids duplicating fixed prompt/envelope overhead for isolated anchors.
        co_packed = False
        for packet in sorted(
            packets,
            key=lambda value: (-value["context"]["rendered_prompt_bytes"], value["packet_id"]),
        ):
            if len(packet["impact_files"]) >= int(graph_limits["packet_max_impact_files"]):
                continue
            candidate = json.loads(json.dumps(packet))
            if candidate["role"] != "impact_graph":
                candidate["role"] = "combined"
                candidate["packet_id"] = "combined-00"
                candidate["invariants"] = sorted(
                    set(candidate["invariants"])
                    | set(config["packet_invariants"]["impact_graph"])
                )
            candidate["impact_files"] = sorted([*candidate["impact_files"], path])
            candidate["impact_file_slices"] = sorted(
                [*candidate["impact_file_slices"], item],
                key=lambda value: (value["path"], value["revision"], value["start_byte"]),
            )
            account_packet(candidate)
            if candidate["context"]["overflow"]:
                continue
            packet.update(candidate)
            co_packed = True
            break
        if co_packed:
            continue
        if current_slices and (
            len(current_paths | {path}) > int(graph_limits["packet_max_impact_files"])
            or sum(int(value["bytes"]) for value in current_slices) + int(item["bytes"]) > payload_limit
        ):
            append_packet("impact_graph", [], [], [], sorted(current_paths), [], [], current_slices)
            current_slices = []
            current_paths = set()
        current_paths.add(path)
        current_slices.append(item)
        if int(item["bytes"]) > payload_limit:
            problems.append(finding("packet_overflow", f"atomic impact anchor slice cannot fit: {path}"))
    if current_slices:
        append_packet("impact_graph", [], [], [], sorted(current_paths), [], [], current_slices)

    # Re-account after impact slices are packed into existing edge packets. The postcondition uses
    # the complete serialized packet envelope and every exact payload slice; no average tokenizer
    # ratio or pre-mutation estimate can authorize a packet.
    for packet in packets:
        account_packet(packet)

    maximum_packets = int(graph_limits.get("max_packets", 64))

    def unique_packet_slices(values: Iterable[dict[str, Any]]) -> list[dict[str, Any]]:
        unique = {
            (item["path"], item["revision"], item["start_byte"], item["end_byte"], item["sha256"]): item
            for item in values
        }
        return sorted(
            unique.values(),
            key=lambda item: (item["path"], item["revision"], item["start_byte"], item["end_byte"], item["sha256"]),
        )

    def merge_packet_pair(left: dict[str, Any], right: dict[str, Any]) -> dict[str, Any]:
        candidate = json.loads(json.dumps(left))
        candidate["role"] = left["role"] if left["role"] == right["role"] else "combined"
        candidate["packet_id"] = f"{candidate['role']}-00"
        for field in (
            "changed_files", "closure_files", "authority_files", "impact_files",
            "impact_edge_ids", "edge_context_files", "invariants",
        ):
            candidate[field] = sorted(set(left[field]) | set(right[field]))
        candidate["impact_edges"] = [
            impact_edge_by_id[edge_id] for edge_id in candidate["impact_edge_ids"]
        ]
        for field in (
            "changed_file_slices", "context_file_slices", "impact_file_slices",
            "edge_context_slices",
        ):
            candidate[field] = unique_packet_slices([*left[field], *right[field]])
        account_packet(candidate)
        return candidate

    def merge_preserves_bounds(candidate: dict[str, Any]) -> bool:
        if len(candidate["changed_files"]) > int(thresholds["packet_max_changed_files"]):
            return False
        if len(candidate["closure_files"]) + len(candidate["authority_files"]) > int(thresholds["packet_max_context_files"]):
            return False
        if len(candidate["impact_files"]) > int(graph_limits["packet_max_impact_files"]):
            return False
        if len(candidate["impact_edge_ids"]) > int(graph_limits["packet_max_impact_edges"]):
            return False
        edge_bytes: dict[str, int] = defaultdict(int)
        for item in candidate["edge_context_slices"]:
            edge_bytes[item["path"]] += int(item["bytes"])
        if any(used > int(graph_limits["edge_context_max_bytes_per_file"]) for used in edge_bytes.values()):
            return False
        return not candidate["context"]["overflow"]

    # A large exact range can contain a few complementary under-filled packets. Coalesce only
    # assignment-preserving compatible pairs near the hard packet cap, retaining two packets of
    # headroom for final evidence-only range growth. The scarcity-first order avoids consuming a
    # packet that is the sole fit for another packet; every accepted pair is re-accounted from the
    # complete merged envelope and exact payload bytes.
    coalescing_target = max(1, maximum_packets - 2)
    while coalescing_target < len(packets) <= maximum_packets * 2:
        compatible: list[tuple[int, int, dict[str, Any]]] = []
        degrees = [0] * len(packets)
        for left_index, left in enumerate(packets):
            for right_index in range(left_index + 1, len(packets)):
                candidate = merge_packet_pair(left, packets[right_index])
                if not merge_preserves_bounds(candidate):
                    continue
                compatible.append((left_index, right_index, candidate))
                degrees[left_index] += 1
                degrees[right_index] += 1
        selected: list[tuple[int, int, dict[str, Any]]] = []
        consumed: set[int] = set()
        for left_index, right_index, candidate in sorted(
            compatible,
            key=lambda item: (degrees[item[0]] + degrees[item[1]], item[0], item[1]),
        ):
            if len(packets) - len(selected) <= coalescing_target:
                break
            if left_index in consumed or right_index in consumed:
                continue
            selected.append((left_index, right_index, candidate))
            consumed.update((left_index, right_index))
        if not selected:
            break
        merged_by_left = {left: (right, candidate) for left, right, candidate in selected}
        right_indexes = {right for _, right, _ in selected}
        coalesced: list[dict[str, Any]] = []
        for index, packet in enumerate(packets):
            if index in right_indexes:
                continue
            if index in merged_by_left:
                coalesced.append(merged_by_left[index][1])
            else:
                coalesced.append(packet)
        packets = coalesced

    role_counts: dict[str, int] = defaultdict(int)
    for packet in packets:
        role_counts[packet["role"]] += 1
        packet["packet_id"] = f"{packet['role']}-{role_counts[packet['role']]:02d}"
        account_packet(packet)
        if packet["context"]["overflow"]:
            problems.append(
                finding("packet_overflow", f"{packet['packet_id']} violates the final rendered-input postcondition")
            )

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
        input_path = Path(args.input).resolve(strict=True)
        input_bytes = input_path.read_bytes()
        document = require_object(json.loads(input_bytes), "observation corpus")
        cases = document.get("cases")
        if not isinstance(cases, list) or not all(isinstance(case, dict) for case in cases):
            raise ReviewSystemError("observation corpus cases must be an object list")
        case_ids = [case.get("case_id") for case in cases]
        if not all(isinstance(case_id, str) and case_id for case_id in case_ids) or len(case_ids) != len(set(case_ids)):
            raise ReviewSystemError("observation corpus case ids must be non-empty and unique")
        started = time.perf_counter_ns()
        observations = [detect(case, args.mode) for case in cases]
        emit_json(
            {
                "schema_version": "polymetrics.ai/pm-review-observations/v1",
                "mode": args.mode,
                "input": str(input_path),
                "input_sha256": hashlib.sha256(input_bytes).hexdigest(),
                "case_set_sha256": hashlib.sha256("\n".join(sorted(case_ids)).encode()).hexdigest(),
                "observations": observations,
                "wall_clock_ms": round((time.perf_counter_ns() - started) / 1_000_000, 6),
            }
        )
        return 0
    except (ReviewSystemError, OSError, json.JSONDecodeError) as exc:
        print(f"pm review observe error: {exc}", file=sys.stderr)
        return 2


def metric_ratio(numerator: int, denominator: int) -> float | None:
    return round(numerator / denominator, 6) if denominator else None


def command_score(args: argparse.Namespace) -> int:
    try:
        observed = require_object(load_json(Path(args.observations)), "observations")
        oracle_document = require_object(load_json(Path(args.oracle)), "oracle")
        oracle = oracle_document.get("cases", {})
        rows = observed.get("observations", [])
        if not isinstance(oracle, dict) or not oracle:
            raise ReviewSystemError("oracle cases must be a non-empty object")
        if not isinstance(rows, list) or not all(isinstance(row, dict) for row in rows):
            raise ReviewSystemError("observations must be an object list")
        observed_ids = [row.get("case_id") for row in rows]
        if not all(isinstance(case_id, str) and case_id for case_id in observed_ids):
            raise ReviewSystemError("observations contain a malformed case id")
        if len(observed_ids) != len(set(observed_ids)):
            raise ReviewSystemError("observations contain duplicate case ids")
        if set(observed_ids) != set(oracle):
            raise ReviewSystemError(
                f"observation/oracle case bijection differs: missing={sorted(set(oracle)-set(observed_ids))} extra={sorted(set(observed_ids)-set(oracle))}"
            )
        input_value = observed.get("input")
        input_digest = observed.get("input_sha256")
        case_set_digest = observed.get("case_set_sha256")
        if not isinstance(input_value, str) or not isinstance(input_digest, str) or not HEX_DIGEST.fullmatch(input_digest):
            raise ReviewSystemError("observations lack exact input-corpus binding")
        try:
            current_input = Path(input_value).resolve(strict=True).read_bytes()
        except OSError as exc:
            raise ReviewSystemError(f"bound observation corpus is unavailable: {exc}") from exc
        if hashlib.sha256(current_input).hexdigest() != input_digest:
            raise ReviewSystemError("bound observation corpus hash differs")
        try:
            input_document = require_object(json.loads(current_input), "bound observation corpus")
            input_case_ids = [item.get("case_id") for item in input_document.get("cases", []) if isinstance(item, dict)]
        except json.JSONDecodeError as exc:
            raise ReviewSystemError(f"bound observation corpus is malformed: {exc}") from exc
        expected_case_set = hashlib.sha256("\n".join(sorted(observed_ids)).encode()).hexdigest()
        if input_case_ids != observed_ids or case_set_digest != expected_case_set:
            raise ReviewSystemError("observation order/case set is not exactly bound to the input corpus")
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


def compile_document(root: Path, config_relative: str, scope_relative: str, base: str, head: str) -> dict[str, Any]:
    root = root.resolve(strict=True)
    base = validate_sha(base, "exact base")
    head = validate_sha(head, "exact head")
    config_relative = validate_relative_path(config_relative, "review config path")
    scope_relative = validate_relative_path(scope_relative, "review scope path")
    run_git(root, ["cat-file", "-e", f"{base}^{{commit}}"])
    run_git(root, ["cat-file", "-e", f"{head}^{{commit}}"])
    head_tree = run_git(root, ["rev-parse", f"{head}^{{tree}}"]).strip()
    merge_base = run_git(root, ["merge-base", base, head]).strip()
    if merge_base != base:
        raise ReviewSystemError(f"exact base is not the candidate merge base: {merge_base} != {base}")

    snapshot_tmp = clone_commit_snapshot(root, head)
    with snapshot_tmp:
        snapshot = Path(snapshot_tmp.name) / "repo"
        config_bytes = git_blob(snapshot, head, config_relative, maximum=2 * 1024 * 1024)
        scope_bytes = git_blob(snapshot, head, scope_relative, maximum=512 * 1024)
        try:
            config = validate_config(json.loads(config_bytes))
            scope = require_object(json.loads(scope_bytes), "review scope")
        except json.JSONDecodeError as exc:
            raise ReviewSystemError(f"exact config/scope JSON is malformed: {exc}") from exc
        if scope.get("schema_version") != SCOPE_SCHEMA:
            raise ReviewSystemError(
                f"review scope migration required: {scope.get('schema_version')!r} != {SCOPE_SCHEMA!r}"
            )
        if not isinstance(scope.get("review_round"), int) or isinstance(scope.get("review_round"), bool) or scope["review_round"] <= 0:
            raise ReviewSystemError("review scope review_round must be a positive integer")
        for field in ("allowed_changed_paths", "forbidden_changed_paths"):
            values = scope.get(field)
            if not isinstance(values, list) or not all(isinstance(item, str) and item for item in values):
                raise ReviewSystemError(f"review scope {field} must be a non-empty string list")
            for pattern in values:
                if CONTROL.search(pattern) or pattern.startswith(("-", "/")) or "\\" in pattern or ".." in PurePosixPath(pattern).parts:
                    raise ReviewSystemError(f"review scope {field} contains an unsafe pattern")

        # The fixed reviewer/workflow/template bytes are part of every rendered input. They are
        # loaded from the exact head and charged once per packet under one token per byte.
        prompt_contract: list[dict[str, Any]] = []
        fixed_prompt_bytes = len(PROMPT_HEADER) + len(PACKET_MARKER) + len(PAYLOAD_MARKER)
        for relative in config.get("prompt_contract_files", []):
            relative = validate_relative_path(relative, "prompt contract file")
            payload = git_blob(snapshot, head, relative, maximum=2 * 1024 * 1024)
            prompt_contract.append({"path": relative, "bytes": len(payload), "sha256": hashlib.sha256(payload).hexdigest()})
            fixed_prompt_bytes += len(payload)
        config = json.loads(json.dumps(config))
        config["thresholds"]["rendered_prompt_fixed_bytes"] = fixed_prompt_bytes

        files, line_counts = changed_files(snapshot, base, head)
        findings: list[dict[str, Any]] = []
        for path in files:
            if path_matches(path, scope.get("forbidden_changed_paths", [])):
                findings.append(finding("changed_path_scope", "changed path is forbidden", path))
            elif not path_matches(path, scope.get("allowed_changed_paths", [])):
                findings.append(finding("changed_path_scope", "changed path is outside the positive allowlist", path))
        closure, edges, closure_findings = compile_closure(snapshot, config)
        findings.extend(closure_findings)
        authority, authority_files, authority_findings = authority_inventory(snapshot, config)
        findings.extend(authority_findings)
        impact_graph, impact_findings = compile_impact_graph(snapshot, base, head, files, config)
        findings.extend(impact_findings)
        domains = {path: classify_domain(path, config) for path in files}
        closure_context = sorted(set(closure) - set(files))
        authority_context = sorted(set(authority_files) - set(files))
        packet_context_paths = set(files) | set(closure_context) | set(authority_context) | set(impact_graph["files"])
        for edge in impact_graph["edges"]:
            packet_context_paths.update(
                endpoint for endpoint in (edge["source"], edge["target"]) if not endpoint.startswith("go-package:")
            )
        maximum_slice_bytes = int(config["impact_graph"]["packet_max_file_slice_bytes"])
        blob_sizes, blob_slices = exact_blob_metadata(
            snapshot, base, head, packet_context_paths, maximum_slice_bytes,
        )
        diff_sizes, diff_slices = exact_diff_metadata(
            snapshot, base, head, files, maximum_slice_bytes,
        )
        _, edge_slices = exact_blob_metadata(
            snapshot, base, head, packet_context_paths, min(256, maximum_slice_bytes),
        )
        _, context_slices = exact_blob_metadata(
            snapshot, base, head, set(closure_context) | set(authority_context),
            min(4096, maximum_slice_bytes),
        )
        selection, packets, packet_findings = build_packets(
            base, head, head_tree, files, line_counts, domains, closure_context, authority_context,
            impact_graph["files"], impact_graph["edges"], blob_sizes, blob_slices, config,
            diff_sizes, diff_slices, edge_slices, context_slices,
        )
        findings.extend(packet_findings)
        status = "blocked" if findings or selection == "blocked" else "ready"
        coverage = {
            "changed_files": files,
            "closure_files": closure_context,
            "authority_files": authority_context,
            "impact_files": impact_graph["files"],
            "impact_edge_ids": [edge["id"] for edge in impact_graph["edges"]],
            "edge_context_files": sorted({path for packet in packets for path in packet.get("edge_context_files", [])}),
            "packet_ids": [packet["packet_id"] for packet in packets],
        }
        document: dict[str, Any] = {
            "schema_version": COMPILE_SCHEMA,
            "status": status,
            "owner": config.get("owner"),
            "config": {
                "schema_version": CONFIG_SCHEMA,
                "path": config_relative,
                "sha256": hashlib.sha256(config_bytes).hexdigest(),
            },
            "scope": {
                "schema_version": SCOPE_SCHEMA,
                "path": scope_relative,
                "sha256": hashlib.sha256(scope_bytes).hexdigest(),
                "issue": scope.get("issue"),
                "candidate_lineage": scope.get("candidate_lineage"),
                "review_round": scope.get("review_round"),
            },
            "exact_base_sha": base,
            "exact_head_sha": head,
            "exact_head_tree": head_tree,
            "source_mode": "detached_exact_commit_snapshot",
            "prompt_contract": prompt_contract,
            "changed_files": files,
            "changed_lines": sum(line_counts.values()),
            "domains": domains,
            "reference_closure": {"files": closure, "edges": edges},
            "authority_inventory": authority,
            "impact_graph": impact_graph,
            "selection": selection,
            "packets": packets,
            "findings": findings,
            "coverage_manifest": coverage,
            "content_policy": "paths, exact revision/blob/slice metadata, and deterministic accounting only; no file contents or environment values",
        }
        document["authentication"] = {
            "algorithm": "sha256-canonical-json-v1",
            "coverage_sha256": sha256_json(coverage),
            "packets_sha256": sha256_json(packets),
            "semantic_manifest_sha256": sha256_json(document),
        }
        return document


def load_trust_bundle(root: Path, path_value: str, document: dict[str, Any]) -> dict[str, Any]:
    if not isinstance(path_value, str) or not path_value:
        raise ReviewSystemError("an external parent-approved trust bundle is required")
    path = Path(path_value).expanduser()
    if path.is_symlink():
        raise ReviewSystemError("trust bundle must not be a symlink")
    try:
        resolved = path.resolve(strict=True)
    except OSError as exc:
        raise ReviewSystemError(f"trust bundle is unavailable: {exc}") from exc
    root_resolved = root.resolve(strict=True)
    if Path(os.path.commonpath((str(root_resolved), str(resolved)))) == root_resolved:
        raise ReviewSystemError("trust bundle must be parent-owned outside the candidate repository")
    with resolved.open("rb") as handle:
        raw = handle.read(64 * 1024 + 1)
    if len(raw) > 64 * 1024:
        raise ReviewSystemError("trust bundle exceeds 64 KiB")
    try:
        bundle = require_object(json.loads(raw), "trust bundle")
    except json.JSONDecodeError as exc:
        raise ReviewSystemError(f"trust bundle JSON is malformed: {exc}") from exc
    required = {
        "schema_version", "exact_base_sha", "exact_head_sha", "exact_head_tree",
        "candidate_lineage", "review_round", "compiler_sha256", "config_sha256",
        "scope_sha256", "prompt_contract_sha256", "approval_reference",
    }
    if set(bundle) != required or bundle.get("schema_version") != TRUST_BUNDLE_SCHEMA:
        raise ReviewSystemError("trust bundle shape/schema is not exact v1")
    compiler_path = Path(__file__).resolve(strict=True)
    compiler_digest = hashlib.sha256(compiler_path.read_bytes()).hexdigest()
    expected = {
        "exact_base_sha": document.get("exact_base_sha"),
        "exact_head_sha": document.get("exact_head_sha"),
        "exact_head_tree": document.get("exact_head_tree"),
        "candidate_lineage": document.get("scope", {}).get("candidate_lineage"),
        "review_round": document.get("scope", {}).get("review_round"),
        "compiler_sha256": compiler_digest,
        "config_sha256": document.get("config", {}).get("sha256"),
        "scope_sha256": document.get("scope", {}).get("sha256"),
        "prompt_contract_sha256": sha256_json(document.get("prompt_contract", [])),
    }
    for field, value in expected.items():
        if bundle.get(field) != value:
            raise ReviewSystemError(f"trust bundle {field} differs from the parent-approved candidate binding")
    if not isinstance(bundle.get("approval_reference"), str) or not bundle["approval_reference"].strip():
        raise ReviewSystemError("trust bundle approval_reference is empty")
    return {
        **bundle,
        "bundle_sha256": hashlib.sha256(raw).hexdigest(),
        "source": "external_parent_owned_exact_binding",
    }


def bind_trust(root: Path, document: dict[str, Any], trust_bundle_path: str) -> dict[str, Any]:
    unsigned = {key: value for key, value in document.items() if key != "authentication"}
    unsigned["trust_root"] = load_trust_bundle(root, trust_bundle_path, document)
    unsigned["authentication"] = {
        "algorithm": "sha256-canonical-json-v1",
        "coverage_sha256": sha256_json(unsigned.get("coverage_manifest")),
        "packets_sha256": sha256_json(unsigned.get("packets")),
        "semantic_manifest_sha256": sha256_json(unsigned),
    }
    return unsigned


def command_compile(args: argparse.Namespace) -> int:
    try:
        root = Path(args.repo_root).resolve(strict=True)
        expected_head = validate_sha(args.head, "exact head")
        before = (
            run_git(root, ["rev-parse", "HEAD"]).strip(),
            run_git(root, ["rev-parse", "HEAD^{tree}"]).strip(),
            run_git(root, ["--no-optional-locks", "status", "--porcelain", "--untracked-files=all"]).strip(),
        )
        if before[0] != expected_head or before[2]:
            raise ReviewSystemError("candidate must be the clean exact reviewed HEAD before compilation")
        document = compile_document(root, args.config, args.scope, args.base, expected_head)
        document = bind_trust(root, document, args.trust_bundle)
        after = (
            run_git(root, ["rev-parse", "HEAD"]).strip(),
            run_git(root, ["rev-parse", "HEAD^{tree}"]).strip(),
            run_git(root, ["--no-optional-locks", "status", "--porcelain", "--untracked-files=all"]).strip(),
        )
        if after != before or after[1] != document.get("exact_head_tree"):
            raise ReviewSystemError("candidate identity or cleanliness drifted during compilation")
        emit_json(document)
        return 1 if document["status"] == "blocked" else 0
    except (ReviewSystemError, OSError, TypeError, KeyError) as exc:
        emit_json({"schema_version": COMPILE_SCHEMA, "status": "blocked", "findings": [finding("compile_input", str(exc))]})
        return 2


def command_render(args: argparse.Namespace) -> int:
    try:
        root = Path(args.repo_root).resolve(strict=True)
        manifest = require_object(load_json(resolve_safe(root, args.manifest)), "compile manifest")
        if manifest.get("schema_version") != COMPILE_SCHEMA:
            raise ReviewSystemError(
                f"compile manifest migration required: {manifest.get('schema_version')!r} != {COMPILE_SCHEMA!r}"
            )
        blockers = verify_manifest_candidate(root, manifest)
        if blockers:
            raise ReviewSystemError(f"compile manifest is not renderable: {blockers[0]['claim']}")
        rebuilt = compile_document(
            root,
            require_object(manifest.get("config"), "manifest config").get("path", ""),
            require_object(manifest.get("scope"), "manifest scope").get("path", ""),
            manifest["exact_base_sha"],
            manifest["exact_head_sha"],
        )
        rebuilt = bind_trust(root, rebuilt, args.trust_bundle)
        if sha256_json(rebuilt) != sha256_json(manifest):
            raise ReviewSystemError("compile manifest differs from deterministic exact-commit reconstruction")
        packet_id = args.packet_id
        if not isinstance(packet_id, str) or not SAFE_ID.fullmatch(packet_id):
            raise ReviewSystemError("packet id is malformed")
        packet = next(
            (item for item in manifest.get("packets", []) if item.get("packet_id") == packet_id),
            None,
        )
        if packet is None:
            raise ReviewSystemError(f"packet id is not assigned: {packet_id}")
        for field in (
            "changed_file_slices", "context_file_slices", "impact_file_slices", "edge_context_slices"
        ):
            for item in packet.get(field, []):
                if item.get("revision") not in {manifest["exact_base_sha"], manifest["exact_head_sha"]}:
                    raise ReviewSystemError(f"packet slice revision is outside exact base/head: {item.get('path')}")
                if item.get("revision_kind") == "diff" and item.get("revision") != manifest["exact_head_sha"]:
                    raise ReviewSystemError(f"packet diff slice is not bound to exact head: {item.get('path')}")
        prompt = bytearray(PROMPT_HEADER)
        for record in manifest.get("prompt_contract", []):
            if not isinstance(record, dict):
                raise ReviewSystemError("prompt contract binding is malformed")
            payload = git_blob(root, manifest["exact_head_sha"], record.get("path", ""), maximum=2 * 1024 * 1024)
            if len(payload) != record.get("bytes") or hashlib.sha256(payload).hexdigest() != record.get("sha256"):
                raise ReviewSystemError("prompt contract content differs from compile binding")
            prompt.extend(payload)
        prompt.extend(PACKET_MARKER)
        prompt.extend(canonical_json_bytes(packet))
        prompt.extend(PAYLOAD_MARKER)
        all_slices = (
            packet.get("changed_file_slices", []) + packet.get("context_file_slices", [])
            + packet.get("impact_file_slices", []) + packet.get("edge_context_slices", [])
        )
        unique_slices = {
            (item["path"], item["revision"], item["start_byte"], item["end_byte"], item["sha256"]): item
            for item in all_slices
        }
        blob_cache: dict[tuple[str, str, str], bytes] = {}
        for item in sorted(
            unique_slices.values(),
            key=lambda value: (value["path"], value["revision"], value["start_byte"], value["end_byte"], value["sha256"]),
        ):
            cache_key = (item["revision_kind"], item["revision"], item["path"])
            if cache_key not in blob_cache:
                if item["revision_kind"] == "diff":
                    blob_cache[cache_key] = run_git_bytes_bounded(
                        root,
                        [
                            "diff", "--no-ext-diff", "--no-renames", "--binary", "--full-index",
                            "--unified=3", f"{manifest['exact_base_sha']}...{manifest['exact_head_sha']}",
                            "--", item["path"],
                        ],
                        MAX_EXACT_DIFF_BYTES,
                    )
                else:
                    blob_cache[cache_key] = git_blob(root, item["revision"], item["path"])
            blob = blob_cache[cache_key]
            if hashlib.sha256(blob).hexdigest() != item["blob_sha256"]:
                raise ReviewSystemError(f"slice blob binding differs: {item['path']}")
            payload = blob[int(item["start_byte"]):int(item["end_byte"])]
            if len(payload) != item["bytes"] or hashlib.sha256(payload).hexdigest() != item["sha256"]:
                raise ReviewSystemError(f"slice content binding differs: {item['path']}")
            prompt.extend(SLICE_SEPARATOR)
            prompt.extend(payload)
        expected = packet.get("context", {}).get("rendered_prompt_bytes", 0)
        if len(prompt) != expected:
            raise ReviewSystemError(f"rendered prompt accounting differs: {len(prompt)} != {expected}")
        sys.stdout.buffer.write(prompt)
        return 0
    except (ReviewSystemError, OSError, TypeError, KeyError) as exc:
        print(f"pm review render error: {exc}", file=sys.stderr)
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
    if not isinstance(hypotheses, list) or not hypotheses:
        blockers.append(finding("hypothesis_evidence", f"{packet_id} lacks structured falsifiable hypotheses"))
    else:
        seen_hypotheses: set[str] = set()
        for hypothesis in hypotheses:
            if not isinstance(hypothesis, dict) or set(hypothesis) != HYPOTHESIS_FIELDS:
                blockers.append(finding("hypothesis_evidence", f"{packet_id} hypothesis shape is not exact v4"))
                continue
            identifier = hypothesis.get("id")
            if not isinstance(identifier, str) or not SAFE_ID.fullmatch(identifier) or identifier in seen_hypotheses:
                blockers.append(finding("hypothesis_evidence", f"{packet_id} hypothesis id is malformed or duplicate"))
            else:
                seen_hypotheses.add(identifier)
            for field in ("claim", "strongest_alternative", "falsifier"):
                if not isinstance(hypothesis.get(field), str) or not hypothesis[field].strip():
                    blockers.append(finding("hypothesis_evidence", f"{packet_id} hypothesis {field} is empty"))
            evidence_paths = hypothesis.get("evidence_paths")
            if not isinstance(evidence_paths, list) or not evidence_paths or not all(isinstance(path, str) and path for path in evidence_paths):
                blockers.append(finding("hypothesis_evidence", f"{packet_id} hypothesis evidence_paths is malformed"))
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
    behavior_hypotheses = response.get("review_behaviors", {}).get("hypotheses", []) if isinstance(response.get("review_behaviors"), dict) else []
    hypothesis_ids = {item.get("id") for item in behavior_hypotheses if isinstance(item, dict)}
    for experiment in experiments:
        if not isinstance(experiment, dict) or set(experiment) != required:
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment does not match the exact v4 shape"))
            continue
        for field in ("hypothesis_id", "claim", "alternative", "temporary_change", "lab_evidence_path", "lab_evidence_sha256"):
            if not isinstance(experiment.get(field), str) or not experiment[field].strip():
                blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment {field} is malformed"))
        if experiment.get("hypothesis_id") not in hypothesis_ids:
            blockers.append(finding("hypothesis_evidence", f"{packet_id} experiment cites an unassigned hypothesis"))
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
            evidence = require_object(json.loads(evidence_bytes), "lab evidence")
            if evidence.get("schema_version") != LAB_EVIDENCE_SCHEMA or evidence.get("status") != "evidence":
                raise ReviewSystemError("lab evidence schema/status is not clean evidence")
            for field in ("exact_base_sha", "exact_head_sha", "exact_head_tree"):
                if evidence.get(field) != manifest.get(field):
                    raise ReviewSystemError(f"lab evidence {field} is stale")
            if evidence.get("packet_id") != packet_id or evidence.get("candidate_unchanged") is not True or evidence.get("lab_cleanup_verified") is not True:
                raise ReviewSystemError("lab evidence packet, candidate, or cleanup proof differs")
            evidence_experiment = require_object(evidence.get("experiment"), "lab experiment")
            for field in EXPERIMENT_BINDING_FIELDS:
                if experiment.get(field) != evidence_experiment.get(field):
                    raise ReviewSystemError(f"response experiment relabels lab evidence field {field}")
            observed = require_object(evidence_experiment.get("observed"), "lab observed result")
            expected = require_object(evidence_experiment.get("expected_discriminator"), "lab discriminator")
            if any(observed.get(key) != value for key, value in expected.items()):
                raise ReviewSystemError("lab discriminator does not match observed evidence")
            if (
                observed.get("limit_hit") is not None
                or observed.get("processes_remaining") != 0
                or observed.get("process_residue_detected") not in {None, False}
                or observed.get("sandbox_denial_observed") is not False
            ):
                raise ReviewSystemError("lab limit, residue, or sandbox-denial proof is incomplete")
            final = evidence.get("final_state")
            if not isinstance(final, dict) or final.get("resource_bounds_satisfied") is not True or final.get("candidate_unchanged") is not True or final.get("cleanup_verified") is not True:
                raise ReviewSystemError("lab final aggregate/state proof is incomplete")
        except (ReviewSystemError, OSError, json.JSONDecodeError) as exc:
            blockers.append(finding("lab_safety", f"{packet_id} lab evidence is invalid: {exc}"))
    return blockers


def validate_response_shape(packet_id: str, response: Any) -> list[dict[str, Any]]:
    if not isinstance(response, dict):
        return [finding("packet_response", f"{packet_id} response must be an object")]
    required = {
        "schema_version", "packet_id", "exact_base_sha", "exact_head_sha", "exact_head_tree",
        "status", "reviewed_files", "changed_file_slices", "closure_files", "authority_files",
        "context_file_slices", "impact_files", "impact_edge_ids", "impact_file_slices",
        "edge_context_files", "edge_context_slices", "invariants", "unreviewed_files",
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
    for slice_field in ("changed_file_slices", "context_file_slices", "impact_file_slices", "edge_context_slices"):
        slices = response.get(slice_field)
        if not isinstance(slices, list) or not all(
            isinstance(item, dict)
            and set(item) == SLICE_FIELDS
            and isinstance(item.get("path"), str)
            and isinstance(item.get("revision"), str) and HEX_SHA.fullmatch(item["revision"])
            and item.get("revision_kind") in {"head", "base_changed", "base_deleted", "diff"}
            and isinstance(item.get("blob_sha256"), str) and HEX_DIGEST.fullmatch(item["blob_sha256"])
            and all(isinstance(item.get(field), int) and not isinstance(item.get(field), bool) for field in ("start_line", "end_line", "start_byte", "end_byte", "bytes"))
            and item["start_line"] >= 1 and item["end_line"] >= item["start_line"]
            and item["start_byte"] >= 0 and item["end_byte"] - item["start_byte"] == item["bytes"] >= 0
            and isinstance(item.get("sha256"), str) and HEX_DIGEST.fullmatch(item["sha256"])
            for item in slices
        ):
            blockers.append(finding("packet_response", f"{packet_id} {slice_field} is malformed"))
    context = response.get("context")
    if not isinstance(context, dict) or set(context) != {"input_tokens", "output_tokens", "cost", "overflow", "truncated"}:
        blockers.append(finding("packet_response", f"{packet_id} context shape is malformed"))
    elif not isinstance(context.get("overflow"), bool) or not isinstance(context.get("truncated"), bool):
        blockers.append(finding("packet_response", f"{packet_id} context flags are malformed"))
    else:
        for field in ("input_tokens", "output_tokens"):
            value = context.get(field)
            if value is not None and (
                not isinstance(value, int) or isinstance(value, bool) or value < 0
            ):
                blockers.append(finding("packet_response", f"{packet_id} context {field} is malformed"))
        cost = context.get("cost")
        if cost is not None and (
            not isinstance(cost, (int, float)) or isinstance(cost, bool) or cost < 0
        ):
            blockers.append(finding("packet_response", f"{packet_id} context cost is malformed"))
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


def dedup_tokens(value: str) -> list[str]:
    return [
        token for token in re.findall(r"[a-z][a-z0-9_.:/-]{1,63}", value.lower())
        if token not in DEDUP_STOP_WORDS
    ]


def category_families(category: str) -> list[str]:
    normalized = re.sub(r"[^a-z0-9]+", "_", category.lower()).strip("_") or "uncategorized"
    families = {normalized}
    for pattern, family in CATEGORY_FAMILY_RULES:
        if pattern.search(normalized):
            families.add(family)
    if re.search(r"(?:exact|version|revision|blob|manifest|synthesi|stale|machine_contract|response|hypothesis|review_behavior|lab_evidence|workflow_contract)", normalized):
        families.add("review_evidence_integrity")
    if re.search(r"(?:authority|authoritative|phase_state|schema_migration|mirror|one_way_migration|temporal|exact_version|machine_contract|evidence_truthfulness)", normalized):
        families.add("authority_state")
    if re.search(r"(?:impact|graph|edge|reference|certainty|path_safety|coverage_active)", normalized):
        families.add("impact_graph")
    if re.search(r"(?:packet|context|slice_bound|slice_bounds|impact_slice_bound|typed_edge_provenance)", normalized):
        families.add("packet_bounds")
    if re.search(r"(?:lab|experiment|resource_bound|information_disclosure|network_test)", normalized):
        families.add("lab_containment")
    if re.search(r"(?:workflow|route|gate|shepherd|revert|liveness|runtime|tool_scope|documentation|exact_identity_safety|human_gate_safety|stale_evidence|authoritative_state_consistency)", normalized):
        families.add("workflow_lifecycle")
    if re.search(r"(?:secret|trace|evidence_isolation|evidence_integrity|path_safety)", normalized):
        families.add("evidence_containment")
    return sorted(families)


def packet_source_anchors(packet: dict[str, Any], path: str) -> list[str]:
    anchors: set[str] = set()
    for field in (
        "changed_file_slices", "context_file_slices", "impact_file_slices", "edge_context_slices"
    ):
        for item in packet.get(field, []):
            if item.get("path") == path:
                anchors.add(
                    ":".join(
                        (
                            path,
                            str(item.get("blob_sha256", "")),
                            str(item.get("start_line", "")),
                            str(item.get("sha256", "")),
                        )
                    )
                )
    return sorted(anchors)


def observation_features(packet: dict[str, Any], raw: dict[str, Any]) -> dict[str, Any]:
    payload = {key: raw.get(key) for key in sorted(raw)}
    evidence_text = " ".join(
        str(raw.get(field, "")) for field in ("category", "evidence", "impact")
    )
    correction_text = str(raw.get("smallest_safe_correction", ""))
    mechanism = sorted(set(dedup_tokens(evidence_text)))
    correction = dedup_tokens(correction_text)
    correction_surface = sorted(set([*dedup_tokens(str(raw.get("path", ""))), *correction[:12]]))
    claim_tokens = dedup_tokens(" ".join((evidence_text, correction_text)))
    shingles = sorted(
        {" ".join(claim_tokens[index:index + 3]) for index in range(max(0, len(claim_tokens) - 2))}
    )
    return {
        "versions": DEDUP_KEY_VERSIONS,
        "payload_hash": sha256_json(payload),
        "invariants": sorted(set(packet.get("invariants", []))),
        "category_families": category_families(str(raw.get("category", ""))),
        "source_anchors": packet_source_anchors(packet, str(raw.get("path", ""))),
        "mechanism_terms": mechanism,
        "correction_surface_terms": correction_surface,
        "claim_shingles": shingles,
    }


def build_observations(
    manifest: dict[str, Any],
    finding_records: list[dict[str, Any]],
) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    scope = require_object(manifest.get("scope"), "manifest scope")
    review_round = scope.get("review_round")
    if not isinstance(review_round, int) or isinstance(review_round, bool) or review_round <= 0:
        raise ReviewSystemError("manifest scope lacks a positive review_round for occurrence identity")
    lineage = scope.get("candidate_lineage")
    if not isinstance(lineage, str) or not lineage:
        raise ReviewSystemError("manifest scope lacks candidate lineage")
    manifest_digest = require_object(manifest.get("authentication"), "manifest authentication").get(
        "semantic_manifest_sha256"
    )
    run_seed = {
        "lineage_id": lineage,
        "review_round": review_round,
        "exact_base_sha": manifest.get("exact_base_sha"),
        "exact_head_sha": manifest.get("exact_head_sha"),
        "exact_head_tree": manifest.get("exact_head_tree"),
        "manifest_sha256": manifest_digest,
    }
    run_id = "RUN-" + sha256_json(run_seed)[:20]
    run = {
        "run_id": run_id,
        **run_seed,
    }
    observations: list[dict[str, Any]] = []
    for global_ordinal, record in enumerate(finding_records, 1):
        packet = require_object(record.get("packet"), "observation packet")
        raw = require_object(record.get("finding"), "observation finding")
        response_digest = record.get("response_sha256")
        raw_ordinal = record.get("raw_ordinal")
        if not isinstance(response_digest, str) or not HEX_DIGEST.fullmatch(response_digest):
            raise ReviewSystemError("observation response digest is malformed")
        if not isinstance(raw_ordinal, int) or raw_ordinal <= 0:
            raise ReviewSystemError("observation raw ordinal is malformed")
        identity_seed = {
            "run_id": run_id,
            "packet_id": packet.get("packet_id"),
            "response_sha256": response_digest,
            "raw_ordinal": raw_ordinal,
        }
        observation_id = "OBS-" + sha256_json(identity_seed)[:24]
        observation = {
            "schema_version": FINDING_OBSERVATION_SCHEMA,
            "observation_id": observation_id,
            "display_alias": f"R{review_round}-F{global_ordinal:03d}",
            "run_id": run_id,
            "lineage_id": lineage,
            "review_round": review_round,
            "exact_base_sha": manifest.get("exact_base_sha"),
            "exact_head_sha": manifest.get("exact_head_sha"),
            "exact_head_tree": manifest.get("exact_head_tree"),
            "manifest_sha256": manifest_digest,
            "packet_id": packet.get("packet_id"),
            "response_sha256": response_digest,
            "raw_ordinal": raw_ordinal,
            "finding": json.loads(json.dumps(raw)),
            "evidence_locations": [
                {
                    "path": raw.get("path"),
                    "line": raw.get("line"),
                    "packet_id": packet.get("packet_id"),
                    "response_sha256": response_digest,
                }
            ],
            "features": observation_features(packet, raw),
        }
        observation["identity_sha256"] = sha256_json(observation)
        observations.append(observation)
    return run, observations


def jaccard(left: Iterable[str], right: Iterable[str]) -> float:
    left_set, right_set = set(left), set(right)
    union = left_set | right_set
    return len(left_set & right_set) / len(union) if union else 0.0


def generate_candidate_pairs(observations: list[dict[str, Any]]) -> dict[str, Any]:
    by_id = {item["observation_id"]: item for item in observations}
    term_frequency: dict[str, int] = defaultdict(int)
    for item in observations:
        for term in set(item["features"]["mechanism_terms"]):
            term_frequency[term] += 1
    candidates: list[dict[str, Any]] = []
    for index, left in enumerate(observations):
        for right in observations[index + 1:]:
            left_features, right_features = left["features"], right["features"]
            reasons: list[str] = []
            if left_features["payload_hash"] == right_features["payload_hash"]:
                reasons.append("exact_payload")
            same_invariant = bool(set(left_features["invariants"]) & set(right_features["invariants"]))
            same_family = bool(
                set(left_features["category_families"]) & set(right_features["category_families"])
            )
            if same_invariant and same_family:
                reasons.append("invariant_category")
            if set(left_features["source_anchors"]) & set(right_features["source_anchors"]) and same_family:
                reasons.append("source_anchor_category")
            mechanism_overlap = set(left_features["mechanism_terms"]) & set(right_features["mechanism_terms"])
            if same_invariant and len(mechanism_overlap) >= 2:
                reasons.append("invariant_mechanism")
            correction_score = jaccard(
                left_features["correction_surface_terms"], right_features["correction_surface_terms"]
            )
            if correction_score >= 0.5 and min(
                len(left_features["correction_surface_terms"]),
                len(right_features["correction_surface_terms"]),
            ) >= 2:
                reasons.append("correction_surface")
            shingle_score = jaccard(left_features["claim_shingles"], right_features["claim_shingles"])
            rare_terms = sorted(term for term in mechanism_overlap if term_frequency[term] <= 2)
            if shingle_score >= 0.45 or len(rare_terms) >= 2:
                reasons.append("text_candidate")
            if not reasons:
                continue
            incompatible_invariants = bool(
                left_features["invariants"] and right_features["invariants"] and not same_invariant
            )
            pair = sorted((left["observation_id"], right["observation_id"]))
            candidates.append(
                {
                    "candidate_id": "PAIR-" + sha256_json(pair)[:20],
                    "observation_ids": pair,
                    "reasons": sorted(set(reasons)),
                    "scores": {
                        "correction_surface_jaccard": round(correction_score, 6),
                        "claim_shingle_jaccard": round(shingle_score, 6),
                    },
                    "hard_non_merge_signals": (
                        ["incompatible_assigned_invariants"] if incompatible_invariants else []
                    ),
                    "rare_terms": rare_terms,
                }
            )
    candidates.sort(key=lambda item: item["candidate_id"])
    possible = len(observations) * (len(observations) - 1) // 2
    return {
        "algorithm": "deterministic-partial-keys/v1",
        "model_tokens": 0,
        "model_assistance": "disabled",
        "possible_pairs": possible,
        "candidate_pair_count": len(candidates),
        "candidate_pair_fraction": len(candidates) / possible if possible else 0.0,
        "pairs": candidates,
        "observation_ids_sha256": sha256_json(sorted(by_id)),
    }


def load_dedup_decisions(root: Path, relative: str | None, run: dict[str, Any], observations: list[dict[str, Any]]) -> list[dict[str, Any]]:
    if not relative:
        return []
    document = require_object(load_json(resolve_safe(root, relative)), "dedup decisions")
    if document.get("schema_version") != DEDUP_DECISIONS_SCHEMA:
        raise ReviewSystemError("dedup decisions schema migration required")
    if document.get("run_id") != run["run_id"] or document.get("lineage_id") != run["lineage_id"]:
        raise ReviewSystemError("dedup decisions are stale for this exact run/lineage")
    events = document.get("events")
    if not isinstance(events, list):
        raise ReviewSystemError("dedup decision events must be a list")
    known = {item["observation_id"] for item in observations}
    previous = None
    result: list[dict[str, Any]] = []
    for sequence, event in enumerate(events, 1):
        if not isinstance(event, dict) or event.get("sequence") != sequence:
            raise ReviewSystemError("dedup decision sequence is not append-only")
        if event.get("previous_event_sha256") != previous:
            raise ReviewSystemError("dedup decision hash chain differs")
        supplied_digest = event.get("event_sha256")
        unsigned = {key: value for key, value in event.items() if key != "event_sha256"}
        if supplied_digest != sha256_json(unsigned):
            raise ReviewSystemError("dedup decision event digest differs")
        decision = event.get("decision")
        ids = event.get("observation_ids")
        if decision not in {"same", "distinct", "ambiguous", "recurrence", "split", "reassign"}:
            raise ReviewSystemError("dedup decision is outside the exact enum")
        if not isinstance(ids, list) or not ids or any(item not in known for item in ids):
            raise ReviewSystemError("dedup decision cites unknown observations")
        if decision in {"same", "distinct", "ambiguous", "split"} and len(set(ids)) != 2:
            raise ReviewSystemError("pair decision must cite exactly two observations")
        if decision in {"recurrence", "reassign"} and len(set(ids)) != 1:
            raise ReviewSystemError("recurrence/reassign decision must cite exactly one observation")
        causal = event.get("causal_test")
        if not isinstance(causal, dict) or set(causal) != {
            "same_invariant", "same_mechanism", "single_treatment", "contradictions",
            "counterfactual_all_cease", "evidence_observation_ids",
        }:
            raise ReviewSystemError("dedup decision lacks the exact causal test")
        if not isinstance(causal.get("contradictions"), list):
            raise ReviewSystemError("dedup contradictions must be a list")
        if sorted(set(causal.get("evidence_observation_ids", []))) != sorted(set(ids)):
            raise ReviewSystemError("dedup causal evidence does not cover every observation")
        if decision in {"same", "recurrence", "reassign"} and not (
            causal.get("same_invariant") is True
            and causal.get("same_mechanism") is True
            and causal.get("single_treatment") is True
            and causal.get("counterfactual_all_cease") is True
            and not causal.get("contradictions")
        ):
            raise ReviewSystemError("same/recurrence/reassign decision does not prove the five-part causal test")
        if decision in {"recurrence", "reassign"} and not isinstance(event.get("root_id"), str):
            raise ReviewSystemError("recurrence/reassign decision lacks a stable root id")
        if decision == "reassign" and not isinstance(event.get("retired_root_id"), str):
            raise ReviewSystemError("reassign decision lacks a retained retired root id")
        previous = supplied_digest
        result.append(event)
    return result


def load_dedup_history(root: Path, relative: str | None, lineage: str) -> dict[str, Any]:
    if not relative:
        return {"schema_version": DEDUP_HISTORY_SCHEMA, "lineage_id": lineage, "roots": []}
    document = require_object(load_json(resolve_safe(root, relative)), "dedup history")
    if document.get("schema_version") != DEDUP_HISTORY_SCHEMA or document.get("lineage_id") != lineage:
        raise ReviewSystemError("dedup history schema or lineage differs")
    if not isinstance(document.get("roots"), list):
        raise ReviewSystemError("dedup history roots must be a list")
    return document


def synthesize_dedup(
    root: Path,
    manifest: dict[str, Any],
    finding_records: list[dict[str, Any]],
    decisions_relative: str | None,
    history_relative: str | None,
) -> dict[str, Any]:
    run, observations = build_observations(manifest, finding_records)
    candidates = generate_candidate_pairs(observations)
    decisions = load_dedup_decisions(root, decisions_relative, run, observations)
    history = load_dedup_history(root, history_relative, run["lineage_id"])
    known_history = {
        item.get("root_id"): item for item in history["roots"]
        if isinstance(item, dict) and isinstance(item.get("root_id"), str)
    }
    pair_decisions: dict[tuple[str, str], str] = {}
    recurrence: dict[str, str] = {}
    for event in decisions:
        ids = sorted(set(event["observation_ids"]))
        if event["decision"] in {"recurrence", "reassign"}:
            if event["root_id"] not in known_history:
                raise ReviewSystemError("recurrence/reassign decision cites an unknown historical root")
            recurrence[ids[0]] = event["root_id"]
        else:
            pair_decisions[(ids[0], ids[1])] = (
                "distinct" if event["decision"] == "split" else event["decision"]
            )

    clusters: list[list[str]] = []
    for observation_id in sorted(item["observation_id"] for item in observations):
        placed = False
        for cluster in clusters:
            relations = [
                pair_decisions.get(tuple(sorted((observation_id, member)))) for member in cluster
            ]
            if relations and all(relation == "same" for relation in relations):
                cluster.append(observation_id)
                placed = True
                break
        if not placed:
            clusters.append([observation_id])
    by_id = {item["observation_id"]: item for item in observations}
    severity_rank = {"critical": 4, "high": 3, "medium": 2, "low": 1}
    roots: list[dict[str, Any]] = []
    occurrences: list[dict[str, Any]] = []
    for members in clusters:
        recurrence_roots = {recurrence[item] for item in members if item in recurrence}
        if len(recurrence_roots) > 1:
            raise ReviewSystemError("one occurrence cannot recur from multiple stable roots")
        root_id = next(iter(recurrence_roots), None)
        if root_id is None:
            root_id = "RC-" + hashlib.sha256(
                (run["lineage_id"] + "\0" + min(members)).encode()
            ).hexdigest()[:16]
        representative = sorted(
            members,
            key=lambda item: (
                -len(by_id[item]["evidence_locations"]),
                -severity_rank.get(by_id[item]["finding"].get("severity"), 0),
                -len(str(by_id[item]["finding"].get("smallest_safe_correction", ""))),
                item,
            ),
        )[0]
        occurrence_id = "OCC-" + hashlib.sha256(
            (root_id + "\0" + run["run_id"]).encode()
        ).hexdigest()[:20]
        occurrences.append(
            {
                "schema_version": OCCURRENCE_SCHEMA,
                "occurrence_id": occurrence_id,
                "root_id": root_id,
                "run_id": run["run_id"],
                "recurrence_of": root_id if root_id in recurrence_roots else None,
                "representative_observation_id": representative,
                "observation_ids": sorted(members),
                "aggregate_severity": max(
                    (by_id[item]["finding"].get("severity", "low") for item in members),
                    key=lambda value: severity_rank.get(value, 0),
                ),
                "evidence_backlinks": [
                    {"observation_id": item, "locations": by_id[item]["evidence_locations"]}
                    for item in sorted(members)
                ],
            }
        )
        historical = known_history.get(root_id, {})
        representative_finding = by_id[representative]["finding"]
        roots.append(
            {
                "schema_version": ROOT_CAUSE_SCHEMA,
                "root_id": root_id,
                "lineage_id": run["lineage_id"],
                "friendly_aliases": sorted(set(historical.get("friendly_aliases", []))),
                "status": "active",
                "pm_root_summary": {
                    "invariant": by_id[representative]["features"]["invariants"],
                    "mechanism": representative_finding.get("evidence"),
                    "correction_owner": representative_finding.get("path"),
                    "regression_treatment": representative_finding.get("smallest_safe_correction"),
                },
                "occurrence_ids": sorted([*historical.get("occurrence_ids", []), occurrence_id]),
                "membership_events": [
                    {
                        "observation_id": item,
                        "action": "assigned",
                        "decision_source": "parent_orchestrator",
                        "run_id": run["run_id"],
                    }
                    for item in sorted(members)
                ],
                "retired_root_ids": sorted(
                    set(historical.get("retired_root_ids", []))
                    | {
                        event["retired_root_id"] for event in decisions
                        if event.get("decision") == "reassign"
                        and event.get("root_id") == root_id
                    }
                ),
            }
        )
    candidate_pair_ids = {
        tuple(item["observation_ids"]): item["candidate_id"] for item in candidates["pairs"]
    }
    candidate_decisions = []
    for pair, candidate_id in sorted(candidate_pair_ids.items()):
        candidate_decisions.append(
            {
                "candidate_id": candidate_id,
                "observation_ids": list(pair),
                "decision": pair_decisions.get(pair, "ambiguous"),
            }
        )
    raw_count = len(observations)
    disclosure = {
        "raw_observation_count": raw_count,
        "flat_observation_ids": [item["observation_id"] for item in observations],
        "observation_identity_sha256": sha256_json(
            [item["identity_sha256"] for item in observations]
        ),
        "retained_observation_count": sum(len(item["observation_ids"]) for item in occurrences),
    }
    if disclosure["retained_observation_count"] != raw_count:
        raise ReviewSystemError("dedup disclosure does not retain every raw observation exactly once")
    result = {
        "schema_version": DEDUP_SCHEMA,
        "run": run,
        "key_versions": DEDUP_KEY_VERSIONS,
        "observations": observations,
        "candidates": candidates,
        "candidate_decisions": candidate_decisions,
        "decision_events": decisions,
        "occurrences": sorted(occurrences, key=lambda item: item["occurrence_id"]),
        "roots": sorted(roots, key=lambda item: item["root_id"]),
        "disclosure": disclosure,
        "raw_finding_count": raw_count,
        "occurrence_count": len(occurrences),
        "root_count": len(roots),
        "duplicate_observation_count": raw_count - len(occurrences),
        "duplicate_ratio": (raw_count - len(occurrences)) / raw_count if raw_count else 0.0,
        "optional_model_assistance": {
            "enabled": False,
            "reason": "disabled until precision, recall, integrity, and trustworthy token/cost gates pass",
        },
    }
    result["deterministic_digest"] = sha256_json(result)
    return result


def aggregate_validated_telemetry(response_records: list[dict[str, Any]], manifest: dict[str, Any]) -> dict[str, Any]:
    contexts = [record["response"].get("context", {}) for record in response_records]
    walls = [record["response"].get("wall_clock_ms") for record in response_records]
    def aggregate(field: str) -> dict[str, Any]:
        values = [context.get(field) for context in contexts if context.get(field) is not None]
        return {
            "available_count": len(values),
            "missing_count": len(contexts) - len(values),
            "total": sum(values) if values else None,
        }
    return {
        "packet_count": len(manifest.get("packets", [])),
        "exact_rendered_prompt_bytes": sum(
            int(packet.get("context", {}).get("rendered_prompt_bytes", 0))
            for packet in manifest.get("packets", [])
        ),
        "input_tokens": aggregate("input_tokens"),
        "output_tokens": aggregate("output_tokens"),
        "cost": aggregate("cost"),
        "wall_clock_ms": {
            "available_count": sum(value is not None for value in walls),
            "missing_count": sum(value is None for value in walls),
            "total": sum(value for value in walls if value is not None) if any(value is not None for value in walls) else None,
        },
        "claim": "only validated provider/runtime telemetry is aggregated; null values are never inferred",
    }


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
    config = manifest.get("config")
    authentication = manifest.get("authentication")
    trust_root = manifest.get("trust_root")
    if not isinstance(trust_root, dict) or trust_root.get("schema_version") != TRUST_BUNDLE_SCHEMA or trust_root.get("source") != "external_parent_owned_exact_binding":
        blockers.append(finding("compile_manifest", "manifest external trust root is absent or incompatible"))
    if not isinstance(scope, dict) or scope.get("schema_version") != SCOPE_SCHEMA:
        blockers.append(finding("compile_manifest", "manifest scope binding is absent or incompatible"))
    if not isinstance(config, dict) or config.get("schema_version") != CONFIG_SCHEMA:
        blockers.append(finding("compile_manifest", "manifest config binding is absent or incompatible"))
    if not isinstance(authentication, dict) or authentication.get("algorithm") != "sha256-canonical-json-v1":
        blockers.append(finding("compile_manifest", "manifest authentication is absent or incompatible"))
    else:
        unsigned = {key: value for key, value in manifest.items() if key != "authentication"}
        expected = {
            "coverage_sha256": sha256_json(manifest.get("coverage_manifest")),
            "packets_sha256": sha256_json(manifest.get("packets")),
            "semantic_manifest_sha256": sha256_json(unsigned),
        }
        for field, digest in expected.items():
            if authentication.get(field) != digest:
                blockers.append(finding("compile_manifest", f"manifest {field} authentication differs"))
    packet_ids: list[str] = []
    for packet in manifest.get("packets", []) if isinstance(manifest.get("packets"), list) else []:
        packet_id = packet.get("packet_id") if isinstance(packet, dict) else None
        if not isinstance(packet_id, str) or not SAFE_ID.fullmatch(packet_id):
            blockers.append(finding("compile_manifest", "manifest contains an unsafe packet id"))
        else:
            packet_ids.append(packet_id)
    if len(packet_ids) != len(set(packet_ids)):
        blockers.append(finding("compile_manifest", "manifest packet ids are not unique"))
    return blockers


def command_synthesize(args: argparse.Namespace) -> int:
    try:
        root = Path(args.repo_root).resolve(strict=True)
        manifest = require_object(load_json(resolve_safe(root, args.manifest)), "compile manifest")
        if manifest.get("schema_version") != COMPILE_SCHEMA:
            raise ReviewSystemError(
                f"compile manifest migration required: {manifest.get('schema_version')!r} != {COMPILE_SCHEMA!r}"
            )
        responses_relative = validate_relative_path(args.responses_dir, "responses directory")
        responses_candidate = root / responses_relative
        if responses_candidate.is_symlink():
            raise ReviewSystemError("responses directory must not be a symlink")
        try:
            responses_dir = responses_candidate.resolve(strict=True)
        except OSError as exc:
            raise ReviewSystemError(f"responses directory is unavailable: {exc}") from exc
        root_resolved = root.resolve(strict=True)
        if Path(os.path.commonpath((str(root_resolved), str(responses_dir)))) != root_resolved or not responses_dir.is_dir():
            raise ReviewSystemError("responses directory escapes repository or is not a directory")
        findings: list[dict[str, Any]] = []
        blockers = verify_manifest_candidate(root, manifest)
        config_binding = manifest.get("config") if isinstance(manifest.get("config"), dict) else {}
        scope_binding = manifest.get("scope") if isinstance(manifest.get("scope"), dict) else {}
        try:
            rebuilt = compile_document(
                root,
                config_binding.get("path", ""),
                scope_binding.get("path", ""),
                manifest.get("exact_base_sha", ""),
                manifest.get("exact_head_sha", ""),
            )
            rebuilt = bind_trust(root, rebuilt, args.trust_bundle)
            if canonical_json_bytes(rebuilt) != canonical_json_bytes(manifest):
                blockers.append(finding("compile_manifest", "manifest semantics differ from deterministic exact-commit recompilation"))
        except (ReviewSystemError, OSError, TypeError, KeyError) as exc:
            blockers.append(finding("compile_manifest", f"deterministic manifest authentication failed: {exc}"))
        response_records: list[dict[str, Any]] = []
        finding_records: list[dict[str, Any]] = []
        packets = manifest.get("packets", []) if isinstance(manifest.get("packets"), list) else []
        expected_response_names = {f"{packet.get('packet_id')}.json" for packet in packets if isinstance(packet, dict)}
        for child in responses_dir.iterdir():
            if child.is_symlink() or not child.is_file() or child.name not in expected_response_names:
                blockers.append(finding("packet_response", f"unexpected or unsafe response child: {child.name}"))
        for packet in packets:
            packet_id = packet.get("packet_id") if isinstance(packet, dict) else None
            if not isinstance(packet_id, str) or not SAFE_ID.fullmatch(packet_id):
                blockers.append(finding("packet_response", "unsafe packet id cannot select a response child"))
                continue
            response_path = responses_dir / f"{packet_id}.json"
            if response_path.is_symlink() or not response_path.is_file() or response_path.parent != responses_dir:
                blockers.append(finding("packet_response", f"missing or unsafe packet response {packet_id}"))
                continue
            try:
                response_bytes = read_file_bounded(response_path, MAX_RESPONSE_BYTES)
                response_value = json.loads(response_bytes)
            except json.JSONDecodeError as exc:
                blockers.append(finding("packet_response", f"{packet_id} response JSON is malformed: {exc}"))
                continue
            if not isinstance(response_value, dict):
                blockers.extend(validate_response_shape(packet_id, response_value))
                continue
            response = response_value
            response_digest = hashlib.sha256(response_bytes).hexdigest()
            response_records.append(
                {"packet": packet, "response": response, "response_sha256": response_digest}
            )
            blockers.extend(validate_response_shape(packet_id, response))
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
            for slice_field in ("changed_file_slices", "context_file_slices", "impact_file_slices", "edge_context_slices"):
                if response.get(slice_field) != packet.get(slice_field, []):
                    blockers.append(finding("packet_coverage", f"{packet_id} {slice_field} differs from exact assignment"))
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
            for raw_ordinal, item in enumerate(packet_findings, 1):
                if isinstance(item, dict):
                    findings.append({"packet_id": packet["packet_id"], **item})
                    finding_records.append(
                        {
                            "packet": packet,
                            "finding": item,
                            "response_sha256": response_digest,
                            "raw_ordinal": raw_ordinal,
                        }
                    )
        blockers.extend(verify_manifest_candidate(root, manifest))
        dedup: dict[str, Any] | None = None
        try:
            dedup = synthesize_dedup(
                root, manifest, finding_records, args.dedup_decisions, args.dedup_history
            )
            if dedup["raw_finding_count"] != len(findings):
                raise ReviewSystemError("dedup raw finding count differs from flat synthesis disclosure")
        except (ReviewSystemError, OSError, TypeError, KeyError, json.JSONDecodeError) as exc:
            blockers.append(finding("dedup_integrity", str(exc)))
        telemetry = aggregate_validated_telemetry(response_records, manifest)
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
                "dedup": dedup,
                "telemetry": telemetry,
                "blockers": blockers,
                "shepherd": {"status": "pending", "rule": "run independently only after clean synthesis"},
                "human_merge_authority": True,
            }
        )
        return 0 if status == "clean" else 1
    except (ReviewSystemError, OSError, TypeError, AttributeError, KeyError, json.JSONDecodeError, subprocess.SubprocessError) as exc:
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
    compile_parser.add_argument("--trust-bundle", default=os.environ.get("PM_REVIEW_TRUST_BUNDLE"))
    compile_parser.set_defaults(func=command_compile)

    render_parser = subparsers.add_parser("render", help="render one authenticated bounded packet prompt")
    render_parser.add_argument("--repo-root", default=".")
    render_parser.add_argument("--manifest", required=True)
    render_parser.add_argument("--packet-id", required=True)
    render_parser.add_argument("--trust-bundle", default=os.environ.get("PM_REVIEW_TRUST_BUNDLE"))
    render_parser.set_defaults(func=command_render)

    synthesize_parser = subparsers.add_parser("synthesize", help="synthesize raw packet responses into one PM verdict")
    synthesize_parser.add_argument("--repo-root", default=".")
    synthesize_parser.add_argument("--manifest", required=True)
    synthesize_parser.add_argument("--responses-dir", required=True)
    synthesize_parser.add_argument("--trust-bundle", default=os.environ.get("PM_REVIEW_TRUST_BUNDLE"))
    synthesize_parser.add_argument("--dedup-decisions")
    synthesize_parser.add_argument("--dedup-history")
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
