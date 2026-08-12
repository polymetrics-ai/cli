# Problem — #3856 immutable polling-watermark conformance corpus

Issue #3856 needs a versioned, executable polling-watermark conformance corpus
that proves state and recovery behavior before #3857 admission, #3858 source
execution, or #3859 apply strategies can claim readiness. The existing generic
`internal/synccontract/testdata/conformance/v1.json` defines shared semantics
and its digest must remain unchanged. The merged #3880 polling tests contain
useful algorithms and scripted fakes, but are package-local tests rather than
an immutable, no-skip corpus another lane can execute.

This phase owns only the reusable polling corpus, its defensive loader/evidence
surface, a registered-lane/no-skip runner, and deterministic fake-lane tests.
It does not add a connector descriptor, product source, apply strategy,
provider code, public CDC capability, credentialed test, or live certification.
