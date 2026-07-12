# Verification: Shepherd Supervision

Status: planning complete; RED and implementation pending.

```bash
bash -n scripts/pi-shepherd-loop.sh scripts/tests/pi-shepherd-supervision.sh
shellcheck --severity=warning scripts/pi-shepherd-loop.sh scripts/tests/pi-shepherd-supervision.sh
bash scripts/tests/pi-shepherd-supervision.sh
bash scripts/tests/auto-loop-control.sh
make agent-loop-test
make verify
git diff --check fix/323-auto-loop-hardening...HEAD
```

Tests use temporary repositories, fake Pi processes, synthetic identifiers, and exact process
groups. They must never inspect credentials, invoke a model/provider, or signal unrelated PIDs.

