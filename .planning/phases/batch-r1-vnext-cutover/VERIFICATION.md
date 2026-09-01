# Batch R1 vNext cutover verification

## Invariants

1. The runtime executes only rendered JSON. `source.lock.json` and provider provenance are authoring inputs, never runtime/admission inputs.
2. A native `mode=fixture` value cannot produce a successful check or canned record. It must reach ordinary validation before any provider request.
3. No native/hook compatibility reader can replace a declared engine execution route.
4. No generated or explanatory surface instructs authors or users to use a removed importer, certification, retention, fixture, compatibility, feature-flag, or second executor route.
5. Every migrated connector has exactly the seven declared lanes and reaches its normal credential/approval boundary without provider I/O.

## TDD ledger execution

| Slice | RED command/proof | GREEN command/proof |
| --- | --- | --- |
| Native fixture bypass | `go test -timeout 20m ./internal/connectors/native/alpha-vantage -run '^TestFixtureModeNoLongerBypassesCredentialBoundary$'` fails while fixture mode returns success. | The same test proves `mode=fixture` returns the ordinary missing-credential error without a request; the native residual guard is green. |
| Compatibility execution | The frozen review maps every `StreamHook`/`CheckHook` native delegation that can bypass declarative execution. | Targeted engine/hook/native tests prove only declared engine execution reaches the credential/approval boundary. |
| Source-lock render | A changed lock is absent or render drift is observed before its output is generated. | `go run ./cmd/connectorgen lock-render <connector>` followed by `--check` is byte-stable for each reference and migrated connector. |
| Credential boundary | A selected declared command has no connector-local witness before migration. | An isolated project with no stored credential returns ordinary missing-credential or approval-boundary output before provider I/O. |

## Exact residual scan

The scan is deliberately separated by data class. Test fixture values and provider evidence are not runtime paths; production Go, execution JSON, generated documentation, and website sources must not retain the forbidden behavior.

| Scope | Exact search expression | Required disposition |
| --- | --- | --- |
| Production Go: `cmd/connectorgen`, `internal/connectors/native`, `internal/connectors/hooks`, `internal/connectors/engine`, `internal/connectors/commandrunner`, excluding `_test.go` | `(?i)(fixtureMode|readFixture|mode[[:space:]]*[:=][[:space:]]*["']fixture|legacy[[:space:]_-]*(reader|shim|execution|admission)|compatib[a-z]*[[:space:]_-]*(reader|shim|execution|admission)|source[[:space:]_-]*(import|projection|admission)|certification[[:space:]_-]*(runtime|admission)|retention[[:space:]_-]*(runtime|admission)|feature[[:space:]_-]*flag)` | Zero execution/admission hits. Standard language-level compatibility or value fallback is classified separately and is not an execution route. |
| Execution definitions: `internal/connectors/defs/**/{spec,metadata,streams,writes,operations,cli_surface,sync_transport,rate_limits}.json` | `(?i)("mode"[[:space:]]*:[[:space:]]*"fixture"|fixture[[:space:]_-]*(mode|execution)|certification|retention[[:space:]_-]*(admission|runtime)|source[[:space:]_-]*(import|projection)|legacy[[:space:]_-]*(reader|shim))` | Zero operational hits; any remaining `source.lock.json` is provider provenance and is never embedded. |
| Human/generated surfaces: `docs`, `website`, `.agents`, `internal/cli` manuals, generated skills | `(?i)(mode[[:space:]]*=[[:space:]]*fixture|fixture[[:space:]_-]*(mode|execution)|source[[:space:]_-]*import|certification[[:space:]_-]*(gate|admission|runtime)|retention[[:space:]_-]*(gate|admission|runtime)|legacy[[:space:]_-]*(reader|shim|execution)|compatib[a-z]*[[:space:]_-]*(reader|shim|execution))` | Remove stale instruction. A retained provider-provenance mention is listed by exact path with why it cannot execute or admit a command. |

