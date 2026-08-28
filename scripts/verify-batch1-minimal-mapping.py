#!/usr/bin/env python3
"""Verify the Batch 1 source-to-runtime handoff remains mechanical and cited."""

from __future__ import annotations

import json
import re
import hashlib
import subprocess
import sys
from collections import Counter
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
PACKAGE = ROOT / "docs" / "architecture" / "batch1-minimal-mapping"
SOURCE_PATH = PACKAGE / "source_operations.json"
MAPPING_PATH = PACKAGE / "mapping.json"
FOUNDATIONS_PATH = PACKAGE / "missing-foundations.json"
PROVIDERS = {
    "asana", "bitbucket", "circleci", "dockerhub", "gitlab",
    "jira", "notion", "sentry", "stripe", "vercel",
}
LANES = {
    "binary_download", "binary_upload", "direct_read", "direct_write",
    "etl", "reverse_etl",
}
STATUSES = {
    "existing_executor", "needs_json_materialization",
    "missing_runtime_foundation", "provider_not_applicable",
}
BANNED_FOUNDATION_TERMS = {
    "certification", "credential", "credentials", "hash", "import",
    "source-lock", "source_lock", "live-test", "live_test", "delete",
    "destructive", "missing mapping", "mapping declaration",
}
NON_GATING_METADATA = {
    "source_lock_hash_retention": "non_gating_metadata_only",
    "source_import_success": "non_gating_metadata_only",
    "certification_status": "non_gating_metadata_only",
    "credential_availability": "non_gating_metadata_only",
    "destructive_method": "non_gating_metadata_only",
    "existing_command_or_runtime_binding": "non_gating_metadata_only",
}
LOCK_CACHE: dict[str, dict[str, Any]] = {}
ARTIFACT_CACHE: dict[str, bytes] = {}


class VerificationError(RuntimeError):
    pass


def strict_object(pairs: list[tuple[str, Any]]) -> dict[str, Any]:
    result: dict[str, Any] = {}
    for key, value in pairs:
        if key in result:
            raise VerificationError(f"duplicate JSON key {key!r}")
        result[key] = value
    return result


def reject_constant(value: str) -> None:
    raise VerificationError(f"non-finite JSON constant {value!r}")


def read_json(path: Path) -> dict[str, Any]:
    try:
        parsed = json.loads(
            path.read_text(encoding="utf-8"),
            object_pairs_hook=strict_object,
            parse_constant=reject_constant,
        )
    except (OSError, json.JSONDecodeError, VerificationError) as exc:
        raise VerificationError(f"{path}: {exc}") from exc
    if not isinstance(parsed, dict):
        raise VerificationError(f"{path}: top-level value must be an object")
    return parsed


def require_fields(record: dict[str, Any], fields: set[str], label: str) -> None:
    absent = sorted(field for field in fields if field not in record)
    if absent:
        raise VerificationError(f"{label}: missing required fields {', '.join(absent)}")


def source_citation(record: dict[str, Any], label: str) -> tuple[str, str, str]:
    require_fields(
        record,
        {
            "source_url", "source_location", "retained_source",
        },
        label,
    )
    retained = record["retained_source"]
    if not isinstance(retained, dict):
        raise VerificationError(f"{label}: retained_source must be an object")
    require_fields(retained, {"sha256", "reference", "lock_reference"}, f"{label}.retained_source")
    digest = retained["sha256"]
    if not isinstance(digest, str) or not re.fullmatch(r"[0-9a-f]{64}", digest):
        raise VerificationError(f"{label}: retained-source digest must be lowercase SHA-256")
    if not isinstance(retained["reference"], str) or not retained["reference"]:
        raise VerificationError(f"{label}: retained-source reference is required")
    return record["source_url"], record["source_location"], digest, retained["reference"], retained["lock_reference"]


