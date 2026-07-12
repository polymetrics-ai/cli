# TDD Ledger: GitLab CLI Surface Metadata (#83)

## 2026-07-09 — planned red test

### GSD / Skill Evidence

- GSD lane prompt: `scripts/gsd prompt execute-phase issue-83-gitlab-cli-surface --tdd`.
- Manual programming-loop fallback is recorded because `scripts/gsd prompt programming-loop ...` is unavailable in this adapter registry.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-spf13-cobra`.

### Red Target

Add `TestBundleLoadEmbeddedGitLabCLISurface` before creating `internal/connectors/defs/gitlab/cli_surface.json`.

Initial failure captured:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
```

Result: failed as expected.

```text
--- FAIL: TestBundleLoadEmbeddedGitLabCLISurface (0.00s)
    bundle_test.go:933: GitLab CLISurface is nil; defs.FS must embed cli_surface.json
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.549s
```

### Green Target

Create schema-valid `internal/connectors/defs/gitlab/cli_surface.json` where:

- `project list`, `group list`, `user list`, and `issue list` are `intent=etl`, `availability=implemented`, and point to the existing streams.
- Future direct-read, reverse-ETL, local workflow, raw API, binary, and admin/destructive commands are planned/unsupported/unsafe with explicit notes, not executable.
- Examples and notes contain no secret-shaped literals.

### Green Evidence

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
```

Result: passed.

```bash
go test ./cmd/connectorgen ./internal/connectors/engine -run 'CLISurface|GitLab' -count=1
```

Result: passed.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json > /tmp/gitlab-validate.json && jq '{connectors_checked, findings: (.findings|length)}' /tmp/gitlab-validate.json
```

Result: `connectors_checked=547`, `findings=0`.

```bash
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1
```

Result: passed.

### Full Verification Evidence

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs
make verify
```

Result: passed; `make verify` ended with `connectorgen validate: 547 connector(s) checked, 0 findings`.

### Website Evidence

```bash
cd website && pnpm run gen:website-data
```

Result: passed and regenerated `website/data/connectors.generated.json` and `website/lib/connectors.catalog.data.generated.json` with GitLab CLI-surface metadata.

```bash
cd website && pnpm run typecheck
```

Result: blocked because `node_modules` is absent and `tsc` is not installed in the checkout.

```bash
cd website && pnpm install --frozen-lockfile
```

Result: blocked by `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`; did not modify tracked files. No `--no-frozen-lockfile` install was run because dependency/lockfile changes are outside this lane.
