# Overview

Nebius Token Factory exposes an OpenAI-compatible API at `https://api.tokenfactory.nebius.com`. This bundle covers the current Swagger/OpenAPI surface plus the legacy `/v1/batches` read stream retained from `internal/connectors/nebius-ai` until the wave6 cutover. Legacy streams keep schema projection; newly added streams use passthrough projection from the documented response shape.

## Auth setup

Provide the `api_key` secret. The connector sends it with bearer auth. `base_url` defaults to `https://api.tokenfactory.nebius.com`. Path-scoped streams require the matching ID config values listed in `spec.json`.

## Streams notes

The bundle declares 26 streams. `models`, `files`, and `batches` preserve the legacy projected fields and OpenAI-style `data[]` records. Additional streams cover file detail/content/link endpoints, fine-tuning jobs/events/checkpoints/trainable models, dedicated endpoints/templates, datasets and multipart-upload metadata, and operations/results/errors.

OpenAI-style list streams use `after` plus the last record `id` and `has_more`; offset-style dataset content uses `limit`/`offset`. Endpoints whose docs do not expose a continuation token are single-page streams.

## Write actions & risks

The bundle declares 20 write actions for JSON or bodyless documented mutations, including inference calls, embeddings, reranking, response creation, image generation, fine-tuning job creation/cancel, dedicated endpoint management, dataset metadata/uploads completion/cancel, operations, and deletes. These actions can incur cost, create resources, start jobs, cancel jobs, or delete resources, so all writes require approval and deletes are marked destructive.

## Known limits

Binary upload endpoints are excluded with `binary_payload`: file upload, custom model archive upload, and dataset upload part. The legacy `batches` stream is retained for record-data parity even though the current Token Factory OpenAPI no longer lists a `/v1/batches` endpoint. Optional filters such as project IDs and sort controls are not exposed as config; only required query parameters and the shared `limit` page-size control are modeled. The operation errors endpoint returns scalar string entries under `data[]`; because the declarative record extractor emits object records only, that stream emits the response envelope as one passthrough record.
