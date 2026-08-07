#!/usr/bin/env python3
"""Expand github's api_surface/writes/cli_surface to the full documented
mutation surface, from GitHub's own OpenAPI description.

Companion to gen_github_gets.py; the same four judgements apply, and the
non-mechanical ones for mutations are spelled out inline where they are made.
"""

import json
import os
import re
import sys

ROOT = os.path.abspath(os.path.dirname(__file__))
SPEC_PATH = os.path.join(ROOT, "api.github.com.json")

CONFIG_PATH_VARS = {"owner", "repo"}
PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults",
}

# Operations that cannot be an honest implemented command, each with the named
# component that refuses it. Nothing here is "we chose not to": every entry
# names something grep-checkable.
OUT_OF_BASE = {
    ("POST", "/repos/{owner}/{repo}/releases/{release_id}/assets"): (
        "the artifact declares an operation-level server override "
        "(https://uploads.github.com) for this endpoint alone",
        "Named dependency: engine resolves every request against the single configured "
        "base_url (spec.json base_url, default https://api.github.com) and "
        "engine.normalizeDirectReadPathForBaseURL only strips a prefix of that one host; "
        "a per-operation host override is not modelled",
    ),
}

NON_JSON_RESPONSE = {
    ("POST", "/markdown"),
    ("POST", "/markdown/raw"),
}

SECRET_IN_BODY = {
    ("POST", "/applications/{client_id}/token"),
    ("PATCH", "/applications/{client_id}/token"),
    ("DELETE", "/applications/{client_id}/token"),
    ("DELETE", "/applications/{client_id}/grant"),
}

# POSTs that are semantically reads and DO return JSON: bounded bulk lookups
# that use a request body only because the subject list is too long for a query
# string. Modelled as operation-backed direct reads, which is the one read shape
# the runtime allows to use POST.
POST_READS = {
    ("POST", "/orgs/{org}/attestations/bulk-list"),
    ("POST", "/users/{username}/attestations/bulk-list"),
}

SCALAR_TYPES = {"string", "integer", "number", "boolean"}


def deref_factory(spec):
    def deref(node):
        seen = 0
        while isinstance(node, dict) and "$ref" in node and seen < 12:
            cur = spec
            for part in node["$ref"].lstrip("#/").split("/"):
                cur = cur[part]
            node = cur
            seen += 1
        return node
    return deref


def first_sentence(text, cap=160):
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


def kebab(name):
    return name.replace("_", "-").replace(".", "-")


def snake(operation_id):
    return operation_id.replace("/", "_").replace("-", "_")


def command_path(operation_id):
    group, _, action = operation_id.partition("/")
    return "%s %s" % (kebab(group), kebab(action))


def scalar_flag_type(node):
    """Flag type for a schema node, or None when no flag can carry it."""
    if not isinstance(node, dict):
        return None
    values = [str(v) for v in node.get("enum", []) if v is not None]
    types = node.get("type")
    if isinstance(types, list):
        types = next((t for t in types if t != "null"), None)
    if values and types in (None, "string"):
        return "enum"
    if types == "array":
        items = node.get("items") or {}
        item_type = items.get("type")
        if isinstance(item_type, list):
            item_type = next((t for t in item_type if t != "null"), None)
        return "string_array" if item_type in ("string", None) else None
    if types in SCALAR_TYPES:
        return {"string": "string", "integer": "integer",
                "number": "number", "boolean": "boolean"}[types]
    return None


def required_leaf_paths(node, prefix, deref, depth=0):
    """Dotted paths of every field the record schema REQUIRES, mirroring
    validate's requiredMappingPaths recursion, so a required leaf can never go
    unflagged and silently ship as implemented."""
    node = deref(node)
    if not isinstance(node, dict) or depth > 4:
        return []
    out = []
    props = node.get("properties") or {}
    for name in node.get("required", []) or []:
        child = deref(props.get(name) or {})
        path = "%s.%s" % (prefix, name) if prefix else name
        nested = required_leaf_paths(child, path, deref, depth + 1)
        if nested:
            out.extend(nested)
        else:
            out.append((path, child))
    return out


