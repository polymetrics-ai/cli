# GSD Roadmap Refresh Trace

Attempted command:

```bash
scripts/gsd prompt roadmap --refresh
```

Result:

```text
scripts/gsd: unknown GSD command: roadmap
```

Resolution:

`roadmap` is not an official command in the current `.gsd/commands.json` registry generated from official `open-gsd/gsd-core@next` `docs/COMMANDS.md`. The roadmap refresh was therefore performed through official available command prompts:

```bash
scripts/gsd prompt onboard --fast --skip-phases
scripts/gsd prompt new-project --from-existing --non-interactive
scripts/gsd prompt milestone-summary --planning-only
scripts/gsd prompt docs-update .planning AGENTS.md .agents --planning-only
```

`.planning/phases/**` was intentionally not regenerated.
