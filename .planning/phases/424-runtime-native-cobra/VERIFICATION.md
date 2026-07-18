# Phase 424 Verification

## Positional-help correction checklist

Worker: session `7050f706-72d2-47df-ac13-0b08979cc1ae`; model `openai-codex/gpt-5.6-sol`; thinking `high`; starting HEAD `8d696cd4c27fad6840e905917e7658e785fa5436`.

- [x] Focused RED captured for `runtime help` and `runtime help --json` before production edits: both exited 2 as unknown commands (`go test ./internal/cli/ -run '^TestRuntimeBareHelpAndInvalidActionSemantics$' -count=1`, expected failure).
- [x] `gofmt -w internal/cli/cobra_router.go internal/cli/runtime_cli_test.go`
- [x] `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1` — pass (`17.302s`).
- [x] `go test ./internal/runtimecheck/... -count=1` — pass (`0.445s`).
- [x] Built binary: `pm help runtime`, `pm runtime help`, `pm runtime help --json`, and invalid `pm runtime bogus --json` parity passed.
- [x] Existing bare namespace and flag help remain green: `pm runtime`, `pm runtime --help`, `pm runtime --json`.
- [x] Invalid runtime action exits 2 with nested usage category and no `CommandManual`.
- [x] Golden/docs/manual/website artifacts unchanged; no delta applicable because canonical help content did not change.
- [x] `go test ./internal/cli/... -count=1` — pass (`192.383s`).
- [x] `go vet ./internal/cli/... ./internal/runtimecheck/...` — pass, no output.
- [x] `go build ./cmd/pm` — pass, no output.
- [x] Full gates: `gofmt -w cmd internal`; `go vet ./...`; `go test -timeout 20m ./...`; `go build ./cmd/pm`; `make verify` — pass.
- [x] `git diff --check` — pass, no output.
- [x] `git diff -- go.mod go.sum` empty; no new dependencies.
- [x] Runtime-backed checks not run; no services or credentials used. `make verify` used only its existing local temp-dir smoke.
- [ ] Coherent green correction committed and existing branch pushed; no new PR/review request.

Correction result: the hidden native positional alias restores canonical runtime text/JSON help. Built-binary output reported `text_alias_match=true`, `json_alias=CommandManual/runtime`, and `invalid_action_exit=2 category=usage command_manual=False`. Full `make verify` passed with `connectorgen validate: 547 connector(s) checked, 0 findings`. No runtime-backed integration test was run.

One verification-script-only attempt initially asserted the JSON usage category at the top level; it failed because the established envelope stores it under `error.category`. The corrected parity script passed, and no product change was needed for that script correction.

## Required gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/cli/...`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum`

## Review-fix gate checklist

- [x] `gofmt -w cmd internal`
- [x] `go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1`
- [x] `go test ./internal/runtimecheck/... -count=1`
- [x] `go vet ./...`
- [x] `go test -timeout 20m ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check`
- [x] `git diff -- go.mod go.sum`

## CLI parity checklist

- [x] Golden transcript diff empty; no fixture changes.
- [x] `./pm help runtime` checked: exit 0, docs-map canonical help.
- [x] Bare `./pm runtime` checked: exit 0, same canonical help as `pm help runtime`.
- [x] `./pm runtime --help` checked: exit 0, docs-map canonical help.
- [x] JSON manual checked: `./pm runtime --json` exit 0 with `CommandManual` envelope.
- [x] Invalid action checked: `./pm runtime bogus --json` exit 2, JSON category `usage`.
- [x] Native doctor semantics checked: `doctor --json`, unknown flags ignored, extra args ignored, late `--json`, late `--root`, and config-file endpoints.
- [x] Runtime service optionality checked: tests use loopback/config-only endpoints; no Podman/PostgreSQL/DragonflyDB/Temporal startup.
- [x] Completion metadata/no-file fallback seam preserved; Phase 15 completion implementation explicitly not included.
- [x] `docs/cli/runtime.md` parity checked by docs-generate-diff; no update needed because help unchanged.
- [x] Website docs/source/generated data checked under `website/**`; no update needed because generated docs unchanged.
- [x] Generated help/manual artifacts checked via existing generator/docs validation.

## Review-fix CLI/docs/security checklist

