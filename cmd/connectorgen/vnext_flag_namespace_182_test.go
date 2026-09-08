package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func flagNamespaceLock182(t *testing.T, names []string) vNextSourceLock {
	t.Helper()
	lock := operationDirectReadLockForSemanticAdmissionTest()
	var op engine.OperationSpec
	if err := json.Unmarshal(lock.Operations[0].Operation, &op); err != nil {
		t.Fatal(err)
	}
	var cmd engine.CLICommand
	if err := json.Unmarshal(lock.Operations[0].Commands[0].Command, &cmd); err != nil {
		t.Fatal(err)
	}
	for i, name := range names {
		// Provider targets are independent of public names, including aliases.
		target := []string{"first", "second", "third", "fourth"}[i]
		op.REST.Parameters = append(op.REST.Parameters, engine.OperationParameter{Name: target, In: "query", Type: "integer"})
		cmd.Flags = append(cmd.Flags, engine.CLIFlag{Name: name, Type: "integer", MapsTo: "query." + target})
	}
	var err error
	lock.Operations[0].Operation, err = json.Marshal(op)
	if err != nil {
		t.Fatal(err)
	}
	lock.Operations[0].Commands[0].Command, err = json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	return lock
}

func TestVNextFlagNamespace182(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input, want []string
		invalid     bool
	}{
		{"reserved", []string{"limit"}, []string{"provider-limit"}, false},
		{"occupied", []string{"limit", "provider-limit"}, []string{"provider-provider-limit", "provider-limit"}, false},
		{"prefix_chain", []string{"provider-provider-limit", "limit", "provider-limit"}, []string{"provider-provider-limit", "provider-provider-provider-limit", "provider-limit"}, false},
		{"multiple", []string{"config", "plan", "page", "root"}, []string{"provider-config", "provider-plan", "provider-page", "provider-root"}, false},
		{"safe", []string{"provider-limit", "ordinary"}, []string{"provider-limit", "ordinary"}, false},
		{"duplicate", []string{"limit", "limit"}, nil, true},
		{"raw_space", []string{" limit"}, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := flagNamespaceLock182(t, tc.input)
			d, err := canonicalizeVNextSourceLock(lock)
			if tc.invalid {
				if err == nil {
					t.Fatal("malformed raw namespace admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, flag := range d.Graph.Operations[0].Commands[0].Spec.Flags {
				got = append(got, flag.Name)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("public namespace %v; want %v", got, tc.want)
			}
		})
	}
}

func TestVNextReservedFlagAdmissionBypass182(t *testing.T) {
	for _, boundary := range []string{"render", "admission"} {
		t.Run(boundary, func(t *testing.T) {
			d, err := canonicalizeVNextSourceLock(flagNamespaceLock182(t, []string{"provider-limit"}))
			if err != nil {
				t.Fatal(err)
			}
			cmd := &d.Graph.Operations[0].Commands[0]
			cmd.Spec.Flags[0].Name = "limit"
			cmd.Raw, err = json.Marshal(cmd.Spec)
			if err != nil {
				t.Fatal(err)
			}
			d.Operations[0].Commands[0].Command = cmd.Raw
			if boundary == "render" {
				_, err = renderVNextExecutionBundle(d)
			} else {
				_, err = admitVNextCanonicalDescriptor(d, vNextSemanticAdmissionInput{})
			}
			if err == nil || !strings.Contains(err.Error(), "reserved") {
				t.Fatalf("ambiguous final provider flag must be refused with reserved-name diagnostic: %v", err)
			}
		})
	}
}
