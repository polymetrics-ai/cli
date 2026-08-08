# Zoom Healthcare official-artifact audit

Retrieved before implementation from `https://developers.zoom.us/docs/api/healthcare.md`.

- UTC retrieval time: `2026-08-08T08:22:55Z`
- HTTP status: `200`
- Download byte count: `13,783`
- Artifact declaration: OpenAPI `3.1.1`, API version `2`, server `https://api.zoom.us/v2`

The live `## Operations` section declares exactly three routes:

| Method | Path | Operation title | Ledger disposition before this slice |
| --- | --- | --- | --- |
| GET | `/clinical_notes/notes` | List clinical notes | blocked `direct_read` |
| GET | `/clinical_notes/notes/{noteId}` | Get a Clinical Note | blocked `direct_read` |
| PATCH | `/clinical_notes/notes/{noteId}` | Update a Clinical Note | blocked `sensitive_reverse_etl` |

All three match `internal/connectors/defs/zoom/api_surface.json`'s Healthcare rows exactly;
delta `0`. The PATCH request body has exactly one required property, `is_note_completed` (boolean),
and its documented successful response is `204 No Content`.

Request-input decision: `note_owner_user_id` and `meeting_id` are explicitly named as values the
caller can provide in the List operation prose. `from`, `to`, `page_size`, and
`next_page_token` are displayed below `Responses → Status: 200`; they are not documented request
parameters and must not become flags.

No source response example content was retained in this trace because it includes clinical data.
