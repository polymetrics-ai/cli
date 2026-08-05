# Verification — issue #3792 provider-search runtime preflight

Status: planned; no production change has been made.

## Completion checklist

- [ ] New commandrunner RED test observed against the original generic-reader preflight.
- [ ] Loaded engine operation preflight rejects unsupported kind, mismatched method/path/policy,
  absent operation, and non-positive cap before any request.
- [ ] Existing bounded provider-search `httptest` coverage stays green.
- [ ] Real implemented-command preflight sweep passes.
- [ ] Focused/package tests, formatting, vet, build, and applicable individual repository gates pass.
- [ ] No declarations, schemas, validators, capabilities, redaction paths/policies, CLI/help/docs,
  #3740 overlap paths, or credentials/live calls changed.
- [ ] GSD verify/code review executed inline and recorded; no-mistakes deferred to firstmate's
  post-commit instruction.
