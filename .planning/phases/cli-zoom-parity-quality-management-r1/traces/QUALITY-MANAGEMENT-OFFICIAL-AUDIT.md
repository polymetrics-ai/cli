# Zoom Quality Management official-artifact audit

- URL: `https://developers.zoom.us/docs/api/quality-management.md`
- Retrieved: `2026-08-08T09:12:11Z`
- HTTP: `200`
- Exact bytes: `40,987`
- OpenAPI / API: `3.1.1` / `2`
- Server: `https://api.zoom.us/v2`

The artifact enumerates exactly six REST operations:

1. `GET /qm/automated_evaluations` — List automated evaluations
2. `GET /qm/evaluation` — List evaluations
3. `GET /qm/evaluation/{evaluationId}` — View evaluation detail
4. `GET /qm/interactions` — List interactions
5. `POST /qm/interactions` — Add an interaction
6. `GET /qm/interactions/{interactionId}` — View interaction detail

All six match `provider_module=quality-management` entries in Zoom's derived API surface. The
derived ledger delta is zero. The live GET entries contain no request-parameter section; listed
pagination/date fields are response schema values, not CLI inputs.
