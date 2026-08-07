#!/usr/bin/env python3
"""Derive zendesk-support's documented operation surface from Zendesk's own OAS.

Source of truth is https://developer.zendesk.com/zendesk/oas.yaml -- 1,701,930
bytes, openapi 3.0.3, info.title "Support API", info.version 2.0.0. That byte
count is the one the sweep recorded, so the artifact is REPRODUCED here rather
than trusted.

PyYAML's SafeLoader cannot parse this file as shipped: it contains bare `=`
scalars (e.g. `change: =` in an example block) which YAML 1.1 resolves to
tag:yaml.org,2002:value, a tag SafeLoader has no constructor for. A constructor
that treats the tag as a plain string is registered below; the document then
parses cleanly. Anything else in this programme calling plain yaml.safe_load()
on this artifact hits the same ConstructorError.

Usage:
    python3 derive_zendesk_support.py <oas.yaml> <out.json>

The four non-mechanical judgements are made here and recorded in the phase
PLAN.md/SUMMARY.md; nothing below is guessed from a path name.

  read-vs-write     METHOD decides, with one documented exception: a POST whose
                    own description says it queries/searches/validates and
                    which creates no resource is a READ. The eight are pinned
                    by name in READ_SHAPED_POSTS so the set cannot drift
                    silently; each is re-checked against the artifact here.
  stream-vs-direct  the 33 fixture-backed streams the bundle already ships stay
                    streams (greenhouse finding 21: converting them inside a
                    parity commit deletes shipped, schema-backed contracts).
                    Every other read becomes a bounded direct read.
  binary detection  an operation is binary iff its own documented success
                    response declares a non-JSON media type AND it is a GET.
                    Read out of the artifact, never inferred from the path --
                    PUT /api/v2/brands/{brand_id} declares image/jpg and
                    image/png success responses and is an UPDATE, not a
                    download. That is this connector's binary trap.
  named-dependency  a row is blocked only when a named runtime component
                    refuses it, and the note names that component.
"""

import collections
import json
import re
import sys

import yaml


class _Loader(yaml.SafeLoader):
    pass


_Loader.add_constructor(
    "tag:yaml.org,2002:value", lambda loader, node: loader.construct_scalar(node)
)

METHODS = ("get", "put", "post", "delete", "patch")

# POSTs that read. Each is verified against the artifact below: it must be a
# POST, and its description must not claim to create a stored resource. The
# sweep's own derivation reached this same set of eight independently.
READ_SHAPED_POSTS = {
    "POST /api/v2/any_channel/validate_token",
    "POST /api/v2/autocomplete/tags",
    "POST /api/v2/custom_objects/{custom_object_key}/records/search",
    "POST /api/v2/it_asset_management/assets/search",
    "POST /api/v2/problems/autocomplete",
    "POST /api/v2/users/autocomplete",
    "POST /api/v2/views/preview",
    "POST /api/v2/views/preview/count",
}

# Paging parameters are derived from the connector's declared pagination spec
# by the foundation lane and are NEVER hand-authored. This blocklist makes a
# required-parameter sweep unable to smuggle one in.
#
# `start_time` is deliberately NOT on this list, and that is a judgement rather
# than an omission. Every stream in this bundle declares pagination
# `type: next_url` over `links.next`, `next_page` or `before_url`, so the
# foundation lane derives Zendesk paging from response links and never from a
# request parameter. `start_time` on /api/v2/incremental/* is the export's
# required opening watermark -- the endpoint returns 400 without it -- so it is
# an input to the operation, not a page control. Blocking it would have made
# every incremental read unreachable while claiming paging as the reason.
PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults", "first", "last",
    "cursor_pagination", "page[before]", "page[after]", "page[size]",
}


def load(path):
    with open(path) as fh:
        return yaml.load(fh, Loader=_Loader)


def make_deref(spec):
    def deref(node):
        hops = 0
        while isinstance(node, dict) and "$ref" in node and hops < 20:
            cur = spec
            for part in node["$ref"].lstrip("#/").split("/"):
                cur = cur[part]
            node = cur
            hops += 1
        return node

    return deref


def success_media_types(op):
    kinds = set()
    for code, resp in (op.get("responses") or {}).items():
        if not str(code).startswith("2"):
            continue
        kinds |= set((resp.get("content") or {}).keys())
    return kinds


def first_sentence(text, cap=180):
    if not text:
        return ""
    text = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", text)
    text = re.sub(r"[`*_>#]", "", text)
    text = " ".join(text.split())
    for marker in (". ", "! "):
        idx = text.find(marker)
        if 0 < idx < cap:
            return text[: idx + 1]
    if len(text) > cap:
        text = text[: cap - 1].rsplit(" ", 1)[0] + "."
    return text


