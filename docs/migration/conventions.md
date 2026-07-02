# Migration conventions — the single recipe

Read this whole file before touching a connector. It is the ground truth every fan-out migration
agent and reviewer follows; deviations are defects, not judgment calls. Goldens (read, don't copy
blindly — port the *pattern*): `internal/connectors/defs/stripe/**` (declarative HTTP + writes),
`internal/connectors/defs/searxng/**` (read-only, no-auth), `internal/connectors/defs/postgres/**`
+ `internal/connectors/native/postgres/**` (Tier-3 split). Engine source of truth:
`internal/connectors/engine/{bundle,interpolate,paginate,read,write,hooks,schema}.go`.

## 1. Target layouts

### Tier 1 — declarative bundle (target ≥90% of connectors)

One directory `internal/connectors/defs/<name>/`, zero Go:

```
metadata.json        # identity + capabilities + risk (engine/bundle.go Metadata)
spec.json            # draft-07 connection spec; x-secret marks secret fields
streams.json         # base HTTP config + streams[] (required unless dynamic_schema)
writes.json          # actions[] (omit entirely when capabilities.write is false)
api_surface.json     # coverage manifest (always required)
schemas/<stream>.json  # one draft-07 schema per stream, x-primary-key/x-cursor-field
fixtures/
  check.json
  streams/<stream>/page_1.json (page_2.json ... when paginated)
  writes/<action>.json
docs.md              # Overview / Auth setup / Streams notes / Write actions & risks / Known limits
```

Worked example — **stripe** (`internal/connectors/defs/stripe/`): `metadata.json` declares
`capabilities.write: true`; `spec.json` has one `x-secret` field (`client_secret`) plus
`base_url`/`account_id`/`start_date`/`page_size`/`max_pages`/`mode`; `streams.json`'s `base` sets
`url`, a conditional `Stripe-Account` header (omitted when `account_id` is unset — see §3), bearer
`auth`, a `cursor` pagination block (`last_record_field: id`, `stop_path: has_more` — Stripe's
`starting_after`/`has_more` convention), a `check` request, and an `error_map`; each of the 5
streams shares the identical shape (`GET`, `records.path: "data"`, `incremental.cursor_field:
created`, `param_format: unix_seconds`) — copy this shape for any list-endpoint API with a uniform
envelope. `writes.json` declares `create_customer`/`update_customer`, both `body_type: form`,
`update_customer` carrying `path_fields: ["id"]`. `fixtures/streams/customers/{page_1,page_2}.json`
is the **required 2-page fixture** for a paginated stream (§4). `docs.md` documents the
`minProperties: 1` parity deviation (§5, item 1) inline as well as in this ledger.

### Tier 1 — read-only, no-auth variant

