package commandrunner

import (
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestSquareOperationDirectReadFlagsBuildSchemaValidBodies(t *testing.T) {
	b, err := engine.Load(defs.FS, "square")
	if err != nil {
		t.Fatalf("load square bundle: %v", err)
	}
	connector := engine.New(b, nil)
	surface := connector.CommandSurface()
	if surface == nil {
		t.Fatalf("square command surface missing")
	}

	operations := map[string]engine.OperationSpec{}
	for _, op := range b.Operations {
		operations[op.ID] = op
	}

	for _, cmd := range surface.Commands {
		if cmd.Intent != "direct_read" || cmd.Availability != "implemented" || cmd.Operation == "" {
			continue
		}
		op := operations[cmd.Operation]
		if op.REST == nil || op.REST.Method != "POST" || len(op.REST.BodySchema) == 0 {
			continue
		}

		flags := map[string][]string{}
		for _, flag := range cmd.Flags {
			flags[flag.Name] = []string{squareSampleFlagValue(flag)}
		}
		_, _, body, err := operationDirectReadOverrides(cmd, flags)
		if err != nil {
			t.Fatalf("%s: build body overrides: %v", cmd.Path, err)
		}
		schema, err := engine.CompileSchema(op.REST.BodySchema)
		if err != nil {
			t.Fatalf("%s: compile body schema: %v", cmd.Path, err)
		}
		if err := schema.Validate(body); err != nil {
			t.Fatalf("%s: generated body does not satisfy schema: %v; body=%+v", cmd.Path, err, body)
		}
	}
}

func TestSquareOperationDirectReadHasNoRawObjectBodyFlags(t *testing.T) {
	b, err := engine.Load(defs.FS, "square")
	if err != nil {
		t.Fatalf("load square bundle: %v", err)
	}
	surface := engine.New(b, nil).CommandSurface()
	for _, cmd := range surface.Commands {
		if cmd.Intent != "direct_read" || cmd.Availability != "implemented" {
			continue
		}
		for _, flag := range cmd.Flags {
			switch flag.MapsTo {
			case "body.query", "body.order", "body.reason_id":
				t.Fatalf("%s flag --%s maps raw object target %q", cmd.Path, flag.Name, flag.MapsTo)
			}
		}
	}
}

func squareSampleFlagValue(flag connectors.CommandSurfaceFlag) string {
	if flag.Format == "date-time" {
		return "2026-01-01T00:00:00Z"
	}
	switch flag.Type {
	case "boolean":
		return "true"
	case "integer":
		return "10"
	case "enum":
		if len(flag.Values) > 0 {
			return flag.Values[0]
		}
		return "FIXTURE"
	case "string_array":
		return "fixture_one,fixture_two"
	default:
		return "fixture_value"
	}
}
