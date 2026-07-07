#!/usr/bin/env python3
"""Generate operations + write actions + CLI commands for every uncovered,
non-deprecated, non-duplicate GitHub endpoint, derived from api_surface.json.

Tiers (gated-execution principle, docs/plans/connector-verify-and-certificate-plan.md §0):
  - direct_read / binary_read  -> open read (operation: rest_read / binary_download)
  - admin_reverse_etl          -> gated reverse-ETL write (approval=typed_confirmation)
  - sensitive_reverse_etl      -> gated sensitive write (sensitive_policy + transform)
  - destructive_action         -> gated destructive write (destructive=true)
  - disallowed / deprecated    -> skipped (never / dead)

This is a first-pass coverage scaffold: every endpoint becomes backed by an
operation + (for writes) a write action + a CLI command with the correct gate.
Flag mappings use path params; write bodies use the declarative fallback
(buildJSONBody sends all non-path record fields as the JSON body).
"""
import json, re, sys

ROOT = "internal/connectors/defs/github"
api = json.load(open(f"{ROOT}/api_surface.json"))
ops_doc = json.load(open(f"{ROOT}/operations.json"))
writes_doc = json.load(open(f"{ROOT}/writes.json"))
cli_doc = json.load(open(f"{ROOT}/cli_surface.json"))

existing_op_ids = {o["id"] for o in ops_doc["operations"]}
existing_write_names = {a["name"] for a in writes_doc["actions"]}
existing_cli_paths = {c["path"] for c in cli_doc["commands"]}

PATH_PARAM_RE = re.compile(r"\{([^}]+)\}")

def slugify(s):
    return re.sub(r"[^a-z0-9]+", "_", s.lower()).strip("_")

def path_params(path):
    return PATH_PARAM_RE.findall(path)

def derive_cli_path(method, api_path, model):
    # strip /repos/{owner}/{repo} prefix -> segments
    p = api_path
    p = re.sub(r"^/repos/\{owner\}/\{repo\}(/)?", "", p)
    p = p.strip("/")
    segs = [s for s in p.split("/") if s and not s.startswith("{")]
    # collapse {id}-style trailing params into a noun
    segs = [s for s in segs if "{" not in s]
    if not segs:
        segs = ["repo"]
    # top-level group
    group = segs[0]
    # verb by model/method
    if model in ("direct_read", "binary_read"):
        verb = "download" if model == "binary_read" else "view"
        if group in ("tarball", "zipball"):
            return f"repo archive {group}"
        if "logs" in p:
            return f"run logs view"
        if "sbom" in p:
            return f"repo sbom {'fetch' if 'fetch' in p else 'generate' if 'generate' in p else 'view'}"
        if "artifacts" in p:
            return "artifact download"
        if "sarifs" in p:
            return "code-sanning upload" if method == "POST" else "code-scanning sarifs view"
        # default: <group> view (or <group> <leaf> view)
        leaf = segs[-1] if len(segs) > 1 else group
        if leaf == group:
            return f"{group} view"
        return f"{group} {leaf} view" if len(segs) <= 3 else f"{group} {segs[-1]} view"
    else:  # admin / sensitive / destructive writes
        verb = {"PATCH": "update", "POST": "create", "PUT": "set", "DELETE": "delete"}.get(method, "update")
        if "permissions" in p:
            return f"{group} permissions {verb}" if verb != "update" else f"{group} permissions update"
        if "secrets" in p and "public-key" not in p:
            return f"secret {verb}" if verb in ("set", "delete") else f"secret {verb}"
        if "variables" in p:
            return f"variable {verb}"
        if "hooks" in p:
            return f"webhook {verb}"
        leaf = segs[-1]
        if leaf == group:
            return f"{group} {verb}"
        return f"{group} {leaf} {verb}" if len(segs) <= 3 else f"{group} {segs[-1]} {verb}"

def make_op_id(group, api_path):
    base = slugify(api_path.replace("/repos/{owner}/{repo}", ""))
    base = re.sub(r"\{[^}]+\}", "", base).strip("_")
    base = re.sub(r"_+", "_", base)
    if not base:
        base = "repo"
    return f"github.{base}"

new_ops, new_writes, new_cmds = [], [], []
covered_added = 0
seen_op, seen_write, seen_cli = set(), set(), set()
seen_cli_path = set()

