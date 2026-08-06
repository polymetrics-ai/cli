# Connector batch authoring pipeline

This is the repeatable control plane for a connector batch: one branch, one
reviewable manifest, one per-connector gate report, and one PR. It deliberately
does **not** convert survey evidence into an executable claim by itself.

The first prepared batch is
[`cli-provider-artifact-sweep-r1-batch-001.json`](batches/cli-provider-artifact-sweep-r1-batch-001.json).
It contains five candidates and 190 measured provider operations (113 read,
77 write). It is a preparation artifact, not an authored connector batch.

## Current external gate

Do not author a bundle from this manifest until this remaining foundation has
merged to `main` and the branch has been rebased on `main` (never on a sibling
branch):

- #3773 / PR #3869 provides `api_surface.json` operation-ledger v2 provenance:
  provider-artifact records and endpoint-local citations.

#3870 is merged at `ee26d20fc`: direct writes can declare non-redacting `json`
or `none`. #3868 is merged at `50deaade9`: command-runner content preservation
is current on `main`. They are consumed by the batch gate below; #3869 is the
only external authoring dependency.

The command-runner, shared schemas, and engine are intentionally untouched by
this pipeline. A v2 provenance writer before #3869 would compete with a shared
contract that is still landing; a new executable policy before #3870 would risk
reintroducing redaction. Neither is acceptable.

## Command ownership

`cmd/connectorgen` remains the one connector authoring tool.

| Command | Existing responsibility | Batch use |
| --- | --- | --- |
| `new` | Scaffold a minimal bundle | Start a post-foundation connector directory; it does not certify an executable surface. |
| `gen` | Regenerate hook/native import sets | Run only when the authored bundle actually introduces an applicable hook/native import. |
| `validate` | Load a bundle and apply structural/semantic checks | Runs once per candidate inside `batch gate`. |
| `surface-sync` | Derive operation-backed command metadata from `operations.json` | Runs in check mode once per candidate inside `batch gate`; it detects drift without rewriting it. |
| `boundary` | Detect connector-specific policy outside definition ownership | Run against the final branch as a repository gate. |
| `ownership` | Check changed paths for an owned connector scope | Run for each authored connector slice. |
| `batch plan` | New: turn live ledger evidence into a deterministic manifest | Prepare a 1–40 connector batch without mutating a bundle or calling a provider. |
| `batch gate` | New: run each candidate's existing checks independently | Produce included/drop results rather than stopping at the first bad bundle. |

The new `batch` command is genuinely needed because none of the existing
commands consumes the survey ledger, preserves candidate evidence, or retains
per-candidate drop decisions. It orchestrates existing checks; it does not
parallel the runtime or write a second validator.

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

## First prepared batch

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
returned HTTP 200 or 206 for all five URLs. Those probes only establish source
availability; their ledger retrieval dates remain the immutable survey evidence
in the manifest. The authoring run must retrieve each artifact again and record
its new full-date retrieval value and content SHA-256 in the v2 artifact table.

## Phase 2: materialize only after the foundations merge

This is intentionally a gated authoring step, not a command to run on the
current base. It must consume #3869's shared v2 provenance contract; it must
not add a sidecar provenance schema or an invented `covered_by` relation.

For each selected manifest record:

1. Fetch or read only its cited public artifact. Store the source URL, exact
   version, full-date retrieval date, and SHA-256 in the v2 provider-artifact
   table. No credential and no live provider API call is allowed.
2. Enumerate each provider operation from the artifact. Give every endpoint
   exactly one result: executable coverage, provider-blocked with a concrete
   reason, or justified exclusion. A generic unclassified/default omission is
   a failure, not a temporary state.
3. Generate the bundle's `operations.json` and `cli_surface.json` from the
   reviewed artifact inventory. Facts that can be derived stay derived:
   `surface-sync` fills command `api_surface`, flag `maps_to`, operation
   byte limits, and supported output policy from `operations.json`. Do not
   hand-copy them.
4. Only promote a command to `availability: implemented` after it has a real
   executor and the v2 `api_surface.json` contains its cited method/path row.
   Every other operation stays provider-blocked or justified-excluded with its
   reason and endpoint-local artifact citation.
5. Add `httptest` fixtures and the usual connector documentation. Fixtures are
   synthetic and contain no secret values. No live provider request is used.
6. For writes, retain plan → preview → approval → execute. Destructive writes
   require typed confirmation; non-idempotent writes remain non-retriable.
   Never introduce a generic HTTP write escape hatch.

The materializer itself is deferred because its output is the shared-contract
surface that #3869 is changing. The work now prepares its deterministic input,
gate, output report, and no-redaction rules; the only safe next code slice
after #3869 merges is the contract-aware artifact-to-bundle emitter.

No-redaction is an explicit batch-gate rule. `batch gate` drops a candidate
that declares `redact_fields`, an output policy containing `redact`, or either
legacy repository-contents policy (which redacts despite its name). #3870's
`json` and `none` direct-write policies are acceptable. Current direct-read
policies are redacting, so an authored batch must leave a direct-read operation
provider-blocked or justified-excluded unless a separate, shared-owned
non-redacting direct-read capability lands; stream and reverse-ETL executors
remain available where their real implementation applies.

## Phase 3: individual gate and batch report

After authoring, run one gate for the manifest:

```bash
go run ./cmd/connectorgen batch gate \
  --manifest docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001.json \
  --defs-root internal/connectors/defs \
  --report docs/migration/batches/cli-provider-artifact-sweep-r1-batch-001-gate.json
```

For every candidate, the gate:

1. Runs the existing bundle validator.
2. Runs `syncBundle(..., true)`, the same non-mutating check behind
   `surface-sync --check`, and drops metadata drift rather than silently fixing
   it during validation.
3. Loads that one bundle and calls the real
   `commandrunner.Preflight` for every implemented command. It does not copy
   the runner's rules. This is the safeguard against commands that validate but
   fail at `pm <connector> <command>` with `connector_command_blocked`.
4. Rejects redacting declarations before runtime preflight: no
   `redact_fields`, redacting output policy, or legacy repository-content
   policy can enter the batch.
5. Requires `api_surface.json` and a non-empty `cli_surface.json`, then reports
   declared endpoint operations as executable, provider-blocked, or excluded.
   The shared v2 validator supplied by #3869 enforces endpoint provenance when
   the post-foundation bundle is validated.

The report includes selected candidates, included candidates, named drops and
their stage/reason, ledger surveyed operations, declared operations, and the
executable/blocked/excluded split. It is written even when one or more
candidates fail, and the command returns nonzero only after evaluating them
all.

## Explicit drop protocol

A failed candidate never makes the rest of the batch fail and never remains
quietly on the branch:

1. Keep the failed `*-gate.json` report in the branch; it is the durable record
   of the connector and concrete failure reason.
2. Remove only the named generated bundle directory from the branch with
   `git rm -r -- internal/connectors/defs/<connector>` after the validation run
   has ended. Do not remove another candidate, shared schema, or runner file.
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
support 30–40 candidates per PR. The honest immediate batch size is **five
prepared, zero authored**: #3869's v2 provenance contract remains unmerged, so
no evidence yet supports a claim that 30–40 fully authored connectors are
merge-ready. After this five-connector proof passes, start at 30 and increase
to 40 only if artifact materialization and the per-candidate report stay
reviewable.
