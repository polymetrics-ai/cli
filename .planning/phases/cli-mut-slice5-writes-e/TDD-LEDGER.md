# TDD Ledger — GitHub mutation certification slice 5 writes-e

## Manual GSD fallback

The repository-local Pi adapter prompts were generated and reviewed. This isolated terminal runtime cannot run Pi subagents, so the lifecycle is completed inline.

## Cycle 1 — certification evidence representability

**Red:** Validate the integration base before writing. The designated base adds schema-v2 `observed_operations` records with salted fingerprints and command-specific `function_kind` values; the pre-rebase checkout did not contain those records.

**Green criterion:** Persist a record only when the command has a real provider proof and validator acceptance. Non-passes are recorded as classified attempt outcomes, not fabricated `status: passed` evidence.

## Cycle 2 — live mutation containment

**Red criterion:** A successful CLI exit alone is insufficient because issue #4221 demonstrates delete success can leave the object present.

**Green criterion:** A passed command must include an independent provider read-back of its state change and a direct-provider deletion followed by a 404 or empty collection.

## Attempt ledger (redacted)

| Ordinal | Command | Result | Evidence |
| ---: | --- | --- | --- |
| 1 | `codespaces add-repository-for-secret-for-authenticated-user` | `no_object` | Plan and preview completed. Run reached GitHub and returned 404 for a disposable, deliberately absent secret/repository pair. Independent `GET /user/codespaces/secrets/<tag>` returned 404. |
| 3 | `codespaces create` | `entitlement` | Plan and preview completed. Run reached GitHub and returned HTTP 400. A raw GitHub provider control using the smallest tier (`basicLinux32gb`) returned the same 400 and created no Codespace; no cleanup was required. |
| 53 | `issues reactions create-2` | `product_defect` | PM POSTed a bodyless reaction and GitHub returned 422. Raw GitHub POST with the required `content` succeeded (201), independently read back as present, then direct DELETE returned 204 and read back absent. |
| 54 | `issues reactions delete-2` | `product_defect` | PM serialized the reaction identifier in exponential notation and GitHub returned 404. Direct DELETE using the provider-issued identifier returned 204 and independent collection read-back proved absence. |
| 105 | `variable delete-2` | `certified` | PM DELETE returned a completed one-record run; independent raw GET returned 404. Published schema-v2 record: `github-manual-slice5-variable-delete-2-rrun-5282e7b8218c.json`. |
| 108 | `variable update` | `product_defect` | PM PATCH had no value payload and GitHub returned 422. Raw GitHub PATCH with the required value returned 204; direct DELETE and independent GET then proved the fixture absent. |

## Contained fixture cleanup

GitHub does not offer deletion for issue resources. The disposable issue used for
the reaction controls was directly PATCHed to `state=closed` after use and an
independent provider read-back confirmed `closed`; it is recorded as
`contained_closed`, never as `verified_absent`. The enclosing private fixture
repository is the disposable container.
