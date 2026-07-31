# Recurly wave05 r1 inventory

Snapshot preserved outside the repo at `/tmp/recurly-wave05-r1-snapshot-20260801023416` before final gate reruns.

Current working tree scope:

- Recurly bundle files under `internal/connectors/defs/recurly/**`
- Generated Recurly docs under `docs/connectors/recurly/**`
- Generated website connector catalog surfaces: `website/data/connectors.generated.json`, `website/lib/connectors.catalog.data.generated.json`, `website/lib/connectors.catalog.generated.ts`
- CLI golden transcript update for changed dynamic connector/manual output: `internal/cli/testdata/golden_transcripts.json`
- GSD artifacts under `.planning/phases/issue-3183-recurly-parity-wave05-r1/**`

No shared runtime files are modified.

Counts:

- Official Recurly v2021-02-25 endpoints: 197 total (GET 97, POST 42, PUT 35, DELETE 23)
- ETL streams: 93
- Reverse ETL write actions: 96
- Operation metadata entries: 8 (5 implemented JSON direct reads + 3 planned binary/export metadata entries)
- API surface coverage: 194 covered; 3 `binary_read:blocked`; 0 exclusions
- Fixtures: 93 stream fixture directories and 96 write fixtures
- CLI direct commands: `invoices preview`, `subscriptions preview renewal`, `subscriptions preview change`, `purchases preview`, `gift cards preview`; planned binary metadata commands: `invoice pdf get`, `export dates get`, `export files get`

Schema/CLI correction applied before final gates:

- Direct-read CLI flags now map only concrete required body leaves, not whole object/array bodies.
- `gift cards preview --unit-amount` uses integer coercion so the JSON body validates against the Recurly numeric `unit_amount` schema without shared-runtime changes.