def build(bundle, spec):
    deref = deref_factory(spec)
    surface = json.load(open(os.path.join(bundle, "api_surface.json")))
    cli = json.load(open(os.path.join(bundle, "cli_surface.json")))
    writes = json.load(open(os.path.join(bundle, "writes.json")))
    operations = json.load(open(os.path.join(bundle, "operations.json")))

    have = {(e["method"], e["path"]) for e in surface["endpoints"]}
    cmd_paths = {c["path"] for c in cli["commands"]}
    action_names = {a["name"] for a in writes["actions"]}
    op_ids = {o["id"] for o in operations["operations"]}

    rows, cmds, actions, ops = [], [], [], []
    stats = {"write": 0, "partial": 0, "post_read": 0, "blocked": 0}
    blocked_reasons = {}

    for path in sorted(spec["paths"]):
        item = spec["paths"][path]
        for method_lower in ("post", "put", "patch", "delete"):
            op = item.get(method_lower)
            if not op:
                continue
            method = method_lower.upper()
            key = (method, path)
            if key in have:
                continue

            operation_id = op["operationId"]
            docs = (op.get("externalDocs") or {}).get("url", "") or "https://docs.github.com/en/rest"
            summary = op.get("summary") or operation_id
            if op.get("deprecated"):
                summary += " (deprecated by GitHub, still documented)"

            body = deref(op.get("requestBody")) if op.get("requestBody") else None
            content = (body or {}).get("content") or {}
            schema = deref((content.get("application/json") or {}).get("schema")) if content else None

            reason = note = None
            if key in OUT_OF_BASE:
                reason, note = OUT_OF_BASE[key]
            elif key in NON_JSON_RESPONSE:
                reason = ("documented success response is text/html, not application/json: this is a "
                          "markdown renderer, semantically a read")
                note = ("Named dependency: engine.decodeDirectReadBody json-decodes every direct-read "
                        "body and commandrunner.supportedDirectReadOutputPolicies declares no "
                        "text/html policy")
            elif key in SECRET_IN_BODY:
                reason = "the documented request body carries an OAuth access_token"
                note = ("Named dependency: AGENTS.md forbids accepting credentials as command "
                        "arguments (add credentials from environment variables or stdin), and a "
                        "record/body flag is a command argument")
            elif content and "application/json" not in content:
                reason = "documented request body is %s, not application/json" % ", ".join(sorted(content))
                note = ("Named dependency: engine write actions declare body_type json and the "
                        "declarative write executor emits a JSON body only")
            elif isinstance(schema, dict) and ("oneOf" in schema or "anyOf" in schema):
                arm = "oneOf" if "oneOf" in schema else "anyOf"
                reason = ("documented request body is rooted at %s with %d arms, which is not one "
                          "executable command contract" % (arm, len(schema[arm])))
                note = ("Named dependency: AGENTS.md 'Command Surface Must Stay Executable' -- runtime "
                        "preflight expands %s arms and rejects promotion; each reachable arm must "
                        "become a separately named action before this can be implemented" % arm)

            if reason:
                rows.append({
                    "method": method,
                    "path": path,
                    "operation": {
                        "model": "disallowed",
                        "status": "blocked",
                        "risk": "low",
                        "blocked_by_default": True,
                        "reason": reason,
                        "source_url": docs,
                        "notes": note,
                    },
                })
                stats["blocked"] += 1
                blocked_reasons[key] = reason
                continue

            cmd_path = command_path(operation_id)
            suffix = 2
            while cmd_path in cmd_paths:
                cmd_path = "%s-%d" % (command_path(operation_id), suffix)
                suffix += 1
            cmd_paths.add(cmd_path)

            params = [deref(p) for p in (item.get("parameters", []) + op.get("parameters", []))]
            by_name = {p["name"]: p for p in params if isinstance(p, dict) and "name" in p}
            path_vars = [v for v in re.findall(r"{(\w+)}", path) if v not in CONFIG_PATH_VARS]

            if key in POST_READS:
                emit_post_read(path, op, operation_id, summary, docs, cmd_path, path_vars,
                               by_name, schema, deref, rows, cmds, ops, op_ids)
                stats["post_read"] += 1
                continue

            action_name = snake(operation_id)
            suffix = 2
            while action_name in action_names:
                action_name = "%s_%d" % (snake(operation_id), suffix)
                suffix += 1
            action_names.add(action_name)

            props, required = {}, []
            for var in path_vars:
                param = by_name.get(var, {})
                kind = scalar_flag_type(deref(param.get("schema")) or {"type": "string"}) or "string"
                props[var] = {"type": "integer" if kind == "integer" else "string"}
                required.append(var)
            body_props = (schema or {}).get("properties") or {}
            for name, node in body_props.items():
                if name in props:
                    continue
                props[name] = strip_schema(deref(node), deref)
            for name in (schema or {}).get("required", []) or []:
                if name not in required:
                    required.append(name)

            record_schema = {
                "$schema": "http://json-schema.org/draft-07/schema#",
                "type": "object",
                "required": required,
                "properties": props,
            }

            flags, unflaggable = [], []
            seen_targets = set()
            for target, node in required_leaf_paths(record_schema, "", deref):
                kind = scalar_flag_type(node)
                if kind is None:
                    unflaggable.append(target)
                    continue
                flags.append(make_flag(target, node, kind, by_name, required=True))
                seen_targets.add(target)
            for name, node in props.items():
                if name in seen_targets or any(t.startswith(name + ".") for t in seen_targets):
                    continue
                kind = scalar_flag_type(node)
                if kind is None:
                    continue
                flags.append(make_flag(name, node, kind, by_name, required=False))
                seen_targets.add(name)

            action = {
                "name": action_name,
                "kind": {"POST": "create", "PUT": "update", "PATCH": "update", "DELETE": "delete"}[method],
                "method": method,
                "path": render_path(path),
                "body_type": "json",
                "record_schema": record_schema,
                "risk": risk_text(method, summary),
            }
            if path_vars:
                action["path_fields"] = path_vars
            if method == "DELETE":
                # GitHub's deletes answer 404 for an already-absent subject, so
                # a replayed delete is a no-op rather than an error.
                action["delete"] = {"idempotent": True, "missing_ok_status": [404]}
            actions.append(action)

            availability = "implemented"
            notes = None
            if unflaggable:
                # A required field with no scalar leaf cannot be supplied from
                # the command line, so the command is honestly partial.
                availability = "partial"
                notes = ("Required record field(s) %s have no scalar leaf, so no flag can carry them "
                         "from the command line; the write action is declared and reachable through "
                         "reverse ETL with a record payload." % ", ".join(sorted(unflaggable)))
                stats["partial"] += 1
            else:
                stats["write"] += 1

            cmd = {
                "path": cmd_path,
                "summary": summary,
                "intent": "reverse_etl",
                "availability": availability,
                "write": action_name,
                "source_cli_path": "",
                "source_url": docs,
                "risk": risk_text(method, summary),
                "approval": "Reverse ETL writes require plan, preview, approval, execute.",
                "flags": flags,
                "api_surface": [{"method": method, "path": path}],
            }
            if notes:
                cmd["notes"] = notes
            cmds.append(cmd)
            rows.append({"method": method, "path": path, "covered_by": {"write": action_name}})

    surface["endpoints"].extend(rows)
    cli["commands"].extend(cmds)
    writes["actions"].extend(actions)
    operations["operations"].extend(ops)
    for name, payload in (("api_surface.json", surface), ("cli_surface.json", cli),
                          ("writes.json", writes), ("operations.json", operations)):
        with open(os.path.join(bundle, name), "w") as fh:
            json.dump(payload, fh, indent=2)
    print("rows %d  commands %d  actions %d  operations %d  %s"
          % (len(rows), len(cmds), len(actions), len(ops), stats))


