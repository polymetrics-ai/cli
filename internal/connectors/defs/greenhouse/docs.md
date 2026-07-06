# Overview

Reads and writes documented Greenhouse Harvest REST API resources through the connectorengine.

Readable streams: `candidates`, `applications`, `jobs`, `offers`, `users`, `activity_feed`,
`application`, `approvals_for_job`, `approval_flow`, `pending_approvals_for_user`, `candidate`,
`close_reasons`, `custom_fields`, `custom_field`, `custom_field_options`,
`demographic_question_sets`, `demographic_question_set`, `demographic_questions`,
`demographic_questions_for_demographic_question_set`, `demographic_question`,
`demographic_answer_options`, `demographic_answer_options_for_demographic_question`,
`demographic_answer_option`, `demographic_answers`, `demographic_answers_for_application`,
`demographic_answer`, `departments`, `department`, `degrees`, `disciplines`, `schools`, `eeoc`,
`eeoc_data_for_application`, `email_templates`, `email_template`, `job_openings`, `opening_for_job`,
`job_posts`, `job_post`, `job_posts_for_job`, `job_post_for_job`, `custom_locations_for_job_post`,
`job_stages`, `job_stages_for_job`, `job_stage`, `job`, `hiring_team`, `offers_for_application`,
`current_offer_for_application`, `offer`, `offices`, `office`, `prospect_pools`, `prospect_pool`,
`rejection_reasons`, `scheduled_interviews`, `scheduled_interviews_for_application`,
`scheduled_interview`, `scorecards`, `scorecards_for_application`, `scorecard`, `sources`,
`candidate_tags`, `tags_applied_to_candidate`, `tracking_link_data_for_token`, `user`,
`job_permissions`, `future_job_permissions`, `user_roles`.

Write actions: `delete_application`, `add_application_to_candidate_prospect`, `update_application`,
`advance_application`, `move_application_different_job`, `move_application_same_job`,
`convert_prospect_to_candidate`, `add_attachment_to_application`, `hire_application`,
`reject_application`, `update_rejection_reason`, `unreject_application`, `request_approvals`,
`replace_an_approver_in_an_approver_group`, `create_or_replace_an_approval_flow`,
`delete_candidate`, `edit_candidate`, `add_attachment`, `add_candidate`, `add_note`,
`add_e_mail_note`, `add_education`, `remove_education_from_candidate`, `add_employment`,
`remove_employment_from_candidate`, `add_prospect`, `anonymize_candidate`, `merge_candidates`,
`create_custom_field`, `update_custom_field`, `delete_custom_field`, `create_custom_field_options`,
`update_custom_field_options`, `remove_custom_field_options`, `edit_department`, `add_department`,
`edit_openings`, `create_new_openings`, `update_job`, `create_job`, `replace_hiring_team`,
`add_hiring_team_members`, `remove_hiring_team_member`, `update_current_offer`, `edit_office`,
`add_office`, `remove_scheduled_interview`, `add_candidate_tag`, `remove_tag_from_candidate`,
`add_a_candidate_tag`, `change_user_permission_level`, `add_user`, `add_e_mail_address_to_user`,
`remove_a_job_permission`, `add_a_job_permission`, `remove_a_future_job_permission`,
`add_a_future_job_permission`.

Service API documentation: https://developers.greenhouse.io/harvest.html.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Greenhouse Harvest API token, sent as the HTTP Basic auth
  username with a blank password (Authorization: Basic base64(api_key:)); never logged.
- `application_id` (optional, string); Application Id used by Greenhouse stream paths.
- `approval_flow_id` (optional, string); Approval Flow Id used by Greenhouse stream paths.
- `base_url` (optional, string); default `https://harvest.greenhouse.io/v1`; format `uri`;
  Greenhouse Harvest API v1 base URL override, for tests or proxies.
- `candidate_id` (optional, string); Candidate Id used by Greenhouse stream paths.
- `custom_field_id` (optional, string); Custom Field Id used by Greenhouse stream paths.
- `demographic_answer_id` (optional, string); Demographic Answer Id used by Greenhouse stream paths.
- `demographic_answer_option_id` (optional, string); Demographic Answer Option Id used by Greenhouse
  stream paths.
- `demographic_question_id` (optional, string); Demographic Question Id used by Greenhouse stream
  paths.
