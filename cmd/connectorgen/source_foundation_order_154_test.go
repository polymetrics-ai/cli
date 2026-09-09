package main

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// Every ordering uses the same four original captured records, including both
// new selected engine proofs. Reordering cannot alter their meaning or bytes.
func TestSourceFoundationFourRecordOrders154(t *testing.T) {
	repo, document := sourceFoundationProofFixture153(t, true)
	if len(document.Records) != 4 {
		t.Fatal("four original proof records required")
	}
	original := append([]sourceFoundationProofRecord{}, document.Records...)
	want := map[string]sourceFoundationProofRecord{}
	for _, r := range original {
		want[r.ID] = r
	}
	count := 0
	for a := range 4 {
		for b := range 4 {
			for c := range 4 {
				for d := range 4 {
					if a == b || a == c || a == d || b == c || b == d || c == d {
						continue
					}
					count++
					t.Run(fmt.Sprintf("%d%d%d%d", a, b, c, d), func(t *testing.T) {
						document.Records = []sourceFoundationProofRecord{original[a], original[b], original[c], original[d]}
						writeSourceFoundationObservationFixture(t, repo, document)
						got, err := readSourceFoundationProofObservations(t.Context(), repo)
						if err != nil {
							t.Fatal(err)
						}
						if len(got) != 4 {
							t.Fatalf("records=%d", len(got))
						}
						seen := map[string]bool{}
						last := ""
						for _, o := range got {
							if seen[o.record.ID] || o.record.ID <= last || o.status != "current" || !reflect.DeepEqual(o.record, want[o.record.ID]) {
								t.Fatalf("original record or deterministic current observation changed: %s status=%s", o.record.ID, o.status)
							}
							seen[o.record.ID] = true
							last = o.record.ID
						}
					})
				}
			}
		}
	}
	if count != 24 {
		t.Fatalf("permutations=%d", count)
	}
}

// The same actual capture can be a 64 MiB capture role and a 4 MiB fixture
// role. The smaller bound must govern both orders before metadata is decoded.
func TestSourceFoundationCaptureFixtureCap154(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse_%t", reverse), func(t *testing.T) {
			repo, document := sourceFoundationProofObservationFixture(t)
			writeSourceFoundationObservationFixture(t, repo, document)
			if _, err := readSourceFoundationProofObservations(t.Context(), repo); err != nil {
				t.Fatal(err)
			}
			capture := document.Records[0].Capture
			if capture.Bytes <= 4<<20 || capture.Bytes > 64<<20 {
				t.Fatal("original capture does not cross fixture cap")
			}
			record := &document.Records[1]
			record.Inputs = append(record.Inputs, sourceFoundationProofInput{Path: capture.Path, SHA256: capture.SHA256, Bytes: capture.Bytes, Role: "fixture"})
			sort.Slice(record.Inputs, func(i, j int) bool { return record.Inputs[i].Path < record.Inputs[j].Path })
			if reverse {
				document.Records[0], document.Records[1] = document.Records[1], document.Records[0]
			}
			writeSourceFoundationObservationFixture(t, repo, document)
			events := []sourceProofReadEvent{}
			got, err := readSourceFoundationProofsObserved(t.Context(), repo, reviewedSourceFoundationProofs(), func(e sourceProofReadEvent) {
				if e.Path == capture.Path {
					events = append(events, e)
				}
			})
			if err == nil || len(got) != 0 {
				t.Fatal("capture role bypassed smaller fixture bound")
			}
			if len(events) != 1 || events[0].Success || events[0].Charged != (4<<20)+1 {
				t.Fatalf("actual bounded failing read not reached: %+v", events)
			}
		})
	}
}