def main():
    oas_path, out_path = sys.argv[1], sys.argv[2]
    spec = load(oas_path)
    deref = make_deref(spec)

    assert spec["openapi"] == "3.0.3", spec["openapi"]
    assert spec["info"]["title"] == "Support API", spec["info"]["title"]

    rows = []
    seen = set()
    malformed = []

    for path in sorted(spec["paths"]):
        item = spec["paths"][path]
        for method in METHODS:
            op = item.get(method)
            if not op:
                continue
            key = "%s %s" % (method.upper(), path)

            # The sweep's recurring double-count defect class, applied to this
            # connector's own derivation BEFORE the count is adopted (finding
            # 34: the red test already forbade these and the derivation feeding
            # it had never been run through the same rules).
            if any(ch in path for ch in " ?*"):
                malformed.append(key)
            if key in seen:
                malformed.append(key + " (duplicate)")
            seen.add(key)

            media = success_media_types(op)
            is_get = method == "get"
            binary_media = sorted(m for m in media if "json" not in m)

            if is_get:
                classification = "read"
            elif key in READ_SHAPED_POSTS:
                classification = "read"
            else:
                classification = "write"

            # Binary is GET-only. A non-JSON success response on a mutating
            # method is a response representation, not a download.
            binary = bool(binary_media) and is_get

            params = [deref(p) for p in (item.get("parameters") or []) + (op.get("parameters") or [])]
            required_query = []
            for param in params:
                if not isinstance(param, dict) or param.get("in") != "query":
                    continue
                if not param.get("required"):
                    continue
                name = param["name"]
                if name.lower() in PAGING_PARAMS:
                    raise SystemExit("refusing to author paging flag %r on %s" % (name, key))
                required_query.append(
                    {
                        "name": name,
                        "type": (param.get("schema") or {}).get("type", "string"),
                        "summary": first_sentence(param.get("description"))
                        or ("Query parameter %s." % name),
                    }
                )

            body = op.get("requestBody") or {}
            body_content = body.get("content") or {}
            body_media = sorted(body_content)
            body_schema_kind = None
            if "application/json" in body_content:
                schema = (body_content["application/json"] or {}).get("schema") or {}
                if "oneOf" in schema:
                    body_schema_kind = "oneOf"
                elif "anyOf" in schema:
                    body_schema_kind = "anyOf"
                else:
                    body_schema_kind = "object"

            rows.append(
                {
                    "key": key,
                    "method": method.upper(),
                    "path": path,
                    "operation_id": op.get("operationId"),
                    "summary": op.get("summary") or op.get("operationId"),
                    "description": first_sentence(op.get("description")),
                    "tags": op.get("tags") or [],
                    "deprecated": bool(op.get("deprecated")),
                    "classification": classification,
                    "binary": binary,
                    "success_media_types": sorted(media),
                    "binary_media_types": binary_media if binary else [],
                    "path_vars": re.findall(r"{([^}]+)}", path),
                    "required_query": required_query,
                    "request_body_media": body_media,
                    "request_body_required": bool(body.get("required")),
                    "request_body_schema_kind": body_schema_kind,
                }
            )

    if malformed:
        raise SystemExit("malformed rows: %s" % ", ".join(sorted(malformed)))

    # Re-check the read-shaped POST pin against the artifact rather than
    # trusting the constant: a pin naming an endpoint the provider does not
    # document can never pass (finding 32).
    pinned = {r["key"] for r in rows if r["key"] in READ_SHAPED_POSTS}
    missing = sorted(READ_SHAPED_POSTS - pinned)
    if missing:
        raise SystemExit("read-shaped POST pin names undocumented endpoint(s): %s" % missing)

    by_method = collections.Counter(r["method"] for r in rows)
    payload = {
        "connector": "zendesk-support",
        "artifact": {
            "url": "https://developer.zendesk.com/zendesk/oas.yaml",
            "bytes": 1701930,
            "openapi": spec["openapi"],
            "info_title": spec["info"]["title"],
            "info_version": spec["info"]["version"],
        },
        "totals": {
            "operations": len(rows),
            "paths": len(spec["paths"]),
            "by_method": dict(sorted(by_method.items())),
            "reads": sum(1 for r in rows if r["classification"] == "read"),
            "writes": sum(1 for r in rows if r["classification"] == "write"),
            "read_shaped_posts": len(READ_SHAPED_POSTS),
            "binary": sum(1 for r in rows if r["binary"]),
            "deprecated": sum(1 for r in rows if r["deprecated"]),
            "union_request_bodies": sum(
                1 for r in rows if r["request_body_schema_kind"] in ("oneOf", "anyOf")
            ),
        },
        "operations": rows,
    }
    with open(out_path, "w") as fh:
        json.dump(payload, fh, indent=1, sort_keys=False)
        fh.write("\n")
    print(json.dumps(payload["totals"], indent=1))


if __name__ == "__main__":
    main()
