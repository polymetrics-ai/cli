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

None yet. Sub-issues will be linked to #3579 and integrated through stacked sub-PRs targeting this parent branch.

## Required slices

1. Target-scope contract and core validator
2. GitHub Actions, tag/label, local hook, and required remote gate
3. PM orchestrator and no-mistakes integration
4. HubSpot and Bitbucket forward remediation
5. Stripe, Freshchat, and Google Ads forward remediation
6. Zendesk Support and Google Ads unrelated-connector/generated remediation
7. Historical audit ledger and end-to-end enforcement proof

## Verification status

- Seed commit only: planning/state artifacts committed.
- Full verification pending implementation integration.
- Final no-mistakes validation pending Firstmate instruction after all sub-issues integrate.

## Safety

- Parent PR is draft.
- Parent PR merge to `main` is human-gated.
- No secrets or credentialed connector checks used.
- No active connector branches modified.
