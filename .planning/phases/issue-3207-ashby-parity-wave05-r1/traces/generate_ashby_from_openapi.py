#!/usr/bin/env python3
from __future__ import annotations
import json, re, urllib.request
from collections import OrderedDict
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[4]
DEFS = ROOT / "internal/connectors/defs/ashby"
NATIVE = ROOT / "internal/connectors/native/ashby"
DOC_URL = "https://developers.ashbyhq.com/reference/candidateaddtag"
REVIEWED_AT = "2026-08-01"
HTTP_METHODS = {"get", "post", "put", "patch", "delete"}
LEGACY_STREAM_NAMES = {"candidate.list":"candidates","job.list":"jobs","application.list":"applications","user.list":"users"}
DIRECT_READ_SUMMARIES = {"candidate.search","job.search","opening.search","project.search","user.search","file.info","notetakerTranscript.info"}
BINARY_BLOCKED_SUMMARIES = {"file.createFileUploadHandle"}
BINARY_JSON_WRITE_SUMMARIES = {"candidate.uploadResume","candidate.uploadFile"}
CDC_STREAM_SUMMARIES = {"auditLog.list","application.listHistory"}
CDC_WRITE_SUMMARIES = {"application.updateHistory"}
PAGINATION_FIELDS = {"cursor","limit","syncToken"}
DESTRUCTIVE_WORDS = ("delete","remove","archive","cancel","reject","close","disable","restore","anonymize")
REDACT_MARKERS = ("id","email","file","resume","handle","url","name","secret")
REQUIRED_ANY_FIELDS = {
    "application.info": ["applicationId", "submittedFormInstanceId"],
    "candidate.info": ["id", "externalMappingId"],
}
DIRECT_READ_MIN_PROPERTIES = {"job.search": 1}
SIGNED_URL_DIRECT_READS = {"file.info", "notetakerTranscript.info"}
SIGNED_URL_DIRECT_READ_RISK = "bounded JSON direct read; credential-marked response fields are redacted, and Ashby signed URL fields are preserved (results.url/results.transcriptUrl) in trusted live local output"


def fetch_schema():
    req = urllib.request.Request(DOC_URL, headers={"User-Agent":"Mozilla/5.0"})
    html = urllib.request.urlopen(req, timeout=45).read().decode()
    m = re.search(r'<script id="ssr-props" type="application/x-ssr-props">([\s\S]*?)</script>', html)
    if not m:
        raise SystemExit("ReadMe ssr-props JSON not found")
    data = json.loads(m.group(1))
    return data["document"]["api"]["schema"]


def snake(name: str) -> str:
    name = re.sub(r"\s*\(.*?\)\s*", "", name).replace(".", "_").replace("-", "_").replace("/", "_")
    name = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", name)
    name = re.sub(r"([A-Z]+)([A-Z][a-z])", r"\1_\2", name)
    name = re.sub(r"[^A-Za-z0-9_]+", "_", name)
    name = re.sub(r"_+", "_", name).strip("_").lower() or "operation"
    if not re.match(r"^[a-z]", name): name = "op_" + name
    return name

def kebab(name: str) -> str: return snake(name).replace("_", "-")
def clean_summary(summary: str) -> str: return re.sub(r"\s*\(Implemented by Partner\)\s*", "", summary).strip()
def ref_slug(op_id: str, summary: str) -> str: return re.sub(r"[^a-z0-9]", "", (op_id or clean_summary(summary)).lower())

def deref(schema: dict[str, Any], ref: str) -> Any:
    obj: Any = schema
    for part in ref.split("/")[1:]: obj = obj[part]
    return obj

def resolve(schema: dict[str, Any], node: Any, seen: set[str] | None = None) -> Any:
    seen = seen or set()
    while isinstance(node, dict) and "$ref" in node:
        ref = node["$ref"]
        if ref in seen: break
        seen.add(ref); node = deref(schema, ref)
    return node

def choose_variant(schema: dict[str, Any], node: Any) -> Any:
    node = resolve(schema, node)
    if isinstance(node, dict):
        for key in ("oneOf","anyOf"):
            if key in node and node[key]:
                variants = [resolve(schema, item) for item in node[key]]
                for item in variants:
                    if isinstance(item, dict) and ("properties" in item or item.get("type") == "object"):
                        return item
                return variants[0]
    return node

