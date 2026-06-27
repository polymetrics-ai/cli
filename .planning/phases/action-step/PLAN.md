# PLAN — Action Step (Phase 1)

Each behavior task (B-N) is immediately paired with a test task (T-N).
Test task must be committed red before the implementation. [DOCS] tasks are docs-only.

---

## Wave 0 — Sentinel errors + FlowStep extension

### T-10 [TEST] Action manifest validation tests
- File: `internal/flow/action_test.go`
- Cover: KindAction is accepted by ValidateManifest; action step missing source_table is invalid;
  action step missing destination_connector is invalid; action step missing mappings is invalid;
  action step missing action field defaults to "upsert".
- Must be RED before B-10.

### B-10 [BEHAVIOR] Add KindAction + ActionConfig to manifest + errors
- Files: `internal/flow/manifest.go`, `internal/flow/errors.go`
- Add `KindAction StepKind = "action"`, `ActionConfig` struct embedded in `FlowStep`.
- Add `ErrApprovalRequired`, `ErrSchemaDrift`, `ErrTokenExpired`, `ErrTokenInvalid` sentinels.
- Extend `ValidateManifest` to validate action step fields.
- Gates T-10 green.

---

## Wave 1 — Idempotent writes

### T-11 [TEST] Idempotent write: re-run never duplicates
- File: `internal/flow/action_test.go`
- Use `httptest.Server` that records all received `_pm_id` values.
- Run action step twice with the same source records.
- Assert server received each record exactly once (second run sends 0 records).
- Must be RED before B-11.

### B-11 [BEHAVIOR] Deterministic record ID + idempotency guard
- File: `internal/flow/action.go`
- `deterministicRecordID(flowName, stepID string, record map[string]any) string`
  — SHA-256 of sorted JSON of record primary key fields; hex-encoded.
- In-run sent-ID set; skip records already in set.
- Gates T-11 green.

---

## Wave 2 — Identity mapping

### T-12 [TEST] Identity mapping: external ID persisted across runs
- File: `internal/flow/action_test.go`
- Fake server returns `{"id": "ext-123"}` in response. Run twice.
- Assert identity map file contains `pm_id → "ext-123"` after first run.
- Assert second run skips the record (already mapped).
- Must be RED before B-12.

### B-12 [BEHAVIOR] Identity map store
- File: `internal/flow/action.go`
- `identityMapStore` backed by a JSON file at `<stateDir>/identity_map.json`.
- Load on ActionRunner init; persist after each successful record write.
- On lookup: if pm_id in map → skip (already sent).
- Gates T-12 green.

---

## Wave 3 — Dedupe/merge

### T-13 [TEST] Dedupe: duplicate email/domain records deduplicated before write
- File: `internal/flow/action_test.go`
- Source table has 3 records with 2 unique emails (one email appears twice).
- Assert fake server receives exactly 2 records.
- Must be RED before B-13.

### B-13 [BEHAVIOR] Pre-write dedup on email/domain/external-id
- File: `internal/flow/action.go`
- `deduplicateRecords(records []map[string]any, keys []string) []map[string]any`
  — last-write-wins per dedup key combination.
- Default dedup keys: `["email", "domain", "external_id"]` (any present field).
- Gates T-13 green.

---

## Wave 4 — Rate-limit handling

### T-14 [TEST] 429 → backoff → success
- File: `internal/flow/action_test.go`
- `httptest.Server` returns 429 for first 2 requests then 200.
- Inject fast sleep (zero duration) via `ActionRunner.Sleep` option.
- Assert action completes successfully and server was called 3 times (2 retries + 1 success).
- Must be RED before B-14.

### B-14 [BEHAVIOR] Exponential backoff + jitter on 429
- File: `internal/flow/action.go`
- Inline backoff: base=500ms, max=30s, jitter ±25%; honor `Retry-After` header if present.
- Respect MaxRetries cap (default 3).
- Injectable `Sleep func(context.Context, time.Duration) error` for tests.
- Gates T-14 green.

---

## Wave 5 — Dead-letter queue

### T-15 [TEST] DLQ: permanently failing records quarantined, not silently dropped
- File: `internal/flow/action_test.go`
- `httptest.Server` always returns 400 for a specific record's `_pm_id`.
- After MaxRetries exhausted, assert:
  - `ActionResult.RecordsFailed == 1`
  - DLQ file exists at expected path and contains the failed record
  - `ActionResult.RecordsSucceeded` does not count the failed record
- Must be RED before B-15.

