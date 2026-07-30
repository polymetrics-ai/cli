#!/usr/bin/env python3
"""Planning-only api_surface validator for Lucid ELD issue #1950.

This script is intentionally scoped under .planning because the issue write scope
forbids shared Go test changes. It validates the #1950 operation ledger against
the public official OpenAPI snapshot fetched for this issue.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import sys
from pathlib import Path
from typing import Any

HTTP_METHODS = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
EXCLUSION_CATEGORIES = {
    "destructive_admin",
    "requires_elevated_scope",
    "binary_payload",
    "deprecated",
    "non_data_endpoint",
    "duplicate_of",
    "out_of_scope",
}
ALLOWED_STREAMS = {"drivers", "vehicles", "vehicle_location_history"}
ALLOWED_DIRECT_READS = {
    "company info get",
    "drivers get",
    "vehicles get",
    "latest driver statuses list",
    "latest vehicle statuses list",
}
ALLOWED_WRITES: set[str] = set()
MAX_REVIEW_AGE_DAYS = 365


def load_json(path: Path) -> Any:
    with path.open("r", encoding="utf-8") as f:
        return json.load(f)


def official_operations(openapi: dict[str, Any]) -> set[tuple[str, str]]:
    out: set[tuple[str, str]] = set()
    for path, path_item in openapi.get("paths", {}).items():
        if not isinstance(path_item, dict):
            continue
        for method in path_item:
            upper = method.upper()
            if upper in HTTP_METHODS:
                out.add((upper, path))
    return out


def classifier_count(endpoint: dict[str, Any]) -> int:
    return sum(1 for key in ("covered_by", "excluded", "operation") if key in endpoint)


def covered_targets(covered_by: dict[str, Any]) -> list[tuple[str, str]]:
    targets: list[tuple[str, str]] = []
    for key in ("stream", "write", "direct_read"):
        value = covered_by.get(key)
        if isinstance(value, str) and value.strip():
            targets.append((key, value))
    for value in covered_by.get("direct_reads", []) or []:
        if isinstance(value, str) and value.strip():
            targets.append(("direct_read", value))
    return targets


def validate(surface: dict[str, Any], openapi: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    official = official_operations(openapi)

    reviewed_at = surface.get("reviewed_at")
    if not isinstance(reviewed_at, str) or not reviewed_at.strip():
        errors.append("reviewed_at is required")
    else:
        try:
            reviewed_date = dt.date.fromisoformat(reviewed_at)
        except ValueError:
            errors.append(f"reviewed_at {reviewed_at!r} is not ISO yyyy-mm-dd")
        else:
            today = dt.date.today()
            if reviewed_date > today:
                errors.append(f"reviewed_at {reviewed_at} is in the future")
            if (today - reviewed_date).days > MAX_REVIEW_AGE_DAYS:
                errors.append(f"reviewed_at {reviewed_at} is stale (> {MAX_REVIEW_AGE_DAYS} days)")

    endpoints = surface.get("endpoints")
    if not isinstance(endpoints, list):
        return errors + ["endpoints must be an array"]

    seen: dict[tuple[str, str], int] = {}
    for i, endpoint in enumerate(endpoints):
        if not isinstance(endpoint, dict):
            errors.append(f"endpoint {i} is not an object")
            continue
        method = str(endpoint.get("method", "")).upper().strip()
        path = str(endpoint.get("path", "")).strip()
        key = (method, path)
        seen[key] = seen.get(key, 0) + 1

        if method not in HTTP_METHODS:
            errors.append(f"endpoint {i} has unsupported method {method!r}")
        if not path.startswith("/"):
            errors.append(f"endpoint {i} path {path!r} must be connector-relative")
        if "*" in path or "{path}" in path or "{proxy+}" in path:
            errors.append(f"endpoint {i} path {path!r} is wildcard-like")
        if classifier_count(endpoint) != 1:
            errors.append(f"endpoint {i} must have exactly one classifier")

        if "covered_by" in endpoint:
            covered = endpoint.get("covered_by")
            if not isinstance(covered, dict):
                errors.append(f"endpoint {i} covered_by must be an object")
                continue
            targets = covered_targets(covered)
            if len(targets) != 1:
                errors.append(f"endpoint {i} covered_by must name exactly one executable target")
                continue
            kind, name = targets[0]
            if kind == "stream" and name not in ALLOWED_STREAMS:
                errors.append(f"endpoint {i} unknown stream target {name!r}")
            if kind == "write" and name not in ALLOWED_WRITES:
                errors.append(f"endpoint {i} unknown write target {name!r}")
            if kind == "direct_read" and name not in ALLOWED_DIRECT_READS:
                errors.append(f"endpoint {i} unknown direct_read target {name!r}")
        if "excluded" in endpoint:
            excluded = endpoint.get("excluded")
            if not isinstance(excluded, dict):
                errors.append(f"endpoint {i} excluded must be an object")
                continue
            category = excluded.get("category")
            if category not in EXCLUSION_CATEGORIES:
                errors.append(f"endpoint {i} invalid exclusion category {category!r}")
            if not str(excluded.get("reason", "")).strip():
                errors.append(f"endpoint {i} exclusion reason is required")

    duplicates = sorted(key for key, count in seen.items() if count > 1)
    for method, path in duplicates:
        errors.append(f"duplicate endpoint {method} {path}")

    found = set(seen)
    missing = sorted(official - found)
    extra = sorted(found - official)
    for method, path in missing:
        errors.append(f"missing official endpoint {method} {path}")
    for method, path in extra:
        errors.append(f"endpoint not present in official OpenAPI {method} {path}")

    return errors


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--surface", required=True, type=Path)
    parser.add_argument("--openapi", required=True, type=Path)
    args = parser.parse_args()

    surface = load_json(args.surface)
    openapi = load_json(args.openapi)
    errors = validate(surface, openapi)
    if errors:
        print(f"FAIL {args.surface}: {len(errors)} error(s)")
        for error in errors:
            print(f"- {error}")
        return 1
    print(f"PASS {args.surface}: {len(surface.get('endpoints', []))} endpoint(s) match official OpenAPI")
    return 0


if __name__ == "__main__":
    sys.exit(main())
