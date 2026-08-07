# DISCUSSION LOG — issues #3745 and #3746

Mode: `scripts/gsd prompt discuss-phase issue-3745-3746-changefeed-discovery-r1 --auto`.

The parent and issue tree resolve all material product choices. The selected defaults are:

- dedicated optional `changefeed.json`, not a metadata boolean;
- no fleet mapping before the provider-artifact survey;
- PostgreSQL described as unsupported, with an evidence source and rationale;
- catalog fail-closed on descriptor/executor mismatch;
- inspect explains the non-capability in JSON; broader docs/manual/website work is deferred to
  #3748;
- no new dependency, credentialed check, live provider call, generic protocol surface, or
  PostgreSQL CDC implementation.

There are no open human decisions in this slice.