def json_type(schema: dict[str, Any], node: Any) -> list[str]:
    node = choose_variant(schema, node)
    if not isinstance(node, dict): return []
    typ = node.get("type")
    if isinstance(typ, list): return [t for t in typ if t in {"string","number","integer","boolean","object","array","null"}]
    if isinstance(typ, str) and typ in {"string","number","integer","boolean","object","array","null"}: return [typ]
    if "properties" in node or "additionalProperties" in node: return ["object"]
    if "items" in node: return ["array"]
    if "enum" in node: return ["string"]
    return []

def union_json_types(schema: dict[str, Any], node: Any) -> list[str]:
    node = resolve(schema, node)
    if not isinstance(node, dict): return []
    variants = None
    for key in ("oneOf", "anyOf"):
        if isinstance(node.get(key), list) and node[key]:
            variants = node[key]
            break
    if not variants: return []
    out: list[str] = []
    for variant in variants:
        resolved = resolve(schema, variant)
        if not isinstance(resolved, dict): return []
        types = json_type(schema, resolved)
        if not types: return []
        for typ in types:
            if typ not in out: out.append(typ)
    return out

def to_draft_schema(schema: dict[str, Any], node: Any, *, close_object: bool=False, depth: int=0) -> dict[str, Any]:
    source = resolve(schema, node)
    union_types = union_json_types(schema, source)
    preserve_union = len(union_types) > 1
    node = choose_variant(schema, node)
    if not isinstance(node, dict): return {}
    out: dict[str, Any] = {}
    if preserve_union:
        out["type"] = union_types
    else:
        types = union_types or json_type(schema, node)
        if types: out["type"] = types[0] if len(types)==1 else types
    if isinstance(source, dict) and isinstance(source.get("description"), str): out["description"] = source["description"][:500]
    elif isinstance(node.get("description"), str): out["description"] = node["description"][:500]
    if node.get("format") in {"date-time","date","uri","email"}: out["format"] = node["format"]
    if isinstance(node.get("enum"), list) and len(node["enum"]) <= 100: out["enum"] = node["enum"]
    props = node.get("properties") if isinstance(node.get("properties"), dict) else None
    if props is not None and not preserve_union:
        out.setdefault("type", "object")
        out["properties"] = {n: to_draft_schema(schema, p, close_object=False, depth=depth+1) for n,p in props.items()}
        req = [r for r in node.get("required", []) if isinstance(r, str)]
        if req: out["required"] = req
        if close_object and depth == 0: out["additionalProperties"] = False
    elif "items" in node and not preserve_union:
        out.setdefault("type", "array")
        out["items"] = to_draft_schema(schema, node.get("items"), close_object=False, depth=depth+1)
    elif not out:
        out["type"] = ["string","number","integer","boolean","object","array","null"]
    return out

def request_schema(schema, op):
    body = op.get("requestBody",{}).get("content",{}).get("application/json",{}).get("schema")
    return choose_variant(schema, body or {"type":"object","properties":{}})

def response_result_schema(schema, op):
    body = op.get("responses",{}).get("200",{}).get("content",{}).get("application/json",{}).get("schema")
    sch = choose_variant(schema, body or {})
    props = sch.get("properties",{}) if isinstance(sch, dict) else {}
    results = choose_variant(schema, props.get("results", {}))
    if isinstance(results, dict) and json_type(schema, results) == ["array"]: return choose_variant(schema, results.get("items",{})), True
    if isinstance(results, dict) and results: return results, False
    return sch if isinstance(sch, dict) else {"type":"object","properties":{}}, False

def properties_of(schema, node) -> OrderedDict[str, dict[str,Any]]:
    node = choose_variant(schema, node); props: OrderedDict[str,dict[str,Any]] = OrderedDict()
    if isinstance(node, dict) and isinstance(node.get("properties"), dict):
        for name, sub in node["properties"].items(): props[name] = to_draft_schema(schema, sub)
    if not props: props["value"] = {"type":["string","number","integer","boolean","object","array","null"]}
    return props

def request_properties(schema, op):
    node = choose_variant(schema, request_schema(schema, op))
    props: OrderedDict[str, dict[str, Any]] = OrderedDict()
    if isinstance(node, dict) and isinstance(node.get("properties"), dict):
        for name, sub in node["properties"].items():
            props[name] = to_draft_schema(schema, sub)
    return props

def required_fields(schema, op):
    sch = request_schema(schema, op)
    return [r for r in sch.get("required", []) if isinstance(r, str)] if isinstance(sch, dict) else []

def choose_pk(props):
    if "id" in props: return "id"
    for n in props:
        if n.endswith("Id") or n.endswith("ID") or n.lower()=="email": return n
    return next(iter(props), "_ashby_row_id")

