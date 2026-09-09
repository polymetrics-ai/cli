package main

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestVNextReservedEnvOnlyAlias185(t *testing.T) {
	lock := flagNamespaceLock182(t, []string{"config"})
	var cmd engine.CLICommand
	if err := json.Unmarshal(lock.Operations[0].Commands[0].Command, &cmd); err != nil {
		t.Fatal(err)
	}
	cmd.Flags[0].EnvOnly = true
	cmd.Flags[0].Required = true
	raw, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}
	lock.Operations[0].Commands[0].Command = raw
	d, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatal(err)
	}
	c := d.Graph.Operations[0].Commands[0]
	f := c.Spec.Flags[0]
	if f.Name != "provider-config" || !f.EnvOnly || !f.Required || f.Type != "integer" || f.MapsTo != "query.first" {
		t.Fatal("source EnvOnly metadata changed")
	}
	if len(c.FlagAliases) != 1 || c.FlagAliases[0].OriginalName != "config" || c.FlagAliases[0].CanonicalName != "provider-config" {
		t.Fatal("alias source witness missing")
	}
	if _, err := admitVNextCanonicalDescriptor(d, vNextSemanticAdmissionInput{}); err != nil {
		t.Fatal(err)
	}
}
