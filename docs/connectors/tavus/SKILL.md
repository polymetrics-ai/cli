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
- api_key (secret)

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

## Command Surface

- Run Tavus's declared streams and reverse-ETL actions.
- Usage: pm tavus <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete v2 deployments deployment-id - Documented DELETE /v2/deployments/{deployment_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-deployments-deployment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 faces face-id - Documented DELETE /v2/faces/{face_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-faces-face-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 lipsync lipsync-id - Documented DELETE /v2/lipsync/{lipsync_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-lipsync-lipsync-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 pals pal-id skills skill-id - Documented DELETE /v2/pals/{pal_id}/skills/{skill_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-pals-pal-id-skills-skill-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 pals pal-id tools tool-id - Documented DELETE /v2/pals/{pal_id}/tools/{tool_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-pals-pal-id-tools-tool-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 replacements replacement-id - Documented DELETE /v2/replacements/{replacement_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-replacements-replacement-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 tools tool-id - Documented DELETE /v2/tools/{tool_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-tools-tool-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v2 transcriptions transcription-id - Documented DELETE /v2/transcriptions/{transcription_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.delete.v2-transcriptions-transcription-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get v2 conversations conversation-id - Documented GET /v2/conversations/{conversation_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-conversations-conversation-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 conversations conversation-id canvas interactions - Documented GET /v2/conversations/{conversation_id}/canvas/interactions (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-conversations-conversation-id-canvas-interactions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 deployments - Documented GET /v2/deployments (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-deployments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 deployments deployment-id - Documented GET /v2/deployments/{deployment_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-deployments-deployment-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 deployments deployment-id init - Documented GET /v2/deployments/{deployment_id}/init (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-deployments-deployment-id-init]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 documents document-id - Documented GET /v2/documents/{document_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-documents-document-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 faces - Documented GET /v2/faces (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-faces]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 faces face-id - Documented GET /v2/faces/{face_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-faces-face-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 guardrails guardrail-id - Documented GET /v2/guardrails/{guardrail_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-guardrails-guardrail-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 guardrails tags - Documented GET /v2/guardrails/tags (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-guardrails-tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 lipsync - Documented GET /v2/lipsync (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-lipsync]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 lipsync lipsync-id - Documented GET /v2/lipsync/{lipsync_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-lipsync-lipsync-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 objectives objectives-id - Documented GET /v2/objectives/{objectives_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-objectives-objectives-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 pals check-username - Documented GET /v2/pals/check-username (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-pals-check-username]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 pals pal-id - Documented GET /v2/pals/{pal_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-pals-pal-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 pals pal-id skills - Documented GET /v2/pals/{pal_id}/skills (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-pals-pal-id-skills]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 pals pal-id skills skill-id - Documented GET /v2/pals/{pal_id}/skills/{skill_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-pals-pal-id-skills-skill-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 pals pal-id tools - Documented GET /v2/pals/{pal_id}/tools (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-pals-pal-id-tools]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 pronunciation-dictionaries dictionary-id - Documented GET /v2/pronunciation-dictionaries/{dictionary_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-pronunciation-dictionaries-dictionary-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 replacements - Documented GET /v2/replacements (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-replacements]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 replacements replacement-id - Documented GET /v2/replacements/{replacement_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-replacements-replacement-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 skills skill-id - Documented GET /v2/skills/{skill_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-skills-skill-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 tools - Documented GET /v2/tools (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-tools]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 tools tool-id - Documented GET /v2/tools/{tool_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-tools-tool-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 transcriptions - Documented GET /v2/transcriptions (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-transcriptions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 transcriptions transcription-id - Documented GET /v2/transcriptions/{transcription_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-transcriptions-transcription-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v2 videos video-id - Documented GET /v2/videos/{video_id} (not implemented) [intent=direct_read availability=not_implemented operation=tavus.get.v2-videos-video-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch v2 deployments deployment-id - Documented PATCH /v2/deployments/{deployment_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-deployments-deployment-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 documents document-id - Documented PATCH /v2/documents/{document_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-documents-document-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 faces face-id name - Documented PATCH /v2/faces/{face_id}/name (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-faces-face-id-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 guardrails guardrail-id - Documented PATCH /v2/guardrails/{guardrail_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-guardrails-guardrail-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 objectives objectives-id - Documented PATCH /v2/objectives/{objectives_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-objectives-objectives-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 pals pal-id - Documented PATCH /v2/pals/{pal_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-pals-pal-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 pals pal-id skills skill-id - Documented PATCH /v2/pals/{pal_id}/skills/{skill_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-pals-pal-id-skills-skill-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 pronunciation-dictionaries dictionary-id - Documented PATCH /v2/pronunciation-dictionaries/{dictionary_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-pronunciation-dictionaries-dictionary-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 tools tool-id - Documented PATCH /v2/tools/{tool_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-tools-tool-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v2 videos video-id name - Documented PATCH /v2/videos/{video_id}/name (not implemented) [intent=direct_write availability=not_implemented operation=tavus.patch.v2-videos-video-id-name]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 conversations conversation-id canvas interactions - Documented POST /v2/conversations/{conversation_id}/canvas/interactions (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-conversations-conversation-id-canvas-interactions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 deployments - Documented POST /v2/deployments (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-deployments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 deployments deployment-id conversations conversation-id end - Documented POST /v2/deployments/{deployment_id}/conversations/{conversation_id}/end (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-deployments-deployment-id-conversations-conversation-id-end]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 deployments deployment-id start - Documented POST /v2/deployments/{deployment_id}/start (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-deployments-deployment-id-start]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 documents document-id recrawl - Documented POST /v2/documents/{document_id}/recrawl (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-documents-document-id-recrawl]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 faces - Documented POST /v2/faces (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-faces]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 lipsync - Documented POST /v2/lipsync (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-lipsync]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 pals pal-id publish - Documented POST /v2/pals/{pal_id}/publish (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-pals-pal-id-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 pals pal-id tools - Documented POST /v2/pals/{pal_id}/tools (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-pals-pal-id-tools]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 replacements - Documented POST /v2/replacements (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-replacements]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 tools - Documented POST /v2/tools (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-tools]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v2 transcriptions - Documented POST /v2/transcriptions (not implemented) [intent=direct_write availability=not_implemented operation=tavus.post.v2-transcriptions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v2 pals pal-id skills - Documented PUT /v2/pals/{pal_id}/skills (not implemented) [intent=direct_write availability=not_implemented operation=tavus.put.v2-pals-pal-id-skills]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v2 pals pal-id skills skill-id - Documented PUT /v2/pals/{pal_id}/skills/{skill_id} (not implemented) [intent=direct_write availability=not_implemented operation=tavus.put.v2-pals-pal-id-skills-skill-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - conversations list - Run the conversations ETL stream [intent=etl availability=implemented stream=conversations]
  - create conversation apply - Plan and execute the create conversation reverse-ETL action [intent=reverse_etl availability=implemented write=create_conversation]; approval: requires plan, preview, approval, and execute; risk: starts a real-time video conversation, which begins consuming conversational-minutes billing immediately and (unless test_mode) places a live call
  - create document apply - Plan and execute the create document reverse-ETL action [intent=reverse_etl availability=implemented write=create_document]; approval: requires plan, preview, approval, and execute; risk: uploads a document to the knowledge base; processing is asynchronous and the document becomes available to PALs only once status reaches ready; flags: --document_url (required)
  - create guardrail apply - Plan and execute the create guardrail reverse-ETL action [intent=reverse_etl availability=implemented write=create_guardrail]; approval: requires plan, preview, approval, and execute; risk: creates a new behavioral guardrail; low-risk external mutation, no approval required; flags: --guardrail_name (required), --guardrail_prompt (required)
  - create objective apply - Plan and execute the create objective reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_objective]; approval: requires plan, preview, approval, and execute; risk: creates one or more new PAL objectives; low-risk external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create pal apply - Plan and execute the create pal reverse-ETL action [intent=reverse_etl availability=implemented write=create_pal]; approval: requires plan, preview, approval, and execute; risk: creates a new PAL persona; low-risk external mutation, no approval required; flags: --default_face_id (required)
  - create pronunciation dictionary apply - Plan and execute the create pronunciation dictionary reverse-ETL action [intent=reverse_etl availability=implemented write=create_pronunciation_dictionary]; approval: requires plan, preview, approval, and execute; risk: creates a new pronunciation dictionary; low-risk external mutation, no approval required; flags: --name (required)
  - create video apply - Plan and execute the create video reverse-ETL action [intent=reverse_etl availability=implemented write=create_video]; approval: requires plan, preview, approval, and execute; risk: generates a new async video render from a face and script/audio; consumes video-generation minutes on the account; flags: --replica_id (required)
  - delete conversation apply - Plan and execute the delete conversation reverse-ETL action [intent=reverse_etl availability=implemented write=delete_conversation]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a conversation and its recorded history; use end_conversation instead for routine call cleanup; flags: --id (required)
  - delete document apply - Plan and execute the delete document reverse-ETL action [intent=reverse_etl availability=implemented write=delete_document]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a knowledge-base document and its processed data; any PAL referencing it via document_ids loses that knowledge source immediately; flags: --id (required)
  - delete guardrail apply - Plan and execute the delete guardrail reverse-ETL action [intent=reverse_etl availability=implemented write=delete_guardrail]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a guardrail; any PAL referencing it via guardrail_ids loses that behavioral boundary immediately; flags: --id (required)
  - delete objective apply - Plan and execute the delete objective reverse-ETL action [intent=reverse_etl availability=implemented write=delete_objective]; approval: requires plan, preview, approval, and execute; risk: permanently deletes an objective; any PAL referencing it via objectives_id loses that goal-oriented instruction immediately; flags: --id (required)
  - delete pal apply - Plan and execute the delete pal reverse-ETL action [intent=reverse_etl availability=implemented write=delete_pal]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a PAL; any conversation still referencing its pal_id will fail to start; flags: --id (required)
  - delete pronunciation dictionary apply - Plan and execute the delete pronunciation dictionary reverse-ETL action [intent=reverse_etl availability=implemented write=delete_pronunciation_dictionary]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a pronunciation dictionary and removes it from every linked PAL; flags: --id (required)
  - delete video apply - Plan and execute the delete video reverse-ETL action [intent=reverse_etl availability=implemented write=delete_video]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a generated video and its hosted/download URLs; flags: --id (required)
  - documents list - Run the documents ETL stream [intent=etl availability=implemented stream=documents]
  - end conversation apply - Plan and execute the end conversation reverse-ETL action [intent=reverse_etl availability=implemented write=end_conversation]; approval: requires plan, preview, approval, and execute; risk: ends an active conversation for every participant; routine call cleanup, not destructive to conversation history (compare delete_conversation); flags: --id (required)
  - guardrails list - Run the guardrails ETL stream [intent=etl availability=implemented stream=guardrails]
  - objectives list - Run the objectives ETL stream [intent=etl availability=implemented stream=objectives]
  - pals list - Run the pals ETL stream [intent=etl availability=implemented stream=pals]
  - pronunciation dictionaries list - Run the pronunciation dictionaries ETL stream [intent=etl availability=implemented stream=pronunciation_dictionaries]
  - replicas list - Run the replicas ETL stream [intent=etl availability=implemented stream=replicas]; notes: discrepancy=present-in-surface-absent-from-artifact
  - skills list - Run the skills ETL stream [intent=etl availability=implemented stream=skills]
  - videos list - Run the videos ETL stream [intent=etl availability=implemented stream=videos]
  - voices list - Run the voices ETL stream [intent=etl availability=implemented stream=voices]

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
