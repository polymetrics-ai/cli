# Phase 423 Verification

## Required gate checklist

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/cli/... -run 'Perf|CobraRouterShell|Golden' -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum`

## CLI parity checklist

- [ ] Golden transcript diff empty or reviewed intentionally.
- [ ] `./pm help perf` checked: exit 0, docs-map canonical help.
- [ ] Bare `./pm perf` checked: exit 0, byte-identical to `pm help perf`.
- [ ] `./pm perf --help` checked: exit 0, byte-identical to docs-map help.
- [ ] JSON manual checked: `./pm perf --json` exit 0 with `CommandManual` envelope.
- [ ] Invalid action checked: `./pm perf bogus --json` exit 2, JSON category `usage`.
- [ ] Native flag semantics checked: `compare --iterations`, `compare --runtime`, `sync-modes --records`; space/equals forms, repeated scalar last-wins, bare bool/value sentinels, unknown flags, extra args, and late `--root`/`--json`.
- [ ] Runtime compare config use checked with loopback endpoints only; no services started.
- [ ] Completion metadata preserved; Phase 15 completion implementation explicitly not included.
- [ ] `docs/cli/perf.md` parity checked by docs-generate-diff/golden docs test; update only if help changes.
- [ ] Website docs/source/generated data checked under `website/**`; update only if generated docs change.
- [ ] Generated help/manual artifacts checked via existing generator/docs validation.

## Optional / safety-limited

- [ ] Runtime-backed integration tests not run unless explicitly requested; no services started.
- [ ] No credentialed connector checks.
- [ ] No external services started.
- [ ] No reverse ETL execution beyond repository local temp-dir smoke inside `make verify`.
- [ ] No new dependencies.

## Results

Pending.
