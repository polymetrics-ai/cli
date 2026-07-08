# Research — Connector Parity Architecture

**Generated via:** Upstream `/gsd:new-project --auto` research step shape  
**Date:** 2026-07-08

## Existing Architecture to Preserve

- JSON definition bundles under `internal/connectors/defs/<name>/`.
- Declarative engine under `internal/connectors/engine/`.
- Tier 2 hooks under `internal/connectors/hooks/<name>/`.
- Tier 3 natives under `internal/connectors/native/<name>/`.
- Conformance under `internal/connectors/conformance/`.
- Certification under `internal/connectors/certify/`.
- CLI surfaces under `cmd/pm` and `internal/cli`.

## Recommended Surface Model

Use one canonical operation inventory per connector with a primary classification:

```text
operation identity = connector + protocol + normalized resource/path/name + method/action + product scope
```

Primary classifications:

- `etl_stream`
- `reverse_etl_write`
- `direct_read`
- `binary_transfer`
- `native_protocol`
- `event_webhook_audit`
- `excluded`

Aliases from docs, generated specs, or product guides should link to a canonical operation rather than duplicate coverage work.

## Runtime Tier Guidance

- Declarative Tier 1 for simple HTTP/JSON request/response streams and writes.
- Hooks for GraphQL body mutation, custom auth, XML/SOAP transforms, CSV/NDJSON parsing, multipart, report polling, sub-resource fanout, binary request shaping, or compound writes.
- Natives for SQL/database/CDC, queues, local files, and non-HTTP protocols where the connector is not naturally request/response JSON.

## Planning Implication

Phase 1 must generate a multi-surface inventory before fanout so later phases can dispatch the right implementation type and avoid forcing all connector work into `api_surface.json` REST semantics.
