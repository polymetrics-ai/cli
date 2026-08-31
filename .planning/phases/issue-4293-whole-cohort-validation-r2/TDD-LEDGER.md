# TDD ledger — issue 4293 whole-cohort validation R2

## Red

Before the production validator changed:

```text
GOCACHE=/private/tmp/gocache-4293-red.3jy0At \
  go test ./cmd/connectorgen \
  -run '^TestBatch1SourceOperationMappingCohortCheckValidatesMatricesAndDeclaredArtifactLinks$' \
  -count=1
```

failed as intended:

```text
cohort CLI output = "connectorgen source-operation-mapping-cohort:
10 connector(s), 4341 source operation(s), 0 finding(s)\n",
want whole-cohort evidence "matrix validation:"
```

That failure demonstrates the original production check neither opened the
matrices nor checked their declared artifact links.

## Green

```text
GOCACHE=/private/tmp/gocache-4293-green.o8TgvF \
  go test ./cmd/connectorgen \
  -run '^(TestBatch1SourceOperationMappingCohortCheckValidatesMatricesAndDeclaredArtifactLinks|TestSourceOperationMappingCohortFullCheckRejectsRealMatrixEvidenceAndArtifactDefects)$' \
  -count=1 -v
```

passed in 91.510s. The real cohort acceptance case reported:

- 4,341 primary source operations;
- 4,343 source rows, including the retained supplemental rows;
- 30,401 source-lane cells;
- 6,790 deferred cells;
- 917 explicit artifact-link records and 930 source-lane link references;
- 5,886 typed unlinked-deferred projection deficits;
- zero executable declarations.

The copied immutable cohort fixture rejects each independently: missing matrix,
malformed matrix schema, citation drift, a hidden source row, missing artifact
target, orphan artifact source identity, ambiguous source-id dialect, and
artifact path traversal.

## Refactor boundary

The report uses derived values from loaded source locks/matrices. Exact counts
above are acceptance evidence only, not production connector branches. The
implementation does not require an artifact for a mapped-unproven or
missing-foundation cell; it records that absence as a typed projection deficit.
