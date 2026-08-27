# Verification — #4366

## Planned checks

- `go test -timeout 20m ./cmd/connectorgen -run 'Test.*(Composition|SourceProjection|SourceImport|DeclarationAdmission|OperationEvidence)' -count=1`
- `go test -timeout 20m ./internal/connectors/engine -run 'Test.*(Schema|Structured|Deferred|Operation)' -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -run 'Test.*(Preflight|Deferred|Structured)' -count=1`
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go run ./cmd/connectorgen source-import --check <touched connectors>` or the repository-supported equivalent
- `go run ./cmd/connectorgen validate <touched connectors>` and `go run ./cmd/connectorgen surface-sync --check <touched connectors>` or the supported all-connector checks
- generated operation-evidence/declaration-admission/endpoint-ledger checks
- `go build ./cmd/pm`
- isolated-project credential-boundary checks for every newly runnable command; no credential or live call
- `gofmt -w` on changed Go files, `go vet ./...`, relevant direct-PR verification gates, and `git diff --check`

## CLI parity assessment

No generic CLI surface was added and no Batch 1 command was promoted. The
source-locked GitHub projection updated two already-implemented
`billing *-budget-org` `--budget-type` flags from an over-broad string to the
closed JSON composition contract; this is a schema correction, not a new
command or lane claim. The generated GitHub manual/skill and website catalog
were refreshed.

- `./pm help github`, `./pm github`, and both changed command `--help` calls
  succeeded.
- `./pm docs generate --dir docs/cli` then
  `./pm docs validate --connectors-dir docs/connectors` passed.
- `cd website && npm run gen:website-data` and `npm run test:scripts` passed.
- `docs/cli/**` had no generated delta; `docs/connectors/github/{MANUAL,SKILL}.md`
  and the two generated website catalog data files carry the exact flag delta.

## Results

### Red / green / reconciliation

- The pre-change red run is retained at `traces/red-composition.txt`.
- Focused green suite passed for source import/projection, engine composition,
  commandrunner preflight, deferred no-artifact behavior, and the 608-row
  fixture.
- `go test -timeout 20m ./cmd/connectorgen -count=1` passed in `215.464s`.
- `go test -timeout 20m ./internal/connectors/engine -count=1` passed in
  `10.729s`.
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1` passed
  in `22.267s`.
- Reconciliation result: **608 / 608 retained as exact
  `missing_foundation` rows; 0 credential-bound runnable commands**. The
  fixture asserts source URL/SHA/location, method/path, canonical target,
  command identity, six-lane classification, and provider totals (Bitbucket
  30, CircleCI 3, GitLab 542, Jira 1, Vercel 32).
- No isolated credential-bound invocation was applicable: the runnable count
  is zero, and no provider credential or live call was used.

### Source / generated contracts

- `go run ./cmd/connectorgen source-import github --defs internal/connectors/defs --out internal/connectors/defs/github/sources/github-operation-descriptor.json --check` passed.
- The five historical Batch 1 providers have no current
  `internal/connectors/defs/<name>/sources/` directory in this checkout, so
  `source-import --check` cannot be truthfully invoked for them. Their
  retained provider evidence is the pinned Batch 1 manifest fixture; the
  608-row reconciliation is the applicable source-cited check.
- `go run ./cmd/connectorgen validate internal/connectors/defs` passed:
  553 connectors, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check` passed:
  553 connectors, 0 corrections.
- `go run ./cmd/connectorgen declaration-admission`,
  `operation-evidence --check`, `certification-subject --check`,
  `certification-matrix --check`, GitHub certification candidates/sweep
  checks, and `boundary . --json` all passed.

### Direct-PR gates

- `go mod tidy` plus `git diff --exit-code -- go.mod go.sum`, `go vet ./...`,
  `go build ./cmd/pm`, `golangci-lint` (applicable packages), and
  `git diff --check` passed.
- `./pm docs validate`, the local smoke flow, GitHub ledger/parity/source-drift
  Node tests, connector canon, pinned build dependency, release target/tooling,
  release size/layout, and installed GitHub certification checks all passed.
- Repository-wide `go test ./...` / `make verify` were not launched as one
  command because this agent's per-command timeout is shorter than the
  repository's 550+ connector suite; the repository instruction directs that
  CI carry that broad aggregate after the focused packages and individual
  `make verify` gates above pass.
