# p2-vitally-ledger — wave1-pilot P-2 (vitally pilot migration)

Scope: `internal/connectors/defs/vitally/**`, `internal/connectors/paritytest/vitally/**` only.
No hooks (Tier 1, no Tier-2 trigger applies). No `git commit` performed by this task. Legacy
`internal/connectors/vitally/{vitally.go,vitally_test.go}` read in full, untouched.

## Legacy summary (read-only reference)

`internal/connectors/vitally/vitally.go` (188 loc): one stream `accounts`
(`GET resources/accounts`), auth via `connsdk.APIKeyHeader("Authorization", auth, "")` where
`auth` is the secret `basic_auth_header` **verbatim** (vitally.go:100-104) — the secret already
contains the complete `Authorization` header value, not a bare token/username-password pair.
Optional `status` config value appended as a query param only when non-empty (vitally.go:72-74).
Records extracted from top-level `results`, mapped field-for-field to `{id, name, traits}`
(vitally.go:79-86). No pagination, no incremental cursor. `Write` always returns
`connectors.ErrUnsupportedOperation` (read-only). Legacy also has a `mode=fixture` short-circuit
(vitally.go:64-65,107-116) — a test affordance, not part of the live record shape (SPEC.md §5.1's
xkcd note documents the same principle generally).

## Red-first protocol

