#!/usr/bin/env python3
"""Expand workday-rest's bundle to its full documented operation surface.

Source of truth is Workday's own service DIRECTORY -- not one spec. The manifest
at services2026.30.json (HTTP 200, 617,538 bytes) names 52 independently
versioned production services, each with its own spec file and its own mount
point. An operation's identity is therefore the resolved (method, base+path)
pair; a service-relative path is ambiguous across 52 modules.

    920 raw rows
     -4  the same endpoint published by two service modules (Custom Object Data
         single- and multi-instance v2 declare the IDENTICAL servers URL)
     -9  query-string variants of an endpoint already counted
    =907 documented operations

Run it against a CLEAN bundle -- it rewrites rather than appends, so it is
re-runnable and never hand-patched:

    python3 gen_workday_rest.py internal/connectors/defs/workday-rest --reads
    python3 gen_workday_rest.py internal/connectors/defs/workday-rest --writes

The four non-mechanical judgements, each recorded in the phase SUMMARY:

  read-vs-write     every documented GET is a read. Non-GETs are read from
                    their summary, not assumed: Workday is read-heavy and the
                    sweep has already found POST-shaped reads in github.
  stream-vs-direct  new GETs become plain direct_read commands, not streams. A
                    stream needs a hand-authored record schema, primary key and
                    fixture; inventing 648 of those would be inventing data
                    contracts Workday never published. The three shipped legacy
                    streams stay streams -- see the legacy note below.
  binary detection  an operation is binary iff its documented `produces`
                    declares application/octet-stream. Read out of the artifact,
                    NEVER guessed from the path: ?type=viewContent and
                    ?type=viewFile sound like downloads and declare JSON only.
  named-dependency  a row is blocked only when a named runtime component
                    refuses it, and the note names that component.
"""

import collections
import json
import os
import re
import sys
from urllib.parse import urlparse

ROOT = os.path.abspath(os.path.dirname(__file__))
REPO = os.path.abspath(os.path.join(ROOT, "..", "..", ".."))
DERIVED = os.path.join(
    REPO, ".planning", "phases", "workday-rest-parity-sweep-r1", "DERIVED-OPERATIONS.json"
)

# Path variables the connection spec supplies, so they never become flags.
# spec.json declares `tenant` required, and Prism Analytics is the only service
# whose mount point carries it (PLAN.md hazard 3).
CONFIG_PATH_VARS = {"tenant"}

# Never hand-authored. streams.json already declares this connector's paging
# (page_number, page/limit); the foundation lane derives flags from it.
PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults", "first", "last",
    "count", "start", "skip", "top",
}

DOCS = "https://community.workday.com/sites/default/files/file-hosting/restapi/"

# The bundle's four shipped rows point at /ccx/api/hcm/v1/{tenant}/... . The
# current directory publishes no `hcm` service and no /ccx/ path, and they are
# not in the archived list either, so they are NOT among the 907.
#
# They are dispositioned rather than deleted, and NOT re-pointed. Re-pointing
# `workers` at staffing/v7/workers would keep the stream name while silently
# changing the response shape out from under schemas/workers.json and its
# fixture -- shipping a stream whose declared contract no longer matches what
# the endpoint returns. Keeping them costs nothing and breaks nothing: tenants
# with the legacy HCM API enabled still reach them.
LEGACY_PREFIX = "/ccx/api/hcm/v1/{tenant}/"
LEGACY_STREAMS = {"workers": "workers", "organizations": "organizations", "jobs": "jobs"}
LEGACY_SUPERSEDED_BY = {
    "workers": "staffing/v7 /workers",
    "organizations": "api/common/v1 /organizations",
    "jobs": "staffing/v7 /jobs",
}


def kebab(text):
    text = re.sub(r"(?<=[a-z0-9])(?=[A-Z])", "-", text)
    return re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")


def first_sentence(text, cap=160):
    if not text:
        return ""
    text = re.sub(r"<[^>]+>", " ", text)          # Workday summaries carry <b> markup
    text = text.replace("\\~", "").replace("~", "")
    text = re.sub(r"[`*_>#]", "", text)
    text = " ".join(text.split())
    idx = text.find(". ")
    if 0 < idx < cap:
        return text[: idx + 1]
    if len(text) > cap:
        text = text[: cap - 1].rsplit(" ", 1)[0] + "."
    return text


