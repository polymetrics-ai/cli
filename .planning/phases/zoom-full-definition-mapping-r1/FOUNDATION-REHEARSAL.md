# Foundation rehearsal — Zoom optional write query

Date: 2026-08-22

## Bound foundation revision

The rehearsal used the detached, exact remote Foundation revision
`c3f83cbf6eabbae00219566fb02719ca2d6c480d`. It was created in the isolated
temporary worktree `.zoom-foundation-rehearsal-c3f83`; it was not merged,
rebased, reset into, or pushed from the preserved Zoom branch. The revision is
not an ancestor of `fm/cli-zoom-full-definition-mapping-r1`.

## Behavioral proof

Command:

```text
go test -count=1 -timeout 20m ./internal/connectors/engine \
  -run '^TestZoomMeetingDeleteOptionalQueryRehearsal$'
```

Result: PASS (`ok polymetrics.ai/internal/connectors/engine`). The temporary
fixture exercised the exact Zoom `zoom_meetings_meetingdelete` declaration:
`DELETE /v2/meetings/{{ record.meetingId }}` with its three declaration-owned
object-form query entries. With `meetingId` and `schedule_for_reminder=true`
present, and `occurrence_id` plus `cancel_meeting_reminder` absent, the
safety-approved fixture request reached a loopback server once, retained
`schedule_for_reminder=true`, and omitted both absent fields. This proves the
`omit_when_absent` behavior required by the preserved Zoom declaration at the
named Foundation SHA.

The test used only synthetic fixture approval data and a loopback HTTP server.
It did not read a credential, contact Zoom, mutate provider data, or promote a
certification cell.

## Boundary and remaining gate

The preserved Zoom branch still reproduces its pre-Foundation conformance RED
for this declaration, because its shared engine has not received the above
revision. This rehearsal is not a substitute for rerunning the connector
fixture, built-binary, and live-proof gates after the Foundation lands on
`main`. Final merge readiness remains false until then.
