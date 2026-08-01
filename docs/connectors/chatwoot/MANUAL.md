# pm connectors inspect chatwoot

```text
NAME
  pm connectors inspect chatwoot - Chatwoot connector manual

SYNOPSIS
  pm connectors inspect chatwoot
  pm connectors inspect chatwoot --json

DESCRIPTION
  Reads fixture-backed Chatwoot account streams and safely plans named reverse ETL actions. The operation ledger covers all 145 official Chatwoot OpenAPI operations.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

SECURITY
  No secret values belong in chat, shell history, docs, fixtures, or JSON output. Reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive/admin actions require --confirm destructive.

IMPLEMENTED STREAMS
  conversations, contacts, inboxes, agents, teams, labels, messages

REVERSE ETL ACTIONS
  73 named actions are declared in writes.json; destructive/admin count: 49. Inspect structured connector JSON for the full action list.

KNOWN LIMITS
  Public client API, write-query gaps, and non-fixture-backed read/report/changefeed operations remain planned/blocked in api_surface.json with evidence. No live certification is claimed.

AGENT WORKFLOW
  Inspect first with --json. Never request credentials in chat. Use plan, preview, approval token, and --confirm destructive for destructive/admin write execution.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
```
