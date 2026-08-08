---
name: pm-greenhouse
description: Greenhouse connector knowledge and safe action guide.
---

# pm-greenhouse

## Purpose

Reads and writes documented Greenhouse Harvest REST API resources through the declarative connector engine.

## Icon

- id: greenhouse
- asset: icons/greenhouse.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.greenhouse.io/harvest.html

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- application_id
- approval_flow_id
- base_url
- candidate_id
- custom_field_id
- demographic_answer_id
- demographic_answer_option_id
- demographic_question_id
- demographic_question_set_id
- department_id
- email_template_id
- field_type
- job_id
- job_post_id
- job_stage_id
- mode
- offer_id
- office_id
- on_behalf_of_user_id
- opening_id
- page_size
- prospect_pool_id
- scheduled_interview_id
- scorecard_id
- token
- user_id
- api_key (secret)

## ETL Streams

- candidates:
  - primary key: id
  - cursor: updated_at
  - fields: company(string), created_at(string), first_name(string), id(integer), is_private(boolean), last_activity(string), last_name(string), title(string), updated_at(string)
- applications:
  - primary key: id
  - cursor: last_activity_at
  - fields: applied_at(string), candidate_id(integer), id(integer), last_activity_at(string), rejected_at(string), source_id(integer), status(string)
- jobs:
  - primary key: id
  - cursor: updated_at
  - fields: closed_at(string), confidential(boolean), created_at(string), id(integer), name(string), opened_at(string), requisition_id(string), status(string), updated_at(string)
- offers:
  - primary key: id
  - cursor: updated_at
  - fields: application_id(integer), candidate_id(integer), created_at(string), id(integer), sent_at(string), starts_at(string), status(string), updated_at(string), version(integer)
- users:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(string), disabled(boolean), employee_id(string), first_name(string), id(integer), last_name(string), name(string), primary_email_address(string), site_admin(boolean), updated_at(string)
- activity_feed:
  - fields: id(integer)
- application:
  - fields: id(integer)
- approvals_for_job:
  - fields: id(integer)
- approval_flow:
  - fields: id(integer)
- pending_approvals_for_user:
  - fields: id(integer)
- candidate:
  - fields: id(integer)
- close_reasons:
  - fields: id(integer)
- custom_fields:
  - fields: id(integer)
- custom_field:
  - fields: id(integer)
- custom_field_options:
  - fields: id(integer)
- demographic_question_sets:
  - fields: id(integer)
- demographic_question_set:
  - fields: id(integer)
- demographic_questions:
  - fields: id(integer)
- demographic_questions_for_demographic_question_set:
  - fields: id(integer)
- demographic_question:
  - fields: id(integer)
- demographic_answer_options:
  - fields: id(integer)
- demographic_answer_options_for_demographic_question:
  - fields: id(integer)
- demographic_answer_option:
  - fields: id(integer)
- demographic_answers:
  - fields: id(integer)
- demographic_answers_for_application:
  - fields: id(integer)
- demographic_answer:
  - fields: id(integer)
- departments:
  - fields: id(integer)
- department:
  - fields: id(integer)
- degrees:
  - fields: id(integer)
- disciplines:
  - fields: id(integer)
- schools:
  - fields: id(integer)
- eeoc:
  - fields: id(integer)
- eeoc_data_for_application:
  - fields: id(integer)
- email_templates:
  - fields: id(integer)
- email_template:
  - fields: id(integer)
- job_openings:
  - fields: id(integer)
- opening_for_job:
  - fields: id(integer)
- job_posts:
  - fields: id(integer)
- job_post:
  - fields: id(integer)
- job_posts_for_job:
  - fields: id(integer)
- job_post_for_job:
  - fields: id(integer)
- custom_locations_for_job_post:
  - fields: id(integer)
- job_stages:
  - fields: id(integer)
- job_stages_for_job:
  - fields: id(integer)
- job_stage:
  - fields: id(integer)
- job:
  - fields: id(integer)
- hiring_team:
  - fields: id(integer)
- offers_for_application:
  - fields: id(integer)
- current_offer_for_application:
  - fields: id(integer)
