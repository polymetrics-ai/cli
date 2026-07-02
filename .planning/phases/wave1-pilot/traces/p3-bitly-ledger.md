# p3-bitly-ledger — wave1-pilot P-3 (bitly)

Scope: `internal/connectors/defs/bitly/**`, `internal/connectors/paritytest/bitly/**`. No
`hooks/bitly/` (no Tier-2 trigger applies — bitly is a pure Tier-1 declarative-HTTP migration).
No `git commit` performed by this task.

Legacy read (read-only reference, not modified): `internal/connectors/bitly/bitly.go` (322 loc),
`internal/connectors/bitly/streams.go` (171 loc), `internal/connectors/bitly/bitly_test.go` (171
loc). Read in full before authoring the bundle.

## Red-first protocol

### RED

`internal/connectors/paritytest/bitly/parity_test.go` was authored FIRST, loading the bundle via
`engine.Load(defs.FS, "bitly")` (per dispatch instructions). Since `defs.FS`'s embed contents were
flaky mid-dispatch (sibling wave1-pilot agents concurrently writing `defs/calendly`,
`defs/chargebee`, `defs/sentry`, `defs/zendesk-support`, etc. — each transiently an incomplete,
non-embeddable directory, breaking `//go:embed all:*` for the WHOLE `defs` package independent of
bitly), RED evidence was captured via an isolated `os.DirFS`-based probe against
`internal/connectors/defs` directly (same `engine.Load` call the production parity test makes,
just bypassing the transiently-broken shared embed for this one-time capture; the probe package
was deleted immediately after capturing the failure — no lingering scratch files):

```
$ go test ./internal/connectors/redcheck_tmp/... -run TestBitlyBundleLoadsFromDisk -v
=== RUN   TestBitlyBundleLoadsFromDisk
    red_check_test.go:13: engine.Load(defs, bitly): load bundle bitly: missing required file metadata.json
--- FAIL: TestBitlyBundleLoadsFromDisk (0.00s)
FAIL
FAIL	polymetrics.ai/internal/connectors/redcheck_tmp	0.436s
FAIL
```

This confirms `internal/connectors/defs/bitly/` genuinely did not exist before authoring began
(`ls` also confirmed: "No such file or directory").

### GREEN

After authoring the full bundle (metadata.json, spec.json, streams.json, 4 schemas,
api_surface.json, docs.md, fixtures/{check.json, streams/{organizations,groups,campaigns,
bitlinks}/page_1.json}) plus the paritytest suite, both green:

```
$ go test ./internal/connectors/paritytest/bitly -v
=== RUN   TestParityBitly_GroupsStreamRecords
--- PASS: TestParityBitly_GroupsStreamRecords (0.00s)
=== RUN   TestParityBitly_OrganizationsStreamRecords
--- PASS: TestParityBitly_OrganizationsStreamRecords (0.00s)
=== RUN   TestParityBitly_CampaignsStreamRecords
--- PASS: TestParityBitly_CampaignsStreamRecords (0.00s)
=== RUN   TestParityBitly_BitlinksStreamPaginates
--- PASS: TestParityBitly_BitlinksStreamPaginates (0.00s)
=== RUN   TestParityBitly_BitlinksAbsoluteNextURLNotRelative
--- PASS: TestParityBitly_BitlinksAbsoluteNextURLNotRelative (0.00s)
=== RUN   TestParityBitly_BearerAuthHeaderParity
--- PASS: TestParityBitly_BearerAuthHeaderParity (0.00s)
=== RUN   TestParityBitly_ErrorPathParity
--- PASS: TestParityBitly_ErrorPathParity (0.00s)
=== RUN   TestParityBitly_BundleLoadsAndValidates
--- PASS: TestParityBitly_BundleLoadsAndValidates (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/paritytest/bitly	0.361s

$ go test ./internal/connectors/conformance -run 'TestConformance/bitly' -v
=== RUN   TestConformance
=== RUN   TestConformance/bitly
--- PASS: TestConformance (0.01s)
    --- PASS: TestConformance/bitly (0.01s)
PASS
```

