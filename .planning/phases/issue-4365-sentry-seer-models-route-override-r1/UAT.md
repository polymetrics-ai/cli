# Issue #4365 automated UAT

Manual GSD fallback: the Pi runtime cannot run the official isolated verifier and
Firstmate explicitly assigned this worker to work without delegation. The coverage
block in `SUMMARY.md` records the automated classification.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| One exact source-bound CLI projection | focused engine and source-projection tests | pass |
| Closed route/base/method/path contract | table-driven preflight rejection tests | pass |
| Slash-stable endpoint | local provider test with both declared base forms | pass |
| Credential boundary | real binary in a fresh project plus zero-I/O transport spy | pass |
| Help and generated documentation | focused help tests, root golden transcripts, docs and website generation | pass |

No human-only visual or external-provider acceptance criterion applies: this slice
does not call Sentry and the tested command halts before transport without a
credential.
