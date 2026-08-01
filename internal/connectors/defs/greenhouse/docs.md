# Overview

Greenhouse reads and writes documented Harvest API resources through the declarative connector engine. The operation ledger was refreshed from the official Harvest HTML reference on 2026-08-01 and records 135 documented operations exactly once: 69 fixture-backed streams (including the activity-feed changefeed surface), 64 fixture-backed typed write actions, and 2 blocked binary attachment upload operations. No generic HTTP method/path/body, raw query, shell, file, or passthrough escape hatch is exposed.

Readable streams: `candidates`, `applications`, `jobs`, `offers`, `users`, `activity_feed`, `application`, `approvals_for_job`, `approval_flow`, `pending_approvals_for_user`, `candidate`, `close_reasons`, `custom_fields`, `custom_field`, `custom_field_options`, `demographic_question_sets`, `demographic_question_set`, `demographic_questions`, `demographic_questions_for_demographic_question_set`, `demographic_question`, `demographic_answer_options`, `demographic_answer_options_for_demographic_question`, `demographic_answer_option`, `demographic_answers`, `demographic_answers_for_application`, `demographic_answer`, `departments`, `department`, `degrees`, `disciplines`, `schools`, `eeoc`, `eeoc_data_for_application`, `email_templates`, `email_template`, `job_openings`, `opening_for_job`, `job_posts`, `job_post`, `job_posts_for_job`, `job_post_for_job`, `custom_locations_for_job_post`, `job_stages`, `job_stages_for_job`, `job_stage`, `job`, `hiring_team`, `offers_for_application`, `current_offer_for_application`, `offer`, `offices`, `office`, `prospect_pools`, `prospect_pool`, `rejection_reasons`, `scheduled_interviews`, `scheduled_interviews_for_application`, `scheduled_interview`, `scorecards`, `scorecards_for_application`, `scorecard`, `sources`, `candidate_tags`, `tags_applied_to_candidate`, `tracking_link_data_for_token`, `user`, `job_permissions`, `future_job_permissions`, `user_roles`.

Write actions: `delete_application`, `add_application_to_candidate_prospect`, `update_application`, `advance_application`, `move_application_different_job`, `move_application_same_job`, `convert_prospect_to_candidate`, `hire_application`, `reject_application`, `update_rejection_reason`, `unreject_application`, `request_approvals`, `replace_an_approver_in_an_approver_group`, `create_or_replace_an_approval_flow`, `delete_candidate`, `edit_candidate`, `add_candidate`, `add_note`, `add_e_mail_note`, `add_education`, `remove_education_from_candidate`, `add_employment`, `remove_employment_from_candidate`, `add_prospect`, `anonymize_candidate`, `merge_candidates`, `create_custom_field`, `update_custom_field`, `delete_custom_field`, `create_custom_field_options`, `update_custom_field_options`, `remove_custom_field_options`, `edit_department`, `add_department`, `edit_openings`, `create_new_openings`, `update_job`, `create_job`, `replace_hiring_team`, `add_hiring_team_members`, `remove_hiring_team_member`, `update_current_offer`, `edit_office`, `add_office`, `remove_scheduled_interview`, `add_candidate_tag`, `remove_tag_from_candidate`, `add_a_candidate_tag`, `change_user_permission_level`, `add_user`, `add_e_mail_address_to_user`, `remove_a_job_permission`, `add_a_job_permission`, `remove_a_future_job_permission`, `add_a_future_job_permission`, `destroy_candidate_tag`, `destroy_openings`, `create_scheduled_interview`, `update_scheduled_interview`, `update_job_post`, `update_job_post_status`, `edit_user_v2`, `disable_user_v2`, `enable_user_v2`.

Service API documentation: https://developers.greenhouse.io/harvest.html.

## Auth setup

Connection fields:

- `api_key` (required, secret, string): Greenhouse Harvest API token. The connector sends it as the HTTP Basic username with a blank password and never logs it.
- `base_url` (optional, string, default `https://harvest.greenhouse.io`): Harvest API origin override for tests/proxies. Bundle paths include `/v1` or `/v2` explicitly.
- `page_size` (optional, string, default `100`): records per page for list streams.
- Resource id config fields such as `candidate_id`, `application_id`, `job_id`, `job_post_id`, `scheduled_interview_id`, and `user_id` are used only by fixed stream paths.
- `on_behalf_of_user_id` (optional, string): sent as `On-Behalf-Of` when configured for audited Harvest requests.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior: HTTP Basic authentication using `secrets.api_key`.

Connection checks call GET `/v1/jobs` with query `per_page=1`.

## Streams notes

List streams use Greenhouse's RFC 5988 Link-header pagination with bounded `per_page` from `page_size`. Detail streams use fixed typed paths and no pagination. The activity-feed stream `/v1/candidates/{{ config.candidate_id }}/activity_feed` is represented as a fixture-backed read stream and is counted as the documented changefeed surface; the connector does not claim CDC certification.

All stream fixtures are sanitized JSON replay fixtures and make no live provider calls. Stream paths are fixed in `streams.json`; callers cannot provide arbitrary provider paths or raw query strings.

## Write actions & risks

Reverse ETL writes remain governed by plan -> preview -> explicit approval -> execute. Each declared action has a fixed method/path, a record schema, sanitized request-shape fixture, and risk text. Destructive/admin actions add `confirm: destructive` and closed schemas so extra arbitrary request fields are rejected.

Destructive/admin-confirmed actions: `delete_application`, `reject_application`, `delete_candidate`, `remove_education_from_candidate`, `remove_employment_from_candidate`, `anonymize_candidate`, `merge_candidates`, `delete_custom_field`, `remove_custom_field_options`, `remove_hiring_team_member`, `remove_scheduled_interview`, `remove_tag_from_candidate`, `remove_a_job_permission`, `remove_a_future_job_permission`, `destroy_candidate_tag`, `destroy_openings`, `disable_user_v2`.

Current v2 parity is implemented for Greenhouse job post updates, scheduled interview create/update replacements, user edit/disable/enable, and job-opening destroy operations where the official docs moved behavior from deprecated v1 endpoints to v2. Hiring-team write paths use the official `/v1/jobs/{id}` write endpoints; the read-only hiring-team stream remains `/v1/jobs/{id}/hiring_team`.

Binary attachment upload operations are intentionally not exposed as writes; see Known limits.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- Certification: fixture-backed only; certified/live-safe count is `0`. No live Greenhouse call or credentialed check is part of this bundle evidence.
- Blocked operations:
- POST `/v1/applications/{application_id}/attachments` — blocked (sensitive_reverse_etl, high): Official Greenhouse attachment upload accepts base64 file content or a machine-accessible URL. The existing JSON write contract cannot enforce byte bounds on content or safely dereference external URLs, so no attachment upload action is exposed until a bounded binary upload contract with redaction is authored.
- POST `/v1/candidates/{candidate_id}/attachments` — blocked (sensitive_reverse_etl, high): Official Greenhouse attachment upload accepts base64 file content or a machine-accessible URL. The existing JSON write contract cannot enforce byte bounds on content or safely dereference external URLs, so no attachment upload action is exposed until a bounded binary upload contract with redaction is authored.
- The blocked attachment rows are documented by Greenhouse as JSON requests carrying base64 `content` or a machine-accessible `url`. The current write contract cannot enforce byte bounds on base64 content or safely dereference arbitrary external URLs, so these remain blocked until a bounded binary upload contract with redaction is authored.
