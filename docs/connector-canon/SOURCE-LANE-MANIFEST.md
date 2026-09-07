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
count verified retained rows. Missing inputs reduce observation, never membership.
The inventory keeps provider IDs verbatim and identifies each operation by connector,
inventory and
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
  Shared reviewed input sets avoid repeating the full dependency closure for
  every record; an empty file does not establish behavioral coverage.

A confirmed receiver-demand annotation must join the source-owned authoring
catalog's exact retained operation, file identity, registration and request/event
citations, actual Atlas owner and complete duplicate-free decision-owner set.
The catalog records the existing demand attribution; receiver implementation and
exposure decisions remain pending. Copied labels cannot create another confirmed
gap. With no gap annotation, registration remains unresolved for CP13. The report
reads and pins `foundations/catalog.json` once when a gap is asserted; runtime
execution does not consume this input or the demand catalog.

A retained-document record separates its actual file hash/size from an upstream
hash/size declared in that file. `upstream_bytes_verified` is true only when the
corresponding raw bytes were read and verified. Each source row points into its
retained document; the full original node is not duplicated in the row. Shared
JSON/YAML facts use exact local pointers and value hashes. Output uses compact
JSON and one trailing newline. Rendered HTML references retain verified heading
sections and text citations; they do not invent request/response schemas
from prose. External references are not fetched.

Each fact envelope states `coverage_confidence` as `machine_readable_snapshot`,
`rendered_reference` or `partial`, with `completeness_limits` and its exact `refs`
as the retained basis. These labels describe available evidence; even a fully
normalized snapshot makes no claim about the current provider universe.

The validator checks copied group values and citation hashes against retained
JSON or rendered document values. The retained row and operation determine
which metadata, request, response and descriptive groups must be present;
deleting both a group and its citation cannot hide those operation facts.
The same check covers effective security (including an explicit empty operation
override), security schemes, root webhooks, retained path parameters, and the
source envelope's bridge, event and batch inventories. Shared values remain
anchored to their original document locations.
Effective parameter nodes, locations, names and requiredness must match the
existing local-reference projection of the verified source groups. These checks
validate retained facts; executable target joins and behavioral proof remain
separate.

Retained inputs have a 64 MiB per-file and 512 MiB aggregate byte limit.
Document accounting permits 1,000,000 nodes per document and 8,000,000 in total;
it charges containers, scalar values and object member names. JSON scanning
stops after one excess token before decoding the object graph. YAML retains its
bounded alias/node conversion, then contributes its converted payload to the
aggregate. Nesting is limited to 256 levels. Refusal preserves anchored source
rows with diagnostics. Rendered content also keeps its separate byte/token
bounds; a one-node string is not an exemption from those bounds.

These retained-inventory limits are distinct from target collection and output
encoding. The collector bounds each execution file to 64 MiB but has no separate
aggregate byte or schema-file-count ceiling. Generation buffers the complete
JSON result without an independent output-byte ceiling; check mode limits the
saved candidate to 512 MiB. These checks do not promise a single bound on total
process memory or an atomic snapshot of all repository files.

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

Affirmative retained hook-registration clauses preserve unresolved sync demands
when callback objects are absent. That visibility creates no receiver gap or
execution claim: CP13 reconciles historical demand identities and ownership.
The existing Vercel demand keeps its explicit reviewed gap; Sentry remains
mapped-unproven and does not become a thirteenth receiver gap.

The report is not the separate source-role/destination-role/mode matrix. It does
not approve credentials, provider exercises, receiver exposure, a relay or a
scope reduction. The existing checkpoint and independent-review gates still
apply.

## Reviewed response interpretations

Collection cardinality is separate from binary shape. A root array of established
object records supports collection applicability even when a nested record field
is unknown. Scalars, scalar arrays and closed fixed scalar objects exclude ETL;
bare or open objects and mixed or unresolved alternatives remain unknown. An
object containing record arrays needs a reviewed response interpretation. Neither
field spelling, pagination, a read method nor the number of arrays supplies that
meaning.

