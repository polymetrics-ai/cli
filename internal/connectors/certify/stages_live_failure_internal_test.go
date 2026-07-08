package certify

import "testing"

func TestResumeSkipsWhenIncrementalDidNotProduceCursor(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "sample"}}
	rep := Report{}
	if err := stageResume(rc, &rep); err != nil {
		t.Fatalf("stageResume: %v", err)
	}
	stage := rep.Stages[len(rep.Stages)-1]
	if stage.Name != "resume" {
		t.Fatalf("stage name = %q, want resume", stage.Name)
	}
	if stage.Passed || !stringsHasPrefix(stage.Error, "skipped:") {
		t.Fatalf("resume stage = %+v, want documented skip", stage)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
