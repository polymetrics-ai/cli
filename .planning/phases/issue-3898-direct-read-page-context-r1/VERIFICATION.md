# VERIFICATION — Issue #3898: direct-read page context

## Goal-backward check

**Does a direct read still report a completeness it cannot prove?** No.

## Evidence

Fixture-backed runs of the real `pm` binary (local `127.0.0.1` fixture serving
120 records, provider default page 30):

```
github  pulls files view              -> records: 100, complete: false, next_number: 2
github  pulls files view --page 2     -> records:  20, complete: true          (100 + 20 = 120)
gong    logs list                     -> records:  30, complete: false, next_cursor: 30
gong    logs list --page-cursor 30/60/90 -> 30 + 30 + 30, final complete: true (120 total)
notion  comment list                  -> records: 100, complete: false, next_cursor: 100
```

gong returns 30 per page because its bundle declares no `size_param`. That is
correct derived behaviour, and it is now reported explicitly instead of being
silent — raising it is a one-line bundle declaration, not engine work.

Human-readable path: stdout stays clean JSON, the notice goes to stderr —
`note: page 1 of a paged result (100 records); more remain — rerun with --page 2`.

## Gates

```
gofmt -l cmd internal                     clean
go vet ./internal/... ./cmd/...           clean
go test ./internal/connectors/...         ok
go test ./internal/cli/ -timeout 30m      ok (660s; exceeds the 600s default — pre-existing)
go test ./internal/app/...                ok
make lint                                 0 issues
make tidy-check docs-check                ok
make agent-contract-check                 ok
make connectorgen-validate                551 connectors, 0 findings
make connectorgen-surface-sync            551 scanned, 0 fields changed  <-- sweep not invalidated
make connector-boundary                   ok
make release-workflow-check               ok
make smoke-no-build                       ok
node website/scripts/cli-surface.test.mjs 7/7 pass
```

## Known, stated plainly

- The GitHub live reverse trace did not send `/issues%22`: the `pm` argv,
  stored connection/plan state, and resolved target immediately before
  `client.Do` were quote-free. `%22` was introduced by the former
  `RedactErrorText` URL regex when it absorbed Go's quoted error delimiter;
  the safety regression now prevents that misleading rendering while preserving
  query redaction. The underlying GitHub result was transport `EOF`, not a
  malformed endpoint.
- `docs/connectors/**` regeneration carries unrelated pre-existing drift
  (field types). Verifiable by running `pm docs generate` on a clean tree, which
  changes 1028 files with no code changes at all.
- `internal/cli` needs ~660s and exceeds `go test`'s 600s default on this
  machine. Pre-existing, not caused by this change.
