# Verification checklist — Zoom connector parity (#3110)

Required gates from the worker brief and issue bodies:

- [x] Focused connectorgen validation for Zoom (run over `internal/connectors/defs`, confirm Zoom has 184 ledger rows and no findings)
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/zoom' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `make verify`
- [x] `git diff --check`
- [x] Regenerate/update connector docs: `docs/connectors/zoom/MANUAL.md`, `docs/connectors/zoom/SKILL.md`, Zoom docs/catalog/website generated data as applicable
- [x] Append/update captain-policy addendum on #3110-#3117 through `gh-axi`
- [x] Commit clean branch; do not push, open PR, merge, or invoke no-mistakes

Safety evidence to preserve:

- No live Zoom API calls.
- No credential request, printed secret, stored secret, or secret-shaped fixture.
- No new dependency.
- No shared runtime/engine behavior edits.
- No raw generic HTTP/query/SQL/shell/file/passthrough escape hatch.
- Reverse ETL remains plan → preview → explicit approval → execute; destructive actions use typed confirmation metadata.
