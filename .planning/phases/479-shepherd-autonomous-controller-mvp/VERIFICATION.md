# Verification

Status: pending.

Required gates:

```bash
node --test .pi/extensions/shepherd/autonomous-controller.test.ts \
  .pi/extensions/shepherd/arguments.test.ts \
  .pi/extensions/shepherd/extension.test.ts
node --test .pi/extensions/shepherd/*.test.ts
git diff --check
```

Also run the repository's strict TypeScript command, offline Pi command discovery, and one local
deterministic autonomous canary. Broad Go/connector gates are parent-only and are not relevant to
this TypeScript Pi-extension slice.
