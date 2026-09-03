package synccontract

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestExecutionAxesAcceptClosedControls(t *testing.T) {
	controls := []ExecutionAxes{
		{Progression: SourceProgressionSnapshot, Apply: DestinationApplyReplace, Object: ObjectClassObject, Binding: BindingKindStream, Key: KeyShapeNone, Delete: DeletePolicyNone},
		{Progression: SourceProgressionCursor, Apply: DestinationApplyAppend, Object: ObjectClassEvent, Binding: BindingKindStream, Key: KeyShapeNone, Delete: DeletePolicyNone, Budget: Budget{MaxRecords: 100}},
		{Progression: SourceProgressionCursor, Apply: DestinationApplyDedupeHistory, Object: ObjectClassObject, Binding: BindingKindAction, Key: KeyShapeComposite, Delete: DeletePolicyHistoryClose},
		{Progression: SourceProgressionChangeCapture, Apply: DestinationApplyChangeApply, Object: ObjectClassObject, Binding: BindingKindStream, Key: KeyShapeStable, Delete: DeletePolicyTombstone},
	}
	for _, axes := range controls {
		if err := axes.Validate(); err != nil {
			t.Fatalf("axes %#v rejected: %v", axes, err)
		}
		encoded, err := json.Marshal(axes)
		if err != nil {
			t.Fatalf("marshal axes: %v", err)
		}
		var decoded ExecutionAxes
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatalf("unmarshal axes: %v", err)
		}
		if !reflect.DeepEqual(decoded, axes) {
			t.Fatalf("round trip = %#v, want %#v", decoded, axes)
		}
	}
}

func TestExecutionAxesRejectUnknownDiscriminants(t *testing.T) {
	valid := ExecutionAxes{Progression: SourceProgressionSnapshot, Apply: DestinationApplyReplace, Object: ObjectClassObject, Binding: BindingKindStream, Key: KeyShapeNone, Delete: DeletePolicyNone}
	mutations := []struct {
		axis string
		edit func(*ExecutionAxes)
	}{
		{"progression", func(axes *ExecutionAxes) { axes.Progression = "poll" }},
		{"apply", func(axes *ExecutionAxes) { axes.Apply = "merge" }},
		{"object", func(axes *ExecutionAxes) { axes.Object = "membership" }},
		{"binding", func(axes *ExecutionAxes) { axes.Binding = "route" }},
		{"key", func(axes *ExecutionAxes) { axes.Key = "nullable" }},
		{"delete", func(axes *ExecutionAxes) { axes.Delete = "physical" }},
	}
	for _, mutation := range mutations {
		axes := valid
		mutation.edit(&axes)
		err := axes.Validate()
		var typed *UnknownAxisValueError
		if !errors.As(err, &typed) || typed.Axis != mutation.axis {
			t.Fatalf("%s error = %T %v, want typed %s error", mutation.axis, err, err, mutation.axis)
		}
	}
}

func TestExecutionAxesRejectBudgetWideningSentinel(t *testing.T) {
	axes := ExecutionAxes{Progression: SourceProgressionSnapshot, Apply: DestinationApplyReplace, Object: ObjectClassObject, Binding: BindingKindStream, Key: KeyShapeNone, Delete: DeletePolicyNone, Budget: Budget{MaxBytes: -1}}
	if err := axes.Validate(); err == nil {
		t.Fatal("negative budget accepted")
	}
}
