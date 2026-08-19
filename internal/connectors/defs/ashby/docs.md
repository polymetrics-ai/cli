# Ashby Connector

## Overview

Ashby is an applicant-tracking connector generated from the public Ashby ReadMe OpenAPI reference (https://developers.ashbyhq.com/reference). The parity ledger was reviewed on 2026-08-01.

Coverage summary:

- REST operations in source: 185
- OpenAPI webhook events in source: 27
- Implemented ETL streams: 71
- Implemented bounded direct reads/search/file metadata operations: 9
- Implemented reverse-ETL write actions: 98
- Reverse-ETL CLI commands with scalar flags: 89; partial nested-object flag surfaces: 9
- Blocked/non-executable ledger rows: 34

## Auth setup

Authentication uses Ashby's documented HTTP Basic API-key flow: the API key is the username and the password is blank. Provide keys via environment variables or stdin only; never paste secrets into prompts, docs, commits, or issue comments.

## Streams notes

Ashby list and info reads are fixed POST endpoints with documented body fields only. The native connector owns Ashby's cursor-in-body pagination and applies page-size, max-pages, and repeated-cursor bounds. Streams are full-refresh only until `ashby-sync-token-checkpoint-foundation` supplies an Ashby-owned persisted opaque-token state seam; timestamp fields are not used as lossy substitutes. Runtime help replaces provider incremental descriptions with full-refresh-only blocker text for every documented sync-token request. Repeatable array stream flags are withheld until `connector-stream-repeatable-array-foundation` preserves every supplied value.

## Write actions & risks

Reverse ETL writes are typed action names with recursively closed modeled JSON schemas and the normal plan → preview → explicit approval → execute gate. Explicitly documented map-valued fields retain their map schemas; all other modeled objects reject undeclared fields. No command exposes a raw HTTP method, raw path, arbitrary request body, raw query, shell, file, SQL, or passthrough escape hatch. The public Ashby OpenAPI did not document an Idempotency-Key or equivalent idempotency header for these actions, so no provider idempotency key is claimed.

## Known limits

Blocked rows are still documented in `api_surface.json`: inbound assessment-partner APIs and webhook events are not pull-executable by a CLI connector, and `file.createFileUploadHandle` remains blocked until a reviewed bounded binary/file workflow can safely return and consume presigned upload handles. `referralForm.info` is blocked pending `ashby-referral-form-info-side-effect-foundation` because it conditionally creates a default form. `applicationForm.submit` is blocked pending `ashby-application-form-typed-multipart-foundation` because the documented request requires multipart form data and typed file parts. Opaque incremental state is blocked pending `ashby-sync-token-checkpoint-foundation`, and repeatable array stream-command variants are blocked pending `connector-stream-repeatable-array-foundation`. `hiringTeamRole.list` defaults to `namesOnly=true`; the `namesOnly=false` object-result variant is blocked pending variant-schema foundation `ashby_hiring_team_role_list_names_only_false`. Fixture replay covers every implemented stream with synthetic values only; no live Ashby credentials or provider calls were used.
