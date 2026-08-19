# pm connectors inspect statsig

```text
NAME
  pm connectors inspect statsig - Statsig connector manual

SYNOPSIS
  pm connectors inspect statsig
  pm connectors inspect statsig --json
  pm credentials add <name> --connector statsig [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and manages Statsig feature gates, dynamic configs, experiments, segments, target apps, tags, keys, holdouts, layers, users, audit logs, and environments through the Statsig Console API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)

ETL STREAMS
  feature_gates:
    primary key: id
    fields: description(string), id(string), isEnabled(boolean), name(string), status(string)
  dynamic_configs:
    primary key: id
    fields: description(string), id(string), isEnabled(boolean), name(string), status(string)
  experiments:
    primary key: id
    fields: description(string), id(string), isEnabled(boolean), name(string), status(string)
  segments:
    primary key: id
    fields: description(string), id(string), isEnabled(boolean), name(string), status(string)
  target_apps:
    primary key: id
    fields: id(string), name(string)
  tags:
    primary key: id
    fields: description(string), id(string), isCore(boolean), name(string)
  keys:
    primary key: key
    fields: description(string), environments(array), key(string), lastUsed(string), primaryTargetApp(string), scopes(array), secondaryTargetApps(array), status(string), type(string)
  holdouts:
    primary key: id
    fields: createdTime(number), creatorEmail(string), creatorID(string), creatorName(string), description(string), experimentIDs(array), gateIDs(array), id(string), idType(string), isEnabled(boolean), isGlobal(boolean), lastModifiedTime(number), lastModifierID(string), layerIDs(array), name(string), passPercentage(number), status(string), team(string), teamID(string)
  layers:
    primary key: id
    fields: createdTime(number), creatorEmail(string), creatorID(string), creatorName(string), description(string), id(string), idType(string), isImplicitLayer(boolean), lastModifiedTime(number), lastModifierID(string), name(string), team(string), teamID(string)
  users:
    primary key: userID
    fields: email(string), firstName(string), lastName(string), role(string), userID(string)
  audit_logs:
    primary key: id
    fields: actionType(string), changeLog(string), date(string), id(string), modifierEmail(string), name(string), tags(array), targetAppIDs(array), time(number), updatedBy(string), updatedByUserID(string)
  environments:
    primary key: name
    fields: id(string), isProduction(boolean), name(string), requiresReleasePipeline(boolean), requiresReview(boolean)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_gate:
    endpoint: POST /gates
    required fields: name
    risk: external mutation; approval required
  update_gate:
    endpoint: PATCH /gates/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_gate:
    endpoint: DELETE /gates/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_dynamic_config:
    endpoint: POST /dynamic_configs
    required fields: name
    risk: external mutation; approval required
  update_dynamic_config:
    endpoint: PATCH /dynamic_configs/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_dynamic_config:
    endpoint: DELETE /dynamic_configs/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_segment:
    endpoint: POST /segments
    required fields: name, type
    risk: external mutation; approval required
  delete_segment:
    endpoint: DELETE /segments/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_tag:
    endpoint: POST /tags
    required fields: name, description
    risk: external mutation; approval required
  update_tag:
    endpoint: PATCH /tags/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_target_app:
    endpoint: POST /target_app
    required fields: name, description
    risk: external mutation; approval required
  update_target_app:
    endpoint: PATCH /target_app/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_target_app:
    endpoint: DELETE /target_app/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_holdout:
    endpoint: POST /holdouts
    required fields: name
    risk: external mutation; approval required
  delete_holdout:
    endpoint: DELETE /holdouts/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_layer:
    endpoint: POST /layers
    required fields: name, idType
    risk: external mutation; approval required
  delete_layer:
    endpoint: DELETE /layers/{{ record.id }}
    required fields: id
    risk: irreversible external deletion; approval required
  create_key:
    endpoint: POST /keys
    required fields: description, type
    risk: external mutation creating a live API credential; approval required
  delete_key:
    endpoint: DELETE /keys/{{ record.key }}
    required fields: key
    risk: irreversible external deletion of a live API credential; approval required

SECURITY
  read risk: external Statsig Console API read of feature gates, dynamic configs, experiments, segments, target apps, tags, keys, holdouts, layers, users, audit logs, and environments
  write risk: external mutation of Statsig feature gates, dynamic configs, segments, tags, target apps, holdouts, layers, and API keys, including irreversible deletes and live-credential creation/deletion
  approval: read: none; write: required for all mutation actions
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect statsig

  # Inspect as structured JSON
  pm connectors inspect statsig --json

AGENT WORKFLOW
  - Run pm connectors inspect statsig before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
