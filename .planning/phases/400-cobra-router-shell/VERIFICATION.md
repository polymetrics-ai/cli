# Issue 400 Verification Checklist — Cobra Router Shell

**Issue:** #400
**Branch:** `refactor/400-cobra-router-shell`
**Base:** `feat/cli-architecture-v2`

## Required gates

- [x] `gofmt -w cmd internal` — pass.
- [x] `go test ./internal/cli/ -run Golden -count=1` — pass: `ok  	polymetrics.ai/internal/cli	6.369s`; post-refactor combined router/golden gate pass: `ok  	polymetrics.ai/internal/cli	6.374s`.
- [x] `go test ./internal/cli/ -run Certify -count=1` — pass: `ok  	polymetrics.ai/internal/cli	95.783s`; post-refactor pass: `ok  	polymetrics.ai/internal/cli	91.679s`.
- [x] `go test ./internal/cli/ -count=1` — pass: `ok  	polymetrics.ai/internal/cli	154.218s`.
- [x] `go vet ./...` — pass, no output.
- [x] `go test ./...` — pass; includes `internal/cli` and connector/certify packages (`internal/connectors/certify` took `341.683s`).
- [x] `go build ./cmd/pm` — pass, no output.
- [x] `make verify` — pass after staging the approved `go.mod`/`go.sum` dependency delta so tidy-check compares post-tidy files to the index. Earlier pre-staging run failed at tidy-check only because dependency files were uncommitted/un-staged; no gate weakened.
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass after implementation commit; no output, exit 0.
- [x] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` — pass/recorded expected dependency delta: Cobra v1.10.2 direct, pflag/mousetrap indirect, go.sum metadata-only checksums.

## Focused TDD gates

- [x] Red validation before implementation: `go list -deps ./internal/cli | grep '^github.com/spf13/cobra$'` failed with no output, exit 1.
- [x] Red focused test before implementation: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` failed with missing Cobra module.
- [x] Green focused test after implementation: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` passed: `ok  	polymetrics.ai/internal/cli	0.532s`; fallback help interception coverage added and passed in combined `TestCobraRouterShell|Golden` gate.
- [x] Golden suite byte-identical: no golden fixture update.
- [x] Certify in-process re-entrancy gate stays green.

## Dependency gate

- [x] Direct dependency added exactly: `github.com/spf13/cobra v1.10.2`.
- [x] Expected indirect go.mod transitives: `github.com/spf13/pflag v1.0.9`, `github.com/inconshreveable/mousetrap v1.1.0`.
- [x] No additional direct dependency introduced.
- [x] `make tidy-check` remained green inside `make verify` after staging `go.mod` / `go.sum`.
- [x] go.sum note: additional `go.mod` checksum lines for Cobra's module metadata (`go-md2man`, `blackfriday`, `go.yaml.in/yaml/v3`) are not direct or imported dependencies; `go mod why -m` says the main module does not need them.

## CLI help / docs / website parity

Applies: yes, because routing/help behavior is CLI-visible even though target is byte-identical.

- [x] Runtime help: `/tmp/pm-400 help connectors` exited 0; stdout 4967 bytes; stderr 0 bytes.
- [x] Bare namespace: `/tmp/pm-400 connectors` exited 0; stdout 4967 bytes; stderr 0 bytes.
- [x] Command help: `/tmp/pm-400 docs --help` exited 0; stdout 818 bytes; stderr 0 bytes.
- [x] Hidden command help: `/tmp/pm-400 worker --help --json` preserved current golden behavior: exit 1; stdout 185 bytes; stderr 37 bytes.
- [x] `docs/cli/**`: not updated; exact reason: no command, flag, help text, or generated docs changed and `TestGoldenDocsGenerateMatchesTrackedCLIManuals` stayed green through `go test ./internal/cli/ -run Golden -count=1`.
- [x] `website/**`: not updated; exact reason: no user-facing CLI help/docs text or command surface changed; router shell is byte-identical by golden suite.
- [x] Generated help/manual artifacts: no update; golden suite remained byte-identical.
- [x] Completion/discovery metadata: not applicable in Phase 2; shell completion is Phase 15.

## Review-fix verification plan — PR #440

Required exact gates after fixes:

- [x] `gofmt -w cmd internal` — pass, no output.
- [x] `go test ./internal/cli/ -run 'TestCobraRouterShell|Golden' -count=1` — pass: `ok  	polymetrics.ai/internal/cli	7.724s`.
- [x] `go test ./internal/cli/ -run Certify -count=1` — pass: `ok  	polymetrics.ai/internal/cli	92.156s`.
- [x] `go test ./internal/cli/ -count=1` — pass: `ok  	polymetrics.ai/internal/cli	155.648s`.
- [x] `go vet ./...` — pass, no output.
- [x] `go test ./...` — pass; notable uncached packages: `polymetrics.ai/internal/cli 162.510s`, `polymetrics.ai/internal/connectors/certify 347.398s`.
- [x] `go build ./cmd/pm` — pass, no output.
- [x] `make verify` — pass; ended with `connectorgen validate: 547 connector(s) checked, 0 findings`.
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output.
- [x] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` — pass/recorded unchanged approved Cobra delta only.

Review-fix targeted assertions:

- [x] Legacy handler/root-fallback plain errors containing `unknown flag` or `unknown command` remain legacy/internal, not usage/exit 2.
- [x] Genuine Cobra parse errors map to existing usage taxonomy through `writeError`.
- [x] Fresh roots define persistent `--root` and `--json` with defaults from parsed invocation state.
- [x] Fresh command trees do not share persistent flag state.
- [x] Every registered top-level wrapper, including `init`, has `DisableFlagParsing` and expected visibility.
- [x] Deterministic dynamic connector passthrough accepts connector-specific flags plus late `--root`/`--json` globals without live credentials.

## Review route

- [x] Open non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #400` and `Refs #397`: PR #440.
- [x] Do not post `@claude review` because repository Claude workflow is already `disabled_manually`.
- [x] Do not request Copilot because quota is exhausted for this blocker window.
- [x] Record review status as human/parent-PR fallback pending; no approval claims.

## Review-fix CLI parity spot checks

- [x] `/tmp/pm-400 help connectors` — exit 0; stdout 4967 bytes; stderr 0 bytes.
- [x] `/tmp/pm-400 connectors` — exit 0; stdout 4967 bytes; stderr 0 bytes.
- [x] `/tmp/pm-400 docs --help` — exit 0; stdout 818 bytes; stderr 0 bytes.
- [x] `docs/cli/**` not updated: no help text/command/flag output changed; golden/docs generation tests remained green.
- [x] `website/**` not updated: no user-facing help/docs text changed.
- [x] Generated help/manual artifacts not updated: golden suite remained byte-identical.

## Full `make verify` result

Review-fix final run passed:

```text
gofmt -w cmd internal
go mod tidy
git diff --exit-code -- go.mod go.sum
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
Validated connector docs in docs/connectors
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.3VKyH34MCu
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/...
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
```
