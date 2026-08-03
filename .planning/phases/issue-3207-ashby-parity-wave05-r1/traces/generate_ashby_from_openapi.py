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
REFERENCE_ROOT = "https://developers.ashbyhq.com/reference"
REVIEWED_AT = "2026-08-01"
HTTP_METHODS = {"get", "post", "put", "patch", "delete"}
LEGACY_STREAM_NAMES = {"candidate.list":"candidates","job.list":"jobs","application.list":"applications","user.list":"users"}
DIRECT_READ_SUMMARIES = {"candidate.search","job.search","opening.search","project.search","user.search","file.info","notetakerTranscript.info","user.interviewerSettings","report.generate"}
BINARY_BLOCKED_SUMMARIES = {"file.createFileUploadHandle"}
BINARY_JSON_WRITE_SUMMARIES = {"candidate.uploadResume","candidate.uploadFile"}
CDC_STREAM_SUMMARIES = {"auditLog.list","application.listHistory"}
CDC_WRITE_SUMMARIES = {"application.updateHistory"}
NON_INCREMENTAL_STREAM_SUMMARIES = {"application.listHistory"}
PAGINATION_FIELDS = {"cursor","limit","syncToken"}
SYNC_TOKEN_FOUNDATION = "ashby-sync-token-checkpoint-foundation"
STREAM_ARRAY_FOUNDATION = "connector-stream-repeatable-array-foundation"
DOCUMENTED_WRITE_MAP_SCHEMAS = {
    "create_department": {("properties", "extraData"): True},
    "update_department": {("properties", "extraData"): True},
    "create_location": {("properties", "extraData"): True},
    "create_interview_schedule": {("properties", "interviewEvents", "items", "properties", "extraData"): True},
    "create_survey_submission": {("properties", "submittedValues"): True},
}
DESTRUCTIVE_WORDS = ("delete","remove","archive","cancel","discard","reject","close","disable","restore","anonymize")
REDACT_MARKERS = ("id","email","file","resume","handle","url","name","secret")
REQUIRED_ANY_FIELDS = {
    "application.info": ["applicationId", "submittedFormInstanceId"],
    "candidate.info": ["id", "externalMappingId"],
}
DIRECT_READ_MIN_PROPERTIES = {"job.search": 1}
SIGNED_URL_DIRECT_READS = {"file.info", "notetakerTranscript.info"}
SIGNED_URL_DIRECT_READ_RISK = "bounded JSON direct read; credential-marked response fields are redacted, and Ashby signed URL fields are preserved (results.url/results.transcriptUrl) in trusted live local output"
DIRECT_READ_RISK_OVERRIDES = {
    "report.generate": "bounded JSON direct read that starts or polls a documented Ashby report generation and returns at most 1 MiB of redacted JSON; the connector does not fetch returned report URLs or poll automatically",
}
DIRECT_READ_COMMAND_SUMMARY_OVERRIDES = {
    "report.generate": "Start an Ashby report generation or check an existing request.",
}
BLOCKED_OPERATION_RULES = {
    "referralForm.info": {
        "model": "admin_reverse_etl",
        "risk": "medium",
        "reason": "fetches the default referral form but conditionally creates one when absent; blocked pending ashby-referral-form-info-side-effect-foundation so an apparent read cannot mutate Ashby without a typed confirmation path",
    },
    "applicationForm.submit": {
        "model": "sensitive_reverse_etl",
        "risk": "high",
        "reason": "requires multipart/form-data with typed file parts for application fields; blocked pending ashby-application-form-typed-multipart-foundation because the Ashby write executor is JSON-only",
    },
}
FIXED_STREAM_REQUEST_FIELDS = {"hiringTeamRole.list": {"namesOnly": "true"}}
FIXED_STREAM_REQUEST_FIELD_GAPS = {"hiringTeamRole.list": {"namesOnly": "ashby_hiring_team_role_list_names_only_false"}}
FIXED_STREAM_REQUEST_FIELD_NOTES = {"hiringTeamRole.list": "Defaults to namesOnly=true role-title results. namesOnly=false object results are blocked pending variant-schema foundation ashby_hiring_team_role_list_names_only_false."}
SCALAR_RESULT_STREAMS = {"hiring_team_role_list"}


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
def terminal_summary(description: str, fallback: str) -> str:
    source_was_truncated = len(description) >= 160
    text = description.strip() or fallback
    text = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", text)
    text = re.sub(r"`([^`]*)`", r"\1", text)
    text = re.sub(r"\((?:ref:|authentication#)[^)]*$", "", text)
    text = re.sub(r"(?m)^\s*>\s?", "", text)
    text = text.replace("**", "").replace("__", "")
    text = text.translate(str.maketrans("", "", "[]`*"))
    text = re.sub(r"\s+", " ", text).strip()
    needs_abbreviation = source_was_truncated or (text and text[-1] not in ".!?")
    if len(text) <= 160 and not needs_abbreviation:
        return text
    candidate = text[:161]
    sentence_end = max(candidate.rfind(". "), candidate.rfind("! "), candidate.rfind("? "))
    if sentence_end >= 40:
        return candidate[:sentence_end + 1]
    word_end = candidate[:157].rfind(" ")
    if word_end < 40:
        word_end = 157
    return candidate[:word_end].rstrip(" ,;:-") + "..."
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
    if props is not None:
        out.setdefault("type", "object")
        out["properties"] = {n: to_draft_schema(schema, p, close_object=close_object, depth=depth+1) for n,p in props.items()}
        req = [r for r in node.get("required", []) if isinstance(r, str)]
        if req: out["required"] = req
        if close_object:
            additional = node.get("additionalProperties", False)
            if isinstance(additional, dict):
                out["additionalProperties"] = {} if not additional else to_draft_schema(schema, additional, close_object=True, depth=depth+1)
            else:
                out["additionalProperties"] = bool(additional)
    elif "items" in node:
        out.setdefault("type", "array")
        out["items"] = to_draft_schema(schema, node.get("items"), close_object=close_object, depth=depth+1)
    elif close_object and out.get("type") == "object":
        additional = node.get("additionalProperties", False)
        if isinstance(additional, dict):
            out["additionalProperties"] = {} if not additional else to_draft_schema(schema, additional, close_object=True, depth=depth+1)
        else:
            out["additionalProperties"] = bool(additional)
    elif not out:
        out["type"] = ["string","number","integer","boolean","object","array","null"]
    return out

