#!/usr/bin/env python3
"""Regenerate internal/connectors/defs/recurly/cli_surface.json with the full
197-command surface (93 ETL streams + 96 typed reverse-ETL writes + 8
direct-read/export operations), derived deterministically from the connector's
streams.json / writes.json / operations.json / api_surface.json.

Every stream and every write gets an individual `pm recurly <command>` entry.
Streams are implemented ETL commands. Writes are implemented reverse-ETL
commands where every required record path is satisfiable by a typed command
flag; otherwise they are marked partial with a named reason (complex object
payload delivered via typed reverse-ETL records). The 8 direct-read/export
operations are preserved from the existing cli_surface.json.

Usage: python3 scripts/gen-recurly-cli-surface.py
"""

import copy
import json
import os
import sys
from collections import OrderedDict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DEFS = os.path.join(ROOT, "internal", "connectors", "defs", "recurly")


def load(name):
    with open(os.path.join(DEFS, name)) as f:
        return json.load(f)


streams = load("streams.json")["streams"]
writes = load("writes.json")["actions"]
ops = load("operations.json")["operations"]
api = load("api_surface.json")
cur = load("cli_surface.json")

# ---------------------------------------------------------------- STREAM paths
# Reserved singular/plural helpers: derive a resource noun per stream name.
_bare_plural = {"accounts", "invoices", "plans", "subscriptions", "transactions"}

# Explicit overrides for names whose default derivation would collide or read
# poorly.
STREAM_OVERRIDES = {
    "get_performance_obligation": "performance-obligations get",
    "get_performance_obligations": "performance-obligations get-all",
    "get_a_billing_info": "billing-infos get",
    "get_billing_info": "accounts billing-infos get",
}


def stream_path(name):
    if name in STREAM_OVERRIDES:
        return STREAM_OVERRIDES[name]
    if name in _bare_plural:
        return name + " list"
    for pfx, verb in (("list_", "list"), ("get_", "get"), ("show_", "get")):
        if name.startswith(pfx):
            rest = name[len(pfx):]
            rest = rest.replace("_", " ")
            if rest.startswith("a "):  # get_a_billing_info
                rest = rest[2:]
            return rest + " " + verb
    raise SystemExit("no stream path rule for " + name)


# --------------------------------------------------------------- WRITE paths
WRITE_OVERRIDES = {
    "create_invoice_retry": "invoices retries create",
    "create_pending_purchase": "purchases create-pending",
    "create_authorize_purchase": "purchases authorize",
    "create_capture_purchase": "purchases capture",
    "cancel_purchase": "purchases cancel",
    "convert_trial": "subscriptions convert-trial",
    "generate_unique_coupon_codes": "unique-coupon-codes generate",
    "generate_unique_coupon_codes_sync": "unique-coupon-codes generate-sync",
    "put_dunning_campaign_bulk_update": "dunning-campaigns bulk-update",
    "record_external_transaction": "invoices record-external-transaction",
    "apply_credit_balance": "invoices apply-credit-balance",
    "put_external_subscription": "external-subscriptions put",
}


def _write_default(name):
    # default: resource = name with leading verb strip; verb = first token
    verbs = [
        "create", "update", "delete", "remove", "deactivate", "reactivate",
        "terminate", "cancel", "pause", "resume", "restore", "collect", "void",
        "reopen", "mark", "verify", "refund", "redeem", "convert", "generate",
        "put", "apply", "record", "redact",
    ]
    for v in verbs:
        if name.startswith(v + "_"):
            rest = name[len(v) + 1:]
            return rest.replace("_", " ") + " " + v
    raise SystemExit("no write path rule for " + name)


def write_path(name):
    if name in WRITE_OVERRIDES:
        return WRITE_OVERRIDES[name]
    return _write_default(name)


# ------------------------------------------------- schema required-path logic
# Faithful port of validator.go checkCLISurfaceWriteFlags requiredMappingPaths.
class Node:
    def __init__(self, m):
        m = m or {}
        t = m.get("type")
        self.types = t if isinstance(t, list) else ([t] if t else [])
        self.required = m.get("required", []) or []
        self.properties = m.get("properties", {}) or {}
        self.items = Node(m.get("items", {})) if m.get("items") else None

    def is_array(self):
        return "array" in self.types

    def is_object(self):
        return bool(self.properties) or "object" in self.types or not self.types

    def required_mapping_paths(self, prefix=""):
        out = []
        for req in self.required:
            path = prefix + "." + req if prefix else req
            child = Node(self.properties.get(req, {}))
            child_paths = child.required_node_mapping_paths(path)
            if not child_paths:
                out.append(path)
            else:
                out.extend(child_paths)
        return out

    def required_node_mapping_paths(self, prefix):
        if self.is_array():
            if self.items is None:
                return []
            paths = self.items.required_node_mapping_paths(prefix + ".0")
            return [prefix] if not paths else paths
        if self.is_object():
            return self.required_mapping_paths(prefix)
        return None


