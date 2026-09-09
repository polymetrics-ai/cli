---
phase: cp12-source-lane-manifest
type: tdd
requirements: [CP12-01, CP12-02, CP12-03, CP12-04, CP12-05, CP12-06, CP12-07, CP12-08, CP12-09]
---
# CP12 selected implementation plan

Task delivery header and discussion: CONTEXT.md. Owner: cli-batch1-pi-takeover, Astra/medium. Source base9c642b873b6a1ba3a4fe55fd7509715050c5b71c. Architecture selected by workflow-owned fresh Astra/xhigh cp12_architecture_plan_097, terminal report SHA256f8227a7830c95367756893e71acfbf0e3d739d541eb923438bffdab4216c5bf4. Firstmate097 authorizes owner execution of the selected in-scope plan without an additional routine permission. All three decisions are within that scope; no new runtime foundation/dependency/product decision identified.

This project plan materializes selected contracts, paths and TDD units from that authored assignment and terminal design. Private supervisor custody/history stays outside project Git. Independent review remains separately assigned. CP11 carry-forward remains unchanged. GSD source/prompt provenance and required skills are in CONTEXT.md; runtime adapter executes locally, no new programming-loop command. The custom phase uses these existing issue/checkpoint contracts rather than fabricating a new ROADMAP/state package.

Red: pending. Green: pending. New compiling seams may be contract scaffolding, but are not behavior evidence. Record meaningful desired failure before corresponding product behavior.

## Decision 1 — retained provider inventory precedes executable admission

**Select a typed, lossless retained-provider inventory in `cmd/connectorgen/source_inventory.go`, backed by an explicit cohort anchor, and restore the four missing immutable files at their historical paths.** Do not relax schema4's execution-form requirement, turn 384/970 execution units into provider membership, restore an old execution importer, or read Git history at generator runtime.

`cmd/connectorgen/vnext_lock.go:90` provides `decodeVNextSourceLock`; `:98` provides `decodeStrictJSON`; `:185` starts canonicalization and `:256` rejects an operation without executable form. Those are the right contracts for authored execution, and the wrong place to admit source-only provider inventory. Reuse strict duplicate/trailing JSON handling and raw JSON cloning, but define a separate, explicitly authoring-only provider-document envelope. This is the 097-authorized normalization of immutable provider evidence, not a second executable descriptor dialect. Update canon wording to make this evidence/accounting boundary explicit while preserving the prohibition on predecessor execution readers.

### Immutable restoration and cohort anchor

Use `2008509f40c90de5d041aa944b19fedfecc15533`, the verified available parent before deletion by `0b214b79eeb871238ce8454cd7b896e71e2746a7`. Read exact blobs and restore only these four paths, without reformatting:

| Path below `internal/connectors/defs/` | Expected SHA256 |
| --- | --- |
| `asana/sources/asana-operation-source-lock.json` | `eb5517f0f1456e4cacb03d4f705fd2244a1ccf6113a9c0d67e29ed2417e407c4` |
| `asana/sources/artifacts/cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56.artifact` | `cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56` (3,066,750 bytes) |
| `gitlab/sources/gitlab-operation-source-lock.json` | `de7010cc2a2088b8da33f50312db4d5129fd167b53ef1e028ccf72cae5c72b8d` |
| `gitlab/sources/gitlab-binary-operation-source-lock.json` | `8ccd5184448121eafe1062b76578c733798c93a3e5e3251c63ce3adae167b924` (3,058 bytes) |

Both supplement artifacts already exist and were independently rehashed during planning: `gitlab/sources/artifacts/f59c93194c095d0e925a5751a08eb7a2176a26c6b5f38bda52f805154219d0f0.artifact` (103,189 bytes), and `gitlab/sources/artifacts/53244a720b8509536290e0058c946a246817c775c797df36f4c9aa1225fdf0a4.artifact` (102,560 bytes). Preserve them. No `dc481bac` inspection is claimed or needed. Do not fetch mutable upstream sources.

Create `data/connector-canon/batch1-source-lane-cohort.json` as a closed, versioned input containing `cohort_id`, ordered connector inventory entries, inventory class, relative lock path, lock-byte SHA256, retained acquisition Git ref/path, `expected_ids`, `expected_count`, and verified raw-artifact pins. Expected IDs are copied verbatim from the verified 095 census and independently compared to the exact retained locks before committing. This is the permanent source-set anchor, not the generated product manifest. A source refresh is an explicit change to this anchor and immutable evidence; the generator never rewrites it from observed output.

| Connector | Primary | Current source-lock SHA256 |
| --- | ---: | --- |
| asana | 249 | restored hash above |
| gitlab | 1752 | restored hash above |
| bitbucket | 297 | `0cd529630a96b1c9fd2a75081d83302dcd651c7d665f3c780719ba4eed80b627` |
| circleci | 111 | `06527eb0012ba8f3396074fb048dad1352f8f0d0c29de9c795a0df9f3be5ca60` |
| dockerhub | 54 | `0a9224a085305dd51037e2f3d723d53cef9659625c0146587a97207747011bc3` |
| jira | 617 | `d37a87d79d658bca06707a787dda1dfd0c0c78527140c3d8f1d8d26426175290` |
| notion | 49 | `977f239db74accbe757d88520e013b8c916eae10dffb440f33ebbfa5f75ecf48` |
| sentry | 223 | `383633e6c8403b78c44d9841cd271ec49dbbcecc5778f6c74db1ec162ef1a059` |
| stripe | 589 | `8ce51371f04d6df6e87d3fb6f7924b52d1f61714ec8f0b88f6a3fd12fd205eb6` |
| vercel | 400 | `2eb0130c9357c90cf74e6b47da48eac3768fec3ae3de9b15890b24a2d3f10f4e` |

Planning compared every actual retained ID set and each lock hash to 095: all ten matched. That read-only comparison is an observation, not behavioral RED/GREEN.

### Selected contracts

Use comparable `sourceOperationKey { Connector, Inventory, ID string }`; `Inventory` is `primary` or a stable supplement inventory ID, with class separately `primary|supplement`. JSON serializes the three fields, not a lossy concatenated key. `.id` is preserved byte-for-byte. Empty/control-character IDs are diagnosed, not trimmed or repaired. Unprefixed GitLab IDs and Bitbucket rows with empty `operation_id` remain valid membership.

`retainedSourceInventory` contains pinned document records, expected keys, observed keys, normalized operation records, raw shared source contracts, and structured diagnostics. Begin from the independently anchored expected key set, then attach observed provider nodes. Report duplicates, missing keys, unexpected keys, and counts separately. Unexpected/output-only IDs go in diagnostics, never into the accepted provider row set. A missing/malformed inventory leaves every anchored row present with `facts.status: unavailable`, source-unavailable diagnostics, seven unresolved cells, and failed source-integrity status. These are expected-but-unverified rows, not fabricated provider facts. No count fallback is permitted.

Schema2 reads `rest.operations`; schema3 flattens every `rest.source_documents[].operations`, retaining document identity per operation. Require `counts.total` and its exact agreement with distinct retained IDs and the independently pinned count; preserve other count dimensions. Preserve the original envelope and provider payload semantics, including null versus absent. Source documents are not required to contain executable units.

Each document record distinguishes `retained_file_sha256`/actual byte size from `upstream_declared_sha256`/declared size, with `upstream_bytes_verified` explicitly true or false. For GitLab primary and the eight embedded-operation archives, an upstream raw-document digest remains metadata unless those raw bytes were actually read. `coverage_confidence` states `machine_readable_snapshot`, `rendered_reference`, or `partial`, with retained basis and completeness limits; never emit a percentage whose denominator is what the generator happened to find. A missing count, unexpectedly small or substituted pin, or mismatched expected set fails source integrity. Supplements are exempt from primary-universe size expectations because they are explicitly scoped references, not small primary specs.

## Decision 2 — source-backed applicability and typed exact references

**Select independently derived facts/applicability, explicit intended bindings, and separately verified materialization/proof.** A valid declaration is not behavior proof. An absent planned target is incomplete work; a claimed present target that dangles or binds different semantics is an error. Both retain the source and every lane.

### Lossless facts and source anchors

Implement `source_facts.go`. Retain `source_operation` as provider JSON, shared `source_contract` once per document, and all relevant schema nodes by document-local JSON Pointer. For Asana, decode the verified YAML artifact with the already installed YAML library, reject duplicate keys/trailing documents, and resolve the exact retained method/path/source location in that document. Do not invent raw bytes for embedded-operation archives. For HTML supplements, retain verified document/section anchors and source-cited typed facts; do not scrape network pages at generation time.

Each normalized fact carries a `sourceFactRef { DocumentID, Pointer, ValueSHA256 }` or a verified rendered-reference section/quote selector. Pointers resolve only within the pinned document; never evaluate a source-location string as code. Canonicalize JSON numbers without float64 rounding (`UseNumber`), preserve opaque extensions, and resolve local `$ref` links with memoization and bounded traversal. Keep recursive schemas as a reference graph; do not blindly inline cycles. Unresolved/external references remain source-keyed diagnostics and cannot become negative applicability evidence. No external reference fetch is allowed.

Required fact groups: protocol/method/original path/provider `operation_id`; path and operation parameters with effective override precedence; required query/header/path/cookie scope; request-body presence/requiredness, every declared media type and schema; every response status/header/media/schema; record-envelope/array shape; paging controls/response continuation/absence-versus-unknown; effective root/operation security and referenced schemes; event/callback/webhook/change-token declarations; retained descriptions and event enums. Preserve operation-level empty `security: []` as an override, and do not silently substitute root auth. Preserve Asana's root event/batch inventories. Raw/shared facts remain available even when a normalized interpretation is unresolved.

