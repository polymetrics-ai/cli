# Connector Certification Harness Design — `pm connectors certify`

Status: approved (2026-07-02). Program PRD: `docs/plans/universal-programming-loop-prd.md`.

## Load-bearing facts (verified in code)

1. **`cli.Run(args, stdout, stderr) int` is a pure in-process entrypoint** (`internal/cli/cli.go:24`).
   The certifier drives the *real* CLI surface — flags, `--json` envelopes, exit codes — without
   spawning processes.
2. **`--root` gives full project isolation.** The Makefile `smoke` target already runs an
   end-to-end pipeline in a `mktemp -d` root. Certify uses the same pattern: one ephemeral root per
   connector run.
3. **Destination sync-mode logic is connector-independent** (`internal/app/app.go` +
   `internal/app/sync_modes.go`): append/overwrite/dedup semantics live in the app layer against
   the local warehouse, not in any connector.
4. **`connectors.LiveConformanceProvider` exists with zero implementations** — it is the intended
   live hook; certify gives it purpose.
5. **Gotchas**: `directConnector` (`cli.go:591`) builds RuntimeConfig from `--config` only — `pm
   etl check/read --connector` never resolves credential Secrets. Live check must go through `pm
   credentials test` (vault-resolved secrets), and `--credential` is added to `etl check/read` as a
   prerequisite fix. `pm reverse plan --json` deliberately strips the approval token — the harness
   parses the token from text output and separately asserts the JSON path keeps hiding it
   (redaction gate).
6. Crontab entries carry a `# pm-schedule-<name>` sentinel (`internal/schedule/crontab.go:88,100`)
   — a verifiable roundtrip marker. `runScheduleRemove` ignores backend removal errors — certify
   verifies independently.

## The honest answer: live-testing all 5 sync modes per connector

Not wrong — mostly redundant, and the redundant part is where cost/flakiness/rate-limit risk live.
The 5 modes decompose into two axes: **source axis** (full vs incremental+resume — connector-
specific, MUST be live-tested: pagination, cursors, rate limits, auth) and **destination axis**
(append/overwrite/dedup — app-layer, connector-independent). Certify covers **all 5 modes per
connector** but only 2 live API reads (full_refresh_append + incremental_append, plus a resume
re-read); the 3 destination variants replay the captured JSONL through the real pipeline via the
built-in `file` connector (still exercising ParseSyncMode, generation IDs, truncation, PK dedup).
Each mode's report entry records `data_source: live | capture`. A `--live-all-modes` escape hatch
exists.

Constraints: writes/deletes only via create-then-cleanup with tagged ephemeral records + mandatory
cleanup verification + orphan sweeper; sandbox strongly recommended (`sandbox: true` in creds file
or `--allow-production-writes`); token-bucket per connector (default 2 rps, 500-call budget;
exhaustion = `skipped: budget_exhausted`); parallelism across connectors only (default 4);
no credential → `uncertified`, never `failed`.

### Tiers

| Tier | What | When | Credentials |
|---|---|---|---|
| 0 | Fixture conformance (conformance v2) | every CI run | none |
| 1 | Recorded HTTP replay: sanitized golden cassettes | every CI run, deterministic | none |
| 2 | Live certification: full command surface + create-then-cleanup writes + flow + schedule | on demand / nightly | gated |

## A) Command spec

```
pm connectors certify <connector>
    [--credential <existing-name> | --from-env field=ENV ... --config k=v ...]
    [--stream <name>]            # default: first stream with a cursor field, else first
    [--limit N]                  # per-read record cap (default 50)
    [--modes m1,m2,...]          # default: all 5
    [--skip write,flow,schedule]
    [--rate-limit RPS] [--budget N]
    [--record | --replay]        # Tier-1 capture / replay
    [--live-all-modes] [--allow-production-writes]
    [--keep-workdir] [--json]
pm connectors certify --all --credentials-file creds.yaml [--parallel N] [--resume] [--json]
pm connectors certify --sweep [--credentials-file creds.yaml] [--older-than 24h] [--json]
```