The final evidence records every nonzero result with path, quoted role, and classification; absence is never inferred from a partial page of search output.

## Required generation and checks

After source changes settle:

```text
go build ./cmd/pm
./pm docs generate --dir docs/cli
./pm skills generate --dir docs/skills --json
node website/scripts/gen-github-cli-surface.mjs
```

Focused cleanup checks:

```text
go test -timeout 20m ./cmd/connectorgen ./internal/connectors/defs ./internal/connectors/native/... ./internal/connectors/hooks/... ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli
go build ./cmd/pm
go run ./cmd/connectorgen lock-render github --check
go run ./cmd/connectorgen lock-render gitlab --check
go run ./cmd/connectorgen lock-render asana --check
./pm docs validate --connectors-dir docs/connectors
```

Broader checks:

```text
go test -timeout 20m ./internal/connectors/engine ./internal/app ./internal/cli
git diff --check
```

The final delivery record also runs the secret/local-state scan for tokens, private keys, `.polymetrics`, vault, and `.env` paths; records available disk space; and records the exact reviewed source SHA, tests, residual result, connector witnesses, and named remaining gaps.

## Results

- G0 direct-parent amendment: `git fetch origin fm/cli-top100-declaration-batch-r1` refreshed the parent. `HEAD` and `origin/fm/cli-top100-declaration-batch-r1` both resolved to `d260b725ce6f53403961d7af1ef48ea6651cdd66`; `HEAD` is an ancestor of that remote tip. The immutable `main` merge base is `813f457a925f7ee3fe3bea101a43e445992c8552`, and #4325 fixes the delivery denominator at 4,341 primary retained source operations.
- G0 routing/local-state: #4325 comment `5500153864` and #4294 comment `5500165004` were read directly. The former certification tree was deleted at the frozen checkpoint. The untracked opaque `internal/connectors/certifications/.fingerprint-salt` has no tracked/index/history entry and is not ignored; it remains unread, unstaged, unmodified, and excluded from this work. It is recorded as local-state residue, never a retained certification path.
- G0 green gate: satisfied by planning commit and ordinary remote update `1655123262586b2eaa395aa75b0e54bd7c4558bd`; remote read-back matched exactly before N1 began.
- #4423 N1 RED: the former named defs proof could not compile because it read removed `Bundle` fields; the two Atlas proof selectors returned `no tests to run`, and explicit `go test -list` count assertions failed. This establishes the inherited handoff as red rather than accepting a zero-test exit status.
- #4423 N1 focused GREEN: `go test -count=1 -timeout 20m` passed for `TestRuntimeEmbedContainsExecutionJSONOnly`, `TestVNextSourceLockDeterministicallyRendersReferenceConnectors`, `TestVNextReferenceConnectorsDiscoverFromExecutionJSONOnly`, `TestVNextReferenceLockClosedSetRejectsUnrenderedArtifact`, and `TestVNextFoundationProofSelectorsResolve`. Each Atlas target was listed exactly once before execution. The full `cmd/connectorgen` and `defs` package suites passed. The closed-set reference test reads GitHub, GitLab, and Asana source locks and verifies their complete in-memory rendered output sets without writes or provider I/O.
- #4423 named remaining gap: `go test -count=1 -timeout 20m ./internal/connectors/engine` fails in unrelated `TestOperationRoutesFailClosedBeforeProviderIO`; N1 changes that package only by renaming the Atlas test selector. The named engine proof passed, but no broader-engine green claim is made. Render/generation, residual scan, secret/local-state scan, and final exact-SHA review remain pending for their later authorized gates.
- #4423 review: fresh-context manual re-review of `1655123262586b2eaa395aa75b0e54bd7c4558bd..eae04af74b6546f9c61130d426f06197723f4f22` found no N1 issue; its exact scope and safety analysis are recorded in `REVIEW-CONVERGENCE.md`. The review applies to the code-bearing N1 commit; the following evidence-only commit does not change production or test code.