An exact-key annotation may include `response_interpretations`. Each entry has
`kind` (`collection` or `single_resource`), a literal `response_schema` citation,
a separately cited `citation` and literal `clause`. A collection additionally
requires `records_pointer`: an instance JSON pointer through named object
properties to an array of object records; `""` selects a root array. A single
resource forbids that field, including null. The response-schema pointer is a
source-document coordinate, not the instance record pointer.

The authoring validator checks the exact successful response occurrence, local
reference lineage, hashes, support location and selected shape. It rejects
request/error/foreign occurrences, stale hashes, duplicate or conflicting claims,
scalar selections and collection claims on mutations. Human review establishes
the cited clause's meaning; the code does not interpret English. External refs,
cycles, unsupported conjunctions and traversal limits retain scoped deficits.
Every intermediate named-property wrapper must have a supported object shape.
Explicit array or scalar types cannot become objects merely by declaring
`properties`; object inference from properties applies only when type is absent.
The same local shape evidence governs structural collection classification and
selected single resources and collection record items. Present unsupported types
and malformed properties stay unresolved; a bare `{}` supplies no object identity.
Properties-only object inference requires a valid map and no competing typeless
items or prefix. Empty `prefixItems: []` has the same uniform-items meaning as an
absent prefix. Nonempty or malformed prefixes cannot establish whole-array
collection coverage from their tail, even if every described prefix item is an
object. Fixed-index projection remains a separate supported coordinate.

The common evidence decoder does no reference traversal or source I/O. Existing
bounded resolvers establish each local node; checked binding projection retains
its stricter ancestry, keyword, requiredness, bounds and physical-use checks.
Object cardinality alone never admits a field mapping or executable reference.
Unknown envelopes without an interpretation emit
`source_collection_interpretation_missing`. Other unresolved response scopes
emit `source_collection_scope_unknown`, including unresolved siblings of a
separately established collection.

The report retains schema, support and selected array/item citations. Its
independent evidence check uses retained documents and the original annotation
input, so agreement between two generated cells cannot establish authority.
The three reviewed Batch One additions select Notion `/templates`, Jira `/keys`
and CircleCI `/org_project_data`; they create no targets or behavioral proof.
The optional annotation envelope is described by
`source-lane-manifest.schema.json#/$defs/annotation_envelope`.

## Typed target projections

Target references use closed kinds and an optional source-schema citation. Schema
roles are `request`, `response` or `record`; the empty role means unspecified.
Each field mapping cites a retained source occurrence and names a target kind
(`schema`, `config` or `parameter`) with an explicit JSON pointer. An empty
pointer selects a schema root; a missing or null pointer is invalid. Config
mappings select a property, while parameter mappings select an indexed typed
parameter declaration. These coordinates do not evaluate templates or rename
provider fields.

Absent or null optional citations and absent, null or empty mapping lists carry
no additional facts. A present citation must include its document, pointer and
literal value digest. Optional GraphQL selectors cite the retained operation
name, document and request schema; an absent, null or empty selector object
supplies no facts. Present selector objects remain closed to unknown fields.
They introduce no provider-envelope format or alternate GraphQL parser.

`python3 scripts/tests/source-lane-graphql-parity.py` checks the published schema
with an installed Python `jsonschema` Draft 2020-12 validator and runs the actual
Go annotation-reader matrix. It requires all eleven literal cases to run and
compares both validators with their expected outcomes; it installs no packages.

Target identity compares every scalar and pointed value. Mapping order and
nil-versus-empty lists do not change identity. Duplicate, conflicting and
ancestor/descendant field claims are refused before comparison; normalization
never removes invalid claims. Accepted copies own their citation and pointer
values. Structural shape checks do not establish source ownership, a valid
consumer projection or execution. Those require the binding join, and a
reference deficit still prevents behavioral-proof promotion.

A schema target uses the exact registry key and an empty artifact-root pointer,
with its canonical operation/schema-role coordinate. Other proof target kinds
cannot use the empty artifact pointer. Sync descriptor references carry their
role and executor identity without schema projections.

