package syncplan

import (
	"testing"

	"polymetrics.ai/internal/synccontract"
)

func FuzzResolveNeverExecutesContradictoryModeAxes(f *testing.F) {
	for modeIndex := range synccontract.AllModes() {
		f.Add(uint64(modeIndex), uint64(0))
		for mutation := uint64(1); mutation <= 10; mutation++ {
			f.Add(uint64(modeIndex), mutation)
		}
	}
	f.Fuzz(func(t *testing.T, modeIndex, mutation uint64) {
		modes := synccontract.AllModes()
		if len(modes) == 0 {
			t.Fatal("canonical mode vocabulary is empty")
		}
		mode := modes[modeIndex%uint64(len(modes))]
		plan := validPlan()
		axes, ok := synccontract.ModeAxes(mode)
		if !ok {
			t.Fatalf("ModeAxes(%q) missing", mode)
		}
		plan.Mode = mode
		plan.Axes = axes
		other := modes[(modeIndex+1)%uint64(len(modes))]
		otherAxes, _ := synccontract.ModeAxes(other)
		switch mutation % 11 {
		case 1:
			plan.Axes.Progression = otherAxes.Progression
		case 2:
			plan.Axes.Apply = otherAxes.Apply
		case 3:
			plan.Axes.Object = otherAxes.Object
		case 4:
			plan.Axes.Key = otherAxes.Key
		case 5:
			plan.Axes.Delete = otherAxes.Delete
		case 6:
			plan.Axes.Retry = otherAxes.Retry
		case 7:
			plan.Axes.Idempotency = otherAxes.Idempotency
		case 8:
			plan.Axes.Receipt = otherAxes.Receipt
		case 9:
			plan.Axes.Acknowledgement = otherAxes.Acknowledgement
		case 10:
			plan.Axes.Checkpoint = otherAxes.Checkpoint
		}
		result := Resolve(plan, synccontract.Budget{})
		if err := result.Validate(); err != nil {
			t.Fatalf("result validation = %v for mode=%q mutation=%d result=%#v", err, mode, mutation, result)
		}
		if result.Kind == ResultKindExecutable {
			if err := synccontract.ValidateModeAxes(result.Plan.Mode, result.Plan.Axes); err != nil {
				t.Fatalf("executable contradiction for mode=%q mutation=%d: %v", mode, mutation, err)
			}
		}
	})
}
