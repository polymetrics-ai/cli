---
name: pm-simplesat
description: Simplesat connector knowledge and safe action guide.
---

# pm-simplesat

## Purpose

Reads and writes Simplesat surveys, answers, questions, customers, and responses (including nested ticket data) through the Simplesat v1 API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- created_after
- page_size
- api_key (secret) (required)

## ETL Streams

- answers:
  - primary key: id
  - fields: choice(string), choice_label(string), choices(array), comment(string), created(string), follow_up_answer(string), follow_up_answer_choice(string), follow_up_answer_choices(array), id(integer), modified(string), published_as_testimonial(boolean), question(object), sentiment(string), survey(object)
- surveys:
  - primary key: id
  - fields: brand_name(string), id(integer), metric(string), name(string), survey_token(string), survey_type(string)
- questions:
  - primary key: id
  - fields: choices(array), id(integer), metric(string), order(integer), rating_scale(boolean), required(boolean), rules(array), survey(object), text(string)
- customers:
  - primary key: id
  - fields: company(string), created(string), custom_attributes(object), email(string), external_id(string), id(integer), language(string), modified(string), name(string), subscribed(boolean), tags(array)
- responses:
  - primary key: id
  - fields: answers(array), created(string), customer(object), id(integer), ip_address(string), language(string), modified(string), source(string), survey(object), tags(array), team_members(array), ticket(object)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_or_update_customer:
  - endpoint: POST /customers
  - risk: creates a new customer or updates the existing one matched by external_id/email; low-risk external mutation, no approval required
- update_customer:
  - endpoint: PUT /customers/{{ record.id }}
  - required fields: id
  - risk: mutates an existing customer's profile fields by id; overwrites tags/custom_attributes wholesale with the submitted value
- create_or_update_team_member:
  - endpoint: POST /team-members
  - risk: creates a new team member or updates the existing one matched by external_id/email; low-risk external mutation, no approval required
- update_answer:
  - endpoint: PUT /answers/{{ record.id }}
  - required fields: id
  - risk: mutates an existing survey answer's recorded choice/comment/follow-up fields; changes the customer-submitted response data an already-collected survey answer represents
- create_or_update_response:
  - endpoint: POST /responses/create-or-update
  - required fields: survey_id
  - risk: creates a new survey response (or updates one matched by the API's own dedup rule) including its nested answers/customer/ticket/team_members sub-objects; commonly used to import or backfill historical survey data with an explicit created timestamp
- update_response:
  - endpoint: PUT /responses/{{ record.id }}/update
  - required fields: id, survey_id
  - risk: mutates an existing survey response's tags/answers/team_members by id; overwrites the identified response's recorded data
- send_survey_email:
  - endpoint: POST /surveys/{{ record.survey_token }}/email
  - required fields: survey_token, customer
  - risk: sends a live survey invitation email to the named customer's real inbox; each call generates one outbound email delivery, not a reversible data mutation

## Security

- read risk: reads survey/answer/customer/response data (including nested ticket/team-member sub-objects) from a connected Simplesat account
- write risk: creates/updates customers and team members, updates individual answers and survey responses, and can trigger a live survey-invitation email to a real customer inbox (send_survey_email)
- approval: none for customer/team-member/answer/response upserts (low-risk CRM-style data); send_survey_email sends a real outbound email and should be reviewed before enabling in a caller with untrusted input
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect simplesat
```

### Inspect as structured JSON

```bash
pm connectors inspect simplesat --json
```

## Agent Rules

- Run pm connectors inspect simplesat before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