- [x] `runtime doctor` help clarifies no live services or credentials are required.
- [x] `runtime doctor` help/docs clarify absent optional services are reported in output status/degraded mode and do not cause runtime-error exit.
- [x] `docs/cli/runtime.md` regenerated/updated with embedded help text.
- [x] `website/content/docs/architecture.mdx` and `website/content/docs/cli-reference.mdx` updated for runtime doctor optional-service semantics.
- [x] Website generated docs data refreshed if generator output changes.
- [x] Golden transcripts updated/reviewed for intentional runtime help wording change.
- [x] DragonflyDB and Temporal reported endpoints are sanitized/redacted for userinfo/query/fragment/control chars.
- [x] PR body records `internal/worker/podman_cmd.go` / `internal/worker/submit.go` high finding as out-of-scope because PR #460 does not change those files.

## Optional / safety-limited

- [x] Runtime-backed integration tests not run; no services started.
- [x] No credentialed connector checks.
- [x] No external services started.
- [x] No reverse ETL execution beyond repository local temp-dir smoke inside `make verify`.
- [x] No new dependencies.

## Results

```bash
gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/runtime_cli_test.go
go test ./internal/cli/ -run 'Runtime|CobraRouterShell' -count=1
```

Result: pass (`ok  \tpolymetrics.ai/internal/cli\t11.749s`).

```bash
go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1
go test ./internal/cli/...
go vet ./...
go build ./cmd/pm
```

Result: pass. Focused/golden passed (`ok  \tpolymetrics.ai/internal/cli\t17.329s`); full internal CLI package passed (`ok  \tpolymetrics.ai/internal/cli\t195.015s`); `go vet` and `go build` emitted no output.

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Result: pass. `go test ./...` passed; slow packages included `internal/connectors/certify 342.381s`. `make verify` passed and ended with `connectorgen validate: 547 connector(s) checked, 0 findings`; `gofmt`, `go vet`, and `go build` emitted no output.

```bash
./pm help runtime
./pm runtime
./pm runtime --help
./pm runtime --json
./pm runtime bogus --json
./pm runtime doctor --unknown ignored extra --root "$root" --json
```

Result: pass. Help/bare/`--help` byte-identical (`help bytes=470`); JSON manual emitted `CommandManual` for `runtime` (`600` bytes); invalid action exited 2 with usage JSON and stderr `error: unknown command "bogus" for "pm runtime"`; loopback doctor emitted `RuntimeDoctor`, redacted PostgreSQL endpoint, Dragonfly/Temporal topology endpoints, all statuses `error`, stderr_bytes=0, and secret_leak=0.

```bash
./pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"
diff -ru docs/cli "$TMP_DOCS/cli"
./pm docs validate --connectors-dir docs/connectors
npm --prefix website run gen:docs
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff -- go.mod go.sum
```

Result: pass. Docs generated to temp dir; docs diff emitted no output; docs validate passed; website docs generator wrote 11 docs pages; diff-check and go.mod/go.sum diff emitted no output.

Sub-PR opened: https://github.com/polymetrics-ai/cli/pull/460. Remote checks/review pending at artifact update.

## Review-fix results

```bash
go test ./internal/cli/... -run 'Runtime|CobraRouterShell|Golden' -count=1
go test ./internal/runtimecheck/... -count=1
```

Result: pass (`ok  \tpolymetrics.ai/internal/cli\t17.208s`; `ok  \tpolymetrics.ai/internal/runtimecheck\t0.364s`).

```bash
gofmt -w cmd internal
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
make verify
```

Result: pass. `go test -timeout 20m ./...` passed; slow packages included `internal/cli 222.610s`, `internal/connectors/certify 369.637s`, and `internal/runtimecheck 2.323s`. `make verify` passed and ended with `connectorgen validate: 547 connector(s) checked, 0 findings`; `gofmt`, `go vet`, and `go build` emitted no output.

```bash
npm --prefix website run gen:docs
TMP_DOCS=$(mktemp -d); ./pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"; diff -ru docs/cli "$TMP_DOCS/cli"; ./pm docs validate --connectors-dir docs/connectors
git diff --check
git diff -- go.mod go.sum
```

Result: pass. Website generator wrote 11 docs pages; docs generate/validate passed; `diff -ru`, `git diff --check`, and go.mod/go.sum diff emitted no output.

```bash
./pm runtime --root
./pm runtime doctor --root
./pm --json runtime --root
./pm --root "$ROOT" --json runtime doctor
./pm help runtime
./pm runtime
./pm runtime --help
```

Result: pass. Malformed known `--root` flags exited 2 with usage JSON when JSON was requested; runtime doctor with absent services emitted degraded status and sanitized endpoints; help/bare/`--help` stayed in parity and documented optional-service semantics.
