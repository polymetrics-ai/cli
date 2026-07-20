package run

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/events"
)

func TestDashboardFramesCoverLifecycleLayoutsAndHygiene(t *testing.T) {
	model := NewModel(Config{
		Title:         "Flow",
		Name:          "likely-customers",
		Width:         100,
		Height:        30,
		NoColor:       true,
		ReducedMotion: true,
		Steps: []Step{
			{ID: "sync-hubspot", Kind: "sync", Detail: "hubspot-prod: contacts"},
			{ID: "score-contacts", Kind: "query", Detail: "shape contacts"},
			{ID: "export-scored", Kind: "action", Detail: "waiting on score-contacts"},
		},
		StartedAt: time.Unix(0, 0),
		Now:       func() time.Time { return time.Unix(51, 0) },
	})

	model.Apply(events.Event{Kind: events.KindStarted, Scope: events.ScopeFlow, RunID: "likely-customers", Status: "running"})
	model.Apply(events.Event{Kind: events.KindStarted, Scope: events.ScopeFlow, RunID: "likely-customers", StepID: "sync-hubspot", Status: "running"})
	model.Apply(events.Event{Kind: events.KindCompleted, Scope: events.ScopeFlow, RunID: "likely-customers", StepID: "sync-hubspot", Status: "success", Counters: events.Counters{RecordsRead: 12480, RecordsWritten: 12480}})
	model.Apply(events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "likely-customers", StepID: "score-contacts", Status: "failed", Message: "connector failed token=abc123\x1b[31m"})
	model.Apply(events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "likely-customers", Status: "failed", Message: "connector failed token=abc123\x1b[31m"})

	frame := model.Frame()
	for _, want := range []string{
		"Flow likely-customers",
		"elapsed 00:51",
		"✓ sync-hubspot",
		"12,480 read → 12,480 written",
		"245 records/s",
		"✗ score-contacts",
		"failed — connector failed token=[redacted]",
		"○ export-scored",
		"✗ Flow likely-customers failed",
		"NORMAL · run",
		"ctrl+c cancel",
	} {
		if !strings.Contains(frame, want) {
			t.Fatalf("frame missing %q:\n%s", want, frame)
		}
	}
	if strings.Contains(frame, "\x1b[") {
		t.Fatalf("no-color frame contains ANSI: %q", frame)
	}
	if strings.Contains(frame, "abc123") {
		t.Fatalf("frame leaked token value:\n%s", frame)
	}
}

func TestDashboardResponsiveAndAccessibleFrames(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want []string
	}{
		{
			name: "wide",
			cfg:  Config{Title: "ETL", Name: "customers", Width: 160, Height: 45, NoColor: true, Steps: []Step{{ID: "customers", Kind: "stream"}}},
			want: []string{"ETL customers", "customers"},
		},
		{
			name: "standard",
			cfg:  Config{Title: "ETL", Name: "customers", Width: 80, Height: 24, NoColor: true, Steps: []Step{{ID: "customers", Kind: "stream"}}},
			want: []string{"ETL customers", "customers"},
		},
		{
			name: "compact",
			cfg:  Config{Title: "ETL", Name: "customers", Width: 64, Height: 20, NoColor: true, Steps: []Step{{ID: "customers", Kind: "stream"}}},
			want: []string{"ETL customers", "compact", "customers"},
		},
		{
			name: "guard",
			cfg:  Config{Title: "ETL", Name: "customers", Width: 50, Height: 12, NoColor: true, Steps: []Step{{ID: "customers", Kind: "stream"}}},
			want: []string{"Terminal too small", "50x12", "pm etl run"},
		},
		{
			name: "ascii",
			cfg:  Config{Title: "ETL", Name: "customers", Width: 80, Height: 24, NoColor: true, ASCII: true, Steps: []Step{{ID: "customers", Kind: "stream"}, {ID: "load", Kind: "warehouse"}}},
			want: []string{"[ ] customers", "|"},
		},
		{
			name: "accessible",
			cfg:  Config{Title: "ETL", Name: "customers", Width: 80, Height: 24, NoColor: true, Accessible: true, Steps: []Step{{ID: "customers", Kind: "stream"}}},
			want: []string{"step customers pending", "mode normal focus run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(tt.cfg)
			frame := model.Frame()
			for _, want := range tt.want {
				if !strings.Contains(frame, want) {
					t.Fatalf("frame missing %q:\n%s", want, frame)
				}
			}
		})
	}
}

