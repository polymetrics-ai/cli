# Mixpanel connector

## Overview

This bundle is generated from the official Mixpanel OpenAPI YAML files linked from <https://docs.mixpanel.com/reference/overview>. It records every documented operation exactly once in `api_surface.json` and keeps duplicate method/path rows distinct across Mixpanel API documents.


## Auth setup

Use `pm connectors inspect mixpanel --json` before adding credentials. Supply secrets from environment variables or stdin only. The connector supports the existing Mixpanel custom Basic-auth hook: `username` config or `username_secret`, plus `password` or `api_secret`. Some documented Mixpanel operation families also require query/config fields such as `project_id`, `projectId`, `workspace_id`, or project-token fields; configure those on the saved credential or pass explicit `--config key=value` overrides.

## Current parity ledger

| Disposition | Count |
| --- | ---: |
| Official operations | 105 |
| Implemented fixture-backed operations | 100 |
| Blocked/planned operations | 5 |
| Excluded/not-applicable operations | 0 |
| Certified live operations | 0 |

| Lane | Official | Implemented | Blocked/planned |
| --- | ---: | ---: | ---: |
| `etl_read` | 24 | 24 | 0 |
| `cdc_changefeed` | 1 | 1 | 0 |
| `direct_read_query_search` | 18 | 15 | 3 |
| `binary_file` | 1 | 0 | 1 |
| `reverse_etl_write` | 61 | 60 | 1 |

## Streams notes


- **Streams:** 25 documented read/changefeed-like operations are declared as bounded replay-tested streams under `streams.json`.
- **Direct reads:** 15 safe JSON Query API GET operations are declared in `operations.json` and exposed through `cli_surface.json` with `json_redacted` output.
- **Writes:** 60 documented mutations are declared as typed reverse-ETL write actions with closed record schemas and sanitized write fixtures.

## Write actions & risks

Reverse-ETL actions are typed and approval-gated. Destructive/admin operations use destructive confirmation metadata and idempotency notes where supported by the provider surface. Use `pm reverse plan`, `pm reverse preview`, and `pm reverse run` rather than attempting direct mutation from a help or inspect command.

## Base URL families

Mixpanel publishes different API documents on different hosts. `base_url` is intentionally configurable per credential or command invocation:

| Family | Default host to use |
| --- | --- |
| Query API direct reads | `https://mixpanel.com/api/query` |
| App API resources (schemas, service accounts, annotations, GDPR, warehouse, management, experiments) | `https://mixpanel.com/api/app` |
| Ingestion and Identity APIs | `https://api.mixpanel.com` |
| Export/Data Pipelines | `https://data.mixpanel.com/api/2.0` |
| Feature flag evaluation | `https://api.mixpanel.com` |

The default `base_url` is the Query API because direct provider-query commands are the most host-specific CLI surface. ETL and reverse-ETL workflows may use credentials configured with the operation family's host.

## Known limits


The connector does **not** expose arbitrary JQL scripts, generic HTTP method/path/body passthrough, raw CSV upload, raw binary export, shell, or local file escape hatches. Unsupported official operations remain blocked/planned in `api_surface.json`:

- `POST /jql` is disallowed as an arbitrary provider-side script surface.
- `POST /cohorts/list` and `POST /engage` need a safe form/query direct-read executor.
- `GET /export` needs a bounded binary/file export executor.
- `PUT /lookup-tables/{id}` needs a typed CSV upload executor.

Destructive/delete/admin operations are in scope when represented as typed reverse-ETL actions. Implemented destructive actions declare `confirm: "destructive"`, idempotent 404 handling where applicable, redaction metadata, and rely on the existing plan → preview → explicit approval → execute flow.

## Fixture and certification status

All implemented streams and write actions have sanitized replay fixtures. No live Mixpanel credentials were requested or used, no provider calls were made, and no live certification is claimed in this bundle.