1. Wrote `internal/connectors/paritytest/vitally/parity_test.go` FIRST (plus a one-line `doc.go`
   per SPEC.md §6's per-connector-package decision), calling
   `engine.Load(defs.FS, "vitally")` before any `defs/vitally/` bundle existed.

### RED evidence

```
$ go test ./internal/connectors/paritytest/vitally/... -v
# polymetrics.ai/internal/connectors/paritytest/vitally
internal/connectors/defs/defs.go:14:12: pattern all:*: cannot embed directory calendly: contains no embeddable files
FAIL	polymetrics.ai/internal/connectors/paritytest/vitally [setup failed]
FAIL
```

Note on this RED evidence: at RED-capture time, DW-1 had 9 other pilot agents writing
concurrently into sibling `internal/connectors/defs/<name>/` directories (per PLAN.md's "all 10
dispatch simultaneously"); `defs.FS`'s `//go:embed all:*` fails the WHOLE tree's build the instant
any sibling directory is transiently incomplete (an empty dir, or one missing a required file
mid-write), which is what the calendly clause captures above — not a defect in this agent's own
work. The SUBSTANTIVE red (the actual thing this task's TDD is proving) is unambiguous
independent of that transient noise: `internal/connectors/defs/vitally/` did not exist at all at
this point, so `engine.Load(defs.FS, "vitally")` could never have resolved a bundle regardless of
sibling state — confirmed directly via an isolated `engine.Load(os.DirFS(defsRoot), "vitally")`
smoke check before any vitally bundle file existed (same "missing bundle" failure mode, decoupled
from the shared embed). Bundle authored immediately after; GREEN captured below once the bundle
existed and (separately, opportunistically) once the shared tree was momentarily stable enough for
the full embed-based suite to run end to end.

## GREEN — bundle authored

Files added under `internal/connectors/defs/vitally/`: `metadata.json`, `spec.json`,
`streams.json`, `api_surface.json`, `docs.md`, `schemas/accounts.json`, `fixtures/check.json`,
`fixtures/streams/accounts/page_1.json` (8 files; no `writes.json` — read-only, matching legacy).

Key authoring decisions:

- **Auth**: `streams.json` `base.auth` = `[{"mode":"api_key_header","header":"Authorization",
  "value":"{{ secrets.basic_auth_header }}","prefix":""}]`. Chose `api_key_header` over the
  engine's `basic` mode deliberately: `basic` mode (`connsdk.Basic`) base64-encodes a
  `username:password` pair at request time, which would require decomposing
  `basic_auth_header` back into constituent parts the connector never receives — the secret
  already IS the complete, pre-built header value (SPEC.md §5.2's explicit call-out: "read the
  legacy value construction and express as engine `basic` mode or `api_key_header` + `base64`
  filter, whichever reproduces the exact header byte-for-byte"). `api_key_header` with an empty
  `prefix` is the byte-exact match, verified by
  `TestParityVitally_AuthorizationHeaderByteExact`.
- **Records**: `records.path: "results"`; schema-as-projection with no `computed_fields` needed —
  legacy's `{id, name, traits}` mapping is already an exact field-for-field match with no renames
  or nested extraction (`traits` passed through as an opaque object bag on both sides).
- **No pagination, no incremental**: matches legacy's single-request, full-refresh-only behavior;
  `x-primary-key: ["id"]`, no `x-cursor-field`.
- **`status` query param — NOT wired (documented known limit, not a defect)**: legacy only sends
  `status` when non-empty; conventions.md §3 states `stream.Query` templating has **no
  absent-key-falsy tolerance** — every `{{ }}` reference in a `query` map value is resolved
  unconditionally, so `"status": "{{ config.status }}"` would hard-error whenever `status` is
  unset (the common case). Identical to searxng's already-accepted optional-passthrough-filter
  limitation (searxng's own `docs.md` "Known limits"; conventions.md §3). `status` is **not
  declared** in `spec.json` at all, per F6 ("a declared-but-unwireable key is worse than an
  absent one"). Documented in `defs/vitally/docs.md` "Known limits" and asserted by
  `TestParityVitally_NoStatusParamSentWhenUnset` (proves the accepted base case: neither side
  sends a status param when unset — the ONLY case this bundle's spec allows).

## Parity-deviation ledger entry (candidate for conventions.md, feeds P-12)

| connector | description | verdict |
|---|---|---|
| vitally | The optional `status` query filter (legacy: appended only when non-empty, vitally.go:72-74) is not modeled — the engine's `stream.Query` templating has no absent-key-falsy tolerance, so an unconditional reference would hard-error on the common no-status-configured path. `status` is not declared in `spec.json` at all (F6: declared-but-unwireable is worse than absent). Base case (no status configured) is byte-identical on both sides; status-configured behavior is out of scope, documented in docs.md "Known limits", not silently approximated. | ACCEPTABLE (documented scope narrowing, same shape as the already-ledgered searxng subreddit-narrowing entry) |

No `ENGINE_GAP` filed: this is the same accepted shape already ledgered for searxng (item 7,
conventions.md §5), not a new blocker.

## Self-verify (final, all green)

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
monday: metadata.json: [surface_fail_first_run] capabilities.write is false but api_surface.json has a non-excluded POST/PUT/PATCH/DELETE endpoint
connectorgen validate: 13 connector(s) checked, 1 finding(s)
```
(The 1 finding belongs to the sibling `monday` connector — a different pilot agent's
work-in-progress, out of this task's writable scope. Isolated re-check of vitally alone —
`go run ./cmd/connectorgen validate <tmp-dir-containing-only-defs/vitally>` — reports
`1 connector(s) checked, 0 findings`.)

```
$ go build ./internal/connectors/... && go vet ./internal/connectors/...
(clean, no output)

$ go test ./internal/connectors/conformance -run 'TestConformance/vitally' -v
=== RUN   TestConformance
=== RUN   TestConformance/vitally
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/vitally (0.00s)
PASS

$ go test ./internal/connectors/paritytest/vitally -v
=== RUN   TestParityVitally_AccountsStreamRecords
--- PASS: TestParityVitally_AccountsStreamRecords (0.00s)
=== RUN   TestParityVitally_NoStatusParamSentWhenUnset
--- PASS: TestParityVitally_NoStatusParamSentWhenUnset (0.00s)
=== RUN   TestParityVitally_AuthorizationHeaderByteExact
--- PASS: TestParityVitally_AuthorizationHeaderByteExact (0.00s)
=== RUN   TestParityVitally_NonSuccessStatusErrorsOnBothSides
--- PASS: TestParityVitally_NonSuccessStatusErrorsOnBothSides (0.00s)
=== RUN   TestParityVitally_WriteUnsupportedOnBothSides
--- PASS: TestParityVitally_WriteUnsupportedOnBothSides (0.00s)
=== RUN   TestParityVitally_BundleLoadsWithSingleAccountsStream
--- PASS: TestParityVitally_BundleLoadsWithSingleAccountsStream (0.00s)
=== RUN   TestParityVitally_CheckRequiresAuthSecretOnBothSides
--- PASS: TestParityVitally_CheckRequiresAuthSecretOnBothSides (0.00s)
PASS
```

## Path guard (this task's writable set only)

```
$ git status --porcelain | grep vitally
?? internal/connectors/defs/vitally/
```
(`internal/connectors/paritytest/` shows as a bare untracked directory in `git status` since
sibling pilot agents also write under it; `find internal/connectors/{defs,paritytest}/vitally
-type f` confirms exactly 10 files, all under this task's two exclusive dirs, nothing elsewhere
touched.)

## Blockers

None. Status: **migrated**.
