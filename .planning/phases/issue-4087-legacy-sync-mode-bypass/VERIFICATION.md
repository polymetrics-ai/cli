# Verification: issue #4087

Status: planned

## Acceptance checklist

- [ ] Both aliases resolve to non-empty canonical typed contracts through normal and persisted-legacy parsing.
- [ ] Both aliases return typed execution or a typed pre-I/O refusal, with no legacy source read.
- [ ] The mapping is single-sourced and connector-neutral.
- [ ] All closed canonical mode names preserve their existing parsed contract/admission behavior.
- [ ] Runtime help and generated surface checks show no stale or divergent compatibility surface.
- [ ] Focused tests, formatting, vet, build, and individual repository gates pass.
- [ ] Inline/manual GSD verify-work and code-review evidence is complete.
