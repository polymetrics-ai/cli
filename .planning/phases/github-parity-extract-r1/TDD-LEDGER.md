# TDD-LEDGER — github-parity-extract-r1

Two red/green cycles. The first re-observes the sweep branch's parity red against
current `main` rather than inheriting it; the second is the captain's unblock order.

---

## Cycle 1 — github's documented REST surface is incomplete on `main`

**Red** — observed after cherry-picking only the surface test (`6511a8a45`), before any
bundle change:

```
$ go test ./cmd/connectorgen/ -run 'GitHub|Github'
--- FAIL: TestGitHubDocumentedRESTSurfaceIsComplete (0.00s)
    github_documented_surface_test.go:168: REST endpoints = 505, want 1220 documented operations
    github_documented_surface_test.go:177: restByMethod = map[DELETE:72 GET:259 PATCH:36 POST:91 PUT:47],
                                           want map[DELETE:187 GET:636 PATCH:70 POST:193 PUT:134]
    github_documented_surface_test.go:191: expected "GET /orgs/{org}" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:191: expected "GET /user" — …
    github_documented_surface_test.go:191: expected "POST /markdown" — …
    github_documented_surface_test.go:139: POST /app/installations/{installation_id}/access_tokens: blocked row must carry a 'Named dependency:' marker
    github_documented_surface_test.go:157: 4 synthetic path row(s) are not documented endpoints
FAIL
```

This is the important line of this ledger: the red was **re-derived here**, not copied.
505 was the true state of `main` at `08cc41c87`.

**Green** — after `8d00a55a9` (GET surface), `997d7391b` (covered_by.writes) and
`5ea17aa41` (parity):

```
$ go test ./cmd/connectorgen/
ok  	polymetrics.ai/cmd/connectorgen	12.321s
```

---

## Cycle 2 — the captain's restored commands (`CAPTAIN-ORDER-unblock-commands.md`)

**Red** — `f45640a27` adds `internal/connectors/commandrunner/github_unblocked_commands_test.go`
with no implementation behind it:

```
$ go test ./internal/connectors/commandrunner/ -run TestGitHubRestoredCommandsAreExecutable
--- FAIL: TestGitHubRestoredCommandsAreExecutable (1.55s)
    --- FAIL: .../repo_create     github "repo create" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../repo_delete     github "repo delete" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../repo_archive    github "repo archive" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../repo_unarchive  github "repo unarchive" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../cache_delete    github "cache delete" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../secret_set      github "secret set" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../secret_delete   github "secret delete" availability = "unsafe_or_disallowed", want implemented
FAIL
```

7/7 subtests red. `TestGitHubHeldCommandsStayBlocked` was green from the start — it asserts
`auth token` and `api` stay blocked, which is the state the captain asked to preserve.

**Red (intermediate, found by the validator, not by me)** — wiring `secret set` to
`actions_secrets_secret_name3` surfaced a defect in the already-`implemented` twin:

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
github: cli_surface.json: [cli_surface_safety] implemented reverse ETL command 88 ("secret set")
  flag --encrypted-value maps outside write "actions_secrets_secret_name3" schema:
  record field "encrypted_value" is not declared
github: cli_surface.json: [cli_surface_safety] … --key-id … "key_id" is not declared
connectorgen validate: 551 connector(s) checked, 2 finding(s)
```

and, after declaring them:

```
github: cli_surface.json: [cli_surface_missing_mapping] implemented reverse ETL command 191
  ("secret set-2") for write "actions_secrets_secret_name3" lacks flag mappings for required
  record fields: encrypted_value, key_id
```

The second finding is the real one: `secret set-2` was already shipping `implemented` while
its write action's `record_schema` carried only the path parameter, so it could only ever
PUT an empty body. Both commands now declare the fields the API requires.

**Green** — `93de56c5a`:

```
$ go test ./internal/connectors/commandrunner/ -run 'TestGitHubRestoredCommandsAreExecutable|TestGitHubHeldCommandsStayBlocked|TestEveryImplementedCommandPassesRuntimePreflight'
ok  	polymetrics.ai/internal/connectors/commandrunner	10.778s

$ go test ./internal/connectors/hooks/github/ -run 'ArchiveRepo|UnarchiveRepo' -v
--- PASS: TestExecuteWrite_ArchiveRepoPinsArchivedTrue (0.00s)
--- PASS: TestExecuteWrite_UnarchiveRepoPinsArchivedFalse (0.00s)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings
```

No test was weakened, skipped, or deleted in either cycle. Three tests were added.
