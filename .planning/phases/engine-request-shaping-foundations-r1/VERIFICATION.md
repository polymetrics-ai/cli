# Verification — engine-request-shaping-foundations-r1

GSD phase verification. Every result below was produced by **running** something, not by reading
code. That distinction is the captain's standing rule for this wave, prompted by the audit that found
174 commands declared `availability: implemented` which fail at runtime because
`cmd/connectorgen/validate.go` exempts a check `commandrunner/runner.go` enforces. The lesson taken
here: declaration-level validation is not evidence of runtime behaviour, so both are exercised
separately below.

## 1. Local gates

| Gate | Command | Result |
| --- | --- | --- |
| Format | `gofmt -l cmd internal` | empty |
| Vet | `go vet ./...` | clean |
| Build | `go build ./cmd/pm` | ok |
| Bundle validation | `make connectorgen-validate` | **550 connectors checked, 0 findings** |
| Path ownership | `make connector-boundary` | **outcome `clean`, 0 findings, 550 connectors loaded** |
| Website data | `pnpm run gen:website-data` | regenerated; **zero drift** (see §4) |

### Test suite

A bare `go test ./...` cannot complete inside a single tool call in this environment — `internal/cli`
takes ~7½ minutes and `internal/connectors/certify` ~9½ minutes on their own, across 550+ connectors.
Runs were therefore scoped per package group. Every package was run with `-count=1`:

| Packages | Result |
| --- | --- |
| `./internal/connectors/engine/...` (the changed package) | ok 1.4s |
| `./internal/connectors/conformance/...` | ok 13.1s |
| `./internal/connectors/commandrunner/...` | ok 4.1s |
| `./internal/connectors/connsdk/...` | ok 0.4s |
| `./internal/safety/...` | ok 2.0s |
| `./cmd/...` (connectorgen, iconregistrygen, pm, prissueguard) | ok |
| `./internal/connectors/` | ok 0.4s |
| `./internal/app/...` | ok 21.5s |
| `./internal/cli/...` | ok 444.4s |
| `./internal/connectors/certify/...` | ok 556.6s |

CI carries the full suite in one run.

## 2. Declaration side — executed, not read

A scratch connector bundle was written **outside the repository** (this is a foundation PR; the
path-ownership guardrail forbids editing any bundle) declaring all four capabilities at once:

- `pagination: { "type": "start_index", "page_size": 100 }` on a SCIM-shaped stream
- `record_schema` with `"minItems": 1, "maxItems": 10` on a documented request array
- `rest.required_query: [ { "any_of": ["email", "id"] } ]` on a direct-read operation
- `body_type: "base64_upload"` with a `base64_upload` block bounded at 3 932 160 decoded /
  5 242 880 encoded bytes

Validated with the **built binary**, not `go run` of a package under test:

```
$ go build -o <scratch>/connectorgen ./cmd/connectorgen
$ <scratch>/connectorgen validate <scratch>/verifydefs
connectorgen validate: 1 connector(s) checked, 0 findings
```

The first run of that command returned 6 findings — all of them the scratch bundle's own incidental
omissions (five missing `docs.md` headings, one missing `operation_ledger_version`), none of them
about the four capabilities. Recorded because it is the evidence that the validator was actually
exercising the bundle rather than skipping it.

### No regression to the existing corpus

```
$ go build -o <scratch>/pm ./cmd/pm
$ <scratch>/pm connectors list --json | (count)
554
$ <scratch>/pm connectors inspect airtable --json
{ api_version, connector, kind, manifest }
```

The built `pm` still loads every connector after the schema-dialect, loader and meta-schema changes.

## 3. Runtime side — executed against live HTTP servers

Declaration validation proves a bundle *loads*. It does not prove the engine *shapes the request
correctly*, which is exactly the gap the 174-command audit exposed. A temporary harness
(`internal/connectors/engine/verifytmp`, run and then deleted — `git status` confirms it left no
trace) loaded the scratch bundle through `engine.Load` and drove `engine.Read`, `engine.Write` and
`engine.OperationDirectRead` against real `httptest` servers.

