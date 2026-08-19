# Issue 4015 Sync Pipeline E2E — Verification

## Verdict

`gaps_found`: the PostgreSQL → PostgreSQL control completed its initial load and second incremental run with independent destination proof, but the same real-binary harness failed its process-death recovery step. The pipeline is therefore `broken-with-evidence` overall for this release gate rather than reported as wholly proven.

## Exact live command

Docker/Colima was already running. It was inspected and reused without a start or restart. Podman was down and was not started.

```bash
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryExecutesPostgresWarehousePostgres$'
```

The command was run twice. The test builds a fresh `pm` binary and programmatically executes the production CLI sequence below. Dynamic roots, loopback ports, plan IDs, and approval tokens are generated per run; secrets are supplied only through bounded stdin and are intentionally not reproduced.

```text
pm init --root <generated-project-root> --json
pm credentials add pg-source --connector postgres --config host=<loopback> --config port=<dynamic> --config database=pm_transport_source --config username=pm_transport --config schema=public --config sslmode=disable --value-stdin password --root <generated-project-root> --json
pm credentials add pg-target --connector postgres --config host=<loopback> --config port=<dynamic> --config database=pm_transport_target --config username=pm_transport --config schema=public --config sslmode=disable --value-stdin password --root <generated-project-root> --json
pm connections create postgres-to-postgres --source postgres:pg-source --destination postgres:pg-target --stream public.events --sync-mode incremental_upsert --cursor sequence --primary-key id --table events --root <generated-project-root> --json
pm etl transport postgres-managed-target plan --connection postgres-to-postgres --stream public.events --root <generated-project-root> --json
pm etl transport postgres-managed-target preview <generated-plan-id> --root <generated-project-root>
pm etl run --connection postgres-to-postgres --stream public.events --batch-size 1000 --approval-plan <generated-plan-id> --approval-token-stdin --confirm destructive --root <generated-project-root> --json
```

The plan/preview/run sequence was repeated for the incremental check. The harness later created a CDC bootstrap connection, killed its live `pm` process after an acknowledged change, prepared a fresh approval, and restarted the same connection to exercise recovery.

## Observable destination proof

The source was a real PostgreSQL table seeded as:

```sql
CREATE TABLE public.events (id bigint PRIMARY KEY, sequence bigint NOT NULL, label text NOT NULL);
INSERT INTO public.events (id, sequence, label)
SELECT value, value * 10, 'event-' || value::text
FROM generate_series(1, 1001) AS value;
```

The destination was a separate PostgreSQL database. After the `pm` child process exited, the Go harness used a direct `pgx` connection—not the writer or its run envelope—to discover and query the managed relation.

Observed first-run output:

```text
independent target read-back: pmmt_namespace_2961a6dc31c5f70416747bf2827bc643.pmmt_relation_d06db8373a79b9a74f24ea7340eabdd8 rows=1001 sample=(1001,10010,"event-1001")
```

The clean rerun observed the same independent state under a fresh generated relation identity:

```text
independent target read-back: pmmt_namespace_8d89a38553dbd88845a9697d68ef9868.pmmt_relation_7005d041fa693adc4ec856fde940b8ad rows=1001 sample=(1001,10010,"event-1001")
second incremental run: records_read=0 records_loaded=0 (checkpoint skip)
independent replay read-back: pmmt_namespace_8d89a38553dbd88845a9697d68ef9868.pmmt_relation_7005d041fa693adc4ec856fde940b8ad rows=1001 sample=(1001,10010,"event-1001")
```

This proves the destination was not empty and the sample content matched the named source row exactly.

## Incremental behavior

The connection declared `incremental_upsert` with cursor `sequence` and primary key `id`.

- First run: 1,001 rows reached the managed target.
- Second run: zero rows were read or loaded.
- Independent target result: still exactly 1,001 rows; sample `1001/10010/event-1001` unchanged.
- Classification: **skip**. No duplicates and no updates were produced because no source row advanced past the acknowledged polling watermark. This matches the declared incremental polling behavior.

