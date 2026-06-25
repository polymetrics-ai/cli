# ADR: Native Go Catalog Metadata Before Runtime Enablement

Decision: import every validated connector as metadata immediately, but only mark entries enabled when native Go code exists and passes conformance tests.

Reason: this gives agents complete discovery and documentation without introducing an opaque container execution bridge or unrestricted external mutation surface.