def emit_post_read(path, op, operation_id, summary, docs, cmd_path, path_vars,
                   by_name, schema, deref, rows, cmds, ops, op_ids):
    op_id = "github." + snake(operation_id)
    ops.append({
        "id": op_id,
        "kind": "rest_read",
        "summary": summary,
        "source_url": docs,
        "risk": "low",
        "approval": "none",
        "output_policy": "json",
        # A rest_read POST must declare the body it will send; the engine
        # compiles and validates it before the request leaves.
        "rest": {
            "method": "POST",
            "path": path,
            "content_type": "application/json",
            "max_bytes": 1048576,
            "body_schema": {
                "$schema": "http://json-schema.org/draft-07/schema#",
                "type": "object",
                "required": list((schema or {}).get("required") or []),
                "properties": {
                    name: strip_schema(deref(node), deref)
                    for name, node in (((schema or {}).get("properties")) or {}).items()
                },
            },
        },
    })
    flags = []
    for var in path_vars:
        param = by_name.get(var, {})
        flags.append(make_flag(var, deref(param.get("schema")) or {"type": "string"},
                               scalar_flag_type(deref(param.get("schema")) or {"type": "string"}) or "string",
                               by_name, required=True, target_prefix="path"))
    for name, node in ((schema or {}).get("properties") or {}).items():
        kind = scalar_flag_type(deref(node))
        if kind is None:
            continue
        flags.append(make_flag(name, deref(node), kind, by_name,
                               required=name in ((schema or {}).get("required") or []),
                               target_prefix="body"))
    cmds.append({
        "path": cmd_path,
        "summary": summary,
        "intent": "direct_read",
        "availability": "implemented",
        "operation": op_id,
        "source_cli_path": "",
        "source_url": docs,
        "flags": flags,
        "api_surface": [{"method": "POST", "path": path}],
        "output_policy": "json_redacted",
    })
    rows.append({"method": "POST", "path": path, "covered_by": {"direct_read": cmd_path}})


