#!/usr/bin/env python3
"""Expand github's api_surface/cli_surface to the full documented GET surface.

Source of truth is GitHub's own OpenAPI description (12,920,264 bytes, openapi
3.0.3, info.version 1.1.4) -- the byte-identical artifact the sweep derived
1220 operations from. Nothing here is invented: every row, command, flag and
summary is read out of that artifact.

Four judgements are encoded here and are NOT mechanical; each is recorded in
the phase SUMMARY:

  read-vs-write        every documented GET is a read; no GET is modelled as a
                       write, and no non-GET is touched by this pass.
  stream-vs-direct     new GETs become plain direct_read commands rather than
                       streams. A stream needs a hand-authored record schema,
                       primary key and fixture; inventing 359 of those would be
                       inventing data contracts the provider never published.
                       A direct read returns exactly what the endpoint returns.
  binary detection     an operation is binary iff its documented success
                       response is a 302 redirect to a file download. Read out
                       of the artifact, never guessed from the path.
  named-dependency     a GET is blocked only when a named runtime component
                       refuses it, and the note names that component.
"""

import json
import os
import re
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__)))
BUNDLE = None  # set from argv
SPEC_PATH = os.path.join(ROOT, "api.github.com.json")

# Path variables the connection spec supplies, so they never become flags.
CONFIG_PATH_VARS = {"owner", "repo"}

# Paging parameters are derived from a connector's declared pagination spec by
# the foundation lane, never hand-authored here. This blocklist exists so a
# required-parameter sweep can never smuggle one in.
PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults", "first", "last",
}


def load_spec():
    with open(SPEC_PATH) as fh:
        return json.load(fh)


def make_deref(spec):
    def deref(node):
        seen = 0
        while isinstance(node, dict) and "$ref" in node and seen < 10:
            cur = spec
            for part in node["$ref"].lstrip("#/").split("/"):
                cur = cur[part]
            node = cur
            seen += 1
        return node
    return deref


def success_kinds(op, deref):
    """Content types (and <302> marker) of an operation's success responses."""
    kinds = set()
    for code, resp in (op.get("responses") or {}).items():
        if not (code.startswith("2") or code == "302"):
            continue
        resp = deref(resp)
        kinds |= set((resp.get("content") or {}).keys())
        if code == "302":
            kinds.add("<302>")
    return kinds


def first_sentence(text, cap=160):
    if not text:
        return ""
    text = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", text)      # markdown links
    text = re.sub(r"[`*_>#]", "", text)
    text = " ".join(text.split())
    for marker in (". ", "! "):
        idx = text.find(marker)
        if 0 < idx < cap:
            return text[: idx + 1]
    if len(text) > cap:
        text = text[: cap - 1].rsplit(" ", 1)[0] + "."
    return text


def flag_spec(schema):
    """(type, values) for a flag, taken from the parameter's own schema. An
    enumerated parameter becomes an enum flag so `--help` lists what the
    provider actually accepts instead of inviting a 422."""
    schema = schema or {}
    values = [str(v) for v in schema.get("enum", []) if v is not None]
    if values:
        return "enum", values
    kind = schema.get("type")
    if kind == "integer":
        return "integer", None
    if kind == "number":
        return "number", None
    if kind == "boolean":
        return "boolean", None
    return "string", None


def typed_flag(name, param, target, summary_fallback):
    kind, values = flag_spec(param.get("schema"))
    flag = {
        "name": kebab(name),
        "type": kind,
        "summary": first_sentence(param.get("description")) or summary_fallback,
        "maps_to": target,
        "required": True,
    }
    if values:
        flag["values"] = values
    return flag


def kebab(name):
    return name.replace("_", "-")


