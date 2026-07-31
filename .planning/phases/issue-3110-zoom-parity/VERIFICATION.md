# Verification checklist — Zoom connector parity (#3110)

Required gates from the worker brief and issue bodies:

- [x] Focused connectorgen validation for Zoom (run over `internal/connectors/defs`, confirm Zoom has 184 ledger rows and no findings)
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `make verify` (pre-correction run included out-of-scope bundle-registry cache)
- [x] `git diff --check`
- [x] Regenerate/update connector docs: `docs/connectors/zoom/MANUAL.md`, `docs/connectors/zoom/SKILL.md`, Zoom docs/catalog/website generated data as applicable
- [x] Append/update captain-policy addendum on #3110-#3117 through `gh-axi`
- [x] Commit clean branch; do not push, open PR, merge, or invoke no-mistakes

Corrective no-cache rerun after scope correction:

- [x] Restore `internal/connectors/bundleregistry/registry.go` to pre-`426eb0e3b` contents in a new commit; do not amend/reset/rebase.
- [x] Rerun focused connectorgen validation for Zoom.
- [x] Rerun `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1`.
- [x] Rerun `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` without the cache change.
- [x] Rerun `make verify` without the cache change; if it times out, record the timeout as a shared performance dependency rather than retaining shared code.
- [x] Rerun `git diff --check`.
- [x] Commit corrective scope fix; do not push, open PR, merge, or invoke no-mistakes.

Safety evidence to preserve:

- No live Zoom API calls.
- No credential request, printed secret, stored secret, or secret-shaped fixture.
- No new dependency.
- No shared runtime/engine behavior edits.
- No raw generic HTTP/query/SQL/shell/file/passthrough escape hatch.
- Reverse ETL remains plan → preview → explicit approval → execute; destructive actions use typed confirmation metadata.
