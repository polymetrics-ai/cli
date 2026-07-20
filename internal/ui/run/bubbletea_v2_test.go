package run

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"polymetrics.ai/internal/events"
)

var _ tea.Model = (*Model)(nil)

func TestBubbleTeaV2ModelAndTeatestProgram(t *testing.T) {
	model := NewModel(Config{
		Title:   "Flow",
		Name:    "teatest-contract",
		Width:   80,
		Height:  24,
		NoColor: true,
		Steps:   []Step{{ID: "extract", Kind: "sync"}},
	})
	if view := model.View(); view.AltScreen {
		t.Fatal("run dashboard enabled the alternate screen")
	}

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	tm.Send(tea.Quit())
	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := final.(*Model); !ok {
		t.Fatalf("final model type = %T, want *run.Model", final)
	}
}

func TestTeatestDashboardLifecycleFrames(t *testing.T) {
	tests := []struct {
		name     string
		messages []tea.Msg
		want     []string
	}{
		{
			name: "success",
			messages: []tea.Msg{
				eventMsg{Event: events.Event{Kind: events.KindStarted, Scope: events.ScopeFlow, RunID: "success", StepID: "extract", Status: "running"}},
				eventMsg{Event: events.Event{Kind: events.KindCompleted, Scope: events.ScopeFlow, RunID: "success", StepID: "extract", Status: "success", Counters: events.Counters{RecordsRead: 10, RecordsWritten: 10}}},
				eventMsg{Event: events.Event{Kind: events.KindCompleted, Scope: events.ScopeFlow, RunID: "success", Status: "success"}},
				eventClosedMsg{},
				runResultMsg{},
			},
			want: []string{"Flow success finished", "10 records", "NORMAL · run"},
		},
		{
			name: "failure",
			messages: []tea.Msg{
				eventMsg{Event: events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "failure", StepID: "extract", Status: "failed", Message: "token=abc123\x1b[31m"}},
				eventMsg{Event: events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "failure", Status: "failed", Message: "token=abc123\x1b[31m"}},
				eventClosedMsg{},
				runResultMsg{Err: errors.New("token=abc123")},
			},
			want: []string{"Flow failure failed", "token=[redacted]", "NORMAL · run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(Config{
				Title:     "Flow",
				Name:      tt.name,
				Width:     100,
				Height:    30,
				NoColor:   true,
				StartedAt: time.Unix(0, 0),
				Now:       func() time.Time { return time.Unix(5, 0) },
				Steps:     []Step{{ID: "extract", Kind: "sync"}},
			})
			final, output := finishTeatest(t, model, 100, 30, tt.messages...)
			for _, want := range tt.want {
				if !strings.Contains(final.Frame(), want) {
					t.Fatalf("final frame missing %q:\n%s", want, final.Frame())
				}
			}
			if !strings.Contains(output, final.finalLine()) {
				t.Fatalf("final truthful frame missing from teatest output:\n%q", output)
			}
			if strings.Contains(final.Frame(), "abc123") {
				t.Fatalf("final frame leaked secret-like value:\n%s", final.Frame())
			}
		})
	}
}

func TestTeatestDashboardCancellationFlowsThroughUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelled := make(chan struct{})
	var once sync.Once
	model := NewModel(Config{
		Title:   "Flow",
		Name:    "cancelled",
		Width:   80,
		Height:  24,
		NoColor: true,
		Cancel: func() {
			cancel()
			once.Do(func() { close(cancelled) })
		},
		Steps: []Step{{ID: "extract", Kind: "sync"}},
	})
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	tm.Send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("ctrl+c key message did not cancel runner context through Update")
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("runner context error = %v, want context.Canceled", ctx.Err())
	}
	tm.Send(eventMsg{Event: events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "cancelled", StepID: "extract", Status: "cancelled", Message: context.Canceled.Error()}})
	tm.Send(eventMsg{Event: events.Event{Kind: events.KindFailed, Scope: events.ScopeFlow, RunID: "cancelled", Status: "cancelled", Message: context.Canceled.Error()}})
	tm.Send(eventClosedMsg{})
	tm.Send(runResultMsg{Err: context.Canceled})
	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	final, ok := finalModel.(*Model)
	if !ok {
		t.Fatalf("final model type = %T, want *run.Model", finalModel)
	}
	for _, want := range []string{"Cancelled after extract", "Resume: pm flow run cancelled", "NORMAL · run"} {
		if !strings.Contains(final.Frame(), want) {
			t.Fatalf("cancel frame missing %q:\n%s", want, final.Frame())
		}
	}
}

func TestTeatestDashboardNavigationAndHelpKeys(t *testing.T) {
	model := NewModel(Config{
		Title:   "Flow",
		Name:    "navigation",
		NoColor: true,
		Steps: []Step{
			{ID: "extract", Kind: "sync"},
			{ID: "shape", Kind: "query"},
			{ID: "load", Kind: "action"},
		},
	})
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(100, 30))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyDown})
	tm.Send(tea.KeyPressMsg{Code: 'j', Text: "j"})
	tm.Send(tea.KeyPressMsg{Code: tea.KeyUp})
	tm.Send(tea.KeyPressMsg{Code: '?', Text: "?"})
	tm.Send(tea.Quit())
	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	final, ok := finalModel.(*Model)
	if !ok {
		t.Fatalf("final model type = %T, want *run.Model", finalModel)
	}
	if got := final.SelectedStep(); got != "shape" {
		t.Fatalf("Tea arrow/Vim selection = %q, want shape", got)
	}
	for _, want := range []string{"up/k", "down/j", "esc close help"} {
		if !strings.Contains(final.Frame(), want) {
			t.Fatalf("Tea help frame missing %q:\n%s", want, final.Frame())
		}
	}
}

func TestTeatestDashboardResponsiveFrames(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		want          string
	}{
		{name: "wide", width: 160, height: 45, want: "Flow responsive"},
		{name: "standard-100", width: 100, height: 30, want: "Flow responsive"},
		{name: "standard-80", width: 80, height: 24, want: "Flow responsive"},
		{name: "compact", width: 64, height: 20, want: "compact"},
		{name: "guard", width: 50, height: 12, want: "Terminal too small: 50x12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel(Config{Title: "Flow", Name: "responsive", NoColor: true, Steps: []Step{{ID: "extract", Kind: "sync"}}})
			final, _ := finishTeatest(t, model, tt.width, tt.height, eventClosedMsg{}, runResultMsg{})
			if !strings.Contains(final.Frame(), tt.want) {
				t.Fatalf("responsive frame missing %q:\n%s", tt.want, final.Frame())
			}
			if final.cfg.Width != tt.width || final.cfg.Height != tt.height {
				t.Fatalf("window size = %dx%d, want %dx%d", final.cfg.Width, final.cfg.Height, tt.width, tt.height)
			}
		})
	}
}

func finishTeatest(t *testing.T, model *Model, width, height int, messages ...tea.Msg) (*Model, string) {
	t.Helper()
	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(width, height))
	for _, msg := range messages {
		tm.Send(msg)
	}
	finalModel := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	final, ok := finalModel.(*Model)
	if !ok {
		t.Fatalf("final model type = %T, want *run.Model", finalModel)
	}
	output, err := io.ReadAll(tm.Output())
	if err != nil {
		t.Fatalf("read teatest output: %v", err)
	}
	return final, string(output)
}
