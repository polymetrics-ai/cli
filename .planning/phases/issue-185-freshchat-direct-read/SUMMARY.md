# Summary — Issue #185 Freshchat direct read

Status: implemented locally; full gates pass.

## Notes

- #181/#182/#183/#184 are merged into the parent branch and this branch starts from that parent state.
- This slice only implements a typed Freshchat `POST /users/fetch` read body command.
- No live Freshchat credentials, writes, or upload surfaces are used.
- Added a narrow `freshchat_users_fetch` direct-read output policy for bounded `POST /users/fetch` ids-array bodies.
- Added commandrunner and engine focused coverage for body-mapped direct reads.
- Updated Freshchat CLI/API metadata plus Freshchat docs/website/generated data for `user fetch`.
- Updated conformance `surface_complete` to honor explicit direct-read output-policy method allowlists instead of hard-coding GET.
