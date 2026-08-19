# Classification report — github's blocked commands

The captain's order (`CAPTAIN-ORDER-unblock-commands.md`) asked for a table of exactly which
commands changed classification, what confirmation each now requires, and why the held items
are held. This is that table, plus the reasoning the PR body summarises.

## The fact that reframes the whole question

The original brief asked whether `repo create` is safe enough to unblock, on the argument that
creating is safer than destroying. That argument turned out not to be the one that matters.

**Every capability behind these names is already reachable after the parity extraction, under a
different name.** The parity enumeration generated executable commands for the same endpoints:

| blocked gh name | already-implemented twin | endpoint |
| --- | --- | --- |
| `repo create` | `repos create-for-authenticated-user` | `POST /user/repos` |
| `repo delete` | `repo delete-2` | `DELETE /repos/{owner}/{repo}` |
| `repo archive` / `repo unarchive` | `repo update` | `PATCH /repos/{owner}/{repo}` |
| `cache delete` | `actions caches delete-2` | `DELETE /repos/{owner}/{repo}/actions/caches/{cache_id}` |
| `secret set` | `secret set-2` | `PUT /repos/{owner}/{repo}/actions/secrets/{secret_name}` |
| `secret delete` | `secret delete-2` | `DELETE /repos/{owner}/{repo}/actions/secrets/{secret_name}` |

So `unsafe_or_disallowed` on the familiar name was not a safety control. It removed no
capability. It only ensured that the destructive path is the one an operator reaches by
accident, through a generated name, rather than the one they reach on purpose.

That is the argument for restoring them, and it is a stronger one than "create is safer than
delete". It is also why `repo delete` is restored here even though the original brief said to
keep it disallowed: keeping the name blocked while `repo delete-2` runs the same DELETE is not
a safety property, it is a naming accident. The captain's later order asks for it explicitly.

## Restored — 7 commands

All 7 become `intent: reverse_etl`, `availability: implemented`, pointed at the write action
their twin already uses. They inherit the reverse-ETL contract unchanged: **plan, preview,
approval, execute**. No new gate was invented and no existing gate was relaxed.

| command | write action | method | mutation class | confirmation now required | source of the challenge |
| --- | --- | --- | --- | --- | --- |
| `repo create` | `repos_create_for_authenticated_user` | POST | create | plan → preview → approval token → execute | none — approval-only create |
| `repo delete` | `repo` | DELETE | delete | the above **plus** closed typed `--confirm destructive`, and the preview digest recomputed at execution | DELETE method **and** declared `confirm: destructive` |
| `repo archive` | `archive_repo` (new) | PATCH | update | the above **plus** closed typed `--confirm destructive` | declared `confirm: destructive` |
| `repo unarchive` | `unarchive_repo` (new) | PATCH | update | the above **plus** closed typed `--confirm destructive` | declared `confirm: destructive` |
| `cache delete` | `actions_caches_cache_id` | DELETE | delete | the above **plus** closed typed `--confirm destructive` | DELETE method |
| `secret set` | `actions_secrets_secret_name3` | PUT | update | plan → preview → approval token → execute; `encrypted_value` in `redact_fields` | none — approval-only create |
| `secret delete` | `actions_secrets_secret_name` | DELETE | delete | the above **plus** closed typed `--confirm destructive` | DELETE method |

The typed confirmation on the three DELETEs is not something this change adds and could
therefore drop. `connectors.ConfirmationForWriteAction` returns `destructive` for any DELETE
regardless of metadata, and `TestGitHubRestoredCommandsAreExecutable` asserts through that
same resolver, so a future edit that severs it fails the test. Those three needed no new
declaration and did not get one.

`repo create` and `secret set` are classified approval-only. The same test asserts they carry
**no** typed challenge, so drifting one onto them fails just as loudly as dropping one from a
delete.

### Two of these needed more than a classification change

**`repo archive` / `repo unarchive`.** Both ride `PATCH /repos/{owner}/{repo}`, the same
endpoint as the generic `repo update`, so the body is the only thing separating them. Wiring
them both to `repo2` would produce a `repo unarchive` that unarchives only when the caller
separately remembered to send `archived: false` — a command that reports success while doing
nothing. The `covered_by.writes` fix earlier in this branch is what allows three write actions
to share one endpoint row.

