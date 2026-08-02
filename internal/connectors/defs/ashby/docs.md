# Ashby Connector

## Overview

Ashby is an applicant-tracking connector generated from the public Ashby ReadMe OpenAPI reference (https://developers.ashbyhq.com/reference/candidateaddtag). The parity ledger was reviewed on 2026-08-01.

Coverage summary:

- REST operations in source: 185
- OpenAPI webhook events in source: 27
- Implemented ETL streams: 72
- Implemented bounded direct reads/search/file metadata operations: 7
- Implemented reverse-ETL write actions: 101
- Reverse-ETL CLI commands with scalar flags: 91; partial nested-object flag surfaces: 10
- Blocked/non-executable ledger rows: 32

## Auth setup

Authentication uses Ashby's documented HTTP Basic API-key flow: the API key is the username and the password is blank. Provide keys via environment variables or stdin only; never paste secrets into prompts, docs, commits, or issue comments.

## Streams notes

Ashby list and info reads are fixed POST endpoints with documented body fields only. The native connector owns Ashby's cursor-in-body pagination, applies page-size and max-pages bounds, and supports client-side incremental filtering when generated stream metadata explicitly declares an incremental cursor.

## Write actions & risks

Reverse ETL writes are typed action names with closed top-level JSON schemas and the normal plan → preview → explicit approval → execute gate. No command exposes a raw HTTP method, raw path, arbitrary request body, raw query, shell, file, SQL, or passthrough escape hatch. The public Ashby OpenAPI did not document an Idempotency-Key or equivalent idempotency header for these actions, so no provider idempotency key is claimed.

## Known limits

Blocked rows are still documented in `api_surface.json`: inbound assessment-partner APIs and webhook events are not pull-executable by a CLI connector, and `file.createFileUploadHandle` remains blocked until a reviewed bounded binary/file workflow can safely return and consume presigned upload handles. `hiringTeamRole.list` defaults to `namesOnly=true`; the `namesOnly=false` object-result variant is blocked pending variant-schema foundation `ashby_hiring_team_role_list_names_only_false`. Fixture replay covers every implemented stream with synthetic values only; no live Ashby credentials or provider calls were used.
