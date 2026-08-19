# Issue #3993 — GitHub live certification measurement

## Delivered harness work

- The case generator and live runner take one immutable, run-owned
  `Polymetrics-Cert` organization/repository boundary.  The historical personal
  owner, repository, and commit SHA no longer influence the live ledger.
- The runner validates that boundary before credential use, releases every
  eligible operation through one barrier, requires an independent read-back for
  any write case, rejects direct and `--config` owner/repository overrides
  before a child starts ([#4020](https://github.com/polymetrics-ai/cli/issues/4020)), bounds every child process at 45,000 ms,
  and redacts output before it can reach a report.
- Current manifests are source-derived: `github-live-lab-manifest.mjs --check`
  reports 1,521 rows (900 repository, 308 organization, 33 App/install, 280
  feature/entitlement) and `github-live-bootstrap-probes.mjs --check` passes.
- The legacy rate-limit artifact is retired to a case-file-bound historical
  observation. It explicitly says current certification is not proven and
  `rate-limit get` remains untestable; bootstrap artifacts are guarded before
  hashing, validation, serialization, and output.

## Final-gate correction history

The completed sequence is `1/#4020 -> 2/#4022 -> 3/#4027 -> 4/#4039 ->
5/#4050`. Round 5 completed only after its three synthetic RED regressions
were observed and the focused 50-test GREEN suite passed. None of the five
corrections changed the historical terminal measurement or implemented
#3990/#3994/#3992 behavior.

## Current canonical classifier

The ordered current classifier contains **1,521** commands: **182 attemptable**
and **1,339 blocked**. Direct read accounts for **169 attemptable / 470
blocked** rows. Its digest, totals, and manifest readiness are derived from the
same case list before a child process, hash, or artifact write can occur.

The static comparison is **-483 attemptable / +483 blocked** against the
separately named historical terminal measurement's 665 failed / 856 untestable
rows. It is not a new provider measurement, and no legacy executable-input
ledger can override it. The current 1,339 blocked rows include every write
without a typed fixture, independent read-back, and inverse cleanup contract.

## Historical terminal-timeout measurement

The App installation token authenticated successfully and a built `pm` binary
read back the single run-owned repository before and after the previously
recorded full-surface sweep. That historical measurement released 665 operations
once:

- launch: `single_barrier_release`, 665 operations (15 ETL, 639 direct reads,
  11 binary downloads);
- result: **0 proven, 665 failed, 856 untestable** across all 1,521 commands;
- failed cause: all 665 reads exceeded the 45,000 ms terminal bound, with no
  provider HTTP status available;
- rate snapshots taken immediately around that barrier: REST core 15,000 →
  14,997 (3 consumed); GraphQL 5,000 → 5,000 (0 consumed).

This historical result is not a credential or quota-policy result hidden behind
a harness pre-skip. A single App-authenticated `repo view` returned one record, and a
separate 16-way simultaneous control returned 16 successes.  The 665-child
all-at-once launch therefore saturated local process admission before the
provider buckets were meaningfully exercised.  It is a #3993 harness scaling
finding, not evidence that #3990 REST/GraphQL policy denied the calls.  The
runner intentionally retains the required single barrier release; no shared
coordinator or rate-policy implementation was added here.

The sanitized R2 proof report, fresh case ledger, rate snapshots, inbound
evidence, and boundary read-back were scanned for token-shaped text before the
temporary run project was deleted.  No credential value appears in a committed
artifact, fixture, report, or transcript.

## Warehouse-mediated inbound proof

The inbound leg is proven with the built binary and one GitHub connector
definition:

1. `pm flow plan` and `pm flow preview` accepted a one-step GitHub repository
   sync into the local warehouse.
2. `pm flow run` read **1** GitHub record and wrote **1** warehouse record.
3. `pm query` returned **1** row from the connection-scoped table through
   DuckDB, and the materialized table had the Parquet `PAR1` signature.

An initial disposable warehouse path was deliberately rejected by its
read-back because the query surface owns the canonical project warehouse root.
The proof was rerun against that canonical root; the failed read-back was not
treated as evidence of a round trip.

No GitHub mutation was dispatched: all live sweep operations were reads and
the flow's only external leg was a source read.  Consequently there were no
provider objects to inverse-clean up.  The final repository read-back returned
exactly one record, and all local warehouse, flow, credential, and report state
lived below the run-owned temporary directory and was removed after evidence
sanitization.

## What remains before certification

- A provider-admission solution must preserve the single barrier semantics if a
  future credentialed measurement is deliberately authorized; it must not turn
  the historical 665-child result into a sequential sweep or a quota claim.
- The currently blocked cases need declared cleanup-safe fixtures and
  independent provider read-backs before any write becomes attemptable.
- The warehouse → GitHub action leg remains blocked by the shared approved
  action path (#3994) and real schedule firing (#3992).  No GitHub-specific
  approval or schedule implementation was added.

GitHub remains **uncertified**.  The 0/665 result is retained as the truth.
