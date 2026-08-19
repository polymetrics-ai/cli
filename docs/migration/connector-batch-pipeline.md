# Connector batch authoring pipeline

This is the repeatable control plane for a connector batch: one branch, one
reviewable manifest, one per-connector gate report, and one PR. It deliberately
does **not** convert survey evidence into an executable claim by itself.

The first delivered batch is
[`cli-provider-artifact-sweep-r1-batch-001.json`](batches/cli-provider-artifact-sweep-r1-batch-001.json).
It contains five candidates and 190 measured provider operations (113 read,
77 write). The immutable survey totals stay separate from the freshly fetched
artifact inventory in the materialization and gate reports.

## Foundation status

The batch was authored after rebasing on `main` with all required foundations:

- #3773 / PR #3869 provides `api_surface.json` operation-ledger v2 provenance:
  provider-artifact records and endpoint-local citations.
- #3870 allows non-redacting `json` and `none` direct-write policies.
- #3868 preserves command-runner content.

The command-runner, shared schemas, and engine remain intentionally untouched
by this pipeline. The materializer consumes the shared v2 contract rather than
adding a competing provenance shape.

## Command ownership

`cmd/connectorgen` remains the one connector authoring tool.

| Command | Existing responsibility | Batch use |
| --- | --- | --- |
| `new` | Scaffold a minimal bundle | Start a post-foundation connector directory; it does not certify an executable surface. |
| `gen` | Regenerate hook/native import sets | Run only when the authored bundle actually introduces an applicable hook/native import. |
| `validate` | Load a bundle and apply structural/semantic checks | Runs once per candidate inside `batch gate`. |
| `surface-sync` | Derive operation-backed command metadata and the shared direct-read endpoint ledger | Its per-bundle metadata check runs inside `batch gate`; run the full branch gate to check the shared ledger without rewriting it. |
| `surface-reconcile` | Derive direct-read `api_surface` coverage and blocked reasons from runtime preflight | Use `--check` to review a stale-reason cohort separately from `batch gate`; it never promotes a command. |
| `boundary` | Detect connector-specific policy outside definition ownership | Run against the final branch as a repository gate. |
| `ownership` | Check changed paths for an owned connector scope | Run for each authored connector slice. |
| `batch plan` | New: turn live ledger evidence into a deterministic manifest | Prepare a 1–40 connector batch without mutating a bundle or calling a provider. |
| `batch materialize` | New: read a selected public OpenAPI/Swagger artifact into the shared v2 operation ledger | Reads a source bundle root and writes only a newly created destination bundle; it fetches only the manifest URL (or an explicit offline cache), records URL/date/SHA evidence, preserves real stream/write bindings, and derives a reachable command surface. |
| `batch gate` | New: run each candidate's existing checks independently | Produce included/drop results rather than stopping at the first bad bundle. |

The new `batch` command is genuinely needed because none of the existing
commands consumes the survey ledger, preserves candidate evidence, or retains
per-candidate drop decisions. It orchestrates existing checks; it does not
parallel the runtime or write a second validator.

A bare `connectorgen batch` prints this namespace's contextual usage and exits
successfully; an invalid subcommand remains a usage error.

## Phase 1: select from the live ledger

The ledger is external to this checkout and intentionally gitignored. Read it
directly at the firstmate-workspace path supplied for the run. `batch plan`
accepts either that absolute path or another explicit snapshot path.

```bash
go run ./cmd/connectorgen batch plan \
  --ledger /Users/karthiksivadas/karthik-agent-workspace/data/cli-provider-artifact-sweep-r1/ledger.json \
  --out docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001.json \
  --size 5 \
  --connector docuseal \
  --connector defillama \
  --connector dockerhub \
  --connector flexmail \
  --connector alpaca-broker-api
```

The command validates an explicit candidate before admitting it:

- `status: done`, `scope_in_current_defs: true`, and a measured total/read/write
  triple;
- `artifact_kind: openapi` or `swagger`, an absolute HTTPS artifact URL, a
  non-empty version, and an ISO full-date `retrieved_at`;
- public access, a stated auth model, evidence source, counting note, and
  processing timestamp;
- a bounded operation count (default 20–250, configurable).

It does not invent a total from the artifact, require total = read + write, or
turn a survey `unknown` into a failure. Nullable counts and a stated reason are
valid for an unknown survey row; they are simply not eligible for a batch.

