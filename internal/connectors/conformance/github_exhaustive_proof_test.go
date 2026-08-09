package conformance

import (
	"strings"
	"testing"
)

func TestGitHubExhaustiveProviderDouble(t *testing.T) {
	report, err := runGitHubExhaustiveProviderDouble(t)
	if err != nil {
		t.Fatal(err)
	}
	wantGraphQLRoots := githubSourceLockedGraphQLRootCount(t)
	if report.Streams != 37 || report.WriteActions != 574 || report.Operations != 377+wantGraphQLRoots {
		t.Fatalf("provider-double totals = streams=%d writes=%d operations=%d, want 37/574/%d", report.Streams, report.WriteActions, report.Operations, 377+wantGraphQLRoots)
	}
	if report.GenericStreams != 23 || report.GenericWrites != 38 {
		t.Fatalf("generic routes = streams=%d writes=%d, want 23/38", report.GenericStreams, report.GenericWrites)
	}
	if report.Failed != 0 {
		t.Fatalf("provider-double report has %d failed rows: %v", report.Failed, report.Failures)
	}
	for _, row := range report.Rows {
		if !strings.HasPrefix(row.Name, "github.graphql.") {
			continue
		}
		want := "exercised"
		if row.Name == "github.graphql.mutation.delete-issue" {
			want = "blocked"
		}
		if row.State != want {
			t.Errorf("fixed GraphQL operation %q state = %q, want %q (%s)", row.Name, row.State, want, row.Reason)
		}
	}
}
