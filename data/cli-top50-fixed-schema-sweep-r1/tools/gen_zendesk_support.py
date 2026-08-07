#!/usr/bin/env python3
"""Promote zendesk-support's inventoried operations to reachable commands.

This connector is NOT from-nothing. #3532 already shipped a complete 631-row
api_surface (625 documented Support-OAS operations + 6 rows carried for Guide /
community / business-hours streams whose endpoints live in separate Zendesk
specs). What it did not ship is reachability: 509 rows are disposed by one
blanket "blocked until shared foundation #2985" sentence and their commands are
`availability: planned` placeholders.

So this generator PROMOTES rather than enumerates. It rewrites in place and is
re-runnable from a clean tree (`git checkout -- internal/connectors/defs/zendesk-support/`).

    python3 gen_zendesk_support.py <bundle-dir> <DERIVED-OPERATIONS.json> --reads
    python3 gen_zendesk_support.py <bundle-dir> <DERIVED-OPERATIONS.json> --writes

Rules encoded here, each of which cost something to learn:

  * Never author a paging flag. The blocklist raises rather than emitting one.
    `start_time` is deliberately absent from it -- see PLAN.md "Paging": every
    stream declares pagination `type: next_url`, so paging comes from response
    links, and start_time is the incremental export's required watermark.
  * Flags cover path variables and REQUIRED query parameters only. Optional
    filters are not operations, and authoring hundreds of them buries the
    parity signal.
  * A read-shaped POST needs BOTH `content_type: application/json` and
    `body_schema` (finding 33). All seven promoted here already carry both;
    the generator asserts it rather than assuming it.
  * Binary is GET-only and read from the artifact's declared media types.
  * A credential-bearing body is a named runtime dependency, not a flag.
"""

import argparse
import collections
import json
import os
import re
import sys

PAGING_PARAMS = {
    "page", "per_page", "limit", "offset", "cursor", "page_size", "pagesize",
    "page_token", "pagetoken", "max_results", "maxresults", "first", "last",
    "cursor_pagination", "page[before]", "page[after]", "page[size]",
}

# POST /api/v2/any_channel/validate_token is read-SHAPED -- it validates and
# stores nothing -- but its documented request body carries a channel token and
# an account push id. Promoting it to a direct read means authoring a --token
# flag, which is exactly the inline-credential input AGENTS.md forbids
# ("Add credentials from environment variables or stdin, not prompt text";
# "Never request, print, summarize, or store secret values"). It therefore
# stays blocked, and its block names the capability that is actually missing
# instead of a shared-foundation issue number.
CREDENTIAL_BODY_BLOCKED = {
    "POST /api/v2/any_channel/validate_token": (
        "documented request body carries a channel token and account push id; promoting it "
        "would author an inline --token flag",
        "Named dependency: commandrunner exposes no stdin/env-sourced secret input for a "
        "request-body credential, and engine redaction covers response fields rather than "
        "request payloads, so this read cannot be reached without inlining a secret",
    ),
}

# The shipped bundle declares these two POSTs `kind: rest_read`. They are not
# reads. Zendesk's own description says the endpoint "enqueues a job to create a
# CSV file" and "sends the requester an email containing a link to the CSV
# file", and /audit_logs/export documents ONLY a 202 Accepted with no body at
# all. Enqueuing a job and sending mail are side effects; a bounded read must
# have none. They are handled by the write slice, which is also how this sweep
# classified GitHub's analogous async export endpoints in the same batch.
ASYNC_EXPORT_WRITES = {
    "POST /api/v2/audit_logs/export",
    "POST /api/v2/suspended_tickets/export",
}

FLAG_TYPES = {
    "integer": "integer",
    "number": "number",
    "boolean": "boolean",
    "array": "string_array",
}


# engine.CompileSchema implements a strict draft-07 SUBSET and treats an
# unknown keyword as a compile error, deliberately, "keeping bundles honest".
# These are the keywords it accepts (engine/schema.go annotationKeywords +
# structuralKeywords); anything else must be removed before a schema lifted out
# of an OpenAPI document can be declared executable.
SCHEMA_KEYWORDS = {
    "format", "default", "title", "description", "$schema",
    "type", "required", "properties", "items", "enum", "pattern",
    "minProperties", "minItems", "maxItems", "additionalProperties",
    "x-secret", "x-primary-key", "x-cursor-field",
}

# OpenAPI annotations that carry no validation meaning in this dialect and are
# safe to drop. Anything NOT on this list and not in SCHEMA_KEYWORDS raises,
# so a keyword that does constrain the payload can never be dropped silently.
# `nullable` is the one that is not merely dropped: OpenAPI 3.0 spells
# "may be null" that way and draft-07 spells it inside `type`, so it is
# translated rather than discarded.
DROPPABLE_ANNOTATIONS = {
    "example", "examples", "externalDocs", "xml", "discriminator",
    "deprecated", "readOnly", "writeOnly",
}


