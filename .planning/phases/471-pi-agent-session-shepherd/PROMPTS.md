# Prompt Snapshots

Phase: `471-pi-agent-session-shepherd`

## 2026-07-21T08:17:01Z — coordinator kickoff

- Agent role: coordinator
- Loop type: manual fallback for `programming-loop run`
- Input refs: issue #471, `AGENTS.md`, universal loop PRD/prompt library, Pi 0.80.6 SDK docs,
  required repository contracts
- Downstream artifact: this phase directory and `.pi/extensions/shepherd/**`
- Verification result: pending

```text
Implement issue #471 using strict RED → GREEN → refactor. Keep the deterministic supervisor core
independent of the Pi SDK, use only public Pi 0.80.6 APIs at the adapter edge, persist bounded
redacted state outside the repository, run the first live canary read-only against PR #438, and
stop at the human merge gate.
```

## Read-only reconnaissance roles

- Repository/Pi-extension scout: inspect main-branch extension and test infrastructure.
- Pi SDK scout: verify exact public APIs and cleanup semantics without editing files.
- GitHub topology scout: reconcile #372/#389/#470/#438 and human gates without mutation.

Their structured findings are recorded under `traces/`; no scout is authorized to edit.
