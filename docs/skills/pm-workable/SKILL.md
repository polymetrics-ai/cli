---
name: pm-workable
description: Workable connector knowledge and safe action guide.
---

# pm-workable

## Purpose

Reads Workable recruiting, account, employee, time tracking, time off, review, subscription, requisition, and offer data; writes Workable candidate, employee, department, member, subscription, time tracking, time off, offer, and requisition mutations.

## Icon

- id: workable
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
- base_url (required)
- candidate_id
- employee_id
- event_id
- job_shortcode
- offer_id
- requisition_code
- review_template_id
- start_date
- timeoff_from_date
- api_key (secret) (required)

## ETL Streams

- jobs:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- candidates:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- members:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- accounts:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- account:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- collaboration_permissions:
  - primary key: _pm_id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- departments:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- disqualification_reasons:
  - primary key: _pm_id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- legal_entities:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- permission_sets:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- recruiters:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- stages:
  - primary key: slug
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- subscriptions:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- employee_fields:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- employees_orgchart:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- employees:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- employee:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- employee_documents:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- review_templates:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- review_template:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- time_entries:
  - primary key: uuid
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- timeoff_balances:
  - primary key: category_id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- timeoff_categories:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- timeoff_requests:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- work_schedules:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- candidate_activities:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- candidate_files:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- candidate_offer:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- custom_attributes:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- events:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- event:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_activities:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_custom_attributes:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_members:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_questions:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_stages:
  - primary key: slug
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_application_form:
  - primary key: _pm_id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- job_recruiters:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- requisitions:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- requisition:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)
