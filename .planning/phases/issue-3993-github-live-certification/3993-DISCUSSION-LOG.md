# Phase 3993: github-live-certification - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in `3993-CONTEXT.md` — this log preserves the alternatives considered.

**Date:** 2026-08-11
**Phase:** 3993-github-live-certification
**Areas discussed:** custody, fresh lineage, credential safety, process/rate provenance, dependency boundary, evidence bar

---

## Custody and lineage

| Option | Description | Selected |
|--------|-------------|----------|
| Reuse #3993 | Continue under the existing issue and PR while preserving history | ✓ |
| Create a duplicate child | Split current-SHA closure into a second overlapping owner | |
| Continue as correction 6 | Re-open the exhausted five-round historical ledger | |

**User's choice:** The ship brief explicitly names #3993, PR #4061, its branch/base, and authorizes a fresh lineage 1/5.
**Notes:** #3993’s current sub-issue list contains no separate exact owner; the historical correction sequence remains immutable.

---

## Provider safety and certification evidence

| Option | Description | Selected |
|--------|-------------|----------|
| Approved App/run-owned boundary | Full-parity GitHub App only, immutable `Polymetrics-Cert` target, typed cleanup | ✓ |
| Personal or weaker credential | Substitute a token or unrelated repository for live proof | |
| Static/historical proof | Treat fixtures or prior provider runs as current-SHA certification | |

**User's choice:** Use actual credentials only under the approved boundary; never disclose a secret; stop as `needs-decision` when safe authority is absent.
**Notes:** Every mutation is independently read back, inversed, idempotently cleaned, and followed by empty-residue enforcement.

---

## Process and rate provenance

| Option | Description | Selected |
|--------|-------------|----------|
| Built one-process proof | Use a fresh built `pm` path which keeps the rate coordinator in process | ✓ |
| External child fan-out | Call one standalone `pm` process per operation behind a Node barrier | |
| Sequentialize silently | Hide local saturation by changing the measured workload | |

**User's choice:** One built binary/process where #3993 requires in-process coordination; existing external runner is not equivalent evidence.
**Notes:** Rate evidence must report observed admission and provider state without exhausting shared credentials.

---

## Dependencies and non-scope

| Option | Description | Selected |
|--------|-------------|----------|
| Continue independent GitHub work | Record exact blocks from #4059/#3994/#3992 and proceed elsewhere | ✓ |
| Copy transport/flow code | Reimplement another issue’s incomplete foundation here | |
| Rebase repeatedly | Chase #4060 while it is settling | |

**User's choice:** Continue independent branch-local work; do one final parent update only after #4060 lands.
**Notes:** The parked Shepherd work is explicitly excluded.

---

## the agent's Discretion

- Apply manual inline GSD because the project contract forbids spawning GSD roles in this issue path.
- Use focused validation before requesting a heavy live validation window.
- Fix actionable current-code scan findings within this PR’s owned scripts/tests only.

## Deferred Ideas

- Full concurrent all-surface live execution awaits complete lab fixtures and executable dependency foundations.
- Real action-flow scheduling waits for #3992 after #3994/#4059 supply the action path.

---

*Phase: 3993-github-live-certification*
*Discussion log generated: 2026-08-11*
