# Summary — certification batch scaling

## Outcome

- Measured live GitHub direct-read throughput at 10 operations and three independent 100-operation cohorts.
- The 100-operation mean is 154.831 seconds / 1.548 seconds per target operation; no measured rate-limit slowdown occurred.
- Found 14 individual product defects in the 100-operation workload, separately from six provider-evidenced missing fixtures and 47 provider refusals.
- Re-read the current sweep: 639 implemented direct reads project to 16 minutes 30 seconds from the measured curve.
- PR #4215 merged while this work ran. The scope contract is now present, but generic GitHub live-evidence importing remains owned by open PR #4216; staged evidence remains unaccepted rather than hand-authored.

## Scope control

This lane changes no connector/runtime/generated surface. It adds reproducible measurement harnesses and durable GSD evidence only. All live work was read-only, serial, disposable-credential execution.

## Handoff

- Read `LIVE-RESULTS.md` for the curve, failure taxonomy, throttle evidence, projection, and rulebook rules.
- Import `STAGED-LIVE-REPORT.json` only through PR #4216's generic importer after it lands; do not promote the sample to `full_parity`.
