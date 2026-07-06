# pm connectors inspect less-annoying-crm

```text
NAME
  pm connectors inspect less-annoying-crm - Less Annoying CRM connector manual

SYNOPSIS
  pm connectors inspect less-annoying-crm
  pm connectors inspect less-annoying-crm --json
  pm credentials add <name> --connector less-annoying-crm [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Less Annoying CRM users, contacts, tasks, notes, and events through the Less Annoying CRM v2 API. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/less-annoying-crm.svg
  source: official
  review_status: official_verified
  review_url: https://www.lessannoyingcrm.com/help/topic/API

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  start_date
  api_key (secret)

ETL STREAMS
  users:
    primary key: UserId
    fields: FirstName(), LastName(), Timezone(), UserId()
  contacts:
    primary key: ContactId
    fields: Address(), AssignedTo(), Company Name(), CompanyId(), ContactId(), DateCreated(), Email(), IsCompany(), Job Title(), LastUpdate(), Name(), Phone(), Website()
  tasks:
    primary key: TaskId
    cursor: DateCreated
    fields: AssignedTo(), CalendarId(), ContactId(), DateCreated(), Description(), DueDate(), IsCompleted(), Name(), TaskId()
  notes:
    primary key: NoteId
    fields: ContactId(), DateCreated(), DateDisplayedInHistory(), IsRichText(), Note(), NoteId(), UserId()
  events:
    primary key: EventId
    cursor: DateUpdated
    fields: ContactIds(), DateCreated(), DateUpdated(), Description(), EndDate(), EventId(), IsAllDay(), IsRecurring(), Location(), Name(), StartDate(), UserIds()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Less Annoying CRM API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect less-annoying-crm

  # Inspect as structured JSON
  pm connectors inspect less-annoying-crm --json

AGENT WORKFLOW
  - Run pm connectors inspect less-annoying-crm before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
