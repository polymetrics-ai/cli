# TDD ledger — Issue #4411 Asana Track C execution proof

## Red

- The existing copied-artifact validator deliberately removes `sync_transport.json`, removes a direct-read backlink, and changes a `mapped_unproven` ETL cell to implemented. Each case must fail before the genuine definitions are accepted.
- The new proof test is intentionally limited to the existing embedded definitions. Its initial run exposed two required boundaries: `tasks create` has no CLI `source_operation` field (it must bind via typed write/API surface), and `tasks` ETL correctly rejects a credential `base_url` override. Those are retained as explicit green assertions, not fixed by changing definitions or runtime.

## Green criteria

- 249 source identities and every seven-lane matrix cell still reconcile to the Track B artifact projection.
- Direct read, direct write, and binary upload each reach the embedded registry/CLI credential boundary with valid typed flags and zero provider I/O.
- The `tasks` stream is the only ETL witness added here: it is source-backed, implemented, materializes one local fixture record through DuckDB, and reads it back from its named connection.
- Existing Asana tests separately retain exact source-bound direct read, every-action direct write, every-action reverse ETL, attachment upload, and event-token sync evidence.
- The 52 mapped-unproven ETL candidates remain a visible gap; N/A or missing cells have no executable claim.

## Executed result

- `TestAsanaTrackCImplementedCommandLanesReachEmbeddedRegistryAndCredentialBoundary` passes for direct read, direct write, and binary upload. Each uses valid typed flags, the production embed/normal registry/CLI, an initialized local project, no credential, and a no-provider-I/O transport spy.
- `TestAsanaTrackCETLThroughDuckDBUsesSourceBoundLocalFixture` passes through the existing `tasks` stream to one owner-scoped local warehouse row and an acknowledged one-record/one-batch checkpoint. Its local response is injected only at the declared `https://app.asana.com/api/1.0` origin.
- `TestAsanaTrackCETLRejectsCredentialBaseURLOverride` passes by refusing the invalid origin before it can be sent.

## Refactor boundary

- Keep proof code connector-local and fixture-only. No new shared helper, definition edit, source import, or runtime behavior is permitted.
