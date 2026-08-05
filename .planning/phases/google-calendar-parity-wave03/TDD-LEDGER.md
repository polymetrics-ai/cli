# Google Calendar parity resume — TDD ledger

## Red evidence

| Slice | Command | Result | Expected failure |
| --- | --- | --- | --- |
| Current-main validation | `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar` | failed | `freebusy query` flags `--time-min`, `--time-max`, and `--calendar` map to required body paths but are not themselves required. |

The failure was reproduced after rebasing PR #3554 onto `origin/main`; the preserved green claims are historical only.

## Green evidence

| Slice | Minimum production change | Focused proof |
| --- | --- | --- |
| Direct read | Made `--calendar`, `--time-min`, and `--time-max` required. | `connectorgen validate` and `surface-sync --check` pass; built help marks all three required. |
| Fixture auth | Focused conformance initially failed when the fixture runtime omitted the preserved hook's sentinel and attempted an OAuth token exchange using synthetic values. | Restore `ProjectDir=__polymetrics_conformance_fixture__` in shared fixture configuration, then rerun all conformance. |
| Mutation truthfulness (superseded) | The prior contract interpretation removed 26 unavailable write actions and replaced each `covered_by.write` with a blocked operation-ledger classifier. | Superseded by the 2026-08-05 section 3b correction: typed `writes.json` actions execute today and must be authored where they fit. |
| Mutation execution | Add 26 typed reverse-ETL actions, fixtures, coverage, CLI commands, and field citations for every documented Google Calendar mutation. | `connectorgen validate`, fixture-backed conformance, generated CLI help, and the reverse-ETL plan surface must prove every action is executable. |
| Runtime activation | Restored the recovered engine-bundle promotion so the legacy native connector cannot shadow the CLI surface. | `bundleregistry` regression test plus built `pm google-calendar` and `freebusy query --help` pass. |
| POST direct-read classification | Restricted mutation-capability detection to operations covered by a write action, so the POST-backed `freeBusy.query` remains a direct read while connector write capability stays false. | Focused connectorgen and conformance regressions cover POST direct reads and covered write operations. |
| Citation research | Added provider field evidence without inventing a shared machine citation schema that is absent from current main. | `REQUEST-FIELD-RESEARCH.md` records 148/148 declared field uses and 38/38 operation-level sources. |
| Generated parity | Regenerated the affected CLI manual/catalog/website outputs and root CLI golden transcripts. | Final docs/website/CLI gates remain in the verification checklist. |
| Review remediation | Corrected stale release/write metadata, removed the implicit event cutoff, restored legacy list-stream projections, and added settings cursor pagination. | `go test ./internal/connectors/hooks/google-calendar -count=1` passes with connector-owned regressions for metadata, initial event queries, projection boundaries, and two-page settings reads. |
