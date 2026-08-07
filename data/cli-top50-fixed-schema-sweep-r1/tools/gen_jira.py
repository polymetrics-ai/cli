#!/usr/bin/env python3
"""Build jira's whole documented surface from Atlassian's own OpenAPI document.

jira is a FROM-NOTHING connector in the same sense workday-rest was: the shipped
bundle carries 15 api_surface rows, of which 12 are comma-joined or wildcard
"and similar" families standing for hundreds of operations, and there is no
`cli_surface.json`, `writes.json` or `operations.json` at all. A wildcard row is
not an operation (finding 24), so this is a restructure, not an extension.

    python3 derive_jira.py /tmp/sweep/jira.json > DERIVED-OPERATIONS.json
    python3 gen_jira.py internal/connectors/defs/jira DERIVED-OPERATIONS.json \
        /tmp/sweep/jira.json --classify        # dry run: print the partition
    python3 gen_jira.py ... --reads
    python3 gen_jira.py ... --all

It REWRITES rather than appends, so it is re-runnable from a clean tree
(`git checkout -- internal/connectors/defs/jira/`). Every fix belongs here, never
in the emitted bundle.

The three shipped streams (issue search, project search, user search) keep their
`covered_by.stream` rows: they carry hand-authored schemas and fixtures, and
replacing them with direct reads inside a parity commit would delete shipped,
contract-backed functionality (finding 21).

Blocking is by BUILT ARTIFACT, not by inputs (finding 44), and every block names
the runtime component that refuses it. The classes, all of them consequences of
what Atlassian's spec does and does not declare:

  raw_binary_body   a request body declared `*/*` with NO schema at all -- the
                    avatar uploads. engine's write body types are
                    json/form/none/graphql/json_array/multipart/base64_upload;
                    not one of them emits a raw byte stream as the request body.
  unbounded_body    a body the spec declares as arbitrary JSON: `schema: {}` for
                    the entity-property PUTs, or an `object` with no properties
                    for the JSON-Patch plan updates. There is no bounded record
                    contract to derive and inventing one is the generic HTTP
                    write AGENTS.md forbids.
  scalar_body       a body that is a bare JSON string. buildJSONBody assembles
                    an object from record fields; `json_array` covers a top-level
                    array and nothing covers a top-level scalar.
  empty_contract    no body, no path variable, no required query parameter.
                    engine.PreflightWriteAction refuses a record_schema that
                    admits only {}, correctly: the action would plan a mutation
                    with no inputs.
"""

import argparse
import collections
import json
import os
import re
import sys

DOCS = "https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/"
ARTIFACT = "https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json"

PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults", "first", "last",
    "start", "startat", "start_at", "skip", "top", "next_page_token",
}

FLAG_TYPE = {"string": "string", "integer": "integer", "number": "number", "boolean": "boolean"}

# Credential-valued record fields never become flags: AGENTS.md requires
# credentials to arrive from environment variables or stdin, never prompt text.
CREDENTIAL_FIELDS = {"password", "token", "secret", "apiToken", "clientSecret",
                     "accessToken", "refreshToken", "sharedSecret"}

WRITE_KIND = {"POST": "create", "PUT": "update", "PATCH": "update", "DELETE": "delete"}

