# Twenty CRM connector

## Overview

Twenty CRM is exposed as a declarative connector for the 28 documented REST object collections:
companies, people, opportunities, notes, tasks, messages, calendar objects, workflow objects,
workspace members, and supporting association objects. The connector supports ETL list streams and
reverse ETL create, update, batch-create, and delete actions for those objects.

## Auth setup

Twenty uses bearer authentication with an `api_key` secret and an optional `base_url` config value.
Provide the secret from an environment variable or stdin; do not paste it into prompts, commit it,
or print it in logs. Inspecting the connector manual or command surface does not read credentials.

## Streams notes

Each stream maps to `GET /rest/<TwentyObject>` and reads records from the Twenty `data.<object>`
envelope with cursor pagination using `pageInfo.endCursor` and `pageInfo.hasNextPage`. The CLI
surface exposes those streams through commands such as `pm twenty companies list --json`.
Get-by-id endpoints (`GET /rest/<TwentyObject>/{id}`) are documented in the command surface as
planned direct-read commands because generic JSON direct-read output policies are outside S6.

## Write actions & risks

For every Twenty object, reverse ETL declares:

- `create_<object>` for `POST /rest/<TwentyObject>`;
- `update_<object>` for `PATCH /rest/<TwentyObject>/{id}`;
- `batch_<object>` for `POST /rest/batch/<TwentyObject>`;
- `delete_<object>` for `DELETE /rest/<TwentyObject>/{id}`.

Reverse ETL remains plan, preview, approval, then execute. Delete actions are destructive and require
typed `--confirm destructive` approval. Batch actions send a validated top-level `records` array
through reverse ETL records; the S6 CLI surface intentionally does not expose a generic JSON/raw
write flag for batch payloads.

## Command Surface

`pm help twenty`, `pm twenty`, and `pm twenty --help` render the Twenty connector manual and command
surface without credentials. Implemented list commands read ETL streams. Implemented create, update,
and delete commands plan reverse ETL writes from commandrunner-safe scalar flags. Batch commands are
marked partial until scalar CLI coercion for `records: []object` is approved.

## Known limits

- Generic JSON direct-read execution for get-by-id commands is not implemented in S6.
- Batch write execution through scalar CLI flags is deferred; use reverse ETL records with a
  top-level `records` array instead of a raw JSON flag.
- No live Twenty credentials are required for connector inspection, help rendering, docs generation,
  validation, or fixture conformance.
- Destructive delete actions are declarative only until a user follows reverse ETL plan, preview,
  approval, and execute.
