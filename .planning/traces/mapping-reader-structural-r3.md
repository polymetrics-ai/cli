# Mapping-reader structural validation R3

## Delivery header

- Scope: mapping-only declaration-admission source-lock reader.
- Base: `14aa19c76c327617216a891f394c9a658819208f`.
- Working branch: `codex/asana-source-operation-compat-r1`.
- R3 code SHA: `72bac825659c13ce9b2219f06b0f031ac0ec43af`.
- Delivery: local commits only; no push, PR creation, merge, source-lock edit,
  connector-definition edit, runtime edit, or source-import behavior change.
- This trace commit is artifact-only. It binds the code SHA above and must not
  be treated as a fresh review of that code.

## Ancestry and review provenance

The branch began from the Batch 1 authoring chain at `14aa19c76c327617216a891f394c9a658819208f`.
It contains the closed Asana strict-import envelope repair
`69224bc7600fda6197c0226473d9e2b444fcc967`, the R3 red-test checkpoint
`815a5121f`, and the normal identical cherry-pick
`aedc7a97d` of the prior R2 mapping-reader code. The cherry-pick is the local
provenance equivalent of reviewed code `2e4f47e6e33c0711aef50089b8a706faf59aab02`;
no no-content merge was made.

The frozen R2 review found three remaining mapping-reader defects:

1. **R2-01:** the common wire could select legacy/v3 variants through empty or
   `null` Go zero values and accept cross-version fields.
2. **R2-02:** legacy `source_reference` primary and supplement source-form pins
   were not validated in mapping-only admission.
3. **R2-03:** the Foundation Atlas lacked proof-test inventory for unavailable,
   v3-envelope, and rendered-reference structure.

## R3 repair

- `decodeDeclarationAdmissionMappingSourceLockWire` selects schema-v1,
  schema-v2 ordinary, schema-v2 source-reference, and schema-v3 closed wire
  shapes before the existing common mapping projection runs.
- Legacy selection uses presence of `rest.source_kind`, not its decoded zero
  value. Ordinary v1/v2 wires reject source-reference-only root/rest/operation
  fields even when their value is empty or `null`.
- The v3 gate selects each document kind before decoding it. Normal,
  rendered-reference, bundle, source-reference, and unavailable documents keep
  their own closed operation field sets, so a foreign `source_url` cannot be
  smuggled through an empty or `null` value.
- `source_operation` and top-level `source_contract` remain `json.RawMessage`;
  malformed hashes, bytes, capture fields, and those opaque provider leaves
  remain non-binding to mapping admission.
- The existing source-reference semantic checks now validate both primary and
  supplement OpenAPI 3.0/3.1 or Swagger 2.0 source-form pins while retaining
  absent pins as valid and leaving retention representation to source-import.
- The Foundation Atlas records the R2 unavailable/v3/rendered tests and the R3
  closed-wire/source-form proofs.

## Red-to-green evidence

The red checkpoint was committed in `815a5121f` before the R3 implementation:

```text
go test ./cmd/connectorgen -run 'TestDeclarationAdmissionMappingReaderR3(RejectsCrossVersionAndVariantDrift|LegacyReferenceValidatesSourceFormPins)' -count=1 -v
```

At that checkpoint the mapping reader accepted every frozen drift class:
ordinary foreign root/rest/operation fields, legacy reference citation fields,
v3 root fields, v3 `source_url` fields, and invalid legacy primary/supplement
source-form pins. The final matrix covers both ordinary v1 and v2 shapes; empty
and `null` discriminators/variant fields; all requested v3 operation document
kinds; valid rendered references; opaque `source_operation`/`source_contract`;
and valid/absent supported source forms with malformed retention.

Green verification against R3 code SHA:

```text
gofmt -w cmd/connectorgen/declarationadmission.go \
  cmd/connectorgen/declarationadmission_mapping_structural_r3_test.go
go test ./cmd/connectorgen -run 'TestDeclarationAdmissionMappingReaderR[23]' -count=1
go test ./cmd/connectorgen -run '^TestDeclarationAdmission' -count=1
go test ./cmd/connectorgen -run 'TestSourceProjection(MappingIgnoresRetentionAndEmbeddedSourceOperation|SourceReferenceIgnoresRetentionButPreservesClosedGap)' -count=1
go test ./cmd/connectorgen -run '^TestParseSourceImportLock' -count=1
make connector-canon-check
git diff --check
```

All commands above passed. `jq empty docs/connector-canon/foundations/catalog.json`
also passed.

## Full-package baseline

`go test ./cmd/connectorgen -count=1` reached this R3 code and failed only on
the five pre-existing, out-of-scope Asana importer/projection cases:

1. `TestOperationEvidenceFixed100RejectsEveryRegression` —
   `asana.rest.getCustomFieldsForWorkspace` ETL classification regressed.
2. `TestOperationEvidenceCheckRunsFixed100Gate` — same fixed-100 ETL
   classification discrepancy.
3. `TestRetainedAsanaSourceImportRejectsReadProjectionDrift` — partial mutation
   disposition cites unknown `asana.rest.createCustomField`.
4. `TestRetainedAsanaSourceImportSelectsSourceBackedFanOutETLStreams` — retained
   source import omits `getSectionsForProject`.
5. `TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation` —
   same unknown partial mutation disposition.

Those paths are source-import/source-projection/operation-evidence ownership,
not the mapping-only parser. They are deliberately not repaired here.

## Required re-review

A fresh-context reviewer must review the exact R3 code SHA
`72bac825659c13ce9b2219f06b0f031ac0ec43af`, including the R2 ancestry noted
above, before certification or merge consideration. Any later code-bearing
commit invalidates that exact-SHA review.
