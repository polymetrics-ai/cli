## Objective

Make credential and connection setup intuitive for humans while preserving a deterministic,
prompt-free API for agents, scripts, pipes, and CI.

## Background

This is a scoped child of #416 (Phase 18 / Track B) under parent program #397. The original Phase
18 issue combined the security-critical human-first `pm reverse` workspace (with
`pm reverse guide` as its explicit alias) with credential/connection setup. This
child owns only the setup journey so it can be delivered as one issue and one stacked PR.

The approved interaction contract is being added by PR #468 in
`.planning/phases/416-guided-reverse-connection-prompts/18-UI-SPEC.md`.

## Scope and write ownership

- `pm credentials add [name]`: on an eligible terminal, guide the user through only missing name,
  connector, auth-mode, non-secret config, and secret-source metadata.
- `pm connections create [name]`: on an eligible terminal, derive valid credential, connector,
  stream, sync-mode, cursor, primary-key, and table choices from existing service metadata.
- Reuse the Phase 7 UI detector and the approved Bubble Tea v2/Bubbles v2/Lip Gloss v2/Huh v2
  foundation. Do not add or change dependencies in this issue.
- Keep runtime help, focused subcommand help, `docs/cli/**`, website docs, generated manuals,
  completion/discovery metadata, golden transcripts, and tests in parity.

Do not edit shared parent orchestration artifacts unless the parent orchestrator explicitly
delegates them.

## Activation and command contract

- Bare `pm credentials` and `pm connections` render contextual help and exit 0; they never launch a
  wizard.
- A fully specified action invocation executes directly, even on a TTY.
- An incomplete `credentials add` or `connections create` launches guidance by default only when
  both stdin and stdout are TTYs, `TERM` is not `dumb`, and `--json`, `--plain`, `--no-input`,
  `PM_NO_TUI`, and `CI` do not select the deterministic path.
- `--json`, `--plain`, `--no-input`, CI, pipes, and non-TTY invocation never prompt. Missing input
  produces one actionable usage/validation error with the exact flag or safe next command.
- Agent/automation documentation uses the existing profile `--json --no-input`; long-running
  commands may additionally use `--progress ndjson`. Do not add a global `--agent-mode`:
  `pm query run --agent-mode summary|stream` remains query-result shaping only.

## Acceptance criteria

- [ ] Incomplete action commands launch the guided flow only under the exact dual-TTY gate and ask
      only for missing fields.
- [ ] Complete action commands bypass the wizard and preserve current flag-path output, exit, and
      side-effect semantics.
- [ ] `--json`, `--plain`, `--no-input`, `PM_NO_TUI`, `CI`, `TERM=dumb`, piped stdin, and piped
      stdout never initialize Bubble Tea/Huh or consume unexpected input.
- [ ] Connector schemas drive required non-secret config and secret field names; missing GitHub
      owner/repo and literal documentation placeholders fail before network activity with
      field-specific validation guidance.
- [ ] Credential guidance never asks for or holds plaintext secrets. It selects only named
      `--from-env field=ENV` metadata or emits a sanitized `--value-stdin field` handoff command.
- [ ] `Save and test` distinguishes “saved and valid” from “saved but validation failed” without
      displaying, logging, or rolling back secret material.
- [ ] Connection choices are capability-aware; stream, sync mode, cursor, primary key, and table
      defaults are visible in a final review.
- [ ] Duplicate credential/connection names are user-correctable validation failures, never
      `internal_error`; no state is overwritten. Connection recovery offers only read-only inspect,
      choose another name, or cancel.
- [ ] Vim and arrow navigation, explicit Normal/Edit/Confirm/Help modes, one-layer `Esc`, `Ctrl+C`
      cancellation, responsive layouts, contextual help, no-color/ASCII, and accessible sequential
      prompting match the approved design contract.
- [ ] Human completion frames teach a sanitized equivalent flag command; agent docs show
      `--json --no-input` and safe secret-source examples.
- [ ] Runtime help, bare namespaces, focused `--help`, `docs/cli/**`, website docs, generated
      help/manual artifacts, completion/discovery metadata, goldens, and tests are updated together.

## Non-goals

