# Prompt snapshot — CLI Architecture v2 Cobra/Viper release split

## Kickoff

- Task: reconstruct and ship a latest-main Cobra + typed Viper/config release candidate without TUI or OpenTelemetry.
- Source plan: captain-approved private release-split scout report.
- Runtime: Pi, manual-GSD fallback because the repo adapter does not expose `programming-loop`.
- Required policy: `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and `docs/prompts/universal-programming-loop-prompts.md`.
- Exact base: `873cd7b251f70c4a35a607a0d4e86051ea0fbd15`.
- Authorized source squashes: `379cb501`, `8900db14`, `7683087d`, `cc2a90e9`, `20475ddf` in that order.
- Execution decision: `local_critical_path`; this is one dependency-ordered reconstruction in the already isolated assigned worktree, with no disjoint mutating slice to delegate.
- Downstream artifact: draft PR #499 (`https://github.com/polymetrics-ai/cli/pull/499`) includes the green Cobra/Viper candidate and checksummed cross-platform release automation.
- Verification result: Cobra/Viper and release-asset campaigns locally green, including deterministic six-target GoReleaser snapshots; PM review, no-mistakes, remote CI/Dependency Review/Snyk remain pending.
