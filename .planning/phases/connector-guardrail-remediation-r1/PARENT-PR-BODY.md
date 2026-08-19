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

Integrated into parent branch:

- Refs #3583 — PM orchestrator and no-mistakes integration; sub-PR #3588 squash-merged at parent commit `86b91fc40f46b8653538531fc40c183913676f05` from validated head `0c321595d7ae4852550a5012a895c3e11f7e8298`; no-mistakes run `01KZ0SEAKBB9TG7N3SMG97XKJS` passed. Review coverage remains provisional through the parent PR fallback/human gate until final parent readiness.

Planned / pending integration:

- Refs #3595 — canonical icon-registry single-source foundation; draft sub-PR #3596 open and must land before #3590 reconciliation.
- Refs #3581 — target-scope contract and core validator; sub-PR #3590 open but blocked on #3595 before integration.
- Refs #3582 — CI, hooks, label, and required remote gate; blocked on #3581.
- Refs #3584 — HubSpot and Bitbucket forward remediation; sub-PR #3591 open, fresh no-mistakes recovery pending.
- Refs #3585 — Stripe Freshchat Google Ads shared remediation; sub-PR #3593 open.
- Refs #3586 — generated and unrelated connector remediation; sub-PR #3589 open.
- Refs #3587 — first-eight audit ledger and enforcement proof; blocked on #3581/#3582 plus remediation rows.

## Dependency graph / current worker queue

- #3595 open: canonical icon-registry foundation; blocks #3590/#3581 R5/R6 reconciliation.
- #3581 open/blocked: core validator waits for #3595 foundation, then requires fresh native-Codex `gpt-5.6-sol` validation before integration; blocks #3582 and #3587 final enforcement proof.
- #3583 provisionally integrated: PM/no-mistakes guidance landed on the parent branch; parent review/final gate pending.
- #3584 open: HubSpot/Bitbucket remediation; recover fresh no-mistakes run before merge arbitration.
- #3585 open: shared engine/runner/connectorgen remediation; ledger-only fallback avoided the prior `cmd/connectorgen/**` collision.
- #3586 open: generated/unrelated connector remediation; canonical stacked PR #3589 remains open.
- #3582 blocked on #3581.
- #3587 blocked on #3581/#3582 plus remediation rows.

## Verification status

- #3583 sub-PR validation: no-mistakes run `01KZ0SEAKBB9TG7N3SMG97XKJS` passed at `0c321595d7ae4852550a5012a895c3e11f7e8298` with review/test/document/lint/push/pr/ci complete.
- Parent branch after #3583 merge: starts from `86b91fc40f46b8653538531fc40c183913676f05`; this roadmap records the restored/updated parent ledger before continuing with #3590.
- #3595 planning scaffold: draft PR #3596 opened at `b814e85a6`; implementation and comprehensive native-Codex `gpt-5.6-sol` validation pending.
- PR #3590 R5/R6 gate remains unanswered: R5's second-registry approach is rejected/superseded by #3595, and R6 waits for the foundation.
- Remaining child validation pending current no-mistakes/review arbitration before merge.
- Full parent verification and final no-mistakes validation pending after all required sub-issues integrate.

## Safety

- Parent PR is draft.
- Parent PR merge to `main` is human-gated.
- No secrets or credentialed connector checks used.
- No active connector branches modified.
