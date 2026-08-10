---
coverage:
  - id: D1
    description: GitHub's documented operation surface is recorded at 1220 REST operations, re-derived from a byte-identical re-fetch of GitHub's own OpenAPI description, with the 4 fixed GraphQL rows counted separately and never folded in.
    verification:
      - kind: unit
        ref: cmd/connectorgen/github_documented_surface_test.go (TestGitHubDocumentedRESTSurfaceIsComplete — row count, method split, uniqueness, GraphQL counted apart)
        status: pass
    human_judgment: false
  - id: D2
    description: The 719 operations the shipped bundle never enumerated — whole org, user, enterprise, app and admin scopes — are enumerated and dispositioned.
    verification:
      - kind: unit
        ref: cmd/connectorgen/github_documented_surface_test.go (one spot-pin per previously absent scope)
        status: pass
    human_judgment: false
  - id: D3
    description: No row encodes a behaviour variant into its path. The four synthetic "(close)"/"(reopen)" rows are gone, and the write actions they existed for are still declared, still hook-backed, and still reachable.
    verification:
      - kind: unit
        ref: cmd/connectorgen/github_documented_surface_test.go (rejects any REST path containing a space, "?" or "*")
        status: pass
      - kind: unit
        ref: cmd/connectorgen/validate_surface_test.go (TestCheckAPISurface_EndpointMayBackMultipleWriteActions)
        status: pass
      - kind: other
        ref: /tmp/pm-gh github issue close --help && … pr reopen --help
        status: pass
    human_judgment: false
  - id: D4
    description: Every blocked row names a machine-checkable dependency; none is a shrug.
    verification:
      - kind: other
        ref: audit of all 98 blocked rows — 0 lack a "Named dependency:" marker outside the duplicate/deprecated models
        status: pass
    human_judgment: true
  - id: D5
    description: Every command that claims to be implemented actually routes when the binary runs it.
    verification:
      - kind: other
        ref: 1079 implemented+partial commands swept through the built binary, routing asserted on the rendered NAME line
        status: pass
    human_judgment: false
  - id: D6
    description: The shared runtime endpoint ledger moves only for github, and only by the two operations that earned an entry.
    verification:
      - kind: other
        ref: object-by-object diff of operation_endpoint_ledger.json against HEAD — exactly one connector key changed, 162 → 164
        status: pass
    human_judgment: false
  - id: D7
    description: No paging flag is hand-authored anywhere in the delivery.
    verification:
      - kind: other
        ref: both generators raise rather than emit a flag whose name is in the paging blocklist; grep of the new commands finds none
        status: pass
    human_judgment: false
---

# github — documented-operation parity

Phase `github-parity-sweep-r1`, program `cli-top50-fixed-schema-sweep-r1`.
Landing order **#1** under the captain's largest-first reversal. Delivered in slices; TDD evidence
per slice is in `TDD-LEDGER.md`, and run state in `RUN-STATE.json`.

## What changed

| | Before | After |
| --- | ---: | ---: |
| `api_surface.json` rows | 509 (505 REST + 4 GraphQL) | **1224 (1220 REST + 4 GraphQL)** |
| of which real documented REST rows | 501 | **1220** |
| synthetic path rows | 4 | **0** |
| covered | 440 | **1126** |
| blocked | 69 | **98** |
| legacy `excluded` | 0 | 0 |
| CLI commands | 461 | **1147** |
| implemented + partial | 383 + 6 | **1042 + 37** |
| write actions | 231 | **553** |
| operations | 341 | **345** |

The shipped bundle's own `scope` said why the gap existed: it enumerated only
`/repos/{owner}/{repo}/…` and declared org-, user-, enterprise- and admin-level surfaces out of
scope. That is 501 real operations out of 1220. The 719 missing ones were missing by whole scope,
not scattered.

## Derivation, reproduced rather than trusted

GitHub's own `rest-api-description` → `descriptions/api.github.com/api.github.com.json`, re-fetched at
**12,920,264 bytes** — byte-identical to the sweep's recorded derivation, which is what makes it the
same artifact rather than a lookalike. `openapi: 3.0.3`, `info.version: 1.1.4`. 808 paths carry
**1220** method entries, all unique: `GET 636 · POST 193 · PUT 134 · DELETE 187 · PATCH 70`, 37
`deprecated: true`. The extraction reproduces `DERIVED-OPERATIONS.json` operation-for-operation, with
zero drift in either direction.

**270 webhook events live under the vendor extension `x-webhooks`,** not a top-level `webhooks`
object — the spec is 3.0.3, which has none. Tooling that checks only a literal `webhooks` key records
zero events for the largest connector in the sweep. Events are excluded from the operation surface by
counting policy; the **28 webhook management operations** are ordinary `paths` entries and are counted.

## The four judgements

1. **read vs write** — decided per operation, not per method. Every GET is a read. Five POSTs are
   semantically reads, and they split three ways: two bulk-list lookups are implemented reads, two
   markdown renderers are blocked because they answer `text/html`, and the token check is blocked
   because its request body carries a credential.
2. **stream vs direct read** — the 364 new GET commands are plain `direct_read`. A stream needs a
   hand-authored record schema, primary key and fixture; inventing 364 would invent contracts GitHub
   never published. It also keeps the shared endpoint ledger untouched. The 37 existing streams are
   unchanged.
3. **binary detection** — read out of the artifact: a documented **302 redirect** whose operationId
   verb is not `check…`. Two new binary downloads (org and user migration archives) join the 8
   already shipped. `orgs/check-membership-for-user` documents a 302 as a *status* and is correctly
   not treated as a download.
4. **named-dependency blocking** — all 98 blocked rows name a grep-checkable blocker. See
   `TDD-LEDGER.md` §3b.3 for the full table.

## One foundation change was required, and it was not optional

`covered_by.write` was a single string, so a bundle modelling several write contracts over one
documented path had nowhere to record the others. That is why the shipped bundle invented
`PATCH .../issues/{issue_number} (close)` and three siblings. Deleting the extra actions was not
available: `internal/connectors/hooks/github/hooks.go` builds their bodies by switching on the action
**name**, and certify's create/cleanup pairing binds `create_issue` to `close_issue`.

`covered_by.writes` now mirrors the `direct_reads` array that already solved this shape for reads.
Nothing was loosened — an undeclared action in a plural entry is still a finding.

## Known-unmet, carried deliberately

- **Shared generated artifacts are not regenerated here.** Docs under `docs/cli/**`, `website/**` and
  the golden transcripts regenerate **once at the end of the sweep**; doing it per connector would
  churn ~1,034 files of pre-existing `main` drift on every commit.
  **`TestGoldenTranscripts/root_bare_manual` therefore fails on this branch** — verified pre-existing
  before this phase, and discharged by the end-of-sweep regeneration.
- **12 `oneOf`/`anyOf`-rooted mutations are blocked, not split into arms.** AGENTS.md sanctions both;
  this takes the "leave it non-implemented" option and names the runtime rule. `TDD-LEDGER.md` §3b.3
  records which are trivially splittable.
- **GitHub's GraphQL schema is still enumerated at 4 fixed operations.** A named scope gap, not a
  coverage claim.
- **`code-sanning` is a typo in a pre-existing shipped top-level command name** (`code-scanning` also
  exists). Found while listing the command surface; not corrected here because renaming a shipped
  command is a user-visible change that belongs in its own commit.
