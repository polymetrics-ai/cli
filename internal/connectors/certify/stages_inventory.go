package certify

import (
	"encoding/json"
	"fmt"

	"polymetrics.ai/internal/connectors/defs"
)

type writesFile struct {
	Actions []writeActionDecl `json:"actions"`
}

type writeActionDecl struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Path string `json:"path"`
	Risk string `json:"risk"`
}

type writeActionInventoryItem struct {
	Action  string
	Path    string
	Risk    string
	Pairing WritePairing
	Reason  string
}

func declaredWriteActions(connector string) ([]writeActionDecl, error) {
	raw, err := defs.FS.ReadFile(connector + "/writes.json")
	if err != nil {
		return nil, fmt.Errorf("read %s writes: %w", connector, err)
	}
	var file writesFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("parse %s writes: %w", connector, err)
	}
	out := make([]writeActionDecl, 0, len(file.Actions))
	for _, action := range file.Actions {
		name := action.Name
		if name == "" {
			name = action.ID
		}
		if name != "" {
			action.Name = name
			out = append(out, action)
		}
	}
	return out, nil
}

func writeActionInventoryFor(connector string) ([]writeActionInventoryItem, error) {
	declared, err := declaredWriteActions(connector)
	if err != nil {
		return nil, err
	}
	curated := map[string]WritePairing{}
	names := make([]string, 0, len(declared))
	for _, action := range declared {
		names = append(names, action.Name)
	}
	for _, pairing := range PairingsFor(connector) {
		curated[pairing.Create] = pairing
		if pairing.Cleanup != "" {
			curated[pairing.Cleanup] = WritePairing{}
		}
	}

	out := make([]writeActionInventoryItem, 0, len(names))
	for _, declaredAction := range declared {
		name := declaredAction.Name
		item := writeActionInventoryItem{Action: name, Path: declaredAction.Path, Risk: declaredAction.Risk}
		if pairing, ok := curated[name]; ok && pairing.Create != "" {
			item.Pairing = pairing
			out = append(out, item)
			continue
		}
		if pairing, ok := InferPairing(name, names); ok {
			item.Pairing = pairing
			item.Reason = "inferred create/cleanup pair exists but no curated verify stream/id field is certified for live execution yet"
			out = append(out, item)
			continue
		}
		item.Reason = "not a safe standalone create action with a certified cleanup lifecycle"
		out = append(out, item)
	}
	return out, nil
}
