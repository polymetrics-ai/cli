# Issue #121 GitHub Full Certificate Plan

## Scope

Stacked sub-issue PR for #121 under parent PR #49 (`feat/44-github-cli-parity`). The goal is a GitHub full-certificate path covering all catalog streams, direct-read coverage, binary-download handling, per-stream flow/schedule checks, and reverse-ETL write lifecycle coverage.

## GSD Runtime Status

The correctly stacked branch is based on `feat/44-github-cli-parity`, which does not currently include the repo-local `scripts/gsd` adapter from the later GSD rebootstrap work. Per `AGENTS.md`, this phase therefore uses the **manual-GSD fallback** until the parent branch contains the adapter.

Manual-GSD fallback requirements for this phase:

1. Plan before code in this directory.
2. Add red tests before production edits.
3. Record green test evidence in `TDD-LEDGER.md` and `VERIFICATION.md`.
4. Keep PR #125 stacked against `feat/44-github-cli-parity`.
5. Do not run live credentialed checks with pasted secrets; only use rotated credentials supplied through local environment variables.

## Implemented Non-Credentialed Slices

- `--full` flag wiring and write-pairing sweep registration.
- All-catalog-stream read sweep machinery.
- Per-stream flow/schedule checks in full mode.
- Direct-read certification stage for a curated GitHub smoke command (`repo read-file`).
- Binary-download safety gate for a curated GitHub binary command (`release download`) that verifies operation-backed binary commands remain safely blocked until a bounded executor exists.
- GitHub catalog bootstrap default stream (`issues`).

## Current GitHub Surface Inventory

The parent branch GitHub bundle currently declares:

- 509 reviewed API endpoints in `api_surface.json`.
- 37 stream-covered endpoints.
- 231 write-covered endpoints/actions.
- 173 direct-read-covered endpoints represented by 2 implemented direct-read command families (`repo read-file`, `repo read-dir`).
- 69 explicitly blocked endpoints with typed model/status/reason.

The certificate must account for all 509 endpoints. It must not blindly execute blocked/admin/sensitive/destructive/local/binary surfaces.

## Remaining Certificate Work

The full GitHub organization certificate is **not live-complete** until these gates pass:

1. Live GitHub run with rotated token and disposable/dev repository.
2. Verify the actual catalog stream count and every stream's read/query stages.
3. Execute both currently implemented direct-read commands (`repo read-file`, `repo read-dir`) against safe paths.
4. Keep binary-read surfaces blocked unless/until a bounded binary executor and destination policy are implemented; current surface accounting verifies the 10 blocked binary-read API endpoints plus the one typed binary operation.
5. Execute live write lifecycle for the safe curated first pairing, including cleanup verification, with a credentialed dev repo. All 231 declared write actions are now accounted for, but not all are safe standalone create-cleanup lifecycles.
6. Enforce typed confirmation / secret transform / destructive/admin runtime gates where not already covered by metadata validation.

## Live Credential Safety

The token previously pasted into chat is considered compromised and must not be used. Before live testing, rotate/revoke it and provide a new token only through a local environment variable such as `PM_GITHUB_DEV_TOKEN`.
