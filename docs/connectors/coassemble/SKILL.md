---
name: pm-coassemble
description: Coassemble connector knowledge and safe action guide.
---

# pm-coassemble

## Purpose

Reads Coassemble courses, screen types, collections, clients, users, learner tracking, and translations, and writes course/collection/client/user/translation lifecycle actions, through the Coassemble headless REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

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
  - fields: active(), description(), id(), identified(), image(), is_sharable(), key(), paid(), private(), title()
- screen_types:
  - fields: icon(), id(), name(), premium(), title()
- trackings:
  - fields: completed(), course_id(), id(), identifier(), progress(), status()
- collections:
  - primary key: id
  - fields: active(), clientIdentifier(), created(), deleted(), description(), id(), identifier(), key(), title(), updated()
- clients:
  - primary key: clientIdentifier
  - fields: clientIdentifier(), created(), updated(), userCount()
- users:
  - primary key: identifier
  - fields: avatar(), clientIdentifier(), created(), identifier(), name(), testMode(), updated()
- user_trackings:
  - primary key: identifier
  - fields: avatar(), clientIdentifier(), identifier(), name(), totals(), trackings()
- collection_trackings:
  - primary key: id
  - fields: collection_id(), commenced(), completed(), courses(), id(), identifier(), name(), progressPercent(), totalTime()
- translations:
  - primary key: id, language
  - fields: course_id(), id(), language(), missingScreens()

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
  - optional fields: id, identifier
  - risk: permanently erases one learner's tracking/progress record for a course; irreversible, approval required
- create_collection:
  - endpoint: POST /api/v1/headless/collection
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
  - required fields: course_id
  - optional fields: language
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
