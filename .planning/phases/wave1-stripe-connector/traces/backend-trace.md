# Agent Trace: backend — wave1-stripe-connector (EXECUTE)

## Rendered Prompt Or Prompt Reference

BACKEND role: implement `stripe` as the reference declarative-HTTP per-system connector on the
`connsdk` toolkit (mirroring `github`), turn the red-first stripe tests GREEN, flip the
`source-stripe` catalog entry to enabled, and keep `make verify` GREEN. See
`.planning/phases/wave1-stripe-connector/{SPEC.md,PLAN.md,ADR.md,THREAT-MODEL.md}`.

## Files Inspected

- `.planning/phases/wave1-stripe-connector/{SPEC.md,PLAN.md,ADR.md,THREAT-MODEL.md}`
- `internal/connectors/stripe/stripe_test.go` (red-first tests)
- `internal/connectors/connsdk/{http.go,auth.go,extract.go,paginate.go,state.go,connsdk.go,http_test.go}`
- `internal/connectors/github/{github.go,streams.go,manifest.go}` (per-system conventions)
- `internal/connectors/connectors.go`, `native_catalog_connector.go`, `native_conformance.go`,
  `native_conformance_test.go`, `catalog.go`, `catalog_test.go`, `manifest.go`
- `internal/connectors/registryset/registry_gen.go`
- `internal/connectors/catalog_data.json` (source-stripe and source-github entries)
- `internal/cli/{cli.go,connector_docs.go,docs.go,catalog_cli_test.go,cli_test.go}`, `Makefile`

## Actions Taken

Files added:
- `internal/connectors/stripe/stripe.go` — `Connector struct{}`, `New()`, `Name()=="stripe"`,
  `init()→RegisterFactory`, Metadata (Check+Catalog+Read+Write), Check, Catalog, Read,
  `InitialState` (StatefulReader), id-cursor pagination (`harvest`), fixture mode (`readFixture`),
  config/secret resolution, SSRF-bounded `base_url` validation.
- `internal/connectors/stripe/streams.go` — stream defs for customers/charges/invoices/
  subscriptions/products (PK ["id"], cursor ["created"], real Field sets) + per-stream flattening
  mappers + a `stripeStreamEndpoints` routing table (data-driven read path).
- `internal/connectors/stripe/write.go` — allow-list {create_customer, update_customer},
  `ValidateWrite` (rejects unknown), `DryRunWrite` (no network), `Write` (form POST; fixture mode no
  mutation).
- `internal/connectors/stripe/manifest.go` — `Manifest()` documenting config/secret/auth/streams/
  write actions/id-cursor pagination.
- `docs/connectors/stripe/{MANUAL.md,SKILL.md}` (generated).

Files changed:
- `internal/connectors/connsdk/http.go` — added `Requester.DoForm` via a shared `do(...)` core
  (the one connsdk addition; behavior-neutral for JSON callers). `applyHeaders` takes a content type.
- `internal/connectors/connsdk/http_test.go` — two tests for `DoForm`.
- `internal/connectors/registryset/registry_gen.go` — blank import `.../stripe`.
- `internal/connectors/catalog_data.json` — flipped ONLY `source-stripe`:
  `implementation_status: enabled`, `pm_connector_name: "stripe"`, runtime_capabilities
  check/catalog/read/write/etl/reverse_etl=true (query=false, metadata=true), removed
  unsupported_reason, updated native_support_notes; runtime_kind stays `declarative_http_go`.
- `docs/connectors/**` regenerated via `pm docs generate` (catalog json/md, README, stripe +
  source-stripe manuals/skills).

Tests updated to keep invariants accurate (NOT weakened; red-first stripe tests untouched). Flipping
source-stripe to enabled raises the enabled count 1→2 and makes its alias resolve live, so:
- `internal/connectors/catalog_test.go` — Enabled 1→2, PlannedNativePort 646→645.
- `internal/connectors/native_conformance_test.go` — `enabled != 1 → != 2`; source-stripe absent→
  present in registry; "planned" example swapped to `source-strava`.
- `internal/cli/catalog_cli_test.go` — enabled 1→2, planned_native_port 646→645, assert
  source-stripe `pm_connector_name: stripe`.
- `internal/cli/cli_test.go` — planned-rejection example swapped `source-stripe`→`source-strava`.

## Commands Run

- `go test ./internal/connectors/stripe/` (red) → build failed (package absent).
- `go build ./internal/connectors/stripe/ ./internal/connectors/connsdk/` → ok.
- `go test ./internal/connectors/stripe/ -v` → 3/3 PASS.
- `go test ./internal/connectors/ ./internal/cli/ ./...` → all ok.
- `gofmt -l internal` → clean; `go vet ./...` → clean.
- `pm docs generate --dir docs/cli` then `pm docs validate --connectors-dir docs/connectors` → ok.
- `pm connectors inspect stripe --json` / `inspect source-stripe --json` → kind Connector, read+write.
- `make verify` → GREEN.

## Findings

- The native conformance harness builds `NewNativeCatalogConnector(def)` directly (not via registry),
  so the catalog flip is what the existing `internal/connectors` tests assert against; three of those
  tests + two CLI tests encoded the pre-flip "exactly 1 enabled / source-stripe planned" invariant
  and had to be updated to the post-flip reality.
- Stripe id-cursor pagination has no prebuilt connsdk paginator; implemented in-package as planned.
- connsdk JSON `Do` could not send form bodies; added `DoForm` rather than bypassing the SDK, so
  Stripe writes still get connsdk auth/retry. Refactor is behavior-neutral.

## Handoff Summary

`make verify` is GREEN. Stripe is implemented as the connsdk-based reference HTTP connector with
read (Bearer auth, has_more/starting_after pagination over `data[]`, incremental `created[gte]`),
write (allow-listed create/update customer, dry-run preview, form POST), fixture mode, manifest, and
docs. `source-stripe` resolves live via `pm_connector_name=stripe`. One connsdk addition
(`Requester.DoForm`, tested). No new dependencies, no schema/auth/policy changes.

## Verification Evidence

- `go test ./internal/connectors/stripe/ -v`: TestReadPaginatesAndAuthenticates,
  TestWriteValidateAllowList, TestRegisteredWithWriteCapability → PASS.
- `go test ./...`: all packages ok.
- `gofmt -l internal` clean; `go vet ./...` exit 0.
- `make verify`: fmt, vet, test (all ok), build, docs validate ("Validated connector docs in
  docs/connectors"), smoke ("smoke ok: ...") all GREEN.
- `NativeConformanceReports` length stays 647 (== catalog length).
- `pm connectors inspect stripe --json` → kind Connector, check/catalog/read/write true.
- `pm connectors inspect source-stripe --json` → kind Connector (aliases to stripe).

## Unresolved Risks

- None blocking. Optional future work (per ADR): extract a reusable `connsdk.IdCursorPaginator` once
  the has_more/starting_after pattern recurs; expand stripe streams/write actions beyond batch 1.
- Threat-model controls upheld: secret only feeds `connsdk.Bearer` (never logged), no record/PII
  debug logging, base_url validated (scheme+host) to bound SSRF, writes stay
  plan→preview→approve→execute, fixture mode performs no external mutation.
