# Engine hardening — unknown-key defense-in-depth (TDD ledger)

Trigger: the re-review found `internal/connectors/defs/rentcast` declaring a `"base.check.query"`
block (`{"limit":"1","offset":"0"}`) that `engine/bundle.go`'s `RequestSpec` (used for `base.check`)
silently ignores — it only has `Method`/`Path` fields. `json.Unmarshal` drops unknown map keys, and
every nested sub-object in the meta-schemas (`streams.schema.json` etc.) was a bare
`{"type":"object"}` with no `additionalProperties:false`, so this invented mechanism passed every
gate (meta-schema validate, `connectorgen validate`, `go build`) while doing nothing at runtime —
`Check()` never sends a query at all.

Scope: `internal/connectors/engine/**` (+tests), `cmd/connectorgen/**` (corpus). No bundle
directories under `internal/connectors/defs/` were edited. No commits made.

---

## Item 1 — meta-schema tightening (additionalProperties:false + explicit allowlists)

### Design

`internal/connectors/engine/schema.go`'s compiled-schema engine (a draft-07 SUBSET, not a full
JSON-Schema implementation — no `$ref`/`definitions` support) already implements
`additionalProperties:false` correctly (`validateObject`), but only rejects keys not present in a
node's `properties` map. Every nested sub-object in the 5 meta-schema files
(`internal/connectors/engine/schema/*.schema.json`) was declared as a bare `{"type":"object"}` with
NO `properties` map at all, so `additionalProperties:false` at that level would have (if set)
rejected EVERY key, not just unknown ones — the fix requires an explicit property allowlist for
every structured sub-object, not just flipping a flag.

Tightened (allowlisted + `additionalProperties:false`):
- `streams.schema.json`: `base` (url/user_agent/headers/auth/pagination/check/error_map/rate_limit),
  `base.auth[]` items, `base.pagination`, `base.check`, `base.error_map[]` items, `base.rate_limit`,
  `streams[]` items, `streams[].records` (+ `.filter`), `streams[].pagination`,
  `streams[].incremental`, `streams[].conformance` (already had it).
- `writes.schema.json`: `actions[]` items (top level already had it), `actions[].delete`.
- `api_surface.schema.json`: `endpoints[]` items, `.covered_by`, `.excluded` (already had it).
- `spec.schema.json`: the top-level spec.json object itself (its fixed keys: `$schema`/`title`/
  `type`/`required`/`properties`) — verified no bundle's spec.json uses any other top-level key.
- `metadata.schema.json`: already fully allowlisted at every level; untouched.

Deliberately LEFT OPEN (free-form maps / user-defined JSON-Schema documents, not a fixed dialect
surface): `base.headers`, `stream.query` (string-or-object dialect, arbitrary param names),
`stream.body`, `stream.computed_fields`, `records.filter.field_equals`, `writes.record_schema`,
`spec.json`'s `properties` (nested JSON-Schema property definitions — legitimately carry `x-secret`/
`format`/`enum`/etc. per bundle author), and all of `schemas/*.json`/`spec.json` beyond its own
fixed top-level keys (these are user draft-07 schemas, explicitly out of scope per the dispatch).

### Verify

```
$ go build ./...
(clean — meta-schemas still compile; no $ref/definitions leaked in)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 411 connector(s) checked, 150 finding(s)
```

150 bundles now fail with `[meta_schema] ... streams.json: /base/check/query: additional property
not allowed` (148 bundles) or `/base/query: additional property not allowed` (7 bundles; 5 overlap
with the check.query set for a combined 150 unique connectors — see "Newly exposed" below). Zero
NEW findings from the writes.json/api_surface.json/spec.json tightening (confirmed by diffing the
finding set before/after those 3 files' changes) — the entire 150-bundle blast radius is exactly the
one `streams.json` `base.query`/`base.check.query` defect class the ledger's trigger named.

`testdata/valid`/`testdata/valid-extra`'s golden bundles still pass with zero findings (confirms the
allowlists match the real, intentional Go struct shapes — no false positives on legitimate fields).

