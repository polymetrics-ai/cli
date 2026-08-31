# Plan — issue 4400 Foundation Atlas demand register

## Task Delivery Header

- Issue: Refs #4400 — Batch R1 Foundation Atlas reconciliation and demand
  register.
- Base branch: `fm/cli-top100-declaration-batch-r1` at
  `0a708dea5e0024a173b19959d2c43f2bf5a6e0f2`.
- Merges into: `fm/cli-top100-declaration-batch-r1` → `main` only after
  independent review and captain approval; this task does not integrate.
- Delivery: one documentation-only candidate commit/push from
  `docs/4400-foundation-atlas-demand-register-r1`, ready for independent review.
- Task: reconcile the frozen Batch R1 deferred evidence with the Foundation
  Atlas, add four planned connector-specific inbound-webhook adapter records,
  and record the exact non-execution/approval boundary.
- Verification: source-accounting report, catalog-schema and unique-ID checks,
  owner/proof-link inspection, documentation link check, JSON parse, and diff
  check.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion |
| --- | --- | --- |
| Each assessed `missing_foundation` cell has a reviewable decision | live frozen cohort report | The register lists all 12 exact source IDs, lane, method/path, citation, current gap ID, disposition, and owner/proof route. |
| Mapping and runtime are not conflated | live source matrices and Atlas catalog | The 6,778 M-U cells remain `source.projection-admission.v1` reuse and all planned entries say no inbound request or sync transport is executable. |
| Existing shared foundations are reused, not duplicated | live owner symbols and proof tests | Each new planned entry names `transport.sync-contract.v1` and `runtime.provider-extension-seams.v1`, with no generic webhook runtime added. |
| Actual adapter gaps are narrowly planned | live connector-local source facts | Four new entries cover only Bitbucket, CircleCI, Jira, and Vercel source IDs; each requires separate captain approval before implementation. |
| Atlas remains documentation only | live catalog and README | No runtime loader, source lock, matrix, artifact, contract, certification, or executor file changes. |

## Documentation slices

1. **Red — absent planning records:** prove the four connector-specific planned
   Atlas IDs and the Batch R1 demand register are absent at the frozen base.
2. **Green — exact reconciliation:** add the register with all 12 exact M-F
   identities and a baseline separation for M-U, certification, importer, and
   existing GitLab/Stripe records.
3. **Green — narrow planned entries:** add Bitbucket, CircleCI, Jira, and
   Vercel connector-specific planned records mirroring the existing GitLab and
   Stripe conventions.  Each names only existing generic owner symbols and
   proof tests, creates no definition or runtime selection, and carries a
   captain-approval boundary.
4. **Refactor and prove:** add the README link, validate catalog shape and
   unique IDs, verify owner/proof links, re-run source accounting without
   modifying its inputs, and inspect the exact diff scope.

## Intended file ownership

| Path | Responsibility |
| --- | --- |
| `docs/connector-canon/foundations/BATCH-R1-DEMAND-REGISTER.md` | Exact Batch R1 reconciliation, classifications, source citations, and future proof/approval boundary. |
| `docs/connector-canon/foundations/catalog.json` | Four planned connector-specific reference records only. |
| `docs/connector-canon/foundations/README.md` | Minimal link to the issue-scoped demand register and its non-runtime status. |
| `.planning/phases/issue-4400-foundation-atlas-demand-register-r1/**` | GSD, TDD, and verification evidence. |

No file under `cmd/`, `internal/`, `data/`, or a connector definition may be
changed.  No `sync_transport.json`, ingress endpoint, worker, CLI command,
source executor, or provider-I/O path is planned.

## Future implementation boundary

The planned entries authorize no implementation.  A later connector-local
adapter must first have separate captain approval and source-backed evidence
for the provider delivery contract.  Its red/green proof must cover, as the
retained source permits: invalid authentication before parsing, source scope,
replay identity/duplicate handling without inference, durable staging,
selected worker/executor registration, and checkpoint only after downstream
acknowledgement.  It must select the existing closed transport and extension
seams; it must not create an arbitrary webhook runtime or connector-name branch.
