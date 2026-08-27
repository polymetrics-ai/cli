# Issue #4359 — Composite provider path identity foundation

## Task Delivery Header

- Issue: Refs #4359 — feat(engine): prove composite provider path identities
- Base branch: main
- Merges into: main
- Delivery: A normal PR from `fm/cli-composite-provider-path-foundation-r1` to `main`, with local verification recorded, exact-head fresh-context Codex review, and the GitHub API-reported base confirmed as `main`.
- Working branch: fm/cli-composite-provider-path-foundation-r1
- Task: Add the smallest fail-closed engine proof that admits CircleCI's cited `{project-slug}` identity only when it maps to precisely `vcs_type/org/repo` in that order. It must unblock Batch-1's eleven existing CircleCI command bindings without adding a generic template, URL, method, body, or provider-I/O escape.
- Verification: Red/green engine equivalence tests; focused engine, commandrunner, CLI, and App tests; source import/check, validation, surface-sync, declaration-admission, operation-evidence, connector-boundary, docs/website, and binary credential-boundary evidence for the eleven commands after their declarative Batch-1 shape is supplied as a test fixture.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| CircleCI's documented identity maps only to the three configured segments in the documented order | live | Engine tests resolve the cited provider endpoint to the existing three-segment transport and report a dedicated equivalence proof. |
| Every incomplete, reordered, conflicting, extra, absolute, cross-connector, or generic mapping is rejected | live | Table-driven engine tests assert endpoint resolution fails before a command can become preflight-reachable. |
| ETL and reverse-ETL remain credential-gated only through their existing commandrunner/App paths | live | Commandrunner/CLI/App tests preflight exact fixtures and assert no provider request occurs before `missing --credential`. |
| Other lanes and provider operations do not gain eligibility | live | The negative matrix covers direct read/write and binary lanes plus unrelated methods, bodies, origins, and paths. |
| The retained source is unchanged | live | `git diff --exit-code` against the source-lock path and source-import `--check` prove no lock byte changed. |

## Locked discovery inputs

- Immutable discovery/review SHA and `origin/main`: `b9b2478b3b2451d632d28b9aa138a170ad835110`.
- Batch-1 is read-only integration evidence, not an ancestor of this foundation branch: `5de7078bfbe2c21db9e200dafe29adfba9e0f91b` (PR #4294 head lineage).
- Retained CircleCI artifact: `https://circleci.com/api/v2/openapi.json`, OpenAPI `3.0.3`, SHA-256 `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`, 621321 bytes, source lock `circleci/sources/circleci-operation-source-lock.json` on the read-only Batch-1 tree.
- Locked provider identity: `/project/{project-slug}`. Existing source-backed runtime transport: `/project/{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}`. The documented component order is `vcs_type`, `org`, `repo`.
- The Batch-1 surface shows eleven affected non-duplicate bindings: `projects list`, `pipelines list`, `schedules list`, `checkout keys list`, `environment variables list`, `insights workflow summary list`, `create schedule apply`, `create environment variable apply`, `create checkout key apply`, `delete environment variable apply`, and `delete checkout key apply`. The first five reads and five writes use the literal `/project/` family; the eleventh uses the same source-documented `{project-slug}` identity in `/insights/{project-slug}/workflows`.
- `origin/main` intentionally has no CircleCI `cli_surface.json` or retained CircleCI source-lock directory. The source lock and concrete CLI surface remain read-only Batch-1 evidence. This foundation therefore introduces the declaration schema/engine proof and exercises the eleven paths with a declaration-shaped test bundle; it must not copy or absorb Batch-1's generated declaration/source files.

## Inline lifecycle fallback

The Pi/GSD adapter commands `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved with `scripts/gsd sources` and their prompts were generated. This disposable Codex worker cannot invoke compatible GSD-isolated project roles without violating the repository's single-worker contract, so the workflow is executed inline. The requested independent audit is a fresh-context Codex reviewer rather than the unavailable Claude audit and is recorded in `REVIEW-CONVERGENCE.md`.

## Required skills

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `connector-lane-build-order`, and `firstmate-exhaustive-review`.

## CLI help/manual/website parity

This changes the proof behind generated connector commands, not command grammar, flags, help text, manual copy, or website content. Existing generated connector docs and website data are checked for drift; runtime root/connector/command help is exercised after the Batch-1-shaped fixture is available. A source or surface generator must be run only if it changes an owned output.