def preserve_documented_write_maps(action_name: str, schema: dict[str, Any]) -> None:
    for path, additional_properties in DOCUMENTED_WRITE_MAP_SCHEMAS.get(action_name, {}).items():
        node: Any = schema
        for part in path:
            if not isinstance(node, dict) or part not in node:
                node = None
                break
            node = node[part]
        if isinstance(node, dict):
            node["additionalProperties"] = additional_properties

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

def result_record_schema(schema, stream_name, op):
    node,_ = response_result_schema(schema, op); props = properties_of(schema, node)
    pk = choose_pk(props); synthetic=[]
    if pk not in props:
        props[pk] = {"type":"string","description":"Synthetic stable key added by the Ashby connector when the response object has no documented id field."}; synthetic=[pk]
    out = {"$schema":"http://json-schema.org/draft-07/schema#","title":stream_name,"type":"object","properties":props,"x-primary-key":[pk]}
    return out, None, synthetic

def cli_flag_type(prop_schema):
    typ = prop_schema.get("type")
    types = typ if isinstance(typ, list) else [typ]
    types = [t for t in types if t != "null"]
    if prop_schema.get("enum"):
        return "enum"
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
    if flag["type"] == "string" and ("date" in name.lower() or name.lower().endswith("at") or name.lower().endswith("after") or name.lower().endswith("before")): flag["format"]="date-time"
    if flag["type"] == "string": flag["allow_empty"] = False
    return flag

def append_flag(flags, seen_targets, label, maps_to, prop_schema, required=False):
    if maps_to in seen_targets:
        return
    flag = field_flag(label, maps_to, prop_schema)
    if flag is None:
        return
    if required:
        flag["required"] = True
    base = flag["name"]
    used_names = {f["name"] for f in flags}
    if base in used_names:
        i = 2
        while f"{base}-{i}" in used_names:
            i += 1
        flag["name"] = f"{base}-{i}"
    flags.append(flag)
    seen_targets.add(maps_to)

def flags_for_props(props, maps_prefix, skip=frozenset(), required=frozenset()):
    flags=[]; seen=set()
    for n,fs in props.items():
        if n in skip:
            continue
        append_flag(flags, seen, n, f"{maps_prefix}.{n}", fs, n in required)
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
    return (c in CDC_STREAM_SUMMARIES or any(t in c for t in (".list",".info",".fetch",".synchronous"))) and c not in DIRECT_READ_SUMMARIES and c not in BLOCKED_OPERATION_RULES and not is_partner(summary) and c not in BINARY_BLOCKED_SUMMARIES
