---
phase: cli-current-foundations-main-integration-r1
fixed_at: 2026-08-21T01:45:05Z
review_path: .planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md
iteration: 1
findings_in_scope: 36
fixed: 36
skipped: 0
status: all_fixed
---

# Foundation Canonical Code Review Fix Report

**Fixed at:** 2026-08-21T01:45:05Z
**Source review:** `.planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md`
**Iteration:** 1

## Summary

- Findings in scope: 36 (27 blockers, 9 warnings)
- Fixed: 36
- Skipped: 0
- Base: `f5b1d15e28d4eedc5018dce7bdd74a28c7eca72e`
- Final code HEAD: `2d656b8b632defbae4f43e08c3f4aaf6fab96942`

All fixes were applied as shared foundation behavior. No connector-specific runtime allowlist or declaration bypass was introduced. Public output preserves ordinary provider values and identifiers while masking only configured credential material at explicit output boundaries.

## Atomic Fix Groups

| Commit | Canonical findings | Applied fix |
|---|---|---|
| `af3e65959` | CF-B04, CF-W01, CF-W07 | Made GraphQL and generated agent/website surfaces declaration-derived and deterministic. |
| `933afaa25` | CF-B08, CF-B09, CF-B10, CF-B12, CF-B16, CF-B18, CF-B23, CF-B24, CF-B26 | Preserved terminal receipts and attempted identity, masked only configured credentials, corrected empty-body semantics, enforced ownership CAS before side effects, and propagated publication outputs. |
| `c9e38798c` | CF-B11 | Added complete shared REST, GraphQL, and binary-download response receipts on success and provider failure. |
| `51aa23665` | CF-B13, CF-B14, CF-B15, CF-B17, CF-B19, CF-B20, CF-B21, CF-B22, CF-W02, CF-W03 | Sealed request construction across approval, redirects, replay, exact numeric validation, duplicate flags, byte bounds, REST body/query declarations, and secret transforms. |
| `bdf54cff7` | CF-B01, CF-B02, CF-B03, CF-B05, CF-W08 | Connected source lock/import through descriptors, validation, generation, uploads, and canonical evidence surfaces. |
| `3fe44681d` | CF-B06, CF-B07, CF-B25, CF-B27, CF-W04, CF-W05, CF-W06, CF-W09 | Enforced authorization/CAS/readback/delivery ordering and durable, serialized rate-parking lifecycle semantics. |
| `a9c22c43a` | Closure for CF-B01, CF-B03, CF-B05, CF-B11, CF-B21, CF-B22, CF-W01, CF-W07, CF-W08 | Derived hyphenated path inputs from executable templates, completed error-bearing read/download envelopes, exercised every declared GitHub route with bounded provider fixtures, and regenerated every dependent evidence surface. |
| `2d656b8b6` | Closure for CF-W01 | Tracked the previously repository-excluded Bahmni generated skill so a clean checkout exactly matches registry-derived generation. |

## Fixed Issues

