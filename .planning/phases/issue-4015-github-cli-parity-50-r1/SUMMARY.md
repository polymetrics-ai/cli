# Summary — issue #4015 GitHub declared-command parity

## Outcome

All 50 supplied declarations have an explicit verdict: **25 implemented + 25 retained = 50**.
No declaration was removed. Implemented commands pass the real commandrunner preflight and name a
fixed API surface; retained commands name the exact provider, composite, local, or safety boundary.

## Per-command verdicts

| # | Command | Verdict | Provider/runtime evidence |
| ---: | --- | --- | --- |
| 1 | `issue pin` | implemented | Fixed GraphQL `pinIssue`; live issue state changed to pinned. |
| 2 | `issue unpin` | implemented | Fixed GraphQL `unpinIssue`; live issue state returned to unpinned. |
| 3 | `pr diff` | implemented | Fixed REST pull endpoint with declaration-owned `application/vnd.github.diff`; live bounded download was 217 bytes and began `diff --git`. |
| 4 | `pr ready` | implemented | Fixed GraphQL `markPullRequestReadyForReview`; live draft state changed true → false. |
| 5 | `repo list` | implemented | Fixed paged GraphQL `RepositoryOwner.repositories`; live result identified the private fixture repository. |
| 6 | `repo autolink create` | implemented | Fixed REST create action; live prefixed autolink was independently listed. |
| 7 | `repo autolink delete` | implemented | Fixed REST delete action with opaque string ID; independent view returned 404. |
| 8 | `repo license list` | implemented | Fixed `GET /licenses`; live result was a non-empty array. |
| 9 | `repo gitignore list` | implemented | Fixed `GET /gitignore/templates`; live result contained `Go`. |
| 10 | `workflow enable` | implemented | Fixed REST enable action; live workflow state returned to `active`. |
| 11 | `workflow disable` | implemented | Fixed REST disable action; live workflow state became `disabled_manually`. |
| 12 | `cache list` | implemented | Fixed repository Actions-cache read; live result returned a bounded collection. |
| 13 | `label clone` | retained | GitHub provides label list/create primitives, not an atomic clone; a paginated multi-write composite with partial-failure policy is absent. |
| 14 | `secret list` | implemented | Fixed repository secret-metadata read; live collection was returned with values redacted. |
| 15 | `variable list` | implemented | Fixed repository variables read; live collection shape was returned. |
| 16 | `variable get` | implemented | Fixed variable-name read; live read observed `PM_CERT_PARITY_50`. |
| 17 | `variable set` | retained | GitHub separates create (`POST`) and update (`PATCH`); no conditional composite executor exists. Typed create/update commands remain available. |
| 18 | `variable delete` | implemented | Fixed REST delete action; independent get returned 404. |
| 19 | `org list` | implemented | Fixed authenticated-user organizations read; live result contained `Polymetrics-Cert`. |
| 20 | `gist list` | implemented | Fixed authenticated-user gist read; live result returned an array. |
| 21 | `gist create` | retained | `POST /gists` exists, but the filename-keyed body is user-scoped and executing it would violate the authorized organization/repository-only fixture boundary. |
| 22 | `codespace list` | implemented | Fixed authenticated-user codespace read; live result reported zero codespaces. |
| 23 | `codespace create` | implemented | Fixed repository-ID write reached GitHub; the authorized repository refused creation and an independent list proved no fixture was created. |
| 24 | `gpg-key list` | implemented | Fixed authenticated-user GPG-key read; live result returned an array. |
| 25 | `ssh-key list` | implemented | Fixed authenticated-user SSH-key read; live result returned an array. |
| 26 | `skill list` | retained | `gh skill list` scans local skill directories; GitHub has no provider collection and the connector has no local-filesystem workflow executor. |
| 27 | `agent-task list` | implemented | Fixed agent-task collection read; live result returned the declared collection shape. |
| 28 | `issue develop` | retained | This is a git branch/worktree composite; no typed local-git executor exists. |
| 29 | `pr checkout` | retained | This mutates a local git worktree; no typed local-git executor exists. |
| 30 | `repo clone` | retained | This invokes local git and writes a worktree; the declared `local_git` metadata has no runtime executor. |
| 31 | `repo sync` | retained | This performs local fetch/reset/branch operations; no typed local-git executor exists. |
| 32 | `repo set-default` | retained | This writes gh-local repository config, not provider state; no connector config executor exists. |
| 33 | `release upload` | retained | GitHub documents a separate upload host and raw binary body; the runtime has no bounded binary-upload executor and generic HTTP writes are prohibited. |
| 34 | `release download` | implemented | Fixed single-asset bounded downloader. The fixture had no assets; a provider 404 left no destination file. |
| 35 | `release verify` | retained | Verification is local cryptographic work over downloaded assets; no signature-verification executor exists. |
| 36 | `run download` | implemented | Fixed single-artifact bounded ZIP downloader with extraction disabled. The fixture had no artifacts; a provider 404 left no destination file. |
| 37 | `run watch` | retained | This requires cancellable polling and terminal rendering; no watch executor exists. |
| 38 | `codespace ssh` | retained | This opens an interactive SSH/process session; no SSH or subprocess executor exists. |
| 39 | `auth login` | retained | This is interactive credential bootstrap; credentials are managed through pm's sealed credential flow. |
| 40 | `auth status` | retained | This reads gh's local credential store, outside the connector credential abstraction. |
| 41 | `auth token` | retained | This prints credential material, which the repository contract prohibits. |
| 42 | `config get` | retained | This reads gh-local configuration; no connector config executor exists. |
| 43 | `config set` | retained | This writes gh-local configuration; no connector config executor exists. |
| 44 | `browse` | retained | This launches a browser; no browser-launch executor exists. |
| 45 | `alias list` | retained | This reads gh-local aliases; no gh-config/local-filesystem executor exists. |
| 46 | `extension list` | retained | This reads locally installed extensions; no local-extension executor exists. |
| 47 | `completion` | retained | This generates shell completion for gh; connector commands use the hand parser and expose no completion contract. |
| 48 | `api` | retained | This is caller-selected generic HTTP/GraphQL passthrough; connector canon explicitly prohibits generic HTTP writes and requires fixed typed operations. |
| 49 | `attestation verify` | retained | GitHub's attestation endpoint only returns bundles; signature, identity, digest, and policy verification remain local cryptographic work. |
| 50 | `copilot` | retained | This launches an external interactive extension; no extension/subprocess executor exists. |