def is_direct(summary): return clean_summary(summary) in DIRECT_READ_SUMMARIES and not is_partner(summary)
def is_write(summary):
    c=clean_summary(summary)
    if is_partner(summary): return False
    if c in BINARY_JSON_WRITE_SUMMARIES or c in CDC_WRITE_SUMMARIES: return True
    if c in DIRECT_READ_SUMMARIES or c in BLOCKED_OPERATION_RULES or c in BINARY_BLOCKED_SUMMARIES or c in CDC_STREAM_SUMMARIES: return False
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

def fixture_type(node: Any) -> str:
    if not isinstance(node, dict): return "string"
    typ = node.get("type")
    if isinstance(typ, list):
        values = [t for t in typ if t != "null"]
        for preferred in ("object", "array", "string", "integer", "number", "boolean"):
            if preferred in values: return preferred
        return values[0] if values else "string"
    if isinstance(typ, str): return typ
    if "properties" in node: return "object"
    if "items" in node: return "array"
    return "string"

def fixture_slug(value: str) -> str:
    value = re.sub(r"[^A-Za-z0-9]+", "_", value).strip("_").lower()
    return value or "ashby"

def fixture_title(value: str) -> str:
    return re.sub(r"[_\-]+", " ", value).strip().title() or "Ashby"

def fixture_value(node: Any, field: str, stream: str, index: int, depth: int = 0) -> Any:
    node = node if isinstance(node, dict) else {}
    enum = node.get("enum")
    if isinstance(enum, list):
        for item in enum:
            if item is not None: return item
    typ = fixture_type(node)
    lower = field.lower()
    stream_slug = fixture_slug(stream)
    field_slug = fixture_slug(field)
    if typ == "object":
        props = node.get("properties") if isinstance(node.get("properties"), dict) else {}
        if props and depth < 20:
            return {name: fixture_value(prop, name, stream, index, depth+1) for name, prop in props.items()}
        return {} if node.get("additionalProperties") is False else {"id": f"{stream_slug}_{field_slug}_fixture_{index}"}
    if typ == "array":
        item = node.get("items") if isinstance(node.get("items"), dict) else {"type": "string"}
        return [fixture_value(item, field.rstrip("s") or field, stream, index, depth+1)]
    if typ == "integer": return index
    if typ == "number": return float(index)
    if typ == "boolean": return index % 2 == 1
    if lower == "timezone": return "UTC"
    day = ((index - 1) % 28) + 1
    if node.get("format") == "date": return f"2026-01-{day:02d}"
    if node.get("format") == "date-time" or lower in {"date", "timestamp"} or lower.endswith(("at", "date", "time")) or "timestamp" in lower:
        return f"2026-01-{day:02d}T0{index % 10}:00:00Z"
    if node.get("format") == "uri" or any(marker in lower for marker in ("url", "link", "profile")):
        return f"https://example.invalid/ashby/{stream_slug}/{field_slug}/{index}"
    if node.get("format") == "email" or "email" in lower:
        return f"{stream_slug}.{field_slug}.{index}@example.invalid"
    if "phone" in lower: return f"+1555010{index:02d}"
    if lower == "name" or lower.endswith("name"): return f"{fixture_title(stream)} {fixture_title(field)} Fixture {index}"
    if lower in {"title", "text", "description", "content", "comment", "reasoning", "failurereason"} or lower.endswith(("title", "text", "description", "content", "comment")):
        return f"{fixture_title(field)} fixture {index}"
    if "html" in lower: return f"<p>{fixture_title(field)} fixture {index}</p>"
    if lower == "status" or lower.endswith("status"): return "active"
    if lower == "type" or lower.endswith("type"): return "fixture"
    if lower == "ipaddress": return f"192.0.2.{index}"
    if lower == "useragent": return "polymetrics-fixture/1.0"
    if lower == "value": return f"{stream_slug}_value_fixture_{index}"
    if lower == "code": return f"{stream_slug.upper()}_{index:03d}"
    if lower == "mimetype": return "application/json"
    if lower == "id" or lower.endswith("id") or lower.endswith("ids") or "handle" in lower:
        return f"{stream_slug}_{field_slug}_fixture_{index}"
    return f"{stream_slug}_{field_slug}_fixture_{index}"

