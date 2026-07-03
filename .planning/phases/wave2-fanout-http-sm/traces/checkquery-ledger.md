# Engine dialect addition — base.check.query (TDD ledger)

Trigger: the strict-key hardening (hardening-ledger.md) closed the "invented mechanism" hole by
tightening streams.schema.json's `base.check` to an explicit property allowlist, which correctly
started REJECTING 148 bundles' `base.check.query` blocks as an unknown key. But those 148 bundles'
`check.query` is not actually invented nonsense the way a bare `base.query` is — legacy connectors
genuinely send query params on their check/probe request (e.g. `limit=1`/`per_page=1` probes), and
`RequestSpec` (engine/bundle.go, Method/Path only) simply had nowhere to put that value. This ledger
implements the follow-up hardening-ledger.md itself named as out of scope: give `RequestSpec` a real
`Query` field, wire `Check()` to send it, and validate/replay-test it like any other query dialect.

Scope: `internal/connectors/engine/**` (+tests), `cmd/connectorgen/**` (+tests, +corpus case). No
bundle directories under `internal/connectors/defs/` were edited. No commits made.

---

## Item 1 — `RequestSpec.Query`

### Design

`RequestSpec.Query` is `map[string]QueryParam` — the SAME type `StreamSpec.Query` already uses, not
a plain `map[string]string`. This mirrors hardening-ledger.md's own suggested follow-up shape
verbatim ("extend RequestSpec with an engine-level Query map[string]QueryParam field, mirroring
StreamSpec.Query's existing string-or-object dialect") and gives check.query the same
`omit_when_absent`/`default` escape hatches stream.Query already has, for free, with no new dialect
to learn. Verified all 148 real bundles' `base.check.query` entries are plain-string templates (no
object-form entries in the wild yet), so this is a pure superset — no existing bundle needs the
object form, but nothing stops one from using it.

