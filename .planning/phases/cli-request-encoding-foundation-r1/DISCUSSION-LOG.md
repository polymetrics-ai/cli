# Discussion log — source-backed request encoding foundation

## Fixed decisions from #4367

- Cohort is exactly 51 source operations: 50 multipart and 1 URL-encoded.
  The rederived provider split is GitLab 47, Sentry 2, Asana 1, Jira 1.
- The selected encoder is source-owned. Callers cannot choose content type,
  HTTP method/path, body map, or generic raw payload.
- Form-field and multipart part names, binary classification, documented
  part encoding/content-type metadata, requiredness, and bounds remain
  declaration-owned.
- This change does not absorb the media-selection, composition, dynamic-map,
  generic HTTP, source-descriptor, or deferred-visibility foundations.
- At launch `main` lacks #4364's manifest and deferred-command projection.
  The code tests derive the expected 51 rows from retained-source-shaped
  fixtures instead of copying the in-flight manifest. Consumer command
  promotion is measured separately and may be zero in this foundation PR.
- No live provider access or credential may be used. A loopback spy is the
  provider-I/O boundary for executor tests.