def fixture_record(stream_name: str, schema: dict[str, Any], index: int) -> dict[str, Any]:
    props = schema.get("properties", {}) if isinstance(schema.get("properties"), dict) else {}
    record = {field: fixture_value(prop, field, stream_name, index) for field, prop in props.items()}
    for pk in schema.get("x-primary-key", []):
        if record.get(pk) in (None, ""):
            record[pk] = f"{fixture_slug(stream_name)}_{fixture_slug(pk)}_fixture_{index}"
    cursor = schema.get("x-cursor-field")
    if cursor:
        record[cursor] = f"2026-02-{((index - 1) % 28) + 1:02d}T{index % 24:02d}:00:00Z"
    return record

def fixture_results_are_array(path: str, stream_name: str) -> bool:
    clean = path.strip("/").lower()
    return clean in {"candidate.list", "job.list", "application.list", "user.list"} or ".list" in clean or stream_name.endswith("_list")

def fixture_request_value(kind: str, field: str) -> str:
    if kind in {"integer", "number"}: return "1"
    if kind == "boolean": return "true"
    if kind == "array": return "[]"
    if kind == "object": return "{}"
    return f"{fixture_slug(field)}_fixture"

def fixture_read_query(shape: dict[str, Any] | None) -> dict[str, str]:
    if not shape: return {}
    fields = list(shape.get("required", []))
    required_any = shape.get("required_any", [])
    if required_any: fields.append(required_any[0])
    field_types = shape.get("field_types", {})
    return {field: fixture_request_value(field_types.get(field, "string"), field) for field in fields}

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
        fixed_request_fields = FIXED_STREAM_REQUEST_FIELDS.get(clean, {})
        fixed_request_field_gaps = FIXED_STREAM_REQUEST_FIELD_GAPS.get(clean, {})
        req_required = [field for field in req_required if field not in fixed_request_fields]
        req_types={n:(json_type(schema, req_sch.get("properties",{}).get(n,{})) or ["string"])[0] if isinstance(req_sch,dict) else "string" for n in req_props}
        if clean in BLOCKED_OPERATION_RULES:
            rule = BLOCKED_OPERATION_RULES[clean]
            api_rows.append({"method":method,"path":path,"operation":{"model":rule["model"],"status":"blocked","risk":rule["risk"],"blocked_by_default":True,"reason":rule["reason"],"source_url":source_url(op),"notes":clean}})
            continue
        if is_read_stream(summary):
            sname=stream_name_for(summary, used_streams); rec_schema,cursor,synthetic=result_record_schema(schema, sname, op)
            if sname == "hiring_team_role_list":
                rec_schema={"$schema":"http://json-schema.org/draft-07/schema#","title":sname,"type":"object","properties":{"value":{"type":"string"}},"x-primary-key":["value"]}; cursor=None; synthetic=[]
            stream_cursor = cursor
            if clean in NON_INCREMENTAL_STREAM_SUMMARIES:
                stream_cursor = None
                rec_schema.pop("x-cursor-field", None)
            stream_schemas[sname]=rec_schema
            se={"name":sname,"method":method,"path":path,"records":{"path":"results"},"schema":f"schemas/{sname}.json","projection":"passthrough"}
            if stream_cursor: se["incremental"]={"cursor_field":stream_cursor,"client_filtered":True}
            catalog_fields=[]
            for fname, fschema in rec_schema.get("properties", {}).items():
                typ=fschema.get("type")
                if isinstance(typ, list):
                    typ=next((t for t in typ if t != "null"), "string")
                if typ not in {"string","integer","number","boolean","array","object"}:
                    typ="string"
                catalog_fields.append({"name":fname,"type":typ})
            required_any = REQUIRED_ANY_FIELDS.get(clean, [])
            stream_entries.append(se); stream_go.append({"name":sname,"path":path,"required":req_required,"required_any":required_any,"fixed":fixed_request_fields,"fixed_gaps":fixed_request_field_gaps,"fields":list(req_props.keys()),"field_types":req_types,"cursor":stream_cursor or "","synthetic":synthetic,"catalog_fields":catalog_fields,"primary_key":rec_schema.get("x-primary-key", [])})
            flags=flags_for_props(req_props, "query", set(PAGINATION_FIELDS) | set(fixed_request_fields), set(req_required))
            cpath=command_path(clean, used_commands)
            notes=f"Fixed Ashby stream for {clean}; flags map only to documented request body fields."
            if clean in FIXED_STREAM_REQUEST_FIELD_NOTES:
                notes = f"Fixed Ashby stream for {clean}; " + FIXED_STREAM_REQUEST_FIELD_NOTES[clean]
            if required_any:
                notes += " Requires at least one documented selector: " + ", ".join(required_any) + "."
            array_flags = [flag for flag in flags if flag.get("type") == "string_array"]
            if array_flags:
                blocked_names = ", ".join("--" + flag["name"] for flag in array_flags)
                flags = [flag for flag in flags if flag.get("type") != "string_array"]
                notes += f" Repeatable array request variants ({blocked_names}) are blocked pending {STREAM_ARRAY_FOUNDATION}."
            if "syncToken" in req_props:
                notes += f" Opaque syncToken checkpointing is blocked pending {SYNC_TOKEN_FOUNDATION}; this stream is full-refresh only."
            command_summary = terminal_summary(op.get("description", ""), clean)
            if "syncToken" in req_props:
                command_summary = f"Full-refresh-only Ashby {clean} read. Opaque syncToken checkpointing is unavailable pending {SYNC_TOKEN_FOUNDATION}."
            cli_commands.append({"path":cpath,"summary":command_summary,"intent":"etl","availability":"implemented","stream":sname,"source_url":source_url(op),"flags":flags,"api_surface":[{"method":method,"path":path}],"notes":notes})
            api_rows.append({"method":method,"path":path,"covered_by":{"stream":sname}}); continue
        if is_direct(summary):
            cpath=command_path(clean, used_commands); opid=operation_id_for("direct", clean, used_ops); body_schema=to_draft_schema(schema, req_sch, close_object=True)
            if clean in DIRECT_READ_MIN_PROPERTIES:
                body_schema["minProperties"] = DIRECT_READ_MIN_PROPERTIES[clean]
            operation_entries.append({"id":opid,"kind":"rest_read","summary":clean,"description":op.get("description","")[:1000],"source_url":source_url(op),"risk":"medium" if ("file" in clean.lower() or "transcript" in clean.lower() or clean == "report.generate") else "low","approval":"none","output_policy":"json_redacted","rest":{"method":method,"path":path,"content_type":"application/json","max_bytes":1048576,"body":default_body_for(req_props),"body_schema":body_schema}})
            flags=flags_for_props(req_props, "body", PAGINATION_FIELDS, set(req_required))
            direct_risk = DIRECT_READ_RISK_OVERRIDES.get(clean, SIGNED_URL_DIRECT_READ_RISK if clean in SIGNED_URL_DIRECT_READS else "bounded JSON direct read; credential-marked response fields are redacted, and non-credential identity fields remain complete in trusted live local output")
            command_summary = terminal_summary(DIRECT_READ_COMMAND_SUMMARY_OVERRIDES.get(clean, op.get("description", "")), clean)
            cli_commands.append({"path":cpath,"summary":command_summary,"intent":"direct_read","availability":"implemented","operation":opid,"source_url":source_url(op),"flags":flags,"api_surface":[{"method":method,"path":path}],"output_policy":"json_redacted","risk":direct_risk,"approval":"none","notes":"Fixed Ashby POST direct read; no raw method/path/body override is exposed."})
            api_rows.append({"method":method,"path":path,"covered_by":{"direct_read":cpath}}); continue
        if is_write(summary):
            wname=write_name_for(clean, used_writes); req_schema=to_draft_schema(schema, req_sch, close_object=True); destructive=is_destructive(clean)
            preserve_documented_write_maps(wname, req_schema)
            risk = f"Executes Ashby {clean} through the documented {method} {path} endpoint; reverse ETL plan, preview, approval, and execute are required."
            if destructive:
                risk = f"Executes Ashby {clean} through the documented {method} {path} endpoint; reverse ETL plan, preview, explicit approval, typed destructive confirmation, and execute are required."
            action={"name":wname,"kind":write_kind(clean),"method":method,"path":path,"record_schema":req_schema,"risk":risk,"redact_fields":redact_fields(req_props, destructive)}
            if destructive: action["confirm"]="destructive"
            write_actions.append(action)
            flags, complex_required = write_cli_flags(req_schema)
            availability = "partial" if complex_required else "implemented"
            note = "Ashby OpenAPI contains no Idempotency-Key or idempotency header evidence; no provider idempotency key is claimed."
            if complex_required:
                note += " This command has nested object/array requirements that are implemented by the reverse-ETL action schema but are not fully expressible as scalar CLI flags; use file/warehouse reverse-ETL inputs for execution."
            cpath=command_path(clean, used_commands)
            cli_commands.append({"path":cpath,"summary":terminal_summary(op.get("description", ""), clean),"intent":"reverse_etl","availability":availability,"write":wname,"source_url":source_url(op),"flags":flags,"api_surface":[{"method":method,"path":path}],"redact_fields":action["redact_fields"],"risk":action["risk"],"approval":"reverse ETL writes require plan -> preview -> explicit approval -> execute","notes":note})
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
        api_rows.append({"method":"WEBHOOK","path":f"webhook:{webhook_name}","operation":{"model":"disallowed","status":"blocked","risk":"medium","blocked_by_default":True,"reason":"official OpenAPI webhook object documents an inbound event callback; implementation requires shared webhook/CDC receiver and state foundations (#2986/#2988)","source_url":source_url(post),"notes":post.get("summary", webhook_name)}})

    for sub in (DEFS/"schemas",):
        if sub.exists():
            for f in sub.glob("*.json"): f.unlink()
    write_json(DEFS/"metadata.json", {"name":"ashby","display_name":"Ashby","description":"Reads Ashby applicant-tracking REST resources and exposes reviewed reverse-ETL/direct-read surfaces from the official Ashby OpenAPI. Fixture-only; not live-certified.","integration_type":"api","release_stage":"alpha","capabilities":{"check":True,"read":True,"write":True,"query":False,"cdc":False,"dynamic_schema":False},"batch":{"read_page_size":100,"write_batch_size":1},"risk":{"read":"bounded Ashby POST reads using documented endpoints, Basic API-key auth, page-size and max-pages bounds, and sanitized replay fixtures","write":"named reverse-ETL actions only; no generic HTTP method/path/body; destructive actions require typed confirmation","approval":"reverse ETL writes require plan -> preview -> explicit approval -> execute"},"docs_url":"https://developers.ashbyhq.com/"})
    write_json(DEFS/"spec.json", {"$schema":"http://json-schema.org/draft-07/schema#","title":"Ashby Connection Specification","type":"object","required":["api_key"],"properties":{"api_key":{"type":"string","x-secret":True,"description":"Ashby API key. Provide from an environment variable or stdin; never inline in prompts or docs."},"base_url":{"type":"string","default":"https://api.ashbyhq.com","description":"Ashby API base URL; override only for local tests."},"page_size":{"type":"string","default":"100","description":"Per-page body limit for Ashby list endpoints; bounded to 1..100 by the native connector."},"max_pages":{"type":"string","default":"0","description":"Maximum pages to read per stream. Defaults to 0 for an exhaustive read; 0, all, and unlimited are equivalent."},"mode":{"type":"string","description":"Set to fixture for credential-free native tests."}}})
    write_json(DEFS/"streams.json", {"base":{"url":"{{ config.base_url }}","user_agent":"polymetrics-go-cli","headers":{"Accept":"application/json; version=1","Content-Type":"application/json"},"auth":[{"mode":"basic","username":"{{ secrets.api_key }}","password":""}],"check":{"method":"POST","path":"/apiKey.info"},"pagination":{"type":"none"}},"streams":stream_entries})
    for name, sch in stream_schemas.items(): write_json(DEFS/"schemas"/f"{name}.json", sch)
    write_json(DEFS/"writes.json", {"actions":write_actions})
    write_json(DEFS/"operations.json", {"operations":operation_entries})
    write_json(DEFS/"api_surface.json", {"api":"Official Ashby developer ReadMe OpenAPI 3.1 schema embedded in the public reference page","docs":REFERENCE_ROOT,"reviewed_at":REVIEWED_AT,"operation_ledger_version":1,"scope":"Complete public Ashby inventory: REST operations plus OpenAPI webhook events. Supported REST read/write/direct surfaces are fixed and typed; conditional side-effect reads, multipart submissions, inbound partner/webhook, and presigned external file-transfer workflows remain blocked with source-backed reasons.","endpoints":api_rows})
    grouped={"streams":[],"direct":[],"writes":[]}
    for c in cli_commands:
        if c["intent"]=="etl": grouped["streams"].append(c["path"])
        elif c["intent"]=="direct_read": grouped["direct"].append(c["path"])
        elif c["intent"]=="reverse_etl": grouped["writes"].append(c["path"])
    write_json(DEFS/"cli_surface.json", {"tagline":"Ashby applicant-tracking connector with typed REST streams, bounded direct reads, and gated reverse-ETL writes.","usage":"pm connectors command ashby <command> [flags]","source_cli":{"name":"Ashby Public API","docs":"https://developers.ashbyhq.com/","reference":REFERENCE_ROOT,"source":"public ReadMe OpenAPI"},"groups":[{"id":"streams","title":"ETL streams","commands":grouped["streams"]},{"id":"direct_reads","title":"Bounded direct reads","commands":grouped["direct"]},{"id":"reverse_etl","title":"Reverse ETL writes","commands":grouped["writes"]}],"commands":cli_commands,"help_topics":[{"name":"ashby safety","summary":"Ashby writes are named, schema-validated actions only; reverse ETL must use plan, preview, explicit approval, and execute."},{"name":"ashby parity","summary":"Public Ashby OpenAPI coverage ledger is recorded in api_surface.json with blocked webhook/partner/binary workflow reasons."}]})
    write_json(DEFS/"certification.json", {"schema_version":1,"source":{"default_stream":"candidates","live_unavailable":[{"kind":"no_credentials_requested","contains":["No live Ashby credentials or provider calls were requested for issue #3207 wave05-r1."]}]},"direct_read_candidates":[{"stage_name":"file_info_fixture_shape","command":"file info","args":[{"literal":"--file-handle"},{"literal":"file_handle_fixture"}]}],"write_pairings":[]})
    fixture_root=DEFS/"fixtures"
    stream_fixture_root=fixture_root/"streams"
    if stream_fixture_root.exists():
        for old in stream_fixture_root.glob("*/page_*.json"):
            old.unlink()
    request_shape_by_stream = {item["name"]: item for item in stream_go}
    for index, stream in enumerate(stream_entries, start=1):
        stream_name = stream["name"]
        record = fixture_record(stream_name, stream_schemas[stream_name], index)
        if stream_name in SCALAR_RESULT_STREAMS:
            body = {"success": True, "results": [record["value"]], "moreDataAvailable": False}
        else:
            body = {"success": True, "results": [record] if fixture_results_are_array(stream["path"], stream_name) else record}
            if fixture_results_are_array(stream["path"], stream_name): body["moreDataAvailable"] = False
        fixture = {"request":{"method":stream.get("method") or "GET","path":stream["path"],"query":{}},"response":{"status":200,"body":body}}
        read_query = fixture_read_query(request_shape_by_stream.get(stream_name))
        if read_query: fixture["read_query"] = read_query
        write_json(stream_fixture_root/stream_name/"page_1.json", fixture)
    check_record = fixture_record("api_key_info", stream_schemas.get("api_key_info", {"properties": {"title": {"type": "string"}, "createdAt": {"type": "string"}, "scopes": {"type": "array", "items": {"type": "string"}}}}), 1)
    check_record["title"] = "Ashby fixture key"
    check_record["createdAt"] = "2026-01-01T00:00:00Z"
    if "scopes" in check_record: check_record["scopes"] = ["fixture_scope"]
    write_json(fixture_root/"check.json", {"request":{"method":"POST","path":"/apiKey.info","query":{}},"response":{"status":200,"body":{"success":True,"results":check_record}}})
    docs = f"""# Ashby Connector\n\n## Overview\n\nAshby is an applicant-tracking connector generated from the public Ashby ReadMe OpenAPI reference ({REFERENCE_ROOT}). The parity ledger was reviewed on {REVIEWED_AT}.\n\nCoverage summary:\n\n- REST operations in source: {len(ops)}\n- OpenAPI webhook events in source: {len(schema.get('webhooks',{}))}\n- Implemented ETL/changefeed streams: {len(stream_entries)}\n- Implemented bounded direct reads/search/file metadata operations: {len(operation_entries)}\n- Implemented reverse-ETL write actions: {len(write_actions)}\n- Reverse-ETL CLI commands with scalar flags: {sum(1 for c in cli_commands if c.get('intent') == 'reverse_etl' and c.get('availability') == 'implemented')}; partial nested-object flag surfaces: {sum(1 for c in cli_commands if c.get('intent') == 'reverse_etl' and c.get('availability') == 'partial')}\n- Blocked/non-executable ledger rows: {len(api_rows) - len(stream_entries) - len(operation_entries) - len(write_actions)}\n\n## Auth setup\n\nAuthentication uses Ashby's documented HTTP Basic API-key flow: the API key is the username and the password is blank. Provide keys via environment variables or stdin only; never paste secrets into prompts, docs, commits, or issue comments.\n\n## Streams notes\n\nAshby list and info reads are fixed POST endpoints with documented body fields only. The native connector owns Ashby's cursor-in-body pagination and applies page-size, max-pages, and repeated-cursor bounds. Streams are full-refresh only until `{SYNC_TOKEN_FOUNDATION}` supplies an Ashby-owned persisted opaque-token state seam; timestamp fields are not used as lossy substitutes. Runtime help replaces provider incremental descriptions with full-refresh-only blocker text for every documented sync-token request. Repeatable array stream flags are withheld until `{STREAM_ARRAY_FOUNDATION}` preserves every supplied value.\n\n## Write actions & risks\n\nReverse ETL writes are typed action names with recursively closed modeled JSON schemas and the normal plan → preview → explicit approval → execute gate. Explicitly documented map-valued fields retain their map schemas. No command exposes a raw HTTP method, raw path, arbitrary request body, raw query, shell, file, SQL, or passthrough escape hatch. The public Ashby OpenAPI did not document an Idempotency-Key or equivalent idempotency header for these actions, so no provider idempotency key is claimed.\n\n## Known limits\n\nBlocked rows are still documented in `api_surface.json`: inbound assessment-partner APIs and webhook events are not pull-executable by a CLI connector, and `file.createFileUploadHandle` remains blocked until a reviewed bounded binary/file workflow can safely return and consume presigned upload handles. `referralForm.info` is blocked pending `ashby-referral-form-info-side-effect-foundation` because it conditionally creates a default form. `applicationForm.submit` is blocked pending `ashby-application-form-typed-multipart-foundation` because the documented request requires multipart form data and typed file parts. Opaque incremental state is blocked pending `{SYNC_TOKEN_FOUNDATION}`, and repeatable array stream-command variants are blocked pending `{STREAM_ARRAY_FOUNDATION}`. `hiringTeamRole.list` defaults to `namesOnly=true`; the `namesOnly=false` object-result variant is blocked pending variant-schema foundation `ashby_hiring_team_role_list_names_only_false`. Fixture replay covers every implemented stream with synthetic values only; no live Ashby credentials or provider calls were used.\n"""
    (DEFS/"docs.md").write_text(docs)

    go_lines=[]
    go_lines.append("package ashby\n\n// Code generated by .planning/phases/issue-3207-ashby-parity-wave05-r1/traces/generate_ashby_from_openapi.py; DO NOT EDIT.\n\n")
    go_lines.append("import \"polymetrics.ai/internal/connectors\"\n\n")
    go_lines.append("var ashbyStreamEndpoints = map[string]streamEndpoint{\n")
    for e in stream_go:
        ft={k:(v if v in {"string","integer","number","boolean","array","object"} else "string") for k,v in e["field_types"].items()}
        catalog="[]connectors.Field{" + ", ".join("{Name: "+go_quote(f["name"])+", Type: "+go_quote(f["type"])+"}" for f in e["catalog_fields"]) + "}"
        required_any = f", requiredAnyFields: {go_string_slice(e['required_any'])}" if e.get("required_any") else ""
        fixed_fields = f", fixedRequestFields: {go_map_string(e['fixed'])}" if e.get("fixed") else ""
        if e.get("fixed_gaps"):
            fixed_fields += f", fixedRequestFieldGaps: {go_map_string(e['fixed_gaps'])}"
        go_lines.append(f"\t{go_quote(e['name'])}: {{path: {go_quote(e['path'].lstrip('/'))}, requestFields: {go_map_string(ft)}{fixed_fields}, requiredFields: {go_string_slice(e['required'])}{required_any}, cursorField: {go_quote(e['cursor'])}, syntheticFields: {go_string_slice(e['synthetic'])}, primaryKey: {go_string_slice(e['primary_key'])}, fields: {catalog}}},\n")
    go_lines.append("}\n")
    (NATIVE/"streams_gen.go").write_text("".join(go_lines))
    print(json.dumps({"rest":len(ops),"webhooks":len(schema.get('webhooks',{})),"streams":len(stream_entries),"writes":len(write_actions),"direct_reads":len(operation_entries),"blocked":len(api_rows)-len(stream_entries)-len(write_actions)-len(operation_entries)}, indent=2))

if __name__ == "__main__": main()
