#!/usr/bin/env python3
"""Derive Jira Cloud platform v3's documented operation surface.

    python3 derive_jira.py /tmp/sweep/jira.json > DERIVED-OPERATIONS.json

Source of truth is Atlassian's own OpenAPI description, fetched from

    https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json

**The pinned `?_v=` URL 404s and the unpinned one serves a ROLLING SNAPSHOT.**
`info.version` is `1001.0.0-SNAPSHOT-<git sha>`, and the sha moved between the
master plan's derivation (2,445,625 bytes, 420 paths, 616 operations) and this
one (2,449,760 bytes, 421 paths, 617 operations) *on the same calendar day*.
So a byte-count match is NOT available for this connector, and the sweep's
"identical bytes prove it is the same artifact" check cannot be satisfied here.
What replaces it is the sha256 of the exact bytes this derivation read, recorded
in the output, so the derivation can be reproduced against that artifact even
though the URL will have moved on.

The derivation is run through the red test's own rules before its count is
adopted (finding 34): no path contains `?`, `*` or a space; no (method, path)
pair repeats either templated or with its variables normalised away.

The four non-mechanical judgements:

  read-vs-write     every GET is a read. A POST is read-shaped only when the
                    provider's own description says it computes, searches,
                    validates or fetches and persists nothing. The list is
                    pinned by name, not inferred by keyword, because keywords
                    got three of them wrong in the upstream derivation.
  stream-vs-direct  the three shipped streams (issue search, project search,
                    user search) stay streams -- they carry hand-authored
                    schemas and fixtures. Every other GET becomes a plain
                    `direct_read` command, which adds nothing to the operation
                    endpoint ledger (finding 26).
  binary detection  binary iff the documented success response declares an
                    image/* or octet-stream media type AND the operation is a
                    GET (finding 45: binary is GET-only, or a mutation gets
                    silently shipped as a download). Jira has exactly three,
                    all avatar image fetches, and their own summaries say so.
  named-dependency  a write is blocked only when a named runtime component
                    refuses it, and the note names that component.
"""

import collections
import hashlib
import json
import re
import sys

METHODS = ("get", "put", "post", "delete", "patch", "head", "options", "trace")

# The three streams the shipped bundle already covers. Their rows keep
# `covered_by.stream` rather than gaining a duplicate direct-read command.
STREAM_ROWS = {
    "GET /rest/api/3/search": "issues",
    "GET /rest/api/3/project/search": "projects",
    "GET /rest/api/3/users/search": "users",
}

# POSTs Jira documents as READS. Every one was checked against its own
# `description` text, not matched by keyword: the upstream keyword pass put
# `bulkSetIssuesPropertiesList` in this set because the path says "list" (it
# sets properties on issues), and left `analyseExpression` and
# `evaluateJiraExpression` out because they match no read keyword at all.
POST_SHAPED_READS = {
    "POST /rest/api/3/app/field/context/configuration/list",
    "POST /rest/api/3/changelog/bulkfetch",
    "POST /rest/api/3/comment/list",
    "POST /rest/api/3/expression/analyse",
    "POST /rest/api/3/expression/eval",
    "POST /rest/api/3/expression/evaluate",
    "POST /rest/api/3/issue/bulkfetch",
    "POST /rest/api/3/issue/{issueIdOrKey}/changelog/list",
    "POST /rest/api/3/jql/autocompletedata",
    "POST /rest/api/3/jql/function/computation/search",
    "POST /rest/api/3/jql/match",
    "POST /rest/api/3/jql/parse",
    "POST /rest/api/3/jql/pdcleaner",
    "POST /rest/api/3/jql/sanitize",
    "POST /rest/api/3/priorityscheme/mappings",
    "POST /rest/api/3/search",
    "POST /rest/api/3/search/approximate-count",
    "POST /rest/api/3/search/jql",
    "POST /rest/api/3/workflow/history/list",
    "POST /rest/api/3/workflows/create/validation",
    "POST /rest/api/3/workflows/preview",
    "POST /rest/api/3/workflows/update/validation",
    "POST /rest/api/3/worklog/list",
    "POST /rest/atlassian-connect/1/migration/workflow/rule/search",
}

PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults", "first", "last",
    "start", "startat", "start_at", "skip", "top", "next_page_token",
    "nextpagetoken",
}


def kebab(text):
    text = re.sub(r"(?<=[a-z0-9])(?=[A-Z])", "-", text)
    return re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")


def command_action(operation_id):
    """A command word derived from Atlassian's own operationId.

    Jira declares a unique operationId on all 617 operations, which is the
    opposite of Workday (21 of 920) -- so the name comes from the provider
    rather than from the path. Thirteen legacy Connect/Forge ids carry a
    `SomeResource.` prefix and a `_get`/`_put` suffix; both are dropped, and
    the result is asserted collision-free rather than assumed.
    """
    name = operation_id.split(".", 1)[-1]
    name = re.sub(r"_(get|put|post|delete|patch)$", "", name)
    return kebab(name)