The prior test expectation demanded a 1,001-row replay and failed after the runtime correctly skipped acknowledged rows. The test-only expectation was corrected to assert the observed contract.

## Recovery finding

After the successful bootstrap snapshot and a committed post-barrier transaction, the harness killed the `pm` process, inserted source row `id=3, sequence=40, label='resumed-after-process-death'`, and restarted the same approved connection. The restarted child exited before the row or a new checkpoint appeared:

```json
{
  "api_version": "polymetrics.ai/v1",
  "error": {
    "category": "internal",
    "code": "internal_error",
    "message": "sync rebootstrap required: invalid_checkpoint: PostgreSQL polling checkpoint mechanism is not resumable"
  },
  "kind": "Error"
}
```

Observed test failure:

```text
PostgreSQL transport process exited before resumed CDC row and checkpoint after process restart: exit status 1
```

This is a product finding, not repaired in this task. Initial delivery, independent read-back, batching across 1,001 rows with `--batch-size 1000`, and the no-change incremental rerun work; process-death recovery does not.

## Per-pipeline verdicts

| Pipeline | Verdict | Evidence / reason |
| --- | --- | --- |
| PostgreSQL → PostgreSQL | **broken-with-evidence** | Initial load and no-change incremental rerun are independently proven; CDC process restart fails with the typed `invalid_checkpoint` error above before the resumed row reaches the destination. |
| PostgreSQL → GitHub | **not-attempted-and-why** | Strict ordering stopped after the control exposed a release-gate recovery defect. The existing live harness targets a repository outside the authorized `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` fixture and intentionally retains labels, so it could not be safely reused under this brief. |
| GitHub → PostgreSQL | **not-attempted-and-why** | Not reached after the control finding within the time box. No GitHub credential was read. |
| GitHub → GitHub | **not-attempted-and-why** | Not reached after the control finding within the time box. No GitHub mutation was made. |

## Cleanup proof

The two attempts generated resources with suffixes `4dcda5c12cfd32c2` and `561637df3242d291`. Docker events independently recorded `volume destroy` and `container destroy` for both database fixtures, and `container destroy` for both capacity probes.

The harness intentionally retains generated run-image tags, so this task explicitly removed only its two exact tags:

```bash
docker --host unix:///Users/karthiksivadas/.colima/default/docker.sock image rm localhost/polymetrics-postgrestransport-it-4dcda5c12cfd32c2:run localhost/polymetrics-postgrestransport-it-561637df3242d291:run
```

Independent follow-up inspection returned:

```text
task run-image tags absent
task containers absent
```

Two exited PostgreSQL harness containers and many run-image tags predated this task and were deliberately left untouched. No GitHub fixture was created, so no GitHub deletion or 404 proof applies.

## Safety

- No secret was printed, persisted in evidence, passed in argv, or read from the macOS Keychain.
- Approval tokens remained bounded stdin-only values inside the harness.
- Docker/Colima was reused without restart; Podman remained stopped.
- Product code was not changed.

## Repository gates

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./internal/cli` | pass (`ok polymetrics.ai/internal/cli 549.685s`) |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `scripts/verify-gsd-workflow` | pass; GSD/TDD evidence recognized |
| `make tidy-check lint` | pass; `0 issues` |
| `make docs-check smoke-no-build` | pass; connector docs validated and smoke fixture materialized/read |
| `make agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` | pass; 552 connectors validated, surface drift `0`, boundary clean, release targets consistent |
| `git diff --check` | pass |
| `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` | pass; `AGENTS.md` unchanged |

`go test -timeout 20m ./...` and the aggregate `make verify` were not run locally because this repository explicitly directs per-command agents to split the 550+ connector suite and let CI carry the full suite; `make verify` includes that same full test command. All non-full-suite `make verify` gates named by `AGENTS.md` were run individually above.