- `demographic_question_set_id` (optional, string); Demographic Question Set Id used by Greenhouse
  stream paths.
- `department_id` (optional, string); Department Id used by Greenhouse stream paths.
- `email_template_id` (optional, string); Email Template Id used by Greenhouse stream paths.
- `field_type` (optional, string); Field Type used by Greenhouse stream paths.
- `job_id` (optional, string); Job Id used by Greenhouse stream paths.
- `job_post_id` (optional, string); Job Post Id used by Greenhouse stream paths.
- `job_stage_id` (optional, string); Job Stage Id used by Greenhouse stream paths.
- `mode` (optional, string).
- `offer_id` (optional, string); Offer Id used by Greenhouse stream paths.
- `office_id` (optional, string); Office Id used by Greenhouse stream paths.
- `on_behalf_of_user_id` (optional, string); Optional Greenhouse user ID sent as On-Behalf-Of for
  audited Harvest mutations; omitted when unset.
- `opening_id` (optional, string); Opening Id used by Greenhouse stream paths.
- `page_size` (optional, string); default `100`.
- `prospect_pool_id` (optional, string); Prospect Pool Id used by Greenhouse stream paths.
- `scheduled_interview_id` (optional, string); Scheduled Interview Id used by Greenhouse stream
  paths.
- `scorecard_id` (optional, string); Scorecard Id used by Greenhouse stream paths.
- `token` (optional, string); Token used by Greenhouse stream paths.
- `user_id` (optional, string); User Id used by Greenhouse stream paths.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://harvest.greenhouse.io/v1`, `page_size=100`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/jobs` with query `per_page`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

Pagination by stream: link_header: `candidates`, `applications`, `jobs`, `offers`, `users`,
`activity_feed`, `approvals_for_job`, `pending_approvals_for_user`, `close_reasons`,
`custom_fields`, `custom_field_options`, `demographic_question_sets`, `demographic_questions`,
`demographic_questions_for_demographic_question_set`, `demographic_answer_options`,
`demographic_answer_options_for_demographic_question`, `demographic_answers`,
`demographic_answers_for_application`, `departments`, `degrees`, `disciplines`, `schools`, `eeoc`,
`email_templates`, `job_openings`, `job_posts`, `job_posts_for_job`,
`custom_locations_for_job_post`, `job_stages`, `job_stages_for_job`, `offers_for_application`,
`offices`, `prospect_pools`, `rejection_reasons`, `scheduled_interviews`,
`scheduled_interviews_for_application`, `scorecards`, `scorecards_for_application`, `sources`,
`candidate_tags`, `tags_applied_to_candidate`, `job_permissions`, `future_job_permissions`,
`user_roles`; none: `application`, `approval_flow`, `candidate`, `custom_field`,
`demographic_question_set`, `demographic_question`, `demographic_answer_option`,
`demographic_answer`, `department`, `eeoc_data_for_application`, `email_template`,
`opening_for_job`, `job_post`, `job_post_for_job`, `job_stage`, `job`, `hiring_team`,
`current_offer_for_application`, `offer`, `office`, `prospect_pool`, `scheduled_interview`,
`scorecard`, `tracking_link_data_for_token`, `user`.

- `candidates`: GET `/candidates` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next.
- `applications`: GET `/applications` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next.
- `jobs`: GET `/jobs` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next.
- `offers`: GET `/offers` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next.
- `users`: GET `/users` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next.
- `activity_feed`: GET `/candidates/{{ config.candidate_id }}/activity_feed` - records at response
  root; query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988
  Link headers with rel=next; emits passthrough records.
- `application`: GET `/applications/{{ config.application_id }}` - records at response root; emits
  passthrough records.
- `approvals_for_job`: GET `/jobs/{{ config.job_id }}/approval_flows` - records at response root;
  query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `approval_flow`: GET `/approval_flows/{{ config.approval_flow_id }}` - records at response root;
  emits passthrough records.
- `pending_approvals_for_user`: GET `/users/{{ config.user_id }}/pending_approvals` - records at
  response root; query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `candidate`: GET `/candidates/{{ config.candidate_id }}` - records at response root; emits
  passthrough records.
