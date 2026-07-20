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
5. **Gotchas**: `directConnector` builds RuntimeConfig from `--config` only — `pm etl check/read
   --connector` never resolves credential Secrets. Live certification creates an isolated
   credential from `--from-env` references and tests it through the vault-backed public commands.
   `pm reverse plan --json` deliberately strips the approval token — the harness parses the token
   from text output and separately asserts the JSON path keeps hiding it (redaction gate).
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
Each mode's report entry records `data_source: live | capture`. The current runner executes this
fixed matrix; it does not expose a mode-selection control.

Constraints: writes/deletes only via create-then-cleanup with tagged ephemeral records + mandatory
cleanup verification + orphan sweeper. Single-connector writes require explicit `--write`; batch
writes require both a credential-file `write: true` entry and `sandbox: true`. User
`--write=false` or `--skip=write` always disables batch writes. Credential-file rate, budget, and
read-limit settings are rejected before a runner starts because the current runner cannot enforce
them. Parallelism is across connectors only; no credential means `uncertified`, never `failed`.

### Tiers

| Tier | What | When | Credentials |
|---|---|---|---|
| 0 | Fixture conformance (conformance v2) | every CI run | none |
| 1 | Recorded HTTP replay: sanitized golden cassettes | every CI run, deterministic | none |
| 2 | Live certification: full command surface + create-then-cleanup writes + flow + schedule | on demand / nightly | gated |

## A) Command spec

```
pm connectors certify <connector>
    [--from-env field=ENV ...] [--config k=v ...]
    [--stream <name>]            # default: first cursor stream, else first
    [--skip write] [--write] [--full] [--keep-workdir] [--json]
pm connectors certify --all --credentials-file creds.yaml
    [--parallel N] [--resume] [--write=false | --skip=write] [--json]
pm connectors certify --sweep [--credentials-file creds.yaml] [--older-than 24h] [--json]
```

Controls not implemented by the runner are hidden and rejected before effects rather than accepted
as no-ops. Mode-specific and unknown controls are likewise rejected before credential loading or
runner/sweep effects. Sweep `--older-than` must be greater than zero and no more than 8760h.

**Exit codes before a report completes**: 1 setup/runtime error · 2 usage error · 3 validation
error. **Completed report exits**: 0 pass · 2 certification failure · **3 leaked resources**
(dominates everything). Wire completed outcomes via typed `certify.ExitError` in
`internal/cli/errors.go`.

**Execution model**: `os.MkdirTemp` root → `pm init --root <dir>` → every stage is an in-process
`cli.Run([..., "--root", dir, "--json"], &out, &err)` whose exit code and envelope `kind` are
asserted → leak verification → report written → workdir deleted.

### Stage list

Source stages: 0 preflight (registry, secrets present, write gates armed) · 1 fixture_conformance
(Tier-0 embedded) · 2 manual_json (`connectors inspect --json`, no secret values) · 3
credentials_add/test (LIVE check via vault) · 4 catalog_live (≥1 stream; PK/cursor recorded) · 5
etl_full_refresh_append (LIVE; records_read>0 or `passed_empty` warning; output JSONL = capture) ·
6 etl_full_refresh_overwrite (capture; run twice; truncate semantics via `pm query`) · 7
etl_full_refresh_overwrite_deduped (capture; no duplicate PK tuples) · 8 etl_incremental_append
(LIVE; cursor set) · 9 resume (LIVE run 2; records_read(run2) ≤ run1; cursor monotonic; no row
below run-1 checkpoint; assert the captured outbound request carried the cursor param) · 10
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
  "effective_options_fingerprint": "sha256:...",
  "started_at": "...", "completed_at": "...",
  "mode": "live",
  "credential_ref": "cert-source",
  "passed": true,
  "leaks": [],
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

`creds.yaml` (environment-variable references only; never secret values):

```yaml
version: 1
defaults: {parallel: 4}
connectors:
  github:
    credential:
      from_env: {token: PM_CERT_GITHUB_TOKEN}
      config:   {repository: polymetrics-ai/cert-sandbox}
    sandbox: true
    write: true
  stripe:
    credential:
      from_env: {api_key: PM_CERT_STRIPE_API_KEY}
    write: false
  salesforce:
    skip: true
    reason: "no sandbox tenant yet"
```

