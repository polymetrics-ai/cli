# Verification Checklist

- [x] Pinned GSD Pi version resolves to 1.11.0.
- [x] `gpt-5.6-sol` appears in the Codex provider catalog with high reasoning support.
- [ ] An actual validation dispatch is observed as GPT-5.6 Sol/high (first coordinator session exposed `off`; admission now blocks that configuration).
- [x] Supported headless query and filtered lifecycle events observed.
- [x] Unsupported/ambiguous headless behavior recorded fail-closed.
- [x] Workflow contract tests pass.
- [x] `go test ./...` passes inside `agent-runtime/shepherd/`.
- [x] `go test -race ./...` passes inside `agent-runtime/shepherd/`.
- [x] Root `go list ./...` excludes the nested module.
- [ ] Root `go test ./...`, `go build ./cmd/pm`, and `make verify` pass.
- [x] Core named incident guard suite passes.
- [x] Governed intake canary emits <=15-second heartbeat and corrects premature upstream success to blocked.
- [ ] Merge-disabled canary reaches exact-head human gate.
- [ ] Cleanup inventory reviewed before deletion.
