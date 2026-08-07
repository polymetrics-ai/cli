# Sweep tooling — committed on purpose

**The previous worker's `tools/` directory was untracked and did not survive its worktree.** The
handoff still pointed at it, so the next worker (me) rebuilt the derivation approach from scratch.
These files are committed so that does not happen a third time. They are working tools, not products:
read them, adapt them, replace them.

Nothing here is a substitute for `PLAN.md`'s hazards. The generators encode the *mechanical* rules;
the four non-mechanical judgements (read-vs-write, stream-vs-direct-read, binary detection,
named-dependency blocking) are made in the per-connector planning docs and only then expressed here.

## Cached artifacts are NOT committed — re-fetch them

They are 12.9 MB and 5.8 MB. **Always check the byte count against the recorded one**: identical
bytes are what prove it is the same artifact rather than a lookalike, and let you reproduce a
derivation instead of trusting it. Both reproduced exactly on 2026-08-07.

```bash
SCRATCH=/tmp/sweep && mkdir -p "$SCRATCH"

# github — expect exactly 12,920,264 bytes (openapi 3.0.3, info.version 1.1.4)
curl -sS -w 'HTTP %{http_code} bytes=%{size_download}\n' \
  -o "$SCRATCH/api.github.com.json" \
  'https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json'

# workday-rest — the manifest is a DIRECTORY, expect exactly 617,538 bytes,
# then fetch each of the 52 production service specs it names.
curl -sS -o "$SCRATCH/workday-services.json" \
  'https://community.workday.com/sites/default/files/file-hosting/restapi/services2026.30.json'
mkdir -p "$SCRATCH/wd" && cd "$SCRATCH/wd"
python3 -c "import json;[print(s['specFilePath']) for s in json.load(open('$SCRATCH/workday-services.json'))['productionConfidenceLevel']]" \
  | xargs -P 8 -I{} curl -sS -o {} \
      "https://community.workday.com/sites/default/files/file-hosting/restapi/{}"
```

## `gen_github_gets.py` / `gen_github_writes.py`

Run against a bundle directory; they append to `api_surface.json`, `cli_surface.json`,
`writes.json` and `operations.json` in place, preserving the files' no-trailing-newline formatting.
They expect the artifact next to themselves as `api.github.com.json`.

```bash
python3 gen_github_gets.py   internal/connectors/defs/github
python3 gen_github_writes.py internal/connectors/defs/github
```

**Regenerate from a clean tree, never hand-patch the output.** Both were re-run from
`git checkout -- internal/connectors/defs/github/` after every rule fix; that is why the two schema
rejections they hit (`rest_read` POST needs `body_schema`, and needs
`content_type: application/json`) are fixed in the generator rather than in the bundle.

Rules worth carrying to the next connector, each of which cost something to learn:

- **Never emit a paging flag.** `PAGING_PARAMS` is a blocklist and the generator raises rather than
  authoring one. A parallel foundation lane derives paging from each connector's pagination spec.
- **Flags cover path variables and REQUIRED query parameters only.** Optional filters are not
  operations.
- **A required record field with no scalar leaf makes the command `partial`, not `implemented`.**
  `validate`'s `checkCLISurfaceWriteFlags` recurses into nested required objects; mirror that
  recursion or you will ship a command the runtime blocks.
- **Plain `direct_read` (no `operation`) adds nothing to `operation_endpoint_ledger.json`.** That is
  the low-blast-radius shape for a detail GET. Operation-backed reads are worth it only when you
  need the derivation guarantee.

## `probe_reachability.sh`

The reachability probe, and the reason it exists in this form:

```bash
go build -o /tmp/pm-gh ./cmd/pm
python3 -c "import json;[print(c['path']) for c in json.load(open('internal/connectors/defs/<name>/cli_surface.json'))['commands'] if c['availability'] in ('implemented','partial')]" > /tmp/cmds.txt
xargs -P 12 -I{} ./probe_reachability.sh "{}" < /tmp/cmds.txt
```

**Exit status proves nothing.** `pm <connector> <nonsense> --help` renders the connector's group help
and **exits 0** — that is the documented namespace behaviour. An exit-code-only probe reports 749
passes it never verified. The script asserts the rendered `NAME` line instead. Edit the connector
name at the top before use.

Expect ~1.2 s per invocation. At `-P 12` a 1000-command sweep takes several minutes — run it in the
background, and do not run two sweeps into the same output file.

## `gen_workday_rest.py`

Builds workday-rest's whole surface — 907 documented rows, 911 commands, 252 write actions — from
`.planning/phases/workday-rest-parity-sweep-r1/DERIVED-OPERATIONS.json` plus the 52 cached service
specs. It **rewrites** rather than appends, so it is re-runnable:

