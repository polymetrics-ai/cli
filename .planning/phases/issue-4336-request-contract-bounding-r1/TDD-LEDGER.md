# TDD ledger — request-contract execution envelopes

| ID | Guarantee | Required red assertion | Green proof |
| --- | --- | --- | --- |
| B1 | Valid missing common bounds do not abort source import | Gong-shaped `workspaceId` fails with `unbounded request schema string has no maxLength`. | The operation imports without a schema gap. |
| B2 | Source schema and PM policy remain distinct | New descriptor assertion cannot find an envelope and observes the old import abort. | Schema remains exactly source-owned; separate envelope carries policy version, origin, unit, default, hard, and effective cap. |
| B3 | Generation fails closed without a finite envelope | Sabotaged/removal path passes today because no envelope exists. | Descriptor/projection validation rejects missing, zero, inconsistent, or over-ceiling envelopes for executable common inputs. |
| B4 | Runtime enforces the exact encoded boundary before I/O | PM-labelled cap assertion fails under the old generic error contract. | Exact cap reaches local HTTP once; cap+1 and UTF-8 expansion reject with zero additional hits. |
| B5 | Numeric parsing is exact and byte-bounded | An integer beyond `int64` and a high-precision decimal are rejected or narrowed by `ParseInt64`/`ParseFloat64`. | `big.Int`/`big.Rat` validation accepts finite lexemes, sends identical bytes, and rejects invalid syntax before I/O. |
| B6 | Semantic ambiguity remains visible | A broad bypass would incorrectly erase dynamic/composition/untyped gaps. | Focused importer cases remain source-traced and merge-blocked. |
| B7 | Existing command limits do not silently drift | Projection snapshot detects any change to current 8 KiB/256/1 MiB compatibility behavior. | Existing bounded fixture remains semantically identical aside from additive descriptor provenance. |

## Red evidence

Before production edits, the Gong-shaped v3 importer test failed at the current
strict `maxLength` preflight, the runtime provenance test exposed the generic
unattributed error, and the exact numeric test exposed int64 narrowing:

```sh
GOFLAGS=-p=3 go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3RepresentsGongWorkspaceQueryWithPMExecutionEnvelope$' -count=1
GOFLAGS=-p=3 go test -timeout 20m ./internal/connectors/engine -run '^(TestOperationParametersReportPMExecutionCapBeforeIO|TestOperationParametersPreserveExactFiniteNumericLexemes)$' -count=1
```

Full outputs are retained in `traces/red-source-import.txt` and
`traces/red-runtime.txt`.

## Green evidence

Focused green after implementation:

```sh
GOFLAGS=-p=3 go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestSourceImportVersion3RepresentsGongWorkspaceQueryWithPMExecutionEnvelope|TestSourceParameterExecutionEnvelopeUsesTighterProviderDerivedByteCap|TestSourceRequestSchemaDispositionSeparatesPolicyBoundsFromMalformedInput|TestSourceImportVersion3KeepsUnboundedHeaderAsMergeBlockingGap|TestSourceImportVersion3RepresentsCommonBodyBoundsWithSeparateEnvelope)$' -count=1
GOFLAGS=-p=3 go test -timeout 20m ./internal/connectors ./internal/connectors/engine ./internal/cli \
  -run '^(TestCommandSurfaceSectionNamesPMEncodedBytePolicy|TestCommandSurfaceProjectsOperationParameterByteCaps|TestConnectorInspectJSONIncludesRequestExecutionLimits|TestOperationParametersReportPMExecutionCapBeforeIO|TestOperationParametersPreserveExactFiniteNumericLexemes)$' -count=1
```

These prove the exact Gong-shaped import, source/envelope separation, typed
schema dispositions, header quarantine, body envelopes, help/inspection
provenance, pre-I/O caps, and exact numeric request lexemes.

Code review added three further red/green controls:

- Altering an otherwise well-shaped body envelope's effective byte limit was
  initially accepted; complete canonical-envelope comparison now rejects it.
- `json.Valid` initially rejected numeric spellings accepted by the prior
  `strconv` contract (`+01`, `+.5`, `01.`, and `0x1p+2`); exact `big.Int` /
  `big.Rat` parsing now preserves that compatibility while continuing to reject
  fraction syntax that `ParseFloat` never accepted.
- A numeric header with provider minimum and maximum still had no finite
  textual byte bound but initially escaped the unbounded-header gap; it is now
  merge-blocked pending the shared header census. Booleans derive the exact
  five-byte `false` ceiling.

The final full `cmd/connectorgen` package, full engine package, lint, vet,
builds, and repository generator/boundary/release gates passed after these
review fixes.

## Deliberate sabotage evidence

After green, the runtime's `len(encoded) > capBytes` condition was temporarily
disabled. `TestOperationParametersReportPMExecutionCapBeforeIO` failed because
the cap+1 request returned no error instead of the PM policy error. The guard
was restored immediately; the same test passed and a path-scoped `git diff
--exit-code` proved no sabotage remained. Full output is retained in
`traces/sabotage-runtime-cap.txt`.
