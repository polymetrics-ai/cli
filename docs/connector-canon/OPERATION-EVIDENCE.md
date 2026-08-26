# Operation evidence

`connectorgen operation-evidence` projects one deterministic evidence row for
every operation named by a connector-owned provider source lock. The checked-in
artifact is [`internal/connectors/operation-evidence.json`](../../internal/connectors/operation-evidence.json);
the executable aggregate contract is
[`internal/connectors/operation-evidence-fixed-100.json`](../../internal/connectors/operation-evidence-fixed-100.json).

Each row records the provider trace, canonical mapping, declaration-owned
runtime target, CLI command, generated website command, fixture, matching
conformance check, and every capability classification (`etl`, `reverse_etl`,
`direct_read`, `direct_write`, `binary_download`, `binary_upload`). A missing
surface is a named gap. It is never silently omitted or represented as `N/A`.

The only non-gap absence is a source lock whose provider evidence explicitly
declares `skipped` or `dynamic`; the artifact carries that reason and evidence
verbatim. An explicit declaration-only `source_reference` remains a row: its
canonical citation and operation identity are visible, while the named
`source_contract_unavailable` gap prevents its six-lane classification from
becoming enabled merely because a similarly shaped declaration exists. The
projector reads source-lock v2 and v3 inputs but does not own or alter their
parser or schema.

Run the generator after a reviewed source-lock or connector-surface change:

```bash
go run ./cmd/connectorgen operation-evidence --write-fixed-100
```

`make verify` runs `connectorgen-operation-evidence`, equivalent to:

```bash
go run ./cmd/connectorgen operation-evidence --check
```

The check rejects artifact drift and validates each source-locked expectation in
the fixed cohort. The cohort is a deterministic, capability-stratified sample
of 100 complete provider operations, so a regression in any selected operation
fails with its source ID rather than being hidden by an aggregate count.
