# Verification — #4093

## Checklist

- [ ] Targeted `synctransport` and connector test packages pass.
- [ ] Race coverage passes for the changed registries/loaders.
- [ ] Docker-backed PostgreSQL native integration test passes using the explicit
  Colima Unix endpoint from the task brief.
- [ ] `go vet` and `go build ./cmd/pm` pass.
- [ ] Required non-full-suite repository gates pass individually.
- [ ] `connectorgen validate` and `surface-sync --check` pass.
- [ ] Inline `verify-work` maps every acceptance criterion to an observable
  passing test or live output.
- [ ] Inline code review has no unresolved actionable findings.
- [ ] Branch is rebased on `origin/integration/4015-mvp-flat-r1` immediately
  before the final push; PR base is API-verified.
