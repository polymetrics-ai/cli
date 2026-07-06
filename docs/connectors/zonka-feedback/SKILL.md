---
name: pm-zonka-feedback
description: Zonka Feedback connector knowledge and safe action guide.
---

# pm-zonka-feedback

## Purpose

Reads Zonka Feedback responses, surveys, contacts, devices, tasks, locations, users, workspaces, stats, and distribution logs; writes responses, contacts, survey sends, and tasks through the Zonka Feedback REST API.

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
- contact_id
- device_id
- device_stats_end_date
- device_stats_start_date
- distribution_logs_channel
- distribution_logs_end_date
- distribution_logs_start_date
- location_id
- response_id
- survey_id
- user_id
- access_token (secret)
- auth_token (secret)

## ETL Streams

- responses:
  - primary key: id
  - fields: id(), name(), rating(), updated_at()
- surveys:
  - primary key: id
  - fields: id(), name(), rating(), updated_at()
- contacts:
  - primary key: id
  - fields: id(), name(), rating(), updated_at()
- response_details:
  - primary key: id
  - fields: channel(), id(), recievedDate(), respondentId(), responseURL(), surveyId(), surveyName()
- workspaces:
  - primary key: id
  - fields: id(), modifiedBy(), modifiedDate(), workspaceName()
- survey_stats:
  - primary key: surveyId
  - fields: CES(), CSAT(), NPS(), averageTime(), completionRate(), responses(), surveyId(), surveyLanguage(), surveyName()
- survey_details:
  - primary key: id
  - fields: id(), isActive(), modifiedBy(), modifiedDate(), surveyDescription(), surveyName(), webSurveyTitle()
- contact_details:
  - primary key: id
  - fields: channel(), email(), externalId(), id(), mobile(), name(), pendingTasks(), totalResponses()
- contact_segments:
  - primary key: id
  - fields: contacts(), id(), name(), type()
- devices:
  - primary key: id
  - fields: appStatus(), description(), deviceCategory(), deviceOS(), friendlyName(), id(), isActive(), lastCommunication(), locationId(), name()
- device_details:
  - primary key: id
  - fields: appStatus(), deviceBrand(), deviceModel(), deviceOS(), friendlyName(), id(), lastCommunication(), name()
- device_uptime:
  - primary key: deviceId
  - fields: deviceId(), totalUptime()
- device_responses:
  - primary key: deviceId
  - fields: deviceId(), responses(), totalResponses()
- tasks:
  - primary key: id
  - fields: assignedTo(), contactId(), description(), dueDateTime(), id(), isCompleted(), name(), reminderSetting(), responseId(), type()
- locations:
  - primary key: id
  - fields: address(), externalId(), id(), isActive(), labels(), name()
- location_details:
  - primary key: id
  - fields: address(), externalId(), id(), isActive(), labels(), name()
- users:
  - primary key: id
  - fields: designation(), email(), id(), isActive(), isOwner(), lastLogin(), locationId(), mobile(), name(), role()
- user_details:
  - primary key: id
  - fields: designation(), email(), id(), isActive(), isOwner(), lastLogin(), mobile(), name(), role()
- distribution_logs:
  - primary key: sentDateTime
  - fields: SurveySubmitted(), channel(), emailOpened(), locationName(), sentDateTime(), status(), surveyId(), surveyName(), surveyOpened(), to()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- add_response:
  - endpoint: POST /responses
  - risk: creates a Zonka Feedback survey response; approval required
- update_response:
  - endpoint: PATCH /responses/{{ record.responseId }}
  - required fields: responseId
  - risk: updates a Zonka Feedback survey response; approval required
- upsert_contacts:
  - endpoint: POST /contacts/upsert
  - risk: creates or updates Zonka Feedback contact records; approval required
- send_email_survey:
  - endpoint: POST /sendemail
  - risk: sends or schedules email survey invitations; approval required
- send_sms_survey:
  - endpoint: POST /sendsms
  - risk: sends or schedules SMS survey invitations; approval required
- send_two_way_sms_survey:
  - endpoint: POST /send2waysms
  - risk: sends or schedules two-way SMS survey invitations; approval required
- send_whatsapp_survey:
  - endpoint: POST /send-wa-message
  - risk: sends or schedules WhatsApp survey invitations; approval required
- add_task:
  - endpoint: POST /tasks/add
  - risk: creates a Zonka Feedback task; approval required
- update_task:
  - endpoint: POST /tasks/{{ record.taskId }}
  - required fields: taskId
  - risk: updates a Zonka Feedback task; approval required
- delete_tasks:
  - endpoint: DELETE /tasks/delete
  - risk: deletes one or more Zonka Feedback tasks; approval required

## Security

- read risk: external Zonka Feedback API read of response, survey, contact, device, task, location, user, workspace, stats, and distribution log data
- write risk: external Zonka Feedback API mutations that add or update responses, upsert contacts, send surveys over email/SMS/WhatsApp, and create, update, or delete tasks
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zonka-feedback
```

### Inspect as structured JSON

```bash
pm connectors inspect zonka-feedback --json
```

## Agent Rules

- Run pm connectors inspect zonka-feedback before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
