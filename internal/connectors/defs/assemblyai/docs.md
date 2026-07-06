# Overview

Reads AssemblyAI transcripts, per-transcript detail, sentences, paragraphs, and word-search matches,
and submits new transcription jobs, through the AssemblyAI REST API.

Readable streams: `transcript`, `transcript_sentences`, `transcript_paragraphs`,
`transcript_subtitle`, `transcript_detail`, `transcript_sentence_items`,
`transcript_paragraph_items`, `transcript_word_search_matches`.

Write actions: `create_transcript`.

Service API documentation: https://www.assemblyai.com/docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); AssemblyAI API key. Sent as the raw value of the
  Authorization header (no Bearer prefix); never logged.
- `base_url` (optional, string); default `https://api.assemblyai.com`; format `uri`; AssemblyAI API
  base URL override for tests or proxies.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page (1-200).
- `word_search_terms` (optional, string); default `the`; Comma-or-space-separated keywords passed
  verbatim as the transcript_word_search_matches stream's required words query parameter
  (AssemblyAI's GET /v2/transcript/{id}/word-search has no server-side default; set this to the
  terms you want counted across every synced transcript). Only consumed by that one stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.assemblyai.com`, `max_pages=0`, `page_size=100`,
`word_search_terms=the`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/v2/transcript` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path
`page_details.next_url`; next URLs stay on the configured API host.

- `transcript`: GET `/v2/transcript` - records path `transcripts`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path
  `page_details.next_url`; next URLs stay on the configured API host.
- `transcript_sentences`: GET `/v2/transcript` - records path `transcripts`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path
  `page_details.next_url`; next URLs stay on the configured API host.
- `transcript_paragraphs`: GET `/v2/transcript` - records path `transcripts`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path
  `page_details.next_url`; next URLs stay on the configured API host.
- `transcript_subtitle`: GET `/v2/transcript` - records path `transcripts`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path
  `page_details.next_url`; next URLs stay on the configured API host.
- `transcript_detail`: GET `/v2/transcript/{{ fanout.id }}` - records at response root; follows a
  next-page URL from the response body; URL path `page_details.next_url`; next URLs stay on the
  configured API host; fan-out; ids from request `/v2/transcript`; id-list records path
  `transcripts`; id field `id`; id inserted into the request path.
- `transcript_sentence_items`: GET `/v2/transcript/{{ fanout.id }}/sentences` - records path
  `sentences`; follows a next-page URL from the response body; URL path `page_details.next_url`;
  next URLs stay on the configured API host; fan-out; ids from request `/v2/transcript`; id-list
  records path `transcripts`; id field `id`; id inserted into the request path; stamps
  `transcript_id`.
- `transcript_paragraph_items`: GET `/v2/transcript/{{ fanout.id }}/paragraphs` - records path
  `paragraphs`; follows a next-page URL from the response body; URL path `page_details.next_url`;
  next URLs stay on the configured API host; fan-out; ids from request `/v2/transcript`; id-list
  records path `transcripts`; id field `id`; id inserted into the request path; stamps
  `transcript_id`.
- `transcript_word_search_matches`: GET `/v2/transcript/{{ fanout.id }}/word-search` - records path
  `matches`; query `words`=`{{ config.word_search_terms }}`; follows a next-page URL from the
  response body; URL path `page_details.next_url`; next URLs stay on the configured API host;
  fan-out; ids from request `/v2/transcript`; id-list records path `transcripts`; id field `id`; id
  inserted into the request path; stamps `transcript_id`.

## Write actions & risks

Overall write risk: submits a new transcription job (POST /v2/transcript) against a caller-supplied
audio_url; consumes account balance/quota.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_transcript`: POST `/v2/transcript` - kind `create`; body type `json`; required record
  fields `audio_url`; accepted fields `audio_url`, `auto_chapters`, `auto_highlights`,
  `entity_detection`, `format_text`, `language_code`, `language_detection`, `punctuate`,
  `redact_pii`, `sentiment_analysis`, `speaker_labels`, `speakers_expected`, `summarization`,
  `webhook_url`; risk: external mutation; submits a new transcription job against a caller-supplied
  audio_url and consumes AssemblyAI account balance/quota; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 8 stream-backed endpoint group(s), 1 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=2, destructive_admin=1, out_of_scope=5.
