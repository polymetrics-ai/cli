# Inline code review — Issue #3975 committed-transaction staging and durable receipts

## Verdict

**PASS after one bounded hardening correction.** The review is an explicit
manual `code-review` fallback: #3975 has no numbered roadmap phase and the
canonical delivery contract permits one worker only, so no reviewer role was
spawned.

## Reviewed boundaries

| Area | Result |
| --- | --- |
| Private stage | Opaque transaction values are hashed before path construction; active chunks have no receiver API and are removed on abort/quota/cancellation. |
| Bounds | Fixed-size stream copies, finite transaction/root limits, reservations for concurrent writes, recovered accounting, and saturating untrusted count diagnostics are fail-closed. |
| Atomicity | Chunks and manifests use temp write → file sync → rename → directory sync. The sealed manifest is verified before any receiver sees data. |
| Whole delivery | The receiver is called once with a transaction stream; it must consume every chunk exactly once before a downstream receipt can be accepted. |
| Receipt/ack | Receiver success alone is insufficient. Receipt bytes are atomically persisted and directory-synced before the private durable marker can adapt to `synccontract.DownstreamAcknowledgement`. |
| Failure/recovery | Begin/append/seal faults discard incomplete data; receiver/receipt failures leave only sealed replay work; a durable receipt survives cleanup failure and startup removes its residual stage. |
| Scope | No PostgreSQL, source feedback, checkpoint, descriptor/admission, generic write, target, CLI, or connector capability behavior changed. |

## Correction and disposition

Review found one bounded safety issue: diagnostic arithmetic for a hostile
`int64` record count could wrap before reporting a named quota refusal. The
implementation now uses limited/saturating addition, and
`TestCommittedTransactionStageSaturatesUntrustedRecordQuotaDiagnostics` proves
the fail-closed result. This is correction round **1/5**, not a newly
discovered repository-gate defect; therefore no extra child issue is required.

After the correction, `gofmt`, `git diff --check`, changed-package lint,
`go vet ./...`, focused/race suites, affected `synccontract`/`app`, build, and
the individual repository gates are green. No remaining actionable source,
security, or scope finding is open.
