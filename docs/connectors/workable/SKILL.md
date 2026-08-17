---
name: pm-workable
description: Workable connector knowledge and safe action guide.
---

# pm-workable

## Purpose

Reads Workable recruiting, account, employee, time tracking, time off, review, subscription, requisition, and offer data; writes Workable candidate, employee, department, member, subscription, time tracking, time off, offer, and requisition mutations.

## Icon

- asset: icons/workable.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://workable.readme.io/reference

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- account_subdomain
- base_url
- candidate_id
- employee_id
- event_id
- job_shortcode
- offer_id
- requisition_code
- review_template_id
- start_date
- timeoff_from_date
- api_key (secret)

## ETL Streams

- jobs:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- candidates:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- members:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- accounts:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- account:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- collaboration_permissions:
  - primary key: _pm_id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- departments:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- disqualification_reasons:
  - primary key: _pm_id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- legal_entities:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- permission_sets:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- recruiters:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- stages:
  - primary key: slug
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- subscriptions:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- employee_fields:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- employees_orgchart:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- employees:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- employee:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- employee_documents:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- review_templates:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- review_template:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- time_entries:
  - primary key: uuid
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- timeoff_balances:
  - primary key: category_id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- timeoff_categories:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- timeoff_requests:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- work_schedules:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- candidate_activities:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- candidate_files:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- candidate_offer:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- custom_attributes:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- events:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- event:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_activities:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_custom_attributes:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_members:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_questions:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_stages:
  - primary key: slug
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_application_form:
  - primary key: _pm_id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- job_recruiters:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- requisitions:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- requisition:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()
- offer:
  - primary key: id
  - fields: _pm_id(), balances(), candidate_id(), categories(), category_id(), code(), created_at(), email(), employee_id(), files(), from_date(), id(), job(), key(), name(), permissions(), questions(), shortcode(), slug(), starts_at(), state(), title(), type(), updated_at(), uuid()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_department:
  - endpoint: POST /departments
  - risk: POST /departments mutates Workable data; approval required
- update_department:
  - endpoint: PUT /departments
  - risk: PUT /departments mutates Workable data; approval required
- merge_department:
  - endpoint: POST /departments/{{ record.department_id }}/merge
  - required fields: department_id
  - risk: POST /departments/{{ record.department_id }}/merge mutates Workable data; approval required
- delete_department:
  - endpoint: DELETE /departments/{{ record.department_id }}?force={{ record.force }}
  - required fields: department_id, force
  - risk: DELETE /departments/{{ record.department_id }}?force={{ record.force }} mutates Workable data; approval required
- invite_member:
  - endpoint: POST /members/invite
  - risk: POST /members/invite mutates Workable data; approval required
- update_member:
  - endpoint: PUT /members
  - risk: PUT /members mutates Workable data; approval required
- deactivate_member:
  - endpoint: DELETE /members/{{ record.member_id }}
  - required fields: member_id
  - risk: DELETE /members/{{ record.member_id }} mutates Workable data; approval required
- enable_member:
  - endpoint: POST /members/{{ record.member_id }}/enable
  - required fields: member_id
  - risk: POST /members/{{ record.member_id }}/enable mutates Workable data; approval required
- create_subscription:
  - endpoint: POST /subscriptions
  - risk: POST /subscriptions mutates Workable data; approval required
- delete_subscription:
  - endpoint: DELETE /subscriptions/{{ record.subscription_id }}
  - required fields: subscription_id
  - risk: DELETE /subscriptions/{{ record.subscription_id }} mutates Workable data; approval required
- create_employee:
  - endpoint: POST /employees
  - risk: POST /employees mutates Workable data; approval required
- update_employee:
  - endpoint: PATCH /employees/{{ record.employee_id }}
  - required fields: employee_id
  - risk: PATCH /employees/{{ record.employee_id }} mutates Workable data; approval required
- create_review_template:
  - endpoint: POST /review-cycles/templates
  - risk: POST /review-cycles/templates mutates Workable data; approval required
- bulk_create_time_entries:
  - endpoint: POST /time-tracking/time-entries
  - risk: POST /time-tracking/time-entries mutates Workable data; approval required
- create_time_entry:
  - endpoint: POST /time-tracking/employees/{{ record.employee_id }}/time-entries
  - required fields: employee_id
  - risk: POST /time-tracking/employees/{{ record.employee_id }}/time-entries mutates Workable data; approval required
