#!/usr/bin/env python3
import hashlib
import json
import os
import re
import shutil
import urllib.request
from collections import Counter
from pathlib import Path

ROOT = Path("internal/connectors/defs/google-ads")
PHASE = Path(".planning/phases/google-ads-parity-wave03-r1")
DISCOVERY_URL = "https://googleads.googleapis.com/$discovery/rest?version=v22"
DISCOVERY_CANONICAL_SHA256 = "c14a489015a3a4664addc58fa429c05b3bce26adc2a519a3a5469d475c18f8f8"
DISCOVERY_CANONICAL_BYTES = 2243707
DISCOVERY_RAW_BYTES = 2937930
DISCOVERY_REVISION = "20260817"

READ_TERMS = ("generate", "suggest", "search", "list", "get", "fetch", "retrieve", "preview")
WRITE_TERMS = (
    "mutate",
    "upload",
    "apply",
    "create",
    "remove",
    "update",
    "start",
    "end",
    "run",
    "cancel",
    "delete",
    "dismiss",
    "configure",
    "resolve",
    "append",
    "addoperations",
    "provide",
    "promote",
    "graduate",
    "schedule",
    "enable",
    "move",
)
RESOURCE_PATH_VARS = ["resourceName", "name", "experiment", "campaignDraft", "adGroupAd"]