Output is deterministic: it contains the ledger schema/version metadata and
only the selected records' measured counts and citations. It has no generated
timestamp, so rerunning against the same ledger yields byte-identical JSON.

## First batch selection

The prepared candidates were selected for direct, provider-published,
versioned machine-readable artifacts and moderate sizes—not provider-name
recognition.

| Connector | Operations (read/write) | Artifact evidence | Why first |
| --- | ---: | --- | --- |
| `docuseal` | 23 (7/16) | Provider OpenAPI YAML; version `DocuSeal API 1.0.0; OpenAPI 3.1.0` | Small enough to prove write lifecycle handling, while retaining real mutations. |
| `defillama` | 31 (31/0) | Provider repository OpenAPI JSON; versioned free-tier artifact | Public, no-auth, read-only control case. |
| `dockerhub` | 41 (24/17) | Docker's current OpenAPI YAML export | Mixed read/write surface with a documented token-flow exclusion note. |
| `flexmail` | 43 (23/20) | Provider `openapi.php` JSON; OpenAPI 3.1.0 | Direct machine-readable URL and mixed HTTP verbs. |
| `alpaca-broker-api` | 52 (28/24) | Alpaca's provider repository OpenAPI YAML | Moderate, version-pinned broker API with a documented semantic split. |

All five existing definition directories were present in the checked-out
connector tree, and an open-issue title scan found no active issue named for
these candidates. Public, credential-free artifact probes on 2026-08-06
returned HTTP 200 or 206 for all five URLs. Those probes establish source
availability only; the ledger retrieval dates remain immutable survey evidence,
and the materializer records its own full-date retrieval value and content
SHA-256 in the v2 artifact table.

## Phase 2: materialize from the cited artifact

This step consumes #3869's shared v2 provenance contract; it never adds a
sidecar provenance schema or invents a `covered_by` relation. The source root
is a read-only pre-batch bundle snapshot. The source and destination bundle
paths must not overlap, and the destination root must not already contain the
selected connector: a pre-existing directory is recorded as a `bundle_collision`
drop before any artifact request or file mutation. Fetch only the manifest URL
(or use a verified offline cache for a repeatable local run):

```bash
go run ./cmd/connectorgen batch materialize \
  --manifest docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001.json \
  --source-defs-root <pre-batch-defs-root> \
  --defs-root <new-batch-output-defs-root> \
  --retrieved-at 2026-08-06 \
  --report docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001-materialize.json
```

For each selected manifest record:

1. Fetch or read only its cited public artifact. An artifact URL must be HTTPS,
   have no userinfo, query, or fragment, and resolve exclusively to public
   addresses before every dial; proxy routing is disabled and every redirect is
   revalidated. Only an HTTP 200 response without `Content-Range` is complete;
   HTTP 206 or any response carrying `Content-Range` is an
   `artifact_inventory_unknown` drop before parsing. Store the source URL,
   exact version, full-date retrieval date, and SHA-256. The shared v2 artifact
   table owns URL/date/SHA; the materialization report preserves the survey's
   exact version, and every endpoint joins to the cited artifact through its
   local provenance row. No credential and no live provider API call is
   allowed.
2. Enumerate each provider operation from the artifact. Local Path Item
   references and every HTTP method, including `TRACE`, are resolved. A form
   that cannot be exhaustively represented—such as a non-empty top-level
   OpenAPI 3.1 `webhooks` container, callbacks, an external reference, or an
   unsupported path-item field—is an `artifact_inventory_unknown` drop with a
   concrete reason; it never yields a partial inventory count. Give every
   enumerated endpoint exactly one result: executable coverage,
   provider-blocked with a concrete reason, or justified exclusion. A generic
   unclassified/default omission is a failure, not a temporary state. An
   executable source binding must match the artifact's exact method and path:
   different slash spelling, method case, or any other equivalence is never
   inferred. A future equivalence exception requires its own cited
   provider-documentation URL and retrieval date; this materializer does not
   create that evidence. `TRACE` and `OPTIONS` remain in the artifact total as
   method-specific protocol-metadata exclusions, never provider-blocked or
   provider-mutation rows.
3. Copy the reviewed source bundle into the fresh, batch-owned destination,
   then generate its `operations.json` and `cli_surface.json` from the
   reviewed artifact inventory. In this first lane `operations.json` is an
   explicit empty direct-executor catalog: provider operation classification
   lives in the v2 `api_surface.json`, and no direct read can be promoted while
   its only runtime policy is redacting. Stream and existing reverse-ETL write
   commands are derived only from real bindings and pass production runtime
   preflight before any write; zero reachable implemented commands or a
   preflight failure is a named drop. A write action with a required structured
   object/object-array record is intentionally not exposed as a scalar-flag
   namespace command; it remains available through the existing generic plan →
   preview → approval → execute workflow.
