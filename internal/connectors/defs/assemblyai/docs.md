# Overview

AssemblyAI is a read-only declarative-HTTP bundle migrated from `internal/connectors/assemblyai`
(the hand-written legacy connector, which stays registered and unchanged until wave6's registry
flip). It reads AssemblyAI transcripts and per-transcript sentence/paragraph/subtitle references
through the AssemblyAI REST API's single `GET /v2/transcript` list endpoint.

## Auth setup

Provide an AssemblyAI API key via the `api_key` secret. It is sent as the **raw** value of the
`Authorization` header (no `Bearer` prefix — AssemblyAI's own convention), expressed as
`{"mode": "api_key_header", "header": "Authorization", "value": "{{ secrets.api_key }}", "prefix":
""}` — matching legacy's `connsdk.APIKeyHeader("Authorization", secret, "")` exactly.

## Streams notes

All 4 streams (`transcript`, `transcript_sentences`, `transcript_paragraphs`,
`transcript_subtitle`) read the identical `GET /v2/transcript` endpoint with records at
`transcripts` — legacy's `assemblyaiStreamEndpoints` table routes all four stream names to the
same `v2/transcript` resource, differing only in which fields their record mapper keeps
(`transcript` keeps `audio_url`/`completed`/`error`; the three reference streams keep only
`id`/`status`/`created`/`resource_url`). This bundle mirrors that exactly via each stream's own
schema projection (`"schema"` mode — only declared properties survive), with no `computed_fields`
needed since every kept field's name already matches the raw API key.

Pagination follows AssemblyAI's `page_details.next_url` absolute-URL convention
(`pagination.type: next_url`, `next_url_path: page_details.next_url`) — the engine's built-in
same-host SSRF guard and loop-guard (same URL requested twice) reproduce legacy's `harvest` loop,
which followed `next_url` verbatim until it was empty/null. `page_size` (default 100, legacy's
`assemblyaiDefaultPageSize`) is sent as the `limit` query param on the first request only; the
next page's `limit` (if any) is embedded in AssemblyAI's own returned `next_url` and is not
re-sent by the connector, matching legacy (`c.harvest` clears `query` to `nil` once `path` becomes
the absolute `next_url`).

No `incremental` block is declared even though `created` is documented as `x-cursor-field` in every
schema: legacy declares `CursorFields: []string{"created"}` on every `connectors.Stream` purely as
catalog metadata but never actually filters requests by it (there is no `created[gte]`-style
request param anywhere in `assemblyai.go`'s `harvest`) — adding a server-side incremental filter
here would be new, legacy-diverging behavior, not a straight port.

## Write actions & risks

None. AssemblyAI is a read-only source in both legacy and this bundle (`capabilities.write: false`,
no `writes.json`).

## Known limits

- Only the 4 legacy-parity `GET /v2/transcript`-backed streams are implemented; the full
  AssemblyAI surface (submit transcript, per-transcript sentences/paragraphs/subtitles sub-resource
  endpoints, LeMUR, real-time streaming, PII redaction, webhooks) is out of scope for this wave —
  see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Legacy itself never called the per-transcript sub-resource endpoints either
  (its "sentences"/"paragraphs"/"subtitle" streams are references into the same list endpoint, not
  the actual sentence/paragraph/subtitle content) — this bundle reproduces that exact (limited)
  legacy behavior, not a new narrowing.
- Per `conventions.md` §4's sanctioned `next_url` exception, every stream ships a single-page
  fixture (the replay server's own absolute URL cannot be embedded in a static fixture ahead of
  time) — `pagination_terminates` exercises the `transcript` stream's 1-page fixture and confirms
  exactly one request is issued and consumed, proving the loop terminates on a null `next_url`
  rather than looping. No live `paritytest/assemblyai` 2-page test exists yet for this migration
  wave (unlike bitly/calendly's dedicated live-parity tests); the single-page fixture is the
  Pass-A-scope proof here.
