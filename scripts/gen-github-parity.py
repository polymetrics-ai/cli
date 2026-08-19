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
import hashlib, json, os, re, sys

ROOT = "internal/connectors/defs/github"
api = json.load(open(f"{ROOT}/api_surface.json"))
ops_doc = json.load(open(f"{ROOT}/operations.json"))
writes_doc = json.load(open(f"{ROOT}/writes.json"))
cli_doc = json.load(open(f"{ROOT}/cli_surface.json"))

existing_op_ids = {o["id"] for o in ops_doc["operations"]}
existing_write_names = {a["name"] for a in writes_doc["actions"]}
existing_cli_paths = {c["path"] for c in cli_doc["commands"]}

PATH_PARAM_RE = re.compile(r"\{([^}]+)\}")
ENGINE_PATH_PARAM_RE = re.compile(r"(?<!\{)\{([^{}]+)\}(?!\})")

# The completion pass derives contracts only from this exact public artifact,
# already pinned by github-operation-source-lock.json. The source is supplied
# explicitly during regeneration rather than fetched implicitly: generated
# connector artifacts remain reviewable and no developer command gains a
# surprise network dependency.
PINNED_OPENAPI_SHA256 = "80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d"
openapi = None

def load_pinned_openapi():
    global openapi
    if openapi is not None:
        return openapi
    source_path = os.environ.get("GITHUB_OPENAPI_PATH", "").strip()
    if not source_path:
        raise ValueError("GITHUB_OPENAPI_PATH is required to regenerate a newly unblocked GitHub REST contract")
    with open(source_path, "rb") as source:
        raw = source.read()
    actual = hashlib.sha256(raw).hexdigest()
    if actual != PINNED_OPENAPI_SHA256:
        raise ValueError(f"GITHUB_OPENAPI_PATH sha256 {actual} does not match pinned source {PINNED_OPENAPI_SHA256}")
    openapi = json.loads(raw)
    return openapi

def resolve_openapi_ref(value):
    """Resolve a local OpenAPI reference without admitting another artifact."""
    document = load_pinned_openapi()
    seen = set()
    while isinstance(value, dict) and "$ref" in value:
        ref = value["$ref"]
        if not isinstance(ref, str) or not ref.startswith("#/") or ref in seen:
            raise ValueError(f"unsupported OpenAPI reference {ref!r}")
        seen.add(ref)
        target = document
        for segment in ref[2:].split("/"):
            target = target[segment.replace("~1", "/").replace("~0", "~")]
        value = target
    return value

def merge_all_of(parts):
    """Merge the narrow object shape used by GitHub's reusable body schemas."""
    merged = {"type": "object", "properties": {}, "required": []}
    for part in parts:
        normalized = openapi_schema_to_engine(part)
        if normalized is None:
            return None
        if normalized.get("type") not in (None, "object"):
            return None
        merged["properties"].update(normalized.get("properties", {}))
        merged["required"].extend(normalized.get("required", []))
        if normalized.get("additionalProperties") is False:
            merged["additionalProperties"] = False
    merged["required"] = sorted(set(merged["required"]))
    if not merged["required"]:
        merged.pop("required")
    if not merged["properties"]:
        merged.pop("properties")
    return merged

def openapi_schema_to_engine(schema):
    """Reduce a resolved OpenAPI schema to the engine's closed draft-07 subset.

    The output intentionally carries only validation facts the runtime knows.
    A nested provider union stays a typed container supplied through a single
    declaration-bound JSON flag; a ROOT union returns None so callers must
    define separately named actions rather than flatten it into a generic body.
    """
    schema = resolve_openapi_ref(schema or {})
    if "allOf" in schema:
        return merge_all_of([resolve_openapi_ref(item) for item in schema["allOf"]])
    if "oneOf" in schema or "anyOf" in schema:
        return None
    out = {}
    value_type = schema.get("type")
    if isinstance(value_type, str):
        out["type"] = value_type
    elif "properties" in schema:
        out["type"] = "object"
    elif "items" in schema:
        out["type"] = "array"
    if "enum" in schema:
        out["enum"] = schema["enum"]
    if "pattern" in schema:
        out["pattern"] = schema["pattern"]
    if "format" in schema:
        out["format"] = schema["format"]
    for key in ("minProperties", "minItems", "maxItems"):
        if key in schema:
            out[key] = schema[key]
    if isinstance(schema.get("additionalProperties"), bool):
        out["additionalProperties"] = schema["additionalProperties"]
    if "properties" in schema:
        properties = {}
        for name, child in schema["properties"].items():
            normalized = openapi_schema_to_engine(child)
            # A nested union remains a bounded container, not an open body.
            # The root union check above is the only place command promotion
            # would otherwise blur distinct documented operations.
            if normalized is None:
                child = resolve_openapi_ref(child)
                normalized = {"type": child.get("type", "object")}
            properties[name] = normalized
        out["properties"] = properties
    if "required" in schema:
        out["required"] = list(schema["required"])
    if "items" in schema:
        items = openapi_schema_to_engine(schema["items"])
        if items is None:
            items = {"type": "object"}
        out["items"] = items
    return out

def openapi_operation(method, api_path):
    document = load_pinned_openapi()
    path_item = resolve_openapi_ref(document["paths"][api_path])
    return path_item, resolve_openapi_ref(path_item[method.lower()])

def openapi_parameters(method, api_path):
    path_item, operation = openapi_operation(method, api_path)
    values = []
    seen = set()
    for raw in list(path_item.get("parameters", [])) + list(operation.get("parameters", [])):
        parameter = resolve_openapi_ref(raw)
        name = parameter.get("name")
        location = parameter.get("in")
        if not name or location not in ("path", "query") or (location, name) in seen:
            continue
        seen.add((location, name))
        schema = openapi_schema_to_engine(parameter.get("schema", {})) or {"type": "string"}
        value_type = schema.get("type", "string")
        if value_type == "array":
            item_type = (schema.get("items") or {}).get("type")
            value_type = "string_array" if item_type == "string" else "json"
        elif value_type not in ("string", "integer", "number", "boolean"):
            value_type = "string"
        values.append({
            "name": name,
            "in": location,
            "type": value_type,
            "schema": schema,
            "required": bool(parameter.get("required")),
            "values": schema.get("enum", []),
            "summary": parameter.get("description", ""),
        })
    return values