def group_of(op):
    parts = [
        p
        for p in op["base"].strip("/").split("/")
        if p != "api" and not re.fullmatch(r"v\d+", p) and not p.startswith("{")
    ]
    return kebab(parts[0])


VERB = {"POST": "create", "PUT": "replace", "PATCH": "update", "DELETE": "delete"}


def action_of(op):
    """Command action derived from the endpoint itself.

    Workday declares an operationId on only 21 of 920 rows, so the name has to
    come from the path. Sweep finding 23: name the action for the ENDPOINT.

    A trailing {var} marks a detail operation; a non-trailing {var} becomes a
    `by-<var>` word so path structure is preserved. Verified to yield 907
    distinct names for 907 operations -- zero collisions.
    """
    segs = [s for s in op["path"].strip("/").split("/") if s]
    words = []
    for i, seg in enumerate(segs):
        if seg.startswith("{"):
            if i != len(segs) - 1:
                words.append("by-" + kebab(seg.strip("{}")))
        else:
            words.append(kebab(seg))
    if not words:
        words = ["root"]
    detail = bool(segs) and segs[-1].startswith("{")
    if op["method"] == "GET":
        return "-".join(words) + ("-get" if detail else "-list")
    # A POST Workday documents as a read must not be named `create-*`. The verb
    # is the one thing a user reads before running a command, and `wql
    # create-data` on an endpoint Workday itself calls "the read-only POST
    # request" would be a lie that the surface test cannot catch.
    if op["full"] in POST_SHAPED_READS:
        return "read-" + "-".join(words)
    return VERB[op["method"]] + "-" + "-".join(words)


def command_path(op):
    return "%s %s" % (group_of(op), action_of(op))


def path_flags(op, extra=()):
    """One flag per path variable the connection spec does not supply, plus any
    explicitly supplied extras. Optional query filters are not authored: they
    are not operations. Paging flags are refused outright."""
    flags = []
    for var in re.findall(r"{(\w+)}", op["full"]):
        if var in CONFIG_PATH_VARS:
            continue
        if var.lower() in PAGING_PARAMS:
            raise SystemExit("refusing to author paging flag %r on %s" % (var, op["full"]))
        flags.append(
            {
                "name": kebab(var),
                "type": "string",
                "summary": "Path parameter %s." % var,
                "maps_to": "path.%s" % var,
                "required": True,
            }
        )
    for flag in extra:
        if flag["name"].lower().replace("-", "_") in PAGING_PARAMS:
            raise SystemExit("refusing to author paging flag %r" % flag["name"])
        flags.append(flag)
    return flags


# --------------------------------------------------------------------------
# Binary and collapsed-behaviour tables, all read from `produces` in the specs.
# --------------------------------------------------------------------------

# Endpoints whose documented success response declares application/octet-stream.
# `dual` means the SAME endpoint also declares application/json, so it needs a
# metadata read AND a download -- covered_by.direct_reads (plural) carries both
# off one documented row rather than inventing a variant path.
BINARY = {
    "/api/prismAnalytics/v3/{tenant}/buckets/{id}/errorFile": "only",
    "/attachments/v1/graphql/{ID}": "dual",
    "/customerAccounts/v1/invoicePDFs/{ID}": "dual",
    "/procurement/v5/requisitions/{ID}/attachments": "dual",
    "/procurement/v5/requisitions/{ID}/attachments/{subresourceID}": "dual",
}

# A collapsed query-string variant's behaviour, re-expressed as a flag on the
# surviving command (sweep finding 23 -- help-scout's --async pattern). Never a
# second path row.
COLLAPSED_FLAG = {
    "GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments": "viewContent",
    "GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments/{subresourceID}": "viewContent",
    "GET /recruiting/v4/prospects/{ID}/resumeAttachments": "viewFile",
    "GET /recruiting/v4/prospects/{ID}/resumeAttachments/{subresourceID}": "viewFile",
    "PATCH /staffing/v7/workers/{ID}/checkInTopics/{subresourceID}": "archive",
    "PATCH /staffing/v7/workers/{ID}/checkIns/{subresourceID}": "archive",
    "POST /api/common/v1/workers/{ID}/businessTitleChanges": "me",
}