4. Only promote a command to `availability: implemented` after it has a real
   executor and the v2 `api_surface.json` contains its cited method/path row.
   Every other operation stays provider-blocked or justified-excluded with its
   reason and endpoint-local artifact citation.
5. Keep the reviewed source bundle's `httptest` fixtures and connector documentation accurate
   before materialization. The materializer copies that bundle; it does not synthesize fixtures or
   prose. Fixtures are synthetic and contain no secret values. No live provider request is used.
6. For writes, retain plan → preview → approval → execute. Destructive writes
   require typed confirmation; non-idempotent writes remain non-retriable.
   Never introduce a generic HTTP write escape hatch.

The full `surface-sync --check` remains a separate branch derivation gate for
implemented operation-backed direct commands and the shared direct-read endpoint
ledger. It deliberately leaves stream and reverse-ETL command bindings intact,
because their executor truth is `streams.json` or `writes.json`, not an invented
direct operation.

No-redaction is an explicit batch-gate rule. `batch gate` drops a candidate
that declares `redact_fields`, an output policy containing `redact`, or either
legacy repository-contents policy (which redacts despite its name). #3870's
`json` and `none` direct-write policies are acceptable. Current direct-read
policies are redacting, so an authored batch must leave a direct-read operation
provider-blocked or justified-excluded unless a separate, shared-owned
non-redacting direct-read capability lands; stream and reverse-ETL executors
remain available where their real implementation applies.

### Batch 001 materialization result

The checked-in
[`materialize report`](batches/cli-provider-artifact-sweep-r1-batch-001-materialize.json)
records five included connectors and zero drops on 2026-08-06.

| Connector | Fresh artifact operations | Executable | Provider-blocked | Excluded | Reachable commands |
| --- | ---: | ---: | ---: | ---: | ---: |
| `docuseal` | 23 | 10 | 0 | 13 | 9 |
| `defillama` | 31 | 10 | 0 | 21 | 10 |
| `dockerhub` | 54 | 4 | 0 | 50 | 4 |
| `flexmail` | 43 | 5 | 22 | 16 | 5 |
| `alpaca-broker-api` | 52 | 10 | 5 | 37 | 10 |

Docker Hub's 54 is intentionally not rewritten to the ledger's 41: the ledger
counting note excludes 13 documented authentication/access-token-flow
operations, while the v2 inventory records and classifies every operation in
the fetched artifact. DocuSeal's `create_submission` remains a real generic
reverse-ETL write action but is not a connector-namespace command because its
required `submitters` value is an object array, not a truthful scalar CLI flag.

### Post-main capability re-gate

After rebasing on current `origin/main` at `5da755596`, batch 001's checked-in
artifact inventory was gated again. It remains
five included, zero dropped, with **39 executable**, 27 provider-blocked, and
137 excluded rows across 203 declared artifact operations.

The executable count did not rise merely because two foundations are present:
the real runner currently accepts only redacting direct-read policies, whereas
`json` and `none` are direct-write policies. The batch's no-redaction rule
therefore still blocks promotion of the 21 unmatched GET rows. Separately,
the cited DocuSeal artifact declares the six document-creation/update requests
as `application/json` structured document payloads, not literal
`multipart/form-data`; #3871's opt-in `rest.multipart` contract cannot be
used to make those different provider contracts executable. The materializer
and runtime gate made that result; no `api_surface.json` reason was hand-edited.

## Phase 3: individual gate and batch report

Run the gate once per candidate before the final batch gate. `--connector`
selects a manifest record without hand-editing its survey evidence:

```bash
go run ./cmd/connectorgen batch gate \
  --manifest docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001.json \
  --defs-root internal/connectors/defs \
  --connector docuseal \
  --report docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001-docuseal-gate.json
```

Then run one final gate for the manifest:

```bash
go run ./cmd/connectorgen batch gate \
  --manifest docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001.json \
  --defs-root internal/connectors/defs \
  --report docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001-gate.json
```

For every candidate, the gate:

