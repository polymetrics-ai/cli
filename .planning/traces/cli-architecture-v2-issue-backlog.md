# CLI Architecture v2 GitHub Issue Backlog

**Prepared:** 2026-07-16
**Repository:** `polymetrics-ai/cli`
**Delivery model:** one parent issue and draft parent PR, stacked sub-issues/PRs, isolated Pi worktrees, GSD programming loop, strict TDD, human-gated merge to `main`

## Sources

- `docs/plans/cli-architecture-v2-improvement-plan.md`
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md`
- `docs/design/tui-ux-design.md`
- `docs/design/terminal-ui-research-and-design-system.md`
- `docs/adr/0002-cobra-viper-cli-framework.md`
- `docs/adr/0003-interactive-tui-layer.md`
- `docs/adr/0004-opentelemetry-observability.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`

## Parent

**Title:** `epic(cli): deliver CLI Architecture v2`
**Parent branch:** `feat/cli-architecture-v2`
**Parent PR:** draft, `feat/cli-architecture-v2` → `main`; merge remains human-gated.
**State before execution:** `planned`; Stage 0 creates the branch, planning scaffold, and draft parent PR.

**Live GitHub milestone:** [CLI Architecture v2](https://github.com/polymetrics-ai/cli/milestone/9)
**Live parent issue:** [#397](https://github.com/polymetrics-ai/cli/issues/397)

The parent orchestrator owns `.planning/ROADMAP.md`, `.planning/PROJECT.md`, the parent PR body,
the integration ledger, dependency scheduling, review coverage, and sub-PR merge arbitration.
Workers must not edit these shared artifacts unless the orchestrator explicitly delegates them.

## Issue roster and dependencies

| ID | Issue title | Track | Blocked by | Primary write scope |
|---|---|---|---|---|
| S0 | `docs(planning): register CLI Architecture v2 milestone and parent PR` | orchestration | — | `.planning/**`, parent branch/PR only |
| P01 | `test(cli): add golden transcript safety net` | A | S0 | `internal/cli/*golden*`, fixtures, docs-diff tests |
| P02 | `refactor(cli): introduce Cobra router shell` | A | P01 | root router/wrappers and focused CLI tests |
| P03 | `feat(config): add typed Viper configuration` | A | P02 | `internal/config/**`, config tests/docs |
| P04 | `refactor(config): migrate scattered environment reads` | A | P03 | runtimecheck, worker/schedule/RLM config call sites |
| P05 | `feat(events): add progress event bus and instrumentation` | B | P04 | `internal/events/**` and named instrumentation points |
| P06 | `feat(obs): add redacted per-run slog foundation` | C | P04 | logging/redaction packages and Temporal logger bridge |
| P07 | `feat(ui): add stdin+stdout TTY gate and NDJSON progress` | B | P05 | CLI run options, global flags, `internal/ui/styles/**` |
| P08 | `refactor(cli): nativize catalog namespace` | A | P04 | catalog command node/tests/docs only |
| P09 | `epic(cli): nativize remaining namespaces` | A | P08 | orchestration-only umbrella; implementation in grandchildren |
| D-TUI | `docs(ui): codify Bubble Tea terminal design research and interaction system` (#462) | B design gate | P07 | design docs, repo-local TUI skill, GSD/UI issue prompts; no production Go |
| P10 | `feat(ui): add flow and ETL run dashboards` (#408) | B | P07, D-TUI/#462 | flow/ETL dashboard models and command wiring |
| P11 | `feat(ui): add flow and schedule creation wizards` (#409) | B | P10, D-TUI/#462 | flow/schedule wizard commands, tests, parity docs |
| P12 | `feat(obs): add opt-in OpenTelemetry tracing` | C | P04 | `internal/telemetry/**` and allowlisted span call sites |
| P13 | `feat(ui): add connector browser and query grid` (#411) | B | P11, D-TUI/#462 | human-first bare query workspace, grid alias, tables/browser, and parity docs |
| P13C | `feat(ui): add read-only query charts and terminal dashboard compositions` (#463) | B | P13, D-TUI/#462 | bounded read-only chart models; renderer dependency remains human-gated |
| P14 | `feat(ui): add terminal docs viewer` (#412) | B | P11, D-TUI/#462 | docs viewer/pager and parity docs |
| P15 | `feat(cli): add connector-aware shell completion` | A | P09 | command registration/completion/tests/docs |
| P16 | `feat(ui): add certify and RLM dashboards` (#414) | B | P09, P10, D-TUI/#462 | certify/RLM dashboard models and command wiring |
| P17 | `feat(obs): add OpenTelemetry metrics` | C | P12 | metrics instruments/exporters/benchmarks |
| P18 | `feat(ui): add guided reverse ETL session` (#416) | B | P11, D-TUI/#462 | human-first bare reverse workspace plus guide alias; approval token stays hidden |
| P18B | `feat(ui): add TTY-progressive credential and connection setup` (#469) | B | P11, D-TUI/#462 | missing-field setup guidance; secret-source metadata only |
| P19 | `feat(cli): deepen help tree and generate man pages` (#417) | A | P13, P14, P15, P16, P18, P18B | help tree, generated manuals, docs/website/goldens |
| P20 | `feat(ui): complete accessibility audit and a11y topic` (#418) | B | P13, P13C after P13 when included, P14, P16, P18, P18B, D-TUI/#462 | all TUI accessibility fixes and parity docs |
| P21 | `feat(obs): add optional OpenTelemetry log bridge` | C | P06, P12 | optional pinned beta log bridge only |
| P22 | `chore(cli): complete Architecture v2 cleanup` | A | P17, P19, P20 | dead parser cleanup, `AGENTS.md`, `CONTEXT.md`, final verification |

P22 treats P21 as a recorded human decision: integrate it if approved and green, or record it as
explicitly skipped. It is not a hard GitHub dependency because the source plan defines it as optional.

## Phase 9 grandchildren

Phase 9 is an umbrella because the source plan requires one GSD loop and one coherent PR per
namespace. These tasks are deliberately serialized: they share central CLI routing/help files and
would collide if multiple Pi sessions edited them concurrently.

`connections → query → perf → runtime → version → skills → docs → agent → credentials → etl → reverse → flow → schedule → rlm → worker → extract → connectors/certify`

Live GitHub range: [#421](https://github.com/polymetrics-ai/cli/issues/421) through
[#437](https://github.com/polymetrics-ai/cli/issues/437), nested under Phase 9
[#407](https://github.com/polymetrics-ai/cli/issues/407).

Each grandchild owns only its namespace command node, handler adaptation, focused tests, and parity
docs. Connector dynamic dispatch remains on the legacy parser; certify migrates last.

## Parallel waves

1. **Bootstrap:** S0 → P01 → P02 → P03 → P04 (strictly serial).
2. **Foundation fan-out after P04:** P05, P06, P08, and P12 may run in separate worktrees.
3. **Track fan-out:** P07 follows P05; P09 follows P08; P10 follows P07; P17 follows P12; P21 follows P06+P12.
4. **TUI design gate:** integrate D-TUI/#462 before P10/#408, P11/#409, P13/#411,
   P14/#412, P16/#414, P18/#416, P18B/#469, P20/#418, or P13C/#463 starts production UI work.
   Parent orchestrator must update GitHub blocked-by metadata; this docs worker must not mutate
   GitHub issue metadata.
5. **UX fan-out:** after P11 and D-TUI/#462, P13, P14, P18, and P18B may run in separate worktrees
   after collision checks. P18 owns reverse only; P18B owns credentials/connections only. P16 can
   run after P09+P10+D-TUI/#462.
6. **Convergence:** P15 after P09; P19 and P20 after their listed UI/CLI dependencies; P22 last.

## Required worker prompt contract

Every Pi session receives exactly one issue URL, the parent issue/PR, its isolated worktree, allowed
write scope, dependencies, and these commands:

```text
/gsd doctor
/gsd plan-phase <issue-number> --skip-research
/gsd-programming-loop init --phase <issue-number> --dry-run
/gsd execute-phase <issue-number>
/gsd verify-work
/gsd-code-review <issue-number>
```

The worker must start with `golang-how-to`, then load the issue-specific Go skills listed in the
issue. CLI-visible work must include `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-documentation`, and `golang-security` when arguments, paths, credentials, or external I/O
are involved. Every TUI worker must load the repo-local `bubble-tea-tui-design` skill and cite both
terminal design documents. Website work loads the applicable web/React skills. Every behavior
change records red evidence before production edits.

## Common verification and safety contract

- Targeted tests first, then `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`,
  `go build ./cmd/pm`, and `make verify`.
- CLI parity covers runtime help, human-first dual-TTY bare query/reverse entry, deterministic
  bypass help, ordinary bare-namespace exit-0 behavior, `docs/cli/**`, `website/**`,
  generated manuals/help fixtures, completion metadata, and tests.
- TUI/Huh prompt activation requires both stdin and stdout TTYs and no `--plain`, `--json`, or
  `--no-input` bypass flag; future RED tests must cover `stdin-piped+stdout-TTY`, `stdout-piped`,
  `CI`, `--json`, `--plain`, and `--no-input` fallback without scripted-stdin consumption, hangs,
  or `/dev/tty` bypass. Bypass flags always skip Bubble Tea, Huh, and prompts; sequential prompts
  are allowed only in explicit accessible mode after the same gate passes.
- Sub-PRs target `feat/cli-architecture-v2` and use `Refs #<sub-issue>` plus `Refs #<parent>`.
- Dependency additions are allowed only in the phases and version lines approved by ADRs 0002–0004;
  any deviation is a fresh human gate.
- `ntcharts/v2` is a researched candidate, not an approved dependency. Chart child #463 must
  receive explicit human approval, pin/wrap the renderer, preserve table/text fallbacks, and test
  bounded data, axes/units, resize, no-color/ASCII, and accessibility before it enters `go.mod`.
- No secrets, interactive secret entry, generic shell, generic HTTP write, generic SQL write,
  destructive/admin execution, quality-gate reduction, production deployment, or merge to `main`.
- Reverse ETL retains plan → preview → approval → execute, and automated tests must not execute an
  unapproved write.

## Live issue ranges

- Parent: [#397](https://github.com/polymetrics-ai/cli/issues/397)
- Stage 0 and phases 1–22: [#398](https://github.com/polymetrics-ai/cli/issues/398) through
  [#420](https://github.com/polymetrics-ai/cli/issues/420)
- Phase 9 namespace grandchildren: [#421](https://github.com/polymetrics-ai/cli/issues/421)
  through [#437](https://github.com/polymetrics-ai/cli/issues/437)
- Terminal design gate: [#462](https://github.com/polymetrics-ai/cli/issues/462)
- Query chart child: [#463](https://github.com/polymetrics-ai/cli/issues/463)
- Parent PR: intentionally absent until Stage 0 #398 creates the parent branch and deliberate seed;
  no implementation issue is `worker_ready` before that draft PR exists.
