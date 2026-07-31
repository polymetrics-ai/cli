Part of #277 (S4). Write scope: `writes.json` (non-destructive rows). Deps: #278, #279.

- 84 `reverse_etl` actions: create `POST /rest/{objects}`, update `PATCH /rest/{objects}/{id}`, batch `POST /rest/batch/{objects}` (≤60) with `record_schema`, `path_fields`, `body_type`, `risk: normal`. Plan → preview → approval → execute only.

Acceptance: every `reverse_etl` row maps to an `api_surface` write via `covered_by`.
Refs #277