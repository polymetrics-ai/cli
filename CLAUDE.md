@AGENTS.md

## Claude Code

- `AGENTS.md` is the cross-agent source of truth for this repository. Keep this file thin so Claude
  Code and other agents do not drift.
- Use `.agents/` for reusable agent contracts, workflows, and YAML role specs. Update those shared
  files when a workflow changes instead of copying long rules here.
- For CodeRabbit review loops, follow `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`.
  In particular, do not post `@coderabbitai review` after every push; wait for automatic
  incremental review when active and use manual review only under the conditions documented in
  `AGENTS.md`.
