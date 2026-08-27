# Schema-v3 source projection/importer foundation — TDD ledger

## Planned red/green/refactor slices

| Slice | Red assertion before production code | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| Source-reference admission | The cited-only Outreach input fails in the current byte-backed reader/importer path. | An explicit declaration-only result preserves source URL, source operation identity, and declaration identity without claiming byte import. | No network/cache fallback, re-pin, descriptor identity rewrite, or weakening of byte-backed verification. |
| Shared reader proof | A second independent schema-v3 fixture fails at the same absent source-reference boundary. | It produces a distinct cited operation descriptor through the same common code. | Unsupported/malformed kinds still fail with contextual errors and size/path bounds. |
| Six-lane projection | A recognized unsupported operation is omitted, incorrectly blocked as provenance, or promoted to `implemented`. | It has six visible lane classifications and a specific `missing_foundation` reason. | Existing executable operations keep their canonical mapping and preflight remains the implementation authority. |

## Red

2026-08-26 — executed before production edits:

```text
go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport(OutreachReferenceProjectsCitedOperationsWithoutFetching|V3SourceReferenceUsesTheSameClosedProjection|SourceReferenceRejectsUnsupportedAndUnsafeKinds|ReferenceFeedsOperationEvidenceWithoutEnabledClassification)$' -count=1
```

Failed as intended at the closed-reader boundary:

- `TestSourceImportOutreachReferenceProjectsCitedOperationsWithoutFetching`:
  `parse source lock: json: unknown field "operation_counts"`.
- `TestSourceImportV3SourceReferenceUsesTheSameClosedProjection`:
  `parse source lock: json: unknown field "source_reference"`.
- malformed-source cases cannot reach their intended validators because the
  parser rejects the new input shape first.
- operation evidence attributes the Custom Objects URL to the primary OpenAPI
  hash/byte record and has no `source_contract_unavailable` classification.

This proves the reader has no explicit cited-only source-reference path and
would otherwise conflate an unavailable raw contract with byte-backed import.

## Green

2026-08-26 — the red command is green after adding the explicit
`source_reference` document kind and the narrow retained-Outreach legacy
adapter. It proves all of the following without a network call:

- the Outreach OpenAPI and Custom Objects operation citations retain their
  distinct URL/digest/byte identities;
- the same v3 reader path accepts a non-Outreach source-reference fixture;
- unsupported source kinds and unsafe citation URLs still fail strict parsing;
- source-import writes and checks the descriptor through its normal CLI path
  with no retained-artifact directory and reports `writes=0 cli=0`;
- source projection does not alter a matching implemented write/command; and
- operation evidence emits all six lane classifications as declared but not
  enabled, with `source_contract_unavailable` rather than a hash,
  certification, or credential failure.

```text
go test -timeout 20m ./cmd/connectorgen -run 'Test(SourceImport(OutreachReferenceProjectsCitedOperationsWithoutFetching|V3SourceReferenceUsesTheSameClosedProjection|ReferenceDigestIsProvenanceNotAnExecutionGate|SourceReferenceRejectsUnsupportedAndUnsafeKinds|ReferenceFeedsOperationEvidenceWithoutEnabledClassification)|RunSourceImportReferenceChecksWithoutRetainedArtifactOrSurfaceWrite|SourceReferenceProjectionDoesNotMaterializeAnExistingWriteOrCommand)$' -count=1
PASS

go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport(RetainedArtifactRejectsMissingAndMismatchedCopies|RejectsArtifactDriftAndSizeBeforeParsing|Version3RenderedReferenceProjectsCapturedCitation|RejectsUnsafeArtifactDestinations)$' -count=1
PASS
```

The second command retains the strict byte/digest identity and captured
cross-origin citation regressions for byte-backed import; the new path does not
reuse or weaken it.

## Refactor / review

Pending. Record error-wrapping, defensive-copy, ordering, provenance, and
generic-escape review outcomes after the final implementation.

## Exact-head R3 B1/B2 repair — planned before production edits (2026-08-27)

| Slice | Red assertion | Green assertion | Boundary retained |
| --- | --- | --- | --- |
| B1 full legacy reference with provider absence | The complete 259-row valid legacy reference fixture plus either `state: skipped` or `state: dynamic` is returned as provider absence and loses its source operations. | Both inputs fail strict reference validation before absence projection. | Null/empty discriminators and ordinary reference-only legacy field drift still fail; ordinary byte-backed absence remains supported. |
| B2 v3 marker/descriptor bypass | Changing only `source_documents[].kind`, retaining raw citation fields, and removing the descriptor’s exact foundation gap lets `surface-sync` project the descriptor. | The lock is parsed and descriptor-bound before projection, so malformed citation evidence and gap drift fail closed. | Exact provenance swap / exact-gap tests stay green; no citation becomes materializable. |