func TestDashboardNavigationHelpAndResize(t *testing.T) {
	model := NewModel(Config{
		Title:  "Flow",
		Name:   "navigation",
		Width:  100,
		Height: 30,
		Steps: []Step{
			{ID: "extract", Kind: "sync"},
			{ID: "shape", Kind: "query"},
			{ID: "load", Kind: "action"},
		},
	})

	if got := model.SelectedStep(); got != "extract" {
		t.Fatalf("initial selection = %q, want extract", got)
	}
	model.HandleKey("j")
	if got := model.SelectedStep(); got != "shape" {
		t.Fatalf("j selection = %q, want shape", got)
	}
	model.HandleKey("down")
	if got := model.SelectedStep(); got != "load" {
		t.Fatalf("down selection = %q, want load", got)
	}
	model.HandleKey("home")
	model.HandleKey("g")
	if got := model.SelectedStep(); got != "extract" {
		t.Fatalf("home/gg selection = %q, want extract", got)
	}
	model.HandleKey("G")
	if got := model.SelectedStep(); got != "load" {
		t.Fatalf("G selection = %q, want load", got)
	}
	model.HandleKey("?")
	if frame := model.Frame(); !strings.Contains(frame, "up/k") || !strings.Contains(frame, "down/j") {
		t.Fatalf("help frame lacks arrow/Vim equivalents:\n%s", frame)
	}
	model.HandleKey("esc")
	if frame := model.Frame(); strings.Contains(frame, "up/k") {
		t.Fatalf("esc did not close one help layer:\n%s", frame)
	}

	model.Resize(50, 12)
	if frame := model.Frame(); !strings.Contains(frame, "Terminal too small: 50x12") {
		t.Fatalf("resize did not enter guard layout:\n%s", frame)
	}
}

func TestDashboardCancelWaitsForTruthfulFinalFrame(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	model := NewModel(Config{Title: "Flow", Name: "likely-customers", Width: 80, Height: 24, NoColor: true, Cancel: cancel, Steps: []Step{{ID: "sync-hubspot", Kind: "sync"}}})

	model.Apply(events.Event{Kind: events.KindStarted, Scope: events.ScopeFlow, RunID: "likely-customers", StepID: "sync-hubspot", Status: "running"})
	if done := model.HandleKey("ctrl+c"); done {
		t.Fatal("HandleKey(ctrl+c) quit before final lifecycle event")
	}
	if err := ctx.Err(); err != context.Canceled {
		t.Fatalf("cancel context err = %v, want context.Canceled", err)
	}
	beforeFinal := model.Frame()
	if !strings.Contains(beforeFinal, "cancelling") || strings.Contains(beforeFinal, "finished") {
		t.Fatalf("pre-final cancel frame is not truthful:\n%s", beforeFinal)
	}

	model.Apply(events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "likely-customers", StepID: "sync-hubspot", Status: "cancelled", Message: context.Canceled.Error()})
	model.Apply(events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "likely-customers", Status: "cancelled", Message: context.Canceled.Error()})
	if done := model.Done(); !done {
		t.Fatal("model not done after terminal cancellation event")
	}
	final := model.Frame()
	for _, want := range []string{"– Cancelled after sync-hubspot", "Resume: pm flow run likely-customers"} {
		if !strings.Contains(final, want) {
			t.Fatalf("cancel final frame missing %q:\n%s", want, final)
		}
	}
}

