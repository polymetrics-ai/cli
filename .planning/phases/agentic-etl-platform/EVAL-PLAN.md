# EVAL PLAN: Agentic ETL Platform

## Evaluation Cases

- Agent can discover GitHub connector streams and required credential fields.
- Agent can generate skills without secrets.
- Agent receives structured JSON errors for bad commands.
- ETL can load multi-page source records through bounded batches.
- Reverse ETL still requires approval before writes.

## Passing Criteria

- Automated tests cover each case.
- `make verify` passes.
- Manual CLI generation commands complete successfully.
