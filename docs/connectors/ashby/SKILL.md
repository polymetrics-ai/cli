---
name: pm-ashby
description: Ashby connector knowledge and safe action guide.
---

# pm-ashby

## Purpose

Reads Ashby applicant-tracking data — candidates, jobs, applications, and users — through the Ashby REST API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/ashby.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.ashbyhq.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- api_key (secret)

## ETL Streams

- candidates:
  - primary key: id
  - cursor: updatedAt
  - fields: company(), createdAt(), id(), locationSummary(), name(), primaryEmailAddress(), primaryPhoneNumber(), timezone(), title(), updatedAt()
- jobs:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), customFields(), defaultInterviewPlanId(), departmentId(), employmentType(), id(), locationId(), status(), title(), updatedAt()
- applications:
  - primary key: id
  - cursor: updatedAt
  - fields: archiveReason(), candidateId(), createdAt(), currentInterviewStageId(), id(), jobId(), source(), status(), updatedAt()
- users:
  - primary key: id
  - cursor: updatedAt
  - fields: email(), firstName(), globalRole(), id(), isEnabled(), lastName(), updatedAt()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Ashby API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect ashby
```

### Inspect as structured JSON

```bash
pm connectors inspect ashby --json
```

## Agent Rules

- Run pm connectors inspect ashby before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
