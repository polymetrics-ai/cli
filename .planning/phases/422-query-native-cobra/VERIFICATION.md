# Phase 422 Verification

## Required gate checklist

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/cli/... -run 'Query|CobraRouterShell|Golden' -count=1`
- [ ] `go test ./internal/cli/ -run Certify -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum`

## CLI parity checklist

- [ ] Golden transcript diff empty or reviewed.
- [ ] `./pm help query` checked: exit 0, docs-map canonical help.
- [ ] Bare `./pm query` checked: exit 0, byte-identical to `pm help query`.
- [ ] `./pm query --help` checked: exit 0, byte-identical to docs-map help.
- [ ] JSON manual checked: `./pm query --json` exit 0 with `CommandManual` envelope.
- [ ] Invalid action checked: `./pm query bogus --json` exit 2, JSON category `usage`, no project open.
- [ ] Query JSON output checked with local temp fixture only.
- [ ] Read-only SQL rejection checked; no generic SQL write exposed or validation weakened.
- [ ] Native flag semantics checked: `--table`, `--sql`, `--limit`, `--fields`, `--agent-mode`, `--sample`; space/equals forms, repeated scalar last-wins, repeated/comma `--fields`, bare bool sentinels, unknown flags, extra args, and late `--root`/`--json`.
- [ ] Completion metadata preserved; Phase 15 completion implementation explicitly not included.
- [ ] `docs/cli/query.md` parity checked by docs-generate-diff/golden docs test; update only if help intentionally changes.
- [ ] Website docs/source/generated data checked under `website/**`; update only if help/output intentionally changes.
- [ ] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [ ] Runtime-backed integration tests not run unless explicitly requested.
- [ ] No credentialed connector checks.
- [ ] No external services started.
- [ ] No reverse ETL execution beyond repository local temp-dir smoke inside `make verify`.
- [ ] No new dependencies.

## Results

Pending.
