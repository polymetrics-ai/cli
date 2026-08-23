package app

import "testing"

// TestFoundationRollupPreservesMultiActionReverseETLComposition keeps the
// production-shaped persisted App path in the rollup's own TDD ledger. The
// exact reverse-ETL foundation supplies the executable fixture; until that
// component is composed this test is intentionally red because the closed
// multi-action destination surface does not exist in this integration tree.
func TestFoundationRollupPreservesMultiActionReverseETLComposition(t *testing.T) {
	TestPersistedConnectionSelectsDeclarativeTypedDestinationAction(t)
}
