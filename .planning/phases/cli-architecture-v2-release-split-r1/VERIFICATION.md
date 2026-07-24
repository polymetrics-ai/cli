# Verification checklist — CLI Architecture v2 Cobra/Viper release split

## Scope and provenance

- [ ] Exact base is recorded and still equals the PR base ancestry.
- [ ] Authorized source SHA mapping is complete.
- [ ] `git diff --check` passes.
- [ ] Diff contains no `internal/events`, `internal/logging`, `internal/telemetry`, `internal/ui`, TUI/OTel ADR/design/phase files, PR #493 routing surfaces, or PM review-system implementation.
- [ ] Module delta contains only approved Cobra/Viper dependencies and their audited transitives; no Charm, Bubble Tea, OTel promotion/addition, or `golang.org/x/term`.

## Focused behavior

- [ ] Config precedence: changed flag > `POLYMETRICS_*` > `PM_*` > effective-root config file > default.
- [ ] Missing config is allowed; malformed config uses existing validation exit 3.
- [ ] Primary env beats legacy alias; undocumented ambient env is ignored.
- [ ] Config load is invocation-scoped, re-entrant, and free of global Viper/`AutomaticEnv`/watchers.
- [ ] Runtime/RLM/schedule/worker/perf consumers use typed non-secret config without silently enabling runtime paths.
- [ ] Credential environment seams remain direct and secret values are neither loaded into typed config nor printed.
- [ ] Reverse smoke enforces plan -> preview -> approval -> execute.
- [ ] Current Gong dynamic connector behavior is preserved.

## CLI compatibility and docs parity

- [ ] 17 deterministic current-main/candidate comparisons match exit code, stdout bytes, and stderr bytes except explicitly documented new config behavior.
- [ ] `pm help config` is accurate.
- [ ] Root and JSON help are accurate.
- [ ] Bare namespace commands render contextual help and exit successfully where expected.
- [ ] Invalid actions retain usage errors.
- [ ] Dynamic connector passthrough/help remains byte-compatible.
- [ ] `docs/cli/**`, website CLI reference, generated docs data, and golden transcripts are fresh.

## Local checks

- [ ] `gofmt -w cmd internal`
- [ ] configured lint target / `golangci-lint` equivalent
- [ ] `go vet ./...`
- [ ] `go test -timeout 20m ./...`
- [ ] `go build ./cmd/pm`
- [ ] `go mod verify`
- [ ] `go mod tidy -diff`
- [ ] focused race coverage for config/runtime/schedule/safety and CLI golden/Cobra/config tests
- [ ] `make verify`
- [ ] fresh vulnerability/security scan

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
