# Overview

Zenefits (TriNet HR Platform) is an HR platform. This bundle reads people, companies, departments,
locations, employments, custom field definitions/values, company and employee bank accounts, labor
groups/labor group types, and time-off vacation types/requests from the Zenefits Core API
(`https://developers.zenefits.com/`). It migrates `internal/connectors/zenefits` (the hand-written
legacy connector) to a declarative Tier-1 bundle; the legacy package stays registered and unchanged
until wave6's registry flip. Pass B full-surface expansion (`api_surface.json`) confirmed, by
reading every reference page in the documentation index (`developers.zenefits.com/llms.txt`), that
the entire documented Zenefits API is GET-only — no POST/PATCH/PUT/DELETE endpoint exists anywhere
in the reference docs — so this bundle has no `writes.json` and `capabilities.write` stays `false`.

## Auth setup

Requires a single secret, `token`, sent as `Authorization: Bearer <token>` — a `bearer` auth spec,
matching legacy's `connsdk.Bearer(token)` exactly (`zenefits.go:118`). `base_url` defaults to
`https://api.zenefits.com/core` (`zenefits.go:17`'s `defaultBaseURL`), materialized via
`spec.json`'s `"default"` when unset.

## Streams notes

Thirteen streams total. The original 3 legacy-parity streams (`people`, `companies`, `departments`)
are unchanged from wave2: single-page GET reads with no pagination and no incremental support,
`records.path: "data"` against a simplified flat `{"data": [...]}` fixture shape that mirrors
legacy's own unconditional single-request behavior exactly — these fixtures/streams are
byte-identical to the pre-Pass-B bundle and were not touched.

The 10 new Pass-B streams (`locations`, `employments`, `custom_fields`, `custom_field_values`,
`company_banks`, `employee_banks`, `labor_group_types`, `labor_groups`, `vacation_types`,
`vacation_requests`) model the REAL Zenefits wire envelope honestly rather than the legacy
simplification: every real Zenefits list response is `{"data": {"data": [...], "next_url": ...,
"object": "/meta/list", "url": ...}, "error": null, "object": "/meta/response", "status": 200}`
(confirmed via `developers.zenefits.com/reference/pagination.md` and every individual resource
reference page's OpenAPI example) — so these 10 streams declare `records.path: "data.data"` and
`pagination: {"type": "next_url", "next_url_path": "data.next_url"}`, matching Zenefits' own
documented `starting_after`-cursor-via-absolute-`next_url` pagination convention (default page size
20, max 100). Fixtures for all 10 are single-page (`data.next_url: null`) per the dialect's
sanctioned `next_url` exception (§4 of the migration conventions): the replay harness cannot embed
a real absolute next-page URL in a static fixture, since that URL is the harness's own
runtime-assigned port.

Every new stream's real API record is a top-level (not nested/scoped) list endpoint — Zenefits
documents both a top-level and a parent-scoped variant (e.g. `/core/employments` vs
`/core/people/{:person_id}/employments`) for `employments`/`custom_field_values`/`labor_groups`/
`vacation_requests`/`locations`/`departments`/`people`, and the top-level path returns the
identical record set since a Zenefits access token is already scoped to one company (Zenefits'
own docs make this explicit for `company_banks`). The scoped variants are excluded as
`duplicate_of` in `api_surface.json`, not modeled as separate streams.

Nested reference objects (`person`, `company`, `vacation_type`, `custom_field`, `labor_group_type`,
`creator`, `assigned_members`, etc. — each a `{"object": "/meta/ref/detail", "ref_object": "...",
"url": "..."}` shape) are preserved as opaque JSON objects in the schema (`type: ["object",
"null"]`) rather than flattened or dereferenced — legacy never modeled these fields at all (its
`mapPerson`/`mapCompany` only copied a handful of scalar fields), so there is no legacy behavior to
match; representing them as-is is the honest, non-lossy choice for genuinely new stream coverage.

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json` — this is a real
property of the Zenefits API (confirmed via full documentation-index research, not an
unimplemented-but-possible gap): every endpoint in `developers.zenefits.com`'s reference section is
GET-only.

## Known limits

Beyond the standard scope narrowing already documented in `api_surface.json` (the `/platform/*`
application/installation-management surface and inbound event/webhook subscriptions are
out-of-scope/non_data_endpoint, since they don't concern this connector's HR-data-read purpose):

`streams.json`'s `base` declares no `check` block, so `Check()` performs no network call at all —
this matches legacy exactly (`zenefits.go:46-60`'s `Check` never calls the API; it only validates
that `base_url` parses as an absolute http(s) URL and that the `token` secret resolves to a
non-empty value). This Known Limit predates Pass B and was deliberately preserved: adding a `check`
block would be a behavior change from legacy in both directions (a syntactically-present but
server-rejected token, or any transient network/API failure, could newly fail Check where legacy
always succeeded; conversely legacy never depended on network reachability at Check time at all).

`custom_field_values`' real "detail" path is a top-level `/core/custom_field_values/{:id}` even
though its person-scoped list variant uses the singular `/core/person/{:person_id}/custom_field_values`
(not the plural `/core/people/...` every other person-scoped endpoint uses) — this is a genuine,
documented irregularity in Zenefits' own API, not a typo introduced by this bundle; recorded here so
a future reader isn't tempted to "fix" the excluded-endpoint path in `api_surface.json`.

`is_sensitive` custom fields (`custom_fields` stream) and bank account numbers/routing numbers
(`company_banks`/`employee_banks` streams) are read and emitted like any other API-returned field —
this is genuinely returned HR/payroll data behind Zenefits' own account-level scopes
(`companies`/`banks`/`company_banks`), not a connector-level secret; `spec.json`'s `x-secret`
discipline governs connection CREDENTIALS only (the bearer `token`), never stream record data.