Do not manufacture `requestBody.required: true` when the source omits it. Vercel file read's JSON object's `required: ["path"]` is explicit and must survive; outer requestBody requiredness is a distinct fact. Missing a fact which exists in retained material is a validator defect/error. A fact genuinely absent from retained material is an explicit unknown with an owner and cannot justify `not_applicable`.

### Reusable rules and bounded semantic annotations

`source_lane_rules.go` owns reusable rules; `data/connector-canon/batch1-source-lane-annotations.json` contains only source-cited noninferable semantic decisions, exact intended bindings, source-linked field mappings, and existing foundation-demand references. It never contains copied executable HTTP bodies, paths as runtime instructions, hand-authored final lane states, or arbitrary provider code. Every annotation names an existing exact key and verified source fact/section; unknown keys, conflicting annotations, changed cited values, or a contradictory source node fail validation. Store an exact cited summary/description clause for reviewed semantic interpretation, not just a source filename. Generic structural predicates are recomputed from the source; opaque semantic interpretation remains clearly identified as reviewed source interpretation.

| Rule | Positive applicability | Exclusion / unresolved behavior |
| --- | --- | --- |
| Direct read | Source-supported bounded read/query/lookup; safe-method read with compatible response, or explicit source-cited POST read semantics | A proven mutation excludes this lane; ambiguous semantics remains unresolved. A word such as `query` in a target name is not an oracle. |
| Direct write | Source-supported creation/update/deletion/action mutation | Source-cited read semantics exclude it regardless of POST. Unknown semantics stays unresolved. |
| Binary download | Successful response with concrete binary media/schema or verified explicit file-download semantics, including media such as PDF | JSON object response with fully known nonbinary contract can exclude. Missing/wildcard media cannot establish binary support or exclusion. |
| Binary upload | Concrete binary request body, a binary multipart part, or verified raw-file/gzip upload semantics | Closed JSON-only nonbinary request can exclude. Include binary fields through local refs; do not treat creation of upload metadata as actual byte upload. |
| ETL | Source-backed extractable record collection, including nonpaginated finite collections; retain paging obligations separately whenever declared | Fixed scalar/file/single-object read or mutation excludes. An unresolved record shape or absent continuation knowledge is an unproven fact, not exclusion. |
| Reverse ETL | Independently evaluate source mutation's destination direction and row-to-request facts; provider mutations retain an explicit cell even if materialization is absent | Source read excludes. Missing executor, idempotency, unattended retry, or batch encoder is not source N/A. Preserve real batch/row-shape constraints as deficits. |
| Sync transport | Explicit source event/change contract or existing cited receiver demand; bind role/mode/ack facts separately | Pagination/listability alone never establishes sync. Webhook registration retains demand information but is not an inbound executor. Unknown event semantics stays unresolved. |

Use common predicates for resolved array/envelope shape, required parameter/body fields, concrete request/response media, event declarations, and source-cited read/mutation classification. Use narrow exact-key semantic annotations when structural inference is insufficient, including known POST reads. Do not port historical method-only mutation rules or the narrow pagination-only ETL predicate unchanged. Historical matrices cited by 095 are evidence for intended mappings, not a new generator input or a state oracle; do not restore their old checkers. Any extraction into the new annotation schema must be revalidated against the pinned provider facts and current target.

At minimum permanently test the Vercel POST file-read case, Notion `notion.rest.post-database-query`, a JSON mutation, a finite nonpaginated collection, a pageable collection, a binary schema under `$ref`, and an unresolved fact. Record-source-backed mutation behavior must produce independent direct_write and reverse_etl cells; an ETL cell cannot disappear because a direct-read mapping exists.

### Cell state reducer

Each of the seven cells has `lane`, `applicability` (`applicable|not_applicable|undetermined`), `state` (the four required values), `rule_id`, source fact references, a reason code/text, intended bindings, verified references, proof references, owner/gap references, and diagnostics. Exactly seven cells appear in the prescribed lane order.

- `not_applicable`: only source-validated exclusion; no positive binding or implementation claim.
- `missing_foundation`: applicability established and exact existing Atlas lookup/gap/decision reference records the unmet shared contract. Missing local code is insufficient. This state never grants approval to build the foundation.
- `implemented`: applicability established, every lane-required present reference valid, required command/stream/action/transport reachability valid through existing authorities, and current lane-specific behavioral proof attached with honest scope.
- `mapped_unproven`: all other mapped or unresolved cases, with explicit missing facts/materialization/schema/behavior/admission diagnostics. `applicability: undetermined` and `facts_unresolved` must remain visible rather than pretending the lane was established applicable.

Manifest structural validity and connector completeness differ. Honest deficits can appear in a valid manifest. A false source fact, omitted anchored source, wrong existing semantic/schema target claimed as materialized, forged/cross-bound proof, or internally contradictory state is a validation error. C1/C2/C3/C4 remain evidence classes; C3 refusal, C4 unsupported, shared-engine tests, credential failure, and preflight alone do not prove an implemented lane. Absence of acceptable real records can legitimately produce zero `implemented` cells; do not preassign a nonzero target or blanket-mark all cells unproven without source rules.

Keep diagnostic severity explicit: absent canonical materialization, unavailable proof, and an observed execution render/admission failure with no accepted materialization claim are visible connector deficits; a manifest can correctly report them and pass its own check. Missing/corrupt required retained source, false copied facts, or an invalid asserted present reference makes source/manifest validation fail. Do not upgrade or downgrade severity merely to obtain a green corpus. A diagnostic always includes exact source key, affected lanes, stage, stable code, source/target pointer, and existing owner; stable generated text uses relative paths and excludes host-dependent raw error formatting.

### Exact target interfaces and rules

Implement `source_lane_bindings.go` against the current contracts:

- `vNextCanonicalDescriptor.Graph.Operations` exposes exact canonical `ID`, raw `Source`, `SchemaRefs`, `Stream`, `Write`, `Operation`, and `Commands`.
- `vNextCanonicalOperation.Index` is for authored diagnostics; `CanonicalIndex` identifies stable provenance fields.
- `vNextStagedGeneration` exposes `Outputs`, `Identity`, `Manifest`, `Index`, `Provenance`, and supplied `Sync`.
- `vNextSourceExecutionProvenance` contains `SourceID`, `FieldPath`, `TargetKind`, `TargetID` (`vnext_admission.go:33`). Its SourceID is a canonical-unit ID, not automatically a retained provider ID.
- `canonicalizeVNextSourceLock` already performs canonical construction/admission. Reuse its staged result once; do not repeatedly rerender/admit a connector per source row.
- `vNextValidateSourceSchemas`, `vNextValidateSourceFacts`, effective schema handling and the existing resolver/preflight remain authoritative for the contracts they actually implement. Direct-only response-schema roles still lack an effective runtime consumer and cannot be invented to make a link pass.

Selected `sourceLaneTargetRef` has `kind` (`canonical_operation|stream|write|operation|command|schema|sync_transport`), connector, exact ID, lane, artifact relative path, JSON field pointer, artifact SHA256, canonical unit ID and canonical field pointer when available, staged identity when applicable, and schema role/provenance links. A command's exact path may identify that target but never source membership. Registry schema paths identify schema targets, not provider operations. Source/canonical identity must be joined explicitly by `sourceOperationKey` and exact canonical unit ID; several canonical units or aliases can legitimately bind one source ID.

Distinguish `intended` references (known planned kind/ID/path, which may be absent) from `materialized` references (all asserted current fields mandatory and resolvable). Missing ID for a future projection is an explicit `target_unspecified` deficit, never a made-up command spelling. `intended` with absent target remains unproven. If an intended target exists, inspect it; never evade a mismatch by labeling a present target absent. A rejected candidate may be reported as an observation/diagnostic, but cannot remain an accepted materialized reference. An explicit materialization claim that is dangling or wrong-semantic fails check, even when its cell says mapped_unproven.

Exact REST semantic joins compare source method/path and operation identity where supplied, explicit path-variable/parameter bindings, body/media constraints, response/record facts and schema role. Apply only the declared GitLab prefix bridge `{source_prefix:"/api/v4", connector_prefix:""}` at a segment boundary, retaining both paths. For existing Go-template paths, accept only a parsed literal route with explicitly mapped `record`/`config` path fields; do not erase arbitrary templates, rename variables speculatively, or strip arbitrary prefixes. GraphQL must match operation name/root field and request schema; equal `/graphql` route is insufficient. A schema reference must bind the appropriate effective schema/field projection, not merely a readable JSON file. Record versus envelope and path-parameter versus body schemas are distinct; explicit source-to-target field projections are required rather than comparing unlike whole schemas.

Current Asana has a usable observed pair: provider `asana.rest.getAgent` GET `/agents/{agent_gid}` → canonical `operation:get_agent` → artifact operation `get_agent` and command `agents get-agent`; independently retain source parameter facts. Provider `asana.rest.approveAccessRequest` → canonical `write:approve_access_request`, action `approve_access_request`, command `access-requests approve-access-request`, with explicit `access_request_gid` → `record.access_request_gid`. These are exact candidate bindings to validate, not pre-certified implemented cells. Existing static examples and approval flow are in 096.

For the eight connectors lacking schema4 locks, observe existing allowlisted execution artifacts through the existing `engine.Load`/binding types where valid, and retain exact artifact references/diagnostics. Mark canonical provenance absent. Do not construct a lock from those artifacts or treat a valid old artifact as canonical admission. An artifact can be structurally observed yet insufficient for the materialized claim; record that distinction. No provider ID may be added by scanning artifacts. Orphan explicit links fail; unrelated existing artifacts can be reported as unmapped observations without becoming source rows.

