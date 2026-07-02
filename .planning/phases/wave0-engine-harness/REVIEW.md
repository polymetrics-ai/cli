# REVIEW — wave0-engine-harness (gsd-loop-reviewer, 2026-07-02, HEAD b3f91af)

Scope: full phase diff `main...HEAD` (~37k insertions). All engine sources, three golden bundles,
native/postgres, conformance, certify, connectorgen/inventorygen/registrygen, lint/Makefile gates,
migration docs read in full. Verification re-run locally: `go build ./...` + all connector/cmd
tests PASS (exit 0); `go test -cover`: engine **85.0%** (gate ≥85 — zero margin), conformance
84.3%, certify 81.0%.

Verdict: **NO-GO for wave1-pilot until the 3 BLOCKs below are fixed.** All are small and
well-scoped; everything else is FLAG (fix during pilot).

---

## Dimension 1 — Correctness: **BLOCK**

**B1 (BLOCK). Stripe golden incremental resume is broken end-to-end through the real app layer.**
The app persists a stream cursor as the stringified record cursor field
(`internal/app/sync_modes.go:163` `recordCursor` → `toComparableString`, `internal/app/app.go:507`).
Live Stripe returns `created` as Unix-seconds → persisted cursor `"1700000100"`. On resume, the
engine's `formatParam(value, "unix_seconds")` (`internal/connectors/engine/read.go:329-334`)
hard-requires RFC3339 input → `time.Parse` fails → **the second incremental sync errors out**.
Legacy forwards the state cursor verbatim (`stripe/stripe.go` incrementalLowerBound), so this is a
real regression the parity suite masks: `TestParityStripe_IncrementalCreatedGTEFromState`
hand-feeds the engine an RFC3339 cursor that no production code path ever produces (nothing
converts `created` to RFC3339 — the docs/test-comment claim that "the engine persists RFC3339" is
false). This is the exact incremental pattern ~77 fan-out agents will copy for every
unix-timestamp API.
Fix: `formatParam` passes a digits-only value through verbatim for `unix_seconds` (matching legacy
semantics); add a parity/app-level round-trip test (read → persist cursor from records → resume).

**B2 (BLOCK). Conformance `cursor_advances` cannot see numeric cursors, and the golden fixtures
were falsified to compensate.** `conformance/dynamic.go:246-249` recognizes a cursor value only via
a Go `string` type assertion — a `json.Number` cursor (the common real-world case) hard-fails
"no cursor value observed". Consequence: `defs/stripe/fixtures/streams/**` commit `created` as
RFC3339 strings (not Stripe's real numeric wire shape) and `schemas/customers.json` was widened to
`["integer","string"]`, and `docs/migration/conventions.md` §4 **institutionalizes** this
("RFC3339 string cursors in conformance fixtures — a documented, deliberate convention"). That
directly contradicts §4's own first rule ("recorded-real-shape, sanitized") and means the dynamic
conformance suite for the flagship golden validates against data the live API can never produce —
artifact-exists-but-substance-is-bent, the exact failure mode this review is meant to catch.
Fix with B1: handle `json.Number`/numeric cursors in `checkCursorAdvances`, restore real wire
shapes in fixtures, delete the conventions.md §4 "RFC3339 fixture cursor" convention and deviation
ledger entry 2.

**B3 (BLOCK). V-21 phase gate incomplete + committed build artifact.** An 11 MB compiled binary
`inventorygen` is committed at the repo root (commit bfad5e5) — the planned path-guard
(`git status --porcelain` limited to planned paths) demonstrably did not catch it. Corroborating:
`RUN-STATE.json` says `status: blocked_missing_artifact`, `coveragePassed: false`; `SUMMARY.md` is
"TBD"; `TDD-GATE.json` claims `passed: true` with **empty** `tasks`/`behaviorTasks` arrays;
`VERIFICATION.md` lists configured commands but records no run results. The Wave H V-21 gate the
plan requires (PLAN.md:289-299) has not actually been completed/recorded.
Fix: remove + gitignore the binary; run and record V-21 (SUMMARY, VERIFICATION results,
TDD-GATE task rows, RUN-STATE).

Other correctness findings (FLAG):
- **F1 (high).** Read path never interpolates `stream.Path` or `check.path`
  (`engine/read.go:123`, `read.go:557`) — a templated path (`/repos/{{ config.owner }}/...`) is
  sent literally. Write path does interpolate (`write.go:202`). `connectorgen validate`
  ResolveChecks stream paths (validate.go:267), so such a bundle validates then breaks at runtime.
  The three goldens don't need it; a github-style wave1 pilot needs it day one — treat as a
  pre-known ENGINE_GAP to close first-thing in pilot (or promote to BLOCK if github is the pilot).
- **F3.** `lastRecordCursor` hardcodes the records path `"data"` (`paginate.go:158`) and requires
  the last-record field to be a non-empty Go `string` (`paginate.go:167-171`) — an API whose list
  key isn't `data`, or whose ids are numeric, silently stops after page 1 (data truncation, no
  error). Derive from `records.path` and stringify numbers.
