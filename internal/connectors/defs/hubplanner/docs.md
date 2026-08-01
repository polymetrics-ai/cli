# Hubplanner connector

## Overview

Hubplanner (`hubplanner`) uses the provider-owned Hubplanner API v1 documentation in `hubplanner/API` at the source set recorded by issue #3239. This bundle now records the complete official operation ledger from `Sections/*.md` plus the `webhooks.md` supported event table: 107 official operations total, with 97 implemented as declarative streams, bounded typed direct reads, or typed reverse-ETL writes, and 10 outbound webhook event deliveries blocked on the shared CDC/changefeed runtime.

## Auth setup

Create credentials with `pm credentials add` using `api_key` from an environment variable or stdin. The connector sends that value in the Hubplanner `Authorization` header. Do not paste API keys into prompts, shell history, fixtures, issue comments, or docs. `base_url` defaults to `https://api.hubplanner.com/v1`.

## Streams notes

List-style GET resources use the shared Hubplanner `page=0&limit=200` pagination shape. The first stream (`resources`) keeps a two-page fixture to prove the zero-indexed page-number paginator terminates; other stream fixtures are sanitized single-page provider-shape samples. Detail GET endpoints that require an object id are exposed as bounded direct reads instead of ETL streams because the declarative stream contract has no per-record path-parameter input. The official operation allocation represented in `api_surface.json` is: 33 read operations, 59 reverse-ETL write operations, 2 provider-search/direct-read operations, and 13 CDC/webhook operations.

## Write actions & risks

`writes.json` declares 61 typed write actions: the 59 official reverse-ETL operations plus create/delete webhook subscription control operations from the CDC lane. Every action uses a fixed documented Hubplanner path, a closed `record_schema`, fixture-backed request-shape validation, and no raw method/path/body/query escape hatch. Delete actions and webhook-subscription creation are marked `confirm: destructive` because webhook creation approves future data egress to an HTTPS target; operators must use plan -> preview -> explicit approval -> execute, with typed confirmation for destructive actions. The Hubplanner docs do not document a universal idempotent-delete status contract, so delete actions do not claim missing-404 success unless a future provider source adds that evidence.

## Known limits

- Fixture-only parity is not live certification. `certification.json` records a default stream but no live-safe write pairing or direct-read candidate; no provider credentials or live Hubplanner calls were used.
- `create_webhook_subscription` intentionally does not expose the optional `authorization_token` request field because command records are persisted for plan/preview evidence; accepting a secret-bearing webhook token requires a future non-persisted secret input path.
- The 10 outbound webhook event names in `webhooks.md` (`project.update`, `resource.update`, `booking.create`, `timeEntry.create`, `timeEntry.update`, `timeEntry.create.update`, `timeEntry.delete`, `booking.update`, `booking.delete`, and `booking.delete.multiple`) are ledgered as blocked because the current connector contract cannot receive provider callbacks as CDC. The documented shared dependency is the CDC/changefeed runtime foundation tracked by #2986 and #2988.
- Relationship/helper endpoints present in the Markdown prose but outside the fixed r2 audit allocation (for example project/resource group membership helpers and project/resource tag attachment helpers) are not added as generic writes; adding them requires a future official-count refresh and typed action evidence.
