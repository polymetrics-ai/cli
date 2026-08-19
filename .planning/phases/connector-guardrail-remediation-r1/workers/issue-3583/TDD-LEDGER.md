# TDD LEDGER — issue 3583 PM/no-mistakes connector lane

## Rule

Capture red validation before production guidance/template edits. This slice is docs/template/prompt guidance; no executable schema currently validates these files, so grep-based validation is the executable docs check.

## Loaded skills

- `gsd-core`
- `no-mistakes`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-security`
- `golang-safety`
- `golang-error-handling`
- `golang-documentation`
- `golang-lint`

Skill rules applied:

- gsd-core workflow rules 1–9: issue-first planning, TDD ledger, verification checklist, evidence, commit checkpoint.
- no-mistakes active validation-step boundary; `ask-user` findings stop for user/coordinator decision.
- golang-testing Best Practices Summary #5: validate observable contract text.
- golang-documentation Writing Principles: preserve `must`/`should` obligations and avoid unsupported claims.
- golang-security Security Thinking Model #1–#3: keep secrets/generic raw write tooling out of connector/no-mistakes guidance.
- golang-safety Best Practices #10: make zero/default connector lane safe by stopping on shared-foundation needs.
- golang-error-handling Best Practices #2/#7/#14: blockers must be explicit and not double-handled or hidden.
- golang-lint Suppressing Lint Warnings #1–#4: do not introduce quality-gate bypass language.

## Evidence log

| Time | Slice | Command | Expected | Result |
| --- | --- | --- | --- | --- |
| 2026-08-02 | GSD preflight | `scripts/gsd doctor` | pass | pass |
| 2026-08-02 | GSD prompt | `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` | prompt generated | pass; `/tmp/gsd-execute-phase-3583.prompt.md` (87 lines) |
| 2026-08-02 | GSD fallback | `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` | alias if available | fallback; `scripts/gsd: unknown GSD command: programming-loop` |
| 2026-08-02 | Red docs validation | `rg` checks for `exactly one target connector`, `ownership guard evidence`, `foundation issue/PR|foundation PR`, and `no-mistakes.*foundation split|foundation split.*no-mistakes` across `.agents .pi .github` | fail before production edits | failed as expected: all four checks reported `RED missing ...`; command exited 1 |
| 2026-08-02 | Green docs validation | `rg` checks for target connector, ownership guard evidence, foundation issue/PR path, and no-mistakes foundation split across `.agents .pi .github` | pass after production edits | pass; all four required patterns found |
| 2026-08-02 | Issue verification grep | `rg -n "connector implementation|foundation|target connector|ownership guard|no-mistakes" .agents .pi .github docs` | pass | pass; exit 0 with required patterns present |
| 2026-08-02 | YAML/template parse | `python3` PyYAML load for edited YAML templates/specs | pass | pass; 5 edited YAML files loaded |
| 2026-08-02 | Hygiene | `git diff --check` | pass | pass; no output |
| 2026-08-02 | GSD verification | `scripts/gsd doctor` | pass | pass |

## Planned red validation

```bash
set -u
fail=0
check() {
  local name="$1"
  local pattern="$2"
  echo "--- ${name}: ${pattern}"
  if rg -n "${pattern}" .agents .pi .github >/tmp/issue-3583-${name}.txt; then
    cat /tmp/issue-3583-${name}.txt
  else
    echo "RED missing ${name}"
    fail=1
  fi
}
check target_connector "exactly one target connector"
check ownership_guard "ownership guard evidence"
check foundation_pr "foundation issue/PR|foundation PR"
check no_mistakes_foundation "no-mistakes.*foundation split|foundation split.*no-mistakes"
exit ${fail}
```

## Docs-only exemption

No Go behavior or runtime CLI surface is edited in this slice, and the repository has no current schema check for these markdown/YAML prompt/template obligations. Grep validation is therefore the executable docs check; Go/package tests are not required unless executable validation is added.
