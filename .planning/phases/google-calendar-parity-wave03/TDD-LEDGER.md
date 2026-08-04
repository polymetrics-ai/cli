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
| Mutation truthfulness | Removed 26 unavailable write actions and replaced each `covered_by.write` with a blocked operation-ledger classifier. | Bundle validation passes; focused conformance and runtime-preflight gates remain in the verification checklist. |
| Runtime activation | Restored the recovered engine-bundle promotion so the legacy native connector cannot shadow the CLI surface. | `bundleregistry` regression test plus built `pm google-calendar` and `freebusy query --help` pass. |
| Citation research | Added provider field evidence without inventing a shared machine citation schema that is absent from current main. | `REQUEST-FIELD-RESEARCH.md` records 25/25 declared field uses and 38/38 operation-level sources. |
| Generated parity | Regenerated the affected CLI manual/catalog/website outputs and root CLI golden transcripts. | Final docs/website/CLI gates remain in the verification checklist. |
