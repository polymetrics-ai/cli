---
name: pm-dremio
description: Dremio connector knowledge and safe action guide.
---

# pm-dremio

## Purpose

Reads and writes Dremio catalog entries, reflections, sources, users, and roles through the Dremio REST API.

## Icon

- asset: icons/dremio.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.dremio.com/software/rest-api/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- page_size
- api_key (secret)

## ETL Streams

- catalog:
  - primary key: id
  - fields: containerType(), createdAt(), datasetType(), id(), path(), tag(), type()
- reflections:
  - primary key: id
  - fields: createdAt(), datasetId(), enabled(), id(), name(), status(), type(), updatedAt()
- sources:
  - primary key: id
  - fields: createdAt(), id(), name(), path(), tag(), type()
- users:
  - primary key: id
  - fields: active(), email(), firstName(), id(), lastName(), name()
- roles:
  - primary key: id
  - fields: description(), id(), memberCount(), name(), roles(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_user:
  - endpoint: POST /user
  - risk: creates a new Dremio user account with instance-wide access, scoped by whatever role assignment follows; external mutation, approval required
- update_user:
  - endpoint: PUT /user/{{ record.id }}
  - required fields: id
  - risk: mutates an existing Dremio user account, including its active flag which can lock the user out; external mutation, approval required
- delete_user:
  - endpoint: DELETE /user/{{ record.id }}
  - required fields: id
  - risk: permanently removes a Dremio user account and revokes its access; destructive, approval required
- create_role:
  - endpoint: POST /role
  - risk: creates a new Dremio role; low external mutation risk on its own until members/grants are attached
- update_role:
  - endpoint: PUT /role/{{ record.id }}
  - required fields: id
  - risk: mutates an existing Dremio role's name/description; external mutation, approval required
- delete_role:
  - endpoint: DELETE /role/{{ record.id }}
  - required fields: id
  - risk: permanently removes a Dremio role and its grants for every member; destructive, approval required
- update_reflection:
  - endpoint: PUT /reflection/{{ record.id }}
  - required fields: id
  - risk: mutates an existing reflection's definition (name/enabled/tag); disabling a reflection removes its query-acceleration benefit until re-enabled and rebuilt; external mutation, approval required
- refresh_reflection:
  - endpoint: POST /reflection/{{ record.id }}/refresh
  - required fields: id
  - risk: forces an immediate reflection rebuild, consuming cluster compute; low external-mutation risk, no data loss, no approval required
- delete_reflection:
  - endpoint: DELETE /reflection/{{ record.id }}
  - required fields: id
  - risk: permanently removes a reflection definition; any query relying on it for acceleration falls back to the raw dataset; destructive, approval required
- create_personal_access_token:
  - endpoint: POST /user/{{ record.user_id }}/token
  - required fields: user_id
  - optional fields: label, millisToExpire
  - risk: mints a new long-lived Personal Access Token credential for the named user; the response body carries the plaintext token exactly once and must never be logged; external mutation, approval required
- delete_personal_access_token:
  - endpoint: DELETE /user/{{ record.user_id }}/token/{{ record.token_id }}
  - required fields: user_id, token_id
  - risk: revokes a single Personal Access Token, immediately invalidating any client still using it; destructive to that credential's holders, approval required

## Security

- read risk: external Dremio API read of catalog, reflection, source, user, and role data
- write risk: external Dremio API mutation of user/role/reflection/PAT lifecycle objects; several actions are destructive (delete_user, delete_role, delete_reflection, delete_personal_access_token) and require approval
- approval: required for create/update/delete user, role, and reflection actions and PAT lifecycle actions; refresh_reflection is low-risk and does not require approval
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect dremio
```

### Inspect as structured JSON

```bash
pm connectors inspect dremio --json
```

## Agent Rules

- Run pm connectors inspect dremio before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
