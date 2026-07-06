# Summary

Issue: #50

## Delivered

- Added a generic parent issue orchestrator contract, workflow, state schema, YAML agent spec, and
  worker handoff template.
- Added a thin Codex custom-agent adapter that points back to `.agents/` as the source of truth.
- Updated existing agentic delivery contracts so worker agents implement and report while the
  orchestrator owns shared parent artifacts, sub-PR merge decisions, CodeRabbit coverage routing,
  and final parent PR readiness.
- Updated issue and PR templates with parent orchestration and CodeRabbit coverage fields.
- Updated `AGENTS.md` and `CLAUDE.md` with short pointers to the orchestrator contract.

## Verification

See `VERIFICATION.md`.

## Safety

- No secrets, auth scope changes, dependencies, production deploys, or destructive external actions.
- Parent PR merge to `main` remains human-gated.
