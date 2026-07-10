# Summary: Gorgias Stream Runner

Parent issue: #196  
Sub-issue: #199  
Branch: `feat/199-gorgias-stream-runner`

## Result

- Expanded Gorgias ETL stream coverage from 4 to 24 stream-backed GET endpoints.
- Added 15 top-level list/search streams for account settings, custom fields, events, integrations, jobs, macros, metric cards, rules, tags, teams, users, views, voice calls, voice call events, and widgets.
- Added 5 single-level fan-out streams for customer custom fields, ticket custom fields, ticket tags, ticket messages, and view items.
- Added minimal schemas and non-secret fixture pages for every new stream.
- Updated `api_surface.json` so the implemented stream endpoints use `covered_by.stream`; remaining direct-read, binary, advanced-query, write/admin/destructive/product-scope rows stay blocked metadata.
- Updated CLI surface metadata for stream-backed list/search commands that now map to implemented streams.
- Updated connector docs and read-risk wording to describe the expanded read-only scope without claiming write/direct-read/binary parity.

## Safety posture

- No write actions were declared.
- No direct-read detail commands were made executable.
- No binary/file or POST query endpoints were implemented.
- No credentialed Gorgias checks were run.
- Fan-out remains single-level and typed; no raw API passthrough was introduced.
