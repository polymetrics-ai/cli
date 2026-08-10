---
name: pm-survicate
description: Survicate connector knowledge and safe action guide.
---

# pm-survicate

## Purpose

Reads Survicate surveys, survey questions, responses, and respondent attributes, and manages GDPR personal-data requests, through the Survicate Data Export API v2. Read-only.

## Icon

- id: survicate
- asset: icons/survicate.svg
- source: official
- review_status: official_verified
- review_url: https://developers.survicate.com/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret)

## ETL Streams

- surveys:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), id(string), name(string), updated_at(string)
- survey_questions:
  - primary key: survey_id, id
  - fields: answer_choices(array), columns(array), fields(array), id(integer), introduction(string), question(string), survey_id(string), type(string)
- responses:
  - primary key: uuid
  - fields: attributes(array), collected_at(string), device_type(string), language(string), operating_system(string), questions(array), respondent_uuid(string), survey_id(string), url(string), uuid(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Survicate API read of survey, response, and respondent data
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Survicate's declared streams and reverse-ETL actions.
- Usage: pm survicate <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete v2 personal-data - Documented DELETE /v2/personal-data (not implemented) [intent=direct_write availability=not_implemented operation=survicate.delete.v2-personal-data]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get v2 personal-data - Documented GET /v2/personal-data (not implemented) [intent=direct_read availability=not_implemented operation=survicate.get.v2-personal-data]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 respondents respondent-uuid attributes - Documented GET /v2/respondents/{respondent_uuid}/attributes (not implemented) [intent=direct_read availability=not_implemented operation=survicate.get.v2-respondents-respondent-uuid-attributes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 respondents respondent-uuid responses - Documented GET /v2/respondents/{respondent_uuid}/responses (not implemented) [intent=direct_read availability=not_implemented operation=survicate.get.v2-respondents-respondent-uuid-responses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 surveys survey-id - Documented GET /v2/surveys/{survey_id} (not implemented) [intent=direct_read availability=not_implemented operation=survicate.get.v2-surveys-survey-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 surveys survey-id responses response-uuid - Documented GET /v2/surveys/{survey_id}/responses/{response_uuid} (not implemented) [intent=direct_read availability=not_implemented operation=survicate.get.v2-surveys-survey-id-responses-response-uuid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - responses list - Run the responses ETL stream [intent=etl availability=implemented stream=responses]
  - survey questions list - Run the survey questions ETL stream [intent=etl availability=implemented stream=survey_questions]
  - surveys list - Run the surveys ETL stream [intent=etl availability=implemented stream=surveys]

## Commands

### Inspect as a manual

```bash
pm connectors inspect survicate
```

### Inspect as structured JSON

```bash
pm connectors inspect survicate --json
```

## Agent Rules

- Run pm connectors inspect survicate before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
