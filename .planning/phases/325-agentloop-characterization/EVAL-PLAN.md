# Evaluation Plan: Phase 0

Phase 0 is deterministic and requires no model-quality evaluation. Its evaluation set is the
thirteen fixture matrix plus malformed-input and driver-fuse negatives. Pass means every expected
semantic violation is derived independently, every malformed fixture is rejected with a stable
class, and both drivers deny before side effects. Any disagreement is a test failure; there is no
score threshold or compensating dimension.
