# Execution prompt — CLI Architecture v2 (pi + GSD)

You are executing the **CLI Architecture v2** program for the Polymetrics `pm` CLI:
cobra+viper adoption via strangler migration, an interactive TUI layer (events bus,
Bubble Tea v2, huh wizards, glamour docs), and observability (slog + opt-in OpenTelemetry).
GSD Universal Programming Loop discipline: strict TDD where new code is written (red test
evidence recorded in TDD-LEDGER.md before production code), commit each green slice so
progress is never lost, deterministic gates between stages. cwd = repo root. Do NOT push to
`main` — work on feature branches per phase with Conventional Commit titles and `Refs #N`
/ `Closes #N` in PR bodies. Dependency additions to go.mod are a human gate: the approvals
are recorded in `docs/adr/0002…0004`; add ONLY the modules named in the phase you are
executing and approved by the relevant ADR/human decision, nothing else, and `make tidy-check`
freezes go.mod again after each phase. `ntcharts/v2` is researched but NOT approved by
ADR-0003; do not add it in #463 without explicit human approval.

SOURCE-OF-TRUTH DOCUMENTS (read before starting, cite in plans):
- `docs/plans/cli-architecture-v2-improvement-plan.md` — the program: gaps, pillars, roadmap, risks.
- `docs/design/tui-ux-design.md` — the UX contract for every interactive surface (Pillar B).
- `docs/design/terminal-ui-research-and-design-system.md` — normative modes, responsive
  layout, chart grammar, reference decisions, and TUI verification matrix.
- `docs/adr/0002-cobra-viper-cli-framework.md`, `docs/adr/0003-interactive-tui-layer.md`,
  `docs/adr/0004-opentelemetry-observability.md` — decisions and their discipline rules.

CONTEXT (verified before you start):
- Dispatch is a hand-rolled switch in `internal/cli/cli.go:57-100` (~21 commands); bespoke
  flag parsing in `internal/cli/parse.go`; help strings in `internal/cli/docs.go`; error
  taxonomy → exit codes in `internal/cli/errors.go:107-127`; `cli.Run(args, stdout, stderr)`
  signature is FROZEN (certify seam: `certify.SetCLIRunFunc`, `cmd/pm/main.go:22`).
- Agentic contract enforced by `internal/cli/agentic_contract_test.go`: one JSON envelope
  per invocation on stdout, stderr for diagnostics, no ANSI (`internal/safety/safety.go:32`),
  deterministic. `pm <connector> <path…>` dynamic dispatch must keep arbitrary flag passthrough.
- No CLI framework, no config package (config.yaml is write-only), no event stream, no
  logging framework (Temporal gets noopLogger, `internal/worker/submit.go:99-106`), no
  telemetry. `.polymetrics/logs/` exists but is never written.
- CLI parity gate (AGENTS.md): runtime help, bare-namespace behavior, `docs/cli/**`,
  website docs, generated artifacts, and tests move together in every CLI-visible change.
- Verification gate: `make verify` (fmt, tidy-check, vet, test 20m, build, docs-check,
  smoke, lint, connectorgen-validate). Library targets: charm.land v2 line (bubbletea
  v2.0.8+, bubbles v2.1.1+, lipgloss v2.0.5+, huh v2.0.3+, glamour v2.0.1+), spf13
  cobra v1.10.x + viper current, otel v1.44.0/v0.66.0 train, evertras/bubble-table v0.22.3+.

STAGE 0 — MILESTONE REGISTRATION (planning only, no code)
1. In pi, run `/gsd new-milestone "CLI Architecture v2"` (or `scripts/gsd prompt new-milestone`
   non-interactively). Update `.planning/PROJECT.md` and `.planning/ROADMAP.md` with the
   22-phase roadmap from `docs/plans/cli-architecture-v2-improvement-plan.md` §5, preserving
   the existing connector-parity workstreams untouched.
2. Record in the milestone notes: the go.mod human-gate approvals (cite the three ADRs),
   the frozen constraints above, and that phases interleave across tracks A/B/C per plan §9.
GATE 0: `.planning/ROADMAP.md` shows the new milestone with 22 phases; `scripts/gsd doctor`
clean; no source files changed. Commit "docs(planning): register CLI Architecture v2 milestone".

