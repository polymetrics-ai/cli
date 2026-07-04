# Overview

AssemblyAI is a declarative-HTTP bundle migrated from `internal/connectors/assemblyai` (the
hand-written legacy connector, which stays registered and unchanged until wave6's registry flip).
It reads AssemblyAI transcripts, per-transcript detail, sentences, paragraphs, and word-search
matches, and submits new transcription jobs, through the AssemblyAI REST API. Pass B full-surface
expansion (this revision) adds every remaining JSON-shaped endpoint on top of the wave2
legacy-parity streams: a `transcript_detail` fan-out stream (`GET /v2/transcript/{id}`), fan-out
`transcript_sentence_items`/`transcript_paragraph_items` streams that flatten each transcript's
sentence/paragraph array into individual rows, a fan-out `transcript_word_search_matches` stream
(`GET /v2/transcript/{id}/word-search`), and a `create_transcript` write action
(`POST /v2/transcript`).

## Auth setup

Provide an AssemblyAI API key via the `api_key` secret. It is sent as the **raw** value of the
`Authorization` header (no `Bearer` prefix — AssemblyAI's own convention), expressed as
`{"mode": "api_key_header", "header": "Authorization", "value": "{{ secrets.api_key }}", "prefix":
""}` — matching legacy's `connsdk.APIKeyHeader("Authorization", secret, "")` exactly.

## Streams notes

**Legacy-parity streams** (unchanged from wave2): all 4 of `transcript`, `transcript_sentences`,
`transcript_paragraphs`, `transcript_subtitle` read the identical `GET /v2/transcript` list
endpoint with records at `transcripts` — legacy's `assemblyaiStreamEndpoints` table routes all four
stream names to the same `v2/transcript` resource, differing only in which fields their record
mapper keeps. Pagination follows AssemblyAI's `page_details.next_url` absolute-URL convention
(`pagination.type: next_url`); `page_size` (default 100) is sent as the `limit` query param on the
first request only.

**New Pass B fan-out streams** each list every transcript via the SAME `GET /v2/transcript`
endpoint (`fan_out.ids_from.request`, `records_path: transcripts`, `id_field: id`), then repeat a
sub-resource GET once per transcript id (`fan_out.into.path_var: transcript_id`):

- `transcript_detail` — `GET /v2/transcript/{id}` (the full transcript object: `text`,
  `audio_duration`, language/PII/summarization flags, etc. — every field the list endpoint omits).
  No `stamp_field` needed; the transcript's own `id` is already in the response body.
- `transcript_sentence_items` — `GET /v2/transcript/{id}/sentences`, records at `sentences`,
  `stamp_field: transcript_id` (the raw sentence object has no id of its own — every emitted row
  is stamped with its parent transcript id, matching the fan-out dialect's standard "stamp the
  parent id" convention); primary key is `[transcript_id, start]` (start-offset-ms is unique per
  transcript, no other stable field is available).
- `transcript_paragraph_items` — `GET /v2/transcript/{id}/paragraphs`, records at `paragraphs`,
  same `stamp_field`/primary-key shape as sentences.
- `transcript_word_search_matches` — `GET /v2/transcript/{id}/word-search`, records at `matches`;
  the `words` query parameter is sourced from `config.word_search_terms` (a new spec.json config
  value with no server-side default — AssemblyAI's word-search endpoint requires at least one
  search term and has no documented default term list, so this bundle exposes it as an explicit,
  user-set config value rather than guessing one). Primary key is `[transcript_id, text]` (one row
  per distinct matched keyword per transcript).

Each fan-out sub-sequence pages/rate-limits/incrementally-filters independently per transcript id
(a fresh paginator per id), per the engine's fan_out dialect; none of the sub-resource endpoints
here are themselves paginated (`sentences`/`paragraphs`/`word-search` each return their full result
in one response), so no `pagination` override is declared on any of the 4 new streams.

No `incremental` block is declared on any stream (legacy-parity or new) even though `created` is
`x-cursor-field` on the legacy streams: legacy declares `CursorFields: []string{"created"}` purely
as catalog metadata but never actually filters requests by it — adding a server-side incremental
filter here would be new, legacy-diverging behavior, not a straight port. The 3 new fan-out child
streams (`transcript_detail`/`transcript_sentence_items`/`transcript_paragraph_items`) have no
cursor field of their own at all: a transcript's sentences/paragraphs/detail don't change once the
transcript reaches `completed`, so there is nothing to cursor.

## Write actions & risks

- **`create_transcript`** (`POST /v2/transcript`, `body_type: json`): submits a new transcription
  job. `audio_url` is the only required field (matching AssemblyAI's own API requirement); every
  other body field (`language_code`, `punctuate`, `speaker_labels`, `auto_chapters`,
  `sentiment_analysis`, `redact_pii`, `webhook_url`, etc.) is a documented, optional
  transcription-configuration flag passed straight through the default JSON body construction (no
  `path_fields`/`body_fields` restriction — every record field the caller sets is sent). **Risk:
  external mutation; approval required.** Submitting a transcript consumes AssemblyAI account
  balance/quota and kicks off asynchronous processing on a caller-supplied `audio_url` — this bundle
  does not poll for completion or manage the resulting job's lifecycle; the `transcript`/
  `transcript_detail` read streams are how a completed job's data is later synced back.

## Known limits

- **Subtitles (`GET /v2/transcript/{id}/{format}`) and redacted audio
  (`GET /v2/transcript/{id}/redacted-audio`) are excluded as `binary_payload`, not modeled as
  streams**: both return a raw non-JSON body (SRT/VTT caption text, or an audio binary) rather than
  a JSON record the engine's declarative read path (`connsdk.RecordsAt`, which only decodes JSON)
  can project into a schema. The pre-existing `transcript_subtitle` legacy-parity stream is a
  reference stream into the JSON transcript-list endpoint (matching legacy's own naming), NOT the
  actual subtitle text — this is unchanged from wave2 and is not a new narrowing.
- **`DELETE /v2/transcript/{id}` is excluded as `destructive_admin`**: permanently deletes the
  transcript's text/PII; not a syncable read or a safe autonomous mutation.
- **LeMUR (`/lemur/v3/generate/*`) and real-time streaming transcription (`/v2/realtime/ws`) are
  out of scope**: LeMUR is a generative summarization/QA task over already-transcribed text (an
  LLM prompt-completion product), not a transcript-domain CRUD resource — there is no
  transcript-shaped record to read or write. Real-time transcription is a stateful WebSocket
  protocol (streamed audio frames in, partial/final transcript events out), not a REST resource
  this engine's declarative HTTP dialect can express at all.
- Per `conventions.md` §4's sanctioned `next_url` exception, every list-backed stream ships a
  single-page fixture (the replay server's own absolute URL cannot be embedded in a static fixture
  ahead of time) — `pagination_terminates` exercises the `transcript` stream's 1-page fixture and
  confirms exactly one request is issued and consumed, proving the loop terminates on a null
  `next_url` rather than looping. The 4 new fan-out streams each ship a 2-page fixture (page 1 =
  the id-listing `GET /v2/transcript` request, page 2 = the one fanned-out sub-resource request for
  the single fixture transcript id) — this is the standard fan-out fixture shape (see
  `docs/migration/conventions.md` §4 and the cisco-meraki golden), not a deviation from the
  next_url exception, since the sub-resource requests themselves are not paginated.
- No live `paritytest/assemblyai` test exists for the 2-page `next_url` pagination shape or the new
  fan-out streams; fixture-replay conformance is the Pass B scope proof here.
