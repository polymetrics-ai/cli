# pm connectors inspect tavus

```text
NAME
  pm connectors inspect tavus - Tavus connector manual

SYNOPSIS
  pm connectors inspect tavus
  pm connectors inspect tavus --json
  pm credentials add <name> --connector tavus [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Tavus faces (replicas), videos, conversations, PALs, guardrails, objectives, documents, pronunciation dictionaries, voices, and skills, and writes approved video/conversation/PAL/guardrail/objective/document/pronunciation-dictionary create-delete mutations through the Tavus API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  replicas:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name()
  videos:
    primary key: id
    fields: download_url(), error_details(), hosted_url(), id(), name(), status(), stream_url()
  conversations:
    primary key: id
    cursor: created_at
    fields: callback_url(), conversation_url(), created_at(), face_id(), id(), name(), pal_id(), status(), updated_at()
  pals:
    primary key: id
    fields: conferencing_email(), default_face_id(), id(), name(), system_prompt()
  guardrails:
    primary key: id
    fields: callback_url(), guardrail_prompt(), id(), modality(), name(), tags()
  objectives:
    primary key: id
    fields: confirmation_mode(), id(), modality(), name(), objective_prompt(), output_variables()
  documents:
    primary key: id
    fields: document_url(), error_message(), id(), name(), progress(), status()
  pronunciation_dictionaries:
    primary key: id
    fields: id(), name(), rules_count()
  voices:
    primary key: voice_name, face_id
    fields: audio_url(), face_id(), voice_name()
  skills:
    primary key: skill_id
    fields: description(), display_name(), skill_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_video:
    endpoint: POST /videos
    risk: generates a new async video render from a face and script/audio; consumes video-generation minutes on the account
  delete_video:
    endpoint: DELETE /videos/{{ record.id }}
    required fields: id
    risk: permanently deletes a generated video and its hosted/download URLs
  create_conversation:
    endpoint: POST /conversations
    risk: starts a real-time video conversation, which begins consuming conversational-minutes billing immediately and (unless test_mode) places a live call
  end_conversation:
    endpoint: POST /conversations/{{ record.id }}/end
    required fields: id
    risk: ends an active conversation for every participant; routine call cleanup, not destructive to conversation history (compare delete_conversation)
  delete_conversation:
    endpoint: DELETE /conversations/{{ record.id }}
    required fields: id
    risk: permanently deletes a conversation and its recorded history; use end_conversation instead for routine call cleanup
  create_pal:
    endpoint: POST /pals
    risk: creates a new PAL persona; low-risk external mutation, no approval required
  delete_pal:
    endpoint: DELETE /pals/{{ record.id }}
    required fields: id
    risk: permanently deletes a PAL; any conversation still referencing its pal_id will fail to start
  create_guardrail:
    endpoint: POST /guardrails
    risk: creates a new behavioral guardrail; low-risk external mutation, no approval required
  delete_guardrail:
    endpoint: DELETE /guardrails/{{ record.id }}
    required fields: id
    risk: permanently deletes a guardrail; any PAL referencing it via guardrail_ids loses that behavioral boundary immediately
  create_objective:
    endpoint: POST /objectives
    risk: creates one or more new PAL objectives; low-risk external mutation, no approval required
  delete_objective:
    endpoint: DELETE /objectives/{{ record.id }}
    required fields: id
    risk: permanently deletes an objective; any PAL referencing it via objectives_id loses that goal-oriented instruction immediately
  create_document:
    endpoint: POST /documents
    risk: uploads a document to the knowledge base; processing is asynchronous and the document becomes available to PALs only once status reaches ready
  delete_document:
    endpoint: DELETE /documents/{{ record.id }}
    required fields: id
    risk: permanently deletes a knowledge-base document and its processed data; any PAL referencing it via document_ids loses that knowledge source immediately
  create_pronunciation_dictionary:
    endpoint: POST /pronunciation-dictionaries
    risk: creates a new pronunciation dictionary; low-risk external mutation, no approval required
  delete_pronunciation_dictionary:
    endpoint: DELETE /pronunciation-dictionaries/{{ record.id }}
    required fields: id
    risk: permanently deletes a pronunciation dictionary and removes it from every linked PAL

SECURITY
  read risk: external Tavus API read of face, video, conversation, PAL, guardrail, objective, document, pronunciation-dictionary, voice, and skill data
  write risk: external Tavus API mutation (create/delete videos, conversations, PALs, guardrails, objectives, documents, pronunciation dictionaries; end conversations); create_video/create_conversation consume billed generation/conversational minutes
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect tavus

  # Inspect as structured JSON
  pm connectors inspect tavus --json

AGENT WORKFLOW
  - Run pm connectors inspect tavus before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
