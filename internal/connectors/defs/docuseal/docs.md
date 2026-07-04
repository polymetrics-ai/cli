# Overview

DocuSeal reads document templates (list + single-template detail), submissions (signing requests),
and submitters (signers on a submission) through the DocuSeal REST API (`https://api.docuseal.com`).
This bundle migrates `internal/connectors/docuseal` (legacy) at capability parity, then extends it
in Pass B with a `template_detail` read stream and 6 write actions covering every
dialect-expressible mutation in DocuSeal's published OpenAPI spec
(`https://console.docuseal.com/openapi.yml`); the legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide a DocuSeal API key via the `api_key` secret; it is sent as the `X-Auth-Token` header
(`api_key_header` auth mode) and never logged, matching legacy's
`connsdk.APIKeyHeader("X-Auth-Token", secret, "")` wiring exactly (no prefix).

## Streams notes

All 3 streams (`templates`, `submissions`, `submitters`) share DocuSeal's list envelope shape:
`GET` against the resource path, records at `data`, primary key `["id"]`. Pagination is `cursor`
with `token_path: pagination.next` and `cursor_param: after` — DocuSeal returns
`{data:[...], pagination:{count,next,prev}}`; the next page is requested with
`after=<pagination.next>`, and the engine's `tokenPathCursor` paginator stops on a null/absent/
non-advancing token exactly like legacy's `harvest` loop (`next == "" || next == "null" || next ==
after`), plus an empty-page stop. No `stop_path` is declared: legacy's stop condition is driven
purely by the token itself (never a separate boolean flag), so this bundle omits `stop_path`,
preserving that exact behavior (conventions.md §3: "a spec that never sets `stop_path` keeps the
exact prior stop-on-empty-token-only behavior").

Every request sends `limit` (default 10, matching legacy's `docusealDefaultPageSize`) via the
stream's static `query`. `submissions` additionally flattens the nested `template` object into
`template_id`/`template_name` via `computed_fields` (`{{ record.template.id }}` /
`{{ record.template.name }}`), matching legacy's `docusealSubmissionRecord` nested-field
flattening exactly.

None of the three list streams have a genuine server-side incremental filter in the legacy connector
(DocuSeal's list endpoints do not accept a since/updated-after query parameter) — legacy exposes
`CursorFields: ["updated_at"]` purely as bookkeeping metadata for a full-refresh-only source
(its own doc comment: "the cursor is informational and used only to record progress"). This bundle
mirrors that: each schema declares `x-cursor-field: updated_at` but no stream declares an
`incremental` block, so only `full_refresh_append`/`full_refresh_append_deduped` sync modes apply —
never `incremental_append*` — matching legacy's real behavior exactly.

`template_detail` (`GET /templates/{template_id}`, Pass B addition) reads a single template's full
detail object, including its `schema`/`fields`/`submitters`/`documents`/`author` nested structures
(typed as generic `object`/`array` in the schema — this dialect's draft-07 subset has no way to
project deeply nested per-field-type shapes, so these survive as opaque JSON exactly as the API
returns them). Requires the new `template_id` config field; `pagination: {"type": "none"}` overrides
the base cursor pagination for this single-object stream, matching the `namespace`-style pattern
used elsewhere in this codebase (see dockerhub).

## Write actions & risks

6 write actions cover every dialect-expressible (plain-JSON-body, non-binary) mutation in DocuSeal's
published OpenAPI spec:

- `create_submission` (`POST /submissions`) — creates a live signature request from a template and
  submitter list. **Dispatches real emails/SMS to every listed submitter unless `send_email`/
  `send_sms` are explicitly set `false`** — the single highest-risk action in this bundle.
- `archive_submission` (`DELETE /submissions/{id}`) — archives (soft-deletes) a submission.
- `update_submitter` (`PUT /submitters/{id}`) — overwrites a submitter's contact info/pre-filled
  values; can re-send notification emails/SMS and can force-mark a submitter completed/auto-signed.
- `update_template` (`PUT /templates/{id}`) — renames/moves a template, updates its role list, and
  can unarchive it (`archived: false`).
- `archive_template` (`DELETE /templates/{id}`) — archives (soft-deletes) a template.
- `clone_template` (`POST /templates/{id}/clone`) — creates a new template by cloning an existing
  one.

All 6 are plain declarative JSON-body writes (`body_type: json`, `path_fields: ["id"]` where the
action targets an existing resource) — no `WriteHook`/compound-write logic was needed since every
DocuSeal mutation this pass covers is a single request with no follow-up call. Every action carries
`risk: "external mutation; ... approval required"` per this codebase's write-action convention; none
are auto-approved.

## Known limits

- `templates`/`submissions`/`submitters`/`template_detail` (read) and `create_submission`/
  `archive_submission`/`update_submitter`/`update_template`/`archive_template`/`clone_template`
  (write) are implemented — every endpoint in DocuSeal's published OpenAPI spec that is expressible
  as a plain-JSON declarative read/write. Excluded: single-submission/single-submitter detail GETs
  (no new data over the list streams' schema-projected fields; not modeled as their own streams this
  pass), `GET /submissions/{id}/documents` (binary document content, not tabular data), and every
  PDF/DOCX/HTML document-authoring endpoint (`/submissions/pdf`, `/submissions/docx`,
  `/submissions/html`, `/submissions/emails`, `/templates/pdf`, `/templates/docx`,
  `/templates/html`, `/templates/merge`, `PUT /templates/{id}/documents`) — these require either a
  base64/URL-embedded binary file payload (multipart/binary bodies are a hook-only capability per
  conventions.md §1, and no hook package is added in this pass) or are document-composition
  operations with no natural single-record write shape. See `api_surface.json`'s `excluded` entries
  for the complete, individually-reasoned list.
- `page_size`/`limit` config precedence: legacy accepts either a `page_size` or a `limit` config key
  (checking `page_size` first, falling back to `limit`). This bundle exposes a single `page_size`
  spec property (the more common, self-descriptive name across this codebase's other bundles) rather
  than declaring two spec properties for the identical value — a documented, accepted config-surface
  simplification since neither name changes any emitted record data, only which config key a caller
  must set.
