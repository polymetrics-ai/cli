# Phase 421 Verification

## Required gate checklist

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1`
- [ ] `go test ./internal/cli/ -run Certify -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum`

## CLI parity checklist

- [ ] Golden transcript diff empty or reviewed with explicit rationale.
- [ ] `./pm help connections` checked: exit 0, docs-map canonical help.
- [ ] Bare `./pm connections` checked: exit 0, byte-identical to `pm help connections` unless intentionally reviewed.
- [ ] `./pm connections --help` checked: exit 0, byte-identical to docs-map help unless intentionally reviewed.
- [ ] JSON manual checked: `./pm connections --json` exit 0 with `CommandManual` envelope.
- [ ] Invalid action checked: `./pm connections bogus --json` exit 2, JSON category `usage`.
- [ ] Native flag semantics checked: `--source`, `--destination`, `--stream`, `--sync-mode`, `--cursor`, `--table`, `--source-config`, `--destination-config`, `--primary-key`; space/equals forms, repeated singleton last-wins, repeated primary key accumulation, bare bool values, unknown flags, extra args, and late `--root`/`--json`.
- [ ] Completion compatibility seam preserved; Phase 15 implementation explicitly not included.
- [ ] `docs/cli/connections.md` parity checked by docs-generate-diff/golden docs test; no update needed unless help changes.
- [ ] Website docs/source/generated data checked under `website/**`; no update needed unless output/docs changes.
- [ ] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [ ] Runtime-backed integration tests not run; no services started.
- [ ] No credentialed connector checks.
- [ ] No reverse ETL external execution beyond repository local temp-dir smoke in `make verify`.
- [ ] No new dependencies.

## Results

Pending.
