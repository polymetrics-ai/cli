---
name: pm-breezy-hr
description: Breezy HR connector knowledge and safe action guide.
---

# pm-breezy-hr

## Purpose

Reads Breezy HR positions, hiring pipelines, per-position candidates, departments, categories, custom attribute definitions, questionnaires, and message templates; writes position create/update/state-change and candidate create/update/pipeline-stage-move mutations, through the Breezy v3 REST API.

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
- api_key (secret)
- company_id (secret)

## ETL Streams

- positions:
  - primary key: position_id
  - fields: country_id(), country_name(), creation_date(), department(), name(), org_type(), pipeline_id(), position_id(), state(), type(), updated_date()
- pipelines:
  - primary key: id
  - fields: id(), name()
- candidates:
  - primary key: id
  - fields: creation_date(), email_address(), headline(), id(), name(), origin(), phone_number(), position_id(), stage(), updated_date()
- departments:
  - primary key: id
  - fields: id(), name()
- categories:
  - primary key: id
  - fields: id(), name()
- custom_attributes_candidate:
  - primary key: id
  - fields: id(), name(), secure()
- custom_attributes_position:
  - primary key: id
  - fields: id(), name(), secure()
- questionnaires:
  - primary key: id
  - fields: id(), name()
- templates:
  - primary key: id
  - fields: id(), name()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_position:
  - endpoint: POST /positions
  - risk: creates a new job opening; if not left in draft state, may become publicly visible on the company's careers page and job boards depending on the configured state
- update_position:
  - endpoint: PUT /position/{{ record.position_id }}
  - required fields: position_id
  - risk: mutates an existing job opening's title/description/location/department; a live (published) posting's public listing reflects the change immediately
- update_position_state:
  - endpoint: PUT /position/{{ record.position_id }}/state
  - required fields: position_id
  - risk: changes a position's lifecycle state (published/draft/closed/archived); setting state to published makes the job publicly visible on the company's careers page and job boards, and closed/archived stops accepting new applicants
- create_candidate:
  - endpoint: POST /position/{{ record.position_id }}/candidates
  - required fields: position_id
  - risk: adds a new candidate to a position's hiring pipeline; low-risk additive mutation, no approval required
- update_candidate:
  - endpoint: PUT /position/{{ record.position_id }}/candidate/{{ record.candidate_id }}
  - required fields: position_id, candidate_id
  - risk: mutates an existing candidate's contact/profile information
- move_candidate_stage:
  - endpoint: PUT /position/{{ record.position_id }}/candidate/{{ record.candidate_id }}/stage
  - required fields: position_id, candidate_id
  - optional fields: stage_id
  - risk: moves a candidate to a different pipeline stage within the SAME position (e.g. Applied to Interviewing to Hired/Disqualified); moving to a terminal stage (hired/disqualified) may trigger configured stage actions (auto-emails, webhook notifications) depending on the position's stage_actions_enabled setting

## Security

- read risk: external Breezy HR API read of company position, hiring pipeline, candidate, department, category, custom-attribute, questionnaire, and template metadata
- write risk: external mutation of Breezy HR positions and candidates; update_position_state can publish a position to the company's public careers page and job boards, and move_candidate_stage to a terminal stage (hired/disqualified) may trigger configured stage-action auto-emails/webhooks — every write ships with an explicit per-action risk string
- approval: none required by default; review update_position_state (publishes job postings) and move_candidate_stage (may trigger candidate-facing communications) before use in an automated pipeline
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect breezy-hr
```

### Inspect as structured JSON

```bash
pm connectors inspect breezy-hr --json
```

## Agent Rules

- Run pm connectors inspect breezy-hr before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