### Red command

```sh
go test -timeout 20m ./cmd/connectorgen \
  -run 'Test(OperationEvidenceLegacyReferenceCannotHideBehindProviderAbsence|SourceReferenceSurfaceSyncRejectsMarkerBypass)$' \
  -count=1 -v
```

Expected pre-green observations:

- `skipped` and `dynamic` report a provider absence for the full source
  inventory rather than rejecting the contradictory reference form.
- A v3 lock with citation raw fields but an altered `kind` avoids the old
  marker-only gate, and a descriptor with the exact gap removed reaches the
  projection path.

Production edits are prohibited until this red command has failed and is
recorded below. The green command is identical plus existing exact-contract and
provenance-swap regressions.

### Red — 2026-08-27

The planned command failed before production edits:

```text
TestOperationEvidenceLegacyReferenceCannotHideBehindProviderAbsence/skipped:
schema-v2 source-reference inventory was accepted behind state="skipped"
TestOperationEvidenceLegacyReferenceCannotHideBehindProviderAbsence/dynamic:
schema-v2 source-reference inventory was accepted behind state="dynamic"
TestSourceReferenceSurfaceSyncRejectsMarkerBypass:
surface-sync accepted a citation descriptor after marker and exact-gap tampering
```

This independently reproduces both R3 blockers on the full fixture. The next
production change may only move strict parsing ahead of the legacy absence
return and make checked-in source projection parse/bind every lock before its
descriptor reaches materialization.

### Green — 2026-08-27

The same focused run plus the pre-existing provenance and exact-gap suite is
green after the smallest shared foundation change:

```sh
go test -timeout 20m ./cmd/connectorgen \
  -run 'Test(OperationEvidenceLegacyReferenceCannotHideBehindProviderAbsence|OperationEvidenceProviderAbsenceRejectsReferenceOnlyFields|SourceReferenceDescriptorsRejectExactContractTampering|SourceReferenceSurfaceSyncRejectsTamperedDescriptor|SourceReferenceSurfaceSyncRejectsMarkerBypass|SourceImportLegacyReferenceRejectsCrossSourceCitationSwap|ExactOutreachReferenceProjectsAgainstCurrentMainBundle)$' \
  -count=1 -v
```

- B1 includes `legacyReference` in the absence guard, so both absent-state
  variants reach `parseSourceImportLock` rather than returning a zero-row
  provider-absence envelope.
- B2 removes the marker-only helper from the source-sync trust boundary.
  `syncCheckedInSourceProjection` fully parses every checked-in lock and
  validates every descriptor before projection. The adversarial altered kind
  fails strict document validation before a descriptor can materialize.
- The exact 259-row, six-lane, provenance-swap, and exact gap tests remain
  green. No new command, request contract, credential flow, provider call, or
  certification condition was added.

## Current-main integration gap — red (2026-08-27)

After merging `origin/main@2165619e`, the global generator backstop failed:

```text
connectorgen validate: github source descriptor provenance drift for
orgs/add-security-manager-team
connectorgen surface-sync: github source projection: validate canonical source
descriptor: source descriptor provenance drift for
actions/set-workflow-access-to-repository
```

The merge had retained an exact full-provenance comparison outside the
citation-only `expectedOperation.reference` branch. That comparison is needed
for a citation descriptor only, where the existing deep equality already owns
connector/protocol/method/path/provenance/empty contract/gaps. For ordinary v3
documents, current main owns the narrower provider contract comparison.

### Current-main integration gap — green (2026-08-27)

The execution retained `reflect.DeepEqual(operation, *expectedOperation.reference)`
for every citation-only descriptor, which already includes exact connector,
protocol, method, path, provenance, empty request/response/output, merge-block,
and exact gap state. It removed duplicate broad source-hash/provenance checks
for ordinary documents, preserving current main’s endpoint/identity comparison.

```text
go test <B1/B2 and provenance set>                 PASS
connectorgen validate internal/connectors/defs     553 / 0 findings
connectorgen surface-sync --check ...              553 scanned / zero drift
connectorgen operation-evidence . --check          1,774 rows / 5 rollups / fixed-100
```

### Current-main projection-compatibility gap — red (2026-08-27)

The full changed-package run failed two established synthetic projection tests:

```text
TestSourceProjection_MissingOperationOrFieldFailsValidateAndSurfaceCheck
TestSurfaceSyncAcceptsSchema3SourceDescriptor
parse source lock: source lock has unsupported schema version 0
```

