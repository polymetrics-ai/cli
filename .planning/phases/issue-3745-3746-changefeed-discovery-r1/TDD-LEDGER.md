# TDD LEDGER — issues #3745 and #3746

| ID | Contract | RED evidence to execute | GREEN evidence | Refactor / verification |
| --- | --- | --- | --- | --- |
| R1 | `--capability cdc` never lists a connector merely because it implements `CDCReader` | Focused catalog test fails on current main because PostgreSQL is listed while `metadata.json` says `cdc: false` | Catalog checks descriptor status plus matching executor, so PostgreSQL is absent | Execute catalog JSON command; assert no PostgreSQL row |
| R2 | A catalogued changefeed is executable | Focused proof fails while PostgreSQL is catalogued and `ReadCDC` returns `ErrUnsupportedOperation` | PostgreSQL is not catalogued; legacy stub remains an honest unsupported implementation | Assert the status/executor mismatch cannot project `cdc: true` |
| R3 | Changefeed taxonomy is closed and unsupported is explicit | Descriptor/inspect test fails because PostgreSQL has no changefeed explanation on current main | Dedicated descriptor yields `unsupported`, source/reason, and mechanism in inspect JSON without inventing an executor, checkpoint, or delivery promise | Absent/zero descriptor cannot become implemented |
| R4 | An implemented declaration needs an executor with the same contract | Focused mismatch case fails before derivation exists | Matching provider contract is required by the public projection | Table-driven absent/mismatched/matching executor cases |

## Executed RED evidence

Command:

```text
go test ./internal/cli -run 'Test(CatalogedCDCRequiresDeclaredExecutableChangefeed|InspectPostgresReportsUnsupportedChangefeed)$' -count=1
```

Result: failed as required before production implementation.

- `TestCatalogedCDCRequiresDeclaredExecutableChangefeed` loaded PostgreSQL metadata with
  `cdc=false`, executed the real catalog command and found `postgres`, then executed its real
  `ReadCDC` stub. The failure records that the stub returned the documented
  `ErrUnsupportedOperation` behind the gated `pglogrepl` implementation plan.
- `TestInspectPostgresReportsUnsupportedChangefeed` failed because the real JSON inspect envelope
  did not contain a `changefeed` descriptor at all.

No test is weakened or deleted to make the result pass.

## GREEN evidence

The descriptor contract is exercised by:

```text
go test ./internal/connectors -run 'Test(HasImplementedChangefeedRequiresMatchingExecutor|ChangefeedDescriptorRejectsUnsupportedExecutionClaims|ChangefeedDescriptorRejectsInvalidSourceURL|ChangefeedDescriptorUsesClosedStatusAndMechanism|DefinitionOfDerivesCDCFromMatchingChangefeedExecutor|RegistryDoesNotTrustLegacyCDCMetadataOrReader|DefinitionJSONShape)$' -count=1
go test ./internal/connectors/engine -run 'TestBundleLoad(OptionalFilesAbsent|ParsesUnsupportedChangefeed|RejectsUnsupportedChangefeedWithExecutor)$' -count=1
go test ./internal/connectors/native/postgres -run 'Test(ConnectorSatisfiesCoreInterfaces|CDCUnsupportedStub)$' -count=1
go test ./internal/cli -run 'Test(CatalogedCDCRequiresDeclaredExecutableChangefeed|InspectPostgresReportsUnsupportedChangefeed)$' -count=1
```

All four commands passed. The first suite proves that legacy metadata and a bare
`CDCReader` cannot advertise CDC, while an `implemented` descriptor with a matching
`ChangefeedExecutor` can. The loader suite proves `changefeed.json` is optional and
rejects an unsupported descriptor that claims an executor. The PostgreSQL and CLI
suites retain the unsupported-stub regression as the visible behavior proof.

## Review-hardening RED evidence

During the required inline code review, this focused test was added before the hardening change:

```text
go test ./internal/connectors -run 'Test(ChangefeedDescriptorRejectsUnsupportedExecutionClaims|ChangefeedDescriptorRejectsInvalidSourceURL)$' -count=1
```

It failed because an `unsupported` descriptor could still carry a checkpoint or delivery claim,
and because a non-URL artifact string passed validation. The descriptor now rejects all three
unsupported execution claims (executor, checkpoint, delivery) and requires an absolute `http` or
`https` source artifact URL. The same command passes after the fix.