- **F9.** Interpolator: multiple piped filters silently ignored (`interpolate.go:83-99` uses only
  `parts[1]`); `ResolveCheck` validates neither filter names nor auth `username/password/
  token_url/client_id/client_secret/scopes` templates (`cmd/connectorgen/validate.go:261-265`) —
  typos pass validate, fail at runtime.
- Error-classification by substring matching (`read.go:266` `isUnresolvedKey`, `read.go:441`
  `isUnresolvedRecordPath`, `validate.go` loadErrorFinding) — brittle; use typed sentinel errors.
- Stale comment: `parity_searxng_test.go:176` references `TestParityStripe_MaxPagesStopEngineGap`,
  which no longer exists.

## Dimension 2 — Security: **FLAG**

Good: CRLF/header-injection guard runs on the **pre-filter** value of every interpolation
(`interpolate.go:90`); `urlencode` is default-on for path insertions with `%`→`%25` double-encode
guard; DryRun previews redact secrets before interpolation (`write.go:112-117`); engine errors
pass through `safety.RedactErrorText` with operator hints exempt (`errors.go:51-57`); fixture
secret scans are real (validate.go:73 pattern, certify base64/URL-encoded scan); postgres golden
has genuine identifier validation + bound parameters + host validation. `selectAuth` errors on
no-match (`auth.go:45`), so absent-key-falsy `when` cannot *silently* disable auth via auth specs.

- **F2.** `next_url` SSRF guard compares host only (`paginate.go:210-215`): an `https`→`http`
  downgrade to the same host passes the guard — credentials/auth headers sent over cleartext on an
  attacker-influenced body value. An unparseable next URL yields host `""` and skips the guard
  entirely (`paginate.go:233-239`). Enforce same-scheme (or https-upgrade-only) and reject
  unparseable URLs.
- **F4.** `resolveHeaders` swallows *any* unresolved-key error (`read.go:253`): a header like
  `Authorization: Bearer {{ secrets.token }}` with the secret absent is **silently omitted** → the
  request goes out unauthenticated instead of failing. Needed for the Stripe-Account pattern, but
  the tolerance should be config-scoped (or headers referencing `secrets.*` should hard-error);
  conventions.md should forbid auth-bearing declared headers in favor of `auth` specs.
- **F9b.** `urlencodeSegment` leaves `.` unescaped, so a record/config value of `..` survives into
  a path segment (`/customers/..`) — single-segment traversal is not blocked (slashes are; `%2e%2e`
  is). Consider rejecting `.`/`..` segments outright in `InterpolatePath`.

## Dimension 3 — Template quality for replication: **BLOCK** (via B1/B2) + FLAGs

