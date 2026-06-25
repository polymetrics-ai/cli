# RUNBOOK: Agentic ETL Platform

## Verify

```bash
make verify
./poly docs generate --dir docs/cli
./poly skills generate --dir /tmp/poly-skills
```

## Runtime Services

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
scripts/runtime.sh ps
scripts/runtime.sh down
```

## Rollback

This phase is local-file and code-only. Roll back by reverting changed files. No migration or destructive data action is required.
