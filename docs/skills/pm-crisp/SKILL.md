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
  - conversations list - List one bounded Crisp conversation page with named provider filters. [intent=etl availability=implemented stream=list_conversations]; flags: --website-id (non-empty) (string): Crisp website identifier; overrides the credential profile config.: maps_to=config.website_id, --page-number (minimum=1) (integer): Positive Crisp conversation page number; defaults to 1 from config.: maps_to=config.page_number, --per-page (max 4096 bytes) (enum): Crisp conversation page size, bounded by the provider to 20 through 50.: values=20|21|22|23|24|25|26|27|28|29|30|31|32|33|34|35|36|37|38|39|40|41|42|43|44|45|46|47|48|49|50: maps_to=query.per_page, --search-query (non-empty, max 4096 bytes) (string): Crisp conversation text, segment, or filter search query.: maps_to=query.search_query, --search-type (max 4096 bytes) (enum): Crisp search interpretation.: values=text|segment|filter: maps_to=query.search_type, --search-operator (max 4096 bytes) (enum): Boolean operator for filter search.: values=and|or: maps_to=query.search_operator, --include-empty (max 4096 bytes) (enum): Include conversations without messages (1) or not (0).: values=0|1: maps_to=query.include_empty, --filter-inbox-id (non-empty, max 4096 bytes) (string): Limit conversations to a Crisp inbox identifier, or all for every inbox.: maps_to=query.filter_inbox_id, --filter-unread (max 4096 bytes) (enum): Filter unread conversations (1) or leave disabled (0).: values=0|1: maps_to=query.filter_unread, --filter-resolved (max 4096 bytes) (enum): Filter resolved conversations (1) or leave disabled (0).: values=0|1: maps_to=query.filter_resolved, --filter-not-resolved (max 4096 bytes) (enum): Filter unresolved conversations (1) or leave disabled (0).: values=0|1: maps_to=query.filter_not_resolved, --filter-mention (max 4096 bytes) (enum): Filter conversations mentioning the token user (1) or leave disabled (0).: values=0|1: maps_to=query.filter_mention, --filter-assigned (non-empty, max 4096 bytes) (string): Limit conversations to the specified Crisp user identifier.: maps_to=query.filter_assigned, --filter-unassigned (max 4096 bytes) (enum): Filter conversations without an assigned user (1) or leave disabled (0).: values=0|1: maps_to=query.filter_unassigned, --filter-date-start (non-empty, max 4096 bytes, format=date-time) (string): Inclusive ISO-8601 lower bound for conversation update date.: maps_to=query.filter_date_start, --filter-date-end (non-empty, max 4096 bytes, format=date-time) (string): Inclusive ISO-8601 upper bound for conversation update date.: maps_to=query.filter_date_end, --order-date-created (max 4096 bytes) (enum): Order by creation date (1) or retain the provider default (0).: values=0|1: maps_to=query.order_date_created, --order-date-updated (max 4096 bytes) (enum): Order by update date (1) or retain the provider default (0).: values=0|1: maps_to=query.order_date_updated, --order-date-waiting (max 4096 bytes) (enum): Order/filter by longest waiting date (1) or retain the provider default (0).: values=0|1: maps_to=query.order_date_waiting
  - conversations suggested-segments - List one Crisp suggested-conversation-segments page. [intent=etl availability=implemented stream=suggested_conversation_segments]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --page-number (minimum=1) (integer): Positive Crisp page number; defaults to 1 from config.: maps_to=config.page_number
  - conversations suggested-data - List one Crisp suggested-conversation-data-keys page. [intent=etl availability=implemented stream=suggested_conversation_data]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --page-number (minimum=1) (integer): Positive Crisp page number; defaults to 1 from config.: maps_to=config.page_number
  - conversations spam-list - List one Crisp spam-conversations page. [intent=etl availability=implemented stream=spam_conversations]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --page-number (minimum=1) (integer): Positive Crisp page number; defaults to 1 from config.: maps_to=config.page_number, --filter-type (non-empty, max 4096 bytes) (string): Optional Crisp spam type filter, such as email.: maps_to=query.filter_type
  - conversations spam-content - Resolve one Crisp spam conversation's content. [intent=etl availability=implemented stream=spam_conversation_content]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --spam-id (non-empty) (string): Crisp spam conversation identifier.: maps_to=config.spam_id
- Crisp conversation-scoped read commands
  - conversation get - Resolve one Crisp conversation. [intent=etl availability=implemented stream=conversation]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation messages - Resolve a Crisp conversation message batch and page it with a named timestamp selector. [intent=etl availability=implemented stream=conversation_messages]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id, --timestamp-before (non-empty, max 4096 bytes) (string): Return a message batch ending before this Crisp timestamp.: maps_to=query.timestamp_before, --timestamp-after (non-empty, max 4096 bytes) (string): Return a message batch starting after this Crisp timestamp.: maps_to=query.timestamp_after, --timestamp-around (non-empty, max 4096 bytes) (string): Return a message batch centered around this Crisp timestamp.: maps_to=query.timestamp_around
  - conversation message - Resolve one Crisp conversation message by fingerprint. [intent=etl availability=implemented stream=conversation_message]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id, --fingerprint (non-empty) (string): Crisp message fingerprint.: maps_to=config.fingerprint
  - conversation routing - Resolve Crisp routing assignment for one conversation. [intent=etl availability=implemented stream=conversation_routing]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation meta - Resolve Crisp metadata for one conversation. [intent=etl availability=implemented stream=conversation_meta]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation original-message - Resolve one original Crisp conversation message. [intent=etl availability=implemented stream=conversation_original_message]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id, --original-id (non-empty) (string): Crisp original-message identifier.: maps_to=config.original_id
  - conversation pages - List one Crisp conversation browser-pages page. [intent=etl availability=implemented stream=conversation_pages]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id, --page-number (minimum=1) (integer): Positive Crisp page number; defaults to 1 from config.: maps_to=config.page_number
  - conversation events - List one Crisp conversation events page. [intent=etl availability=implemented stream=conversation_events]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id, --page-number (minimum=1) (integer): Positive Crisp page number; defaults to 1 from config.: maps_to=config.page_number
  - conversation files - List one Crisp conversation files page without downloading file contents. [intent=etl availability=implemented stream=conversation_files]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id, --page-number (minimum=1) (integer): Positive Crisp page number; defaults to 1 from config.: maps_to=config.page_number
  - conversation state - Resolve Crisp state for one conversation. [intent=etl availability=implemented stream=conversation_state]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation relations - Resolve Crisp related conversations for one conversation. [intent=etl availability=implemented stream=conversation_relations]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation participants - Resolve Crisp participants for one conversation. [intent=etl availability=implemented stream=conversation_participants]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation block-status - Resolve whether a Crisp conversation is blocked. [intent=etl availability=implemented stream=conversation_block_status]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation verify-status - Resolve Crisp identity-verification status for one conversation. [intent=etl availability=implemented stream=conversation_verify_status]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation browsing - List browsing sessions for one Crisp conversation. [intent=etl availability=implemented stream=conversation_browsing]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
  - conversation call - Resolve an ongoing Crisp call session for one conversation. [intent=etl availability=implemented stream=conversation_call]; flags: --website-id (non-empty) (string): Crisp website identifier.: maps_to=config.website_id, --session-id (non-empty) (string): Crisp conversation session identifier.: maps_to=config.session_id
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
