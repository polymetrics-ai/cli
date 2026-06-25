# Agent Trace: backend

## Rendered Prompt Or Prompt Reference

EXECUTE stage for phase `wave0-github-native-package`: make red-first tests pass by
(A) adding a self-registration mechanism to package connectors and (B) migrating the
GitHub connector into its own package `internal/connectors/github`, keeping `make verify`
green. Spec/plan: `.planning/phases/wave0-github-native-package/{PLAN,SPEC,ADR,THREAT-MODEL}.md`.

## Files Inspected

- `internal/connectors/connectors.go`, `manifest.go`, `guide.go`, `catalog.go`,
  `native_catalog_connector.go`, `catalog_data.json` (source-github entry).
- `internal/connectors/github.go`, `github_auth.go`, `github_streams.go` (originals).
- Red tests: `internal/connectors/registry_factory_test.go`,
  `internal/connectors/github/github_test.go`.
- Moving tests: `internal/connectors/github_test.go`, `github_auth_test.go`,
  `github_expanded_test.go`, `manifest_test.go`, `guide_test.go`, `native_port_test.go`.
- Wiring: `internal/app/app.go`, `internal/cli/cli.go`, `internal/app/github_sync_modes_test.go`.
- `Makefile` (verify = fmt vet test build docs-check smoke).

## Actions Taken

Wave A (package connectors):
- Added mutex-guarded, order-preserving `factoryRegistry` + `RegisterFactory`,
  unexported `unregisterFactory` (test cleanup) and `registeredFactories()` in
  `connectors.go`.
- `NewRegistry()`: removed `r.Register(Github{})`; kept built-ins
  Sample/File/Warehouse/Outbox; registers factories AFTER built-ins, BEFORE the
  enabled catalog-alias loop, skipping existing names.

Wave B (github package + wiring):
- Created `internal/connectors/github/` (`package github`):
  - `github.go`: `type Connector struct{Client *http.Client}`, `New() connectors.Connector`,
    `init()` → `connectors.RegisterFactory("github", New)`, all core interface methods,
    `ValidateWrite`, write-action execution, 6 generic helpers + all `githubXxx` helpers,
    all `connectors.*` types qualified.
  - `auth.go`: auth modes + `firstNonEmptyString` + `githubAuthModeSpecs`.
  - `streams.go`: stream/field/record builders + `nestedMap` (Field literals made named).
  - `manifest.go`: relocated `Manifest()` (was in connectors/manifest.go) + local copies
    of unexported `allSyncModes`/`readSourceSyncModes`.
- Created `internal/connectors/registryset/registry_gen.go` (`package registryset`)
  blank-importing github, exposing `New()`.
- Wired `internal/app/app.go:98` and `internal/cli/cli.go:902` to `registryset.New()`.
- Deleted old `github.go`, `github_auth.go`, `github_streams.go`; removed
  `(Github).Manifest()` from `connectors/manifest.go`.
- Moved tests: deleted connectors `github_test.go`, `github_auth_test.go`,
  `github_expanded_test.go`, `manifest_test.go`; created `github/connector_test.go`
  (`package github`), `github/expanded_test.go` (`package github`),
  `github/manifest_test.go` (`package github_test`). Removed the github guide test from
  `connectors/guide_test.go`, kept the generic guide test.

## Commands Run

- `go test ./internal/connectors/ -run TestRegisterFactoryIsResolvedByRegistry` → ok.
- `gofmt -l internal cmd` → clean.
- `go vet ./...` → clean.
- `go test ./...` → all ok.
- `go build -o /tmp/pm-verify ./cmd/pm` then `pm connectors inspect github --json` and
  `... source-github --json` → both `kind=Connector`, caps read+write.
- `make verify` → exit 0 (fmt, vet, test, build, docs-check, smoke all green).

## Findings

- The prompt said github.go had no `Manifest()`; in fact `(Github).Manifest()` lived in
  `internal/connectors/manifest.go`. It was relocated into the github package as
  `(Connector).Manifest()` so `connectors.ManifestOf(github.New())` and
  `RenderConnectorManual` still return the full github manifest/guide.
- Github-specific guide examples/links are keyed off the connector NAME "github" inside
  `connectors/guide.go` (`examplesForManifest`, `linksForManifest`), so they keep firing
  for the migrated connector with no change.
- More test files referenced github than the prompt listed (`github_expanded_test.go`,
  plus the github case in `guide_test.go`); all handled.
- App/CLI tests resolve github through the App/registry (now `registryset.New()`); no
  app/cli test constructs `Github{}` directly, so none needed a blank import.

## Handoff Summary

`make verify` is GREEN (exit 0). Self-registration mechanism added; GitHub migrated to
`internal/connectors/github` as the reference per-system package; wired via the new
`registryset` package. Import cycle avoided (connectors never imports github; only
registryset blank-imports it). Behavior parity preserved: both `github` and
`source-github` resolve with Read+Write, `pm connectors inspect github --json` →
`kind=Connector` with 25 write actions.

## Verification Evidence

```
go vet ./...                       # clean
gofmt -l internal cmd              # clean
go test ./...                      # all packages ok
make verify                        # exit 0
  ./pm docs validate ...           # Validated connector docs in docs/connectors
  smoke ok: /var/folders/.../tmp.9DqvoRvEqM
pm connectors inspect github --json        → kind=Connector caps read+write, 25 actions
pm connectors inspect source-github --json → kind=Connector caps read+write
TestNewContract / TestCatalogStreams / TestRegisteredInRegistry → PASS
TestRegisterFactoryIsResolvedByRegistry → PASS
```

## Unresolved Risks

- None blocking. `connsdk` adoption (optional in SPEC) was intentionally deferred to
  preserve exact behavior and green tests — the proven GitHub HTTP/JWT code was kept
  in-package. Reviewer may opt to refactor onto `connsdk.Requester` later.
- `cmd/pm-cataloggen` was not touched; it does not build a runtime registry. Worth a
  glance if a reviewer expects all registry builders to route through `registryset`.