The goldens are largely exemplary (postgres especially), but ship anti-patterns that would
replicate 557×:
- The falsified-fixture convention (B2) is *taught* by conventions.md §4 — must be removed.
- **F6.** Dead/inert declared config: searxng `spec.json` declares 9 keys the bundle never wires
  (`api_key`, `subreddit`, `categories`, `engines`, `language`, `time_range`, `safesearch`,
  `page_size`, `max_pages`) — `api_key` is an "optional Bearer" that is **never applied** (an
  instance behind an auth proxy silently 401s), and conventions.md §2 endorses declaring such
  fields. Stripe: `metadata.json` `rate_limit.strategy: "token_bucket"` is an unknown key silently
  ignored, and `requests_per_minute: 100` lives in metadata.json which the read path **never
  consults** (`read.go:91` uses `b.HTTP.RateLimit` only) → stripe has *no* enforced rate limit
  while appearing to declare one; base pagination carries dead `limit_param`/`page_size` for a
  cursor paginator (deviation #3 keeps them deliberately — dead config in a few-shot example).
  Rule to add: connectorgen should flag spec keys/pagination fields not consumed by anything.
- **F7.** conventions.md §5's meta-rule ("ACCEPTABLE iff it never changes the emitted record DATA")
  is violated by its own entries 4 (searxng `engines`: array vs legacy comma-joined string) and 6
  (legacy `stream` marker field dropped) — both change emitted record shape for identical inputs
  (warehouse schema drift on cutover). Needs an explicit policy decision (human-gate material for
  wave6 cutover) plus engine features (array-join filter; static-literal computed fields).
- Accuracy spot-check (10 claims): 8/10 accurate against code (urlencode default, absent-key
  hard-error scope, when-grammar usage, wholesale pagination override, MaxPages pre-request check,
  form-body sorted/empty-omitted, delete missing_ok semantics, Tier-3 loader tolerance). Two slips:
  §2 attributes `pk_fields_exist`/`cursor_fields_exist` to `connectorgen validate` (those are
  conformance names; validate's are `primary_key_missing`/`cursor_field_missing`), and §3 calls
  cursor `token_path`+`last_record_field` conflicts a "load-time error" (it's read-time —
  `newPaginator` runs per-read; validate doesn't check pagination specs at all).

## Dimension 4 — API design: **FLAG**

- **F5.** `Definition.Spec` is a lossy reconstruction (`engine/connector.go:293-315`): every
  property becomes `{"type":"string"}` (+`x-secret`); types, enums, defaults, required,
  descriptions all dropped — postgres's port/sslmode constraints vanish. The loader has the raw
  `spec.json` bytes in hand (`bundle.go:394-407`) and discards them. Wave6 (Definition-driven
  config UX/validation) will be wrong. Fix: retain raw bytes on `Bundle`, serve verbatim.
- **F8.** `AuthHook.Authenticator` is invoked with `context.Background()` (`auth.go:150`) —
  `selectAuth` takes no ctx. A github_app JWT→installation-token exchange (network call, wave1)
  won't honor cancellation/deadlines. Thread ctx through `newRuntime`/`selectAuth` before pilots
  write AuthHooks. Otherwise the 5-hook seam looks adequate for wave1 (github_app via AuthHook,
  fan-out via StreamHook, compound writes via WriteHook); gmail-style 3-legged OAuth token
  *refresh* is expressible as an AuthHook but token acquisition/storage is out of engine scope —
  confirm the credentials layer covers it before selecting gmail as a pilot.
- Manifest synthesis maps `RequiredFields=path_fields`/`OptionalFields=body_fields`
  (`connector.go:210-217`) — an approximation of legacy manifests (parity only asserts action
  names); acceptable, note for wave6.
- Exported surface is otherwise tight; `nextURL.BaseHost`-must-be-set-by-caller
  (`paginate.go:23`) is a fragile implicit contract — constructor injection would be safer.

## Dimension 5 — Test integrity: **PASS** (with B2 carve-out)

- Parity tests genuinely drive the **legacy connectors live** against the same httptest servers as
  the engine (`parity_stripe_test.go`, `parity_searxng_test.go`, `native/postgres/parity_test.go`)
  — not copies of expectations. Legacy-side sanity assertions ("test fixture bug" fatals) guard
  against dead comparisons.
- The flipped `TestParitySearxng_MaxPagesStop` is a legitimate **strengthening**: commit 97dc754
  replaced the gap-documenting `TestParitySearxng_MaxPagesStopEngineGap` (asserted engHits > 1)
  with a hard-cap parity assertion (engHits == 1), alongside the actual engine fix and new RED
  evidence (traces/waveF-repair-ledger.md). Not a weakening.
- `withSearxngUnboundedMaxPages` is legitimate isolation: used only in the short-page-stop test,
  with legacy symmetrically fed `max_pages: "all"`, and the cap has its own dedicated test.
- TDD ledger evidence is real (per-wave RED/GREEN transcripts in traces/), though TDD-GATE.json's
  empty arrays (B3) undercut the machine-readable gate.
- Carve-out: stripe conformance-fixture realism (B2) — dynamic conformance passes against
  synthetic wire shapes.

---

## Go / No-Go

**NO-GO** for starting wave1-pilot until:
1. B1 — `formatParam` unix_seconds digits-passthrough + app-level cursor round-trip test.
2. B2 — numeric-cursor support in `cursor_advances`; real-wire-shape stripe fixtures; delete the
   conventions.md §4 fixture-cursor convention (+ deviation ledger entry 2 rewrite).
3. B3 — remove/gitignore the `inventorygen` binary; complete and record the V-21 gate
   (SUMMARY.md, VERIFICATION.md results, TDD-GATE.json task rows, RUN-STATE.json).

FLAGs F1–F10 may be fixed during pilot; F1 (stream-path interpolation) and F8 (AuthHook ctx) must
land before any pilot connector that needs templated paths or a custom auth hook. F7 (record-shape
deviations vs the §5 meta-rule) requires an explicit human decision before any cutover wave.

Handoff: B1/B2 → backend+tester (engine/read.go, conformance/dynamic.go, defs/stripe fixtures,
docs/migration/conventions.md); B3 → coordinator/verifier (V-21 re-run).
