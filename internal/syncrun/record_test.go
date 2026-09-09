package syncrun

import (
	"encoding/json"
	"testing"
)

func TestRecordAcceptsVersionedAdditiveTransitions(t *testing.T) {
	record := Record{Version: RecordVersion, RunID: "run-test-1", Transitions: []Transition{
		{Sequence: 1, To: PhasePlanned},
		{Sequence: 2, From: PhasePlanned, To: PhasePreviewed},
		{Sequence: 3, From: PhasePreviewed, To: PhaseApproved},
		{Sequence: 4, From: PhaseApproved, To: PhaseStaging},
		{Sequence: 5, From: PhaseStaging, To: PhaseStaged},
		{Sequence: 6, From: PhaseStaged, To: PhaseApplying},
		{Sequence: 7, From: PhaseApplying, To: PhaseApplied},
		{Sequence: 8, From: PhaseApplied, To: PhaseReadbackVerified},
		{Sequence: 9, From: PhaseReadbackVerified, To: PhaseAcknowledged},
		{Sequence: 10, From: PhaseAcknowledged, To: PhaseCheckpointed},
		{Sequence: 11, From: PhaseCheckpointed, To: PhaseCompleted},
	}}
	if err := record.Validate(); err != nil {
		t.Fatalf("valid record: %v", err)
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Record
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRecordRejectsUnknownOrOutOfOrderTransitions(t *testing.T) {
	cases := []Record{
		{Version: RecordVersion, RunID: "run-test-1", Transitions: []Transition{{Sequence: 1, To: Phase("unknown")}}},
		{Version: RecordVersion, RunID: "run-test-1", Transitions: []Transition{{Sequence: 2, To: PhasePlanned}}},
		{Version: RecordVersion, RunID: "run-test-1", Transitions: []Transition{{Sequence: 1, To: PhasePlanned}, {Sequence: 2, From: PhasePlanned, To: PhaseCompleted}}},
	}
	for _, record := range cases {
		if err := record.Validate(); err == nil {
			t.Fatalf("invalid record accepted: %#v", record)
		}
	}
}
