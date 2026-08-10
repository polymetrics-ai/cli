---
name: pm-coassemble
description: Coassemble connector knowledge and safe action guide.
---

# pm-coassemble

## Purpose

Reads Coassemble courses, screen types, collections, clients, users, learner tracking, and translations, and writes course/collection/client/user/translation lifecycle actions, through the Coassemble headless REST API.

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
- mode
- user_id (secret)
- user_token (secret)

## ETL Streams

- courses:
  - primary key: id
  - fields: active(boolean), description(string), id(integer), identified(boolean), image(string), is_sharable(boolean), key(string), paid(boolean), private(boolean), title(string)
- screen_types:
  - fields: icon(string), id(integer), name(string), premium(boolean), title(string)
- trackings:
  - fields: completed(boolean), course_id(integer), id(integer), identifier(string), progress(number), status(string)
- collections:
  - primary key: id
  - fields: active(boolean), clientIdentifier(string), created(string), deleted(boolean), description(string), id(integer), identifier(string), key(string), title(string), updated(string)
- clients:
  - primary key: clientIdentifier
  - fields: clientIdentifier(string), created(string), updated(string), userCount(integer)
- users:
  - primary key: identifier
  - fields: avatar(string), clientIdentifier(string), created(string), identifier(string), name(string), testMode(boolean), updated(string)
- user_trackings:
  - primary key: identifier
  - fields: avatar(string), clientIdentifier(string), identifier(string), name(string), totals(object), trackings(array)
- collection_trackings:
  - primary key: id
  - fields: collection_id(string), commenced(string), completed(string), courses(array), id(integer), identifier(string), name(string), progressPercent(number), totalTime(number)