**Exit codes**: 0 pass · 1 usage/internal · 2 certification failures · **3 leaked resources**
(dominates everything). Wire via typed `certify.ExitError` in `internal/cli/errors.go`.

**Execution model**: `os.MkdirTemp` root → `pm init --root <dir>` → every stage is an in-process
`cli.Run([..., "--root", dir, "--json"], &out, &err)` whose exit code and envelope `kind` are
asserted → leak verification → report written → workdir deleted.

### Stage list

Source stages: 0 preflight (registry, secrets present, budget armed) · 1 fixture_conformance
(Tier-0 embedded) · 2 manual_json (`connectors inspect --json`, no secret values) · 3
credentials_add/test (LIVE check via vault) · 4 catalog_live (≥1 stream; PK/cursor recorded) · 5
etl_full_refresh_append (LIVE; records_read>0 or `passed_empty` warning; output JSONL = capture) ·
6 etl_full_refresh_overwrite (capture; run twice; truncate semantics via `pm query`) · 7
etl_full_refresh_overwrite_deduped (capture; no duplicate PK tuples) · 8 etl_incremental_append
(LIVE; cursor set) · 9 resume (LIVE run 2; records_read(run2) ≤ run1; cursor monotonic; no row
below run-1 checkpoint; in --record mode assert outbound request carried the cursor param) · 10
etl_incremental_append_deduped (capture) · 11 query_contract.

Write stages: 12 write_plan_preview (text yields plan id + token; `--json` contains NO token) · 13
write_create (`reverse run --approve`: succeeded=1, failed=0) · 14 write_verify (live read-back
finds tag; else `unverified` warning) · 15 write_cleanup · 16 cleanup_verify (entity gone —
failure ⇒ `leaked_resource`) · 17 approval_idempotency (consumed plan+token re-run must fail).

Glue stages: 18 flow_roundtrip (ephemeral manifest: capture-backed etl step + query step; plan
order, preview dry_run with zero side effects, run completed, status per-step) · 19
schedule_roundtrip (snapshot crontab → create → list → install --crontab → sentinel present →
remove → sentinel absent AND `crontab -l` byte-identical → manifest deleted; residue ⇒
`leaked_schedule`, harness force-removes sentinel before reporting) · 20 secret_redaction_live
(scan ALL captured stdout/stderr + report for secret values: exact, base64, URL-encoded) · 21
json_contract (meta-stage aggregating envelope kind + exit-code assertions).

### Report artifact — `.polymetrics/certifications/<connector>.json`

```json
{
  "kind": "ConnectorCertification",
  "schema_version": 1,
  "connector": "github",
  "pm_version": "v0.x.y",
  "connector_manifest_hash": "sha256:...",
  "started_at": "...", "completed_at": "...",
  "mode": "live",
  "credential_ref": "cert-github",
  "passed": true,
  "leaks": [],
  "budget": {"calls_used": 143, "calls_budget": 500, "rate_limit_rps": 2},
  "fixture": { "...embedded conformance report...": true },
  "capabilities": {
    "check":   {"live": "pass"},
    "catalog": {"live": "pass", "streams": 21},
    "read":    {"live": "pass", "stream": "issues", "records": 50},
    "sync_modes": {
      "full_refresh_append":            {"result": "pass", "data_source": "live"},
      "full_refresh_overwrite":         {"result": "pass", "data_source": "capture"},
      "full_refresh_overwrite_deduped": {"result": "pass", "data_source": "capture"},
      "incremental_append":             {"result": "pass", "data_source": "live", "cursor_advanced": true},
      "incremental_append_deduped":     {"result": "pass", "data_source": "capture"}
    },
    "resume": {"result": "pass"},
    "write_actions": {
      "create_issue": {"result": "pass", "cleanup": "close_issue", "verify": "read_back", "tag": "pm-cert-github-ab12cd34-1751450000"},
      "merge_pull_request": {"result": "skipped", "reason": "no safe cleanup pairing"}
    },
    "flow": {"result": "pass"},
    "schedule": {"result": "pass", "backend": "crontab", "residue": false},
    "json_contract": {"result": "pass", "stages_checked": 23},
    "secret_redaction": {"result": "pass"}
  },
  "stages": [
    {"name": "credentials_test", "tier": 2, "passed": true, "duration_ms": 812,
     "cli": {"argv_redacted": "pm credentials test cert-github --json", "exit_code": 0, "kind": "CredentialTest"}}
  ]
}
```