def name_words(name: str) -> list[str]:
    return [w.lower() for w in re.findall(r"[A-Z]?[a-z]+|[A-Z]+(?![a-z])|\d+", name.replace("_", " ").replace("-", " "))]

def is_time_like_name(name: str) -> bool:
    lower = name.lower()
    if lower in {"date", "timestamp"}: return True
    words = name_words(name)
    if not words: return False
    if words[-1] in {"date", "time", "timestamp", "after", "before"}: return True
    return words[-1] == "at" and len(words) > 1

def cursor_field_supported(name: str, prop_schema: dict[str, Any]) -> bool:
    typ = prop_schema.get("type") if isinstance(prop_schema, dict) else None
    types = {typ} if isinstance(typ, str) else set(typ or []) if isinstance(typ, list) else set()
    types.discard("null")
    if types and types <= {"object", "array"}: return False
    if isinstance(prop_schema, dict) and prop_schema.get("format") in {"date-time", "date"}: return True
    return is_time_like_name(name)

def choose_cursor(props):
    for n in ("updatedAt","createdAt","submittedAt","completedAt","sentAt","date","timestamp"):
        if n in props and cursor_field_supported(n, props[n]): return n
    for n, prop_schema in props.items():
        if cursor_field_supported(n, prop_schema): return n
    return None

def result_record_schema(schema, stream_name, op):
    node,_ = response_result_schema(schema, op); props = properties_of(schema, node)
    pk = choose_pk(props); synthetic=[]
    if pk not in props:
        props[pk] = {"type":"string","description":"Synthetic stable key added by the Ashby connector when the response object has no documented id field."}; synthetic=[pk]
    cursor = choose_cursor(props)
    out = {"$schema":"http://json-schema.org/draft-07/schema#","title":stream_name,"type":"object","properties":props,"x-primary-key":[pk]}
    if cursor: out["x-cursor-field"] = cursor
    return out, cursor, synthetic

def cli_flag_type(prop_schema):
    typ = prop_schema.get("type")
    types = typ if isinstance(typ, list) else [typ]
    types = [t for t in types if t != "null"]
    type_set = set(types)
    if prop_schema.get("enum"):
        return "enum"
    if len(type_set) > 1:
        if type_set <= {"string", "number", "integer", "boolean"} and "string" in type_set:
            return "string"
        if type_set <= {"integer", "number"}:
            return "integer"
        return None
    if "boolean" in types:
        return "boolean"
    if "integer" in types or "number" in types:
        return "integer"
    if "string" in types or not types:
        return "string"
    if "array" in types:
        return "string_array"
    return None

def field_type_for_cli(prop_schema):
    return cli_flag_type(prop_schema) or "string"

def field_flag(name, maps_to, prop_schema):
    ftype = cli_flag_type(prop_schema)
    if ftype is None:
        return None
    flag_name = kebab(name)
    if flag_name in {"body", "path", "query", "method", "raw"}:
        flag_name = "message-" + flag_name
    flag={"name":flag_name,"type":ftype,"maps_to":maps_to}
    if prop_schema.get("description"): flag["summary"] = str(prop_schema["description"])[:180]
    if flag["type"] == "enum" and isinstance(prop_schema.get("enum"), list): flag["values"] = [str(v) for v in prop_schema["enum"] if v is not None][:50]
    if flag["type"] == "string" and is_time_like_name(name): flag["format"]="date-time"
    if flag["type"] == "string": flag["allow_empty"] = False
    return flag

def append_flag(flags, seen_targets, label, maps_to, prop_schema):
    if maps_to in seen_targets:
        return
    flag = field_flag(label, maps_to, prop_schema)
    if flag is None:
        return
    base = flag["name"]
    used_names = {f["name"] for f in flags}
    if base in used_names:
        i = 2
        while f"{base}-{i}" in used_names:
            i += 1
        flag["name"] = f"{base}-{i}"
    flags.append(flag)
    seen_targets.add(maps_to)

def flags_for_props(props, maps_prefix, skip=frozenset()):
    flags=[]; seen=set()
    for n,fs in props.items():
        if n in skip:
            continue
        append_flag(flags, seen, n, f"{maps_prefix}.{n}", fs)
    return flags

def schema_child(node, part):
    if not isinstance(node, dict):
        return {}
    if part == "0" and node.get("type") == "array":
        return node.get("items", {})
    return node.get("properties", {}).get(part, {})

