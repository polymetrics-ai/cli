package conformance

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestTwentyFixturesCoverAllStreamsAndWrites(t *testing.T) {
	b, err := engine.Load(os.DirFS("../defs"), "twenty")
	if err != nil {
		t.Fatalf("Load(twenty): %v", err)
	}
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
