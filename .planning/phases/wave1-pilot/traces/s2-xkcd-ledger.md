# S2 xkcd repair — trace ledger (wave1-pilot gap-loop cycle-1, Step 2)

Connector: xkcd. Source: `GAP-LOOP-PLAN.md` Step 2 bullet 1 + `REVIEW-B.md`'s xkcd block
(verdict: FAIL, 1 blocker + 1 minor). Legacy: `internal/connectors/xkcd/xkcd.go` (read entirely,
re-confirmed lines 93-97 before any edit). Files touched: `internal/connectors/defs/xkcd/**`,
`internal/connectors/paritytest/xkcd/parity_test.go`, this ledger. No other files touched (scope
guard: `git diff --stat` confirms only these 8 files changed).

## Finding recap (REVIEW-B.md)

1. **BLOCKER** — legacy's read path is a raw passthrough (`json.Unmarshal(resp.Body, &rec);
   emit(rec)`, xkcd.go:93-97): EVERY field of the real API response reaches the caller. The real
   XKCD JSON API (https://xkcd.com/json.html) returns 11 fields: `month, num, link, year, news,
   safe_title, transcript, alt, img, title, day`. The bundle's schemas declared only 6
   (`num,title,safe_title,year,month,day`, copied from legacy's `Catalog()` field list — a
   capability-discovery listing, NOT the record-shaping function conventions.md §2 mandates as the
   schema source). In `"schema"` projection mode (the engine default; `engine/read.go`'s
   `projectRecord`), `link/news/transcript/alt/img` were silently dropped on every real read — a
   real record-DATA divergence from legacy for every input legacy accepts. The parity suite's own
   6-field fixture body masked it (DeepEqual never saw the dropped fields), and that same 6-field
   fixture also violated conventions.md §4's recorded-real-shape rule.
2. **minor** — `docs.md:60-63` claimed `spec.json`'s `base_url` "still declares `"default":
   "https://xkcd.com"`" — it does not (no `default` key anywhere in `spec.json`). Doc/code drift.

## Decision: passthrough, not full-schema-only (picked per legacy fidelity — legacy is ground truth)

REVIEW-B.md offered two fixes: (a) `"projection": "passthrough"` on both streams (exact analog of
legacy's own behavior), or (b) declare all 11 real fields under `"schema"` projection. Chose a
**combination that satisfies the "legacy is ground truth" rule literally**: legacy does not
project or filter at all — it unmarshals the whole response body and emits it unconditionally, so
the emission mechanism must be passthrough, not a fixed-field schema (option (b) alone would still
silently drop any 12th field the real API adds later, which legacy would emit and the bundle
would not — the same class of bug, just with a wider net). Applied BOTH: `streams.json` now
declares `"projection": "passthrough"` on `latest` and `comic` (engine already supports this
mode — `engine/bundle.go:182`, `engine/read.go:598-614` — no engine change needed, confirmed
available per Step 1/`docs/migration/conventions.md` §3 "Projection" section), AND the schemas
were widened to document all 11 real fields (`required`/`x-primary-key`/types), since under
passthrough mode the schema's `properties` set is no longer the emission gate — it is pure
documentation/validation surface — so there is no reason to leave it stale at 6 fields. This
matches conventions.md §2's rule precisely once corrected: "Schema-as-projection... derived
field-for-field from what the legacy connector's own record-shaping function actually emits" —
for a raw-passthrough legacy connector, the record-shaping function IS the identity function over
the whole raw body, so passthrough projection is the literal, mechanical translation of that rule,
not a workaround.

## RED-first evidence

Before touching `streams.json`/schemas, added `TestParityXkcd_AllElevenRealFieldsSurvivePassthrough`
to `internal/connectors/paritytest/xkcd/parity_test.go` and rewrote `xkcdFixtureBody` (used by
`TestParityXkcd_LatestStreamRecord`) and the comic stream's `comicBody` literal (used by
`TestParityXkcd_ComicStreamTemplatedPath`) to the full realistic 11-field shape BEFORE changing any
bundle file, then ran the suite against the unmodified (6-field-schema, `"schema"`-projection)
bundle:

```
$ go test -count=1 ./internal/connectors/paritytest/xkcd/... \
    -run 'TestParityXkcd_AllElevenRealFieldsSurvivePassthrough|TestParityXkcd_LatestStreamRecord|TestParityXkcd_ComicStreamTemplatedPath' -v