def schema_at_path(node, dotted):
    cur=node
    for part in dotted.split("."):
        cur=schema_child(cur, part)
    return cur if isinstance(cur, dict) else {}

def required_mapping_paths(node, prefix=""):
    if not isinstance(node, dict):
        return []
    if node.get("type") == "array":
        child_paths = required_mapping_paths(node.get("items", {}), prefix + (".0" if prefix else "0"))
        return child_paths or ([prefix] if prefix else [])
    req = [r for r in node.get("required", []) if isinstance(r, str)]
    paths=[]
    for r in req:
        path = f"{prefix}.{r}" if prefix else r
        child = schema_child(node, r)
        if child.get("type") == "array":
            nested = required_mapping_paths(child, path)
        elif child.get("type") == "object" or child.get("properties"):
            nested = required_mapping_paths(child, path)
        else:
            nested = []
        paths.extend(nested or [path])
    return paths

def write_cli_flags(record_schema):
    flags=[]; seen=set(); complex_required=False; required_roots=set()
    required_paths = required_mapping_paths(record_schema)
    for path in required_paths:
        required_roots.add(path.split(".", 1)[0])
        leaf = schema_at_path(record_schema, path)
        before=len(flags)
        append_flag(flags, seen, path.replace(".", "-"), f"record.{path}", leaf)
        if len(flags) == before:
            complex_required=True
    for name, fs in record_schema.get("properties", {}).items():
        if name in required_roots:
            continue
        append_flag(flags, seen, name, f"record.{name}", fs)
    mapped={f["maps_to"].removeprefix("record.") for f in flags}
    for path in required_paths:
        if path not in mapped:
            complex_required=True
    return flags, complex_required

def operation_parts(summary):
    c=clean_summary(summary)
    if "." not in c: return c,"run"
    return tuple(c.split(".",1))
def unique_name(name, used):
    base=name; i=2
    while name in used:
        name=f"{base}_{i}"; i+=1
    used.add(name); return name
def stream_name_for(summary, used):
    c=clean_summary(summary)
    if c in LEGACY_STREAM_NAMES: name=LEGACY_STREAM_NAMES[c]
    else:
        r,a=operation_parts(c); name=snake(r+"_"+a)
    return unique_name(name, used)
def write_name_for(summary, used):
    r,a=operation_parts(summary); words=snake(a).split("_"); verb=words[0] if words else "run"; rest="_".join(words[1:]); pieces=[verb,snake(r)]
    if rest: pieces.append(rest)
    return unique_name("_".join(pieces), used)
def operation_id_for(prefix, name, used): return unique_name("ashby."+snake(prefix+"_"+name).replace("_","."), used)
def command_path(summary, used):
    r,a=operation_parts(summary); base=f"{kebab(r)} {kebab(a)}"; p=base; i=2
    while p in used:
        p=f"{base}-{i}"; i+=1
    used.add(p); return p

def is_partner(summary): return "Implemented by Partner" in summary
def is_read_stream(summary):
    c=clean_summary(summary)
    return (c in CDC_STREAM_SUMMARIES or any(t in c for t in (".list",".info",".fetch",".synchronous"))) and c not in DIRECT_READ_SUMMARIES and not is_partner(summary) and c not in BINARY_BLOCKED_SUMMARIES
def is_direct(summary): return clean_summary(summary) in DIRECT_READ_SUMMARIES and not is_partner(summary)
def is_write(summary):
    c=clean_summary(summary)
    if is_partner(summary): return False
    if c in BINARY_JSON_WRITE_SUMMARIES or c in CDC_WRITE_SUMMARIES: return True
    if c in DIRECT_READ_SUMMARIES or c in BINARY_BLOCKED_SUMMARIES or c in CDC_STREAM_SUMMARIES: return False
    if any(t in c for t in (".list",".info",".fetch",".synchronous",".search")): return False
    return True

def write_kind(summary):
    a=snake(operation_parts(summary)[1])
    if a.startswith(("create","add","start","submit","send")): return "create"
    if a.startswith(("update","change","transfer","restore","set")): return "update"
    if a.startswith(("delete","remove","cancel","archive")): return "delete"
    return "custom"
def is_destructive(summary): return any(w in snake(operation_parts(summary)[1]).split("_") for w in DESTRUCTIVE_WORDS)
def redact_fields(props, destructive=False): return [n for n in props if any(m in n.lower() for m in REDACT_MARKERS)][:40]
def default_body_for(props): return {"limit":100} if "limit" in props else {}
def source_url(op):
    slug=ref_slug(op.get("operationId",""), op.get("summary","")); return f"https://developers.ashbyhq.com/reference/{slug}" if slug else DOC_URL

