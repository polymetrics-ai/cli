---
name: pm-zendesk-sunshine
description: Zendesk Sunshine connector knowledge and safe action guide.
---

# pm-zendesk-sunshine

## Purpose

Reads and writes Zendesk Sunshine legacy custom object types, objects, relationship types, and relationships.

## Icon

- asset: icons/zendesk-sunshine.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.zendesk.com/api-reference/custom-data/custom-objects-api/introduction/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- email
- object_type
- relationship_type
- api_token (secret)

## ETL Streams

- object_types:
  - primary key: id
  - fields: created_at(), id(), schema()
- relationship_types:
  - primary key: id
  - fields: created_at(), id(), source_type(), target_type(), updated_at()
- objects:
  - primary key: id
  - fields: attributes(), id(), type()
- relationships:
  - primary key: id
  - fields: id(), source(), target(), type()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_object_record:
  - endpoint: POST /objects/records
  - risk: creates a new record of an EXISTING legacy object type; low-risk external mutation, no approval required. Legacy custom object TYPE creation is blocked by Zendesk as of January 15, 2026, but creating records of already-existing types remains supported until the full legacy API sunset (July 27, 2026)
- upsert_object_record_by_external_id:
  - endpoint: PUT /objects/records
  - risk: creates a new record with the given external_id if none exists, or overwrites the attributes object of the existing record with that external_id and type -- an overwrite, not a merge: any attribute omitted from this call's attributes is cleared on an existing record
- delete_object_record:
  - endpoint: DELETE /objects/records/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a legacy object record; Zendesk requires every relationship record referencing this object record to be deleted first, or this request fails
- delete_object_type:
  - endpoint: DELETE /objects/types/{{ record.key }}
  - required fields: key
  - risk: permanently deletes a legacy object type definition; only succeeds if every object record of that type has already been deleted. Standard Zendesk object types (users, tickets, organizations) cannot be deleted this way and the API rejects the request
- create_relationship_record:
  - endpoint: POST /relationships/records
  - risk: links two existing object records (legacy custom object records, or a mix of custom and standard Zendesk records) via an EXISTING relationship type; low-risk external mutation. Legacy relationship TYPE creation is blocked by Zendesk as of January 15, 2026, but creating relationship records of already-existing types remains supported until the full legacy API sunset (July 27, 2026)
- delete_relationship_record:
  - endpoint: DELETE /relationships/records/{{ record.id }}
  - required fields: id
  - risk: permanently removes a relationship record between two object records; irreversible, does not affect either underlying record
- delete_relationship_type:
  - endpoint: DELETE /relationships/types/{{ record.key }}
  - required fields: key
  - risk: permanently deletes a legacy relationship type; only succeeds if every relationship record of that type has already been deleted

## Security

- read risk: external Zendesk account read of legacy custom object types, objects, relationship types, and relationships
- write risk: external mutation: creates/upserts/deletes object records of existing legacy object types, creates/deletes relationship records between them, and deletes object/relationship type definitions. Zendesk blocked creating new legacy object/relationship TYPES as of January 15, 2026, and will fully remove the legacy custom objects API on July 27, 2026 -- this bundle's write surface is scoped to what remains supported today
- approval: required for all write actions; read access uses a read-only Basic-auth API token with no approval needed
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zendesk-sunshine
```

### Inspect as structured JSON

```bash
pm connectors inspect zendesk-sunshine --json
```

## Agent Rules

- Run pm connectors inspect zendesk-sunshine before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