- `close_reasons`: GET `/close_reasons` - records at response root; query `per_page` from template
  `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `custom_fields`: GET `/custom_fields/{{ config.field_type }}` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `custom_field`: GET `/custom_field/{{ config.custom_field_id }}` - records at response root; emits
  passthrough records.
- `custom_field_options`: GET `/custom_field/{{ config.custom_field_id }}/custom_field_options` -
  records at response root; query `per_page` from template `{{ config.page_size }}`, default `100`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `demographic_question_sets`: GET `/demographics/question_sets` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `demographic_question_set`: GET `/demographics/question_sets/{{ config.demographic_question_set_id
  }}` - records at response root; emits passthrough records.
- `demographic_questions`: GET `/demographics/questions` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `demographic_questions_for_demographic_question_set`: GET `/demographics/question_sets/{{
  config.demographic_question_set_id }}/questions` - records at response root; query `per_page` from
  template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next;
  emits passthrough records.
- `demographic_question`: GET `/demographics/questions/{{ config.demographic_question_id }}` -
  records at response root; emits passthrough records.
- `demographic_answer_options`: GET `/demographics/answer_options` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `demographic_answer_options_for_demographic_question`: GET `/demographics/questions/{{
  config.demographic_question_id }}/answer_options` - records at response root; query `per_page`
  from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `demographic_answer_option`: GET `/demographics/answer_options/{{
  config.demographic_answer_option_id }}` - records at response root; emits passthrough records.
- `demographic_answers`: GET `/demographics/answers` - records at response root; query `per_page`
  from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `demographic_answers_for_application`: GET `/applications/{{ config.application_id
  }}/demographics/answers` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `demographic_answer`: GET `/demographics/answers/{{ config.demographic_answer_id }}` - records at
  response root; emits passthrough records.
- `departments`: GET `/departments` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `department`: GET `/departments/{{ config.department_id }}` - records at response root; emits
  passthrough records.
- `degrees`: GET `/degrees` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `disciplines`: GET `/disciplines` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `schools`: GET `/schools` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `eeoc`: GET `/eeoc` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `eeoc_data_for_application`: GET `/applications/{{ config.application_id }}/eeoc` - records at
  response root; emits passthrough records.
- `email_templates`: GET `/email_templates` - records at response root; query `per_page` from
  template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next;
  emits passthrough records.
- `email_template`: GET `/email_templates/{{ config.email_template_id }}` - records at response
  root; emits passthrough records.
- `job_openings`: GET `/jobs/{{ config.job_id }}/openings` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `opening_for_job`: GET `/jobs/{{ config.job_id }}/openings/{{ config.opening_id }}` - records at
  response root; emits passthrough records.
- `job_posts`: GET `/job_posts` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `job_post`: GET `/job_posts/{{ config.job_post_id }}` - records at response root; emits
  passthrough records.
- `job_posts_for_job`: GET `/jobs/{{ config.job_id }}/job_posts` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `job_post_for_job`: GET `/jobs/{{ config.job_id }}/job_post` - records at response root; emits
  passthrough records.
- `custom_locations_for_job_post`: GET `/job_posts/{{ config.job_post_id }}/custom_locations` -
  records at response root; query `per_page` from template `{{ config.page_size }}`, default `100`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `job_stages`: GET `/job_stages` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `job_stages_for_job`: GET `/jobs/{{ config.job_id }}/stages` - records at response root; query
  `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `job_stage`: GET `/job_stages/{{ config.job_stage_id }}` - records at response root; emits
  passthrough records.
- `job`: GET `/jobs/{{ config.job_id }}` - records at response root; emits passthrough records.
- `hiring_team`: GET `/jobs/{{ config.job_id }}/hiring_team` - records at response root; emits
  passthrough records.
- `offers_for_application`: GET `/applications/{{ config.application_id }}/offers` - records at
  response root; query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `current_offer_for_application`: GET `/applications/{{ config.application_id
  }}/offers/current_offer` - records at response root; emits passthrough records.
- `offer`: GET `/offers/{{ config.offer_id }}` - records at response root; emits passthrough
  records.
- `offices`: GET `/offices` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `office`: GET `/offices/{{ config.office_id }}` - records at response root; emits passthrough
  records.
