# Overview

Zonka Feedback is a declarative-HTTP connector for the Zonka Feedback REST API v2.1. It reads responses, surveys, contacts, workspaces, survey stats, devices, tasks, locations, users, and distribution logs. It also exposes typed write actions for response ingestion, contact upsert, survey distribution, and task management. The legacy Go connector remains registered until the wave6 registry cutover.

## Auth setup

Generate an Auth Token in Company Settings > Developers > API, then provide it as the `auth_token` secret. `access_token` remains supported as a fallback secret name for compatibility with earlier bundle revisions. Current Zonka docs and the published Postman collection send the token in the `Z-API-TOKEN` header.

Set `base_url` to the account datacenter URL. Zonka documents US, EU, India, and Australia hosts (`https://us1.apis.zonkafeedback.com`, `https://e.apis.zonkafeedback.com`, `https://in.apis.zonkafeedback.com`, and `https://au.apis.zonkafeedback.com`). The default is the US host.

## Streams notes

The original three streams remain legacy-parity streams: `responses` reads `GET /responses` from the `responses` envelope, `surveys` reads `GET /surveys` from the `surveys` envelope, and `contacts` reads `GET /contacts` from the `contacts` envelope. They retain the legacy static `page`/`per_page` pagination shape and the legacy four-field schema projection.

Pass B adds the current v2.1 streams from the Zonka Developer Guide and Postman collection: `response_details`, `workspaces`, `survey_stats`, `survey_details`, `contact_details`, `contact_segments`, `devices`, `device_details`, `device_uptime`, `device_responses`, `tasks`, `locations`, `location_details`, `users`, `user_details`, and `distribution_logs`. Detail and stats streams use explicit config IDs or date filters when the API requires them.

## Write actions & risks

Write actions are typed, one HTTP request per approved record: `add_response`, `update_response`, `upsert_contacts`, `send_email_survey`, `send_sms_survey`, `send_two_way_sms_survey`, `send_whatsapp_survey`, `add_task`, `update_task`, and `delete_tasks`. These actions can create customer-visible responses, send survey invitations, mutate contact data, and create/update/delete tasks, so reverse ETL must follow plan, preview, approval, execute.

## Known limits

- Legacy `id`/`name`/`updated_at` alternate-key backfill remains unmodeled for the original three streams. The declarative engine has no coalesce filter, so the bundle keeps the already-approved schema projection and documents the narrowing.
- Legacy runtime-configurable `page_size`, `max_pages`, and `mode=fixture` are not modeled because the engine pagination fields are static bundle fields. The bundle keeps the legacy default page size for the three parity streams and uses fixture replay for credential-free tests.
- Zonka's Postman collection shows detail endpoint URLs with trailing slashes and parameter names in prose. This bundle models those as REST path parameters such as `/responses/{responseId}` and `/users/{userId}`.
- The current docs state that many filters accept arrays. The engine query dialect sends string query values, so optional list filters are not declared globally; callers can still use the broad list streams and the typed detail/stat streams.