def openapi_body_schema(method, api_path):
    _, operation = openapi_operation(method, api_path)
    request_body = operation.get("requestBody")
    if not request_body:
        return {"type": "object", "additionalProperties": False, "properties": {}}
    request_body = resolve_openapi_ref(request_body)
    content = request_body.get("content", {})
    media = content.get("application/json")
    if media is None:
        raise ValueError(f"{method} {api_path} has no supported application/json request body")
    return openapi_schema_to_engine(media.get("schema", {}))

def source_contract_flags(parameters, schema, include_optional=True, write=False):
    """Build command flags only from declared path/query/body parameters."""
    flags = []
    parameter_names = set()
    for parameter in parameters:
        if parameter["name"] in ("owner", "repo"):
            continue
        parameter_names.add(parameter["name"])
        kind = parameter["type"]
        values = parameter["values"]
        if values and kind == "string":
            kind = "enum"
        target_scope = "record" if write else parameter["in"]
        flag = contract_flag(parameter["name"].replace("_", "-"), kind, f"{target_scope}.{parameter['name']}", required=parameter["required"], values=values)
        if parameter["summary"]:
            flag["summary"] = parameter["summary"]
        flags.append(flag)
    if schema is not None:
        for name, field in sorted((schema.get("properties") or {}).items()):
            # A path/query parameter already owns this record field. Emitting
            # a second body-derived flag would give one command two spellings
            # for the same write value and make their types drift.
            if write and name in parameter_names:
                continue
            required = name in set(schema.get("required") or [])
            if not required and not include_optional:
                continue
            flag = materialized_record_flag(name, field, required=required)
            flag["maps_to"] = "record." + name
            flag["summary"] = ("Required " if required else "Optional ") + name.replace("_", " ") + " record field."
            flags.append(flag)
    return flags

def normalize_reverse_etl_path_flags():
    """A reverse-ETL action owns path values in its record, never path.*.

    Operation reads/writes use `path.<name>` because their executor separates
    request interpolation from body values. A declarative write action uses its
    own `path_fields` to remove `record.<name>` from the JSON body after
    interpolation. Normalize existing generated completion commands as well as
    fresh ones, so rerunning the generator cannot retain an invalid first-pass
    mapping.
    """
    actions = {action["name"]: action for action in writes_doc["actions"]}
    changed = 0
    for command in cli_doc["commands"]:
        if command.get("intent") != "reverse_etl" or not command.get("write"):
            continue
        action = actions.get(command["write"])
        if action is None:
            continue
        path_fields = set(action.get("path_fields") or [])
        for flag in command.get("flags", []):
            target = flag.get("maps_to", "")
            if not target.startswith("path."):
                continue
            name = target[len("path."):]
            if name in path_fields:
                flag["maps_to"] = "record." + name
                changed += 1
        # An earlier completion generation made path fields twice: once from
        # OpenAPI parameters and again from the temporary record schema. Keep
        # the parameter flag (the first source-derived declaration), so a
        # regeneration repairs an interrupted run rather than retaining an
        # invalid duplicate mapping.
        deduped = []
        targets = set()
        for flag in command.get("flags", []):
            target = flag.get("maps_to")
            if target and target in targets:
                changed += 1
                continue
            if target:
                targets.add(target)
            deduped.append(flag)
        command["flags"] = deduped
        # The same interrupted generation used the legacy name heuristic for
        # path-field schemas. When a duplicate revealed the mismatch, align
        # the action with the surviving source-derived command flag. This is
        # intentionally limited to scalar path fields; complex path values
        # are not an executable URL contract.
        props = (action.get("record_schema") or {}).get("properties") or {}
        flags_by_target = {flag.get("maps_to"): flag for flag in deduped}
        for name in path_fields:
            flag = flags_by_target.get("record." + name)
            if not flag or flag.get("type") not in ("string", "integer", "number", "boolean", "enum"):
                continue
            field = props.get(name)
            if field is None:
                continue
            field["type"] = "string" if flag["type"] == "enum" else flag["type"]
            if flag["type"] == "enum":
                field["enum"] = flag.get("values", [])
    return changed

def completion_write_model(method, api_path):
    if method == "DELETE":
        return "destructive_action"
    if "secret" in api_path or api_path.startswith("/applications/"):
        return "sensitive_reverse_etl"
    return "admin_reverse_etl"

SENSITIVE_COMPLETION_FIELDS = {
    ("DELETE", "/applications/{client_id}/grant"): ["access_token"],
    ("POST", "/applications/{client_id}/token"): ["access_token"],
    ("PATCH", "/applications/{client_id}/token"): ["access_token"],
    ("DELETE", "/applications/{client_id}/token"): ["access_token"],
    ("PUT", "/repos/{owner}/{repo}/agents/secrets/{secret_name}"): ["encrypted_value"],
    ("PATCH", "/repos/{owner}/{repo}/hooks/{hook_id}/config"): ["secret"],
}


def engine_write_path(api_path):
    """Translate provider placeholders into the engine's write path dialect."""

    if "{{{{" in api_path or "}}}}" in api_path:
        return api_path.replace("{{{{", "{{").replace("}}}}", "}}")

    def replace(match):
        name = match.group(1)
        namespace = "config" if name in ("owner", "repo") else "record"
        return "{{ %s.%s }}" % (namespace, name)

    return ENGINE_PATH_PARAM_RE.sub(replace, api_path)


operation_paths = {
    operation["id"]: operation.get("rest", {}).get("path", "")
    for operation in ops_doc["operations"]
    if operation.get("rest", {}).get("path")
}
surface_write_paths = {}
for endpoint in api.get("endpoints", []):
    covered = endpoint.get("covered_by") or {}
    targets = []
    if covered.get("write"):
        targets.append(covered["write"])
    targets.extend(covered.get("writes", []))
    for target in targets:
        surface_write_paths[target] = endpoint["path"]