- update_time_entry:
  - endpoint: PATCH /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }}
  - required fields: employee_id, uuid
  - risk: PATCH /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }} mutates Workable data; approval required
- archive_time_entry:
  - endpoint: DELETE /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }}
  - required fields: employee_id, uuid
  - risk: DELETE /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }} mutates Workable data; approval required
- decide_timeoff_approval:
  - endpoint: PATCH /timeoff/approvals/{{ record.approval_key }}
  - required fields: approval_key
  - risk: PATCH /timeoff/approvals/{{ record.approval_key }} mutates Workable data; approval required
- create_timeoff_request:
  - endpoint: POST /timeoff/requests
  - risk: POST /timeoff/requests mutates Workable data; approval required
- update_candidate_custom_attribute:
  - endpoint: PATCH /candidates/{{ record.candidate_id }}/update_custom_attribute_value
  - required fields: candidate_id
  - risk: PATCH /candidates/{{ record.candidate_id }}/update_custom_attribute_value mutates Workable data; approval required
- comment_on_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/comments
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/comments mutates Workable data; approval required
- copy_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/copy
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/copy mutates Workable data; approval required
- disqualify_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/disqualify
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/disqualify mutates Workable data; approval required
- create_job_candidate:
  - endpoint: POST /jobs/{{ record.job_shortcode }}/candidates
  - required fields: job_shortcode
  - risk: POST /jobs/{{ record.job_shortcode }}/candidates mutates Workable data; approval required
- move_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/move
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/move mutates Workable data; approval required
- relocate_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/relocate
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/relocate mutates Workable data; approval required
- revert_candidate_disqualification:
  - endpoint: POST /candidates/{{ record.candidate_id }}/revert
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/revert mutates Workable data; approval required
- update_candidate_tags:
  - endpoint: PUT /candidates/{{ record.candidate_id }}/tags
  - required fields: candidate_id
  - risk: PUT /candidates/{{ record.candidate_id }}/tags mutates Workable data; approval required
- rate_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/ratings
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/ratings mutates Workable data; approval required
- update_candidate_rating:
  - endpoint: PUT /candidates/{{ record.candidate_id }}/ratings
  - required fields: candidate_id
  - risk: PUT /candidates/{{ record.candidate_id }}/ratings mutates Workable data; approval required
- update_candidate:
  - endpoint: PATCH /candidates/{{ record.candidate_id }}
  - required fields: candidate_id
  - risk: PATCH /candidates/{{ record.candidate_id }} mutates Workable data; approval required
- approve_offer:
  - endpoint: PATCH /offers/{{ record.offer_id }}/approve
  - required fields: offer_id
  - risk: PATCH /offers/{{ record.offer_id }}/approve mutates Workable data; approval required
- reject_offer:
  - endpoint: PATCH /offers/{{ record.offer_id }}/reject
  - required fields: offer_id
  - risk: PATCH /offers/{{ record.offer_id }}/reject mutates Workable data; approval required
- create_requisition:
  - endpoint: POST /requisitions
  - risk: POST /requisitions mutates Workable data; approval required
- update_requisition:
  - endpoint: PATCH /requisitions/{{ record.requisition_id }}
  - required fields: requisition_id
  - risk: PATCH /requisitions/{{ record.requisition_id }} mutates Workable data; approval required
- approve_requisition:
  - endpoint: PATCH /requisitions/{{ record.requisition_code }}/approve
  - required fields: requisition_code
  - risk: PATCH /requisitions/{{ record.requisition_code }}/approve mutates Workable data; approval required
- reject_requisition:
  - endpoint: PATCH /requisitions/{{ record.requisition_code }}/reject
  - required fields: requisition_code
  - risk: PATCH /requisitions/{{ record.requisition_code }}/reject mutates Workable data; approval required
- create_talent_pool_candidate:
  - endpoint: POST /talent_pool/{{ record.stage }}/candidates
  - required fields: stage
  - risk: POST /talent_pool/{{ record.stage }}/candidates mutates Workable data; approval required

## Security

- read risk: external Workable SPI v3 reads across recruiting, account, employee, time tracking, time off, review, subscription, requisition, and offer endpoints
- write risk: creates, updates, approves, rejects, archives, deactivates, or deletes Workable recruiting/HR resources according to the selected action
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect workable
```

### Inspect as structured JSON

```bash
pm connectors inspect workable --json
```

## Agent Rules

- Run pm connectors inspect workable before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
