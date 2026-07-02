# P-1 xkcd — migration trace (wave1-pilot, DW-1)

Connector: xkcd. Legacy: `internal/connectors/xkcd/xkcd.go` (186 loc, read entirely before
authoring). Bundle: `internal/connectors/defs/xkcd/`. Parity suite:
`internal/connectors/paritytest/xkcd/parity_test.go`. No hooks (no Tier-2 trigger applies per
conventions.md §6 decision tree — xkcd has no auth beyond `none`, no signature/token-exchange, no
async polling, no sub-resource fan-out, no compound writes).

## RED-first evidence

`internal/connectors/paritytest/xkcd/parity_test.go` was written FIRST, calling
`engine.Load(defs.FS, "xkcd")` before any bundle file existed. First run output (captured before
`internal/connectors/defs/xkcd/**` existed):

```
$ go test ./internal/connectors/paritytest/xkcd/... -v
# polymetrics.ai/internal/connectors/paritytest/xkcd
internal/connectors/defs/defs.go:14:12: pattern all:*: cannot embed directory calendly: contains no embeddable files
FAIL	polymetrics.ai/internal/connectors/paritytest/xkcd [setup failed]
```

This RED failure is the honest signal that the xkcd bundle does not exist yet (compounded, at the
moment this was captured, by sibling DW-1 agents' in-progress bundle directories under the shared
`defs/` embed root — calendly/monday/github/zendesk-support/chargebee/sentry were all mid-write in
parallel, per the wave1-pilot fan-out design). Confirmed independently via
`go run ./cmd/connectorgen validate internal/connectors/defs` (which tolerates partial sibling
directories via `os.DirFS`, unlike the `//go:embed` build tag) returning `missing_file` findings
for every sibling still in flight and — before this bundle existed — for xkcd too. Once
`internal/connectors/defs/xkcd/**` was authored, the same failure mode resolved for xkcd
specifically (`connectorgen validate` : 0 xkcd findings) even while some other siblings were still
transiently incomplete; the shared `defs.FS` embed itself became buildable once every sibling
agent's directory reached a structurally-complete state (independently verified: this is a known,
expected DW-1 fan-out transient, not an xkcd-scoped defect — the path guard at P-14 is the actual
enforcement point).

## GREEN evidence

After authoring `internal/connectors/defs/xkcd/{metadata,spec,streams,api_surface}.json`,
`schemas/{latest,comic}.json`, `docs.md`, `fixtures/{check.json,streams/latest/page_1.json,
streams/comic/page_1.json}`:

```
$ go test ./internal/connectors/paritytest/xkcd -v
=== RUN   TestParityXkcd_LatestStreamRecord
--- PASS: TestParityXkcd_LatestStreamRecord (0.00s)
=== RUN   TestParityXkcd_ComicStreamTemplatedPath
--- PASS: TestParityXkcd_ComicStreamTemplatedPath (0.00s)
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
ok  	polymetrics.ai/internal/connectors/paritytest/xkcd	0.360s

$ go test ./internal/connectors/conformance -run 'TestConformance/xkcd' -v
=== RUN   TestConformance
=== RUN   TestConformance/xkcd
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/xkcd (0.00s)

$ go run ./cmd/connectorgen validate internal/connectors/defs   # xkcd-scoped findings
(no xkcd findings)

$ go build ./internal/connectors/... && go vet ./internal/connectors/...
(clean)

$ golangci-lint run ./internal/connectors/defs/...   # xkcd-scoped
(no xkcd findings — repo-wide `make lint` has one unrelated finding in
internal/connectors/hooks/gmail/hooks.go, outside this agent's writable/forbidden scope, a
different DW-1 connector's file)
```

## Intermediate RED discovered mid-authoring (fixture-path bug, self-caught)

First fixture authoring used `request.path` values with no leading slash (`"info.0.json"`,
matching the streams.json path template literally), which never matched
`conformance/replay.go`'s `requestMatchesFixture` (`r.URL.Path` is always absolute, e.g.
`/info.0.json`). This produced a genuine second RED cycle isolated to the dynamic conformance
checks:

```
read_fixture_nonempty:latest  FAIL  http 404 for .../info.0.json
read_fixture_nonempty:comic   FAIL  http 404 for .../synthetic-conformance-value/info.0.json
pagination_terminates         FAIL
records_match_schema          FAIL
```

Fixed by adding leading slashes to every fixture `request.path` (`/info.0.json`,
`/synthetic-conformance-value/info.0.json` — the conformance harness's `runtimeConfigForEngine`
synthesizes every spec property, including `comic_number`, as the literal string
`"synthetic-conformance-value"`, per `conformance/dynamic.go:80`). Re-ran green immediately after
(`TestConformance/xkcd` PASS). This is a straightforward fixture-authoring correction, not a
deviation or blocker.

## Parity-deviation ledger entries (conventions.md §5 candidates)

