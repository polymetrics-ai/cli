---
name: pm-chatwoot
description: Safe Chatwoot connector usage for pm.
---

# Chatwoot connector

Use `pm connectors inspect chatwoot --json` before selecting streams or write actions.

## Agent Rules

- Streams: conversations, contacts, inboxes, agents, teams, labels, messages.
- Writes: 73 named reverse-ETL actions in `internal/connectors/defs/chatwoot/writes.json`.
- Safety: plan -> preview -> explicit approval -> execute; destructive/admin actions require typed `--confirm destructive`.
- Never ask for or print `api_access_token`; load it from environment or stdin.
- Public client API and non-fixture-backed direct/report/changefeed operations remain planned/blocked in `api_surface.json`; do not route through a raw HTTP escape hatch.
