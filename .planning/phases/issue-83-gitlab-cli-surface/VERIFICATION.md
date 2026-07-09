# Verification: GitLab CLI Surface Metadata (#83)

## Red Command

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
```

Result: failed as expected before `internal/connectors/defs/gitlab/cli_surface.json` existed.

```text
--- FAIL: TestBundleLoadEmbeddedGitLabCLISurface (0.00s)
    bundle_test.go:933: GitLab CLISurface is nil; defs.FS must embed cli_surface.json
FAIL
FAIL	polymetrics.ai/internal/connectors/engine	0.549s
```

## Focused Green Commands

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
go test ./cmd/connectorgen ./internal/connectors/engine -run 'CLISurface|GitLab' -count=1
go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner -count=1
go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json > /tmp/gitlab-validate.json && jq '{connectors_checked, findings: (.findings|length)}' /tmp/gitlab-validate.json
```

Result: passed. `connectorgen validate` reported `connectors_checked=547`, `findings=0`.

## Full Go / Connector Gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs
make verify
```

Result: passed. `make verify` completed gofmt, tidy-check, vet, full tests, build, docs validate, smoke, lint, and connector definition validation.

## Website / Generated Data

```bash
cd website && pnpm run gen:website-data
```

Result: passed; regenerated website connector data includes GitLab `cli_surface` metadata.

```bash
cd website && pnpm run typecheck
```

Result: blocked; local checkout has no `website/node_modules`, so `tsc` is not available.

```bash
cd website && pnpm install --frozen-lockfile
```

Result: blocked by `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`. No `--no-frozen-lockfile` install was run because dependency/lockfile updates are outside #83 and require separate review.

## CLI Help / Docs / Website Parity

- #83 adds CLI-surface metadata and a docs.md metadata-only note.
- Runtime help/manual behavior remains a #84/#85 follow-up; this slice does not claim direct reads, writes, local workflows, or binary downloads are executable.
- Website generated data was refreshed; typecheck/build remain blocked locally by dependency installation state above.
