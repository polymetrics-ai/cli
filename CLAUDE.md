@AGENTS.md

## Claude Code

- `AGENTS.md` is the cross-agent source of truth for this repository. Keep this file thin so Claude
  Code and other agents do not drift.
- Use `.agents/` for reusable agent contracts, workflows, and YAML role specs. Update those shared
  files when a workflow changes instead of copying long rules here.
- For parent issues with sub-issues, follow
  `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`; its compatibility filename
  now defines single-worker parent ownership, not a dedicated role. The canonical worker processes
  ready waves inline and persists state in issues, branches, PRs, and GSD artifacts; do not spawn an
  orchestrator, shepherd, planner, reviewer, verifier, GSD role, or extra worker for the job.
- For automated review routing, follow
  `.agents/agentic-delivery/workflows/automated-review-routing-loop.md` and
  `.agents/agentic-delivery/workflows/claude-review-loop.md`. Claude Code is the primary automated
  reviewer via the `.github/workflows/claude-review.yml` Action (auto-review on PR open for trusted
  authors, plus on-demand `@claude` review); GitHub Copilot review is fallback-only when Claude is
  unavailable. Do not comment `@claude review` after every push.
