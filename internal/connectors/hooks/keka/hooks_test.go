package keka

import (
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestMapRecordStampsLegacyNullFields(t *testing.T) {
	h := New()
	raw := connsdk.Record{
		"id":             "attendance_1",
		"attendanceDate": "2026-01-01",
	}
	projected := connsdk.Record{
		"id":             "attendance_1",
		"attendanceDate": "2026-01-01",
	}

	got, keep, err := h.MapRecord("attendance", raw, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("MapRecord keep=false, want true")
	}
	if _, ok := got["employeeId"]; !ok {
		t.Fatalf("employeeId key absent, want explicit nil to match legacy: %+v", got)
	}
	if got["employeeId"] != nil {
		t.Fatalf("employeeId = %#v, want nil", got["employeeId"])
	}
	if got["attendanceDate"] != "2026-01-01" {
		t.Fatalf("attendanceDate = %#v, want raw value", got["attendanceDate"])
	}
}

func TestMapRecordLeavesExpandedStreamsUntouched(t *testing.T) {
	h := New()
	projected := connsdk.Record{"id": "employee_1"}

	got, keep, err := h.MapRecord("employee", connsdk.Record{"id": "employee_1"}, projected)
	if err != nil {
		t.Fatalf("MapRecord: %v", err)
	}
	if !keep {
		t.Fatal("MapRecord keep=false, want true")
	}
	if len(got) != 1 || got["id"] != "employee_1" {
		t.Fatalf("got %+v, want projected record unchanged", got)
	}
}
