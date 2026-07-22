# Research: Pitfalls

**Generated via:** official GSD Core Pi adapter command path.

## Connector Pitfalls

- Treating documented connector APIs as REST-only.
- Duplicating operation work because docs pages, generated references, aliases, and product guides describe the same upstream operation.
- Mapping binary/download/export operations into durable stream abstractions without product fit.
- Mapping direct-read operations into ETL when they are point lookups or operational reads.
- Exposing destructive/admin/elevated-scope writes instead of typed exclusions or human gates.
- Treating missing live credentials as certification failure instead of `uncertified`.

## Planning Pitfalls

- Using stale legacy `.planning/` counts for fanout.
- Regenerating `.planning/phases/**` when the user requested non-phase refresh only.
- Shipping CLI-visible work without updating runtime help, bare namespace command behavior such as `pm connectors`, `docs/cli/**`, website docs, generated help/manual artifacts, and tests.
- Claiming upstream Pi support when the official GSD docs do not list Pi as a runtime.
- Letting agents use stale Claude-local slash command assumptions instead of the repo-local Pi adapter.

## Delivery Pitfalls

- Posting redundant Claude manual review commands after every push.
- Treating skipped/rate-limited automated review as approval.
- Adding dependencies without human approval.
- Running credentialed connector checks or reverse ETL execution during planning-only work.

---
*Pitfall research refreshed: 2026-07-08.*
