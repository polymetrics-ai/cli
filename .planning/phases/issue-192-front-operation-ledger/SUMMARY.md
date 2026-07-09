# Summary: Front Operation Ledger (#192)

Status: plan checkpoint in progress.

## Completed

- Created #192 branch `feat/192-front-operation-ledger` from parent branch `feat/188-front-cli-parity`.
- Generated GSD plan prompt and recorded programming-loop adapter fallback.
- Captured public ReadMe registry sources for Front Core and Channel OpenAPI definitions.
- Planned red validation for current 10-row non-ledger `api_surface.json`.

## Key source facts

- Official public ReadMe OpenAPI registry currently exposes 255 REST operations.
- Method split: `GET=123`, `POST=76`, `PATCH=26`, `PUT=3`, `DELETE=27`.
- Parent issue #188 records a 342-operation baseline, but the current public registry does not reproduce that count.
- `llms.txt` has 346 API Reference links; 91 are category/guide/plugin-SDK/data-model pages outside REST method/path execution.

## Next steps

1. Commit and push the plan checkpoint before production edits.
2. Run the red ledger validation.
3. Rewrite `internal/connectors/defs/front/api_surface.json` in `operation_ledger_version: 1` mode for the 255 captured REST operations.
4. Run focused validation and update this summary.
