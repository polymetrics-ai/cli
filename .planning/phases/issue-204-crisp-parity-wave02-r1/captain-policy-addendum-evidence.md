# Captain policy addendum evidence

Command path: `python3 .planning/phases/issue-204-crisp-parity-wave02-r1/append_captain_addendum.py` (uses `gh-axi issue view` and `gh-axi issue edit`).

First run appended the marker to #204-#211:

```text
#204: appended
#205: appended
#206: appended
#207: appended
#208: appended
#209: appended
#210: appended
#211: appended
```

Second run proved idempotency:

```text
#204: already-present
#205: already-present
#206: already-present
#207: already-present
#208: already-present
#209: already-present
#210: already-present
#211: already-present
```

The addendum text explicitly preserves existing r2 body content and count tables; no count table edits were made.
