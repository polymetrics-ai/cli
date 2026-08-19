# PLAN — github-parity-extract-r1

## Goal

Land GitHub's full documented-operation parity on `main` as its own PR, extracted from the paused
sweep branch `fm/cli-top50-sweep-resume2-r1`, with shared artifacts **regenerated** rather than
hand-merged, and every reachable command **verified by running the binary**.

Then fold in the captain's order (`CAPTAIN-ORDER-unblock-commands.md`): restore the
`unsafe_or_disallowed` commands to full parity behind the confirmation gate the write path already
imposes — or report plainly which ones cannot be restored without a real runtime capability.

## Scope

In scope:

- GitHub bundle: `api_surface.json`, `cli_surface.json`, `operations.json`, `writes.json`.
- The foundation fix `covered_by.writes` (one endpoint backing several write actions) — not
  github-specific, it lands with this PR because github is its first consumer.
- Regenerated shared artifacts: runtime operation endpoint ledger, website connector catalog,
  golden CLI transcripts, generated connector docs.
- Reclassification of blocked github commands where an already-implemented twin proves the
  capability exists.

Out of scope (deliberately):

- The sweep's other connectors (workday-rest, zendesk-support, jira, stripe...). This PR is
  github-only. The sweep branch rebases onto the new `main` afterwards.
- The sweep's consolidated PR.
- `data/cli-top50-fixed-schema-sweep-r1/PROGRESS.md` — sweep bookkeeping, dropped from the pick.

## Approach

1. Cherry-pick the four github commits in TDD order, preserving red-before-green.
2. Reset the operation endpoint ledger to `main`'s content, then regenerate with
   `go run ./cmd/connectorgen surface-sync`. Assert the delta is confined to `github`.
3. Regenerate website catalog, golden transcripts, and connector docs from their own generators.
4. Sweep every `implemented`/`partial` github command through the real `pm` binary and assert none
   answers `unknown command`. Do **not** inherit the branch's 1079 figure — a prior worker found
   gmail reporting 79 successes while the binary rejected all 79.
5. Assess each blocked command against the runtime it would have to honour, and only reclassify
   where the runtime already executes that exact shape.

## Constraints

- Never weaken, skip, or delete a test.
- Never print a credential or token-derived value.
- `availability: implemented` is a claim the runtime has to honour
  (`TestEveryImplementedCommandPassesRuntimePreflight`). A classification flip that would fail
  preflight is the claim-before-establish defect, not a delivery.
- Never invent an `api_surface` endpoint to make a command look implemented.

## Verification gates

- `go test ./cmd/connectorgen/ ./internal/connectors/... ./internal/cli/ -timeout 20m`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `make connector-boundary`, `make docs-check`, `make lint`, `make agent-contract-check`
- Binary reachability sweep over the full github command surface.
