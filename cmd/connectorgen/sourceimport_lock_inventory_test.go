package main

import "testing"

func TestValidateSourceImportLockInventoryRejectsContradictoryRESTOnlyTotal(t *testing.T) {
	lock := sourceImportLock{
		Counts: sourceImportCounts{REST: 1, Total: 2},
		Rest: sourceImportREST{Operations: []sourceImportRESTOperation{{
			ID:             "alpha.widgets.list",
			Protocol:       "rest",
			Method:         "GET",
			Path:           "/widgets",
			SourceLocation: "paths./widgets.get",
		}}},
	}
	if err := validateSourceImportLockInventory(lock); err == nil {
		t.Fatal("REST-only contradictory total was accepted")
	}

	lock.Counts.Total = 1
	if err := validateSourceImportLockInventory(lock); err != nil {
		t.Fatalf("valid REST-only exact inventory: %v", err)
	}
}
