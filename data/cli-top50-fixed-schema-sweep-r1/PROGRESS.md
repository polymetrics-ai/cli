# ⇒⇒ HANDOFF — READ THIS FIRST, THEN THE WORK ORDER BELOW ⇒⇒

**Rewritten 2026-08-07 by the worker that delivered github. A fresh worker continues SERIALLY on the
same consolidated branch (firstmate decision, captain caps the fleet at 2–3 lanes: parallel branches
would collide on the shared generated artifacts). Nothing is uncommitted; everything below is pushed.**

## Where the work is

| | |
| --- | --- |
| Worktree | `/Users/karthiksivadas/.treehouse/cli-83d592/20/cli` (was `…/42/cli`) |
| **Branch** | **`fm/cli-top50-sweep-resume2-r1`** — cut from the tip of `fm/cli-top50-sweep-continue-r1` (`b1ceedc32`), which was itself cut from `fm/cli-top50-sweep-consolidated` @ `bedb9b0a3`. **History is continuous across all three names**; only the name changed, because each dispatch scaffold named its own branch. This is the branch the single sweep PR opens from. |
| **HEAD** | Use the **branch tip** — it is authoritative. `git fetch origin && git checkout fm/cli-top50-sweep-resume2-r1 && git reset --hard origin/fm/cli-top50-sweep-resume2-r1`. A commit cannot record its own hash, so do not treat any sha written here as the tip. |
| Cut from | `origin/main` @ `5d43d7c00` (post-Mailchimp #3562) |
| Model | **ONE consolidated sweep PR** (captain, ordered repeatedly; firstmate re-confirmed 2026-08-07). PR **not yet opened.** |
| Master plan | `MASTER-PLAN.json` (sibling of this file) — all 30 counts, hazards, issue links. **Its counts are candidates, not truth**: github's 1220 held, workday-rest's 920 did not (see below). |
| Tooling | **`tools/` (sibling) — NOW COMMITTED.** The previous worker's `tools/` was untracked and died with its worktree, so this handoff pointed at nothing and the approach was rebuilt from scratch. Read `tools/README.md` first: it holds both generators, the reachability probe, and the exact artifact re-fetch commands with their expected byte counts. |

**Do NOT open the PR yet.** Shared artifacts (endpoint ledger, website catalogs, golden transcripts,
docs) are regenerated **once at the very end**, not per connector.

### ⚠️ Known-red on this branch — both are expected, neither is yours to "fix"

1. **The current connector's own surface test is red until its last slice lands.** Right now that is
   `TestWorkdayRESTDocumentedSurfaceIsComplete`. Every *shared* gate is green:
   `connectorgen validate` 551/0, `surface-sync --check` clean, runtime preflight passing. **Do not
   weaken a red test to make the package green** — it goes green when the connector lands.
2. **`TestGoldenTranscripts/root_bare_manual` fails**, and has since before github. Verified
   pre-existing: chatwoot and gmail already added command surfaces the committed transcript predates,
   and github added ~686 more commands. It is discharged by the **end-of-sweep regeneration**, which
   must happen before the PR merges. Carried as a known-unmet item in every connector's
   `VERIFICATION.md`. **Do not regenerate it per connector** — finding F6: a per-connector
   `pm docs generate` run rewrites ~1,034 files of pre-existing `main` drift.

## ⇒ RESUME POINT — workday-rest slice 2 of 5 (its red is now observed at **907** — not 920, not 916)

> **Updated by the slice-2 worker.** The section below is slice 1's account and is kept because its
> reasoning about the cross-service duplication is correct and still load-bearing. **Its number is
> not.** The artifact was re-fetched (manifest HTTP 200, **617,538 bytes**, byte-identical; all 52
> specs re-fetched) and the derivation reproduced independently to 920 raw → 916 — confirming slice 1
> exactly to that point — and then collapsed a **second** time to **907** on nine query-string
> variants slice 1 never looked for. **See findings 34 and 35.** The red test now asserts 907, plus
> the stricter requirement that each collapsed variant's base endpoint is *present*.

**`github` is COMPLETE and pushed.** Its full documented surface is enumerated at **1220 REST + 4
GraphQL**, and `TestGitHubDocumentedRESTSurfaceIsComplete` — red since slice 1 — now **passes**. The
whole `cmd/connectorgen` package is green. See `.planning/phases/github-parity-sweep-r1/`
(`SUMMARY.md`, `VERIFICATION.md`, `TDD-LEDGER.md`, `RUN-STATE.json`).

**Branch note:** work continued on **`fm/cli-top50-sweep-resume2-r1`**, cut from the tip of
`fm/cli-top50-sweep-continue-r1` (`b1ceedc32`). History is continuous, so the one consolidated PR is
unaffected; only the branch name changed, because the dispatch scaffold named the new branch.

### workday-rest slice 1 is done, and its count is NOT the planned 920

`.planning/phases/workday-rest-parity-sweep-r1/` holds the red test, `DERIVED-OPERATIONS.json`,
`PLAN.md` and `RUN-STATE.json`. **Read `PLAN.md`'s hazards 1–3 before authoring a single row.**

**The count is 916, not 920.** The 52 service specs hold 920 raw rows, matching the master plan
exactly — and then four of them turn out to be the *same endpoint documented twice*: `Custom Object
Data (multi-instance) v2` and `(single-instance) v2` are two directory entries declaring the
**identical** `servers` URL. Deduped on the resolved `(method, base+path)`, it is **916**.

**Why it was missed, and why it will be missed again:** 49 of the 52 specs are Swagger 2.0 with
`basePath`; **3 are OpenAPI 3.0.1 and have no `basePath` at all** — they carry `servers`. A reader
that looks only at `basePath` records an EMPTY base for those three and they appear to collide.
Those same three are the only specs *not* stamped `20260727`, so the master plan's derivation note
("every spec filename is date-stamped 20260727") **is wrong**, and wrong exactly where it changes the
count.

**And the 4 rows it already has are NOT among the 916.** They point at
`/ccx/api/hcm/v1/{tenant}/workers|organizations|jobs`. The current directory publishes **no `hcm`
service and no `/ccx/` path anywhere** — worker resources live under `staffing/v7`,
`absenceManagement/v5`, `compensation/v3`, `performanceEnablement/v5`, `timeTracking/v5` and
`api/common/v1`. They are not in the archived list either (that holds only older *versions* of the 52
listed services). So they cannot be counted as documented, **and they must not simply vanish** —
deleting them deletes three shipped, schema- and fixture-backed streams inside a parity commit. The
red test pins them by name and **counts them apart**, so the re-point-or-supersede call is made
deliberately rather than by arithmetic. See `PLAN.md` hazard 7.

`RUN-STATE.json.next_action` names the resume point. The bundle has **4 rows** and no
`cli_surface.json`, `writes.json` or `operations.json` — the largest before/after gap in the sweep.
`capabilities.write` is `false` and will have to become `true`.

### Shape of the github delivery, as a template for the next 20

Slices that actually worked, against the 5-slice plan that did not survive contact: **slice 2 and 3
cannot be separated.** An `api_surface` row needs a disposition, and a `covered_by` disposition must
name a command that already exists, so rows and commands land together. What worked was
**reads in one commit, then mutations in another**, each green on all shared gates:

| | rows | commands |
| --- | ---: | ---: |
| GETs | 377 new | 364 direct reads + 2 binary downloads + 11 blocked |
| mutations | 342 new | 322 write actions + 2 POST-shaped reads + 18 blocked |

Budget honestly: **github alone consumed most of one worker context**, and it is one of 21.

### One shared-code change landed with github, and every later connector inherits it

**`covered_by.writes` now exists** — a plural array mirroring `covered_by.direct_reads`. Before it,
a bundle modelling several write contracts over ONE documented path had nowhere to record the
others, which is why github shipped `PATCH .../issues/{issue_number} (close)` and three siblings.
**If a connector ahead of you has two write actions on one endpoint, use the array — do not invent a
variant path.** notion (merged #3894) still carries the old defect and can now be converted; that is
deliberately not folded in here.

## Done — 5 on the branch, plus 3 already resolved elsewhere

| Connector | Ops | State |
| --- | ---: | --- |
| chatwoot | 148 | ✅ on branch (folded from its own lane, verified: 101 covered + 47 blocked) |
| gmail | 79 | ✅ on branch `d1bacbfe1` (66 covered + 13 blocked; `pm gmail` did not exist before) |
| lever-hiring | 106 | ✅ on branch `0829543e6` (60 covered + 46 blocked) |
| greenhouse | 138 | ✅ on branch (127 covered + 11 blocked; `pm greenhouse` did not exist before) |
| help-scout | **144** | ✅ on branch (139 covered + 5 blocked; ledger said 146, derivation 145 — see finding 22) |
| **github** | **1220** | ✅ **on branch** (1126 covered + 98 blocked, +4 GraphQL counted separately; 461→1147 commands, 231→553 write actions; 1079 verified reachable by running the binary) |
| gorgias | 114 | ✅ **separate open PR #3896** — leave it alone, do not fold |
| notion | 51 | ✅ MERGED #3894 |
| gong | 69 | ✅ MERGED #3895 |

## Deferred — do not work these

- **mixpanel (104)** — own task. Its `api_surface` collapses operations into comma-joined rows
  (23 of 47 rows, one holding TEN), so it needs a 47→104 **restructure**, plus it spans five base
  URLs. Red test + WIP preserved on `fm/cli-mixpanel-parity-sweep-r1`. **Its WIP deletes
  `schemas/annotations.json` — that deletion is WRONG, Mixpanel still documents annotations. Do not
  fold it.**
- ~~**marketo**~~ — **UNBLOCKED 2026-08-07 by captain ruling**, see `DECISION-marketo-scope.md`
  (sibling of this file, now on the branch). Marketo's surface is the **full** surface **including
  Adobe's v2 Asset API**. Two conditions both apply and are the operative test: the endpoint must
  **genuinely belong to Marketo under Adobe** (not merely sit adjacent in Adobe's wider docs), **and**
  it must be in the **current published spec** (retired endpoints are out regardless of history).
  Anything failing either condition is excluded **with the reason recorded**, never silently dropped.
  **The count is NOT settled.** The ruling resolved scope, not a number. Do **not** adopt 367 because
  it is the largest — re-derive under the stated rule and let the count fall out, then explain the
  delta against **all three** prior derivations (320 HTML reference · 327 earlier lane's spec ·
  367 fresh recount; the ledger's own 322 matches none of them), saying for each whether it was
  missing asset operations, counting retired endpoints, or double-counting query-string variants —
  the last being exactly the error found in lever-hiring.
  Re-entered in the work order at its size position (see the ordered list below).

## Remaining work order — 21 connectors, ⇄ LARGEST FIRST

Commit **and push** after every single connector, and tick it here in the same step.

> ### ⇄ ORDER REVERSED — LARGEST FIRST (captain, 2026-08-07). Eat the frog.
>
> **This supersedes the smallest-first order.** The four giants land FIRST, not last. Re-sorted by
> re-derived operation count descending from `MASTER-PLAN.json`.

~~github 1220~~ ✅ · `workday-rest 907 (NOT 920, NOT 916 — re-derived twice, see findings 34-35) · zendesk-support 625 · jira 616 · stripe 589 · linear 538 ·
chargebee 438 · **marketo ~320-367 (count TBD)** · square 334 · bitbucket 331 · bamboo-hr 311 ·
monday 292 · trello 261 · asana 249 · front 244 · xero 235 · intercom 231 · quickbooks 198 ·
segment 197 · twilio 197 · google-ads 163`

Already delivered on this branch and NOT in the order above: ~~greenhouse 138~~ ✅,
~~help-scout 144~~ ✅ (both landed under the previous smallest-first order).

**marketo is slotted on its UPPER candidate (367), between chargebee 438 and square 334.** Its true
size is unknown until re-derived under the captain's scope ruling, and the three candidate counts
(320 / 327 / 367) straddle that slot. If the derivation lands materially outside the range,
**re-slot it by its real count** rather than forcing it through in place.

### What largest-first changes about the method — one addition, nothing removed

Every existing rule stands unchanged: red test **written AND run AND captured** before authoring ·
one commit per connector · **push after each** · no paging flags hand-authored · GSD planning
evidence under `.planning/` (`PLAN` + `RUN-STATE` + `TDD-LEDGER` + `SUMMARY` + `VERIFICATION`).

**Added, because the giants will not fit one context:** for a connector too large to finish in one
context, **commit and push incremental progress inside it** rather than holding one giant
uncommitted change. A workable slicing for github 1220 / workday-rest 920:

1. red test + derived-operation inventory (`DERIVED-OPERATIONS.json`) + `PLAN.md`/`RUN-STATE.json` —
   commit, push;
2. `api_surface.json` dispositions — commit, push;
3. streams / writes / operations — commit, push;
4. `cli_surface.json` + full binary reachability sweep — commit, push;
5. `SUMMARY.md` + `VERIFICATION.md` — commit, push.

Each slice must leave the tree **green on its own gates** (`connectorgen validate`, whole
`cmd/connectorgen`, runtime preflight) — an intermediate commit may leave the connector's own red
test still red, and that is fine and honest, but it must never leave a **different** connector or a
shared gate broken. Say which slice you stopped on in `RUN-STATE.json.next_action`.

Per connector: **write the red test against the REAL bundle → run it → watch it fail → record the
verbatim failure** in `RUN-STATE.json` + `TDD-LEDGER.md` → author to green → `connectorgen validate`
+ **whole** `cmd/connectorgen` package + runtime preflight → build and **run the binary** over every
command → one commit → push → tick.

## Open findings and hard-won constraints — all of these will bite you

1. **The red test must genuinely fail first.** If you author before testing (I did once, so did the
   gorgias agent), `git stash push -u` the bundle, write the test, observe red against the restored
   real bundle, commit red, then `git stash pop`. `tools/check_red_observed.py <name>` enforces this
   and rejects placeholder text.
2. **⛔ Never hand-author paging flags** (`page`, `per_page`, `limit`, `offset`, `cursor`,
   `pageToken`, `maxResults`, `page_size`, …). A foundation lane derives them from each connector's
   pagination spec. chatwoot's 6 were removed and verified not required. **If one ever seems
   genuinely required, record it here and escalate — do not author it.**
3. **The operation endpoint ledger is the binding gate**, not the URL builder.
   `deriveOperationDirectReadEndpointLedger` only emits an entry when `operation.rest.path` **equals
   an `api_surface` row verbatim**. Consequence: an out-of-base path cannot be both correctly
   constructed and ledgered. This is why **chatwoot's 47 blocks are CORRECT** (probed and settled —
   my earlier suspicion that they were wrong was refuted). **Streams do NOT go through this ledger**,
   which is why multi-base *streams* work while multi-base *direct reads* do not.
4. **`api_surface.json` schema constraints** learned the hard way on gmail — expect these:
   `delete` is an object not a boolean · `operation.dependency` is NOT an allowed field (fold it into
   `notes` as `"Named dependency: …"`) · `operation.model` is a fixed enum
   (`direct_read binary_read sensitive_reverse_etl admin_reverse_etl destructive_action
   local_workflow duplicate deprecated disallowed`) · `output_policy` is a fixed enum (binary uses
   `binary_file_bounded`) · `source_cli` is an object not a string · **implemented reverse-ETL
   commands must expose a flag for every required record field**, and a required field with no
   scalar leaf means `availability: partial`, not `implemented`.
5. **Binary downloads** use `intent: binary_download` + `kind: binary_download` +
   `engine.OperationBinaryDownload`. GET-only.
6. **F4 — docs generator drift.** `./pm docs generate --dir docs/cli` rewrites ~1,034 files of
   pre-existing `main` drift. Revert every non-connector path:
   `git checkout -- $(git diff --name-only docs/connectors/ | grep -v '^docs/connectors/<name>/')`
7. **F5 — a connector can have MORE THAN ONE surface test.** gong had two and a targeted `-run`
   missed the second. Always run the **whole** `cmd/connectorgen` package before pushing.
8. **Inspect every regenerated diff BY OBJECT, not by line** — 551 connectors before and after, none
   added, none removed, exactly one changed. This is not delegable; it is the safeguard.
9. **The ledger errs in BOTH directions.** notion +1, intercom **−93** (324→231, corroborated three
   ways), lever-hiring **+39** (67 is wrong at every point in time). A reconciling total is evidence;
   a non-reconciling one is a finding. Never adopt the ledger number.
10. **Webhook EVENTS are excluded** from the operation surface and must not appear as `api_surface`
    rows. Webhook **MANAGEMENT** endpoints stay in scope and are counted.
11. **linear 538 vs 539** — one Mutation field; resolve at red-test time by diffing the field-name
    sets. Non-blocking.
12. **Three `connectorgen batch` blockers** (OAS 3.1 + top-level `webhooks`): notion, chargebee
    (227 events), bamboo-hr (6 events). Hand-author around the gate; do not relax it.
13. **github's 270 webhook events live under `x-webhooks`**, not the standard block — any inventory
    checking only `webhooks` records zero for the largest connector.
14. Spend limits killed two sub-agents earlier. **Preserve and push partial work immediately** if an
    agent dies; five inherited lanes were one teardown from permanent loss before being pushed to
    `…-preserve-20260807` branches.
15. **Re-fetch the artifact and check its BYTE COUNT against `MASTER-PLAN.json`.** greenhouse's
    re-fetch was 1,636,662 bytes — identical to the recorded derivation, which proves it is the same
    artifact rather than a lookalike, and lets you reproduce the derivation instead of trusting it.
    The whole extraction reproduced to the operation, so 138 was confirmed, not adopted.
16. **A provider's own docs markup can silently eat an operation.** greenhouse's
    `DELETE /v1/tags/candidate/{tag id}` carries a stray unescaped `&#39;` before the URL *and* a
    path placeholder holding a literal space. A regex ending at `[^<\s]+` truncates it to
    `/tags/candidate/{tag`; a regex ending at `[^<]+` recovers it. **The shipped bundle had dropped
    this row** — the count was 129, not 130 v1 rows. Terminate on `<` and strip, never on whitespace.
17. **Counting `<h2>` headings under-counts when one heading documents two versions.** greenhouse has
    3 sections that each declare a deprecated v1 operation *and* its v2 replacement (`Destroy
    Openings`, `Create Scheduled Interview`, `Update Scheduled Interview`). 135 headings, 138
    operations. Always count the `HTTP Request` **declarations**, never the section headings.
18. **Deprecated is a disposition, not an exclusion.** greenhouse parked its 3 deprecated v1
    mutations under a legacy `excluded` stub. The counting policy counts deprecated operations, and
    `excluded` is not one of the three dispositions this sweep accepts. They are now `blocked` with
    `model: deprecated`, each naming its **v2 successor** as the dependency — which is checkable, and
    is not the same thing as "we chose not to".
19. **`surface-sync` fills nothing for a stream+write-only connector, and that is correct.** It only
    derives metadata for operation-backed `direct_read`/`direct_write`/`binary_download` commands.
    greenhouse has none, so `--check` is clean and `operation_endpoint_ledger.json` is untouched —
    which is also why greenhouse's ledger delta is trivially confined to greenhouse.
20. **Verifying 127 commands by running the binary takes ~1.3 s each — serialise it and you time
    out.** Fan the `--help` invocations out with `xargs -P 8`; the whole sweep then takes seconds.
    Do not skip this step: authoring a command is not evidence that it routes.
21. **Follow-up recorded, deliberately not done here: greenhouse models 43 singleton detail GETs as
    config-parameterised streams** (`/candidates/{{ config.candidate_id }}`) rather than as
    `direct_read` operations. Parity holds either way and all 43 are reachable, so converting them
    would have deleted shipped, schema- and fixture-backed functionality inside a parity commit.
    Instead each got a **required named flag bound to its config key**
    (`pm greenhouse candidate list --candidate-id 42` rather than `--config candidate_id=42`), which
    is the reachability win without the restructure. Whoever revisits greenhouse should model them as
    direct reads. **This is the stream-vs-direct-read judgement for this connector; it is a choice,
    not an oversight.**
22. **⚠️ THE SWEEP'S OWN DERIVATIONS CAN DOUBLE-COUNT QUERY-STRING VARIANTS TOO.** help-scout came in
    at 146 (ledger, = doc pages), 145 (sweep derivation) and **144** (strict policy). The derivation
    deduped on the **literal example** path, which catches a variant only when the two example URLs
    happen to be byte-identical. It caught `original-source` (documented once per `Accept` header,
    same literal line) and **missed** `DELETE /v2/customers/{customerId}` vs
    `…/100?async=true`. **Always dedup on the TEMPLATED path the provider publishes**, never on the
    example URL. This is the lever-hiring defect class recurring in a different provider's docs, and
    it will recur again. Put `if strings.Contains(ep.Path, "?")` in every red test.
23. **A collapsed count can silently misname the surviving thing.** When two doc pages collapse to
    one operation, whichever page wins the dedup also supplies the title — and therefore the write
    action's name. help-scout's async page won, which would have shipped
    `delete-customer-asynchronously` on a path that performs the **synchronous** delete. **Name
    actions for the endpoint, and re-express the losing page's behaviour as a flag** (here `--async`
    with `omit_when_absent`), so collapsing the count never loses a documented behaviour.
24. **A wildcard row is not an operation.** help-scout shipped `GET /v2/reports/*` as one excluded
    row standing for **33** distinct report endpoints. Add `if strings.Contains(ep.Path, "*")` to
    every red test.
25. **Do not assume `id` is the primary key.** help-scout's user statuses key on `userId`;
    `connectorgen validate` rejects an assumed key as `primary_key_missing`. Take stream primary keys
    from the record the provider actually returns.
26. **A `direct_read` command needs NO `operations.json` entry.**
    `commandrunner.validateDirectReadCommand` accepts exactly one GET `api_surface` endpoint plus an
    explicit `output_policy`. Only **operation-backed** reads go through the endpoint ledger, which
    is why help-scout's 49 direct reads add zero ledger entries. Simplest reachable shape for a
    detail GET; gmail uses it too.
27. **`api_surface` rows carry the DOCUMENTED path; streams/writes carry base-relative paths.**
    `engine.normalizeDirectReadPathForBaseURL` strips a base path prefix the request path actually
    starts with — so `/v2/…` resolves correctly against a `/v2` base, and `/v3/…` does not. That
    asymmetry *is* the out-of-base rule, and it is worth checking rather than assuming per connector.
28. **The GSD contract in `AGENTS.md` outranks a dispatch brief that omits it** (captain, 2026-08-07).
    Every connector needs `PLAN.md` + `RUN-STATE.json` + `TDD-LEDGER.md` **and** `SUMMARY.md` (with
    the `coverage:` frontmatter block) + `VERIFICATION.md`, with the lifecycle and
    `scripts/gsd sources` provenance recorded in `.planning/traces/`. Gate:
    `GSD_BASE_REF=origin/main ./scripts/verify-gsd-workflow`. Note `scripts/gsd` is a **Node** script
    — `bash scripts/gsd` dies with a syntax error; use `./scripts/gsd`.
29. **Host hazard: a full disk corrupts the SHARED 55 GB Go build cache.** `go test` then fails to
    *build* with `could not import <pkg> (open …/go-build/XX/<hash>-d: no such file or directory)`,
    **a different missing entry on each retry**, so it looks like a code error and is not
    self-healing while disk is tight. It recovered on its own once space was freed. If it happens
    mid-red, **commit the test with `red_confirmed: false` and an explicit blocker rather than
    authoring on unobserved evidence** — that is what help-scout did.

30. **A namespace miss exits 0.** `pm <connector> <nonsense> --help` renders the connector's group
    help and **succeeds**, which is the documented namespace behaviour. **Exit status therefore
    proves nothing about reachability.** The first github probe checked only the exit code and passed
    749 commands it had not actually verified. Assert the rendered `NAME` line reads
    `pm <connector> <path> - …`. This is the gmail lesson in a form that silently passes.
31. **`covered_by.write` was singular, and that is what forced synthetic variant paths.**
    `covered_by.writes` (plural) now exists, mirroring `direct_reads`. Use it when two write actions
    land on one documented endpoint. Deleting the extra actions is usually NOT available: github's
    close/reopen bodies are assembled by a Go hook switching on the action **name**, and certify's
    create/cleanup pairing binds `create_issue` to `close_issue`. **notion (merged #3894) still ships
    the old defect** (`(body=markdown)`, `(parent=data_source_id)`) and can now be converted.
32. **A spot-pin can name an endpoint the provider does not document.** github's slice-1 red test
    pinned `GET /enterprises/{enterprise}/copilot/billing/seats`; Copilot billing is org-scoped and
    that path is nowhere in the artifact. Once the GET surface was complete the pin could never pass.
    **Replace such a pin with a real endpoint from the same scope — never delete it** — and write the
    substitution into the test. Check every spot-pin against the artifact when you write it.
33. **A `rest_read` POST needs BOTH `body_schema` and `content_type: application/json`.** Two
    separate validate rules, caught one run apart. Fix them in the generator and regenerate; never
    hand-patch a generated bundle.
34. **⚠️⚠️ DEDUPING ON THE RESOLVED PATH IS NOT ENOUGH — DEDUP ON THE TEMPLATED PATH TOO.**
    **workday-rest is 907, not the 916 slice 1 derived and not the 920 the master plan recorded.**
    Slice 1 deduped on the resolved `(method, base+path)` pair, which correctly caught four
    custom-object rows published by two service modules (920 → 916) — and **never looked for a `?`**.
    Nine of the 916 are query-string variants of endpoints already counted, each published by
    Workday as its own Swagger path key with the query string baked in, each with its base-path
    sibling documented separately. **Fifth appearance of this defect class** (notion, lever-hiring,
    help-scout, github, now workday-rest), and the first where **both** collapses were needed in one
    connector — passing one dedup is what made the second look unnecessary. Finding 22 already said
    "put `if strings.Contains(ep.Path, "?")` in every red test"; the test **had** that guard, and the
    derivation feeding it did not. **Run the derivation through the red test's own rules before
    trusting its count.**
35. **The query-string variants ARE the binary-detection judgement in disguise.** Six of
    workday-rest's nine (`?type=viewContent` / `getFileContent` / `viewFile`) are the **binary** mode
    of an attachment endpoint whose default mode returns JSON metadata — Workday's own procurement
    summary says "Retrieves the metadata **or the attachment content**". So collapsing the count is
    not the end of it: modelling the survivor as `direct_read` alone drops the file fetch, and as
    `binary_download` alone drops the metadata read. **Hang both off ONE row with
    `covered_by.direct_reads` (plural)** — this is the second connector to need github's plural-array
    foundation fix, which confirms it generalised rather than solving one connector's problem.
    Seven of the nine carry an **empty summary**, which is the provider telling you it is an addendum
    to the base row and not an operation; the two that carry text describe a *behaviour*
    ("…to archived or un-archived") and become a flag, per finding 23.

## Honest note on pace

Each of the three delivered connectors hit real structural surprises well beyond "mechanical
generation" — gmail's six schema rejections, chatwoot's ledger constraint, lever-hiring's
simultaneous double-count and missed endpoint. Budget for that residue; it is not going away, and
github (1220), zendesk-support (625), jira (616) and stripe (589) are still ahead.

---

# cli-top50-fixed-schema-sweep-r1 — progress

Durable resumable state for the fixed-schema top-50 connector parity sweep. **Read this file first.**
Updated immediately after each connector, never batched.

## ⇒ `MASTER-PLAN.json` is now THE GSD artifact (captain efficiency ruling, 2026-08-07)

Per-connector GSD reasoning was burning tokens on work that is the same shape 30 times over. The
sweep is restructured:

1. **ONE upfront planning pass** derives every remaining connector's operation count from its
   provider artifact in a single parallel fan-out → `MASTER-PLAN.json` (sibling of this file).
2. **Each connector's `.planning` slice is generated MECHANICALLY** from that master by
   `gen_planning_slice.py` — no fresh reasoning pass per connector. This satisfies
   `gsd-workflow-evidence` at near-zero cost.
3. **Implementation parallelises** across Sonnet sub-agents, several connectors in flight.
4. **Validation parallelises too** — start a connector's no-mistakes run and move straight to the
   next; do NOT walk them through one at a time. **Never restart the shared daemon.**

### Four limits that do NOT move, and how each is enforced

| Limit | Enforcement |
| --- | --- |
| Each connector needs its OWN red test that **genuinely fails first** against its real bundle | `gen_planning_slice.py` deliberately generates **no** red test and **no** `TDD-LEDGER.md` — there is no optimistic template to fill. `RUN-STATE.json` ships `red_confirmed:false, red_failure:null`, and **`check_red_observed.py` refuses to clear a connector** until that field holds real captured `go test` output (it pattern-matches `--- FAIL`, `<file>_test.go:<line>`, `want N` and rejects placeholders). |
| Per-connector findings are **not** pre-planned | Every generated `PLAN.md` carries a banner: if the connector surprises you, **STOP and record it** rather than forcing it into the batch shape. |
| Each connector gets its **own PR with an issue-first body** | `require-linked-issue` is a required check; validate with `go run ./cmd/prissueguard` before opening. |
| **This lane inspects every regenerated diff itself** | Not delegable. It is the safeguard, not a formality. |

Already-built work **stands** and is not redone to fit the new shape: notion (merged #3894), gong
(PR #3895), gmail (red committed).

## Owner / lane

- **The sweep branch is now PUSHED** to `origin/fm/cli-top50-fixed-schema-sweep-r1` (2026-08-07) —
  it was previously worktree-only, the same hazard that nearly lost the five inherited lanes.
  It carries the dynamodb red test (`bdb6efc0a`), which is **permanently red by design** now that
  dynamodb is removed by correction 3. **Never merge the sweep branch**; it is a work ledger.
  Per-connector PRs are cut fresh from `origin/main`.
- Lane: `cli-top50-fixed-schema-sweep-r1`, branch `fm/cli-top50-fixed-schema-sweep-r1`,
  worktree `/Users/karthiksivadas/.treehouse/cli-83d592/42/cli` (moved from `…/49/cli` when the
  pane was rebuilt on 2026-08-07; no work was lost, this file was the recovery point).
- One connector per PR. Never batch connectors into one PR.

## Scope: 30 connectors

Original brief listed 35. Three captain coordination corrections applied on 2026-08-07:

1. Correction 1 removed 8 connectors owned by other live lanes.
2. Correction 2 folded 5 of those back in (their individual lanes were stopped to avoid paying
   twice to read the same provider docs) — these are **resumed from their existing branches**,
   not restarted.
3. **Correction 3 (classification) removed `dynamodb` and `postgres`** — both are
   `integration_type: database`, not `api`. See "Correction 3" below.

Net: **30 connectors**.

## Correction 3 — database-typed connectors removed (captain ruling, 2026-08-07)

**The captain caught a classification error before dynamodb was implemented. He is right, and the
audit found a second instance.**

DynamoDB's wire protocol happens to be HTTP (`POST /` selected by `X-Amz-Target`), which is what
made it look API-shaped. But `integration_type` is the authority, and it says `database`.

### Verified against the metadata, not taken on report

| Fact | dynamodb | postgres |
| --- | --- | --- |
| `integration_type` | `database` | `database` |
| Native Go package | `internal/connectors/native/dynamodb` | `internal/connectors/native/postgres` |
| `capabilities.write` | `false` | `false` |
| `capabilities.cdc` | `false` | `false` (yet `cdc.go`, `cdc_decode.go` exist) |
| `capabilities.dynamic_schema` | `false` (yet DynamoDB is schemaless) | `true` |

**Repo-wide audit: 551 connectors, exactly 2 typed `database` — dynamodb and postgres.** Every one
of the other 28 connectors in this sweep is `api`. This confirms the captain's "only TWO
database-typed connectors" exactly, and the error does not recur anywhere else in my list.

**One decisive fact the ruling did not yet have.** dynamodb's own `conformance.reason` states it is
a **Tier-3 native package** whose "Check/Catalog/Read are hand-written Go … never routed through
`engine.Check`/`engine.Read`", and that its bundle files "exist only to keep identity/spec/docs/
schema uniform with every other connector." So this sweep's entire machinery — declarative
`api_surface.json` dispositions, engine-executed commands, runtime preflight — **does not route for
dynamodb at all**. Authoring 58 API-shaped commands would have built against an architecture the
connector does not use. That is a stronger reason to remove it than the type field alone.

### postgres — the second instance, and why its ledger count was `unknown`

postgres sat at order #30 recorded `unknown — must evidence or record unenumerable`. That `unknown`
was **never a research gap**: PostgreSQL has no documented REST/RPC operation surface to enumerate.
Its surface is SQL, discovered from `information_schema`. Under the sweep's counting policy — "one
documented REST/RPC action or public top-level GraphQL query/mutation field" — the question does not
apply. The `unknown` was a symptom of the same misclassification, not missing evidence.

The other three `unknown` records (**linear, quickbooks, workday-rest**) are all `integration_type:
api` and **remain in scope**; they genuinely require artifact research and stay last in the order.

### Disposition

Both **removed from this sweep and rerouted to the database engine queue**. Their real gaps are
database-shaped, not API-shaped: read-only with `write: false`, no CDC (DynamoDB Streams for
dynamodb), and — for dynamodb — `dynamic_schema: false` despite being a schemaless store where items
carry arbitrary attributes, so it likely needs the dynamic schema discovery foundation merged as
**#3892**. The proper test story is a container running DynamoDB Local, the harness the MySQL lane
just proved.

The committed red test **stays on the branch** (`cmd/connectorgen/dynamodb_api_surface_test.go` plus
its GSD phase artifacts, commit `bdb6efc0a`) — per captain, it is good work the database lane will
want. It is **not** carried into any PR from this sweep.

### Excluded — recorded skips

| Connector | Reason | Recorded |
| --- | --- | --- |
| hubspot | Dynamic-schema; deliberately excluded by brief, waits on PR #3892 | brief |
| airtable | Dynamic-schema; deliberately excluded by brief, waits on PR #3892 | brief |
| mailchimp | Own live lane — complete and pushing at time of handover | correction 2 |
| aws-cloudtrail | Own live lane (`fm/cli-aws-cloudtrail-parity-resume-r1`) | correction 2 |
| shopify | Own live/preserved lane (`fm/cli-shopify-parity-wave01-r1`) | correction 2 |
| **dynamodb** | `integration_type: database` + Tier-3 native Go pkg → **rerouted to database engine queue** | correction 3 |
| **postgres** | `integration_type: database` + native Go pkg → **rerouted to database engine queue** | correction 3 |

### Inherited lanes (resume, do not restart)

All five confirmed `paused: folded into cli-top50-fixed-schema-sweep-r1` before being taken, so no
worker was raced. Resume points read from
`/Users/karthiksivadas/karthik-agent-workspace/state/<task-id>.status`.

> ⚠️ **2026-08-07 — the recorded handover heads were NOT on any remote branch.** Verified for
> lever-hiring: `bcacf3490` existed only in the worktree object store, reachable from **zero** remote
> branches, while `origin/fm/cli-lever-hiring-parity-wave05-r1` still pointed at the **pre-rebase**
> `2df0342da` (forked Aug 1). One pane teardown would have destroyed it. **Now preserved** at
> `origin/fm/cli-lever-hiring-parity-wave05-r1-preserve-20260807`.
>
> **Check the other four the same way before assuming they are safe** — the recorded head is not
> proof it is pushed:
>
> ```
> git cat-file -t <head>                 # exists locally?
> git branch -r --contains <head>        # empty  => UNPUSHED, preserve it now
> git push origin <head>:refs/heads/fm/cli-<name>-parity-<wave>-r1-preserve-20260807
> ```

| Connector | Branch | Head at handover | Rebased onto | Remote state |
| --- | --- | --- | --- | --- |
| lever-hiring | `fm/cli-lever-hiring-parity-wave05-r1` | `bcacf3490` | `8a2971a0c` | **preserved 2026-08-07** as `…-preserve-20260807`; origin branch itself is stale at `2df0342da` |
| greenhouse | `fm/cli-greenhouse-parity-wave05-r1` | `da98de1de` (3 commits) | `8a2971a0c` | **preserved 2026-08-07** |
| monday | `fm/cli-monday-parity-wave02-r1` | `0ae91aa44` (3 commits) | `8a2971a0c` | **preserved 2026-08-07** |
| marketo | `fm/cli-marketo-parity-wave05-r1` | `4f45b7a11` (6 commits) | `8a2971a0c` | **preserved 2026-08-07** |
| intercom | `fm/cli-intercom-parity-wave01-r1` | `6f0de2c24` (6 commits) | `8a2971a0c` | **preserved 2026-08-07** |

**All five were unpushed, and all five are now preserved.** Checked on 2026-08-07: every one of the
recorded handover heads existed only in the worktree object store with `git branch -r --contains`
returning **zero** remote branches — **22 commits of rebased, review-fixed work across five
connectors**, one pane teardown from permanent loss. Each is now pushed to
`fm/cli-<name>-parity-<wave>-r1-preserve-20260807`. Resume from the **preserve** branch, never from
the same-named origin branch, which is stale pre-rebase in at least lever-hiring's case.

**lever-hiring resume detail.** The preserved branch carries 4 commits on top of `8a2971a0c`:
`f0b754fce` expand lever hiring parity → `f9283c2b9` harden Lever parity metadata →
`ad0252cd8` fix Lever CLI object-array flags → `bcacf3490` sync lever-hiring runtime endpoint
ledger. The last two are **no-mistakes review fixes**, so it has already been through a review
cycle. It needs rebasing from `8a2971a0c` onto current `origin/main` (`d71c6206b`, post-notion), its
count re-derived from the provider artifact, and then the standard bar. Its GSD phase artifacts
(PLAN/RUN-STATE/SUMMARY/TDD-LEDGER/VERIFICATION) and a 1,220-line
`sources/lever-official-http-ops.json` are already on the branch.

## Ledger targets — FLOOR ONLY, NOT THE TARGET (captain correction, 2026-08-07)

Source: `../cli-provider-artifact-sweep-r1/ledger.json` (522-record completed sweep).
Counting policy: one documented REST/RPC action or public top-level GraphQL query/mutation field;
webhook events excluded; GET/HEAD are reads; POST classified semantically.

**The ledger counts for this cohort are stale by design and must not be used as targets.**
Verified against the ledger: **28 of my 32 records carry `counting_note: "... Carried forward from
cli-top50-surface-audit-r1"`** — copied from an older audit, never freshly derived. The only 4
without that marker are the 4 recorded `unknown` (linear, postgres, quickbooks, workday-rest).

### Mandatory method

1. For **every** connector, **re-derive** the operation count from the provider artifact URL
   recorded in the ledger (or a current artifact if that URL is dead).
2. Treat the ledger number as a **floor and a cross-check**, never the target.
3. If the fresh count differs from the ledger, **that is a finding** — record both numbers and the
   reason in the Landed table below and in the PR body.
4. Never invent, estimate, or infer a count.

Hitting a stale target and declaring parity is the exact defect this programme exists to eliminate.

**Proven case — notion:** ledger says 50; developers.notion.com today lists ~57 action-shaped
endpoints (~63 counting query/trash/upload pages) — roughly 25% understated, because Notion has
since shipped an entire views API (create-view, update-a-view, delete-view, create-view-query,
get-view-query-results, list-views) plus meeting notes, file uploads, and token introspection.
Notion documents webhooks but exposes **no webhook management endpoints**, so there is nothing to
count there; webhook events remain excluded per the counting policy.

Ledger floor total across the counted connectors: **7,861** documented operations across the 27 that
carry a number (was 7,918 across 28; **−57 for dynamodb**, removed by correction 3). The true sweep
total will be higher and is established connector-by-connector as each is re-derived.

## Landing order — with re-derived counts (MASTER-PLAN.json, 2026-08-07)

All 30 derived from live provider artifacts in one parallel pass. **Zero undetermined.**
Ledger total across the 27 that carried a number: **7,861**. Re-derived across those same 27:
**7,807** (net −54). Plus **1,656** newly enumerable from the three the ledger recorded `unknown`
(linear 538, workday-rest 920, quickbooks 198). **Sweep total: 9,463 documented operations.**

**18 of 27 reconcile exactly; 9 drift.** Full provenance, hazards, webhook inventories and issue
links per connector live in `MASTER-PLAN.json`.

| # | Connector | Ledger | **Re-derived** | Delta | Status |
| --- | --- | ---: | ---: | ---: | --- |
| 1 | notion | 50 | **51** | +1 | **MERGED #3894** |
| 2 | lever-hiring | 67 | **106** | +39 | planned |
| 3 | gong | 69 | **69** | +0 | **PR #3895** |
| 4 | gmail | 79 | **79** | +0 | red committed |
| 5 | mixpanel | 105 | **104** | -1 | planned |
| 6 | gorgias | 114 | **114** | +0 | planned |
| 7 | greenhouse | 138 | **138** | +0 | planned |
| 8 | chatwoot | 146 | **148** | +2 | planned |
| 9 | help-scout | 146 | **145** | -1 | planned |
| 10 | google-ads | 163 | **163** | +0 | planned |
| 11 | segment | 197 | **197** | +0 | planned |
| 12 | twilio | 197 | **197** | +0 | planned |
| 13 | xero | 235 | **235** | +0 | planned |
| 14 | asana | 249 | **249** | +0 | planned |
| 15 | front | 255 | **244** | -11 | planned |
| 16 | trello | 261 | **261** | +0 | planned |
| 17 | monday | 280 | **292** | +12 | planned |
| 18 | bamboo-hr | 311 | **311** | +0 | planned |
| 19 | marketo | 322 | **320** | -2 | **BLOCKED — 3-way count disagreement** |
| 20 | intercom | 324 | **231** | -93 | planned |
| 21 | bitbucket | 331 | **331** | +0 | planned |
| 22 | square | 334 | **334** | +0 | planned |
| 23 | chargebee | 438 | **438** | +0 | planned |
| 24 | stripe | 589 | **589** | +0 | planned |
| 25 | jira | 616 | **616** | +0 | planned |
| 26 | zendesk-support | 625 | **625** | +0 | planned |
| 27 | github | 1220 | **1220** | +0 | planned |
| 28 | linear | unknown | **538** | — | planned |
| 29 | quickbooks | unknown | **198** | — | planned |
| 30 | workday-rest | unknown | **920** | — | planned |

## ⇒ RESUME HERE — ⇄ LARGEST-FIRST work order, one commit+push per connector

Branch: **`fm/cli-top50-sweep-continue-r1`** (cut from `fm/cli-top50-sweep-consolidated` @
`bedb9b0a3`). Order is **DESCENDING by re-derived count — eat the frog** (captain, 2026-08-07,
reversing the original smallest-first order). **github 1220 and workday-rest 920 go first.**
**Commit AND PUSH after every single connector** so a context loss costs at most one, and **within**
an oversized connector commit and push each slice rather than holding one giant uncommitted change.
Update the tick-box below in the same step.

### 🛑 mixpanel (#2) — STOPPED AND RECORDED, not forced. Bigger than the plan said.

Red is committed and honest (`14bb74556` cherry-picked; re-run here: `endpoints = 47, want 104`).
**Authoring deliberately not attempted** — two structural surprises make it a restructure, not an
extension, and the standing rule is to stop and record rather than force it into the batch shape.

**1. `api_surface.json` collapses many operations into single comma-joined rows.**
23 of its 47 rows carry comma-separated paths — e.g. one row is
`GET /insights, /funnels, /retention, /retention/addiction, /segmentation, /segmentation/numeric,
/segmentation/sum, /segmentation/average, /events, /events/properties` (**ten** operations in one
row). Splitting every row on commas yields 108 paths against a derived 104. So mixpanel is not
"47 rows needing 57 more" — it is **47 collapsed rows needing a full restructure into 104 discrete
rows**, and the 108-vs-104 gap must be reconciled path by path first. This is the same defect class
as DynamoDB's `X-Amz-Target` collapse and chatwoot's trailing slash, but already baked into the
shipped bundle.

**2. Five different base URLs across 13 spec files.** `/api/app` (annotations, experiments,
feature-flags-management, gdpr, lexicon-schemas, service-accounts, warehouse-connectors),
`/api/2.0` (data-pipelines, export), `/api/query` (query), bare `{regionAndDomain}.com`
(feature-flags), `{region}.mixpanel.com` (identity, ingestion). The declared base is
`https://mixpanel.com/api/2.0`. Existing streams already cope by using absolute URLs.

**Also absent:** `writes.json`, `cli_surface.json`, `operations.json` — so, as with gmail, none of
mixpanel's operations are reachable as commands today.

**Absolute-path probe: INCONCLUSIVE, do not assume either way.** A probe operation with an
implemented absolute-path direct read got as far as two *unrelated* schema findings (`approval`
required; an implemented direct read must reference exactly one `api_surface` endpoint) before the
question could be answered, and the probe was then removed to leave mixpanel clean. **The absolute
path itself was never accepted or rejected on its merits.** Whoever resumes should settle it with a
minimal probe wired to a real single-path `api_surface` row — the answer decides both mixpanel's
shape and whether chatwoot's 47 blocks stand.

### ✅ SETTLED: chatwoot's 47 blocked rows are CORRECT. My suspicion was wrong.

Probed properly on a real chatwoot blocked row (`GET /api/v2/accounts/{account_id}/reports`) wired
as an **implemented** absolute-templated-path direct read, then reverted. **The 47 blocks stand.**
Recording the full mechanism because the isolated code reading that made me doubt them was
genuinely misleading, and the same trap will recur.

**Why I doubted it:** `connsdk.(*Requester).resolveURL` really does bypass the base for any
`http(s)://` path, so *in isolation* an absolute path looks like a clean per-operation base override.
That part of my reading was correct.

**Why the blocks are nonetheless right — it is the LEDGER gate, not the URL builder:**
`engine.deriveOperationDirectReadEndpointLedger` only emits a ledger entry when
`hasOperationDirectReadSurfaceEndpoint(surface, method, operation.REST.Path)` holds — i.e. the
operation's `rest.path` must **exactly equal a row in `api_surface.json`**. That closes both doors:

| Attempt | Outcome |
| --- | --- |
| Operation path = documented relative `/api/v2/...` | Ledger entry emitted, **but** at request time `normalizeDirectReadPathForBaseURL` cannot strip a non-matching prefix, so `resolveURL` joins it to the `/api/v1/accounts/{id}` base → `…/api/v1/accounts/{id}/api/v2/accounts/{id}/reports`. **Broken URL.** |
| Operation path = absolute `{{ config.base_url }}/api/v2/...` | Correct URL at request time, **but** it no longer equals any `api_surface` row, so **no ledger entry is emitted** and runtime preflight refuses it: `runtime operation endpoint ledger does not contain GET {{ config.base_url }}/api/v2/… kind "rest_read"`. Observed verbatim. |

Making it work would require putting the absolute templated string into `api_surface.json` itself,
which would corrupt that file's role as the record of **documented provider endpoints** and break the
surface test. So the honest disposition really is blocked-with-named-dependency, and chatwoot's
sub-agent reached the right answer.

**Sharpen the wording, not the disposition.** The blocked reasons say "per-operation base URL
override is not supported". More precisely: *the operation endpoint ledger requires
`operation.rest.path` to match an `api_surface` row verbatim, and an out-of-base path cannot satisfy
both that and correct URL construction at once.* Worth tightening if chatwoot is ever revisited, but
**not worth reopening 47 rows** — the disposition is right.

**Consequence for mixpanel:** its five base URLs hit exactly this. Its existing *streams* cope via
absolute paths because **streams do not go through the operation endpoint ledger** — only direct
reads, direct writes and binary downloads do. So mixpanel can keep multi-base *streams* but faces the
same wall for out-of-base *direct reads*.

**Lesson:** `resolveURL` alone does not tell you whether a path is usable. The ledger gate is the
binding constraint for anything marked `implemented`.

### ⚠️ (superseded — kept for provenance) REVISIT chatwoot's 47 blocked rows — the stated reason looks WRONG

chatwoot blocked 47 of its 148 operations on the grounds that "the engine binds exactly one
`HTTPBase` per bundle and **rejects an absolute-path override for direct reads**", so anything
outside `/api/v1/accounts/{account_id}` was deemed unreachable.

**That premise does not hold at the HTTP layer.** `connsdk.(*Requester).resolveURL`
(`internal/connectors/connsdk/http.go:371`) reads:

```go
if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
    base := strings.TrimRight(r.BaseURL, "/")
    raw = base + "/" + strings.TrimLeft(path, "/")
}
```

An absolute `http(s)://` path **bypasses the base entirely and is used as-is**. Corroborating
evidence: **mixpanel's existing, shipped streams already do this** — `saved_funnels`,
`activity_stream`, `top_events` and the three `project_annotation*` streams all carry absolute
`https://mixpanel.com/api/query/...` and `.../api/app/...` paths against a declared base of
`https://mixpanel.com/api/2.0`. And `youtube-analytics` declares an absolute-path `rest_read`
(though at `availability: planned`, so it is not itself runtime-proven).

**Not yet confirmed:** whether `connectorgen validate` or the runtime preflight *separately* reject
an absolute path on an **implemented** direct read. That is the only remaining question, and it is
being answered empirically while authoring mixpanel, which needs absolute paths across **five**
distinct base URLs (`/api/app`, `/api/2.0`, `/api/query`, bare `{regionAndDomain}.com`,
`{region}.mixpanel.com`).

**Action:** if mixpanel's implemented absolute-path direct reads validate and pass preflight, then
chatwoot's 47 blocks are wrong and must be revisited — a large correction (47 operations moving from
blocked to executable). Do not let this finding get lost; chatwoot is already folded into the
consolidated branch.

### ⛔ Do NOT hand-author paging flags (standing constraint, 2026-08-07)

**Never author `page`, `per_page`, `limit`, `offset`, `cursor`, `page_token`, `pageToken`,
`maxResults`, `page_size` or `starting_after`/`ending_before` flags into any `cli_surface.json`.**
A parallel foundation lane is making these derive automatically from each connector's declared
pagination spec; hand-authored ones will collide with it.

If a paging flag ever *seems* required for a specific command, **record it here and escalate — do
not author it.**

Applied retroactively: chatwoot carried 6 (`audit-logs list`, `automation-rules list`,
`contacts filter`, `contacts search`, `conversations filter`, `reporting-events list`, all `--page`)
from before the constraint. Removed and **verified not required rather than assumed**:
`connectorgen validate` 0 findings without them, all 6 commands still reachable, `TestChatwoot`
still passes. gmail was checked and carried none. **No case has yet arisen where a per-command
paging flag was genuinely required.**

Cheap check before every commit:

```
python3 -c "import json,re,sys;c=json.load(open(sys.argv[1]));P=re.compile(r'^(page|per_page|per-page|limit|offset|cursor|page_token|pageToken|maxResults|page_size|pageSize|starting_after|ending_before)$',re.I);print([(x['path'],f['name']) for x in c['commands'] for f in (x.get('flags') or []) if P.match(f['name'])])" internal/connectors/defs/<name>/cli_surface.json
```

Per connector, in order: write its red test against the **real** bundle → **run it and watch it
fail** → record the verbatim failure in `RUN-STATE.json` + `TDD-LEDGER.md` → author to green →
`connectorgen validate` + whole `cmd/connectorgen` package + runtime preflight → **one commit** →
**push** → update this list. Shared artifacts (endpoint ledger, website catalogs, golden transcripts)
are regenerated **once at the very end**, not per connector.

### Delivered (landed under the previous smallest-first order — not redone)

| Connector | Ops | Done |
| --- | ---: | --- |
| gmail | 79 | ✅ `d1bacbfe1` — 79 rows, 66 covered + 13 blocked, 66 cmds all reachable |
| lever-hiring | 106 | ✅ `0829543e6` — 117 preserved rows → **106**; 60 covered + 46 blocked; 60 cmds all reachable |
| greenhouse | 138 | ✅ — 138 rows, 127 covered + 11 blocked, 127 cmds all reachable |
| help-scout | **144** | ✅ — 144 rows (ledger 146, derivation 145; both deltas dedup), 139 covered + 5 blocked, 139 cmds all reachable |

### ⇄ REMAINING — LARGEST FIRST (captain, 2026-08-07). Work top to bottom.

| # | Connector | Ops | Done |
| ---: | --- | ---: | :--: |
| 1 | **github** | **1220** | ☐ — 🐸 **the frog.** Its 270 webhook events live under `x-webhooks`, NOT the standard block (finding 13): an inventory checking only `webhooks` records zero for the largest connector. Slice it (see the work order above). |
| 2 | **workday-rest** | **920** | ☐ — ledger recorded `unknown`; 920 is freshly derived, so it has no cross-check. Derive it twice by different routes before trusting it. Slice it. |
| 3 | zendesk-support | 625 | ☐ |
| 4 | jira | 616 | ☐ |
| 5 | stripe | 589 | ☐ |
| 6 | linear | 538 | ☐ — **538 vs 539**, one Mutation field (finding 11); resolve at red-test time by diffing the field-name sets. |
| 7 | chargebee | 438 | ☐ — `connectorgen batch` blocker: OAS 3.1 + top-level `webhooks` (227 events). Hand-author around the gate; do not relax it (finding 12). |
| 8 | marketo | **TBD** | ☐ — **unblocked by captain ruling `DECISION-marketo-scope.md`.** Full surface incl. Adobe v2 Assets, subject to BOTH conditions. Count must be **re-derived, not picked**; explain the delta vs 320 / 327 / 367 and the ledger's 322. Slotted on its upper candidate; re-slot if the real count lands outside the range. |
| 9 | square | 334 | ☐ |
| 10 | bitbucket | 331 | ☐ |
| 11 | bamboo-hr | 311 | ☐ — `connectorgen batch` blocker: OAS 3.1 + top-level `webhooks` (6 events). |
| 12 | monday | 292 | ☐ — resume from `fm/cli-monday-parity-wave02-r1-preserve-20260807`, do not restart. |
| 13 | trello | 261 | ☐ |
| 14 | asana | 249 | ☐ |
| 15 | front | 244 | ☐ — ledger 255, derived 244 (−11); expect a dedup story, and check for query-string variants first (finding 22). |
| 16 | xero | 235 | ☐ |
| 17 | intercom | 231 | ☐ — ledger 324, derived **231** (−93), corroborated three ways. Resume from `fm/cli-intercom-parity-wave01-r1-preserve-20260807`. |
| 18 | quickbooks | 198 | ☐ — ledger recorded `unknown`; 198 is freshly derived and has no cross-check. |
| 19 | segment | 197 | ☐ |
| 20 | twilio | 197 | ☐ |
| 21 | google-ads | 163 | ☐ |
| — | ~~mixpanel~~ | 104 | 🛑 **DEFERRED to its own task.** Red test moved OFF the sweep branch (it would keep `cmd/connectorgen` red for a connector this PR does not deliver); preserved on `fm/cli-mixpanel-parity-sweep-r1`. Its WIP deletion of `schemas/annotations.json` is **WRONG** and must never be folded in. |

Already on the branch: **chatwoot** (148, folded, verified). Separate and open: **gorgias #3896**
(114). Merged: **notion #3894** (51), **gong #3895** (69).
**marketo is no longer excluded** — the captain ruled on 2026-08-07 that its surface includes Adobe's
v2 Asset API; see `DECISION-marketo-scope.md` and row 18a. Still deferred: **mixpanel** (own task).

## ⇒ MODEL CHANGE: ONE CONSOLIDATED SWEEP PR (captain, ordered twice — 2026-08-07)

**Supersedes the founding "one connector, one PR" constraint** and the interim one-agent-in-flight
steer. Confirmed explicitly in-conversation after this lane escalated the conflict rather than
choosing between two directives.

- Consolidated branch: **`fm/cli-top50-sweep-consolidated`**, cut from `origin/main` (`5d43d7c00`).
- **gorgias PR #3896 stays as-is** — open, nearly green, not discarded or folded.
- **chatwoot folded in**: its 5 pushed commits cherry-picked (red → feat → docs → website →
  planning). Verified intact after the fold: 148 rows, 101 covered / 47 blocked, `ledger_v 1`,
  0 duplicates, 101 CLI commands, red test green.
- **Retained, non-negotiable**: red test per connector that genuinely fails first; one commit per
  connector inside the single PR so a bad connector can be dropped without redoing the rest;
  shared artifacts regenerated **exactly once** at the end; this lane verifies counts itself.
- marketo stays excluded pending the captain's Adobe v2 Asset API scope call.

### Injection scare — CLOSED, not an attack

A sub-agent twice received instructions this lane never wrote, one appearing to confirm something as
this lane. Escalated and held. **Resolved: they were firstmate's own `fm-send` steers landing in the
sub-agent's view because the pane was focused there.** Not impersonation. Recorded because the
sub-agent's refusal to act on an unverified instruction that contradicted its brief was correct
behaviour and should stay the norm.

### mixpanel — the flagged deletion was WRONG, and the check caught it

The wave-1 WIP deleted `internal/connectors/defs/mixpanel/schemas/annotations.json` and its fixture,
removing the annotations stream with no justification. **Checked against the provider artifact:
`annotations` is still fully documented** — Mixpanel ships a dedicated `annotations.yaml` OpenAPI
file carrying `/annotations`, `/annotations/{annotationId}` and `/annotations/tags`.

**The deletion must NOT be folded in.** Had it shipped, mixpanel would have silently lost a
documented stream — exactly the defect class this programme exists to remove. mixpanel is re-done in
the generation pass with annotations retained; only its **red test** (`14bb74556`, genuinely
observed: `endpoints = 47, want 104`) is worth carrying forward.

### Technical constraint found when starting the mechanical pass — read before continuing

"Generate mechanically from the existing planning slice" is **not literally sufficient**. The slices
carry the **count, artifact URL, hazards, issue links and webhook inventory** — they do **not** carry
a per-operation inventory. Authoring `cli_surface.json`/`operations.json` needs the actual endpoint
list, so the generator must read each **provider artifact**, not just the slice.

Most artifacts are already fetched and cached under the sweep tooling (`tools/derive/raw/`), so this
is a cost already paid — but the generator has to be pointed at them. And note the residue that is
genuinely **not** mechanical: semantic read/write classification, stream-vs-direct-read shape, binary
detection, and blocked dispositions with named dependencies all needed real judgement on gorgias
(6 blocked rows) and chatwoot (47 blocked by a single-`HTTPBase` architectural limit). Expect a
mechanical pass to leave those for per-connector adjudication.

## Implementation wave 1 — halted by an ACCOUNT SPEND LIMIT, work preserved (2026-08-07)

**Both failures were an account monthly-spend limit, NOT code faults.** Per the captain: retry
serially, **at most ONE implementation agent in flight** until told otherwise. Do not re-fan out.

All partial work was committed and **pushed** the moment the agents died — applying the lesson from
the five inherited lanes that unpushed work is one teardown from gone.

| Connector | Branch | Red observed? | State |
| --- | --- | --- | --- |
| **gorgias** | `fm/cli-gorgias-parity-sweep-r1` | **YES** — `api_surface declares 11 rows, want 114`, confirmed against `main`'s real 11-row bundle | **LANDED — PR #3896.** 11 → 114. Rebased onto post-Mailchimp `main`; both conflicts were generated files, resolved by regeneration not hand-merge. All diffs inspected by this lane. |
| **mixpanel** | `fm/cli-mixpanel-parity-sweep-r1` | **YES** — `endpoints = 47, want 104` | **INCOMPLETE** — died mid-docs-update; catalogs/ledger never regenerated. **Plus an unexplained deletion (below).** |
| **chatwoot** | — | — | still in flight at time of writing |

Both red tests were **genuinely observed failing and committed before production edits** — the
constraint held under batching, which was the thing most at risk from parallelising.

### mixpanel — one item needing lane judgement before it can ship

Its WIP commit **deletes `internal/connectors/defs/mixpanel/schemas/annotations.json` and its
fixture**, removing the annotations stream. That deletion is **unexplained**. A parity pass should
not silently drop an existing stream: either justify it against the provider artifact (the endpoint
genuinely no longer documented) or **revert it**. Recorded rather than waved through.

### Resume instructions

1. Wait for chatwoot, then retry **gorgias** serially (it is closest to done — likely only needs
   lane diff inspection and a PR).
2. Then **mixpanel**, resolving the annotations deletion first.
3. Inspect every regenerated diff **by object, not by line** — gorgias touches
   `operation_endpoint_ledger.json`, `golden_transcripts.json` and both website catalogs.
4. Never restart the shared no-mistakes daemon.

## Captain rulings on the master-plan open decisions (2026-08-07)

| Decision | Ruling |
| --- | --- |
| **front** product scope | **244, Core API only.** The ledger's 255 is an un-deduplicated sum across two products; counting the same operation twice because it appears in two catalogues is the exact miscount discipline applied everywhere else in this sweep. The true union would be 254 (`PATCH /channels/{channel_id}` is in both specs), not 255 — itself proof 255 was never deduplicated. **Recorded in `MASTER-PLAN.json`; do not re-litigate.** |
| **linear** 538 vs 539 | Resolve at **red-test time**. A one-item delta is exactly what a failing test settles. Non-blocking. |
| **lever-hiring** 106 vs 107 | Resolve at **red-test time**. Non-blocking. |
| **marketo** 320 / 327 / 367 | **ESCALATED TO THE CAPTAIN. marketo is SKIPPED**; the sweep proceeds with the other 29 rather than blocking on one connector. Adopt none of the three until he rules. |

### Ledger defect worth recording on its own: lever-hiring 67 is wrong by about 40

Independently of the unresolved 106-vs-107 delta, **the ledger's lever-hiring figure of 67 is simply
wrong**, by roughly 40 operations, whichever of 106/107 turns out correct:

- this sweep's HTML derivation: **106**
- the preserved lever-hiring lane's own `api_surface.json`: **107** HTTP operations
  (its 117 rows include 10 webhook-*event* rows, which this sweep's policy excludes from the total)
- a 2020 Wayback snapshot of the same page, parsed with identical methodology: **129**

So 67 does not match the current surface, does not match an independent lane's derivation, and does
not match the page as it stood six years ago. It is not stale-by-drift — **no point in time supports
it.** This sits alongside intercom (324 → 231) as the second demonstrable ledger defect, and the two
err in opposite directions.

## Re-derived counts (findings)

### notion — re-derived 2026-08-07

**Ledger said 50 (carried forward). Re-derived: 51 documented operations. Finding: ledger
understated by 1.**

Method — the ledger recorded `artifact_kind: html_reference` at
`https://developers.notion.com/reference/intro`, but **Notion publishes a real machine-readable
OpenAPI 3.1.0 artifact** at `https://developers.notion.com/openapi.json` (876 KB, `info.version`
1.0.0, fetched 2026-08-07). That is a better artifact than the one recorded and is what I counted
from. Ledger record should be upgraded from `html_reference` to `openapi`.

| Source | Count |
| --- | --- |
| Official OpenAPI `paths` — HTTP operations (20 GET, 17 POST, 8 PATCH, 4 DELETE over 34 paths) | 49 |
| Legacy endpoints documented only in the nav's explicit **"Databases (deprecated)"** group — `GET /v1/databases`, `POST /v1/databases/{database_id}/query` (absent from current OAS) | +2 |
| **Total documented operations** | **51** |
| OpenAPI `webhooks` block — webhook *events* | 31 (**excluded per counting policy**) |

Notion exposes **no webhook management endpoints** (no `/v1/webhooks` path in the OAS), so there is
nothing countable there — confirms the captain's note.

**Correction to the briefed figure.** The coordination message stated Notion "today lists 57
action-shaped endpoints and about 63 counting query/trash/upload pages — roughly 25 percent
understated". The direction of that claim is right about *content* — Notion has indeed shipped the
whole views API (`list-views`, `create-view`, `retrieve-a-view`, `update-a-view`, `delete-view`,
`create-view-query`, `get-view-query-results`), meeting notes, file uploads and token introspection
since the old audit, and all of those are present in the 51. But the *magnitude* does not hold up
against the artifact: **55 documentation pages carry an `openapi` nav entry, and those collapse to
51 unique method+path actions**, because four operations are each documented on two pages —

- `POST /v1/databases` → `create-database` + `create-a-database`
- `GET /v1/databases/{database_id}` → `retrieve-database` + `retrieve-a-database`
- `PATCH /v1/databases/{database_id}` → `update-database` + `update-a-database`
- `POST /v1/oauth/token` → `create-a-token` + `refresh-a-token`

Counting *pages* yields ~55–57; the sweep's counting policy counts **one documented REST/RPC
action**, which yields 51. Pages such as `trash-page`, `update-data-source-properties`,
`post-database-query-filter`, and `sort-data-source-entries` carry **no** `openapi` nav entry — they
are usage guides for endpoints already counted (trashing a page is `PATCH /v1/pages/{page_id}` with
`in_trash`, not a distinct endpoint), so they are correctly not counted.

So the real gap for notion is not the count — it is **coverage**: the bundle in `main` declares only
**6** api_surface endpoints (3 read streams: databases, pages, users; 3 excluded) against 51
documented operations, `capabilities.write: false`, and no `cli_surface.json`, `operations.json`, or
`writes.json` at all. Bringing it to parity is a full build, not a top-up.

Evidence retained: `<scratchpad>/notion/openapi.json`, `ops.json`, `nav-ops.json`, plus all 91
fetched `/reference/*` pages.

### dynamodb — re-derived 2026-08-07 — ⚠️ PRESERVED FOR THE DATABASE ENGINE QUEUE

> **Connector REMOVED from this sweep by correction 3 (database-typed). These findings are
> verified, complete, and carry over — the database lane MUST NOT re-derive them.** Retained here
> in full at the captain's explicit instruction. Companion red test: commit `bdb6efc0a`,
> `cmd/connectorgen/dynamodb_api_surface_test.go` + `.planning/phases/dynamodb-parity-sweep-r1/`.
>
> Carry-over summary: **58 operations (27 read / 31 write)**; ledger's 57 was **correct when taken**,
> the delta being newly-added `SearchVectors`; the `X-Amz-Target` path-collapse hazard; and
> **Streams (4) + DAX (21) are separate services**, never to be folded in.

**Ledger said 57 (26 read / 31 write, carried forward). Re-derived: 58 (27 read / 31 write).
Finding: ledger understated by exactly 1, and the missing operation is named and dated.**

Artifact: AWS's own service model,
`https://raw.githubusercontent.com/boto/botocore/develop/botocore/data/dynamodb/2012-08-10/service-2.json`
(524,841 bytes; `serviceId: DynamoDB`, `targetPrefix: DynamoDB_20120810`, `apiVersion: 2012-08-10`).
Byte-identical at stable tag `1.43.66`. Cross-checked against the ledger's own cited HTML page,
`API_Operations.html` — its 58 core-slug operations are **set-equal** to the service model. No
operation is marked deprecated.

**The +1 is `SearchVectors`** — vector similarity search over a DynamoDB vector index (`POST /`,
with 18 supporting `*Vector*` shapes). Proven by version probe, not inferred:

| botocore | operations | `SearchVectors` |
| --- | --- | --- |
| 1.35.0 | 57 | absent |
| 1.38.0 | 57 | absent |
| 1.40.0 | 57 | absent |
| develop / 1.43.66 | **58** | **present** |

So the original audit's 57 was **correct when taken** and has since gone stale by exactly one
operation. `SearchVectors` classifies as a read, which moves 26→27 reads and leaves writes at 31 —
and my independent semantic classification produced **exactly 31 writes**, reconciling with the
ledger's write count with no adjustment. That is strong mutual corroboration of both numbers.

**Classification note.** Three PartiQL operations — `ExecuteStatement`, `BatchExecuteStatement`,
`ExecuteTransaction` — are documented by AWS as reading *or* writing depending on the statement
supplied at call time. Counted as writes, because an operation that can mutate must be gated as a
mutation. `ExportTableToPointInTime` counts as a write: it creates an export job.

**Separate services, deliberately NOT merged into the 58:**
- **DynamoDB Streams** — 4 operations (`DescribeStream`, `GetRecords`, `GetShardIterator`,
  `ListStreams`), own botocore directory and `targetPrefix`, own `API_streams_*` doc slug.
- **DAX (DynamoDB Accelerator)** — 21 operations under the `API_dax_*` doc slug.

**Miscount hazard worth carrying to other connectors:** a naive scrape of every `API_*.html` link on
that single AWS page yields **84**, not 58, by silently folding in DAX and Streams. Counting must key
on the slug prefix. This is the same class of error as counting Notion's doc pages instead of its
unique actions.

No webhook or subscription-management operations exist in either the 58 or the Streams 4 (keyword
search across names and documentation text).

### gong — re-derived 2026-08-07

**Ledger said 69 (29 read / 40 write, carried forward). Re-derived: 69. Finding: the total
reconciles EXACTLY — zero drift. The read/write split is wrong, the total is not.**

Artifact: Gong's own OpenAPI document at the exact URL the ledger records —
`https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=` — HTTP 200,
453,605 bytes, `openapi: 3.0.1`, `info.title: "Gong API"`, `info.version: V2`, fetched 2026-08-07.
**Publicly reachable with no authentication**, despite the `gong.app.gong.io/ajax/...` URL shape
suggesting an internal endpoint.

**69 operations across 59 paths: GET 29, POST 28, PUT 8, DELETE 3, PATCH 1.**

**Split correction.** The ledger's `29 read / 40 write` is a naive HTTP-method split. The counting
policy is "GET/HEAD are reads; **POST classified semantically**", and Gong documents three official
POST read-query endpoints — `/v2/calls/extensive`, `/v2/calls/transcript`, `/v2/stats/interaction`.
The honest split is **32 read / 37 write**. The total of 69 is unaffected. Worth carrying: a
method-split read/write ratio in the ledger is *not* reliable even when the total is.

**Baseline measured by running the binary, not by reading files:** 67 declared endpoints, all 67
commands reachable (`--help` exit 0 on every one), 0 blank dispositions, 43 `implemented` /
24 `partial`. gong was **not** a rebuild like notion — it was already substantially complete.

**The gap was exactly 2 operations, both in Targets, and zero stale rows** (the bundle was a clean
subset of the spec — nothing to retire):

| Operation | opId | Shape | Landed as |
| --- | --- | --- | --- |
| `GET /v2/targets` | `listTargetDefinitions` | uncursored `{requestId, targets}` | **direct read** `pm gong targets list` |
| `POST /v2/targets/{targetId}/assignments` | `uploadAssignments` | `multipart/form-data`, one `file` part, `format: binary` | **binary write** `pm gong targets upload-assignments` |

The GET is a **direct read, not a stream** — declaring it a stream would advertise pagination the
provider does not offer. The POST is a genuine **binary** operation routed through the **existing**
`engine.MultipartSpec` capability, copying the shape gong already ships and runs in
`upload_call_media` and `upload_crm_entities`.

**`partial` is honest here, not a defect.** The 24 `partial` commands are reverse-ETL writes whose
records contain complex object/array fields that flat CLI flags cannot express; the typed
reverse-ETL path is required. Named engine limitation, not a false `implemented` claim. Recorded
against **#3001**, not closed.

Deprecated: exactly 1 (`POST /v2/flows/prospects/assign/cool-off-override`) — documented, so counted,
and already present in the bundle.

### gmail — re-derived 2026-08-07, NOT yet built (next connector; research is done, do not redo it)

**Ledger said 79 (30 read / 49 write, carried forward). Re-derived: 79. Reconciles exactly.**

Artifact: Google's official Discovery document `https://gmail.googleapis.com/$discovery/rest?version=v1`
— HTTP 200, 217,687 bytes, `kind: discovery#restDescription`, `id: gmail:v1`,
**`revision: 20260803`** (3 Aug 2026, current), fetched 2026-08-07. Publicly reachable, no auth.

**79 methods: GET 30, POST 28, DELETE 10, PUT 8, PATCH 3.** Counted by walking
`resources.*.methods.*` **recursively** — Gmail nests resources (`users.messages.attachments`,
`users.settings.sendAs.smimeInfo`), so a non-recursive walk undercounts. There are **0** top-level
`methods` outside `resources`. Ledger's `30 read / 49 write` also matches on a GET-vs-rest split.

**Branch check: `fm/cli-gmail-parity-wave03-r1` is stale** (forked 31 Jul, `main` 52 commits ahead).
gmail is this lane's to build.

**Baseline — this is a much bigger job than gong, closer to notion:**

| | State |
| --- | --- |
| `api_surface.json` | **79 rows already present**, method split matches the artifact exactly |
| Dispositions | 10 `covered_by.stream` · 35 `covered_by.write` · **34 `excluded`** · 0 blank |
| `operation_ledger_version` | **unset** (`null`) — the v2 provenance ledger is missing |
| `cli_surface.json` | **ABSENT** |
| `operations.json` | **ABSENT** |
| `streams.json` / `writes.json` | present: 10 streams, 35 actions |

**So the gap is not the count — it is the 34 `excluded` rows plus a missing command surface.**
With no `cli_surface.json`, gmail cannot meet "individually reachable as `pm gmail <command>`" for
the direct-read scope at all.

**The 34 exclusions are reasoned and source-cited, not blank** — they fall into clear families, and
each family needs its own disposition decision rather than blanket promotion:

| Family | Count | Current reason |
| --- | ---: | --- |
| Single-resource detail GETs called redundant with a list stream | ~10 | "identical resource shape the list stream already paginates" |
| Settings singleton get-after-write reads (`auto_forwarding`, `vacation`, `language`, `imap`, `pop`) | 5 | "reads back the singleton this bundle already writes" |
| CSE (Client-Side Encryption) Workspace add-on surface | ~9 | elevated scope + Enterprise Plus/Education Plus admin gating |
| S/MIME elevated-scope certificate surface | ~5 | requires the narrow `https://mail.google.com/` full-mailbox scope |
| Bulk `batchModify` / `batchDelete` variants | 2 | engine write dialect is one-request-per-record |
| `watch()` / `stop()` Pub/Sub push subscription | 2 | control-plane side effect |
| `attachments.get` raw base64url bytes | 1 | not a syncable structured record |

**ADJUDICATION IS DONE** and committed as `.planning/phases/gmail-parity-sweep-r1/PLAN.md`
(commit `ee7ea6cd6`). **Do not re-adjudicate; author from it.** Summary:

| Family | Rows | Disposition |
| --- | ---: | --- |
| Single-resource detail GETs | 8 | **PROMOTE → direct read** |
| Settings singleton GETs | 5 | **PROMOTE → direct read** |
| S/MIME certificate surface | 5 | **PROMOTE — stated block is FALSE** |
| `watch` / `stop` | 2 | **PROMOTE — webhook management, in scope** |
| CSE (Client-Side Encryption) | 11 | blocked-with-named-dependency (Workspace add-on) |
| Bulk `batchDelete`/`batchModify` | 2 | blocked-with-named-dependency (**#514**) |
| `attachments.get` | 1 | binary download — capability check required first |
| **Legacy `excluded` remaining** | **0** | |

**Three findings from the adjudication, two of which contradict standing captain rulings:**

1. **`watch`/`stop` excluded as `non_data_endpoint` contradicts the webhook ruling (F2 point 5).**
   They register and cancel a Pub/Sub push **subscription** — webhook *management*, which stays
   fully in scope; only webhook *events* are deferred.
2. **`cse/keypairs/{id}:obliterate` excluded as `destructive_admin` contradicts the captain's
   2026-07-30 destructive-operations policy** (recorded verbatim on the gong issues): destructive,
   admin, DELETE and file-upload operations are all in scope when modeled as typed operations with
   schema-bounded inputs, redaction, and typed destructive confirmation. **Risk is not grounds for
   exclusion.** It is blocked on the CSE entitlement, and must say that instead.
3. **The 5 S/MIME exclusions are factually false against the bundle's own spec.** Their reason
   claims S/MIME needs "the narrow `https://mail.google.com/` full-mailbox scope" — but gmail's
   `spec.json` **already declares exactly that scope**, and documents why ("mutating write actions
   … that gmail.readonly cannot authorize"). The stated dependency does not exist.

CSE is genuinely blocked and is **not** a scope problem: it is a Workspace Enterprise Plus /
Education Plus **add-on** gated on organization-level admin enablement, which no OAuth scope the
connector declares can unlock. That distinction (entitlement vs scope) is what separates family 5
from family 4.

**RED is committed** — `fd1c07d01`, `cmd/connectorgen/gmail_api_surface_test.go` plus
`.planning/phases/gmail-parity-sweep-r1/{PLAN,TDD-LEDGER,RUN-STATE}`. F5 check done first: gmail had
**no** pre-existing surface test, so the file is strictly additive. Observed failure:

```
gmail_api_surface_test.go:80:  operation_ledger_version = 0, want 1
gmail_api_surface_test.go:139: 34 legacy excluded row(s) remain, want 0
gmail_api_surface_test.go:146: covered(45)+blocked(0) = 45, want 79
```

**The `endpoints = 79` and `totalByMethod` assertions already PASS.** That is the whole point:
gmail's *count* is correct and its *dispositions* are not. Beyond counts the test enforces exactly
one disposition per row, and requires every blocked row to carry a reason **and** a source citation
**and** a named dependency — so `blocked` can never degrade into a shrug.

### Remaining work for gmail (authoring — this is the bulk, and it is delegable)

1. Author **`cli_surface.json` from scratch** (absent today) — the largest single item; gong's
   equivalent is ~87 KB for 67 commands.
2. Author **`operations.json`** (absent today) for the typed/sensitive rows.
3. Re-disposition all **34** `excluded` rows per the table above; set `operation_ledger_version: 1`.
4. **Check whether a bounded engine binary *download* capability exists** before deciding
   `attachments.get`. `engine.Base64UploadSpec` covers uploads; a download equivalent must be
   confirmed, not assumed. If absent → `blocked-with-named-dependency` naming it.
5. Then: `surface-sync`, docs (revert the F4 repo-wide churn), website catalog by-object check,
   full `cmd/connectorgen` + `internal/cli` runs, `certify-timing`, PR.

## Sweep-wide findings (read before starting any connector)

### F1 — ledger staleness is concentrated in `html_reference` connectors, not uniform

The captain's correction (re-derive everything, ledger is a floor) stands and is being followed. But
measured evidence refines *where* the drift actually is. Freshly counted from the live artifacts on
2026-08-07:

| Connector | Artifact | Ledger | Re-derived | Drift |
| --- | --- | --- | --- | --- |
| stripe | OAS 3.0.0 | 589 | 589 | 0 |
| square | OAS 3.0.0 | 334 | 334 | 0 |
| twilio | OAS 3.0.1 | 197 | 197 | 0 |
| trello | swagger/OAS 3.0.0 | 261 | 261 | 0 |
| bitbucket | swagger/OAS 3.0.0 | 331 | 331 | 0 |
| chargebee | OAS 3.1.0 | 438 | 438 | 0 |
| gong | OAS 3.0.1 | 69 | 69 | 0 |
| notion | **html_reference** → real OAS 3.1.0 | 50 | **51** | **+1** |

**Seven of eight** OpenAPI/swagger-backed counts reconcile **exactly** (gong added 2026-08-07, also
zero drift). The one that drifted is the one the ledger had recorded as `html_reference`.

**But gong sharpens the rule: a matching total does not mean a matching record.** gong's total
reconciled at 69 while its recorded `29 read / 40 write` split was **wrong** — the ledger split on
HTTP method, and the policy classifies POST semantically. Trust a reconciled *total*; never trust a
carried-forward *split*. **Still re-derive every connector** — it is cheap for a
machine-readable artifact and it is the only way to know — but expect real drift to concentrate in
the `html_reference` records (mixpanel, gorgias, segment, bamboo-hr, marketo, monday, greenhouse,
lever-hiring — plus notion, already done, and dynamodb, now removed by correction 3), where the
original audit scraped doc pages rather than a spec.

### F2 — `connectorgen batch materialize` is blocked by OAS 3.1 top-level `webhooks`

`cmd/connectorgen/batch_materialize.go:891` (`batchArtifactWebhooksUnknown`) fails any artifact
declaring one or more non-`x-` top-level `webhooks` entries with
`artifact_inventory_unknown: "top-level webhooks (...) cannot be represented as provider request
paths"`. It is deliberately fail-closed.

This locks the sweep's canonical artifact→bundle authoring path out of every OAS 3.1 provider that
declares webhooks — **even though the sweep's own counting policy already excludes webhook events**.
Confirmed blocked so far: **notion** (31 webhook entries), **chargebee** (227 webhook entries).
OAS 3.0 artifacts (stripe, square, twilio, trello, bitbucket) are unaffected.

**CAPTAIN RULING (2026-08-07) — settled, do not relitigate:**

1. **Do NOT relax the gate.** Webhooks are deferred to their own dedicated sweep, filed durably as
   **`cli-webhook-surface-sweep-r1`**, with this finding recorded in it.
2. **Hand-author around the gate** where it blocks a connector (option (c), already chosen here).
3. **When the gate blocks a connector, record it below as blocked-on-the-shared-gate** — it is not
   this lane's failure.
4. **Record every connector's webhook EVENT inventory** (count, plus names where cheap) in the
   Webhook inventory section below. This sweep is the only pass that reads every provider artifact,
   so that inventory is the input `cli-webhook-surface-sweep-r1` needs.
5. **Webhook *management* endpoints stay fully in scope.** Create/list/update/delete of a webhook
   *subscription* is an ordinary REST operation, is already inside the counts, and is **not**
   deferred. Only webhook *events* are deferred. Notion happens to have no management endpoints;
   most other connectors do — do not skip them.

Note also that the batch path *by design* "does not promote direct reads" — so even where it works,
direct reads must be hand-authored on top of it to meet this sweep's bar.

### F3 — the count method is confirmed

The captain's ~25% claim for notion was withdrawn: it counted documentation pages; this lane counts
unique method+path actions per the counting policy, which is correct. **Keep re-deriving exactly
that way** — unique method+path, webhook events excluded, webhook management endpoints included.

### F5 — a connector can have MORE THAN ONE surface test; find them all before authoring

**Cost me a red PR push. Do this check first on every connector.**

gong has **two** independent surface tests in `cmd/connectorgen/`, not one:

| File | Asserts |
| --- | --- |
| `gong_api_surface_test.go` | endpoint total, per-method split, covered/blocked/excluded partition, duplicate absence |
| `gong_full_surface_test.go` | `len(writes.actions)`, `len(operations.operations)`, a `wantCoverage` map of `{stream, direct_read, write}`, plus an exemplar table of specific commands |

I raised the first to 69, went green, ran the targeted test, and **pushed**. The second test then
failed the full-package run with `write actions = 27, want 26` — a real failure on the PR branch that
the targeted `-run TestGongAPISurface...` invocation never touched. CI would have caught it, but only
after a red PR.

**The check, before authoring anything:**

```
ls cmd/connectorgen/ | grep '^<name>_'          # every test file for this connector
grep -rn '<name>' cmd/connectorgen/*_test.go | grep -v '^cmd/connectorgen/<name>_'
go test ./cmd/connectorgen/                     # the WHOLE package, never just -run <one test>
```

Today gong is the only connector with 2; every other has 1. That will change as this sweep adds
lock tests, so re-check per connector rather than trusting this count.

**Corollary — always run the whole package.** A targeted `-run` is for the red/green inner loop
only. Before pushing, run `go test ./cmd/connectorgen/` entire, plus `internal/cli`.

### F4 — `pm docs generate` rewrites all 551 connectors: pre-existing drift in `main`

**Every connector in this sweep will hit this. Budget for it; do not be alarmed by the file count.**

Running `./pm docs generate --dir docs/cli` — required by the CLI/docs parity contract — rewrote
**1,034 files** across all 551 connectors, not just the one being worked on. Inspected rather than
accepted: the churn is **pre-existing drift in `main`**, unrelated to any connector change. The
committed docs carry stream fields as `field()` with an **empty** type; the current generator emits
`field(string)`, `field(boolean)`, `field(integer)`. So `main`'s committed connector docs are stale
against `main`'s own generator, repo-wide.

**Handling (matches the notion precedent, which reverted `docs/connectors/amazon-sqs/**`):**

1. Run the generator.
2. `git checkout --` every `docs/connectors/*` path **except the connector being worked on**.
3. **Keep the worked connector's own file as generated**, even though it also carries the type
   correction. Hand-stripping the drift would commit a file that disagrees with its own generator,
   which is worse than a slightly larger scoped diff.

```
git checkout -- $(git diff --name-only docs/connectors/ | grep -v '^docs/connectors/<name>/')
```

`docs/cli/**` is **not** affected — it regenerates byte-identical.

**This is a real repo condition someone should fix in its own PR** (one mechanical commit reconciling
all 1,034 files). It is deliberately *not* fixed inside a connector parity PR: 14,000 lines of
unrelated churn would bury a 2-operation change and violate the scoped-diff constraint.

## Deployment model (captain ruling 2026-08-07, durable in `data/captain.md`)

Effective from the connector **after** notion — notion was already built and is not to be redone.

**Opus (this lane) keeps every judgement call:** deriving operation counts from provider artifacts;
deciding whether a regenerated diff is genuinely mechanical; answering pipeline gates; resolving
ask-user findings; deciding when to escalate a shared-gate change rather than make it.

**Delegate the mechanical bulk to Sonnet sub-agents:** generating and editing bundle files, running
commands and collecting output, regenerating catalogs and ledgers, fixture assembly, and repetitive
per-operation authoring.

Three safety rules, non-negotiable:

1. A sub-agent **never** answers a gate, resolves an ask-user finding, or decides a diff is safe to
   commit. It produces work; this lane accepts or rejects it.
2. **This lane inspects every regenerated artifact diff itself** before committing. Do not delegate
   "confirm this is purely mechanical" — that judgement is the safeguard. It is what turned the
   1,866-endpoint surface-sync line and the 1,920-line website diff into verified one-object changes.
3. When a sub-agent reports something done, **verify it against reality**. A report is evidence, not
   proof; treating the two as equal is the exact defect this programme has spent two weeks removing.

## PR requirements — get this right on every connector

Two CI gates now block merge. Both were hit on notion (#3894) and both are solved; reuse the pattern.

### `require-linked-issue` (REQUIRED status check, added 2026-08-07)

Enforced by `.github/workflows/pr-issue-guard.yml` → `cmd/prissueguard` →
`internal/coordination/issueguard/guard.go`. Two independent requirements:

1. **Title must be Conventional Commits** —
   `^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\(scope\))?!?: .+`
2. **Body must link an issue** — `Closes #N` / `Fixes #N` / `Resolves #N` for completed work,
   `Refs #N` for stacked or incremental work, or explicit parent/delivery issue wording.

**Validate locally before opening — this costs seconds and avoids a red PR:**

```
go run ./cmd/prissueguard --title "<title>" --body-file <body.md>
```

**Connector issues already exist.** Do not create new ones without checking. Notion had a full tree
(parent #3062 with #3063–#3069) found via `gh-axi search issues "<connector> parity repo:polymetrics-ai/cli"`.
Use `Closes` only for what the PR genuinely completes and `Refs` for the rest — notion used
`Closes #3063/#3064/#3066/#3067/#3068`, `Refs #3062` (parent), `Refs #3065` (binary blocked on the
shared upload runner), `Refs #3069` (certification evidence not in that PR).

### `gsd-workflow-evidence` (also blocking)

`scripts/verify-gsd-workflow` fails any PR touching `cmd/**` or `internal/**` without a matching
change under `.planning/`. So **the GSD phase must be in the same PR as the implementation** — it is
not optional paperwork, it is a merge gate. Reproduce locally with
`GSD_BASE_REF=origin/main bash scripts/verify-gsd-workflow`.

### One PR per connector — split the branch

The sweep branch accumulates work across connectors. Cut a per-connector branch from current
`origin/main` and cherry-pick only that connector's commits, or the PR carries a neighbour's work.
Notion needed this: `fm/cli-notion-parity-sweep-r1` was cut fresh and the dynamodb red test excluded.
Generated website files conflict on cherry-pick — **regenerate them, do not hand-merge**.

## GSD / TDD per connector (captain correction 2026-08-07)

The programme contract is **red-green-refactor TDD driven through GSD, then no-mistakes**. This lane
was running the no-mistakes half only. **From dynamodb (connector 2) onward**, every connector gets
a GSD phase at `.planning/phases/<connector>-parity-sweep-r1`. **Notion is NOT retrofitted** —
already built and committed; redoing it wastes the work.

Established shape, read from `.planning/phases/{stripe-parity-3005, xero-parity-wave04-r1,
youtube-analytics-parity-3456}`:

| Artifact | Purpose |
| --- | --- |
| `PLAN.md` | the operation surface planned before any authoring |
| `TDD-LEDGER.md` | red/validation baseline before production edits → planned assertions → green evidence → refactor/safety notes |
| `VERIFICATION.md` | gate results |
| `RUN-STATE.json` | machine-readable phase state |
| `SUMMARY.md` | optional (stripe has one) |

**The red test is a real, load-bearing artifact, not paperwork.** The repo already carries
per-connector surface tests — `cmd/connectorgen/{xero,github,gong}_api_surface_test.go` (146–210
lines each). Each asserts the derived operation total, the per-method split, the covered/blocked/
excluded partition, `operation_ledger_version`, and duplicate-endpoint absence. That is exactly the
test that makes a re-derived count *stick*: any later drift fails the build instead of silently
passing.

Order per connector: research artifact → **this lane judges the count** → `PLAN.md` → write
`cmd/connectorgen/<name>_api_surface_test.go` and watch it FAIL (red) → sub-agents author the bundle
→ green → `TDD-LEDGER.md`/`VERIFICATION.md` → no-mistakes.

### Cost assessment (requested by the captain)

**Verdict: justified, do not take it back.** Roughly 4 markdown/JSON artifacts (~12–18 KB) plus one
~150–200 line test per connector. The test earns its cost outright — it is the only mechanism that
locks a re-derived count against future drift, which is this sweep's central risk. The four GSD
artifacts are moderate overhead but they carry the red/green evidence that `PROGRESS.md` does not,
and this sweep is explicitly resumable-by-design.

**One real gap it exposes:** notion landed *without* its `cmd/connectorgen/notion_api_surface_test.go`,
so its verified 51 is currently unlocked — nothing fails if a later edit drifts it. That is a missing
test, not a rebuild, so it does not conflict with "do not retrofit notion". Recommend adding just
that one test file for notion.

## Webhook event inventory (input for `cli-webhook-surface-sweep-r1`)

Webhook **events** only. Webhook **management endpoints** are counted as ordinary operations in the
main table and are not listed here.

| Connector | Artifact | Webhook events | Blocks batch pipeline? | Names |
| --- | --- | --- | --- | --- |
| notion | OAS 3.1.0 | **31** | yes — hand-authored instead | commentCreated, commentDeleted, commentUpdated, dataSourceContentUpdated, dataSourceCreated, dataSourceDeleted, dataSourceMoved, dataSourceSchemaUpdated, dataSourceUndeleted, databaseContentUpdated, databaseCreated, databaseDeleted, databaseMoved, databaseSchemaUpdated, databaseUndeleted, fileUploadCompleted, fileUploadCreated, fileUploadExpired, fileUploadUploadFailed, pageContentUpdated, pageCreated, pageDeleted, pageLocked, pageMoved, pagePropertiesUpdated, pageTranscriptionBlockTranscriptDeleted, pageUndeleted, pageUnlocked, viewCreated, viewDeleted, viewUpdated |
| gong | OAS 3.0.1 | **0** | no — OAS 3.0.1 has no top-level `webhooks` construct | — (webhook *management* endpoints are in the 69 and already covered) |
| chargebee | OAS 3.1.0 | **227** | yes — will hand-author | not yet captured (capture at order #23) |
| stripe | OAS 3.0.0 | 0 top-level | no | — |
| square | OAS 3.0.0 | 0 top-level | no | — |
| twilio | OAS 3.0.1 | 0 top-level | no | — |
| trello | OAS 3.0.0 | 0 top-level | no | — |
| bitbucket | OAS 3.0.0 | 0 top-level | no | — |

Notion additionally has **no webhook management endpoints** (no `/v1/webhooks` path in the OAS), so
there is nothing in-scope to implement for it.

## Landed

One line per connector, appended immediately on landing. Format:
`connector | ledger target | landed | blocked-with-named-dependency | PR | certify outcome`

**notion | ledger 50 (stale) → re-derived 51 | 48 executable + 3 partial | 1 blocked (shared file
upload runner) + 5 not-executable | PR https://github.com/polymetrics-ai/cli/pull/3894 | certify:
`make certify-timing` PASSES, 89.1s of a 3m30s budget, 92 real CLI invocations at budget** — branch
`fm/cli-notion-parity-sweep-r1`, cut fresh from `origin/main` (which had advanced to `c057bb81d`,
including #3892 dynamic schema discovery). PR opened; `require-linked-issue`, `gsd-workflow-evidence`,
`pr-title`, `branch-name`, `govulncheck`, `Website generated data`, `Dependency Review` all pass,
0 failed, heavy jobs (verify, connector-boundary, Website checks, CodeQL, snyk) pending.
GSD phase `.planning/phases/notion-parity-sweep-r1` added — honestly labelled **retrospective**,
because it was written after the implementation; dynamodb onward is genuinely red-first.

Detail: 6 ETL streams, 18 direct reads, 24 write actions (21 implemented + 3 partial), 1 blocked
upload, 5 not-executable rows (3 OAuth `disallowed`, 2 provider-`deprecated`). 54 api_surface rows
for 51 documented operations — `POST /v1/search` is one operation carried on two rows (object=page,
object=database, the qualified-path convention already shipped for notion in `main`) and
`PATCH /v1/comments/{comment_id}` likewise carries its two modelled union arms.
Every row carries exactly one disposition; none blank.

Verified: `connectorgen validate` 0 findings · `TestEveryImplementedCommandPassesRuntimePreflight`
passes · `surface-sync --check` clean · `go build ./cmd/pm` OK · binary run: `pm connectors inspect
notion --json`, bare `pm notion` renders 16 command groups, `pm notion page get --help` (direct_read,
json_redacted), `pm notion comment create --help` (reverse_etl), `pm notion block delete --help`
(destructive confirmation + high risk), `pm notion file-upload send --help` (availability planned,
named dependency).

Commits on `fm/cli-top50-fixed-schema-sweep-r1`:
`a886b9ac3` bundle · `22e1b66eb` docs + catalogs + golden transcripts · `e3f0b97e9` website catalog.

**Gates run and green:** `connectorgen validate` (551 connectors, 0 findings) ·
`TestEveryImplementedCommandPassesRuntimePreflight` · `surface-sync --check` clean ·
`make certify-timing` **passes** (89.1s of a 3m30s budget, 92 real CLI invocations at budget) ·
conformance suite · `make docs-check` · `make connector-boundary` · `make agent-contract-check` ·
`make tidy-check` · `go test ./internal/connectors/{commandrunner,engine}`.

**Regenerated diffs, each inspected by this lane before committing:**
- `operation_endpoint_ledger.json` — 110 insertions, all inside the `"notion": []` block. Trap 1 hit
  exactly as briefed: 18 direct reads failed preflight until surface-sync filled the ledger.
- `golden_transcripts.json` — trap 2 handled: diff read first. Exactly one added line per root-help
  transcript, notion joining the connector command-surface list. No other transcript content moved.
- website catalog — 1,920 insertions looked alarming; compared **by object**, not by line: 551
  connectors before and after, none added, none removed, exactly one changed (`Notion`).
- `docs/connectors/amazon-sqs/**` — pre-existing generator drift in `main`, unrelated to notion,
  **reverted** to keep the diff scoped.

**notion is COMPLETE and MERGED** — PR #3894 squash-merged onto `main` as `d71c6206b` on 2026-08-07.
Nothing remains for it. The lock test gap noted in the cost assessment above **also closed**:
`cmd/connectorgen/notion_api_surface_test.go` (170 lines) is present in `main` and byte-identical to
the version staged on this branch, so the re-derived 51 is locked against future drift. Verified by
cherry-picking it onto `main` and getting an empty change.

**gong | ledger 69 → re-derived 69 (exact reconcile) | 67 → 69, all 69 covered and individually
reachable | 0 blocked | PR https://github.com/polymetrics-ai/cli/pull/3895 | certify:
`make certify-timing` PASSES, exit 0, 92 real CLI invocations at budget, 108.6s total** —
PR branch `fm/cli-gong-parity-sweep-r1`, cut fresh from `origin/main` (`d71c6206b`, the notion
merge) with only gong's commits cherry-picked, so the dynamodb red test does not ride along:
`548836773` red test + GSD phase · `6f8381324` bundle + docs · `6a4561ae8` website catalog ·
`8c416761e` full-surface count raise (see **F5**). 13 files, zero non-gong paths.
**A second surface test (`gong_full_surface_test.go`) failed after the first push** with
`write actions = 27, want 26`; caught by running the whole `cmd/connectorgen` package rather than a
targeted `-run`, fixed by raising its counts (26→27 actions, 16→17 operations, coverage
direct_read 29→30 / write 26→27) and locking both new commands in its exemplar table. Nothing
weakened. `go test ./cmd/connectorgen/` now green in full. `prissueguard` validated locally before opening: `ok (5 linked
issues)`; `GSD_BASE_REF=origin/main scripts/verify-gsd-workflow` exit 0.
Issue links: `Closes #2998` (api_surface complete/duplicate-free/partitions every operation exactly
once — fully evidenced); `Refs #2997` parent, `Refs #3000` (direct+binary halves landed, but its
search/query half depends on foundation **#2985**, so not closed), `Refs #3001` (the 24 `partial`
complex-record writes), `Refs #3002`.

Detail: 12 ETL streams, **30** direct reads (+1), **27** write actions (+1, of which **3 binary
multipart**), 69 api_surface rows for 69 documented operations, every row carrying exactly one
disposition, none blank. Availability 45 `implemented` / 24 `partial`; the 24 `partial` are the
pre-existing complex-record reverse-ETL limitation recorded against **#3001**, not introduced here.

Verified: `connectorgen validate` 0 findings · `TestGongAPISurfaceOperationLedger` PASS at 69
(red→green) · `TestEveryImplementedCommandPassesRuntimePreflight` PASS · `surface-sync --check`
clean · conformance PASS · `TestGoldenTranscripts` PASS **unchanged** · `connector-boundary`,
`docs-check`, `tidy-check`, `lint`, `smoke`, `agent-contract-check`, `release-workflow-check` all
PASS · `gofmt`/`go vet` clean · binary run: `pm gong` renders the new `targets` group,
`pm gong targets list --help` (direct_read, json_redacted), `pm gong targets upload-assignments
--help` (reverse_etl, destructive `--confirm`, 4 typed flags), and in a scratch `pm init` project the
write `--preview` advances past project resolution to `missing --credential` — proving real runtime
routing. No credential introduced or printed.

**Traps:** trap 1 did **not** bite (gong's 13 ledger rows are all POST `rest_read`; an unambiguous
GET needs no disambiguating entry, and `surface-sync --check` confirms 0 drift). Trap 2 did **not**
bite (gong was already in the root-help surface list, so transcripts pass unchanged — nothing
regenerated blindly). **F4 did bite** and was handled per the notion precedent.

**Regenerated diffs, each inspected by this lane before committing:**
- connector docs — 1,034 files rewritten by the generator; **1,032 reverted** as pre-existing
  `main` drift (finding F4). Kept gong's own 2 files at 19+/12- each.
- website catalog — compared **by object**: 551 connectors before and after, none added, none
  removed, **exactly one changed (`Gong`)**. Generator summary independently corroborated
  correction 3 with `"database": 2`.

## Notes

- Pre-existing `fm/cli-<name>-parity-wave0N-r1` branches exist on origin for most connectors, but
  they are **stale** — cut before the connector-architecture-v2 rewrite landed, so they diff against
  current `main` with ~800k–1M deletions. They are not live lanes and are not resumed. Only the five
  inherited branches above carry current rebased work.

  **Re-verified with evidence on 2026-08-07** (do not take the paragraph above on trust — this is
  the check to repeat per connector). `fm/cli-dynamodb-parity-wave04-r1`, the one that looked most
  like a live lane, forks at `86d510927` (Fri 31 Jul) and `main` is **52 commits / 639,997
  insertions** ahead of it. Its open PR **#3548 is a draft** whose own body reads "unvalidated cloud
  checkpoint — do not merge yet … did not run review, tests, lint, documentation generation,
  validation pipelines", 0 checks passing, 0 reviews. Decisively, its diff is all
  `internal/connectors/native/dynamodb/*.go` — the **pre-v2 architecture**, replaced by JSON bundles
  under `internal/connectors/defs/`. Stale, not live.

  **Per-connector check to run before starting each one:**

  ```
  git fetch origin 'refs/heads/fm/cli-<name>-parity-*:refs/remotes/origin/fm/cli-<name>-parity-*'
  git merge-base origin/main origin/fm/cli-<name>-parity-<branch>   # ancient fork ⇒ stale
  gh-axi search prs "<name> parity repo:polymetrics-ai/cli"          # draft/unvalidated ⇒ stale
  ```
