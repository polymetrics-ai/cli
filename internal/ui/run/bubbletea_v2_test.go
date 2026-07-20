package run

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/exp/teatest/v2"
)

var _ tea.Model = (*Model)(nil)

func TestBubbleTeaV2ModelAndTeatestProgram(t *testing.T) {
	model := NewModel(Config{
		Title:  "Flow",
		Name:   "teatest-contract",
		Width:  80,
		Height: 24,
		Steps:  []Step{{ID: "extract", Kind: "sync"}},
	})

	tm := teatest.NewTestModel(t, model, teatest.WithInitialTermSize(80, 24))
	tm.Send(tea.Quit())
	final := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := final.(*Model); !ok {
		t.Fatalf("final model type = %T, want *run.Model", final)
	}
}