- offer:
  - fields: id(integer)
- offices:
  - fields: id(integer)
- office:
  - fields: id(integer)
- prospect_pools:
  - fields: id(integer)
- prospect_pool:
  - fields: id(integer)
- rejection_reasons:
  - fields: id(integer)
- scheduled_interviews:
  - fields: id(integer)
- scheduled_interviews_for_application:
  - fields: id(integer)
- scheduled_interview:
  - fields: id(integer)
- scorecards:
  - fields: id(integer)
- scorecards_for_application:
  - fields: id(integer)
- scorecard:
  - fields: id(integer)
- sources:
  - fields: id(integer)
- candidate_tags:
  - fields: id(integer)
- tags_applied_to_candidate:
  - fields: id(integer)
- tracking_link_data_for_token:
  - fields: id(integer)
- user:
  - fields: id(integer)
- job_permissions:
  - fields: id(integer)
- future_job_permissions:
  - fields: id(integer)
- user_roles:
  - fields: id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- delete_application:
  - endpoint: DELETE /applications/{{ record.application_id }}
  - required fields: application_id
  - risk: Destructive Greenhouse mutation: DELETE: Delete Application.
- add_application_to_candidate_prospect:
  - endpoint: POST /candidates/{{ record.candidate_id }}/applications
  - required fields: candidate_id
  - risk: Greenhouse mutation: POST: Add Application to Candidate/Prospect.
- update_application:
  - endpoint: PATCH /applications/{{ record.application_id }}
  - required fields: application_id
  - risk: Greenhouse mutation: PATCH: Update Application.
- advance_application:
  - endpoint: POST /applications/{{ record.application_id }}/advance
  - required fields: application_id
  - risk: Greenhouse mutation: POST: Advance Application.
- move_application_different_job:
  - endpoint: POST /applications/{{ record.application_id }}/transfer_to_job
  - required fields: application_id
  - risk: Greenhouse mutation: POST: Move Application (Different Job).
- move_application_same_job:
  - endpoint: POST /applications/{{ record.application_id }}/move
  - required fields: application_id
  - risk: Greenhouse mutation: POST: Move Application (Same Job).
- convert_prospect_to_candidate:
  - endpoint: PATCH /applications/{{ record.application_id }}/convert_prospect
  - required fields: application_id
  - risk: Greenhouse mutation: PATCH: Convert Prospect To Candidate.
- add_attachment_to_application:
  - endpoint: POST /applications/{{ record.application_id }}/attachments
  - required fields: application_id
  - risk: Greenhouse mutation: POST: Add Attachment to Application.
- hire_application:
  - endpoint: POST /applications/{{ record.application_id }}/hire
  - required fields: application_id
  - risk: Greenhouse mutation: POST: Hire Application.
- reject_application:
  - endpoint: POST /applications/{{ record.application_id }}/reject
  - required fields: application_id
  - risk: Destructive Greenhouse mutation: POST: Reject Application.
- update_rejection_reason:
  - endpoint: PATCH /applications/{{ record.application_id }}/reject
  - required fields: application_id
  - risk: Greenhouse mutation: PATCH: Update Rejection Reason.
- unreject_application:
  - endpoint: POST /applications/{{ record.application_id }}/unreject
  - required fields: application_id
  - risk: Greenhouse mutation: POST: Unreject Application.
- request_approvals:
  - endpoint: POST /approval_flows/{{ record.approval_flow_id }}/request_approvals
  - required fields: approval_flow_id
  - risk: Greenhouse mutation: POST: Request Approvals.
- replace_an_approver_in_an_approver_group:
  - endpoint: PUT /approver_groups/{{ record.approver_group_id }}/replace_approvers
  - required fields: approver_group_id
  - risk: Greenhouse mutation: PUT: Replace an approver in an approver group.
- create_or_replace_an_approval_flow:
  - endpoint: PUT /jobs/{{ record.job_id }}/approval_flows
  - required fields: job_id
  - risk: Greenhouse mutation: PUT: Create or replace an approval flow.
- delete_candidate:
  - endpoint: DELETE /candidates/{{ record.candidate_id }}
  - required fields: candidate_id
  - risk: Destructive Greenhouse mutation: DELETE: Delete Candidate.
