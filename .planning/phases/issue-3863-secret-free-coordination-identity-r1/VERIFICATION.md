# VERIFICATION — issue #3863 secret-free credential coordination identity

Status: planned; no production edit has occurred.

## Checklist

- [ ] A linked compatible pair shares an opaque auth cohort without vault participation in identity derivation.
- [ ] A rate key requires explicit policy/kind/non-secret subject; same binding with different subject differs.
- [ ] Unlinked credentials remain isolated; mismatched profiles cannot link.
- [ ] No raw binding, secret, revision, or raw rate subject appears in identity JSON, errors, CLI JSON, logs, state outside protected credential metadata, or tests.
- [ ] Legacy credential metadata migrates safely and rotation remains approval-only.
- [ ] Runtime config carries opaque identity only; #3754/#3865/#3867 seams are present but unimplemented.
- [ ] CLI runtime help/manual/website/generated docs and bare namespace behavior match the actual commands.
- [ ] Targeted tests, package suites, formatting/static checks, repository gates, GSD verify, and code review are recorded.

## Intentionally excluded

- Rate registry/admission/shared coordinator (#3754), requester policy (#3752/#3753), auth fence
  (#3865), persisted parking (#3867), transport dispatch (#3864), live provider checks, and
  reverse-ETL execution.
