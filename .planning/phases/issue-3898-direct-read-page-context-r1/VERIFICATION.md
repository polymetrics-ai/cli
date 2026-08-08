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
- The follow-up same-machine controls isolate the EOF to PM's transport path:
  curl POST (including PM's HTTP/1.1, `Connection: close`, User-Agent, and API
  version shape) and `gh api` POST each received HTTP 201, while a fresh PM
  plan → preview → approved run read and declared exactly 44 body bytes then
  received EOF. Header values and credentials were never captured. The private
  repository contains only the four intentional curl/`gh` diagnostic issues;
  no PM write reached it. The live-write completion gate remains blocked.
- A deterministic stale-idle connection test disproved the proposed missing
  `GetBody` explanation. JSON already reaches `http.NewRequest` as a concrete
  `*bytes.Reader`; strict writes deliberately clear replay capability only
  after construction and instead force a fresh connection. Re-enabling replay
  would weaken the non-idempotent mutation safety contract, so no such change
  was made.
- `httptrace` places the GitHub EOF after `WroteRequest` and before any first
  response byte; TCP and TLS both completed. The same binary successfully
  dispatched one normal reverse write to a loopback fixture, while a self-signed
  local TLS fixture was correctly refused rather than weakening certificate
  verification. PM and curl inherit no configured HTTP proxy, but the machine
  does have one connected VPN service. No network change was made; the live
  GitHub write gate remains blocked.
- `docs/connectors/**` regeneration carries unrelated pre-existing drift
  (field types). Verifiable by running `pm docs generate` on a clean tree, which
  changes 1028 files with no code changes at all.
- `internal/cli` needs ~660s and exceeds `go test`'s 600s default on this
  machine. Pre-existing, not caused by this change.
