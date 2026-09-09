package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestVNextFlagAliasIntegrity182(t *testing.T) {
	for _, mutation := range []string{"healthy", "omitted_flag", "same_count_swapped_targets", "erased_metadata", "duplicate_alias", "ledger_target", "ledger_source", "erased_ledger"} {
		t.Run(mutation, func(t *testing.T) {
			lock := flagNamespaceLock182(t, []string{"limit", "plan"})
			d, err := canonicalizeVNextSourceLock(lock)
			if err != nil {
				t.Fatal(err)
			}
			cmd := &d.Graph.Operations[0].Commands[0]
			switch mutation {
			case "omitted_flag":
				cmd.Spec.Flags = cmd.Spec.Flags[:1]
			case "same_count_swapped_targets":
				cmd.Spec.Flags[0].MapsTo, cmd.Spec.Flags[1].MapsTo = cmd.Spec.Flags[1].MapsTo, cmd.Spec.Flags[0].MapsTo
			case "erased_metadata":
				cmd.Spec.Flags[0].Type = "string"
			case "duplicate_alias":
				cmd.Spec.Flags[1].Name = cmd.Spec.Flags[0].Name
			case "ledger_target":
				cmd.FlagAliases[0].MapsTo = "query.wrong"
			case "ledger_source":
				cmd.FlagAliases[0].SourceID = "wrong-source"
			case "erased_ledger":
				cmd.FlagAliases = nil
			}
			cmd.Raw, err = json.Marshal(cmd.Spec)
			if err != nil {
				t.Fatal(err)
			}
			d.Operations[0].Commands[0].Command = cmd.Raw
			_, err = admitVNextCanonicalDescriptor(d, vNextSemanticAdmissionInput{})
			if mutation == "healthy" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("corrupted exact source/alias/metadata join admitted")
			}
		})
	}
}

func TestVNextFlagAliasSourceReorder182(t *testing.T) {
	lock := flagNamespaceLock182(t, []string{"limit", "provider-limit", "plan"})
	second := flagNamespaceLock182(t, []string{"config"}).Operations[0]
	second.ID = "operation:widgets.second"
	var op map[string]json.RawMessage
	if err := json.Unmarshal(second.Operation, &op); err != nil {
		t.Fatal(err)
	}
	op["id"] = json.RawMessage(`"widgets.second"`)
	second.Operation, _ = json.Marshal(op)
	var command map[string]json.RawMessage
	if err := json.Unmarshal(second.Commands[0].Command, &command); err != nil {
		t.Fatal(err)
	}
	command["operation"] = json.RawMessage(`"widgets.second"`)
	command["path"] = json.RawMessage(`"widgets second"`)
	second.Commands[0].Command, _ = json.Marshal(command)
	second.OperationOrder = 1
	second.Commands[0].Order = 1
	lock.Operations = append(lock.Operations, second)
	first, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	lock.Operations[0], lock.Operations[1] = lock.Operations[1], lock.Operations[0]
	last, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	if !executionBundlesEqual(first.Staged.Outputs, last.Staged.Outputs) {
		t.Fatal("source reordering changed emitted execution")
	}
	if len(first.Graph.Operations) != 2 || len(last.Graph.Operations) != 2 {
		t.Fatal("source member omitted")
	}
	for i, source := range first.Graph.Operations {
		other := last.Graph.Operations[i]
		if source.ID != other.ID || source.Index == other.Index {
			t.Fatal("source identity/original index lost")
		}
		for j, cmd := range source.Commands {
			if !reflect.DeepEqual(cmd.Spec, other.Commands[j].Spec) {
				t.Fatal("canonical alias metadata changed")
			}
			for k, alias := range cmd.FlagAliases {
				want := alias
				want.OperationIndex = other.Index
				if !reflect.DeepEqual(want, other.Commands[j].FlagAliases[k]) {
					t.Fatal("original-to-canonical source join lost")
				}
			}
		}
	}
}