1. Runs the existing bundle validator.
2. Runs `syncBundle(..., true)`, the per-bundle metadata portion of
   `surface-sync --check`, and drops metadata drift rather than silently fixing
   it during validation. The full branch gate checks the shared runtime ledger.
3. Loads that one bundle and requires exactly one complete v2 provenance
   artifact whose URL and every endpoint-local source URL exactly match the
   manifest's cited artifact URL. Legacy, incomplete, missing, or
   URL-mismatched provenance is a named `provenance` drop before runtime
   preflight; the report records its explicit `provenance_refusals` count.
4. Rejects `TRACE` and `OPTIONS` `covered_by` bindings and accepts those methods
   only as their method-specific protocol-metadata exclusions. They remain in
   declared operation totals and are never counted executable or
   provider-blocked.
5. Calls the real `commandrunner.Preflight` for every implemented command. It
   does not copy the runner's rules. This is the safeguard against commands
   that validate but fail at `pm <connector> <command>` with
   `connector_command_blocked`.
6. Rejects redacting declarations before runtime preflight: no
   `redact_fields`, redacting output policy, or legacy repository-content
   policy can enter the batch.
7. Requires `api_surface.json` and a non-empty `cli_surface.json`, then reports
   declared endpoint operations as executable, provider-blocked, or excluded.
   The shared v2 validator supplied by #3869 enforces endpoint provenance when
   the post-foundation bundle is validated.

The report includes selected candidates, included candidates, named drops and
their stage/reason, its explicit provenance-refusal count, ledger surveyed
operations, declared operations, and the executable/blocked/excluded split. It
is written even when one or more
candidates fail, and the command returns nonzero only after evaluating them
all.

## Explicit drop protocol

A failed candidate never makes the rest of the batch fail and never remains
quietly on the branch:

1. Keep the failed `*-gate.json` report in the branch; it is the durable record
   of the connector and concrete failure reason.
2. Establish ownership from the pre-batch revision before changing a failed
   bundle. Only a directory that was absent before the batch and was created by
   the batch may be removed with `git rm -r --
   internal/connectors/defs/<connector>`. For a pre-existing bundle, restore or
   revert only the branch-owned changes beneath that directory from the
   pre-batch revision; never delete the directory. A `bundle_collision` is
   refused without writing or deleting its target. Batch 001's five definition
   directories existed before this branch, so they always use the restore/revert
   path if dropped.
3. Regenerate the manifest from the same ledger with only the surviving
   `--connector` values. Do not hand-edit the count/evidence fields.
4. Re-run `batch gate` and commit the updated manifest, bundle removal, and
   final clean gate report together. A PR is not opened with any unresolved
   report drop.

This is intentionally explicit instead of an automatic recursive deletion
inside a validation command: the checked-in report identifies the exact target,
and the operator keeps normal Git review/recovery for the branch mutation.

## Required branch gates after a clean batch gate

Run the normal targeted connector checks and repository gates on the same
branch, including `surface-sync --check`, connector-boundary, ownership checks,
and the actual runtime sweep:

```bash
go test ./internal/connectors/commandrunner \
  -run '^TestEveryImplementedCommandPassesRuntimePreflight$'
```

Then run the standard individual `make` gates documented in `AGENTS.md`, rebase
on the current `main`, and repeat the local validation before requesting one
PR. Do not merge the PR.

## Capacity claim

`batch plan` has a hard maximum of 40 candidates and `batch gate` performs
failure-isolated checks for each one, so the control plane can mechanically
support 30–40 candidates per PR. The proven first batch is **five authored,
five individually gated, zero dropped**. The honest operating estimate is
**30 connectors per batch only for small, complete, provenance-ready provider
surfaces**: it retains a reviewable failure-isolated report and one PR/CI merge
cost. Increase to 40 only after another 30-connector run shows that its
generated artifact diff and per-candidate report remain reviewable; the limit
is review throughput, not the tool's loop.

The Top-50 provider programme is intentionally not treated as such a uniform
batch. Its 13,761-operation cohort ranges from an 11-operation provider to
Zoom's 1,913-operation multi-module surface. The planned size tiers, readiness
gates, and proposed merge units are in
[`cli-top50-surface-audit-r1-batch-plan.md`](batches/cli-top50-surface-audit-r1-batch-plan.md).
For that cohort, use one-provider PRs for 250+ operations, 2–3 complete
provider bundles for 100–249, and 4–6 only below 100; large-provider work may
be sequenced by module inside a branch, but its final artifact inventory and
runtime gate must remain whole-provider.
