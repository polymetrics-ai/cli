# Concerns

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Planning Concerns

### Legacy planning contamination

The previous `.planning/` tree was custom/legacy GSD-style state with many stale phase directories. It must remain archived outside active planning and should not be used as the new milestone source of truth.

### Inventory drift

Generated counts have changed across prior docs and worktrees. Current onboarding observed 547 defs bundles, 78 hook dirs, 37 native dirs, and 29,123 `api_surface.json` endpoint rows, but Phase 1 must regenerate authoritative counts from current code and docs before fanout.

### REST-only blind spot

Planning that says only "API endpoints" risks missing GraphQL, XML/SOAP, CSV/NDJSON exports, binary transfers, file/object storage, SQL/CDC, queues/events/webhooks, and other connector technologies. The new requirements and roadmap explicitly model connector surfaces across protocols.

### Duplicate surface counting

Generated docs, product guides, OpenAPI specs, GraphQL schemas, webhook pages, and binary export docs can describe the same capability in multiple places. Inventory needs canonical operation identity and alias tracking to avoid duplicate streams/actions/commands.

## Runtime and Safety Concerns

### Secret exposure

No issue #122 work requires secrets. Credentialed checks and live certification remain separate and must never print or summarize secret values.

### Destructive/admin write paths

Repository deletion, admin mutations, elevated scopes, destructive deletes, and dependency-sensitive capabilities must remain human-gated or typed exclusions unless explicitly approved.

### Generic raw tool exposure

The repo rules explicitly forbid generic shell, generic HTTP write, and generic SQL write tools. Parity work must expose product-specific safe commands/actions, not raw escape hatches.

### Dependency gates

Some native/CDC or protocol-specific work may require dependencies. New dependencies are human-gated and out of scope for issue #122.

## Review Concerns

- CodeRabbit skipped/disabled/rate-limited review is not completed review coverage.
- Parent/main PR merges remain human-gated.
- Any follow-on connector fanout should use disjoint worktrees/paths to avoid `defs/` index and `go:embed` collisions.

---
*Concern analysis: 2026-07-08*
