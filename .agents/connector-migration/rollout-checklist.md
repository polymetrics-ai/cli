# Connector Rollout Checklist

Use this checklist for every connector rolled out after the GitHub pilot. It is the single
source of truth a connector agent follows end-to-end. Each item must be true before the
connector's slice is considered integrated.

## 1. Inputs and scope

- [ ] One connector assigned to one implementation agent (no shared-file collisions).
- [ ] Delivery profile is explicit: `pm_worker` for issue/branch/stacked-PR delivery, or
      `coordinator_fanout` for an explicitly coordinator-owned no-commit bulk task.
- [ ] Issue names the connector, branch, PR base, primary agent, verification, and human gates when
      the profile is `pm_worker`; the coordinator prompt names the isolated directory and handoff
      boundary when the profile is `coordinator_fanout`.
- [ ] Branch starts from `feat/44-github-cli-parity` (or the active rollout parent branch), not `main`.
- [ ] Worker has an isolated working directory or git worktree before any edit.
- [ ] No production `internal/connectors/defs/<name>/` edits unless the issue explicitly assigns this connector.

## 2. Inventory and source links

- [ ] Provider CLI/API inventory captured with **official source URLs** for every endpoint group.
- [ ] Parity matrix: `gh`/provider-CLI command → Polymetrics stream/write/operation (or `gap`).
- [ ] Every API surface row has a `source_url` pointing at official provider docs.
- [ ] No GitHub-specific assumptions (e.g. `owner/repo` path shape, `gh` subcommands) left unparameterized.

## 3. CLI surface and operation ledger

- [ ] `cli_surface.json` drafted with commands, flags, `intent` (`etl`/`direct_read`/`reverse_etl`/`local_workflow`), `availability`, and `write` refs.
- [ ] `api_surface.json` rows classify each endpoint into an execution model (`stream_read`, `direct_read`, `sensitive_reverse_etl`, `admin_reverse_etl`, `destructive_admin`, …).
- [ ] `writes.json` actions declare `record_schema`, `path_fields`, `body_type`, `risk`, and `hook` only when a compound action needs Go behavior.
- [ ] `reverse_etl` commands that map to a write action have `record.*` flag mappings; unmapped commands stay `availability=partial` with a note.

## 4. Help preview and docs

- [ ] `pm connectors inspect <name> --json` runs without reading credentials.
- [ ] Help preview rendered from metadata (`pm <name> --help` shape) and attached to the slice.
- [ ] Connector `docs/connectors/<name>/**` generated/validated where the repo requires it.

## 5. Validation gates (must all pass)

- [ ] `jq .` on every edited JSON file (JSON parses).
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs --json` → 0 findings, 0 warnings.
- [ ] Secret scan: no secret values, tokens, or PEMs in any artifact (docs, examples, previews, fixtures, errors).
- [ ] Source-link gate: every `api_surface` row has a non-empty `source_url`.
- [ ] Operation-classification gate: every row has an `execution_model` (no `partial`/`planned`/`unsupported_api` for API-backed commands unless explicitly gap-documented).
- [ ] `gofmt`, `go vet`, `go build ./cmd/pm`, focused package tests, `make verify` when feasible.
- [ ] Website `pnpm run gen:website-data` idempotent (regen twice, no diff) when connector docs change.

## 6. Handoff and merge

- [ ] `pm_worker` handoff uses `.agents/agentic-delivery/contracts/pm-worker-handoff-template.md`;
      `coordinator_fanout` returns an uncommitted path inventory and verification handoff instead.
- [ ] `pm_worker` committed/pushed coherent green slices only to its assigned branch and its sub-PR
      targets the parent branch with `Refs #<sub-issue>` and `Refs #44` (no closing keywords).
      `coordinator_fanout` did not commit, push, or mutate a PR.
- [ ] V4 exact-head compilation is ready, every packet was rendered by the canonical renderer,
      one authenticated synthesis is clean, every finding is dispositioned, and independent
      Shepherd returns `PROCEED` for the same exact identities (`pm_worker` profile).
- [ ] Shared/generated files remain coordinator-owned; the worker did not commit generated files unless authorized.

## 7. Human gates (stop)

- [ ] New dependencies, auth-scope changes, secrets, destructive external actions, production deploys, quality-gate reductions, generic write tools, or reverse ETL outside plan/preview/approval/execute.
