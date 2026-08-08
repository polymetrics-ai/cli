# Verification — MySQL container harness R1

## Fresh post-rebase evidence — 2026-08-08

The branch was rebased onto current `origin/main` before this verification. Generated files were
regenerated from source (`pm docs generate`, website `gen:catalog`, and `connectorgen surface-sync`)
rather than hand-merged. `surface-sync --check` reports 552 connectors and zero fields to fill or
correct.

### Tests and build

- [x] `go test -count=1 -timeout 20m ./internal/connectors/native/dbtest ./internal/connectors/native/mysql ./internal/connectors/native/postgres ./internal/connectors/native/sqltls ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/commandrunner`
- [x] `go test -count=1 -timeout 20m ./internal/cli` — including the existing golden transcript
      coverage; no blind transcript regeneration was performed.
- [x] `go test -count=1 -timeout 20m ./cmd/connectorgen` — including
      `TestGen_NativesetImportsRuntimePackagesAndExcludesSupportLibraries` after regeneration.
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`, `make docs-check-no-build`, `make smoke-no-build`, `make lint`,
      `make agent-contract-check`, `make connectorgen-validate`,
      `make connectorgen-surface-sync`, `make connector-boundary`, and
      `make release-workflow-check`.
- [x] `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — **No vulnerabilities found**.

### Live MySQL proof

Ran the documented tagged test with a new, task-owned 2 CPU / 2 GiB / 8 GiB Podman machine and an
explicit connection/machine name:

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_PODMAN_CONNECTION=<task-owned-machine> \
  POLYMETRICS_PODMAN_MACHINE=<task-owned-machine> \
  POLYMETRICS_DATABASE_RECLAIM_DISK=1 \
  go test -tags=databaseintegration -count=1 -timeout 20m \
    -run '^TestMySQLContainerHarness$' ./internal/connectors/native/mysql
```

- [x] The non-cached live invocation completed successfully; a follow-up identical invocation was
      reported by Go as cached pass.
- [x] It asserted a real check, dynamic catalog, full five-record read with `page_size=2`, exact
      incremental read, live transport-security modes, and insert/update/delete binary-log events.
- [x] The harness logged integer disk-free values before and after teardown and its leak threshold
      passed with reclaim enabled.
- [x] Scoped post-run `ps --all`, `volume ls`, and `images` returned no records.
- [x] The task-owned machine was explicitly stopped and removed immediately after the proof. No
      other Podman machine, container, or volume was touched by this verification.

### Contract reconciliation

- [x] #3902 direct-read page context does not apply to MySQL's ETL `Read`: MySQL has no HTTP/API
      direct-read surface and no `DirectReader` implementation. `TestReadIsETLNotPagewiseDirectRead`
      records the distinction, while the live result proves its internal SQL pages are complete.
- [x] PostgreSQL accepts and enforces the shared SQL TLS options in the pool used by Check, Catalog,
      and Read. `TestDefinitionAcceptsSharedTransportSecurityVocabulary` and
      `TestPostgresPoolConfigUsesSharedTransportSecurityOptions` pass. No PostgreSQL write-path
      source was edited.
- [x] Recovered the old pipeline's native-wiring review finding without discarding its custody
      history: `dbtest` and shared `sqltls` are support libraries, not production connector
      registrations. The generator's source and test were updated, then `connectorgen gen`
      regenerated `nativeset_gen.go`; the focused generator test passed.

### Dependency approval evidence

- [x] Sole direct addition: `github.com/go-mysql-org/go-mysql v1.16.0`, resolved by
      `go mod why` through `internal/connectors/native/mysql` → `client`.
- [x] Upstream release timestamp is 2026-07-15; repository activity and release are linked in the
      PR body. The module's `LICENSE` is MIT.
- [x] Module delta is one direct module plus nine newly listed indirect modules and three indirect
      version bumps; the TiDB parser is the largest new transitive component.
- [x] At dependency introduction, the clean darwin/arm64 `go build -trimpath ./cmd/pm` comparison
      was 94,191,266 bytes before and 100,101,650 bytes after: +5,910,384 bytes (+5.64 MiB,
      +6.27%). The current rebased branch also builds successfully above.

## Inline GSD / review evidence

The inline fallback recorded in `CONTEXT.md` applies: required GSD prompts and contract check were
resolved, the Red/Green ledger is current, and the changed implementation/generation/durability
paths received an inline review in `REVIEW.md`. PR-stage automated review and CI remain to be
driven after opening the PR.