History appends to `.polymetrics/certifications/history/<connector>/<timestamp>.json`.

### Enablement (replacing the manual flip)

1. Curated artifacts committed under `internal/connectors/certifications/<name>.json` (go:embed)
   when a maintainer accepts a run.
2. A generator (`cmd/certstatus`) rewrites catalog capability flags from artifacts: enabled
   requires `passed=true`, `mode=live`, matching `connector_manifest_hash`, current
   `schema_version`, age < 90 days. Sources need sync_modes + resume green; destinations need
   write_actions green with cleanup verified.
3. A guard test fails the build if any enabled connector lacks a valid artifact. `pm connectors
   inspect` renders live gates from the artifact (absence ⇒ `uncertified`).

## B) Batch mode

`creds.yaml` (env/exec references only — safe to commit; never secret values):

```yaml
version: 1
defaults: {limit: 50, rate_limit_rps: 2, budget_calls: 500, parallel: 4}
connectors:
  github:
    credential:
      from_env: {token: PM_CERT_GITHUB_TOKEN}
      config:   {repository: polymetrics-ai/cert-sandbox}
    sandbox: true
    write: true
  stripe:
    credential:
      exec: {api_key: ["op", "read", "op://certs/stripe-test/api_key"]}
    write: false
    rate_limit_rps: 1
  salesforce:
    skip: true
    reason: "no sandbox tenant yet"
```

