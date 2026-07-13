# Verification Checklist

- [x] Pinned GSD Pi version resolves to 1.11.0.
- [x] `gpt-5.6-sol` appears in the Codex provider catalog with high reasoning support.
- [x] Supported headless query and filtered lifecycle events observed.
- [x] Unsupported/ambiguous headless behavior recorded fail-closed.
- [ ] Workflow contract tests pass.
- [ ] `go test ./...` passes inside `agent-runtime/shepherd/`.
- [ ] `go test -race ./...` passes inside `agent-runtime/shepherd/`.
- [ ] Root `go list ./...` excludes the nested module.
- [ ] Root `go test ./...`, `go build ./cmd/pm`, and `make verify` pass.
- [ ] Incident replay suite passes.
- [ ] Merge-disabled canary reaches exact-head human gate.
- [ ] Cleanup inventory reviewed before deletion.

