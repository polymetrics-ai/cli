---
name: pm-statsig
description: Statsig connector knowledge and safe action guide.
---

# pm-statsig

## Purpose

Reads and manages Statsig feature gates, dynamic configs, experiments, segments, target apps, tags, keys, holdouts, layers, users, audit logs, and environments through the Statsig Console API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- feature_gates:
  - primary key: id
  - fields: description(), id(), isEnabled(), name(), status()
- dynamic_configs:
  - primary key: id
  - fields: description(), id(), isEnabled(), name(), status()
- experiments:
  - primary key: id
  - fields: description(), id(), isEnabled(), name(), status()
- segments:
  - primary key: id
  - fields: description(), id(), isEnabled(), name(), status()
- target_apps:
  - primary key: id
  - fields: id(), name()
- tags:
  - primary key: id
  - fields: description(), id(), isCore(), name()
- keys:
  - primary key: key
  - fields: description(), environments(), key(), lastUsed(), primaryTargetApp(), scopes(), secondaryTargetApps(), status(), type()
- holdouts:
  - primary key: id
  - fields: createdTime(), creatorEmail(), creatorID(), creatorName(), description(), experimentIDs(), gateIDs(), id(), idType(), isEnabled(), isGlobal(), lastModifiedTime(), lastModifierID(), layerIDs(), name(), passPercentage(), status(), team(), teamID()
- layers:
  - primary key: id
  - fields: createdTime(), creatorEmail(), creatorID(), creatorName(), description(), id(), idType(), isImplicitLayer(), lastModifiedTime(), lastModifierID(), name(), team(), teamID()
- users:
  - primary key: userID
  - fields: email(), firstName(), lastName(), role(), userID()
- audit_logs:
  - primary key: id
  - fields: actionType(), changeLog(), date(), id(), modifierEmail(), name(), tags(), targetAppIDs(), time(), updatedBy(), updatedByUserID()
- environments:
  - primary key: name
  - fields: id(), isProduction(), name(), requiresReleasePipeline(), requiresReview()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_gate:
  - endpoint: POST /gates
  - risk: external mutation; approval required
- update_gate:
  - endpoint: PATCH /gates/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_gate:
  - endpoint: DELETE /gates/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_dynamic_config:
  - endpoint: POST /dynamic_configs
  - risk: external mutation; approval required
- update_dynamic_config:
  - endpoint: PATCH /dynamic_configs/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_dynamic_config:
  - endpoint: DELETE /dynamic_configs/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_segment:
  - endpoint: POST /segments
  - risk: external mutation; approval required
- delete_segment:
  - endpoint: DELETE /segments/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_tag:
  - endpoint: POST /tags
  - risk: external mutation; approval required
- update_tag:
  - endpoint: PATCH /tags/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_tag:
  - endpoint: DELETE /tags/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_target_app:
  - endpoint: POST /target_app
  - risk: external mutation; approval required
- update_target_app:
  - endpoint: PATCH /target_app/{{ record.id }}
  - required fields: id
  - risk: external mutation; approval required
- delete_target_app:
  - endpoint: DELETE /target_app/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_holdout:
  - endpoint: POST /holdouts
  - risk: external mutation; approval required
- delete_holdout:
  - endpoint: DELETE /holdouts/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_layer:
  - endpoint: POST /layers
  - risk: external mutation; approval required
- delete_layer:
  - endpoint: DELETE /layers/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion; approval required
- create_key:
  - endpoint: POST /keys
  - risk: external mutation creating a live API credential; approval required
- delete_key:
  - endpoint: DELETE /keys/{{ record.key }}
  - required fields: key
  - risk: irreversible external deletion of a live API credential; approval required

## Security

- read risk: external Statsig Console API read of feature gates, dynamic configs, experiments, segments, target apps, tags, keys, holdouts, layers, users, audit logs, and environments
- write risk: external mutation of Statsig feature gates, dynamic configs, segments, tags, target apps, holdouts, layers, and API keys, including irreversible deletes and live-credential creation/deletion
- approval: read: none; write: required for all mutation actions
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect statsig
```

### Inspect as structured JSON

```bash
pm connectors inspect statsig --json
```

## Agent Rules

- Run pm connectors inspect statsig before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