Worked example — **searxng** (`internal/connectors/defs/searxng/`): `metadata.json` has
`capabilities.write: false` and no `writes.json` file at all. `spec.json` declares only `base_url`
(required) and `api_key` (`x-secret`, optional) — every OTHER config property this bundle once
declared (`subreddit`/`categories`/`engines`/`language`/`time_range`/`safesearch`/`page_size`/
`max_pages`) was removed as genuinely-dead config (F6, REVIEW.md): `stream.Query` templating has no
absent-key-falsy tolerance, so an optional passthrough filter cannot be wired without guaranteeing
the caller always sets it, and `page_size`/`max_pages` have no runtime config-driven override
mechanism at all (they are fixed values baked into `streams.json`'s `base.pagination` block). A
`spec.json` property that no template anywhere in the bundle consumes should not be declared — see
§3's "Conditional headers"/query-templating note. `api_key`, by contrast, IS wired: `streams.json`
`base.auth` gates a `bearer` spec on `{{ secrets.api_key }}`'s truthiness, falling back to `none`
when unset (§3's `when`-grammar absent-key-falsy tolerance makes this safe). Pagination is
`page_number` with `size_param` intentionally empty (`""`) — no page-size query param is ever
sent; `page_size` in `streams.json`'s base pagination block exists purely as the client-side
short-page stop threshold, and `max_pages: 1` IS enforced as a hard request-count cap by the engine
read path. `streams.json`'s two streams differ only in query scoping (`q` vs `site:reddit.com {{
config.query }}`) and both use `computed_fields` to rename the raw API's `publishedDate` to the
schema's `published_date`, join the raw `engines[]` array into a comma-separated string
(`join:,`), and stamp a static-literal `stream` marker field (`"search"`/`"reddit"`) — see §2 and
§3.

### Tier 3 — native component split

Non-REST protocols (SQL, CDC, message queues, filesystems) implement `connectors.Connector`
directly under `internal/connectors/native/<name>/`, following the mandated Ruby-style component
split (design §B.7), each file well under ~400 lines:

- `connector.go` — entry/wiring: `type Connector struct{ engine.Base }`, `New()` (loads the defs
  bundle via `engine.Load`, panics on load failure — a broken build, not a runtime error),
  `Metadata()` override (only refine `Description`/`DisplayName` wording; capability flags stay
  sourced from `metadata.json`), `Write()` stub if read-only.
- `connection.go` — config struct, DSN/connection-string builder, `resolveConfig` validation,
  identifier-safety helpers (SQL injection / SSRF guards), `Check()`.
- `reader.go` — `Read()`, snapshot/incremental query builder, `InitialState()`.
- `cataloger.go` — `Catalog()` / dynamic schema discovery.
- `cdc.go` — `ReadCDC()` (a documented stub is acceptable when the CDC dependency is gated — see
  postgres below).

Still ship a bundle (`internal/connectors/defs/<name>/{metadata.json,spec.json,api_surface.json,
docs.md}`) so identity/spec/docs stay uniform with every other connector; `metadata.json` sets
`capabilities.dynamic_schema: true` and the bundle ships **no `streams.json`** (the loader
(`bundle.go`'s `loadStreams`) only tolerates a missing `streams.json` when `dynamic_schema` is
true). The package embeds `engine.Base` (via `engine.NewBase(bundle)`) purely to serve
`Name()`/`Metadata()`/`Definition()`; `engine.Base` does **not** provide
Check/Catalog/Read/Write — those remain hand-written Go.

Worked example — **postgres** (`internal/connectors/native/postgres/` +
`internal/connectors/defs/postgres/`): `api_surface.json` declares `endpoints: []` with a `scope`
prose explaining there is no REST surface to enumerate (schema-valid: no `minItems` on
`endpoints`); `spec.json` has `password` marked `x-secret` and a `mode: fixture` config value that
short-circuits all network access for credential-free testing (test/conformance-harness affordance
only, never set in production); `cdc.go`'s `ReadCDC` is a **documented stub** returning
`connectors.ErrUnsupportedOperation` wrapped with the recorded future implementation plan (gated
`pglogrepl` dependency, not present in `go.mod`) — this is the sanctioned way to represent an
out-of-scope capability without faking it. The package has **no `init()`/`RegisterFactory`/
`RegisterNativeLive` call** in wave0 (registration flip is wave6); a grep-guard test enforces this.

### Tier 2 — hooks (bundle + Go escape hatch, target ~8%)

The bundle still declares every stream/write/schema; a single `internal/connectors/hooks/<name>/
hooks.go` (+ `hooks_test.go`) attaches Go at named extension points only, hard-capped at ~300
lines (design §F.1) — past that, escalate to Tier 3. Five interfaces
(`internal/connectors/engine/hooks.go`), a hook set implements any subset:

| Interface | Signature | Use for |
|---|---|---|
| `AuthHook` | `Authenticator(ctx, cfg, spec AuthSpec) (connsdk.Authenticator, error)` | signature auth (SigV4, HMAC), token-exchange auth (GitHub App JWT→installation token) |
| `RecordHook` | `MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error)` | per-record post-processing beyond schema projection |
| `StreamHook` | `ReadStream(ctx, stream, req, rt *Runtime, emit) (handled bool, err error)` | whole-stream override: async report jobs, CSV downloads, sub-resource fan-out (issue → comments per issue) |
| `WriteHook` | `ExecuteWrite(ctx, action, rec, rt *Runtime) (handled bool, err error)` | compound/multi-request writes (create_pull_request + reviewer follow-up) |
| `CheckHook` | `Check(ctx, cfg, rt *Runtime) (handled bool, err error)` | non-declarative health check |

Legitimate Tier-2 triggers (design §B.7) — anything else belongs in Tier 1 (JSON) or, if still
insufficient, is an `ENGINE_GAP`/Tier-3 escalation, never invented Go in a declarative field:
signature auth, token-exchange auth, multipart/XML bodies, async report polling, response
decompression/CSV parsing, sub-resource fan-out reads, compound writes. A hook returning
`handled=false` falls back to the declarative path for that dispatch point — hooks are additive,
not a full override by default.

## 2. Authoring rules

- **Naming** (design §F.3): connector name = directory name = `metadata.json.name` = registry key,
  matching `^[a-z0-9][a-z0-9-]*$` (enforced by `engine/bundle.go`'s `namePattern` at load time —
  mismatch is a hard load error, not a validate warning). Stream and write-action names are
  `snake_case`. Write actions are verb-first: `create_customer`, `update_customer`,
  `merge_pull_request` — never `customer_create` or bare nouns.
- **`spec.json` x-secret discipline**: every credential-shaped field (API keys, tokens, passwords,
  client secrets) is `x-secret: true` in `spec.json`, never a plain `properties` entry. Only
  `x-secret` fields end up in `Schema.SecretKeys()`, which governs what is redacted from
  `DryRunWrite` previews (see §3) and never logged. A field that merely *looks* sensitive but is
  documentation-only (an optional Bearer-proxy key never wired into `auth`, e.g. searxng's
  `api_key`) is still marked `x-secret: true` — the marker is about the *field's nature*, not
  whether this bundle currently exercises it.
- **Schema-as-projection**: a stream's `schemas/<stream>.json` `properties` set is derived
  **field-for-field** from what the legacy connector's own `mapRecord`/record-shaping function
  actually emits — not from guessing the raw API shape. In `"schema"` projection mode (the
  default; see §3), only declared properties survive; anything legacy emitted that this bundle's
  schema omits is silently dropped from parity. When a legacy field name differs from the raw API
  field (e.g. searxng's `published_date` vs the raw `publishedDate`), add a `computed_fields`
  rename (§3) — don't just omit it. Every schema declares `x-primary-key` and, when the stream is
  incremental, `x-cursor-field`; both must name properties that actually exist in that same schema
  (`connectorgen validate`'s `primary_key_missing`/`cursor_field_missing` rules, and
  `conformance`'s static `pk_fields_exist`/`cursor_fields_exist` checks — same underlying
  requirement, two differently-named rule sets — enforce this).
- **Sync-mode derivation — never declared** (design §B.6): `full_refresh_append`/
  `full_refresh_overwrite` always apply; `*_deduped` variants apply iff `x-primary-key` is
  present; `incremental_append[_deduped]` applies iff the stream has an `incremental` block. Do
  not add a "supported_sync_modes" field anywhere — there isn't one in this dialect; the engine
  derives it from schema/stream shape at runtime.
- **`api_surface.json` depth — minimal-honest for wave0/pilot** (DECISIONS.md #4): list every
  implemented stream/write under `covered_by`; everything else documented as `excluded: {category:
  "out_of_scope", reason: "Pass B capability expansion"}` (see stripe's `api_surface.json` for the
  pattern — 5 covered streams, 2 covered writes, the remaining known Stripe surface excluded
  out-of-scope, one `non_data_endpoint` for `/v1/balance`). Full API-surface research (every
  documented endpoint actually implemented) is Pass B (wave5), not wave0/pilot/Pass-A fan-out. The
  closed exclusion-category vocabulary (design §E.1 rule 3, enforced by the loader's meta-schema
  enum): `destructive_admin`, `requires_elevated_scope`, `binary_payload`, `deprecated`,
  `non_data_endpoint`, `duplicate_of`, `out_of_scope`.
- **`docs.md` required headings** (exact text, `#`/`##` either level; `conformance`'s
  `docs_present` and `connectorgen validate`'s `docs_heading` rule both check presence by trimmed
  text only): `Overview`, `Auth setup`, `Streams notes`, `Write actions & risks`, `Known limits`.
  `Known limits` is not decorative — every deliberate simplification, `ENGINE_GAP`, or
  scope-narrowing goes there (see the parity-deviation ledger, §5) with enough detail that a
  reviewer or a future capability-expansion agent doesn't have to re-derive the reasoning.