def sanitize_schema(node, where):
    """Strip OpenAPI-only annotations so the schema compiles in this dialect."""
    if isinstance(node, list):
        return [sanitize_schema(item, where) for item in node]
    if not isinstance(node, dict):
        return node

    out = {}
    nullable = bool(node.get("nullable"))
    for key, value in node.items():
        if key == "nullable":
            continue
        if key in DROPPABLE_ANNOTATIONS:
            continue
        if key not in SCHEMA_KEYWORDS:
            raise SystemExit(
                "%s: schema keyword %r is neither supported by engine.CompileSchema nor a known "
                "droppable annotation; decide what it means before dropping it" % (where, key)
            )
        if key in ("properties",):
            out[key] = {k: sanitize_schema(v, where) for k, v in value.items()}
        elif key in ("items", "additionalProperties") and isinstance(value, dict):
            out[key] = sanitize_schema(value, where)
        else:
            out[key] = value

    if nullable:
        declared = out.get("type")
        if isinstance(declared, str):
            out["type"] = sorted({declared, "null"})
        elif isinstance(declared, list):
            out["type"] = sorted(set(declared) | {"null"})
    return out


def flatten_union(node, where):
    """Rewrite a oneOf/anyOf node into one compilable schema, or refuse.

    AGENTS.md: a schema rooted at oneOf/anyOf is not one executable command
    contract -- runtime preflight expands its arms and rejects promotion.
    `engine.CompileSchema` says the same thing more bluntly: `unknown keyword
    "anyOf"`. The rule's remedy is to model each reachable arm, and there are
    two ways a Zendesk union turns out to have only ONE reachable arm:

      * the arms are structurally identical and differ only by `title`. Both
        filtered-search bodies are exactly {"type":"object",
        "additionalProperties":true} twice over, labelled Basic and Complex.
        That is documentation prose, not a contract union, and collapsing it
        loses nothing the provider published.
      * the arms are type variants of one value -- "all" or an integer for
        brand_id, an enum string or an object for filter. draft-07 expresses
        that natively as a type array, which keeps EVERY documented arm
        reachable rather than picking a winner.

    An arm carrying its own `required` list is a genuinely different contract
    and is NOT flattened: this raises, and the caller blocks the command with
    the missing capability named. Refusing loudly is the point -- silently
    keeping one arm would ship a command that rejects half its documented
    inputs.
    """
    if not isinstance(node, dict):
        return node
    key = "oneOf" if "oneOf" in node else ("anyOf" if "anyOf" in node else None)
    if key is None:
        return {k: flatten_union(v, where) if isinstance(v, dict) else v for k, v in node.items()}

    arms = node[key]
    for arm in arms:
        if arm.get("required"):
            raise SystemExit(
                "%s: %s arm %r declares its own required fields; it is a distinct contract "
                "and must be modelled as a separate named action, not flattened"
                % (where, key, arm.get("title") or arm.get("type"))
            )

    types, properties = set(), {}
    for arm in arms:
        arm_type = arm.get("type")
        if isinstance(arm_type, list):
            types.update(arm_type)
        elif arm_type:
            types.add(arm_type)
        for name, schema in (arm.get("properties") or {}).items():
            properties.setdefault(name, flatten_union(schema, where))

    merged = {k: v for k, v in node.items() if k not in (key, "title")}
    if types:
        merged["type"] = sorted(types)[0] if len(types) == 1 else sorted(types)
    if properties:
        merged["properties"] = properties
    if "object" in types:
        # Every Zendesk union arm here is open. Keeping it open is what makes
        # a collapsed arm lossless: nothing the provider documented is rejected.
        merged["additionalProperties"] = True
    return merged


def kebab(name):
    return re.sub(r"[^a-z0-9]+", "-", name.lower()).strip("-")


def flag_type(declared):
    return FLAG_TYPES.get(declared, "string")


def load(path):
    with open(path) as fh:
        return json.load(fh)


def write_json(path, payload):
    with open(path, "w") as fh:
        json.dump(payload, fh, indent=2)


def command_flags(derived_row, key):
    """One flag per path variable, plus every REQUIRED query parameter."""
    flags = []
    for var in derived_row["path_vars"]:
        flags.append(
            {
                "name": kebab(var),
                "type": "string",
                "summary": "Path parameter %s." % var,
                "maps_to": "path.%s" % var,
                "required": True,
            }
        )
    for param in derived_row["required_query"]:
        name = param["name"]
        if name.lower() in PAGING_PARAMS:
            raise SystemExit("refusing to author paging flag %r on %s" % (name, key))
        flags.append(
            {
                "name": kebab(name),
                "type": flag_type(param["type"]),
                "summary": param["summary"],
                "maps_to": "query.%s" % name,
                "required": True,
            }
        )
    return flags


def index_operations(operations):
    """(METHOD, path) -> operation, over both rest and binary declarations."""
    index = {}
    for op in operations:
        rest = op.get("rest") or op.get("binary") or {}
        method = (rest.get("method") or "").upper()
        if method:
            index[(method, rest.get("path"))] = op
    return index


