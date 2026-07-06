# Summary

Issue: #50

## Delivered

- Added a generic parent issue orchestrator contract, workflow, state schema, YAML agent spec, and
  worker handoff template.
- Added a thin Codex custom-agent adapter that points back to `.agents/` as the source of truth.
- Updated existing agentic delivery contracts so worker agents implement and report while the
  orchestrator owns shared parent artifacts, sub-PR merge decisions, automated review coverage routing,
  and final parent PR readiness.
- Updated issue and PR templates with parent orchestration and automated review coverage fields.
- Updated `AGENTS.md` and `CLAUDE.md` with short pointers to the orchestrator contract.
- Corrected CodeRabbit usage policy after PR #51 showed that a manual `@coderabbitai full review`
  on a non-draft `main` PR can consume review allowance and hit fair-usage limits. The workflow now
  waits for automatic review by default and treats manual review commands as fallback-only.
- Added a source-backed automated review routing loop that keeps CodeRabbit as the primary reviewer
  and uses GitHub Copilot as backup only when CodeRabbit is rate-limited, skipped, disabled,
  paused, or unavailable and review coverage is blocking progress. Copilot comments require the
  same disposition process, but Copilot review is not approval.

## Verification

See `VERIFICATION.md`.

## Safety

- No secrets, auth scope changes, dependencies, production deploys, or destructive external actions.
- Parent PR merge to `main` remains human-gated.