## 3. The engine dialect reference

All of this is read from `internal/connectors/engine/{bundle,interpolate,paginate,read,write,
schema}.go` directly — this section is a map, not a spec; when in doubt, read the source.

**Template references + filters** (`interpolate.go`). `{{ <ref> | <filter1> | <filter2> ... }}` —
**a filter chain of any length** applies every stage left-to-right (not just one filter); an
unknown filter name anywhere in the chain is a hard error (validate-time via `ResolveCheck`'s
filter-name check, wired into both `connectorgen validate` and static-checked chains, and
runtime). References: `config.<key>`, `secrets.<key>`, `record.<dotted.path>` (walks nested
`map[string]any`), bare `cursor`. Filters: `urlencode` (percent-encodes for path-segment
insertion; applied **by default** to every `InterpolatePath` resolution unless an explicit filter
chain overrides it — general `Interpolate`/header interpolation do NOT default-apply it),
`unix_seconds` (RFC3339 string → Unix seconds integer string), `base64` (`base64.StdEncoding`),
`join:<sep>` (joins an array-valued reference — e.g. `record.tags` when it resolves to a raw
`[]any` — with `<sep>`; a non-array value is a hard error, not a silent stringify; the separator is
everything after the first `:`, so `join:, ` or a multi-character separator both work
unambiguously, but a literal `|` cannot be the separator since the outer chain-split takes
precedence). Any resolved value destined for a header (`InterpolateHeader`) or that itself contains
`\r`/`\n` is rejected outright as a header/injection guard (THREAT-MODEL §2) — this check runs on
the **pre-filter** value for every interpolation call, not just headers. An absent `config.*`/
`secrets.*` key is a **hard error** naming the key and namespace in ordinary `Interpolate`/
`InterpolatePath`/`InterpolateHeader` calls — do not rely on a missing key silently becoming `""`
outside `when` conditions (next paragraph).

**Path interpolation** (`InterpolatePath`, wired into BOTH reads and writes): a stream's `path` and
`check.path` are interpolated exactly like a write action's `path` — `{{ }}` templates are legal in
any of them, urlencoded by default per-segment. A resolved path segment that is exactly `..`, or
that (after percent-decoding) contains `/../`, ends in `/..`, or starts with `../`, is rejected
outright even though the slash itself is already percent-encoded — closing the "single dot-dot
segment survives as an intact same-value encoded literal" traversal gap.

