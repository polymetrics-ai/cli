---
name: pm-split-io
description: Split.io connector knowledge and safe action guide.
---

# pm-split-io

## Purpose

Reads Split.io workspaces, environments, feature flags, segments, groups, traffic types, and users, and writes feature-flag kill/restore/archive/unarchive and segment-key mutations through the Split Admin API.

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
- mode
- workspace_id
- api_key (secret)

## ETL Streams

- workspaces:
  - primary key: id
  - fields: id(), name(), status()
- environments:
  - primary key: id
  - fields: id(), name(), status()
- splits:
  - primary key: id
  - cursor: updatedAt
  - fields: environment(), id(), name(), status(), trafficType(), updatedAt()
- segments:
  - primary key: id
  - cursor: updatedAt
  - fields: id(), name(), status(), updatedAt()
- groups:
  - primary key: id
  - fields: description(), id(), name(), type()
- traffic_types:
  - primary key: id
  - fields: displayAttributeId(), id(), name()
- users:
  - primary key: id
  - fields: email(), groups(), id(), name(), status()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- kill_feature_flag_in_environment:
  - endpoint: PUT /internal/api/v2/splits/ws/{{ record.workspace_id }}/{{ record.feature_flag_name }}/environments/{{ record.environment_id }}/kill
  - required fields: workspace_id, feature_flag_name, environment_id
  - risk: immediately forces every SDK evaluating this feature flag in the given environment onto its off/default treatment; high-impact production traffic-shaping mutation, approval required
- restore_feature_flag_in_environment:
  - endpoint: PUT /internal/api/v2/splits/ws/{{ record.workspace_id }}/{{ record.feature_flag_name }}/environments/{{ record.environment_id }}/restore
  - required fields: workspace_id, feature_flag_name, environment_id
  - risk: reverts a previously-killed feature flag in the given environment back to its configured targeting rules; production traffic-shaping mutation, approval required
- archive_feature_flag:
  - endpoint: PUT /internal/api/v2/splits/ws/{{ record.workspace_id }}/{{ record.feature_flag_name }}/archive
  - required fields: workspace_id, feature_flag_name
  - risk: archives a feature flag account-wide (all SDKs calling it return control); approval required
- unarchive_feature_flag:
  - endpoint: PUT /internal/api/v2/splits/ws/{{ record.workspace_id }}/{{ record.feature_flag_name }}/unarchive
  - required fields: workspace_id, feature_flag_name
  - risk: restores an archived feature flag to active use account-wide; approval required
- add_segment_keys_in_environment:
  - endpoint: PUT /internal/api/v2/segments/{{ record.environment_id }}/{{ record.segment_name }}/uploadKeys
  - required fields: environment_id, segment_name
  - risk: adds member keys to a segment in the given environment, changing which end-users match segment-based targeting rules for every feature flag using it; production traffic-shaping mutation, approval required
- remove_segment_keys_from_environment:
  - endpoint: PUT /internal/api/v2/segments/{{ record.environment_id }}/{{ record.segment_name }}/removeKeys
  - required fields: environment_id, segment_name
  - risk: removes member keys from a segment in the given environment, changing which end-users match segment-based targeting rules for every feature flag using it; production traffic-shaping mutation, approval required

## Security

- read risk: external Split.io API read of workspace, environment, feature-flag, segment, group, traffic-type, and user data
- write risk: external Split.io API mutation that reshapes live feature-flag targeting/rollout state or segment membership for every SDK evaluating it
- approval: reverse ETL plan approval required before all writes; every write action in this bundle changes production feature-flag or segment-targeting behavior
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect split-io
```

### Inspect as structured JSON

```bash
pm connectors inspect split-io --json
```

## Agent Rules

- Run pm connectors inspect split-io before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