for action in writes_doc["actions"]:
    # Full-surface imports carry the canonical provider path in the matching
    # rest_write operation. Re-derive the engine path from that source so a
    # mixed path (some old rows already used {{ }}, some used { }) cannot be
    # normalized by regex into a malformed nested template.
    provider_path = surface_write_paths.get(
        action.get("name"), operation_paths.get(f"github.{action.get('name')}", action.get("path", ""))
    )
    action["path"] = engine_write_path(provider_path)

# These are the only GitHub rows whose documented response/body shapes fall
# outside the generic JSON GET reader. Keep them explicit: a broad promotion of
# every blocked direct_read/disallowed row would turn a classifier into a
# bypass for future foundation gaps. Each generated declaration below is still
# routed through OperationDirectRead's real preflight and bounded requester.
def contract_flag(name, kind, maps_to, required=False, values=None):
    flag = {"name": name, "type": kind, "maps_to": maps_to}
    if required:
        flag["required"] = True
    if values:
        flag["values"] = values
    return flag


STATUS_READ_MAX_BYTES = 1024
DEFAULT_TEXT_READ_MAX_BYTES = 1024 * 1024
GITHUB_RAW_MARKDOWN_MAX_BYTES = 400 * 1024

EXPLICIT_DIRECT_READ_CONTRACTS = {
    ("GET", "/gists/{gist_id}/star"): {
        "cli_path": "gists star check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [contract_flag("gist-id", "string", "path.gist_id", required=True)],
    },
    ("GET", "/orgs/{org}/blocks/{username}"): {
        "cli_path": "orgs blocks check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [
            contract_flag("org", "string", "path.org", required=True),
            contract_flag("username", "string", "path.username", required=True),
        ],
    },
    ("GET", "/orgs/{org}/members/{username}"): {
        "cli_path": "orgs members check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [
            contract_flag("org", "string", "path.org", required=True),
            contract_flag("username", "string", "path.username", required=True),
        ],
    },
    ("GET", "/orgs/{org}/public_members/{username}"): {
        "cli_path": "orgs public-members check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [
            contract_flag("org", "string", "path.org", required=True),
            contract_flag("username", "string", "path.username", required=True),
        ],
    },
    ("GET", "/teams/{team_id}/members/{username}"): {
        "cli_path": "teams members check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [
            contract_flag("team-id", "integer", "path.team_id", required=True),
            contract_flag("username", "string", "path.username", required=True),
        ],
    },
    ("GET", "/user/blocks/{username}"): {
        "cli_path": "user blocks check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [contract_flag("username", "string", "path.username", required=True)],
    },
    ("GET", "/user/following/{username}"): {
        "cli_path": "user following check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [contract_flag("username", "string", "path.username", required=True)],
    },
    ("GET", "/user/starred/{owner}/{repo}"): {
        "cli_path": "user starred check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        # These are the repository being checked, not this connection's
        # owner/repo. Make them explicit path flags so they override the
        # connector config rather than silently checking the configured repo.
        "flags": [
            contract_flag("owner", "string", "path.owner", required=True),
            contract_flag("repo", "string", "path.repo", required=True),
        ],
    },
    ("GET", "/users/{username}/following/{target_user}"): {
        "cli_path": "users following check",
        "output_policy": "none",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [
            contract_flag("username", "string", "path.username", required=True),
            contract_flag("target-user", "string", "path.target_user", required=True),
        ],
    },
    ("GET", "/zen"): {
        "cli_path": "meta zen view",
        "output_policy": "text",
        "max_bytes": STATUS_READ_MAX_BYTES,
        "flags": [],
    },
    ("GET", "/octocat"): {
        "cli_path": "meta octocat view",
        "output_policy": "text",
        "max_bytes": 16 * 1024,
        "flags": [contract_flag("s", "string", "query.s")],
    },
    ("POST", "/markdown"): {
        "cli_path": "markdown render",
        "output_policy": "text",
        "max_bytes": DEFAULT_TEXT_READ_MAX_BYTES,
        "content_type": "application/json",
        "body_schema": {
            "type": "object",
            "additionalProperties": False,
            "required": ["text"],
            "properties": {
                "text": {"type": "string"},
                "mode": {"type": "string", "enum": ["markdown", "gfm"]},
                "context": {"type": "string"},
            },
        },
        "flags": [
            contract_flag("text", "string", "body.text", required=True),
            contract_flag("mode", "enum", "body.mode", values=["markdown", "gfm"]),
            contract_flag("context", "string", "body.context"),
        ],
    },
    ("POST", "/markdown/raw"): {
        "cli_path": "markdown raw render",
        "output_policy": "text",
        # GitHub documents raw Markdown content as at most 400 KiB. This is
        # both the request and response cap because the declared text executor
        # uses one bounded operation limit for each side of the request.
        "max_bytes": GITHUB_RAW_MARKDOWN_MAX_BYTES,
        "content_type": "text/plain",
        "body_schema": {"type": "string"},
        "flags": [contract_flag("text", "string", "body", required=True)],
    },
}

# These documented endpoints put genuinely alternative request
# contracts at the root of a oneOf. A oneOf is not one executable command
# contract: each arm gets a closed, separately named write action and command.
# Do not broaden this table into a generic oneOf promotion. Each arm below was
# taken from the GitHub OpenAPI artifact recorded in the parity plan and is
# covered by a generator contract test through the real preflight.
def one_of_arm(action, cli_path, required, properties, destructive=False, body_type=None, body_field=None, body_schema=None):
    return {
        "action": action,
        "cli_path": cli_path,
        "required": required,
        "properties": properties,
        "destructive": destructive,
        "body_type": body_type,
        "body_field": body_field,
        "body_schema": body_schema,
    }

CAMPAIGN_COMMON_PROPERTIES = {
    "name": {"type": "string"},
    "description": {"type": "string"},
    "managers": {"type": "array"},
    "team_managers": {"type": "array"},
    "ends_at": {"type": "string"},
    "contact_link": {"type": "string"},
    "generate_issues": {"type": "boolean"},
}

PROJECT_FIELD_PROPERTIES = {
    "name": {"type": "string"},
    "data_type": {"type": "string"},
}

