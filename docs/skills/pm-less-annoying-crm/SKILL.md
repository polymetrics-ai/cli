---
name: pm-less-annoying-crm
description: Less Annoying CRM connector knowledge and safe action guide.
---

# pm-less-annoying-crm

## Purpose

Reads Less Annoying CRM users, contacts, tasks, notes, and events through fixed v2 RPC requests.

## Icon

- id: less-annoying-crm
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

- start_date (required)
- api_key (secret) (required)

## ETL Streams

- users:
  - primary key: UserId
  - fields: FirstName(string), LastName(string), Timezone(string), UserId(string)
- contacts:
  - primary key: ContactId
  - fields: Address(array), AssignedTo(number), Company Name(string), CompanyId(string), ContactId(string), DateCreated(string), Email(array), IsCompany(boolean), Job Title(string), LastUpdate(string), Name(string), Phone(array), Website(string)
- tasks:
  - primary key: TaskId
  - cursor: DateCreated
  - fields: AssignedTo(string), CalendarId(string), ContactId(string), DateCreated(string), Description(string), DueDate(string), IsCompleted(boolean), Name(string), TaskId(string)
- notes:
  - primary key: NoteId
  - fields: ContactId(string), DateCreated(string), DateDisplayedInHistory(string), IsRichText(boolean), Note(string), NoteId(string), UserId(string)
- events:
  - primary key: EventId
  - cursor: DateUpdated
  - fields: ContactIds(array), DateCreated(string), DateUpdated(string), Description(string), EndDate(string), EventId(string), IsAllDay(boolean), IsRecurring(boolean), Location(string), Name(string), StartDate(string), UserIds(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: Bounded POST reads use the fixed Less Annoying CRM v2 origin and declared API-key header authentication.
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