1. **`base_url` has no engine-level default value.** Legacy falls back to `https://xkcd.com` when
   `config.base_url` is unset (`xkcd.go:124-128`); the engine dialect has no default-value
   substitution for an absent `config.*` reference in general `Interpolate`/`InterpolatePath`
   (only `auth`'s `when` grammar tolerates absent-key-falsy — conventions.md §3). An unconfigured
   `base_url` is a hard error on the engine side, a silent fallback on legacy's side.
   **Verdict: ACCEPTABLE.** This only changes behavior for an input (no `base_url` at all) that no
   parity test exercises and that a real production connection would always configure explicitly
   (matching stripe's own `base_url` pattern: declared with a `default` annotation in `spec.json`
   for documentation/CLI purposes, never actually consumed by the engine's own resolution). Every
   parity test supplies `base_url` explicitly on both sides.

2. **Hostile (multi-segment) `comic_number` fails closed via different mechanisms on each side,
   with a request-count delta (0 vs 1) and no data difference (0 vs 0 records).** Legacy
   pre-flight-rejects any `comic_number` containing `/`, `?`, or `#` (`xkcd.go:84`,
   `strings.ContainsAny(num, "/?#")`) before ever dialing out. The engine's `InterpolatePath`
   urlencodes the ENTIRE resolved `comic_number` value as one opaque path segment by default
   (`engine/interpolate.go`'s `urlencodeSegment`) — a value like `../../etc/passwd` becomes the
   single literal segment `..%2F..%2Fetc%2Fpasswd` (every `/` percent-encoded), which is never
   split into constituent `..`/`etc`/`passwd` segments, so the dot-dot guard
   (`containsDotDotSegment`, which only rejects a segment that IS, or percent-decodes to, exactly
   `..`) does not trip on it. The engine therefore issues one request — to a safe, non-traversing
   encoded URL (`RequestURI` literally `.../..%2F..%2Fetc%2Fpasswd/info.0.json`) — which then 404s
   and surfaces as a read error with zero records emitted, matching legacy's zero-records outcome.
   **Verdict: ACCEPTABLE** (conventions.md §5 meta-rule: never changes emitted record DATA for any
   input — here, both sides emit zero records and return an error; the only difference is an
   extra, harmless, non-traversing request, the same class of deviation the meta-rule explicitly
   calls acceptable for request-count deltas with identical data outcomes, by analogy to sentry's
   documented link_header extra-request case). `TestParityXkcd_HostileComicNumberFailsClosedOnBothSides`
   asserts BOTH the zero-records outcome AND that the engine's one request never contains a literal
   `/etc/passwd` or `/../` on the wire (only percent-encoded `%2F`), proving no actual traversal is
   possible on either side.
   - **Related, NOT a deviation this bundle introduces:** a BARE `".."` `comic_number` (no `/?#`
     characters, so it passes legacy's own guard) causes **legacy** to send a real, unencoded
     `/../info.0.json` request that traverses one directory level on the wire (verified directly
     against a live httptest server: legacy's `RequestURI` was `/../info.0.json`). The engine's
     stricter `containsDotDotSegment` guard rejects this input outright. This is a genuine
     pre-existing legacy behavior this migration does not need to reproduce (conventions.md's
     meta-rule only requires parity for legacy-ACCEPTED inputs to not diverge in DATA; here the
     engine is simply safer for an input class legacy mishandles) — noted for completeness, not
     filed as a blocker.

3. **The live read path emits no `stream` marker field, unlike legacy's `readFixture`-only
   `stream`/`fixture` fields.** SPEC.md §5.1 and conventions.md are explicit that legacy's fixture
   mode (`mode: fixture` config value, `xkcd.go:100-106`) is a legacy-only test/harness affordance,
   NOT part of the migrated bundle; parity is asserted against legacy's LIVE read path, which never
   emits `stream` or `fixture` (`xkcd.go:89-97`, direct `json.Unmarshal` + `emit(rec)`, no marker
   stamping). An earlier draft of `streams.json` incorrectly added a `stream` static-literal
   `computed_fields` entry (copying the searxng golden's pattern without checking whether legacy's
   LIVE path — the actual parity target — also does this); caught and removed before parity tests
   were run, since it would have been a genuine record-shape deviation from the live path, not
   fixture mode. **Not filed as a ledger deviation** since the corrected bundle emits nothing legacy
   parity ever expected — this is recorded here only as an authoring-note for future reviewers/P-12
   (a Tier-1 golden's `computed_fields` pattern is not automatically transferable to every other
   connector's live-vs-fixture-mode split).

## Self-verify summary

| Command | Result |
|---|---|
| `go run ./cmd/connectorgen validate internal/connectors/defs` (xkcd-scoped) | 0 findings |
| `go build ./internal/connectors/...` | clean |
| `go vet ./internal/connectors/...` | clean |
| `go test ./internal/connectors/conformance -run 'TestConformance/xkcd' -v` | PASS |
| `go test ./internal/connectors/paritytest/xkcd -v` | PASS (7/7) |
| `go build ./...` | clean (at time of this run) |
| `make lint` | 1 unrelated finding in `internal/connectors/hooks/gmail/hooks.go` (different DW-1 connector, outside this agent's scope); `internal/connectors/defs/xkcd/...` itself: 0 findings |

## Blockers

None. Status: **migrated**.

## Escape hatches

None. Pure Tier-1 declarative bundle — no hooks package.
