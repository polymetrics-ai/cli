# CLI Architecture v2 Cobra/Viper release split

Branch: `fm/cli-architecture-v2-release-split-r1`
Target: `main`
Primary architecture issue: #397
Foundation issues: #399, #400, #401, #402
Release-safety issue: #453

## GSD path

- `scripts/gsd doctor` passed on 2026-07-24.
- `scripts/gsd prompt programming-loop init --phase cli-architecture-v2-release-split-r1 --dry-run` returned `unknown GSD command: programming-loop`; the documented manual-GSD fallback is active.
- Source plan: the captain-approved release-split report at `data/cli-architecture-v2-release-split-r1/report.md` in the private task workspace.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `golang-documentation`, `golang-design-patterns`, `golang-spf13-cobra`, `golang-spf13-viper`, `golang-dependency-management`, and `golang-continuous-integration`.

## Objective

Reconstruct the dependency-closed Cobra + typed, invocation-scoped Viper configuration foundation on exact latest `origin/main`, include reverse-smoke preview safety, and open a review-ready PR to `main` without TUI, events, logging, or OpenTelemetry code.

## Exact base and source provenance

Base at branch creation:

```text
873cd7b251f70c4a35a607a0d4e86051ea0fbd15
```

Apply source squashes in this order, adapting their patches to current `main`:

1. `379cb5015335ff7c9b20e5bb780952ead22c53b2` — #399 / PR #439 golden/docs safety.
2. `8900db141cc289b65491365d2ebcab490af57789` — #400 / PR #440 Cobra strangler shell.
3. `7683087d41646c92b2bd7f677f47cf2bc9d88462` — #401 / PR #441 typed invocation-scoped Viper config.
4. `cc2a90e918b2814a64516d6bad6d14462b3ac079` — #402 / PR #448 config consumer/env migration.
5. `20475ddf8ae3486282ead4fc7d2129f2bd1129b3` — #453 / PR #454 reverse smoke plan -> preview -> approval -> execute safety.

The historical SHAs are patch/provenance sources, not ancestors to preserve by rewriting history.

## Approved dependency boundary

Only the audited Cobra/Viper dependency set is authorized:

- direct: `github.com/spf13/cobra v1.10.2`, `github.com/spf13/viper v1.21.0`;
- their audited transitive module entries as selected by the source patches and `go mod tidy`.

No Charm, Bubble Tea, OpenTelemetry, `golang.org/x/term`, or any other dependency may be added.

## Implementation slices

1. **Planning checkpoint:** establish this plan, TDD ledger, verification checklist, prompt snapshot, and run state before production edits.
2. **Red/baseline checkpoint:** record latest-main absence of the config package/config help and preserve current-main CLI transcripts for deterministic compatibility comparison.
3. **Foundation reconstruction:** apply the five source squashes in order. Resolve only the known `internal/cli/cli.go` current-main/Gong conflict while preserving current Gong behavior.
4. **Latest-main adaptations:** adapt Cobra help routing to current `runHelp(..., jsonOut)`, remove the obsolete assumption that GitHub lacks a dynamic connector surface, and regenerate only affected golden/docs outputs.
5. **Governance/docs:** add ADR 0002 and an additive, truthful release-split record. Do not import the historical bootstrap commit or post-foundation architecture state.
6. **Verification:** run focused tests, exact CLI transcript comparison, docs/golden freshness, module gates, formatting/lint/vet/full tests/build/race where practical, `make verify`, security/dependency checks, and allowlist checks.
7. **Review/delivery:** use the canonical PM exact-version Codex review when the PM review system is clean, then independent Shepherd after a clean synthesis. Run `/no-mistakes`/`no-mistakes axi` through a green open PR. Do not merge.

## Mandatory behavior contracts

- Preserve `cli.Run(args, stdout, stderr) int`, re-entrant in-process execution, dynamic connector passthrough, stdout/stderr bytes, exact JSON envelopes, and exit codes for unchanged commands.
- Preserve non-TTY behavior and isolated worker behavior.
- Keep credential values outside typed config; never print or store secrets.
- Config precedence is: changed flag > explicit `POLYMETRICS_*` > legacy `PM_*` > `<effective-root>/.polymetrics/config.yaml` > default.
- Use a fresh `viper.New()` per invocation; no global Viper instance, `AutomaticEnv`, watcher, or config mutation.
- Keep reverse ETL plan -> preview -> approval -> execute ordering.
- Do not execute credentialed connector/runtime checks or external reverse ETL.

## Explicit exclusions

Exclude all of:

```text
internal/events/**
internal/logging/**
internal/telemetry/**
internal/ui/**
docs/adr/0003-interactive-tui-layer.md
docs/adr/0004-opentelemetry-observability.md
docs/design/tui-ux-design.md
```

Also exclude all TUI/event/logging/telemetry phase artifacts and dependencies, PR #493-owned delivery-skill/routing surfaces, PM review-system implementation files, the historical bootstrap commit, and unrelated generated artifacts.

## CLI help/docs/website parity

- Verify `pm help config`, root help, `pm --help`, bare namespace commands, invalid actions, JSON help, and dynamic connector help.
- Keep `docs/cli/config.md`, `docs/cli/connectors.md`, `website/content/docs/cli-reference.mdx`, generated website data, and golden transcripts synchronized only where the candidate changes them.
- Preserve existing stdout/stderr/JSON/exit behavior outside documented config activation.

## Commit and delivery checkpoints

- Commit the plan before production edits.
- Commit each coherent green reconstruction/adaptation slice.
- Push only through the authorized no-mistakes delivery pipeline after local gates.
- Never push `main`, rewrite published history, alter `feat/cli-architecture-v2`, merge PR #438, or publish a release before all authorized gates.

## Human gates and external waits

- The captain already approved the audited Cobra/Viper dependency set and implementation/PR creation.
- No merge to `main` is authorized.
- No parent-branch reconciliation, history rewrite, or parent PR merge is authorized.
- No prerelease publication is allowed until no-mistakes, required CI/security/Snyk, and exact-version PM review gates pass.
- If release version is not mechanically unambiguous, stop after code/PR readiness with concrete version options.
- If the PM review system is not clean/integrated, preserve the exact committed candidate and report a bounded external wait; do not substitute Claude, Copilot, or ad hoc review.