def make_flag(target, node, kind, by_name, required, target_prefix="record"):
    param = by_name.get(target.split(".")[-1], {})
    summary = (first_sentence((node or {}).get("description"))
               or first_sentence(param.get("description"))
               or "Sets %s." % target)
    flag = {
        "name": kebab(target),
        "type": kind,
        "summary": summary,
        "maps_to": "%s.%s" % (target_prefix, target),
    }
    values = [str(v) for v in (node or {}).get("enum", []) if v is not None]
    if kind == "enum" and values:
        flag["values"] = values
    if required:
        flag["required"] = True
    return flag


def strip_schema(node, deref, depth=0):
    """Keep the published contract's shape; drop prose and vendor noise so the
    record schema stays a contract rather than a copy of the docs."""
    node = deref(node)
    if not isinstance(node, dict) or depth > 3:
        return {"type": "string"}
    out = {}
    for field in ("type", "enum", "format"):
        if field in node:
            out[field] = node[field]
    if isinstance(out.get("type"), list):
        out["type"] = next((t for t in out["type"] if t != "null"), "string")
    if out.get("type") == "object" and node.get("properties"):
        out["properties"] = {k: strip_schema(v, deref, depth + 1)
                             for k, v in node["properties"].items()}
        if node.get("required"):
            out["required"] = node["required"]
    if out.get("type") == "array":
        out["items"] = strip_schema(node.get("items") or {"type": "string"}, deref, depth + 1)
    if not out:
        out = {"type": "string"}
    return out


def render_path(path):
    def sub(match):
        name = match.group(1)
        if name in CONFIG_PATH_VARS:
            return "{{ config.%s }}" % name
        return "{{ record.%s }}" % name
    return re.sub(r"{(\w+)}", sub, path)


def risk_text(method, summary):
    if method == "DELETE":
        return "Destructive: %s. Removes provider-side state." % summary
    if method == "POST":
        return "Creates provider-side state: %s." % summary
    return "Mutates existing provider-side state: %s." % summary


if __name__ == "__main__":
    build(sys.argv[1], json.load(open(SPEC_PATH)))
