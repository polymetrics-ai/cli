# Context — issue 4400 Foundation Atlas demand register

## Decision record

- **Issue:** Refs #4400 — Batch R1 Foundation Atlas reconciliation and demand
  register.
- **Base:** `fm/cli-top100-declaration-batch-r1` at
  `0a708dea5e0024a173b19959d2c43f2bf5a6e0f2`.
- **Delivery:** a documentation-only candidate branch
  `docs/4400-foundation-atlas-demand-register-r1`; no parent integration or
  main merge in this task.
- **Authority:** captain-approved Atlas and issue-evidence documentation only.
  Runtime code, source locks, matrices, connector definitions and artifacts,
  contracts, certification, imports, and provider I/O are out of scope.

## Evidence boundary

`connectorgen deferred-visibility` is a check-only reporter over the frozen
Batch R1 cohort.  At the stated base it reports 4,341 primary source operations,
4,343 source rows (including two supplements), 30,401 matrix cells, 6,790
deferred cells, and zero executable declarations.

Of the deferred cells, 6,778 are `mapped_unproven`: they retain source-backed
mapping and cite the existing authoring prerequisite
`source.projection-admission.v1`.  They are not assertions that a runtime
foundation is absent.

The remaining 12 are `missing_foundation` `sync_transport` cells.  They are
source-backed provider webhook-registration or update operations:

- Bitbucket: four repository/workspace create or update hook identities;
- CircleCI: `createWebhook` and `updateWebhook`;
- GitLab: group, system, and project hook creation;
- Jira: `registerDynamicWebhooks`;
- Stripe: `PostWebhookEndpoints`;
- Vercel: `createWebhook`.

The existing `transport.sync-contract.v1` and
`runtime.provider-extension-seams.v1` cover generic closed composition and the
connector-owned extension seam.  They deliberately do not create HTTP ingress,
validate an inbound provider delivery, infer replay semantics, or register a
source executor from a provider webhook-registration request.

## Atlas classification

| Evidence class | Atlas result | Consequence |
| --- | --- | --- |
| 6,778 `mapped_unproven` cells | Reuse `source.projection-admission.v1` | Preserve mapping visibility; do not call this a runtime gap. |
| GitLab regex and typed-alias records | Existing reusable/partial contracts; connector-local mapping restriction where stated | Preserve their current status; do not introduce a generic runtime demand. |
| GitLab and Stripe webhook cells | Existing planned connector-specific entries | Retain their non-executing, captain-gated plan. |
| Bitbucket, CircleCI, Jira, and Vercel webhook cells | Reuse generic sync/extension seams plus a planned connector-specific inbound adapter | Add only an authoring-only planned Atlas record; no executor claim. |

No generic webhook receiver or other shared runtime foundation is proposed.
Every future connector-specific adapter remains subject to a separate captain
approval, source evidence, red tests, and conformance proof.

## Manual GSD record

The project-local GSD adapter was checked with `scripts/gsd doctor`; source
resolution succeeded for `discuss-phase`, `plan-phase`, `execute-phase`,
`verify-work`, and `code-review`.  This is an Atlas/documentation-only slice,
not a behavior change.  The task uses the repository-required inline/manual
record in this context, plan, TDD ledger, and verification file; no compatible
interactive Pi worker is available in this isolated worktree.

## Skills

Loaded: `connector-lane-build-order`, `firstmate-exhaustive-review`,
`github-issue-first-delivery`, and `golang-documentation`.