# Jira's 78 URL groups, folded into the product domains Atlassian's own REST
# reference uses. Presentation only -- a command's path is unaffected.
DOMAINS = [
    ("issues", "Issues, comments, worklogs and attachments", {
        "issue", "issues", "comment", "worklog", "attachment", "issue-link",
        "issue-link-type", "issuetype", "issuetypescheme", "issuetypescreenscheme",
        "changelog", "bulk", "redact", "label", "status", "statuses",
        "statuscategory", "priority", "priorityscheme", "resolution", "version",
        "component", "securitylevel", "issuesecurityschemes",
    }),
    ("projects", "Projects, versions, roles and categories", {
        "project", "projects", "project-category", "project-template",
        "projectvalidate", "role", "classification-levels", "data-policy",
    }),
    ("search", "Search, JQL and expressions", {
        "search", "jql", "expression", "filter",
    }),
    ("users", "Users, groups and permissions", {
        "user", "users", "group", "groups", "groupuserpicker", "myself",
        "mypermissions", "mypreferences", "permissions", "permissionscheme",
        "avatar", "universal-avatar",
    }),
    ("fields", "Fields, screens and configuration schemes", {
        "field", "fieldconfiguration", "fieldconfigurationscheme", "screens",
        "screenscheme", "custom-field-option", "config", "configuration",
        "settings", "ui-modifications",
    }),
    ("workflows", "Workflows and notification schemes", {
        "workflow", "workflows", "workflowscheme", "notificationscheme",
    }),
    ("dashboards", "Dashboards, plans and announcements", {
        "dashboard", "plans", "announcement-banner",
    }),
    ("platform", "Instance administration, apps and webhooks", {
        "app", "applicationrole", "application-properties", "auditing", "events",
        "instance", "license", "server-info", "task", "webhook", "connect",
        "forge", "internal",
    }),
]


def kebab(text):
    text = re.sub(r"(?<=[a-z0-9])(?=[A-Z])", "-", text)
    return re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")


def first_sentence(text, cap=180):
    if not text:
        return ""
    text = re.sub(r"<[^>]+>", " ", text)
    text = re.sub(r"\[([^\]]*)\]\([^)]*\)", r"\1", text)
    text = re.sub(r"[`*_>#]", "", text)
    text = " ".join(text.split())
    idx = text.find(". ")
    if 0 < idx < cap:
        return text[: idx + 1]
    if len(text) > cap:
        text = text[: cap - 1].rsplit(" ", 1)[0] + "."
    return text


# --------------------------------------------------------------------------
# Schema handling
# --------------------------------------------------------------------------

MAX_DEPTH = 4


def make_deref(components):
    def deref(node, depth=0, seen=frozenset()):
        """Resolve $refs with cycle and depth protection.

        Jira's component schemas are heavily self-referential (an issue holds
        fields that hold an issue), so an unbounded expansion never terminates.
        A truncated branch degrades to a plain object rather than disappearing.
        """
        if not isinstance(node, dict):
            return {}
        while "$ref" in node:
            ref = node["$ref"]
            if ref in seen or depth > MAX_DEPTH:
                return {"type": "object"}
            seen = seen | {ref}
            node = components.get(ref.split("/")[-1], {})
            if not isinstance(node, dict):
                return {"type": "object"}
        for combinator in ("allOf",):
            if combinator in node:
                merged = {"type": "object", "properties": {}, "required": []}
                for arm in node[combinator]:
                    sub = deref(arm, depth, seen)
                    merged["properties"].update(sub.get("properties") or {})
                    for req in sub.get("required") or []:
                        if req not in merged["required"]:
                            merged["required"].append(req)
                if not merged["required"]:
                    merged.pop("required")
                if not merged["properties"]:
                    merged.pop("properties")
                return merged
        out = {}
        for key in ("type", "format", "enum", "description"):
            if key in node:
                out[key] = node[key]
        if node.get("required"):
            out["required"] = list(node["required"])
        if depth >= MAX_DEPTH:
            out.setdefault("type", "object")
            out.pop("required", None)
            return {"type": out.get("type", "object")}
        if "properties" in node:
            out["properties"] = {
                name: deref(sub, depth + 1, seen) for name, sub in node["properties"].items()
            }
            out.setdefault("type", "object")
        if "items" in node:
            out["items"] = deref(node["items"], depth + 1, seen)
            out.setdefault("type", "array")
        if "additionalProperties" in node and isinstance(node["additionalProperties"], dict):
            out.setdefault("type", "object")
        if "type" not in out:
            out["type"] = "object" if "properties" in out else "string"
        # A required name with no declared property cannot be bound to a flag
        # and would make validate's recursion point at nothing.
        if out.get("required"):
            declared = out.get("properties") or {}
            out["required"] = [r for r in out["required"] if r in declared]
            if not out["required"]:
                out.pop("required")
        return out

    return deref