def leaf_types(node):
    return node.types or (["array"] if node.is_array() else (["object"] if node.is_object() else ["any"]))


def required_paths(rs):
    return Node(rs).required_mapping_paths("")


def flag_type_for(types):
    if "boolean" in types:
        return "boolean"
    if "integer" in types or "number" in types:
        return "integer"
    if "array" in types:
        return "string_array"
    return "string"


def schema_desc(rs, dotted):
    node = rs
    parts = dotted.split(".")
    for p in parts:
        if node is None:
            return ""
        props = node.get("properties", {}) if isinstance(node, dict) else {}
        node = props.get(p)
    if isinstance(node, dict):
        return node.get("description", "")
    return ""


# Choose mappable flag targets for a required field. Returns a list of
# (target, node) pairs (scalar/array leaves to map) or [] if the field's
# required object shape can't be expressed as flat command flags.
def choose_target(schema_prop, prefix_path):
    node = Node(schema_prop)
    if node.is_array():
        return [(prefix_path, node)]
    if node.is_object():
        chosen = []
        for req in node.required:
            child = node.properties.get(req, {})
            child_path = (prefix_path + "." + req) if prefix_path else req
            chosen.extend(choose_target(child, child_path))
        return chosen
    # scalar
    return [(prefix_path, node)]


def leaf_node_at(rs, dotted):
    """Walk a dotted path through the record schema, descending array '0'
    indices into items. Returns the schema node dict or None."""
    node = rs
    for part in dotted.split("."):
        if isinstance(node, dict) and node.get("type") == "array" and node.get("items") is not None and part.isdigit():
            node = node["items"]
            continue
        props = node.get("properties", {}) if isinstance(node, dict) else {}
        node = props.get(part)
        if node is None:
            return None
    return node


def is_object_node(node):
    if not isinstance(node, dict):
        return False
    t = node.get("type")
    if node.get("properties"):
        return True
    return t == "object"


def build_write_command(w):
    name = w["name"]
    path = write_path(name)
    rs = w.get("record_schema", {})

    # Required mapping paths: every scalar/array leaf the Go validator demands
    # a flag for (a mapped target must equal-or-descend the required path).
    req_paths = set(Node(rs).required_mapping_paths(""))
    for pf in (w.get("path_fields", []) or []):
        req_paths.add(pf)

    flags = []
    ok = True
    for P in sorted(req_paths):
        leaf = leaf_node_at(rs, P)
        if leaf is None or is_object_node(leaf):
            ok = False
            break
        flags.append((P, flag_type_for(leaf_types(Node(leaf)))))

    if not ok:
        availability = "partial"
        # still expose mappable scalar/array leaf flags for documentation
        flags = []
        for P in sorted(req_paths):
            leaf = leaf_node_at(rs, P)
            if leaf is None or is_object_node(leaf):
                continue
            flags.append((P, flag_type_for(leaf_types(Node(leaf)))))
        notes = ("Complex required object/array payloads for this action are delivered via typed "
                 "reverse-ETL records rather than flat command flags; command form is partial.")
    else:
        availability = "implemented"
        notes = ""

    # dedupe flags by target (a scalar leaf may appear via multiple required paths)
    seen_targets = {}
    flag_list = []
    for target, ftype in flags:
        if target in seen_targets:
            continue
        seen_targets[target] = True
        dotted = "record." + target
        flag_name = target.replace(".", "-").replace("_", "-")
        flag_list.append({
            "name": flag_name,
            "type": ftype,
            "summary": schema_desc(rs, target) or f"Type {ftype} value mapped to Recurly `{target}`.",
            "maps_to": dotted,
        })

    cmd = {
        "path": path,
        "summary": summary_for_write(w),
        "intent": "reverse_etl",
        "availability": availability,
        "write": name,
        "risk": w.get("risk", ""),
        "approval": "Reverse ETL mutations require plan, preview, explicit approval, and execute; "
                    "destructive lifecycle actions require destructive confirmation.",
    }
    if flag_list:
        cmd["flags"] = flag_list

    # api_surface
    ep = endpoint_for("write", name)
    if ep is not None:
        cmd["api_surface"] = [{"method": ep["method"], "path": ep["path"]}]

    examples = ["pm recurly " + path + (" --json" if availability == "implemented" else "")]
    if flag_list:
        ex = "pm recurly " + path
        for fl in flag_list[:2]:
            ex += " --" + fl["name"] + ' "value"'
        ex += " --json"
        examples[0] = ex
    cmd["examples"] = examples

    if notes and availability == "partial":
        cmd["notes"] = notes
    return cmd