def write_json(path: Path, data: Any):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, indent=2, ensure_ascii=False)+"\n")

def go_quote(s: str) -> str: return json.dumps(s)
def go_string_slice(items): return "[]string{" + ", ".join(go_quote(x) for x in items) + "}"
def go_map_string(items): return "map[string]string{" + ", ".join(go_quote(k)+":"+go_quote(v) for k,v in sorted(items.items())) + "}"

def main():
    schema=fetch_schema()
    ops=[]
    for path, methods in schema.get("paths",{}).items():
        for method, op in methods.items():
            if method.lower() in HTTP_METHODS:
                op=dict(op); op["method"]=method.upper(); op["path"]=path; ops.append(op)
    used_streams=set(); used_writes=set(); used_commands=set(); used_ops=set()
    stream_entries=[]; stream_go=[]; write_actions=[]; operation_entries=[]; api_rows=[]; cli_commands=[]; stream_schemas={}
    priority={"candidate.list":0,"job.list":1,"application.list":2,"user.list":3}
    ops_sorted=sorted(enumerate(ops), key=lambda pair:(priority.get(clean_summary(pair[1].get("summary","")),100), pair[0]))
    for _,op in ops_sorted:
        summary=op.get("summary",""); clean=clean_summary(summary); method=op["method"]; path=op["path"]
        req_props=request_properties(schema, op); req_required=required_fields(schema, op); req_sch=request_schema(schema, op)
        req_types={n:(json_type(schema, req_sch.get("properties",{}).get(n,{})) or ["string"])[0] if isinstance(req_sch,dict) else "string" for n in req_props}
        if is_read_stream(summary):
            sname=stream_name_for(summary, used_streams); rec_schema,cursor,synthetic=result_record_schema(schema, sname, op); stream_schemas[sname]=rec_schema
            se={"name":sname,"method":method,"path":path,"records":{"path":"results"},"schema":f"schemas/{sname}.json","projection":"passthrough"}
            if cursor: se["incremental"]={"cursor_field":cursor,"client_filtered":True}
            catalog_fields=[]
            for fname, fschema in rec_schema.get("properties", {}).items():
                typ=fschema.get("type")
                if isinstance(typ, list):
                    typ=next((t for t in typ if t != "null"), "string")
                if typ not in {"string","integer","number","boolean","array","object"}:
                    typ="string"
                catalog_fields.append({"name":fname,"type":typ})
            required_any = REQUIRED_ANY_FIELDS.get(clean, [])
            stream_entries.append(se); stream_go.append({"name":sname,"path":path,"required":req_required,"required_any":required_any,"fields":list(req_props.keys()),"field_types":req_types,"cursor":cursor or "","synthetic":synthetic,"catalog_fields":catalog_fields,"primary_key":rec_schema.get("x-primary-key", [])})
            flags=flags_for_props(req_props, "query", PAGINATION_FIELDS)
            cpath=command_path(clean, used_commands)
            notes=f"Fixed Ashby stream for {clean}; flags map only to documented request body fields."
            if required_any:
                notes += " Requires at least one documented selector: " + ", ".join(required_any) + "."
            cli_commands.append({"path":cpath,"summary":op.get("description","")[:160] or clean,"intent":"etl","availability":"implemented","stream":sname,"source_url":source_url(op),"flags":flags,"api_surface":[{"method":method,"path":path}],"notes":notes})
            api_rows.append({"method":method,"path":path,"covered_by":{"stream":sname}}); continue
        if is_direct(summary):
            cpath=command_path(clean, used_commands); opid=operation_id_for("direct", clean, used_ops); body_schema=to_draft_schema(schema, req_sch, close_object=True)
            if clean in DIRECT_READ_MIN_PROPERTIES:
                body_schema["minProperties"] = DIRECT_READ_MIN_PROPERTIES[clean]
            operation_entries.append({"id":opid,"kind":"rest_read","summary":clean,"description":op.get("description","")[:1000],"source_url":source_url(op),"risk":"medium" if ("file" in clean.lower() or "transcript" in clean.lower()) else "low","approval":"none","output_policy":"json_redacted","rest":{"method":method,"path":path,"content_type":"application/json","max_bytes":1048576,"body":default_body_for(req_props),"body_schema":body_schema}})
            flags=flags_for_props(req_props, "body", PAGINATION_FIELDS)
            direct_risk = SIGNED_URL_DIRECT_READ_RISK if clean in SIGNED_URL_DIRECT_READS else "bounded JSON direct read; response fields with secret/download markers are redacted"
            cli_commands.append({"path":cpath,"summary":op.get("description","")[:160] or clean,"intent":"direct_read","availability":"implemented","operation":opid,"source_url":source_url(op),"flags":flags,"api_surface":[{"method":method,"path":path}],"output_policy":"json_redacted","redact_fields":redact_fields(req_props),"risk":direct_risk,"approval":"none","notes":"Fixed Ashby POST direct read; no raw method/path/body override is exposed."})
            api_rows.append({"method":method,"path":path,"covered_by":{"direct_read":cpath}}); continue
        if is_write(summary):
            wname=write_name_for(clean, used_writes); req_schema=to_draft_schema(schema, req_sch, close_object=True); destructive=is_destructive(clean)
            action={"name":wname,"kind":write_kind(clean),"method":method,"path":path,"record_schema":req_schema,"risk":f"Executes Ashby {clean} through the documented {method} {path} endpoint; reverse ETL plan, preview, approval, and execute are required.","redact_fields":redact_fields(req_props, destructive)}
            if destructive: action["confirm"]="destructive"
            write_actions.append(action)
            flags, complex_required = write_cli_flags(req_schema)
            availability = "partial" if complex_required else "implemented"
            note = "Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed."
            if complex_required:
                note += " This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution."
            cpath=command_path(clean, used_commands)
            cli_commands.append({"path":cpath,"summary":op.get("description","")[:160] or clean,"intent":"reverse_etl","availability":availability,"write":wname,"source_url":source_url(op),"flags":flags,"api_surface":[{"method":method,"path":path}],"redact_fields":action["redact_fields"],"risk":action["risk"],"approval":"reverse ETL writes require plan -> preview -> explicit approval -> execute","notes":note})
            api_rows.append({"method":method,"path":path,"covered_by":{"write":wname}}); continue
        if is_partner(summary):
            reason="official OpenAPI marks this as implemented by an assessment partner and called by Ashby; the CLI connector is a client of Ashby, not an inbound partner API server"; model="disallowed"; risk="medium"
        elif clean in BINARY_BLOCKED_SUMMARIES:
            reason="creates a presigned upload workflow whose useful result is an external file-transfer URL; blocked until a reviewed bounded binary/file workflow can return and consume the handle without exposing raw file passthrough"; model="binary_read"; risk="high"
        else:
            reason="official webhook/changefeed surface requires a shared webhook/CDC receiver and state foundation (#2986/#2988), not a pull-style stream or direct read"; model="disallowed"; risk="medium"
        api_rows.append({"method":method,"path":path,"operation":{"model":model,"status":"blocked","risk":risk,"blocked_by_default":True,"reason":reason,"source_url":source_url(op),"notes":clean}})
    for webhook_name, webhook in schema.get("webhooks",{}).items():
        post=webhook.get("post",{}) if isinstance(webhook, dict) else {}
        api_rows.append({"method":"WEBHOOK","path":f"webhook:{webhook_name}","operation":{"model":"disallowed","status":"blocked","risk":"medium","blocked_by_default":True,"reason":"official OpenAPI webhook object documents an inbound event callback; implementation requires shared webhook/CDC receiver and state foundations (#2986/#2988)","source_url":DOC_URL,"notes":post.get("summary", webhook_name)}})

    for sub in (DEFS/"schemas",):
        if sub.exists():
            for f in sub.glob("*.json"): f.unlink()
    write_json(DEFS/"metadata.json", {"name":"ashby","display_name":"Ashby","description":"Reads Ashby applicant-tracking REST resources and exposes reviewed reverse-ETL/direct-read surfaces from the official Ashby OpenAPI. Fixture-only; not live-certified.","integration_type":"api","release_stage":"alpha","capabilities":{"check":True,"read":True,"write":True,"query":False,"cdc":False,"dynamic_schema":False},"batch":{"read_page_size":100,"write_batch_size":1},"risk":{"read":"bounded Ashby POST reads using documented endpoints, Basic API-key auth, page-size and max-pages bounds, and redacted fixtures","write":"named reverse-ETL actions only; no generic HTTP method/path/body; destructive actions require typed confirmation","approval":"reverse ETL writes require plan -> preview -> explicit approval -> execute"},"conformance":{"skip_dynamic":True,"reason":"Ashby list streams require POST body cursor pagination and many generated write/direct surfaces; native Ashby tests plus connectorgen/static validation cover the connector-local engine delegation while fixture replay remains credential-free and no live calls are made."},"docs_url":"https://developers.ashbyhq.com/"})
    write_json(DEFS/"spec.json", {"$schema":"http://json-schema.org/draft-07/schema#","title":"Ashby Connection Specification","type":"object","required":["api_key","start_date"],"properties":{"api_key":{"type":"string","x-secret":True,"description":"Ashby API key. Provide from an environment variable or stdin; never inline in prompts or docs."},"start_date":{"type":"string","format":"date-time","description":"Lower bound used by client-side incremental filtering when a stream has a timestamp cursor."},"base_url":{"type":"string","default":"https://api.ashbyhq.com","description":"Ashby API base URL; override only for local tests."},"page_size":{"type":"string","default":"100","description":"Per-page body limit for Ashby list endpoints; bounded to 1..100 by the native connector."},"max_pages":{"type":"string","default":"1","description":"Maximum pages to read per stream. Use 0, all, or unlimited for an exhaustive read."},"mode":{"type":"string","description":"Set to fixture for credential-free native tests."}}})
    write_json(DEFS/"streams.json", {"base":{"url":"{{ config.base_url }}","user_agent":"polymetrics-go-cli","headers":{"Accept":"application/json; version=1","Content-Type":"application/json"},"auth":[{"mode":"basic","username":"{{ secrets.api_key }}","password":""}],"check":{"method":"POST","path":"/apiKey.info"},"pagination":{"type":"none"}},"streams":stream_entries})
    for name, sch in stream_schemas.items(): write_json(DEFS/"schemas"/f"{name}.json", sch)
    write_json(DEFS/"writes.json", {"actions":write_actions})
    write_json(DEFS/"operations.json", {"operations":operation_entries})
    write_json(DEFS/"api_surface.json", {"api":"Official Ashby developer ReadMe OpenAPI 3.1 schema embedded in the public reference page","docs":DOC_URL,"reviewed_at":REVIEWED_AT,"operation_ledger_version":1,"scope":"Complete public Ashby inventory: REST operations plus OpenAPI webhook events. Supported REST read/write/direct surfaces are fixed and typed; inbound partner/webhook and presigned external file-transfer workflows remain blocked with source-backed reasons.","endpoints":api_rows})
    grouped={"streams":[],"direct":[],"writes":[]}
    for c in cli_commands:
        if c["intent"]=="etl": grouped["streams"].append(c["path"])
        elif c["intent"]=="direct_read": grouped["direct"].append(c["path"])
        elif c["intent"]=="reverse_etl": grouped["writes"].append(c["path"])
    write_json(DEFS/"cli_surface.json", {"tagline":"Ashby applicant-tracking connector with typed REST streams, bounded direct reads, and gated reverse-ETL writes.","usage":"pm connectors command ashby <command> [flags]","source_cli":{"name":"Ashby Public API","docs":"https://developers.ashbyhq.com/","reference":DOC_URL,"source":"public ReadMe OpenAPI"},"groups":[{"id":"streams","title":"ETL streams","commands":grouped["streams"]},{"id":"direct_reads","title":"Bounded direct reads","commands":grouped["direct"]},{"id":"reverse_etl","title":"Reverse ETL writes","commands":grouped["writes"]}],"commands":cli_commands,"help_topics":[{"name":"ashby safety","summary":"Ashby writes are named, schema-validated actions only; reverse ETL must use plan, preview, explicit approval, and execute."},{"name":"ashby parity","summary":"Public Ashby OpenAPI coverage ledger is recorded in api_surface.json with blocked webhook/partner/binary workflow reasons."}]})
    write_json(DEFS/"certification.json", {"schema_version":1,"source":{"default_stream":"candidates","live_unavailable":[{"kind":"no_credentials_requested","contains":["No live Ashby credentials or provider calls were requested for issue #3207 wave05-r1."]}]},"direct_read_candidates":[{"stage_name":"candidate_search_fixture_shape","command":"candidate search","args":[{"literal":"--email"},{"literal":"candidate@example.invalid"}]}],"write_pairings":[]})
    fixture_dir=DEFS/"fixtures/streams/candidates"; fixture_dir.mkdir(parents=True, exist_ok=True)
    write_json(fixture_dir/"page_1.json", {"success":True,"results":[{"id":"candidate_fixture_1","name":"Fixture Candidate","primaryEmailAddress":{"value":"fixture@example.invalid"},"primaryPhoneNumber":{"value":"+15550100"},"company":"Example Inc","title":"Engineer","locationSummary":"Remote","timezone":"UTC","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-02T00:00:00Z"}],"moreDataAvailable":False})
    docs = f"""# Ashby Connector\n\n## Overview\n\nAshby is an applicant-tracking connector generated from the public Ashby ReadMe OpenAPI reference ({DOC_URL}). The parity ledger was reviewed on {REVIEWED_AT}.\n\nCoverage summary:\n\n- REST operations in source: {len(ops)}\n- OpenAPI webhook events in source: {len(schema.get('webhooks',{}))}\n- Implemented ETL/changefeed streams: {len(stream_entries)}\n- Implemented bounded direct reads/search/file metadata operations: {len(operation_entries)}\n- Implemented reverse-ETL write actions: {len(write_actions)}\n- Reverse-ETL CLI commands with scalar flags: {sum(1 for c in cli_commands if c.get('intent') == 'reverse_etl' and c.get('availability') == 'implemented')}; partial nested-object flag surfaces: {sum(1 for c in cli_commands if c.get('intent') == 'reverse_etl' and c.get('availability') == 'partial')}\n- Blocked/non-executable ledger rows: {len(api_rows) - len(stream_entries) - len(operation_entries) - len(write_actions)}\n\n## Auth setup\n\nAuthentication uses Ashby's documented HTTP Basic API-key flow: the API key is the username and the password is blank. Provide keys via environment variables or stdin only; never paste secrets into prompts, docs, commits, or issue comments.\n\n## Streams notes\n\nAshby list and info reads are fixed POST endpoints with documented body fields only. The native connector owns Ashby's cursor-in-body pagination, applies page-size and max-pages bounds, and supports client-side incremental filtering when a documented cursor field exists.\n\n## Write actions & risks\n\nReverse ETL writes are typed action names with closed top-level JSON schemas and the normal plan → preview → explicit approval → execute gate. No command exposes a raw HTTP method, raw path, arbitrary request body, raw query, shell, file, SQL, or passthrough escape hatch. The public Ashby OpenAPI did not document an Idempotency-Key or equivalent idempotency header for these actions, so no provider idempotency key is claimed.\n\n## Known limits\n\nBlocked rows are still documented in `api_surface.json`: inbound assessment-partner APIs and webhook events are not pull-executable by a CLI connector, and `file.createFileUploadHandle` remains blocked until a reviewed bounded binary/file workflow can safely return and consume presigned upload handles. The current wave is fixture/static validated only; no live Ashby credentials or provider calls were used.\n"""
    (DEFS/"docs.md").write_text(docs)

    go_lines=[]
    go_lines.append("package ashby\n\n// Code generated by .planning/phases/issue-3207-ashby-parity-wave05-r1/traces/generate_ashby_from_openapi.py; DO NOT EDIT.\n\n")
    go_lines.append("import \"polymetrics.ai/internal/connectors\"\n\n")
    go_lines.append("var ashbyStreamEndpoints = map[string]streamEndpoint{\n")
    for e in stream_go:
        ft={k:(v if v in {"string","integer","number","boolean","array","object"} else "string") for k,v in e["field_types"].items()}
        catalog="[]connectors.Field{" + ", ".join("{Name: "+go_quote(f["name"])+", Type: "+go_quote(f["type"])+"}" for f in e["catalog_fields"]) + "}"
        required_any = f", requiredAnyFields: {go_string_slice(e['required_any'])}" if e.get("required_any") else ""
        go_lines.append(f"\t{go_quote(e['name'])}: {{path: {go_quote(e['path'].lstrip('/'))}, requestFields: {go_map_string(ft)}, requiredFields: {go_string_slice(e['required'])}{required_any}, cursorField: {go_quote(e['cursor'])}, syntheticFields: {go_string_slice(e['synthetic'])}, primaryKey: {go_string_slice(e['primary_key'])}, fields: {catalog}}},\n")
    go_lines.append("}\n")
    (NATIVE/"streams_gen.go").write_text("".join(go_lines))
    print(json.dumps({"rest":len(ops),"webhooks":len(schema.get('webhooks',{})),"streams":len(stream_entries),"writes":len(write_actions),"direct_reads":len(operation_entries),"blocked":len(api_rows)-len(stream_entries)-len(write_actions)-len(operation_entries)}, indent=2))

if __name__ == "__main__": main()
