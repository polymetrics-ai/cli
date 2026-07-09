# TDD Ledger — Issue #185

## Red target

Add tests for:

- commandrunner maps `--id` on Freshchat `user fetch` to a direct-read request body (`ids` array) for `POST /users/fetch`;
- engine direct read can safely execute the bounded Freshchat users-fetch POST policy.

Expected initial failure: direct-read POST commands are rejected because only GET direct reads and GitHub output policies are supported.

## Green target

Implement the narrow body-mapped Freshchat direct-read policy and mark Freshchat `user fetch` implemented.

## Verification ledger

Pending.
