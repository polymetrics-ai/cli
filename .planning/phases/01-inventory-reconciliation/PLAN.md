# Phase 1 Plan — Inventory and Surface Reconciliation

**Phase:** 1 — Inventory and Surface Reconciliation
**Generated via:** Upstream `/gsd:plan-phase 1 --skip-research` workflow shape
**Issue:** #122

## Objective

Rebootstrap GSD planning and then produce a trusted, generated connector inventory before any connector fanout. The inventory must classify every documented connector surface across protocols and avoid duplicated operations.

## Scope

### Plan 01-01 — Rebootstrap Planning

- Archive prior `.planning/` outside active planning.
- Replace active `.planning/` with upstream GSD Core style artifacts.
- Record command/workflow sources and onboarding prompt.
- Verify no Go source changes.

### Plan 01-02 — Generate Inventory

- Generate current counts from repository state:
  - defs bundles
  - hooks
  - natives
  - schemas
  - streams
  - writes
  - `api_surface.json` entries
  - docs files
  - conformance fixtures
  - certification metadata
  - direct-read/binary/operation metadata where present
  - quarantine/blocker state
- Classify documented surfaces across:
  - REST/HTTP JSON
  - GraphQL
  - XML/SOAP/XML feeds
  - CSV/TSV/NDJSON/report exports
  - binary transfers
  - file/object storage
  - SQL/database/CDC
  - queues/events/webhooks/audit logs
  - native protocols
  - direct-read
  - mutations/writes
  - typed exclusions

### Plan 01-03 — Review and Gate Fanout

- Review generated inventory for duplicates, gaps, stale counts, and unsafe classifications.
- Confirm each upstream operation has exactly one primary classification.
- Update blocker/quarantine lists as needed.
- Decide whether connector fanout may proceed.

## Non-goals

- Do not implement connector runtime changes.
- Do not edit `cmd/` or `internal/` in issue #122.
- Do not add dependencies.
- Do not run credentialed checks.
- Do not run real reverse ETL execution.

## TDD / Validation Evidence

### Red evidence

- Legacy `.planning/` contained custom phase artifacts.
- `.planning/codebase/` was absent.
- Prior counts in docs/planning artifacts were stale or inconsistent with current repo scans.

### Green evidence

- Active `.planning/` exists in upstream GSD Core shape.
- Codebase maps exist.
- Roadmap starts with inventory/surface reconciliation.
- Requirements cover non-REST surfaces and de-duplication.
- Config parses.
- Grep confirms connector parity, reverse ETL, binary, direct-read, GraphQL, XML/SOAP, CDC, queue, human-gated, certification, and conformance language.
- `git diff --name-only -- cmd internal` is empty.

## Verification Commands

```bash
node -e "JSON.parse(require('fs').readFileSync('.planning/config.json','utf8')); console.log('config ok')"
test -f .planning/PROJECT.md
test -f .planning/REQUIREMENTS.md
test -f .planning/ROADMAP.md
test -f .planning/STATE.md
test -d .planning/codebase
rg -n "connector parity|reverse ETL|binary|direct-read|GraphQL|XML|SOAP|CDC|queue|human-gated|certification|conformance" .planning

git diff --check
git diff --name-only -- cmd internal
```

## Human Gates

- New dependencies.
- Credentialed live connector checks.
- Reverse ETL execution.
- Destructive/admin/elevated-scope operation enablement.
- Quality-gate reductions.
- Merge to `main`.
