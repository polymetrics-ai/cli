# Docker Hub parity corrective slice — code review

## Method

Standard inline review after resolving scripts/gsd sources code-review and reading the generated GSD prompt. The canonical parent-worker contract forbids role spawning in this worktree, so this is the documented manual fallback.

## Reviewed scope

- internal/safety/safety.go and safety_test.go
- internal/connectors/engine/direct_read.go and direct_read_test.go
- internal/connectors/defs/dockerhub/scim_schema_urn_test.go
- Docker Hub reconciliation evidence under .planning/phases/dockerhub-parity-sweep-r1

## Review evidence

- Read the full corrective diff from the red SCIM test through the final reconciliation commit.
- Re-ran focused safety, engine, Docker Hub definition, fleet command-preflight, race, generator, lint, docs, smoke, release, vet, build, and CLI regression gates.
- Mechanically checked the final verification ledger against cli_surface.json: all 54 implemented commands occur once, with no missing, duplicate, or extra row.
- Ran the rebuilt binary through all 54 Docker Hub help routes, bare namespace help, and the SCIM schema command help.

## Findings

No Critical, Warning, or Info findings.

The fix correctly separates opaque provider path values from local connector and credential identifiers. It allows only the documented colon, plus, and at-sign cases before url.PathEscape while retaining rejection of separators, traversal, raw percent escapes, whitespace, controls, dangerous Unicode, query/fragment delimiters, and generic identifier loosening.

The Docker Hub regression test uses an isolated local server with the bundle's auth hook removed only in test memory. It proves the canonical SCIM URN reaches the declared path without a credential or live Docker Hub request. The email regression independently proves the shared direct-read route transports a documented opaque email segment.

The live ledger makes no certification claim. Explicit HTTP 403s remain provider-permission dependencies, documented SCIM remains Enterprise-only, and the three intentionally non-dispatched writes name their exact provider dependency.

## Verdict

PASS — no source, security, test-quality, or evidence-integrity issue found in the corrective slice. The PR must remain explicitly **not certified** because live provider acceptance is account-tier and permission dependent.

## Automated-review corrective round (2026-08-08)

The review above covered the SCIM-URN/opaque-path slice only. A subsequent
automated review of the full branch raised nine findings; all were dispositioned
under binding captain decisions and are recorded in TDD-LEDGER.md under
"Corrective slice — automated-review dispositions".

| Finding | Disposition |
| --- | --- |
| `CLAUDE.md` replaced by a symlink to `AGENTS.md` | Fixed — restored to its prior regular-file content (out of scope for this phase). |
| `AGENTS.md` HEAD-capability addition | Fixed — reverted; the capability stays documented in the connector's own `docs.md`. |
| SCIM token gated behind `docker_pat` | Fixed — second `when: "{{ secrets.scim_bearer_token }}"` auth spec plus a SCIM-only `dualAuth` that fails closed on non-SCIM paths. |
| SCIM prefix routing fails open under a proxy `base_url` | Fixed — prefixes derived from the resolved base path, both the base-relative write form and the unstripped declared direct-read form. |
| `namespace` required with no add-time enforcement | Fixed — declarative required-key admission at the credential boundary; no silent fallback to `docker_username`. |
| Status-only HEAD check cannot report absence | Fixed — 404 returns `{"status_code": 404}` for HEAD only; 401/403/429/5xx and every non-HEAD direct read unchanged. |
| Auth commands persist plaintext credentials | Fixed — the three exchanges are withdrawn from the executable surface and marked blocked on a named dependency. |
| TDD ledger claims a dropped `connsdk` ALPN fix | Fixed — historical entries preserved; a correction states the fix was dropped in the rebase and is unreachable against current `main`. |
| Stale generated connector catalog | Fixed — regenerated to a temp directory and byte-compared; only Docker Hub records updated, the stale `warehouse` description deliberately untouched, no generator or validator broadened. |

## Verdict (corrective round)

PASS with corrections applied. The PR remains explicitly **not certified**: live
provider acceptance is account-tier and permission dependent, and the corrected
bucket split is 11 PROVEN / 0 PROVIDER-PLAN-LIMIT / 31 PROVIDER-PERMISSION /
9 ENTERPRISE-ONLY over 51 implemented operations, plus 3 NOT-IMPLEMENTED.
