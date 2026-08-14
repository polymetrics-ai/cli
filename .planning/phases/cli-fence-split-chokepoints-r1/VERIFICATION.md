# Verification checklist — Fence split chokepoints r1

## Behaviour-preservation checks

- [x] GitHub and PostgreSQL shard union reconstructs their exact in-memory aggregate payload.
- [x] A normal scoped generation run leaves the other allowlisted shard and shared status artifact
      byte-identical.
- [x] All generated anchors name source symbols and no `:<digits>` source anchor remains.
- [x] `connectorgen certification-matrix --check` detects allowlisted shard drift and source
      disappearance while ignoring non-allowlisted certification claims, including a malformed
      real `mysql` runtime-ledger record that must not re-enter the native PostgreSQL factory.
- [x] A one-line shared-file insertion above an anchor produces zero generated shard changes after
      the refactor (baseline old-matrix diff count is recorded below).
- [x] Existing app Open/ETL tests and PostgreSQL metadata/manifest tests pass unchanged.
- [x] No capability values, error text, CLI output, or connector definition semantics change.

## Local commands

- [x] `go test -count=1 ./cmd/connectorgen`
- [x] `go test -count=1 ./internal/app`
- [x] `go test -count=1 ./internal/connectors/native/postgres`
- [x] `go test -count=1 ./internal/connectors/certifications`
- [x] `go run ./cmd/connectorgen certification-matrix --all`
- [x] `go run ./cmd/connectorgen certification-matrix --check`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connectorgen-certification-matrix`
- [x] `make connector-boundary`
- [x] `make connector-canon-check`
- [x] `make release-workflow-check`
- [ ] no-mistakes pipeline returns `checks-passed`.

## Generator-isolation correction — 2026-08-14

- [x] **Red:** command-level regression reproduces the pre-fix panic by corrupting only the
      non-allowlisted `mysql` entry in the real generator ledger.
- [x] **Green:** the same real `certification-matrix --check` path succeeds without a native
      PostgreSQL factory reload.
- [x] `git diff --exit-code 2df18ee3a083fe507cbe1c07e0270e82c5ab0182 --
      internal/connectors/engine/bundle.go` confirms production loader byte identity.
- [ ] Full `./cmd/connectorgen` and `./internal/connectors/engine` suites pass, followed by the
      new no-mistakes run to CI green.

## CI repair — 2026-08-14

- [x] `GOTOOLCHAIN=go1.25.12 go test -count=1 -timeout 20m -run '^TestCurrentRepositoryBaselinePasses$' ./internal/connectors/boundary`
- [x] `GOTOOLCHAIN=go1.25.12 go run ./cmd/connectorgen boundary . --json`
- [x] `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...`

## Required report measurements

- Database test package on base: absent (`internal/connectors/database` does not exist on
  `2df18ee`).
- Old one-line-insertion capability-matrix diff: one changed generated line (`1` addition / `1`
  deletion) at the moved `binary_download` executor anchor in the shared 6,632,701-byte file.
- Post-change insertion result: zero generated shard/status changes; the GitHub, PostgreSQL, and
  status SHA-256 values were identical before and after the inserted source comment.
- Shards produced: 2. Largest shard: GitHub, 37,118 bytes. PostgreSQL: 30,940 bytes. The compact
  status projection is 525 bytes; total checked-in certification surface: 68,583 bytes.
- Aggregate consumers: only `cmd/connectorgen` produced/checked the two retired matrices and the
  Make target called that generator; runtime `internal/connectors/certifications/status.go` consumes
  only `status.json`. No website, docs build, or CLI consumer reads either retired aggregate.
- Flow share: direct GitHub/PostgreSQL flow rows were 23,562 / 10,692,336 bytes (0.22%). The old
  compressed global pair-set records mentioning either connector occupied 317,373 additional bytes;
  their non-exclusive union was 340,935 bytes (3.19%), demonstrating why those global sets are now
  reconstructed in memory rather than committed in either shard.