**`when` grammar** (`interpolate.go`'s `EvalWhen`, used only by `auth.go`'s `selectAuth`/
`authSpecMatches` today): `config.k == 'literal'` (equality against a single-quoted string
literal), `config.k in ['a','b']` (membership), or a bare `config.k`/`secrets.k` (truthiness — any
non-empty resolved string is true). **ABSENT-KEY-FALSY semantics apply only in `when`, never in
general interpolation**: a `config.*`/`secrets.*` reference whose key is entirely absent at
runtime resolves to `""` (empty string) here — truthiness is false, `==` compares against `""`,
`in [...]` treats it as not-contained unless the list itself contains the empty-string literal.
This is what makes an *optional* credential/config-gated `auth` candidate possible without a
separate "is this key present" primitive — e.g. an optional Bearer-proxy secret:
`[{"mode":"bearer","token":"{{ secrets.api_key }}","when":"{{ secrets.api_key }}"},{"mode":"none"}]`
safely falls through to `none` when `api_key` is unset, and applies the bearer token when it is set
(see searxng's `streams.json`). Static validation (`ResolveCheck`, run by `connectorgen
validate`/conformance's `interpolations_resolve`) is unaffected by this runtime-absence tolerance: a
`when` template referencing a key **not even declared** in `spec.json`'s properties is still a hard
validate-time error. `connectorgen validate` validates EVERY templated `AuthSpec` field this way —
not just `token`/`value`/`when` but also `username`/`password`/`token_url`/`client_id`/
`client_secret`/`scopes` — via `engine.ResolveCheckAuthSpec(spec, specKeys)`.

**Pagination — 6 types + none** (`bundle.go`'s `PaginationSpec`, `paginate.go`'s `newPaginator`):

| `type` | Fields used | When to use |
|---|---|---|
| `none` (default/omitted) | — | single-page endpoints |
| `link_header` | — | RFC 5988 `Link: <url>; rel="next"` header (GitHub-style) |
| `page_number` | `page_param`, `size_param` (empty string = never send a size param), `start_page`, `page_size` | 1-based/N-based page-number APIs; short-page stop when a page returns fewer than `page_size` records |
| `offset_limit` | `limit_param`, `offset_param`, `page_size` | offset+limit APIs |
| `cursor` (`token_path`) | `cursor_param`, `token_path` | next-page token read from the response body |
| `cursor` (`last_record_field`+`stop_path`) | `cursor_param`, `last_record_field`, `stop_path` | Stripe-style `starting_after`/`has_more`: next cursor = a named field on the **last record** of the current page; `stop_path` names a body path whose falsy value stops (its absence, or an empty/malformed page, is defensive "never loop forever") |
| `next_url` | `next_url_path`, `allow_cross_host` | absolute next-page URL read from the body (aircall-style); same-host SSRF guard by default (THREAT-MODEL §3) — set `allow_cross_host: true` to opt out; also loop-guards against requesting the same URL twice |

`cursor`'s `token_path` and `last_record_field` are **mutually exclusive**; declaring both, or
neither, is a **read-time** error, not a load-time one — `paginate.go`'s `newPaginator` (which
enforces this) runs from `read.go`'s `newRuntime`, once per `Read`/`Check` call, not once at bundle
load. `connectorgen validate` does not check pagination specs at all (no field/rule references
`PaginationSpec` anywhere in `cmd/connectorgen/validate.go`) — a malformed `token_path`+
`last_record_field` combination passes `connectorgen validate` cleanly and only surfaces the first
time the stream is actually read. `MaxPages` is a hard request-count cap enforced in `read.go`'s
`readDeclarative` loop, independent of page fullness, checked *before* issuing the request for
that page number; `MaxPages <= 0` (absent/zero) is unbounded. Stream-level `pagination` replaces
the base-level spec **wholesale** (no field-by-field merge) when present.

**`param_format`** (`incremental.param_format`, applied by `read.go`'s `formatParam` to the
RFC3339 lower-bound value before sending it as `request_param`): `rfc3339` (default; send
verbatim), `unix_seconds`, `date` (`2006-01-02`), `github_date_range` (`>=<value>`, a
lower-bound-only GitHub search qualifier).

**`computed_fields`** (`streams.json` per-stream): a map of output-field-name → template resolved
against the **raw** (pre-projection) record, so it can rename a field (`published_date` from
`{{ record.publishedDate }}`), reach into nested JSON (`{{ record.user.login }}`), join an array
(`{{ record.engines | join:, }}`), or inject a **static literal** — a value with no `{{ }}` markers
at all (e.g. `"stream": "search"`) passes through `Interpolate` verbatim, since a template with no
markers is a no-op by construction. This is the sanctioned way to stamp a per-stream constant onto
every emitted record (e.g. "which stream did this come from", matching a legacy connector's own
derived marker field — see searxng's `stream` field, §5). A computed field whose source path is
absent on a given record is silently skipped for that record (not an error) — common for
optional/differently-shaped nested fields across record variants (issues vs PRs); this only applies
to `record.*`-templated fields, not static literals (which never fail to resolve).

**Conditional headers — omission semantics are namespace-scoped, not blanket-tolerant**
(`read.go`'s `resolveHeaders`/`classifyHeaderResolutionError`): a header template resolving to an
unresolved key is handled by a decision table, not a single blanket rule:
- `secrets.*` (any key) — **always a hard error**. A header templating an absent secret (e.g.
  `Authorization: Bearer {{ secrets.token }}` with no `token` configured) is NEVER silently
  omitted — that would send the request unauthenticated instead of failing loudly (F4,
  REVIEW.md/SECURITY-REVIEW.md). Prefer an `auth` spec over a declared header for anything
  secret-bearing.
- `config.*`, key **declared in `spec.json` but NOT in `required[]`** — **omitted** (the
  Stripe-Account/`account_id` pattern: an optional per-account header with no value configured is
  left off entirely, not sent empty).
- `config.*`, key in `required[]` or **not declared in `spec.json` at all** — **hard error**
  (required-but-missing, or a reference to an undeclared key).
- Any other interpolation failure (CRLF, unknown namespace/filter) always propagates as a hard
  error regardless of namespace.

`stream.Query` templating (per-request query params) has **no such omission tolerance at all** —
every `{{ }}` reference in a `query` map value is resolved unconditionally via `Interpolate`, so an
absent optional `config.*` key referenced there is always a hard error. This is why a query-driven
"optional passthrough filter" (e.g. searxng's would-be `categories`/`language`/`time_range`/etc.)
cannot be expressed today without guaranteeing the caller always supplies a value — do not declare
such a `spec.json` property unless the query template can actually tolerate its absence (currently:
never, for `query`; `auth`'s `when` is the only templating surface with absent-key-falsy
tolerance). A `spec.json` property that no template anywhere in the bundle ever consumes is dead
config — don't declare it (F6, REVIEW.md).

**`client_filtered`** (`incremental.client_filtered: true`): for APIs with no server-side
incremental filter parameter, the engine drops already-seen records client-side by comparing each
record's cursor field against the lower bound (strictly greater survives). Use only when the API
truly cannot filter server-side — prefer `request_param` whenever the API supports it.

**Projection**: `stream.projection` is `"schema"` (default: only schema-declared properties
survive) or `"passthrough"` (every raw field survives, unfiltered). `computed_fields` are always
applied after projection regardless of mode.

**`MaxPages` hard cap**: see the pagination table above; this is the only page-count bound the
engine enforces on the declarative read path (a short/empty final page from the paginator is the
*other* stop signal, and both must be considered independently when reasoning about termination).

**`metadata.json`'s `rate_limit` is informational-only, NEVER enforced** (F6, REVIEW.md):
`Metadata.RateLimit.RequestsPerMinute` documents a connector's published rate limit for operator
awareness but is never consumed by the read path. The ONLY rate-limit field the engine actually
enforces is `streams.json`'s `base.rate_limit` (`HTTPBase.RateLimit`, a distinct field —
`read.go` reads only `b.HTTP.RateLimit`) — declare it there, not in `metadata.json`, if a connector
genuinely needs client-side throttling. Do not add a `streams.json` `rate_limit` block for a
connector whose legacy implementation enforced no client-side rate limit (that would be new,
behavior-changing throttling introduced under the guise of a migration, not a parity port) — an
informational `metadata.json.rate_limit.requests_per_minute` with no `streams.json` counterpart is
the correct, honest representation of "legacy documented this limit but never enforced it
client-side" (see stripe's `docs.md`). Any key on `metadata.json.rate_limit` beyond
`requests_per_minute` (e.g. a `strategy` field) is not even a field on the Go type and is silently
dropped — don't declare one.

**Write body construction** (`write.go`): `body_type` is `"json"` (default), `"form"`, or
`"none"`. Default body = every record field **except** those named in `path_fields` (the path
already carries them, e.g. `{{ record.id }}` for an update). `body_fields` (if set) restricts the
body to an explicit allow-list instead (used for delete-with-body actions). `"form"` builds a
`url.Values` body (Stripe-shape — compare `stripe/write.go`'s `customerForm`), sorted keys for
deterministic encoding, empty-string values omitted. `"none"` with no `body_fields` sends no body
at all (pure path-parameterized mutation/delete).

**Delete semantics**: `kind: "delete"` + `delete.missing_ok_status: [404, ...]` means those HTTP
statuses on the delete request count as **written, not failed** (idempotent delete) — any other
status, or an unlisted 404, is a genuine per-record failure. `Write`'s overall accounting is
fail-fast, matching legacy (e.g. `stripe/write.go:66`): on the first real failure (validation, a
per-record request error, or ctx cancellation), the loop stops immediately;
`RecordsWritten`/`RecordsFailed` reflect exactly what completed, not a best-effort continuation.

## 4. Fixture rules

- Fixtures are **recorded-real-shape, sanitized** pages: capture what the live API actually
  returns (field names, nesting, null-vs-absent behavior), then replace every real value with a
  synthetic one (`cus_fixture_1`, `fixture1@example.com`, `+15550100`, `2026-01-01T00:00:0*Z`
  timestamps). Never commit a real ID, name, email, token, or account identifier.
- **A 2-page fixture is REQUIRED whenever the bundle declares pagination** for at least one
  stream (`conformance`'s `pagination_terminates` dynamic check needs a second page to prove the
  engine consumes each page exactly once and terminates). See
  `fixtures/streams/customers/{page_1,page_2}.json` in the stripe golden: page 1 sets
  `has_more: true` and 2 records; page 2's request carries the expected `starting_after` query
  param and sets `has_more: false`.
- **Fixtures use the API's REAL wire shape for every field, including cursor fields — no
  string-ification workaround, ever** (B2, REVIEW.md, now RESOLVED). A numeric cursor field
  (Stripe's `created` is a Unix-seconds integer) is committed as a bare JSON NUMBER in
  `fixtures/streams/**`, exactly matching the "recorded-real-shape" rule above — there is no
  carve-out for cursor fields specifically. `conformance/dynamic.go`'s `checkCursorAdvances`
  recognizes BOTH shapes: a `string` cursor value (compared/maxed lexicographically — correct for a
  fixed-width RFC3339 representation) and a `json.Number`/`float64` cursor value (compared/maxed
  numerically, then canonicalized to a digit string before formatting), so a numeric fixture value
  is fully supported, not a hard-fail. Schema types stay tight (`"integer"`, not
  `["integer","string"]`) — declare the field's REAL type, do not widen it to accommodate a
  conformance limitation that no longer exists.
- `fixtures/check.json`: `{"request": {...}, "response": {"status": 200, "body": {...}}}` — used
  by `check_fixture`.
- `fixtures/writes/<action>.json`: `{"record": {...}, "expect": {"method", "path", "body"?},
  "response"?: {"status", "body"}}` — used by `write_validate`/`write_request_shape`; the engine's
  dry-run/actual request must match `expect` exactly for a valid `record`, and a deliberately
  invalid record (missing a required field) must fail validation in its own dedicated fixture/test
  case. The optional `response` block (R3) lets you declare what the write-replay capture server
  answers with, instead of the default `200 {}` — needed whenever a `WriteHook`'s follow-up logic
  reads its own write response (github's `create_pull_request` fixture declares `"response":
  {"status": 201, "body": {"number": 42}}` because `hooks/github/hooks.go`'s `createPullRequest`
  decodes the POST response's `number` field before issuing its follow-up requests; a fixture with
  no `response` block is unaffected, still gets `200 {}`).
- **No secret-looking values, ever.** `connectorgen validate`'s `secret_literal` rule and
  `conformance`'s `secret_redaction` check both regex-scan every fixture file (and `docs.md`) for
  Bearer-header shapes, `api_key`/`access_token`/`secret`/`password`-adjacent long tokens, and
  vendor-recognizable prefixes (e.g. `sk_live_`/`sk_test_`) — a match is a hard validate failure,
  not a warning. Keep fixture "secrets" obviously synthetic (`sk_test_FAKE...` still trips the
  scanner deliberately — use `fixture_token_placeholder`-style strings instead, or better, don't
  put secret-shaped strings in fixtures at all since fixtures never carry auth).

### Conformance is hook-aware; the skip-marker rule replaces the old "shadow path" pattern (R3)

`conformance`'s dynamic (fixture-replay) checks run the REAL engine, including any registered
Tier-2 hook (`engine.HooksFor(b.Name)`, `conformance` blank-imports `hooks/hookset`) — a hook that
CAN run against a declarative-shaped fixture replay is now genuinely exercised, not silently
bypassed. **github is the worked example**: its `AuthHook` (token-or-app_jwt "auto" resolution) and
`WriteHook` (the 4 compound write actions) both run for real against conformance's replay harness
and are fully covered by it — prefer this outcome (full dynamic coverage) wherever a hook CAN run
in replay; do not add a skip marker "just in case" or to save fixture-authoring effort.

Some hooks genuinely **cannot** be exercised by a declarative fixture replay no matter how the
fixture is shaped — a custom-auth `AuthHook` whose real request needs a config value (e.g.
`token_url`) that conformance's synthetic non-secret config (`"synthetic-conformance-value"` for
every non-x-secret spec property) can never meaningfully populate, or a `StreamHook` whose real
wire shape (a POST body carrying query text, in-body pagination state) has no declarative
equivalent for the replay server to match against at all. For exactly this case, declare an
OPTIONAL, EXPLICIT skip marker instead of shaping a fictional "shadow" fixture just to fool the
replay harness (the pre-R3 pattern this rule replaces):

- **Stream-level** (`streams.json`, one stream): `"conformance": {"skip_dynamic": true, "reason":
  "..."}` on that stream's object. Skips that stream's `read_fixture_nonempty:<name>` and excludes
  it from every other check's candidate-stream selection (`pagination_terminates`'
  first-eligible-stream pick, `records_match_schema`'s per-stream iteration, `cursor_advances`'
  first-incremental-stream pick) — as if the stream did not exist for dynamic-check purposes. A
  bundle with OTHER, unmarked streams still runs their dynamic checks normally (this is the monday/
  sentry shape: every stream shares the same StreamHook limitation, so every stream is marked, but
  the mechanism is genuinely per-stream).
- **Bundle-level** (`metadata.json`, top level): the identical `"conformance": {"skip_dynamic":
  true, "reason": "..."}` shape. Skips EVERY auth-dependent dynamic check outright
  (`check_fixture`, every `read_fixture_nonempty:<stream>`, `pagination_terminates`,
  `records_match_schema`, `cursor_advances`) — this is gmail's shape: a sole `mode: custom` auth
  candidate with no `when`-gated non-custom fallback means literally every dynamic check that
  resolves auth would otherwise fail identically and uninformatively.
- **`reason` is REQUIRED whenever `skip_dynamic` is `true`** — `connectorgen validate`'s
  `conformance_skip_reason` rule hard-fails a marker with an empty/whitespace-only reason, at
  either level. The reason MUST name the authoritative substitute that actually proves the skipped
  behavior, not just assert the skip — e.g. `"hook-covered; proven live by
  internal/connectors/paritytest/<name>"`, optionally citing the specific test
  (`TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop`). **The parity suite
  (`paritytest/<name>`) is the authoritative correctness bar for any hook-covered behavior a skip
  marker names** — a marked connector's `migrated` status still requires its parity suite (and the
  hook's own unit tests) to be green; the marker is an honest description of WHERE the proof lives,
  never a way to avoid proving the behavior at all.
- A `CheckResult` produced by a marker is always `Skipped: true` with `Error` set to the marker's
  `reason` text — never `Passed` (a skip is not a pass) and never a silent, no-reason `Skipped`
  (every pre-existing structural Skip — e.g. "no incremental stream" — still carries no reason;
  only a MARKER-caused Skip does, so a report reader can immediately tell the two apart).
- Do not reach for a skip marker to avoid writing a genuinely fixable fixture. github's own
  `write_request_shape:create_pull_request` failure (the write-replay capture server always
  answered `200 {}`, so the `WriteHook`'s post-POST `number` decode always saw `0`) was a real
  fixture bug, not a hook-vs-replay mismatch — it was fixed by declaring the fixture's `response`
  block (see above), not by adding a marker. A marker is for behavior a replay genuinely cannot
  reach, not a substitute for fixture-authoring diligence.

## 5. Parity-deviation ledger

**Meta-rule**: a deviation from legacy behavior is **ACCEPTABLE** iff it never changes the emitted
record DATA for any input legacy itself would accept, AND it is documented both in the migration
agent's `result.schema.json` `parity_deviations[]` entry and in this ledger. Anything that *would*
change accepted-input behavior, or that the engine genuinely cannot express at all, is an
`ENGINE_GAP` blocker (§6) — never a silent workaround.

| id | connector | description | verdict |
|---|---|---|---|
| 1 | stripe | `create_customer`'s legacy "email OR name required" rule (a named-field OR) is approximated by `minProperties: 1` over the four optional fields (`email`,`name`,`description`,`phone`) — the engine's draft-07 subset has no `anyOf`/`oneOf`. Strictly more permissive (a record with only `phone` set passes here, would fail legacy); never stricter; parity tests only exercise legacy-valid records. | ACCEPTABLE |
| 2 | stripe | ~~Bundle's own `fixtures/streams/**` represented the `created` cursor as an RFC3339 string rather than Stripe's real Unix-seconds wire integer, to work around `conformance`'s `cursor_advances` string-only type assertion.~~ **RESOLVED**: `cursor_advances` now recognizes both string AND numeric (`json.Number`/`float64`) cursor values (B2, REVIEW.md); fixtures were rewritten to Stripe's real numeric wire shape and the schema type tightened back to `"integer"`-only (no more `["integer","string"]` widening). See §4. | RESOLVED |
| 3 | stripe | ~~`limit=100` was sent via a static per-stream `query: {"limit": "100"}` while `pagination.limit_param`/`page_size` were ALSO declared on the spec despite being dead (the `cursor`+`last_record_field` paginator constructor never reads them).~~ **RESOLVED**: the dead `limit_param`/`page_size` fields were removed from `streams.json`'s `base.pagination` block (F6, REVIEW.md) — `limit=100` still flows via the static per-stream `query`, unchanged, but the bundle no longer declares config that does nothing. See §3's rate_limit rule for the same "informational vs. enforced" distinction applied to `metadata.json.rate_limit`. | RESOLVED |
| 4 | searxng | ~~Raw API's `engines[]` array was passed through unjoined (schema typed it `["array","string","null"]`) rather than legacy's comma-joined string, because the dialect had no array-join filter.~~ **RESOLVED**: the engine's `join:<sep>` filter (R1) lets `computed_fields` emit `"engines": "{{ record.engines | join:, }}"`, producing legacy's exact comma-joined string; schema tightened to `["string","null"]`. `parity_searxng_test.go` asserts RAW record equality (the prior normalization workaround was removed). | RESOLVED |
| 5 | searxng | `published_date` requires a `computed_fields` rename from the raw API's camelCase `publishedDate` — plain schema projection copies by exact key match only; without the rename the field would silently drop. Fixed via `computed_fields`, not a deviation once fixed, but recorded because it is a required, non-obvious authoring step for any camelCase-vs-snake_case API. | ACCEPTABLE (mitigated) |
| 6 | searxng | ~~Legacy's derived `stream` marker field (which stream a record came from) was not modeled: it isn't present on the raw API response and `computed_fields` had no static-literal namespace to synthesize it.~~ **RESOLVED**: `computed_fields` now supports static-literal values (a template with no `{{ }}` markers passes through verbatim) — `"stream": "search"`/`"reddit"` stamps the identical marker legacy emits. `parity_searxng_test.go` asserts RAW record equality (no more field-stripping workaround). | RESOLVED |
| 7 | searxng | Subreddit-narrowing (`site:reddit.com/r/<sub>`) is not modeled — the dialect's `stream.Query` templating has no absent-key-falsy tolerance (unlike `auth`'s `when`), so a subreddit-present-vs-absent branch risks an unresolved-key hard error when `subreddit` is unset (the common case). Only the base case (`site:reddit.com <query>`, legacy's own no-subreddit fallback) is implemented; out of scope, not silently wrong. `subreddit` is no longer even declared in `spec.json` (F6: a declared-but-unwireable key is worse than an absent one). | ACCEPTABLE (documented scope narrowing) |
| 8 | searxng | ~~Optional Bearer-proxy `api_key` secret was not wired into a conditional `auth` rule — this golden predated the `when`-on-an-absent-secret absent-key-falsy fix and omitted the `auth` block entirely.~~ **RESOLVED**: `streams.json` `base.auth` now declares `[{"mode":"bearer","token":"{{ secrets.api_key }}","when":"{{ secrets.api_key }}"},{"mode":"none"}]` (F6, REVIEW.md), parity-tested both ways (`TestParitySearxng_ApiKeySecretSendsBearerAuth`/`ApiKeyAbsentSendsNoAuth`). | RESOLVED |
| 9 | postgres | Config-validation error **classification** parity, not exact string-match parity: both connectors are asserted to reject the same input for the same *reason* (bucketed: `missing_host`/`missing_database`/`missing_username`/`missing_password`/`invalid_sslmode`/`invalid_port`/`invalid_host`), but native/postgres keeps its own descriptive (still secret-free) error text rather than byte-copying legacy's strings. | ACCEPTABLE |
| 10 | postgres | Parity is proven against **fixture mode**, not a live pgx connection: `mode=fixture` short-circuits all network access on both legacy and native sides. This is not a coverage reduction — legacy itself has zero live-DB tests; fixture mode exercises 100% of the branches either side's test suite ever covered. `api_surface.json` for a DB connector declares `endpoints: []` (no REST surface to enumerate) rather than the REST-shaped surface pattern — this is the correct Tier-3 minimal-db surface, not a shortcut. | ACCEPTABLE |

Several engine gaps discovered by the goldens and two review passes were **closed**, not left as
deviations (see `.planning/phases/wave0-engine-harness/traces/waveF-repair-ledger.md`,
`gaploop-r1-ledger.md`, `gaploop-r2-ledger.md`): `PaginationSpec.MaxPages` enforcement; `EvalWhen`
absent-key-falsy semantics (both `waveF-repair-ledger.md`; item 8 above predates the latter);
`formatParam`/`parseLowerBoundTime` digits-only (Unix-seconds) passthrough for
`unix_seconds`/`date`/`github_date_range` (B1 — the honest shape `internal/app`'s persisted cursor
actually takes); stream/check path interpolation (F1); filter chains, `join:<sep>`, and
static-literal `computed_fields` (F9/F7 enablement); `link_header`'s SSRF host/scheme guard and
fail-closed-on-unparseable-URL (M1/F2/m2); `lastRecordCursor`'s configurable records path + numeric
last-record ids (F3); header resolution's hard-error-on-absent-secret (F4); `Bundle.RawSpec`
verbatim `Definition().Spec` (F5); `AuthHook`'s real caller context (F8); numeric-cursor support in
`cursor_advances` (B2, motivating item 2's RESOLVED above); certify's `secret_redaction` full
stdout/stderr/report-JSON scan and JSON-escaped-secret detection (M2/m4); `connectorgen validate`'s
`engine.ResolveCheckAuthSpec` wiring. When you hit any of these shapes again, rely on the
fixed/wired behavior directly — do not re-document as an open gap or re-introduce a workaround.

## 6. Escape-hatch decision tree

1. **Can it be expressed in `streams.json`/`writes.json`/`spec.json`/schemas alone (Tier 1)?** —
   default assumption; almost everything is (target ≥90%).
2. **Does it need Go, but only at one of the 5 named hook points (§1's Tier-2 table)?** — write a
   `hooks/<name>/hooks.go` (≤300 lines, ≤2 hook interfaces per `connectorgen validate`'s LOC/
   interface-count report). If it needs more than that, or a 3rd hook interface, escalate to
   Tier 3 rather than stretching a hook package.
3. **Is the protocol not HTTP/REST at all, or does it need direct control over connection
   lifecycle (SQL, message queues, filesystems, CDC)?** — Tier 3 native package, component split
   per §1.
4. **None of the above cleanly fits** — file a typed blocker, do not fake it. Taxonomy (exactly
   these values; matches `result.schema.json`'s `blockers[].type` enum):
   - `AUTH_COMPLEX` — an auth scheme the dialect/hooks genuinely cannot express (e.g. a signing
     scheme requiring canonicalized-request state hooks don't have access to).
   - `NON_REST` — a protocol Tier 3 hasn't been scoped for yet.
   - `DOCS_UNREACHABLE` — the connector's `documentation_url` is dead/paywalled/unreadable and no
     alternative source of truth (legacy code, OpenAPI spec) exists.
   - `SCHEMA_AMBIGUOUS` — the record shape cannot be determined confidently from any available
     source (docs, legacy code, live response).
   - `NEEDS_NEW_DEP` — requires a Go module not in `go.mod` (human gate; never add it yourself).
   - `ENGINE_GAP` — the engine dialect is missing a feature needed for **correct** (not
     merely convenient) behavior; document exactly what's missing and why a Tier-1/2 workaround
     would silently diverge from legacy. `ENGINE_GAP`s recur ≥3 times → the orchestrator extends
     the engine in a mini wave-0 increment, not per-connector patches.
   - `OTHER` — anything not covered above; always include `evidence`.

Quarantine (blocked connectors keep their current legacy implementation, tracked in
`docs/migration/quarantine.json`) is the outcome of an unresolved blocker after one repair attempt
— never silently ship an approximation you haven't documented as a deviation (§5) or a blocker.

## 7. Self-verify command block

Run all of these before reporting a connector "migrated" or "partial":

```
go run ./cmd/connectorgen validate internal/connectors/defs/<name>
go build ./internal/connectors/... && go vet ./internal/connectors/...
go test ./internal/connectors/conformance -run 'TestConformance/<name>'
```

When a parity suite exists for your connector (goldens only in wave0; may extend later):
`go test ./internal/connectors/engine -run TestParity<Name> -v`. Whole-repo hygiene (run whenever
touching anything beyond your exclusive dirs is even a remote possibility):
`go build ./... && go vet ./...` and `make lint`.

**FORBIDDEN files (never touch, regardless of connector)**: `internal/connectors/registryset/
registry_gen.go` (regenerated by `cmd/registrygen`/`cmd/connectorgen gen` only, orchestrator-run),
`catalog_data.json`, `icon_data.json`, any top-level `internal/connectors/*.go`, `go.mod`/`go.sum`
(a new dependency is a `NEEDS_NEW_DEP` blocker, never a self-serve add), any other connector's
`defs/`/`native/`/`hooks/` directory, legacy connector packages under
`internal/connectors/<name>/` (read-only reference until the wave6 registry flip — do not edit,
do not delete).

**No-commit rule**: migration agents do not run `git commit`. The orchestrator commits once per
wave-close after the path-guard (`git status --porcelain` limited to assigned dirs) passes.