=== RUN   TestParityXkcd_LatestStreamRecord
    parity_test.go:155: latest record mismatch:
        engine:  map[day:1 month:1 num:42 safe_title:Geography title:Geography year:2006]
        legacy:  map[alt:alt text day:1 img:https://imgs.xkcd.com/comics/geography.png link: month:1 news: num:42 safe_title:Geography title:Geography transcript:transcript text year:2006]
--- FAIL: TestParityXkcd_LatestStreamRecord (0.00s)
=== RUN   TestParityXkcd_ComicStreamTemplatedPath
    parity_test.go:195: comic record mismatch:
        engine:  map[day:9 month:9 num:614 safe_title:Woodpecker title:Woodpecker year:2009]
        legacy:  map[alt:woodpecker alt text day:9 img:https://imgs.xkcd.com/comics/woodpecker.png link: month:9 news: num:614 safe_title:Woodpecker title:Woodpecker transcript:woodpecker transcript year:2009]
--- FAIL: TestParityXkcd_ComicStreamTemplatedPath (0.00s)
=== RUN   TestParityXkcd_AllElevenRealFieldsSurvivePassthrough
    parity_test.go:240: engine record dropped real API field "link" (schema-projection silently discarding a field legacy passes through) — got map[day:1 month:1 num:42 safe_title:Geography title:Geography year:2006], want field present as in legacy map[alt:alt text day:1 img:https://imgs.xkcd.com/comics/geography.png link: month:1 news: num:42 safe_title:Geography title:Geography transcript:transcript text year:2006]
--- FAIL: TestParityXkcd_AllElevenRealFieldsSurvivePassthrough (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/paritytest/xkcd	0.352s
```

This RED evidence reproduces REVIEW-B.md's finding exactly (engine record missing `link`, `news`,
`transcript`, `alt`, `img`; legacy record has all 11) — proof the failure is real and the fix
targets the right thing, not a fixture artifact.

## GREEN evidence

Applied the fix (`streams.json`: `"projection": "passthrough"` on both streams; schemas widened to
11 fields; fixtures re-recorded in the full 11-field real shape — `fixtures/check.json`,
`fixtures/streams/{latest,comic}/page_1.json`):

```
$ go test -count=1 ./internal/connectors/paritytest/xkcd/... -v
=== RUN   TestParityXkcd_LatestStreamRecord
--- PASS: TestParityXkcd_LatestStreamRecord (0.00s)
=== RUN   TestParityXkcd_ComicStreamTemplatedPath
--- PASS: TestParityXkcd_ComicStreamTemplatedPath (0.00s)
=== RUN   TestParityXkcd_AllElevenRealFieldsSurvivePassthrough
--- PASS: TestParityXkcd_AllElevenRealFieldsSurvivePassthrough (0.00s)
=== RUN   TestParityXkcd_HostileComicNumberFailsClosedOnBothSides
--- PASS: TestParityXkcd_HostileComicNumberFailsClosedOnBothSides (0.00s)
=== RUN   TestParityXkcd_NotFoundErrorPathParity
--- PASS: TestParityXkcd_NotFoundErrorPathParity (0.00s)
=== RUN   TestParityXkcd_WriteUnsupportedOnBothSides
--- PASS: TestParityXkcd_WriteUnsupportedOnBothSides (0.00s)
=== RUN   TestParityXkcd_ManifestSurface
--- PASS: TestParityXkcd_ManifestSurface (0.00s)
=== RUN   TestParityXkcd_BundleLoadsAndValidates
--- PASS: TestParityXkcd_BundleLoadsAndValidates (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/paritytest/xkcd	0.435s
```

`TestParityXkcd_AllElevenRealFieldsSurvivePassthrough` asserts, per real field name in
`xkcdAllRealFieldNames` (`month, num, link, year, news, safe_title, transcript, alt, img, title,
day`), that the field is present on BOTH sides, then asserts full `reflect.DeepEqual` AND an exact
field-count match (guards against the schema silently regressing to a narrower set again in the
future — a widening-only guard, matching the recorded-real-shape rule's intent).

## Full self-verify (re-run after fix; commands from the task's SELF-VERIFY line)

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 13 connector(s) checked, 0 findings

$ go test -count=1 ./internal/connectors/conformance/... -run 'TestConformance/xkcd' -v
=== RUN   TestConformance
=== RUN   TestConformance/xkcd
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/xkcd (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/conformance	0.453s

$ go test -count=1 ./internal/connectors/paritytest/xkcd/... -v
(all 8 subtests PASS — see GREEN evidence above)

$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.

$ go build ./... && go vet ./internal/connectors/...
(clean, no output)
```

Also ran the broader scoped suite (`paritytest/...`, `conformance/...`, `engine/...`,
`cmd/connectorgen/...`) to check for cross-connector regressions from this change: xkcd, bitly,
calendly, chargebee, github, monday, sentry, vitally, zendesk-support all PASS. One unrelated
pre-existing failure was observed — `paritytest/gmail`'s
`TestParityGmail_ComputedFieldsStringifyLabelCountFields` — in a sibling connector directory this
task does not touch (gmail is a different Step-2 gap-loop bullet, being repaired by a different
parallel agent per `GAP-LOOP-PLAN.md`'s "parallel, disjoint dirs" instruction; `git diff --stat`
confirms zero xkcd-repair files touch anything under `defs/gmail` or `paritytest/gmail`). Not in
scope for this ledger; flagged here only so the coordinator doesn't misattribute it.

## Docs fix (minor finding 2)

`docs.md`'s "Known limits" section corrected: removed the false claim that `spec.json`'s
`base_url` property "still declares `"default": "https://xkcd.com"`" (verified: no `default` key
exists anywhere in `spec.json`); replaced with an accurate statement that the legacy default is
documented in the property's `description` text only, with no CLI-affordance `default` annotation
(unlike stripe's `base_url` pattern, which does have one). Also rewrote the "Streams notes" section
to explain the passthrough decision, why `Catalog()` was the wrong schema source, and what the
schema's 11 properties now mean under passthrough mode (documentation/validation surface, not an
emission filter) — closing the "neither `docs.md`... mentions it" half of finding 1.

## Files changed

- `internal/connectors/defs/xkcd/streams.json` — `"projection": "passthrough"` added to both
  `latest` and `comic` streams.
- `internal/connectors/defs/xkcd/schemas/latest.json`,
  `internal/connectors/defs/xkcd/schemas/comic.json` — widened from 6 to 11 properties
  (`link`, `news`, `transcript`, `alt`, `img` added), matching the real XKCD JSON API's field list.
- `internal/connectors/defs/xkcd/fixtures/check.json`,
  `internal/connectors/defs/xkcd/fixtures/streams/latest/page_1.json`,
  `internal/connectors/defs/xkcd/fixtures/streams/comic/page_1.json` — re-recorded in the real
  11-field shape (synthetic values, recorded-real-shape rule).
- `internal/connectors/defs/xkcd/docs.md` — "Streams notes" rewritten to document the passthrough
  decision and its rationale; "Known limits" `base_url` claim corrected.
- `internal/connectors/paritytest/xkcd/parity_test.go` — `xkcdFixtureBody` and the comic stream's
  `comicBody` rewritten to the real 11-field shape; added
  `TestParityXkcd_AllElevenRealFieldsSurvivePassthrough` (previously-dropped-field regression
  guard, per-field presence check + exact field-count check).

## Note: transient shared-tree `connectorgen validate` finding (not xkcd-scoped)

At the time of this ledger's final self-verify pass, `go run ./cmd/connectorgen validate
internal/connectors/defs` reported 1 finding — entirely in `github`'s `streams.json`
(`interpolation_unresolved` on an `auth`/`when` field referencing an undeclared spec key), zero
`xkcd` findings. `git status --porcelain` at the same moment shows `github/{docs.md,streams.json,
schemas/**}` modified by a different parallel Step-2 agent (github is its own Step-2 gap-loop
bullet, "auth_type + secret-alias surface restored"), consistent with the same shared-`defs/`-tree
transient documented in `p1-xkcd-ledger.md`'s RED-first evidence for sibling in-flight bundles.
Re-run in isolation confirms: `go test -count=1 ./internal/connectors/conformance/... -run
'TestConformance/xkcd'` and `go test -count=1 ./internal/connectors/paritytest/xkcd/...` both PASS
cleanly at this same HEAD state; the xkcd bundle itself is not implicated. Flagged for the
orchestrator, not actionable from this ledger's scope (xkcd-only, no touches to `defs/github/**`).

## Verdict

Both REVIEW-B.md xkcd findings (1 blocker, 1 minor) closed. No test weakened; parity coverage
strictly widened (11-field assertion added, 6-field assertions upgraded in place to 11-field
bodies). No engine change was required — `"projection": "passthrough"` and the C3 spec-default
materialization mentioned in Step 1 were both already available at this HEAD; only the latter was
inspected and found not applicable here (xkcd's `base_url` deviation is a documented, ACCEPTABLE
fail-loud strictness per conventions.md §5, not something this ledger's scope asked to change).