---

## Item 2 — loader strict-decode (independent of the meta-schema layer)

### Design

Added `engine/bundle.go`'s `strictDecode(raw []byte, dst any) error` — wraps
`json.NewDecoder(...).DisallowUnknownFields()`. Wired into `loadMetadata` (metadata.json),
`loadStreams` (streams.json), `loadWrites` (writes.json), and `loadAPISurface` (api_surface.json),
replacing their plain `json.Unmarshal` calls. `loadSpec` (spec.json) deliberately UNCHANGED — it
compiles through `CompileSchema`'s own custom schema-node representation, not a fixed Go struct, and
per the dispatch spec.json/schemas/* must not be restricted beyond the meta-level.

This is deliberate defense-in-depth, not redundant with Item 1: a future edit to the meta-schema
files themselves (e.g. someone loosening `additionalProperties:false` back to `true` while adding a
new field) would silently reopen the exact hole this ledger closes; the loader's independent
strict-decode does not depend on the meta-schema files staying correct. `DisallowUnknownFields`
only rejects keys unmatched by a STRUCT field — every free-form map field (`Headers`, `Query`,
`Body`, `ComputedFields`, `FieldEquals`, `RecordSchema`) stays open, verified by a dedicated
"still accepts free-form map keys" test.

### RED (engine/bundle_test.go)

```
$ go test ./internal/connectors/engine/... -run 'TestBundleLoadRejectsUnknown|TestBundleLoadStillAcceptsFreeFormMapKeys' -v
(against meta-schemas reverted to their pre-Item-1 shape, loader unpatched)
--- FAIL: TestBundleLoadRejectsUnknownBaseLevelKey        (Load returned nil error)
--- FAIL: TestBundleLoadRejectsUnknownBaseCheckQueryKey    (Load returned nil error — the exact rentcast shape)
--- FAIL: TestBundleLoadRejectsUnknownStreamLevelKey       (Load returned nil error)
--- FAIL: TestBundleLoadRejectsUnknownAuthCandidateKey     (Load returned nil error)
--- FAIL: TestBundleLoadRejectsUnknownWritesActionKey      (Load returned nil error)
--- PASS: TestBundleLoadRejectsUnknownMetadataTopLevelKey  (metadata.json top level was already additionalProperties:false)
--- PASS: TestBundleLoadStillAcceptsFreeFormMapKeys        (trivially true pre-fix too)
FAIL
```

### GREEN

```
$ go test ./internal/connectors/engine/... -run 'TestBundleLoadRejectsUnknown|TestBundleLoadStillAcceptsFreeFormMapKeys' -v
--- PASS (all 8 subtests, meta-schemas restored + strictDecode wired)

$ go test ./internal/connectors/engine/... -v
ok   # full package, zero regressions
```

Also added `TestBundleLoadRejectsUnknownAPISurfaceEndpointKey` (api_surface.json's newly-tightened
endpoint-item allowlist), same RED/GREEN pattern.

---

## Item 3 — `cmd/connectorgen` invalid-corpus seed case

New fixture: `cmd/connectorgen/testdata/invalid/unknown-base-key/` — reproduces the EXACT rentcast
shape (`base.check.query`), added to `TestValidate_RejectsSeededInvalidBundles`'s table under rule
`ruleMetaSchema` (the loader's `engine.Load` error is already classified into `ruleMetaSchema` by
`loadErrorFinding`'s file-name substring match — no new classification code needed).

```
$ go test ./cmd/connectorgen/... -run TestValidate_RejectsSeededInvalidBundles -v
--- PASS (all cases including unknown-base-key)

$ go test ./cmd/connectorgen/... -v
ok   # full package, zero regressions (TestValidate_AcceptsGoodBundle's golden corpus unaffected)
```

---

## Item 4 — `LoadAll` fleet-wide resilience (unplanned, discovered during self-verify)

### The problem

`engine.LoadAll` (production bundle-discovery path; also used by `TestConformance`,
`TestBundleLoadAllDefsFS`, and the stripe/searxng/chargebee/github/monday golden parity suites) was
fail-fast: the FIRST bundle that failed to `Load` aborted the entire batch, returning `nil` bundles
and a single-bundle error. With Item 1+2 correctly exposing 150 pre-existing, out-of-scope-to-fix
defects across ~400 independently-authored bundles, this meant `LoadAll(defs.FS)` started returning
zero bundles and a useless "first alphabetical failure" error — hiding the ~260 unaffected bundles
from every test (and from production discovery) entirely. `cmd/connectorgen validate` already
avoids exactly this failure mode (`validateBundleDir` isolates one bundle's load error into a
`Finding` and keeps going); `LoadAll` did not.

This is a latent robustness gap Item 1+2 newly exposed, not a new defect introduced by them — but
leaving it as fail-fast would make the mandated self-verify command (`go test
./internal/connectors/engine ... ./internal/connectors/conformance`) report a false, uninformative
signal (zero bundles loadable) instead of the true, complete one (260 fine, 150 named and why).

### Fix

`LoadAll` now attempts every bundle directory unconditionally and returns every bundle that DID load
alongside a new typed `*LoadAllError` (`Failures []BundleLoadFailure{Name, Err}`) whenever one or
more failed — `errors.As`-compatible, preserves per-bundle detail instead of a flattened string.
Callers that must treat any failure as fatal still get a non-nil `error`; callers that want the
loadable subset (this repo's own fleet-wide tests) can proceed with the returned bundles.

Consumers updated to use the new resilient contract (test-only changes, no assertion-strength
loosening — same or stronger checks, just no longer coupled to the ENTIRE fleet's cleanliness):
- `internal/connectors/engine/bundle_test.go`: new `TestBundleLoadAllOneBadBundleDoesNotHideTheRest`
  (RED-verified against the pre-fix fail-fast `LoadAll`); `TestBundleLoadAllDefsFS` updated to still
  hard-require the stripe/postgres goldens load cleanly, while tolerating (and shape-checking) the
  now-expected `*LoadAllError` for the 150 known bundles.
- `internal/connectors/conformance/conformance_test.go`: `TestConformance` now emits one FAILING
  subtest per bundle that failed to `Load` (named after the bundle, so `-run
  'TestConformance/<name>'` still isolates it) instead of a single `t.Fatalf` that previously
  produced ZERO subtests and zero information about any other bundle. This is a strengthening: more
  information surfaces per run, not less.
- `internal/connectors/engine/parity_stripe_test.go`, `parity_searxng_test.go`,
  `internal/connectors/paritytest/{chargebee,github,monday}/parity_test.go`: their single-bundle
  `load<Name>Bundle` helpers no longer treat `LoadAll`'s non-nil error as fatal — they only fail if
  their OWN named bundle is missing from the returned set. This brings them in line with
  `paritytest/calendly`/`paritytest/sentry`'s ALREADY-established pattern (those two use
  `engine.Load(defs.FS, "<name>")` directly, specifically documented in-repo as avoiding "LoadAll
  fails hard on the first sibling's defect" — SPEC.md §6, pre-dating this ledger).

### RED/GREEN

```
$ go test ./internal/connectors/engine/... -run TestBundleLoadAllOneBadBundleDoesNotHideTheRest -v
(pre-fix, fail-fast LoadAll)
--- FAIL: LoadAll bundles = map[], want acme and beta still returned despite broken's failure

(post-fix)
--- PASS

$ go test ./internal/connectors/engine ./cmd/connectorgen -count=1
ok   ok   # both fully green

$ go test ./internal/connectors/conformance -count=1
FAIL  # 150 named subtests fail (the 150 bundles below) — EXPECTED, not a regression; every OTHER
      # bundle's subtest passes; zero collateral/unexplained failures.

$ go test ./internal/connectors/paritytest/... -count=1
ok (all 10 packages, including chargebee/github/monday)

$ go build ./... && make lint
BUILD OK; 0 issues.
```

---

## Final self-verify

```
$ go build ./...
(clean)

$ go test ./internal/connectors/engine ./cmd/connectorgen -count=1
ok   ok

$ go test ./internal/connectors/conformance -count=1
FAIL — exactly 150 named TestConformance/<bundle> subtests fail, 1:1 with connectorgen validate's
150 findings; every other bundle (~260) passes. This is the CORRECT, intended, complete consequence
of closing the hole: these 150 bundles have a real defect (an invented "base.check.query"/
"base.query" mechanism the engine silently no-ops) that was invisible before this hardening and is
now loudly, individually surfaced. Repairing them is explicitly out of scope for this dispatch.

$ make lint
0 issues.

$ go vet ./...
(clean)

$ gofmt -l <all touched files>
(clean)
```

## Newly exposed (150 connectors) — follow-up required, NOT fixed here

148 via `base.check.query` (a `RequestSpec` has no `query` field — `Check()` never sends one):
aha, aircall, apptivo, assemblyai, bigmailer, bluetally, boldsign, breezometer, brex, buildkite,
callrail, capsule-crm, cin7, circa, codefresh, dbt, defillama, deputy, e-conomic, easypost,
employment-hero, encharge, eventzilla, financial-modelling, finnhub, flexmail, flexport, float,
freshbooks, freshchat, freshdesk, getgist, getlago, giphy, gong, gorgias, granola, greenhouse,
hellobaton, high-level, hugging-face-datasets, humanitix, huntr, inflowinventory, insightly,
intruder, invoiced, invoiceninja, jobnimbus, katana, kisi, kissmetrics, klarna, launchdarkly,
leadfeeder, lever-hiring, lightspeed-retail, lob, mailchimp, mailerlite, mailersend, newsdata-io,
ninjaone-rmm, nylas, omnisend, oncehub, openaq, openfda, outreach, paddle, pagerduty, pandadoc,
paperform, papersign, pardot, partnerize, partnerstack, payfit, paystack, pendo, perigon, picqer,
pipedrive, pipeliner, pivotal-tracker, piwik, planhat, pokeapi, polygon-stock-api, poplar,
postmarkapp, productboard, productive, pylon, qonto, qualaroo, railz, rd-station-marketing,
recreation, recruitee, recurly, reddit, referralhero, **rentcast**, repairshopr, rollbar,
salesflare, sendgrid, shippo, shipstation, shopwired, signnow, simfin, simplecast, split-io,
spotify-ads, spotlercrm, squarespace, statsig, statuspage, stockdata, survey-sparrow, survicate,
taboola, tempo, tickettailor, tinyemail, trustpilot, tvmaze-schedule, twelve-data,
twilio-taskrouter, twitter, typeform, unleash, woocommerce, wordpress, workday, workday-rest,
workflowmax, workramp, wrike, zonka-feedback, zoom.

7 via bare `base.query` (an `HTTPBase` has no `Query` field at all — a per-bundle "shared query
params applied to every stream" mechanism that never existed): reply-io, retailexpress-by-maropost,
retently, revenuecat, revolut-merchant, ringcentral, you-need-a-budget-ynab.

Note: `internal/connectors/defs/rentcast`'s OWN `base.query` (the block this ledger's trigger
originally named) was already found repaired to per-stream `query` blocks in the working tree
before this dispatch started (uncommitted, pre-existing change, not made by this dispatch) — but its
`base.check.query` (`{"limit":"1","offset":"0"}`) is a SEPARATE, still-unrepaired instance of the
identical invented-mechanism class, and is included in the 148-bundle list above.

Suggested follow-up shape per case (not applied here): either (a) delete the dead
`query`/`check.query` block entirely if the fixtures/parity tests prove it was never load-bearing,
or (b) if `Check()` genuinely needs those query params for the real API, extend `RequestSpec` with
an engine-level `Query map[string]QueryParam` field (mirroring `StreamSpec.Query`'s existing
string-or-object dialect) — an ENGINE_GAP-shaped change, out of scope here since it changes runtime
behavior for `base.check`, not just validation.