def group_of(path):
    segs = [s for s in path.split("/") if s]
    if segs[:3] == ["rest", "api", "3"]:
        rest = segs[3:]
        return kebab(rest[0]) if rest else "root"
    if segs[:2] == ["rest", "atlassian-connect"]:
        return "connect"
    if segs[:2] == ["rest", "forge"]:
        return "forge"
    if segs[:2] == ["rest", "internal"]:
        return "internal"
    raise SystemExit("unclassifiable path prefix: %s" % path)


def success_media(op):
    media = set()
    for code, resp in (op.get("responses") or {}).items():
        if code.startswith("2"):
            media |= set((resp.get("content") or {}).keys())
    return media


def request_content(op):
    return ((op.get("requestBody") or {}).get("content") or {})


def main():
    raw = open(sys.argv[1], "rb").read()
    doc = json.loads(raw)
    paths = doc["paths"]

    rows = []
    for path, item in paths.items():
        for method, op in item.items():
            if method.lower() not in METHODS:
                continue
            rows.append((method.upper(), path, op))

    # --- the red test's own rules, applied to the derivation itself ---------
    malformed = [m + " " + p for m, p, _ in rows if re.search(r"[ ?*]", p)]
    if malformed:
        raise SystemExit("malformed paths in artifact: %s" % malformed)
    dup_templated = [k for k, n in collections.Counter((m, p) for m, p, _ in rows).items() if n > 1]
    if dup_templated:
        raise SystemExit("duplicate templated (method, path): %s" % dup_templated)
    normalised = collections.Counter((m, re.sub(r"\{[^}]*\}", "{}", p)) for m, p, _ in rows)
    dup_resolved = [k for k, n in normalised.items() if n > 1]
    if dup_resolved:
        raise SystemExit("duplicate normalised (method, path): %s" % dup_resolved)

    operations, names = [], {}
    for method, path, op in sorted(rows, key=lambda r: (r[1], r[0])):
        key = method + " " + path
        media = success_media(op)
        content = request_content(op)
        path_vars = re.findall(r"\{([^}]+)\}", path)

        required_query = []
        for param in (op.get("parameters") or []) + (paths[path].get("parameters") or []):
            if param.get("in") != "query" or not param.get("required"):
                continue
            name = param["name"]
            if name.lower() in PAGING_PARAMS:
                raise SystemExit("required paging parameter %r on %s" % (name, key))
            required_query.append(
                {
                    "name": name,
                    "type": (param.get("schema") or {}).get("type") or "string",
                    "summary": (param.get("description") or "Query parameter %s." % name),
                }
            )

        binary = method == "GET" and any(
            m.startswith("image/") or "octet-stream" in m for m in media
        )
        if method != "GET" and any(m.startswith("image/") or "octet-stream" in m for m in media):
            raise SystemExit(
                "%s declares a binary success media type on a non-GET; binary is GET-only "
                "and modelling this as a download would drop its mutation" % key
            )

        is_read = method == "GET" or key in POST_SHAPED_READS
        action = command_action(op["operationId"])
        group = group_of(path)
        command = group + " " + action
        if command in names:
            raise SystemExit("command collision %r: %s and %s" % (command, names[command], key))
        names[command] = key

        operations.append(
            {
                "key": key,
                "method": method,
                "path": path,
                "operation_id": op["operationId"],
                "group": group,
                "command": command,
                "summary": (op.get("summary") or "").strip(),
                "deprecated": bool(op.get("deprecated")),
                "read": is_read,
                "binary": binary,
                "stream": STREAM_ROWS.get(key),
                "path_vars": path_vars,
                "required_query": required_query,
                "request_media": sorted(content),
                "request_body_required": bool((op.get("requestBody") or {}).get("required")),
            }
        )

    by_method = collections.Counter(o["method"] for o in operations)
    json.dump(
        {
            "connector": "jira",
            "artifact_url": "https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json",
            "artifact_bytes": len(raw),
            "artifact_sha256": hashlib.sha256(raw).hexdigest(),
            "oas_version": doc.get("openapi"),
            "info_version": doc["info"].get("version"),
            "paths_count": len(paths),
            "operations_total": len(operations),
            "by_method": dict(sorted(by_method.items())),
            "reads": sum(1 for o in operations if o["read"]),
            "writes": sum(1 for o in operations if not o["read"]),
            "binary_downloads": sorted(o["key"] for o in operations if o["binary"]),
            "post_shaped_reads": sorted(POST_SHAPED_READS),
            "deprecated_count": sum(1 for o in operations if o["deprecated"]),
            "operations": operations,
        },
        sys.stdout,
        indent=2,
    )


if __name__ == "__main__":
    main()
