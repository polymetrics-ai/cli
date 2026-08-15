# Verification — #3860 polling-watermark truth surfaces

| Area | Command/evidence | Status |
| --- | --- | --- |
| TDD red/green | `TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt`; `TestPlannedPollingWatermarkJSONOmitsUndeclaredRuntimeContract`; ledger | pass |
| Required truth surfaces | `go test -count=1 -timeout 20m ./internal/cli -run '^(TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt\|TestPostgresNativeAPISurfaceHasNoFabricatedRESTEndpoints\|TestConnectorsHelpExplainsDeclaredNoneInspectionPolicy)$'` | pass — observed planned state/reason; no executable CDC words; native `api_surface.endpoints` length is zero; help includes all limits |
| Unsafe/eligible polling contracts | `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestBundleRejectsUnsafePollingWatermarkDefinition\|TestPollingModeEligibilityComesFromRealPreflight\|TestPollingSourceDefinitionSoftDeleteRequiresCursorAdvancingMapping\|TestPollingSourceDefinitionRejectsSourceIdentityRebootstrapWithoutReason)$'` | pass — unsafe mode is rejected, a valid registered declaration is the only implemented state, and soft-delete/identity requirements remain observable |
| Fresh binary | `go build -o ./bin/pm ./cmd/pm` before `./bin/pm docs generate --dir docs/cli` | pass |
| Runtime parity | `./bin/pm help connectors`; `./bin/pm connectors`; inspect and CDC catalog JSON with `jq` assertion | pass — bare namespace gives help; PostgreSQL is `planned`, carries a reason, and has no `source`/`target` binding |
| Generated docs/website | `npm run gen:website-data`; `./bin/pm docs validate --connectors-dir docs/connectors`; normalized generated-data hashes | pass — sanctioned generation, docs validation, and hash audit prove only PostgreSQL's connector data changed |
| Connector validation | `go run ./cmd/connectorgen validate internal/connectors/defs`; `go run ./cmd/connectorgen surface-sync --check`; `go vet ./internal/cli ./internal/connectors/...` | pass — 552 definitions, zero findings; zero surface corrections; vet clean |
| Live database lane | Supplied `POLYMETRICS_DATABASE_INTEGRATION=1 ... go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres` | waived by supervisor — one attempt timed out before PostgreSQL startup because Docker blocked in `dbtest` image-store validation; independent `docker info` also stalled under machine saturation. No retry authorized. |
| GSD verification/review | generated execute/verify/review prompts and inline manual source/diff review | pass — no correctness or security findings; incompatible Pi runtime / no delegated roles is recorded in plan |
| Delivery | rebased commit `38e46afe1`, pushed task-owned branch, PR [#4157](https://github.com/polymetrics-ai/cli/pull/4157), `gh api /repos/polymetrics-ai/cli/pulls/4157 --jq .base.ref` | pass — API reported `integration/4015-mvp-flat-r1`; CI pending |

## Safety and scope

- No credentials, provider calls, generic SQL/HTTP/shell write, reverse ETL execution, dependency, or driver change.
- #4125, #4136, #4090, and #4154 remain excluded.
- Generated artifacts are only modified by their sanctioned generator after building a fresh binary.
- Full `./internal/cli/...` and `./internal/connectors/...` sweeps were started before the live lane but contended with six other full Go/container jobs on a saturated machine; the targeted required proofs above are the authoritative local verification, with CI retained as the full-suite gate.