- translations:
  - primary key: id, language
  - fields: course_id(string), id(integer), language(string), missingScreens(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- publish_course:
  - endpoint: POST /api/v1/headless/course/{{ record.id }}/publish
  - required fields: id
  - risk: publishes the current draft of a course, making it live for learners; no approval required
- duplicate_course:
  - endpoint: POST /api/v1/headless/course/{{ record.id }}/duplicate
  - required fields: id
  - risk: creates a full copy of an existing course; low-risk external mutation, no approval required
- delete_course:
  - endpoint: DELETE /api/v1/headless/course/{{ record.id }}
  - required fields: id
  - risk: soft-deletes a course (recoverable via restore_course within Coassemble's retention window); approval required
- restore_course:
  - endpoint: POST /api/v1/headless/course/{{ record.id }}/restore
  - required fields: id
  - risk: restores a previously soft-deleted course; no approval required
- delete_tracking:
  - endpoint: DELETE /api/v1/headless/tracking
  - required fields: id, identifier
  - risk: permanently erases one learner's tracking/progress record for a course; irreversible, approval required
- create_collection:
  - endpoint: POST /api/v1/headless/collection
  - required fields: title
  - risk: creates a new collection of courses; low-risk external mutation, no approval required
- delete_collection:
  - endpoint: DELETE /api/v1/headless/collection/{{ record.id }}
  - required fields: id
  - risk: soft-deletes a collection (recoverable via restore_collection); approval required
- restore_collection:
  - endpoint: POST /api/v1/headless/collection/{{ record.id }}/restore
  - required fields: id
  - risk: restores a previously soft-deleted collection; no approval required
- update_client:
  - endpoint: PUT /api/v1/headless/client/{{ record.clientIdentifier }}
  - required fields: clientIdentifier
  - optional fields: metadata
  - risk: overwrites a client's arbitrary metadata bag; no approval required
- delete_client:
  - endpoint: DELETE /api/v1/headless/client/{{ record.clientIdentifier }}
  - required fields: clientIdentifier
  - risk: irreversibly removes a client (multi-tenant sub-account) and its documented on-delete effects on associated users; approval required
- update_user:
  - endpoint: PUT /api/v1/headless/user/{{ record.identifier }}
  - required fields: identifier
  - optional fields: clientIdentifier, metadata, name, avatar
  - risk: overwrites a learner's profile fields (name/avatar/metadata) or reassigns their client; no approval required
- delete_user:
  - endpoint: DELETE /api/v1/headless/user/{{ record.identifier }}
  - required fields: identifier
  - risk: irreversibly removes a learner identity, applying Coassemble's server-side DEFAULT handling for that identity's course progress (the real endpoint also accepts optional action=reallocate|delete|ignore/reallocateTo/clientIdentifier query params to control that handling explicitly, and Coassemble's own docs do not fully specify their exact semantics beyond "choose what to do with any courses associated with this identifier" — this action deliberately does not expose them, since the write-action path/query dialect has no way to send an optional record field only when present, and silently defaulting an ambiguous, irreversible per-learner-data-retention choice would be worse than declaring it out of scope; approval required
- translate_course:
  - endpoint: POST /api/v1/headless/translation/translate/{{ record.course_id }}
  - required fields: course_id, language
  - risk: kicks off machine translation of a course into a new BCP-47 language variant; low-risk external mutation, no approval required
- set_default_translation:
  - endpoint: POST /api/v1/headless/translation/default/{{ record.course_id }}/{{ record.language }}
  - required fields: course_id, language
  - risk: changes which language variant learners see by default for this course; no approval required
- sync_translation:
  - endpoint: POST /api/v1/headless/translation/sync/{{ record.course_id }}/{{ record.language }}
  - required fields: course_id, language
  - risk: re-syncs a translated variant's content with upstream changes to the source-language course, which can overwrite manual edits made directly in the translated variant; no approval required
- delete_translation:
  - endpoint: DELETE /api/v1/headless/translation/{{ record.course_id }}/{{ record.language }}
  - required fields: course_id, language
  - risk: permanently removes a language variant of a course; irreversible, approval required

## Security

- read risk: external Coassemble headless API read of course, screen type, collection, client, user, tracking, and translation data
- write risk: external mutation of Coassemble courses, collections, clients, users, and translations (publish/duplicate/restore/delete a course; delete a tracking record; create/delete/restore a collection; update/delete a client; update/delete a user; translate/set-default/sync/delete a course translation)
- approval: publish/duplicate/restore/create/update actions: none; delete_course/delete_collection/delete_client/delete_user/delete_tracking/delete_translation: approval required (irreversible or high-blast-radius)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Coassemble's declared streams and reverse-ETL actions.
- Usage: pm coassemble <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete v1 headless client identifier - Documented DELETE /v1/headless/client/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-client-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless collection id - Documented DELETE /v1/headless/collection/{id} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-collection-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless course id - Documented DELETE /v1/headless/course/{id} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-course-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless themes id - Documented DELETE /v1/headless/themes/{id} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-themes-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless tracking - Documented DELETE /v1/headless/tracking (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-tracking]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless translation id language - Documented DELETE /v1/headless/translation/{id}/{language} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-translation-id-language]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless user identifier - Documented DELETE /v1/headless/user/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-user-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v1 headless webhooks id - Documented DELETE /v1/headless/webhooks/{id} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.delete.v1-headless-webhooks-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get api v1 headless client identifier - Documented GET /api/v1/headless/client/{identifier} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-client-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless collection tracking id - Documented GET /api/v1/headless/collection/tracking/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-collection-tracking-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless collections id - Documented GET /api/v1/headless/collections/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-collections-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless course action - Documented GET /api/v1/headless/course/{action} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-course-action]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless course content id - Documented GET /api/v1/headless/course/content/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-course-content-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless course scorm id - Documented GET /api/v1/headless/course/scorm/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-course-scorm-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless courses id - Documented GET /api/v1/headless/courses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-courses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless screen trackings - Documented GET /api/v1/headless/screen/trackings (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-screen-trackings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless tracking id - Documented GET /api/v1/headless/tracking/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-tracking-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get api v1 headless user identifier - Documented GET /api/v1/headless/user/{identifier} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.api-v1-headless-user-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless client identifier - Documented GET /v1/headless/client/{identifier} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-client-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless clients - Documented GET /v1/headless/clients (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-clients]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless collection tracking id - Documented GET /v1/headless/collection/tracking/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-collection-tracking-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless collection trackings - Documented GET /v1/headless/collection/trackings (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-collection-trackings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless collections - Documented GET /v1/headless/collections (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-collections]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless collections id - Documented GET /v1/headless/collections/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-collections-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless course action - Documented GET /v1/headless/course/{action} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-course-action]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless course content id - Documented GET /v1/headless/course/content/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-course-content-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless course scorm id - Documented GET /v1/headless/course/scorm/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-course-scorm-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless courses - Documented GET /v1/headless/courses (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-courses]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless courses id - Documented GET /v1/headless/courses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-courses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless screen trackings - Documented GET /v1/headless/screen/trackings (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-screen-trackings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless themes - Documented GET /v1/headless/themes (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-themes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless themes id - Documented GET /v1/headless/themes/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-themes-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless tracking id - Documented GET /v1/headless/tracking/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-tracking-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless trackings - Documented GET /v1/headless/trackings (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-trackings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless translations id - Documented GET /v1/headless/translations/{id} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-translations-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless usage allowances - Documented GET /v1/headless/usage/allowances (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-usage-allowances]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless usage client identifier - Documented GET /v1/headless/usage/client/{identifier} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-usage-client-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless usage clients - Documented GET /v1/headless/usage/clients (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-usage-clients]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless user identifier - Documented GET /v1/headless/user/{identifier} (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-user-identifier]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless user trackings - Documented GET /v1/headless/user/trackings (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-user-trackings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless users - Documented GET /v1/headless/users (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v1 headless webhooks - Documented GET /v1/headless/webhooks (not implemented) [intent=direct_read availability=not_implemented operation=coassemble.get.v1-headless-webhooks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post api v1 headless course generate - Documented POST /api/v1/headless/course/generate (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.api-v1-headless-course-generate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 headless course id revert - Documented POST /api/v1/headless/course/{id}/revert (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.api-v1-headless-course-id-revert]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post api v1 headless course url - Documented POST /api/v1/headless/course/url (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.api-v1-headless-course-url]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless collection - Documented POST /v1/headless/collection (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-collection]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless collection id restore - Documented POST /v1/headless/collection/{id}/restore (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-collection-id-restore]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless course generate - Documented POST /v1/headless/course/generate (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-course-generate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless course id duplicate - Documented POST /v1/headless/course/{id}/duplicate (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-course-id-duplicate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless course id publish - Documented POST /v1/headless/course/{id}/publish (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-course-id-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless course id restore - Documented POST /v1/headless/course/{id}/restore (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-course-id-restore]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless course id revert - Documented POST /v1/headless/course/{id}/revert (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-course-id-revert]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless course url - Documented POST /v1/headless/course/url (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-course-url]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless embed course - Documented POST /v1/headless/embed/course (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-embed-course]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless generate course - Documented POST /v1/headless/generate/course (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-generate-course]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless generate course id quiz - Documented POST /v1/headless/generate/course/{id}/quiz (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-generate-course-id-quiz]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless generate course id screen - Documented POST /v1/headless/generate/course/{id}/screen (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-generate-course-id-screen]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless themes - Documented POST /v1/headless/themes (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-themes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless translation default id language - Documented POST /v1/headless/translation/default/{id}/{language} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-translation-default-id-language]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless translation sync id language - Documented POST /v1/headless/translation/sync/{id}/{language} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-translation-sync-id-language]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless translation translate id - Documented POST /v1/headless/translation/translate/{id} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-translation-translate-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v1 headless webhooks - Documented POST /v1/headless/webhooks (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.post.v1-headless-webhooks]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 headless client identifier - Documented PUT /v1/headless/client/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.put.v1-headless-client-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 headless themes id - Documented PUT /v1/headless/themes/{id} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.put.v1-headless-themes-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 headless usage client identifier - Documented PUT /v1/headless/usage/client/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.put.v1-headless-usage-client-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v1 headless user identifier - Documented PUT /v1/headless/user/{identifier} (not implemented) [intent=direct_write availability=not_implemented operation=coassemble.put.v1-headless-user-identifier]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - clients list - Run the clients ETL stream [intent=etl availability=implemented stream=clients]; notes: discrepancy=present-in-surface-absent-from-artifact
  - collection trackings list - Run the collection trackings ETL stream [intent=etl availability=implemented stream=collection_trackings]; notes: discrepancy=present-in-surface-absent-from-artifact
  - collections list - Run the collections ETL stream [intent=etl availability=implemented stream=collections]; notes: discrepancy=present-in-surface-absent-from-artifact
  - courses list - Run the courses ETL stream [intent=etl availability=implemented stream=courses]; notes: discrepancy=present-in-surface-absent-from-artifact
  - create collection apply - Plan and execute the create collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_collection]; approval: requires plan, preview, approval, and execute; risk: creates a new collection of courses; low-risk external mutation, no approval required; flags: --title (required)
  - delete client apply - Plan and execute the delete client reverse-ETL action [intent=reverse_etl availability=implemented write=delete_client]; approval: requires plan, preview, approval, and execute; risk: irreversibly removes a client (multi-tenant sub-account) and its documented on-delete effects on associated users; approval required; flags: --clientIdentifier (required)
  - delete collection apply - Plan and execute the delete collection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_collection]; approval: requires plan, preview, approval, and execute; risk: soft-deletes a collection (recoverable via restore_collection); approval required; flags: --id (required)
  - delete course apply - Plan and execute the delete course reverse-ETL action [intent=reverse_etl availability=implemented write=delete_course]; approval: requires plan, preview, approval, and execute; risk: soft-deletes a course (recoverable via restore_course within Coassemble's retention window); approval required; flags: --id (required)
  - delete tracking apply - Plan and execute the delete tracking reverse-ETL action [intent=reverse_etl availability=implemented write=delete_tracking]; approval: requires plan, preview, approval, and execute; risk: permanently erases one learner's tracking/progress record for a course; irreversible, approval required; flags: --id (required), --identifier (required)
  - delete translation apply - Plan and execute the delete translation reverse-ETL action [intent=reverse_etl availability=implemented write=delete_translation]; approval: requires plan, preview, approval, and execute; risk: permanently removes a language variant of a course; irreversible, approval required; flags: --course_id (required), --language (required)
  - delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: irreversibly removes a learner identity, applying Coassemble's server-side DEFAULT handling for that identity's course progress (the real endpoint also accepts optional action=reallocate|delete|ignore/reallocateTo/clientIdentifier query params to control that handling explicitly, and Coassemble's own docs do not fully specify their exact semantics beyond "choose what to do with any courses associated with this identifier" — this action deliberately does not expose them, since the write-action path/query dialect has no way to send an optional record field only when present, and silently defaulting an ambiguous, irreversible per-learner-data-retention choice would be worse than declaring it out of scope; approval required; flags: --identifier (required)
  - duplicate course apply - Plan and execute the duplicate course reverse-ETL action [intent=reverse_etl availability=implemented write=duplicate_course]; approval: requires plan, preview, approval, and execute; risk: creates a full copy of an existing course; low-risk external mutation, no approval required; flags: --id (required)
  - publish course apply - Plan and execute the publish course reverse-ETL action [intent=reverse_etl availability=implemented write=publish_course]; approval: requires plan, preview, approval, and execute; risk: publishes the current draft of a course, making it live for learners; no approval required; flags: --id (required)
  - restore collection apply - Plan and execute the restore collection reverse-ETL action [intent=reverse_etl availability=implemented write=restore_collection]; approval: requires plan, preview, approval, and execute; risk: restores a previously soft-deleted collection; no approval required; flags: --id (required)
  - restore course apply - Plan and execute the restore course reverse-ETL action [intent=reverse_etl availability=implemented write=restore_course]; approval: requires plan, preview, approval, and execute; risk: restores a previously soft-deleted course; no approval required; flags: --id (required)
  - screen types list - Run the screen types ETL stream [intent=etl availability=implemented stream=screen_types]; notes: discrepancy=present-in-surface-absent-from-artifact
  - set default translation apply - Plan and execute the set default translation reverse-ETL action [intent=reverse_etl availability=implemented write=set_default_translation]; approval: requires plan, preview, approval, and execute; risk: changes which language variant learners see by default for this course; no approval required; flags: --course_id (required), --language (required)
  - sync translation apply - Plan and execute the sync translation reverse-ETL action [intent=reverse_etl availability=implemented write=sync_translation]; approval: requires plan, preview, approval, and execute; risk: re-syncs a translated variant's content with upstream changes to the source-language course, which can overwrite manual edits made directly in the translated variant; no approval required; flags: --course_id (required), --language (required)
  - trackings list - Run the trackings ETL stream [intent=etl availability=implemented stream=trackings]; notes: discrepancy=present-in-surface-absent-from-artifact
  - translate course apply - Plan and execute the translate course reverse-ETL action [intent=reverse_etl availability=implemented write=translate_course]; approval: requires plan, preview, approval, and execute; risk: kicks off machine translation of a course into a new BCP-47 language variant; low-risk external mutation, no approval required; flags: --course_id (required), --language (required)
  - translations list - Run the translations ETL stream [intent=etl availability=implemented stream=translations]; notes: discrepancy=present-in-surface-absent-from-artifact
  - update client apply - Plan and execute the update client reverse-ETL action [intent=reverse_etl availability=implemented write=update_client]; approval: requires plan, preview, approval, and execute; risk: overwrites a client's arbitrary metadata bag; no approval required; flags: --clientIdentifier (required)
  - update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: overwrites a learner's profile fields (name/avatar/metadata) or reassigns their client; no approval required; flags: --identifier (required)
  - user trackings list - Run the user trackings ETL stream [intent=etl availability=implemented stream=user_trackings]; notes: discrepancy=present-in-surface-absent-from-artifact
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect coassemble
```

### Inspect as structured JSON

```bash
pm connectors inspect coassemble --json
```

## Agent Rules

- Run pm connectors inspect coassemble before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
