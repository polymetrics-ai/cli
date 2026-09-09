package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

type flagOwnershipCase181 struct {
	Connector      string            `json:"connector"`
	SourceID       string            `json:"source_id"`
	OperationIndex int               `json:"operation_index"`
	CommandIndex   int               `json:"command_index"`
	OldName        string            `json:"old_name"`
	NewName        string            `json:"new_name"`
	MapsTo         string            `json:"maps_to"`
	Command        engine.CLICommand `json:"command"`
}

// The frozen179 source witnesses specify all22 identities and complete metadata,
// including optional siblings. They are independent of the alias allocator.
func TestVNextFlagOwnershipCohort181(t *testing.T) {
	raw, err := os.ReadFile("testdata/flag-ownership-181.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []flagOwnershipCase181
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 22 {
		t.Fatal("incomplete reviewed cohort")
	}
	for _, name := range []string{"github", "gitlab"} {
		raw, err := os.ReadFile(filepath.Join("../../internal/connectors/defs", name, "source.lock.json"))
		if err != nil {
			t.Fatal(err)
		}
		lock, err := decodeVNextSourceLock(raw)
		if err != nil {
			t.Fatal(err)
		}
		before, err := json.Marshal(lock)
		if err != nil {
			t.Fatal(err)
		}
		descriptor, err := canonicalizeVNextSourceLock(lock)
		if err != nil {
			t.Fatal(err)
		}
		after, err := json.Marshal(lock)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("canonicalization mutated caller lock")
		}
		for _, tc := range cases {
			if tc.Connector != name {
				continue
			}
			t.Run(name+"/"+tc.Command.Path, func(t *testing.T) {
				want := tc.Command
				want.Flags = append([]engine.CLIFlag(nil), want.Flags...)
				changed := 0
				for i := range want.Flags {
					if want.Flags[i].Name == tc.OldName && want.Flags[i].MapsTo == tc.MapsTo {
						want.Flags[i].Name = tc.NewName
						changed++
					}
				}
				if changed != 1 {
					t.Fatal("ambiguous witness")
				}
				found := false
				for _, op := range descriptor.Graph.Operations {
					if op.ID != tc.SourceID {
						continue
					}
					for _, command := range op.Commands {
						if command.Spec.Path != tc.Command.Path {
							continue
						}
						found = true
						if !reflect.DeepEqual(command.Spec, want) {
							t.Errorf("canonical command did not preserve exact source metadata and expected alias: got flags %+v; want %+v", command.Spec.Flags, want.Flags)
						}
					}
				}
				if !found {
					t.Fatal("source-bound command omitted")
				}
			})
		}
	}
}
