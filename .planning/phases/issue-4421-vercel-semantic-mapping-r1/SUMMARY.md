# Summary — Issue #4421 Vercel semantic mapping repair

The Vercel seven-lane matrix now preserves source semantics that were previously lost to a verb-only classifier:

- `HEAD artifactExists` is a source-GET-equivalent direct read.
- `POST artifactQuery` and `POST readSessionFile` are bounded source-declared direct reads, not writes or reverse-ETL mutations.
- `POST readSessionFile` is a source-cited octet-stream binary download.
- `POST writeSessionFiles` retains its documented gzipped-tarball `application/gzip` fact and is a source-cited binary upload.

The correction changes only source-mapped evidence, not runtime behavior. Counter reconciliation is: direct reads `+3`, direct writes `-2`, reverse ETL `-2`, binary downloads `+1`, and binary uploads `+1`; ETL and sync transport are unchanged.
