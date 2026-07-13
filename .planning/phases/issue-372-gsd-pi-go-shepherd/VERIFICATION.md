# Verification Checklist

- [x] Pinned GSD Pi version resolves to 1.11.0.
- [x] `gpt-5.6-sol` appears in the Codex provider catalog with high reasoning support.
- [x] A disposable governed session was observed as GPT-5.6 Sol/high after the initial `off` mismatch was made admission-fatal.
- [x] Supported headless query and filtered lifecycle events observed.
- [x] Unsupported/ambiguous headless behavior recorded fail-closed.
- [x] Workflow contract tests pass.
- [x] `go test ./...` passes inside `agent-runtime/shepherd/`.
- [x] `go test -race ./...` passes inside `agent-runtime/shepherd/`.
- [x] Root `go list ./...` excludes the nested module.
- [x] Root `go test ./...` and `go build ./cmd/pm` passed before the adversarial hardening slice.
- [ ] Root gates rerun after the adversarial hardening slice.
- [x] Core named incident guard suite passes.
- [x] Governed intake canary emits <=15-second heartbeat and corrects premature upstream success to blocked.
- [x] A delayed human response keeps emitting heartbeats; the Go deadline cancels before GSD's
  fallback timer, so confirmation cannot auto-approve.
- [x] Stable issue identity, typed context binding, persistent attempts, explicit blocked resume,
  real `skip`/`stop` query shapes, tool errors, project setting overrides, and symlink escape have
  adversarial tests.
- [ ] Merge-disabled canary reaches exact-head human gate.
- [ ] Cleanup inventory reviewed before deletion.
