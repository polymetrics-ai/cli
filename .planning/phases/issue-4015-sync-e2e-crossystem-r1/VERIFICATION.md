# Verification — issue #4015 cross-system pipelines

## Live invocation

The credential was read inline from macOS Keychain service `pm-cert-classic` into an environment variable for this process only. Its value was never printed, persisted, placed in argv, or supplied to GitHub CLI. The existing Colima Docker endpoint was used without a start, stop, or restart.

```text
env POLYMETRICS_GITHUB_TOKEN="$(security find-generic-password -s pm-cert-classic -w)" \
  POLYMETRICS_DATABASE_INTEGRATION=1 \
  POLYMETRICS_CROSS_SYSTEM_LIVE_PROOF=1 \
  POLYMETRICS_CONTAINER_RUNTIME=docker \
  POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock \
  go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli \
  -run '^TestPMBinaryExecutesLiveCrossSystemPipelines$'
```

Fresh binary SHA-256: `c26691fa0925122d61163d5c9a58e56bb62ecbc6e39a18824858cfedacc2e205`; size: `180056098` bytes. The test returned failure because its route-2 assertions deliberately characterize the product defect. It continued through routes 3 and 4 and unconditional cleanup.

## Route verdicts

### 1. PostgreSQL → GitHub — proven

Commands:

```text
pm connections create postgres-to-warehouse --source postgres:pg-source --destination warehouse:warehouse-cross-system --stream public.label_updates --sync-mode incremental_upsert --cursor sequence --primary-key id --table db_api_label --root <root> --json
pm etl run --connection postgres-to-warehouse --stream public.label_updates --batch-size 1 --root <root> --json
pm reverse plan pm-cert-db-api-first-e10940f636b8 --source-table db_api_label --connection postgres-to-warehouse --destination github:github-cross-system --action update_label --map name:name --map new_name:new_name --map color:color --map description:description --root <root>
pm reverse preview <plan-id> --root <root> --json
printf '<approval-token>\\n' | pm reverse approve <plan-id> --from-stdin --root <root> --json
pm reverse run <plan-id> --root <root> --json
```

- First ETL: read/loaded `1/1`; independent warehouse query: exactly one row named `pm-cert-db-api-e10940f636b8`.
- Independent destination GET: one label, color `1d76db`, description `pm-cert-db-api-updated-e10940f636b8`.
- Replay: `incremental_upsert` read/loaded `0/0`, matching its acknowledged cursor; a separately approved typed replay left exactly one label with unchanged content.

### 2. GitHub → PostgreSQL — broken-with-evidence

Commands:

```text
pm connections create github-to-postgres --source github:github-cross-system --destination postgres:pg-target --stream labels --sync-mode full_overwrite --primary-key name --table github_labels --root <root> --json
pm etl transport postgres-managed-target plan --connection github-to-postgres --stream labels --batch-size 100 --root <root> --json
pm etl transport postgres-managed-target preview <plan-id> --root <root> --json
printf '<approval-token>\\n' | pm etl transport postgres-managed-target approve <plan-id> --from-stdin --root <root> --json
pm etl run --connection github-to-postgres --stream labels --batch-size 100 --approval-plan <plan-id> --confirm destructive --root <root> --json
```

- First run: independently counted GitHub source `10`; run read/loaded `10/10`; a separate pgx connection read exactly 10 PostgreSQL rows and matched label `pm-cert-db-api-e10940f636b8`, color `1d76db`, and its description.
- Replay using the same standing authorization: status `completed`, read/loaded `0/0`; independent pgx read-back found zero rows and the named sample absent.
- Declared mode is `full_overwrite`, so the expected replay was another complete `10/10` replacement, not checkpoint skip plus empty replacement. No product fix is included.

### 3. GitHub → GitHub — proven

Commands:

```text
pm connections create github-to-warehouse --source github:github-comment-source --destination warehouse:warehouse-cross-system --stream issue_comments --sync-mode full_refresh_overwrite --cursor updated_at --primary-key id --table api_api_comments --root <root> --json
pm etl run --connection github-to-warehouse --stream issue_comments --batch-size 10 --root <root> --json
pm reverse plan pm-cert-api-api-e10940f636b8 --source-table api_api_comments --connection github-to-warehouse --destination github:github-cross-system --action update_issue_comment --limit 1 --map id:comment_id --map repository:body --root <root>
pm reverse preview <plan-id> --root <root> --json
printf '<approval-token>\\n' | pm reverse approve <plan-id> --from-stdin --root <root> --json
pm reverse run <plan-id> --root <root> --json
```

- The source credential used an exact pre-creation `since` boundary; independent listing found only task comment `5328121289`.
- ETL read/loaded `1/1`; independent warehouse query found exactly that one ID.
- Independent destination GET found comment `5328121289` with body `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`; independent listing found exactly one occurrence.
- Replay is the scheduled flow below: `full_refresh_overwrite` read one row again and the typed update wrote one row, leaving one unchanged destination comment. This matches full refresh plus idempotent update semantics.

### 4. ETL/reverse-ETL through schedule and flow — proven

Commands:

```text
pm flow plan --file <flow.json> --root <root> --json
pm flow preview --file <flow.json> --root <root> --json
pm flow create --file <flow.json> --root <root> --json
pm schedule create --name pm-cert-schedule-e10940f636b8 --cron '0 2 * * *' --flow pm-cert-cross-system-e10940f636b8 --root <root> --json
pm schedule install pm-cert-schedule-e10940f636b8 --crontab --root <root> --json
<fresh-pm> --root <root> flow run pm-cert-cross-system-e10940f636b8 --json
pm schedule inspect pm-cert-schedule-e10940f636b8 --root <root> --json
```

Scheduling entry point: the exact payload emitted by `pm schedule install ... --crontab` into the task-local `PM_CRONTAB_FILE`:

```text
/var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/polymetrics-cli-pm-2617849072/pm --root /tmp/fm-cli-sync-e2e-crossystem-r1/gotmp/TestPMBinaryExecutesLiveCrossSystemPipelines3823562150/001/cross-system-project flow run pm-cert-cross-system-e10940f636b8 --json
```

The certification invoked that installed payload, rather than a one-shot ETL shortcut.

The persisted flow honored both declared steps: `issue_comments` full refresh read one row and the approved action updated one row. The result was `ok`, carried a non-empty prepared execution identity, and `schedule inspect` recorded terminal state. Independent GitHub GET/list read-back found exactly comment `5328121289` with the expected unchanged body.

## Cleanup and containment

- Final decisive run: label `pm-cert-db-api-e10940f636b8` returned HTTP 404 after deletion; issue comment `5328121289` returned HTTP 404 after deletion.
- All earlier diagnostic fixtures were also removed and checked independently: labels `pm-cert-db-api-ebeb78245608`, `pm-cert-db-api-648f3727cf22`, `pm-cert-db-api-48f28d289f66`, and `pm-cert-db-api-a725c16ef17b`; release `372322374`; tag `pm-cert-api-api-48f28d289f66`; comment `5328103780`.
- The project tree scan found no credential or approval token material. No task-owned PostgreSQL container or volume remained.
- The optional 5 GB run was not attempted because all four routes did not pass; correctness remained the gate for scale.

## GSD verify-work result

Inline/manual fallback: the official `verify-work` prompt was resolved and executed without spawning a delivery role, as required by the canonical single-worker contract. D1, D3, D4, and cleanup pass from direct live observations. D2 fails its incremental/full-overwrite requirement with the exact evidence above. Overall certification result: `verified_with_finding`.
