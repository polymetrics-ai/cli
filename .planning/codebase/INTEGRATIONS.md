# Integrations

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Connector Integrations

**Definition bundles:**
- Location: `internal/connectors/defs/<name>/`
- Current working-tree count during onboarding: 547 connector definition directories.
- Each complete bundle may include `metadata.json`, `spec.json`, `streams.json`, optional `writes.json`, `api_surface.json`, schemas, fixtures, and `docs.md`.

**Hooks:**
- Location: `internal/connectors/hooks/<name>/`
- Current working-tree count during onboarding: 78 hook directories.
- Used for connector-specific auth, stream, record, write, and check escape hatches.

**Native connectors:**
- Location: `internal/connectors/native/<name>/`
- Current working-tree count during onboarding: 37 native directories.
- Used for non-REST protocols and full custom connector behavior such as databases, queues, custom data services, and built-ins.

## External Systems and Protocol Families

Connector parity must reconcile all documented upstream product surfaces, including:

| Surface family | Examples / evidence | Planning treatment |
|---|---|---|
| REST/HTTP JSON | Most `integration_type: api` bundles | ETL stream, write action, direct-read, binary, or typed exclusion |
| GraphQL | `github`, `linear`, `monday`, `notion`, `plaid`, `stigg` mentions | Treat query/mutation/subscription operations as first-class surfaces, not duplicated REST gaps |
| XML/SOAP/XML feeds | `rss`, `workday`, `amazon-sqs`, `tally-prime` mentions | Treat XML parsing/serialization as protocol-specific coverage or hook/native work |
| CSV/NDJSON/report exports | `amplitude`, `appsflyer`, `mixpanel`, `vercel` mentions | Classify as report/export streams or binary/direct-read surfaces based on durability |
| Binary downloads/uploads | artifacts, attachments, archives, files, documents, media | Classify as binary transfer, not JSON stream, unless metadata records are separately documented |
| File/object storage | signed downloads, S3-like resources, local files | Use file/binary/native surface classes with safety gates |
| Databases and CDC | `postgres`, `dynamodb`; CDC stubs/gates | Native/database/CDC class; dependency-gated live replication remains human-gated |
| Queues/events/webhooks/audit logs | `amazon-sqs`, webhook/event/audit resources | Classify as queue/event/audit surfaces; avoid duplicating same event as stream + direct-read |
| Admin/destructive/elevated scope | org/admin/delete/security operations | Human-gated or typed exclusion unless product-safe and explicitly approved |

## Credential and Secret Integrations

- Credentials are managed through CLI/vault flows; planning must not request, print, store, or summarize secret values.
- Connector specs mark secret fields with `x-secret` in JSON Schema.
- Certification can treat missing live credentials as `uncertified`, not failed.

## CI and Review Integrations

- GitHub Actions workflows under `.github/workflows/` run verification, security, scorecard, release, website, and PR issue guard checks.
- Automated review route follows `.agents/agentic-delivery/workflows/automated-review-routing-loop.md` and CodeRabbit/Copilot fallback rules from `AGENTS.md`.

---
*Integration analysis: 2026-07-08*
