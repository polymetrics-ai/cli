# Verification: Chatwoot Help Renderer And Docs Parity

## Planned gates

```bash
./pm help docs
./pm connectors inspect chatwoot | grep -E 'COMMAND SURFACE|Usage: pm chatwoot|conversation list|message create'
go test ./internal/connectors/bundleregistry -run ChatwootGuide -count=1
( cd website && pnpm test:unit -- tests/api/connector-data.test.ts )
./pm docs validate --connectors-dir docs/connectors
make verify
git diff --check
```

## Results

Pending.