def load_discovery():
    with urllib.request.urlopen(DISCOVERY_URL, timeout=60) as response:
        raw = response.read()
    document = json.loads(raw)
    canonical = json.dumps(document, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()
    actual_sha256 = hashlib.sha256(canonical).hexdigest()
    if len(raw) != DISCOVERY_RAW_BYTES or len(canonical) != DISCOVERY_CANONICAL_BYTES or actual_sha256 != DISCOVERY_CANONICAL_SHA256:
        raise RuntimeError(
            "Google Ads Discovery source drift: "
            f"got canonical_sha256={actual_sha256} canonical_bytes={len(canonical)} raw_bytes={len(raw)}, "
            f"want canonical_sha256={DISCOVERY_CANONICAL_SHA256} canonical_bytes={DISCOVERY_CANONICAL_BYTES} raw_bytes={DISCOVERY_RAW_BYTES}"
        )
    if str(document.get("revision")) != DISCOVERY_REVISION:
        raise RuntimeError(
            f"Google Ads Discovery revision drift: got {document.get('revision')}, want {DISCOVERY_REVISION}"
        )
    return document


def collect_methods(rest_description):
    methods = []

    def walk(resource, prefix=""):
        for method_name, method in resource.get("methods", {}).items():
            methods.append((prefix + method_name, method))
        for resource_name, child in sorted(resource.get("resources", {}).items()):
            walk(child, prefix + resource_name + ".")

    walk(rest_description)
    return sorted(methods)


def snake_words(value):
    value = value.replace("-", "_").replace(":", "_").replace("/", "_").replace(".", "_")
    value = re.sub(r"([a-z0-9])([A-Z])", r"\1_\2", value)
    value = re.sub(r"([A-Z]+)([A-Z][a-z])", r"\1_\2", value)
    value = value.lower()
    value = re.sub(r"[^a-z0-9]+", "_", value).strip("_")
    value = re.sub(r"_+", "_", value)
    return value or "operation"


def command_path(method_name):
    return " ".join(snake_words(part).replace("_", "-") for part in method_name.split("."))


def action_name(method_name):
    parts = method_name.split(".")
    last = parts[-1]
    parent = parts[-2] if len(parts) > 1 else parts[0]
    if last.lower() == "mutate":
        return "mutate_" + snake_words(parent)
    return snake_words(last + "_" + parent)


def operation_id(method_name):
    return "google_ads." + snake_words(method_name).replace("_", ".")


def unique(base, seen):
    candidate = base
    index = 2
    while candidate in seen:
        candidate = f"{base}_{index}"
        index += 1
    seen.add(candidate)
    return candidate


def path_vars(path):
    return re.findall(r"\{\+?([^}]+)\}", path or "")


def surface_path(path):
    result = "/" + path.lstrip("/")
    result = result.replace("{+customerId}", "{customer_id}").replace("{customerId}", "{customer_id}")
    for var in RESOURCE_PATH_VARS:
        result = result.replace("{+" + var + "}", "{" + snake_words(var) + "}")
        result = result.replace("{" + var + "}", "{" + snake_words(var) + "}")
    return result


def write_path(path):
    result = "/" + path.lstrip("/")
    result = result.replace("{+customerId}", "{{ config.customer_id }}")
    result = result.replace("{customerId}", "{{ config.customer_id }}")
    return result


def classify(method_name, method):
    if method_name in ("customers.listAccessibleCustomers", "customers.googleAds.search"):
        return "stream"
    if method_name == "customers.googleAds.searchStream":
        return "blocked_duplicate"
    bad_vars = [var for var in path_vars(method.get("path", "")) if var != "customerId"]
    if bad_vars:
        return "blocked_reserved_path"
    last_raw = method_name.split(".")[-1]
    last = last_raw.lower()
    tokens = set(snake_words(last_raw).split("_"))
    http_method = (method.get("httpMethod") or "GET").upper()
    write_verbs = set(WRITE_TERMS)
    if bool(tokens & write_verbs) or last in write_verbs or http_method == "DELETE":
        return "write"
    if any(term in last for term in READ_TERMS):
        return "direct_read"
    return "blocked_unclassified"


def schema_from_ref(schemas, ref, depth=0, seen=None):
    if seen is None:
        seen = set()
    if not ref or ref not in schemas:
        return {"type": "object", "additionalProperties": False, "properties": {}}
    # Recursive discovery types are cut at the cycle boundary, and unusually
    # deep acyclic shapes are capped at the engine's own structured-body depth.
    # Neither case may reopen arbitrary JSON properties in a generated command.
    if ref in seen or depth >= 16:
        return {"type": "object", "additionalProperties": False, "properties": {}}
    seen = set(seen)
    seen.add(ref)
    return schema_from_discovery(schemas, schemas[ref], depth=depth, seen=seen)


def schema_from_discovery(schemas, node, depth=0, seen=None):
    if seen is None:
        seen = set()
    if "$ref" in node:
        return schema_from_ref(schemas, node["$ref"], depth + 1, seen)
    typ = node.get("type")
    if typ == "array":
        return {
            "type": "array",
            "maxItems": 1024,
            "items": schema_from_discovery(schemas, node.get("items", {}), depth + 1, seen),
        }
    if typ == "object" or "properties" in node:
        props = {}
        for key, value in (node.get("properties") or {}).items():
            props[key] = schema_from_discovery(schemas, value, depth + 1, seen)
        out = {"type": "object", "additionalProperties": False, "properties": props}
        required = node.get("required")
        if isinstance(required, list) and required:
            out["required"] = required
        return out
    if typ in ("string", "integer", "number", "boolean"):
        out = {"type": typ}
        if isinstance(node.get("enum"), list) and node["enum"]:
            out["enum"] = node["enum"]
        return out
    return {"type": ["string", "number", "integer", "boolean", "object", "array", "null"]}


def discovery_required_field(node):
    if node.get("required") is True:
        return True
    description = str(node.get("description") or "")
    return bool(re.search(r"(^|[.\n]\s*)Required\.", description))


def discovery_required_body_fields(schemas, ref):
    if not ref:
        return []
    node = schemas.get(ref, {})
    required = list(node.get("required") or [])
    for field, value in (node.get("properties") or {}).items():
        if discovery_required_field(value) and field not in required:
            required.append(field)
    return required


def request_schema(schemas, ref, force_required=None, min_props=False):
    if ref:
        result = schema_from_ref(schemas, ref, depth=0)
    else:
        result = {"type": "object", "additionalProperties": False, "properties": {}}
    result.setdefault("type", "object")
    result["additionalProperties"] = False
    props = result.setdefault("properties", {})
    required = list(result.get("required") or [])
    if force_required:
        for field in force_required:
            if field in props and field not in required:
                required.append(field)
    if required:
        result["required"] = required
    elif min_props:
        result["minProperties"] = 1
    return result


def schema_has_open_objects(schema):
    if isinstance(schema, dict):
        if schema.get("additionalProperties") is True:
            return True
        return any(schema_has_open_objects(value) for value in schema.values())
    if isinstance(schema, list):
        return any(schema_has_open_objects(value) for value in schema)
    return False


def raw_write_block_row(method_name, method, http_method, spath, source, name):
    return {
        "method": http_method,
        "path": spath,
        "operation": {
            "model": "disallowed",
            "status": "blocked",
            "risk": write_risk(method_name, method),
            "blocked_by_default": True,
            "reason": "Blocked because the discovery request schema contains open-ended operation objects; exposing it would be a raw Google Ads write escape hatch rather than a typed connector-owned action.",
            "source_url": source,
            "notes": "Author a closed record_schema for " + name + " before enabling this reverse ETL action.",
        },
    }


def raw_query_block_row(method_name, http_method, spath, source):
    return {
        "method": http_method,
        "path": spath,
        "operation": {
            "model": "disallowed",
            "status": "blocked",
            "risk": "medium",
            "blocked_by_default": True,
            "reason": "Blocked because body.query accepts arbitrary GAQL text; Google Ads query surfaces must use fixed connector-owned stream queries.",
            "source_url": source,
            "notes": "Use the fixed campaigns and ad_groups streams instead of exposing googleAdsFields.search as a raw query command.",
        },
    }


def required_body_direct_read_block_row(method_name, http_method, spath, source, unsupported_required):
    fields = ", ".join(unsupported_required)
    return {
        "method": http_method,
        "path": spath,
        "operation": {
            "model": "direct_read",
            "status": "blocked",
            "risk": "medium",
            "blocked_by_default": True,
            "reason": "Blocked because required Google Ads request body fields cannot be supplied by the typed CLI surface without exposing a raw JSON body.",
            "source_url": source,
            "notes": "Unsupported required body fields: " + fields + ". Add connector-owned typed mappings before enabling this direct read.",
        },
    }


def required_query_direct_read_block_row(method_name, http_method, spath, source, unsupported_required):
    fields = ", ".join(unsupported_required)
    return {
        "method": http_method,
        "path": spath,
        "operation": {
            "model": "direct_read",
            "status": "blocked",
            "risk": "medium",
            "blocked_by_default": True,
            "reason": "Blocked because required Google Ads query parameters cannot be supplied by the typed CLI surface.",
            "source_url": source,
            "notes": "Unsupported required query parameters: " + fields + ". Add connector-owned typed mappings before enabling this direct read.",
        },
    }


def dummy_for_schema(schema):
    typ = schema.get("type")
    if isinstance(typ, list):
        for candidate in typ:
            if candidate != "null":
                return dummy_for_schema({**schema, "type": candidate})
        return None
    if schema.get("enum"):
        return schema["enum"][0]
    if typ == "string":
        return "fixture_value"
    if typ == "integer":
        return 1
    if typ == "number":
        return 1.0
    if typ == "boolean":
        return True
    if typ == "array":
        return [dummy_for_schema(schema.get("items", {}))]
    if typ == "object":
        return {key: dummy_for_schema((schema.get("properties") or {}).get(key, {"type": "string"})) for key in schema.get("required") or []}
    return "fixture_value"


def fixture_record(schema):
    props = schema.get("properties") or {}
    keys = list(schema.get("required") or [])
    if not keys and props:
        keys = [next(iter(props))]
    record = {}
    for key in keys:
        if key == "operations":
            record[key] = [{"create": {}}]
        elif key == "resourceName":
            record[key] = "customers/synthetic-conformance-value/resources/fixture"
        elif key in props:
            record[key] = dummy_for_schema(props[key])
    return record


def cli_flag_for_schema(name, schema, target_prefix="body", summary_prefix="Google Ads request field"):
    typ = schema.get("type")
    if isinstance(typ, list):
        non_null = [item for item in typ if item != "null"]
        if len(non_null) == 1:
            return cli_flag_for_schema(name, {**schema, "type": non_null[0]}, target_prefix, summary_prefix)
        return None
    flag = {"name": snake_words(name).replace("_", "-"), "maps_to": target_prefix + "." + name, "summary": summary_prefix + " " + name + "."}
    if typ == "boolean":
        flag["type"] = "boolean"
        return flag
    if typ == "integer":
        flag["type"] = "integer"
        return flag
    if typ == "string":
        if schema.get("enum"):
            flag["type"] = "enum"
            flag["values"] = schema["enum"]
        else:
            flag["type"] = "string"
        return flag
    if typ == "array":
        item = schema.get("items") or {}
        item_type = item.get("type")
        if item_type == "string" or item.get("enum"):
            flag["type"] = "string_array"
            if item.get("enum"):
                flag["values"] = item["enum"]
            return flag
        flag["type"] = "json"
        return flag
    if typ == "object":
        flag["type"] = "json"
        return flag
    return None


def cli_flags_for_body_schema(body_schema):
    flags = []
    required = set(body_schema.get("required") or [])
    for name, schema in sorted((body_schema.get("properties") or {}).items()):
        flag = cli_flag_for_schema(name, schema)
        if flag:
            if name in required:
                flag["required"] = True
            flags.append(flag)
    return flags


def discovery_required_parameter(parameter):
    if parameter.get("required") is True:
        return True
    description = str(parameter.get("description") or "").strip().lower()
    return description.startswith("required.")


def cli_flag_for_query_parameter(name, parameter):
    if parameter.get("location") != "query" or not discovery_required_parameter(parameter):
        return None
    schema = {"type": parameter.get("type")}
    if parameter.get("enum"):
        schema["enum"] = parameter["enum"]
    flag = cli_flag_for_schema(name, schema, target_prefix="query", summary_prefix="Google Ads query parameter")
    if not flag:
        return None
    description = " ".join(str(parameter.get("description") or "").split())
    if description:
        flag["summary"] = description
    flag["required"] = True
    return flag


def cli_flags_for_query_parameters(method):
    flags = []
    for name, parameter in sorted((method.get("parameters") or {}).items()):
        flag = cli_flag_for_query_parameter(name, parameter)
        if flag:
            flags.append(flag)
    return flags


def unsupported_required_body_fields(body_schema, flags):
    mapped = {flag.get("maps_to") for flag in flags}
    unsupported = []
    for field in body_schema.get("required") or []:
        if "body." + field not in mapped:
            unsupported.append(field)
    return unsupported


def unsupported_required_query_parameters(method, flags):
    mapped = {flag.get("maps_to") for flag in flags}
    unsupported = []
    for name, parameter in sorted((method.get("parameters") or {}).items()):
        if parameter.get("location") == "query" and discovery_required_parameter(parameter) and "query." + name not in mapped:
            unsupported.append(name)
    return unsupported


def example_value_for_flag(flag):
    target = flag.get("maps_to")
    if target == "query.billingSetup":
        return "customers/synthetic-conformance-value/billingSetups/fixture"
    if target == "query.issueMonth":
        return "JANUARY"
    if target == "query.issueYear":
        return "2026"
    values = flag.get("values") or []
    if values:
        return str(values[0])
    typ = flag.get("type")
    if typ == "boolean":
        return "true"
    if typ == "integer":
        return "1"
    return "fixture_value"


def command_example(cmd, flags, required_fields):
    parts = ["pm", "google-ads", *cmd.split()]
    required = set(required_fields or [])
    for flag in flags:
        target = str(flag.get("maps_to", ""))
        field = target.removeprefix("body.").removeprefix("query.")
        if target in required or field in required:
            parts.extend(["--" + flag["name"], example_value_for_flag(flag)])
    parts.append("--json")
    return " ".join(parts)


def write_kind(method_name, method):
    lower = method_name.lower()
    http_method = (method.get("httpMethod") or "GET").upper()
    if http_method == "DELETE":
        return "delete"
    if any(term in lower for term in ("remove", "delete", "cancel", "endexperiment", "dismiss")):
        return "custom"
    if any(term in lower for term in ("create", "upload", "addoperations", "append")):
        return "create"
    if any(term in lower for term in ("update", "mutate", "apply", "configure", "enable", "move")):
        return "update"
    return "custom"


def write_risk(method_name, method):
    lower = method_name.lower()
    http_method = (method.get("httpMethod") or "GET").upper()
    if http_method == "DELETE" or any(term in lower for term in ("remove", "delete", "cancel", "endexperiment", "dismiss")):
        return "critical"
    if any(term in lower for term in ("mutate", "upload", "apply", "create", "update", "run", "start", "move", "promote", "graduate", "schedule", "configure", "append", "addoperations", "provide", "enable")):
        return "high"
    return "medium"


def blocked_model(method_name, method, bad_vars):
    http_method = (method.get("httpMethod") or "GET").upper()
    last = method_name.split(".")[-1].lower()
    if any(term in last for term in READ_TERMS) and not any(term in last for term in WRITE_TERMS):
        return "direct_read", "medium"
    if http_method == "DELETE" or any(term in method_name.lower() for term in ("remove", "delete", "cancel", "endexperiment", "dismiss")):
        return "destructive_action", "critical"
    return "admin_reverse_etl", "high"


def build_artifacts(rest_description):
    schemas = rest_description["schemas"]
    methods = collect_methods(rest_description)
    class_counts = Counter()
    api_endpoints = [
        {"method": "GET", "path": "/v22/customers:listAccessibleCustomers", "covered_by": {"stream": "accessible_customers"}},
        {"method": "POST", "path": "/v22/customers/{customer_id}/googleAds:search (fixed GAQL: campaign)", "covered_by": {"stream": "campaigns"}},
        {"method": "POST", "path": "/v22/customers/{customer_id}/googleAds:search (fixed GAQL: ad_group)", "covered_by": {"stream": "ad_groups"}},
    ]
    writes = []
    write_fixtures = {}
    operations = []
    cli_commands = []
    blocked_rows = []
    seen_write_names, seen_commands, seen_ops = set(), set(), set()

    for method_name, method in methods:
        cls = classify(method_name, method)
        class_counts[cls] += 1
        if method_name in ("customers.listAccessibleCustomers", "customers.googleAds.search"):
            continue
        http_method = (method.get("httpMethod") or "GET").upper()
        spath = surface_path(method.get("path", ""))
        req_ref = (method.get("request") or {}).get("$ref")
        source = DISCOVERY_URL + "#" + method_name

        if cls == "write":
            name = unique(action_name(method_name), seen_write_names)
            discovery_schema = schemas.get(req_ref, {}) if req_ref else {}
            force_required = ["operations"] if "operations" in (discovery_schema.get("properties") or {}) else None
            body_schema = request_schema(schemas, req_ref, force_required=force_required, min_props=bool(req_ref))
            if "resourceName" in (body_schema.get("properties") or {}):
                required = list(body_schema.get("required") or [])
                if "resourceName" not in required:
                    required.append("resourceName")
                body_schema["required"] = required
                body_schema.pop("minProperties", None)
            if schema_has_open_objects(body_schema):
                class_counts[cls] -= 1
                class_counts["blocked_raw_write_schema"] += 1
                row = raw_write_block_row(method_name, method, http_method, spath, source, name)
                api_endpoints.append(row)
                blocked_rows.append(row)
                continue
            risk = write_risk(method_name, method)
            action = {
                "name": name,
                "kind": write_kind(method_name, method),
                "method": http_method,
                "path": write_path(method.get("path", "")),
                "record_schema": body_schema,
                "risk": f"Executes Google Ads API v22 method {method_name} against the configured customer. Review provider-side effects before approval.",
            }
            sensitive_terms = ("upload", "offline_user_data", "customer_user_access", "user_data", "user_list")
            if any(term in name for term in sensitive_terms):
                redact_fields = [field for field in ("operations", "operation", "conversions", "conversionAdjustments", "customerMatchUserListMetadata", "job") if field in (body_schema.get("properties") or {})]
                if redact_fields:
                    action["redact_fields"] = redact_fields
            if risk in ("high", "critical"):
                action["confirm"] = "destructive"
            if http_method == "DELETE":
                action["body_type"] = "none"
                action["delete"] = {"idempotent": True, "missing_ok_status": [404]}
            writes.append(action)
            api_endpoints.append({"method": http_method, "path": spath, "covered_by": {"write": name}})
            record = fixture_record(body_schema)
            expected_path = spath.replace("{customer_id}", "synthetic-conformance-value")
            fixture = {"record": record, "expect": {"method": http_method, "path": expected_path}, "response": {"status": 200, "body": {}}}
            if record and http_method != "DELETE":
                fixture["expect"]["body"] = record
            write_fixtures[name] = fixture
            continue

        if cls == "direct_read":
            required_fields = discovery_required_body_fields(schemas, req_ref)
            body_schema = request_schema(schemas, req_ref, force_required=required_fields, min_props=False)
            if method_name == "googleAdsFields.search" and "query" in (body_schema.get("properties") or {}):
                class_counts[cls] -= 1
                class_counts["blocked_raw_query"] += 1
                row = raw_query_block_row(method_name, http_method, spath, source)
                api_endpoints.append(row)
                blocked_rows.append(row)
                continue
            flags = cli_flags_for_body_schema(body_schema) if http_method == "POST" else []
            flags.extend(cli_flags_for_query_parameters(method))
            unsupported_required = unsupported_required_body_fields(body_schema, flags)
            if unsupported_required:
                class_counts[cls] -= 1
                class_counts["blocked_required_body_direct_read"] += 1
                row = required_body_direct_read_block_row(method_name, http_method, spath, source, unsupported_required)
                api_endpoints.append(row)
                blocked_rows.append(row)
                continue
            unsupported_query = unsupported_required_query_parameters(method, flags)
            if unsupported_query:
                class_counts[cls] -= 1
                class_counts["blocked_required_query_direct_read"] += 1
                row = required_query_direct_read_block_row(method_name, http_method, spath, source, unsupported_query)
                api_endpoints.append(row)
                blocked_rows.append(row)
                continue
            op = unique(operation_id(method_name), seen_ops)
            cmd = unique(command_path(method_name), seen_commands)
            operation = {
                "id": op,
                "kind": "rest_read",
                "summary": f"Read Google Ads API v22 method {method_name} with a fixed connector-owned operation definition.",
                "source_url": source,
                "risk": "medium",
                "approval": "none; bounded direct read with JSON redaction and a fixed endpoint",
                "output_policy": "json_redacted",
                "rest": {"method": http_method, "path": spath, "max_bytes": 1048576},
            }
            if http_method == "POST":
                operation["rest"]["content_type"] = "application/json"
                operation["rest"]["body_schema"] = body_schema
                operation["rest"]["body"] = {}
            operations.append(operation)
            command = {
                "path": cmd,
                "summary": f"Read Google Ads {method_name}.",
                "intent": "direct_read",
                "availability": "implemented",
                "operation": op,
                "output_policy": "json_redacted",
                "source_url": source,
                "risk": "Bounded JSON direct read; response fields with secret-like names are redacted.",
                "approval": "none",
                "api_surface": [{"method": http_method, "path": spath}],
                "examples": [command_example(cmd, flags, ["body." + field for field in (body_schema.get("required") or [])] + [flag["maps_to"] for flag in flags if flag.get("required")])],
            }
            if flags:
                command["flags"] = flags
            if method_name == "customers.generateKeywordIdeas":
                seed_targets = [
                    "body.keywordAndUrlSeed",
                    "body.urlSeed",
                    "body.keywordSeed",
                    "body.siteSeed",
                ]
                if all(any(flag.get("maps_to") == target for flag in flags) for target in seed_targets):
                    command["constraints"] = [{
                        "kind": "exactly_one",
                        "fields": seed_targets,
                        "message": "exactly one Google Ads keyword seed must be provided",
                    }]
            cli_commands.append(command)
            api_endpoints.append({"method": http_method, "path": spath, "covered_by": {"direct_read": cmd}})
            continue

        bad_vars = [var for var in path_vars(method.get("path", "")) if var != "customerId"]
        if cls == "blocked_duplicate":
            row = {
                "method": http_method,
                "path": spath,
                "operation": {
                    "model": "duplicate",
                    "status": "blocked",
                    "risk": "low",
                    "blocked_by_default": True,
                    "reason": "SearchStream is the streaming transport for the same GAQL search surface; the connector implements fixed, bounded campaign/ad_group reads through googleAds:search instead.",
                    "source_url": source,
                    "duplicate_of": "customers.googleAds.search fixed campaign/ad_group streams",
                },
            }
        else:
            model, risk = blocked_model(method_name, method, bad_vars)
            row = {
                "method": http_method,
                "path": spath,
                "operation": {
                    "model": model,
                    "status": "blocked",
                    "risk": risk,
                    "blocked_by_default": True,
                    "reason": "Blocked by connector-local path contract: Google Ads uses reserved-expansion resource-name path variables (" + ",".join(bad_vars or ["unknown"]) + ") that may contain slashes, while current direct/write path interpolation safely URL-encodes slash characters. Executing this without shared reserved-expansion support would call a different path.",
                    "source_url": source,
                    "notes": "Needs shared reserved-expansion path-variable support or a connector-local declarative raw-slash path field type.",
                },
            }
        api_endpoints.append(row)
        blocked_rows.append(row)

    return methods, class_counts, api_endpoints, writes, write_fixtures, operations, cli_commands, blocked_rows


def write_json(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        json.dump(data, handle, indent=2, sort_keys=False)
        handle.write("\n")


def update_definition_files(rest_description, artifacts):
    methods, class_counts, api_endpoints, writes, write_fixtures, operations, cli_commands, blocked_rows = artifacts

    metadata = json.load((ROOT / "metadata.json").open())
    metadata["description"] = "Declarative Google Ads connector for v22 customer, campaign, ad group, direct-read, and limited guarded reverse-ETL API surfaces."
    metadata.pop("categories", None)
    metadata.pop("status", None)
    caps = metadata.setdefault("capabilities", {})
    caps["read"] = True
    caps["write"] = True
    caps.pop("bulk", None)
    caps.pop("webhook", None)
    metadata["risk"] = {
        "read": "external Google Ads API reads of customer, campaign, ad-group, and bounded direct-read metadata",
        "write": "limited guarded Google Ads API reverse/write actions with closed record schemas; destructive/admin actions require explicit approval",
        "approval": "reads require no approval; writes remain gated by plan -> preview -> explicit approval -> execute"
    }
    write_json(ROOT / "metadata.json", metadata)

    spec_doc = {
        "$schema": "http://json-schema.org/draft-07/schema#",
        "title": "Google Ads Connection Specification",
        "type": "object",
        "required": ["access_token", "developer_token"],
        "properties": {
            "access_token": {
                "type": "string",
                "x-secret": True,
                "description": "Google OAuth 2.0 access token with Google Ads API scopes. Used only for Bearer auth; never logged. Acquisition/refresh is out of scope for this connector (credentials layer already owns it)."
            },
            "developer_token": {
                "type": "string",
                "x-secret": True,
                "description": "Google Ads developer token, sent as the developer-token header on every request. Never logged."
            },
            "login_customer_id": {
                "type": "string",
                "description": "Optional manager (MCC) customer id sent as the login-customer-id header. Omitted entirely when unset."
            },
            "base_url": {
                "type": "string",
                "format": "uri",
                "default": "https://googleads.googleapis.com",
                "description": "Google Ads REST API root URL override for tests or proxies; declarative paths include /v22."
            },
            "customer_id": {
                "type": "string",
                "description": "Google Ads customer id without dashes. Required for customer-scoped streams, direct reads, and reverse/write actions."
            },
            "page_size": {
                "type": "integer",
                "default": 1000,
                "description": "GAQL search pageSize (1-10000)."
            },
            "max_pages": {
                "type": "string",
                "description": "Maximum pages fetched per GAQL search stream. A positive integer, or 'all'/'unlimited' (default) for no cap."
            },
            "mode": {
                "type": "string",
                "description": "Runtime mode: live (default) or fixture for credential-free conformance."
            }
        }
    }
    write_json(ROOT / "spec.json", spec_doc)

    streams = json.load((ROOT / "streams.json").open())
    streams.setdefault("base", {}).setdefault("check", {})["method"] = "GET"
    streams.setdefault("base", {}).setdefault("check", {})["path"] = "v22/customers:listAccessibleCustomers"
    for stream in streams["streams"]:
        if stream["name"] == "accessible_customers":
            stream["path"] = "v22/customers:listAccessibleCustomers"
        elif stream["name"] in ("campaigns", "ad_groups"):
            stream["path"] = "v22/customers/{{ config.customer_id }}/googleAds:search"
    write_json(ROOT / "streams.json", streams)

    api_surface = {
        "api": "Google Ads API v22 REST discovery revision " + str(rest_description.get("revision")),
        "docs": "https://developers.google.com/google-ads/api/reference/rpc/v22/overview",
        "reviewed_at": "2026-07-31",
        "operation_ledger_version": 1,
        "scope": "Fixture-only parity ledger generated from the public v22 discovery document; no live Google Ads credentials or API calls were used. customers.googleAds.search is split into two fixed GAQL stream rows (campaigns and ad_groups) because connector streams are separate covered_by units. Reserved-expansion resource-name path variables remain blocked because current connector-local templating safely URL-encodes slash characters.",
        "endpoints": api_endpoints,
    }
    write_json(ROOT / "api_surface.json", api_surface)

    write_json(ROOT / "writes.json", {"actions": writes})
    write_json(ROOT / "operations.json", {"operations": operations})
    write_json(ROOT / "cli_surface.json", {
        "tagline": "Google Ads v22 fixed direct reads and limited guarded reverse/write actions.",
        "usage": "pm google-ads <resource> <operation> [flags]",
        "source_cli": {
            "name": "Google Ads API v22 REST discovery",
            "docs": "https://developers.google.com/google-ads/api/rest/overview",
            "reference": "https://developers.google.com/google-ads/api/reference/rpc/v22/overview",
            "source": DISCOVERY_URL
        },
        "commands": cli_commands
    })

    fixtures_dir = ROOT / "fixtures" / "writes"
    if fixtures_dir.exists():
        shutil.rmtree(fixtures_dir)
    fixtures_dir.mkdir(parents=True, exist_ok=True)
    for name, fixture in sorted(write_fixtures.items()):
        write_json(fixtures_dir / (name + ".json"), fixture)

    # Update request fixture paths for v22 root paths.
    fixture_paths = [ROOT / "fixtures" / "check.json"]
    fixture_paths.extend(sorted((ROOT / "fixtures" / "streams").glob("*/page_*.json")))
    for fixture_path in fixture_paths:
        fixture = json.load(fixture_path.open())
        req = fixture.get("request", {})
        path = req.get("path", "")
        if path.startswith("/customers"):
            req["path"] = "/v22" + path
        write_json(fixture_path, fixture)

    docs = f"""# Google Ads connector notes\n\n## Overview\n\nGoogle Ads is implemented as a declarative preview connector against the public Google Ads API v22 REST discovery document. This wave ships sanitized fixture coverage plus executable credential-backed reads, fixed direct reads, and guarded reverse/write actions, but does not claim certification.\n\nPublic source audit:\n\n- Source: `{DISCOVERY_URL}`\n- API version: `{rest_description.get('version')}`\n- Discovery revision: `{rest_description.get('revision')}`\n- Raw discovery method count: `{len(methods)}` (`POST={Counter(m.get('httpMethod') for _, m in methods).get('POST', 0)}`, `GET={Counter(m.get('httpMethod') for _, m in methods).get('GET', 0)}`, `DELETE={Counter(m.get('httpMethod') for _, m in methods).get('DELETE', 0)}`)\n- Local operation ledger rows: `{len(api_endpoints)}`. The row count is one greater than the raw method count because the single `customers.googleAds.search` method is intentionally represented by two fixed GAQL stream rows: `campaigns` and `ad_groups`.\n\n## Auth setup\n\nProvide `access_token` and `developer_token` through the credentials layer or environment. Optional `login_customer_id` is sent only when present. `customer_id` is required for customer-scoped streams, fixed direct reads, and reverse/write actions. Do not place secret values in plans, docs, fixtures, or command text.\n\n## Streams notes\n\nImplemented streams are `accessible_customers`, `campaigns`, and `ad_groups`. The campaign and ad group streams use fixed connector-owned GAQL statements; the connector does not expose arbitrary GAQL or raw search passthrough.\n\nDirect reads: `{len(operations)}` fixed connector-owned operations with JSON-redacted output, bounded response size, and typed CLI body/query fields where a POST body or GET query parameters are required.\n\n## Write actions & risks\n\nReverse/write actions: `{len(writes)}` guarded write actions whose request schemas are closed and connector-owned.\n\n- Write actions use closed record schemas derived from public discovery fields that can be represented without raw operation objects.\n- Destructive or account-admin actions carry explicit `confirm: destructive` metadata and remain subject to the platform reverse ETL plan -> preview -> approval -> execute lifecycle.\n- Secret-like fields are redacted; `access_token` and `developer_token` are never stored in fixtures.\n- No generic Google Ads SQL/GAQL shell, generic HTTP write, or raw request passthrough is exposed.\n\n## Known limits\n\nBlocked/planned operations: `{len(blocked_rows)}` rows. These are not advertised as executable. Reserved-expansion resource-name path variables, open-ended discovery write schemas, raw GAQL query commands, and direct reads with required complex request bodies remain blocked.\n\nGoogle Ads methods whose REST paths use `{{+resourceName}}`, `{{+name}}`, `{{+experiment}}`, `{{+campaignDraft}}`, or `{{+adGroupAd}}` are blocked in `api_surface.json`. These path variables are reserved expansions and may contain slash-separated Google Ads resource names. The current connector-local path interpolation intentionally URL-encodes slashes for safety, so enabling those methods without shared reserved-expansion support would call the wrong URL.\n"""
    docs = docs.replace(
        "- Raw discovery method count:",
        f"- Canonical Discovery SHA-256: `{DISCOVERY_CANONICAL_SHA256}` (`{DISCOVERY_CANONICAL_BYTES}` canonical bytes; `{DISCOVERY_RAW_BYTES}` source bytes)\n- Raw discovery method count:",
    )
    (ROOT / "docs.md").write_text(docs, encoding="utf-8")


def write_audit(rest_description, artifacts):
    methods, class_counts, api_endpoints, writes, write_fixtures, operations, cli_commands, blocked_rows = artifacts
    http_counts = Counter(m.get("httpMethod") for _, m in methods)
    var_counts = Counter()
    for _, method in methods:
        for var in path_vars(method.get("path", "")):
            var_counts[var] += 1
    audit = {
        "source_url": DISCOVERY_URL,
        "version": rest_description.get("version"),
        "revision": rest_description.get("revision"),
        "source_canonical_sha256": DISCOVERY_CANONICAL_SHA256,
        "source_canonical_bytes": DISCOVERY_CANONICAL_BYTES,
        "source_raw_bytes": DISCOVERY_RAW_BYTES,
        "method_count": len(methods),
        "http_method_counts": dict(sorted(http_counts.items())),
        "schema_count": len(rest_description.get("schemas", {})),
        "path_variable_counts": dict(sorted(var_counts.items())),
        "classification_counts": dict(sorted(class_counts.items())),
        "generated_counts": {
            "api_surface_rows": len(api_endpoints),
            "streams": 3,
            "direct_reads": len(operations),
            "write_actions": len(writes),
            "write_fixtures": len(write_fixtures),
            "blocked_rows": len(blocked_rows),
        },
        "blocked_methods": [row for row in blocked_rows],
    }
    write_json(PHASE / "SOURCE-AUDIT.json", audit)
    lines = [
        "# Google Ads v22 source audit",
        "",
        f"- Source: `{DISCOVERY_URL}`",
        f"- Version: `{rest_description.get('version')}`",
        f"- Revision: `{rest_description.get('revision')}`",
        f"- Canonical source SHA-256: `{DISCOVERY_CANONICAL_SHA256}`",
        f"- Canonical source bytes: `{DISCOVERY_CANONICAL_BYTES}`",
        f"- Raw source bytes: `{DISCOVERY_RAW_BYTES}`",
        f"- Raw methods: `{len(methods)}` ({dict(sorted(http_counts.items()))})",
        f"- Schemas: `{len(rest_description.get('schemas', {}))}`",
        f"- Path variable counts: `{dict(sorted(var_counts.items()))}`",
        f"- Classification counts: `{dict(sorted(class_counts.items()))}`",
        f"- Generated counts: `{audit['generated_counts']}`",
        "",
        "The `api_surface.json` ledger has one extra row relative to raw discovery because `customers.googleAds.search` backs two fixed connector streams (`campaigns`, `ad_groups`).",
    ]
    (PHASE / "SOURCE-AUDIT.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


def main():
    rest_description = load_discovery()
    artifacts = build_artifacts(rest_description)
    update_definition_files(rest_description, artifacts)
    write_audit(rest_description, artifacts)
    counts = artifacts[1]
    print("generated", {
        "classification": dict(sorted(counts.items())),
        "api_surface_rows": len(artifacts[2]),
        "writes": len(artifacts[3]),
        "write_fixtures": len(artifacts[4]),
        "direct_reads": len(artifacts[5]),
        "blocked": len(artifacts[7]),
    })


if __name__ == "__main__":
    main()
