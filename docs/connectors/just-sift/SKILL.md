---
name: pm-just-sift
description: JustSift connector knowledge and safe action guide.
---

# pm-just-sift

## Purpose

Reads JustSift people directory profiles and person field definitions through the Sift REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_token (secret)

## ETL Streams

- peoples:
  - primary key: id
  - fields: companyName(string), connector(string), department(string), directReportCount(number), directoryId(string), displayName(string), email(string), firstName(string), id(string), isTeamLeader(boolean), lastName(string), officeCity(string), officeState(string), phone(string), pictureUrl(string), title(string)
- fields:
  - primary key: id
  - fields: connector(string), displayName(string), filterable(boolean), id(string), objectKey(string), searchable(boolean), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external JustSift API read of people directory profiles and field definitions
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run JustSift's declared streams and reverse-ETL actions.
- Usage: pm just-sift <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api get fields person - Documented GET /fields/person (not implemented) [intent=direct_read availability=not_implemented operation=just-sift.get.fields-person]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get media people idoremail mediakind - Documented GET /media/people/{idOrEmail}/{mediaKind} (not implemented) [intent=direct_read availability=not_implemented operation=just-sift.get.media-people-idoremail-mediakind]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get people idoremail - Documented GET /people/{idOrEmail} (not implemented) [intent=direct_read availability=not_implemented operation=just-sift.get.people-idoremail]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get search people - Documented GET /search/people (not implemented) [intent=direct_read availability=not_implemented operation=just-sift.get.search-people]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post search people - Documented POST /search/people (not implemented) [intent=direct_write availability=not_implemented operation=just-sift.post.search-people]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - fields list - Run the fields ETL stream [intent=etl availability=implemented stream=fields]; notes: discrepancy=present-in-surface-absent-from-artifact
  - peoples list - Run the peoples ETL stream [intent=etl availability=implemented stream=peoples]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect just-sift
```

### Inspect as structured JSON

```bash
pm connectors inspect just-sift --json
```

## Agent Rules

- Run pm connectors inspect just-sift before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