### Behavioral evidence interface

`source_lane_proof.go` validates optional records in `data/connector-canon/batch1-source-lane-proofs.json`. Each record binds exact source key/lane, target identity/digests, test file and symbol, source/test/dependency hashes or an exact tree plus reviewed unchanged-binding receipt, execution class, original result receipt path/hash, selected test/subtest, non-skipped successful result, tested observable contract, and limitations. Scope is per lane. Test declaration, receipt syntax, preflight, and hash equality alone cannot establish behavioral sufficiency: the owner supplies a reviewed test/assertion-to-lane mapping, and the independent reviewer challenges it. Never execute commands found in this data. Invalid claimed proof fails; absent/outdated evidence remains unproven. The production proof file starts empty if no existing exact source/lane-bound behavioral receipt meets this contract; fixture proof records are confined to tests and cannot promote real connector rows.

### Required example cells

Vercel facts are retained at `sources/vercel-operation-source-lock.json`, `.id=vercel.rest.readSessionFile`, source location `paths["/v2/sandboxes/sessions/{sessionId}/fs/read"].post`, and `.id=vercel.rest.createWebhook`, location `paths["/v1/webhooks"].post`, under the pinned Vercel lock hash above. Generated fact references must use actual immutable-document pointers/digests, not these prose abbreviations.

| Lane | readSessionFile | createWebhook |
| --- | --- | --- |
| direct_read | mapped_unproven; source-supported POST file read; target absent | not_applicable; registration mutation |
| direct_write | not_applicable; source read semantics | mapped_unproven; registration mutation |
| binary_download | mapped_unproven; 200 octet-stream string/binary | not_applicable; known JSON registration result |
| binary_upload | not_applicable; closed JSON `path`/`cwd` request | not_applicable; closed nonbinary JSON registration |
| etl | not_applicable; fixed file response, no record collection | not_applicable; mutation |
| reverse_etl | not_applicable; read semantics | mapped_unproven; destination mutation |
| sync_transport | not_applicable; fixed file read with no event-intake contract | missing_foundation; existing receiver demand, not an implemented receiver |

The latter uses `cli-webhook-event-surface-foundation-r1`, consulted `transport.sync-contract.v1`, and existing decision owners `cli-batch1-vercel-inbound-sync-decision-r1` and `cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary`. Preserve CP13 reconciliation of twelve historical candidates (Bitbucket4, CircleCI2, Jira1, Vercel1, GitLab3, Stripe1) as retained demands, not fresh approvals. Sentry's service-hook case remains mapped_unproven; do not create a thirteenth gap or rediscover a receiver merely from an events field. Other newly uncertain demands go to the existing CP13 owner as unresolved diagnostics, without inventing a confirmed gap.

Concrete selected cell JSON (a cell within source key `{connector:"vercel", inventory:"primary", id:"vercel.rest.readSessionFile"}`, not a generated/certified result):

```json
{
  "lane": "binary_download",
  "applicability": "applicable",
  "state": "mapped_unproven",
  "rule_id": "successful_binary_response",
  "fact_refs": [
    {
      "document_id": "vercel:primary",
      "pointer": "/rest/operations/299/source_operation/responses/200/content/application~1octet-stream/schema",
      "value_sha256": "2db098654a77dbd75ff3dd3bda2aa4fec4c1bb261ed413634b61832152672d3e"
    }
  ],
  "reason": {
    "code": "materialization_absent",
    "text": "The retained successful response is binary; no current file-read operation or command materializes this source mapping."
  },
  "intended_bindings": [],
  "references": [],
  "proof_refs": [],
  "owner_refs": ["CP26"],
  "gap_refs": [],
  "diagnostics": []
}
```

The pointer's index299 was checked against the pinned source's exact ID order (createWebhook is index381). Its value hash is canonical JSON `{"format":"binary","type":"string"}` with sorted keys and no whitespace/newline. `vercel:primary` identifies the retained lock bytes with the pinned Vercel hash above; it does not claim the original upstream raw document was verified. Empty intended bindings deliberately avoid inventing a PM command. Use these snake_case JSON names consistently in closed Go tags and the JSON schema. `reason` is the closed `{code,text}` object; references and proof refs are arrays even when empty. Other cells follow the same shape; no optional omission of a lane is permitted.

## Decision 3 — deterministic authoring JSON and a read-only connectorgen seam

**Select one `source-lanes` command in existing `runContext`, a checked-in manifest at `data/connector-canon/batch1-source-lane-manifest.json`, and buffered JSON generation to stdout plus read-only `--check`.** The generator owns no filesystem publication transaction. The sole project owner saves a fully generated, validated candidate and normally commits it. Do not add a second CP11 generation publisher, CURRENT reader, journal, lease, or cleanup protocol for this advisory document.

Command contract:

- `go run ./cmd/connectorgen source-lanes` generates the entire anchored Batch One document to stdout; source totals precede availability summaries in the document. Exit0 means the report is internally/source-reference consistent, not that every lane is implemented.
- `go run ./cmd/connectorgen source-lanes --check` reads the fixed tracked manifest, independently regenerates from current pinned inputs, validates its source sets/facts/cells/claims, compares deterministic bytes, and returns1 on drift/error. No write, mkdir, temporary file, lock, publisher, or recovery operation is permitted.
- `--repo <dir>` selects a repository root for isolated tests; default uses existing `repoRoot`. `--manifest <relative-path>` is accepted only with `--check`, for checking a candidate before the owner installs it at the fixed path. Require confined regular files, reject absolute/traversal/control-character/symlink escapes in input-derived references, and reject conflicting/duplicate/unknown args with exit2.
- `source-lanes --help` gives accurate authoring-only help, exit0; missing optional flags are fine. No PM command, credential flag, availability filter, source-refresh flag, arbitrary output path, or runtime route is added.
- Generate/validate the full in-memory result before stdout. A failing run may emit a complete JSON diagnostic document with `validation.status:"invalid"`, the complete anchored rows/cells and error findings; it must never emit a success-shaped reduced census. Fatal cohort-anchor syntax failure cannot honestly enumerate keys: return an explicit anchor-invalid error and no purported manifest. A short/error stdout write returns1; never use `logf`'s ignored writer result for successful manifest output.
- Owner captures successful output into a fresh candidate, checks it with `--check --manifest`, then replaces only the owned tracked manifest. Never redirect generation straight over the existing tracked manifest. On any nonzero result preserve the old tracked bytes; keep invalid diagnostic output only as a separately named receipt. This is controlled artifact saving, not runtime-generation publication. Original source/code/evidence receipts remain outside the generated manifest's self-hash cycle.

Use `sourceLaneManifest { SchemaVersion, Kind, CohortID, SourceTotals, Inputs, Documents, SourceOperations, LaneSummary, Diagnostics, Validation }`. Set schema_version to integer1 and kind to `retained_source_lane_manifest`; this is a document schema identity, not a reduced feature version. Closed Go envelope/cell/reference types and `docs/connector-canon/source-lane-manifest.schema.json` describe the same contract. Source provider JSON remains an explicitly opaque JSON-object boundary. The manifest stores the normalized retained representation; do not add a second generated normalized file that can drift from it.

Every source row contains exact key, inventory/document refs, original location and canonical pointer, preserved source fact groups, independently derived semantics, exactly seven cells, and diagnostics. Root totals separate primary/supplement/expected/observed counts and rows/cells; all summaries are recomputed. Counts never stand in for set equality. Sort connectors and source rows by exact key, documents/refs/diagnostics by declared stable keys, and cells by the seven-lane order. Object keys and semantically unordered collections canonicalize; retain provider arrays whose order matters. End with one newline. No current timestamps, host paths, random IDs, credentials, or process details occur in generated bytes; acquisition timestamps already in immutable provider evidence are retained as source metadata.

Determinism compares the same immutable input bytes. Permute enumeration/map/normalized-record order while preserving the input anchors and assert equal output. If a test physically rewrites source JSON/YAML, its byte hash legitimately changes: the pinned-source check must reject it, or an explicitly repinned synthetic fixture can compare only the semantic projection. Never claim full-byte equality across changed immutable source digests.

Public pipeline: parse args → load/validate cohort anchor → allocate anchored source rows and seven cells → read/hash/normalize retained facts → source-semantic classification → collect canonical/artifact observations once per connector → resolve exact bindings → assess proof → validate/report summaries → buffer encode → write stdout; check additionally reads/validates/compares the candidate. Catch import/normalization, render, admission, and proof failures into source-keyed diagnostics after membership allocation. Preserve every affected row/cell even if connector-wide admission fails. Expected missing schema4 materialization is a deficit; a fabricated materialization claim is an error.

## Executable GSD TDD units

