# Prompts: Gorgias CLI Parity Parent Orchestration

## Prompt commands used

```bash
scripts/gsd prompt plan-phase 196 --skip-research
scripts/gsd prompt execute-phase issue-196-gorgias-cli-parity --dry-run
scripts/gsd prompt programming-loop init --phase issue-196-gorgias-cli-parity --dry-run
```

## Prompt results

- `plan-phase`: generated official GSD prompt for `/gsd-plan-phase 196 --skip-research`.
- `execute-phase`: generated official GSD prompt for `/gsd-execute-phase issue-196-gorgias-cli-parity --dry-run`.
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop`.

## Manual fallback source

Manual fallback follows:

- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.pi/prompts/pm-gsd-loop.md`
- `.pi/prompts/pm-orchestrate.md`
- `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md`