# --------------------------------------------------------------------------
# validate.go transcription (cmd/connectorgen/validate.go requiredMappingPaths)
# --------------------------------------------------------------------------


def required_mapping_paths(node, prefix=""):
    out = []
    for req in node.get("required", []):
        child = (node.get("properties") or {}).get(req)
        path = "%s.%s" % (prefix, req) if prefix else req
        child_paths = required_node_mapping_paths(child, path)
        out.extend(child_paths if child_paths else [path])
    return out


def required_node_mapping_paths(node, prefix):
    if not isinstance(node, dict):
        return []
    if node.get("type") == "array":
        items = node.get("items")
        if not items:
            return []
        paths = required_node_mapping_paths(items, prefix + ".0")
        return paths if paths else [prefix]
    if node.get("type") == "object":
        return required_mapping_paths(node, prefix)
    return []


def resolve_leaf(node, dotted):
    cur = node
    for part in dotted.split("."):
        if not isinstance(cur, dict):
            return None
        if part == "0":
            cur = cur.get("items")
            continue
        cur = (cur.get("properties") or {}).get(part)
    return cur


def flag_type_for(leaf):
    if not isinstance(leaf, dict):
        return None
    declared = leaf.get("type")
    if declared == "array":
        items = leaf.get("items") or {}
        return "string_array" if items.get("type") in FLAG_TYPE else None
    return FLAG_TYPE.get(declared)


# --------------------------------------------------------------------------
# Classification
# --------------------------------------------------------------------------

BLOCK_NOTES = {
    "raw_binary_body": (
        "the provider declares this request body as */* with no schema at all: it is a raw image "
        "byte stream",
        "Named dependency: engine's write body types are json, form, none, graphql, json_array, "
        "multipart and base64_upload (engine/bundle.go WriteAction.BodyType); none of them emits a "
        "raw byte stream as the request body, and accepting inline raw bytes is banned outright. "
        "Unblocks when a bounded raw-upload body type exists.",
    ),
    "unbounded_body": (
        "the provider's published spec declares this request body as arbitrary JSON with no "
        "properties, so there is no bounded record contract to derive",
        "Named dependency: connectorgen validate's checkCLISurfaceWriteFlags has no required "
        "fields to bind and engine.CompileSchema has no shape to compile, because Atlassian "
        "documents this body as an unconstrained JSON value. Deriving one would mean inventing a "
        "payload the provider never published, which is the generic HTTP write AGENTS.md forbids. "
        "Unblocks when Atlassian publishes the request schema.",
    ),
    "scalar_body": (
        "the provider declares this request body as a bare JSON string rather than an object",
        "Named dependency: engine.buildJSONBody assembles the request body as an object from the "
        "record's fields, and body_type json_array covers a top-level array; no body type emits a "
        "top-level JSON scalar. Modelling it as body_type none would send the request with the "
        "documented value silently dropped. Unblocks when a scalar body type exists.",
    ),
    "dynamic_key_map": (
        "the provider declares this request body as an open map keyed by a custom field or scheme "
        "id, whose values are objects rather than scalars",
        "Named dependency: engine's dynamic_fields region (engine/bundle.go DynamicFieldsSpec) is "
        "the one declared capability for a dynamic-key payload, and validateDynamicFields accepts "
        "scalar value_types only, so an object-valued map has no bounded record contract. "
        "Unblocks when dynamic_fields accepts a declared object value shape.",
    ),
    "unbindable_read_body": (
        "the provider declares a required request-body field for this read that has no scalar leaf",
        "Named dependency: cli_surface flag types are boolean, string, integer, number, enum and "
        "string_array (engine/schema/cli_surface.schema.json), so an object-valued required body "
        "field has no flag form; connectorgen validate's checkCLISurfaceOperationBodyMappings then "
        "refuses the 'implemented' claim, and covered_by.direct_read accepts only an implemented "
        "command. Unblocks when a structured-document flag type exists.",
    ),
    "empty_contract": (
        "the provider's published spec documents no bindable request field for this operation: no "
        "request body, no path variable and no required query parameter",
        "Named dependency: engine.PreflightWriteAction refuses a write whose record_schema admits "
        "only an empty object (engine/record_schema_promotion.go), correctly -- the action would "
        "plan a mutation with no inputs. Atlassian documents this operation's selector as optional "
        "query parameters only. Unblocks when the spec declares one required.",
    ),
}


