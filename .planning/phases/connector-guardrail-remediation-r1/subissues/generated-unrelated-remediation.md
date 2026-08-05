## Objective

Forward-remediate and record dispositions for Zendesk Support #3532 and Google Ads #3535 unrelated-connector/generated/manual/website churn without rewriting `main` history.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3586-generated-unrelated-connector-remediation`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`
- Dependency: none for ledger/generated remediation; coordinate with #3581 for guard fixture names if needed.

## Scope

Allowed write scope:

- remediation ledger/proof entries for PR #3532 and #3535 generated/unrelated connector findings
- generated/manual/docs/website correction for the audited unrelated connector outputs only
- guard fixtures proving unrelated connector docs/defs/generated website churn fails in connector lanes
- generator validation tests if needed

Primary audited paths include:

- unrelated connector manuals/skills in `docs/connectors/{bahmni,bitbucket,gong,hubspot,xero}/**` from #3532
- unrelated `internal/connectors/defs/gong/cli_surface.json` from #3535
- broad website app/tooling/test paths from #3532/#3535 only when demonstrated stale/unrelated
- target outputs for `zendesk-support` and `google-ads` only as needed to preserve valid generated state

Do not edit shared engine/commandrunner/connectorgen remediation code, HubSpot/Bitbucket shared CLI/app code, workflow/ruleset wiring, or PM/no-mistakes guidance.

## Required reading / skills

Read `AGENTS.md`, issue-agent contract, GSD universal loop, GSD adapter reference, required-skills routing, CLI help/docs/website parity reference, first-eight audit report sections for #3532/#3535, #3579, #3580, and #3581. Load `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, and website/design skills if editing `website/**` beyond generated data.

## GSD/TDD plan

Use `scripts/gsd doctor` and `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`; record fallback if `programming-loop` alias remains unavailable. Update phase PLAN/TDD/VERIFICATION before production edits.

Red evidence first:

- ledger completeness test/fixture fails until #3532/#3535 generated/unrelated paths have disposition;
- guard fixture fails until unrelated connector docs/defs/generated paths are rejected;
- generator-specific regression fails before removing/regenerating stale outputs.

## Acceptance criteria

- Every unrelated connector/generated/manual/website path listed for #3532 and #3535 has an explicit disposition.
- Unrelated connector manuals/skills/defs are removed or regenerated only when demonstrably stale/unrelated.
- Valid shared indexes and target outputs are preserved with evidence.
- Unrelated generated churn is rejected by the new guard in connector lanes.
- No history rewrite, no force-push, no blanket revert.
- Ledger cites PR URLs, merge SHAs, path classifications, disposition, and verification evidence.

## Verification

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
rg -n "zendesk-support|google-ads|gong|hubspot|bitbucket|xero|bahmni" docs/connectors website internal/connectors/defs
```

Run docs/website tests if editing non-generated website code.

## PR / review requirements

Sub-PR targets the parent branch and uses:

```markdown
Refs #3586
Refs #3579
```

Include GSD/TDD/skill evidence, no-mistakes validation, CI, automated review route, and worker handoff.

## Safety

No secrets, no credentialed connector checks, no new dependencies, no broad generated rewrites not named here, no push to `main`.
