# Issue #4359 — Review convergence record

## Immutable discovery frame

- Base SHA: `b9b2478b3b2451d632d28b9aa138a170ad835110`
- Review SHA: `b9b2478b3b2451d632d28b9aa138a170ad835110`
- Merge base: `b9b2478b3b2451d632d28b9aa138a170ad835110`
- Branch: `fm/cli-composite-provider-path-foundation-r1`
- Issue: #4359
- Production changed files at discovery start: none
- Generated files at discovery start: none
- Read-only integration witness: Batch-1 CircleCI tree `5de7078bfbe2c21db9e200dafe29adfba9e0f91b`; it is not merged, reset, checked out, or modified by this branch.
- Review policy substitution: the usual Claude audit is unavailable in this environment. A separate fresh-context Codex audit will review this immutable SHA and this record before any production mutation.

## Source-to-execution map

| Stage | Current owner | CircleCI proof/reachability observation |
| --- | --- | --- |
| Provider source | Batch-1 `sources/circleci-operation-source-lock.json` | Pinned OpenAPI 3.0.3, artifact `https://circleci.com/api/v2/openapi.json`, digest `61c6…d07`; all affected source paths use `{project-slug}`. |
| Source declaration | Batch-1 `api_surface.json` plus retained source lock; foundation `composite_provider_path_identity.json` | Canonical provider method/path stays visible in provider declaration JSON. The separately embedded foundation record carries the minimum source-cited per-binding identity needed at runtime; no provider route is rewritten. |
| CLI declaration | Batch-1 `cli_surface.json` | Eleven implemented ETL/reverse-ETL commands reference one canonical provider endpoint each. At `b9…`, this generated file is intentionally absent, so the proof is tested with a faithful in-memory declaration fixture rather than copied into this foundation branch. |
| Runtime transport | `streams.json`, `writes.json` | Existing fixed templates send `{{ config.vcs_type }}/{{ config.org }}/{{ config.repo }}`. |
| Equivalence boundary | `internal/connectors/engine/command_binding.go:170` | `proveCommandEndpointEquivalence` presently only equates same-number placeholders, base paths, hooks, GraphQL, query/absolute/suffix/annotation variants; it rejects the one-to-three identity. |
| Commandrunner | `internal/connectors/commandrunner` preflight | It consumes the resolved command surface and must remain unchanged except to observe the proof. |
| App/direct API | `internal/app` transport dispatch plus engine read/write | Existing request dispatch is downstream of preflight. The proof must not interpolate or dispatch a request. |
| CLI | `internal/cli` hand parser/command runner bridge | A valid command must only reach credential resolution after engine proof and existing required-input validation. |

## Six-lane eligibility matrix

| Lane | Status at review SHA | Required invariant |
| --- | --- | --- |
| ETL | affected | Only the six cited CircleCI stream bindings may use the proof; records, schema, pager, method, and origin remain unchanged. |
| Reverse ETL | affected | Only the five cited write bindings may use the proof; plan/preview/approval/execute and typed action/body semantics remain unchanged. |
| Direct read | not admitted | A direct-read operation cannot borrow this stream/write proof or acquire a route. |
| Direct write | not admitted | A direct-write operation cannot borrow this stream/write proof or accept a raw body/method/path. |
| Binary download | not admitted | No binary operation can match, and byte/response contracts remain untouched. |
| Binary upload | not admitted | No upload action can match, and filesystem/media/approval contracts remain untouched. |

## Adversarial input matrix

| Class | Required outcome |
| --- | --- |
| Missing/partial/reordered/repeated/extra config segment | reject equivalence |
| Other connector, wrong cited source URL/artifact, wrong provider placeholder | reject equivalence |
| Runtime absolute URL, query/suffix, changed literal prefix/suffix, path traversal, changed method | reject equivalence unless already covered by its pre-existing independent proof branch |
| Different binding kind, origin, action, body, operation, binary contract | reject equivalence and preserve its own eligibility gate |
| Valid exact CircleCI source identity + `vcs_type/org/repo` templates | resolve one named composite proof; later commandrunner stops at missing credential before I/O |

## Frozen initial findings

| ID | Severity | Lens | Reachability trace | Finding/disposition |
| --- | --- | --- | --- | --- |
| F-4359-001 | high | declaration reachability / provider semantics | source `{project-slug}` → API surface → CLI → stream/write transport → `proveCommandEndpointEquivalence` | Existing complete CircleCI paths are rejected because placeholder cardinalities differ. Fix with a closed, source-cited engine proof only. |
| F-4359-002 | high | security / closed surface | prospective declaration → equivalence → preflight → App request | A general placeholder expansion would turn source metadata into a route escape. Require exact connector, source citation, placeholder, component sequence, and literal path identity. |
| F-4359-003 | medium | lane isolation | all six command intents → resolver | Proof must be constrained to the currently resolved exact command endpoint and must not alter direct/binary eligibility or execution code. |
| F-4359-004 | high | audit substitution / source identity | API surface → command endpoint ref → resolver | The current API surface/endpoint reference lacks source ID, operation ID, location, and artifact digest; a URL alone cannot prove the pinned identity. Accepted: the separately embedded, source-cited per-binding identity record carries and validates the full closed identity without embedding repository-only API-surface coverage. |
| F-4359-005 | high | admission reachability | `b9…` CircleCI bundle → `synthesizeCommandSurface` | Current main has no CircleCI `cli_surface.json`, so it cannot run the eleven command paths in the shipped binary. Accepted: do not import Batch-1 declarations; use a faithful in-memory command-surface fixture for foundation tests and record the binary proof as pending Batch-1 integration. |