def classify_write(op, body_media, body_schema, dynamic_key_map=False):
    """The blocked-or-covered judgement, made on the BUILT contract.

    `body_schema is None` means the provider declared a body and constrained it
    with NOTHING -- either the `schema` key is absent or it is the empty object.
    That distinction has to survive derefencing: an empty schema resolved
    through a "default to string" fallthrough looks exactly like a documented
    scalar body, and the two need different dispositions.
    """
    if body_media and body_schema is None:
        binary_media = [m for m in body_media
                        if m == "*/*" or m.startswith("multipart/") or "octet-stream" in m]
        if binary_media and "application/json" not in body_media:
            return "raw_binary_body"
        return "unbounded_body"
    if body_media and body_schema:
        declared = body_schema.get("type")
        if declared == "string":
            return "scalar_body"
        if declared == "object" and not body_schema.get("properties"):
            if dynamic_key_map:
                return "dynamic_key_map"
            return "unbounded_body"
    if not (op["path_vars"] or op["required_query"] or (body_schema or {}).get("properties")
            or (body_schema or {}).get("type") == "array"):
        return "empty_contract"
    return None


def body_schema_of(artifact_op, deref):
    """(declared media types, resolved body schema or None).

    None means the provider declared no shape at all. Jira spells that two
    ways -- an absent `schema` key on a `*/*` avatar upload, and a literal
    `"schema": {}` on the entity-property PUTs -- and both must stay
    distinguishable from a schema that genuinely says `type: string`.
    """
    content = ((artifact_op.get("requestBody") or {}).get("content") or {})
    if not content:
        return [], None, False
    pick = content.get("application/json")
    if pick is None and "*/*" in content:
        pick = content["*/*"]
    if pick is None:
        pick = list(content.values())[0]
    raw = pick.get("schema")
    if not raw:
        return sorted(content), None, False
    # An open map keyed by an id the caller supplies -- Jira's field-scheme
    # bodies are `additionalProperties: <object>` keyed by custom field id.
    # engine's dynamic_fields region is the closest declared capability and it
    # accepts SCALAR value types only, so this is its own blocked class rather
    # than "arbitrary JSON".
    dynamic_key_map = bool(
        raw.get("type") == "object"
        and not raw.get("properties")
        and isinstance(raw.get("additionalProperties"), dict)
    )
    return sorted(content), deref(raw), dynamic_key_map


# --------------------------------------------------------------------------
# Emission
# --------------------------------------------------------------------------


def read_body_flags(body_schema):
    """Flags binding a read-shaped POST's REQUIRED body fields, or the reason it
    cannot be `implemented`.

    connectorgen validate's checkCLISurfaceOperationBodyMappings requires an
    implemented operation-backed direct read to bind every required body path,
    and it uses the SAME requiredMappingPaths recursion the write side uses --
    so this mirrors the write-side flag builder rather than restating it, and
    binds the scalar LEAF: `body.queries.0.query`, not `body.queries`.
    (google-calendar already ships `body.items.0.id`, so an array-element leaf
    is a supported mapping target rather than a guess.)

    A required path with no scalar leaf -- cli_surface flag types are boolean,
    string, integer, number, enum and string_array, and none of them carries an
    object -- returns as unbindable. `covered_by.direct_read` in turn requires
    an IMPLEMENTED command, so such a row cannot be covered at all and the
    caller blocks it. That is finding 4's rule reaching its stronger conclusion
    on the read side: a write can be `partial` and still cover its row; a read
    cannot.
    """
    flags, unbindable = [], []
    for dotted in required_mapping_paths(body_schema):
        leaf = resolve_leaf(body_schema, dotted)
        if dotted.split(".")[-1] in CREDENTIAL_FIELDS:
            unbindable.append(dotted)
            continue
        ftype = flag_type_for(leaf)
        if ftype is None:
            unbindable.append(dotted)
            continue
        flag = {
            "name": kebab(dotted.replace(".", "-")),
            "type": ftype,
            "summary": first_sentence((leaf or {}).get("description"))
            or ("Request body field %s." % dotted),
            "maps_to": "body.%s" % dotted,
            "required": True,
        }
        if ftype == "string" and (leaf or {}).get("enum"):
            flag["type"] = "enum"
            flag["values"] = [str(v) for v in leaf["enum"]]
        flags.append(flag)
    return flags, unbindable