for e in api["endpoints"]:
    op = e.get("operation")
    if not op:
        continue
    model = op.get("model")
    if model in ("duplicate", "deprecated", "disallowed"):
        continue
    if e.get("covered_by"):
        continue  # already converted
    e.pop("operation", None)  # covered_by replaces the operation classifier
    method = e.get("method", "GET")
    api_path = e.get("path", "")
    params = path_params(api_path)
    # derive names
    cli_path = derive_cli_path(method, api_path, model)
    # de-dup cli paths
    base_cli = cli_path
    i = 2
    while cli_path in seen_cli_path or cli_path in existing_cli_paths:
        cli_path = f"{base_cli}-{i}"; i += 1
    seen_cli_path.add(cli_path)

    op_id = make_op_id(model, api_path)
    base_op = op_id
    i = 2
    while op_id in existing_op_ids or op_id in seen_op:
        op_id = f"{base_op}{i}"; i += 1
    seen_op.add(op_id)

    # direct_read coverage must be GET; a non-GET endpoint in a read tier is an upload/generator -> gated admin write
    if method != "GET" and model in ("direct_read", "binary_read"):
        model = "admin_reverse_etl"
    is_write = model in ("admin_reverse_etl", "sensitive_reverse_etl", "destructive_action")
    is_binary = model == "binary_read"

    if is_write:
        # write action
        wname = op_id.replace("github.", "").replace(".", "_")
        base_w = wname
        i = 2
        while wname in existing_write_names or wname in seen_write:
            wname = f"{base_w}{i}"; i += 1
        seen_write.add(wname)
        path_fields = [p for p in params if p not in ("owner", "repo")]
        record_props = {}
        required = []
        for pf in path_fields:
            record_props[pf] = {"type": "integer"} if "id" in pf or "number" in pf else {"type": "string"}
            required.append(pf)
        wa = {
            "name": wname, "kind": {"POST": "create", "PATCH": "update", "PUT": "update", "DELETE": "delete"}[method],
            "method": method, "path": api_path, "path_fields": path_fields, "body_type": "json",
            "record_schema": {"$schema": "http://json-schema.org/draft-07/schema#", "type": "object",
                              "required": required, "properties": record_props},
            "risk": op.get("risk", "medium"),
        }
        mut_class = "admin"
        if model == "sensitive_reverse_etl":
            mut_class = "secret"
            if "secrets" in api_path and "public-key" not in api_path:
                wa["hook"] = "github"
        elif model == "destructive_action":
            mut_class = "destructive"
            wa["confirm"] = "destructive"
        new_writes.append(wa)
        # operation carries the gate metadata (mutation_class/sensitive_policy/destructive)
        oprec = {"id": op_id, "kind": "rest_write", "summary": f"{method} {api_path}",
                 "source_url": op.get("source_url", ""), "risk": op.get("risk", "medium"),
                 "approval": "plan, preview, approval, execute (typed confirmation)" if model != "admin_reverse_etl" else "plan, preview, approval, execute",
                 "output_policy": "json", "mutation_class": mut_class,
                 "rest": {"method": method, "path": api_path}}
        if model == "sensitive_reverse_etl":
            oprec["secret_sensitive"] = True
            oprec["sensitive_policy"] = {"input_mode": "env", "redact_fields": ["value", "encrypted_value"],
                                         "transform": "github_secret_encryption" if "secrets" in api_path else "none",
                                         "approval_mode": "typed_confirmation"}
        elif model == "destructive_action":
            oprec["destructive"] = True
        new_ops.append(oprec)
        # cli command
        flags = [{"name": p.replace("_", "-"), "type": "integer" if ("id" in p or "number" in p) else "string",
                  "summary": f"{p} path parameter", "maps_to": f"record.{p}"} for p in path_fields]
        cmd = {"path": cli_path, "summary": f"{method} {api_path}", "intent": "reverse_etl",
               "availability": "implemented", "write": wname, "source_cli_path": "",
               "risk": op.get("risk", "medium"),
               "approval": "Reverse ETL writes require plan, preview, approval, execute.",
               "flags": flags}
        if model == "destructive_action":
            cmd["notes"] = "destructive; requires --allow-destructive + typed confirmation"
        new_cmds.append(cmd)
        e["covered_by"] = {"write": wname}
    elif is_binary:
        new_ops.append({"id": op_id, "kind": "binary_download", "summary": f"Download {api_path}",
                        "source_url": op.get("source_url", ""), "risk": op.get("risk", "medium"),
                        "approval": "none", "output_policy": "binary",
                        "binary": {"method": "GET", "path": api_path, "max_bytes": 104857600,
                                   "allow_overwrite": False, "extract_archives": "tarball" in api_path or "zipball" in api_path}})
        flags = [{"name": p.replace("_", "-"), "type": "string", "summary": f"{p} path parameter"} for p in params if p not in ("owner", "repo")]
        new_cmds.append({"path": cli_path, "summary": f"Download {api_path}", "intent": "direct_read",
                         "availability": "implemented", "operation": op_id, "source_cli_path": "",
                         "flags": flags})
        e["covered_by"] = {"direct_reads": [cli_path]}
    else:  # direct_read
        new_ops.append({"id": op_id, "kind": "rest_read", "summary": f"Read {api_path}",
                        "source_url": op.get("source_url", ""), "risk": op.get("risk", "low"),
                        "approval": "none", "output_policy": "json",
                        "rest": {"method": "GET", "path": api_path}})
        flags = [{"name": p.replace("_", "-"), "type": "integer" if ("id" in p or "number" in p) else "string",
                  "summary": f"{p} path parameter"} for p in params if p not in ("owner", "repo")]
        cmd = {"path": cli_path, "summary": f"Read {api_path}", "intent": "direct_read",
               "availability": "implemented", "operation": op_id, "source_cli_path": "", "flags": flags}
        new_cmds.append(cmd)
        e["covered_by"] = {"direct_reads": [cli_path]}
    covered_added += 1

ops_doc["operations"].extend(new_ops)
writes_doc["actions"].extend(new_writes)
cli_doc["commands"].extend(new_cmds)

json.dump(ops_doc, open(f"{ROOT}/operations.json", "w"), indent=2)
json.dump(writes_doc, open(f"{ROOT}/writes.json", "w"), indent=2)
json.dump(cli_doc, open(f"{ROOT}/cli_surface.json", "w"), indent=2)
json.dump(api, open(f"{ROOT}/api_surface.json", "w"), indent=2)

print(f"generated: {covered_added} endpoints covered")
print(f"  new operations: {len(new_ops)} (total {len(ops_doc['operations'])})")
print(f"  new write actions: {len(new_writes)} (total {len(writes_doc['actions'])})")
print(f"  new cli commands: {len(new_cmds)} (total {len(cli_doc['commands'])})")
