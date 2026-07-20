## Objective

Reduce manual token/ID relay through a human-first bare `pm reverse` workspace, with
`pm reverse guide` retained as an explicit alias, while preserving the exact reverse ETL plan →
preview → approval → execute safety gate.

## Ownership split

This is Phase 18A / Track B under parent #397. It owns only the guided reverse ETL session.
TTY-progressive credential and connection setup is the scoped child #469 and must use a separate
GSD loop, isolated worktree, and stacked PR.

## Scope

- Make eligible dual-TTY bare `pm reverse` enter the guided workspace; retain `pm reverse guide` as
  an explicit alias to the same state model.
- Keep `pm reverse --help` help-only and make bypass/non-TTY bare `pm reverse` render deterministic
  contextual help and exit 0.
- Project the existing reverse plan, preview, approval, typed confirmation, execute, and status
  services into the approved Bubble Tea/Huh interaction system.
- Reduce manual relay inside the session without creating a second mutation path.
- Preserve plain, JSON, stdout/stderr, exit-code, cancellation, help/manual/docs/website, generated
  artifact, completion, and discovery contracts.

## Acceptance criteria

- [ ] Guided state transitions are equivalent to the existing flag flow; no execution occurs before
      the existing approval and typed-confirmation gates pass.
- [ ] Approval tokens live only ephemerally in memory and never appear in final frames, transcripts,
      logs, screenshots, accessibility output, JSON, shell-equivalent commands, or fixtures.
- [ ] Eligible dual-TTY bare `pm reverse` and `pm reverse guide` activate the same guided state
      model; invalid actions remain usage errors.
- [ ] Help flags always render help. Bypass/non-TTY bare `pm reverse` renders deterministic
      contextual help and exits 0 without initializing Bubble Tea/Huh.
- [ ] `--json`, `--plain`, `--no-input`, PM_NO_TUI, CI, TERM=dumb, piped stdin, and piped stdout
      never initialize Bubble Tea/Huh, consume scripted input, hang, or use `/dev/tty`.
- [ ] Vim/arrows, explicit modes, one-layer escape, Ctrl+C cancellation, responsive layouts,
      accessible sequential prompting, no-color/ASCII, sanitation, and redaction match the approved
      Phase 18 UI contract.
- [ ] Runtime help, focused `--help`, `docs/cli/**`, website docs, generated manuals, completion,
      discovery metadata, goldens, and tests are updated together.

## Non-goals

- Credential or connection setup; owned by #469.
- Interactive secret entry or approval-token display.
- A new reverse mutation, approval, or credential storage path.
- Generic shell, generic HTTP write, generic SQL write, or generic file-write actions.
- New dependencies, auth-scope changes, credentialed external tests, or parent/main merges.

## RED → GREEN → REFACTOR

1. RED — record failing state/service equivalence, bypass/reader-spy, cancellation, token-marker,
   model/key/resize/accessibility, and help/docs/website parity tests before production edits.
2. GREEN — add the smallest guide state model and service adapters required to pass each slice;
   keep `Update` deterministic and I/O in cancellation-aware `tea.Cmd` values.
3. REFACTOR — deduplicate state/copy/adapters while green, then rerun race/leak, transcript,
   sanitation, token-marker, and parity gates.
4. Commit and push coherent plan, RED, GREEN, refactor, and review-fix checkpoints to an isolated
   issue branch based on `feat/cli-architecture-v2`.

## Required GSD and skills

```text
/gsd doctor
/gsd plan-phase 416
/gsd-programming-loop init --phase 416 --dry-run
/gsd execute-phase 416
/gsd verify-work
/gsd-code-review 416
```

Load and record `bubble-tea-tui-design`, `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-security`, `golang-safety`, `golang-error-handling`, `golang-context`,
`golang-concurrency`, `golang-documentation`, and `golang-spf13-cobra`.

## Required reading

- `.planning/phases/416-guided-reverse-connection-prompts/18-UI-SPEC.md`
- `docs/design/tui-ux-design.md`
- `docs/design/terminal-ui-research-and-design-system.md`
- `docs/adr/0003-interactive-tui-layer.md`
- CLI help/docs/website parity and issue-agent contracts named by `AGENTS.md`

## Dependencies and gates

- Blocked by #409 and the completed, reviewed #462 design gate including PR #468.
- Setup child #469 may run separately only after its blockers clear and write scopes do not collide.
- The stacked implementation PR targets `feat/cli-architecture-v2` and uses `Refs #416` and
  `Refs #397`. Parent PR #438 remains human-gated.

Refs #397
Refs #469
