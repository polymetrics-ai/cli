# VERIFICATION — issue-3062 Notion parity wave03 r1

Required local gates (fixture-only; no live provider calls):

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/notion` — exit 0; 1 connector checked, 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/notion' -count=1` — exit 0.
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — exit 0.
- [x] `go build ./cmd/pm` — exit 0.
- [x] `make connector-boundary` — exit 0; clean report.
- [x] `make verify` — exit 0 on final run.
- [x] `git diff --check` — exit 0.

Additional focused checks:

- [x] `go test ./internal/connectors/hooks/notion -count=1` — exit 0.
- [x] `go test ./cmd/connectorgen -run 'TestValidatePath|TestValidate_AcceptsGoodBundle' -count=1` — exit 0.
- [x] `./pm docs validate --connectors-dir docs/connectors` — exit 0.
- [x] `cd website && npm run gen:website-data` — exit 0; generated website connector data refreshed.
- [x] `go test -timeout 20m ./internal/connectors/certify` — exit 0; seeded cache for final `make verify` after parallel `go test ./...` timing variance.

Notes:

- Runtime-backed integration checks were intentionally not run.
- No credentials or Notion workspace API calls were used.
- `make verify` timing note: two earlier attempts hit package `-timeout 20m` in existing long-running test packages under parallel load (`internal/cli`, then `internal/connectors/certify`). Focused package reruns passed, and the final required `make verify` run passed cleanly.
