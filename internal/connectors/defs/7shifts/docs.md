# Overview

7shifts is a restaurant/hospitality scheduling and labor-management platform. This bundle reads
companies, locations, departments, roles, users, shifts, and time punches through the 7shifts v2
REST API. It migrates `internal/connectors/7shifts` (the hand-written connector) at capability
parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a 7shifts API access token via the `access_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <access_token>`) and is never logged. Also provide `company_id`, the
7shifts company id every stream except `companies` is scoped under
(`/v2/company/{{ config.company_id }}/...`) — matching legacy's `endpoint.companyScoped` check,
which hard-errors a company-scoped stream read when `company_id` is unset. `company_id` is
declared `required` in `spec.json` (not merely optional-and-conditionally-used) because the
engine's path interpolation has no absent-key-tolerance mechanism: an unresolved `config.*`
reference inside a stream `path` is always a hard error, regardless of whether the key is
`required[]`-declared or not (see `docs/migration/conventions.md` §3's path-interpolation note).

## Streams notes

All 7 streams share the same shape: `GET`, records at `data`, primary key `["id"]`, incremental
cursor field `modified`. Pagination follows 7shifts's `meta.cursor.next` next-page-token
convention (`pagination.type: cursor` with `token_path: meta.cursor.next`, `cursor_param: cursor`):
the engine reads the next cursor token from the response body and stops when the token is
absent/empty, exactly matching legacy `harvest`'s `strings.TrimSpace(next) == ""` stop condition.
Every request sends `limit=100` (matches legacy's `defaultPageSize`) via each stream's static
`query: {"limit": "100"}`.

Incremental reads send `modified_since` as a bare `YYYY-MM-DD` date (`param_format: date`),
computed either from the sync's persisted cursor or, on a fresh sync, from the RFC3339 `start_date`
config value — identical to legacy `incrementalLowerBound`/`toDate`, which always reduces the
cursor or `start_date` to its date-only prefix before sending it as `modified_since` (7shifts
filters by date only, not a full timestamp).

`companies` is the one top-level (non-company-scoped) stream, matching legacy's
`streamEndpoints["companies"].companyScoped == false`; its `path` has no `company_id` template.

## Write actions & risks

None. 7shifts is read-only here (`capabilities.write: false`), matching legacy
(`Capabilities.Write: false`) exactly — `Write` returns `ErrUnsupportedOperation` on the legacy
side and this bundle ships no `writes.json` at all.

## Known limits

- `page_size`/`max_pages` config-driven overrides from legacy (`pageSizeFor`/`maxPagesFor`,
  bounded 1-200 and 0/all/unlimited respectively) are not modeled: the engine's `cursor`
  paginator's `PaginationSpec.PageSize`/`MaxPages` fields are static JSON literals with no
  templating support, so they cannot be wired to a runtime `config.*` value (unlike
  `stream.Query`, which does support templated, optionally-absent values). `limit=100` is sent as
  a fixed literal via the static per-stream `query` block, matching legacy's *default* exactly;
  a caller can no longer override the page size or cap total pages via config. This is a
  documented, accepted config-surface narrowing (`docs/migration/conventions.md` §5's
  meta-rule: it never changes emitted record DATA for any input legacy itself would accept at
  its own default, since the default behavior is preserved byte-for-byte) — declaring dead
  `page_size`/`max_pages` spec properties that no template consumes would itself violate F6.
- Full 7shifts API surface (schedules, time off, availability, wages, punch edit requests, and any
  write endpoints) is out of scope for this pass; see `api_surface.json`'s `excluded:
  {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