def mapping_citation(record: dict[str, Any], label: str) -> tuple[str, str]:
    require_fields(record, {"source_url", "source_location"}, label)
    if not isinstance(record["source_url"], str) or not record["source_url"]:
        raise VerificationError(f"{label}: source URL is required")
    if not isinstance(record["source_location"], str) or not record["source_location"]:
        raise VerificationError(f"{label}: source location is required")
    return record["source_url"], record["source_location"]


def git_show(reference: str) -> bytes:
    if not reference.startswith("git:"):
        raise VerificationError(f"retained-source reference must be a git object: {reference!r}")
    spec = reference.removeprefix("git:")
    try:
        return subprocess.run(
            ["git", "show", spec],
            cwd=ROOT,
            check=True,
            capture_output=True,
        ).stdout
    except subprocess.CalledProcessError as exc:
        raise VerificationError(f"unable to read retained-source reference {reference!r}") from exc


def verify_retained_source(source: dict[str, Any], label: str) -> None:
    citation = source_citation(source, label)
    source_url, source_location, digest, artifact_ref, lock_ref = citation
    artifact = ARTIFACT_CACHE.get(artifact_ref)
    if artifact is None:
        artifact = git_show(artifact_ref)
        ARTIFACT_CACHE[artifact_ref] = artifact
    if hashlib.sha256(artifact).hexdigest() != digest:
        raise VerificationError(f"{label}: retained artifact digest does not match citation")
    lock = LOCK_CACHE.get(lock_ref)
    if lock is None:
        try:
            lock = json.loads(
                git_show(lock_ref).decode("utf-8"),
                object_pairs_hook=strict_object,
                parse_constant=reject_constant,
            )
        except (UnicodeDecodeError, json.JSONDecodeError, VerificationError) as exc:
            raise VerificationError(f"{label}: invalid retained source lock") from exc
        LOCK_CACHE[lock_ref] = lock
    rest = lock.get("rest")
    if not isinstance(rest, dict) or rest.get("source_url") != source_url or rest.get("sha256") != digest:
        raise VerificationError(f"{label}: retained source lock does not match citation")
    candidates = [
        operation for operation in rest.get("operations", [])
        if operation.get("id") == source["operation_id"]
        and operation.get("method") == source["method"]
        and operation.get("path") == source["path"]
        and operation.get("source_location") == source_location
    ]
    if len(candidates) != 1:
        raise VerificationError(f"{label}: retained source lock cannot prove the cited operation exactly once")


def check_command(command: dict[str, Any], label: str) -> None:
    require_fields(command, {"artifact_ref", "path", "availability"}, label)
    path = command["path"]
    if not isinstance(path, str) or not path:
        raise VerificationError(f"{label}: command path is required")
    normalized = path.lower()
    if "api op-" in normalized:
        raise VerificationError(f"{label}: synthetic api op-* command path")
    if re.search(
        r"(?:^|\s)[^\s]+-(?:binary-download|binary-upload|direct-read|direct-write|etl|reverse-etl)$",
        normalized,
    ):
        raise VerificationError(f"{label}: lane-suffixed command path")


def check_non_gating_policy(document: dict[str, Any], label: str) -> None:
    if document.get("non_gating_metadata") != NON_GATING_METADATA:
        raise VerificationError(f"{label}: non-gating metadata policy drift")


