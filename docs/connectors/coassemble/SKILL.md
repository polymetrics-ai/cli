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
- user_id (secret) (required)
- user_token (secret) (required)

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

- ETL sync modes: full_refresh_append, full_refresh_overwrite

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