# The bundle already ships five `search *` commands as availability "planned",
# named after gh's own subcommands rather than after GitHub's operationIds.
# Promoting those rows in place is the honest move: adding a second command for
# the same endpoint would ship two names for one operation. `search prs` is the
# exception -- gh models it as a preset over the SAME issues-and-pull-requests
# endpoint, so it stays planned and its note now names the command that covers
# the endpoint instead of claiming a missing capability.
OPERATION_ID_ALIASES = {
    "search/code": "search code",
    "search/commits": "search commits",
    "search/repos": "search repos",
    "search/issues-and-pull-requests": "search issues",
}


def command_path(operation_id):
    alias = OPERATION_ID_ALIASES.get(operation_id)
    if alias:
        return alias
    group, _, action = operation_id.partition("/")
    return "%s %s" % (kebab(group), kebab(action))


def main():
    global BUNDLE
    BUNDLE = sys.argv[1]
    spec = load_spec()
    deref = make_deref(spec)

    surface_path = os.path.join(BUNDLE, "api_surface.json")
    cli_path = os.path.join(BUNDLE, "cli_surface.json")
    ops_path = os.path.join(BUNDLE, "operations.json")
    with open(surface_path) as fh:
        surface = json.load(fh)
    with open(cli_path) as fh:
        cli = json.load(fh)
    with open(ops_path) as fh:
        operations = json.load(fh)

    have = {(e["method"], e["path"]) for e in surface["endpoints"]}
    existing_cmd_paths = {c["path"] for c in cli["commands"]}

    new_rows = []
    new_cmds = []
    new_ops = []
    stats = {"direct_read": 0, "binary": 0, "blocked_status": 0, "blocked_text": 0}
    collisions = []
    promoted = {}

    for path in sorted(spec["paths"]):
        item = spec["paths"][path]
        op = item.get("get")
        if not op or ("GET", path) in have:
            continue

        operation_id = op["operationId"]
        cmd_path = command_path(operation_id)
        replacing = operation_id in OPERATION_ID_ALIASES
        if cmd_path in existing_cmd_paths and not replacing:
            collisions.append((cmd_path, operation_id))
            continue
        existing_cmd_paths.add(cmd_path)

        docs = (op.get("externalDocs") or {}).get("url", "")
        summary = op.get("summary") or operation_id
        if op.get("deprecated"):
            summary = summary + " (deprecated by GitHub, still documented)"
        kinds = success_kinds(op, deref)

        # BINARY: a documented 302 redirect to a downloadable file.
        verb = operation_id.partition("/")[2]
        if "<302>" in kinds and not verb.startswith("check"):
            op_id = "github." + operation_id.replace("/", "_").replace("-", "_")
            new_ops.append({
                "id": op_id,
                "kind": "binary_download",
                "summary": summary,
                "source_url": docs or "https://docs.github.com/en/rest",
                "risk": "medium",
                "approval": "none",
                "output_policy": "binary",
                "binary": {
                    "method": "GET",
                    "path": path,
                    "max_bytes": 104857600,
                    "allow_overwrite": False,
                    "extract_archives": False,
                },
            })
            cmd = {
                "path": cmd_path,
                "summary": summary,
                "intent": "binary_download",
                "availability": "implemented",
                "operation": op_id,
                "source_cli_path": "",
                "flags": path_flags(path, item, op, deref),
                "api_surface": [{"method": "GET", "path": path}],
            }
            new_cmds.append(cmd)
            new_rows.append({
                "method": "GET",
                "path": path,
                "covered_by": {"direct_read": cmd_path},
            })
            stats["binary"] += 1
            continue

        # BLOCKED: the engine's direct-read executor decodes a JSON body. An
        # endpoint that documents no JSON success body cannot be an honest
        # implemented direct read, and the block names the component that
        # refuses it.
        if "application/json" not in kinds:
            if not kinds or kinds <= {"<302>"}:
                reason = (
                    "documented success response is 204 No Content: this is a boolean "
                    "status check, not a JSON body"
                )
                note = (
                    "Named dependency: engine.decodeDirectReadBody requires a JSON body and "
                    "commandrunner.supportedDirectReadOutputPolicies declares no status-only "
                    "policy, so a 204 endpoint cannot be an implemented direct_read"
                )
                stats["blocked_status"] += 1
            else:
                reason = (
                    "documented success response is %s, not application/json"
                    % ", ".join(sorted(kinds))
                )
                note = (
                    "Named dependency: engine.decodeDirectReadBody json-decodes the response "
                    "body and commandrunner.supportedDirectReadOutputPolicies declares no "
                    "text/plain policy"
                )
                stats["blocked_text"] += 1
            new_rows.append({
                "method": "GET",
                "path": path,
                "operation": {
                    "model": "direct_read",
                    "status": "blocked",
                    "risk": "low",
                    "blocked_by_default": True,
                    "reason": reason,
                    "source_url": docs or "https://docs.github.com/en/rest",
                    "notes": note,
                },
            })
            continue

        # DIRECT READ: the default for a documented JSON GET.
        cmd = {
            "path": cmd_path,
            "summary": summary,
            "intent": "direct_read",
            "availability": "implemented",
            "source_cli_path": "",
            "source_url": docs or "https://docs.github.com/en/rest",
            "flags": path_flags(path, item, op, deref),
            "api_surface": [{"method": "GET", "path": path}],
            "output_policy": "json_redacted",
        }
        if replacing:
            promoted[cmd_path] = cmd
        else:
            new_cmds.append(cmd)
        new_rows.append({
            "method": "GET",
            "path": path,
            "covered_by": {"direct_read": cmd_path},
        })
        stats["direct_read"] += 1

    if collisions:
        print("COMMAND PATH COLLISIONS (%d):" % len(collisions))
        for c in collisions:
            print("  ", c)
        sys.exit(1)

    # Promote the planned placeholders in place, keeping their gh provenance.
    for i, cmd in enumerate(cli["commands"]):
        upgrade = promoted.get(cmd["path"])
        if upgrade is None:
            continue
        upgrade["source_cli_path"] = cmd.get("source_cli_path", "")
        cli["commands"][i] = upgrade
    if promoted and len(promoted) != sum(
        1 for c in cli["commands"] if c["path"] in promoted and c["availability"] == "implemented"
    ):
        raise SystemExit("a planned search command was not promoted in place")

    surface["endpoints"].extend(new_rows)
    cli["commands"].extend(new_cmds)
    operations["operations"].extend(new_ops)

    write_json(surface_path, surface)
    write_json(cli_path, cli)
    write_json(ops_path, operations)
    print("new rows: %d  %s" % (len(new_rows), stats))
    print("new commands: %d  new operations: %d" % (len(new_cmds), len(new_ops)))


