# Plan — issue #4342 binary upload CLI and certification foundation

## GSD command path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, `code-review`: resolved to the pinned official command sources.
- `scripts/gsd prompt discuss-phase issue-4342-binary-upload-surface-foundation-r2`: generated and executed inline as recorded in `CONTEXT.md`.
- `scripts/gsd prompt plan-phase issue-4342-binary-upload-surface-foundation-r2 --tdd`: generated and executed inline as recorded in `CONTEXT.md`.
- `scripts/gsd prompt execute-phase issue-4342-binary-upload-surface-foundation-r2`, `verify-work`, and `code-review`: generated and executed inline. The canonical worker contract forbids role spawning in this direct-PR lane and no compatible isolated Pi runtime was available, so the same worker recorded the red/green evidence, verification, and review below as the documented manual-GSD fallback.

## Implementation slices

1. **Red — closed command admission.** Complete. Commandrunner tests prove raw binary, base64, and multipart required-file actions plan only through their declaration; malformed mappings and JSON actions fail closed.
2. **Green — runtime and public surface.** Complete. The schema, engine preflight, commandrunner, static validator, GitHub declaration, help, manuals, and website surface route the command into the existing approval-bound lifecycle.
3. **Red — truthful certification.** Complete. The stage test drives the actual in-process harness and proves a rejected command records a non-passing `blocked` result. A plan without a live transfer is `not_live`, never `pass`.
4. **Green — candidate, stage, and sweep.** Complete. Download remains backward-compatible at `binary`; upload is independent at `binary_upload`, with separate candidate/sweep/evidence classes. `file_upload` remains declarable but non-executable.
5. **Refactor and generated projections.** Complete. Regenerated the GitHub manual/skills/website, certification candidates/sweep, operation evidence fixed-100 projection, and current certification subject.
6. **Verify/review.** Superseded by the 2026-08-24 independent review finding set below; the original manual source/diff review did not catch the upload-origin, state-redaction, status-contract, media-policy, lifecycle-proof, live-host, guidance, or source-root gaps.

## Review remediation gap plan — 2026-08-24

This is the documented inline/manual `plan-phase --gaps` fallback: the task's direct-PR
single-worker contract does not permit a compatible isolated GSD worker. No finding is
considered closed by a declaration check or a green refusal.

1. **Provider contract (F-4343-01 and F-4343-03).** Write red tests that arm public and
   Enterprise transports separately. An Enterprise API origin must refuse the fixed public
   upload host before either transport receives a request or an authorization header. Also
   assert only GitHub's declared `201` creates a written result; `200`, `202`, and `204`
   retain their provider receipt as failures. Green implementation must bind the upload
   origin to public github.com and thread declared successful statuses into binary writes.
2. **Source confidentiality and lifecycle proof (F-4343-02, F-4343-05, and F-4343-09).** Write a real
   App/CLI plan → preview → approval → execute test against a declaration-bound server.
   It must assert exact bytes, SHA-256, request content type, `201` response, changed-file
   zero-I/O refusal, and no caller local-path sentinel in plan JSON, preview JSON, output,
   or state. Green implementation must redact and require re-supply of `file_path` while
   retaining only identity/digest/size metadata. It must also prove a binary-upload plan emits no
   approval token before preview, the persisted preview issues exactly one human-only bounded token,
   and execution before preview is refused before provider I/O.
3. **Executable media policy (F-4343-04).** Red tests must demonstrate that raw and base64
   public uploads reject unsatisfiable/mismatched media declarations before transport I/O.
   Green validation may promote raw binary only when its allow-list admits the executor's
   fixed `application/octet-stream`; base64 requires an equally explicit enforceable policy.
4. **Truthful certification and live-host proof (F-4343-06).** Keep the generated stage
   `not_live` until the separately authorized run completes. The live record must name only
   non-secret fixture metadata, exact transmitted/read-back digest and size, `201` provider
   receipt, and complete cleanup ledger. No refusal or plan may become `pass`.
5. **Guidance and source-root parity (F-4343-07 and F-4343-08).** Red CLI/App tests must
   prove that a documented project-root file is found and that a path relative to the actual
   containment root is not double-prefixed. The public help must explicitly say one coherent
   contract: supply a path relative to the project root, and the runtime must resolve it there
   while confining it beneath that root. It must not require users to discover or name
   `.polymetrics`. Update the stale legacy unsupported note to describe its actual unsupported
   multi-file/clobber semantics rather than deny the bounded executor used by its sibling. The PR
  body must separately disclose the shipped command's change from optional preview under
  `reverse_etl` to required preview under `binary_upload`.

## Review remediation completion — 2026-08-24

All nine findings were addressed with the limits recorded in `CONTEXT.md` and
`VERIFICATION.md`: binary upload now has its own persisted-preview approval gate;
the generic no-confirmation `reverse_etl` population remains unchanged by the
Firstmate scope ruling. The observed live transfer is recorded in `LIVE-PROOF.md`.
It does not alter the generated stage's non-pass status until that stage can own
the same transfer/read-back/cleanup protocol.

## CLI help/manual/website parity

- [x] `pm github releases assets upload --help` lists the declared upload source flag and approval behavior.
- [x] The same command's help resolves without a credential or project when invoked with `--help`.
- [x] Existing bare connector namespace behavior remains unchanged (`pm github` renders contextual group help at exit 0).
- [x] `docs/connectors/github/{MANUAL,SKILL}.md` and website generated CLI surface use `binary_upload` and its closed safety rule.
- [x] No `docs/cli` generic upload command is added because no generic command exists.
- [x] Generated documentation and website drift checks are run.

## Verification plan

Use `GOFLAGS='-p=3'` and one heavy suite at a time. The targeted package set is `internal/connectors/engine`, `internal/connectors/commandrunner`, `internal/cli`, `internal/connectors/certify`, and `cmd/connectorgen`, plus docs/generator checks. The full test suite and `make verify` are not run as one process in this memory-bound worktree; every omitted local gate will be recorded with its exact reason before PR creation.