Credential files are bounded to 1 MiB, decoded with known fields only, require version 1 and at
least one locally registered connector, validate connector/environment references, reject symlinks
and secret-schema values under `config`, and reject `exec` before runner effects; certification never
executes an external credential command. Worker pool concurrency is across connectors, is limited to
1–32 when explicit, and never exceeds queued connectors; stages within one connector remain strictly
serial. Resumability markers are written to `certifications/batch-<runid>/progress.json`;
`--resume` reuses a completed prior report only when its exact schema, connector manifest, and
secret-free effective-options/environment-reference fingerprint match. Summary matrix rows are connectors; columns are
check/catalog/read/5 modes/resume/write/flow/schedule/redaction/**leaks**; any leak row prints first
and forces exit 3.

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
ETL → `pm reverse plan --limit 1` → token from text output → successful `preview --json` →
`run --approve` → verify → cleanup plan → successful cleanup preview → approval/run → verify again.

**Write-ahead leak ledger**: before any live write, append connector/run/tag plus curated create,
cleanup, verification, and timestamp provenance directly to
`.polymetrics/certifications/ledger/<connector>/certify-ledger.jsonl`; after verified cleanup append
`{tag, cleaned_at}`. Appends are synced before mutation. The durable layout is the exact layout a
fresh-process sweeper consumes; ephemeral workdir deletion cannot disconnect recovery authority.

**Failure semantics**: create fails ⇒ stage fail, no leak. Create ok + cleanup/verify fails ⇒
`leaked_resource`: top-level `leaks[]`, `passed=false`, exit 3, LEAKED block printed first.
Cleanup ok but unverifiable ⇒ `leak_unverified` warning.

**Orphan sweeper** (`--sweep`): ledger entries without `cleaned_at` + optional live scan of
VerifyStreams for aged `pm-cert-<slug>-*` tags. It rejects connector/run/tag/action provenance that
does not match a curated pairing, provisions only validated environment references in a fresh
temporary workspace, and cleans through plan → successful preview → approval → run. CI batch jobs
always run `--sweep` as a trailing step; nightly sweep backs it up.

## D) Flow + schedule stages

Flow manifest (`<workdir>/.polymetrics/flows/cert_<slug>.json`): etl step (capture-backed
connection) + dependent query step; assert plan order, preview `dry_run` with zero side effects,
run completed, status shows both steps done.

Schedule: bind every in-process invocation to an invocation-local temporary crontab file → snapshot
that file → create (cron `0 3 * * *`) → list contains → install `--crontab` → sentinel present →
remove → sentinel absent AND the file byte-identical to its snapshot AND manifest deleted. The
system crontab backend is unreachable, including during parallel certification. Residue ⇒
`leaked_schedule` (same severity class as leaked_resource).

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
  report.go         # CertificationReport schema, save/load, history, matrix rendering
  credsfile.go      # creds.yaml parsing, env references, prohibited-exec rejection
  record.go         # Tier-1 recording RoundTripper + sanitizer
  replay.go         # Tier-1 replay transport + cassette store
internal/cli/certify_cli.go   # wire `certify` into runConnectors
cmd/certstatus/               # enablement generator from artifacts
```

**Tier-1 record/replay internals**: `record.go` and `replay.go` provide sanitized cassette
primitives for tests and internal harness use. They are not exposed as public certify controls.
Sanitization removes secret values (exact/base64/URL-encoded), credential-class headers, and
manifest secret fields; unmatched replay requests fail closed.

**Local and CI invocation**:

```bash
./pm connectors certify sample --json
./pm connectors certify --all --credentials-file certify/creds.yaml --json
./pm connectors certify --sweep --credentials-file certify/creds.yaml
```

The dependency-free sample command is suitable for local/CI verification. Credential-file batch
and sweep commands remain explicit, credential-gated operations and are not part of the default
verification target.

## Implementation order

1. `report.go` + `cliharness.go`; prove against `sample` connector end-to-end.
2. Source stages (5-mode matrix, capture replay, resume) against sample/smoke-test/e2e-test +
   github/httptest. Live checks go through the isolated credential lifecycle and `credentials test`.
3. Write protocol + ledger + sweeper, github pairing table first.
4. Flow + schedule stages.
5. CLI wiring (single, `--all`, `--sweep`), exit-code mapping.
6. `httpx` seam + record/replay + sanitizer + `certify-replay` make target.
7. `cmd/certstatus` + enablement guard test.
