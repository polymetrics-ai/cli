# ADR 0002 — Adopt cobra + viper via strangler migration

- Status: Accepted (2026-07-16)
- Deciders: user (approved plan; explicitly chose full viper over a thin config package)
- Context docs: `docs/plans/cli-architecture-v2-improvement-plan.md` (Pillar A),
  `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md` §9/§16/§17,
  `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md`

## Context

The `pm` CLI routes ~21 top-level commands and ~60+ subcommand paths through a hand-rolled
`switch` (`internal/cli/cli.go:57-100`) with a bespoke untyped flag parser
(`internal/cli/parse.go:47-67`: bare `--flag` → `"true"`, unknown flags silently collected,
repeated flags accumulate). Help is hand-authored man-page strings (`internal/cli/docs.go`)
that also generate `docs/cli/**`; there is no shell completion, no per-subcommand help, no
man pages. Configuration is ~11 ad-hoc `os.Getenv` reads across six packages under two
prefixes (`POLYMETRICS_*`, `PM_*`), and `.polymetrics/config.yaml` is written by `pm init`
but read by nothing. The PRD (§17) has always called for "a CLI framework with strong
subcommand/help support". Hard constraints: the agentic contract (JSON envelopes, exit-code
taxonomy `internal/cli/errors.go:107-127`, no ANSI, bare-namespace-help-exit-0), the frozen
`cli.Run(args, stdout, stderr) int` signature (certify drives the CLI in-process and
recursively), the dynamic `pm <connector> <path…>` dispatch with arbitrary flag
passthrough, the CLI parity gate (help/docs/website/tests move together), and the go.mod
human gate.

## Decision

1. **Adopt `spf13/cobra` as the router via a strangler migration.** Phase one wraps every
   existing handler in a `DisableFlagParsing: true` cobra command so behavior is
   byte-identical; the root command uses `ArbitraryArgs` + a `RunE` fallthrough to preserve
   dynamic connector dispatch and exact legacy errors. A fresh command tree is built per
   `Run()` invocation (certify re-entrancy; no shared `Command` state). Namespaces are then
   promoted to native cobra parsing one GSD loop at a time (pilot: `catalog`; certify
   subtree last; connector dispatch never migrates off the legacy `parseFlags`).
2. **The existing error funnel stays the only exit-code authority.**
   `SilenceErrors/SilenceUsage` on the root; a `mapCobraErr` shim classifies cobra/pflag
   errors into the existing taxonomy feeding `writeError`.
3. **The docs map stays the single help source** until the deliberate help-tree phase:
   custom `SetHelpFunc`/`SetUsageFunc` render the existing manual text, honor the `--json`
   `CommandManual` envelope, and keep bare-namespace help exiting 0. Cobra boilerplate
   (`Usage:`, `Available Commands:`) must never leak.
4. **Flag-promotion conventions**: `StringArrayVar` for repeated flags (never
   `StringSliceVar` — comma-splitting), `NoOptDefVal="true"` where bare `--flag` means
   true, `FParseErrWhitelist{UnknownFlags: true}` to preserve legacy tolerance until a
   separately documented tightening.
5. **Adopt `spf13/viper` in instance mode** for a new `internal/config` package:
   `viper.New()` inside `config.Load` (never the package-level singleton),
   `SetEnvPrefix("POLYMETRICS")`, explicit `BindEnv` per key with `PM_*` legacy aliases as
   fallbacks (no `AutomaticEnv`), `BindPFlag` for global flags,
   `SetConfigFile(.polymetrics/config.yaml)` — the file's first-ever reader — **no
   `WatchConfig`** (deterministic CLI), `Unmarshal` into a typed `Config` struct injected
   down the call tree. Precedence: flags > env > file > defaults. Scattered `os.Getenv`
   config reads migrate to it; `credentials add --from-env` and certify's credsfile keep
   raw `os.Getenv` (user-named data plumbing, not configuration).
6. **A golden-transcript safety net precedes any cobra code**: ~80 recorded invocations
   (exit code + stdout + stderr) across help paths, flag-form edge cases, JSON envelopes,
   error categories, and connector dispatch — kept green through every phase; intentional
   changes are reviewed fixture diffs.
7. **UX additions once native**: shell completion (`ValidArgsFunction` for
   connector/credential/connection names; silent `ShellCompDirectiveNoFileComp` when no
   project exists), typo suggestions appended to the legacy unknown-command message,
   `cobra/doc.GenManTree` behind a new `pm docs man`, and per-subcommand `--help`.

## Alternatives considered

- **Keep the hand-rolled dispatcher**: rejected — every new command re-implements
  parsing/help/errors; no completion or help tree; the PRD gap persists.
- **Big-bang rewrite onto cobra**: rejected — a single PR touching 60+ paths, all help
  strings, docs/cli, and the website cannot be reviewed against the parity gate; one pflag
  semantic difference blocks everything.
- **Legacy switch primary, cobra as fallback**: rejected — two dispatchers to keep in
  parity while cobra contributes nothing until the end.
- **cobra without viper (thin `internal/config`)**: viable and lighter (viper brings ~10
  transitive modules; config.yaml is currently write-only), but the user explicitly chose
  the standard cobra+viper pairing; the instance-mode discipline above removes viper's
  determinism risks (globals, file watching, ambient env).
- **urfave/cli or kong**: rejected — smaller ecosystems for completion/man-page tooling;
  cobra's `DisableFlagParsing` wrapper is what makes the byte-identical strangler cheap.
- **charmbracelet/fang** (styled cobra wrapper): rejected outright — lipgloss-styled help
  and errors violate the no-ANSI agentic contract enforced by
  `internal/safety/safety.go:32` and `internal/cli/agentic_contract_test.go`.

## Consequences

- (+) Completion, focused sub-help, man pages, and typo suggestions arrive without
  touching handler logic; new commands (`flow create`, `docs view`, `query tables`) get a
  standard home.
- (+) Configuration becomes one typed struct with documented precedence; config.yaml
  becomes useful; the `PM_*`/`POLYMETRICS_*` split is resolved by aliasing, breaking no one.
- (+) The golden harness makes the parity gate mechanical instead of reviewer-heroic.
- (−) go.mod grows by cobra (+pflag, mousetrap) and viper (+afero, cast, fsnotify, gotenv,
  toml, mapstructure, conc) → accepted through the human gate this ADR records; tidy-check
  freezes it again afterward.
- (−) Two parsing regimes coexist during phases 2–9 → mitigated by the wrapper generator
  being ~30 lines, per-namespace promotion loops, and the golden suite.
- (−) pflag error wording differs from legacy messages when namespaces go native →
  category/exit-code pinned by `mapCobraErr`; wording changes are fixture-reviewed.
