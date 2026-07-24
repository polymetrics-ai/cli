# ADR 0002 — Adopt Cobra + Viper via strangler migration

- Status: Accepted; foundation release candidate (2026-07-24)
- Deciders: user (approved the architecture plan and the audited Cobra/Viper dependency set)
- Foundation issues: #399, #400, #401, #402
- Release-safety issue: #453
- Release-split record: `docs/architecture/cli-architecture-v2-release-split.md`

## Context

Before this foundation, `pm` routed its top-level commands through a hand-written switch and used a
bespoke untyped flag parser. Help came from hand-authored man-page strings that also generated
`docs/cli/**`. Configuration was spread across direct environment reads under both
`POLYMETRICS_*` and `PM_*`, while the `.polymetrics/config.yaml` written by `pm init` was not read.

The migration must preserve the existing agentic and Unix contracts:

- exact JSON envelopes, exit-code taxonomy, and stdout/stderr ownership;
- no ANSI output;
- bare namespace help with exit 0 while invalid actions remain usage errors;
- `cli.Run(args, stdout, stderr) int` and re-entrant in-process execution;
- dynamic `pm <connector> <path...>` dispatch with legacy flag passthrough;
- credential handling and reverse ETL plan -> preview -> approval -> execute;
- CLI help, manual, website, generated-artifact, and test parity.

## Decision

1. **Adopt `spf13/cobra` as a strangler router.** The foundation wraps existing handlers in
   `DisableFlagParsing: true` commands. The root uses `ArbitraryArgs` plus a `RunE` fallback for
   dynamic connector dispatch. A fresh command tree is built for every `Run` invocation. Native
   namespace promotion remains separate follow-up work.
2. **Keep the existing error funnel authoritative.** Cobra runs with `SilenceErrors` and
   `SilenceUsage`; `mapCobraErr` maps framework parse errors into the existing error taxonomy before
   `writeError` determines JSON shape and exit code.
3. **Keep the existing docs map authoritative.** Custom help and usage functions render the existing
   manuals, including JSON `CommandManual` envelopes. Cobra boilerplate must not leak into output.
4. **Preserve legacy parsing until a command is deliberately promoted.** Repeated values, bare
   booleans, unknown-flag tolerance, and dynamic connector flags continue to use existing semantics.
5. **Use Viper only in invocation-scoped instance mode.** `internal/config.Load` creates
   `viper.New()` for each invocation, binds an explicit allowlist of `POLYMETRICS_*` variables with
   `PM_*` aliases, binds only changed global `root` and `json` flags, reads
   `<effective-root>/.polymetrics/config.yaml`, and unmarshals into typed configuration. It does not
   use the package singleton, `AutomaticEnv`, `WatchConfig`, or a file watcher.
6. **Use this precedence exactly:** changed flag > explicit `POLYMETRICS_*` > matching legacy
   `PM_*` > `<effective-root>/.polymetrics/config.yaml` > built-in default. A root provided by a
   changed flag or explicit environment variable determines config-file discovery. A `root` inside
   that file does not relocate the same read.
7. **Keep secrets outside typed configuration.** User-selected `credentials add --from-env`,
   provider API keys, and certification credential seams remain direct data/secret plumbing. Typed
   config contains only non-secret runtime, RLM, scheduling, project, and warehouse settings.
8. **Require golden transcript safety before router changes.** The tracked suite records exit code,
   stdout, and stderr across help, flag forms, JSON envelopes, error categories, and connector
   dispatch. Intentional behavior changes require reviewed fixture diffs.

## Foundation release boundary

The default-branch candidate is reconstructed from latest `main` using these reviewed source
squashes in order:

```text
379cb5015335ff7c9b20e5bb780952ead22c53b2  # #399 / PR #439 golden safety
8900db141cc289b65491365d2ebcab490af57789  # #400 / PR #440 Cobra shell
7683087d41646c92b2bd7f677f47cf2bc9d88462  # #401 / PR #441 typed Viper config
cc2a90e918b2814a64516d6bad6d14462b3ac079  # #402 / PR #448 config consumers
20475ddf8ae3486282ead4fc7d2129f2bd1129b3  # #453 / PR #454 reverse smoke safety
```

This release boundary excludes events, logging, TUI, OpenTelemetry, native namespace promotions,
and their dependencies. It does not claim that CLI Architecture v2 is complete.

## Alternatives considered

- **Keep the hand-written dispatcher:** rejected because every new command would continue to
  duplicate parsing, help, and error behavior.
- **Big-bang Cobra rewrite:** rejected because command, help, docs, and flag-semantic changes would be
  too broad to review and roll back safely.
- **Legacy switch primary with Cobra as fallback:** rejected because it creates two primary routers
  without moving ownership.
- **Cobra without Viper:** viable but not selected; the approved design uses the standard pair under
  strict instance-scoped and explicit-environment constraints.
- **Styled Cobra wrappers:** rejected because ANSI/styled help would violate the machine-readable and
  non-TTY contracts.

## Consequences

- (+) Command ownership can move to Cobra incrementally without changing current handlers.
- (+) Configuration has one typed, testable precedence model and `config.yaml` becomes active.
- (+) A fresh router and Viper instance preserve in-process isolation.
- (+) Golden transcripts make compatibility review mechanical.
- (-) The module adds Cobra, Viper, and their audited transitive dependencies.
- (-) Legacy and framework parsing coexist until later namespace migrations.
- (-) Existing malformed config files that were previously ignored now fail through the validation
  error path; release notes must call out config activation.
