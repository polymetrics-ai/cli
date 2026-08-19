# Verification — MySQL container harness R1

## Current endpoint-contract correction — 2026-08-09

The earlier task-owned-machine evidence below applies to a removed lifecycle and is retained only as
historical context. The current harness accepts a direct local Unix endpoint, never changes a Podman
global default, always retains the source image, and rechecks endpoint identity plus target-store
capacity before every command. The outer validation phase must run the tagged MySQL proof against a
direct endpoint after the focused dbtest check recorded in `TDD-LEDGER.md`.

- [x] `go test -count=1 -timeout 5m ./internal/connectors/native/dbtest` passed for the current
      endpoint-contract correction and Podman 5.3 forwarded-Unix capacity handling.
- [x] The tagged live MySQL proof passed against the supplied task-owned direct local Unix forward:
      check, catalog discovery, full/incremental reads, every TLS mode including `verify-ca`, and
      internal row-event CDC all completed; generated container, volume, and run-image cleanup
      completed with equal before/after daemon image-store capacity.

## CI Snyk verify-ca resumption remediation — 2026-08-09

`verify-ca` previously installed its manual certificate-chain check through
`tls.Config.VerifyPeerCertificate`. Go skips that callback for resumed connections when
`InsecureSkipVerify` is needed to omit hostname verification, so this failed Snyk's TLS hardening
gate despite the initial handshake proof. The callback is now `VerifyConnection`, which Go invokes
on all connections, including resumptions.

- [x] Red: the focused `TestVerifyCA...` regression failed before the implementation because no
      connection-wide verifier was installed.
- [x] `TestVerifyCARevalidatesResumedSessions` proves the chain-only verifier runs for both the
      initial and resumed TLS 1.2 handshakes.
- [x] `go test -count=10 -timeout 1m ./internal/connectors/native/sqltls -run '^TestVerifyCARevalidatesResumedSessions$'`
- [x] `go test -count=1 -timeout 5m ./internal/connectors/native/sqltls`
- [x] `go test -count=1 -timeout 20m ./internal/connectors/native/sqltls ./internal/connectors/native/mysql ./internal/connectors/native/postgres`
- [x] `go vet ./internal/connectors/native/sqltls ./internal/connectors/native/mysql ./internal/connectors/native/postgres`
- [x] `govulncheck ./internal/connectors/native/sqltls ./internal/connectors/native/mysql ./internal/connectors/native/postgres` — **No vulnerabilities found.**

The outer CI phase must re-run Snyk against its pushed commit; this local phase intentionally did
not push or control the gate.

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

### Live MySQL proof — re-run against the ownership-gated reclaim (review round 3)

The proof above predates the review fix that replaced the name-matching reclaim gate with an
ownership record, so it was re-run end to end against that gate. The machine is now created by the
test infrastructure itself, which is the only thing that establishes ownership:

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_DATABASE_OWN_MACHINE=1 \
  go test -tags=databaseintegration -count=1 -v -timeout 30m \
    -run '^TestMySQLContainerHarness$' ./internal/connectors/native/mysql
```

```text
mysql_integration_test.go:81: MySQL database test created task-owned Podman machine "pmdb-mysql-eddc7350dec5"
mysql_integration_test.go:120: MySQL database test disk free bytes: before=18050224128 after=22241910784 reclaimed=true
--- PASS: TestMySQLContainerHarness (80.93s)
ok      polymetrics.ai/internal/connectors/native/mysql  81.690s
```

- [x] The run created its own uniquely named 2 CPU / 2 GiB / 16 GiB machine
      (`pmdb-mysql-eddc7350dec5`) rather than being handed one.
- [x] `reclaimed=true`: the ownership gate permitted the trim, so this is observed rather than
      inferred. Host free went from 18,050,224,128 to 22,241,910,784 bytes across teardown — the
      trim returned more than the run consumed, and the leak assertion was evaluated rather than
      skipped.
- [x] All four `sslmode` subtests, the reads, and the binary-log CDC assertions passed on the
      task-owned machine.
- [x] `podman machine list` afterwards showed no `pmdb-` machine and the two pre-existing machines
      (`polymetrics-runtime`, `fm-bahmni-lab-r1-machine`) still present and stopped. Nothing else
      was touched.

One disclosed residue: `podman machine rm --force` leaves a 128 KiB
`efi-bl-<machine>` file in Podman's own machine state directory. This is Podman-side behaviour, not
harness behaviour — the same file exists for machines this repository never created
(`efi-bl-podman-machine-default`) — and the harness deliberately does not reach into Podman's state
directory to delete files, because guessing at that layout risks deleting another machine's state
for 128 KiB.

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
