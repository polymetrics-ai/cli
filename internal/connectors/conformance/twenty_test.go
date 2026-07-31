package conformance

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestTwentyFixturesCoverAllStreamsAndWrites(t *testing.T) {
	b := loadTwentyBundle(t)
	rep := RunBundle(b)
	checks := map[string]CheckResult{}
	for _, c := range rep.Checks {
		checks[c.Name] = c
	}

	var missing []string
	for _, stream := range b.Streams {
		name := "read_fixture_nonempty:" + stream.Name
		if c, ok := checks[name]; !ok || !c.Passed {
			missing = append(missing, name)
		}
	}
	for _, action := range b.Writes {
		name := "write_request_shape:" + action.Name
		if c, ok := checks[name]; !ok || !c.Passed {
			missing = append(missing, name)
		}
	}
	if len(missing) != 0 {
		t.Fatalf("Twenty fixture coverage missing/failing %d checks: %s", len(missing), strings.Join(missing, ", "))
	}
}

func TestTwentyBatchAndDeleteFixturesAssertExactBodies(t *testing.T) {
	b := loadTwentyBundle(t)

	batch := loadTwentyWriteFixtureRaw(t, b, "batch_companies")
	batchExpect := rawFixtureExpect(t, batch, "batch_companies")
	bodyExact, ok := batchExpect["body_exact"].([]any)
	if !ok || len(bodyExact) != 1 {
		t.Fatalf("batch_companies expect.body_exact = %#v, want one top-level array element", batchExpect["body_exact"])
	}

	deleteFixture := loadTwentyWriteFixtureRaw(t, b, "delete_companies")
	deleteExpect := rawFixtureExpect(t, deleteFixture, "delete_companies")
	noBody, ok := deleteExpect["no_body"].(bool)
	if !ok || !noBody {
		t.Fatalf("delete_companies expect.no_body = %#v, want true", deleteExpect["no_body"])
	}

	checks := checkWriteRequestShape(b)
	wantPass := map[string]bool{
		"write_request_shape:batch_companies":  false,
		"write_request_shape:delete_companies": false,
	}
	for _, check := range checks {
		if _, ok := wantPass[check.Name]; ok {
			wantPass[check.Name] = check.Passed
			if !check.Passed {
				t.Errorf("%s failed: %s", check.Name, check.Error)
			}
		}
	}
	for name, passed := range wantPass {
		if !passed {
			t.Fatalf("%s did not pass", name)
		}
	}
}

func loadTwentyBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(os.DirFS("../defs"), "twenty")
	if err != nil {
		t.Fatalf("Load(twenty): %v", err)
	}
	return b
}

func loadTwentyWriteFixtureRaw(t *testing.T, b engine.Bundle, action string) map[string]any {
	t.Helper()
	raw, err := fs.ReadFile(b.Fixtures, "writes/"+action+".json")
	if err != nil {
		t.Fatalf("read %s fixture: %v", action, err)
	}
	var doc map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode %s fixture: %v", action, err)
	}
	return doc
}

func rawFixtureExpect(t *testing.T, doc map[string]any, action string) map[string]any {
	t.Helper()
	expect, ok := doc["expect"].(map[string]any)
	if !ok {
		t.Fatalf("%s fixture expect = %#v, want object", action, doc["expect"])
	}
	return expect
}