PROJECT_ITEM_PROPERTIES = {
    "type": {"type": "string"},
    "id": {"type": "integer"},
    "owner": {"type": "string"},
    "repo": {"type": "string"},
    "number": {"type": "integer"},
}

EMAIL_ARRAY_SCHEMA = {
    "type": "array",
    "items": {"type": "string"},
    "minItems": 1,
}

CUSTOM_PATTERN_PROPERTIES = {
    "pattern": {"type": "string"},
    "start_delimiter": {"type": "string"},
    "end_delimiter": {"type": "string"},
    "must_match": {"type": "array", "items": {"type": "string"}},
    "must_not_match": {"type": "array", "items": {"type": "string"}},
    # GitHub requires this optimistic-concurrency version. The published
    # nullable marker does not make omitting it safe: the endpoint's required
    # list and documented purpose require the current string version.
    "custom_pattern_version": {"type": "string"},
}

EXPLICIT_ONE_OF_WRITE_CONTRACTS = {
    # GitHub's object and array forms are both independently documented. A
    # scalar email is the exact one-element case of the array form, so the
    # typed array action retains the complete provider semantics without a
    # raw scalar-body escape hatch.
    ("POST", "/user/emails"): {
        "arms": [
            one_of_arm(
                "user_emails_add_object",
                "user emails add",
                ["emails"],
                {"emails": EMAIL_ARRAY_SCHEMA},
            ),
            one_of_arm(
                "user_emails_add_array",
                "user emails add-array",
                ["emails"],
                {"emails": EMAIL_ARRAY_SCHEMA},
                body_type="json_array",
                body_field="emails",
                body_schema=EMAIL_ARRAY_SCHEMA,
            ),
        ],
    },
    ("DELETE", "/user/emails"): {
        "arms": [
            one_of_arm(
                "user_emails_delete_object",
                "user emails delete",
                ["emails"],
                {"emails": EMAIL_ARRAY_SCHEMA},
                destructive=True,
            ),
            one_of_arm(
                "user_emails_delete_array",
                "user emails delete-array",
                ["emails"],
                {"emails": EMAIL_ARRAY_SCHEMA},
                destructive=True,
                body_type="json_array",
                body_field="emails",
                body_schema=EMAIL_ARRAY_SCHEMA,
            ),
        ],
    },
    ("PATCH", "/orgs/{org}/secret-scanning/custom-patterns/{pattern_id}"): {
        "arms": [
            one_of_arm("org_custom_pattern_update_pattern", "org secret-scanning custom-pattern update-pattern", ["custom_pattern_version", "pattern"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("org_custom_pattern_update_start_delimiter", "org secret-scanning custom-pattern update-start-delimiter", ["custom_pattern_version", "start_delimiter"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("org_custom_pattern_update_end_delimiter", "org secret-scanning custom-pattern update-end-delimiter", ["custom_pattern_version", "end_delimiter"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("org_custom_pattern_update_must_match", "org secret-scanning custom-pattern update-must-match", ["custom_pattern_version", "must_match"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("org_custom_pattern_update_must_not_match", "org secret-scanning custom-pattern update-must-not-match", ["custom_pattern_version", "must_not_match"], CUSTOM_PATTERN_PROPERTIES),
        ],
    },
    ("PATCH", "/repos/{owner}/{repo}/secret-scanning/custom-patterns/{pattern_id}"): {
        "arms": [
            one_of_arm("repo_custom_pattern_update_pattern", "repo secret-scanning custom-pattern update-pattern", ["custom_pattern_version", "pattern"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("repo_custom_pattern_update_start_delimiter", "repo secret-scanning custom-pattern update-start-delimiter", ["custom_pattern_version", "start_delimiter"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("repo_custom_pattern_update_end_delimiter", "repo secret-scanning custom-pattern update-end-delimiter", ["custom_pattern_version", "end_delimiter"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("repo_custom_pattern_update_must_match", "repo secret-scanning custom-pattern update-must-match", ["custom_pattern_version", "must_match"], CUSTOM_PATTERN_PROPERTIES),
            one_of_arm("repo_custom_pattern_update_must_not_match", "repo secret-scanning custom-pattern update-must-not-match", ["custom_pattern_version", "must_not_match"], CUSTOM_PATTERN_PROPERTIES),
        ],
    },
    ("POST", "/orgs/{org}/attestations/delete-request"): {
        "arms": [
            one_of_arm(
                "orgs_attestations_delete_request_by_subject_digests",
                "orgs attestations delete-by-subject-digests",
                ["subject_digests"],
                {"subject_digests": {"type": "array"}},
                destructive=True,
            ),
            one_of_arm(
                "orgs_attestations_delete_request_by_attestation_ids",
                "orgs attestations delete-by-attestation-ids",
                ["attestation_ids"],
                {"attestation_ids": {"type": "array"}},
                destructive=True,
            ),
        ],
    },
    ("POST", "/users/{username}/attestations/delete-request"): {
        "arms": [
            one_of_arm(
                "users_attestations_delete_request_by_subject_digests",
                "users attestations delete-by-subject-digests",
                ["subject_digests"],
                {"subject_digests": {"type": "array"}},
                destructive=True,
            ),
            one_of_arm(
                "users_attestations_delete_request_by_attestation_ids",
                "users attestations delete-by-attestation-ids",
                ["attestation_ids"],
                {"attestation_ids": {"type": "array"}},
                destructive=True,
            ),
        ],
    },
    ("POST", "/orgs/{org}/campaigns"): {
        "arms": [
            one_of_arm(
                "orgs_campaigns_create_code_scanning",
                "orgs campaigns create-code-scanning",
                ["code_scanning_alerts", "description", "ends_at", "name"],
                dict(CAMPAIGN_COMMON_PROPERTIES, code_scanning_alerts={"type": "array"}),
            ),
            one_of_arm(
                "orgs_campaigns_create_secret_scanning",
                "orgs campaigns create-secret-scanning",
                ["description", "ends_at", "name", "secret_scanning_alerts"],
                dict(CAMPAIGN_COMMON_PROPERTIES, secret_scanning_alerts={"type": "array"}),
            ),
        ],
    },
    ("POST", "/orgs/{org}/projectsV2/{project_number}/fields"): {
        "arms": [
            one_of_arm(
                "orgs_projectsv2_fields_create_existing_issue_field",
                "orgs projects fields create-existing-issue-field",
                ["issue_field_id"],
                {"issue_field_id": {"type": "integer"}},
            ),
            one_of_arm(
                "orgs_projectsv2_fields_create_new_field",
                "orgs projects fields create-new-field",
                ["data_type", "name"],
                PROJECT_FIELD_PROPERTIES,
            ),
            one_of_arm(
                "orgs_projectsv2_fields_create_single_select",
                "orgs projects fields create-single-select",
                ["data_type", "name", "single_select_options"],
                dict(PROJECT_FIELD_PROPERTIES, single_select_options={"type": "array"}),
            ),
            one_of_arm(
                "orgs_projectsv2_fields_create_iteration",
                "orgs projects fields create-iteration",
                ["data_type", "iteration_configuration", "name"],
                dict(PROJECT_FIELD_PROPERTIES, iteration_configuration={"type": "object"}),
            ),
        ],
    },
    ("POST", "/users/{username}/projectsV2/{project_number}/fields"): {
        "arms": [
            one_of_arm(
                "users_projectsv2_fields_create_new_field",
                "users projects fields create-new-field",
                ["data_type", "name"],
                PROJECT_FIELD_PROPERTIES,
            ),
            one_of_arm(
                "users_projectsv2_fields_create_single_select",
                "users projects fields create-single-select",
                ["data_type", "name", "single_select_options"],
                dict(PROJECT_FIELD_PROPERTIES, single_select_options={"type": "array"}),
            ),
            one_of_arm(
                "users_projectsv2_fields_create_iteration",
                "users projects fields create-iteration",
                ["data_type", "iteration_configuration", "name"],
                dict(PROJECT_FIELD_PROPERTIES, iteration_configuration={"type": "object"}),
            ),
        ],
    },
    ("POST", "/orgs/{org}/projectsV2/{project_number}/items"): {
        "arms": [
            one_of_arm(
                "orgs_projectsv2_items_create_by_id",
                "orgs projects items create-by-id",
                ["id", "type"],
                PROJECT_ITEM_PROPERTIES,
            ),
            one_of_arm(
                "orgs_projectsv2_items_create_by_repo_number",
                "orgs projects items create-by-repo-number",
                ["number", "owner", "repo", "type"],
                PROJECT_ITEM_PROPERTIES,
            ),
        ],
    },
    ("POST", "/users/{username}/projectsV2/{project_number}/items"): {
        "arms": [
            one_of_arm(
                "users_projectsv2_items_create_by_id",
                "users projects items create-by-id",
                ["id", "type"],
                PROJECT_ITEM_PROPERTIES,
            ),
            one_of_arm(
                "users_projectsv2_items_create_by_repo_number",
                "users projects items create-by-repo-number",
                ["number", "owner", "repo", "type"],
                PROJECT_ITEM_PROPERTIES,
            ),
        ],
    },
    ("POST", "/user/codespaces"): {
        "arms": [
            one_of_arm(
                "user_codespaces_create_from_repository",
                "codespaces create-from-repository",
                ["repository_id"],
                {
                    "repository_id": {"type": "integer"},
                    "ref": {"type": "string"},
                    "location": {"type": "string"},
                    "geo": {"type": "string"},
                    "client_ip": {"type": "string"},
                    "machine": {"type": "string"},
                    "devcontainer_path": {"type": "string"},
                    "multi_repo_permissions_opt_out": {"type": "boolean"},
                    "working_directory": {"type": "string"},
                    "idle_timeout_minutes": {"type": "integer"},
                    "display_name": {"type": "string"},
                    "retention_period_minutes": {"type": "integer"},
                },
            ),
            one_of_arm(
                "user_codespaces_create_from_pull_request",
                "codespaces create-from-pull-request",
                ["pull_request"],
                {
                    "pull_request": {"type": "object"},
                    "location": {"type": "string"},
                    "geo": {"type": "string"},
                    "machine": {"type": "string"},
                    "devcontainer_path": {"type": "string"},
                    "working_directory": {"type": "string"},
                    "idle_timeout_minutes": {"type": "integer"},
                },
            ),
        ],
    },
}

def slugify(s):
    return re.sub(r"[^a-z0-9]+", "_", s.lower()).strip("_")

def path_params(path):
    return PATH_PARAM_RE.findall(path)

# The two contracts the runtime actually enforces, mirroring
# connectors.ConfirmationForWriteAction: a DELETE is destructive by
# construction, and destructive_action declares confirm=destructive on any
# method. Everything else is a safe write, and PlanConnectorCommand hands those
# an approval token at plan time, so a preview is available but not required.
# Emitting the destructive sentence for all of them documented a gate that is
# not there.
SAFE_WRITE_APPROVAL = "Reverse ETL writes require plan, approval, execute; preview is optional."
DESTRUCTIVE_WRITE_APPROVAL = "Reverse ETL writes require plan, preview, approval, execute."

def command_approval(method, model):
    destructive = method.upper() == "DELETE" or model == "destructive_action"
    return DESTRUCTIVE_WRITE_APPROVAL if destructive else SAFE_WRITE_APPROVAL

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

def scalar_path_schema(name):
    return {"type": "integer"} if "id" in name or "number" in name else {"type": "string"}

def structured_record_flag(name, schema, required=False):
    kind = schema.get("type")
    if kind in ("object", "array"):
        flag_type = "json"
    elif kind in ("string", "integer", "number", "boolean"):
        flag_type = kind
    else:
        raise ValueError(f"oneOf field {name!r} has unsupported schema type {kind!r}")
    return contract_flag(name.replace("_", "-"), flag_type, f"record.{name}", required=required)

def materialized_record_flag(name, schema, required=False):
    """Derive one closed record flag from a concrete top-level property.

    This mirrors the shared connectorgen materializer: scalar fields retain a
    scalar flag, while a declared object or non-string array becomes one
    declaration-bound JSON record flag. It does not accept a raw body or a
    nested mapping, and rejects a union so the endpoint must be expanded into
    separate named actions first.
    """
    kind = schema.get("type")
    if isinstance(kind, list) or kind is None:
        raise ValueError(f"record field {name!r} has no single concrete type")
    values = schema.get("enum")
    if values and kind == "string":
        return contract_flag(name.replace("_", "-"), "enum", f"record.{name}", required=required, values=values)
    if kind in ("string", "integer", "number", "boolean", "object"):
        flag_type = "json" if kind == "object" else kind
        return contract_flag(name.replace("_", "-"), flag_type, f"record.{name}", required=required)
    if kind == "array":
        items = schema.get("items") or {}
        item_type = items.get("type")
        flag_type = "string_array" if item_type == "string" else "json"
        return contract_flag(name.replace("_", "-"), flag_type, f"record.{name}", required=required)
    raise ValueError(f"record field {name!r} has unsupported type {kind!r}")

def promote_partial_structured_write_commands():
    """Promote partial GitHub write aliases once their schema has a faithful CLI form.

    The partial commands were not provider/runtime limitations: each already
    referenced an executable reverse-ETL action, but their required object or
    non-string array inputs lacked a command form. Build the form from the
    action's top-level closed record schema, preserving the existing plan,
    approval, and destructive-confirmation paths. Root unions are deliberately
    rejected here; they require their own named actions rather than a generic
    JSON escape hatch.
    """
    actions = {action["name"]: action for action in writes_doc["actions"]}
    promoted = []
    for command in cli_doc["commands"]:
        if command.get("availability") != "partial" or not command.get("write"):
            continue
        action = actions.get(command["write"])
        if action is None:
            raise ValueError(f"partial command {command['path']!r} references missing write {command['write']!r}")
        schema = action.get("record_schema") or {}
        if "oneOf" in schema or "anyOf" in schema:
            raise ValueError(f"partial command {command['path']!r} has a root union and needs separately named actions")
        properties = schema.get("properties") or {}
        required = set(schema.get("required") or [])
        existing = {flag.get("maps_to"): flag for flag in command.get("flags", [])}
        for name in sorted(required):
            field = properties.get(name)
            if field is None:
                raise ValueError(f"partial command {command['path']!r} requires undeclared field {name!r}")
            target = f"record.{name}"
            flag = existing.get(target)
            if flag is not None:
                flag["required"] = True
                continue
            flag = materialized_record_flag(name, field, required=True)
            flag["summary"] = f"Required {name.replace('_', ' ')} record field."
            command.setdefault("flags", []).append(flag)
        command["flags"].sort(key=lambda flag: flag["name"])
        command["availability"] = "implemented"
        command.pop("notes", None)
        promoted.append(command["path"])
    return promoted

# These gh-familiar paths are compatibility aliases, never hand-authored
# transports. Each is cloned from the exact already-generated typed command it
# represents, so the alias inherits the provider endpoint, fixed GraphQL
# document, approval lifecycle, and typed confirmation (where applicable).
# This keeps legacy discoverability from becoming a second execution surface.
LEGACY_ALIAS_SOURCE_COMMANDS = {
    "issue view": "issues view",
    "pr view": "pulls view",
    "release view": "releases view",
    "ruleset view": "rulesets view",
    "run view": "actions runs view",
    "workflow view": "actions workflows view",
    "discussion create": "graphql mutation create-discussion",
    "issue status": "issues list-for-authenticated-user",
    "pr checks": "commits check-runs view",
    "pr status": "pr list",
    "project create": "graphql mutation create-project-v2",
    "ruleset check": "rules branches view",
    "search prs": "search issues",
    "status": "graphql query viewer",
    "issue delete": "graphql mutation delete-issue",
    "issue transfer": "graphql mutation transfer-issue",
    "pr revert": "graphql mutation revert-pull-request",
}

# These two legacy commands deliberately remain non-executable: printing a
# stored credential and accepting a caller-controlled authenticated request
# are both incompatible with the connector boundary. `unsupported_local`
# documents that no provider endpoint is omitted while ensuring the command
# can neither acquire a credential value nor bypass declared operations.
NONEXECUTABLE_SAFETY_ALIASES = {
    "auth token": "Named dependency: a credential-export policy would violate pm's no-secret-disclosure boundary.",
    "api": "Named dependency: a constrained declared-operation catalog; generic authenticated API dispatch is intentionally absent.",
}

def restore_legacy_command_aliases():
    commands_by_path = {command.get("path"): command for command in cli_doc["commands"]}
    endpoints_by_key = {
        (endpoint.get("method"), endpoint.get("path")): endpoint
        for endpoint in api.get("endpoints", [])
    }
    restored = []
    for alias_path, source_path in LEGACY_ALIAS_SOURCE_COMMANDS.items():
        target = commands_by_path.get(alias_path)
        source = commands_by_path.get(source_path)
        if target is None:
            raise ValueError(f"legacy GitHub alias {alias_path!r} is missing from cli_surface.json")
        if source is None:
            raise ValueError(f"legacy GitHub alias {alias_path!r} has no generated source command {source_path!r}")
        # JSON round-trip makes a deep copy without adding another dependency
        # to the deterministic generator. It also preserves only artifact data.
        restored_command = json.loads(json.dumps(source))
        restored_command["path"] = alias_path
        restored_command["summary"] = target.get("summary", source.get("summary", alias_path))
        restored_command["source_cli_path"] = "gh " + alias_path
        restored_command["notes"] = f"Compatibility alias of {source_path}; uses that declaration-owned provider contract."
        restored_command.pop("examples", None)
        target.clear()
        target.update(restored_command)
        # REST direct-read coverage is command-path based rather than
        # operation-ID based. Name the alias beside its canonical generated
        # command in the single endpoint row so validation proves both paths
        # reach the same declared request; the documented endpoint count stays
        # unchanged.
        if restored_command.get("intent") == "direct_read" and not str(restored_command.get("operation", "")).startswith("github.graphql."):
            for endpoint_ref in restored_command.get("api_surface", []):
                key = (endpoint_ref.get("method"), endpoint_ref.get("path"))
                endpoint = endpoints_by_key.get(key)
                if endpoint is None:
                    raise ValueError(f"legacy GitHub alias {alias_path!r} references missing endpoint {key!r}")
                covered = endpoint.setdefault("covered_by", {})
                existing_reads = []
                if covered.get("direct_read"):
                    existing_reads.append(covered.pop("direct_read"))
                existing_reads.extend(covered.get("direct_reads", []))
                if alias_path not in existing_reads:
                    existing_reads.append(alias_path)
                covered["direct_reads"] = sorted(set(existing_reads))
        restored.append(alias_path)

    for alias_path, reason in NONEXECUTABLE_SAFETY_ALIASES.items():
        target = commands_by_path.get(alias_path)
        if target is None:
            raise ValueError(f"non-executable GitHub safety alias {alias_path!r} is missing from cli_surface.json")
        target["availability"] = "unsupported_local"
        target["notes"] = reason
        target.pop("operation", None)
        target.pop("write", None)
        target.pop("api_surface", None)
    return restored

def append_explicit_one_of_write_contract(endpoint, operation, contract):
    """Append all concrete write contracts for one documented oneOf endpoint."""
    method = endpoint["method"]
    api_path = endpoint["path"]
    params = path_params(api_path)
    path_fields = [p for p in params if p not in ("owner", "repo")]
    path_properties = {p: scalar_path_schema(p) for p in path_fields}
    actions = []

    for arm in contract["arms"]:
        action_name = arm["action"]
        cli_path = arm["cli_path"]
        op_id = f"github.{action_name}"
        if action_name in existing_write_names or action_name in seen_write:
            raise ValueError(f"oneOf action name collision: {action_name}")
        if cli_path in existing_cli_paths or cli_path in seen_cli_path:
            raise ValueError(f"oneOf CLI path collision: {cli_path}")
        if op_id in existing_op_ids or op_id in seen_op:
            raise ValueError(f"oneOf operation id collision: {op_id}")

        record_properties = dict(path_properties)
        record_properties.update(arm["properties"])
        required = list(path_fields) + list(arm["required"])
        action = {
            "name": action_name,
            "kind": {"POST": "create", "PATCH": "update", "PUT": "update", "DELETE": "delete"}[method],
            "method": method,
            "path": engine_write_path(api_path),
            "path_fields": path_fields,
            "body_type": "json",
            "record_schema": {
                "$schema": "http://json-schema.org/draft-07/schema#",
                "type": "object",
                "additionalProperties": False,
                "required": required,
                "properties": record_properties,
            },
            "risk": operation.get("risk", "medium"),
        }
        if arm.get("body_type"):
            action["body_type"] = arm["body_type"]
            action["body_field"] = arm["body_field"]
            action["body_schema"] = arm["body_schema"]
        mutation_class = "admin"
        approval = SAFE_WRITE_APPROVAL
        operation_approval = "plan, approval, execute"
        if arm.get("destructive"):
            mutation_class = "destructive"
            approval = DESTRUCTIVE_WRITE_APPROVAL
            operation_approval = "plan, preview, approval, execute (caller-supplied intent acknowledgement)"
            action["confirm"] = "destructive"

        new_writes.append(action)
        new_ops.append({
            "id": op_id,
            "kind": "rest_write",
            "summary": f"{method} {api_path} ({arm['action']})",
            "source_url": operation.get("source_url", ""),
            "risk": operation.get("risk", "medium"),
            "approval": operation_approval,
            "output_policy": "json",
            "mutation_class": mutation_class,
            "destructive": True if arm.get("destructive") else None,
            "rest": {"method": method, "path": api_path},
        })
        # Omit a false JSON key instead of making an authored non-destructive
        # operation look like it needs an explicit exception from the safety
        # rule. The write action remains the confirmation source of truth.
        if not arm.get("destructive"):
            new_ops[-1].pop("destructive")

        flags = [contract_flag(p.replace("_", "-"), scalar_path_schema(p)["type"], f"record.{p}", required=True) for p in path_fields]
        required_fields = set(arm["required"])
        flags.extend(
            structured_record_flag(name, schema, required=name in required_fields)
            for name, schema in arm["properties"].items()
        )
        new_cmds.append({
            "path": cli_path,
            "summary": f"{method} {api_path} ({arm['action']})",
            "intent": "reverse_etl",
            "availability": "implemented",
            "write": action_name,
            "source_cli_path": "",
            "risk": operation.get("risk", "medium"),
            "approval": approval,
            "flags": flags,
        })
        seen_write.add(action_name)
        seen_cli_path.add(cli_path)
        seen_op.add(op_id)
        actions.append(action_name)

    endpoint["covered_by"] = {"writes": actions}

promoted_partial_write_commands = promote_partial_structured_write_commands()

new_ops, new_writes, new_cmds = [], [], []
covered_added = 0
seen_op, seen_write, seen_cli = set(), set(), set()
seen_cli_path = set()

for e in api["endpoints"]:
    op = e.get("operation")
    if not op:
        continue
    model = op.get("model")
    method = e.get("method", "GET")
    api_path = e.get("path", "")
    explicit_read = EXPLICIT_DIRECT_READ_CONTRACTS.get((method, api_path))
    explicit_one_of_write = EXPLICIT_ONE_OF_WRITE_CONTRACTS.get((method, api_path))
    completion_params = None
    completion_body = None
    completion_contract = model in ("duplicate", "deprecated", "disallowed")
    if completion_contract and not explicit_read and not explicit_one_of_write:
        # Former exclusion labels are accounting metadata, not an execution
        # contract. Re-open each pinned endpoint with its own method, path,
        # declared parameters, and (for writes) closed request schema.
        completion_params = openapi_parameters(method, api_path)
        if method == "GET":
            model = "direct_read"
        else:
            completion_body = openapi_body_schema(method, api_path)
            # A root union cannot become one permissive action. Leave it for
            # the explicit arm table below, where every arm has a stable name.
            if completion_body is None:
                continue
            if completion_body.get("type") not in (None, "object"):
                raise ValueError(f"{method} {api_path} body is not a concrete object contract")
            model = completion_write_model(method, api_path)
    if e.get("covered_by"):
        continue  # already converted
    e.pop("operation", None)  # covered_by replaces the operation classifier
    if explicit_one_of_write:
        append_explicit_one_of_write_contract(e, op, explicit_one_of_write)
        covered_added += 1
        continue
    params = path_params(api_path)
    if explicit_read:
        model = "direct_read"
    # derive names
    cli_path = explicit_read["cli_path"] if explicit_read else derive_cli_path(method, api_path, model)
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
    if not explicit_read and method != "GET" and model in ("direct_read", "binary_read"):
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
        completion_params_by_name = {parameter["name"]: parameter for parameter in completion_params or []}
        for pf in path_fields:
            source_param = completion_params_by_name.get(pf)
            record_props[pf] = dict(source_param["schema"]) if source_param else scalar_path_schema(pf)
            required.append(pf)
        if completion_body is not None:
            for name, field in (completion_body.get("properties") or {}).items():
                record_props.setdefault(name, field)
            required.extend(completion_body.get("required", []))
        wa = {
            "name": wname, "kind": {"POST": "create", "PATCH": "update", "PUT": "update", "DELETE": "delete"}[method],
            "method": method, "path": engine_write_path(api_path), "path_fields": path_fields, "body_type": "json",
            "record_schema": {"$schema": "http://json-schema.org/draft-07/schema#", "type": "object",
                              "required": sorted(set(required)), "properties": record_props},
            "risk": op.get("risk", "medium"),
        }
        if completion_body is not None and completion_body.get("additionalProperties") is False:
            wa["record_schema"]["additionalProperties"] = False
        sensitive_fields = SENSITIVE_COMPLETION_FIELDS.get((method, api_path), [])
        if sensitive_fields:
            # The generic plan lifecycle withholds these exact declared fields
            # from state, samples, previews, and run records. The source field
            # list remains narrow rather than redacting an entire request.
            wa["redact_fields"] = sensitive_fields
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
        if completion_body is not None:
            flags = source_contract_flags(completion_params, wa["record_schema"], write=True)
        else:
            flags = [{"name": p.replace("_", "-"), "type": "integer" if ("id" in p or "number" in p) else "string",
                      "summary": f"{p} path parameter", "maps_to": f"record.{p}"} for p in path_fields]
        cmd = {"path": cli_path, "summary": f"{method} {api_path}", "intent": "reverse_etl",
               "availability": "implemented", "write": wname, "source_cli_path": "",
               "risk": op.get("risk", "medium"),
               "approval": command_approval(method, model),
               "flags": flags}
        # No `notes` marker. The typed confirmation is derived from the bound
        # write action by commandrunner.ConfirmationChallengeForCommand and
        # rendered once by the help/manual/skill CONFIRMATION field. The marker
        # this used to emit named an opt-in flag pm has never parsed, and a
        # per-command note is silent on every command nobody annotated, which
        # reads as "no confirmation needed". See the phase TDD ledger, red 7b.
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
        if explicit_read:
            rest = {"method": method, "path": api_path, "max_bytes": explicit_read["max_bytes"]}
            if "content_type" in explicit_read:
                rest["content_type"] = explicit_read["content_type"]
            if "body_schema" in explicit_read:
                rest["body_schema"] = explicit_read["body_schema"]
            output_policy = explicit_read["output_policy"]
            flags = explicit_read["flags"]
        else:
            rest = {"method": "GET", "path": api_path}
            output_policy = "json"
            if completion_params is not None:
                flags = source_contract_flags(completion_params, None)
            else:
                flags = [{"name": p.replace("_", "-"), "type": "integer" if ("id" in p or "number" in p) else "string",
                          "summary": f"{p} path parameter"} for p in params if p not in ("owner", "repo")]
        new_ops.append({"id": op_id, "kind": "rest_read", "summary": f"Read {api_path}",
                        "source_url": op.get("source_url", ""), "risk": op.get("risk", "low"),
                        "approval": "none", "output_policy": output_policy, "rest": rest})
        cmd = {"path": cli_path, "summary": f"Read {api_path}", "intent": "direct_read",
               "availability": "implemented", "operation": op_id, "source_cli_path": "", "flags": flags,
               "output_policy": output_policy}
        new_cmds.append(cmd)
        e["covered_by"] = {"direct_reads": [cli_path]}
    covered_added += 1

ops_doc["operations"].extend(new_ops)
writes_doc["actions"].extend(new_writes)
cli_doc["commands"].extend(new_cmds)

normalized_reverse_etl_path_flags = normalize_reverse_etl_path_flags()
restored_legacy_aliases = restore_legacy_command_aliases()

def write_generated_json(path, value):
    # Keep generated artifacts byte-stable when their semantic content did not
    # change. Existing source artifacts intentionally differ on their final
    # newline convention, so preserve that convention as well as Unicode text.
    with open(path, encoding="utf-8") as source:
        existing = source.read()
    encoded = json.dumps(value, indent=2, ensure_ascii=False)
    if existing.endswith("\n"):
        encoded += "\n"
    if encoded == existing:
        return
    with open(path, "w", encoding="utf-8") as output:
        output.write(encoded)

write_generated_json(f"{ROOT}/operations.json", ops_doc)
write_generated_json(f"{ROOT}/writes.json", writes_doc)
write_generated_json(f"{ROOT}/cli_surface.json", cli_doc)
write_generated_json(f"{ROOT}/api_surface.json", api)

print(f"generated: {covered_added} endpoints covered")
print(f"  new operations: {len(new_ops)} (total {len(ops_doc['operations'])})")
print(f"  new write actions: {len(new_writes)} (total {len(writes_doc['actions'])})")
print(f"  new cli commands: {len(new_cmds)} (total {len(cli_doc['commands'])})")
print(f"  promoted partial structured writes: {len(promoted_partial_write_commands)}")
print(f"  normalized reverse-ETL path flags: {normalized_reverse_etl_path_flags}")
print(f"  restored legacy aliases: {len(restored_legacy_aliases)}")
