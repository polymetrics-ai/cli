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

## Frozen independent-audit repair wave — planned red/green (2026-08-27)

The fresh independent audit of head `21480fcd9ce5701164bafb82666ffe5bbc3934c4`
identified three remaining, independently reachable contract failures. No
production source file has been edited for this wave.

| Finding | Red observable assertion | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| F1 | `source-import` writes a descriptor that `validate` rejects for legacy-v2 reference and v3 `source_reference` locks. | Both CLI-path source-import → validate sequences succeed with the descriptor's actual v3 provenance preserved. | The shared helper cannot change byte-backed v1/v2 expected schemas or turn a citation into an executor/certification condition. |
| F2 | A valid byte-backed v2 lock with a reference-only overlay (including a null discriminator) reaches operation evidence. | Every overlay is rejected before tolerant evidence projection; an untouched byte-backed v2 lock still produces its legacy row. | Inspect raw field presence and retain duplicate-member detection; do not silently discard or reinterpret unknown reference fields. |
| F3 | Swapping one primary and one supplemental operation URL preserves counts yet passes legacy-reference admission. | Admission rejects the swap because each operation's exact source location belongs to its declared primary/supplemental document. | Preserve all declared source artifacts, counts, URLs, and no-fetch semantics; do not normalize or rewrite citations. |

### Red — pending execution

Add the regressions first and run the focused command. It must fail against
the frozen head before implementation changes, naming F1/F2/F3 independently.

### Green — pending execution

After the smallest shared repair, repeat the focused command, merge current
`origin/main` normally, then repeat it and record all generator/engine/runner
gates in `VERIFICATION.md`.