## Mandatory lens status

| Lens | Status | Evidence |
| --- | --- | --- |
| Architecture/data flow | complete | Source-to-execution map above. |
| Happy/bad/edge behavior | complete | Adversarial input matrix above. |
| State machine/concurrency | not applicable | Equivalence is pure load/preflight validation; no state, callback, retry, or goroutine changes are planned. |
| Security/secret taint | complete | The proof reads declaration metadata only; no credentials, headers, query values, or responses are introduced. |
| Retry/rate-limit/resume/idempotency | complete | Existing ETL/write semantics are explicitly out of scope and must remain covered by lane tests. |
| Output integrity | complete | No response or output path changes; credential-boundary tests prohibit provider I/O. |
| Declaration reachability/closed surface | complete | Exact source, connector, placeholder, component, method, and literal-path negative cases are planned. |
| CLI/App parity | complete | Commandrunner and fresh-project CLI/App evidence are planned. |
| Provider semantics | complete | Source citation/digest plus exact provider endpoint identity are retained. |
| Tests/evidence | complete | Red/green, six-lane, generated checks, binary proof, and fresh review are enumerated in plan/checklist. |

## Independent audit substitution and convergence

Fresh-context Codex audit completed against `b9b2478b3b2451d632d28b9aa138a170ad835110` with the read-only Batch-1 witness `5de7078bfbe2c21db9e200dafe29adfba9e0f91b`; Claude Code was unavailable in this environment. It found F-4359-004 and F-4359-005 above and required an explicit per-binding source identity plus an honest no-CLI-on-base limitation before red work. It independently confirmed the existing safe boundary: CLI preflight occurs before `ResolveConnectorCredential`, `ResolveConnectorCredential` returns `missing --credential`, and reverse-ETL repeats preflight before plan construction.

The auditor counted ten literal `/project/` bindings. The task's eleven-command set additionally contains `insights workflow summary list`, whose source lock uses the same documented `{project-slug}` identity at `/insights/{project-slug}/workflows`; it remains in scope but must be individually enumerated in the closed proof. The initial finding set is frozen with that clarification. Production code remains unmodified; commit this record before red work.

## Final independent re-review

Fresh-context Codex independently re-reviewed the complete immutable range
`b9b2478b3b2451d632d28b9aa138a170ad835110..880c3c452274b227d91450aa5680188087f95a0e`
after the final code commit. Verdict: **PASS — no blocker, critical, high,
medium, or low code findings.** It verified the exact source URL/digest,
ordered `vcs_type/org/repo` identity, all eleven source rows, six-lane refusal
matrix, and explicit runtime inventory classification. Claude Code remained
unavailable, so this is the recorded independent-Codex substitution.

The re-review confirmed the one deliberate limitation: this foundation branch
does not contain Batch 1's generated CircleCI `cli_surface.json` or retained
source lock. The built binary consequently reports `unknown command "circleci"`
for all eleven proposed commands here. The required fresh-project,
credential-free `missing --credential` proofs are therefore an explicit Batch 1
integration gate after this foundation is applied, never a reason to fabricate
or import connector command declarations into this branch.

## CI boundary correction pending final re-review

CI run `33039407885` correctly found that the first implementation duplicated
provider policy in shared engine Go. The corrective slice removes every
CircleCI URL, digest, placeholder, component-order, and binding-list literal
from production `internal/connectors/engine/**`. The declaration now carries
its owning connector and ordered row number; the provider-neutral validator
checks only the closed source-cited record shape, while the equivalence
boundary permits the one exact inverse constructed from that record. This
preserves the frozen six-lane matrix and adds no execution capability. A
fresh-context reviewer must assess the full range including this correction.

## Fresh exact-head review after CI correction

Fresh-context Codex reviewed
`b9b2478b3b2451d632d28b9aa138a170ad835110..56808f8d2732f9d545982af4a1934a1e9b8dba5d`.
It found no production-code blocker and confirmed that Batch-1's retained
source lock exactly matches all eleven declaration rows, current source-backed
stream/write templates match the declaration inverse, all six lanes remain
closed, and preflight still precedes credential resolution and provider I/O.

It found one low-severity PR-body overstatement: the provider-neutral engine
validates a closed declaration shape, rather than hard-coding the declaration's
specific URL/digest/placeholder values. This is resolved in `PR-BODY.md` by
describing the retained source lock as the exact CircleCI witness and the
runtime as the closed row/inverse validator. This evidence-only correction did
not change the reviewed code SHA.