PER-PHASE LOOP (applies to every stage below): run `/gsd plan-phase <N>` (research →
PLAN.md with TDD-ordered tasks) → `/gsd execute-phase <N>` → `/gsd verify-work`. Every
phase ends with `make verify` green plus the phase GATE. If a gate cannot pass, stop and
flag it loudly in SUMMARY.md — never fake or skip a gate.

TUI-PHASE PREFLIGHT (stages 10, 11, 13, chart child #463, 14, 16, 18A/#416, 18B/#469, and 20): issue #462/D-TUI
must be integrated into the active parent branch and its accepted correction PR/review blocker must
be resolved by the parent orchestrator before production UI work starts. Load the repo-local
`bubble-tea-tui-design` skill plus the routed Go CLI/testing/security/safety/context/concurrency/
documentation skills. The shared RED contract must cover the TUI fallback matrix for every TUI
phase: stdin+stdout TTY activation, `stdin-piped+stdout-TTY` fallback, `stdout-piped`, `CI`,
`--json`, `--plain`, and `--no-input`. `--plain`, `--json`, and `--no-input` must bypass Bubble
Tea, Huh, and all prompts; sequential prompts are allowed only in explicit accessible mode after the
same stdin+stdout TTY gate passes and no bypass flag is set. Retain each surface's own RED cases for
ordinary bare-namespace help behavior and the narrowly allowlisted human-first bare `pm query`/
`pm reverse` dual-TTY entry behavior, Normal/Filter/Edit mode conflicts, arrows+Vim equivalence,
contextual help, wide/standard/compact/guard layouts, accessible/plain fallback,
sanitation/redaction, approval-token non-display, cancellation, and unchanged
JSON/stdout/stderr/exit semantics. Record the skill and evidence in PLAN/TDD-LEDGER/VERIFICATION
and the PR body.

STAGE 1 — GOLDEN SAFETY NET (track A, no deps)
Record ~80 golden transcripts (exit code + stdout + stderr) against the CURRENT dispatcher:
all help paths (bare pm, --help, -h, help, man, per-namespace, --json manuals), flag-form
edges (--flag=v, --flag v, repeated flags, bare --flag, unknown-flag tolerance,
--root both forms, --json positioned late), error categories (unknown command exit 2,
unsafe identifier exit 3, `pm help nosuchtopic` exit 1), connector dispatch errors, hidden
commands (extract, worker). Add a docs-generate-diff test (`pm docs generate` into temp dir,
diff against `docs/cli/**`).
GATE 1: `go test ./internal/cli/ -run Golden` green; `make verify` green; `git diff go.mod` empty.
Commit "test(cli): golden transcript safety net for architecture v2".

STAGE 2 — COBRA ROUTER SHELL (track A; deps: cobra)
`go get github.com/spf13/cobra@v1.10.x`. Build `newRootCmd()` per ADR-0002: root with
ArbitraryArgs + DisableFlagParsing + SilenceErrors/Usage, persistent --root/--json
definitions, DisableFlagParsing wrappers for all 21 commands (extract/worker Hidden),
custom help/man commands, root RunE fallthrough to connector dispatch, mapCobraErr into
writeError, fresh tree per Run() call. Delete the switch; keep parseGlobal/parseFlags and
all handlers. cli.Run signature unchanged.
GATE 2: golden suite byte-identical; full `make verify`; certify smoke
(`go test ./internal/cli/ -run Certify`). Commit "refactor(cli): cobra router shell — strangler phase (arch-v2)".

STAGE 3 — CONFIG PACKAGE (track A; deps: viper)
`internal/config` per ADR-0002 §5: viper.New() instance inside config.Load, SetEnvPrefix
POLYMETRICS, explicit BindEnv with PM_* aliases, BindPFlag, SetConfigFile
(.polymetrics/config.yaml — its first reader), NO WatchConfig, Unmarshal into typed Config.
Table tests: precedence per key (flag > env > file > default), alias fallback, malformed
file → validation error exit 3.
GATE 3: `go test ./internal/config/...`; `make verify`; docs: config keys documented in
docs/cli + website (parity checklist). Commit "feat(config): typed config with viper instance mode (arch-v2)".

STAGE 4 — ENV MIGRATION (track A, no deps)
Move config-shaped os.Getenv reads onto Config: runtimecheck (add FromConfig, keep FromEnv
delegating), worker_cli temporalAddr, schedule/select.go + crontab.go (certify's
stages_glue.go PM_CRONTAB_FILE save/restore must keep working), agent_image_cli,
worker/submit, rlm_cli LLM config. Do NOT touch credentials --from-env or certify credsfile.
GATE 4: full test suite + goldens unchanged; `POLYMETRICS_INTEGRATION=1 go test ./...` if
runtime available. Commit "refactor(config): route env reads through config (arch-v2)".

STAGE 5 — EVENTS BUS (track B, no deps)
`internal/events` per ADR-0003 §1: Event struct, Emitter, context carriage (Nop default),
sinks Nop/NDJSON/Chan/Throttle/Multi (NDJSON sanitized via internal/safety). Instrument
flow.Engine.Run (beside appendLedger sites), app ETL flush(), certify.RunBatch workers,
worker/submit DescribeWorkflowExecution poller. Emission-sequence tests with a collector
emitter; -race over concurrent Emit.
GATE 5: `go test -race ./...`; `git diff go.mod` empty;
`go list -deps ./internal/events | grep -v '^internal\|std'` shows nothing external.
Commit "feat(events): progress event bus + instrumentation (arch-v2)".

STAGE 6 — SLOG FOUNDATION (track C, no deps)
Per ADR-0004 §1: RedactingHandler (key-based from SecretFields + value registry fed at
vault.Get + safety.RedactErrorText), fan-out to per-run
.polymetrics/logs/<run-id>.jsonl (ctx-routed) + stderr warn+; retention pruning; Temporal
noopLogger → tlog.NewStructuredLogger. RED first: redaction tests, run-routing tests, then
the end-to-end secret test (smoke flow with known token; grep logs — absent; test hook
proves the grep can fail).
GATE 6: `make verify`; `test -s .polymetrics/logs/run-*.jsonl` in a smoke dir;
secret grep clean; stdout envelope unchanged (`--json | jq .kind`). `git diff go.mod` empty.
Commit "feat(obs): slog foundation with redaction + per-run logs (arch-v2)".

STAGE 7 — STDIN+STDOUT TTY GATE + NDJSON PROGRESS (track B; deps: golang.org/x/term)
ui.Detect per ADR-0003 §2 (stdin TTY ∧ stdout TTY ∧ ¬json ∧ ¬plain ∧ ¬no-input ∧ ¬PM_NO_TUI ∧
¬CI ∧ TERM≠dumb); piped/non-TTY stdin always falls back to deterministic plain/noninteractive
behavior, is never consumed by prompts, and is never bypassed through `/dev/tty`;
cli.RunWithOptions (Run delegates ModePlain); global --plain/--no-input/--progress flags;
--progress ndjson wires events.NDJSON to stderr; internal/ui/styles foundation (palette
tokens, LightDark, glyph vocabulary + ASCII fallbacks per design doc §1).
GATE 7: Detect table tests must cover stdin+stdout TTY green path, `stdin-piped+stdout-TTY`
fallback, `stdout-piped`, `CI=1`, `PM_NO_TUI=1`, `TERM=dumb`, `--json`, `--plain`, and
`--no-input`; contract tests prove those fallback cases force plain/noninteractive behavior and
ndjson writes nothing to stdout beyond the envelope; `make verify`. Commit "feat(ui): stdin/stdout
TTY gate, plain/no-input flags, ndjson progress (arch-v2)".

STAGE 8 — NATIVIZE PILOT: catalog (track A)
Promote `catalog` to native cobra flags per ADR-0002 §4 conventions (StringArrayVar,
NoOptDefVal, FParseErrWhitelist). Custom SetHelpFunc/SetUsageFunc rendering docs map;
bare `pm catalog` exit 0 preserved.
GATE 8: golden diff empty (or reviewed fixture change); `make verify`.
Commit "refactor(cli): nativize catalog namespace (arch-v2)".

STAGE 9 — NATIVIZE REMAINING NAMESPACES (track A; several loops)
Order: connections, query, perf, runtime, version, skills, docs, agent, credentials, etl,
reverse, flow, schedule, rlm, worker, extract, connectors (certify subtree LAST, own loop).
One loop per namespace; delete that namespace's parseFlags call sites when done. Connector
dynamic dispatch stays on parseFlags forever.
GATE 9 (per loop): goldens unchanged; `make verify`; certify loop additionally runs the
certify smoke. Commits "refactor(cli): nativize <ns> namespace (arch-v2)".

STAGE 10 — RUN DASHBOARDS (track B; blocked by #462/D-TUI; deps: charm.land bubbletea/bubbles/lipgloss v2, teatest test-only)
Flagship per design doc §2.1: pipeline-rail model for flow run + etl run; inline mode;
ctrl+c → engine ctx cancel → DoneMsg → truthful final frame; exit codes identical to plain.
teatest goldens (happy/fail/cancel at 160×45, 100×30, 80×24, compact, size guard); command
wiring with Chan + runner goroutine. Use the operator-workspace hierarchy and persistent
mode/focus/help contract; throttle redraws without dropping lifecycle events.
GATE 10: `go test -race ./...`; goldens; manual TTY check; `pm flow run x | cat` byte-equal
to pre-TUI golden; `stdin-piped+stdout-TTY`, `stdout-piped`, `CI=1`, `--json`, `--plain`, and
`--no-input` all force plain/noninteractive behavior. Commit "feat(ui): live run dashboards for
flow/etl (arch-v2)".

STAGE 11 — WIZARDS WAVE 1 (track B; blocked by #462/D-TUI; deps: huh v2)
`pm flow create` per design doc §2.2 (kind-dependent huh groups, upstream-only in/out
pickers, rail preview, manifest round-trips flow.ParseManifest, prints sanitized scripted
equivalent) and `pm schedule create` interactive per §2.3 (flow existence validation, cron
presets + next-3-fire-times via schedule.Next, backend select, optional install).
WithAccessible wired to --accessible/PM_ACCESSIBLE_PROMPTER/ACCESSIBLE from day one.
Parity: new command docs (docs map, docs/cli, website, help tests). Printable Vim letters
must remain input while a field is focused; each step uses Gum-like focus and retains the
growing pipeline preview.
GATE 11: wizard-written manifests pass `pm flow plan`; accessible-mode transcript test;
`stdin-piped+stdout-TTY`, `stdout-piped`, `CI=1`, `--json`, `--plain`, and `--no-input` bypass Huh
prompts and name required flags/files; `make verify`. Commit "feat(ui): flow create + schedule
create wizards (arch-v2)".

STAGE 12 — TRACES (track C; deps: otel api/sdk + stdout/otlp trace exporters)
Per ADR-0004 §2-4: internal/telemetry Init/Shutdown (none/file/otlp; disabled = no SDK),
root pm.command span, pm.etl.run/pm.flow.step/pm.certify.* spans, pm.connector.http in
connsdk.Requester.do (path-only URLs, attempt events). Attribute allowlist test; exit-code
invariance test (unwritable telemetry dir → normal exit code, warn on stderr).
GATE 12: `PM_TELEMETRY=file` run produces spans JSONL with expected names; secret grep over
telemetry dir clean; `PM_TELEMETRY=off` → no dir; envelope-only stdout; `make verify`.
Commit "feat(obs): opt-in tracing with file/otlp exporters (arch-v2)".

STAGE 13 — BROWSE WAVE (track B; blocked by #462/D-TUI; deps: evertras/bubble-table)
`pm query tables` (plain/JSON warehouse enumerator — lands FIRST), connectors browser per
design doc §2.5 (fuzzy list + manual preview + pager; fix the %+v dump on the plain path),
the human-first query workspace per §2.4 (LIMIT/OFFSET paging through existing QuerySQL guard).
Eligible dual-TTY bare `pm query` opens it; `pm query grid` remains an explicit alias. Help flags
and bypass/non-TTY bare invocations render contextual help and exit 0. Use fzf's filter/list/preview
interaction without shell-backed preview execution. Query grid export is typed read-only only:
project-scoped default, clean/confined path, reject control characters/traversal/broad paths/
symlink races, no overwrite by default, confirmation only when stdin and stdout are TTYs,
noninteractive `--output` plus `--force`, sanitized command echo, and exact `--no-input` guidance.
A query chart is a separate
dependency-gated child issue #463 after #411 and #462/D-TUI: it may visualize only the returned
read-only rows; must preserve table/text access, axes/units/exact values, bounded deterministic
sampling, no-color/accessibility fallbacks; and may use `ntcharts/v2` only after explicit human
approval and an exact pinned wrapper. Otherwise use the minimal internal renderer.
GATE 13: query tables golden + JSON envelope; browser/grid teatest goldens; read-only export/path
guard tests; modal-key/resize/chart-bounds tests as applicable; `stdin-piped+stdout-TTY`,
`stdout-piped`, `CI=1`, `--json`, `--plain`, and `--no-input` fall back without consuming scripted
stdin; `make verify`. Commit "feat(ui): query tables, connectors browser, query grid (arch-v2)".

STAGE 14 — DOCS VIEWER (track B; blocked by #462/D-TUI; deps: glamour v2)
`pm docs view [topic|connector]` per design doc §2.6: glamour in viewport, auto light/dark,
piped → plain text identical to `pm help`. Use Normal/Search/Help modes; `/`, `n/N`,
`ctrl+u/d`, `gg/G`, arrows, and one-layer `esc` follow the shared contract. Parity checklist
for the new subcommand.
GATE 14: teatest goldens; piped-output golden equals help text; `stdin-piped+stdout-TTY`,
`stdout-piped`, `CI=1`, `--json`, `--plain`, and `--no-input` bypass the pager; `make verify`.
Commit "feat(ui): glamour docs viewer (arch-v2)".

STAGE 15 — COMPLETION (track A)
Hidden per-connector registered commands (still DisableFlagParsing) + root
ValidArgsFunction; credential/connection completion (silent NoFileComp on missing project);
un-hide cobra completion command. Full parity checklist (docs map topic, docs/cli/completion.md,
website).
GATE 15: `pm __complete "" ""` snapshot test; goldens; `make verify`.
Commit "feat(cli): shell completion + connector command registration (arch-v2)".

STAGE 16 — CERTIFY + RLM DASHBOARDS (track B; blocked by #462/D-TUI)
Certify batch table per design doc §2.7 (concurrent row updates from events; exit contract
0/1/2/3 untouched); RLM viewer per §2.8 (heartbeat age from the Temporal poller). Pair
bpytop-style exact metrics with any small graph, disclose units/ranges, throttle redraws,
and preserve text-only/reduced-motion frames.
GATE 16: teatest goldens; certify exit tests green; shared TTY fallback RED matrix green for
stdin+stdout TTY activation, `stdin-piped+stdout-TTY`, `stdout-piped`, `CI=1`, `--json`,
`--plain`, and `--no-input`; `make verify` + certify smoke.
Commit "feat(ui): certify batch table + rlm agent viewer (arch-v2)".

STAGE 17 — METRICS (track C; deps: otel sdk/metric + exporters + temporal contrib)
PRD §15.2 instruments per ADR-0004 §5; RunCounters accumulate-then-flush (NO per-record
instrument calls — benchmark guard: emit path allocs not regressed); Temporal tracing
interceptor + metrics handler gated on enabled.
GATE 17: file-exporter metrics reconcile with envelope counts (jq equality test);
`go test -bench BenchmarkEmit -benchmem ./internal/app` no regression; `make verify`.
Commit "feat(obs): metrics per PRD §15.2 + temporal otel contrib (arch-v2)".

STAGE 18A — GUIDED REVERSE ETL (#416; track B; blocked by #462/D-TUI)
Human-first reverse workspace per design doc §2.9: eligible dual-TTY bare `pm reverse` opens it and
`pm reverse guide` remains an explicit alias. Preserve the existing plan → preview → approval →
execute seam, typed confirmation, and exit semantics. Carry the approval token only ephemerally in
memory; never display it in final frames, transcripts, logs, screenshots, accessibility output,
JSON, shell-equivalent command text, or fixtures. Help flags and bypass/non-TTY bare `pm reverse`
render contextual help and exit 0. GATE 18A: both entries reach the same model, session tests prove
parity with the flag flow, and every noninteractive bypass avoids
prompt initialization and scripted-stdin consumption; secret-marker greps and `make verify` pass.
Commit "feat(ui): add guided reverse-etl session (arch-v2)".

STAGE 18B — PROGRESSIVE SETUP (#469; track B; blocked by #409 and #462/D-TUI)
Implement the approved Phase 18 UI contract for `pm credentials add [name]` and
`pm connections create [name]`. These setup namespaces remain contextual help. Incomplete actions launch
Bubble Tea/Huh by default only when stdin and stdout are TTYs and no bypass applies; fully specified
actions run directly; complete-but-invalid input returns the normal field-specific error. Ask only
for missing fields. Credential setup accepts non-secret config and secret-source metadata only:
named `--from-env` selections or a sanitized `--value-stdin` handoff, never plaintext secret input.
Connection setup derives capability-compatible choices, shows a final review, and recovers from a
duplicate with inspect, rename, or cancel—never overwrite. Agent docs use `--json --no-input` and
add `--progress ndjson` only for long-running commands; do not add a global `--agent-mode`.
GATE 18B: RED/GREEN/refactor evidence covers complete/incomplete × dual-TTY/bypass matrices,
reader-spy non-consumption, schema/placeholders, secret markers, capability choices, duplicate race,
responsive/accessibility frames, and help/manual/docs/website parity; `make verify` passes. Commit
"feat(ui): add progressive credential and connection setup (arch-v2)".

STAGE 19 — HELP TREE + MAN PAGES (track A; THE deliberate help-churn phase)
Per-subcommand focused --help; `pm docs man` via cobra/doc.GenManTree. Regenerate goldens,
`docs/cli/**`, website mdx + `node website/scripts/gen-docs-data.mjs`, help tests — full
parity checklist in the PR body.
GATE 19: `make verify`; docs diff reviewed; goldens regenerated deliberately.
Commit "feat(cli): help tree deepening + generated man pages (arch-v2)".

STAGE 20 — ACCESSIBILITY AUDIT (track B; blocked by #462/D-TUI)
`pm a11y` help topic per design doc §3; audit pass: every TUI surface against the
checklist (accessible mode, reduced motion, NO_COLOR, glyph+word pairing, min-size,
keyboard nav, modal printable-key conflicts, arrows+Vim equivalence, chart text/table
fallback, wide/standard/compact/guard layouts); fix findings.
GATE 20: a11y topic in docs map + docs/cli + website; checklist items each have a test or
a recorded manual verification; `make verify`. Commit "feat(ui): a11y topic + accessibility audit (arch-v2)".

STAGE 21 — OTEL LOG BRIDGE (track C; deps: otel log beta, pinned; OPTIONAL — skippable by
human decision without affecting anything else)
otelslog handler branch inside RedactingHandler, otlp-only; trace/span ID correlation test.
GATE 21: in-memory exporter test shows redacted records with matching trace_id;
file/none modes untouched; `make verify`. Commit "feat(obs): otel log bridge (arch-v2)".

STAGE 22 — CLEANUP
Delete dead parse code (parseGlobal if unused, wrapper generator, legacy help interception),
update AGENTS.md (dispatcher is cobra; how to add a command: docs map entry + cobra node +
parity checklist) and CONTEXT.md (new flags: --plain/--no-input/--progress/--accessible).
GATE 22 (FINAL): `make verify` + `make verify-duckdb` if CGO available; goldens green;
`grep -rn "parseGlobal" internal/` returns only sanctioned uses; ADR statuses confirmed.
Commit "chore(cli): architecture v2 cleanup + docs (arch-v2)".

REPORT at the end of every stage (in SUMMARY.md and the PR body): what was implemented,
TDD ledger reference (red evidence per behavior), gate outputs verbatim, parity-checklist
status, any deviation from the source-of-truth docs with justification. At program end:
a final report enumerating each phase's status, the full go.mod delta vs the ADR budgets,
and any constraint you could not preserve — flag it loudly instead of faking it.
