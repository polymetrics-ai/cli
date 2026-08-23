---
name: pm-crisp
description: Crisp connector knowledge and safe action guide.
---

# pm-crisp

## Purpose

Reads the first Wave 1 set of Crisp REST API conversation resources through HTTP Basic authentication.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- fingerprint
- original_id
- page_number
- session_id
- spam_id
- token_tier
- website_id (required)
- identifier (secret)
- key (secret)

## ETL Streams

- list_conversations:
- suggested_conversation_segments:
- suggested_conversation_data:
- spam_conversations:
- spam_conversation_content:
- conversation:
- conversation_messages:
- conversation_message:
- conversation_routing:
- conversation_meta:
- conversation_original_message:
- conversation_pages:
- conversation_events:
- conversation_files:
- conversation_state:
- conversation_relations:
- conversation_participants:
- conversation_block_status:
- conversation_verify_status:
- conversation_browsing:
- conversation_call:

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external Crisp REST API reads of conversation records, messages, routing, metadata, participants, browsing sessions, and call state
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Read Crisp conversation resources through bounded, typed Wave 1 REST commands.
- Usage: pm crisp <conversations|conversation> <command> [flags]
- Source CLI: Crisp REST API Reference (V1) (Official Postman Collection v2.1.0, retrieved 2026-08-05; ETag 6a6b6306-32de7)
- Global flags:
  - --credential (non-empty) (string): Credential profile name; never pass Crisp token values as flags.: maps_to=config.credential
  - --config (string_array): Connector config override as key=value; never pass secret values here.
  - --json (boolean): Render machine-readable JSON output.
  - --limit (integer): Maximum records to emit from a read command.
- Crisp conversation-list read commands
  - conversations list - List one bounded Crisp conversation page with named provider filters. [intent=etl availability=implemented stream=list_conversations]; flags: --website-id (non-empty), --page-number (minimum=1), --per-page (max 4096 bytes), --search-query (non-empty, max 4096 bytes), --search-type (max 4096 bytes), --search-operator (max 4096 bytes), --include-empty (max 4096 bytes), --filter-inbox-id (non-empty, max 4096 bytes), --filter-unread (max 4096 bytes), --filter-resolved (max 4096 bytes), --filter-not-resolved (max 4096 bytes), --filter-mention (max 4096 bytes), --filter-assigned (non-empty, max 4096 bytes), --filter-unassigned (max 4096 bytes), --filter-date-start (non-empty, max 4096 bytes, format=date-time), --filter-date-end (non-empty, max 4096 bytes, format=date-time), --order-date-created (max 4096 bytes), --order-date-updated (max 4096 bytes), --order-date-waiting (max 4096 bytes)
  - conversations suggested-segments - List one Crisp suggested-conversation-segments page. [intent=etl availability=implemented stream=suggested_conversation_segments]; flags: --website-id (non-empty), --page-number (minimum=1)
  - conversations suggested-data - List one Crisp suggested-conversation-data-keys page. [intent=etl availability=implemented stream=suggested_conversation_data]; flags: --website-id (non-empty), --page-number (minimum=1)
  - conversations spam-list - List one Crisp spam-conversations page. [intent=etl availability=implemented stream=spam_conversations]; flags: --website-id (non-empty), --page-number (minimum=1), --filter-type (non-empty, max 4096 bytes)
  - conversations spam-content - Resolve one Crisp spam conversation's content. [intent=etl availability=implemented stream=spam_conversation_content]; flags: --website-id (non-empty), --spam-id (non-empty)
- Crisp conversation-scoped read commands
  - conversation get - Resolve one Crisp conversation. [intent=etl availability=implemented stream=conversation]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation messages - Resolve a Crisp conversation message batch and page it with a named timestamp selector. [intent=etl availability=implemented stream=conversation_messages]; flags: --website-id (non-empty), --session-id (non-empty), --timestamp-before (non-empty, max 4096 bytes), --timestamp-after (non-empty, max 4096 bytes), --timestamp-around (non-empty, max 4096 bytes)
  - conversation message - Resolve one Crisp conversation message by fingerprint. [intent=etl availability=implemented stream=conversation_message]; flags: --website-id (non-empty), --session-id (non-empty), --fingerprint (non-empty)
  - conversation routing - Resolve Crisp routing assignment for one conversation. [intent=etl availability=implemented stream=conversation_routing]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation meta - Resolve Crisp metadata for one conversation. [intent=etl availability=implemented stream=conversation_meta]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation original-message - Resolve one original Crisp conversation message. [intent=etl availability=implemented stream=conversation_original_message]; flags: --website-id (non-empty), --session-id (non-empty), --original-id (non-empty)
  - conversation pages - List one Crisp conversation browser-pages page. [intent=etl availability=implemented stream=conversation_pages]; flags: --website-id (non-empty), --session-id (non-empty), --page-number (minimum=1)
  - conversation events - List one Crisp conversation events page. [intent=etl availability=implemented stream=conversation_events]; flags: --website-id (non-empty), --session-id (non-empty), --page-number (minimum=1)
  - conversation files - List one Crisp conversation files page without downloading file contents. [intent=etl availability=implemented stream=conversation_files]; flags: --website-id (non-empty), --session-id (non-empty), --page-number (minimum=1)
  - conversation state - Resolve Crisp state for one conversation. [intent=etl availability=implemented stream=conversation_state]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation relations - Resolve Crisp related conversations for one conversation. [intent=etl availability=implemented stream=conversation_relations]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation participants - Resolve Crisp participants for one conversation. [intent=etl availability=implemented stream=conversation_participants]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation block-status - Resolve whether a Crisp conversation is blocked. [intent=etl availability=implemented stream=conversation_block_status]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation verify-status - Resolve Crisp identity-verification status for one conversation. [intent=etl availability=implemented stream=conversation_verify_status]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation browsing - List browsing sessions for one Crisp conversation. [intent=etl availability=implemented stream=conversation_browsing]; flags: --website-id (non-empty), --session-id (non-empty)
  - conversation call - Resolve an ongoing Crisp call session for one conversation. [intent=etl availability=implemented stream=conversation_call]; flags: --website-id (non-empty), --session-id (non-empty)
- Help topics:
  - auth - Crisp uses HTTP Basic identifier/key credentials plus X-Crisp-Tier; add secrets from environment variables or stdin only.
  - wave-1 - Wave 1 exposes 21 read-only conversation GET operations; later provider operations remain explicitly blocked in api_surface.json.

## Commands

### Inspect as a manual

```bash
pm connectors inspect crisp
```

### Inspect as structured JSON

```bash
pm connectors inspect crisp --json
```

## Agent Rules

- Run pm connectors inspect crisp before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