- the human-first `pm reverse` workspace and `pm reverse guide` alias; those remain owned by #416.
- Connector list/browser/inspect/capability presentation (#411 and existing connector follow-ups).
- Help-tree/man-page architecture (#417), except the parity changes required by this issue.
- Interactive secret entry, masked password prompts, clipboard secret intake, raw secret flags, or
  a second vault path.
- Connection update/remove/replace/overwrite/`--force` behavior.
- A new global `--agent`, `--agent-mode`, or `--machine` flag.
- Generic shell, generic HTTP write, generic SQL write, or generic file-write tools.
- New dependencies, auth-scope changes, credentialed tests, or external mutations.

## RED -> GREEN -> REFACTOR plan

1. RED — write table-driven activation/completeness tests; reader-spy tests for piped stdin and
   `--value-stdin`; schema/placeholder validation tests; duplicate no-overwrite tests; Bubble Tea
   state/key/resize/accessibility tests; help/docs/website parity tests. Record exact failures.
2. GREEN — add the smallest shared preflight/completeness seam, credential guidance, connection
   guidance, and typed validation mapping needed to pass each slice. Reuse existing app services and
   approved UI components.
3. REFACTOR — deduplicate the gate, field model, error copy, and service adapters while green; keep
   Bubble Tea `Update` deterministic and I/O in `tea.Cmd`; rerun race/leak, redaction, transcript,
   and parity suites.
4. Commit and push coherent plan, RED, GREEN, refactor, and review-fix checkpoints to an isolated
   issue branch based on `feat/cli-architecture-v2`.

## Required GSD and skills

In the isolated issue worktree:

```text
/gsd doctor
/gsd plan-phase <this-issue-number>
/gsd-programming-loop init --phase <this-issue-number> --dry-run
/gsd execute-phase <this-issue-number>
/gsd verify-work
/gsd-code-review <this-issue-number>
```

Load and record `bubble-tea-tui-design`, `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-security`, `golang-safety`, `golang-error-handling`, `golang-context`,
`golang-concurrency`, `golang-documentation`, and `golang-spf13-cobra`.

## Verification

Focused gates:

- gate/completeness table: dual TTY, complete/incomplete, all bypasses, pipes, help routes;
- credential schema, placeholder, reader-spy, secret-marker, persistence, and partial-test cases;
- connection eligibility, defaults, duplicate-at-entry/final-race, no-overwrite, and cancellation;
- Bubble Tea model/key/focus/resize frames at wide, standard, compact, and guard sizes;
- accessible sequential, `NO_COLOR`, ASCII, sanitation/redaction, and no-ANSI non-TTY transcripts;
- `pm help credentials`, `pm credentials`, `pm credentials add --help`, and corresponding
  connections routes, docs generator, website generator, golden transcripts, and discovery data.

Broader gate:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test -race ./internal/ui/... ./internal/cli/...
go build ./cmd/pm
make verify
```

No credentialed connector or external mutation test is required. Use fake services, temporary
roots, synthetic secret markers, and injected TTY facts.

## Safety and human gates

- Never request, print, store in planning/docs, log, render, screenshot, or echo secret values.
- Never expose approval tokens; this child does not own reverse execution.
- Do not add dependencies. If the approved Charm/Huh foundation has not landed, stop at the
  dependency gate.
- Do not weaken tests, quality gates, JSON/stdout/stderr contracts, or exit-code taxonomy.
- The stacked implementation PR targets `feat/cli-architecture-v2` and uses `Refs` for this issue,
  #416, and #397. Parent PR #438 remains human-gated.

## Dependencies and downstream

- Blocked by #409 and the completed/reviewed #462 design gate (including PR #468).
- Parent feature issue: #416.
- Must complete before #417 help/man convergence and #418 accessibility convergence.

## Sources

- [Phase 18 UI contract in PR #468](https://github.com/polymetrics-ai/cli/pull/468)
- [Bubble Tea v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)
- [Huh v2 forms and accessible mode](https://github.com/charmbracelet/huh)
- [Bubbles v2](https://github.com/charmbracelet/bubbles)
- [GitHub CLI interactive/non-interactive create pattern](https://cli.github.com/manual/gh_repo_create)
- [GitHub CLI prompt-disable and accessibility environment](https://cli.github.com/manual/gh_help_environment)
- [GitHub CLI accessibility guide](https://accessibility.github.com/documentation/guide/cli/)
- [Command Line Interface Guidelines: interactivity](https://clig.dev/)
- [Pulumi non-interactive and structured-output conventions](https://www.pulumi.com/docs/iac/cli/commands/pulumi_do/)

Refs #416
Refs #397