## Live certification and cleanup

The provider credential was read inline from the macOS Keychain into pm's sealed temporary project
credential and was never placed in argv, logs, fixtures, planning artifacts, or branch files.

- Safe reads exercised repository, license, gitignore, cache, secret, variable, organization, gist,
  codespace, GPG-key, SSH-key, and agent-task surfaces.
- Disposable variable, autolink, issue, workflow file, and temporary PR branch were created only in
  `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`, independently observed, and independently 404 after
  cleanup.
- The existing certification PR used for `pr ready` was restored to its original closed/non-draft
  state.
- The workflow fixture was restored to absence after enable/disable state assertions.
- Download destination directories were removed after success or provider failure; failed downloads
  left no partial files.
- GitHub refused codespace creation for the authorized fixture; the independent codespace list was
  empty, so there was nothing to delete.
- Gist creation was not executed because a gist is user-scoped and the launch brief authorizes writes
  only inside the named organization/repository fixture.

## Implementation notes

- The binary downloader gained one fixed, declaration-owned `Accept` value. Callers still cannot
  supply headers, and cross-host credential stripping, byte caps, destination confinement, and the
  archive-extraction refusal remain unchanged.
- Live certification exposed provider IDs being persisted as floating-point numbers and rendered in
  scientific notation. Autolink and workflow identifiers are now opaque strings, and a regression
  test protects the contract.
- Variable create/update schemas now require only their real provider body fields and reject
  undeclared top-level fields.
