# Hubplanner official inventory summary

Authoritative sources: provider-owned https://github.com/hubplanner/API `Sections/*.md` at tree SHA `91217d34486e43fa590e9f9e3e477aee20da157a` from issue #3239. No live Hubplanner API calls or credentials were used.

| total | implemented | fixture-tested | blocked/planned | excluded/not-applicable | certified |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 107 | 97 | 81 | 10 | 0 | 0 |

| covered stream rows | covered write rows | covered direct-read rows | blocked webhook event rows |
| ---: | ---: | ---: | ---: |
| 20 | 61 | 16 | 10 |

## Blocked rows
- `WEBHOOK event:project.update`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:resource.update`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:booking.create`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:timeEntry.create`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:timeEntry.update`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:timeEntry.create.update`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:timeEntry.delete`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:booking.update`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:booking.delete`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
- `WEBHOOK event:booking.delete.multiple`: Hubplanner documents this as an outbound webhook event delivered to an external service. The current connector contract cannot receive provider callbacks as CDC; implementation depends on the shared CDC/changefeed runtime tracked by #2986 and #2988.