Those tests used deliberately invalid `{}` lock fixtures to cover the
materializer independently of source-import admission. They are not a reason
to restore a marker-only source-reference gateway or to tolerate an invalid
lock. The green design is unconditional: parse every checked-in lock strictly,
bind every descriptor to that parsed lock, and reject before projection. The
tests now build minimal valid legacy-v1 and v3 inventories for their projection
assertions and separately assert that `{}` fails the strict parser.

### Current-main projection-compatibility gap — green (2026-08-27)

`TestSourceProjection_MissingOperationOrFieldFailsValidateAndSurfaceCheck`
now writes a minimal valid legacy lock whose operation identity and provenance
match the synthetic descriptor; it also asserts `parseSourceImportLock({})`
rejects schema version zero. `TestSurfaceSyncAcceptsSchema3SourceDescriptor`
uses the standard valid v3 document fixture and its derived descriptor
identity/provenance. With those valid inputs, `surface-sync` has one trust
boundary for all locks: strict parse followed by descriptor validation before
materialization.

## Verify repair red

2026-08-26 — GitHub Verify runs `32978972653` and `32978973162` both failed
at the same source-import help contract. Local reproduction:

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportCommandContractAndMigrationDocumentation$' -count=1
FAIL: source-import help is incomplete
```

The new correct help text wraps the historical literal `retained artifact`
across a line boundary and explicitly introduces the safe declaration-only
source-reference path. The assertion must move from fragile layout-dependent
text to the real contract: byte-backed retained documents remain strict and
the reference path emits `source_contract_unavailable`; no importer behavior
will change for this repair.

## Verify repair green

2026-08-26 — the assertion now checks the explicit closed contracts instead of
line layout. It requires byte-backed provenance language, the declaration-only
reference path, its precise `source_contract_unavailable` gap, and retained
content-addressed document behavior while continuing to refuse arbitrary
request/cache flags.

```text
go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportCommandContractAndMigrationDocumentation$' -count=1
PASS

go test -timeout 20m ./cmd/connectorgen -count=1
PASS (146.529s)

go vet ./cmd/connectorgen
PASS

git diff --check
PASS
```

## Independent-audit repair wave — planned red/green

2026-08-27 — independent audit report
`cli-source-reference-projection-reaudit-codex-r1/report.md` rejected head
`2738c6a9ff7172c74bedbcede092a77f16a05ba2` on four findings. Before changing
production code, add and run tests that prove:

- cited-only GET descriptors cannot downgrade an existing direct-read command
  or replace its API-surface coverage;
- the Outreach proof parses the exact 259-row lock with real source IDs,
  source counts, and citations against the current-main bundle without
  materializing declarations;
- ordinary byte-backed schema-v1/v2 locks reject each reference-only field;
  and
- v3 cited-only operations reject malformed protocol, HTTP method, identity,
  location, mixed citation fields, and duplicate routes/identities through the
  same closed rules as legacy references.

Green may only filter non-materializing reference descriptors, split the
legacy decoder, and share the closed validator. It must not relax byte-backed
identity verification, add provider I/O, create a command, or convert a hash
into a credential/certification gate.

### Red — 2026-08-27

```text
go test -timeout 20m ./cmd/connectorgen -run 'Test(SourceImportOutreachReferenceProjectsCitedOperationsWithoutFetching|SourceReferenceProjectionDoesNotDowngradeExistingDirectReadSurface|SourceImportLegacyByteBackedLocksRejectReferenceOnlyFields|SourceImportV3SourceReferenceRejectsClosedOperationIdentityViolations)$' -count=1
FAIL
```

- The former Outreach proof yielded two descriptors, not 259, and used the
  shortened source IDs.
- A cited-only GET changed a matching implemented direct-read command
  (`CLI:1`) before the operation-loop skip.
- Every listed reference-only field was accepted by ordinary v1/v2 byte-backed
  decoding (the v2 discriminator field instead reached a later reference
  validation error).
- V3 cited-only operations accepted lowercase/unsupported methods, untrimmed
  IDs, control-bearing locations, and untrimmed provider operation IDs.

This is the required red state for H1, H2, M3, and M4. No production file had
been changed when it was observed.

### Green — 2026-08-27

```text
go test -timeout 20m ./cmd/connectorgen -run 'Test(SourceImport(OutreachReferenceProjectsCitedOperationsWithoutFetching|V3SourceReferenceUsesTheSameClosedProjection|ReferenceDigestIsProvenanceNotAnExecutionGate|SourceReferenceRejectsUnsupportedAndUnsafeKinds|ReferenceEncodesEveryDeclaredLaneWithoutPromotion|LegacyByteBackedLocksRejectReferenceOnlyFields|V3SourceReferenceRejectsClosedOperationIdentityViolations|V3SourceReferenceRejectsDuplicateOperationIdentityAndRoute)|ExactOutreachReferenceProjectsAgainstCurrentMainBundle|RunSourceImportReferenceChecksWithoutRetainedArtifactOrSurfaceWrite|SourceReferenceProjectionDoesNotMaterializeAnExistingWriteOrCommand|SourceReferenceProjectionDoesNotDowngradeExistingDirectReadSurface)$' -count=1
PASS