def path_flags(path, item, op, deref):
    """One flag per path variable the connection spec does not supply, plus
    every REQUIRED query parameter. Optional query filters are deliberately not
    authored: they are not operations, and authoring hundreds of them would
    bury the parity signal. Paging parameters are never authored at all."""
    params = [deref(p) for p in (item.get("parameters", []) + op.get("parameters", []))]
    by_name = {p["name"]: p for p in params if isinstance(p, dict) and "name" in p}
    flags = []
    for var in re.findall(r"{(\w+)}", path):
        if var in CONFIG_PATH_VARS:
            continue
        param = by_name.get(var, {})
        flags.append(typed_flag(var, param, "path.%s" % var, "Path parameter %s." % var))
    for param in params:
        if not isinstance(param, dict) or param.get("in") != "query":
            continue
        if not param.get("required"):
            continue
        name = param["name"]
        if name.lower() in PAGING_PARAMS:
            raise SystemExit("refusing to author paging flag %r on %s" % (name, path))
        flags.append(typed_flag(name, param, "query.%s" % name, "Query parameter %s." % name))
    return flags


def write_json(path, payload):
    with open(path, "w") as fh:
        # The shipped bundle files carry no trailing newline; preserving that
        # keeps the diff to the rows this pass actually adds.
        json.dump(payload, fh, indent=2)


if __name__ == "__main__":
    main()
