---
name: pm-less-annoying-crm
description: Less Annoying CRM connector knowledge and safe action guide.
---

# pm-less-annoying-crm

## Purpose

Reads Less Annoying CRM users, contacts, tasks, notes, and events through the Less Annoying CRM v2 API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

## Icon

- asset: icons/less-annoying-crm.svg
- source: official
- review_status: official_verified
- review_url: https://www.lessannoyingcrm.com/help/topic/API

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

- users:
  - primary key: UserId
  - fields: FirstName(), LastName(), Timezone(), UserId()
- contacts:
  - primary key: ContactId
  - fields: Address(), AssignedTo(), Company Name(), CompanyId(), ContactId(), DateCreated(), Email(), IsCompany(), Job Title(), LastUpdate(), Name(), Phone(), Website()
- tasks:
  - primary key: TaskId
  - cursor: DateCreated
  - fields: AssignedTo(), CalendarId(), ContactId(), DateCreated(), Description(), DueDate(), IsCompleted(), Name(), TaskId()
- notes:
  - primary key: NoteId
  - fields: ContactId(), DateCreated(), DateDisplayedInHistory(), IsRichText(), Note(), NoteId(), UserId()
- events:
  - primary key: EventId
  - cursor: DateUpdated
  - fields: ContactIds(), DateCreated(), DateUpdated(), Description(), EndDate(), EventId(), IsAllDay(), IsRecurring(), Location(), Name(), StartDate(), UserIds()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Less Annoying CRM API reads performed by the legacy connector via a Tier-2 hook
- write risk: unsupported
- approval: none; read-only
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect less-annoying-crm
```

### Inspect as structured JSON

```bash
pm connectors inspect less-annoying-crm --json
```

## Agent Rules

- Run pm connectors inspect less-annoying-crm before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