def path_flags(op):
    flags = []
    for var in op["path_vars"]:
        if var.lower() in PAGING_PARAMS:
            raise SystemExit("refusing to author paging flag %r on %s" % (var, op["key"]))
        flags.append({
            "name": kebab(var),
            "type": "string",
            "summary": "Path parameter %s." % var,
            "maps_to": "path.%s" % var,
            "required": True,
        })
    for param in op["required_query"]:
        if param["name"].lower() in PAGING_PARAMS:
            raise SystemExit("refusing to author paging flag %r on %s" % (param["name"], op["key"]))
        flags.append({
            "name": kebab(param["name"]),
            "type": FLAG_TYPE.get(param["type"], "string"),
            "summary": first_sentence(param["summary"]) or ("Query parameter %s." % param["name"]),
            "maps_to": "query.%s" % param["name"],
            "required": True,
        })
    return flags


def op_id(op):
    return "jira." + re.sub(r"[^a-z0-9]+", "_", op["command"]).strip("_")


def templated_path(op):
    path = op["path"]
    for var in op["path_vars"]:
        path = path.replace("{%s}" % var, "{{ record.%s }}" % var)
    return path


def summary_of(op):
    text = first_sentence(op["summary"]) or ("%s %s." % (op["method"], op["path"]))
    if not text.endswith("."):
        text += "."
    if op["deprecated"]:
        text += " (deprecated by Atlassian, still documented)"
    return text


