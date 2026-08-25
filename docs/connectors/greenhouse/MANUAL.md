# pm connectors inspect greenhouse

```text
NAME
  pm connectors inspect greenhouse - Greenhouse connector manual

SYNOPSIS
  pm connectors inspect greenhouse
  pm connectors inspect greenhouse --json
  pm credentials add <name> --connector greenhouse [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented Greenhouse Harvest REST API resources through the declarative connector engine.

ICON
  id: greenhouse
  asset: icons/greenhouse.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.greenhouse.io/harvest.html

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  application_id
  approval_flow_id
  base_url
  candidate_id
  custom_field_id
  demographic_answer_id
  demographic_answer_option_id
  demographic_question_id
  demographic_question_set_id
  department_id
  email_template_id
  field_type
  job_id
  job_post_id
  job_stage_id
  mode
  offer_id
  office_id
  on_behalf_of_user_id
  opening_id
  page_size
  prospect_pool_id
  scheduled_interview_id
  scorecard_id
  token
  user_id
  api_key (secret) (required)

ETL STREAMS
  candidates:
    primary key: id
    cursor: updated_at
    fields: company(string), created_at(string), first_name(string), id(integer), is_private(boolean), last_activity(string), last_name(string), title(string), updated_at(string)
  applications:
    primary key: id
    cursor: last_activity_at
    fields: applied_at(string), candidate_id(integer), id(integer), last_activity_at(string), rejected_at(string), source_id(integer), status(string)
  jobs:
    primary key: id
    cursor: updated_at
    fields: closed_at(string), confidential(boolean), created_at(string), id(integer), name(string), opened_at(string), requisition_id(string), status(string), updated_at(string)
  offers:
    primary key: id
    cursor: updated_at
    fields: application_id(integer), candidate_id(integer), created_at(string), id(integer), sent_at(string), starts_at(string), status(string), updated_at(string), version(integer)
  users:
    primary key: id
    cursor: updated_at
    fields: created_at(string), disabled(boolean), employee_id(string), first_name(string), id(integer), last_name(string), name(string), primary_email_address(string), site_admin(boolean), updated_at(string)
  activity_feed:
    fields: id(integer)
  application:
    fields: id(integer)
  approvals_for_job:
    fields: id(integer)
  approval_flow:
    fields: id(integer)
  pending_approvals_for_user:
    fields: id(integer)
  candidate:
    fields: id(integer)
  close_reasons:
    fields: id(integer)
  custom_fields:
    fields: id(integer)
  custom_field:
    fields: id(integer)
  custom_field_options:
    fields: id(integer)
  demographic_question_sets:
    fields: id(integer)
  demographic_question_set:
    fields: id(integer)
  demographic_questions:
    fields: id(integer)
  demographic_questions_for_demographic_question_set:
    fields: id(integer)
  demographic_question:
    fields: id(integer)
  demographic_answer_options:
    fields: id(integer)
  demographic_answer_options_for_demographic_question:
    fields: id(integer)
  demographic_answer_option:
    fields: id(integer)
  demographic_answers:
    fields: id(integer)
  demographic_answers_for_application:
    fields: id(integer)
  demographic_answer:
    fields: id(integer)
  departments:
    fields: id(integer)
  department:
    fields: id(integer)
  degrees:
    fields: id(integer)
  disciplines:
    fields: id(integer)
  schools:
    fields: id(integer)
  eeoc:
    fields: id(integer)
  eeoc_data_for_application:
    fields: id(integer)
  email_templates:
    fields: id(integer)
  email_template:
    fields: id(integer)
  job_openings:
    fields: id(integer)
  opening_for_job:
    fields: id(integer)
  job_posts:
    fields: id(integer)
  job_post:
    fields: id(integer)
  job_posts_for_job:
    fields: id(integer)
  job_post_for_job:
    fields: id(integer)
  custom_locations_for_job_post:
    fields: id(integer)
  job_stages:
    fields: id(integer)
  job_stages_for_job:
    fields: id(integer)
  job_stage:
    fields: id(integer)
  job:
    fields: id(integer)
  hiring_team:
    fields: id(integer)
  offers_for_application:
    fields: id(integer)
  current_offer_for_application:
    fields: id(integer)
  offer:
    fields: id(integer)
  offices:
    fields: id(integer)
  office:
    fields: id(integer)
  prospect_pools:
    fields: id(integer)
  prospect_pool:
    fields: id(integer)
  rejection_reasons:
    fields: id(integer)
  scheduled_interviews:
    fields: id(integer)
  scheduled_interviews_for_application:
    fields: id(integer)
  scheduled_interview:
    fields: id(integer)
  scorecards:
    fields: id(integer)
  scorecards_for_application:
    fields: id(integer)
  scorecard:
    fields: id(integer)
  sources:
    fields: id(integer)
  candidate_tags:
    fields: id(integer)
  tags_applied_to_candidate:
    fields: id(integer)
  tracking_link_data_for_token:
    fields: id(integer)
  user:
    fields: id(integer)
  job_permissions:
    fields: id(integer)
  future_job_permissions:
    fields: id(integer)
  user_roles:
    fields: id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_application:
    endpoint: DELETE /applications/{{ record.application_id }}
    required fields: application_id
    risk: Destructive Greenhouse mutation: DELETE: Delete Application.
  add_application_to_candidate_prospect:
    endpoint: POST /candidates/{{ record.candidate_id }}/applications
    required fields: candidate_id
    risk: Greenhouse mutation: POST: Add Application to Candidate/Prospect.
  update_application:
    endpoint: PATCH /applications/{{ record.application_id }}
    required fields: application_id
    risk: Greenhouse mutation: PATCH: Update Application.
  advance_application:
    endpoint: POST /applications/{{ record.application_id }}/advance
    required fields: application_id
    risk: Greenhouse mutation: POST: Advance Application.
  move_application_different_job:
    endpoint: POST /applications/{{ record.application_id }}/transfer_to_job
    required fields: application_id
    risk: Greenhouse mutation: POST: Move Application (Different Job).
  move_application_same_job:
    endpoint: POST /applications/{{ record.application_id }}/move
    required fields: application_id
    risk: Greenhouse mutation: POST: Move Application (Same Job).
  convert_prospect_to_candidate:
    endpoint: PATCH /applications/{{ record.application_id }}/convert_prospect
    required fields: application_id
    risk: Greenhouse mutation: PATCH: Convert Prospect To Candidate.
  add_attachment_to_application:
    endpoint: POST /applications/{{ record.application_id }}/attachments
    required fields: application_id
    risk: Greenhouse mutation: POST: Add Attachment to Application.
  hire_application:
    endpoint: POST /applications/{{ record.application_id }}/hire
    required fields: application_id
    risk: Greenhouse mutation: POST: Hire Application.
  reject_application:
    endpoint: POST /applications/{{ record.application_id }}/reject
    required fields: application_id
    risk: Destructive Greenhouse mutation: POST: Reject Application.
  update_rejection_reason:
    endpoint: PATCH /applications/{{ record.application_id }}/reject
    required fields: application_id
    risk: Greenhouse mutation: PATCH: Update Rejection Reason.
  unreject_application:
    endpoint: POST /applications/{{ record.application_id }}/unreject
    required fields: application_id
    risk: Greenhouse mutation: POST: Unreject Application.
  request_approvals:
    endpoint: POST /approval_flows/{{ record.approval_flow_id }}/request_approvals
    required fields: approval_flow_id
    risk: Greenhouse mutation: POST: Request Approvals.
  replace_an_approver_in_an_approver_group:
    endpoint: PUT /approver_groups/{{ record.approver_group_id }}/replace_approvers
    required fields: approver_group_id
    risk: Greenhouse mutation: PUT: Replace an approver in an approver group.
  create_or_replace_an_approval_flow:
    endpoint: PUT /jobs/{{ record.job_id }}/approval_flows
    required fields: job_id
    risk: Greenhouse mutation: PUT: Create or replace an approval flow.
  delete_candidate:
    endpoint: DELETE /candidates/{{ record.candidate_id }}
    required fields: candidate_id
    risk: Destructive Greenhouse mutation: DELETE: Delete Candidate.
  edit_candidate:
    endpoint: PATCH /candidates/{{ record.candidate_id }}
    required fields: candidate_id
    risk: Greenhouse mutation: PATCH: Edit Candidate.
  add_attachment:
    endpoint: POST /candidates/{{ record.candidate_id }}/attachments
    required fields: candidate_id
    risk: Greenhouse mutation: POST: Add Attachment.
  add_candidate:
    endpoint: POST /candidates
    risk: Greenhouse mutation: POST: Add Candidate.
  add_note:
    endpoint: POST /candidates/{{ record.candidate_id }}/activity_feed/notes
    required fields: candidate_id
    risk: Greenhouse mutation: POST: Add Note.
  add_e_mail_note:
    endpoint: POST /candidates/{{ record.candidate_id }}/activity_feed/emails
    required fields: candidate_id
    risk: Greenhouse mutation: POST: Add E-mail Note.
  add_education:
    endpoint: POST /candidates/{{ record.candidate_id }}/educations
    required fields: candidate_id
    risk: Greenhouse mutation: POST: Add Education.
  remove_education_from_candidate:
    endpoint: DELETE /candidates/{{ record.candidate_id }}/educations/{{ record.education_id }}
    required fields: candidate_id, education_id
    risk: Destructive Greenhouse mutation: DELETE: Remove Education From Candidate.
  add_employment:
    endpoint: POST /candidates/{{ record.candidate_id }}/employments
    required fields: candidate_id
    risk: Greenhouse mutation: POST: Add Employment.
  remove_employment_from_candidate:
    endpoint: DELETE /candidates/{{ record.candidate_id }}/employments/{{ record.employment_id }}
    required fields: candidate_id, employment_id
    risk: Destructive Greenhouse mutation: DELETE: Remove Employment From Candidate.
  add_prospect:
    endpoint: POST /prospects
    risk: Greenhouse mutation: POST: Add Prospect.
  anonymize_candidate:
    endpoint: PUT /candidates/{{ record.candidate_id }}/anonymize?fields={{ record.field_names }}
    required fields: candidate_id, field_names
    risk: Destructive Greenhouse mutation: PUT: Anonymize Candidate.
  merge_candidates:
    endpoint: PUT /candidates/merge
    risk: Destructive Greenhouse mutation: PUT: Merge Candidates.
  create_custom_field:
    endpoint: POST /custom_fields
    risk: Greenhouse mutation: POST: Create Custom Field.
  update_custom_field:
    endpoint: PATCH /custom_fields/{{ record.custom_field_id }}
    required fields: custom_field_id
    risk: Greenhouse mutation: PATCH: Update Custom Field.
  delete_custom_field:
    endpoint: DELETE /custom_fields/{{ record.custom_field_id }}
    required fields: custom_field_id
    risk: Destructive Greenhouse mutation: DELETE: Delete Custom Field.
  create_custom_field_options:
    endpoint: POST /custom_field/{{ record.custom_field_id }}/custom_field_options
    required fields: custom_field_id
    risk: Greenhouse mutation: POST: Create Custom Field Options.
  update_custom_field_options:
    endpoint: PATCH /custom_field/{{ record.custom_field_id }}/custom_field_options
    required fields: custom_field_id
    risk: Greenhouse mutation: PATCH: Update Custom Field Options.
  remove_custom_field_options:
    endpoint: DELETE /custom_field/{{ record.custom_field_id }}/custom_field_options
    required fields: custom_field_id
    risk: Destructive Greenhouse mutation: DELETE: Remove Custom Field Options.
  edit_department:
    endpoint: PATCH /departments/{{ record.department_id }}
    required fields: department_id
    risk: Greenhouse mutation: PATCH: Edit Department.
  add_department:
    endpoint: POST /departments
    risk: Greenhouse mutation: POST: Add Department.
  edit_openings:
    endpoint: PATCH /jobs/{{ record.job_id }}/openings/{{ record.opening_id }}
    required fields: job_id, opening_id
    risk: Greenhouse mutation: PATCH: Edit Openings.
  create_new_openings:
    endpoint: POST /jobs/{{ record.job_id }}/openings
    required fields: job_id
    risk: Greenhouse mutation: POST: Create New Openings.
  update_job:
    endpoint: PATCH /jobs/{{ record.job_id }}
    required fields: job_id
    risk: Greenhouse mutation: PATCH: Update Job.
  create_job:
    endpoint: POST /jobs
    risk: Greenhouse mutation: POST: Create Job.
  replace_hiring_team:
    endpoint: PUT /jobs/{{ record.job_id }}/hiring_team
    required fields: job_id
    risk: Greenhouse mutation: PUT: Replace Hiring Team.
  add_hiring_team_members:
    endpoint: POST /jobs/{{ record.job_id }}/hiring_team
    required fields: job_id
    risk: Greenhouse mutation: POST: Add Hiring Team Members.
  remove_hiring_team_member:
    endpoint: DELETE /jobs/{{ record.job_id }}/hiring_team
    required fields: job_id
    risk: Destructive Greenhouse mutation: DELETE: Remove Hiring Team Member.
  update_current_offer:
    endpoint: PATCH /applications/{{ record.application_id }}/offers/current_offer
    required fields: application_id
    risk: Greenhouse mutation: PATCH: Update Current Offer.
  edit_office:
    endpoint: PATCH /offices/{{ record.office_id }}
    required fields: office_id
    risk: Greenhouse mutation: PATCH: Edit Office.
  add_office:
    endpoint: POST /offices
    risk: Greenhouse mutation: POST: Add Office.
  remove_scheduled_interview:
    endpoint: DELETE /scheduled_interviews/{{ record.scheduled_interview_id }}
    required fields: scheduled_interview_id
    risk: Destructive Greenhouse mutation: Delete: Remove Scheduled Interview.
  add_candidate_tag:
    endpoint: POST /tags/candidate
    risk: Greenhouse mutation: POST: Add New Candidate Tag.
  remove_tag_from_candidate:
    endpoint: DELETE /candidates/{{ record.candidate_id }}/tags/{{ record.tag_id }}
    required fields: candidate_id, tag_id
    risk: Destructive Greenhouse mutation: DELETE: Remove tag from candidate.
  destroy_candidate_tag:
    endpoint: DELETE /tags/candidate/{{ record.tag_id }}
    required fields: tag_id
    risk: Destructive Greenhouse mutation: DELETE: Destroy a candidate tag. Greenhouse queues an asynchronous job that strips this tag from EVERY candidate it is applied to before deleting the tag itself, so the blast radius is organisation-wide and is not bounded by any candidate id.
  add_a_candidate_tag:
    endpoint: PUT /candidates/{{ record.candidate_id }}/tags/{{ record.tag_id }}
    required fields: candidate_id, tag_id
    risk: Greenhouse mutation: PUT: Add a candidate tag.
  change_user_permission_level:
    endpoint: PATCH /users/permission_level
    risk: Greenhouse mutation: PATCH: Change user permission level.
  add_user:
    endpoint: POST /users
    risk: Greenhouse mutation: POST: Add User.
  add_e_mail_address_to_user:
    endpoint: POST /users/{{ record.user_id }}/email_addresses
    required fields: user_id
    risk: Greenhouse mutation: POST: Add E-mail Address To User.
  remove_a_job_permission:
    endpoint: DELETE /users/{{ record.user_id }}/permissions/jobs
    required fields: user_id
    risk: Destructive Greenhouse mutation: DELETE: Remove a Job Permission.
  add_a_job_permission:
    endpoint: PUT /users/{{ record.user_id }}/permissions/jobs
    required fields: user_id
    risk: Greenhouse mutation: PUT: Add a Job Permission.
  remove_a_future_job_permission:
    endpoint: DELETE /users/{{ record.user_id }}/permissions/future_jobs
    required fields: user_id
    risk: Destructive Greenhouse mutation: DELETE: Remove a Future Job Permission.
  add_a_future_job_permission:
    endpoint: PUT /users/{{ record.user_id }}/permissions/future_jobs
    required fields: user_id
    risk: Greenhouse mutation: PUT: Add a Future Job Permission.

SECURITY
  read risk: external Greenhouse Harvest API read of candidates, applications, jobs, offers, users, and other documented resources
  write risk: external Greenhouse Harvest API mutations including candidate, application, job, office, department, tag, and user changes
  approval: required for every write action; destructive and identity-changing actions are marked confirm: destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read Greenhouse Harvest records and safely plan typed Greenhouse mutations.
  Usage: pm greenhouse <command> [flags]
  Source CLI: Greenhouse Harvest API (Greenhouse Harvest reference fetched 2026-08-07 (HTTP 200, 1,636,662 bytes))
  Global flags:
    --credential (string): Credential name to use for the Greenhouse request.
    --connection (string): Alias for --credential.
    --config (string_array): Connector config override as key=value; never pass secret values here.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum ETL records to emit.
    --max-bytes (integer): Maximum bounded direct-read response bytes.
    --preview (boolean): Preview a reverse-ETL write command without making a network mutation.
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
    --confirm (string): Typed confirmation challenge for destructive reverse-ETL writes.
  ETL streams
    candidates list - Read Greenhouse List Candidates as ETL records. [intent=etl availability=implemented stream=candidates]
    applications list - Read Greenhouse List Applications as ETL records. [intent=etl availability=implemented stream=applications]
    jobs list - Read Greenhouse List Jobs as ETL records. [intent=etl availability=implemented stream=jobs]
    offers list - Read Greenhouse List Offers as ETL records. [intent=etl availability=implemented stream=offers]
    users list - Read Greenhouse List Users as ETL records. [intent=etl availability=implemented stream=users]
    activity-feed list - Read Greenhouse Retrieve Activity Feed as ETL records. [intent=etl availability=implemented stream=activity_feed]; flags: --candidate-id (required, non-empty)
    application list - Read Greenhouse Retrieve Application as ETL records. [intent=etl availability=implemented stream=application]; flags: --application-id (required, non-empty)
    approvals-for-job list - Read Greenhouse List Approvals For Job as ETL records. [intent=etl availability=implemented stream=approvals_for_job]; flags: --job-id (required, non-empty)
    approval-flow list - Read Greenhouse Retrieve Approval Flow as ETL records. [intent=etl availability=implemented stream=approval_flow]; flags: --approval-flow-id (required, non-empty)
    pending-approvals-for-user list - Read Greenhouse Pending Approvals For User as ETL records. [intent=etl availability=implemented stream=pending_approvals_for_user]; flags: --user-id (required, non-empty)
    candidate list - Read Greenhouse Retrieve Candidate as ETL records. [intent=etl availability=implemented stream=candidate]; flags: --candidate-id (required, non-empty)
    close-reasons list - Read Greenhouse List Close Reasons as ETL records. [intent=etl availability=implemented stream=close_reasons]
    custom-fields list - Read Greenhouse List Custom Fields as ETL records. [intent=etl availability=implemented stream=custom_fields]; flags: --field-type (required, non-empty)
    custom-field list - Read Greenhouse Retrieve Custom Field as ETL records. [intent=etl availability=implemented stream=custom_field]; flags: --custom-field-id (required, non-empty)
    custom-field-options list - Read Greenhouse List Custom Field Options as ETL records. [intent=etl availability=implemented stream=custom_field_options]; flags: --custom-field-id (required, non-empty)
    demographic-question-sets list - Read Greenhouse List Demographic Question Sets as ETL records. [intent=etl availability=implemented stream=demographic_question_sets]
    demographic-question-set list - Read Greenhouse Retrieve Demographic Question Set as ETL records. [intent=etl availability=implemented stream=demographic_question_set]; flags: --demographic-question-set-id (required, non-empty)
    demographic-questions list - Read Greenhouse List Demographic Questions as ETL records. [intent=etl availability=implemented stream=demographic_questions]
    demographic-questions-for-demographic-question-set list - Read Greenhouse List Demographic Questions For Demographic Question Set as ETL records. [intent=etl availability=implemented stream=demographic_questions_for_demographic_question_set]; flags: --demographic-question-set-id (required, non-empty)
    demographic-question list - Read Greenhouse Retrieve Demographic Question as ETL records. [intent=etl availability=implemented stream=demographic_question]; flags: --demographic-question-id (required, non-empty)
    demographic-answer-options list - Read Greenhouse List Demographic Answer Options as ETL records. [intent=etl availability=implemented stream=demographic_answer_options]
    demographic-answer-options-for-demographic-question list - Read Greenhouse List Demographic Answer Options For Demographic Question as ETL records. [intent=etl availability=implemented stream=demographic_answer_options_for_demographic_question]; flags: --demographic-question-id (required, non-empty)
    demographic-answer-option list - Read Greenhouse Retrieve Demographic Answer Option as ETL records. [intent=etl availability=implemented stream=demographic_answer_option]; flags: --demographic-answer-option-id (required, non-empty)
    demographic-answers list - Read Greenhouse List Demographic Answers as ETL records. [intent=etl availability=implemented stream=demographic_answers]
    demographic-answers-for-application list - Read Greenhouse List Demographic Answers For Application as ETL records. [intent=etl availability=implemented stream=demographic_answers_for_application]; flags: --application-id (required, non-empty)
    demographic-answer list - Read Greenhouse Retrieve Demographic Answer as ETL records. [intent=etl availability=implemented stream=demographic_answer]; flags: --demographic-answer-id (required, non-empty)
    departments list - Read Greenhouse List Departments as ETL records. [intent=etl availability=implemented stream=departments]
    department list - Read Greenhouse Retrieve Department as ETL records. [intent=etl availability=implemented stream=department]; flags: --department-id (required, non-empty)
    degrees list - Read Greenhouse List Degrees as ETL records. [intent=etl availability=implemented stream=degrees]
    disciplines list - Read Greenhouse List Disciplines as ETL records. [intent=etl availability=implemented stream=disciplines]
    schools list - Read Greenhouse List Schools as ETL records. [intent=etl availability=implemented stream=schools]
    eeoc list - Read Greenhouse List EEOC as ETL records. [intent=etl availability=implemented stream=eeoc]
    eeoc-data-for-application list - Read Greenhouse Retrieve EEOC Data for Application as ETL records. [intent=etl availability=implemented stream=eeoc_data_for_application]; flags: --application-id (required, non-empty)
    email-templates list - Read Greenhouse List Email Templates as ETL records. [intent=etl availability=implemented stream=email_templates]
    email-template list - Read Greenhouse Retrieve Email Template as ETL records. [intent=etl availability=implemented stream=email_template]; flags: --email-template-id (required, non-empty)
    job-openings list - Read Greenhouse List Job Openings as ETL records. [intent=etl availability=implemented stream=job_openings]; flags: --job-id (required, non-empty)
    opening-for-job list - Read Greenhouse Single Opening For Job as ETL records. [intent=etl availability=implemented stream=opening_for_job]; flags: --job-id (required, non-empty), --opening-id (required, non-empty)
    job-posts list - Read Greenhouse List Job Posts as ETL records. [intent=etl availability=implemented stream=job_posts]
    job-post list - Read Greenhouse Retrieve Job Post as ETL records. [intent=etl availability=implemented stream=job_post]; flags: --job-post-id (required, non-empty)
    job-posts-for-job list - Read Greenhouse List Job Posts for Job as ETL records. [intent=etl availability=implemented stream=job_posts_for_job]; flags: --job-id (required, non-empty)
    job-post-for-job list - Read Greenhouse Retrieve Job Post for Job as ETL records. [intent=etl availability=implemented stream=job_post_for_job]; flags: --job-id (required, non-empty)
    custom-locations-for-job-post list - Read Greenhouse Retrieve Custom Locations for Job Post as ETL records. [intent=etl availability=implemented stream=custom_locations_for_job_post]; flags: --job-post-id (required, non-empty)
    job-stages list - Read Greenhouse List Job Stages as ETL records. [intent=etl availability=implemented stream=job_stages]
    job-stages-for-job list - Read Greenhouse List Job Stages for Job as ETL records. [intent=etl availability=implemented stream=job_stages_for_job]; flags: --job-id (required, non-empty)
    job-stage list - Read Greenhouse Retrieve Job Stage as ETL records. [intent=etl availability=implemented stream=job_stage]; flags: --job-stage-id (required, non-empty)
    job list - Read Greenhouse Retrieve Job as ETL records. [intent=etl availability=implemented stream=job]; flags: --job-id (required, non-empty)
    hiring-team list - Read Greenhouse Hiring Team as ETL records. [intent=etl availability=implemented stream=hiring_team]; flags: --job-id (required, non-empty)
    offers-for-application list - Read Greenhouse List Offers for Application as ETL records. [intent=etl availability=implemented stream=offers_for_application]; flags: --application-id (required, non-empty)
    current-offer-for-application list - Read Greenhouse Retrieve Current Offer for Application as ETL records. [intent=etl availability=implemented stream=current_offer_for_application]; flags: --application-id (required, non-empty)
    offer list - Read Greenhouse Retrieve Offer as ETL records. [intent=etl availability=implemented stream=offer]; flags: --offer-id (required, non-empty)
    offices list - Read Greenhouse List Offices as ETL records. [intent=etl availability=implemented stream=offices]
    office list - Read Greenhouse Retrieve Office as ETL records. [intent=etl availability=implemented stream=office]; flags: --office-id (required, non-empty)
    prospect-pools list - Read Greenhouse List Prospect Pools as ETL records. [intent=etl availability=implemented stream=prospect_pools]
    prospect-pool list - Read Greenhouse Retrieve Prospect Pool as ETL records. [intent=etl availability=implemented stream=prospect_pool]; flags: --prospect-pool-id (required, non-empty)
    rejection-reasons list - Read Greenhouse List Rejection Reasons as ETL records. [intent=etl availability=implemented stream=rejection_reasons]
    scheduled-interviews list - Read Greenhouse List Scheduled Interviews as ETL records. [intent=etl availability=implemented stream=scheduled_interviews]
    scheduled-interviews-for-application list - Read Greenhouse List Scheduled Interviews for Application as ETL records. [intent=etl availability=implemented stream=scheduled_interviews_for_application]; flags: --application-id (required, non-empty)
    scheduled-interview list - Read Greenhouse Retrieve Scheduled Interview as ETL records. [intent=etl availability=implemented stream=scheduled_interview]; flags: --scheduled-interview-id (required, non-empty)
    scorecards list - Read Greenhouse List Scorecards as ETL records. [intent=etl availability=implemented stream=scorecards]
    scorecards-for-application list - Read Greenhouse List Scorecards for Application as ETL records. [intent=etl availability=implemented stream=scorecards_for_application]; flags: --application-id (required, non-empty)
    scorecard list - Read Greenhouse Retrieve Scorecard as ETL records. [intent=etl availability=implemented stream=scorecard]; flags: --scorecard-id (required, non-empty)
    sources list - Read Greenhouse List Sources as ETL records. [intent=etl availability=implemented stream=sources]
    candidate-tags list - Read Greenhouse List Candidate Tags as ETL records. [intent=etl availability=implemented stream=candidate_tags]
    tags-applied-to-candidate list - Read Greenhouse List tags applied to candidate as ETL records. [intent=etl availability=implemented stream=tags_applied_to_candidate]; flags: --candidate-id (required, non-empty)
    tracking-link-data-for-token list - Read Greenhouse Tracking Link Data for Token as ETL records. [intent=etl availability=implemented stream=tracking_link_data_for_token]; flags: --token (required, non-empty)
    user list - Read Greenhouse Retrieve User as ETL records. [intent=etl availability=implemented stream=user]; flags: --user-id (required, non-empty)
    job-permissions list - Read Greenhouse List Job Permissions as ETL records. [intent=etl availability=implemented stream=job_permissions]; flags: --user-id (required, non-empty)
    future-job-permissions list - Read Greenhouse List Future Job Permissions as ETL records. [intent=etl availability=implemented stream=future_job_permissions]; flags: --user-id (required, non-empty)
    user-roles list - Read Greenhouse List User Roles as ETL records. [intent=etl availability=implemented stream=user_roles]
  Reverse ETL write plans
    add-a-candidate-tag plan - Add a candidate tag. [intent=reverse_etl availability=implemented write=add_a_candidate_tag]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PUT: Add a candidate tag.; flags: --candidate-id (required, non-empty), --tag-id (required, non-empty)
    add-a-future-job-permission plan - Add a Future Job Permission. [intent=reverse_etl availability=implemented write=add_a_future_job_permission]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PUT: Add a Future Job Permission.; flags: --user-id (required, non-empty)
    add-a-job-permission plan - Add a Job Permission. [intent=reverse_etl availability=implemented write=add_a_job_permission]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PUT: Add a Job Permission.; flags: --user-id (required, non-empty)
    add-application-to-candidate-prospect plan - Add Application to Candidate/Prospect. [intent=reverse_etl availability=implemented write=add_application_to_candidate_prospect]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Application to Candidate/Prospect.; flags: --candidate-id (required, non-empty)
    add-attachment plan - Add Attachment. [intent=reverse_etl availability=implemented write=add_attachment]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Attachment.; flags: --candidate-id (required, non-empty)
    add-attachment-to-application plan - Add Attachment to Application. [intent=reverse_etl availability=implemented write=add_attachment_to_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Attachment to Application.; flags: --application-id (required, non-empty)
    add-candidate plan - Add Candidate. [intent=reverse_etl availability=implemented write=add_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Candidate.
    add-candidate-tag plan - Add New Candidate Tag. [intent=reverse_etl availability=implemented write=add_candidate_tag]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add New Candidate Tag.
    add-department plan - Add Department. [intent=reverse_etl availability=implemented write=add_department]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Department.
    add-e-mail-address-to-user plan - Add E-mail Address To User. [intent=reverse_etl availability=implemented write=add_e_mail_address_to_user]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add E-mail Address To User.; flags: --user-id (required, non-empty)
    add-e-mail-note plan - Add E-mail Note. [intent=reverse_etl availability=implemented write=add_e_mail_note]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add E-mail Note.; flags: --candidate-id (required, non-empty)
    add-education plan - Add Education. [intent=reverse_etl availability=implemented write=add_education]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Education.; flags: --candidate-id (required, non-empty)
    add-employment plan - Add Employment. [intent=reverse_etl availability=implemented write=add_employment]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Employment.; flags: --candidate-id (required, non-empty)
    add-hiring-team-members plan - Add Hiring Team Members. [intent=reverse_etl availability=implemented write=add_hiring_team_members]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Hiring Team Members.; flags: --job-id (required, non-empty)
    add-note plan - Add Note. [intent=reverse_etl availability=implemented write=add_note]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Note.; flags: --candidate-id (required, non-empty)
    add-office plan - Add Office. [intent=reverse_etl availability=implemented write=add_office]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Office.
    add-prospect plan - Add Prospect. [intent=reverse_etl availability=implemented write=add_prospect]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add Prospect.
    add-user plan - Add User. [intent=reverse_etl availability=implemented write=add_user]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Add User.
    advance-application plan - Advance Application. [intent=reverse_etl availability=implemented write=advance_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Advance Application.; flags: --application-id (required, non-empty)
    anonymize-candidate plan - Anonymize Candidate. [intent=reverse_etl availability=implemented write=anonymize_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: PUT: Anonymize Candidate.; flags: --candidate-id (required, non-empty), --field-names (required, non-empty)
    change-user-permission-level plan - Change user permission level. [intent=reverse_etl availability=implemented write=change_user_permission_level]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Change user permission level.
    convert-prospect-to-candidate plan - Convert Prospect To Candidate. [intent=reverse_etl availability=implemented write=convert_prospect_to_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Convert Prospect To Candidate.; flags: --application-id (required, non-empty)
    create-custom-field plan - Create Custom Field. [intent=reverse_etl availability=implemented write=create_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Create Custom Field.
    create-custom-field-options plan - Create Custom Field Options. [intent=reverse_etl availability=implemented write=create_custom_field_options]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Create Custom Field Options.; flags: --custom-field-id (required, non-empty)
    create-job plan - Create Job. [intent=reverse_etl availability=implemented write=create_job]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Create Job.
    create-new-openings plan - Create New Openings. [intent=reverse_etl availability=implemented write=create_new_openings]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Create New Openings.; flags: --job-id (required, non-empty)
    create-or-replace-an-approval-flow plan - Create or replace an approval flow. [intent=reverse_etl availability=implemented write=create_or_replace_an_approval_flow]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PUT: Create or replace an approval flow.; flags: --job-id (required, non-empty)
    delete-application plan - Delete Application. [intent=reverse_etl availability=implemented write=delete_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Delete Application.; flags: --application-id (required, non-empty)
    delete-candidate plan - Delete Candidate. [intent=reverse_etl availability=implemented write=delete_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Delete Candidate.; flags: --candidate-id (required, non-empty)
    delete-custom-field plan - Delete Custom Field. [intent=reverse_etl availability=implemented write=delete_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Delete Custom Field.; flags: --custom-field-id (required, non-empty)
    destroy-candidate-tag plan - Destroy a candidate tag. [intent=reverse_etl availability=implemented write=destroy_candidate_tag]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Destroy a candidate tag. Greenhouse queues an asynchronous job that strips this tag from EVERY candidate it is applied to before deleting the tag itself, so the blast radius is organisation-wide and is not bounded by any candidate id.; flags: --tag-id (required, non-empty)
    edit-candidate plan - Edit Candidate. [intent=reverse_etl availability=implemented write=edit_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Edit Candidate.; flags: --candidate-id (required, non-empty)
    edit-department plan - Edit Department. [intent=reverse_etl availability=implemented write=edit_department]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Edit Department.; flags: --department-id (required, non-empty)
    edit-office plan - Edit Office. [intent=reverse_etl availability=implemented write=edit_office]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Edit Office.; flags: --office-id (required, non-empty)
    edit-openings plan - Edit Openings. [intent=reverse_etl availability=implemented write=edit_openings]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Edit Openings.; flags: --job-id (required, non-empty), --opening-id (required, non-empty)
    hire-application plan - Hire Application. [intent=reverse_etl availability=implemented write=hire_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Hire Application.; flags: --application-id (required, non-empty)
    merge-candidates plan - Merge Candidates. [intent=reverse_etl availability=implemented write=merge_candidates]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: PUT: Merge Candidates.
    move-application-different-job plan - Move Application (Different Job). [intent=reverse_etl availability=implemented write=move_application_different_job]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Move Application (Different Job).; flags: --application-id (required, non-empty)
    move-application-same-job plan - Move Application (Same Job). [intent=reverse_etl availability=implemented write=move_application_same_job]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Move Application (Same Job).; flags: --application-id (required, non-empty)
    reject-application plan - Reject Application. [intent=reverse_etl availability=implemented write=reject_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: POST: Reject Application.; flags: --application-id (required, non-empty)
    remove-a-future-job-permission plan - Remove a Future Job Permission. [intent=reverse_etl availability=implemented write=remove_a_future_job_permission]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove a Future Job Permission.; flags: --user-id (required, non-empty)
    remove-a-job-permission plan - Remove a Job Permission. [intent=reverse_etl availability=implemented write=remove_a_job_permission]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove a Job Permission.; flags: --user-id (required, non-empty)
    remove-custom-field-options plan - Remove Custom Field Options. [intent=reverse_etl availability=implemented write=remove_custom_field_options]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove Custom Field Options.; flags: --custom-field-id (required, non-empty)
    remove-education-from-candidate plan - Remove Education From Candidate. [intent=reverse_etl availability=implemented write=remove_education_from_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove Education From Candidate.; flags: --candidate-id (required, non-empty), --education-id (required, non-empty)
    remove-employment-from-candidate plan - Remove Employment From Candidate. [intent=reverse_etl availability=implemented write=remove_employment_from_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove Employment From Candidate.; flags: --candidate-id (required, non-empty), --employment-id (required, non-empty)
    remove-hiring-team-member plan - Remove Hiring Team Member. [intent=reverse_etl availability=implemented write=remove_hiring_team_member]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove Hiring Team Member.; flags: --job-id (required, non-empty)
    remove-scheduled-interview plan - Remove Scheduled Interview. [intent=reverse_etl availability=implemented write=remove_scheduled_interview]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: Delete: Remove Scheduled Interview.; flags: --scheduled-interview-id (required, non-empty)
    remove-tag-from-candidate plan - Remove tag from candidate. [intent=reverse_etl availability=implemented write=remove_tag_from_candidate]; approval: reverse ETL writes require plan, preview, explicit approval, then execute. This action is destructive and additionally requires a typed confirmation.; risk: Destructive Greenhouse mutation: DELETE: Remove tag from candidate.; flags: --candidate-id (required, non-empty), --tag-id (required, non-empty)
    replace-an-approver-in-an-approver-group plan - Replace an approver in an approver group. [intent=reverse_etl availability=implemented write=replace_an_approver_in_an_approver_group]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PUT: Replace an approver in an approver group.; flags: --approver-group-id (required, non-empty)
    replace-hiring-team plan - Replace Hiring Team. [intent=reverse_etl availability=implemented write=replace_hiring_team]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PUT: Replace Hiring Team.; flags: --job-id (required, non-empty)
    request-approvals plan - Request Approvals. [intent=reverse_etl availability=implemented write=request_approvals]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Request Approvals.; flags: --approval-flow-id (required, non-empty)
    unreject-application plan - Unreject Application. [intent=reverse_etl availability=implemented write=unreject_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: POST: Unreject Application.; flags: --application-id (required, non-empty)
    update-application plan - Update Application. [intent=reverse_etl availability=implemented write=update_application]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Update Application.; flags: --application-id (required, non-empty)
    update-current-offer plan - Update Current Offer. [intent=reverse_etl availability=implemented write=update_current_offer]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Update Current Offer.; flags: --application-id (required, non-empty)
    update-custom-field plan - Update Custom Field. [intent=reverse_etl availability=implemented write=update_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Update Custom Field.; flags: --custom-field-id (required, non-empty)
    update-custom-field-options plan - Update Custom Field Options. [intent=reverse_etl availability=implemented write=update_custom_field_options]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Update Custom Field Options.; flags: --custom-field-id (required, non-empty)
    update-job plan - Update Job. [intent=reverse_etl availability=implemented write=update_job]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Update Job.; flags: --job-id (required, non-empty)
    update-rejection-reason plan - Update Rejection Reason. [intent=reverse_etl availability=implemented write=update_rejection_reason]; approval: reverse ETL writes require plan, preview, explicit approval, then execute.; risk: Greenhouse mutation: PATCH: Update Rejection Reason.; flags: --application-id (required, non-empty)
  Help topics:
    operation-ledger - Greenhouse Harvest documents 138 operations; 127 are covered by this bundle and 11 are blocked with a named dependency.
    write-safety - Greenhouse write commands use reverse ETL plan -> preview -> approval -> execute; destructive actions require typed confirmation.

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

EXAMPLES
  # Inspect as a manual
  pm connectors inspect greenhouse

  # Inspect as structured JSON
  pm connectors inspect greenhouse --json

AGENT WORKFLOW
  - Run pm connectors inspect greenhouse before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