They were first pinned in the github hook's `ExecuteWrite`, exactly as `close_issue` and
`reopen_issue` pin `state`. The captain's later order to give them the typed destructive
confirmation is incompatible with that shape, and the engine says so itself:
`prepareDeclarativeWrite` refuses to prepare a destructive action whose hook overrides
execution, because the preview an operator approves would not be the request that runs.
Declaring `confirm: destructive` on the hook-executed actions would have made both commands
fail at preview rather than gate at run — a worse outcome than leaving them ungated.

So the pin moved from the request to the record. `engine.WriteRecordHook` is a new optional
interface that lets a hook fix the body fields an action's own name implies, before the
declarative body is built. `prepareDeclarativeWrite` and `executeApprovedWrite` both apply it,
so the approved body is the dispatched body by construction, and the exactness guard is
satisfied rather than bypassed. `archive_repo`/`unarchive_repo` also declare
`body_fields: ["archived"]`, so the pinned field remains the only field sent — the tightness
the hook-executed version had, now with an exact preview. The interface is opt-in; a connector
that does not implement it sees the records it staged, unchanged.

**`secret set`.** Wiring it surfaced a real defect in the already-`implemented` `secret set-2`:
the write action's `record_schema` declared only the `secret_name` path parameter, so the
command could only ever send an empty PUT body and take a 422. `encrypted_value` and `key_id`
are now declared and flagged on both commands. `pm` does not encrypt and does not store the
value — the caller seals it with the repository public key, which is what the GitHub API
requires — and `encrypted_value` is in `redact_fields` so it stays out of previews and stored
plan samples.

## Not restored — 3 commands, each needs a runtime capability this PR does not add

Marking any of these `implemented` would be the claim-before-establish defect the captain
warned about. Each would fail at runtime, or has no endpoint to fail against.

| command | why it stays blocked |
| --- | --- |
| `issue delete` | Its operation `github.issue.delete` is fully declared — but as `kind: graphql_mutation`. `engine.operationDirectWriteSpec` requires `rest_write`, so the executor refuses it. GitHub exposes no REST issue-delete endpoint; `deleteIssue` is GraphQL-only and takes a node ID, not an issue number. Restoring it means a GraphQL write executor plus node-ID resolution. |
| `issue transfer` | GraphQL-only (`transferIssue`). `POST /repos/{owner}/{repo}/transfer` is **repository** transfer, not issue transfer — wiring it there would be a different operation wearing this command's name. No documented REST endpoint exists. |
| `pr revert` | No documented REST endpoint at all — it is absent from github's own `api_surface.json` and from the 1220-row documented enumeration. `AGENTS.md` is explicit: never invent an `api_surface` endpoint to make a command look implemented. |

**bahmni `documents upload`** is in the same category and is also left blocked. It has no write
action at all, and its own notes name the missing capability: "current inline JSON content
surface lacks the claimed file snapshot/SHA-256 approval binding". Unblocking it means building
that binding, not changing a string.

## Held for the captain's decision — 2 commands

**`auth token` — held, as ordered.** It prints a credential. The standing rule is never print,
summarize, or store a secret value, and runtime output is never stripped, so a printed token
stays printed. Unblocking it contradicts a rule the captain set, so it waits for him to say so
himself. `TestGitHubHeldCommandsStayBlocked` pins it.

**`api` — reported, not acted on.** Unblocking the arbitrary authenticated-request escape hatch
would mean, concretely:

- **no declared parameters** — the request shape is whatever the caller types;
- **no enum validation** — none of the value checking every other command gets;
- **no required-field enforcement** — nothing establishes a call is well-formed before it goes out;
- **no operation ledger entry** — the endpoint is absent from `operation_endpoint_ledger.json`,
  so it has no `kind` and no `max_bytes`, and nothing bounds the response;
- **no `api_surface` row** — it bypasses the 1220-endpoint surface this PR exists to enumerate,
  and with it every gate keyed on that surface, including the destructive-write confirmation.

It is the same shape already rejected for a generic `--param` passthrough. One `pm github api`
call can reach every endpoint the other 1146 commands model, under none of their controls —
including `DELETE /repos/{owner}/{repo}` with no typed confirmation. The captain's call.
