# Automated UAT

Coverage-aware verification classifies D1-D3 in `SUMMARY.md` as automated passes. The installed
firing test executes the rendered command through a freshly built `cmd/pm`, observes destination
read-back and terminal schedule state, and scans durable/rendered surfaces for approval material.
The separate fresh-binary rate test observes `policy/shared_coordinator_unavailable` and zero
transport sends.

The corrective certification proof also auto-passes: the exact fresh-binary CLI test executes
schedule create/list/install/remove without any schedule authority carrier, observes the direct
rooted flow command, and restores the redirected crontab byte-for-byte.

No manual judgment item remains. Credentialed live-provider certification was not possible without
an available credential/runtime and remains explicitly tracked by #4166.