Worker pool over connectors (default 4), one ephemeral root each; stages within a connector
strictly serial; per-connector token bucket. Resumability: `certifications/batch-<runid>/
progress.json`; `--resume` skips connectors whose report is newer than batch start. Summary
matrix: rows = connectors; columns = check/catalog/read/5 modes/resume/write/flow/schedule/
redaction/**leaks**; any leak row printed first and forces exit 3.

## C) Create-then-cleanup write protocol

**Tag**: `pm-cert-<slug>-<runid8>-<unix-ts>` in the primary human-visible field, plus
`{"pm_certify":{"run","created_at","host"}}` metadata where schema allows.

**Data generation from write action `record_schema`**: required-field heuristics by name
(`name|title|label ⇒ tag`; `email ⇒ pm-cert+<runid>@example.com`; `url ⇒
https://example.com/pm-cert/<runid>`; numeric ⇒ 1; bool ⇒ false); optional fields unset.
Per-connector overrides in a declarative pairing table:

```go
type WritePairing struct {
    Create, Cleanup string   // "create_issue" -> "close_issue"
    CleanupKind     string   // delete | close | archive
    IDField         string   // e.g. "number"
    VerifyStream    string   // e.g. "issues"
    VerifyField     string   // e.g. "title"
    Overrides       map[string]any
}
```

Default pairing inference: `create_X ↔ delete_X | close_X | archive_X`. **Unpaired mutating
actions are never executed live** (`skipped: no cleanup pairing`) unless a profile supplies one.

**Mechanics per pair** (all via public CLI): write tagged record to local JSONL → file→warehouse
ETL → `pm reverse plan --limit 1` → token from text output → `preview --json` → `run --approve` →
verify → cleanup plan → verify again.

**Write-ahead leak ledger**: before any live write, append `{action, tag, connector, entity_hint,
planned_at}` to `certify-ledger.jsonl`; after verified cleanup append `{tag, cleaned_at}`. Ledger
copied into `.polymetrics/certifications/ledger/` even on crash.

**Failure semantics**: create fails ⇒ stage fail, no leak. Create ok + cleanup/verify fails ⇒
`leaked_resource`: top-level `leaks[]`, `passed=false`, exit 3, LEAKED block printed first.
Cleanup ok but unverifiable ⇒ `leak_unverified` warning.

**Orphan sweeper** (`--sweep`): ledger entries without `cleaned_at` + optional live scan of
VerifyStreams for aged `pm-cert-<slug>-*` tags; cleanup through the same plan/approve/run path.
CI batch jobs always run `--sweep` as a trailing step; nightly sweep job backs it up.

## D) Flow + schedule stages

Flow manifest (`<workdir>/.polymetrics/flows/cert_<slug>.json`): etl step (capture-backed
connection) + dependent query step; assert plan order, preview `dry_run` with zero side effects,
run completed, status shows both steps done.

Schedule: snapshot `crontab -l` → create (cron `0 3 * * *`) → list contains → install `--crontab`
(skip-with-reason if no crontab binary) → sentinel present → remove → sentinel absent AND crontab
byte-identical to snapshot AND manifest file deleted. Residue ⇒ `leaked_schedule` (same severity
class as leaked_resource).

## E) Package layout & integration

```
internal/connectors/certify/
  certify.go        # Runner + Options; per-connector orchestration
  cliharness.go     # in-process cli.Run driver: capture stdout/stderr, parse envelope, assert kind+exit
  stages_source.go  # check/catalog/read/5-mode matrix/resume
  stages_write.go   # create-then-cleanup protocol
  stages_glue.go    # flow, schedule, query, redaction
  pairing.go        # WritePairing tables + data generation
  ledger.go         # write-ahead leak ledger
  sweeper.go        # --sweep
  budget.go         # token bucket + call budget (transport wrapper)
  report.go         # CertificationReport schema, save/load, history, matrix rendering
  credsfile.go      # creds.yaml parsing, env/exec secret resolution
  record.go         # Tier-1 recording RoundTripper + sanitizer
  replay.go         # Tier-1 replay transport + cassette store
internal/cli/certify_cli.go   # wire `certify` into runConnectors
cmd/certstatus/               # enablement generator from artifacts
```

**Tier-1 record/replay**: a tiny `internal/connectors/httpx` package —
`httpx.Client(cfg RuntimeConfig) *http.Client` honoring `base_url` plus a process-global
`httpx.SetTransport(rt)` hook (safe: certify runs in-process; `--record` forces `--parallel 1`).
Cassettes at `internal/connectors/<pkg>/testdata/golden/<stage>/<seq>.json`; matched by (method,
path, per-stage sequence); unmatched request in replay ⇒ failure (catches drift). Sanitizer runs
at record time: secret values (exact/base64/URL-encoded), Authorization/Cookie/X-Api-Key-class
headers, manifest secret fields; tags/emails normalized for deterministic replay. Write stages
replay too — Tier-1 CI exercises the write protocol with zero leak risk.

**Makefile/CI**:

```make
certify-replay: build   # ./pm connectors certify --all --replay --json
certify-live: build     # ./pm connectors certify --all --credentials-file certify/creds.yaml --json
certify-sweep: build    # ./pm connectors certify --sweep --credentials-file certify/creds.yaml
verify: fmt tidy-check vet test build docs-check smoke certify-replay
```

Nightly `certify-live.yml` GitHub Action: matrix over connectors with repo secrets, `--parallel 2`,
always-run sweep step, uploads report artifacts; green runs open a PR refreshing
`internal/connectors/certifications/` + regenerated statuses.

## Implementation order

1. `report.go` + `cliharness.go`; prove against `sample` connector end-to-end.
2. Source stages (5-mode matrix, capture replay, resume) against sample/smoke-test/e2e-test +
   github/httptest. Prerequisite fix: `--credential` on `pm etl check/read` (or live check
   exclusively via `credentials test`).
3. Write protocol + ledger + sweeper, github pairing table first.
4. Flow + schedule stages.
5. CLI wiring (single, `--all`, `--sweep`), exit-code mapping.
6. `httpx` seam + record/replay + sanitizer + `certify-replay` make target.
7. `cmd/certstatus` + enablement guard test.
