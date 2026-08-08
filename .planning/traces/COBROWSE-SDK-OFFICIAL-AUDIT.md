# Zoom Cobrowse SDK — official artifact audit

- URL: `https://developers.zoom.us/docs/api/cobrowse-sdk.md`
- Retrieved: `2026-08-08T10:10:47Z`
- HTTP status: `200`
- Byte count: `11,697`
- Artifact header: Zoom API `2`, OpenAPI `3.1.1`, `https://api.zoom.us/v2`

| Method | Origin-relative path | Provider operation | Declared request contract |
| --- | --- | --- | --- |
| GET | `/cobrowsesdk/live_sessions` | List live sessions | Optional monthly `from` and `to` query dates are stated in operation prose. |
| GET | `/cobrowsesdk/past_sessions` | List past sessions | Optional monthly `from` and `to` query dates are stated in operation prose. |
| GET | `/cobrowsesdk/sessions/{sessionId}` | Get session details | Required `{sessionId}` path identifier. |
| GET | `/cobrowsesdk/sessions/{sessionId}/users` | List session users | Required `{sessionId}` path identifier. |

The documentation lists `page_size` and `next_page_token` in responses only. They are not CLI input
flags. The four inherited `provider_module=cobrowse-sdk` ledger rows match the live artifact exactly
(delta `0`).
