# Verification checklist — CLI Architecture v2 Cobra/Viper release split

## Scope and provenance

- [x] Exact base is recorded and still equals the PR base ancestry (`873cd7b251f70c4a35a607a0d4e86051ea0fbd15`).
- [x] Authorized source SHA mapping is complete.
- [x] `git diff --check` passes.
- [x] Diff contains no `internal/events`, `internal/logging`, `internal/telemetry`, `internal/ui`, TUI/OTel ADR/design/phase files, PR #493 routing surfaces, or PM review-system implementation.
- [x] Module delta contains only approved Cobra/Viper dependencies and their audited transitives; no Charm, Bubble Tea, OTel promotion/addition, or `golang.org/x/term`.

At local verification checkpoint `9c5cb9b5902ffadeb50c727780312007c0acead1`, the candidate was 74 paths and 10 additive commits from the exact base. `go mod why` attributed Cobra only to `internal/cli` and Viper only to `internal/config`.

## Focused behavior

- [x] Config precedence: changed flag > `POLYMETRICS_*` > `PM_*` > effective-root config file > default.
- [x] Missing config is allowed; malformed config uses existing validation exit 3.
- [x] Primary env beats legacy alias; undocumented ambient env is ignored.
- [x] Config load is invocation-scoped, re-entrant, and free of global Viper/`AutomaticEnv`/watchers.
- [x] Runtime/RLM/schedule/worker/perf consumers use typed non-secret config without silently enabling runtime paths.
- [x] Credential environment seams remain direct and secret values are neither loaded into typed config nor printed.
- [x] Reverse smoke enforces plan -> preview -> approval -> execute.
- [x] Current Gong dynamic connector behavior is preserved.

Evidence: focused config/runtime/schedule/safety and worker/RLM/perf tests passed; the configured reverse smoke passed against temporary local sample/warehouse/outbox state; source inspection found only `viper.New()` and no prohibited Viper global/watcher calls.

## CLI compatibility and docs parity

- [x] 17 deterministic current-main/candidate comparisons match exit code, stdout bytes, and stderr bytes except explicitly documented new config behavior.
- [x] `pm help config` is accurate.
- [x] Root and JSON help are accurate.
- [x] Bare namespace commands render contextual help and exit successfully where expected.
- [x] Invalid actions retain usage errors.
- [x] Dynamic connector passthrough/help remains byte-compatible.
- [x] `docs/cli/**`, website CLI reference, generated docs data, and golden transcripts are fresh.

The comparison ran under an empty home/project with `CI=1` and `TERM=dumb`; all 17 status/stdout/stderr triplets matched byte-for-byte. `npm run gen:docs` regenerated only `website/lib/docs.generated.ts` after the GitHub dynamic-command wording correction.

## Local checks

- [x] `gofmt -w cmd internal`
- [x] configured lint target / `golangci-lint` equivalent (`make verify`: 0 issues)
- [x] `go vet ./...`
- [x] `go test -timeout 20m ./...`
- [x] `go build ./cmd/pm`
- [x] `go mod verify`
- [x] `go mod tidy -diff`
- [x] focused race coverage for config/runtime/schedule/safety and CLI golden/Cobra/config tests
- [x] `make verify`
- [x] fresh vulnerability/security scan under repository/CI toolchain Go 1.25.12: no vulnerabilities found

The workstation's newer ambient Go 1.26.4 separately reports standard-library GO-2026-5856, fixed in Go 1.26.5. The repository and CI pin Go 1.25.12; a fresh `GOTOOLCHAIN=go1.25.12 govulncheck ./...` returned `No vulnerabilities found`. No candidate dependency caused the ambient standard-library result.

## GitHub/release gates

- [ ] PR targets `main`, has a Conventional Commit title, truthful source provenance, GSD/TDD/skills evidence, exclusions, compatibility results, and no merge authorization.
- [ ] Verify, CodeQL, Dependency Review, docs/website/generated checks, and all required branch checks pass.
- [ ] Fresh Snyk passes; no PR #495 deferral reused.
- [ ] Exact-version bounded PM Codex packets complete.
- [ ] One PM synthesis is clean.
- [ ] Independent Shepherd completes only after clean synthesis.
- [ ] No-mistakes reaches `checks-passed` or `passed` with an open PR.
- [ ] Release version is mechanically selected or escalated as one concrete decision.
- [ ] No prerelease tag/release is published before every required gate; no merge is performed.

`verificationPassed` remains false until the full declared campaign, including `make verify`, succeeds.
