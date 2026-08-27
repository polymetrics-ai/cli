# Discussion log — Stripe provider-dialect tolerance foundation

The launch brief resolves the material choices:

1. A global limit increase is insufficient. Reference lookup must be normalized and memoized per source document, while the reference count, descriptor byte, schema byte, and finite-depth bounds remain enforced.
2. A depth traversal that cannot yield an exact operation contract must not discard unrelated operations. For a source lock that permits source-contract gaps, retain a skeletal, source-cited operation descriptor marked merge-blocked with a concrete `source_descriptor` missing-foundation gap at that operation's source location.
3. Only the bounded depth condition is retained operation-locally. External references, malformed pointers, ambiguous `$ref` siblings, non-schema cycles, target-kind errors, operation count/byte/reference-count exhaustion, and all other unsafe failures stay hard rejections.
4. The original Stripe artifact is historical retained evidence, not authority to restore a connector source tree. Reduced hermetic source fixtures retain exact Stripe IDs, paths, methods, operation IDs, source locations, and source URL without provider I/O.
5. No command becomes runnable. This PR produces no Stripe action, CLI binding, executor, credential check, write/delete flow, or generic escape hatch.

GSD command prompts were generated and executed inline because this runtime cannot provide compatible isolated GSD roles and the canonical delivery contract prohibits role spawning.
