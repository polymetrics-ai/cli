# Twenty CRM delivery summary

## Outcome

Recovered and reconciled the Twenty CRM declarative bundle for issue #277.
The bundle now exposes 168 executable CLI commands: 28 ETL lists, 28
operation-backed direct gets, and 112 typed reverse-ETL actions, including 28
typed destructive deletes. Twenty is intentionally not promoted to the
separate certification allowlist.

## Live proof

An isolated, disposable published-Compose instance pinned to
`twentycrm/twenty@sha256:cb80b05bc2619a88a3a83293f45f2be495a55ac77a90946fa1f7d85f0b7fde24`
was seeded headlessly. Firstmate directed the seed/key path; the captain's
standing requirement was the real-instance proof itself. The image's emitted
444-byte key streamed directly into the CLI credential vault with
`--value-stdin`; no secret entered argv, a file, evidence, or source control.

The built CLI authenticated, listed 600 companies and 1,000 people over
provider pagination, bounded-listed all 28 object groups, used direct gets
where seeded records existed, created a company, read its matching name,
updated and reread it, then deleted it through plan/preview/approval/typed
confirmation. The created record was absent afterward and its get returned
404. All mutation runs reported one staged, one succeeded, zero failed.

The explicitly named Compose project was then brought down with volumes, and
the disposable directory was moved to Trash. No captain-owned instance was
contacted.

## Validation and review

`VERIFICATION.md` contains the exact command set and results. The final
rebased-head full `internal/cli` package run passed in 596.926s, golden transcripts passed, and
the freshly rebuilt binary completed a count-only authenticated live read of
three records. Manual code review found no actionable issues; automatic Claude
review is requested by opening the non-draft PR.

## Typed-destination reconciliation

PR #4304's published head `d814875a902be684cb2a38b94f7a8077f66b70b1` is merged
into this local branch and GitHub API confirms PR #4298 targets
`fm/cli-reverse-etl-destination-r1` at that exact SHA. `sync_transport.json` is a
connector-owned declaration: all 28 REST streams are eligible as a bounded
full-append source and `create_companies` is the only destination action. Its
closed mapping copies only a company's `name` into the action's `name` field;
the declaration names durable acknowledgement and never provides a
caller-selected endpoint or body template. The refreshed foundation persists a
selected action, but its source-binding lookup has one mapping per executor and
stream. It cannot preserve Twenty's distinct mappings for the other 55
record-shaped actions, so this is not an application-deployable all-ops
reverse-ETL path.

`write_eligibility.json` closes the one-action accounting gap without
misrepresenting it as complete transport coverage. It records one currently
bound action, 55 schema-intersecting record actions blocked on a
provider-neutral per-action source-binding capability, 28 batch actions whose required
`records` array cannot be formed by the single-record contract, and 28 deletes
whose tombstone workset is incompatible with its no-tombstone delivery.
These dispositions do not alter CLI reachability: all 112 typed actions remain
implemented. Safety, privilege, and destructive labels remain execution gates,
not eligibility reasons.

`SEVEN-SURFACE-LEDGER.json` is the machine-readable reconciliation. It records
all 168 source-locked REST operations as declared with zero exclusions: 28 ETL
reads, 28 operation-backed direct reads, and 112 typed reverse-ETL actions.
The published source contract exposes neither file-transfer nor direct-write
operations in the source-locked REST inventory, so both binary surfaces and
direct write are truthfully zero. Twenty's GraphQL and Metadata APIs are
workspace-schema-generated and are not inferred from that REST contract.
The no-tombstone destination does not weaken CLI reachability: 28 destructive
delete actions remain implemented, approval-gated, and typed-confirmed.

`FOUNDATION-HANDOFF.{md,json}` now provides the separate foundation lane with
source hashes, descriptor/App call-path evidence, the 55/28/28 exact
membership selectors, and a committed connector-local red behavioral witness.
It preserves the acceptance criteria and makes no shared engine change.
