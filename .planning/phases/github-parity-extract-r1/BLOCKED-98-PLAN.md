# PLAN — github's 98 blocked endpoints (captain's complete-parity order)

Governing rule from the captain: **exclude ONLY genuine duplicates and genuine missing
foundations.** A named dependency is no longer a deferral — if a foundation is missing, extend it
in this PR.

## The key question, answered

**Does `covered_by.writes` (60719bfbe) already express one endpoint as N write actions?
YES — for request-body arms too. No extension needed for that group.**

Three independent proofs:

1. **60719bfbe's own commit message** names exactly this case: "a bundle may model several distinct
   write contracts over that one path — github's `update_issue`, `close_issue` and `reopen_issue`
   all PATCH the same endpoint **with different bodies**."
2. **`engine/record_schema_promotion.go:69`** — `ValidatePromotableRecordSchema` rejects a union
   root with the literal instruction *"declare a separate named write action for each arm"*, and
   `expandRecordSchemaArms` already merges the wrapper base into each arm and recurses, so the
   concrete per-arm schema is computed for us.
3. **Existence proof already in this PR** — `PATCH /repos/{owner}/{repo}` carries
   `covered_by.writes: ["repo2","archive_repo","unarchive_repo"]`, three actions with different
   bodies on one endpoint, `connectorgen validate` at 0 findings.

### But the 12 are not homogeneous — splitting all of them would be a modelling error

| group | count | correct model |
| --- | --- | --- |
| genuine alternatives (disjoint required sets) | 8 endpoints | one write action per arm → **19 actions** |
| `anyOf` used as "at least one of" | 2 endpoints | **one** action each, base schema flattened |
| root-type polymorphism (object \| array \| string) | 2 endpoints | **one** action each, object arm |

The `anyOf` pair (`secret-scanning/custom-patterns/{pattern_id}`, org + repo) carries **all six
properties on the base**, `required: [custom_pattern_version]`, and five arms that each add one
required field. That is the JSON-Schema idiom for *at least one of*. Splitting it into five named
actions would manufacture five commands each claiming to be the only way to set one field, when one
call may set several. One action; the provider enforces the at-least-one constraint.

`POST` and `DELETE /user/emails` are `object | array-of-string | bare string`. A `record_schema`
here is an object contract (`buildJSONBody` builds a map from record fields), so a bare array or
string body is not expressible as a record at all. The object arm is the implementable superset.

## The 19 genuine alternative actions

| endpoint | actions |
| --- | --- |
| `POST /orgs/{org}/attestations/delete-request` | `by_subject_digests`, `by_attestation_ids` |
| `POST /users/{username}/attestations/delete-request` | `by_subject_digests`, `by_attestation_ids` |
| `POST /orgs/{org}/campaigns` | `code_scanning`, `secret_scanning` |
| `POST /orgs/{org}/projectsV2/{n}/fields` | `existing_issue_field`, `new_field`, `single_select`, `iteration` |
| `POST /users/{username}/projectsV2/{n}/fields` | `new_field`, `single_select`, `iteration` |
| `POST /orgs/{org}/projectsV2/{n}/items` | `by_id`, `by_repo_number` |
| `POST /users/{username}/projectsV2/{n}/items` | `by_id`, `by_repo_number` |
| `POST /user/codespaces` | `from_repository`, `from_pull_request` |

Every arm has a distinct `required` set — verified against the OpenAPI description, which was
fetched at 12,920,264 bytes, byte-exact to the `artifact_bytes` recorded in
`DERIVED-OPERATIONS.json`, so these are the same documents the parity enumeration was derived from.

## Two genuine foundation gaps — EXTEND in this PR, own commits

**F1 — status-only read policy.** The 9 boolean checks are `GET`s answering `204`/`404` with no
body. `validateDirectReadOutputPolicy` (`engine/direct_read.go:409`) accepts only the four
JSON-shaped policies; every read path calls `decodeDirectReadBody`. Note `direct_write` *already*
has `directWritePolicyNone` returning a nil body, so the gap is specifically the read side.

**F2 — text read policy.** `POST /markdown` and `/markdown/raw` return `text/html`; `GET /zen`
returns `text/plain`; `GET /octocat` returns `application/octocat-stream`. No read or write policy
accepts a non-JSON body today.

Both are small, both unblock endpoints for every connector, and the PR body must name them as
foundation changes and say which connectors they unblock.

## No foundation gap — implement directly

- **4 OAuth token-body endpoints** (`applications/{client_id}/grant|token`).
  `directWritePolicyJSONRedacted` / `WriteResultRedacted` already exist and the methods are
  POST/PATCH/DELETE, so `direct_write` covers them.
  **Precondition:** all four take `access_token` as a *required request-body field*, and
  `check-token`/`reset-token` return token material in the 200 body. Implementing them safely
  depends on the `RedactFields` fix the captain ordered — without it a live OAuth token would be
  written to the project state file. Each needs `redact_fields: ["access_token"]`.
- **`force-cancel`** (POST, no body) and **`rerun-failed-jobs`** (POST) — plain write actions. The
  "narrow escalation of an existing write" call was a judgement the captain overrode.

## Stays blocked — genuinely excluded

- **`POST /app/installations/{installation_id}/access_tokens`** — mints a credential and is already
  consumed internally by the `github_app` AuthHook. Same class as the held `auth token`; exposing
  it as a user-invoked command would print a token.

## Still to put to the captain — the 67 "duplicate" rows

He reserved "genuine duplicate" as a judgement. The audit:

| group | count | assessment |
| --- | --- | --- |
| single-detail GET covered by a list stream | 33 | **not** clearly duplicate — 6 already have `partial` CLI commands saying "single X lookup is planned for direct reads" (`issue view`, `pr view`, `release view`, `workflow view`, `run view`, `ruleset view`), so these are acknowledged gaps, and implementing them would promote those 6 from `partial` to `implemented` |
| `agents/*` preview mirror of `actions/secrets`/`variables` | 12 | genuinely the same resources under a preview path prefix |
| "other" | 17 | mixed; needs per-row review |
| field already on an existing stream record | 4 | genuinely duplicate data |
| narrow escalation (`force-cancel`) | 1 | already reclassified above — implement |
| deprecated (`GET /repos/{owner}/{repo}/events`) | 1 | neither duplicate nor missing foundation; GitHub recommends against it, but it is documented |
