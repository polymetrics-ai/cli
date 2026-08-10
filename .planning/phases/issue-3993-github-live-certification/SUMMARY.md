# Issue #3993 — GitHub live certification measurement

## Delivered harness work

- The case generator and live runner take one immutable, run-owned
  `Polymetrics-Cert` organization/repository boundary.  The historical personal
  owner, repository, and commit SHA no longer influence the live ledger.
- The runner validates that boundary before credential use, releases every
  eligible operation through one barrier, requires an independent read-back for
  any write case, bounds every child process at 45,000 ms, and redacts output
  before it can reach a report.
- Current manifests are source-derived: `github-live-lab-manifest.mjs --check`
  reports 1,521 rows (900 repository, 308 organization, 33 App/install, 280
  feature/entitlement) and `github-live-bootstrap-probes.mjs --check` passes.

## Reclassified live ledger

The fresh immutable-boundary ledger has **665 attemptable** and **856 blocked**
rows, a movement of **+505 / -505** from the frozen R1 160 / 1,361 result.
This is a reclassification, not preservation of the frozen pre-skip policy.

| Reason family | Frozen R1 blocked | Fresh boundary blocked | Movement |
| --- | ---: | ---: | ---: |
| mutation outside pinned repository | 748 | 0 | -748 |
| organization or enterprise | 175 | 0 | -175 |
| secret material | 62 | 85 | +23 |
| App authentication | 23 | 0 | -23 |
| binary resource | 9 | 4 | -5 |
| no cleanup-safe fixture | 37 | 767 | +730 |
| retained historical target (reported separately in R1) | 3 | 0 | -3 |
| other | 304 | 0 | -304 |
| **total** | **1,361** | **856** | **-505** |

Every family moved.  The 767 cleanup-safe-fixture blocks are a current finding:
they are no longer falsely attributed to the old pinned repository, but they
still have no declared fixture/read-back/inverse-cleanup lifecycle.

## Credentialed full-surface result

The App installation token authenticated successfully and a built `pm` binary
read back the single run-owned repository before and after the sweep.  The
whole applicable surface was then released once:

- launch: `single_barrier_release`, 665 operations (15 ETL, 639 direct reads,
  11 binary downloads);
- result: **0 proven, 665 failed, 856 untestable** across all 1,521 commands;
- failed cause: all 665 reads exceeded the 45,000 ms terminal bound, with no
  provider HTTP status available;
- rate snapshots taken immediately around that barrier: REST core 15,000 →
  14,997 (3 consumed); GraphQL 5,000 → 5,000 (0 consumed).

This is not a credential or quota-policy result hidden behind a harness
pre-skip.  A single App-authenticated `repo view` returned one record, and a
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

- A provider-admission solution must let the 665 barrier-released operations
  reach GitHub without turning this result into a sequential sweep.  That is
  separate from the R2 rate telemetry and must not be faked as a quota result.
- The 767 mutation cases need declared cleanup-safe fixtures and independent
  provider read-backs before they become attemptable.
- The warehouse → GitHub action leg remains blocked by the shared approved
  action path (#3994) and real schedule firing (#3992).  No GitHub-specific
  approval or schedule implementation was added.

GitHub remains **uncertified**.  The 0/665 result is retained as the truth.
