Refs #3579

## Intent

Forward-only connector path ownership guardrail remediation parent PR. Draft remains open while stacked sub-PRs land into `fix/3579-connector-path-ownership-guardrails`.

## Parent roadmap

- Parent issue: https://github.com/polymetrics-ai/cli/issues/3579
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Final target branch: `main`
- Orchestrator: active PM parent-issue orchestrator
- State ledger: `.planning/phases/connector-guardrail-remediation-r1/RUN-STATE.json`

## Integrated sub-issues

Planned / pending integration:

- Refs #3581 — target-scope contract and core validator
- Refs #3582 — CI, hooks, label, and required remote gate
- Refs #3583 — PM orchestrator and no-mistakes integration
- Refs #3584 — HubSpot and Bitbucket forward remediation
- Refs #3585 — Stripe Freshchat Google Ads shared remediation
- Refs #3586 — generated and unrelated connector remediation
- Refs #3587 — first-eight audit ledger and enforcement proof

No sub-PRs integrated yet.

## Dependency graph / current worker queue

- #3581 ready: core validator; blocks #3582 and #3587.
- #3583 ready: PM/no-mistakes guidance; disjoint from core/remediation code.
- #3584 ready: HubSpot/Bitbucket remediation; disjoint from generated remediation.
- #3586 ready: generated/unrelated connector remediation; disjoint docs/generated scope.
- #3585 queued: shared engine/runner/connectorgen remediation; `cmd/connectorgen/**` may collide with #3581, so defer implementation until #3581 stabilizes or narrow to ledger-only.
- #3582 blocked on #3581.
- #3587 blocked on #3581/#3582 plus remediation rows.

## Verification status

- Seed commit only: planning/state artifacts committed.
- Full verification pending implementation integration.
- Final no-mistakes validation pending Firstmate instruction after all sub-issues integrate.

## Safety

- Parent PR is draft.
- Parent PR merge to `main` is human-gated.
- No secrets or credentialed connector checks used.
- No active connector branches modified.