def build(bundle, derived, artifact, mode):
    paths = artifact["paths"]
    deref = make_deref(artifact.get("components", {}).get("schemas", {}))

    rows, commands, operations, actions = [], [], [], []
    stats = collections.Counter()
    blocked_detail = collections.defaultdict(list)

    for op in derived["operations"]:
        artifact_op = paths[op["path"]][op["method"].lower()]
        key = op["key"]

        # ---------------- streams stay streams ----------------------------
        if op["stream"]:
            rows.append({"method": op["method"], "path": op["path"],
                         "covered_by": {"stream": op["stream"]}})
            commands.append({
                "path": "%s list" % op["stream"],
                "summary": "Read Jira %s as ETL records." % op["stream"],
                "intent": "etl",
                "availability": "implemented",
                "stream": op["stream"],
                "source_url": DOCS,
                "examples": ["pm jira %s list --json" % op["stream"]],
            })
            stats["stream"] += 1
            continue

        # ---------------- binary downloads (GET only) ---------------------
        if op["binary"]:
            oid = op_id(op)
            operations.append({
                "id": oid,
                "kind": "binary_download",
                "summary": summary_of(op),
                "source_url": DOCS,
                "risk": "low",
                "approval": "none",
                "output_policy": "binary_file_bounded",
                "binary": {
                    "method": "GET",
                    "path": op["path"],
                    "max_bytes": 10485760,
                    "allow_overwrite": False,
                    "extract_archives": False,
                },
            })
            commands.append({
                "path": op["command"],
                "summary": summary_of(op),
                "intent": "binary_download",
                "availability": "implemented",
                "operation": oid,
                "source_url": DOCS,
                "flags": path_flags(op),
                "api_surface": [{"method": "GET", "path": op["path"]}],
            })
            rows.append({"method": op["method"], "path": op["path"],
                         "covered_by": {"direct_reads": [op["command"]]}})
            stats["binary_download"] += 1
            continue

        # ---------------- plain GET reads ---------------------------------
        if op["method"] == "GET":
            commands.append({
                "path": op["command"],
                "summary": summary_of(op),
                "intent": "direct_read",
                "availability": "implemented",
                "source_url": DOCS,
                "flags": path_flags(op),
                "api_surface": [{"method": "GET", "path": op["path"]}],
                "output_policy": "json_redacted",
            })
            rows.append({"method": op["method"], "path": op["path"],
                         "covered_by": {"direct_reads": [op["command"]]}})
            stats["direct_read"] += 1
            continue

        media, schema, dynamic_key_map = body_schema_of(artifact_op, deref)

        # ---------------- read-shaped POSTs -------------------------------
        if op["read"]:
            if "application/json" not in media:
                raise SystemExit("%s: read-shaped POST has no application/json body" % key)
            if not schema:
                raise SystemExit("%s: read-shaped POST has no body_schema" % key)
            oid = op_id(op)
            operations.append({
                "id": oid,
                "kind": "rest_read",
                "summary": summary_of(op),
                "source_url": DOCS,
                "risk": "low",
                "approval": "none",
                "output_policy": "json_redacted",
                "rest": {
                    "method": op["method"],
                    "path": op["path"],
                    "content_type": "application/json",
                    "max_bytes": 1048576,
                    "body_schema": schema,
                },
            })
            body_flags, unbindable = read_body_flags(schema)
            if unbindable:
                reason, note = BLOCK_NOTES["unbindable_read_body"]
                rows.append({
                    "method": op["method"],
                    "path": op["path"],
                    "operation": {
                        "model": "direct_read",
                        "status": "blocked",
                        "risk": "low",
                        "blocked_by_default": True,
                        "reason": reason + " (" + ", ".join(sorted(unbindable)) + ")",
                        "source_url": ARTIFACT,
                        "notes": note,
                    },
                })
                operations.pop()
                stats["blocked_unbindable_read_body"] += 1
                blocked_detail["unbindable_read_body"].append(key)
                continue
            command = {
                "path": op["command"],
                "summary": summary_of(op),
                "intent": "direct_read",
                "availability": "implemented",
                "operation": oid,
                "source_url": DOCS,
                "flags": path_flags(op) + body_flags,
                "api_surface": [{"method": op["method"], "path": op["path"]}],
                "output_policy": "json_redacted",
            }
            commands.append(command)
            rows.append({"method": op["method"], "path": op["path"],
                         "covered_by": {"direct_reads": [op["command"]]}})
            stats["post_shaped_read"] += 1
            continue

        # ---------------- writes ------------------------------------------
        blocked = classify_write(op, media, schema, dynamic_key_map)
        if blocked:
            reason, note = BLOCK_NOTES[blocked]
            rows.append({
                "method": op["method"],
                "path": op["path"],
                "operation": {
                    "model": "deprecated" if op["deprecated"] else "sensitive_reverse_etl",
                    "status": "blocked",
                    "risk": "medium",
                    "blocked_by_default": True,
                    "reason": reason,
                    "source_url": ARTIFACT,
                    "notes": note,
                },
            })
            stats["blocked_" + blocked] += 1
            blocked_detail[blocked].append(key)
            continue

        name = re.sub(r"[^a-z0-9]+", "_", op["command"]).strip("_")
        record = {
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "required": list(op["path_vars"]),
            "properties": {v: {"type": "string"} for v in op["path_vars"]},
        }
        for param in op["required_query"]:
            record["properties"].setdefault(
                param["name"], {"type": FLAG_TYPE.get(param["type"], "string")})
            if param["name"] not in record["required"]:
                record["required"].append(param["name"])

        action = {
            "name": name,
            "kind": WRITE_KIND[op["method"]],
            "method": op["method"],
            "path": templated_path(op),
            "record_schema": record,
            "risk": "Jira %s on %s; changes site data." % (op["method"], op["path"]),
        }
        if op["path_vars"]:
            action["path_fields"] = list(op["path_vars"])
        if op["required_query"]:
            action["query"] = {
                p["name"]: {"template": "{{ record.%s }}" % p["name"]} for p in op["required_query"]
            }

        if schema and schema.get("type") == "array":
            # A top-level array body: the record carries it in one named field.
            field = "items"
            record["properties"][field] = schema
            record["required"].append(field)
            action["body_type"] = "json_array"
            action["body_field"] = field
            action["body_schema"] = schema
            stats["write_json_array"] += 1
        elif schema and schema.get("properties"):
            for prop, sub in schema["properties"].items():
                record["properties"].setdefault(prop, sub)
            for req in schema.get("required") or []:
                if req not in record["required"]:
                    record["required"].append(req)
            action["body_type"] = "json"
            action["body_fields"] = sorted(schema["properties"])
        else:
            action["body_type"] = "none"

        if op["method"] == "DELETE":
            action["delete"] = {"idempotent": True, "missing_ok_status": [404]}
            action["confirm"] = "destructive"

        if not record["required"]:
            record.pop("required")
        if not record["properties"]:
            raise SystemExit("%s: built record_schema has no properties" % key)

        flags, unbindable, redacted = [], [], []
        for dotted in required_mapping_paths(record):
            leaf = resolve_leaf(record, dotted)
            if dotted.split(".")[-1] in CREDENTIAL_FIELDS:
                redacted.append(dotted)
                unbindable.append(dotted)
                continue
            ftype = flag_type_for(leaf)
            if ftype is None:
                unbindable.append(dotted)
                continue
            flag = {
                "name": kebab(dotted.replace(".", "-")),
                "type": ftype,
                "summary": first_sentence((leaf or {}).get("description"))
                or ("Record field %s." % dotted),
                "maps_to": "record.%s" % dotted,
                "required": True,
            }
            if ftype == "string" and (leaf or {}).get("enum"):
                flag["type"] = "enum"
                flag["values"] = [str(v) for v in leaf["enum"]]
            flags.append(flag)
        if redacted:
            action["redact_fields"] = sorted(set(redacted))

        actions.append(action)
        command = {
            "path": op["command"],
            "summary": summary_of(op),
            "intent": "reverse_etl",
            "availability": "partial" if unbindable else "implemented",
            "write": name,
            "source_url": DOCS,
            "risk": action["risk"],
            "approval": "Reverse ETL writes require plan, preview, approval, execute.",
            "flags": flags,
        }
        if unbindable:
            command["notes"] = (
                "availability: partial -- Atlassian declares required record field(s) %s with no "
                "scalar leaf (or carrying a credential), so no flag can carry them and the runtime "
                "would refuse an 'implemented' claim. Supply them with --record instead."
                % ", ".join(sorted(unbindable))
            )
            stats["write_partial"] += 1
        else:
            stats["write_implemented"] += 1
        commands.append(command)
        rows.append({"method": op["method"], "path": op["path"],
                     "covered_by": {"writes": [name]}})

    if mode == "classify":
        print(json.dumps(dict(sorted(stats.items())), indent=2))
        covered = sum(v for k, v in stats.items()
                      if not k.startswith("blocked_") and k != "write_json_array")
        blocked = sum(v for k, v in stats.items() if k.startswith("blocked_"))
        print("covered=%d blocked=%d total=%d" % (covered, blocked, covered + blocked))
        for cls, keys in sorted(blocked_detail.items()):
            print("== %s (%d)" % (cls, len(keys)))
            for k in keys:
                print("   " + k)
        return

    surface_path = os.path.join(bundle, "api_surface.json")
    surface = json.load(open(surface_path))
    surface["api"] = "Jira Cloud platform REST API v3 -- the full documented operation surface"
    surface["docs"] = "https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/"
    surface["reviewed_at"] = "2026-08-07"
    surface["scope"] = (
        "cli-top50-fixed-schema-sweep-r1: every operation Atlassian's own OpenAPI description "
        "declares under `paths` -- %d operations across %d path keys, counted one per "
        "(method, path). The shipped bundle's 15 rows were 12 comma-joined and wildcard families "
        "standing for hundreds of endpoints; a wildcard row is not an operation, so they are "
        "replaced rather than extended. Artifact sha256 %s (info.version %s): Atlassian serves a "
        "rolling SNAPSHOT at a single unpinned URL, so this sha, not a byte count, is what "
        "identifies the document this surface was derived from."
        % (derived["operations_total"], derived["paths_count"],
           derived["artifact_sha256"], derived["info_version"])
    )
    surface["operation_ledger_version"] = 1
    surface["endpoints"] = rows
    write_json(surface_path, surface)

    write_json(os.path.join(bundle, "cli_surface.json"), build_cli_surface(commands))
    write_json(os.path.join(bundle, "operations.json"), {"operations": operations})
    write_json(os.path.join(bundle, "writes.json"), {"actions": actions})

    meta_path = os.path.join(bundle, "metadata.json")
    meta = json.load(open(meta_path))
    meta["capabilities"]["write"] = True
    meta["description"] = (
        "Reads and writes the full documented Jira Cloud platform REST API v3 surface: %d "
        "operations across issues, projects, users, fields, workflows, dashboards and instance "
        "administration." % derived["operations_total"]
    )
    meta["risk"]["approval"] = (
        "reverse ETL writes require plan, preview, approval, execute; reads are HTTP Basic "
        "authenticated with an Atlassian API token"
    )
    write_json(meta_path, meta)

    print("api_surface rows: %d" % len(rows))
    print("commands: %d  operations: %d  write actions: %d" % (len(commands), len(operations), len(actions)))
    print("stats: %s" % dict(sorted(stats.items())))