git diff --check
PASS
```

- `sourceProjectionMaterializableResult` now removes every
  `source_contract_unavailable` descriptor before any read, CLI, API-surface,
  flag, or write transform. The red direct-read GET now leaves all three
  declaration files byte-identical in normal and `--check` modes.
- The exact gzip-encoded candidate fixture decodes to 100124 bytes whose
  SHA-256 is `f733248bfd484625b8f2bae3490b3211f7e158ab375d3c8de5ede83b1f369f89`.
  It proves all 259 rows, 253 OpenAPI/6 custom-document citations, the two
  real selected operation IDs, normal source-import write/check behavior, and
  an unchanged checked-in Outreach canonical bundle.
- A separate six-lane encoding unit keeps each classification visible while
  every cited-only lane remains disabled. The real current-bundle evidence
  proof confirms the `GET /api/v2/prospects` canonical mapping and the same
  unavailable disposition without manufacturing a command.
- The normal legacy decoder is again closed. Schema-v2 has a distinct
  `source_kind` discriminator and strict cited-only wire type; reference-only
  root, REST, and operation fields are rejected everywhere else.
- One reference-operation validator now requires the REST protocol, an
  uppercase allow-listed method, a valid route, canonical source ID/location,
  no control code points, and a canonical optional provider operation ID. V3
  reference rows also reject citation URL/binding mixtures.

### Refactor / review — repair wave

- Kept the provenance bytes in a compressed test constant with a runtime
  SHA-256 assertion; no provider source or production connector file is
  rewritten.
- Reviewed the projection entry point for generic HTTP/command/shell/SQL
  expansion: it only filters non-materializable descriptors and cannot create
  a request or executor.
- Reviewed source-reference digest flow: the digest remains copied citation
  provenance and never enters credential, certification, or execution
  admission logic.

## Fresh admission hardening — planned red/green (2026-08-27)

The audit of exact head `6a372ee216112d4a83c00d0d687a96dc438abf84` found two
post-repair bypasses. Before production edits, add regressions proving:

- a schema-v2 `skipped` or `dynamic` record carrying the complete legacy
  inventory plus `rest.source_kind: null` cannot return an absence record; and
- a legacy/v3 citation-only descriptor cannot remove/replace its sole gap,
  change connector/protocol/method/path/provider identity/provenance, or add
  a request, response, or output contract. The same tamper must stop
  `surface-sync` before it can project a command/action change.

The implementation may only reuse raw v2 presence validation before absence
projection and exact reference-operation construction for descriptor checking.
The untouched byte-backed legacy reader and all cited Outreach operations must
remain non-executable.

### Red — 2026-08-27

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'Test(OperationEvidenceProviderAbsenceRejectsReferenceOnlyFields|SourceReferenceDescriptorsRejectExactContractTampering|SourceReferenceSurfaceSyncRejectsTamperedDescriptor)$' \
  -count=1 -v
FAIL
```

- Both `skipped` and `dynamic` returned provider-absence records before v2 raw
  field inspection, accepting every reference-only overlay, including explicit
  `rest.source_kind: null`.
- Descriptor validation accepted connector/protocol/method/path changes,
  removed/replaced gaps, and invented request/response/output contracts for
  both the retained legacy reference and a v3 source-reference document.
- `surface-sync` accepted the same gap-removed descriptor in both forms,
  reaching projection instead of refusing the lock/descriptor mismatch.

### Green — 2026-08-27

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'Test(OperationEvidenceProviderAbsenceRejectsReferenceOnlyFields|SourceReferenceDescriptorsRejectExactContractTampering|SourceReferenceSurfaceSyncRejectsTamperedDescriptor)$' \
  -count=1 -v