`streams.schema.json`'s `base.check` gained a `"query": {"type": "object"}` property — identical
shape to how `stream.query` itself is declared at the meta-schema level (no per-key constraint
there either; the Go struct's `QueryParam.UnmarshalJSON` handles string-or-object).

### Verify

```
$ go build ./...
(clean)

$ go test ./internal/connectors/engine -run 'TestBundleLoadAcceptsBaseCheckQueryKey|TestBundleLoadParsesCheckQueryOptionalDialect|TestBundleLoadRejectsUnknownBaseLevelKey' -v
PASS (base.check.query now loads and round-trips into RequestSpec.Query; bare base.query, a
SEPARATE and still out-of-scope 7-bundle defect class, remains correctly rejected)
```

`TestBundleLoadRejectsUnknownBaseCheckQueryKey` (the seeded test that asserted `base.check.query` is
an unknown-key error) is now OBSOLETE by design — this dispatch's entire point is to make that key
legitimate. Replaced with `TestBundleLoadAcceptsBaseCheckQueryKey` (proves the exact rentcast shape
now loads cleanly and round-trips) and `TestBundleLoadParsesCheckQueryOptionalDialect` (proves the
object-form omit_when_absent/default dialect works identically to stream.Query's). No coverage was
deleted — the REJECTION behavior that test pinned no longer describes a real defect, so the test was
converted to pin the new, correct ACCEPTANCE behavior instead, at the same specificity.

---

## Item 2 — `Check()` builds and sends the query

### Design

Extracted `buildInitialQuery`'s per-entry resolution loop (bundle.go: hard-error on unresolved
config/secrets/incremental key for a plain-string entry; `OmitWhenAbsent`/`Default` escape hatches
for an object-form entry) into a shared `resolveQueryParams(params map[string]QueryParam, vars Vars)`
helper, used by BOTH `buildInitialQuery` (stream.Query) and a new `buildCheckQuery` (RequestSpec.
Query). This makes "same semantics as stream static query strings" true by construction, not by
convention — a future change to one path's resolution rules cannot silently drift from the other's
without both call sites' tests catching it.

`Check()` now calls `buildCheckQuery(b.HTTP.Check, cfg)` and passes the result as `Requester.Do`'s
query argument (previously hard-coded `nil`). A nil/absent `check.query` returns a nil url.Values —
`Check()` sends no query string at all, byte-for-byte identical to every bundle's behavior before
this dialect existed.

### Verify

```
$ go test ./internal/connectors/engine -run TestCheck -v
PASS (13 TestCheck* tests, including 6 new: query sent on request, template resolved against
config, unresolved key hard-errors, omit_when_absent drops the param, default is sent when absent,
no-query-declared sends no query string)

$ go test ./internal/connectors/engine -count=1
ok
```

---

## Item 3 — meta-schema (streams.schema.json)

Already covered by item 1 — `base.check.query` is now an allowlisted property, not a rejected one.

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 411 connector(s) checked, 7 finding(s)
```

Down from 150 findings (hardening-ledger.md's post-hardening count) to exactly 7 — the bare
`base.query` class (reply-io, retailexpress-by-maropost, retently, revenuecat, revolut-merchant,
ringcentral, you-need-a-budget-ynab), which remains out of scope for this dispatch (HTTPBase still
has no Query field; that is a genuinely different, still-invented mechanism). Zero collateral
findings — every one of the 148 `base.check.query` bundles now validates with zero errors, and no
new interpolation findings appeared (all 148 are plain-string templates against declared config, per
item 1's corpus check).

The `unknown-base-key` seeded-invalid corpus case (cmd/connectorgen/testdata/invalid/unknown-base-key)
tested exactly the shape this dispatch legitimizes, so it was repointed to the still-genuinely-invalid
bare `base.query` shape (the 7-bundle class above) rather than deleted — same rule
(`ruleMetaSchema`), same specificity, still a real defect.

---

## Item 4 — ResolveCheck / connectorgen validate

### Design

`cmd/connectorgen/validate.go`'s `checkInterpolations` now validates `b.HTTP.Check.Path` (previously
never checked at all — a pre-existing gap, fixed here as a natural side effect of touching this
function for Query) and every `b.HTTP.Check.Query` entry's `.Template`, via the same
`engine.ResolveCheck` call every other template in the bundle already goes through. An unresolvable
spec key in a check.query template is a `ruleInterpolationUnresolved` finding — the same rule
`auth-field-unknown-spec-key` already exercises for `base.auth`.

New seeded-invalid corpus case: `cmd/connectorgen/testdata/invalid/check-query-unknown-spec-key`
(`base.check.query.limit` templates `{{ config.nope_limit }}`, an undeclared spec key). Added to
`TestValidate_RejectsSeededInvalidBundles`'s table under `ruleInterpolationUnresolved`.

### Verify

```
$ go test ./cmd/connectorgen -run TestValidate_RejectsSeededInvalidBundles -v
PASS (21 seeded cases, all passing, including the new check-query-unknown-spec-key case)

$ go test ./cmd/connectorgen -count=1
ok
```

---

## Item 5 — conformance check_fixture: recorded request.query comparison

### Design (two false starts corrected before landing)

**First attempt** (rejected): reused `requestMatchesFixture`'s full method+path+query comparison
verbatim for `fixtures/check.json`'s (newly-added) `request` field. This broke ~37 real bundles that
had NOTHING to do with check.query — `internal/connectors/defs/github/fixtures/check.json` records
`request.path: "/repos/octocat/hello-world"` as human-readable documentation of what was captured,
while conformance's `runtimeConfigForEngine` synthesizes `"synthetic-conformance-value"` for every
config property at replay time — so a templated check.Path never equals the fixture's documentary
path. (Stream page fixtures follow the OPPOSITE convention — recording the synthetic placeholder
verbatim — specifically so THEIR path matching works; check.json fixtures predate any matching at
all and were never written to that convention.) Path/method matching was dropped entirely; only
query is compared, per the dispatch's own instructions ("compares the check fixture's recorded
request.query", not request generally).

**Second attempt** (also rejected): compared the FULL live query string against the fixture's
recorded query, with an empty/absent fixture query strictly requiring zero query params on the wire.
This broke ~28 more real bundles using `api_key_query` auth mode (nasa, openweather, aviationstack,
ticketmaster, tmdb, ...) — that auth mode injects its own `api_key=...` param onto EVERY request
including Check(), completely independent of `check.query`/`RequestSpec.Query`, and those bundles'
fixtures (correctly, for their era) never recorded it. It also surfaced several bundles
(appcues, buzzsprout, chargebee, pexels-api, sonar-cloud, weatherstack, zoho-*) whose
`fixtures/check.json` records a `request.query` the bundle's OWN `streams.json` doesn't even declare
in `check.query` — a real but entirely separate pre-existing data-drift concern this dispatch has no
mandate to adjudicate.

**Final design**: `checkCheckFixture` passes the bundle's OWN declared `base.check.query` KEY SET
(not the fixture's) into the replay server. Matching is scoped to exactly those keys: for each key
the bundle declares in check.query, the fixture's recorded `request.query` must carry that same key
with the same value, or the match fails (this is precisely the ledger's named scenario — a fixture
recorded before check.query existed, or never updated, must fail loudly). Any OTHER query param on
the wire (auth-injected params, or a fixture's own aspirational/stale extra keys) is never inspected.
A bundle that declares no check.query is completely unaffected — its replay server matches
unconditionally, exactly as before this dispatch existed.

`checkFixtureFile` gained an optional `Request fixtureRequest` field (reusing the existing
`fixtureRequest` type stream pages already use — method/path/query — even though only `.Query` is
read); `newCheckReplayServer` takes an additional `checkQueryKeys []string` parameter.

Two new conformance testdata bundles: `testdata/dynamic-invalid/check-query-mismatch` (streams.json
declares check.query, fixtures/check.json's `request` field is absent — proves check_fixture now
FAILS on the ledger's exact scenario) and `testdata/good/acme-check-query` (fixtures/check.json's
recorded request.query matches check.query exactly — proves check_fixture correctly PASSES, not just
unconditionally fails once check.query exists).

### Verify

```
$ go test ./internal/connectors/conformance -run TestCheckFixture -v
PASS (TestCheckFixture_AndReadFixtureNonempty, TestCheckFixture_FailsWhenBundleSendsQueryFixtureDid
NotRecord, TestCheckFixture_PassesWhenFixtureRecordsMatchingQuery)

$ go test ./internal/connectors/conformance -count=1
FAIL — exactly 15 named TestConformance/<bundle> subtests fail (see below); every other bundle
(~396) passes, including every unit test in the package.

$ go test ./internal/connectors/paritytest/... -count=1
ok (all 10 packages)
```

---

## Final self-verify

```
$ go build ./...
(clean)

$ go test ./internal/connectors/engine ./cmd/connectorgen -count=1
ok   ok

$ go test ./internal/connectors/conformance -count=1 -run TestConformance
FAIL — exactly 15 named subtests fail (below); DOWN from the 151-bundle pre-dispatch baseline
(150 from hardening-ledger.md's "Newly exposed" + 1 unrelated concurrent-repair-in-progress drift).
136 bundles moved from failing to passing. Zero bundles moved from passing to failing (verified by
diffing the full failing-subtest list before/after, twice, after each design correction above).

$ go vet ./...
(clean)

$ make lint
0 issues.

$ gofmt -l <every touched file>
(clean)
```

## `conformance_failing` — final list (15)

7 bundles fail to LOAD at all (bare `base.query`, NOT `base.check.query` — a separate, still
out-of-scope invented-mechanism class per hardening-ledger.md's "Newly exposed" section; HTTPBase
still has no Query field):

reply-io, retailexpress-by-maropost, retently, revenuecat, revolut-merchant, ringcentral,
you-need-a-budget-ynab

8 bundles load and validate cleanly but fail `check_fixture` specifically because their
`fixtures/check.json` was recorded before `base.check.query` existed (or never updated to match it)
— exactly the scenario this dispatch's own item 5 verify step exists to catch. Not fixed here (bundle
dirs are out of scope for this dispatch):

aha (check.query={per_page:1}, fixture request.query={}), aircall (per_page:1 / {}),
callrail (per_page:1 / {}), financial-modelling (limit:1 / {}), finnhub (exchange:US / {}),
flexport (per:1 / {}), referralhero (page:1,per_page:1 / {}), rentcast (limit:1,offset:0 / {})

Suggested follow-up per case (not applied here): re-record fixtures/check.json's `request.query` to
match the bundle's declared check.query (if the params are genuinely sent and real), or remove the
check.query block if it turns out to be another dead invented-mechanism instance (a live parity check
against the real API would settle which, per-bundle).