def build_cli_surface(commands):
    present = {c["path"].split(" ", 1)[0] for c in commands}
    known = set()
    for _, _, members in DOMAINS:
        known |= members
    known |= {"issues", "projects", "users"}
    missing = present - known
    if missing:
        raise SystemExit("URL group(s) not assigned to a domain: %s" % sorted(missing))
    return {
        "tagline": "Read Jira issues, projects and users, run bounded reads, and plan typed Jira mutations.",
        "usage": "pm jira <group> <action> [flags]",
        "source_cli": {
            "name": "Jira Cloud platform REST API v3",
            "docs": "https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/",
            "reference": "Atlassian's own OpenAPI description, " + ARTIFACT,
            "source": "provider_api",
        },
        "groups": [
            {"id": "etl", "title": "ETL streams", "commands": ["issues", "projects", "users"]},
        ] + [
            {"id": gid, "title": title,
             "commands": sorted(m for m in members if m in present)}
            for gid, title, members in DOMAINS
        ],
        "global_flags": [
            {"name": "json", "type": "boolean", "summary": "Write machine-readable JSON output."},
            {"name": "connection", "type": "string",
             "summary": "Use a saved Jira connector credential and site base URL.",
             "maps_to": "connection"},
        ],
        "commands": commands,
        "help_topics": [
            {"name": "authentication",
             "summary": "Use pm credentials to store the Atlassian account email and API token. "
                        "Never print stored tokens."},
            {"name": "execution-model",
             "summary": "ETL commands map to streams. Bounded direct reads map to documented GET "
                        "and read-shaped POST operations. Reverse ETL commands map to approved "
                        "write actions and keep plan, preview, approval, execute."},
        ],
    }


def write_json(path, payload):
    with open(path, "w") as fh:
        json.dump(payload, fh, indent=2)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("bundle")
    parser.add_argument("derived")
    parser.add_argument("artifact")
    parser.add_argument("mode", choices=["classify", "all"])
    args = parser.parse_args()
    build(args.bundle, json.load(open(args.derived)), json.load(open(args.artifact)), args.mode)


if __name__ == "__main__":
    main()