## Bundle shape decisions

- **Auth**: Bearer, `Authorization: Bearer {{ secrets.api_key }}` — matches legacy
  `connsdk.Bearer(secret)` (`bitly.go:239`) byte-for-byte (asserted in
  `TestParityBitly_BearerAuthHeaderParity`).
- **Streams**: `organizations`, `groups`, `campaigns` — simple `GET`, no pagination, records at
  the top-level key matching the stream name, no incremental cursor (legacy `bitly.go:36`'s own
  comment: none of the core list endpoints expose one). `bitlinks` — group-scoped via
  `{{ config.group_guid }}` path templating (urlencoded by `InterpolatePath`'s default, matching
  legacy's `url.PathEscape(guid)`), `pagination.type: next_url` /
  `next_url_path: "pagination.next"` (Bitly's real wire shape).
- **N3 verification (per dispatch note)**: confirmed bitly's `pagination.next` is always an
  ABSOLUTE URL — legacy's own `bitly_test.go:64` (`TestReadBitlinksPaginates`) fixture serves
  `srv.URL + "/groups/g1/bitlinks?search_after=tok2"`, and `bitly.go:180-183`'s own comment says
  so explicitly. The wave0 N3 relative-URL fail-closed guard does NOT bite; no `ENGINE_GAP` filed
  for N3.
- **Fixture-mode-only fields excluded**: legacy's `readFixture` path stamps `connector`/
  `fixture`/`previous_cursor` onto records only when `config.mode == "fixture"` — these are not
  part of the live wire shape. Per SPEC.md §5.2's explicit instruction, this bundle's schemas and
  parity suite target ONLY the live path (`bitly.go`'s `harvest`), matching every parity test in
  `parity_test.go` (all driven against `httptest.Server`s, never `mode=fixture`).

## Deviations (parity-deviation ledger candidates, conventions.md §5 meta-rule)

1. **`page_size`/`max_pages` config-driven overrides not modeled (ACCEPTABLE).** Legacy exposes
   both as config-driven overrides on the `bitlinks` stream (`bitlyPageSize`/`bitlyMaxPages`,
   `bitly.go:286-314`). The engine's `next_url` paginator never reads `PaginationSpec.PageSize`/
   has no `MaxPages`-equivalent knob wired to a runtime config value, and `stream.Query`
   templating has no absent-key-falsy tolerance (a `{{ config.page_size }}` template would
   hard-error whenever `page_size` is unset — the common case). This bundle sends bitly's own
   default (`size=50`) as a static per-stream query literal — the identical "static default value,
   no runtime override" pattern stripe's golden already established for its `limit=100`
   (conventions.md's stripe worked example). `page_size` is consequently NOT declared in
   `spec.json` at all (F6, REVIEW.md precedent: a declared-but-unwireable config key is worse than
   an absent one). Never diverges for any input legacy itself would accept when no override is
   configured (the common/default case); only diverges from legacy's non-default,
   config-override-supplied page sizes, which is a strictly-additive legacy capability this
   engine dialect cannot express yet. Documented in docs.md "Known limits".
2. **`group_guid`-missing error text differs, same failure classification (ACCEPTABLE, precedent:
   ledger item 9, postgres).** Legacy: `"bitly bitlinks stream requires config group_guid"`
   (a hand-written `errors.New`). Engine: an unresolved `config.group_guid` path-interpolation
   hard error (different literal wording, naming the key/namespace). Both fail closed for the
   identical input (an unset `group_guid` on the `bitlinks` stream); no behavior-changing
   divergence for any accepted input.

No `ENGINE_GAP` blockers. No Tier-2/Tier-3 escalation needed — pure Tier-1 declarative bundle.

## Self-verify (conventions.md §7)

```
$ go run ./cmd/connectorgen validate internal/connectors/defs --json | jq
    # (validateDir's actual contract validates a PARENT dir of named bundle subdirs, per
    # cmd/connectorgen/main_test.go's TestValidate_AcceptsGoodBundle calling
    # validateDir(os.DirFS("testdata/valid")) — never the bundle dir itself; running validate
    # directly against internal/connectors/defs/bitly mis-treats fixtures/ and schemas/ as
    # candidate bundle dirs and produces 2 spurious missing_file findings unrelated to bitly's
    # real bundle content. Ran against the parent instead, per the tool's actual contract, and
    # filtered for connector=="bitly".)
    -> connectors_checked: 13, zero findings/warnings for connector "bitly"
      (3 pre-existing findings are unrelated siblings: github docs.md missing_file, monday
      surface_fail_first_run, zendesk-support docs.md missing_file — other agents' in-flight work,
      not bitly's)

$ go build ./internal/connectors/defs/... ./internal/connectors/paritytest/bitly/... ./internal/connectors/bitly/...
    -> clean

$ go vet ./internal/connectors/defs/... ./internal/connectors/paritytest/bitly/... ./internal/connectors/bitly/...
    -> clean

$ go test ./internal/connectors/conformance -run 'TestConformance/bitly' -v
    -> PASS

$ go test ./internal/connectors/paritytest/bitly -v
    -> PASS (8/8)

$ go build ./...
    -> clean (confirmed at time of this report; whole-repo build was transiently broken several
       times during the dispatch window purely by OTHER wave1-pilot sibling agents' in-flight
       writes to defs/calendly, defs/chargebee, defs/sentry, defs/zendesk-support,
       hooks/gmail, paritytest/monday, paritytest/zendesk-support — never by bitly)

$ golangci-lint run ./internal/connectors/defs/bitly/... ./internal/connectors/paritytest/bitly/...
    -> 0 issues
```

## Handoff notes for P-11 review / P-12 conventions patch

- Confirms conventions.md's existing stripe precedent (2-page fixture required only for the
  bundle's FIRST stream, since `pagination_terminates` only exercises `b.Streams[0]`) generalizes
  cleanly to a bundle whose ONLY paginated stream is NOT first: `organizations` (non-paginated) is
  `bitly`'s `Streams[0]`, so `bitlinks` ships a single-page fixture (`pagination.next: ""`,
  immediate stop) rather than a 2-page one.
- **Worth flagging for P-12**: `next_url`-type pagination fixtures cannot express a REAL 2-page
  follow through `conformance`'s static-JSON-file replay harness at all, for ANY bundle, regardless
  of stream ordering — `newStreamReplayServer` builds one `httptest.Server` per test run, so its
  URL isn't known until after the fixture JSON is already loaded from disk, but a `next_url`
  fixture's `pagination.next` value must be a literal absolute URL string committed to that JSON
  ahead of time. (Contrast: the engine's OWN unit tests, e.g. `read_test.go`'s
  `TestReadNextURLPaginationSetsBaseHostFromRequester`, forward-declare `var srv *httptest.Server`
  and close over `srv.URL` inside the handler at response-write time — a trick unavailable to a
  static fixture file.) This bundle's own live-httptest `paritytest` suite
  (`TestParityBitly_BitlinksStreamPaginates`) DOES prove real 2-page `next_url` pagination
  correctness using exactly that forward-declaration trick, so parity coverage is not weakened —
  only `conformance`'s fixture-driven 2-page proof specifically cannot reach a `next_url` stream.
  Any other pilot connector using `next_url` pagination (calendly, P-4) will hit the identical
  limit. Not filed as an `ENGINE_GAP` blocker for bitly since conformance's own
  `pagination_terminates` check never actually requires it here (organizations is stream[0]); flag
  for conventions.md as a documented harness limitation, worked-example candidate for P-12.
