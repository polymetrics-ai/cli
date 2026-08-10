# SUMMARY — github-parity-extract-r1

## What landed

GitHub at full documented-operation parity on `main`, as its own PR, extracted from the paused
sweep branch `fm/cli-top50-sweep-resume2-r1` — plus the captain's order to restore the commands
that were classified `unsafe_or_disallowed` without his authorisation.

461 commands → 1147. 505 REST endpoint rows → 1224 (1220 documented + 4 synthetic
close/reopen). 1126 endpoints covered, 98 operation-blocked. 1086 commands reachable, every one
verified by running the binary.

## How it was extracted

Four commits cherry-picked in TDD order — the red surface test, the GET enumeration, the
`covered_by.writes` foundation fix, and the parity commit. The foundation fix is not
github-specific; it lands here because github is its first consumer, and it is what lets
`repo update`, `archive_repo` and `unarchive_repo` share one endpoint row.

Shared artifacts were **regenerated**, never hand-merged. The ledger the cherry-pick
auto-merged was thrown away, reset to `main`, and rebuilt with `surface-sync` — and came back
byte-identical to the branch's, which independently confirms the branch had derived it
correctly. Website catalogs, connector docs and golden transcripts were rebuilt from their own
generators. Every delta was diffed per connector and per transcript: all github.

The sweep branch's `data/cli-top50-fixed-schema-sweep-r1/PROGRESS.md` was dropped — sweep
bookkeeping, not github parity.

## The captain's classification order

`repo create` was `unsafe_or_disallowed` on both `main` and the sweep branch, so full parity
alone would not have made it runnable. But the framing in the original brief — is creating
safer than destroying? — turned out not to be the deciding question.

**The parity enumeration already made every one of these capabilities reachable under a
generated name.** `repos create-for-authenticated-user`, `repo delete-2`, `repo update`,
`secret set-2`, `secret delete-2`, `actions caches delete-2` are all `implemented`. Blocking
`repo create` and `repo delete` removed nothing; it only guaranteed that the destructive path
is the one an operator reaches by accident rather than on purpose.

So 7 commands were restored by pointing them at the write action their twin already uses,
inheriting the reverse-ETL contract unchanged — plan, preview, approval, execute, plus the
closed typed `--confirm destructive` on the three DELETEs. No gate was invented and none was
relaxed. Full table in `CLASSIFICATION-REPORT.md`.

Two needed more than a classification change, and got it: `repo archive`/`repo unarchive` are
new hook-pinned write actions, because a declarative PATCH cannot pin `archived` and a command
named "archive" that only archives when the caller separately remembers to say so is a command
that lies. `secret set` exposed a live defect — the already-`implemented` `secret set-2` could
only ever send an empty PUT body, because its write action declared just the path parameter.
Both now declare `encrypted_value` and `key_id`, with `encrypted_value` redacted.

Three are **not** restored, because each needs a runtime capability this PR does not add, and
marking them `implemented` would fail at runtime: `issue delete` (a `graphql_mutation`, and the
direct-write executor requires `rest_write`), `issue transfer` (GraphQL-only, no documented
REST endpoint), `pr revert` (no documented REST endpoint at all). bahmni's `documents upload`
is the same category — it has no write action, and its own notes name the missing
file-snapshot/SHA-256 approval binding.

`auth token` is held, as ordered. `api` is reported, not acted on.

## What a reviewer should look at hardest

- **`93de56c5a`** — the only behaviour change. Everything else is enumeration or generated output.
- **`internal/connectors/hooks/github/hooks.go`** — the one new Go path, ~5 lines, mirroring
  `closeResource`/`reopenResource`.
- **`CLASSIFICATION-REPORT.md`** — every classification change with the confirmation it now
  requires, and every command deliberately left blocked with the reason.

## Follow-ups this surfaced, deliberately not fixed here

1. **Repo-wide doc drift on `main`.** `pm docs generate` and `pm skills generate` want to
   rewrite all 551 connectors' `MANUAL.md`/`SKILL.md` and four `pm-*` skill files — field type
   annotations render as empty parens (`created_at()` rather than `created_at(string)`), and an
   `## Icon` section is missing. Reverted here to keep this diff github's; needs its own PR.
2. **Thin `record_schema`s on generated write actions.** `secret set-2` was `implemented` while
   able only to send an empty body. That is the shape issue #3899 (derive accepted operation
   params from the OpenAPI ledger) exists to fix; only the one blocking this order was fixed
   here. Other generated actions with empty `properties` — `repo2` among them — are worth a
   sweep.
3. **`pm github api`** needs the captain's decision before anyone else re-litigates it.

## The sweep branch needs a rebase

`fm/cli-top50-sweep-resume2-r1` carries these same four github commits. Once this merges, that
branch must be rebased onto the new `main` before it resumes — its github commits will be
already-applied, and its `operation_endpoint_ledger.json` will conflict against the regenerated
one. The sweep is paused at stripe, so this is cheap now. It must not be a surprise later.
