---
phase: issue-4015-api-to-api-github-github-r1
plan: 01
type: tdd
status: completed
---

# #4015 — API → API GitHub route proof

## Foundation check

| Need | Existing executable proof | This delivery's evidence |
| --- | --- | --- |
| Source | `github` declarative `issues` stream, `declarative_stream_source`, and real command runner | `pm connectors inspect github --json`, definition validation, targeted tests |
| Durable materialization | connection-owned JSONL WAL plus DuckDB-produced Parquet and manifest | fresh binary run with durable artifact and receipt facts |
| Reverse delivery | `issue_label_destination`, closed plan → preview → stdin approval → `add_issue_labels` | supported full-append live run and independent provider read-back |
| Acknowledgement/checkpoint | `ReadBackDestination` precedes durable acknowledgement and checkpoint commit | run state reports an acknowledged checkpoint after independent read-back |
| Honest deletion semantics | destination says `deletes: not_available`; cleanup is explicit `remove_issue_label` | no inferred delete propagation; explicit missing-label typed cleanup succeeds |

## TDD slices

1. **Red — establish refusal coverage before live I/O.** Run the existing
   executable regression cases which must fail if an ineligible `issues`
   source stream or unsupported action/strategy reaches source or destination
   I/O. This proof-only lane adds no production behavior, so it does not invent
   a synthetic failing code change; the named existing RED regressions are the
   safety baseline.
2. **Green — execute the accepted binary route.** Build a fresh `pm`, initialise
   an isolated project, add saved GitHub source/destination credentials through
   environment-backed PM credential input, create the closed connection, plan,
   preview, approve through stdin, and run one bounded issue. Read the target
   independently from GitHub and assert exact label values. Preserve only
   sanitized counts, hashes, actions, and receipt/checkpoint facts.
3. **Refactor/edge — prove no accidental semantics.** Verify zero source,
   missing record mapping, replay, and `deletes: not_available` behavior with
   the existing transport regressions plus the live rerun and explicit typed
   cleanup. No general mapper, raw HTTP writer, SQL writer, or extra GitHub
   action may be introduced.

## Verification plan

1. Targeted transport, app, and CLI binary tests with `-timeout 20m`.
2. Build the binary, run metadata-only connector inspection and CLI help checks.
3. Run definition, generated-surface, connector runtime preflight, lint,
   documentation, contract, and release checks individually as required by
   `AGENTS.md`.
4. Run the real GitHub proof only after the no-I/O bad-case evidence passes.
   If the available credential lacks write permission in the controlled
   repository, record `needs-decision` with the exact missing scope and stop.
5. Run `verify-work`, then `code-review`, and record any finding disposition.

## CLI/docs parity

No command, flag, help, documentation, or generated surface changes are
planned. The existing carrier remains nevertheless verified through `pm etl
transport`, `pm etl transport github-issue-label --help`, and the generated
transcript/docs checks; therefore no docs change is applicable.