func TestSessionRendersLiveUpdatesAndPersistsFinalFrame(t *testing.T) {
	var output bytes.Buffer
	session := NewSession(context.Background(), SessionOptions{
		Config: Config{
			Title:  "Flow",
			Name:   "live",
			Width:  80,
			Height: 24,
			Steps:  []Step{{ID: "extract", Kind: "sync"}},
		},
		Output:   &output,
		Interval: time.Hour,
	})
	err := session.Execute(func(ctx context.Context) error {
		events.Emit(ctx, events.Event{Kind: events.KindStarted, Scope: events.ScopeFlow, RunID: "live", StepID: "extract", Status: "running"})
		events.Emit(ctx, events.Event{Kind: events.KindCompleted, Scope: events.ScopeFlow, RunID: "live", StepID: "extract", Status: "success", Counters: events.Counters{RecordsRead: 10, RecordsWritten: 10}})
		events.Emit(ctx, events.Event{Kind: events.KindCompleted, Scope: events.ScopeFlow, RunID: "live", Status: "success"})
		return nil
	})
	if err != nil {
		t.Fatalf("session Execute: %v", err)
	}
	if strings.Count(output.String(), "Flow live") < 2 {
		t.Fatalf("session did not render live lifecycle updates:\n%q", output.String())
	}
	if !strings.Contains(output.String(), "Flow live finished") {
		t.Fatalf("session final frame missing from scrollback:\n%q", output.String())
	}
}

func TestSessionCancellationPropagatesAndDrainsFinalLifecycle(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	collector := events.NewCollector()
	session := NewSession(parent, SessionOptions{
		Config: Config{
			Title:  "Flow",
			Name:   "cancel-session",
			Width:  80,
			Height: 24,
			Steps:  []Step{{ID: "extract", Kind: "sync"}},
		},
		Upstream: collector,
		Interval: time.Hour,
	})
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- session.Execute(func(ctx context.Context) error {
			events.Emit(ctx, events.Event{Kind: events.KindStarted, Scope: events.ScopeFlow, RunID: "cancel-session", StepID: "extract", Status: "running"})
			close(started)
			<-ctx.Done()
			events.Emit(ctx, events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "cancel-session", StepID: "extract", Status: "cancelled", Message: ctx.Err().Error()})
			events.Emit(ctx, events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "cancel-session", Status: "cancelled", Message: ctx.Err().Error()})
			return ctx.Err()
		})
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("session error = %v, want context.Canceled", err)
	}
	if !session.Model().Done() {
		t.Fatal("session returned before final lifecycle reached model")
	}
	if frame := session.Model().Frame(); !strings.Contains(frame, "Cancelled after extract") {
		t.Fatalf("session final frame is not truthful:\n%s", frame)
	}
	got := collector.Events()
	if len(got) != 3 || got[len(got)-1].Status != "cancelled" {
		t.Fatalf("upstream lifecycle events = %+v, want started + terminal step/run", got)
	}
}

func TestBridgeThrottlesProgressWithoutDroppingLifecycle(t *testing.T) {
	collector := events.NewCollector()
	bridge := NewBridge(BridgeOptions{Interval: time.Hour, Sink: collector})
	ctx := context.Background()

	bridge.Emit(ctx, events.Event{Kind: events.KindStarted, Scope: events.ScopeETL, RunID: "run-1"})
	bridge.Emit(ctx, events.Event{Kind: events.KindProgress, Scope: events.ScopeETL, RunID: "run-1", StepID: "customers", Counters: events.Counters{RecordsRead: 1}})
	bridge.Emit(ctx, events.Event{Kind: events.KindProgress, Scope: events.ScopeETL, RunID: "run-1", StepID: "customers", Counters: events.Counters{RecordsRead: 2}})
	bridge.Emit(ctx, events.Event{Kind: events.KindCompleted, Scope: events.ScopeETL, RunID: "run-1"})
	bridge.Flush(ctx)

	got := collector.Events()
	if len(got) != 4 {
		t.Fatalf("bridge events = %d, want 4: %+v", len(got), got)
	}
	if got[0].Kind != events.KindStarted || got[3].Kind != events.KindCompleted {
		t.Fatalf("lifecycle events not preserved at edges: %+v", got)
	}
	if got[2].Kind != events.KindProgress || got[2].Counters.RecordsRead != 2 {
		t.Fatalf("latest progress not flushed before terminal: %+v", got)
	}
	if bridge.Dropped() == 0 {
		t.Fatal("bridge did not account throttled progress")
	}
}