def type_flag(value, note):
    # Not `required`: absent means the endpoint's documented default mode, which
    # is what the surviving row models. help-scout's --async, same shape.
    return {
        "name": "type",
        "type": "enum",
        "values": [value],
        "summary": note,
        "maps_to": "query.type",
    }


def load_ops():
    with open(DERIVED) as fh:
        derived = json.load(fh)
    ops = derived["operations"]
    if len(ops) != 907:
        raise SystemExit("expected 907 derived operations, got %d" % len(ops))
    return ops, derived


def legacy_rows():
    """The four shipped /ccx/ rows, each dispositioned deliberately."""
    rows = []
    for name in ("jobs", "organizations", "workers"):
        rows.append(
            {
                "method": "GET",
                "path": LEGACY_PREFIX + name,
                "covered_by": {"stream": LEGACY_STREAMS[name]},
            }
        )
    rows.append(
        {
            "method": "POST",
            "path": LEGACY_PREFIX + "workers",
            "operation": {
                "model": "deprecated",
                "status": "blocked",
                "risk": "low",
                "blocked_by_default": True,
                "reason": (
                    "legacy Workday HCM REST v1 shape: the current service directory publishes no "
                    "'hcm' service and no /ccx/ path, and this row is not in the archived list "
                    "either. Worker mutation is documented by staffing/v7, which this bundle now "
                    "enumerates."
                ),
                "source_url": DOCS,
                "notes": (
                    "Named dependency: superseded by staffing/v7 /workers, enumerated in this same "
                    "surface. Kept rather than deleted so the three shipped, schema- and "
                    "fixture-backed legacy read streams beside it are not removed inside a parity "
                    "commit (sweep finding 21)."
                ),
            },
        }
    )
    return rows


def build_reads(ops):
    rows, cmds, operations = [], [], []
    stats = {"direct_read": 0, "binary_dual": 0, "binary_only": 0, "collapsed_flag": 0}
    for op in ops:
        if op["method"] != "GET":
            continue
        key = "GET " + op["full"]
        summary = first_sentence(op["summary"]) or ("Reads %s." % op["full"])
        if op.get("deprecated"):
            summary += " (deprecated by Workday, still documented)"
        base_cmd = command_path(op)
        binary = BINARY.get(op["full"])
        extra = []
        collapsed = COLLAPSED_FLAG.get(key)
        if collapsed:
            extra.append(
                type_flag(
                    collapsed,
                    "Request the '%s' mode Workday documents on this endpoint as a separate "
                    "query-string page." % collapsed,
                )
            )
            stats["collapsed_flag"] += 1

        if binary == "only":
            op_id = "workday_rest." + base_cmd.replace(" ", "_").replace("-", "_")
            operations.append(binary_operation(op_id, summary, op))
            cmds.append(
                {
                    "path": base_cmd,
                    "summary": summary,
                    "intent": "binary_download",
                    "availability": "implemented",
                    "operation": op_id,
                    "source_cli_path": "",
                    "flags": path_flags(op),
                    "api_surface": [{"method": "GET", "path": op["full"]}],
                }
            )
            rows.append(
                {"method": "GET", "path": op["full"], "covered_by": {"direct_reads": [base_cmd]}}
            )
            stats["binary_only"] += 1
            continue

        read_cmd = {
            "path": base_cmd,
            "summary": summary,
            "intent": "direct_read",
            "availability": "implemented",
            "source_cli_path": "",
            "source_url": DOCS,
            "flags": path_flags(op, extra),
            "api_surface": [{"method": "GET", "path": op["full"]}],
            "output_policy": "json_redacted",
        }
        cmds.append(read_cmd)

        if binary == "dual":
            dl_cmd = base_cmd + "-download"
            op_id = "workday_rest." + dl_cmd.replace(" ", "_").replace("-", "_")
            dl_summary = "Downloads the file content of %s." % op["full"]
            operations.append(binary_operation(op_id, dl_summary, op))
            cmds.append(
                {
                    "path": dl_cmd,
                    "summary": dl_summary,
                    "intent": "binary_download",
                    "availability": "implemented",
                    "operation": op_id,
                    "source_cli_path": "",
                    "flags": path_flags(op),
                    "api_surface": [{"method": "GET", "path": op["full"]}],
                }
            )
            rows.append(
                {
                    "method": "GET",
                    "path": op["full"],
                    "covered_by": {"direct_reads": [base_cmd, dl_cmd]},
                }
            )
            stats["binary_dual"] += 1
            continue

        rows.append(
            {"method": "GET", "path": op["full"], "covered_by": {"direct_reads": [base_cmd]}}
        )
        stats["direct_read"] += 1
    return rows, cmds, operations, stats


