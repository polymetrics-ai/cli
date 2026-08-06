# Plan — #3771 command-runner runtime content

## Goal

Remove the last runtime connector-command redaction site while retaining declaration compatibility
and all unrelated safety controls.

## Slices

1. **#3782 — runner red test and implementation** — completed
   - Change the legacy masking assertions in `internal/connectors/commandrunner/runner_test.go`
     deliberately into content-preservation tests, and add a failing test for request forwarding
     where necessary.
   - Run the narrow tests to capture the red failure before production edits.
   - In only the owned runner functions, preserve a write command's record clone, emit ETL records
     unchanged, return original errors, stop setting `RedactFields` on executor requests, and
     delete the runner redaction helpers while retaining unrelated safety utilities used by input
     validation.

2. **#3784 — application persistence proof** — completed
   - Add a credential-free fake connector test under `internal/app` that calls the public
     `PlanConnectorCommand` path, persists a reverse command plan, reopens state, and verifies
     nested/token/content fields remain in the saved sample.
   - Keep the test preview-only and assert it does not invoke the fake write method.

3. **#3786 — CLI, manual, and website parity** — completed
   - Update the reverse help text, CLI manual, golden transcript fixture/test, and website reverse
     ETL guide. Explain that connector command content is complete; approval-token and destructive
     lifecycle protections remain; generic source-table output handling is out of scope.
   - Regenerate or deliberately update only the relevant golden artifact using the repository's
     established test path.

4. **#3790 — behavior regression matrix** — completed
   - Add a focused commandrunner test file that exercises ETL, reverse preview, direct read,
     operation direct read, binary download, and their error paths through public behavior.
   - Assert executor request structs have nil/empty `RedactFields`, records retain exact nested and
     heuristic-sensitive values, and errors retain their original text.

5. **Verification and review**
   - Run focused commandrunner, application, and CLI tests; format changed Go; run `go vet` and
     build; execute `make verify`'s individual non-aggregate gates specified in `AGENTS.md`.
   - Run the local GSD verification/review fallback checklist, inspect the final diff, and commit
     the verified branch. No push, PR, or merge occurs in this worker stage.

## Constraints

- Test first: preserve red-then-green terminal evidence in `TDD-LEDGER.md`.
- Do not touch runner functions owned by #3775 or #3769, engine paths, bundles, or generic app
  output redaction.
- Use fixtures and fakes only; never provide credentials or make provider calls.
