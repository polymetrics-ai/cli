# Official Surface Capture: Gorgias
## Source
- LLM index: https://developers.gorgias.com/llms.txt
- Operation pages: linked `https://developers.gorgias.com/reference/*.md` pages with embedded `# OpenAPI definition` JSON blocks.
- Capture command: public unauthenticated HTTP GET with a browser-like User-Agent; no credentials used.

## Capture summary
- Parsed operation pages: 114 unique method/path/operationId rows.
- Method split from ReadMe markdown blocks: {'DELETE': 18, 'GET': 46, 'POST': 23, 'PUT': 27}.
- Parent issue baseline records 114 official operations with a different method taxonomy: GET 59, PATCH 22, DELETE 16, POST 17. #200 owns reconciliation of ReadMe `PUT` update verbs vs the parent taxonomy and direct-read/binary candidates.
- Official paths include `/api/...`; the current Gorgias bundle asks users for a `base_url` ending in `/api`, so connector-relative paths strip the `/api` prefix (for example `/api/tickets` -> `/tickets`).

## Resource families observed
- `account`: 4
- `custom-fields`: 5
- `customers`: 12
- `events`: 2
- `integrations`: 5
- `jobs`: 5
- `macros`: 7
- `messages`: 1
- `metric-cards`: 3
- `phone`: 7
- `reporting`: 1
- `rules`: 6
- `satisfaction-surveys`: 4
- `search`: 1
- `stats`: 2
- `tags`: 7
- `teams`: 5
- `tickets`: 18
- `upload`: 1
- `users`: 5
- `views`: 7
- `widgets`: 5
- `{file_type}`: 1

## #197 classification boundary
- Implemented CLI metadata maps only current stream-backed commands to existing streams and existing `api_surface.json` rows.
- Planned write/direct-read/binary/admin commands are listed without executable targets; later lanes (#199-#203) own implementation and full operation-ledger classification.
- No raw API, direct write, generic HTTP write, generic shell, or generic SQL write tool is exposed.
