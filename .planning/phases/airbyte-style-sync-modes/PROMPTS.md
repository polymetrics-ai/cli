# Prompt Snapshots

Phase: airbyte-style-sync-modes

## 2026-06-24T20:41:49.429Z - universal-kickoff

- Agent role: coordinator
- Loop type: run
- Input refs: docs/plans/universal-programming-loop-prd.md, docs/prompts/universal-programming-loop-prompts.md, docs/architecture/repo-profile.json
- Downstream artifact: pending
- Verification result: pending

```text
Run the GSD Universal Programming Loop using the repo PRD, prompt library, strict TDD gate, local verification, and committed phase traces.
```

## 2026-06-25 - implementation-prompt

- Agent role: implementation
- Input refs: docs/prompts/airbyte-sync-modes-implementation-prompt.md
- Downstream artifact: dependency-free Airbyte-style sync-mode implementation
- Verification result: passed through local Go gates and smoke workflow

```text
Implement Airbyte-style sync modes for PM using the GSD universal programming loop and strict TDD, starting with the dependency-free local JSONL warehouse path.
```
