# #4089 — bounded stdin approval carrier summary

`pm reverse run` and provider-style connector writes now require the bare
`--approval-token-stdin` marker. The approval token travels as the existing
bounded one-line stdin input; it is no longer accepted from argv.

- Red: a real binary ignored the valid stdin token before CLI wiring existed.
- Green: the same binary test passed with live process argv evidence, six
  independent token-surface assertions, invalid-input no-write assertions, and
  end-to-end replay rejection.
- Parity: CLI help/manual/transcripts, generated skills and connector manuals,
  website source/generated data/blog copy, smoke coverage, and connector
  surface definitions now use the stdin marker.
