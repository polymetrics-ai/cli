# Summary — Google Search Console documented-operation parity resume

Status: VERIFIED — ready to commit on current `origin/main`.

Google's current provider-published Discovery description enumerates **11 unique Search Console
operations**. All **11/11 are genuinely reachable** through the connector: five typed ETL read
operations, four typed reverse-ETL write actions, and two typed bounded direct reads. The generated
surface contains **15 honest commands**: the five Search Analytics ETL commands are safe
dimension-specific conveniences for one provider operation.

There are **zero planned or non-reachable documented operations**. The recovered sixteenth command,
`direct search-analytics query`, was removed rather than claimed implemented because current-main
path safety rejects Google's required URL-valued `siteUrl` before dispatch. Its provider operation
remains reachable through the five ETL streams that correctly path-escape the property value.

The field research matrix records **32/32 operation-specific request-field paths** with
provider-owned citations, including nested Search Analytics filter fields and optional fields not
exposed by the bounded CLI. Current `origin/main` has not yet landed the shared machine-readable
citation convention, so this phase deliberately keeps the required matrix instead of inventing a
competing bundle encoding.

The human REST reference index visibly lists ten operations; it no longer links the old
Mobile-Friendly Test page. Google's current Discovery description still publishes that operation,
so the provider-published total for parity is eleven. The matrix records this source-availability
caveat and its limited requiredness inference explicitly.