def binary_operation(op_id, summary, op):
    return {
        "id": op_id,
        "kind": "binary_download",
        "summary": summary,
        "source_url": DOCS,
        "risk": "medium",
        "approval": "none",
        "output_policy": "binary_file_bounded",
        "binary": {
            "method": "GET",
            "path": op["full"],
            "max_bytes": 104857600,
            "allow_overwrite": False,
            "extract_archives": False,
        },
    }


# Workday's own product domains. Every group below is a service module the
# directory manifest names; the grouping is presentational only and does not
# affect a command's path.
GROUPS = [
    ("hcm", "Human Capital Management", [
        "staffing", "person", "compensation", "absence-management", "time-tracking",
        "performance-enablement", "talent-management", "recruiting", "learning",
        "payroll", "global-payroll", "journeys", "skill", "holiday",
        "benefit-enrollment-event-offerings", "benefit-partner",
    ]),
    ("financials", "Financials", [
        "accounts-payable", "customer-accounts", "expense", "procurement", "projects",
        "revenue", "budgets", "core-accounting", "contract-compliance", "fin-tax-public",
        "worktag", "asor",
    ]),
    ("student", "Student", [
        "student-core", "student-academic-foundation", "student-curriculum",
        "student-recruiting", "student-advising", "student-engagement", "student-finance",
    ]),
    ("platform", "Platform And Extensibility", [
        "common", "business-process", "custom-object", "custom-object-definition",
        "prism-analytics", "wql", "graph", "attachments", "connect", "communications",
        "help-article", "help-case", "request", "privacy", "system-metrics",
        "o-auth-client",
    ]),
]


def build_cli_surface(cmds):
    known = {g for _, _, members in GROUPS for g in members}
    present = {c["path"].split(" ", 1)[0] for c in cmds}
    missing = present - known
    if missing:
        raise SystemExit("service group(s) not assigned to a domain: %s" % sorted(missing))
    return {
        "tagline": "Read and write Workday HCM, Financials, Student and Platform data.",
        "usage": "pm workday-rest <service> <action> [flags]",
        "groups": [
            {
                "id": gid,
                "title": title,
                "commands": [m for m in members if m in present],
            }
            for gid, title, members in GROUPS
        ],
        "global_flags": [
            {"name": "json", "type": "boolean", "summary": "Write machine-readable JSON output."},
            {
                "name": "connection",
                "type": "string",
                "summary": "Use a saved Workday REST connector credential and tenant.",
                "maps_to": "connection",
            },
        ],
        "commands": cmds,
        "help_topics": [
            {
                "name": "authentication",
                "summary": (
                    "Use pm credentials to store the Workday tenant and bearer access token. "
                    "Never print stored tokens."
                ),
            },
            {
                "name": "services",
                "summary": (
                    "Workday publishes 52 independently versioned services, each mounted at its "
                    "own base path. A command names its service first, then the action."
                ),
            },
            {
                "name": "execution-model",
                "summary": (
                    "ETL commands map to streams. Reverse ETL commands map to approved write "
                    "actions and keep plan, preview, approval, execute."
                ),
            },
        ],
    }


def write_json(path, payload):
    with open(path, "w") as fh:
        json.dump(payload, fh, indent=2)


