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
`capabilities.write: false` and no `writes.json` file at all. `spec.json` has only `base_url`
required; `api_key` is declared `x-secret` for documentation but **not** wired into an `auth`
block (see §3's when-grammar note — an optional secret cannot safely gate an `auth` candidate in
this dialect without risking a hard error on the common no-credential path; omit the whole `auth`
list instead, which resolves to "no auth", exactly matching legacy's own default). Pagination is
`page_number` with `size_param` intentionally empty (`""`) — no page-size query param is ever
sent; `page_size` in the bundle exists purely as the client-side short-page stop threshold.
`streams.json`'s two streams differ only in query scoping (`q` vs `site:reddit.com {{
config.query }}`) and both use `computed_fields` to rename the raw API's `publishedDate` to the
schema's `published_date` (a schema-projection rename — see §2 and §3).

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
  (`connectorgen validate`'s `pk_fields_exist`/`cursor_fields_exist` checks, and
  `conformance`'s static checks of the same names, enforce this).
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

**Template references + filters** (`interpolate.go`). `{{ <ref> | <filter> }}`, one optional
filter. References: `config.<key>`, `secrets.<key>`, `record.<dotted.path>` (walks nested
`map[string]any`), bare `cursor`. Filters: `urlencode` (percent-encodes for path-segment
insertion; applied **by default** to every `InterpolatePath` resolution unless an explicit filter
overrides it — general `Interpolate`/header interpolation do NOT default-apply it),
`unix_seconds` (RFC3339 string → Unix seconds integer string), `base64` (`base64.StdEncoding`).
Any resolved value destined for a header (`InterpolateHeader`) or that itself contains `\r`/`\n`
is rejected outright as a header/injection guard (THREAT-MODEL §2) — this check runs on the
**pre-filter** value for every interpolation call, not just headers. An absent `config.*`/
`secrets.*` key is a **hard error** naming the key and namespace in ordinary `Interpolate`/
`InterpolatePath`/`InterpolateHeader` calls — do not rely on a missing key silently becoming `""`
outside `when` conditions (next paragraph).

**`when` grammar** (`interpolate.go`'s `EvalWhen`, used only by `auth.go`'s `selectAuth`/
`authSpecMatches` today): `config.k == 'literal'` (equality against a single-quoted string
literal), `config.k in ['a','b']` (membership), or a bare `config.k`/`secrets.k` (truthiness — any
non-empty resolved string is true). **ABSENT-KEY-FALSY semantics apply only in `when`, never in
general interpolation**: a `config.*`/`secrets.*` reference whose key is entirely absent at
runtime resolves to `""` (empty string) here — truthiness is false, `==` compares against `""`,
`in [...]` treats it as not-contained unless the list itself contains the empty-string literal.
This is what makes an *optional* credential/config-gated `auth` candidate possible without a
separate "is this key present" primitive. Static validation (`ResolveCheck`, run by
`connectorgen validate`/conformance's `interpolations_resolve`) is unaffected by this
runtime-absence tolerance: a `when` template referencing a key **not even declared** in
`spec.json`'s properties is still a hard validate-time error.

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
neither, is a load-time error. `MaxPages` is a hard request-count cap enforced in `read.go`'s
`readDeclarative` loop, independent of page fullness, checked *before* issuing the request for
that page number; `MaxPages <= 0` (absent/zero) is unbounded. Stream-level `pagination` replaces
the base-level spec **wholesale** (no field-by-field merge) when present.

**`param_format`** (`incremental.param_format`, applied by `read.go`'s `formatParam` to the
RFC3339 lower-bound value before sending it as `request_param`): `rfc3339` (default; send
verbatim), `unix_seconds`, `date` (`2006-01-02`), `github_date_range` (`>=<value>`, a
lower-bound-only GitHub search qualifier).

**`computed_fields`** (`streams.json` per-stream): a map of output-field-name → template resolved
against the **raw** (pre-projection) record, so it can rename a field (`published_date` from
`{{ record.publishedDate }}`) or reach into nested JSON (`{{ record.user.login }}`). **No
static-literal injection exists** — a computed field can only derive from `record.*`; there is no
namespace for a per-call constant (e.g. a "which stream did this come from" marker cannot be
synthesized this way — see the searxng `stream`-field omission in §5). A computed field whose
source path is absent on a given record is silently skipped for that record (not an error) —
common for optional/differently-shaped nested fields across record variants (issues vs PRs).

**Conditional headers**: a header template that resolves to an unresolved key (the referenced
config/secret is simply absent — e.g. an optional per-account header with no value configured) is
**omitted entirely**, not sent empty. Any other interpolation failure (CRLF, unknown
namespace/filter) still propagates as a hard error.

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
- **RFC3339 string cursors in conformance fixtures** — a documented, deliberate convention, not a
  bug: even when the real API's wire format for a cursor field is numeric (Stripe's `created` is a
  Unix-seconds integer), **this bundle's own `fixtures/streams/**` files** represent that field as
  an RFC3339 string. Why: `conformance/dynamic.go`'s `checkCursorAdvances` recognizes a cursor
  value only via a Go `string` type assertion and then parses it as RFC3339 to compute
  `unix_seconds`-formatted request params; a numeric fixture value makes that check hard-fail
  ("no cursor value observed"), not skip gracefully. This does **not** affect engine-vs-legacy
  parity: the schema type is declared permissively (`["integer","string"]`), and the read path
  performs no type coercion — it copies the raw value verbatim regardless of declared type. Real
  parity tests (e.g. `parity_stripe_test.go`) drive their own `httptest` payloads with the field in
  its real (numeric) wire shape; only this bundle's *own* conformance/`connectorgen validate`
  fixtures use the RFC3339-string convention. See stripe's `docs.md` "Known limits" for the
  in-bundle explanation to copy.
- `fixtures/check.json`: `{"request": {...}, "response": {"status": 200, "body": {...}}}` — used
  by `check_fixture`.
- `fixtures/writes/<action>.json`: `{"record": {...}, "expect": {"method", "path", "body"?}}` —
  used by `write_validate`/`write_request_shape`; the engine's dry-run/actual request must match
  `expect` exactly for a valid `record`, and a deliberately invalid record (missing a required
  field) must fail validation in its own dedicated fixture/test case.
- **No secret-looking values, ever.** `connectorgen validate`'s `secret_literal` rule and
  `conformance`'s `secret_redaction` check both regex-scan every fixture file (and `docs.md`) for
  Bearer-header shapes, `api_key`/`access_token`/`secret`/`password`-adjacent long tokens, and
  vendor-recognizable prefixes (e.g. `sk_live_`/`sk_test_`) — a match is a hard validate failure,
  not a warning. Keep fixture "secrets" obviously synthetic (`sk_test_FAKE...` still trips the
  scanner deliberately — use `fixture_token_placeholder`-style strings instead, or better, don't
  put secret-shaped strings in fixtures at all since fixtures never carry auth).

## 5. Parity-deviation ledger

**Meta-rule**: a deviation from legacy behavior is **ACCEPTABLE** iff it never changes the emitted
record DATA for any input legacy itself would accept, AND it is documented both in the migration
agent's `result.schema.json` `parity_deviations[]` entry and in this ledger. Anything that *would*
change accepted-input behavior, or that the engine genuinely cannot express at all, is an
`ENGINE_GAP` blocker (§6) — never a silent workaround.

| id | connector | description | verdict |
|---|---|---|---|
| 1 | stripe | `create_customer`'s legacy "email OR name required" rule (a named-field OR) is approximated by `minProperties: 1` over the four optional fields (`email`,`name`,`description`,`phone`) — the engine's draft-07 subset has no `anyOf`/`oneOf`. Strictly more permissive (a record with only `phone` set passes here, would fail legacy); never stricter; parity tests only exercise legacy-valid records. | ACCEPTABLE |
| 2 | stripe | Bundle's own `fixtures/streams/**` represent the `created` cursor as an RFC3339 string rather than Stripe's real Unix-seconds wire integer — a fixture-authoring accommodation for `conformance`'s `cursor_advances` string type-assertion (see §4). Schema type is permissive (`["integer","string"]`); read path performs no coercion; real parity tests use the numeric wire shape. | ACCEPTABLE |
| 3 | stripe | `limit=100` is sent via a static per-stream `query: {"limit": "100"}` rather than `PaginationSpec.LimitParam`/`PageSize`, because the `cursor`+`last_record_field` paginator constructor does not read those fields (only `page_number`/`offset_limit` do). `limit_param`/`page_size` remain declared on the spec anyway as an honest statement of intended page size for a future engine enhancement. | ACCEPTABLE |
| 4 | searxng | Raw API's `engines[]` array is passed through unjoined (schema types it `["array","string","null"]`) rather than legacy's comma-joined string, because the dialect has no array-join filter. Parity tests normalize both representations to a canonical form before comparing — the underlying DATA (which engines contributed) is unchanged. | ACCEPTABLE |
| 5 | searxng | `published_date` requires a `computed_fields` rename from the raw API's camelCase `publishedDate` — plain schema projection copies by exact key match only; without the rename the field would silently drop. Fixed via `computed_fields`, not a deviation once fixed, but recorded because it is a required, non-obvious authoring step for any camelCase-vs-snake_case API. | ACCEPTABLE (mitigated) |
| 6 | searxng | Legacy's derived `stream` marker field (which stream a record came from) is not modeled: it isn't present on the raw API response and `computed_fields` has no static-literal/stream-name namespace to synthesize it. Not part of the PK (`url`)/cursor (`published_date`) contract, so no dedup/incremental impact. | ACCEPTABLE |
| 7 | searxng | Subreddit-narrowing (`site:reddit.com/r/<sub>`) is not modeled — the dialect's `stream.Query` templating has no conditional/default-value filter, so a subreddit-present-vs-absent branch risks an unresolved-key hard error when `subreddit` is unset (the common case). Only the base case (`site:reddit.com <query>`, legacy's own no-subreddit fallback) is implemented; out of scope, not silently wrong. | ACCEPTABLE (documented scope narrowing) |
| 8 | searxng | Optional Bearer-proxy `api_key` secret is not wired into a conditional `auth` rule — `when`-on-an-absent-secret's absent-key-falsy fix (see below) now makes this *possible*, but this golden predates the fix and still omits the `auth` block entirely (absent/empty `auth` = no authentication, matching legacy's real default). Future connectors with an optional secret-gated auth spec should use the now-fixed `when` tolerance instead of omitting `auth`. | ACCEPTABLE (superseded pattern; prefer `when` going forward) |
| 9 | postgres | Config-validation error **classification** parity, not exact string-match parity: both connectors are asserted to reject the same input for the same *reason* (bucketed: `missing_host`/`missing_database`/`missing_username`/`missing_password`/`invalid_sslmode`/`invalid_port`/`invalid_host`), but native/postgres keeps its own descriptive (still secret-free) error text rather than byte-copying legacy's strings. | ACCEPTABLE |
| 10 | postgres | Parity is proven against **fixture mode**, not a live pgx connection: `mode=fixture` short-circuits all network access on both legacy and native sides. This is not a coverage reduction — legacy itself has zero live-DB tests; fixture mode exercises 100% of the branches either side's test suite ever covered. `api_surface.json` for a DB connector declares `endpoints: []` (no REST surface to enumerate) rather than the REST-shaped surface pattern — this is the correct Tier-3 minimal-db surface, not a shortcut. | ACCEPTABLE |

Two engine gaps discovered by the searxng golden were **closed** by a follow-up repair (see
`.planning/phases/wave0-engine-harness/traces/waveF-repair-ledger.md`), not left as deviations:
`PaginationSpec.MaxPages` is now wired into `read.go`'s loop (previously silently ignored — an
always-full-page source would have paginated unbounded), and `EvalWhen` now evaluates an absent
`config.*`/`secrets.*` key as falsy instead of hard-erroring (previously made an
optional-secret-gated `auth` rule impossible; item 8 above predates this fix). When you hit either
shape again, rely on the fixed behavior directly — do not re-document them as open gaps.

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