PASS
```

- `operationEvidenceLegacyReferenceWire` now runs before either v2 absence
  return, so the existing raw-presence matrix rejects all reference-only
  fields for both `skipped` and `dynamic` and retains ordinary v2 absence
  behavior when no such field is present.
- Reference descriptor validation constructs the exact expected lock
  projection and rejects any connector/protocol/method/path/provider identity,
  provenance, execution-contract, operation runtime, or pure-reference root
  gap-state drift.
- `surface-sync` identifies a source-reference lock, performs strict lock and
  descriptor validation before materialization, and therefore rejects the
  tampered descriptor without reaching a command/action projection path.

### Refactor correction — 2026-08-27

The first complete `cmd/connectorgen` run found that carrying descriptor schema
version into the existing materializer activated a v3 execution-envelope rule
for an unrelated synthetic source-projection fixture. That propagation was not
needed for the new pre-materialization citation check and was removed. The
focused source-reference plus `TestSurfaceSyncAcceptsSchema3SourceDescriptor`
set then passed, followed by the full package in `168.282s`.

## Frozen independent-audit repair wave — planned red/green (2026-08-27)

The fresh independent audit of head `21480fcd9ce5701164bafb82666ffe5bbc3934c4`
identified three remaining, independently reachable contract failures. No
production source file has been edited for this wave.

| Finding | Red observable assertion | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| F1 | `source-import` writes a descriptor that `validate` rejects for legacy-v2 reference and v3 `source_reference` locks. | Both CLI-path source-import → validate sequences succeed with the descriptor's actual v3 provenance preserved. | The shared helper cannot change byte-backed v1/v2 expected schemas or turn a citation into an executor/certification condition. |
| F2 | A valid byte-backed v2 lock with a reference-only overlay (including a null discriminator) reaches operation evidence. | Every overlay is rejected before tolerant evidence projection; an untouched byte-backed v2 lock still produces its legacy row. | Inspect raw field presence and retain duplicate-member detection; do not silently discard or reinterpret unknown reference fields. |
| F3 | Swapping one primary and one supplemental operation URL preserves counts yet passes legacy-reference admission. | Admission rejects the swap because each operation's exact source location belongs to its declared primary/supplemental document. | Preserve all declared source artifacts, counts, URLs, and no-fetch semantics; do not normalize or rewrite citations. |

### Red — 2026-08-27

Before any production edit, the added regressions failed as required:

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'Test(SourceImportReferenceDescriptorsValidateThroughCLI|OperationEvidenceLegacyByteBackedLocksRejectReferenceOnlyFields|SourceImportLegacyReferenceRejectsCrossSourceCitationSwap)$' \
  -count=1 -v
FAIL
```

- F1 legacy-v2 reference: `source descriptor schema_version = 3, want 2`.
- F1 schema-v3 reference: `source descriptor provenance drift for
  outreach.rest.outreach-source.get`.
- F2: each of `operations_found`, `coverage_confidence`,
  `rest.operation_counts`, `rest.supplements`, operation `source_url`, and
  `rest.source_kind: null` was accepted by the operation-evidence reader.
- F3: the two-operation primary/supplemental citation URL swap remained
  admitted even though the source counts were unchanged.

### Green — 2026-08-27, before current-main integration

```text
go test -timeout 20m ./cmd/connectorgen \
  -run 'Test(SourceImportReferenceDescriptorsValidateThroughCLI|OperationEvidenceLegacyByteBackedLocksRejectReferenceOnlyFields|SourceImportLegacyReferenceRejectsCrossSourceCitationSwap)$' \
  -count=1 -v
PASS
```

- One shared reference-provenance constructor is used by import and validation.
  It selects `sourceArtifact()` for schema-v3 references, preserves the
  reference form/version, and makes legacy cited descriptors schema v3 without
  changing byte-backed v1/v2 expectations.
- The operation-evidence reader scans raw v2 field presence before its tolerant
  path. Only a non-empty `rest.source_kind` reaches the strict legacy-reference
  parser; null and all reference-only overlays fail closed while an unchanged
  byte-backed v2 fixture still yields one evidence operation.
- Supplement-attributed operations must repeat the supplement's declared exact
  source location, so the count-preserving cross-source swap is rejected.
- Next: merge `origin/main` normally and repeat this suite plus the required
  generator, engine, runner, docs, boundary, and release checks.

## Exact-head R3 B1/B2 final green evidence — 2026-08-27

The corrected strict-parser implementation passed the full changed package:

```text
go test -timeout 20m ./cmd/connectorgen
PASS (196.142s)
```

The focused suite also passed against the exact complete fixture. It covers
both B1 absence overlays, the v3 altered-marker/exact-gap bypass, every
reference-only absence field including `source_kind: null`, exact
connector/protocol/method/path/provenance/empty-contract/root-gap tampering,
and the primary/supplemental provenance swap.
