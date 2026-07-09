# Sources: Front Operation Ledger (#192)

No credentials were used. No secret values were requested, printed, stored, or summarized.

## Repo sources

- `AGENTS.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `docs/architecture/connector-operation-kernel.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `internal/connectors/engine/schema/api_surface.schema.json`
- `cmd/connectorgen/validate.go`
- `internal/connectors/defs/front/api_surface.json`
- `internal/connectors/defs/front/streams.json`
- `internal/connectors/defs/front/metadata.json`

## Public Front sources

- `https://dev.frontapp.com/llms.txt`
- `https://dev.frontapp.com/reference/introduction`
- ReadMe public API registry metadata discovered in the rendered docs:
  - `core-api.json`, UUID `103mteemr3o8hk5`
  - `channel-api.json`, UUID `48cgkyj15mnqc4ttp`
- Public registry fetches:
  - `https://dash.readme.com/api/v1/api-registry/103mteemr3o8hk5`
  - `https://dash.readme.com/api/v1/api-registry/48cgkyj15mnqc4ttp`

## Captured counts

- Core REST OpenAPI registry: 244 operations.
- Channel OpenAPI registry: 11 operations.
- Combined REST OpenAPI registry: 255 operations.
- Combined method split: `GET=123`, `POST=76`, `PATCH=26`, `PUT=3`, `DELETE=27`.
- `llms.txt` API Reference markdown links: 346.
- Non-OpenAPI API Reference links: 91 category/guide/plugin SDK/data-model pages.

## Parent baseline mismatch

Parent issue #188 records a 342-operation baseline. The current public ReadMe OpenAPI registries do not reproduce that count. This phase uses the public registry as the source of REST method/path truth and records non-REST reference links as source context rather than inventing method/path rows.
