#!/usr/bin/env python3
"""Generate GitLab full operation parity scaffolding from api_surface.json.

This intentionally mirrors scripts/gen-github-parity.py:
  - non-deprecated operation-ledger rows become covered_by entries;
  - GET/HEAD/binary rows become operation-backed direct_read commands;
  - mutating rows become typed reverse-ETL write actions and commands;
  - deprecated rows remain blocked operation-ledger rows.

Execution remains gated by existing runtime policy:
  - operation-backed direct reads are feature-gated until an operation executor exists;
  - generated writes execute only through reverse ETL plan -> preview -> approval -> execute;
  - no raw generic GitLab API command, arbitrary GraphQL mutation, or shell/SQL escape hatch is added.
"""
import json
import re
from pathlib import Path

ROOT = Path("internal/connectors/defs/gitlab")
api = json.loads((ROOT / "api_surface.json").read_text())
ops_doc = json.loads((ROOT / "operations.json").read_text())
cli_doc = json.loads((ROOT / "cli_surface.json").read_text())
writes_path = ROOT / "writes.json"
if writes_path.exists():
    writes_doc = json.loads(writes_path.read_text())
else:
    writes_doc = {"actions": []}

existing_write_names = {a["name"] for a in writes_doc.get("actions", [])}
existing_cli_paths = {c["path"] for c in cli_doc.get("commands", [])}
existing_ops_by_endpoint = {}
COMPOSITE_ENDPOINT_RE = re.compile(r"(?:evaluate|operation)\s+(GET|POST|PUT|PATCH|DELETE|HEAD)\s+([^\s]+)")
for op in ops_doc.get("operations", []):
    method = ""
    path = ""
    for block_name in ("rest", "binary"):
        block = op.get(block_name) or {}
        if block.get("method") and block.get("path"):
            method = block["method"].upper()
            path = block["path"]
            break
    if not method and op.get("composite"):
        for step in op.get("composite", {}).get("steps", []):
            match = COMPOSITE_ENDPOINT_RE.search(step)
            if match:
                method = match.group(1).upper()
                path = match.group(2)
                break
    if method and path:
        existing_ops_by_endpoint[(method, path)] = op

PATH_PARAM_RE = re.compile(r"\{([^}]+)\}")

def path_params(path: str) -> list[str]:
    return PATH_PARAM_RE.findall(path)


def slugify(value: str) -> str:
    return re.sub(r"[^a-z0-9]+", "_", value.lower()).strip("_")


def path_to_template(path: str) -> str:
    def repl(match: re.Match[str]) -> str:
        name = match.group(1)
        return "{{ record." + name + " }}"
    return PATH_PARAM_RE.sub(repl, path)


def field_type(name: str) -> dict:
    lower = name.lower()
    if any(marker in lower for marker in ("id", "iid", "number", "page", "per_page", "attempt", "pipeline_id", "job_id")):
        return {"type": ["integer", "string"]}
    return {"type": "string"}


def operation_for(method: str, path: str, model: str) -> dict:
    op = existing_ops_by_endpoint.get((method, path))
    if op:
        return op
    # HEAD endpoints were represented as composite metadata in the first ledger.
    op_id = "gitlab." + slugify(method + "_" + path)
    return {
        "id": op_id,
        "kind": "composite",
        "summary": f"{method} {path}",
        "source_url": "https://docs.gitlab.com/ee/api/rest/",
        "risk": "low" if model == "direct_read" else "medium",
        "approval": "none",
        "output_policy": "json_redacted",
        "composite": {"steps": [f"feature-gated GitLab operation {method} {path}"]},
    }


def derive_cli_path(method: str, api_path: str, model: str) -> str:
    p = api_path.strip("/")
    segs = [s for s in p.split("/") if s and not s.startswith("{")]
    segs = [s.replace("_", "-") for s in segs if "{" not in s]
    if not segs:
        segs = ["gitlab"]
    group = segs[0]
    leaf = segs[-1]
    lowered = p.lower()
    if model == "binary_read":
        verb = "download"
    elif method == "HEAD":
        verb = "check"
    elif method in ("GET",):
        verb = "view"
    else:
        verb = {"POST": "create", "PUT": "set", "PATCH": "update", "DELETE": "delete"}.get(method, "update")
    if "variables" in lowered:
        group = "variable"
    elif "secrets" in lowered or "tokens" in lowered or "token" in lowered:
        group = "token"
    elif "members" in lowered or "access" in lowered:
        group = "access"
    elif "runners" in lowered:
        group = "runner"
    elif "pipelines" in lowered:
        group = "pipeline"
    elif "merge_requests" in lowered or "merge-requests" in lowered:
        group = "mr"
    elif "issues" in lowered:
        group = "issue"
    elif "repository" in lowered:
        group = "repo"
    elif "groups" in lowered:
        group = "group"
    elif "projects" in lowered:
        group = "project"
    parts = [group]
    if leaf != group and leaf not in {"projects", "groups", "issues", "merge_requests", "repository"}:
        parts.append(leaf)
    parts.append(verb)
    return " ".join(parts)


