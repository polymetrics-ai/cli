# PR 346 Current-Head Review

## Review scope

- Parent PR #346 against `main` at `ce0c816a`.
- Stacked GitHub-only auth PR #394 at `fa46381e`.
- Auth/session, comments/bookmarks APIs, PostgreSQL migrations, annotations UI,
  sidebars, production image, and deployment workflow.

## Findings

### Warning: bookmark identity crossed block namespaces

`bookmark_unique_anchor` and its conflict lookup omitted `block_type`. A body
paragraph and bullet point use independent block indexes, so matching text and
offsets could return the wrong existing bookmark. Fixed with migration 4 and a
matching lookup predicate. Regression test proved red then green.

### Warning: optimistic deletes did not handle rejected fetches

Comment and bookmark removal restored state for non-2xx responses but not for a
network rejection. This caused an unhandled promise rejection and permanently
hid the item until refresh. All delete paths now catch both failure modes,
restore without overwriting concurrent state, and announce the failure.

### Warning: Better Auth session state could mismatch during hydration

An immediately resolved client session rendered signed-out markup while the
server rendered the loading skeleton. React discarded the right-rail subtree.
All auth-dependent rendered surfaces now use a shared post-mount session hook.

### Info: secret-free image build logged false errors

Next route collection instantiated Better Auth without a runtime secret and
logged default-secret errors even though the image build succeeded. The builder
now has an explicit public placeholder that is not copied to the runner stage;
the runtime still reads the real secret from its environment file.

## Stacked PR disposition

PR #394 is necessary and reviewed clean: it removes nonfunctional Google and
LinkedIn launch paths, keeps GitHub-only OAuth, preserves account linking, and
has green CI. It is included in the local assembled tree.

## Review route

Claude review is unavailable and the recorded Copilot review exhausted quota.
This document records the current-head human/Codex fallback required by the
parent contract. No unresolved Critical or Warning finding remains locally.
