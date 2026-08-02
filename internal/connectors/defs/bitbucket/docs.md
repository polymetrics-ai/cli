# Bitbucket connector

## Overview

This bundle records the complete official Bitbucket Cloud REST API 2.0 OpenAPI inventory from `https://developer.atlassian.com/cloud/bitbucket/swagger.v3.json`. The inventory observed 331 operations with method counts {'DELETE': 54, 'GET': 179, 'POST': 50, 'PUT': 48} and source SHA-256 `efd23d3948bf0b4d4ff5ad6ae6bab9479d8f00200783751e871df66b8db232c2`. The connector declares declarative streams for JSON GET operations that the current engine can express, a narrow closed-schema reverse-ETL surface for typed repository creation and path-only deletes, and connector-local blocked operation rows for untyped JSON body mutations, binary transfer, typed JSON direct reads, multipart upload, and provider-search/query surfaces whose shared bounded-command foundations are not present in this worktree.

## Auth setup

Use credentials from the credential store only. Configure `access_token` as a secret for bearer auth, or configure `username` plus secret `app_password` for Bitbucket app-password basic auth. Do not pass tokens, app passwords, OAuth refresh material, or authorization headers in command text, fixtures, docs examples, issue comments, or logs. `base_url` defaults to `https://api.bitbucket.org/2.0`.

## Streams notes

The stream ledger is generated from every official JSON GET operation that has typed stream coverage in this bundle. Paginated Bitbucket collection responses use `page` and `pagelen` with a bounded `max_pages` cap; singleton GET responses are modeled as single-object streams. Resource-scoped streams use optional config properties matching Bitbucket path parameter names such as `workspace`, `repo_slug`, `pull_request_id`, `commit`, and `issue_id`; a read of those streams fails closed if the needed config is absent. Repository, workspace, project, account, pipeline, and configuration-resource streams use provider `uuid`, `account_id`, `fingerprint`, `key`, `name`, or `external_id` fields when those keys are present; commit streams use commit `hash` values, and streams without a confirmed top-level provider key do not claim deduped sync. The `/user/emails` and `/user/emails/{email}` endpoints remain blocked because the official OpenAPI document publishes no provider-backed 2xx email-address response schema for this bundle to model. Fixtures are sanitized synthetic Bitbucket-shaped pages and do not come from live credentials.

## Write actions & risks

Executable reverse-ETL actions now use closed record schemas. The retained JSON-body write is repository creation, which requires `workspace`, `repo_slug`, and `scm` and accepts only the typed body fields `scm` and optional `is_private`. `DELETE` actions use `kind: delete`, `body_type: none`, idempotent `missing_ok_status: [404]`, `additionalProperties: false`, and `confirm: "destructive"`; they remain in scope under the captain policy correction only through the existing plan -> preview -> explicit approval -> execute path plus typed destructive confirmation. JSON mutations without connector-owned typed body fields are blocked in `api_surface.json`/`operations.json` until the bundle declares closed request-body schemas and redaction evidence. File/multipart upload operations are not exposed as raw writes; they are blocked until a bounded file-transfer foundation can bind file snapshots, byte caps, and approvals without generic raw bytes.

Secret-bearing JSON-body writes for pipeline variables, webhooks, deployment variables, and repository pipeline SSH key pairs are blocked until those actions declare closed typed body schemas with required secret fields and matching `redact_fields`. Redaction evidence in this branch applies to plans created after action metadata is persisted; hydrating saved pre-upgrade plans that lack `redact_fields` is a shared runtime dependency and is not claimed by the Bitbucket bundle.

## Known limits

- No live Bitbucket credentials, provider calls, writes, certification, VPS, or Thaalam work were performed. Certification remains fixture-only and `uncertified` until an approved live-safe executor records redacted artifacts.
- Provider search/query, user email reads, planned JSON direct reads, and direct/binary command execution are connector-local metadata only here. User email endpoints are blocked on provider-backed response schema fixtures; provider query/search is blocked on shared foundation #2985; binary/file transfer execution needs the bounded transfer foundation described in the direct/binary lane.
- CDC/changefeed semantics for webhook delivery are not claimed because CDC truthfulness/state foundations #2986 and #2988 are outside this connector-local scope. Webhook REST resources are ledgered as ordinary REST rows only; no CDC capability is advertised.
- Generated streams use passthrough projection except for secret-bearing Bitbucket resources. Pipeline variable streams, webhook streams, deployment-environment variable streams, and the repository pipeline SSH key-pair stream use schema projection and omit provider secret fields such as `value`, `secret`, and `private_key` from ordinary read output.
- The repository operation-count tables in GitHub issues are preserved by the captain-policy addendum and are not rewritten by this connector-local bundle.
