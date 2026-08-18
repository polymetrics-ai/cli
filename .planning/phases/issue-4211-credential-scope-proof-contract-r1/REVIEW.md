# Deep review — Issue 4211 provable credential-scope contract

**Route:** inline/manual deep review. The active runtime forbids the GSD command's
reviewer subagent, so the changed construction, validation, consumer, generated
surface, test, and re-issued-evidence paths were inspected together.

## Review focus

1. **Construction boundary:** no caller-provided credential-scope boolean,
   free-form scope, note, or proof survives. The only full claim construction
   requires `Report.FullParityVerified()`; the default path is bounded by the
   serialized protocol transcript.
2. **Validation boundary:** the former full-parity string validator is retained
   in the full-claim path. The added validator binds both supported scopes to an
   exact proof discriminator and rejects schema v1 accepted evidence.
3. **Reader/consumer parity:** certification-matrix pointers preserve the proof
   discriminator and `internal/agentcontract` independently validates it and
   includes it in pointer-to-record equality.
4. **Evidence truthfulness:** every one of the fourteen freshly generated
   PostgreSQL records and its matrix pointer is v2
   `observed_operations/protocol_exchanges`; no record is silently stamped full
   parity.
5. **Safety:** proofs retain only existing redacted fingerprints, no credentials
   or response bodies. The shared Go changes contain no connector-specific
   identifier; `make connector-boundary` passes.

## Findings

No critical, warning, or informational findings remain after the cross-file
review. The renamed historical narrow-credential test now explicitly documents
the required behavior change rather than appearing to weaken or skip a prior
guard.

## Evidence

- `git diff HEAD --check` passed.
- Targeted tests, consumer tests, matrix/sweep checks, and the fresh PostgreSQL
  evidence re-issue are recorded in `VERIFICATION.md`.
