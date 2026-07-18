# Phase 428 Verification

Session `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `235233f7cfde4a24612be6b0f95fb37a412d388a`.

## TDD and behavior

- [x] Six phase artifacts created before production edits.
- [x] Exact focused RED captured before production edits (missing native injected-runtime/action seams; build failed as expected).
- [x] Native agent/plan/image/build/pull/ensure/help tree; legacy wrapper removed.
- [x] Typed repeated/bare `--request`; unknown flags and extra args compatible.
- [x] Bare/text/JSON/long/short/positional help parity in focused tests.
- [x] Invalid actions and global assigned booleans preserve categories/output in focused tests.
- [x] Request/path/Podman-bin/image-reference validation covered.
- [x] All image actions covered through injected fakes/temp dirs; no Podman/Docker or image operation executed.
- [x] Agent trailing help and literal `--` compatibility preserved in focused tests.
- [x] Only the agent `parseFlags` call removed; shared/dynamic parser remains.
- [x] Deterministic output verified in focused tests.

## Gates

- [x] Focused agent/router tests (`4.408s`; expanded rerun `4.480s`).
- [ ] Focused agent/router/golden tests.
- [ ] Golden transcript test.
- [ ] Full `internal/cli/...`.
- [ ] Runtime dependency-free tests; optional services not started.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.
- [ ] `git diff --check`.

## CLI help/manual/website parity

- [ ] `pm help agent`, bare `pm agent`, `pm agent --help`, `pm agent -h`, `pm agent help` text parity.
- [ ] JSON help routes emit canonical `CommandManual/agent`.
- [ ] Built-binary plan output/error/global matrix passes; no image action invoked.
- [ ] `docs/cli/agent.md` update or no-change exemption proven by temp docs diff.
- [ ] `website/**` update or no-change exemption proven by generator/diff.
- [ ] Golden/generated help update or no-change exemption proven.
- [ ] Completion/discovery names unchanged; no-file seam tested; Phase 15 remains deferred.
- [ ] Focused subcommand help/man churn remains deferred to Phase 19.

## Safety/scope/delivery

- [ ] No secrets/credentials requested, read, printed, summarized, or stored.
- [ ] No Podman/Docker command, image build, image pull, publish, Temporal, PostgreSQL, or Dragonfly service executed.
- [ ] Image behavior tested only with injected fakes and temporary directories.
- [ ] Worker/RLM behavior unchanged.
- [ ] No dependency or `go.mod`/`go.sum` delta.
- [ ] No connector-def, unrelated namespace, or Phase 19 production delta.
- [ ] Coherent planning/RED/GREEN/verification commits pushed.
- [ ] No PR or external review requested.

Result: pending implementation and verification.
