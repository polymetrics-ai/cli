# Issue #4359 — TDD plan

## Slice 1 — Freeze the exact contract

1. Read the current main tree, the read-only Batch-1 tree, and the retained CircleCI source lock; map provider source → api surface → CLI declaration → engine resolver → commandrunner → App/CLI.
2. Write `REVIEW-CONVERGENCE.md` before any production edit and obtain a fresh-context Codex audit in place of unavailable Claude Code.
3. Commit the frozen plan/review record before the red test.

## Slice 2 — Red

1. Add behavioral engine tests for all eleven real CircleCI-shaped bindings. They must fail at the current endpoint-equivalence boundary, not merely assert helper internals.
2. Add a closed negative matrix: absent configuration, wrong connector, source URL/artifact mismatch, wrong placeholder, reordered/partial/extra/repeated keys, non-config template, an absolute runtime path, a changed method, a changed literal path, and direct-read/direct-write/binary binding attempts.
3. Record the failing command and output in `TDD-LEDGER.md` and `traces/` before production code.

## Slice 3 — Green and refactor

1. Add a declaration-owned, schema-validated identity proof record at the existing API-surface/equivalence boundary. It carries a per-binding source ID, provider operation ID, source URL/digest/location, command intent, binding identity, method, canonical path, provider placeholder, and exact config-key order.
2. Validate it fail-closed: it is accepted only for the eleven named CircleCI source/binding rows, exact `https://circleci.com/api/v2/openapi.json` citation/digest, `{project-slug}`, and exactly `[vcs_type, org, repo]`; no other connector/configuration is admitted.
3. At the resolver, allow the proof only when command intent, binding, method, all non-identity literals, and relative transport match and the provider placeholder expands to exactly the three declared config segments. Preserve current proof branches and do not alter request execution.
4. Add focused loader/schema coverage and retain exact resolution tests for all six lanes.

## Slice 4 — Integration proof

1. Build `pm`; run the eleven declared command paths using a fresh initialized project and their source-declared fixture inputs. Every implemented path must stop at `missing --credential`, with neither unknown command nor provider I/O.
2. Run the focused and repository generator/check gates named in `VERIFICATION.md`. Regenerate only outputs proven stale by their owning generator.
3. Run fresh-context exact-head Codex re-review, commit the review/evidence-only record (if the code SHA remains fixed), push normally, open a `main` PR, and API-read its base.
