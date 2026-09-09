package synccontract

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestModeAxesCoverEveryCanonicalMode(t *testing.T) {
	for _, mode := range AllModes() {
		axes, ok := ModeAxes(mode)
		if !ok {
			t.Fatalf("ModeAxes(%q) missing", mode)
		}
		if err := axes.Validate(); err != nil {
			t.Fatalf("ModeAxes(%q) invalid: %v", mode, err)
		}
		if err := ValidateModeAxes(mode, axes); err != nil {
			t.Fatalf("ValidateModeAxes(%q) rejected canonical axes: %v", mode, err)
		}
		encoded, err := json.Marshal(axes)
		if err != nil {
			t.Fatal(err)
		}
		var decoded ExecutionAxes
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, axes) {
			t.Fatalf("%q round trip = %#v, want %#v", mode, decoded, axes)
		}
	}
}

func TestExecutionAxesRejectUnknownDiscriminants(t *testing.T) {
	valid, ok := ModeAxes(ModeFullOverwrite)
	if !ok {
		t.Fatal("full_overwrite mode is missing")
	}
	mutations := []struct {
		axis string
		edit func(*ExecutionAxes)
	}{
		{"progression", func(axes *ExecutionAxes) { axes.Progression = "poll" }},
		{"apply", func(axes *ExecutionAxes) { axes.Apply = "merge" }},
		{"object", func(axes *ExecutionAxes) { axes.Object = "membership" }},
		{"key", func(axes *ExecutionAxes) { axes.Key = "nullable" }},
		{"delete", func(axes *ExecutionAxes) { axes.Delete = "physical" }},
		{"retry", func(axes *ExecutionAxes) { axes.Retry = "forever" }},
		{"idempotency", func(axes *ExecutionAxes) { axes.Idempotency = "best_effort" }},
		{"receipt", func(axes *ExecutionAxes) { axes.Receipt = "ephemeral" }},
		{"acknowledgement", func(axes *ExecutionAxes) { axes.Acknowledgement = "eventual" }},
		{"checkpoint", func(axes *ExecutionAxes) { axes.Checkpoint = "before_acknowledgement" }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.axis, func(t *testing.T) {
			axes := valid
			mutation.edit(&axes)
			err := axes.Validate()
			var typed *UnknownAxisValueError
			if !errors.As(err, &typed) || typed.Axis != mutation.axis {
				t.Fatalf("error = %T %v, want typed %s error", err, err, mutation.axis)
			}
		})
	}
}

func TestValidateModeAxesReportsFirstRealContradictoryAxis(t *testing.T) {
	axes, ok := ModeAxes(ModeIncrementalUpsert)
	if !ok {
		t.Fatal("incremental_upsert mode is missing")
	}
	axes.Apply = DestinationApplyAppend
	err := ValidateModeAxes(ModeIncrementalUpsert, axes)
	var typed *AxisError
	if !errors.As(err, &typed) || typed.Axis != "apply" {
		t.Fatalf("incremental_upsert + append error = %T %v, want apply axis", err, err)
	}
}

func TestExecutionAxesRejectNegativeBudget(t *testing.T) {
	axes, _ := ModeAxes(ModeFullOverwrite)
	axes.Budget.MaxBytes = -1
	if err := axes.Validate(); err == nil {
		t.Fatal("negative budget accepted")
	}
}
