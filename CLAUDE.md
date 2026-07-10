@AGENTS.md

## Claude Code

- `AGENTS.md` is the cross-agent source of truth for this repository. Keep this file thin so Claude
  Code and other agents do not drift.
- Use `.agents/` for reusable agent contracts, workflows, and YAML role specs. Update those shared
  files when a workflow changes instead of copying long rules here.
- For parent issues with sub-issues, follow
  `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`; runtime-specific agent
  files are thin adapters over `.agents/`. Treat parent issue orchestration as active ownership:
  spawn or assign ready workers until human-ready, blocked, or explicitly limited by the user.
- For automated review routing, follow
  `.agents/agentic-delivery/workflows/automated-review-routing-loop.md` and
  `.agents/agentic-delivery/workflows/claude-review-loop.md`. Claude Code is the primary automated
  reviewer via the `.github/workflows/claude-review.yml` Action (auto-review on PR open for trusted
  authors, plus on-demand `@claude` review); GitHub Copilot review is fallback-only when Claude is
  unavailable. Do not comment `@claude review` after every push.
