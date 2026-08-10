---
name: pm-tavus
description: Tavus connector knowledge and safe action guide.
---

# pm-tavus

## Purpose

Reads Tavus faces (replicas), videos, conversations, PALs, guardrails, objectives, documents, pronunciation dictionaries, voices, and skills, and writes approved video/conversation/PAL/guardrail/objective/document/pronunciation-dictionary create-delete mutations through the Tavus API.

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
- api_key (secret) (required)

## ETL Streams

- replicas:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), name(string)
- videos:
  - primary key: id
  - fields: download_url(string), error_details(string), hosted_url(string), id(string), name(string), status(string), stream_url(string)
- conversations:
  - primary key: id
  - cursor: created_at
  - fields: callback_url(string), conversation_url(string), created_at(string), face_id(string), id(string), name(string), pal_id(string), status(string), updated_at(string)
- pals:
  - primary key: id
  - fields: conferencing_email(string), default_face_id(string), id(string), name(string), system_prompt(string)
- guardrails:
  - primary key: id
  - fields: callback_url(string), guardrail_prompt(string), id(string), modality(string), name(string), tags(array)
- objectives:
  - primary key: id
  - fields: confirmation_mode(string), id(string), modality(string), name(string), objective_prompt(string), output_variables(array)
- documents:
  - primary key: id
  - fields: document_url(string), error_message(string), id(string), name(string), progress(integer), status(string)
- pronunciation_dictionaries:
  - primary key: id
  - fields: id(string), name(string), rules_count(integer)
- voices:
  - primary key: voice_name, face_id
  - fields: audio_url(string), face_id(string), voice_name(string)
- skills:
  - primary key: skill_id
  - fields: description(string), display_name(string), skill_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_video:
  - endpoint: POST /videos
  - required fields: replica_id
  - risk: generates a new async video render from a face and script/audio; consumes video-generation minutes on the account
- delete_video:
  - endpoint: DELETE /videos/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a generated video and its hosted/download URLs
- create_conversation:
  - endpoint: POST /conversations
  - risk: starts a real-time video conversation, which begins consuming conversational-minutes billing immediately and (unless test_mode) places a live call
- end_conversation:
  - endpoint: POST /conversations/{{ record.id }}/end
  - required fields: id
  - risk: ends an active conversation for every participant; routine call cleanup, not destructive to conversation history (compare delete_conversation)
- delete_conversation:
  - endpoint: DELETE /conversations/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a conversation and its recorded history; use end_conversation instead for routine call cleanup
- create_pal:
  - endpoint: POST /pals
  - required fields: default_face_id
  - risk: creates a new PAL persona; low-risk external mutation, no approval required
- delete_pal:
  - endpoint: DELETE /pals/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a PAL; any conversation still referencing its pal_id will fail to start
- create_guardrail:
  - endpoint: POST /guardrails
  - required fields: guardrail_name, guardrail_prompt
  - risk: creates a new behavioral guardrail; low-risk external mutation, no approval required
- delete_guardrail:
  - endpoint: DELETE /guardrails/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a guardrail; any PAL referencing it via guardrail_ids loses that behavioral boundary immediately
- create_objective:
  - endpoint: POST /objectives
  - required fields: data
  - risk: creates one or more new PAL objectives; low-risk external mutation, no approval required
- delete_objective:
  - endpoint: DELETE /objectives/{{ record.id }}
  - required fields: id
  - risk: permanently deletes an objective; any PAL referencing it via objectives_id loses that goal-oriented instruction immediately
- create_document:
  - endpoint: POST /documents
  - required fields: document_url
  - risk: uploads a document to the knowledge base; processing is asynchronous and the document becomes available to PALs only once status reaches ready
- delete_document:
  - endpoint: DELETE /documents/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a knowledge-base document and its processed data; any PAL referencing it via document_ids loses that knowledge source immediately
- create_pronunciation_dictionary:
  - endpoint: POST /pronunciation-dictionaries
  - required fields: name
  - risk: creates a new pronunciation dictionary; low-risk external mutation, no approval required
- delete_pronunciation_dictionary:
  - endpoint: DELETE /pronunciation-dictionaries/{{ record.id }}
  - required fields: id
  - risk: permanently deletes a pronunciation dictionary and removes it from every linked PAL

## Security

- read risk: external Tavus API read of face, video, conversation, PAL, guardrail, objective, document, pronunciation-dictionary, voice, and skill data
- write risk: external Tavus API mutation (create/delete videos, conversations, PALs, guardrails, objectives, documents, pronunciation dictionaries; end conversations); create_video/create_conversation consume billed generation/conversational minutes
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect tavus
```

### Inspect as structured JSON

```bash
pm connectors inspect tavus --json
```

## Agent Rules

- Run pm connectors inspect tavus before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
