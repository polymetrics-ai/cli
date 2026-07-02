# Overview

DocuSeal reads document templates, submissions (signing requests), and submitters (signers on a
submission) through the DocuSeal REST API (`https://api.docuseal.com`). This bundle migrates
`internal/connectors/docuseal` (legacy) at capability parity; the legacy package stays registered
and unchanged until wave6's registry flip.

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

None of the three streams have a genuine server-side incremental filter in the legacy connector
(DocuSeal's list endpoints do not accept a since/updated-after query parameter) — legacy exposes
`CursorFields: ["updated_at"]` purely as bookkeeping metadata for a full-refresh-only source
(its own doc comment: "the cursor is informational and used only to record progress"). This bundle
mirrors that: each schema declares `x-cursor-field: updated_at` but no stream declares an
`incremental` block, so only `full_refresh_append`/`full_refresh_append_deduped` sync modes apply —
never `incremental_append*` — matching legacy's real behavior exactly.

## Write actions & risks

None. DocuSeal's write endpoints (submission creation, template upload, submitter updates) have no
legacy-parity implementation to migrate; legacy itself is read-only (`Capabilities.Write: false`),
so no `writes.json` is shipped.

## Known limits

- Only the 3 legacy-parity read streams (`templates`, `submissions`, `submitters`) are implemented;
  the full DocuSeal API surface (submission/template/submitter writes, template detail retrieval)
  is out of scope for this wave — see `api_surface.json`'s `excluded` entries.
- `page_size`/`limit` config precedence: legacy accepts either a `page_size` or a `limit` config key
  (checking `page_size` first, falling back to `limit`). This bundle exposes a single `page_size`
  spec property (the more common, self-descriptive name across this codebase's other bundles) rather
  than declaring two spec properties for the identical value — a documented, accepted config-surface
  simplification since neither name changes any emitted record data, only which config key a caller
  must set.