- offer:
  - primary key: id
  - fields: _pm_id(string), balances(array), candidate_id(string), categories(array), category_id(string), code(string), created_at(string), email(string), employee_id(string), files(array), from_date(string), id(string), job(object), key(string), name(string), permissions(object), questions(array), shortcode(string), slug(string), starts_at(string), state(string), title(string), type(string), updated_at(string), uuid(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_department:
  - endpoint: POST /departments
  - required fields: name
  - risk: POST /departments mutates Workable data; approval required
- update_department:
  - endpoint: PUT /departments
  - required fields: id
  - risk: PUT /departments mutates Workable data; approval required
- merge_department:
  - endpoint: POST /departments/{{ record.department_id }}/merge
  - required fields: department_id, target_department_id
  - risk: POST /departments/{{ record.department_id }}/merge mutates Workable data; approval required
- delete_department:
  - endpoint: DELETE /departments/{{ record.department_id }}?force={{ record.force }}
  - required fields: department_id, force
  - risk: DELETE /departments/{{ record.department_id }}?force={{ record.force }} mutates Workable data; approval required
- invite_member:
  - endpoint: POST /members/invite
  - required fields: email
  - risk: POST /members/invite mutates Workable data; approval required
- update_member:
  - endpoint: PUT /members
  - required fields: id
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
  - required fields: target, event
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
  - required fields: name
  - risk: POST /review-cycles/templates mutates Workable data; approval required
- bulk_create_time_entries:
  - endpoint: POST /time-tracking/time-entries
  - required fields: time_entries
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
  - required fields: approval_key, state
  - risk: PATCH /timeoff/approvals/{{ record.approval_key }} mutates Workable data; approval required
- create_timeoff_request:
  - endpoint: POST /timeoff/requests
  - required fields: from_date
  - risk: POST /timeoff/requests mutates Workable data; approval required
- update_candidate_custom_attribute:
  - endpoint: PATCH /candidates/{{ record.candidate_id }}/update_custom_attribute_value
  - required fields: candidate_id, custom_attribute_id, value
  - risk: PATCH /candidates/{{ record.candidate_id }}/update_custom_attribute_value mutates Workable data; approval required
- comment_on_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/comments
  - required fields: candidate_id, comment
  - risk: POST /candidates/{{ record.candidate_id }}/comments mutates Workable data; approval required
- copy_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/copy
  - required fields: candidate_id, member_id, target_job_shortcode, target_stage
  - risk: POST /candidates/{{ record.candidate_id }}/copy mutates Workable data; approval required
- disqualify_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/disqualify
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/disqualify mutates Workable data; approval required
- create_job_candidate:
  - endpoint: POST /jobs/{{ record.job_shortcode }}/candidates
  - required fields: job_shortcode, name
  - risk: POST /jobs/{{ record.job_shortcode }}/candidates mutates Workable data; approval required
- move_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/move
  - required fields: candidate_id, target_stage
  - risk: POST /candidates/{{ record.candidate_id }}/move mutates Workable data; approval required
- relocate_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/relocate
  - required fields: candidate_id, target_job_shortcode
  - risk: POST /candidates/{{ record.candidate_id }}/relocate mutates Workable data; approval required
- revert_candidate_disqualification:
  - endpoint: POST /candidates/{{ record.candidate_id }}/revert
  - required fields: candidate_id
  - risk: POST /candidates/{{ record.candidate_id }}/revert mutates Workable data; approval required
- update_candidate_tags:
  - endpoint: PUT /candidates/{{ record.candidate_id }}/tags
  - required fields: candidate_id, tags
  - risk: PUT /candidates/{{ record.candidate_id }}/tags mutates Workable data; approval required
- rate_candidate:
  - endpoint: POST /candidates/{{ record.candidate_id }}/ratings
  - required fields: candidate_id, rating
  - risk: POST /candidates/{{ record.candidate_id }}/ratings mutates Workable data; approval required
- update_candidate_rating:
  - endpoint: PUT /candidates/{{ record.candidate_id }}/ratings
  - required fields: candidate_id, rating
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
  - required fields: stage, name
  - risk: POST /talent_pool/{{ record.stage }}/candidates mutates Workable data; approval required

## Security

- read risk: external Workable SPI v3 reads across recruiting, account, employee, time tracking, time off, review, subscription, requisition, and offer endpoints
- write risk: creates, updates, approves, rejects, archives, deactivates, or deletes Workable recruiting/HR resources according to the selected action
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Workable's declared typed write actions.
- Usage: pm workable <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Reverse ETL writes
- Other Commands
  - approve offer apply - Typed action approve_offer [intent=reverse_etl availability=partial write=approve_offer]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /offers/{offer_id}/approve disagrees with covered api_surface path /spi/v3/offers/{id}/approve.; risk: PATCH /offers/{{ record.offer_id }}/approve mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /offers/{offer_id}/approve disagrees with covered api_surface path /spi/v3/offers/{id}/approve.; flags: --offer-id (required)
  - approve requisition apply - Typed action approve_requisition [intent=reverse_etl availability=partial write=approve_requisition]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /requisitions/{requisition_code}/approve disagrees with covered api_surface path /spi/v3/requisitions/{code}/approve.; risk: PATCH /requisitions/{{ record.requisition_code }}/approve mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /requisitions/{requisition_code}/approve disagrees with covered api_surface path /spi/v3/requisitions/{code}/approve.; flags: --requisition-code (required)
  - archive time entry apply - Typed action archive_time_entry [intent=reverse_etl availability=partial write=archive_time_entry]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /time-tracking/employees/{employee_id}/time-entries/{uuid} disagrees with covered api_surface path /spi/v3/time-tracking/employees/{id}/time-entries/{uuid}.; risk: DELETE /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /time-tracking/employees/{employee_id}/time-entries/{uuid} disagrees with covered api_surface path /spi/v3/time-tracking/employees/{id}/time-entries/{uuid}.; flags: --employee-id (required), --uuid (required)
  - bulk create time entries apply - Typed action bulk_create_time_entries [intent=reverse_etl availability=partial write=bulk_create_time_entries]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /time-tracking/time-entries disagrees with covered api_surface path /spi/v3/time-tracking/time-entries.; risk: POST /time-tracking/time-entries mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /time-tracking/time-entries disagrees with covered api_surface path /spi/v3/time-tracking/time-entries.; flags: --time-entries (required)
  - comment on candidate apply - Typed action comment_on_candidate [intent=reverse_etl availability=partial write=comment_on_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/comments disagrees with covered api_surface path /spi/v3/candidates/{id}/comments.; risk: POST /candidates/{{ record.candidate_id }}/comments mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/comments disagrees with covered api_surface path /spi/v3/candidates/{id}/comments.; flags: --candidate-id (required), --comment (required)
  - copy candidate apply - Typed action copy_candidate [intent=reverse_etl availability=partial write=copy_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/copy disagrees with covered api_surface path /spi/v3/candidates/{id}/copy.; risk: POST /candidates/{{ record.candidate_id }}/copy mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/copy disagrees with covered api_surface path /spi/v3/candidates/{id}/copy.; flags: --candidate-id (required), --member-id (required), --target-job-shortcode (required), --target-stage (required)
  - create department apply - Typed action create_department [intent=reverse_etl availability=partial write=create_department]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /departments disagrees with covered api_surface path /spi/v3/departments.; risk: POST /departments mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /departments disagrees with covered api_surface path /spi/v3/departments.; flags: --name (required)
  - create employee apply - Typed action create_employee [intent=reverse_etl availability=partial write=create_employee]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /employees disagrees with covered api_surface path /spi/v3/employees.; risk: POST /employees mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /employees disagrees with covered api_surface path /spi/v3/employees.
  - create job candidate apply - Typed action create_job_candidate [intent=reverse_etl availability=partial write=create_job_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /jobs/{job_shortcode}/candidates disagrees with covered api_surface path /spi/v3/jobs/{shortcode}/candidates.; risk: POST /jobs/{{ record.job_shortcode }}/candidates mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /jobs/{job_shortcode}/candidates disagrees with covered api_surface path /spi/v3/jobs/{shortcode}/candidates.; flags: --job-shortcode (required), --name (required)
  - create requisition apply - Typed action create_requisition [intent=reverse_etl availability=partial write=create_requisition]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /requisitions disagrees with covered api_surface path /spi/v3/requisitions.; risk: POST /requisitions mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /requisitions disagrees with covered api_surface path /spi/v3/requisitions.
  - create review template apply - Typed action create_review_template [intent=reverse_etl availability=partial write=create_review_template]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /review-cycles/templates disagrees with covered api_surface path /spi/v3/review-cycles/templates.; risk: POST /review-cycles/templates mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /review-cycles/templates disagrees with covered api_surface path /spi/v3/review-cycles/templates.; flags: --name (required)
  - create subscription apply - Typed action create_subscription [intent=reverse_etl availability=partial write=create_subscription]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /subscriptions disagrees with covered api_surface path /spi/v3/subscriptions.; risk: POST /subscriptions mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /subscriptions disagrees with covered api_surface path /spi/v3/subscriptions.; flags: --event (required), --target (required)
  - create talent pool candidate apply - Typed action create_talent_pool_candidate [intent=reverse_etl availability=partial write=create_talent_pool_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /talent_pool/{stage}/candidates disagrees with covered api_surface path /spi/v3/talent_pool/{stage}/candidates.; risk: POST /talent_pool/{{ record.stage }}/candidates mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /talent_pool/{stage}/candidates disagrees with covered api_surface path /spi/v3/talent_pool/{stage}/candidates.; flags: --name (required), --stage (required)
  - create time entry apply - Typed action create_time_entry [intent=reverse_etl availability=partial write=create_time_entry]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /time-tracking/employees/{employee_id}/time-entries disagrees with covered api_surface path /spi/v3/time-tracking/employees/{id}/time-entries.; risk: POST /time-tracking/employees/{{ record.employee_id }}/time-entries mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /time-tracking/employees/{employee_id}/time-entries disagrees with covered api_surface path /spi/v3/time-tracking/employees/{id}/time-entries.; flags: --employee-id (required)
  - create timeoff request apply - Typed action create_timeoff_request [intent=reverse_etl availability=partial write=create_timeoff_request]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /timeoff/requests disagrees with covered api_surface path /spi/v3/timeoff/requests.; risk: POST /timeoff/requests mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /timeoff/requests disagrees with covered api_surface path /spi/v3/timeoff/requests.; flags: --from-date (required)
  - deactivate member apply - Typed action deactivate_member [intent=reverse_etl availability=partial write=deactivate_member]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /members/{member_id} disagrees with covered api_surface path /spi/v3/members/{id}.; risk: DELETE /members/{{ record.member_id }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /members/{member_id} disagrees with covered api_surface path /spi/v3/members/{id}.; flags: --member-id (required)
  - decide timeoff approval apply - Typed action decide_timeoff_approval [intent=reverse_etl availability=partial write=decide_timeoff_approval]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /timeoff/approvals/{approval_key} disagrees with covered api_surface path /spi/v3/timeoff/approvals/{key}.; risk: PATCH /timeoff/approvals/{{ record.approval_key }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /timeoff/approvals/{approval_key} disagrees with covered api_surface path /spi/v3/timeoff/approvals/{key}.; flags: --approval-key (required), --state (required)
  - delete department apply - Typed action delete_department [intent=reverse_etl availability=partial write=delete_department]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /departments/{department_id} disagrees with covered api_surface path /spi/v3/departments/{id}.; risk: DELETE /departments/{{ record.department_id }}?force={{ record.force }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /departments/{department_id} disagrees with covered api_surface path /spi/v3/departments/{id}.; flags: --department-id (required), --force (required)
  - delete subscription apply - Typed action delete_subscription [intent=reverse_etl availability=partial write=delete_subscription]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /subscriptions/{subscription_id} disagrees with covered api_surface path /spi/v3/subscriptions/{id}.; risk: DELETE /subscriptions/{{ record.subscription_id }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /subscriptions/{subscription_id} disagrees with covered api_surface path /spi/v3/subscriptions/{id}.; flags: --subscription-id (required)
  - disqualify candidate apply - Typed action disqualify_candidate [intent=reverse_etl availability=partial write=disqualify_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/disqualify disagrees with covered api_surface path /spi/v3/candidates/{id}/disqualify.; risk: POST /candidates/{{ record.candidate_id }}/disqualify mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/disqualify disagrees with covered api_surface path /spi/v3/candidates/{id}/disqualify.; flags: --candidate-id (required)
  - enable member apply - Typed action enable_member [intent=reverse_etl availability=partial write=enable_member]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /members/{member_id}/enable disagrees with covered api_surface path /spi/v3/members/{id}/enable.; risk: POST /members/{{ record.member_id }}/enable mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /members/{member_id}/enable disagrees with covered api_surface path /spi/v3/members/{id}/enable.; flags: --member-id (required)
  - invite member apply - Typed action invite_member [intent=reverse_etl availability=partial write=invite_member]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /members/invite disagrees with covered api_surface path /spi/v3/members/invite.; risk: POST /members/invite mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /members/invite disagrees with covered api_surface path /spi/v3/members/invite.; flags: --email (required)
  - merge department apply - Typed action merge_department [intent=reverse_etl availability=partial write=merge_department]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /departments/{department_id}/merge disagrees with covered api_surface path /spi/v3/departments/{id}/merge.; risk: POST /departments/{{ record.department_id }}/merge mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /departments/{department_id}/merge disagrees with covered api_surface path /spi/v3/departments/{id}/merge.; flags: --department-id (required), --target-department-id (required)
  - move candidate apply - Typed action move_candidate [intent=reverse_etl availability=partial write=move_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/move disagrees with covered api_surface path /spi/v3/candidates/{id}/move.; risk: POST /candidates/{{ record.candidate_id }}/move mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/move disagrees with covered api_surface path /spi/v3/candidates/{id}/move.; flags: --candidate-id (required), --target-stage (required)
  - rate candidate apply - Typed action rate_candidate [intent=reverse_etl availability=partial write=rate_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/ratings disagrees with covered api_surface path /spi/v3/candidates/{id}/ratings.; risk: POST /candidates/{{ record.candidate_id }}/ratings mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/ratings disagrees with covered api_surface path /spi/v3/candidates/{id}/ratings.; flags: --candidate-id (required), --rating (required)
  - reject offer apply - Typed action reject_offer [intent=reverse_etl availability=partial write=reject_offer]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /offers/{offer_id}/reject disagrees with covered api_surface path /spi/v3/offers/{id}/reject.; risk: PATCH /offers/{{ record.offer_id }}/reject mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /offers/{offer_id}/reject disagrees with covered api_surface path /spi/v3/offers/{id}/reject.; flags: --offer-id (required)
  - reject requisition apply - Typed action reject_requisition [intent=reverse_etl availability=partial write=reject_requisition]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /requisitions/{requisition_code}/reject disagrees with covered api_surface path /spi/v3/requisitions/{code}/reject.; risk: PATCH /requisitions/{{ record.requisition_code }}/reject mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /requisitions/{requisition_code}/reject disagrees with covered api_surface path /spi/v3/requisitions/{code}/reject.; flags: --requisition-code (required)
  - relocate candidate apply - Typed action relocate_candidate [intent=reverse_etl availability=partial write=relocate_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/relocate disagrees with covered api_surface path /spi/v3/candidates/{id}/relocate.; risk: POST /candidates/{{ record.candidate_id }}/relocate mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/relocate disagrees with covered api_surface path /spi/v3/candidates/{id}/relocate.; flags: --candidate-id (required), --target-job-shortcode (required)
  - revert candidate disqualification apply - Typed action revert_candidate_disqualification [intent=reverse_etl availability=partial write=revert_candidate_disqualification]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/revert disagrees with covered api_surface path /spi/v3/candidates/{id}/revert.; risk: POST /candidates/{{ record.candidate_id }}/revert mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/revert disagrees with covered api_surface path /spi/v3/candidates/{id}/revert.; flags: --candidate-id (required)
  - update candidate apply - Typed action update_candidate [intent=reverse_etl availability=partial write=update_candidate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id} disagrees with covered api_surface path /spi/v3/candidates/{id}.; risk: PATCH /candidates/{{ record.candidate_id }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id} disagrees with covered api_surface path /spi/v3/candidates/{id}.; flags: --candidate-id (required)
  - update candidate custom attribute apply - Typed action update_candidate_custom_attribute [intent=reverse_etl availability=partial write=update_candidate_custom_attribute]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/update_custom_attribute_value disagrees with covered api_surface path /spi/v3/candidates/{id}/update_custom_attribute_value.; risk: PATCH /candidates/{{ record.candidate_id }}/update_custom_attribute_value mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/update_custom_attribute_value disagrees with covered api_surface path /spi/v3/candidates/{id}/update_custom_attribute_value.; flags: --candidate-id (required), --custom-attribute-id (required), --value (required)
  - update candidate rating apply - Typed action update_candidate_rating [intent=reverse_etl availability=partial write=update_candidate_rating]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/ratings disagrees with covered api_surface path /spi/v3/candidates/{id}/ratings.; risk: PUT /candidates/{{ record.candidate_id }}/ratings mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/ratings disagrees with covered api_surface path /spi/v3/candidates/{id}/ratings.; flags: --candidate-id (required), --rating (required)
  - update candidate tags apply - Typed action update_candidate_tags [intent=reverse_etl availability=partial write=update_candidate_tags]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /candidates/{candidate_id}/tags disagrees with covered api_surface path /spi/v3/candidates/{id}/tags.; risk: PUT /candidates/{{ record.candidate_id }}/tags mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /candidates/{candidate_id}/tags disagrees with covered api_surface path /spi/v3/candidates/{id}/tags.; flags: --candidate-id (required), --tags (required)
  - update department apply - Typed action update_department [intent=reverse_etl availability=partial write=update_department]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /departments disagrees with covered api_surface path /spi/v3/departments.; risk: PUT /departments mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /departments disagrees with covered api_surface path /spi/v3/departments.; flags: --id (required)
  - update employee apply - Typed action update_employee [intent=reverse_etl availability=partial write=update_employee]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /employees/{employee_id} disagrees with covered api_surface path /spi/v3/employees/{id}.; risk: PATCH /employees/{{ record.employee_id }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /employees/{employee_id} disagrees with covered api_surface path /spi/v3/employees/{id}.; flags: --employee-id (required)
  - update member apply - Typed action update_member [intent=reverse_etl availability=partial write=update_member]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /members disagrees with covered api_surface path /spi/v3/members.; risk: PUT /members mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /members disagrees with covered api_surface path /spi/v3/members.; flags: --id (required)
  - update requisition apply - Typed action update_requisition [intent=reverse_etl availability=partial write=update_requisition]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /requisitions/{requisition_id} disagrees with covered api_surface path /spi/v3/requisitions/{id}.; risk: PATCH /requisitions/{{ record.requisition_id }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /requisitions/{requisition_id} disagrees with covered api_surface path /spi/v3/requisitions/{id}.; flags: --requisition-id (required)
  - update time entry apply - Typed action update_time_entry [intent=reverse_etl availability=partial write=update_time_entry]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /time-tracking/employees/{employee_id}/time-entries/{uuid} disagrees with covered api_surface path /spi/v3/time-tracking/employees/{id}/time-entries/{uuid}.; risk: PATCH /time-tracking/employees/{{ record.employee_id }}/time-entries/{{ record.uuid }} mutates Workable data; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /time-tracking/employees/{employee_id}/time-entries/{uuid} disagrees with covered api_surface path /spi/v3/time-tracking/employees/{id}/time-entries/{uuid}.; flags: --employee-id (required), --uuid (required)

## Sync Transport

- Source transport: declared
- Destination transport: unsupported
- A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
- Source executor: declarative_api/declarative_stream_source

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
