# Prompts: Gorgias CLI Surface Metadata

## Prompt commands used

```bash
scripts/gsd prompt plan-phase 197 --skip-research
scripts/gsd prompt execute-phase issue-197-gorgias-cli-surface-metadata --dry-run
scripts/gsd prompt programming-loop init --phase issue-197-gorgias-cli-surface-metadata --dry-run
```

## Prompt results

- `plan-phase`: generated official GSD prompt for `/gsd-plan-phase 197 --skip-research`.
- `execute-phase`: generated official GSD prompt for `/gsd-execute-phase issue-197-gorgias-cli-surface-metadata --dry-run`.
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop`.

## Manual fallback source

Manual fallback follows `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` and `.pi/prompts/pm-gsd-loop.md`.
