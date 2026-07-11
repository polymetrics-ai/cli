package conformance

import (
	"bytes"
	"encoding/json"
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

	batch, err := loadWriteFixture(b.Fixtures, "batch_companies")
	if err != nil {
		t.Fatalf("load batch_companies fixture: %v", err)
	}
	if len(batch.Expect.BodyExact) == 0 {
		t.Fatalf("batch_companies fixture must assert body_exact top-level array")
	}
	var exact any
	dec := json.NewDecoder(bytes.NewReader(batch.Expect.BodyExact))
	dec.UseNumber()
	if err := dec.Decode(&exact); err != nil {
		t.Fatalf("decode batch_companies body_exact: %v", err)
	}
	arr, ok := exact.([]any)
	if !ok || len(arr) != 1 {
		t.Fatalf("batch_companies body_exact = %#v, want one top-level array element", exact)
	}

	deleteFixture, err := loadWriteFixture(b.Fixtures, "delete_companies")
	if err != nil {
		t.Fatalf("load delete_companies fixture: %v", err)
	}
	if !deleteFixture.Expect.NoBody {
		t.Fatalf("delete_companies fixture must assert no_body=true")
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
