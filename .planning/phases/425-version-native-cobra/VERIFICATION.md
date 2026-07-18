# Phase 425 Verification

Invocation session: `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc`.

## TDD checklist

- [x] Planning artifacts created before production edits.
- [x] Exact focused RED captured before production edits (`0.612s`; native registration assertions failed as intended).
- [x] Smallest native Cobra version implementation green (`0.553s`).
- [x] Focused version/router/golden tests green (`7.814s`); no golden fixture delta.

## Focused behavior checklist

- [x] Version is a native Cobra top-level command and absent from legacy wrappers.
- [x] `pm version` deterministic plain output and exit 0 (in-process focused test).
- [x] `pm version --json` deterministic `Version` envelope and exit 0 (in-process focused test).
- [x] `pm help version`, `pm version --help`, and `pm version -h` match canonical manual (in-process focused test).
- [x] `pm version help` and `pm version help --json` preserve positional help compatibility (in-process focused test).
- [x] JSON flag help returns `CommandManual/version` (in-process focused test).
- [x] Unknown version flag remains usage exit 2 (in-process focused test).
- [x] Invalid version action remains usage exit 2 and not a manual (in-process focused test).
- [x] `cli.Run` signature/stdout/stderr/JSON/exit semantics unchanged in focused/golden tests; broader gates pending.

## CLI parity checklist

- [ ] Runtime help checked using built binary.
- [ ] Bare leaf behavior checked (version metadata, not group help).
- [ ] Flag and positional help checked.
- [ ] `docs/cli/version.md` updated or N/A with docs-generator diff proof.
- [ ] `website/**` updated or N/A with source/generated diff proof.
- [ ] Golden/generated help artifacts updated or N/A with golden test/diff proof.
- [ ] Completion/discovery metadata updated or N/A; top-level command name unchanged and Phase 15 completion remains out of scope.

## Required gates

- [ ] `gofmt -w cmd internal`
- [x] focused version/router/golden tests (`7.814s`)
- [ ] `go test ./internal/cli/...`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] docs generate temp diff + docs validate
- [ ] website docs generator and tracked diff check
- [ ] `git diff --check`
- [ ] no `go.mod`/`go.sum` delta
- [ ] no connector definitions or unrelated namespaces changed

## Safety-limited / N/A

- Runtime-backed checks: N/A; version parsing has no service dependency.
- Credentialed checks: prohibited/not run.
- Reverse ETL: not in scope/not run outside existing local repository smoke gates.
- Dependencies: prohibited/not changed.
- External review and PR: explicitly prohibited by user for this run.

`verificationPassed` remains false until the complete `make verify` exits 0.
