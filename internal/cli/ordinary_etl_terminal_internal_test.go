package cli

import (
	"errors"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	state "polymetrics.ai/internal/state"
)

func TestETLTerminalPresentationRequiresDurableUncertainCommit(t *testing.T) {
	terminal := app.Run{ID: "run_durable_terminal", Status: "failed", CompletedAt: time.Now().UTC()}
	for _, tt := range []struct {
		name string
		run  app.Run
		err  error
		want bool
	}{
		{name: "clean terminal success", run: terminal, want: true},
		{name: "durably persisted source failure", run: terminal, err: errors.New("source is unavailable"), want: true},
		{name: "durably persisted provider failure", run: terminal, err: errors.New("provider rejected credential"), want: true},
		{name: "committed unlock uncertainty", run: terminal, err: &state.CommitOutcomeError{Outcome: state.CommitOutcomeCommitted, Err: errors.New("unlock state")}, want: true},
		{name: "indeterminate directory uncertainty", run: terminal, err: &state.CommitOutcomeError{Outcome: state.CommitOutcomeIndeterminate, Err: errors.New("sync directory")}, want: true},
		{name: "definite no-commit has zero run", err: &state.CommitOutcomeError{Outcome: state.CommitOutcomeNotCommitted, Err: errors.New("lock state")}, want: false},
		{name: "uncertain error without terminal identity", err: &state.CommitOutcomeError{Outcome: state.CommitOutcomeIndeterminate, Err: errors.New("sync directory")}, want: false},
		{name: "uncertain error with nonterminal run", run: app.Run{ID: "run_running", Status: "running"}, err: &state.CommitOutcomeError{Outcome: state.CommitOutcomeIndeterminate, Err: errors.New("sync directory")}, want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldPresentETLTerminalRun(tt.run); got != tt.want {
				t.Fatalf("shouldPresentETLTerminalRun(%#v) = %t, want %t", tt.run, got, tt.want)
			}
		})
	}
}
