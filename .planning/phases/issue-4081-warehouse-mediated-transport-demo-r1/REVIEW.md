# Code review — #4081 warehouse-mediated Transport demo

**Mode:** inline/manual `code-review` fallback. The official prompt was
resolved with `scripts/gsd prompt code-review
issue-4081-warehouse-mediated-transport-demo-r1 --auto`; the repository's
canonical single-worker contract forbids spawning the reviewer role.

## Scope and method

Reviewed the #4081 range from the admitted base through `6220144db`, with
special attention to command ownership, token handling, persisted approval
state, plan/binding enforcement, legacy ETL dispatch, declarative GitHub
read-back, and help/docs parity. Re-read the relevant App/Transport paths and
the executable faithful-server test; checked the staged diff, schema validation,
goldens, full affected packages, race approval suite, vet, and surface sync.

## Findings and disposition

| Severity | Finding | Disposition | Evidence |
| --- | --- | --- | --- |
| warning | A bare `pm etl transport github-issue-label cleanup` could reach `App.Open` instead of rendering the closed namespace manual. | Fixed before `6220144db`. | `TestETLTransportBareAndLeafHelpAreContextual` executes it with an uninitialized root and receives contextual help at exit 0. |
| warning | A full 4096-byte token plus newline was incorrectly rejected by the initial bounded reader. | Fixed before `6220144db`. | `TestETLRunTransportApprovalReadsExactlyOneEphemeralStdinLine` accepts exact 4 KiB LF and CRLF inputs and rejects overlong/multiline/trailing-byte input. |
| warning | Passing engine preview warnings and `ApprovalTarget` through the carrier could disclose resolved destination base URL/target details in JSON or human output. | Fixed before `6220144db`; carrier outputs only staged count, action, digest, and issuance fact. | `TestETLTransportSafeOutputOmitsApprovalAndDestinationInternals` rejects token, grant, seal, credential/config, binding, target path, base URL, and `approval_target` leakage. |
| warning | The initial strict carrier admitted the unrelated legacy `--runtime` extension even though the accepted route permits only the plan selector, stdin marker, and destructive confirmation. | Fixed before `6220144db`; legacy runtime remains ordinary ETL behavior. | `TestETLRunTransportApprovalRejectsUnsafeOrIncompleteCarriage/runtime_is_not_part_of_closed_carrier` and `TestETLRunTransportApprovalLeavesLegacyRuntimeAlone`. |
| info | The engine’s declared GitHub `issues` page is `per_page=100`; batch-size 1 bounds the emitted/staged singleton rather than silently changing wire pagination. | Verified, not changed. | Faithful test requires `GET:source:100`, rejects any other page size, emits exactly one source/reopened record, and documents the distinction. |

No unresolved Critical, Warning, or Info findings remain.

## Security and scope confirmation

- No generic HTTP/SQL/shell writer, arbitrary connector pair, provider URL,
  source/target issue, label, action, record, or credential CLI input was
  added. The connection/App derives all provider fields.
- The raw one-time token is read only from a single bounded stdin line after
  the exact closed tuple is validated. It is not accepted in argv, environment,
  files, JSON, state, runtime records, logs, or reader error text. The faithful
  test scans argv/environment/output/project artifacts for leakage.
- `App.RunETL` rejects an approval envelope off the exact persisted Transport
  route before ordinary source read. The existing legacy GitHub five-mode JSON
  suite remains green.
- The schema addition only projects GitHub issue label names necessary for an
  independent declarative read-back; GitHub validation and `surface-sync --check`
  pass. No connector-wide framework or certification claim was added.
- Deferred: PostgreSQL legs, schedules/flows/CDC, auth/rate coordination,
  exhaustive mode certification, TOCTOU/hash syntax/collision hardening, and
  final release promotion.