def unique(value: str, existing: set[str], seen: set[str]) -> str:
    candidate = value
    i = 2
    while candidate in existing or candidate in seen:
        candidate = f"{value}-{i}"
        i += 1
    seen.add(candidate)
    return candidate


def make_write_name(op_id: str, method: str, path: str, seen: set[str]) -> str:
    base = op_id.removeprefix("gitlab.")
    if not base:
        base = slugify(method + "_" + path)
    base = base.replace(".", "_")
    candidate = base
    i = 2
    while candidate in existing_write_names or candidate in seen:
        candidate = f"{base}{i}"
        i += 1
    seen.add(candidate)
    return candidate


def write_kind(method: str) -> str:
    return {"POST": "create", "PUT": "update", "PATCH": "update", "DELETE": "delete"}.get(method, "custom")


def mutation_class(model: str, method: str) -> str:
    if model == "destructive_action" or method == "DELETE":
        return "destructive"
    if model == "sensitive_reverse_etl":
        return "secret"
    if model == "admin_reverse_etl":
        return "admin"
    return "update"


def is_write_model(model: str, method: str) -> bool:
    return model in {"admin_reverse_etl", "sensitive_reverse_etl", "destructive_action"} or method in {"POST", "PUT", "PATCH", "DELETE"}

seen_cli: set[str] = set()
seen_writes: set[str] = set()
new_commands = []
new_writes = []
converted = 0

for endpoint in api.get("endpoints", []):
    if endpoint.get("covered_by"):
        continue
    op_meta = endpoint.get("operation")
    if not op_meta:
        continue
    model = op_meta.get("model", "")
    if model in {"deprecated", "duplicate", "disallowed"}:
        continue

    method = endpoint.get("method", "GET").upper()
    api_path = endpoint.get("path", "")
    params = path_params(api_path)
    op = operation_for(method, api_path, model)
    op_id = op["id"]
    cli_path = unique(derive_cli_path(method, api_path, model), existing_cli_paths, seen_cli)

    if is_write_model(model, method):
        write_name = make_write_name(op_id, method, api_path, seen_writes)
        required = list(params)
        props = {name: field_type(name) for name in params}
        action = {
            "name": write_name,
            "kind": write_kind(method),
            "method": method,
            "path": path_to_template(api_path),
            "path_fields": params,
            "body_type": "json",
            "record_schema": {
                "$schema": "http://json-schema.org/draft-07/schema#",
                "type": "object",
                "required": required,
                "properties": props,
            },
            "risk": op_meta.get("reason") or op.get("risk", "medium"),
        }
        if mutation_class(model, method) == "destructive":
            action["confirm"] = "destructive"
        new_writes.append(action)
        flags = [
            {
                "name": name.replace("_", "-"),
                "type": "string",
                "summary": f"{name} path parameter",
                "maps_to": f"record.{name}",
            }
            for name in params
        ]
        cmd = {
            "path": cli_path,
            "summary": f"{method} {api_path}",
            "intent": "reverse_etl",
            "availability": "implemented",
            "write": write_name,
            "source_cli_path": "",
            "risk": op_meta.get("reason") or op.get("risk", "medium"),
            "approval": "Reverse ETL writes require plan, preview, approval, execute.",
            "flags": flags,
        }
        if mutation_class(model, method) in {"destructive", "secret", "admin"}:
            cmd["notes"] = "typed operation; use reverse ETL preview and approval before execution"
        new_commands.append(cmd)
        endpoint.pop("operation", None)
        endpoint["covered_by"] = {"write": write_name}
    else:
        flags = [
            {
                "name": name.replace("_", "-"),
                "type": "string",
                "summary": f"{name} path parameter",
            }
            for name in params
        ]
        cmd = {
            "path": cli_path,
            "summary": f"{method} {api_path}",
            "intent": "direct_read",
            "availability": "implemented",
            "operation": op_id,
            "source_cli_path": "",
            "flags": flags,
        }
        new_commands.append(cmd)
        endpoint.pop("operation", None)
        endpoint["covered_by"] = {"direct_reads": [cli_path]}
    converted += 1

writes_doc["actions"].extend(new_writes)
cli_doc["commands"].extend(new_commands)
api["scope"] = (
    "Full GitLab OpenAPI parity scaffold. All non-deprecated official GitLab REST operations "
    "are covered by existing streams/direct reads, operation-backed read/binary/HEAD commands, "
    "or typed reverse-ETL write actions. Operation-backed read commands remain feature-gated until "
    "their executor exists; generated writes remain gated by reverse ETL plan, preview, approval, execute. "
    "Deprecated operations remain blocked-by-default operation-ledger rows. The /users compatibility "
    "stream row is retained because the current OpenAPI snapshot omits GET /users."
)

(ROOT / "api_surface.json").write_text(json.dumps(api, indent=2) + "\n")
(ROOT / "writes.json").write_text(json.dumps(writes_doc, indent=2) + "\n")
(ROOT / "cli_surface.json").write_text(json.dumps(cli_doc, indent=2) + "\n")

print(f"converted endpoints: {converted}")
print(f"new write actions: {len(new_writes)}")
print(f"new commands: {len(new_commands)}")
print(f"total commands: {len(cli_doc['commands'])}")
print(f"total write actions: {len(writes_doc['actions'])}")