### B-15 [BEHAVIOR] DLQ write on permanent failure
- File: `internal/flow/action.go`
- After MaxRetries exhausted for a record: append to `.polymetrics/dlq/<flow>/<step>/<run>.ndjson`.
- Each line: JSON with `pm_id`, `error`, `attempts`, and redacted `record`.
- `RecordsFailed` incremented; run continues with remaining records.
- Gates T-15 green.

---

## Wave 6 — Schema-drift detection

### T-16 [TEST] Schema drift: breaking change halts before any write
- File: `internal/flow/action_test.go`
- Store a schema snapshot with field `email string`.
- At run time, fake connector reports field `email` as `integer` (type change = breaking).
- Assert `Execute` returns `ErrSchemaDrift` and fake server received 0 records.
- Must be RED before B-16.

### T-16b [TEST] Schema drift: additive field is non-breaking (warning only)
- New field added → run proceeds, no error.

### B-16 [BEHAVIOR] Schema-drift detection + pause
- File: `internal/flow/action.go`
- `schemaSnapshotStore` — JSON file at `<stateDir>/schema_snap_<flow>_<step>.json`.
- On first run: save snapshot. On subsequent runs: compare; breaking change → `ErrSchemaDrift`.
- Breaking = field removed or type changed; additive = warning only.
- Gates T-16, T-16b green.

---

## Wave 7 — Receipts/audit

### T-17 [TEST] Receipt written to ledger after successful batch
- File: `internal/flow/action_test.go`
- Stub ledger (same as flow_test.go `stubLedger`).
- After successful action execution, assert ledger contains at least one receipt entry
  with `Mode=="action"` and `Status=="receipt"`.
- Must be RED before B-17.

### B-17 [BEHAVIOR] Ledger receipt writes
- File: `internal/flow/action.go`
- After each successful batch: call `ledger.Append(ctx, LedgerRecord{Mode:"action", ...})`.
- Redact sensitive fields (no values, only field names).
- Gates T-17 green.

---

## Wave 8 — Engine integration (action dispatch)

### T-18 [TEST] Engine dispatches action step via ActionRunner
- File: `internal/flow/action_test.go`
- Build a FlowManifest with a `sync` step + an `action` step.
- Wire Engine with stub ActionRunner.
- Without approval token: assert engine returns `ErrApprovalRequired`.
- With approval token: assert ActionRunner.Execute called.
- Must be RED before B-18.

### T-18b [TEST] Approval required: no token → step not executed
- Flow with only action steps; Run with no token → `ErrApprovalRequired` before any write.

### B-18 [BEHAVIOR] Engine action dispatch + approval gate
- File: `internal/flow/engine.go` (extend)
- `RunOptions` gains `ApprovalToken string` and `PerAction bool`.
- KindAction dispatch: check token before calling ActionRunner.Execute.
- If no token and !DryRun → return `ErrApprovalRequired`.
- Gates T-18, T-18b green.

---

## Wave 9 — CLI extension

### T-19 [TEST] CLI flow run with --token flag
- File: `internal/cli/flow_cli_test.go` (extend)
- `pm flow run --token abc123 myflow` calls engine with token.
- `pm flow plan myflow` returns a plan with `approval_token` field.
- Must be RED before B-19.

### B-19 [BEHAVIOR] CLI --token + --per-action flags
- File: `internal/cli/flow_cli.go` (extend)
- Parse `--token`, `--per-action` from args.
- `pm flow plan` calls Engine.Run(DryRun=true), returns plan + token.
- `pm flow preview` calls Engine.Run(DryRun=true, Preview=true).
- `pm flow run --token <tok>` executes.
- Gates T-19 green.

---

## Human gates in this phase

- **HUMAN GATE (already cleared per prompt)**: real network writes behind approval gate.
  The HUMAN GATE CLEARED note in the prompt covers this.
- Any new go.mod dependency would require a fresh human gate. None are needed.

## Execution order (wave summary)

| Wave | Tasks | Depends on |
|------|-------|-----------|
| 0 | T-10, B-10 | Phase 0 green |
| 1 | T-11, B-11 | Wave 0 |
| 2 | T-12, B-12 | Wave 1 |
| 3 | T-13, B-13 | Wave 0 |
| 4 | T-14, B-14 | Wave 0 |
| 5 | T-15, B-15 | Wave 4 |
| 6 | T-16, T-16b, B-16 | Wave 0 |
| 7 | T-17, B-17 | Wave 1 |
| 8 | T-18, T-18b, B-18 | Waves 1–7 |
| 9 | T-19, B-19 | Wave 8 |
