# THREAT-MODEL — Action Step (Phase 1)

## Trust boundary

The action step is the first component that can send data to an external system.
All writes must flow through: `plan → preview → approval-token → execute`.

## Threat inventory

### TH-01 — Unapproved write
**Description**: Engine executes an action step without a valid approval token.
**Mitigation**: `ErrApprovalRequired` sentinel returned before ActionRunner.Execute is
called. Token validated by hash comparison (SHA-256 of random 18-byte token, same pattern
as `internal/app` ReversePlan). Engine never calls Execute without a matching token hash.
**Residual risk**: None if token validation is correctly placed in engine dispatch.

### TH-02 — Approval token replay / theft
**Description**: Attacker replays a consumed or stolen approval token.
**Mitigation**: Token is single-use (hash zeroed after consume). Token expires after 24h
(same as ReversePlan). Token is never logged or printed in structured JSON output; only
shown once on plan creation (redacted thereafter).

### TH-03 — Duplicate sends (idempotency failure)
**Description**: Retry or re-run causes the same record to be sent twice.
**Mitigation**: Deterministic `_pm_id` guards in-run and across-run (identity map).
Both layers must independently block the duplicate.

### TH-04 — Secret leakage in DLQ / ledger
**Description**: Credential values written to DLQ files or ledger entries.
**Mitigation**: DLQ records store only `pm_id`, `error`, `attempts`, and a redacted record
(field names only, no values for fields declared as secrets). `safety.RedactErrorText`
applied to all error strings before write.

### TH-05 — SSRF via destination_config base_url override
**Description**: Attacker supplies a `base_url` in `DestinationConfig` pointing to an
internal service.
**Mitigation**: `internal/safety` SSRF validation applied to any URL resolved from
`DestinationConfig`. Block RFC-1918 addresses and loopback except in test mode.

### TH-06 — Schema drift exploitation (partial write)
**Description**: Destination schema silently accepts a mismatched field type, corrupting
existing data.
**Mitigation**: Schema-drift detection compares live catalog to stored snapshot before
any write. Breaking change → `ErrSchemaDrift`, zero records sent. Snapshot updated only
after a successful run.

### TH-07 — Unbounded goroutine growth
**Description**: Batched sends with unbounded concurrency exhaust file descriptors.
**Mitigation**: Worker pool bounded by `BatchSize` (default 100, max 1000). Implemented
with a semaphore channel pattern (Go stdlib only, no external library).

### TH-08 — DLQ unbounded growth
**Description**: Permanent failures fill disk with DLQ files.
**Mitigation**: DLQ files are per-run (bounded by run record count). Operator runbook
documents cleanup procedure. Future: DLQ size limit flag.

## Gates not triggered

- No new network endpoints (reuses existing connector resolution).
- No new credentials (reuses vault).
- No new go.mod dependency.
