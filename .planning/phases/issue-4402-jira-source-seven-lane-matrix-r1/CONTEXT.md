# Jira Track A context

## Locked source authority

- Lock: `internal/connectors/defs/jira/sources/jira-operation-source-lock.json`.
- Denominator: 617 retained REST rows; lock schema 2.
- Source document: `https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json`.
- Immutable source binding: SHA-256 `e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf`, 2,456,011 bytes.

## Deliberate non-inferences

- A GET verb is a read-verb candidate, not direct-read runtime proof.
- A non-GET verb is a mutation-verb candidate, not proof of a safe provider write.
- `maxResults` identifies the requested 95-row source paging cohort; it neither proves a continuation protocol nor admits ETL execution.
- Existing `streams.json` rows are artifact backlinks only. The three legacy rows do not cover the remaining 92 paging candidates or certify any stream.
- Source media is used only where the exact selected media/text rule is retained. No connector-wide media inference is introduced.
- Dynamic webhook registration is source-visible, but no inbound receiver is present. It remains a typed `missing_foundation` cell.

## Artifact boundary

The matrix is connector-local mapping evidence. It changes no current CLI projection, stream, write, sync transport, runtime registry, generated artifact, or shared foundation. The local Go test is the no-hidden-row and no-promotion guard for its JSON.