The binding join checks retained source ownership before accepting a projection,
including uniquely linked local references, and compares the loaded consumer's
actual body, template, record extraction and GraphQL variable declarations.
Separate successful response scopes retain separate references; an accepted
schema observation alone does not complete executable coverage. Source-schema
traversal is limited to 4,096 visits and 128 levels. Unsupported compositions,
ambiguous ownership, external references and missing consumers remain explicit.
A physical component citation needs a complete ownership search: a skipped
schema branch or exhausted bound cannot establish uniqueness. Literal referring
occurrences select their own use sites without borrowing that unproved uniqueness.
A definition registry nested beneath the schema root is still a physical
definition, not a literal property/item occurrence.
A syntactically direct pointer is only a traversal hint. The join carries checked
property/item/fixed-index steps, requiredness and array bounds separately from
physical-search completeness. Selected ancestors must have supported semantics;
malformed requiredness, unsupported composition and unknown reference/scope
keywords remain unverified even when encoded as scalar JSON. Unselected stored
definitions do not become instance uses, and unrelated unknown siblings do not
invalidate an independently selected direct path.

Local reference pointers use the same closed escape/index syntax in uncached and
cached lookup, preserving literal citation hashes and document prefixes. Uniform
items, fixed prefix positions and tail-only items have distinct coverage: a tail
schema cannot certify all records returned from a prefix-plus-tail array.
Record extraction and GraphQL variable placement check supported ancestry;
GraphQL uncertainty is not reported as a proven variable mismatch. Missing or
malformed target selectors remain pointer mismatches, while unsupported target
semantics remain unverified. Budget-consuming property traversal is sorted,
reference cycles are path-local, and the original visit/depth bounds remain.

Canonical-operation checks read already-observed source-lock bytes from a
separate authoring map. Execution artifacts never contain those bytes. A valid
sync descriptor or supplied plan does not prove registered transport or delivery
authority. Direct-only response schemas, absent binary consumers and unsupported
transformations are not promoted into executable capabilities.

## Proof record format and limits

The [proof assertion schema](source-lane-proofs.schema.json) describes additive
closed schema version 1. The existing empty `records` array remains valid. A
record uses exactly one of inline `inputs` or `input_set_sha256`; even an explicit
`inputs: null` cannot accompany a shared-set reference. Optional same-document
`input_sets` declare complete `{path, sha256, role}` tuples. A set digest covers
compact canonical JSON of `{version: 1, inputs: [...]}`, with tuples sorted by
path, digest and role. Sets are immutable evidence declarations, not executable
commands or external registries.

The caller separately supplies trusted reviewed sets and exact record/assertion
mappings. Decoding a record never adds it to that catalog. Source keys come from
the independently validated cohort, allowing at most one proof per exact
key/lane: 30,401 possible slots for Batch One, without making every lane
applicable. Global and orphan errors remain in the report even when they cannot
be attached to an anchored row.

The loader bounds the document to 128 MiB, each record to 64 targets, each set
to 4,096 pins, aggregate input tuples to 131,072, and unique file entries to
65,536. Unique charged bytes, including the document, are bounded to 512 MiB.
Go/module/receipt files retain 4 MiB limits; target artifacts use 64 MiB limits.
These independent ceilings do not promise admission of their full product.
Structural decoding enforces counts before allocating further typed entries.

An invocation-local cache records actual file identity, content digest and read
status independently of a claimant's expected digest. Every successful unique
file is checked again by identity and content before return: two bounded
content-read passes, not one physical read or an atomic repository snapshot.
Failed reads retain a conservative allowance charge including lookahead; that
charge is an upper bound, not a claimed exact byte count. Physical allowance is
at most twice the unique-byte budget plus bounded per-file lookahead. Context
cancellation invalidates proof acceptance. No receipt command is executed.

Schema validity, current hashes and successful test events still cannot establish
assertion meaning or complete lane behavior. Those remain responsibilities of
the independently reviewed assertion mapping and the checkpoint review gate.
