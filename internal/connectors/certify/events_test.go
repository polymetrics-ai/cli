package certify_test

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/events"
)

func TestRunBatchEmitsDeterministicEvents(t *testing.T) {
	cf := certify.CredsFile{
		Defaults: certify.CredsDefaults{Parallel: 1},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {},
			"stripe": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		return &fakeRunnable{rep: passingReport(name)}
	}
	collector := events.NewCollector()

	batch, err := certify.RunBatch(events.WithEmitter(context.Background(), collector), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if batch.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", batch.ExitCode)
	}

	got := certifyEventSequence(collector.Events())
	want := []string{
		"certify::started:running",
		"certify:github:queued:queued",
		"certify:stripe:queued:queued",
		"certify:github:started:running",
		"certify:github:completed:success",
		"certify:stripe:started:running",
		"certify:stripe:completed:success",
		"certify::completed:success",
	}
	if len(got) != len(want) {
		t.Fatalf("sequence length = %d, want %d\ngot  %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %q, want %q\ngot  %#v\nwant %#v", i, got[i], want[i], got, want)
		}
	}
}

func TestRunBatchEmitsSkipAndFailureEvents(t *testing.T) {
	cf := certify.CredsFile{
		Defaults: certify.CredsDefaults{Parallel: 1},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {Skip: true, Reason: "not requested"},
			"stripe": {},
		},
	}
	factory := func(name string, _ certify.Options) certify.Runnable {
		if name == "stripe" {
			return &fakeRunnable{err: errors.New("certify failed")}
		}
		return &fakeRunnable{rep: passingReport(name)}
	}
	collector := events.NewCollector()

	batch, err := certify.RunBatch(events.WithEmitter(context.Background(), collector), certify.BatchOptions{
		CredsFile:     cf,
		RunnerFactory: factory,
		BatchDir:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunBatch() error = %v", err)
	}
	if batch.ExitCode != 2 {
		t.Fatalf("ExitCode = %d, want 2", batch.ExitCode)
	}

	got := certifyEventSequence(collector.Events())
	want := []string{
		"certify::started:running",
		"certify:github:skipped:skipped",
		"certify:stripe:queued:queued",
		"certify:stripe:started:running",
		"certify:stripe:failed:failed",
		"certify::failed:failed",
	}
	if len(got) != len(want) {
		t.Fatalf("sequence length = %d, want %d\ngot  %#v\nwant %#v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sequence[%d] = %q, want %q\ngot  %#v\nwant %#v", i, got[i], want[i], got, want)
		}
	}
}

func certifyEventSequence(in []events.Event) []string {
	out := make([]string, 0, len(in))
	for _, ev := range in {
		out = append(out, string(ev.Scope)+":"+ev.StepID+":"+string(ev.Kind)+":"+ev.Status)
	}
	return out
}
