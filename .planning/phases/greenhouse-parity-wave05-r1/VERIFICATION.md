# Verification checklist — Greenhouse parity wave05 r1

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 550 connector(s), 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/greenhouse' -count=1` — passed.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — passed.
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors` — passed.
- [x] `go vet ./internal/connectors/... ./internal/cli/...` — passed.
- [x] `go build ./cmd/pm` — passed.
- [x] `make connector-boundary` — passed, no findings.
- [x] `git diff --check` — passed.
- [x] `make verify` — final rerun passed (`exit=0`, duration 905.2s). First attempt hit the harness timeout in `internal/connectors/certify` after 20m; focused certify reruns passed, then full `make verify` passed.
- [x] Final pre-commit `git status --short` — changes limited to Greenhouse defs/fixtures/docs, connector catalog generated counts, and this GSD phase artifact.

Additional note: `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` was run as required; it exited with `conflict: both AGENTS.md and CLAUDE.md are real files ... reconcile them manually` and made no changes.
