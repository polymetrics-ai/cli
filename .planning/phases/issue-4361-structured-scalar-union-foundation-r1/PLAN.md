# Plan — issue 4361: typed nullable-string request values

## Task Delivery Header

- Issue: Refs #4361 — feat(engine): validate declared nullable string request fields
- Base branch: main
- Merges into: main
- Delivery: Open one PR from `fm/cli-structured-scalar-union-foundation-r1` to `main`, with current-main integration, local gates, and an independent exact-head audit recorded.
- Working branch: fm/cli-structured-scalar-union-foundation-r1
- Task: Add the smallest shared, declaration-bound representation for an exact `string|null` request field. It must preserve source schema validation, refuse undeclared values and paths, and leave source locks, credentials, provider I/O, certification, safety classification, and consumer-lane promotion untouched.
- Verification: Focused engine, commandrunner, and connectorgen red/green tests; source/definition and generated-surface checks; credential-free command-surface census; build; repository gates; diff check; independent exact-head audit.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An exact declared `string|null` field can use the bounded JSON field contract | live | The real Twilio `create_address.City` and Xero `delete_payment.Status` definitions pass the shared declaration preflight; pre-change they fail with the object-or-array-only error. |
| A declared string and explicit null materialize | live | `BuildWriteCommand` produces the declared string and Go `nil` from strict JSON inputs. |
| Other JSON values remain refused | live | The same command path rejects number, boolean, array, object, unknown flag, missing required field, and null against a string-only schema before provider I/O. |
| Generator preserves the closed source boundary | live | Source projection emits a named `json` flag without `allow_bare_string` only for an exact `string|null` declaration; it does not emit a body/route/method escape hatch. |
| Existing bounds, enum, empty, nested object, and array behavior remains intact | live | Existing focused schema/body/commandrunner suites remain green; targeted regressions retain their former assertions. |
| No executable claim is fabricated | live | Current `main` bundle inventory shows no Twilio command surface and no affected Xero command; the report records a zero direct usable-surface delta and named Batch 2–3 consumer work. |

## GSD / TDD execution record

The installed GSD prompts were resolved with `scripts/gsd sources` for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`. This isolated direct-PR worker executes those stages inline: the canonical contract does not authorize spawning the generated role workers from this task environment. This is an explicit manual-GSD fallback, not a lifecycle exemption.

1. **Discuss / discover.** Freeze issue, base SHA, ownership, source declarations, affected lanes, and consumer boundary in `DISCOVERY.md`.
2. **Plan / red.** Add the failing definition-backed tests before production edits and record their exact failure in `TDD-LEDGER.md`.
3. **Green.** Make the smallest common preflight/projection change. Do not add connector code, artifacts, source locks, raw JSON body input, routes, methods, or credentials.
4. **Refactor.** Keep the exact-union classifier single-purpose and reuse existing schema validation and strict JSON decoding.
5. **Verify / review.** Run the stated local gates, merge current `origin/main` normally, then request an independent exact-head Codex audit and disposition every finding.

## Scope and non-goals

- In scope: `internal/connectors/engine`, `internal/connectors/commandrunner` tests, and `cmd/connectorgen` projection tests only as required to preserve a source-declared typed field.
- Out of scope: source locks/artifacts, credentials, provider network calls, Twilio/Xero bundle materialization, availability promotion, direct DELETE/reverse-ETL policy, certification, and Batch 1 / Batch 2–3 / #4313 worktrees.
- CLI docs/help/website: no current command or generated help changes on `main`; each future consumer adoption owns its own parity work. Existing generated/docs checks still run.

