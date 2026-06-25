# SUMMARY — Wave 0: GitHub Native Package + Data-Driven Registry

Status: **completed** (GO). `make verify` green; reviewer GO; TDD red-before-code enforced.

## What shipped
1. **Self-registration mechanism** (`internal/connectors/connectors.go`): a mutex-guarded,
   order-preserving factory store with `RegisterFactory(name, func() Connector)` +
   `unregisterFactory` (test helper). `NewRegistry()` now registers built-ins → factories →
   catalog-alias loop (in that order, so `source-github` resolves to the live target), and no
   longer hard-codes `r.Register(Github{})`.
2. **GitHub as the reference per-system package** (`internal/connectors/github/`,
   `package github`): github.go, auth.go, streams.go, manifest.go + tests. Exposes
   `New() connectors.Connector`; `Name()=="github"`; `init()` calls
   `connectors.RegisterFactory("github", New)`. The 6 generic helpers + all githubXxx helpers
   moved with it; `connectors.*` types qualified.
3. **Cycle-free wiring** (`internal/connectors/registryset/registry_gen.go`): the single place that
   blank-imports connector packages and returns `connectors.NewRegistry()`. `internal/app/app.go`
   and `internal/cli/cli.go` build the registry via `registryset.New()`. This file is the future
   codegen target — adding a connector = one blank-import line. (Avoids the import cycle: package
   connectors never imports a connector package.)
4. Old flat `github*.go` + their in-package tests deleted; github tests relocated into the package.

## Verification
- `make verify` exit 0 (gofmt, vet, `go test ./...` = 10 pkgs ok, build, docs-check, smoke).
- Parity: `github` and `source-github` both → kind "Connector", read+write, 25 write actions.
- Secret safety, auth (incl. GitHub App RS256 JWT), write allow-list: unchanged (reviewer-confirmed).
- TDD: 3 behavior tasks red-confirmed in TDD-LEDGER before code; gate passed.

## Boundary / deferred
- **connsdk adoption deferred** for github (SPEC-permitted): the proven in-package HTTP/JWT code was
  kept to guarantee parity. A later pass can refactor onto `connsdk.Requester`/`LinkHeaderPaginator`.
- **DuckDB warehouse + real Querier (Wave 0 item 2): NOT done here** — it requires a third-party
  dependency + CGO, which is a human-approval gate. Tracked as a separate gated step.

## Next
- Human gate: approve DuckDB dependency + CGO build-tag approach, then run that as its own phase.
- Then Wave 1 (GA connectors) using this github package as the template + registryset codegen.
