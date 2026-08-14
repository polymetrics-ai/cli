# Migration conventions — bundle authoring reference

> **Current delivery procedure:** [Connector Delivery Canon](../connector-canon/INDEX.md) and its
> [Implementation Procedure](../connector-canon/IMPLEMENTATION-PROCEDURE.md) decide whether a
> connector can be called complete. This document is the detailed bundle-authoring reference; it
> cannot grant a capability, bypass the Foundation Check, or replace warehouse-mediated flow proof.

Read this whole file before touching a connector. It is the bundle-authoring ground truth every
migration agent and reviewer follows; deviations are defects, not judgment calls. Goldens (read, don't copy
blindly — port the *pattern*): `internal/connectors/defs/stripe/**` (declarative HTTP + writes),
`internal/connectors/defs/searxng/**` (read-only, no-auth), `internal/connectors/defs/postgres/**`
+ `internal/connectors/native/postgres/**` (Tier-3 split). Engine source of truth:
`internal/connectors/engine/{bundle,interpolate,paginate,read,write,hooks,rate_limits,schema}.go`.

## 1. Target layouts

### Tier 1 — declarative bundle (target ≥90% of connectors)

One directory `internal/connectors/defs/<name>/`, zero Go:

```
metadata.json        # identity + capabilities + risk (engine/bundle.go Metadata)
changefeed.json      # optional evidence-backed changefeed declaration
spec.json            # draft-07 connection spec; x-secret marks secret fields
streams.json         # base HTTP config + streams[] (required unless dynamic_schema)
writes.json          # actions[] (omit entirely when capabilities.write is false)
api_surface.json     # coverage manifest (always required)
cli_surface.json     # optional provider-style CLI/help metadata
certification.json   # optional certify metadata: defaults, safe candidates, pairings
rate_limits.json     # optional provider-cited HTTP pacing policy (see §3)
schemas/<stream>.json  # one draft-07 schema per stream, x-primary-key/optional x-cursor-field
fixtures/
  check.json
  streams/<stream>/page_1.json (page_2.json ... when paginated)
  writes/<action>.json
docs.md              # Overview / Auth setup / Streams notes / Write actions & risks / Known limits
```