```bash
python3 gen_workday_rest.py internal/connectors/defs/workday-rest --reads
python3 gen_workday_rest.py internal/connectors/defs/workday-rest --all /path/to/specs
go run ./cmd/connectorgen surface-sync     # fills the 7 rest_read ledger entries
```

Always regenerate from a clean tree (`git checkout -- internal/connectors/defs/`) — it was re-run
three times that way, which is why each fix lives in the generator and not in the bundle.

Rules it encodes that cost something to learn:

- **Command names come from the endpoint**, because Workday declares an `operationId` on only 21 of
  920 rows. A trailing `{var}` marks a detail read; a non-trailing one becomes a `by-<var>` word.
  Verified to yield **907 distinct names for 907 operations** — check collisions before authoring.
- **A read-only POST is named `read-*`, never `create-*`.** The verb is what a user reads before
  running the command; no test can catch a name that lies.
- **Binary is read from `produces`, never from the path.** `?type=viewContent` and `?type=viewFile`
  sound like downloads and declare `application/json` only.
- **A collapsed query-string behaviour uses `omit_when_absent`**, or the variant becomes the default.
- `required_mapping_paths` is a **transcription** of `validate.go`'s recursion, not a restatement of
  the rule — `AGENTS.md` is explicit that hand-copied runtime rules drift.

## `derive_jira.py` / `gen_jira.py`

jira is a **from-nothing** connector in workday-rest's sense: 15 `api_surface` rows, twelve of them
comma-joined or wildcard "and similar" families standing for 602 endpoints, and no
`cli_surface.json`, `writes.json` or `operations.json` at all. A wildcard row is not an operation
(finding 24), so this is a restructure and the count moves 15 → 617.

```bash
curl -sS -o /tmp/sweep/jira.json \
  'https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json'
python3 derive_jira.py /tmp/sweep/jira.json > \
  .planning/phases/jira-parity-sweep-r1/DERIVED-OPERATIONS.json
python3 gen_jira.py internal/connectors/defs/jira \
  .planning/phases/jira-parity-sweep-r1/DERIVED-OPERATIONS.json /tmp/sweep/jira.json classify
python3 gen_jira.py ... all
go run ./cmd/connectorgen surface-sync
```

**⚠️ THE BYTE-COUNT CHECK DOES NOT WORK FOR THIS ARTIFACT, AND THAT IS A FINDING.** The ledger's
version-pinned URL (`?_v=1.8516.72`) returns 404; the unpinned URL serves a **rolling snapshot**
whose `info.version` is `1001.0.0-SNAPSHOT-<git sha>`. On one calendar day it went 2,445,625 →
2,449,760 bytes, 420 → 421 path keys, 616 → 617 operations. The derivation therefore records the
artifact's **sha256** instead, which identifies the document even after the URL has moved on. Any
connector whose artifact URL is unpinned needs the same treatment — check `info.version` for a
snapshot marker before trusting a recorded byte count.

Rules `gen_jira.py` encodes that cost something to learn:

- **"No schema declared" and "declared as a string" must stay distinguishable.** Jira spells the
  first two ways — an absent `schema` key on a `*/*` avatar upload and a literal `"schema": {}` on
  the entity-property PUTs — and a deref that falls through to `type: string` collapses them into
  the second. That misfiled 12 blocked rows into the wrong class on the first run.
- **Binary is GET-only and read from 2xx responses only.** Atlassian attaches the same content map
  to every response code, so the avatar reads declare `image/png` on their 401/403/404 as well; and
  the avatar *upload* sits in the same resource family, so a rule keyed on the path or on "declares
  a non-JSON media type" ships a mutation as a download.
- **Command names come from `operationId`**, the opposite of workday-rest: Jira declares one on all
  617 and they are collision-free after kebab-casing. Thirteen legacy Connect/Forge ids carry a
  `SomeResource.` prefix and a `_get` suffix; both are stripped and the result is asserted unique.

## `check_red_observed.py`

**PROGRESS.md has referenced this tool since github, and it was never committed** — running it
raised `No such file or directory`, so three workers were told an enforcement existed that did not.
It now does:

```bash
python3 check_red_observed.py --all          # every *-parity-sweep-r1 phase
python3 check_red_observed.py jira
```

It fails a connector whose `RUN-STATE.json` claims `red_confirmed: true` without output that could
only have come from a real `go test` run: a `--- FAIL:` line, a `<file>_test.go:<line>:` assertion
location, a `want <expected>` clause, and no placeholder text. Placeholders are matched with
**anchored** boundaries — a bare substring search flags `n/a` inside `expression/analyse`, a real
Jira endpoint, and a check that cries wolf gets disabled.