def main():
    bundle = sys.argv[1]
    mode = sys.argv[2] if len(sys.argv) > 2 else "--reads"
    ops, derived = load_ops()

    surface_path = os.path.join(bundle, "api_surface.json")
    with open(surface_path) as fh:
        surface = json.load(fh)

    rows, cmds, operations, stats = build_reads(ops)
    actions = []
    if mode == "--all":
        spec_dir = sys.argv[3]
        specs = load_specs(spec_dir, os.path.join(os.path.dirname(spec_dir), "services.json"))
        mrows, mcmds, actions, mops, mstats = build_mutations(ops, specs)
        rows += mrows
        cmds += mcmds
        operations += mops
        stats.update(mstats)
    elif mode != "--reads":
        raise SystemExit("mode must be --reads or --all <spec-dir>")

    surface["api"] = (
        "Workday REST API -- the full documented surface of all 52 independently versioned "
        "production services named by Workday's own service directory manifest"
    )
    surface["docs"] = DOCS + "services2026.30.json"
    surface["reviewed_at"] = "2026-08-07"
    surface["scope"] = (
        "cli-top50-fixed-schema-sweep-r1: every operation the service directory documents. 920 raw "
        "rows across 52 service specs, minus 4 published by two service modules at the identical "
        "servers URL, minus 9 query-string variants of an endpoint already counted, = 907. The four "
        "shipped /ccx/api/hcm/v1/{tenant}/ rows are a superseded legacy HCM shape the current "
        "directory does not publish; they are carried with their own dispositions and counted apart."
    )
    surface["operation_ledger_version"] = 1
    surface["endpoints"] = legacy_rows() + rows
    write_json(surface_path, surface)

    cli_path = os.path.join(bundle, "cli_surface.json")
    write_json(cli_path, build_cli_surface(cmds))

    ops_path = os.path.join(bundle, "operations.json")
    write_json(ops_path, {"operations": operations})

    if actions:
        write_json(os.path.join(bundle, "writes.json"), {"actions": actions})
        meta_path = os.path.join(bundle, "metadata.json")
        with open(meta_path) as fh:
            meta = json.load(fh)
        meta["capabilities"]["write"] = True
        meta["description"] = (
            "Reads and writes the full documented Workday REST surface: 907 operations across the "
            "52 independently versioned services Workday's own directory publishes."
        )
        meta["risk"]["read"] = (
            "external Workday REST API read across HCM, Financials, Student and Platform services "
            "(HR/PII-adjacent)"
        )
        meta["risk"]["approval"] = (
            "reverse ETL writes require plan, preview, approval, execute; reads are bearer-token "
            "authenticated"
        )
        write_json(meta_path, meta)

    print("api_surface rows: %d (%d documented + %d legacy)" % (len(surface["endpoints"]), len(rows), 4))
    print("commands: %d  operations: %d  writes: %d" % (len(cmds), len(operations), len(actions)))
    print("stats: %s" % stats)




# ==========================================================================
# Mutations
# ==========================================================================

SPEC_DIR = None  # set from --specs

# POST endpoints Workday documents as READS, not writes. This is the
# read-vs-write judgement and it is made per operation from the provider's own
# summary, never from the method. /wql/v1/data is decisive: Workday itself calls
# it "the read-only POST request". The rest validate, calculate against a
# hypothetical, check a permission, or look an ID up -- none persists anything.
POST_SHAPED_READS = {
    "/globalPayroll/v1/authorizations",
    "/person/v4/phoneValidation",
    "/skill/v1/mlSkills",
    "/studentAdvising/v1/hypotheticalAcademicProgress",
    "/studentRecruiting/v1/academicRequirementEvaluation",
    "/worktag/v1/validateWorktags",
    "/wql/v1/data",
}

WRITE_KIND = {"POST": "create", "PUT": "upsert", "PATCH": "update", "DELETE": "delete"}


def load_specs(spec_dir, manifest):
    """(resolved endpoint -> (spec, operation)) across all 52 service specs."""
    with open(manifest) as fh:
        services = json.load(fh)["productionConfidenceLevel"]
    out = {}
    for svc in services:
        with open(os.path.join(spec_dir, svc["specFilePath"])) as fh:
            spec = json.load(fh)
        oas3 = "basePath" not in spec
        base = urlparse(spec["servers"][0]["url"]).path if oas3 else spec["basePath"]
        for path, item in spec["paths"].items():
            if "?" in path:
                continue
            for method, op in item.items():
                if method.lower() not in ("get", "post", "put", "patch", "delete"):
                    continue
                out.setdefault((method.upper(), base + path), (spec, op, oas3))
    return out


