# Prompts — Issue #204 Crisp CLI parity parent

## GSD plan prompt

Generated with:

```bash
scripts/gsd prompt plan-phase 204 --skip-research
```

Output path during this session: `/tmp/gsd-plan-phase-204.txt`.

Key instruction excerpt followed:

> Execute the requested workflow using Pi tools. If the official workflow expects runtime subagents that Pi does not provide, perform the documented inline/manual fallback and record that fallback in the generated planning artifacts.

Downstream artifact: `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`.
Verification result: planning artifacts created; production edits not started.

## Programming-loop fallback prompt

Attempted:

```bash
scripts/gsd prompt programming-loop init --phase issue-204-crisp-cli-parity --dry-run
```

Result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback: manual GSD/TDD loop using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and recorded ledgers.
