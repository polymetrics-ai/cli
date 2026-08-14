---
name: pm-everhour
description: Everhour connector knowledge and safe action guide.
---

# pm-everhour

## Purpose

Reads Everhour projects, clients, team members, team time records, per-project tasks and sections, time-off types, time-off allocations, expenses, expense categories, and invoices, and writes client/project/task/section/time-record/expense mutations, through the Everhour REST API.

## Icon

- id: everhour
- asset: icons/everhour.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://everhour.docs.apiary.io/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret) (required)

## ETL Streams

- projects:
  - primary key: id
  - fields: createdAt(string), favorite(boolean), foreign(boolean), id(string), name(string), platform(string), status(string), type(string), workspaceId(string), workspaceName(string)
- clients:
  - primary key: id
  - fields: createdAt(string), email(string), favorite(boolean), id(string), name(string), status(string)
- users:
  - primary key: id
  - fields: capacity(integer), createdAt(string), email(string), headline(string), id(string), isEmailVerified(boolean), name(string), role(string), status(string), type(string)
- time:
  - primary key: id
  - fields: createdAt(string), date(string), id(string), time(integer), user(integer)
- tasks:
  - primary key: id
  - fields: completed(boolean), createdAt(string), id(string), name(string), project_id(string), status(string), type(string), url(string)
- sections:
  - primary key: id
  - fields: id(integer), name(string), position(integer), project_id(string), status(string)
- time_off_types:
  - primary key: id
  - fields: color(string), description(string), id(integer), name(string), paid(boolean)
- allocations:
  - primary key: id
  - fields: accrualFrequency(string), completed(boolean), days(number), endDate(string), id(integer), notes(string), restrictOverAllocation(boolean), startDate(string), timeOffType(integer)
- expense_categories:
  - primary key: id
  - fields: color(string), id(integer), name(string), unitBased(boolean), unitName(string), unitPrice(number)
- expenses:
  - primary key: id
  - fields: amount(number), billable(boolean), category(integer), date(string), details(string), id(integer), project(string), quantity(number), user(integer)
- invoices:
  - primary key: id
  - fields: clientId(string), createdAt(string), discount(number), dueDate(string), id(integer), issueDate(string), publicId(string), reference(string), status(string), tax(number)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_client:
  - endpoint: POST /clients
  - required fields: name
  - risk: creates a new client record; low-risk external mutation, no approval required
- update_client:
  - endpoint: PUT /clients/{{ record.id }}
  - required fields: id
  - risk: renames or otherwise mutates an existing client's metadata; low-risk external mutation
- delete_client:
  - endpoint: DELETE /clients/{{ record.id }}
  - required fields: id
  - risk: permanently removes a client and its association with any linked projects; irreversible, approval required
- create_project:
  - endpoint: POST /projects
  - required fields: name, type
  - risk: creates a new project; low-risk external mutation, no approval required
- update_project:
  - endpoint: PUT /projects/{{ record.id }}
  - required fields: id
  - risk: renames or reconfigures an existing project; low-risk external mutation
- archive_project:
  - endpoint: PATCH /projects/{{ record.id }}/archive
  - required fields: id, archived
  - risk: archives or unarchives a project, hiding it from active project lists and blocking new time entries against it while archived; approval required for archiving a project still in active use
- delete_project:
  - endpoint: DELETE /projects/{{ record.id }}
  - required fields: id
  - risk: permanently removes a project and its tasks/sections/time associations; irreversible, approval required
- create_task:
  - endpoint: POST /projects/{{ record.project_id }}/tasks
  - required fields: project_id, name, section
  - risk: creates a new task under an existing project section; low-risk external mutation, no approval required
- update_task:
  - endpoint: PUT /tasks/{{ record.id }}
  - required fields: id
  - risk: renames or reconfigures an existing task; low-risk external mutation
- delete_task:
  - endpoint: DELETE /tasks/{{ record.id }}
  - required fields: id
  - risk: permanently removes a task and its logged time association; irreversible, approval required
- create_section:
  - endpoint: POST /projects/{{ record.project_id }}/sections
  - required fields: project_id, name
  - risk: creates a new task section within a project; low-risk external mutation, no approval required
- delete_section:
  - endpoint: DELETE /sections/{{ record.id }}
  - required fields: id
  - risk: permanently removes a task section; any tasks in it become unsectioned, approval required
- create_time_record:
  - endpoint: POST /time
  - required fields: time, date
  - risk: logs a new time entry against a task, which can feed directly into client billing/invoicing; low-risk external mutation, no approval required
- update_time_record:
  - endpoint: PUT /time/{{ record.id }}
  - required fields: id
  - risk: changes the logged duration/date/comment of an existing time entry that may already be invoiced; approval required if the entry is locked or billed
- delete_time_record:
  - endpoint: DELETE /time/{{ record.id }}
  - required fields: id
  - risk: permanently removes a logged time entry, which can affect billing/invoicing history; irreversible, approval required
- create_expense:
  - endpoint: POST /expenses
  - required fields: category, date
  - risk: logs a new billable/non-billable expense, which can feed directly into client invoicing; low-risk external mutation, no approval required
- delete_expense:
  - endpoint: DELETE /expenses/{{ record.id }}
  - required fields: id
  - risk: permanently removes a logged expense, which can affect billing/invoicing history; irreversible, approval required

## Security

- read risk: external Everhour API read of project, client, member, time, task, section, time-off, expense, and invoice data
- write risk: external mutation of clients, projects, tasks, sections, time records, and expenses; deletes are irreversible and time-record/expense mutations can affect client billing/invoicing, every write ships with an explicit per-action risk string
- approval: required for every delete_* action, archive_project, and update_time_record (time entries may already be invoiced/locked); create/update of clients, projects, tasks, sections, time records, and expenses are low-risk
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect everhour
```

### Inspect as structured JSON

```bash
pm connectors inspect everhour --json
```

## Agent Rules

- Run pm connectors inspect everhour before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