This is one bounded architecture plan with five sequential execution units, each two tasks. The sole existing owner executes them; waves express dependencies, not authorization to spawn writers. Materialize as `.planning/phases/cp12-source-lane-manifest/12-01-PLAN.md` through `12-05-PLAN.md` (or retain the established custom phase's single PLAN with these exact unit IDs). Include the objective/context/verification/threat/summary fields below. No new planning/review loop is needed.

All units reference this immutable report, 097, and the mandatory REVIEW-EVIDENCE-REQUIREMENTS. Production tasks are TDD: write and execute failing desired-behavior tests first, record original RED, implement the dependency group, GREEN, then refactor. Compiler/unknown-subcommand/zero-test/setup failures are not intended RED. New API declarations and a compiling test harness are contract scaffolding, not feature evidence: record them separately and do not call their setup failure RED. Before adding each normalization/classification/reference/validation behavior, use a compilable seam and a valid fixture reaching the intended stage, record the failed desired assertion with its actual expected/observed IDs/facts/state, then implement that behavior. The first meaningful selected RED must precede the product behavior it proves. Never claim all eventual negative cases were red before the first edit; preserve which witnesses were originally RED and which were later edge/oracle controls, without retrofitting history.

### 12-01 — independent inventory and exact recovery

Frontmatter: `phase: cp12-source-lane-manifest`, `plan: "01"`, `type: tdd`, `wave: 1`, `depends_on: []`, `autonomous: true`, `requirements: [CP12-01, CP12-02, CP12-07]`. Context target about40%; no checkpoints. Files modified are the exact union in its tasks. Must-haves: independently anchored exact membership, restored byte-identical sources, missing source refusal without execution-count fallback. Artifacts: `source_inventory.go`, cohort JSON, expected-ID fixture, four restored files. Key link: cohort source key → retained lock `.id`, never schema4 operations.

<tasks>
<task type="auto" tdd="true">
<name>Task 1: Define and test the source-set contract before availability</name>
<files>cmd/connectorgen/source_inventory.go, cmd/connectorgen/source_inventory_test.go, cmd/connectorgen/testdata/source_lanes/batch1-expected-ids.json, data/connector-canon/batch1-source-lane-cohort.json</files>
<behavior>Exact source sets survive; duplicate/removal/same-count replacement/stale counts identify their keys; primary and supplement identities remain separate; no generated output can create membership.</behavior>
<action>Per 097-S1/S2, write the independent expected-ID test fixture directly from the verified retained research set, retaining its source hash. Define the closed cohort, sourceOperationKey, document/inventory contracts and source-set validation. Expected fixture must not call production normalization or read generated manifest data to construct expected keys. Add a tiny inline synthetic cohort with hand-listed IDs and a same-count replacement counterexample to falsify the set oracle.</action>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceInventory(ExactSet|MembershipCounterexample|Counts|PrimarySupplement)$' -count=1 -v</automated></verify>
<done>Assertions compare exact retained IDs/classes and distinguish duplicates, missing IDs, unexpected IDs, and stale totals; the independent membership helper rejects a readable wrong set while accepting its positive control.</done>
</task>
<task type="auto">
<name>Task 2: Restore the exact absent retained documents</name>
<files>internal/connectors/defs/asana/sources/asana-operation-source-lock.json, internal/connectors/defs/asana/sources/artifacts/cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56.artifact, internal/connectors/defs/gitlab/sources/gitlab-operation-source-lock.json, internal/connectors/defs/gitlab/sources/gitlab-binary-operation-source-lock.json, cmd/connectorgen/source_inventory_test.go</files>
<action>Per 097-S2, observe the expected missing-input test before restoration, restore only the four listed Git blobs verbatim, and verify the specified hashes, raw byte sizes and exact 249/1752+2 ID sets. Leave existing supplement artifacts and schema4 locks unchanged. Add missing/corrupt-input fixtures which leave existing executable definitions in place and assert unavailable-source diagnostics, not384/970 fallback. This is immutable data recovery, with behavioral coverage in the inventory tests.</action>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceInventory(HistoricalRestoration|MissingSourceNoExecutionFallback|RawArtifactIntegrity)$' -count=1 -v</automated></verify>
<done>All restored bytes match pinned evidence; missing/corrupt copies preserve anchored keys and fail visibly; all ten source denominators match independently.</done>
</task>
</tasks>

### 12-02 — complete facts and independent lane applicability

Frontmatter: `plan: "02"`, `type: tdd`, `wave: 2`, `depends_on: [12-01]`, `autonomous: true`, `requirements: [CP12-03, CP12-04, CP12-05, CP12-07]`, same phase. Context target45–50%. Must-haves: retained facts can be audited at exact pointers; source semantics produce every lane; genuine unknowns retain diagnostics. Artifacts: facts/rules Go files and tests, annotations. Key link: cell rule → independently reread retained fact pointer/value, not cell's own copied summary.

<tasks>
<task type="auto" tdd="true">
<name>Task 1: Normalize provider facts without losing source payloads</name>
<files>cmd/connectorgen/source_facts.go, cmd/connectorgen/source_facts_test.go, cmd/connectorgen/source_inventory.go</files>
<behavior>Schema2 embedded source objects and schema3 raw-document locations normalize equally; required parameters/body/media/security/record/event facts remain source-bound; unknown/ref failures retain rows.</behavior>
<action>Per 097-S2/S5/S6, implement retained provider envelope decoding and bounded local JSON/YAML fact resolution described above, preserving raw shared/provider payloads and exact absence semantics. Cover Asana raw YAML, GitLab bridge, local refs, explicit empty auth override, Vercel inner required path, webhook required events/minItems/enum, and pagination record shape. Use fixed expected fact values and independent raw source-pointer reads in tests, not the extractor to produce expected facts.</action>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceFacts' -count=1 -v</automated></verify>
<done>Each required fact survives with a resolvable immutable citation; deletion or contradiction of copied required facts is detected; absent source knowledge is diagnostic and preserves the row.</done>
</task>
<task type="auto" tdd="true">
<name>Task 2: Derive seven cells from reusable semantic rules</name>
<files>cmd/connectorgen/source_lane_rules.go, cmd/connectorgen/source_lane_rules_test.go, data/connector-canon/batch1-source-lane-annotations.json</files>
<action>Per 097-S3/S4/S6/S7, implement typed cells, rule evaluation and source-bound annotation validation. Apply the rule table to the full retained cohort, using exact noninferable semantic annotations where needed. Preserve actual documented POST reads, binary request/response distinctions, finite collections, mutation direct/reverse independence, source event semantics and existing receiver owners. Add positive/negative controls which contradict a claimed rule using readable source nodes. Do not copy historical all-green states, narrow ETL heuristics, or invent gap IDs.</action>
<behavior>POST read stays read; JSON mutation occupies independent direct/reverse cells; finite collection gets ETL; paging alone does not get sync; missing media stays unresolved; Sentry is not a thirteenth receiver.</behavior>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceLane(Applicability|Vercel|PostRead|Collection|Receiver|RuleCounterexample)' -count=1 -v</automated></verify>
<done>All rows have seven independently justified cells; every exclusion names source semantics; unknowns and source-backed demands have precise existing owners.</done>
</task>
</tasks>

### 12-03 — exact bindings and scoped behavioral proof

Frontmatter: `plan: "03"`, `type: tdd`, `wave: 3`, `depends_on: [12-02]`, `autonomous: true`, `requirements: [CP12-04, CP12-05, CP12-06, CP12-07, CP12-08]`. Context target45%. Must-haves: wrong existing target/schema refuses; absent intended target is unproven; only lane-specific current behavior can promote. Artifacts: bindings/proof Go files and tests, proof data. Key links: source key → exact canonical unit → staged provenance/loaded binding → evidence record.

<tasks>
<task type="auto" tdd="true">
<name>Task 1: Resolve typed canonical and artifact references exactly</name>
<files>cmd/connectorgen/source_lane_bindings.go, cmd/connectorgen/source_lane_bindings_test.go</files>
<action>Per 097-S4/S5/S8, collect each current canonical/staged result once using existing strict/admission APIs and observe existing allowlisted artifact targets separately. Implement exact typed intended/materialized references and explicit provider-to-canonical/field/schema joins. Include existing-but-wrong semantic/schema targets, same-route GraphQL swaps, declared GitLab prefix bridge versus unrelated-prefix refusal, absent target controls, and orphan output-only references. Intercept failures after anchored rows exist and record their actual stage. Reuse existing admission tests as patterns, not as source-set oracles.</action>
<behavior>Absent planned targets remain unproven; dangling or wrong-semantic materialized claims fail; valid exact joins pass; output orphans never add a provider ID.</behavior>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceLane(Binding|ReferenceCounterexample|GitLabBridge|Orphan)' -count=1 -v</automated></verify>
<done>Every accepted present reference resolves exact kind/ID/field/schema/provenance; missing or rejected candidates retain all source rows with precise diagnostics.</done>
</task>
<task type="auto" tdd="true">
<name>Task 2: Reduce cell states using explicit proof scope</name>
<files>cmd/connectorgen/source_lane_proof.go, cmd/connectorgen/source_lane_proof_test.go, data/connector-canon/batch1-source-lane-proofs.json</files>
<action>Per 097-S3/S4, implement proof record parsing/binding and the state reducer. Require actual lane-specific assertions and current immutable receipt scope for promotion, without invoking receipt commands or reading credentials. Prove syntax-only/preflight-only/shared-engine/C3/C4/cross-lane/stale/skip receipts do not promote; a fixture-scoped matching positive control does. Preserve the empty real proof set if existing receipts cannot substantiate exact rows. A real invalid claimed receipt is diagnosed; absence does not erase source membership or block runtime.</action>
<behavior>Valid fixture-scoped lane evidence can promote its fixture; equivalent syntax with wrong source/lane/digest or no actual behavior cannot; unproven is not falsely reported implemented.</behavior>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceLane(Proof|StateReduction)' -count=1 -v</automated></verify>
<done>Every real implemented cell has a precise accepted behavioral evidence mapping; no availability/certification condition affects denominator or runtime.</done>
</task>
</tasks>

### 12-04 — complete manifest validation and CLI/check

Frontmatter: `plan: "04"`, `type: tdd`, `wave: 4`, `depends_on: [12-03]`, `autonomous: true`, `requirements: [CP12-03, CP12-05, CP12-06, CP12-07, CP12-08]`. Context target45–50%. Must-haves: complete JSON report, precise invalid-claim diagnostics, source totals before availability, deterministic no-write check. Artifacts: manifest/CLI/oracle files. Key link: real `runContext` route → complete build/validator → checked tracked candidate.

<tasks>
<task type="auto" tdd="true">
<name>Task 1: Build and independently validate the complete document</name>
<files>cmd/connectorgen/source_lane_manifest.go, cmd/connectorgen/source_lane_manifest_test.go, cmd/connectorgen/source_lane_oracle_test.go</files>
<action>Per 097-S1/S3/S4/S5, implement the document schema/build/validate pipeline and deterministic serializer. Validator compares supplied manifest rows/facts/claims to immutable inputs, not just regenerated summary counts; aggregate all named findings without dropping unaffected rows. Add injectable package-local narrow stage callbacks only where necessary for tests, with production paths using the real operations. Each import/render/admission/proof fault must record a stage event and preserve independently expected source keys and seven cells.</action>
<behavior>Every required mutation witness fails at the named contract; faults retain 4341+2 anchored rows; repeated generation and enumeration permutations are byte-identical; a reduced-but-self-consistent report fails.</behavior>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceLaneManifest' -count=1 -v</automated></verify>
<done>Exact sets, fact contents, lane coverage, refs, counts and deterministic bytes are checked independently; helper falsification controls pass.</done>
</task>
<task type="auto" tdd="true">
<name>Task 2: Wire the real authoring-only source-lanes command</name>
<files>cmd/connectorgen/source_lane_cli.go, cmd/connectorgen/source_lane_cli_test.go, cmd/connectorgen/main.go</files>
<action>Per 097-S8, add the command/parser/help described in Decision3 and call the complete pipeline through runContext. Generate to buffered stdout; check reads and compares only. Source/reference errors return1 with explicit invalid report/diagnostics; invalid flags return2. Honor context cancellation, writer errors/short writes, bounded reads, and repository-confined source references. Test the public route using isolated roots, pre-retained file byte/type/identity snapshots, missing input and injected stage failures. Never call the CP11 publisher or modify PM parsing/dispatch.</action>
<behavior>Help is accurate; normal/check JSON exposes exact source totals and limitations; check on missing or stale output writes nothing; failed generation never overwrites the tracked manifest; short stdout write fails.</behavior>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceLanesCLI' -count=1 -v</automated></verify>
<done>Real CLI entry exercises generation and validation; read-only evidence compares original expected state, not repeated reads of current state.</done>
</task>
</tasks>

### 12-05 — tracked product artifact and review-ready evidence

Frontmatter: `plan: "05"`, `type: execute`, `wave: 5`, `depends_on: [12-04]`, `autonomous: true`, `requirements: [CP12-01, CP12-02, CP12-03, CP12-08, CP12-09]`. Context target30–40%. Must-haves: full tracked generated product, authoring help/schema/Atlas ownership, complete evidence packet and no runtime drift. Artifacts: generated manifest, schema/docs/Atlas and owner's established phase evidence. Key link: generated product → immutable inputs/typed references and exact receipts → independently assigned review.

<tasks>
<task type="auto">
<name>Task 1: Generate the full corpus twice and retain its exact evidence</name>
<files>data/connector-canon/batch1-source-lane-manifest.json, .planning/phases/cp12-source-lane-manifest/TDD-LEDGER.md, .planning/phases/cp12-source-lane-manifest/VERIFICATION.md</files>
<action>Produce two independently invoked successful buffered generations, compare their full bytes, check the candidate before replacing only the owned tracked manifest, then run default --check. Assert all ten exact ID sets, total4341 primary/2 supplement/30401 cells, seven per row, every present reference valid, and explicitly enumerate state and unresolved-owner totals. Preserve original RED/GREEN/edge command output/exit/source/test/dependency bindings as immutable receipts; do not claim past REDs or provider proof. Owner's existing evidence directory may be used, with exact paths recorded in the obligation table.</action>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestSourceLaneManifest(BatchOneCorpus|Deterministic|NoHiddenRows)$' -count=1 -v; go run ./cmd/connectorgen source-lanes --check</automated></verify>
<done>Tracked generated manifest is the actual validated corpus result, not the private census; equal bytes and all source/lane counts are recorded with honest unproven rows.</done>
</task>
<task type="auto">
<name>Task 2: Document the authoring seam and freeze a complete review handoff</name>
<files>docs/connector-canon/SOURCE-LANES.md, docs/connector-canon/source-lane-manifest.schema.json, docs/connector-canon/SOURCE-LOCK-VNEXT.md, docs/connector-canon/foundations/catalog.json, .planning/phases/cp12-source-lane-manifest/SUMMARY.md</files>
<action>Document the selected schema/command, source-versus-execution distinction, evidence states, saving/checking candidate safely, and exact downstream boundaries. Extend the existing authoring Atlas owner/selector/proof list, preserving CP11 publication proofs. Record PM runtime help/manual/website/completion changes as not applicable because no PM surface changes; connectorgen help plus canonical authoring docs are required. Run focused verification below, record outstanding CP11 review, complete every obligation disposition and #4325 completion-proof field, and return exact code/evidence SHA for Firstmate's separately prompted xhigh review. Do not launch no-mistakes, reviewers, close issues, or claim independent acceptance.</action>
<verify><automated>go test -timeout 20m ./cmd/connectorgen -run '^TestFoundationAtlasSelectorsResolve$' -count=1 -v; go test -timeout 20m ./internal/connectors/defs -run '^TestRuntimeEmbedContainsExecutionJSONOnly$' -count=1 -v; git diff --check</automated></verify>
<done>Authoring contract and Atlas agree; no runtime source reader or request-path change exists; exact-SHA evidence handoff and remaining gaps are ready for assigned review.</done>
</task>
</tasks>

## Required RED/edge witnesses and independent controls

No tests below were executed by the planner. Record real RED before corresponding production behavior, actual GREEN afterward, and distinguish test-helper counterexample PASS from desired-behavior RED. Tiny fixtures should run in under60seconds after compilation; the full corpus and candidate gates may need longer, always with `-timeout 20m` and preserved selection/output.

| Obligation | Fixture / intentional counterexample | Independent expected state and positive control |
| --- | --- | --- |
| E01 exact membership | Duplicate raw ID; remove one; replace one by a different ID with identical cardinality; stale `counts.total`; modify cohort roster | Hand-listed synthetic IDs plus retained 095 expected fixture; assert missing/unexpected ID, duplicate pointer and exact connector. Positive unmodified set succeeds. |
| E02 historical source | Delete/corrupt Asana/GitLab retained lock or raw Asana artifact while leaving384/970 executable units | Expected249/1752 keys already retained; source unavailable/hash diagnostic, no execution-count fallback. Positive exact restored hashes and actual artifact bytes pass. |
| E03 all lanes | Remove ETL cell from collection; remove reverse_etl from mutation; duplicate/extra lane | Explicit literal seven-lane list and collection/mutation facts; validate omitted lane by exact source key. Positive seven-cell report passes. |
| E04 facts | Remove Vercel sessionId, JSON inner required path, binary media/schema, webhook events/minItems/enum; remove collection/paging/auth facts | Read known pinned source pointers directly and assert literal expected values, independently of production extraction. Positive unchanged source facts pass. |
| E05 applicability oracle | Misclassify Vercel POST read as write, JSON response as binary, finite collection as non-ETL, Stripe list pagination as sync, Sentry as receiver gap | Direct hand-asserted facts and reviewed source clauses; counterexample must be readable/valid JSON. Positive source read/mutation/collection/binary/event cases pass. |
| E06 exact references | Existing other operation, equal route/wrong GraphQL operation, wrong existing effective schema, wrong field projection, arbitrary GitLab prefix stripping | Retain exact target identity/schema before call; known good target passes. Include same bytes under wrong target identity to falsify existence/byte-only oracle. |
| E07 absent versus invalid | Absent intended target; absent materialized claim; present wrong-semantic intended target; orphan output source ID | Planned absence stays unproven; false materialized claim fails; rejected candidate remains diagnostic. Positive valid existing binding passes. No output adds a provider row. |
| E08 failure preservation | Real normalization/import error; render error; semantic admission error; proof/certification evaluation error injected after membership exists | Independently expected full keys/cells before fault, stage event history showing actual boundary, exact source/lane diagnostic; unfailed sibling rows unchanged. Real operation executes before post-stage fault; do not relabel an earlier hook as render/admission. |
| E09 proof semantics | Declaration-only, preflight-only, shared-engine-only, cross-lane, stale/different-digest, skipped/zero-selected receipt | Nonimplementation remains visible; fixture-scoped matching behavioral proof can promote its fixture only. No fabricated real provider receipt. |
| E10 supplement split | Move a GitLab supplement into primary; use supplemental equivalence to delete primary row | Exact1752+2 keys/classes and4341 primary remain; relationships do not deduplicate membership. Positive disjoint classes pass. |
| E11 deterministic output | Permute source enumeration, map iteration and normalized records; run twice | Same immutable source bytes and retained anchors yield equal complete bytes. Physically repinned bytes are a separate hash-change/refusal control. |
| E12 read-only/no partial success | Check missing/stale manifest; import failure after successful sibling; early cancellation; writer returns short count/error | Pre-retain exact tracked bytes/types/identity and tree listing; after call same owned file/roots and no new state. Complete invalid diagnostic result cannot masquerade as success; valid check passes. |
| E13 runtime isolation | Generate/check with source files only in authoring root; inspect embedded FS and changed runtime files | Existing allowlist excludes sources/new manifest, no PM reader/route changed. Positive unchanged runnable execution inventory remains independent of source proof. |

The membership helper must reject E01 same-count substitution; the source-fact/applicability helper must reject E04/E05 readable contradictions; the reference helper must reject E06 wrong existing target/schema/identity. These three controls are mandatory even if the main feature tests already fail on malformed JSON. Do not add a generic mutation framework or dozens of implementation-mirroring tests.

## Firstmate118 correction discussion and plan adoption

Completed100 ledger is committed at6e5d4506ea9a9a9e9daa84aa24a648b04683fcf3 without product changes. Full118/118A/118B envelope hashes verified;119 resolves next action only. Owner Astra/medium; same099d Astra/medium resumed on clean isolated worktree then fast-forwarded to6e5d4506; fresh118B Astra/xhigh read-only immutable9dc. Only118B/Firstmate resolves CR05 representation and WR02 historical-policy disposition; other corrections proceed.

GSD sources and full generated discuss-phase, plan-phase --tdd, execute-phase prompts resolved/read at pinned Core20297a8f. Execute inline via documented non-Pi fallback: this Codex runtime runs shell-generated GSD prompts, not interactive Pi; authored118 limits roles to owner, same binding helper and bounded authority/evidence specialist. Existing adapter doctor provenance/known missing issue-122 prompt is retained, not repaired. No generic planner/verifier/reviewer spawned. Skills applied: Go how-to, testing, errors, safety, security, performance, benchmark, context, design-patterns and structs/interfaces; prior connector lane-build-order and exhaustive overlay remain applicable. No dependency/tool installation or model fallback.

| Group | Owner and allowed paths | Genuine RED and unchanged GREEN contract | Next gate |
| --- | --- | --- | --- |
| CR01 | parent source_facts.go/tests and actual-builder/validator seam | Retained duplicate equal/conflicting parameters in both scopes; exact pointers/facts/diagnostics, legal override, independent wrong/omitted facts and7 cells | focused GREEN then coherent commit |
| WR01 | parent same facts paths | Actual completed component decoder count+bytes including preparation, multiple operations/doc roles; semantic facts/citations/cells; extra actual decode falsifies oracle | no new uninstrumented unchanged corpus; corrected corpus race at final20m |
| CR02/03/06 | parent source_lane_rules.go/tests and builder | Real Notion/Jira/CircleCI collections; scalar/metadata negatives; monotonic unknown composition; deterministic bounded traversal/exhaustion with real100000visit fixture | coherent source-shape GREEN |
| CR04/07 | same099d isolated binding files only | Admitted correct200 plus unresolved201/root, trusted-proof cannot promote; admitted direct204 and emitted fixed query contract | helper full snapshot/RED/GREEN/commit handoff then normal integration |
| CR05/WR02 |118B read-only then Firstmate disposition; parent implementation only when supplied | Existing authority exact join/negative controls; genuine original evidence recovery only, no reconstructed original | preserve pending |
| final integrated | parent shared source/evidence/artifact | exact name-union normal/race, corpus isolated20m, two generations full-byte equality,4343keys30401cells/schema/pins/read-only identity, builds/vet/embed/agentcontract/docs | freeze handoff; Firstmate-authored independent closure |

Before every substantive edit/test record full original source/test/dependency snapshots in private content-addressed evidence, command/cwd/Go settings/result/raw bytes and selected events. Each row names phase reached, prior effects and ownership, expected/actual output and oracle controls. Preserve setup/compiler failures distinct from behavior RED, and later controls/replay distinct from original chronology. No provider/archive changes or runtime semantics. Authoring docs/Atlas/schema/manifest updated on coherent final source; PM/App/help/manual/website surface unchanged remains justified N/A unless actual correction changes that boundary.

## Correction118 implementation completed for frozen review under126

Coherent implementation through48179506 integrates118A/124, implements CR01–07 and measured WR01, preserves121 WR02 custody disposition, and adopts the corrected A/B-identical authoring report. The required normal source601/601 and isolated real corpus race1/1 pass. Current contract/evidence table and explicit PENDING final read-only validation live in VERIFICATION.md; all original RED/GREEN distinctions remain in TDD-LEDGER. This is implementation readiness for a later Firstmate-bound closure, not checkpoint acceptance. No new implementation is authorized during source-freeze overlap.


## Firstmate130 coordinated correction plan (pre-production)

Full authored130 adopted:18152 bytes SHA256 1b136a183f4edba1b2ab20bcc0629aec2d1e1bbd820a409efb428a7996da8214. Complete127 report adopted verbatim in REVIEW-CONVERGENCE before any test/production change. Current source87ad0bb3/treec27423ff049c075000f0d00bb2df6413a8a8e74f. Firstmate130 resolves discussion: no new solver, no source-byte changes, no expanded runtime/product contract. Same Astra/medium sole owner executes sequentially; no extra helper or review authorized.

GSD discuss-phase, plan-phase --tdd and execute-phase generated prompts read at pinned20297a8f; source commands resolved. Existing adapter doctor provenance remains valid. Documented non-Pi inline execution is required because this canonical single-owner assignment does not authorize workflow roles; no silent GSD exemption. Previously loaded required Go how-to/testing/errors/security/safety/design-patterns/structs/CLI/documentation and connector/exhaustive evidence skills remain applied. Verify-work and final Firstmate-authored fresh closure follow, never self-acceptance.

| Group/checklist | Owned paths and contract | Pre-fix RED / unchanged GREEN obligations |
| --- | --- | --- |
| A:CR08/CR10, CP12-05 | source_lane_collection.go, distinct130 regression tests; existing consistent supported-object predicate. Every intermediate path rejects unsupported composition; explicit scalar/array cannot become object through properties. | Actual retained loader/builder exact keys/seven lanes/scoped deficits; allOf/oneOf/anyOf at multiple depths; scalar/array single-resource, scalar records/wrappers; explicit-object/type-absent/deeper positives. Independent emitted-evidence reconstruction must reject matching fabricated claims. Preserve source/direct-read/sibling truth, three real seeds and exact occurrence/ref custody. |
| B:CR09, CP12-06/07 | source_lane_bindings.go, distinct130 regression tests; complete bounded deterministic physical-component ownership. | Start from real admitted099F stream/engine + actual retained builder. Unique physical positive, composition second-use, unresolved/exhausted second-use, direct occurrence positive; reach actual binding before refusal. Shared REST/stream/GraphQL/body/record callers and independent reviewed-proof boundary must not promote incomplete binding. Budget uncertainty is same invariant, not fabricated fourth original defect. |
| Final:CP12-03/08/09 | deterministic generated report, authoring docs/Atlas and phase evidence only as needed | Full normal once; predeclared anchored whole-parent serial race partitions including corpus and full budget parent separately; exact required-name union with601 prior names plus real additions. A/B full bytes/schema/pins/4343keys30401cells and exact cell deltas. Current static56/runtime/vet/build/help/contract/format/diff plus actual before/after two checks. Commit coherent artifact/docs, freeze exact source, full handoff to Firstmate. |

Full literal source/test/dependency CAS capture includes initially untracked regression tests before RED. Setup failures do not count; old127 private probes remain immutable and only adapted into distinct owner tests. Preserve all100/118/121/127/128 failures and chronology. No input fixture/budget reduction, increased timeout, missing-name substitution or aggregate PASS invention. PM/App/runtime help/manual/website unchanged: N/A; authoring behavior docs/help and existing Atlas proof entry remain required. No new dependency/foundation/provider/live authority.130 permits explicitly pending read-only gates at final frozen handoff under126, not pending code changes.


## Firstmate132 checked-occurrence correction

commit complete131 ledger/R1–R10 before production; incorporate immutable diagnostic assertions with explicit name mapping and augment every selected obligation before production RED. Implement checked occurrence lineage, separate physical completeness, closed ref syntax and equivalent target/record/GraphQL consumers. Complete focused GREEN and compatibility, docs/Atlas, final normal/name census, full A/B artifact adoption and source-bound static gates. Deliver frozen complete review-ready handoff under126 while continuing declared serial read-only race/check gates. Firstmate alone supplies separate fresh reviewer; no acceptance or CP13 advance.

R1 storage/use/root refs; R2 selected ancestors; R3 scalar unknown edges; R4 uniform/fixed/tail coverage; R5 GraphQL uncertainty/mismatch/delegation; R6 pointer/ref identity and syntax; R7 all role/target consumers; R8 good/bad sibling identity, seven cells, severity and proof; R9 unchanged bounds and deterministic repeats; R10 independently falsifiable exact byte oracle. Full normative rows are adopted verbatim in REVIEW-CONVERGENCE; all new augmented evidence remains pending.

Red: pending augmented owner-managed complete R1–R10 matrix on unchanged production; original131 private diagnostic REDs independently adopted, not relabeled.
Green: pending same augmented assertions after coherent implementation; all130 fixes and original evidence preserved.


## Firstmate137 — coherent CR-11 correction (type: tdd)

Task header: existing CONTEXT.md; issue https://github.com/polymetrics-ai/cli/issues/4293, working branch fm/cli-top100-declaration-batch-r1-cp12, integration fm/cli-top100-declaration-batch-r1, existing https://github.com/polymetrics-ai/cli/pull/4294 targets main per preserved API receipt. Delivery: local coherent committed candidate and complete evidence for separately authored independent CP11/CP12 reviews; no push or acceptance authorized. Sole owner Astra/medium. Existing136 completes the discuss/design boundary. Installed scripts/gsd sources plan-phase/execute-phase/verify-work and prompt plan-phase --tdd/execute-phase resolved at pinned20297a8f. Execute inline via documented non-Pi/single-owner fallback; no extra role. Previously loaded required GSD, Go how-to/testing/errors/safety/security/design/structs/CLI/docs/GraphQL skills and exhaustive review overlay remain applied.

Owned production: source_lane_collection.go local checked facts/object/coverage and explicit adapters; source_lane_rules.go structural item/root/closed-object/envelope; source_lane_bindings.go narrow decoding/projection/GraphQL/record adapters. Only meaningful tests, existing phase evidence, generated manifest and necessary existing guidance/Atlas updates accompany them. Provider/annotation/schema/runtime/CP11 source unchanged. No new solver, dependencies, execution foundation, source read/cache/budget or role.

## One ownership/state table

Notation: **O** established local object; **X** supported known nonobject; **U** insufficient/unsupported/malformed local evidence. **C/N/?** mean collection/noncollection/unknown cardinality, which a read maps to applicable/not_applicable/undetermined. A noncollection result needs positive exclusion evidence; unknown is not false. “Explicit” assumes correct retained role/citation/custody and a claim selecting the stated node. Invalid means a structured contradictory/malformed assertion error; unresolved means a scoped deficit. No interpretation result supplies target/proof authority.

| State / representative input | Common local evidence and owner | Structural read result | Explicit interpretation / projection consequence |
| --- | --- | --- | --- |
| Explicit `type:object`, absent properties or valid properties map, including `{}` | O; actual object identity does not depend on nonempty fields | As homogeneous array item: C. Bare/open root object: ?. Closed fixed root can be N only under the separate complete exclusion check | Valid single_resource: N; collection may traverse declared named properties to a supported array. Projection still checks selected constraints/requiredness. |
| No type, **valid properties map** (empty or nonempty), no competing typeless items/prefix | O inferred by the already supported properties-only convention; preserve that inference provenance | Array item: C; open root: ?; closed fixed root: N if every actual child is established N | Same object eligibility as above. No arbitrary object inference from raw-byte length. |
| Bare `{}`, absent schema, or typeless `items`/`prefixItems` without established array type; typeless properties plus items/prefix | U for the required local shape; items keywords do not establish an array or override ambiguous typeless structure | ? | Unresolved shape for an unestablished object/array; no accepted collection or single-resource claim. An actually absent selected property/pointer is still the existing invalid-selector error. Preserve binding's typeless-array unverified guard. |
| Explicit scalar string/number/integer/boolean/null, with or without stray properties | X; explicit supported type wins | Root or homogeneous scalar-only items: N; never object records | Object or collection-target assertion: invalid. A scalar property mapping may still be supported; a property traversal beneath it is a contradiction, not object evidence. |
| Explicit array item containing properties (confirmed negative1), including nested array of objects | X relative to object requirement; nested sequence is not one object record | Outer array: ?; no flattening. Root array's properties cannot activate the closed-object fallback or envelope rule | Claim that the item/wrapper is an object: invalid. Separate supported array coordinates remain subject to their real consumer. |
| Type array such as `["object","null"]`, `null`, numeric type, empty/unsupported string type, even with properties/closed fields (confirmed negative2 includes mixed item type) | U; preserve type presence and decode/enum failure rather than treating it as type absent | ?; neither properties nor agreeing descendants rescue it into C or N | Unresolved shape; checked selected lineage unverified. No union-type solver. |
| `items` missing, `null`, boolean, number, string, array-valued old tuple encoding; or present item schema `{}` | U after existing object resolver or local evidence; `{}` is an unconstrained schema, not known nonobject | ? | Unresolved collection shape, not a successful assertion. A malformed *authored pointer/tag/ref* remains an error under its separate syntax contract. |
| Item properties malformed (`[]`, `null`, scalar), with or without `type:object` | U for supported object shape; valid declared object identity must not certify malformed object structure | ?; root closed-object exclusion also ? | Unresolved effective shape; checked projection already refuses malformed properties. This closes the local malformed-properties sibling without recursively validating unrelated children. |
| Explicit array; absent prefix; supported object `items` | Uniform coverage + O items | C, even with no minItems or with zero runtime records | Valid collection: C. Whole-array consumer may use `[]` only after its stricter item/field/target checks. |
| Explicit array; `prefixItems:[]`; supported object or scalar items | Valid empty prefix is uniform, exactly as131's existing binding contract | O items: C; scalar items: N | Same collection eligibility as absent prefix. Current explicit raw-presence refusal is to be replaced; require a genuine pre-edit RED for this compatibility sibling. Empty prefix is not a fixed tuple. |
| Nonempty prefix plus object tail, including scalar prefix/object tail (confirmed negative3), all-object prefix/object tail, or object prefix/scalar tail | Prefix/tail coverage, never uniform from `items` alone | ? for whole response collection; do not claim all records from a tail | Collection interpretation unresolved. Whole-array stream remains unverified. Supported fixed `[N]` mappings retain their distinct coordinates; tail remains tail. |
| Fixed nonempty prefix with absent/false items, even all object entries; malformed/null/scalar prefix | Fixed or unknown coverage, not a homogeneous sequence in the supported classifier | ?; no tuple enumeration, flattening, “all prefix objects” shortcut or new bounds solver | Whole collection unresolved. Valid fixed-index consumers remain available under existing131 bounds/lineage rules; malformed prefix is unverified. |
| Object item with an unrelated unknown descendant, e.g. `detail:$ref external`; valid unknown/open additionalProperties | O locally; child/cardinality/field completeness remain separate | Homogeneous outer array: C. Root open object still ?; closed-object exclusion fails if a collection-bearing child is unknown | Valid selected object role/array can survive. A mapping into the unknown field or a whole unsupported field contract remains unverified. |
| Closed fixed object, every property established scalar/fixed N and no open/unresolved collection branch | O + `additionalProperties:false` + complete child exclusion | N; open/bare object or unresolved child remains ? | An independently supported single_resource can establish N even with open metadata. Explicit collection must select a real supported array, not that object. |
| Root composition-only with existing bounded analysis proving all alternatives the same cardinality; scalar-only agreeing alternatives; mixed/unknown branches | Same-schema consensus/effective conjunction, not distinct-response positive merge | Existing supported agreement C or N; mixed/unknown/incompatible branch ?. Preserve simple supported cases; no solver expansion | Selected interpretation/field lineage crossing composition remains unresolved under current supported path grammar. It can be stricter than aggregate cardinality. |
| Composition at selected wrapper/item, or explicit known type with unresolved/incompatible same-node composition | U for object/path eligibility; an unknown conjunct cannot be erased by properties/type or another known branch | Affected array/item ?. Existing root conjunction analysis must retain unresolved/incompatible constraints; unsupported type cannot be “recovered” through a known branch | Unresolved interpretation/selected lineage. `allOf` that merely omits type is not automatically a known contradiction; absent evaluator support still cannot establish the requested path. |
| Valid local ref to a supported object/array, annotation-only ref siblings | Existing resolver proves the effective shape; local helper sees resolved node, custody retains literal ref occurrence and physical lineage | Same C/N/? as the supported resolved shape | Exact operation/status/media/instance-path ownership, literal digest and support citation required. Equal bytes/shared target are not interchangeable occurrences. |
| External/missing/cyclic/invalid-escape ref, structural ref sibling, unresolved composition/ref on selected ancestry | U at affected shape/path; existing resolver/lineage owner | ? at that scope; no convenient sibling substitution | Scoped unresolved/unverified; independently bad supplied hash/key/selector remains error. No accepted reference for this claim. |
| One supported collection success scope plus distinct unknown success scope; all N scopes; only N and ? scopes; error-only object array | Existing scope reducer, independent of local evidence | Respectively C + scoped unknown limit; N; ?; error response supplies no C | Interpret only the exact selected success occurrence. Unknown siblings cannot supply field/extraction proof; an invalid explicit interpretation retains existing ETL invalidation. |
| Mutation or unknown read/mutation semantics, regardless response shape | Operation semantics owner, not shape helper | Mutation ETL N; unknown semantics ETL ? | Interpretation cannot silently redefine operation semantics or invent receiver demand. Binary/write/reverse/sync retain their own truth. |
| Local selected resolution depth/cycle failure; whole shared analysis budget exhausted; projection search incomplete | Keep current resolver/shape/analysis/projection limits and sorted traversal; no partial-positive promotion from required unresolved work | Affected scope ?; when `facts.analysis.Exhausted` is set, existing whole-operation response/request reset at rules153–155 remains authoritative | Interpretation unresolved; projection unverified/incomplete. Object locality cannot defeat actual global exhaustion. No lower test-only limits or new unbounded pass. |

The table intentionally makes three consumer distinctions: an array is known nonobject yet an array-of-arrays is **unknown collection**, a scalar-only array is **known noncollection**, and a supported object can establish **record cardinality without complete field/lineage proof**. It also distinguishes an unsupported source shape from a known wrong authored target. Mapping both to a Boolean is the recurring design error.

## Executable regression obligations

The owner must put the **whole state table** into table-driven fixtures with literal expected states before production edits. Use existing `sourceCollection122Fixture`/retained inventory writers so each case travels `source.json` → pinned inventory loader → normalized facts → classifier → real builder. Do not hand-populate `sourceFacts` for the main oracle. A focused pure table can test the local evidence/adapter boundary, but cannot replace actual-builder cases.

For every ordinary read fixture assert the exact sorted two-key set `(fixture,primary,source.a)` and `(fixture,primary,source.b)`, each observed with available normalized facts, and the literal lane order `direct_read,direct_write,binary_download,binary_upload,etl,reverse_etl,sync_transport`. Assert14 cells with no duplicate/omitted lane. Check independently expected ETL applicability, state, rule and scoped deficit/error semantics; direct_read remains applicable. For unbound fixtures assert empty accepted References/ProofRefs and no `implemented` across **every** cell. Establish expectations before calling production; never derive want from `sourceSchemaShape`, the new helper, a regenerated manifest, or matching counts alone.

For those unbound structural read fixtures, literal ETL expectations are C=`applicable/mapped_unproven/source_record_collection`, N=`not_applicable/not_applicable/fixed_noncollection_read`, and ?=`undetermined/mapped_unproven/facts_unresolved` with the exact-key, ETL-only `source_record_shape_unresolved` deficit. Add the appropriate occurrence-scoped unknown/missing-interpretation diagnostic for each case; do not weaken its key/pointer/lane checks into a global code search. Valid explicit C/N uses `source_response_interpretation`; unsupported explicit shape is ? with its selected-scope unresolved diagnostic, and an incompatible known target is ? with its selected-scope invalid error. Source.a-only annotations must preserve source.b and every unrelated cell exactly.

| Regression group / concrete obligations | Existing evidence and genuinely missing work |
| --- | --- |
| **A: exact three134 negatives**, byte-equivalent schemas printed above; controls object items, scalar items, object items with external-ref descendant | Sealed134 supplies independent actual-builder RED for all3 negatives and3 passing controls. Add permanent owner regressions, run them pre-edit and retain actual RED receipt. Current normal769 does not contain these negative structural cases. |
| **B: type/property decision table**: object with absent/empty/nonempty properties; typeless valid empty/nonempty properties; scalar+properties for each supported scalar; nested-array+properties; mixed/null/numeric/unsupported type+properties; bare `{}`; typeless items with and without properties; malformed/null properties with and without object type |122 covers bare root/closed malformed properties and a positive object descendant;130 covers many **explicit** incompatible-type interpretations. Missing structural item/root siblings and cross-adapter agreement must be run pre-edit with literal expectations. Add closed-object override witnesses: explicit array or unsupported mixed type plus `additionalProperties:false` and scalar properties must not become a fixed object. |
| **C: items/prefix coverage**: missing/null/boolean/number/string/array-valued items and `{}`; absent versus `[]` prefix for object/scalar items; scalar/object prefix plus object/scalar tail; all-object fixed tuple with absent/false items; malformed prefix |134 covers one prefix/tail negative.132 `TestSourceLane132PointersAndConsumers` covers record-coordinate refusal and valid empty-prefix/fixed-index consumers, not structural ETL or explicit empty-prefix equivalence. Add actual-builder structural and annotation cases; capture pre-edit RED for currently wrong cases, including explicit empty-prefix positive and unconstrained-shape unresolved diagnostics. Other rows may already pass and are guard controls, not manufactured RED. |
| **D: composition/ref/state boundaries**: preserve agreeing root oneOf arrays and anyOf scalars; mixed/unknown branches; unresolved allOf with a known type; composition at items/wrapper; invalid type plus known composition; valid local ref to object items/root array; nested missing/external/cycle/structural-sibling ref; valid escaped local ref and equal-byte sibling |122 boundary/reference/lineage tests,130 nested compositions,131 escape tests already cover many path-specific cases. Run them unchanged; add missing structural-versus-explicit paired siblings and exact occurrence witnesses. Do not require support for an item union/composite path the existing grammar cannot establish. |
| **E: locality and operation/scope isolation**: known object with unrelated unknown descendant; unknown field selected by mapping; root closed object with unknown child; positive200 collection plus unknown201/media; scalar-only successes; error-only collection; annotation on source.a while source.b has equal bytes; mutation/unknown semantics |122 retained boundary, response occurrence custody, interpretation order/lineage and132 direct-unknown-sibling already provide controls. New cases must check exact literal response/media FactRefs, source-scoped diagnostics and full non-ETL cell preservation against an independently valid baseline. Distinct-success C cannot supply the unknown sibling's field proof. |
| **F: actual admission and binding consumers**: negative structural ETL with intended/materialized stream claim; separate valid direct-read binding; uniform object collection with valid source/target projection; fixed-index and tail consumers; GraphQL variable object, REST/body field consumers | Preserve130/131/132 actual-builder/GraphQL/body tests. Add at least one CR11 negative with an actual shaped ETL target: `target_applicability_conflict`, intended claim retained, no accepted ETL reference/proof; valid unrelated source/lane ref remains. Positive control must reach the existing accepted-reference frontier without fabricated proof. A malformed fixture that stops before applicability/projection is not evidence. |
| **G: bounds and determinism**: existing shape atomic-budget test, projection depth/visits/cycles, normalizer/retained-node limits, sorted repeated report equivalence | Existing `TestSourceLane118ShapeBudgetAtomic`, inventory budget and131/132 tests retain their real limits. Reuse their test bodies/name sets; run after changing the helper because its callers share traversal state. Add a focused helper test showing it performs no resolution/budget increment. Do not lower100000 analysis visits,128 shape/interpretation depth,256 object-resolution depth, or4096/128 projection limits to obtain RED. |
| **H: exact retained seeds and evidence validators** | Keep122's actual Notion templates, Jira keys, CircleCI org_project_data expectations and immutable source/annotation hashes. Retain `TestSourceLane122IndependentInterpretationEvidence`,130's forged emitted interpretation and manifest citation/identity tests. No annotation/provider update is required to pass a classifier repair. |

A reusable assertion should return structured violations rather than merely terminating the test, so its own falsification is executable. Give it a literal case specification containing expected key set, lane sequence, ETL state/rule/diagnostic and reference/proof limits. Feed it a genuine builder result; then clone that result and deliberately promote a known negative's ETL cell from unknown to applicable/source_record_collection, recomputing summary fields as well. Even if both candidate and a supplied comparison snapshot are forged to agree, the assertion must emit the **specific ETL-state violation for the exact key**, because its expected state came from the independent table. A control clone must pass. This is the required critical reusable-oracle falsification, not a production injection or a second classifier. Also keep the existing production interpretation-evidence forgery test, which independently consults retained source/annotation custody.

Proposed permanent parent tests may be named `TestSourceLane136CollectionInvariant`, `TestSourceLane136CollectionInterpretationAgreement`, `TestSourceLane136CollectionAdmission`, `TestSourceLane136LocalShapeEvidence` and `TestSourceLane136CollectionOracleFalsification`. Names are implementation suggestions; record the actual names. Include complete parent/subtest names in RED/GREEN receipts. Missing behavior gets actual pre-edit failing assertions at the intended frontier; existing passes remain controls. The diagnosis itself did not execute any of these proposed tests or claim them covered.


### Stable acceptance matrix / contract to evidence

| ID | Contract and observable boundary | Evidence / reason | Status |
| --- | --- | --- | --- |
| CP12-05 / CR-11 A–E | Full literal state table through retained source→inventory→facts→classifier→builder; exact two keys/14 ordered cells, scoped ETL states/diagnostics and unaffected lanes | Fake provider documents: hermetic retained input intentionally expresses unsupported schemas; actual production pipeline and returned cells are asserted | RED pending |
| CP12-06 / CR-11 F | Actual admitted stream/direct target, negative applicability conflict and no accepted ETL ref; separate positive accepted ref and unrelated direct ref preserved | Disposable local authoring target; no provider or credential authority; real admission/consumer executes | RED pending |
| CP12-09 / CR-11 G–H | Independent reusable oracle rejects exact-key coherently forged output plus matching forged comparator; all existing source-owned proof/boundary/seed defenses retained | Actual builder result mutated only in assertion falsification; genuine control clone passes | RED pending |
| CP12-09 final | Whole cmd/connectorgen normal names; exact disjoint race union including CP11 and new137; artifact A/B/schema/census/pins/check-mode identity; original-range lint/vet/build/docs | Real local package/CLI/artifacts, hermetic provider scope | pending |
| CP11-08/09 and CP12 acceptance | Separate fresh exact-candidate judgments, complete disposition and Firstmate acceptance beforeCP13 | Reviews not yet assigned;136 designer cannot accept own design | pending |

Before production: commit full ledger/plan, then anchored new parent run with immutable source/test/dependency/raw receipts; distinguish controls, semantic failures and fixture/compile failures. After complete coherent GREEN/refactor and lint, commit/freeze and notify immediately, then whole-package normal and serial race groups each -timeout20m. Compare630 CP11 and769 CP12 predecessor test names as sets; no omissions/skips/aggregate-fiction. Explain actual artifact differences without changing source pins. No partial success or failure suppression.


### Firstmate137 candidate freeze boundary

A/B generation exit0 at61.627790292s/59.357604875s; both101749993bytes SHA25629dd14f5974911543f7b176b72bafabe02c79a1d984dfd139191f6b2d9076bbb. Actual closed-schema/census exit0/10.816240708s/raw793d8d78196f34cad7cd16c068d0d823c87121eb0ae0f5914f6bb95fee89cd9e:4341primary+2supplement/4343exactkeys/30401orderedcells/14documents17pins;18501not_applicable11899mapped_unproven1missing_foundation0implemented;22185deficits0errors; no accepted binding/proof/intended refs. Every source/fact/document/lane/diagnostic equals739 predecessor. Only Atlas input hash changes; complete byte-validated generated candidate adopted after census. Vet exit0/1.186137167s; build exit0/6.154566042s.

Verify-work installed prompt resolved and applied inline under prior single-owner/non-Pi fallback. CP12-05/06/09 focused137 repair assertions and impacted391controls pass, source/table implementation complete for independent judgment; full current package normal/race/name union and candidate/default preservation checks remain pending at freeze. This is deliberately not whole-checkpoint or CP11 acceptance. Complete current full-package validation must include inherited CP11 parents, preserve630094 and769133 historical names as sets, and add137 names. Independent fresh CP11/CP12 reviews require separate Firstmate prompts. Source will remain frozen through088 overlap; all terminal future private receipts bind this committed candidate.
