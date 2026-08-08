package conformance

import "testing"

func TestGitHubExhaustiveProviderDouble(t *testing.T) {
	report, err := runGitHubExhaustiveProviderDouble(t)
	if err != nil {
		t.Fatal(err)
	}
	if report.Streams != 37 || report.WriteActions != 574 || report.Operations != 377 {
		t.Fatalf("provider-double totals = streams=%d writes=%d operations=%d, want 37/574/377", report.Streams, report.WriteActions, report.Operations)
	}
	if report.GenericStreams != 23 || report.GenericWrites != 38 {
		t.Fatalf("generic routes = streams=%d writes=%d, want 23/38", report.GenericStreams, report.GenericWrites)
	}
	if report.Failed != 0 {
		t.Fatalf("provider-double report has %d failed rows: %v", report.Failed, report.Failures)
	}
}
