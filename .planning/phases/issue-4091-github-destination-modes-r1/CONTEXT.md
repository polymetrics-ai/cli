# #4091 — GitHub destination-mode authorization context

## Task delivery header

- Issue: Refs #4091 — authorize set-replace and keyed issue-label destination modes under explicit per-connection consent.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with required checks green and API read-back of the target base.
- Working branch: `fm/cli-4091-github-destination-modes-r1`
- Task: Declare GitHub safe-additive, set-replace, and keyed issue-label destination modes and allow non-additive execution only after an explicit, durable, revocable per-connection authorization.
- Verification: `go test -timeout 20m ./internal/app/... ./internal/connectors/hooks/github/...`, targeted zero-send regression tests, generated-definition checks, local repository gates, PR checks, and API base read-back.

## Fixed decisions

- Safe additive GitHub reverse ETL remains the default and does not require the non-additive opt-in.
- The opt-in is a destination-connection fact. It is explicit, off by default, non-inherited, and can be disabled without disabling safe additive execution.
- `plan` → `preview` → `proceed` consumes the approval token once and mints the durable #4132 authorization record. A later identical scope may run unattended; any changed scope, revocation, or expiry stops before a provider request.
- Authorization scope binds the connection, credential revision, streams/tables, field mappings, action/mode, destination configuration, enabled non-additive operations, confirmation policy, and expiry. It never binds record contents, counts, timestamps, cursors, or run ID.
- Do not change shared delivery-ledger behavior, generic reverse-plan hashing, authorization primitives, or unrelated connectors. The only target connector is GitHub.

## Manual lifecycle fallback

The active lane uses an issue identifier rather than a numeric GSD phase, and the canonical contract forbids spawning workflow roles. Generated `scripts/gsd prompt` commands were resolved and this lifecycle is being performed inline with the same discussion, TDD, execution, verification, and review evidence.
