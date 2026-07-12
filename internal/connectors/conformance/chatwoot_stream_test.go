package conformance

import (
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestChatwootStreamRunnerSweep(t *testing.T) {
	b, err := engine.Load(realDefsFS(), "chatwoot")
	if err != nil {
		t.Fatalf("engine.Load(chatwoot): %v", err)
	}

	if len(b.Streams) != 7 {
		t.Fatalf("Chatwoot streams = %d, want 7", len(b.Streams))
	}

	streamByName := map[string]engine.StreamSpec{}
	for _, s := range b.Streams {
		streamByName[s.Name] = s
		pages, err := loadFixturePages(b.Fixtures, s.Name)
		if err != nil {
			t.Fatalf("load fixtures for stream %q: %v", s.Name, err)
		}
		if len(pages) == 0 {
			t.Fatalf("stream %q has no fixture pages; full read sweep coverage requires every Chatwoot stream to replay", s.Name)
		}
	}

	messages := streamByName["messages"]
	if messages.Name == "" {
		t.Fatal("messages stream missing")
	}
	if messages.Pagination == nil || messages.Pagination.Type != "cursor" || messages.Pagination.CursorParam != "after" || messages.Pagination.LastRecordField != "id" {
		t.Fatalf("messages pagination = %+v, want cursor after/id", messages.Pagination)
	}
	if messages.FanOut == nil || messages.FanOut.IDsFrom.Request == nil {
		t.Fatalf("messages fan_out ids_from.request missing: %+v", messages.FanOut)
	}
	parentPagination := messages.FanOut.IDsFrom.Request.Pagination
	if parentPagination == nil || parentPagination.Type != "page_number" || parentPagination.PageParam != "page" || parentPagination.PageSize != 25 {
		t.Fatalf("messages parent fan_out pagination = %+v, want page_number page size 25", parentPagination)
	}

	schema := b.Schemas["messages"]
	if schema == nil {
		t.Fatal("messages schema missing")
	}
	if got := schema.CursorField; got != "id" {
		t.Fatalf("messages schema cursor = %q, want id for Chatwoot after pagination", got)
	}

	rep := RunBundle(b)
	if !rep.Passed {
		t.Fatalf("Chatwoot conformance failed: %+v", rep.Checks)
	}
	for _, name := range []string{"conversations", "contacts", "inboxes", "agents", "teams", "labels", "messages"} {
		check := assertHasCheck(t, rep, "read_fixture_nonempty:"+name)
		if !check.Passed || check.Skipped {
			t.Fatalf("read fixture sweep for %q = %+v, want passed and not skipped", name, check)
		}
	}
	for _, name := range []string{"pagination_terminates", "records_match_schema"} {
		check := assertHasCheck(t, rep, name)
		if !check.Passed || check.Skipped {
			t.Fatalf("%s = %+v, want passed and not skipped", name, check)
		}
	}
}
