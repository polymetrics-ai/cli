---
name: pm-aws-cloudtrail
description: AWS CloudTrail connector knowledge and safe action guide.
---

# pm-aws-cloudtrail

## Purpose

Reads CloudTrail management events through fixed, declaration-bound LookupEvents requests signed with AWS Signature Version 4.

## Icon

- id: aws-cloudtrail
- asset: icons/aws-cloudtrail.svg
- source: upstream_registry
- review_status: upstream_seeded

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- aws_key_id (secret) (required)
- aws_secret_key (secret) (required)
- aws_session_token (secret)

## ETL Streams

- management_events:
  - primary key: EventId
  - cursor: EventTime
  - fields: AccessKeyId(string), CloudTrailEvent(string), EventId(string), EventName(string), EventSource(string), EventTime(integer), ReadOnly(string), Resources(array), Username(string)
- read_only_events:
  - primary key: EventId
  - cursor: EventTime
  - fields: AccessKeyId(string), CloudTrailEvent(string), EventId(string), EventName(string), EventSource(string), EventTime(integer), ReadOnly(string), Resources(array), Username(string)
- write_only_events:
  - primary key: EventId
  - cursor: EventTime
  - fields: AccessKeyId(string), CloudTrailEvent(string), EventId(string), EventName(string), EventSource(string), EventTime(integer), ReadOnly(string), Resources(array), Username(string)
- console_logins:
  - primary key: EventId
  - cursor: EventTime
  - fields: AccessKeyId(string), CloudTrailEvent(string), EventId(string), EventName(string), EventSource(string), EventTime(integer), ReadOnly(string), Resources(array), Username(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded CloudTrail LookupEvents requests are signed by the declaration-bound generic SigV4 authenticator against the fixed CloudTrail us-east-1 origin.
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect aws-cloudtrail
```

### Inspect as structured JSON

```bash
pm connectors inspect aws-cloudtrail --json
```

## Agent Rules

- Run pm connectors inspect aws-cloudtrail before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
