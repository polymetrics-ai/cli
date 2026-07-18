# Phase 424 Verification

## Required gate checklist

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`
- [ ] `go test ./internal/cli/...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum`

## CLI parity checklist

- [ ] Golden transcript diff empty, or any fixture change explicitly reviewed.
- [ ] `./pm help runtime` checked: exit 0, docs-map canonical help.
- [ ] Bare `./pm runtime` checked: exit 0, same canonical help as `pm help runtime`.
- [ ] `./pm runtime --help` checked: exit 0, docs-map canonical help.
- [ ] JSON manual checked: `./pm runtime --json` exit 0 with `CommandManual` envelope.
- [ ] Invalid action checked: `./pm runtime bogus --json` exit 2, JSON category `usage`.
- [ ] Native doctor semantics checked: `doctor --json`, unknown flags ignored, extra args ignored, late `--json`, late `--root`, and config-file endpoints.
- [ ] Runtime service optionality checked: tests use loopback/config-only endpoints; no Podman/PostgreSQL/DragonflyDB/Temporal startup.
- [ ] Completion metadata/no-file fallback seam preserved; Phase 15 completion implementation explicitly not included.
- [ ] `docs/cli/runtime.md` parity checked by docs-generate-diff/golden docs test; update only if help/output intentionally changes.
- [ ] Website docs/source/generated data checked under `website/**`; update only if user-facing docs intentionally change.
- [ ] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [ ] Runtime-backed integration tests not run unless explicitly requested.
- [ ] No credentialed connector checks.
- [ ] No external services started.
- [ ] No reverse ETL execution beyond repository local temp-dir smoke inside `make verify`.
- [ ] No new dependencies.

## Results

Pending.