- edit_candidate:
  - endpoint: PATCH /candidates/{{ record.candidate_id }}
  - required fields: candidate_id
  - risk: Greenhouse mutation: PATCH: Edit Candidate.
- add_attachment:
  - endpoint: POST /candidates/{{ record.candidate_id }}/attachments
  - required fields: candidate_id
  - risk: Greenhouse mutation: POST: Add Attachment.
- add_candidate:
  - endpoint: POST /candidates
  - risk: Greenhouse mutation: POST: Add Candidate.
- add_note:
  - endpoint: POST /candidates/{{ record.candidate_id }}/activity_feed/notes
  - required fields: candidate_id
  - risk: Greenhouse mutation: POST: Add Note.
- add_e_mail_note:
  - endpoint: POST /candidates/{{ record.candidate_id }}/activity_feed/emails
  - required fields: candidate_id
  - risk: Greenhouse mutation: POST: Add E-mail Note.
- add_education:
  - endpoint: POST /candidates/{{ record.candidate_id }}/educations
  - required fields: candidate_id
  - risk: Greenhouse mutation: POST: Add Education.
- remove_education_from_candidate:
  - endpoint: DELETE /candidates/{{ record.candidate_id }}/educations/{{ record.education_id }}
  - required fields: candidate_id, education_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove Education From Candidate.
- add_employment:
  - endpoint: POST /candidates/{{ record.candidate_id }}/employments
  - required fields: candidate_id
  - risk: Greenhouse mutation: POST: Add Employment.
- remove_employment_from_candidate:
  - endpoint: DELETE /candidates/{{ record.candidate_id }}/employments/{{ record.employment_id }}
  - required fields: candidate_id, employment_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove Employment From Candidate.
- add_prospect:
  - endpoint: POST /prospects
  - risk: Greenhouse mutation: POST: Add Prospect.
- anonymize_candidate:
  - endpoint: PUT /candidates/{{ record.candidate_id }}/anonymize?fields={{ record.field_names }}
  - required fields: candidate_id, field_names
  - risk: Destructive Greenhouse mutation: PUT: Anonymize Candidate.
- merge_candidates:
  - endpoint: PUT /candidates/merge
  - risk: Destructive Greenhouse mutation: PUT: Merge Candidates.
- create_custom_field:
  - endpoint: POST /custom_fields
  - risk: Greenhouse mutation: POST: Create Custom Field.
- update_custom_field:
  - endpoint: PATCH /custom_fields/{{ record.custom_field_id }}
  - required fields: custom_field_id
  - risk: Greenhouse mutation: PATCH: Update Custom Field.
- delete_custom_field:
  - endpoint: DELETE /custom_fields/{{ record.custom_field_id }}
  - required fields: custom_field_id
  - risk: Destructive Greenhouse mutation: DELETE: Delete Custom Field.
- create_custom_field_options:
  - endpoint: POST /custom_field/{{ record.custom_field_id }}/custom_field_options
  - required fields: custom_field_id
  - risk: Greenhouse mutation: POST: Create Custom Field Options.
- update_custom_field_options:
  - endpoint: PATCH /custom_field/{{ record.custom_field_id }}/custom_field_options
  - required fields: custom_field_id
  - risk: Greenhouse mutation: PATCH: Update Custom Field Options.
- remove_custom_field_options:
  - endpoint: DELETE /custom_field/{{ record.custom_field_id }}/custom_field_options
  - required fields: custom_field_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove Custom Field Options.
- edit_department:
  - endpoint: PATCH /departments/{{ record.department_id }}
  - required fields: department_id
  - risk: Greenhouse mutation: PATCH: Edit Department.
- add_department:
  - endpoint: POST /departments
  - risk: Greenhouse mutation: POST: Add Department.
- edit_openings:
  - endpoint: PATCH /jobs/{{ record.job_id }}/openings/{{ record.opening_id }}
  - required fields: job_id, opening_id
  - risk: Greenhouse mutation: PATCH: Edit Openings.