def build_stream_command(s):
    name = s["name"]
    path = stream_path(name)
    cmd = {
        "path": path,
        "summary": summary_for_stream(s),
        "intent": "etl",
        "availability": "implemented",
        "stream": name,
    }
    ep = endpoint_for("stream", name)
    if ep is not None:
        cmd["api_surface"] = [{"method": ep["method"], "path": ep["path"]}]
    cmd["examples"] = ["pm recurly " + path + " --json"]
    return cmd


def summary_for_stream(s):
    n = s["name"]
    if n.startswith("list_"):
        return "List Recurly " + n[5:].replace("_", " ") + " as ETL records."
    if n.startswith("get_") or n.startswith("show_"):
        return "Get Recurly " + n[4:].replace("_", " ") + " as an ETL record."
    return "List Recurly " + n.replace("_", " ") + " as ETL records."


def summary_for_write(w):
    kind = {  # kind -> human verb
        "create": "Create",
        "update": "Update",
        "delete": "Delete",
    }.get(w.get("kind"), "Write")
    return f"{kind} Recurly {w['name'].replace('_', ' ')} via typed reverse ETL."


def endpoint_for(kind, name):
    for ep in api["endpoints"]:
        cb = ep.get("covered_by") or {}
        if kind == "stream" and cb.get("stream") == name:
            return ep
        if kind == "write" and cb.get("write") == name:
            return ep
    return None


def main():
    commands = []
    paths = {}

    # 1. ETL streams
    for s in streams:
        c = build_stream_command(s)
        commands.append(c)
    # 2. Reverse-ETL writes
    for w in writes:
        c = build_write_command(w)
        commands.append(c)
    # 3. direct-read / export ops preserved from current surface (only the
    #    direct_read family survives regeneration; etl/reverse_etl are rebuilt)
    for c in cur["commands"]:
        if c["intent"] != "direct_read":
            continue
        c = copy.deepcopy(c)
        if c.get("availability") == "implemented" and c.get("flags"):
            # Executable examples must include declared required inputs so
            # copying the help example does not fail validation.
            ex = "pm recurly " + c["path"]
            for f in c["flags"]:
                ex += " --" + f["name"] + ' "<value>"'
            ex += " --json"
            c["examples"] = [ex]
        else:
            c["examples"] = ["pm recurly " + c["path"] + " --json"]
        commands.append(c)

    # collision check across all commands
    for c in commands:
        p = c["path"]
        if p in paths:
            raise SystemExit(f"DUPLICATE path: {p} ({paths[p]} vs {c['intent']}:{c.get('stream') or c.get('write')})")
        paths[p] = c["intent"]

    # counts
    etl = [c for c in commands if c["intent"] == "etl"]
    rev = [c for c in commands if c["intent"] == "reverse_etl"]
    dr = [c for c in commands if c["intent"] == "direct_read"]
    impl_rev = [c for c in rev if c["availability"] == "implemented"]
    part_rev = [c for c in rev if c["availability"] == "partial"]
    print(f"etl={len(etl)} (want {len(streams)}), reverse_etl={len(rev)} (want {len(writes)}), "
          f"direct_read={len(dr)}, total={len(commands)} (want {len(streams)+len(writes)+len(dr)})")
    print(f"  reverse_etl implemented={len(impl_rev)}, partial={len(part_rev)}")
    if part_rev:
        print("  partial writes:", [c["path"] for c in part_rev])

    # groups derived from leading noun token of each command path
    groups = OrderedDict()
    for c in commands:
        head = c["path"].split(" ", 1)[0] if " " in c["path"] else c["path"]
        gid = head.replace("-", "_")
        if gid not in groups:
            groups[gid] = {"id": gid, "title": " ".join(x.capitalize() for x in head.split("-")), "commands": []}
        groups[gid]["commands"].append(c["path"])

    out = {
        "tagline": cur["tagline"],
        "usage": "pm recurly <command> [flags] --json",
        "source_cli": cur["source_cli"],
        "groups": list(groups.values()),
        "commands": commands,
        "help_topics": cur.get("help_topics", []),
    }

    with open(os.path.join(DEFS, "cli_surface.json"), "w") as f:
        json.dump(out, f, indent=2)
        f.write("\n")
    print("wrote", os.path.join(DEFS, "cli_surface.json"))


if __name__ == "__main__":
    main()