def deref(spec, node, depth=0, seen=None):
    """Resolve $refs with cycle and depth protection. Workday's schemas are
    deeply self-referential (a worker references an organization that
    references a worker), so an unbounded expansion never terminates."""
    seen = seen or set()
    if not isinstance(node, dict):
        return node
    while "$ref" in node:
        ref = node["$ref"]
        if ref in seen or depth > 6:
            return {"type": "object"}
        seen = seen | {ref}
        cur = spec
        for part in ref.lstrip("#/").split("/"):
            cur = cur.get(part, {})
        node = cur
        if not isinstance(node, dict):
            return {"type": "object"}
    if depth > 6:
        return {"type": node.get("type", "object")}
    out = {}
    for key in ("type", "format", "enum", "description"):
        if key in node:
            out[key] = node[key]
    if node.get("required"):
        out["required"] = list(node["required"])
    if "properties" in node:
        out["properties"] = {
            name: deref(spec, sub, depth + 1, seen) for name, sub in node["properties"].items()
        }
        out.setdefault("type", "object")
    if "items" in node:
        out["items"] = deref(spec, node["items"], depth + 1, seen)
        out.setdefault("type", "array")
    out.setdefault("type", "object" if "properties" in out else out.get("type", "string"))
    return out


def body_schema_for(spec, op, oas3):
    if oas3:
        rb = op.get("requestBody") or {}
        content = rb.get("content") or {}
        if not content:
            return None
        schema = list(content.values())[0].get("schema")
    else:
        params = [p for p in (op.get("parameters") or []) if p.get("in") == "body"]
        schema = params[0].get("schema") if params else None
    if not schema:
        return None
    return deref(spec, schema)


def effective_types(node):
    t = node.get("type")
    return [t] if t else ["any"]


def required_mapping_paths(node, prefix=""):
    """Mirrors cmd/connectorgen/validate.go requiredMappingPaths exactly. A rule
    restated by hand drifts (the sweep has that scar); this is a transcription,
    and the generator's output is checked by `connectorgen validate` itself."""
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


FLAG_TYPE = {"string": "string", "integer": "integer", "number": "number", "boolean": "boolean"}


def flag_type_for(leaf):
    """None when the leaf has no scalar form -- which makes the command
    `partial`, never `implemented` (sweep finding 4)."""
    if not isinstance(leaf, dict):
        return None
    t = leaf.get("type")
    if t == "array":
        items = leaf.get("items") or {}
        return "string_array" if items.get("type") in FLAG_TYPE else None
    return FLAG_TYPE.get(t)


def action_name(op):
    return re.sub(r"[^a-z0-9]+", "_", command_path(op).replace(" ", "_")).strip("_")


def templated_path(op):
    path = op["full"]
    for var in re.findall(r"{(\w+)}", path):
        if var in CONFIG_PATH_VARS:
            path = path.replace("{%s}" % var, "{{ config.%s }}" % var)
        else:
            path = path.replace("{%s}" % var, "{{ record.%s }}" % var)
    return path