```
loaded bundle "scimdemo": 1 streams, 2 writes, 1 operations

start_index pagination (unblocks 2 Airtable operations)
  [PASS] walks every page — err=<nil> records=5 requests=[count=100&startIndex=1 count=100&startIndex=3 count=100&startIndex=5]
  [PASS] stops at totalResults — issued 3 requests, want 3
array cardinality (unblocks 25 Airtable operations)
  [PASS] empty array refused — err=scimdemo action=delete_records record=0: /records: minItems 1 not satisfied (got 0) issued=0
  [PASS] non-empty array accepted — err=<nil> issued=1
  [PASS] maxItems enforced — err=scimdemo action=delete_records record=0: /records: maxItems 10 exceeded (got 11) issued=1
required_query (unblocks 1 Airtable operation)
  [PASS] unfiltered refused — err=operation "scimdemo.enterprise_users.list" requires at least one of query parameters email, id issued=0
  [PASS] filtered accepted — err=<nil> issued=1
bounded base64 upload (unblocks 1 Airtable operation)
  [PASS] payload base64-encoded on the wire — err=<nil> file="bGl2ZSBhdHRhY2htZW50IGJ5dGVz"
  [PASS] local path never transmitted — body keys=[file filename contentType]
  [PASS] declared fields still sent — filename=payload.txt contentType=text/plain
  [PASS] symlink escape refused — err=... base64_upload source_field "attachment_file_path": openat escape.txt: path escapes from parent
  [PASS] traversal refused — err=... base64_upload source_field "attachment_file_path": base64 upload source path must not escape the project root

all runtime checks passed
```

Four results are worth calling out because they are the properties the ledger actually asked for:

1. **`count=100&startIndex=1` … `startIndex=3` … `startIndex=5`** — the SCIM parameter names are
   defaulted correctly and the walk advances by the records the engine extracted, not by a claimed
   `itemsPerPage`. Three requests for five records at page size 100 also shows the `totalResults`
   stop firing rather than the page-size stop.
2. **`issued=0` on the empty array** — the malformed request never reaches the provider. That is the
   whole point of the 25 blocked operations: they could not be declared executable "without risking
   a malformed request".
3. **`body keys=[file filename contentType]`** — `attachment_file_path` is absent. A local
   filesystem path was never transmitted, while the ordinary declared fields still travel.
4. **`openat escape.txt: path escapes from parent`** — that error text is `os.Root`'s own. The
   symlink pointing outside the project root was refused by the kernel-level containment primitive,
   not by a lexical prefix check, which is what the parallel download lane established and what
   `safety.ValidateLocalWritePath` alone cannot do.

## 4. Website catalog

`pnpm run gen:website-data` was **run**, not reasoned about:

```
Wrote 11 docs pages to lib/docs.generated.ts.
Wrote 550 connectors to data/connectors.generated.json; 334 icons copied.
Wrote 550 connectors to lib/connectors.catalog.generated.ts and lib/connectors.catalog.data.generated.json (11413 KB).
Wrote 550 connectors to lib/connectors.generated.ts (73 famous first, 477 alphabetical).
```

`git status -- website/` is empty afterwards: the regenerated data is byte-identical. That is the
expected outcome — this phase adds no connector and edits no bundle — but it was verified rather than
assumed. The `Website checks` job additionally only triggers on `website/**`,
`internal/connectors/icon_data.json`, `docs/connectors/icons/**`, `.github/workflows/website.yml` and
`.gitlab-ci.yml` (`.github/workflows/website.yml:4-9`), none of which this branch touches.

## 5. Constraint compliance

| Constraint | Evidence |
| --- | --- |
| No connector bundle modified | `git status` lists only `internal/connectors/engine/**` and `.planning/**`; `make connector-boundary` outcome `clean` |
| No raw request escape hatch | Method, path and body **structure** stay bundle metadata in all four capabilities. `base64_upload` fills one declared body property from one declared record field; every other field is governed by the action's closed `record_schema`. Nothing added lets a caller choose a method, a path, or an arbitrary body shape. |
| Additive and opt-in | A bundle declaring none of the four behaves identically: the 550-connector validate and the 554-connector binary load both pass unchanged |
| Bounds consistent with the download lane | `os.Root` containment, read one past the limit and reject, clamp declared → engine ceiling (16 MiB, matching `maxOperationDirectReadBytes`), SHA-256 verified against the approved payload digest |

## 6. Deliberate omissions, recorded

- **`items_per_page_path`** is not a declarable field on `start_index`, although SCIM reports
  `itemsPerPage` and the ledger names it. Every job it could do is done more safely by `recordCount`
  (the advance) or `totalResults` (the stop), and using it as a short-page stop would silently
  truncate against a server that caps the page size below the requested count. An unused knob in a
  bundle schema is worse than an absent one; the reasoning is recorded in `paginate.go` beside the
  type.
- **`required_query` on streams** is not implemented. The blocked endpoint is a direct read, and its
  ledger reason explicitly rules out "claiming an unfiltered executable stream", so a stream is the
  wrong surface for it.
- **Go's base64 decoder skips `\r` and `\n` unconditionally, in `Strict()` mode too.** Discovered by
  a failing test, not by reading the docs. `base64.StdEncoding.Strict()` alone would therefore have
  accepted a MIME-wrapped payload and silently re-encoded it into something the operator never
  wrote, so an explicit alphabet check runs first. This is the single most surprising finding of the
  phase.