def promote_reads(bundle, derived):
    surface_path = os.path.join(bundle, "api_surface.json")
    cli_path = os.path.join(bundle, "cli_surface.json")
    ops_path = os.path.join(bundle, "operations.json")

    surface = load(surface_path)
    cli = load(cli_path)
    ops_doc = load(ops_path)

    by_endpoint = index_operations(ops_doc["operations"])
    cmd_by_operation = {c["operation"]: c for c in cli["commands"] if c.get("operation")}
    derived_by_key = {r["key"]: r for r in derived["operations"]}

    stats = collections.Counter()
    promoted_paths = {}

    for row in surface["endpoints"]:
        if not row.get("operation"):
            continue
        key = "%s %s" % (row["method"], row["path"])
        op = by_endpoint.get((row["method"], row["path"]))
        if op is None:
            continue

        # This one is checked BEFORE the kind filter on purpose. The shipped
        # bundle declares validate_token `rest_write`, so a read-kind filter
        # would skip it entirely and leave it carrying the blanket
        # shared-foundation sentence -- blocked for a real reason, but not a
        # stated one.
        if key in CREDENTIAL_BODY_BLOCKED:
            reason, note = CREDENTIAL_BODY_BLOCKED[key]
            row["operation"]["reason"] = reason
            row["operation"]["notes"] = note
            stats["blocked_credential_body"] += 1
            continue

        if op["kind"] not in ("rest_read", "binary_download"):
            continue

        if key in ASYNC_EXPORT_WRITES:
            stats["deferred_to_write_slice"] += 1
            continue

        command = cmd_by_operation.get(op["id"])
        if command is None:
            raise SystemExit("no cli command declares operation %s (%s)" % (op["id"], key))

        derived_row = derived_by_key.get(key)
        # The six carried rows are outside the Support OAS, so the derivation
        # has nothing to say about them; their path variables still do.
        if derived_row is None:
            derived_row = {"path_vars": re.findall(r"{([^}]+)}", row["path"]), "required_query": []}

        if op["kind"] == "binary_download":
            # A binary_download command declares no output_policy of its own:
            # the policy lives on the operation, and `surface-sync --check`
            # strips one set here. Deriving it wrongly and letting surface-sync
            # correct it is exactly the hand-maintained drift AGENTS.md forbids.
            intent, policy = "binary_download", None
            stats["binary_download"] += 1
        else:
            rest = op["rest"]
            if rest["method"].upper() == "POST":
                # Finding 33: two separate validate rules, caught one run apart.
                # Assert rather than assume -- a POST read missing either is
                # rejected by engine.operationDirectReadSpec at preflight.
                if (rest.get("content_type") or "").lower() != "application/json":
                    raise SystemExit("%s: POST read has no application/json content_type" % key)
                if not rest.get("body_schema"):
                    raise SystemExit("%s: POST read has no body_schema" % key)
                unioned = flatten_union(rest["body_schema"], key)
                if unioned != rest["body_schema"]:
                    stats["union_body_flattened"] += 1
                compilable = sanitize_schema(unioned, key)
                if compilable != rest["body_schema"]:
                    rest["body_schema"] = compilable
                    stats["body_schema_rewritten"] += 1
                stats["direct_read_post"] += 1
            else:
                stats["direct_read_get"] += 1
            intent, policy = "direct_read", op["output_policy"]

        command["intent"] = intent
        command["availability"] = "implemented"
        command["operation"] = op["id"]
        if policy is None:
            command.pop("output_policy", None)
        else:
            command["output_policy"] = policy
        command["api_surface"] = [{"method": row["method"], "path": row["path"]}]
        command["flags"] = command_flags(derived_row, key)
        command["approval"] = "none: bounded read, response capped and redacted by output policy"
        command.pop("risk", None)
        command["notes"] = (
            "Bounded direct read of an official Zendesk Support operation. Responses are capped "
            "at the operation's declared max_bytes and redacted by its output policy; this is not "
            "a raw API passthrough."
        )

        # The operation's own prose still claimed it was blocked by default.
        # Leaving it would ship a command whose metadata contradicts it.
        op["description"] = (
            "Bounded, connector-owned read of an official Zendesk Support operation. Executed "
            "through the declarative direct-read executor with a capped, policy-redacted response; "
            "it does not enable a raw API escape hatch."
        )
        op["approval"] = "none: bounded read, response capped and redacted by output policy"

        row.pop("operation", None)
        row["covered_by"] = {"direct_read": command["path"]}
        promoted_paths[command["path"]] = key

    write_json(surface_path, surface)
    write_json(cli_path, cli)
    write_json(ops_path, ops_doc)
    print("reads promoted: %s" % dict(sorted(stats.items())))
    print("total commands promoted: %d" % len(promoted_paths))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("bundle")
    parser.add_argument("derived")
    parser.add_argument("--reads", action="store_true")
    parser.add_argument("--writes", action="store_true")
    args = parser.parse_args()

    derived = load(args.derived)
    if args.reads:
        promote_reads(args.bundle, derived)
    if args.writes:
        raise SystemExit("--writes is slice 3; not implemented in this commit")
    if not (args.reads or args.writes):
        parser.error("pass --reads or --writes")


if __name__ == "__main__":
    main()