def build_mutations(ops, specs):
    """Write actions, POST-shaped reads, and their commands."""
    rows, cmds, actions, operations = [], [], [], []
    stats = collections.Counter()
    for op in ops:
        if op["method"] == "GET":
            continue
        key = (op["method"], op["full"])
        spec, sop, oas3 = specs[key]
        summary = first_sentence(op["summary"]) or ("%s %s." % (op["method"], op["full"]))
        if op.get("deprecated"):
            summary += " (deprecated by Workday, still documented)"
        cmd_path = command_path(op)
        schema = body_schema_for(spec, sop, oas3)
        collapsed = COLLAPSED_FLAG.get(op["method"] + " " + op["full"])

        # ---- POST-shaped reads: documented as retrieval, not mutation. -----
        if op["full"] in POST_SHAPED_READS:
            op_id = "workday_rest." + action_name(op)
            rest = {
                "method": op["method"],
                "path": op["full"],
                "content_type": "application/json",
                "max_bytes": 1048576,
                # Both fields are required by validate, and they were caught one
                # run apart on github (sweep finding 33).
                "body_schema": schema or {"type": "object"},
            }
            operations.append({
                "id": op_id,
                "kind": "rest_read",
                "summary": summary,
                "source_url": DOCS,
                "risk": "low",
                "approval": "none",
                "output_policy": "json",
                "rest": rest,
            })
            cmds.append({
                "path": cmd_path,
                "summary": summary,
                "intent": "direct_read",
                "availability": "implemented",
                "operation": op_id,
                "source_cli_path": "",
                "source_url": DOCS,
                "flags": path_flags(op),
                "api_surface": [{"method": op["method"], "path": op["full"]}],
                "output_policy": "json_redacted",
            })
            rows.append({
                "method": op["method"],
                "path": op["full"],
                "covered_by": {"direct_reads": [cmd_path]},
            })
            stats["post_shaped_read"] += 1
            continue

        # ---- Write actions -------------------------------------------------
        name = action_name(op)
        path_vars = [v for v in re.findall(r"{(\w+)}", op["full"]) if v not in CONFIG_PATH_VARS]
        record = {
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "required": list(path_vars),
            "properties": {v: {"type": "string"} for v in path_vars},
        }
        if schema and schema.get("type") == "object" and schema.get("properties"):
            for prop, sub in schema["properties"].items():
                record["properties"].setdefault(prop, sub)
            for req in schema.get("required", []):
                if req not in record["required"]:
                    record["required"].append(req)

        action = {
            "name": name,
            "kind": WRITE_KIND[op["method"]],
            "method": op["method"],
            "path": templated_path(op),
            "body_type": "json" if schema else "none",
            "record_schema": record,
            "risk": "Workday %s on %s; changes tenant data." % (op["method"], op["full"]),
        }
        if path_vars:
            action["path_fields"] = list(path_vars)
        if collapsed:
            # The collapsed query-string page's behaviour, re-expressed as an
            # OPTIONAL record field surfaced as a flag -- help-scout's --async
            # shape (sweep finding 23). It must be omit_when_absent: baking
            # ?type=archive into the action would make every partial update
            # archive the record, which is not what the base endpoint does.
            record["properties"]["type"] = {"type": "string", "enum": [collapsed]}
            action["query"] = {
                "type": {"template": "{{ record.type }}", "omit_when_absent": True}
            }
            stats["collapsed_flag"] += 1
        actions.append(action)

        # Every required mapping path needs a flag, and a required field with no
        # scalar leaf makes the command `partial` rather than `implemented`.
        flags, blocked_by = [], []
        for dotted in required_mapping_paths(record):
            leaf = resolve_leaf(record, dotted)
            ftype = flag_type_for(leaf)
            if ftype is None:
                blocked_by.append(dotted)
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

        if collapsed:
            flags.append({
                "name": "type",
                "type": "enum",
                "values": [collapsed],
                "summary": (
                    "Request the '%s' mode Workday documents on this endpoint as a separate "
                    "query-string page. Omitted by default, which is the endpoint's own behaviour."
                    % collapsed
                ),
                "maps_to": "record.type",
            })

        cmd = {
            "path": cmd_path,
            "summary": summary,
            "intent": "reverse_etl",
            "availability": "partial" if blocked_by else "implemented",
            "write": name,
            "source_cli_path": "",
            "source_url": DOCS,
            "risk": "Writes to the configured Workday tenant.",
            "approval": "Reverse ETL writes require plan, preview, approval, execute.",
            "flags": flags,
        }
        if blocked_by:
            cmd["notes"] = (
                "availability: partial -- Workday declares required record field(s) %s with no "
                "scalar leaf, so no flag can carry them and the runtime would refuse an "
                "'implemented' claim. Supply them with --record instead."
                % ", ".join(sorted(blocked_by))
            )
            stats["partial"] += 1
        else:
            stats["implemented_write"] += 1
        cmds.append(cmd)
        rows.append({
            "method": op["method"],
            "path": op["full"],
            "covered_by": {"writes": [name]},
        })
    return rows, cmds, actions, operations, dict(stats)

if __name__ == "__main__":
    main()
