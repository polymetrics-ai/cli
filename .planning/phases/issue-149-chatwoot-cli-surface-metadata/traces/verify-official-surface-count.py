#!/usr/bin/env python3
"""Verify Chatwoot api_surface.json matches the official Swagger operation set.

This script performs no credentialed checks. It fetches the public official Swagger
and compares only method/path inventory plus method counts.
"""

from __future__ import annotations

import collections
import json
import sys
import urllib.request
from pathlib import Path

OFFICIAL_URL = "https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json"
SURFACE_PATH = Path("internal/connectors/defs/chatwoot/api_surface.json")
METHODS = {"get", "post", "put", "patch", "delete"}
EXPECTED_METHOD_COUNTS = {"POST": 41, "GET": 62, "PATCH": 21, "DELETE": 18, "PUT": 2}
EXPECTED_TOTAL = 144


def load_official() -> list[tuple[str, str]]:
    with urllib.request.urlopen(OFFICIAL_URL, timeout=30) as response:
        swagger = json.load(response)
    operations: list[tuple[str, str]] = []
    for path, item in swagger.get("paths", {}).items():
        for method in item:
            if method.lower() in METHODS:
                operations.append((method.upper(), path))
    return sorted(operations)


def load_surface() -> list[tuple[str, str]]:
    with SURFACE_PATH.open() as fh:
        surface = json.load(fh)
    return sorted((entry.get("method", "").upper(), entry.get("path", "")) for entry in surface.get("endpoints", []))


def main() -> int:
    official = load_official()
    surface = load_surface()
    official_counts = dict(collections.Counter(method for method, _ in official))
    surface_counts = dict(collections.Counter(method for method, _ in surface))
    missing = sorted(set(official) - set(surface))
    extra = sorted(set(surface) - set(official))

    result = {
        "official_total": len(official),
        "surface_total": len(surface),
        "official_method_counts": official_counts,
        "surface_method_counts": surface_counts,
        "missing_count": len(missing),
        "extra_count": len(extra),
        "missing_first_10": [f"{method} {path}" for method, path in missing[:10]],
        "extra_first_10": [f"{method} {path}" for method, path in extra[:10]],
    }
    print(json.dumps(result, indent=2, sort_keys=True))

    ok = (
        len(official) == EXPECTED_TOTAL
        and official_counts == EXPECTED_METHOD_COUNTS
        and len(surface) == EXPECTED_TOTAL
        and surface_counts == EXPECTED_METHOD_COUNTS
        and not missing
        and not extra
    )
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
