# Bitbucket connector

## Overview

This bundle records the complete official Bitbucket Cloud REST API 2.0 OpenAPI inventory from `https://developer.atlassian.com/cloud/bitbucket/swagger.v3.json`. The inventory observed 331 operations with method counts {'DELETE': 54, 'GET': 179, 'POST': 50, 'PUT': 48} and source SHA-256 `efd23d3948bf0b4d4ff5ad6ae6bab9479d8f00200783751e871df66b8db232c2`. The connector declares declarative streams for JSON GET operations that the current engine can express, typed reverse-ETL write actions for JSON/path mutations, and connector-local blocked operation rows for binary transfer, multipart upload, and provider-search/query surfaces whose shared bounded-command foundations are not present in this worktree.

## Auth setup

Use credentials from the credential store only. Configure `access_token` as a secret for bearer auth, or configure `username` plus secret `app_password` for Bitbucket app-password basic auth. Do not pass tokens, app passwords, OAuth refresh material, or authorization headers in command text, fixtures, docs examples, issue comments, or logs. `base_url` defaults to `https://api.bitbucket.org/2.0`.

## Streams notes

The stream ledger is generated from every official JSON GET operation not classified as binary artifact transfer or provider search/query; JSON listing endpoints such as repository downloads and pull request commits/diffstat are modeled as paginated streams. Paginated Bitbucket collection responses use `page` and `pagelen` with a bounded `max_pages` cap; singleton GET responses are modeled as single-object streams. Resource-scoped streams use optional config properties matching Bitbucket path parameter names such as `workspace`, `repo_slug`, `pull_request_id`, `commit`, and `issue_id`; a read of those streams fails closed if the needed config is absent. Fixtures are sanitized synthetic Bitbucket-shaped pages and do not come from live credentials.

## Write actions & risks

Supported mutations are limited to executable reverse-ETL actions with closed record schemas and sanitized request-shape fixtures; JSON mutations that only had path-parameter validation are blocked in the operation ledger until connector-local typed request schemas and fixtures prove the body contract. The implemented `create_repositories_workspace_repo_slug` action sends only JSON body fields `scm`, `is_private`, and optional `description`; `workspace` and `repo_slug` are path-only. `DELETE` actions use `kind: delete`, `body_type: none`, idempotent `missing_ok_status: [404]`, and `confirm: "destructive"`; they remain in scope under the captain policy correction only through the existing plan -> preview -> explicit approval -> execute path plus typed destructive confirmation. File/multipart operations, including issue imports, issue attachments, repository `/src` commits, and snippet file bodies, are blocked in `api_surface.json`/`operations.json` until a bounded file-transfer foundation can bind file snapshots, byte caps, and approvals without generic raw bytes. The issue-import ZIP operation is also destructive and must require destructive confirmation if it is later implemented.

Secret-bearing POST/PUT actions for pipeline variables, webhooks, and repository pipeline SSH key pairs remain blocked until connector-local typed request schemas, fixtures, and redaction evidence exist. If later implemented, those write actions must redact variable `value`, webhook `secret`, and SSH private key/passphrase fields from plans, previews, and write errors.

## Known limits

- No live Bitbucket credentials, provider calls, writes, certification, VPS, or Thaalam work were performed. Certification remains fixture-only and `uncertified` until an approved live-safe executor records redacted artifacts.
- Legacy or stale reverse plans created before `redact_fields` existed, or before a connector adds new `redact_fields`, still depend on a shared runtime fix that hydrates or unions `ReversePlan.RedactFields` from the current connector manifest during `GetReversePlan`, `ListReversePlans`, and output redaction before plan samples are rendered. This Bitbucket bundle keeps executable write `redact_fields` metadata current but does not claim stale-plan redaction is complete without that shared dependency.
- Provider search/query and direct/binary command execution are connector-local metadata only here. Provider query/search is blocked on shared foundation #2985; binary/file transfer execution needs the bounded transfer foundation described in the direct/binary lane.
- Bitbucket `/addon` mutations are blocked because the official API requires Atlassian Connect-app JWT authentication, while this bundle only declares bearer/basic/none credential shapes.
- Path-only POST/PUT mutations are blocked rather than advertised as runnable reverse ETL until connector-local typed request schemas, fixtures, and conformance request-shape evidence exist.
- CDC/changefeed semantics for webhook delivery are not claimed because CDC truthfulness/state foundations #2986 and #2988 are outside this connector-local scope. Webhook REST resources are ledgered as ordinary REST rows only; no CDC capability is advertised.
- Generated streams use passthrough projection except for secret-bearing Bitbucket resources. Pipeline variable streams, webhook streams, deployment-environment variable streams, and the repository pipeline SSH key-pair stream use schema projection and omit provider secret fields such as `value`, `secret`, and `private_key` from ordinary read output.
- The repository operation-count tables in GitHub issues are preserved by the captain-policy addendum and are not rewritten by this connector-local bundle.