- `prospect_pools`: GET `/prospect_pools` - records at response root; query `per_page` from template
  `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `prospect_pool`: GET `/prospect_pools/{{ config.prospect_pool_id }}` - records at response root;
  emits passthrough records.
- `rejection_reasons`: GET `/rejection_reasons` - records at response root; query `per_page` from
  template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next;
  emits passthrough records.
- `scheduled_interviews`: GET `/scheduled_interviews` - records at response root; query `per_page`
  from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `scheduled_interviews_for_application`: GET `/applications/{{ config.application_id
  }}/scheduled_interviews` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `scheduled_interview`: GET `/scheduled_interviews/{{ config.scheduled_interview_id }}` - records
  at response root; emits passthrough records.
- `scorecards`: GET `/scorecards` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `scorecards_for_application`: GET `/applications/{{ config.application_id }}/scorecards` - records
  at response root; query `per_page` from template `{{ config.page_size }}`, default `100`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `scorecard`: GET `/scorecards/{{ config.scorecard_id }}` - records at response root; emits
  passthrough records.
- `sources`: GET `/sources` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `candidate_tags`: GET `/tags/candidate` - records at response root; query `per_page` from template
  `{{ config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `tags_applied_to_candidate`: GET `/candidates/{{ config.candidate_id }}/tags` - records at
  response root; query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `tracking_link_data_for_token`: GET `/tracking_links/{{ config.token }}` - records at response
  root; emits passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - records at response root; emits passthrough records.
- `job_permissions`: GET `/users/{{ config.user_id }}/permissions/jobs` - records at response root;
  query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `future_job_permissions`: GET `/users/{{ config.user_id }}/permissions/future_jobs` - records at
  response root; query `per_page` from template `{{ config.page_size }}`, default `100`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `user_roles`: GET `/user_roles` - records at response root; query `per_page` from template `{{
  config.page_size }}`, default `100`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.

## Write actions & risks

Overall write risk: external Greenhouse Harvest API mutations including candidate, application, job,
office, department, tag, and user changes.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `delete_application`: DELETE `/applications/{{ record.application_id }}` - kind `delete`; body
  type `none`; path fields `application_id`; required record fields `application_id`; accepted
  fields `application_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation:
  DELETE: Delete Application.
- `add_application_to_candidate_prospect`: POST `/candidates/{{ record.candidate_id }}/applications`
  - kind `create`; body type `json`; path fields `candidate_id`; required record fields
  `candidate_id`; accepted fields `candidate_id`; risk: Greenhouse mutation: POST: Add Application
  to Candidate/Prospect.
- `update_application`: PATCH `/applications/{{ record.application_id }}` - kind `update`; body type
  `json`; path fields `application_id`; required record fields `application_id`; accepted fields
  `application_id`; risk: Greenhouse mutation: PATCH: Update Application.
- `advance_application`: POST `/applications/{{ record.application_id }}/advance` - kind `create`;
  body type `json`; path fields `application_id`; required record fields `application_id`; accepted
  fields `application_id`; risk: Greenhouse mutation: POST: Advance Application.
- `move_application_different_job`: POST `/applications/{{ record.application_id }}/transfer_to_job`
  - kind `create`; body type `json`; path fields `application_id`; required record fields
  `application_id`; accepted fields `application_id`; risk: Greenhouse mutation: POST: Move
  Application (Different Job).
- `move_application_same_job`: POST `/applications/{{ record.application_id }}/move` - kind
  `create`; body type `json`; path fields `application_id`; required record fields `application_id`;
  accepted fields `application_id`; risk: Greenhouse mutation: POST: Move Application (Same Job).
- `convert_prospect_to_candidate`: PATCH `/applications/{{ record.application_id
  }}/convert_prospect` - kind `update`; body type `json`; path fields `application_id`; required
  record fields `application_id`; accepted fields `application_id`; risk: Greenhouse mutation:
  PATCH: Convert Prospect To Candidate.
- `add_attachment_to_application`: POST `/applications/{{ record.application_id }}/attachments` -
  kind `create`; body type `json`; path fields `application_id`; required record fields
  `application_id`; accepted fields `application_id`; risk: Greenhouse mutation: POST: Add
  Attachment to Application.
- `hire_application`: POST `/applications/{{ record.application_id }}/hire` - kind `create`; body
  type `json`; path fields `application_id`; required record fields `application_id`; accepted
  fields `application_id`; risk: Greenhouse mutation: POST: Hire Application.
- `reject_application`: POST `/applications/{{ record.application_id }}/reject` - kind `create`;
  body type `json`; path fields `application_id`; required record fields `application_id`; accepted
  fields `application_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation: POST:
  Reject Application.
- `update_rejection_reason`: PATCH `/applications/{{ record.application_id }}/reject` - kind
  `update`; body type `json`; path fields `application_id`; required record fields `application_id`;
  accepted fields `application_id`; risk: Greenhouse mutation: PATCH: Update Rejection Reason.
- `unreject_application`: POST `/applications/{{ record.application_id }}/unreject` - kind `create`;
  body type `json`; path fields `application_id`; required record fields `application_id`; accepted
  fields `application_id`; risk: Greenhouse mutation: POST: Unreject Application.
- `request_approvals`: POST `/approval_flows/{{ record.approval_flow_id }}/request_approvals` - kind
  `create`; body type `json`; path fields `approval_flow_id`; required record fields
  `approval_flow_id`; accepted fields `approval_flow_id`; risk: Greenhouse mutation: POST: Request
  Approvals.
- `replace_an_approver_in_an_approver_group`: PUT `/approver_groups/{{ record.approver_group_id
  }}/replace_approvers` - kind `update`; body type `json`; path fields `approver_group_id`; required
  record fields `approver_group_id`; accepted fields `approver_group_id`; risk: Greenhouse mutation:
  PUT: Replace an approver in an approver group.
- `create_or_replace_an_approval_flow`: PUT `/jobs/{{ record.job_id }}/approval_flows` - kind
  `update`; body type `json`; path fields `job_id`; required record fields `job_id`; accepted fields
  `job_id`; risk: Greenhouse mutation: PUT: Create or replace an approval flow.
- `delete_candidate`: DELETE `/candidates/{{ record.candidate_id }}` - kind `delete`; body type
  `none`; path fields `candidate_id`; required record fields `candidate_id`; accepted fields
  `candidate_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation: DELETE: Delete
  Candidate.
- `edit_candidate`: PATCH `/candidates/{{ record.candidate_id }}` - kind `update`; body type `json`;
  path fields `candidate_id`; required record fields `candidate_id`; accepted fields `candidate_id`;
  risk: Greenhouse mutation: PATCH: Edit Candidate.
- `add_attachment`: POST `/candidates/{{ record.candidate_id }}/attachments` - kind `create`; body
  type `json`; path fields `candidate_id`; required record fields `candidate_id`; accepted fields
  `candidate_id`; risk: Greenhouse mutation: POST: Add Attachment.
- `add_candidate`: POST `/candidates` - kind `create`; body type `json`; risk: Greenhouse mutation:
  POST: Add Candidate.
- `add_note`: POST `/candidates/{{ record.candidate_id }}/activity_feed/notes` - kind `create`; body
  type `json`; path fields `candidate_id`; required record fields `candidate_id`; accepted fields
  `candidate_id`; risk: Greenhouse mutation: POST: Add Note.
- `add_e_mail_note`: POST `/candidates/{{ record.candidate_id }}/activity_feed/emails` - kind
  `create`; body type `json`; path fields `candidate_id`; required record fields `candidate_id`;
  accepted fields `candidate_id`; risk: Greenhouse mutation: POST: Add E-mail Note.
- `add_education`: POST `/candidates/{{ record.candidate_id }}/educations` - kind `create`; body
  type `json`; path fields `candidate_id`; required record fields `candidate_id`; accepted fields
  `candidate_id`; risk: Greenhouse mutation: POST: Add Education.
- `remove_education_from_candidate`: DELETE `/candidates/{{ record.candidate_id }}/educations/{{
  record.education_id }}` - kind `delete`; body type `none`; path fields `candidate_id`,
  `education_id`; required record fields `candidate_id`, `education_id`; accepted fields
  `candidate_id`, `education_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation:
  DELETE: Remove Education From Candidate.
- `add_employment`: POST `/candidates/{{ record.candidate_id }}/employments` - kind `create`; body
  type `json`; path fields `candidate_id`; required record fields `candidate_id`; accepted fields
  `candidate_id`; risk: Greenhouse mutation: POST: Add Employment.
- `remove_employment_from_candidate`: DELETE `/candidates/{{ record.candidate_id }}/employments/{{
  record.employment_id }}` - kind `delete`; body type `none`; path fields `candidate_id`,
  `employment_id`; required record fields `candidate_id`, `employment_id`; accepted fields
  `candidate_id`, `employment_id`; confirmation `destructive`; risk: Destructive Greenhouse
  mutation: DELETE: Remove Employment From Candidate.
- `add_prospect`: POST `/prospects` - kind `create`; body type `json`; risk: Greenhouse mutation:
  POST: Add Prospect.
- `anonymize_candidate`: PUT `/candidates/{{ record.candidate_id }}/anonymize?fields={{
  record.field_names }}` - kind `update`; body type `none`; path fields `candidate_id`,
  `field_names`; required record fields `candidate_id`, `field_names`; accepted fields
  `candidate_id`, `field_names`; confirmation `destructive`; risk: Destructive Greenhouse mutation:
  PUT: Anonymize Candidate.
- `merge_candidates`: PUT `/candidates/merge` - kind `update`; body type `json`; confirmation
  `destructive`; risk: Destructive Greenhouse mutation: PUT: Merge Candidates.
- `create_custom_field`: POST `/custom_fields` - kind `create`; body type `json`; risk: Greenhouse
  mutation: POST: Create Custom Field.
- `update_custom_field`: PATCH `/custom_fields/{{ record.custom_field_id }}` - kind `update`; body
  type `json`; path fields `custom_field_id`; required record fields `custom_field_id`; accepted
  fields `custom_field_id`; risk: Greenhouse mutation: PATCH: Update Custom Field.
- `delete_custom_field`: DELETE `/custom_fields/{{ record.custom_field_id }}` - kind `delete`; body
  type `none`; path fields `custom_field_id`; required record fields `custom_field_id`; accepted
  fields `custom_field_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation:
  DELETE: Delete Custom Field.
- `create_custom_field_options`: POST `/custom_field/{{ record.custom_field_id
  }}/custom_field_options` - kind `create`; body type `json`; path fields `custom_field_id`;
  required record fields `custom_field_id`; accepted fields `custom_field_id`; risk: Greenhouse
  mutation: POST: Create Custom Field Options.
- `update_custom_field_options`: PATCH `/custom_field/{{ record.custom_field_id
  }}/custom_field_options` - kind `update`; body type `json`; path fields `custom_field_id`;
  required record fields `custom_field_id`; accepted fields `custom_field_id`; risk: Greenhouse
  mutation: PATCH: Update Custom Field Options.
- `remove_custom_field_options`: DELETE `/custom_field/{{ record.custom_field_id
  }}/custom_field_options` - kind `delete`; body type `json`; path fields `custom_field_id`;
  required record fields `custom_field_id`; accepted fields `custom_field_id`; confirmation
  `destructive`; risk: Destructive Greenhouse mutation: DELETE: Remove Custom Field Options.
- `edit_department`: PATCH `/departments/{{ record.department_id }}` - kind `update`; body type
  `json`; path fields `department_id`; required record fields `department_id`; accepted fields
  `department_id`; risk: Greenhouse mutation: PATCH: Edit Department.
- `add_department`: POST `/departments` - kind `create`; body type `json`; risk: Greenhouse
  mutation: POST: Add Department.
- `edit_openings`: PATCH `/jobs/{{ record.job_id }}/openings/{{ record.opening_id }}` - kind
  `update`; body type `json`; path fields `job_id`, `opening_id`; required record fields `job_id`,
  `opening_id`; accepted fields `job_id`, `opening_id`; risk: Greenhouse mutation: PATCH: Edit
  Openings.
- `create_new_openings`: POST `/jobs/{{ record.job_id }}/openings` - kind `create`; body type
  `json`; path fields `job_id`; required record fields `job_id`; accepted fields `job_id`; risk:
  Greenhouse mutation: POST: Create New Openings.
- `update_job`: PATCH `/jobs/{{ record.job_id }}` - kind `update`; body type `json`; path fields
  `job_id`; required record fields `job_id`; accepted fields `job_id`; risk: Greenhouse mutation:
  PATCH: Update Job.
- `create_job`: POST `/jobs` - kind `create`; body type `json`; risk: Greenhouse mutation: POST:
  Create Job.
- `replace_hiring_team`: PUT `/jobs/{{ record.job_id }}/hiring_team` - kind `update`; body type
  `json`; path fields `job_id`; required record fields `job_id`; accepted fields `job_id`; risk:
  Greenhouse mutation: PUT: Replace Hiring Team.
- `add_hiring_team_members`: POST `/jobs/{{ record.job_id }}/hiring_team` - kind `create`; body type
  `json`; path fields `job_id`; required record fields `job_id`; accepted fields `job_id`; risk:
  Greenhouse mutation: POST: Add Hiring Team Members.
- `remove_hiring_team_member`: DELETE `/jobs/{{ record.job_id }}/hiring_team` - kind `delete`; body
  type `json`; path fields `job_id`; required record fields `job_id`; accepted fields `job_id`;
  confirmation `destructive`; risk: Destructive Greenhouse mutation: DELETE: Remove Hiring Team
  Member.
- `update_current_offer`: PATCH `/applications/{{ record.application_id }}/offers/current_offer` -
  kind `update`; body type `json`; path fields `application_id`; required record fields
  `application_id`; accepted fields `application_id`; risk: Greenhouse mutation: PATCH: Update
  Current Offer.
- `edit_office`: PATCH `/offices/{{ record.office_id }}` - kind `update`; body type `json`; path
  fields `office_id`; required record fields `office_id`; accepted fields `office_id`; risk:
  Greenhouse mutation: PATCH: Edit Office.
- `add_office`: POST `/offices` - kind `create`; body type `json`; risk: Greenhouse mutation: POST:
  Add Office.
- `remove_scheduled_interview`: DELETE `/scheduled_interviews/{{ record.scheduled_interview_id }}` -
  kind `delete`; body type `none`; path fields `scheduled_interview_id`; required record fields
  `scheduled_interview_id`; accepted fields `scheduled_interview_id`; confirmation `destructive`;
  risk: Destructive Greenhouse mutation: Delete: Remove Scheduled Interview.
- `add_candidate_tag`: POST `/tags/candidate` - kind `create`; body type `json`; risk: Greenhouse
  mutation: POST: Add New Candidate Tag.
- `remove_tag_from_candidate`: DELETE `/candidates/{{ record.candidate_id }}/tags/{{ record.tag_id
  }}` - kind `delete`; body type `none`; path fields `candidate_id`, `tag_id`; required record
  fields `candidate_id`, `tag_id`; accepted fields `candidate_id`, `tag_id`; confirmation
  `destructive`; risk: Destructive Greenhouse mutation: DELETE: Remove tag from candidate.
- `add_a_candidate_tag`: PUT `/candidates/{{ record.candidate_id }}/tags/{{ record.tag_id }}` - kind
  `update`; body type `json`; path fields `candidate_id`, `tag_id`; required record fields
  `candidate_id`, `tag_id`; accepted fields `candidate_id`, `tag_id`; risk: Greenhouse mutation:
  PUT: Add a candidate tag.
- `change_user_permission_level`: PATCH `/users/permission_level` - kind `update`; body type `json`;
  risk: Greenhouse mutation: PATCH: Change user permission level.
- `add_user`: POST `/users` - kind `create`; body type `json`; risk: Greenhouse mutation: POST: Add
  User.
- `add_e_mail_address_to_user`: POST `/users/{{ record.user_id }}/email_addresses` - kind `create`;
  body type `json`; path fields `user_id`; required record fields `user_id`; accepted fields
  `user_id`; risk: Greenhouse mutation: POST: Add E-mail Address To User.
- `remove_a_job_permission`: DELETE `/users/{{ record.user_id }}/permissions/jobs` - kind `delete`;
  body type `json`; path fields `user_id`; required record fields `user_id`; accepted fields
  `user_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation: DELETE: Remove a Job
  Permission.
- `add_a_job_permission`: PUT `/users/{{ record.user_id }}/permissions/jobs` - kind `update`; body
  type `json`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`;
  risk: Greenhouse mutation: PUT: Add a Job Permission.
- `remove_a_future_job_permission`: DELETE `/users/{{ record.user_id }}/permissions/future_jobs` -
  kind `delete`; body type `json`; path fields `user_id`; required record fields `user_id`; accepted
  fields `user_id`; confirmation `destructive`; risk: Destructive Greenhouse mutation: DELETE:
  Remove a Future Job Permission.
- `add_a_future_job_permission`: PUT `/users/{{ record.user_id }}/permissions/future_jobs` - kind
  `update`; body type `json`; path fields `user_id`; required record fields `user_id`; accepted
  fields `user_id`; risk: Greenhouse mutation: PUT: Add a Future Job Permission.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 69 stream-backed endpoint group(s), 57 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  deprecated=3.