Worked example — **stripe** (`internal/connectors/defs/stripe/`): `metadata.json` declares
`capabilities.write: true`; `spec.json` has one `x-secret` field (`client_secret`) plus
`base_url`/`account_id`/`start_date`/`page_size`/`max_pages`/`mode`; `streams.json`'s `base` sets
`url` from `{{ config.base_url }}` with a defaulted Stripe API URL, a conditional `Stripe-Account`
header (omitted when `account_id` is unset — see §3), bearer `auth`, a `cursor` pagination block
(`last_record_field: id`, `stop_path: has_more` —
Stripe's `starting_after`/`has_more` convention), a `check` request, and an `error_map`; each of the
5 streams shares the identical shape (`GET`, `records.path: "data"`,
`incremental.cursor_field: created`, `param_format: unix_seconds`) — copy this shape for any
list-endpoint API with a uniform envelope. `writes.json` declares `create_customer` and
`update_customer` as approval-gated form writes plus `delete_customer` as a destructive no-body
write with the legacy `confirm: "destructive"` declaration (normalized by the shared typed gate)
and idempotent 404 handling; both mutating customer-by-id actions carry `path_fields: ["id"]`.
`fixtures/streams/customers/{page_1,page_2}.json` is the
**required 2-page fixture** for a paginated stream (§4). `docs.md` documents the `minProperties: 1`
parity deviation (§5, item 1) inline as well as in this ledger.

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
- `cdc.go` — `ReadCDC()` (a planned CDC boundary must fail closed until its execution contract is
  complete — see postgres below).

Still ship a bundle (`internal/connectors/defs/<name>/{metadata.json,changefeed.json?,spec.json,
api_surface.json,docs.md}`) so identity/spec/docs stay uniform with every other connector. A
native database bundle may additionally ship `database.json` as described below;
`metadata.json` sets
`capabilities.dynamic_schema: true` and the bundle ships **no `streams.json`** (the loader
(`bundle.go`'s `loadStreams`) only tolerates a missing `streams.json` when `dynamic_schema` is
true). The package embeds `engine.Base` (via `engine.NewBase(bundle)`) purely to serve
bundle-derived `Name()`/`Metadata()` and the base `Definition()`; `engine.Base` does **not**
provide Check/Catalog/Read/Write — those remain hand-written Go. A native transport declaration
extends that base rather than replacing it.

Worked example — **postgres** (`internal/connectors/native/postgres/` +
`internal/connectors/defs/postgres/`): `api_surface.json` declares `endpoints: []` with a `scope`
prose explaining there is no REST surface to enumerate (schema-valid: no `minItems` on
`endpoints`); `spec.json` has `password` marked `x-secret` and a `mode: fixture` config value that
short-circuits all network access for credential-free testing (test/conformance-harness affordance
only, never set in production); `cdc.go` retains a `pglogrepl`-backed logical-replication
foundation, but `ReadCDC` fails closed before source contact while its changefeed is planned. Its
admission conditions are owned by the [PostgreSQL bundle docs](../../internal/connectors/defs/postgres/docs.md),
so the capability stays false rather than faking a runnable CDC surface. Its `database.json` is the
reference strict policy-only declaration described below; it does not register a driver or promote
a capability. The package has **no `init()`/`RegisterFactory`/
`RegisterNativeLive` call** in wave0 (the factory-registration flip is wave6); that is distinct
from its explicit, definition-selected source-transport helper. A grep-guard test enforces the
factory boundary.

#### Database-native policy declaration (`database.json`)

Use this optional bundle file only for a native database connector; it remains absent for the
broad existing fleet and non-database natives. When present, normal `engine.Load` and
`connectorgen` loading run it through `database.Load`. The versioned, closed schema at
`internal/connectors/database/schema/database.schema.json` is the authoritative field contract;
do not copy it into another validator or document variant.

The declaration is limited to a driver identity/protocol/API version, structured catalog and
identifier policies, finite resource limits, explicit native-to-logical type mappings, and the
existing closed `synccontract.Mode` vocabulary. It is policy, not connection configuration: never
put a credential, DSN, raw SQL/query, DDL, session/transaction/receipt/checkpoint state, CDC or
polling declaration, generic operation, or capability claim in it. `metadata.json` remains the
only public capability owner. A syntactically valid declaration neither registers a driver nor
admits execution; its exact driver identity and matching native evidence are required separately.

Managed-target owners, refs, control records, and provisioning plans are runtime values, not
`database.json` fields. A declaration must never encode an author-supplied target name or control
record; their ownership and execution contract is owned by the
[warehouse-mediation architecture](../architecture/connector-architecture-v2-design.md#b71-database-warehouse-mediation).

The database layer has no direct connector-pair or zero-copy command. Its source-to-warehouse and
warehouse-to-target admissions are separate legs; see the
[warehouse-mediation architecture](../architecture/connector-architecture-v2-design.md#b71-database-warehouse-mediation)
for the shared boundary.

### Tier 2 — hooks (bundle + Go escape hatch, target ~8%)

The bundle still declares every stream/write/schema; a single `internal/connectors/hooks/<name>/
hooks.go` (+ `hooks_test.go`) attaches Go at named extension points only. Line-cap wording (gap-loop
cycle-1 REVIEW-A.md §C1 — the prior "~300, hard-capped" phrasing was self-contradictory, a tilde is
not a hard cap): **~300 lines is a soft target; exceeding it requires a self-reported justification
in the connector's trace/deviation ledger (name which mandated interfaces/shapes account for the
size, e.g. github's RS256 JWT exchange + 4 compound writes, monday's 2 pagination shapes + 5 record
mappers + GraphQL-errors-in-HTTP-200 envelope); 400 lines is a hard ceiling; exceeding 400, OR
needing a 3rd hook interface regardless of line count, escalates to Tier 3.** Five interfaces
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
  `x-secret` fields end up in `Schema.SecretKeys()`, which governs secret/config partitioning.
  `DryRunWrite` preview warnings intentionally preserve their resolved request line rather than
  substitute secret values. Write actions that place sensitive non-secret identifiers or clinical
  values into templated paths must also declare `redact_fields` for those record paths so
  reverse-plan source-table samples and write errors remain redacted; those declarations likewise
  do not substitute values in the resolved dry-run request line. A field that merely *looks*
  sensitive but is documentation-only (an optional Bearer-proxy key never
  wired into `auth`, e.g. searxng's `api_key`) is still marked `x-secret: true` — the marker is
  about the field's nature, not whether this bundle currently exercises it.
- **JSON Schema `format` enforcement is intentionally narrow**: the engine validates string
  instances declared with `format: "uri"` as absolute URIs and rejects whitespace, control
  characters, backslashes, and malformed escaping. Other schema formats remain annotations unless
  a surface-specific validator enforces them. Pair `format: "uri"` with `pattern` when the contract
  is narrower than an absolute URI; Google Calendar watch actions use that combination to require
  HTTPS callback URLs.
- **Schema-as-projection**: a stream's `schemas/<stream>.json` `properties` set is derived
  **field-for-field** from what the legacy connector's own `mapRecord`/record-shaping function
  actually emits — not from guessing the raw API shape. In `"schema"` projection mode (the
  default; see §3), only declared properties survive; anything legacy emitted that this bundle's
  schema omits is silently dropped from parity. When a legacy field name differs from the raw API
  field (e.g. searxng's `published_date` vs the raw `publishedDate`), add a `computed_fields`
  rename (§3) — don't just omit it. Every declared `x-primary-key`, `x-cursor-field`, and
  `incremental.cursor_field` must name a property in that stream's schema. `x-cursor-field` is
  optional: an executable `incremental.cursor_field` supplies the effective cursor when the schema
  omits it; when both cursor declarations are present, they must agree. `connectorgen validate`
  and conformance enforce the primary-key and incremental-cursor field checks, and the generator
  also rejects a mismatch between the two cursor declarations.
- **Sync-mode derivation — never declared**: `full_refresh_append` and
  `full_refresh_overwrite` always apply. `full_refresh_overwrite_deduped` requires both
  `x-primary-key` and an effective cursor (`x-cursor-field`, or an executable
  `incremental.cursor_field`); `incremental_append` requires an executable `incremental` block and
  that effective cursor; `incremental_append_deduped` requires all three. The public compatibility
  names and their closed contracts are owned by `internal/synccontract/public_modes.go`; do not add
  a `supported_sync_modes` field anywhere — the engine derives this projection from schema/stream
  shape at runtime.
- **`api_surface.json` depth — minimal-honest for wave0/pilot** (DECISIONS.md #4): list every
  implemented stream/write under `covered_by`; everything else is documented as blocked/planned or
  excluded operation-ledger metadata until typed schemas, bounds, fixtures, and safety evidence are
  authored (see stripe's `api_surface.json` for the current full-ledger pattern — 5 covered streams,
  3 covered customer writes, and the remaining official Stripe operations tracked exactly once as
  blocked/planned or excluded rows). Full API-surface research (every documented endpoint actually
  implemented) is Pass B (wave5), not wave0/pilot/Pass-A fan-out. The
  closed exclusion-category vocabulary (design §E.1 rule 3, enforced by the loader's meta-schema
  enum): `destructive_admin`, `requires_elevated_scope`, `binary_payload`, `deprecated`,
  `non_data_endpoint`, `duplicate_of`, `out_of_scope`.
- **`changefeed.json` is evidence-backed and fail-closed**: add this optional file only after
  reviewing a provider artifact. It records a closed status (`implemented`, `planned`,
  `unsupported`, or `unknown`), a closed mechanism (`logical_replication`, `binlog_replication`,
  `incremental_cursor`, `webhook`, `event_stream`, or `polling_watermark` — the `mechanism` enum
  of the schema named below), and a source artifact URL, version, and
  retrieval date. An `implemented` declaration also requires a named executor, checkpoint and
  recovery contract, delivery guarantees, and covered streams, but it does not set public `cdc`
  by itself: the registered connector must expose a matching `ChangefeedExecutor`. An
  `unsupported` declaration requires a reason and must not claim an executor, checkpoint, or
  delivery contract. An absent descriptor is unknown and non-capable. The structural schema is
  `internal/connectors/engine/schema/changefeed.schema.json`; runtime semantic checks live on
  `connectors.ChangefeedDescriptor`.
- **`polling_watermark` is bounded replay, not a hard-delete feed**: an implemented polling
  declaration uses the fixed executor `{"kind":"engine","id":"polling_watermark"}` and must
  supply `polling_watermark` beside the normal evidence/checkpoint/delivery fields. Its
  `watermark` declares `{kind, path}` where kind is exactly `timestamp`,
  `monotonic_sequence`, or `opaque_cursor`; `tie_breaker.path` names the stable secondary order
  value. Each timestamp or monotonic-sequence source page must be ordered ascending by that tuple;
  an unordered page is refused before delivery or checkpoint advancement. `checkpoint.keys` must
  list those paths in that order. The only supported `boundary` is
  `inclusive`: the next read starts at `>=` the committed boundary and therefore deliberately
  replays the edge record. Set `delivery.duplicates` to `at_least_once`; `>` would silently lose
  tied timestamps and is not an allowed substitute. Declare positive `safety_lag_seconds` for a
  timestamp source whose provider can publish late/clock-skewed records; the executor rereads that
  overlap on the next run. An unfinished bounded overlap persists an internal scan tuple separately
  from the durable high-water tuple, retaining the provider's exact tie-breaker text and numeric
  ordering metadata when applicable; it resumes after restart and clears once the overlap is
  exhausted so later polls begin at the lag boundary again. `0` is an explicit no-lag opt-out and
  can lose such records. Non-time watermark kinds must declare `0` because a timestamp lag has no
  honest meaning for them.
  `page_size`, `max_pages`, and `request_budget` are all positive required bounds. Each physical
  declared-source request consumes the one shared request budget: the source adapter must consume
  a budget token before its primary read and before every additional fixed read, including a
  declared deletion endpoint. A deletion endpoint requires `request_budget >= 2`, and the executor
  does not start a primary read unless both its primary and deletion tokens remain. Each primary or
  deletion physical page must contain no more than `page_size` records; an overflow is a typed,
  non-advancing refusal. Timestamp and monotonic-sequence sources merge primary and deletion
  records through a safe shared tuple boundary while preserving independent durable frontiers, so
  one source cannot skip the other source's page. A declared deletion endpoint also persists its
  own query/start frontier from the first checkpoint, before any tombstone is accepted. Once a
  deletion source has a durable timestamp tuple, restart derives that source's cursor and safety
  overlap from the tuple and scan, never from primary progress. An uninitialized opaque deletion
  source starts without a primary cursor. Opaque cursor sources retain each source cursor
  verbatim and promise only provider-local ordering; they do not claim a cross-source tuple merge.
  Reaching `max_pages` or exhausting that budget returns a typed,
  resumable stop after only the last durably accepted position; it is never a successful silent
  truncation. The executor checks cancellation between source fetch, delivery, and checkpoint
  commit. It advances the tuple only after every emitted record in its page is durably accepted by
  the destination and the checkpoint committer succeeds; a failure in either step replays the
  page. The committer is a consumer-facing adapter for the durable database sync contract rather
  than a second checkpoint store.

  A polling scan cannot see a hard delete after the record disappears. With no delete source,
  declare `delivery.deletes: "not_available"`; inspect/catalog output carries that exact truth and
  must not describe the stream as delete-aware. To advertise `"tombstone"`, declare exactly one
  of `soft_delete: {"path":"..."}` (a truthy provider field becomes a delete event) or
  `deletion_endpoint: {"path":"/fixed/provider/path","records_path":"..."}` (the closed
  source adapter returns its records as delete events). A soft-delete field or deletion endpoint
  must itself carry the declared watermark and tie-breaker values so it participates in the same
  ordered, replay-safe checkpoint. Do not use a generic endpoint, SQL, or caller-provided path.
- **`api_surface.json` v2 provenance**: follow the authoritative
  [provider-artifact contract](../architecture/connector-architecture-v2-design.md#version-2-provider-artifact-provenance).
  Upgrade only bundles in the dedicated provider-artifact sweep; do not bulk-edit unrelated bundles
  or infer v2 evidence from a bare URL.
- **`docs.md` required headings** (exact text, `#`/`##` either level; `conformance`'s
  `docs_present` and `connectorgen validate`'s `docs_heading` rule both check presence by trimmed
  text only): `Overview`, `Auth setup`, `Streams notes`, `Write actions & risks`, `Known limits`.
  `Known limits` is not decorative — every deliberate simplification, `ENGINE_GAP`, or
  scope-narrowing goes there (see the parity-deviation ledger, §5) with enough detail that a
  reviewer or a future capability-expansion agent doesn't have to re-derive the reasoning.
- **`cli_surface.json` validation stays definition-owned**: command-specific flags and constraints
  belong in the connector's CLI surface metadata, never in provider-named shared runner branches.
  Use `flags[].type:"number"` for finite decimal-valued inputs; it maps to JSON Schema `number`
  (not `integer`), and the runtime rejects NaN and infinite values.
  Use `flags[].format:"date-time"` for RFC3339/ISO-8601 timestamp flags,
  `flags[].allow_empty:false` to reject present blank string flags, `flags[].required:true` for
  command inputs that must be present before execution, `flags[].max_items`/`flags[].min_items`
  to bound a `string_array` flag's item count against the flag the user typed (before the
  assembled body is validated), and optional `flags[].minimum` on `integer` or `number` flags to
  enforce a provider-cited numeric lower bound. A declared minimum rejects a supplied value below
  the bound; for an ETL flag mapped to `config.*`, it also checks the effective configured value
  when the command flag is absent. Omit `minimum` rather than inferring a bound from a parameter
  name or schema default, so existing command behavior stays unchanged. Use
  `constraints[]` `kind:"order"` /
  `op:"lt"` / `value_type:"date-time"` over mapped `query.*` or `body.*` targets for provider
  date-range rules. Optional `config.*` fallbacks preserve connection-level defaults when the
  command flag is absent. For implemented direct-read POST operations, every required
  `operations.json` `rest.body_schema` path must either be supplied by static `rest.body` or by a
  required command flag mapped to `body.*`.
  `internal/connectors/engine/schema/cli_surface.schema.json` is the schema source of truth, and
  `connectorgen validate` rejects unsupported formats, operators, fallback namespaces, unmapped
  constraint targets, missing required body mappings, and multi-line validation messages. A
  `direct_read` command's page-navigation flags are runtime-owned exactly as the binary-download
  destination flags are: `--page` and `--page-cursor` are declared once in
  `internal/connectors/direct_read_page_flags.json` and answered from the connector's own declared
  pagination spec, so no bundle declares them (see AGENTS.md, "Direct Reads Return One Page, And
  Say So"). The command fields that are derivable from `operations.json` — `api_surface`, flag
  `maps_to`, `output_policy`, and `rest.max_bytes` — are not hand-authored:
  `go run ./cmd/connectorgen surface-sync` fills them and `--check` fails on drift (see AGENTS.md,
  "Command Surface Must Stay Executable"). The same command generates the embedded
  `operation_endpoint_ledger.json` runtime projection from `api_surface.json` and
  `operations.json`; it contains only direct-read method, path, operation kind, and response cap,
  so preflight fails closed if a shipped bundle has no matching projection entry. The shared-code boundary guard
  (`docs/migration/connector-boundary-guard.md`) enforces this ownership rule outside connector
  defs/hooks/native escape hatches.
- **`api_surface` operation rows are reconciled, never hand-edited**: use
  `go run ./cmd/connectorgen surface-reconcile --check` to derive the current
  result for direct-read operation rows. The tool loads the disk bundle and
  calls the real `commandrunner.Preflight`; it writes `covered_by.direct_read`
  only for a matching command that passes that runtime path. A missing,
  planned, or preflight-failing command stays blocked with a deterministic
  reason; unknown models are refused unchanged. Use `--reason-contains` to
  audit a named stale-reason cohort and `--json` for a reviewable count report.
  Run without `--check` only after reviewing the proposed reclassification.
- **`output_policy` is intent-specific, closed, and must be chosen deliberately**: for a
  `direct_write` command whose caller needs the complete decoded JSON response, declare `"json"`;
  the engine returns that decoded body unchanged. Declare `"none"` only when the operation
  intentionally returns no response body to the caller. `json_redacted`, `write_result_redacted`,
  and `gong_bounded_input_redacted` remain direct-write compatibility choices for explicitly
  selected response contracts, not defaults for complete output. For `direct_read`, use
  `repository_contents_file_metadata` for a single repository-file metadata response and
  `repository_contents_directory` for repository directory listings; those bounded shapes reject
  sensitive repository paths before network access and redact `content` plus download URLs from
  returned JSON. `json_redacted` and `clinical_json_redacted` remain the direct-read compatibility
  choices for explicitly selected response contracts. `binary_file_bounded` is neither a
  direct-read nor direct-write policy: it remains schema-valid only for existing binary-download
  metadata, while new file/binary commands use the distinct intent below. Do not add
  provider-prefixed output-policy names in shared Go; new response families need a generic policy
  name and regression tests proving reuse by more than one connector shape.
- **File/binary commands are their own intent**: declare `intent:"binary_download"` with an
  `operation` of kind `binary_download` and exactly one connector-relative GET `api_surface`
  endpoint, and leave the command's `output_policy` unset — the response becomes a file on disk,
  not a JSON body, so no direct-read policy applies. Do not declare the destination flags in the
  bundle: `--dest-root` (required), `--file-name`, and `--max-bytes` are runtime-owned, read from
  `internal/connectors/binary_download_flags.json` by runtime help, the generated `MANUAL.md`/
  `SKILL.md`, and the website generator, so one declaration keeps all three documenting the same
  flag surface. `--max-bytes` may only lower the operation's declared `binary.max_bytes`, never
  raise it.
- **`certification.json` stays definition-owned and harness-only**: connector-specific certify
  contracts belong beside the connector bundle, never in provider-named shared certify branches.
  This optional file may declare `source.default_stream`, source credential defaults,
  live-unavailable classifiers, direct-read candidates, binary-download candidates, and safe write
  pairings. `source.default_stream` and every pairing's `verify_stream` must name declared streams;
  pairing `create`/`cleanup` actions must name declared `writes.json` actions. The write record
  schema remains the action's own `record_schema` in `writes.json`; certification metadata only
  chooses the safe lifecycle and overrides. Absent metadata means no candidates, no pairings, and no
  connector-specific cleanup behavior. `internal/connectors/engine/schema/certification.schema.json`
  is the schema source of truth, and `connectorgen validate` scans the raw file for secret-shaped
  literals.
- **`certification-matrix.json` is generated proof, not harness input**: only a connector named in
  `cmd/connectorgen/certificationallowlist.go` receives this sibling artifact. Generate one
  connector with `go run ./cmd/connectorgen certification-matrix --connector <name>`; it writes
  that shard only. It is not part of `defs.FS`, is never hand-authored, and contains no global
  baseline or count. Use `--all` only for deliberate regeneration and `--check` for the drift gate.

## 2.9 Command parameters and paging are DERIVED — never hand-author them

Two parts of a `direct_read` command's surface are generated from declarations
you already own. Hand-writing either one is a bug, not a shortcut.

### Paging

Never declare an opaque provider cursor (`cursor`, `start_cursor`,
`page_token`, or an equivalent) on a command. Navigation comes from the
connector's `streams.json` `base.pagination` spec, which the direct-read
executor consumes through the same seven strategies the ETL path uses. The
runtime supplies `--page` (for `page_number` and `offset_limit`, the only
strategies with an addressable page number) and `--page-cursor` (for `cursor`,
`next_url`, `link_header`). A raw cursor flag gives a caller a second,
unchecked route that bypasses the completeness contract and is removed by
`surface-sync`.

A declared page/window or addressable-position control such as Notion's
`page-size` or Bahmni's `start-index` is different: keep it only when the
runtime sends the caller's value and reports the window it actually used. The
executor honours that value over a declared default and measures completeness
against the effective size. Legacy opaque cursors such as Notion
`start-cursor` and Gong `cursor` are generated-surface drift, not precedent.

### Other parameters

Do not hand-author a flag for a parameter the provider specification already
documents. Import it instead:

```sh
go run ./cmd/connectorgen params-import <connector> --artifact <spec.json>
go run ./cmd/connectorgen surface-sync internal/connectors/defs
```

`params-import` writes the accepted parameter set into `operations.json` under
`rest.parameters` (name, location, type, requiredness, enum values, summary).
`surface-sync` then derives the command flags from it. The split exists so CI
stays hermetic: `surface-sync --check` verifies drift with no artifact and no
network.

The import deliberately drops two classes of parameter:

- **paging parameters**, for the reason above. The test is the parameter's
  MEANING, not the names your bundle happens to declare: a well-known paging
  name (`page`, `per_page`, `cursor`, `offset`, `limit`, `start_index`, …) is
  dropped, and so is any parameter whose own specification describes it as a
  cursor or as pagination. github's `after`/`before` cursors are dropped that
  way even though its spec declares `page`/`per_page`; the `before` on
  `/repos/{owner}/{repo}/notifications` is an ISO 8601 timestamp filter and is
  kept.
- **a path variable the connection supplies** — `{owner}`/`{repo}` in an
  operation's own `rest.path`, resolved from `spec.json` config through
  templating, so turning it into a flag would make every command demand a value
  the connection already knows. This is scoped to the path template on purpose:
  a config key nothing interpolates into the request (github's ETL-only `since`)
  IS imported, because no other mechanism supplies it.

### What you still author by hand

A flag the derivation cannot know about, and a better summary or narrower type
than the specification carries. `surface-sync` only ever **adds** derived flags;
a flag you already declared is left exactly as written.

### Verifying

```sh
go run ./cmd/connectorgen params-import <connector> --artifact <spec.json> --check
go run ./cmd/connectorgen surface-sync --check internal/connectors/defs
go run ./cmd/connectorgen validate internal/connectors/defs
```

`validate` rejects a malformed `rest.parameters` entry at build time — an `in`
outside `query|path` fails with the exact JSON pointer — so a declaration that
could not produce a valid command never reaches the runtime.

## 3. The engine dialect reference

All of this is read from `internal/connectors/engine/{bundle,interpolate,paginate,read,write,
schema}.go` directly — this section is a map, not a spec; when in doubt, read the source.

**Template references + filters** (`interpolate.go`). `{{ <ref> | <filter1> | <filter2> ... }}` —
**a filter chain of any length** applies every stage left-to-right (not just one filter); an
unknown filter name anywhere in the chain is a hard error (validate-time via `ResolveCheck`'s
filter-name check, wired into both `connectorgen validate` and static-checked chains, and
runtime). References: `config.<key>`, `secrets.<key>`, `record.<dotted.path>` (walks nested
`map[string]any`), bare `cursor`, `incremental.lower_bound` (S3 engine mini-wave item 1, below).
Filters: `urlencode` (percent-encodes for path-segment
insertion; applied **by default** to every `InterpolatePath` resolution unless an explicit filter
chain overrides it — general `Interpolate`/header interpolation do NOT default-apply it),
`unix_seconds` (RFC3339 string → Unix seconds integer string), `base64` (`base64.StdEncoding`),
`join:<sep>` (joins an array-valued reference — e.g. `record.tags` when it resolves to a raw
`[]any` — with `<sep>`; a non-array value is a hard error, not a silent stringify; the separator is
everything after the first `:`, so `join:, ` or a multi-character separator both work
unambiguously, but a literal `|` cannot be the separator since the outer chain-split takes
precedence), `last_path_segment` (gap-loop cycle-1 item 4, REVIEW-B.md finding 1/adjudication 1:
returns the final non-empty `/`-delimited segment of the resolved value — a trailing slash is
ignored, a value with no `/` at all passes through unchanged, never errors — the sanctioned way to
derive a HAL/URI-keyed API's trailing-segment id, e.g. `"id": "{{ record.uri | last_path_segment
}}"` for calendly's `idFromURI(uri)` legacy convention; use this instead of a RecordHook for any
URI-shaped derived-id field), `length` (counts a raw array value; non-array values count as `0`;
inside `computed_fields`, the exact single-reference form `{{ record.items | length }}` emits a
native integer, while ordinary string interpolation still returns the decimal text form),
`const:<value>` (S3 engine mini-wave item 1, below: discards the
resolved value entirely and always returns the literal text after the first `:` — its purpose is
composing with `omit_when_absent`/`default` to express "send this FIXED literal iff a reference
resolves" without depending on the reference's own value; the gating reference must still fail to
resolve for absence to propagate, `const` only replaces a SUCCESSFULLY resolved value). Any resolved
value destined for a header (`InterpolateHeader`) or that itself contains
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
(see searxng's `streams.json`). Static validation is unaffected by this runtime-absence tolerance: a
`when` template referencing a key **not even declared** in `spec.json`'s properties is still a hard
validate-time error. `connectorgen validate` validates EVERY templated `AuthSpec` field this way —
`token`/`username`/`password`/`value`/`token_url`/`client_id`/`client_secret`/`scopes` via
`ResolveCheck`, and `when` via **`ResolveCheckWhen`** (S3 engine mini-wave item 2, wave1-pilot
SUMMARY.md carried queue / REVIEW-A.md re-review R1/R3) — both wired through
`engine.ResolveCheckAuthSpec(spec, specKeys)`.

**`ResolveCheckWhen` — full when-grammar static validation** (`interpolate.go`, S3 engine mini-wave
item 2): before this increment, `ResolveCheck`'s bare `namespace.key`-only parsing treated an ENTIRE
`==`/`in`-shaped `when` expression as a single dotted reference — `{{ config.auth_type ==
'public' }}` hard-failed validate as an "unknown spec key `auth_type == 'public'`" finding even when
`auth_type` genuinely IS a declared spec property, because `resolve check: unknown spec key` reads
the whole clause after the first split-on-`.`, not just the intended left-hand-side reference. Every
`when` clause in every bundle in this repo was therefore restricted to bare truthiness only — not
because `EvalWhen` (the runtime evaluator) couldn't handle `==`/`in`, but because static validation
couldn't statically ACCEPT one. `ResolveCheckWhen(template, specKeys)` parses the IDENTICAL grammar
`EvalWhen` evaluates at runtime (equality, membership, truthiness, and the same rejected-operator set
— `!=`/`>=`/`<=`/`>`/`<`/`&&`/`||`) and validates only the left-hand-side reference against
`specKeys` (via the same `checkNamespaceRef` helper `ResolveCheck` itself uses) — the right-hand-side
literal/list syntax is checked for well-formedness (a missing quote or bracket is still a
validate-time error) but its literal VALUES are never checked against anything (there is no enum to
validate against; the bundle author chooses the literal set). Use `ResolveCheckWhen` for ANY `when`
field check; `ResolveCheck` remains correct for every other templated field (paths, headers, query,
computed_fields, non-`when` AuthSpec fields) which are plain `{{ }}` interpolation, never the
when-grammar. `internal/connectors/conformance/static.go`'s `checkInterpolationsResolve` routes
`AuthSpec.When` through `ResolveCheckWhen` the same way `cmd/connectorgen`'s `checkInterpolations`
does — keep both wired consistently if either changes.

**Dual-auth ordering is load-bearing — the golden pattern** (gap-loop cycle-1 item 7, lifting
zendesk-support's ledger item 3 per REVIEW-B.md's fan-out-readiness note): `base.auth` is evaluated
as a **first-match-wins candidate list** (`selectAuth`), so when legacy accepts more than one
credential shape with a defined precedence (e.g. an OAuth bearer token taking priority over
Basic-auth email+API-token when BOTH secrets happen to be configured), the candidate list's
DECLARATION ORDER must reproduce that exact precedence — reordering silently changes which
credential wins when both are present, an accepted-input-behavior change the §5 meta-rule forbids.
zendesk-support's golden shape:

```json
"auth": [
  { "mode": "bearer", "token": "{{ secrets.access_token }}", "when": "{{ secrets.access_token }}" },
  { "mode": "basic", "username": "{{ secrets.email }}/token", "password": "{{ secrets.api_token }}",
    "when": "{{ secrets.api_token }}" }
]
```

— bearer is declared FIRST (matches legacy's access-token-first precedence under its own
first-match-wins auth selection), Basic is the fallback candidate, and the both-secrets-present
corner case is explicitly parity-tested (not just each candidate in isolation) to prove the ordering
itself, not just that each mode individually works. Any dual/multi-candidate `auth` list must carry
an equivalent both-present parity test asserting which candidate actually wins.

**`oauth2_client_credentials` extra token-request form params — `auth[].extra_params`** (S4 engine
mini-wave item 4: auth0's M2M client-credentials exchange always includes an `audience` form
parameter alongside the standard `grant_type`/`client_id`/`client_secret`/`scope`; box's
`box_subject_type`/`box_subject_id` token-scoping params are the same shape):

```json
{
  "mode": "oauth2_client_credentials",
  "token_url": "{{ config.base_url }}/oauth/token",
  "client_id": "{{ config.client_id }}",
  "client_secret": "{{ secrets.client_secret }}",
  "extra_params": { "audience": "{{ config.base_url }}/api/v2/" }
}
```

Each `extra_params` value is an ordinary `{{ }}` template resolved via `Interpolate` against the
same Vars every other `AuthSpec` field uses (config/secrets) — an unresolved key HARD ERRORS exactly
like `client_id`/`client_secret` do; this is deliberately NOT given `stream.Query`'s
`omit_when_absent`/`default` tolerance, since a misconfigured audience/subject param should fail
loudly rather than silently omit a value a real OAuth2 provider may require. `connsdk.
OAuth2ClientCredentials` already exposed an `ExtraParams url.Values` field before this
increment — the gap was purely that `AuthSpec` had nothing to populate it from;
`buildOAuth2ClientCredentials` (engine/auth.go) now resolves `extra_params` and wires it through, no
connsdk change needed. `ResolveCheckAuthSpec` validates every `extra_params` value template
statically, exactly like `token_url`/`client_id`/`client_secret`/`scopes` — this flows into
`connectorgen validate` for free via the existing `engine.ResolveCheckAuthSpec(a, specKeys)` call in
`checkInterpolations`, no `cmd/connectorgen/validate.go` change was needed.

**`oauth2_refresh_token` — user-context tokens that expire** (`connsdk.OAuth2RefreshToken`,
`engine/auth.go`'s `buildOAuth2RefreshToken`). Use this, NOT `oauth2_client_credentials`, whenever
the provider's endpoints act on behalf of an end user. Client-credentials obtains an **app-only**
token: it authenticates the application, so it cannot reach a user's own resources however many
scopes it carries. A user-context token comes from the authorization-code flow and is renewed with
a refresh token, which is what this mode does. Reddit is the motivating case — its bearer tokens
expire one hour after issuance, so a caller-supplied `access_token` cannot survive a scheduled sync.

```json
{
  "mode": "oauth2_refresh_token",
  "token_url": "https://www.reddit.com/api/v1/access_token",
  "client_id": "{{ config.client_id }}",
  "client_secret": "{{ secrets.client_secret }}",
  "refresh_token": "{{ secrets.refresh_token }}",
  "refresh_token_store_key": "refresh_token",
  "scopes": "identity read"
}
```

`token_url`/`client_id`/`client_secret`/`scopes`/`extra_params` are the SAME fields
`oauth2_client_credentials` uses, with the same `Interpolate` semantics (an unresolved key hard
errors; never a silently unauthenticated request). Only the grant and two fields differ:

- **`refresh_token`** — templated, normally `{{ secrets.refresh_token }}`.
- **`refresh_token_store_key`** — OPTIONAL, and the one field that needs a decision. Many providers
  rotate the refresh token on every exchange and invalidate the previous one; dropping the new value
  means the connector works for one process lifetime and then fails with `invalid_grant` on the next
  run, long after the run that caused it. Declaring this key (normally the same key `refresh_token`
  reads from) persists the rotated value into the credential's encrypted vault entry, so the next
  run picks it up through the ordinary `secrets` path. **Omit it only if you have confirmed the
  provider does not rotate** — omitted means nothing is ever written, and the engine deliberately
  does NOT guess a key name from the `refresh_token` template (a resolved template yields a value,
  not a key; guessing would silently overwrite the wrong secret whenever it guessed wrong).

Runtime behaviour a bundle gets for free, none of which is declarable: the access token is exchanged
once and reused until shortly before expiry (safety margin clamped to half the lifetime, so a very
short-lived token still caches rather than re-exchanging per request); a missing/zero/unparseable
`expires_in` is treated as a conservative five-minute life, never as "never expires"; an otherwise
retry-eligible request gets at most ONE 401 refresh-and-retry (declarative writes follow the retry
policy below), so a revoked grant terminates instead of hammering the token endpoint; and the
exchange is serialised, so concurrent callers sharing an authenticator produce one exchange whose
result they share. Rotation persistence is per-credential and local —
`internal/vault`, AES-256-GCM, `0600` under `.polymetrics/vault` — with zero centralized custody.
The refresh token, client secret and access token never appear in argv, logs or error text: the
token endpoint's error body is never surfaced (RFC 6749 §5.2 lets it echo the grant back), and
transport errors are rendered through `safety.RedactErrorText`.

`ResolveCheckAuthSpec` statically validates `refresh_token` alongside
`token_url`/`client_id`/`client_secret`/`scopes`, and checks `refresh_token_store_key` as an
identifier, so both flow into `connectorgen validate` for free.

**Pagination — 7 types + none** (`bundle.go`'s `PaginationSpec`, `paginate.go`'s `newPaginator`):

| `type` | Fields used | When to use |
|---|---|---|
| `none` (default/omitted) | — | single-page endpoints |
| `link_header` | — | RFC 5988 `Link: <url>; rel="next"` header (GitHub-style) |
| `page_number` | `page_param`, `size_param` (empty string = never send a size param), `start_page`, `page_size` | 1-based/N-based page-number APIs; short-page stop when a page returns fewer than `page_size` records |
| `offset_limit` | `limit_param`, `offset_param`, `page_size` | offset+limit APIs |
| `cursor` (`token_path`) | `cursor_param`, `token_path`, optional `stop_path` | next-page token read from the response body; optional `stop_path` (gap-loop cycle-1 item 5) names a body path whose falsy value stops pagination REGARDLESS of whether the token itself is still non-empty (Zendesk's `meta.has_more`: its own docs warn the cursor properties may be populated even when `has_more` is false) — a spec that never sets `stop_path` keeps the exact prior stop-on-empty-token-only behavior; also loop-guards against the same token repeating twice in a row |
| `cursor` (`last_record_field`+`stop_path`) | `cursor_param`, `last_record_field`, `stop_path` | Stripe-style `starting_after`/`has_more`: next cursor = a named field on the **last record** of the current page; `stop_path` names a body path whose falsy value stops (its absence, or an empty/malformed page, is defensive "never loop forever") |
| `next_url` | `next_url_path`, `allow_cross_host` | absolute next-page URL read from the body (aircall-style); same-host SSRF guard by default (THREAT-MODEL §3) — set `allow_cross_host: true` to opt out; also loop-guards against requesting the same URL twice |
| `start_index` | `start_index_param` (default `startIndex`), `count_param` (default `count`), `total_path` (default `totalResults`), `start_index_path` (default `startIndex`), `start_index_base` (pointer; default 1 — an explicit `0` is honored for a 0-based server), `page_size` | 1-based index-plus-total pagination — the SCIM 2.0 list shape, RFC 7644 §3.4.2.4: `?startIndex=1&count=N` → `{totalResults, itemsPerPage, startIndex, Resources}`. Named for the mechanism, not for SCIM: any API that pages by an index and reports a total is served by the same walk. Every field defaults to SCIM's name, so a SCIM stream declares only `{"type":"start_index","page_size":N}` |

The `start_index` walk advances by the records the engine **actually extracted at the stream's records path** (`recordCount`), never by a server-claimed `itemsPerPage` — a server that claims 100 while returning 2 would otherwise make the walk skip 98 records silently. There is deliberately no declarable `items_per_page_path` field (the reason is recorded beside the type in `bundle.go`/`paginate.go`); a non-advancing index is a sticky `Err()` rather than a silent stop, mirroring the `tokenPathCursor`/`nextURL` loop guards, and `total_path` bounds the walk so an over-reported total can never make it loop.

A `stop_path` body value (both cursor variants) is read via `connsdk.StringAt`; ANY value other than
the literal string `"true"` (a JSON `false`, a missing path, or a read error) is falsy and stops
pagination — this is the exact rule to use for Zendesk-shaped `has_more` boolean stop signals: any
API whose real stop signal is a boolean, not just Stripe's shape, should declare `stop_path` on
whichever cursor variant it uses.

`cursor`'s `token_path` and `last_record_field` are **mutually exclusive**; declaring both, or
neither, is a **read-time** error, not a load-time one — `paginate.go`'s `newPaginator` (which
enforces this) runs from `read.go`'s `newRuntime`, once per `Read`/`Check` call, not once at bundle
load. `connectorgen validate` does not check pagination specs at all (no field/rule references
`PaginationSpec` anywhere in `cmd/connectorgen/validate.go`) — a malformed `token_path`+
`last_record_field` combination passes `connectorgen validate` cleanly and only surfaces the first
time the stream is actually read. `PaginationSpec.MaxPages` is a hard request-count cap enforced
in `read.go`'s `readDeclarative` loop, independent of page fullness. A positive
`connectors.ReadRequest.MaxPages` can tighten but never widen that declared cap; `0` leaves the
declared cap unchanged, and a negative request is rejected. The effective cap is checked *before*
issuing the request for that page number; when both caps are absent or zero, pagination is
unbounded. Stream-level `pagination` replaces the base-level spec **wholesale** (no field-by-field
merge) when present.

**`page_number`'s `start_page` supports an explicit 0-indexed start** (S4 engine mini-wave item 1:
algolia/auth0/beamer/braze/clickup-api/concord/customerly/dolibarr/harness/hubplanner and more —
each genuinely paginates from page 0, not 1). `PaginationSpec.StartPage` is `*int`, not a plain
`int`, specifically so `"start_page": 0` is distinguishable at the Go layer from an omitted
`start_page` key (both would otherwise decode to the same zero value). `newPaginator`'s
`page_number` case builds an engine-local `pageNumberPaginator` (NOT `connsdk.PageNumberPaginator`
— that type is unchanged and still coerces a zero `StartPage` field to 1, since connsdk is
read-only and every legacy Go connector still constructs it directly with an ordinary `int`);
`startPageOrDefault` maps a nil pointer to 1 (the historical default for every bundle that never
declares `start_page`) and returns a non-nil pointer's value verbatim, including 0. JSON
representation is unaffected — `"start_page": 0` was already schema-valid (`"type": "integer"`);
only the Go-side unset-vs-zero ambiguity needed fixing.

**`param_format` + `operator_prefix`** (`incremental.param_format`, applied by `read.go`'s
`formatIncrementalParam` to the lower-bound value before sending it as `request_param` or exposing
it as `{{ incremental.lower_bound }}`): `rfc3339` (default; send verbatim), `rfc3339_utc` (parse
digits, RFC3339, or date-only and emit UTC RFC3339), `unix_seconds`, and `date` (`2006-01-02`).
Optional `operator_prefix` prepends a bounded comparison operator (`>=`, `>`, `<=`, `<`) after
formatting, so connector definitions can express lower-bound query-qualifier shapes without
provider-named engine policies.

**`parseLowerBoundTime` accepts a bare `YYYY-MM-DD` date-only input, alongside digits/RFC3339**
(S4 engine mini-wave item 5: marketstack's `eod`/`splits`/`dividends` streams' real wire cursor
value for `param_format: date` is a bare date string with no time/offset component at all — neither
the all-digits Unix-seconds shape nor strict RFC3339 parses it). `parseLowerBoundTime` (used by
every `param_format` that needs to parse a timestamp — `unix_seconds`/`date`/`rfc3339_utc`, not
`date` alone) now tries three shapes in a fixed order: all-digits (Unix seconds) first, then
strict RFC3339, then bare `YYYY-MM-DD` (parsed as midnight UTC that date) last. The order is
unambiguous — a valid RFC3339 string always contains a `T` time-of-day separator (never all-digits,
never date-only-parseable), and a valid date-only string can never itself parse as RFC3339 — so no
input shape is ever misclassified as another. A malformed value (`"2026-13-40"`,
`"2026/01/02"`) still hard-errors at every stage; this is a strictly additive third fallback, not a
loosening of the existing two shapes.

**`{{ incremental.lower_bound }}` in `stream.Query`** (S3 engine mini-wave item 1, wave1-pilot
SUMMARY.md carried queue / REVIEW-A.md re-review R2 ACCEPT-WITH-QUEUE): exposes the RESOLVED,
post-`formatParam` incremental lower bound (state cursor, falling back to `start_config_key`; ""
when neither resolves — a fresh full sync, or a stream with no `incremental` spec at all) to
`stream.Query` template resolution, closing the recurring "a query param legacy sends ONLY when the
computed incremental lower bound resolves" gap class (chargebee's `sort_by[asc]=updated_at`, sent in
the SAME branch as `updated_at[after]`, never on a full-refresh read). Mechanically: `read.go`'s
`buildInitialQuery` now computes the lower bound and formats it via `formatParam` BEFORE
`stream.Query`'s own resolution loop runs, and wires the formatted value into that loop's `Vars`
(`Vars.IncrementalLowerBound`) — this is why ORDER matters and why this could not be expressed with
the pre-existing `omit_when_absent`/`default` dialect alone (item 3, below): that dialect's
absence-detection is keyed to a CONFIG/SECRET reference resolving or not, and there is no
config/secret key whose presence tracks "the incremental lower bound resolved" on the common
state-cursor-driven repeat-sync path (an app-persisted `State["cursor"]` is not a config key at
all). An absent lower bound is represented as the SAME `unresolvedKeyError` shape (namespace
`"incremental"`) the config/secrets dialect uses, so `omit_when_absent`/`default` compose with it
identically — no new tolerance mechanism was needed, just a new resolvable reference. Because the
literal VALUE a legacy connector sends alongside the lower bound is often a fixed constant (not the
lower bound's own value — chargebee always sends the literal string `"updated_at"`, never the
timestamp itself), pair this reference with the `const:<value>` filter (above) rather than sending
the raw lower-bound value:

```json
"query": {
  "sort_by[asc]": { "template": "{{ incremental.lower_bound | const:updated_at }}", "omit_when_absent": true }
}
```

— present with the literal `updated_at` iff the incremental lower bound resolves (state cursor or
`start_config_key`), absent on a full-refresh read. If a param genuinely needs the lower bound's OWN
value (not a fixed literal), omit the `const:` filter and reference `{{ incremental.lower_bound }}`
directly — it resolves to the exact same formatted string `request_param` itself receives.

**`computed_fields`** (`streams.json` per-stream): a map of output-field-name → template resolved
against the **raw** (pre-projection) record AND `config.*` (Config only — **never** `secrets.*`, see
below), so it can rename a field (`published_date` from `{{ record.publishedDate }}`), reach into
nested JSON (`{{ record.user.login }}`), join an array (`{{ record.engines | join:, }}`), stamp a
config-scoped marker (`{{ config.owner }}/{{ config.repo }}`, below), or inject a **static literal**
— a value with no `{{ }}` markers at all (e.g. `"stream": "search"`) passes through `Interpolate`
verbatim, since a template with no markers is a no-op by construction. This is the sanctioned way to
stamp a per-stream constant onto every emitted record (e.g. "which stream did this come from",
matching a legacy connector's own derived marker field — see searxng's `stream` field, §5). A
computed field whose source path is absent on a given record is silently skipped for that record
(not an error) — common for optional/differently-shaped nested fields across record variants (issues
vs PRs); this only applies to `record.*`-templated fields, not static literals (which never fail to
resolve).

**Typed extraction — bare `{{ record.<path> }}` copies the raw typed value** (gap-loop cycle-1 item
1, REVIEW-A.md adjudication A1: chargebee/gmail/github all had to widen a real integer/boolean
schema type to `["string", ...]` because `computed_fields`' `Interpolate` always stringified its
result — a JSON type change is an emitted-record-DATA change, not cosmetic, since every warehouse
destination derives column types from it; this recurred ≥3 times in one wave, meeting the §6
recurrence threshold). **When a `computed_fields` entry is a SINGLE bare `{{ record.<path> }}`
reference — no filter stage, no surrounding literal text, no second `{{ }}` segment — the engine
copies the RAW (pre-stringify) JSON value found at that record path straight into the projected
record, preserving its native type** (number/bool/null/object/array survive; schema types should
declare the field's REAL wire type, e.g. `"integer"`/`"boolean"`, never a widened
`["integer","string"]` union). Any other shape — a filter chain (`{{ record.tags | join:, }}`), a
mixed template with literal text or more than one reference (`"count={{ record.count }}"`), or a
static literal — is UNCHANGED and still produces a STRING via `Interpolate`, exactly as before. This
is what makes the increment backward compatible: only the bare-single-reference shape gets typed
extraction; every existing `join:`/rename/marker-stamp computed_fields keeps its current string
output untouched. When porting a pilot's stringify-workaround (chargebee's ~30 fields, gmail's 4,
github's `user_id`/`author_id`/`committer_id`/`workflow_run_id`), re-tighten the schema back to the
real type and flip the parity lock-in test from string-equality to native-type equality — do not
leave the widened union in place now that typed extraction exists.

**`coalesce` in `computed_fields` — first present, non-null, non-empty-string raw record value** (Pass B fidelity
follow-up mini-wave): a computed field may be a single expression of the form
`{{ coalesce record.a record.b record.c }}`. Each argument MUST be a `record.*` path; `config.*` and
`secrets.*` are not accepted in this form. The engine walks the paths left-to-right and copies the
first value that is present, non-null, and not the empty string `""` into the projected record,
preserving the raw JSON type exactly like the bare `{{ record.path }}` typed-extraction path above.
Only strings get the empty-value skip: numeric `0`, boolean `false`, empty arrays, and empty objects
are retained. If every path is absent, null, or `""`, the computed field is omitted; when filling the
field itself, put that field's current raw key first (e.g.
`"id": "{{ coalesce record.id record._id }}"`) to express legacy `id = id ?? _id` without
clobbering an already-present `id`.

**`length` in `computed_fields` — array length as a native integer** (Pass B fidelity follow-up
mini-wave): the exact single-reference form `{{ record.items | length }}` writes `len(items)` as a
Go integer into the projected record. A missing path, JSON null, or non-array value writes `0`,
matching legacy Go `len(x)`-style mapper behavior used for line/item count fields. This typed
integer behavior is scoped to `computed_fields`; elsewhere `length` remains an ordinary string
filter because `Interpolate` returns strings by contract.

**`response_fields` — top-level response values stamped onto extracted records**: when a legacy
mapper copied a response-level sibling onto every record extracted from a nested section (for
example OpenWeather's top-level `timezone` beside `hourly[]`/`daily[]`), declare
`"response_fields": { "timezone": "timezone" }` on the stream. Each map entry is
output-field-name → dotted response-body path. The raw value at that response path is copied into
each extracted raw record before projection and `computed_fields`; missing/null response paths are
omitted. Use this only for response metadata genuinely shared by every record in that response, not
as a replacement for ordinary `record.*` computed fields.

**`config.*` in `computed_fields` — Config only, Secrets is EXCLUDED by design** (gap-loop cycle-1
item 2, REVIEW-A.md adjudication A3/`ENGINE_GAP` G0): before this increment, `computed_fields`
templates were resolved against the record ONLY, so a legacy connector's config-derived per-record
marker field (github's `repository: "{{ config.owner }}/{{ config.repo }}"`, stamped on every record
of every stream) could not be expressed at all short of a 3rd Tier-2 hook interface (forbidden — see
§1's Tier-2 cap). `config.*` is now resolvable inside `computed_fields` exactly like every other
templating surface. **`secrets.*` is deliberately NEVER wired into this Vars environment** — a
computed field must never be able to copy a secret value into emitted record data (a record is what
flows to a destination warehouse; a secret leaking there is a credential-exfiltration path). A
`{{ secrets.* }}` reference inside a `computed_fields` template therefore still hard-errors exactly
as it always has (unresolved key) — this is intentional and permanent, not a gap to close later.

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

**`stream.Query` — an explicit opt-in optional-query dialect** (gap-loop cycle-1 item 3, REVIEW-B.md
cross-cutting adjudication 2: this recurring gap — vitally `status`, bitly `page_size`/`max_pages`,
calendly's `count`/page_size workaround, zendesk's dead keys, gmail's two filters, searxng wave0
F6 — hit the ≥3-occurrence threshold several times over). Each `streams.json` `query` entry
(`engine.QueryParam`) is declared EITHER as a plain JSON string (today's exact pre-existing
semantics: the string IS the template, resolved unconditionally via `Interpolate`; an absent
referenced `config.*`/`secrets.*` key is **always a hard error**, zero migration risk for every
existing bundle) OR as an object:

```json
"query": {
  "page[size]": "100",
  "status":   { "template": "{{ config.status }}", "omit_when_absent": true },
  "count":    { "template": "{{ config.page_size }}", "default": "100" }
}
```

`omit_when_absent: true` means the param is left off the request ENTIRELY when its template hits an
unresolved `config.*`/`secrets.*`/`incremental.*` key (vitally's `status`, bitly's `size` config
override, chargebee's `sort_by[asc]` incremental-lower-bound gate above) —
**scoped to config/secrets/incremental absence only** (`read.go`'s
`isUnresolvedConfigSecretOrIncremental`, extended by S3 item 1 to cover the `incremental` namespace
alongside `config`/`secrets`): any OTHER interpolation failure (CRLF injection, an
unknown filter/namespace) still hard-errors regardless. `default` sends that literal instead of
erroring when the referenced key is absent (closes calendly's `page_size`-defaults-to-100 gap and
the same class of legacy-default-base-URL-style shape at the per-param level — see also `default`
materialization at the config layer, below). Declaring both is unusual (contradictory intents); pick
one. **This is deliberately NOT a blanket absent-key-falsy change to query templating** — a plain
string entry keeps hard-erroring on a typo'd/missing REQUIRED key exactly as before; only an entry
that explicitly opts into the object form gets any tolerance at all, so a mis-declared required
query param can never silently degrade into an unfiltered request (the F4 fail-open class the engine
deliberately rejects elsewhere). Static validation (`connectorgen validate`'s
`checkInterpolations`) still resolves EVERY entry's `template` field, string or object — the
referenced key must still be DECLARED in `spec.json`'s properties either way (`config.*` keys) or a
known engine pseudo-namespace key (`incremental.lower_bound`, checked against
`knownIncrementalKeys` rather than `specKeys` — it is not a spec.json property). A `spec.json`
property that no template anywhere in the bundle ever consumes is still dead config — don't declare
it (F6, REVIEW.md).

**GraphQL stream variables** (`stream.graphql.variables`): GraphQL reads use a fixed `document` and
`operation_name` in `streams.json`; the document itself must not be templated. Variable values may
be constants, nested JSON values, or template objects:

```json
"graphql": {
  "operation_name": "ViewDiscussion",
  "document": "query ViewDiscussion($owner: String!, $repo: String!, $number: Int!) { ... }",
  "variables": {
    "owner": "{{ config.owner }}",
    "repo": "{{ config.repo }}",
    "number": { "template": "{{ query.number }}", "type": "integer", "default": "7" },
    "after": { "template": "{{ cursor }}", "omit_when_empty": true }
  }
}
```

`query.*` references come from connector command flags or read request query values, not
`spec.json`. Static validation accepts that namespace so parameterized command streams can be
declared without adding command-only fields to connection config. `omit_when_empty` removes an
empty resolved variable, used for first-page cursors. `default` is a string fallback used only when
the template hits an unresolved `config.*`, `query.*`, or `incremental.*` key; the default is then
converted through `type` (`string`, `integer`, `number`, or `boolean`). Do not use GraphQL variables
as a raw API escape hatch: documents stay fixed and reviewed, and mutations still require explicit
write actions with plan/preview/approval/execute.

**`client_filtered`** (`incremental.client_filtered: true`): for APIs with no server-side
incremental filter parameter, the engine drops already-seen records client-side by comparing each
record's cursor field against the lower bound (strictly greater survives). Use only when the API
truly cannot filter server-side — prefer `request_param` whenever the API supports it.

**Projection**: `stream.projection` is `"schema"` (default: only schema-declared properties
survive) or `"passthrough"` (every raw field survives, unfiltered). `computed_fields` are always
applied after projection regardless of mode.

**Sub-resource fan-out — `stream.fan_out`** (S4 engine mini-wave item 2: appfollow/bigmailer/
breezy-hr/campayn/eventzilla/everhour/finnworlds/k6-cloud/metricool/cisco-meraki/configcat and 15+
quarantined/partial connectors whose real read is "list N parent ids, then repeat the WHOLE
per-stream request/pagination/incremental sequence once per id, stamping the parent id onto every
child record" — this is the one legitimate Tier-2 `StreamHook` trigger this dialect addition
retires for every bundle it now covers):

```json
"fan_out": {
  "ids_from": { "config_key": "app_collection_ids" },
  "into": { "query_param": "apps_id" },
  "stamp_field": "app_id"
}
```

`ids_from` is EXACTLY ONE of `config_key` (a comma-separated config value, split/trimmed/
empty-entries-dropped — appfollow's `app_collection_ids`) or `request` (one preliminary GET, fully
paginated to exhaustion using the CHILD STREAM'S OWN effective pagination spec — base or
stream-level override, the id-listing request declares no pagination block of its own —
extracting `id_field` off every record found at `records_path`):

```json
"fan_out": {
  "ids_from": { "request": { "path": "/projects", "records_path": "data", "id_field": "id" } },
  "into": { "path_var": "parent_id" },
  "stamp_field": "project_id"
}
```

Declaring both `config_key` and `request`, or neither, is a **read-time** error (mirroring cursor
pagination's `token_path`/`last_record_field` mutual exclusivity) — `connectorgen validate` does
not check this (same reasoning as pagination specs above: no field/rule references `FanOutSpec`).
`into` is EXACTLY ONE of `query_param` (the resolved id is added as a query parameter on every
request of that id's sub-sequence) or `path_var` (the resolved id becomes referenceable in the
stream's own `path` as `"{{ fanout.id }}"` — a new engine pseudo-namespace, resolved via
`interpolate.go`'s `resolveFanoutRef`; unlike `incremental.lower_bound`, an unresolved `fanout.id`
reference is a HARD ERROR, never absent-tolerant, since it only ever appears inside a fan_out-driven
path and a missing id there is a real bug). `stamp_field`, when set, writes the current fan-out id
onto that field of every emitted record AFTER projection/`computed_fields`, exactly once per
sub-sequence — the bundle author never declares it as a `computed_fields` entry themselves.
Resolution happens ONCE per `Read()` call, before any per-id sub-sequence starts; each id then runs
the identical declarative request/pagination/incremental/filter/project/computed_fields/hook
sequence an ordinary (non-fan-out) stream runs — pagination, incremental state, the effective page
cap, and rate-limiting are all independent PER id (a fresh paginator + fresh base query per id),
never shared across the fan-out. The caller cap also bounds a request-form ID listing.
`connectorgen validate`'s `checkInterpolations` walks
`fan_out.ids_from.request.path` with the same `ResolveCheck` coverage `stream.path` gets; `fanout.id`
is checked against a `knownFanoutKeys` set (mirroring `knownIncrementalKeys`) rather than
`specKeys`, since it is an engine-provided pseudo-namespace, not a spec.json property.

**Keyed-object flatten — `records.keyed_object`/`records.key_field`** (S4 engine mini-wave item 3:
appfigures' `products`/`sales`/`ratings`/`categories` streams and similar APIs whose list endpoint
returns a JSON OBJECT keyed by an arbitrary id — `{"111": {...}, "222": {...}}` — rather than an
array):

```json
"records": { "path": "products", "keyed_object": true, "key_field": "product_id" }
```

`records.path` still selects where the object lives in the page body (the SAME dotted-path
convention `connsdk.RecordsAt` uses elsewhere); `keyed_object: true` explodes EACH VALUE of that
object into its own record — `connsdk.RecordsAt`'s ordinary behavior of returning a bare object as
ONE whole-object record does not apply once this flag is set. A value that is not itself a JSON
object (a scalar, array, or null) is silently skipped, mirroring `RecordsAt`'s own tolerance for
non-object array elements. `key_field`, when set, stamps the source map key onto that field of the
exploded record BEFORE projection, so it participates in ordinary schema projection/computed_fields
like any other raw field. Records are emitted in ascending sorted-key order for deterministic
output (Go map iteration order is randomized). Implemented as an engine-local `recordsAtKeyed`
(read.go) reimplementing connsdk's unexported decode+selectPath plumbing rather than an exported
connsdk addition — connsdk itself needed no change. No `connectorgen validate` rule was added (no
templated field on `RecordsSpec` to statically check); the positive-control corpus case
(`keyed-object-valid`) proves the shape loads and validates cleanly instead.

**Page-count cap**: see the pagination table above. The effective hard cap and the paginator's
short/empty-page stop signal are independent and must both be considered when reasoning about
termination.

**`spec.json` `"default"` values ARE now materialized into `RuntimeConfig.Config`** (gap-loop
cycle-1 item 6, REVIEW-A.md C3 — RESOLVED: previously `default` was accepted-but-only-preserved,
never read back out anywhere, so EVERY migrated connector hard-errored on a config shape legacy
accepted, e.g. an unset `base_url` when legacy derived `https://api.github.com`/
`https://api.monday.com/v2`/etc. in code). Every engine request path that resolves bundle
configuration (`engine.Check`, `engine.Read`, direct-read operations, `DryRunWrite`, and `Write`)
calls `materializeConfigDefaults` before template resolution: for every `spec.json` property that
declares a `"default"` AND is genuinely ABSENT from the caller's `RuntimeConfig.Config` (a key
already present — even as an explicit empty string — is NEVER overridden), the default's JSON value
is stringified and filled in. This is the single, uniform mechanism for every legacy base-URL
default: `base_url: {{ config.base_url }}` + `"default": "https://api.github.com"` on the `base_url`
property now round-trips exactly like legacy's in-code fallback, with no template-level special
case needed. For a DERIVED default (sentry's `hostname`-based URL, chargebee's `site`-based URL —
the base URL is a function of another config value, not a fixed literal) this mechanism alone is not
enough; either require `base_url` and drop the derivation (documented config-surface narrowing,
ledgered), or express the derivation as a `computed_fields`-style template if/when the dialect grows
one for base-URL construction — do not invent ad hoc Go for it (Tier-2 escalation only if genuinely
needed). **Validate rule**: `connectorgen validate`'s `default_type_mismatch` rule hard-fails a
`spec.json` property whose `"default"` value's JSON type does not match its own declared `"type"`
(e.g. `"type":"integer","default":"not-a-number"`) — a mismatched default would otherwise silently
materialize a wrong-shaped config value into every defaulted engine request.

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

**Provider-cited `rate_limits.json` is an enforced HTTP policy declaration**: this optional,
closed-schema file records a provider policy only when it can be cited. A declared policy
must carry an HTTPS provider artifact URL without userinfo, query parameters, or credential-like
query-style fragment parameters (an ordinary documentation anchor fragment is allowed) and an ISO
`retrieved_at` date. Every file starts with `schema_version: 1`; `version` is optional
context, not a substitute for a retrieval date. Model the provider shape rather than flattening it
to requests per minute: selectors can target an endpoint, tier, and auth type, while
`exclude_endpoints` refines such a selector by removing exact declared endpoints; it cannot stand
alone or combine with `all`. Budgets label their
`burst` or `sustained` dimension, `requests` or `points` unit, and fixed/sliding-window or
token/leaky-bucket replenishment model. Cost-weighted APIs can declare a default cost and the
provider response header that reports cost. A policy may name at most one actual-cost header;
when two policies can match one request, they must use the same header when both declare one.

The root state is deliberately explicit: `declared` requires one or more cited policies, while
`unknown` and `not_applicable` require a nonblank reason and cannot carry a policy. A policy's
`scope.subject_kind` is one of `account`, `installation`, `application`, `endpoint`, or `ip`, and
`scope.subject_config` names the corresponding **non-secret** `spec.json` configuration property.
The loader rejects a missing or `x-secret` property. The declaration records only that property
name; it never carries a runtime subject. At runtime the
engine gives the transient config value to `connectors.CoordinationIdentity.RateScopeKey`, so the
registry key is the credential binding plus policy ID plus the non-secret subject as an opaque salted
projection. Never put a credential, token-derived value, or runtime subject value in the declaration,
logs, events, or persisted state. A subject kind outside this vocabulary is refused.

For each outbound engine request, the runtime resolves every matching declared policy and admits all
of their budgets before the logical requester send. `all` matches the whole HTTP connector and cannot
combine with any other selector; an endpoint selector matches its declared method/path pair;
`exclude_endpoints` removes exact declared method/path pairs from a non-`all` selector; optional
`tiers` and `auth_types` match the non-secret `config.tier` and
`config.auth_type` values as additional AND conditions. Check requests, stream pages (including
pagination), direct and operation direct reads, declarative and operation writes (form, JSON, and
multipart), and binary downloads use this same requester admission path. A selector with
`exclude_endpoints` is not attached to a whole-connector hook requester because that requester has
no declared path to evaluate.
The process-local registry enforces each declared burst/sustained request/point budget with its
fixed-window, sliding-window, token-bucket, or leaky-bucket model. A declared actual-cost response
header tightens a point budget when it reports a higher cost; it never credits capacity from a lower
or absent value. A reset timestamp hard-blocks only after `Retry-After`, a 429, or an exhausted
remaining budget; non-exhausted reset metadata only tightens state. #3755 still
owns operator-visible output; this mechanism does not emit rate-limit events itself.

**Coordination remains process-local unless a policy explicitly says otherwise.** Omit
`coordination` for dependency-free, per-`pm`-process protection. That is not account-wide or
cross-process coordination, and connector inspection says so. A provider policy may instead set
`"coordination": "require_shared"` to use the optional Redis-compatible coordinator. The engine
then verifies that coordinator before it sends a request and returns a typed unavailable reason if
it cannot enforce the shared budget; it never falls back to a local limiter. This opt-in is per
policy only: do not infer it from an endpoint, a runtime address, another policy, a connector, or
credential configuration. The shared key remains the existing opaque
`CoordinationIdentity.RateScopeKey` projection; never put a raw subject, binding, credential, or
coordinator address in a declaration or output.

An absent declaration, `unknown`, `not_applicable`, or a non-matching selector leaves the requester
unchanged. `streams.json` `base.rate_limit` remains the legacy page-loop limiter: it is neither
created nor replaced by `rate_limits.json`, and when both apply its old wait runs independently of
the declared requester admission. Do not use this declaration to add a `streams.json`
`base.rate_limit` throttle or change legacy behavior. Because Go's `embed` directives reject an
unmatched optional wildcard, `internal/connectors/defs/defs.go` intentionally adds
`*/rate_limits.json` only with the first production declaration; include that embed update in the
same migration so the shipped CLI can read the file.

**Write body construction** (`write.go`): `body_type` is one of `"json"` (default), `"form"`, `"none"`, `"graphql"`, `"json_array"`, `"multipart"`, or `"base64_upload"` (the `body_type` enum in `internal/connectors/engine/schema/writes.schema.json` is the authority). Default body = every record field **except** those named in `path_fields` (the path already carries them, e.g. `{{ record.id }}` for an update). `body_fields` (if set) restricts the body to an explicit allow-list instead (used for delete-with-body actions). `"form"` builds a `url.Values` body (Stripe-shape — compare `stripe/write.go`'s `customerForm`), sorted keys for deterministic encoding, empty-string values omitted. `"none"` with no `body_fields` sends no body at all (pure path-parameterized mutation/delete). For a JSON endpoint whose provider contract requires a body even when no body fields resolve, set `body_required: true`; it sends `{}` and is rejected on every non-JSON body type.

**Write retries require operation-scoped evidence** (`write.go`'s `writeRequester`): an action may
declare `idempotency_key_header` only when provider documentation covers that operation. The runtime
then generates one fresh 128-bit key per record and reuses it only for automatic attempts of that
record, overriding any same-named base header. An unkeyed action retains automatic retries only when
it is a delete with `delete.idempotent: true`; every other unkeyed mutation runs once, including no
automatic 401 refresh, so an ambiguous transport/provider result is surfaced instead of replayed.
Do not infer idempotency from the HTTP method or apply a provider-wide header without
operation-level evidence.

`"base64_upload"` is a **typed** JSON body, not a raw request escape hatch — method, path and body structure stay bundle metadata. It carries a base64-encoded payload in exactly one declared JSON property, everything else being an ordinary record field governed by the action's closed `record_schema`. Declare it via a `base64_upload` block with `source` (`"path"` default — read a local file — or `"base64"` — take an already-encoded string), `source_field`, `content_field`, `max_decoded_bytes` (required, positive, clamped to a 16 MiB engine ceiling — an unbounded inline upload is a memory-exhaustion vector), and optional `max_encoded_bytes` (defaults to the base64 length of the decoded bound). Two properties every author must know: the `source_field` is **removed from the body before transmission** — in `"path"` mode it holds a local filesystem path, and transmitting it would leak the operator's directory layout to the provider; and in path mode the file is read under `os.Root` containment (refuses traversal, symlink escape and the check-then-open TOCTOU race in one primitive, not a lexical prefix check), bounded (read one byte past the limit and REJECT rather than truncate — a truncated attachment is a silently corrupt upload), and verified against the approved-payload SHA-256 digest exactly as the `multipart` file part does, so plan/preview/approve/execute still binds the approved bytes.

`"multipart"` is the streaming counterpart for `multipart/form-data` uploads, and is what a provider documenting a file/image endpoint with a size cap wants. Each entry in `multipart.parts` is a `field` part (a record value sent as a form field) or a `file` part (a local file streamed as an upload). A `file` part is subject to the same containment and bounding rules as `base64_upload`'s path mode — `os.Root` confinement, a size pre-check, a read one byte past the limit that REJECTS rather than truncates, and approved-digest verification — but it streams through a bounded temp snapshot instead of buffering the whole payload, so the per-part `max_bytes` and the form-wide `multipart.max_bytes` are the real memory bound rather than an engine ceiling.

A `file` part may also declare `allowed_media_types`, which bounds what the file's **own bytes** may sniff as (`http.DetectContentType`), checked before any request is made. Three consequences worth knowing: it is the only restriction on a part's type, so **a single-entry list is how a bundle demands exactly one type**; when it is present the part's wire `Content-Type` header is set from the sniffed type rather than from the declared `content_type`, because the sniffed type has just been proven to be one the bundle accepts and asserting the declared one over disagreeing bytes would be a claim we had already falsified; and when it is absent the declared `content_type` is sent untouched, since nothing verified the sniff and `http.DetectContentType` is coarse (every CSV sniffs as `text/plain`). Omit the key to leave a part unconstrained — a present-but-empty list is a load-time error rather than a silent "allow anything", and a `content_type` outside its own `allowed_media_types` is rejected at load time as an unsatisfiable declaration.

**Operation-level multipart `rest_write`** (`operations.json`, not `writes.json`): use this only when a connector author has provider evidence for one fixed multipart mutation and is ready to bind it to a typed command. It extends the existing `rest_write`/direct-write executor; it does **not** introduce a generic uploader or make a MIME annotation executable. Preserve the provider operation URL in `source_url` so the method, fields, media types, and size limits remain reviewable. This authoring evidence does not make shipped endpoint validation provider-surface provenance; the endpoint-validation boundary below owns that distinction. The complete, closed declaration has this shape:

```json
{
  "id": "example.attachments.create",
  "kind": "rest_write",
  "summary": "Create one attachment",
  "source_url": "https://provider.example/docs/attachment-upload",
  "risk": "high",
  "approval": "plan-preview-confirm-execute",
  "output_policy": "json",
  "mutation_class": "create",
  "rest": {
    "method": "POST",
    "path": "/attachments",
    "content_type": "multipart/form-data",
    "max_bytes": 1048576,
    "body_schema": {
      "type": "object",
      "additionalProperties": false,
      "required": ["message", "media_file_path"],
      "properties": {
        "message": { "type": "string" },
        "media_file_path": { "type": "string" }
      }
    },
    "multipart": {
      "max_bytes": 26214400,
      "parts": [
        { "name": "message", "type": "field", "field": "message", "required": true },
        {
          "name": "attachment",
          "type": "file",
          "field": "media_file_path",
          "required": true,
          "max_bytes": 26214400,
          "content_type": "application/pdf",
          "allowed_media_types": ["application/pdf"]
        }
      ]
    }
  }
}
```

`rest.multipart` is valid only for `kind: "rest_write"`. `content_type` is the exact literal `multipart/form-data` — do not add a boundary; the requester owns it. The method, connector-relative path, output policy, part names, field mappings, declared media policy, and every cap are bundle-owned. `rest.max_bytes` is a required, positive **response-capture** cap; `multipart.max_bytes` is the required, positive aggregate upload cap; and each `file` part has its own required positive cap. The body schema is closed, every part must name a declared typed body field, and each file field is a required string path. A file source field may use any declared name; runtime identity binding comes from the file-part declaration, not a `file_path` naming convention. That is what rules out an inline byte body, an arbitrary file part, arbitrary headers, a caller-selected method/path, or a raw request body. A command adoption may map only its fixed flags to declared `body.*` fields; it must not add a generic file/method/path/header/body input.

A command adoption must declare exactly one connector-relative mutating endpoint in
its `cli_surface.json` `api_surface` field.
For a disk-backed bundle, the operation's fixed method/path must also cross-check
an `api_surface.json` entry marked as an `operation`. Shipped builds cannot read
that file: their endpoint validation is derived only from the included
`rest_write` declarations, so it proves consistency within those declarations,
not provider documented-surface provenance. #3773 owns the separate per-operation
`api_surface` provenance foundation. Neither check is an execution allowlist.

All multipart mutations retain the plan → preview → approval → execute lifecycle. Preview is network-free and binds the resolved method, target/query, the full multipart declaration, typed form values, source-path identities, and the approved SHA-256 for every file. Execution re-prepares that canonical request, refuses a stale preview or changed/missing file before network dispatch, then uses the existing project-root confinement, regular-file, cap, snapshot, digest, and media checks. Declared `rest_write` multipart calls are single-attempt (`DisableRetries=true`): non-idempotent calls are never retried, and redirect replay remains refused. Runtime response and error content is preserved in full subject only to the declaration's response-capture bound; this flow does not add a masking policy.

This is separate from the already executable reverse-ETL `writes.json` `body_type: "multipart"` path. Gong's `upload_call_media` action and `pm gong calls upload-media` command are its existing proof; no operation-level connector adoption is implied here. The legacy `operations.json` `kind: "file_upload"` remains planned/non-executable until a connector moves each endpoint to a complete declared contract and proves it. This shared-runtime documentation makes **no** GitLab, Freshchat, Gong, or other provider operation newly available. CLI/help/manual/website parity is therefore not applicable to this foundation: each adoption lane must update its own runtime help, `docs/cli/**`, website docs, generated manuals, command surface, and executable evidence before claiming `availability: implemented`.

`redact_fields` is an action-local list of record paths whose values remain masked in generic
source-table plan samples and returned write errors. It is for non-secret identifiers or clinical
values that can appear in templated paths or upstream error text; reverse-plan creation persists the
list and masks matching sample fields. `DryRunWrite` deliberately does not apply that masking to its
resolved request line, so approval sees the method and URL that execution will use. `Write` redacts
raw and URL-encoded literal forms from returned write errors while preserving typed error wrapping.
`cli_surface.json` declarations remain load-compatible metadata, but `commandrunner` does not use
them to mutate connector-command records or errors, or to forward them to executors.

`confirmation` is the closed confirmation declaration for new actions:
`"confirmation": {"kind": "destructive"}`. The writes and operations schemas are authoritative;
existing `confirm: "destructive"` bundles remain compatible, but do not copy that legacy spelling
into new authoring. See the architecture design's [write semantics](../architecture/connector-architecture-v2-design.md#b5-write-path-enginewritego)
for fail-closed normalization and preview-bound execution. Declaring confirmation in
`operations.json` does not make an operation executable or create a command binding.

`batchable` declares whether the action may run from a **bulk** reverse ETL plan — the
`pm reverse plan --source-table ...` shape that fans one action out over many warehouse rows under
a single approval. It defaults to `true`; omit it unless you mean to restrict the action. Declaring
`"batchable": false` makes `PlanReverseETL` refuse the action before it stores a plan or mints an
approval token, and makes `RunReverseETL` re-check the live manifest before executing an
already-stored bulk plan. The action remains fully executable as its own `pm <connector> <command>`
via `cli_surface.json`, which is the entire point: it stays available to a human invoking it one
record at a time.

Declare it for operations that must never be bulk-automated — moderation actions, irreversible
sends, rate-sensitive endpoints, and anything governed by a provider rule about human intent. Do
**not** reach for it as a severity signal: that is `confirmation`'s job, and the two are
independent. An action can be non-batchable without being destructive (casting a vote) or
destructive without being non-batchable (a bulk delete), so neither one implies the other.

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
- Stream fixtures may include `"read_query": {"flag_or_param": "value"}` for parameterized reads
  whose runtime `ReadRequest.Query` values are not URL query parameters, such as fixed GraphQL
  documents that take command flags in the POST body. Use this only for replay input; do not model
  required command flags as GraphQL variable defaults.
- A stream fixture response may include `"headers"`, with each header value written as either one
  string or an array of strings. The replay harness expands `{{ base_url }}` inside those values to
  its loopback origin; use that placeholder for recorded `Link` pagination instead of committing an
  ephemeral test-server URL. When `Content-Type` is absent, replay still defaults it to
  `application/json`.
- **A 2-page fixture is REQUIRED whenever the bundle declares pagination** for at least one
  stream (`conformance`'s `pagination_terminates` dynamic check needs a second page to prove the
  engine consumes each page exactly once and terminates). See
  `fixtures/streams/customers/{page_1,page_2}.json` in the stripe golden: page 1 sets
  `has_more: true` and 2 records; page 2's request carries the expected `starting_after` query
  param and sets `has_more: false`.
- **Sanctioned exception — `next_url` pagination MAY ship a single-page conformance fixture**
  (gap-loop cycle-1 item 7, formalizing bitly's REVIEW-B.md finding 3 / fan-out-readiness item 4,
  and calendly's identical accepted shape for the same reason): a `next_url` stream's next-page URL
  is the REPLAY SERVER'S OWN address, unknown until the harness picks a port at runtime — a static
  fixture file literally cannot embed the correct absolute URL for a second page. This is a genuine
  harness limitation, not a fixture-authoring shortcut: ship a single-page fixture for the
  `next_url` stream (satisfies `fixtures_present`/`read_fixture_nonempty`; `pagination_terminates`
  exercises a DIFFERENT, non-paginated stream in the same bundle instead — see bitly's
  `organizations` stream), and prove real 2-page `next_url` correctness with a LIVE
  `paritytest/<name>` test that drives an actual `httptest.Server` and asserts the second page is
  requested with the expected query/params (bitly's
  `TestParityBitly_BitlinksStreamPaginates`, calendly's
  `TestParityCalendly_ScheduledEventsTwoPagePagination`). Do not invent a fake absolute URL in the
  fixture to work around this (it would never match the actual replay server's origin and would
  either fail the SSRF guard or require weakening it) — the single-page-fixture-plus-live-parity-test
  pattern is the correct, honest representation, not a corner cut.
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
- `fixtures/writes/<action>.json`: `{"record": {...}, "expect": {"method", "path", "query"?, "body"?},
  "response"?: {"status", "body"}}` — used by `write_validate`/`write_request_shape`; the engine's
  dry-run/actual request must match the declared method/path and every expected query/body entry
  for a valid `record`, and a deliberately
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
  behavior, not just assert the skip — e.g. `"hook-covered; proven live by hook tests and archived
  pre-deletion parity evidence"`, optionally citing the specific archived test
  (`TestParitySentry_IssuesTwoPagePaginationAndResultsFalseStop`). Hook/native unit tests plus
  archived parity evidence are the authoritative correctness bar for any hook-covered behavior a
  skip marker names; a marked connector's `migrated` status still requires those checks to be
  green. The marker is an honest description of WHERE the proof lives, never a way to avoid proving
  the behavior at all.
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
| 11 | ip2whois | Legacy publishes registrant/admin/tech/billing contact records under ONE catalog stream named `contacts` (primary key `[domain, role]`). This bundle publishes 4 separately named streams instead (`contacts_registrant`/`contacts_admin`/`contacts_tech`/`contacts_billing`), one per role, because `records.path` supports exactly one dotted path per stream declaration — there is no declarative way to select several independent named sub-object keys (`registrant`/`admin`/`tech`/`billing`, siblings of an unrelated `registrar` key at the same nesting level) into a single stream; `records.keyed_object` would wrongly explode ALL object-valued sibling keys at that path (including `registrar`), not just the 4 contact roles. Every field legacy emits for every role (name/organization/street_address/city/region/zip_code/country/phone/fax/email, plus domain/role) is preserved verbatim; only the catalog stream-name/count changed, never a record's data for any domain a legacy-valid config would produce. | ACCEPTABLE |
| 12 | ip2whois | `nameservers` is NOT migrated: legacy's `nameserverRecords` fans the raw `nameservers` field — a bare JSON array of scalar strings, not objects — out into one record per nameserver (`{domain, nameserver}`). Neither `connsdk.RecordsAt` (keeps only array elements that decode as a JSON object; a bare string element is silently dropped, yielding zero records) nor `records.keyed_object` (explodes a JSON object's VALUES, which must themselves be objects) fans a scalar-valued array into one record per element. Emitting a single `join:,`-joined string instead (as the `whois` stream's own `nameservers` field already does) would change record CARDINALITY versus legacy's genuine 1-record-per-nameserver stream — an accepted-input emitted-DATA change, not cosmetic. See `docs/migration/quarantine.json`'s superseded ip2whois entry (the original per-domain-request-iteration + whole-object-fan-out blocker is now solved by the `fan_out` dialect; this is a narrower, different, still-real gap). | ENGINE_GAP |
| 13 | jamf-pro | Legacy also stops pagination early when the running record count reaches the response body's `totalCount` field (`jamfTotalCount`), in addition to the short-page stop. The engine's `page_number` paginator implements only the short-page stop signal. This can cause, at most, one harmless extra request on the rare page where a full-size page happens to exactly exhaust `totalCount` (the following request then returns an empty/short page and stops normally) — it never omits, duplicates, or reorders any record for any input legacy itself would accept. | ACCEPTABLE |

Several engine gaps discovered by the goldens and two review passes were **closed**, not left as
deviations (see `.planning/phases/wave0-engine-harness/traces/waveF-repair-ledger.md`,
`gaploop-r1-ledger.md`, `gaploop-r2-ledger.md`): `PaginationSpec.MaxPages` enforcement; `EvalWhen`
absent-key-falsy semantics (both `waveF-repair-ledger.md`; item 8 above predates the latter);
`formatParam`/`parseLowerBoundTime` digits-only (Unix-seconds) passthrough for
`unix_seconds`/`date`/`rfc3339_utc` with `operator_prefix` (B1 — the honest shape
`internal/app`'s persisted cursor actually takes); stream/check path interpolation (F1); filter
chains, `join:<sep>`, and
static-literal `computed_fields` (F9/F7 enablement); `link_header`'s SSRF host/scheme guard and
fail-closed-on-unparseable-URL (M1/F2/m2); `lastRecordCursor`'s configurable records path + numeric
last-record ids (F3); header resolution's hard-error-on-absent-secret (F4); `Bundle.RawSpec`
verbatim `Definition().Spec` (F5); `AuthHook`'s real caller context (F8); numeric-cursor support in
`cursor_advances` (B2, motivating item 2's RESOLVED above); certify's `secret_redaction` full
stdout/stderr/report-JSON scan and JSON-escaped-secret detection (M2/m4); `connectorgen validate`'s
`engine.ResolveCheckAuthSpec` wiring. When you hit any of these shapes again, rely on the
fixed/wired behavior directly — do not re-document as an open gap or re-introduce a workaround.

The wave1-pilot gap-loop cycle-1 engine mini-wave (`.planning/phases/wave1-pilot/traces/
gaploop-s1-ledger.md`) closed six more, per REVIEW-A.md adjudications A1/A3/C3 and REVIEW-B.md
cross-cutting adjudications 1-2 + zendesk finding 2: typed `computed_fields` extraction for a bare
`{{ record.<path> }}` reference (A1 — see the typed-extraction paragraph above); `config.*` (never
`secrets.*`) wired into `computed_fields`' Vars (A3/`ENGINE_GAP` G0); the opt-in optional-query
dialect on `stream.Query` (`engine.QueryParam`, REVIEW-B.md adjudication 2); the `last_path_segment`
interpolation filter (REVIEW-B.md adjudication 1); `stop_path` + a loop guard on the token_path
cursor paginator (`tokenPathCursor`, zendesk finding 2); and spec.json `"default"` materialization
into `RuntimeConfig.Config` at `Read`/`Check` time (C3), plus its `default_type_mismatch` validate
rule. Pilots carrying the pre-increment workarounds for any of these (chargebee/gmail/github's
stringify-widened schemas for A1; calendly's dropped `id` for the filter; zendesk-support's invented
incremental filter and unguarded `has_more` for the paginator; sentry/chargebee's dead
hostname/site config for C3) are Step-2 (pilot repair wave) scope — re-tighten those bundles to use
the now-available engine mechanism instead of re-deriving the workaround.

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

When hook/native behavior is touched, prefer the connector's focused hook/native tests plus
`go test ./internal/connectors/conformance -run 'TestConformance/<name>'`. Whole-repo hygiene
(run whenever touching anything beyond your exclusive dirs is even a remote possibility):
`go build ./... && go vet ./...`, `make lint`, and `make connector-boundary`.

**FORBIDDEN files (never touch, regardless of connector)**: generated hook/native import sets
(`internal/connectors/hooks/hookset/hookset_gen.go` and generated nativeset files) except via
`go run ./cmd/connectorgen gen`, `icon_data.json`, any top-level `internal/connectors/*.go`,
`go.mod`/`go.sum` (a new dependency is a `NEEDS_NEW_DEP` blocker, never a self-serve add), and
any other connector's `defs/`/`native/`/`hooks/` directory.

**No-commit rule**: migration agents do not run `git commit`. The orchestrator commits once per
wave-close after the path-guard (`git status --porcelain` limited to assigned dirs) passes.

## §8 Post-wave2 review rules (mandatory for wave3+)

1. **Projection decision**: declare `projection: "passthrough"` iff the legacy read path emits
   records verbatim (no field-built `connectors.Record{...}` mapping). Schema-mode projection on a
   verbatim-emitting legacy silently drops fields — meta-rule violation.
2. **Incremental truth table**: bare `incremental.cursor_field` iff legacy publishes CursorFields
   in its catalog; `request_param` iff legacy sends a server-side filter; neither → no incremental
   block (keep `x-cursor-field` in schemas only).
3. **Fixture/live separation**: live config (page sizes, limits) must reproduce legacy defaults —
   never inherit fixture conveniences. Fixture request values for templated config are the literal
   `"synthetic-conformance-value"`; page-1 fixtures are FULL pages when page_number/offset
   pagination is declared (token/cursor types stop on token absence instead).
