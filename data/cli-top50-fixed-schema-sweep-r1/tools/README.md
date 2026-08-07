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
