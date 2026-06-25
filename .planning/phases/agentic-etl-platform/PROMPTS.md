# Prompt Snapshots

Phase: agentic-etl-platform

## 2026-06-24T15:39:46.700Z - universal-kickoff

- Agent role: coordinator
- Loop type: run
- Input refs: docs/plans/universal-programming-loop-prd.md, docs/prompts/universal-programming-loop-prompts.md, docs/architecture/repo-profile.json
- Downstream artifact: pending
- Verification result: pending

```text
Run the GSD Universal Programming Loop using the repo PRD, prompt library, strict TDD gate, local verification, and committed phase traces.
```

## Canonical Implementation Prompt

Source: docs/prompts/gsd-agentic-etl-full-rewrite-tdd-prompt.md

This phase executes the canonical prompt with strict red-green-refactor TDD. The implementation is scoped to the highest-impact local slices that can be verified without adding dependencies or using live external secrets: structured CLI errors, validation/sanitization, connector manifests, generated skills/docs, and bounded streaming ETL.