def main() -> int:
    source_document = read_json(SOURCE_PATH)
    mapping_document = read_json(MAPPING_PATH)
    foundations_document = read_json(FOUNDATIONS_PATH)
    check_non_gating_policy(source_document, "source_operations.json")
    check_non_gating_policy(mapping_document, "mapping.json")
    check_non_gating_policy(foundations_document, "missing-foundations.json")

    source_operations = source_document.get("source_operations")
    if not isinstance(source_operations, list):
        raise VerificationError("source_operations.json: source_operations must be an array")
    sources: dict[str, dict[str, Any]] = {}
    providers: set[str] = set()
    for index, source in enumerate(source_operations):
        if not isinstance(source, dict):
            raise VerificationError(f"source_operations[{index}] must be an object")
        label = f"source_operations[{index}]"
        require_fields(
            source,
            {
                "source_identity", "provider", "operation_id", "method", "path",
                "source_url", "source_location", "retained_source",
            },
            label,
        )
        identity = source["source_identity"]
        if not isinstance(identity, str) or not identity or identity in sources:
            raise VerificationError(f"{label}: source identity must be unique and non-empty")
        if source["provider"] not in PROVIDERS:
            raise VerificationError(f"{label}: unrecognised provider {source['provider']!r}")
        source_citation(source, label)
        verify_retained_source(source, label)
        sources[identity] = source
        providers.add(source["provider"])
    if providers != PROVIDERS:
        raise VerificationError(
            "source_operations.json: providers do not match Batch 1 scope: "
            f"expected {sorted(PROVIDERS)}, got {sorted(providers)}"
        )
    if source_document.get("source_operation_count") != len(sources):
        raise VerificationError("source_operations.json: source_operation_count is not exact")

    mappings = mapping_document.get("mappings")
    if not isinstance(mappings, list):
        raise VerificationError("mapping.json: mappings must be an array")
    mapping_ids: set[str] = set()
    missing_foundation_mappings: list[dict[str, Any]] = []
    lane_counts: Counter[str] = Counter()
    for index, mapping in enumerate(mappings):
        if not isinstance(mapping, dict):
            raise VerificationError(f"mappings[{index}] must be an object")
        label = f"mappings[{index}]"
        require_fields(
            mapping,
            {
                "mapping_id", "source_identity", "lane", "status", "source_citation",
                "existing_command", "runtime_target", "lane_evidence",
            },
            label,
        )
        mapping_id = mapping["mapping_id"]
        if not isinstance(mapping_id, str) or not mapping_id or mapping_id in mapping_ids:
            raise VerificationError(f"{label}: mapping_id must be unique and non-empty")
        mapping_ids.add(mapping_id)
        identity = mapping["source_identity"]
        if identity not in sources:
            raise VerificationError(f"{label}: unknown source identity {identity!r}")
        if mapping["lane"] not in LANES:
            raise VerificationError(f"{label}: unrecognised requested lane")
        if mapping["status"] not in STATUSES:
            raise VerificationError(f"{label}: unrecognised status")
        citation = mapping["source_citation"]
        if not isinstance(citation, dict):
            raise VerificationError(f"{label}: source_citation must be an object")
        if mapping_citation(citation, f"{label}.source_citation") != (
            sources[identity]["source_url"], sources[identity]["source_location"]
        ):
            raise VerificationError(f"{label}: source citation does not match source identity")
        runtime_target = mapping["runtime_target"]
        command = mapping["existing_command"]
        if mapping["status"] == "existing_executor":
            if not isinstance(runtime_target, dict) or not runtime_target.get("artifact_ref"):
                raise VerificationError(f"{label}: existing executor requires exact runtime target")
            if not isinstance(mapping["lane_evidence"], dict) or not mapping["lane_evidence"].get("artifact_ref"):
                raise VerificationError(f"{label}: existing executor requires exact lane evidence")
            if not isinstance(command, dict):
                raise VerificationError(f"{label}: existing executor requires exact command")
            check_command(command, f"{label}.command")
            if command["availability"] != "implemented":
                raise VerificationError(f"{label}: executor command must be implemented")
        if mapping["status"] == "needs_json_materialization":
            if command is not None:
                if not isinstance(command, dict):
                    raise VerificationError(f"{label}: existing_command must be an object or null")
                check_command(command, f"{label}.command")
        if mapping["status"] == "missing_runtime_foundation":
            missing_foundation_mappings.append(mapping)
        lane_counts[mapping["lane"]] += 1
    if mapping_document.get("mapping_count") != len(mappings):
        raise VerificationError("mapping.json: mapping_count is not exact")
    expected_lane_counts = {lane: lane_counts[lane] for lane in sorted(LANES)}
    if mapping_document.get("mapping_count_by_lane") != expected_lane_counts:
        raise VerificationError("mapping.json: mapping_count_by_lane is not exact")

    non_lane_dispositions = mapping_document.get("non_lane_dispositions")
    if not isinstance(non_lane_dispositions, list):
        raise VerificationError("mapping.json: non_lane_dispositions must be an array")
    non_lane_ids: set[str] = set()
    non_lane_counts: Counter[str] = Counter()
    for index, disposition in enumerate(non_lane_dispositions):
        if not isinstance(disposition, dict):
            raise VerificationError(f"non_lane_dispositions[{index}] must be an object")
        label = f"non_lane_dispositions[{index}]"
        require_fields(disposition, {"source_identity", "status", "source_citation", "evidence"}, label)
        identity = disposition["source_identity"]
        if identity not in sources or identity in non_lane_ids:
            raise VerificationError(f"{label}: source identity must be known and unique")
        if disposition["status"] != "provider_not_applicable":
            raise VerificationError(f"{label}: only source-cited provider-not-applicable exclusions may be non-lane")
        if mapping_citation(disposition["source_citation"], f"{label}.source_citation") != (
            sources[identity]["source_url"], sources[identity]["source_location"]
        ):
            raise VerificationError(f"{label}: source citation does not match source identity")
        non_lane_ids.add(identity)
        non_lane_counts[disposition["status"]] += 1
    mapped_source_ids = {mapping["source_identity"] for mapping in mappings}
    if mapped_source_ids & non_lane_ids or mapped_source_ids | non_lane_ids != set(sources):
        raise VerificationError("mapping.json: every source must be covered exactly by a lane mapping or non-lane disposition")
    expected_non_lane_counts = {status: non_lane_counts[status] for status in sorted(non_lane_counts)}
    if mapping_document.get("non_lane_disposition_count_by_status") != expected_non_lane_counts:
        raise VerificationError("mapping.json: non_lane_disposition_count_by_status is not exact")

    entries = foundations_document.get("entries")
    if not isinstance(entries, list):
        raise VerificationError("missing-foundations.json: entries must be an array")
    foundation_by_mapping: dict[str, dict[str, Any]] = {}
    for index, entry in enumerate(entries):
        if not isinstance(entry, dict):
            raise VerificationError(f"entries[{index}] must be an object")
        label = f"entries[{index}]"
        require_fields(entry, {"mapping_id", "source_identity", "lane", "foundation_id", "runtime_capability", "evidence", "affected_json_target"}, label)
        mapping_id = entry["mapping_id"]
        if mapping_id in foundation_by_mapping:
            raise VerificationError(f"{label}: duplicate report for mapping {mapping_id!r}")
        foundation_by_mapping[mapping_id] = entry
        forbidden = " ".join(str(entry[field]).lower() for field in ("foundation_id", "runtime_capability", "evidence"))
        if any(term in forbidden for term in BANNED_FOUNDATION_TERMS):
            raise VerificationError(f"{label}: policy/provenance term used as foundation")
    expected_foundation_ids = {mapping["mapping_id"] for mapping in missing_foundation_mappings}
    if set(foundation_by_mapping) != expected_foundation_ids:
        raise VerificationError("missing-foundations.json: report does not match missing_runtime_foundation mappings")
    if foundations_document.get("missing_foundation_count") != len(entries):
        raise VerificationError("missing-foundations.json: missing_foundation_count is not exact")

    print(
        f"PASS: {len(sources)} source operations; {len(mappings)} explicit lane mappings; "
        f"{len(non_lane_dispositions)} non-lane dispositions; {len(entries)} missing foundations"
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except VerificationError as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        raise SystemExit(1)
