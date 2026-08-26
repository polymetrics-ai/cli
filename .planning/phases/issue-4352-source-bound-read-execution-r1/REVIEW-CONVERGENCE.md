# Firstmate exhaustive review convergence — source-bound read repair r4

## Immutable review custody

- Review source: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-source-bound-read-execution-reaudit-codex-r3/report.md`
- PR/issue: [#4356](https://github.com/polymetrics-ai/cli/pull/4356) / Refs #4352.
- Base: `main` at `b33983927d863032dac8220949990506e812937d`; immutable review SHA and PR head: `19b2bd8dc470d6fa92da1a500173c8c8c30ba59c`.
- Branch: `fm/cli-source-bound-read-execution-r1-continuation`, verified locally and remotely at the immutable SHA before this record.
- Delivery: direct-PR; only normal fast-forward commits to the existing PR branch. Do not merge, force-push, alter `main`, source locks, retained bytes, other worktrees, credentials, or provider state.
- GSD: inline/manual `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`. The repository contract forbids generic role spawning and no compatible isolated runtime is available; this records the fallback without weakening any gate.

## Architecture and complete lane map

`retained Asana source → sourceprojection → operations.source_operation → cli_surface source_operation/flags → hand parser → commandrunner → App credential boundary → engine OperationDirectRead or Connector.Read → authentication admission → requester`.

The review maps provider operations across ETL, reverse ETL, direct read, direct write, binary download, and binary upload. A source-complete operation must take its exact existing executor lane; a true gap remains declared with the source citation and exact `missing_foundation`. No generic route, method, body, opaque cursor, or write executor is admissible.

## Frozen initial finding set

| ID | Invariant and repair | Required red/green regression | State |
| --- | --- | --- | --- |
| AUDIT-001 | Projection must not emit raw paging `offset` or dead/colliding command `limit`; use only the declared closed page contract. | Red: current generated source-bound Asana flags contain raw controls. Green: projection/help/direct APIs omit them while `--page`/`--page-cursor` retain derived navigation. | green; final re-audit pending pushed SHA |
| AUDIT-002 | Caller-selected source-bound origin must fail before credential lookup, auth cohort/protected state, or provider I/O. Coordinate rather than duplicate #4351's generic input-ordering repair. | Red: a CLI/App observer sees credential/auth work before rejection. Green: origin error with zero credential reads, auth construction, and requester calls. | green; final re-audit pending pushed SHA |
| AUDIT-003 | Source-bound ETL source identity, method, path, records, pagination, and origin must be enforced at direct App `Read`/`ReadWithOutcome`, not only CLI preflight. | Red: direct stream literal drift reaches the later boundary. Green: each mismatch rejects before auth/requester use. | green; final re-audit pending pushed SHA |
| AUDIT-004 | Nineteen DELETEs and two no-body POSTs have the existing reverse-ETL/delete action foundation and must be promoted; absence of local action is not a foundation. | Red: 21 source-complete mutations remain declared blocked. Green: all reach the credential boundary through plan → preview → approval → execute; real named gaps remain. | green; final re-audit pending pushed SHA |
| AUDIT-005 | Fixed-100 isolated inputs must include every connector its cohort references. | Red: fixed-100 isolated tests name missing Asana input. Green: both tests and full generator package pass deterministically. | green; final re-audit pending pushed SHA |
| AUDIT-006 | Generated Asana docs/manual/help/website must state actual source-backed counts, implemented lanes, closed pagination, and only real named gaps. | Red: source-derived semantic assertions find stale counts/blockers. Green: regenerated artifacts assert current truth. | green; final re-audit pending pushed SHA |

## r4 resolution record before final re-audit

- **AUDIT-001:** `sourceProjectionReadParametersComplete` and
  `sourceProjectionSyncReadParameters` now exclude provider pagination by the
  shared paging classifier. The regression includes source `limit` and
  `offset`, commandrunner rejects both as unknown command flags, and built help
  renders only `--page` and `--page-cursor` under PAGE FLAGS.
- **AUDIT-002:** CLI calls `commandrunner.PreflightSourceBoundOrigin` after
  public `--config` parsing and before `withApp`/credential resolution. The
  engine implements the public-only source origin preflight for operations and
  streams. PR #4351 remains open at
  `fd400c501d99daa22210d42f736742706b4d8f1a`; its generic declaration-admission
  work is related but not a prerequisite and no local bypass was added.
- **AUDIT-003:** `engine.Read` and `ReadWithOutcome` invoke the declared
  source-bound stream proof before origin/authentication. The direct-read test
  rejects method, path, record, and pagination drift before I/O.
- **AUDIT-004:** 19 DELETE actions and `approve_access_request` /
  `reject_access_request` are promoted to 94 declared reverse-ETL actions.
  Their API endpoints use `covered_by.write`; no generic write executor was
  introduced. The remaining source rows carry exact foundation or
  not-applicable notes.
- **AUDIT-005:** fixed-100 creates a fully isolated Asana + GitHub source and
  website cohort. Its targeted regression and the final full generator package
  pass.
- **AUDIT-006:** the Asana definition, generated manual/skill/catalog, help,
  website data, operation evidence, and certification subject are regenerated
  from the actual 106 direct-read + 12 ETL + 94 write surface. The sole generic
  batch wrapper is explicitly not applicable; every remaining source gap is
  named.

## Lens coverage

Architecture/data flow, declaration reachability, happy/bad/edge behavior, CLI/App parity, secret-taint ordering, provider semantics, output integrity, and tests/evidence are complete in the independent audit. State/concurrency and retry/rate-limit/resume are not applicable: this repair introduces no state machine, persistence, goroutine, timer, retry, or rate-limit change. Each frozen finding has a behavioral red/green row in `TDD-LEDGER.md`.

## Fix and re-review protocol

Production work may now address only this frozen set and required generated artifacts. Each dependency group records exact files/symbols, intended regression, observed red failure, green result, and staged file list before its coherent fast-forward checkpoint. A new finding is classified `initial_snapshot_miss`, `fix_created:<sha>`, `requirement_changed:<decision>`, or `new_external_evidence:<source>` and requires another cycle. Final acceptance is a fresh independent Codex audit of the whole original PR change plus repair delta at the pushed code SHA; a later artifact-only audit commit may name that code SHA only if it changes no behavior.

## Fresh independent audit — repair SHA `35888c27c51b8c1168c2d6f08ffa505f5ffdb6bd`

The required independent `codex review --base
b33983927d863032dac8220949990506e812937d` completed after the first repair
push. It found two actionable `fix_created:35888c27` findings, both within the
already frozen AUDIT-001/AUDIT-006 scope. The tracked-skill parity test also
provided an AUDIT-006 `initial_snapshot_miss` red result.

| Finding | Classification | Red evidence | Required green result |
| --- | --- | --- | --- |
| F3 / AUDIT-001 | `fix_created:35888c27` | 52 source-bound next-URL reads have no derived `limit` on page one; a returned `?limit=...&offset=...` continuation fails the closed cursor admission. | The source descriptor derives `next_url` plus its exact `limit` size and `offset` continuation controls; neither becomes a raw command flag, page one sends the bounded limit, and `--page-cursor` admits the provider continuation. |
| F4 / AUDIT-006 | `fix_created:35888c27` | Eight source-complete implemented reads still render historical `Blocked until …` notes. | Promotion clears that historical blocker note only for the now source-complete command; truly missing foundations remain exact named gaps. |
| F5 / AUDIT-006 | `initial_snapshot_miss` | `TestSkillsGenerateMatchesTrackedSkills` fails because `docs/skills/pm-asana/SKILL.md` still lists raw `--limit` and `--offset`. | `pm skills generate --dir docs/skills` updates the tracked skill and generated docs/website/manual validation is clean. |

No source lock, retained provider byte, credential, provider state, or unrelated
worktree is in scope for this convergence slice. The follow-up must be pushed
normally, then independently re-audited at its exact remote SHA.

## Follow-up resolution and pre-push proof

- **F3 / AUDIT-001 Green:** `TestOperationDirectReadNextURLUsesClosedLimitOffsetContinuation`
  proves page one sends `limit=100`, a provider-issued `limit`/`offset`
  continuation is accepted only through `--page-cursor`, and the generated
  direct command owns neither raw flag. The full `./cmd/connectorgen` package
  passed in `182.924s`; the final source import check verified all 249 Asana
  operations.
- **F4 / AUDIT-006 Green:** source-complete reads clear only inherited
  historical `Blocked until …` notes. The final declaration scan finds neither
  raw direct-read paging flags nor historical blocker notes.
- **F5 / AUDIT-006 Green:** tracked Asana skill/manual, root-help goldens, and
  website catalog were regenerated. `go test -timeout 20m -count=1
  ./internal/cli` passed in `420.427s`; connector docs validation and 34
  website script tests passed.
- The whole-tree connector-boundary gate initially rejected Asana-specific
  shared projection branches. The final implementation makes the source-gap
  annotation generic; rerun result is clean (317 files, 553 connectors).
- The final credential-bound binary census is `212/212` implemented Asana
  commands at `missing --credential` (exit 1), with no configured credential
  or provider I/O. `validate` is clean for 553 connectors, surface-sync has
  zero corrections, operation evidence is current at 1,774 rows/5 rollups with
  fixed-100 passed, runtime preflight/canon are green, and certification
  subject/matrix/candidates/sweep are current.

The next and only acceptance action is a normal push of the reviewed commit to
the existing #4356 branch followed by a fresh independent audit of that exact
remote SHA. Do not merge.

## Fresh-audit F6 / AUDIT-002 convergence

- **Classification:** `fix_created:07251df15c904cad0f91a43724810dffa133b81d`.
- **Red:** the exact-SHA independent audit found that CLI preflight considered
  only `--config`. A selected credential's persisted `base_url` was merged only
  by `ResolveConnectorCredential`, after vault access and authentication-cohort
  construction, and could then be refused by the engine.
- **Green required:** before `withApp` and before any vault/authentication
  work, read only the selected credential's public persisted configuration,
  overlay command configuration with the normal caller-wins order, and run the
  existing source-bound origin preflight. The regression must use a persisted
  Asana credential configuration and prove the declared-origin error, with no
  provider request.