- create_new_openings:
  - endpoint: POST /jobs/{{ record.job_id }}/openings
  - required fields: job_id
  - risk: Greenhouse mutation: POST: Create New Openings.
- update_job:
  - endpoint: PATCH /jobs/{{ record.job_id }}
  - required fields: job_id
  - risk: Greenhouse mutation: PATCH: Update Job.
- create_job:
  - endpoint: POST /jobs
  - risk: Greenhouse mutation: POST: Create Job.
- replace_hiring_team:
  - endpoint: PUT /jobs/{{ record.job_id }}/hiring_team
  - required fields: job_id
  - risk: Greenhouse mutation: PUT: Replace Hiring Team.
- add_hiring_team_members:
  - endpoint: POST /jobs/{{ record.job_id }}/hiring_team
  - required fields: job_id
  - risk: Greenhouse mutation: POST: Add Hiring Team Members.
- remove_hiring_team_member:
  - endpoint: DELETE /jobs/{{ record.job_id }}/hiring_team
  - required fields: job_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove Hiring Team Member.
- update_current_offer:
  - endpoint: PATCH /applications/{{ record.application_id }}/offers/current_offer
  - required fields: application_id
  - risk: Greenhouse mutation: PATCH: Update Current Offer.
- edit_office:
  - endpoint: PATCH /offices/{{ record.office_id }}
  - required fields: office_id
  - risk: Greenhouse mutation: PATCH: Edit Office.
- add_office:
  - endpoint: POST /offices
  - risk: Greenhouse mutation: POST: Add Office.
- remove_scheduled_interview:
  - endpoint: DELETE /scheduled_interviews/{{ record.scheduled_interview_id }}
  - required fields: scheduled_interview_id
  - risk: Destructive Greenhouse mutation: Delete: Remove Scheduled Interview.
- add_candidate_tag:
  - endpoint: POST /tags/candidate
  - risk: Greenhouse mutation: POST: Add New Candidate Tag.
- remove_tag_from_candidate:
  - endpoint: DELETE /candidates/{{ record.candidate_id }}/tags/{{ record.tag_id }}
  - required fields: candidate_id, tag_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove tag from candidate.
- add_a_candidate_tag:
  - endpoint: PUT /candidates/{{ record.candidate_id }}/tags/{{ record.tag_id }}
  - required fields: candidate_id, tag_id
  - risk: Greenhouse mutation: PUT: Add a candidate tag.
- change_user_permission_level:
  - endpoint: PATCH /users/permission_level
  - risk: Greenhouse mutation: PATCH: Change user permission level.
- add_user:
  - endpoint: POST /users
  - risk: Greenhouse mutation: POST: Add User.
- add_e_mail_address_to_user:
  - endpoint: POST /users/{{ record.user_id }}/email_addresses
  - required fields: user_id
  - risk: Greenhouse mutation: POST: Add E-mail Address To User.
- remove_a_job_permission:
  - endpoint: DELETE /users/{{ record.user_id }}/permissions/jobs
  - required fields: user_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove a Job Permission.
- add_a_job_permission:
  - endpoint: PUT /users/{{ record.user_id }}/permissions/jobs
  - required fields: user_id
  - risk: Greenhouse mutation: PUT: Add a Job Permission.
- remove_a_future_job_permission:
  - endpoint: DELETE /users/{{ record.user_id }}/permissions/future_jobs
  - required fields: user_id
  - risk: Destructive Greenhouse mutation: DELETE: Remove a Future Job Permission.
- add_a_future_job_permission:
  - endpoint: PUT /users/{{ record.user_id }}/permissions/future_jobs
  - required fields: user_id
  - risk: Greenhouse mutation: PUT: Add a Future Job Permission.

## Security

- read risk: external Greenhouse Harvest API read of candidates, applications, jobs, offers, users, and other documented resources
- write risk: external Greenhouse Harvest API mutations including candidate, application, job, office, department, tag, and user changes
- approval: required for every write action; destructive and identity-changing actions are marked confirm: destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect greenhouse
```

### Inspect as structured JSON

```bash
pm connectors inspect greenhouse --json
```

## Agent Rules

- Run pm connectors inspect greenhouse before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
