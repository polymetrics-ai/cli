# Overview

Pylon is a declarative HTTP bundle for the documented Pylon REST API. It keeps the legacy streams (`issues`, `accounts`, `contacts`, `users`, and legacy `/messages`) and adds the remaining documented GET endpoints from the official OpenAPI as streams.

## Auth setup

Provide `api_token` as a secret. Requests use `Authorization: Bearer <token>`. `base_url` defaults to `https://api.usepylon.com`.

## Streams notes

Legacy streams preserve the existing schema-projection behavior and cursor token path `pagination.next_cursor`. New OpenAPI-backed streams use passthrough projection with records extracted from `data`; cursor-paginated OpenAPI streams use `pagination.cursor` with `pagination.has_next_page` as the stop signal. Path-scoped streams use config values such as `account_id`, `issue_id`, `knowledge_base_id`, `article_id`, and `custom_object_type`.

Streams covered: issues, accounts, contacts, users, messages, account_relationships, account, activity_types, audit_logs, call_recording, contact, custom_fields, custom_field, custom_objects, custom_object, feature_request, issue_statuses, issue, issue_followers, issue_messages, issue_threads, issue_voice_calls, knowledge_bases, knowledge_base, articles, article, collections, collection, macro_groups, macros, macro, me, milestone, project, surveys, survey, survey_responses, tags, tag, tasks, task, task_comments, teams, team, ticket_forms, ticket_form, training_data, training_data_detail, user_roles, user.

## Write actions & risks

`writes.json` contains concrete actions for every documented JSON or path-only POST/PATCH/DELETE operation. Delete, redact, and merge-style actions are marked destructive. Multipart upload endpoints are excluded because the current dialect does not support multipart/form-data file uploads.

## Known limits

The legacy `/messages` stream is retained although current Pylon OpenAPI documents messages under `/issues/{id}/messages`. Existing legacy record-shape limitations remain: the legacy `raw` nested field copies the entire source item, which is not expressible in the current schema-projection dialect.
