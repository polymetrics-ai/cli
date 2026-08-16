package certify

import "testing"

func TestDeclaredWriteActionsTreatsAbsentWritesFileAsEmptyInventory(t *testing.T) {
	actions, err := declaredWriteActions("postgres")
	if err != nil {
		t.Fatalf("declaredWriteActions(postgres) error = %v", err)
	}
	if len(actions) != 0 {
		t.Fatalf("declaredWriteActions(postgres) = %#v, want no direct writes", actions)
	}
}