| Finding | Title | Commit(s) | Resolution |
|---|---|---|---|
| CF-B01 | Authoritative source import is nonfunctional and silently ignores GraphQL | `bdf54cff7`, `a9c22c43a` | Source import now retains REST and GraphQL declarations and drives regenerated surfaces. |
| CF-B02 | Imported descriptors are orphaned from generation and validation | `bdf54cff7` | Imported descriptors are consumed by validation and generation with exact regression coverage. |
| CF-B03 | Reverse-ETL actions omit provider inputs and advertise invalid requests | `bdf54cff7`, `a9c22c43a` | Required request inputs, including hyphenated template fields, are derived from executable declarations and exposed to CLI schemas. |
| CF-B04 | Generated GraphQL operations expose placeholder results and incomplete pagination | `af3e65959` | GraphQL result selections, variables, and pagination metadata are source-derived and executable. |
| CF-B05 | Installed GitHub binary upload is a JSON request to the wrong origin | `bdf54cff7`, `a9c22c43a` | Binary upload remains mapped to its declaration-owned upload origin and multipart/binary executor. |
| CF-B06 | Reusable typed-destination authorization can fail after write and checkpoint | `3fe44681d` | Authorization identity and CAS are checked before provider I/O and checkpoint advancement. |
| CF-B07 | Generic destination read-back never reads the provider | `3fe44681d` | Generic destinations use declared provider-owned readback with identity, bounds, and checkpoint ordering. |
| CF-B08 | Public provider receipts disclose configured credentials | `933afaa25` | Explicit public receipt boundaries mask exact configured secret values while retaining ordinary provider output. |
| CF-B09 | CLI discards persisted failed runs and their receipts | `933afaa25` | Failed runs and terminal receipts are persisted and rendered with nonzero status. |
| CF-B10 | Direct-write no-response failures lose attempted operation identity | `933afaa25` | Attempt identity is sealed before I/O and retained when no response arrives. |
| CF-B11 | Direct-read and binary-download receipts omit provider truth and error responses | `c9e38798c`, `a9c22c43a` | Complete REST/GraphQL/binary receipts and safe typed errors are emitted together on failures; binary bytes remain confined to files. |
| CF-B12 | Accepted-success-status mismatch discards terminal direct-write receipt | `933afaa25` | Undeclared accepted-status responses retain bounded terminal metadata and typed mismatch errors. |
| CF-B13 | Idempotency-enabled reverse writes can redirect outside approval and lose final failure receipts | `51aa23665` | Redirect targets and final retry receipts remain inside the approved, sealed execution identity. |
| CF-B14 | Buffered declarative HTTP follows cross-origin redirects with custom credentials | `51aa23665` | Cross-origin redirects cannot forward custom credential material. |
| CF-B15 | Reverse-write previews print resolved credentials and sensitive identifiers | `51aa23665` | Approval previews use a safe public projection without mutating the sealed execution request. |
| CF-B16 | REST and binary error formatting bypasses safe HTTP diagnostics | `933afaa25` | REST and binary failures use bounded, credential-safe diagnostics and complete receipts. |
| CF-B17 | Declarative writes rematerialize records after approval | `51aa23665` | Approved prepared requests are executed without record rematerialization. |
| CF-B18 | Write receipt parsers fabricate body presence for empty JSON | `933afaa25` | Absent, empty, null, and present response bodies retain distinct receipt semantics. |
| CF-B19 | Numeric enum validation collapses integers above 2^53 | `51aa23665` | Numeric enum validation uses exact integer representation. |
| CF-B20 | Singleton typed CLI flags silently use the last occurrence | `51aa23665` | Duplicate singleton flags are rejected before request construction. |
| CF-B21 | Caller-controlled path and query values have no byte bound | `51aa23665`, `a9c22c43a` | Path/query components and generated provider fixtures are bounded before network I/O. |
| CF-B22 | REST direct reads remain an open query/body escape hatch | `51aa23665`, `a9c22c43a` | Direct-read query/body inputs must be declaration-owned and schema-valid. |
| CF-B23 | Name-based JSON redaction deletes ordinary provider identifiers | `933afaa25` | Heuristic key-name deletion was removed; occurrence IDs and other ordinary provider IDs remain intact. |
| CF-B24 | Provider side effects occur before stale-writer ownership CAS | `933afaa25` | Ownership CAS precedes provider side effects. |
| CF-B25 | Post-publication readback failure invokes pre-publication abort | `3fe44681d` | Committed publication remains committed when later readback fails. |
| CF-B26 | Full-overwrite and Arrow publication outputs never reach results | `933afaa25` | Full-overwrite and Arrow publication outputs are propagated through standard and pipeline results. |
| CF-B27 | Declared delivery guarantees are never enforced | `3fe44681d` | Mode, strategy, delivery guarantees, and tombstone compatibility are enforced before I/O. |
| CF-W01 | Generated connector skills are hard-coded to five connector names | `af3e65959`, `a9c22c43a`, `2d656b8b6` | Skills are registry-derived, regenerated for changed declarations, and complete in a clean checkout. |
| CF-W02 | Path interpolation validates before filters, not after final path context | `51aa23665` | Validation applies to the final filtered/interpolated path value. |
| CF-W03 | Accepted GitHub secret transform has no executor | `51aa23665` | Accepted secret transforms have matching execution support. |
| CF-W04 | Transient rate-parking failures remove the only retry timer | `3fe44681d` | Retry state and timer ownership survive transient resume failures. |
| CF-W05 | Same provider-scope parked runs resume concurrently | `3fe44681d` | Same-scope resumes are serialized while independent scopes remain available. |
| CF-W06 | Route selection hides actionable transport preflight errors | `3fe44681d` | Route selection returns typed reasons and preserves actionable preflight errors. |
| CF-W07 | Generated website data is deterministic but semantically stale | `af3e65959`, `a9c22c43a` | Website data is regenerated from current declarations and checked by script tests. |
| CF-W08 | Evidence status is inconsistent and overstates coverage | `bdf54cff7`, `a9c22c43a` | Canonical evidence and generated certification artifacts now agree with executable coverage. |
| CF-W09 | Rate-parking store APIs can persist unloadable state | `3fe44681d` | Parking writes validate reloadable state before persistence. |

## Verification

- Literal combined `go test ./internal/app ./internal/cli ./internal/connectors/... ./internal/coordination ./internal/synctransport ./cmd/connectorgen -count=1 -timeout 30m`: PASS, wall `19:33.60`; App `383.829s`, CLI `1155.435s`, boundary `571.443s`, commandrunner `37.420s`, conformance `71.124s`, engine `18.509s`, coordination `10.072s`, synctransport `5.583s`, connectorgen `176.140s`, with all other connector packages green or correctly reporting no test files.
- Full `internal/cli`: PASS, package `1052.232s`, wall `17:37.92`.
- Full `internal/app`: PASS, package `316.310s`, wall `5:27.66`.
- Full `cmd/connectorgen`: PASS, package `161.921s`, wall `2:46.92`.
- Focused commandrunner/conformance/native Postgres/synctransport: PASS (`27.090s`, `40.392s`, `7.614s`, `4.581s`).
- Generated skills/docs: PASS, package `389.322s`, wall `6:38.43`.
- `connectorgen validate`: PASS, 552 connectors, 0 findings.
- `surface-sync --check`: PASS, 552 scanned, 0 changes.
- GitHub certification sweep: PASS, 1,584 rows / 1,580 CLI commands current.
- Website script tests: PASS, 33/33.
- Focused `go vet`, modified JSON parse, and `git diff --check`: PASS.
- Website TypeScript/unit commands were not runnable because `tsc` and `vitest` are not installed in this isolated worktree (exit 127); the dependency-free website generated-data suite passed.

## Skipped Issues

None.

---

_Fixed: 2026-08-21T01:45:05Z_
_Fixer: the agent (gsd-code-fixer)_
_Iteration: 1_
