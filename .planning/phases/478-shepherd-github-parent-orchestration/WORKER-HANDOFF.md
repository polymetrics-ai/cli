## Worker Handoff

Sub-issue: #478

Parent issue: #471

Worker agent: Codex `gpt-5.6-sol` / high

Branch: `feat/478-shepherd-github-parent-orchestration`

Sub-PR: https://github.com/polymetrics-ai/cli/pull/487

Parent PR: #472

Base branch: `feat/471-pi-agent-session-shepherd`

Worker directory: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-478-shepherd-github-parent-orchestration`

Implementation head: `40ce66d4b5010b92089895a05709687143d15a05`

Verification evidence head: `568c98e2bf09ac751eb474df20cd37a5af3cbd70`

## Scope Delivered

- Typed, bounded, fakeable GitHub orchestration port for parent objectives, child issues, stacked
  PRs, rosters, exact child integration receipts, and parent ready transitions.
- Exact-shape authoritative checks, requested changes, review threads, review findings,
  dispositions, and exact-range independent Codex review evidence.
- Reuse of dependency graph scheduling, autonomy reconciliation, workspace handoff evidence, and
  the existing request/poll/consume decision broker without controller/session wiring.
- Retry-safe marker reconciliation, timeout-after-publish recovery, merged-child restart reuse,
  exact generation/head binding, and fail-closed parent human gating without merge capability.

## Files Changed

- `.pi/extensions/shepherd/github-orchestrator.ts` and matching test: orchestration domain/port.
- `.pi/extensions/shepherd/github-evidence.ts` and matching test: authoritative evidence policy.
- `.pi/extensions/shepherd/review-router.ts` and matching test: declarative independent review work.
- `.pi/extensions/shepherd/fixtures/issue-478/**`: bounded fake evidence/objective fixtures.
- `.planning/phases/478-shepherd-github-parent-orchestration/**`: plan, TDD, verification, and handoff.

## GSD / TDD / Skill Evidence

- GSD mode: `manual_gsd_fallback`.
- GSD command: `scripts/gsd prompt programming-loop init --phase
  478-shepherd-github-parent-orchestration --dry-run` returned `unknown GSD command:
  programming-loop`; `scripts/gsd doctor` passed.
- GSD adapter source: `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Required skills source: `.agents/agentic-delivery/references/required-skills-routing.md`.
- Required Go skills loaded: not applicable; this is a TypeScript-only bounded slice.
- Required design skills loaded: not applicable.
- Skills loaded: `gsd-programming-loop`, `github-issue-first-delivery`, `gsd-workstreams`,
  `architecture-patterns`, `javascript-testing-patterns`.
- Red test evidence: initial absent-module RED was 0 pass / 3 file failures; adversarial correction
  RED at `db9fbc33` was 17 pass / 10 expected failures with production unchanged at `90321ffb`.
- Green implementation evidence: focused 27/27 and strict owned TypeScript pass at `40ce66d4`.
- Refactor evidence: serialized Shepherd, strict production, offline RPC, and diff/base/scope pass.

## CLI Help / Docs / Website Parity

- Applies: no; no CLI command, flag, help, docs, website, or generated manual surface changed.
- Runtime help checked: not applicable.
- Bare namespace behavior checked: not applicable.
- `docs/cli/**` updated: not applicable.
- `website/**` updated: not applicable.
- Generated help/manual artifacts updated: not applicable.
- Parity exemptions: internal Pi extension orchestration boundary only.

## Verification

```bash
node --test .pi/extensions/shepherd/github-orchestrator.test.ts \
  .pi/extensions/shepherd/review-router.test.ts \
  .pi/extensions/shepherd/github-evidence.test.ts
node --test --test-concurrency=1 .pi/extensions/shepherd/*.test.ts
# strict TypeScript 5.9.3 with cached Pi 0.80.6 package/types
# pinned Pi 0.80.6 offline RPC get_commands discovery
git diff --check 3addb1f48be1afe8b1e2b59b54247679d7293805..HEAD
```

Result: pass. Focused 27/27; serialized Shepherd 290 pass, 0 fail, 1 intentional skip; strict owned
and all-production TypeScript pass; pinned offline RPC returns `true`; diff/base/scope pass.

## Automated Review

- Primary route: `codex_independent` using `openai-codex/gpt-5.6-sol:xhigh` (parent policy).
- Fallback route: none.
- Coverage route: parent-owned stable-head specialist campaign.
- Coverage status: pending.
- Review URL: pending.
- Disposition summary: no review was started by this worker.
- Unresolved findings: none known locally; exact-head independent coverage remains required.

## Merge Recommendation

- Recommended state: `provisional_parent_integration` after parent-owned review coverage.
- Reason: all authorized local gates pass; the exact-head automated review gate is intentionally
  still open.
- Human gates: parent ready, exact-head parent merge, and default-branch merge remain active.
- Follow-up issues: #479 owns controller/session integration; this worker must not add it.
