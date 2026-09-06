# Retained source and lane report

`connectorgen source-lanes` accounts for retained provider operations before
execution admission. It emits an authoring-only report. Runtime never reads the
report, its cohort inventory, semantic annotations or behavioral-proof records.
The schema-4 source-lock → canonical descriptor → closed execution generation
path remains the only execution-authoring pipeline.

The Batch One cohort has 4,341 primary operations and two separate GitLab
supplements. Its report contains 4,343 source rows and 30,401 cells. These are
source-accounting totals, not executable capability counts. Root `primary`,
`supplement` and `operations` totals describe the anchored expected universe;
`observed_primary`, `observed_supplement` and `observed_operations` separately
count verified retained rows. Missing inputs reduce observation, never membership. The inventory keeps
provider IDs verbatim and identifies each operation by connector, inventory and
ID; command names and canonical execution-unit IDs do not define membership.

## Generate and check

```bash
go run ./cmd/connectorgen source-lanes --help
go run ./cmd/connectorgen source-lanes > candidate-source-lanes.json
go run ./cmd/connectorgen source-lanes --check --manifest candidate-source-lanes.json
go run ./cmd/connectorgen source-lanes --check
```

Capture generation into a **fresh candidate**, check its exit status and validate
that candidate before replacing the tracked
`data/connector-canon/batch1-source-lane-manifest.json`. Do not redirect a possibly
failing generation over the tracked report. A failed run can emit a complete
invalid report for investigation; those bytes are not a replacement candidate.

`--repo <dir>` selects the authoring root. `--manifest` is accepted only with
`--check` and must name a confined relative regular file. Check mode regenerates
from current inputs, compares source identity, facts, cells, references,
diagnostics and summaries, and requires exact deterministic bytes. It does not
create files, stage a generation, acquire publication locks, recover journals or
execute receipt commands.

Exit 0 means report consistency. Exit 1 means invalid retained input/report or
an I/O failure. Exit 2 means invalid arguments. A missing source preserves its
anchored rows and seven cells with errors. An invalid cohort anchor cannot
supply trustworthy membership, so it produces an error without a fabricated
report. A short or failed stdout write fails the command.

## Inputs and evidence boundaries

- `data/connector-canon/batch1-source-lane-cohort.json` independently pins exact
  source sets, retained lock hashes and raw artifact hashes. Generation never
  rewrites this anchor from whichever operations happened to load.
- Retained archives under `internal/connectors/defs/*/sources/` may describe
  provider operations without executable forms. Their historical envelope is
  parsed only by this authoring report. They are not predecessor execution-lock
  readers or runtime fallbacks.
- `batch1-source-lane-annotations.json` contains source-cited semantic decisions
  and intended/current target claims. Unknown or duplicate source keys,
  contradictory semantics and wrong present targets remain errors.
- `batch1-source-lane-proofs.json` contains optional scoped behavioral records.
  The real set remains empty until an original per-source/lane receipt and its
  assertion mapping have been reviewed. A proof record cannot approve itself.
  Current loader capacity is undergoing reconciliation against the full cohort;
  an empty file does not establish that the loader is complete.

A retained-document record separates its actual file hash/size from an upstream
hash/size declared in that file. `upstream_bytes_verified` is true only when the
corresponding raw bytes were read and verified. Each source row points into its
retained document; the full original node is not duplicated in the row. Shared
JSON/YAML facts use exact local pointers and value hashes. Output uses compact
JSON and one trailing newline. Rendered HTML references retain verified heading
sections and text citations; they do not invent request/response schemas
from prose. External references are not fetched.

## Seven separate cells

Every source row retains direct read, direct write, binary download, binary
upload, ETL, reverse ETL and sync transport. Applicability is separately
`applicable`, `not_applicable` or `undetermined`.

| State | Meaning |
| --- | --- |
| `not_applicable` | Retained source facts support an explicit exclusion. |
| `missing_foundation` | A source-backed unmet contract has an existing owner/gap reference. This does not authorize a receiver or another foundation. |
| `mapped_unproven` | Facts, exact materialization or lane-specific behavior are still unresolved or unproven. |
| `implemented` | Applicable source semantics, all required independently validated targets, and current reviewed lane-specific behavior agree. |

Artifact existence, engine loading, a test symbol, preflight, HTTP success and
counts cannot establish the last state. A fixture proves its declared fixture
scope; it cannot certify a real connector or provider-live behavior. Source
exclusions and existing foundation decisions remain independent of proof
availability. Invalid asserted claims remain errors even if their cells are
unproven.

The report is not the separate source-role/destination-role/mode matrix. It does
not approve credentials, provider exercises, receiver exposure, a relay or a
scope reduction. The existing checkpoint and independent-review gates still
apply.
