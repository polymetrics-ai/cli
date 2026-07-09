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
}

type writeActionInventoryItem struct {
	Action  string
	Pairing WritePairing
	Reason  string
}

func declaredWriteActionNames(connector string) ([]string, error) {
	raw, err := defs.FS.ReadFile(connector + "/writes.json")
	if err != nil {
		return nil, fmt.Errorf("read %s writes: %w", connector, err)
	}
	var file writesFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("parse %s writes: %w", connector, err)
	}
	out := make([]string, 0, len(file.Actions))
	for _, action := range file.Actions {
		name := action.Name
		if name == "" {
			name = action.ID
		}
		if name != "" {
			out = append(out, name)
		}
	}
	return out, nil
}

func writeActionInventoryFor(connector string) ([]writeActionInventoryItem, error) {
	names, err := declaredWriteActionNames(connector)
	if err != nil {
		return nil, err
	}
	curated := map[string]WritePairing{}
	for _, pairing := range PairingsFor(connector) {
		curated[pairing.Create] = pairing
		if pairing.Cleanup != "" {
			curated[pairing.Cleanup] = WritePairing{}
		}
	}

	out := make([]writeActionInventoryItem, 0, len(names))
	for _, name := range names {
		if pairing, ok := curated[name]; ok && pairing.Create != "" {
			out = append(out, writeActionInventoryItem{Action: name, Pairing: pairing})
			continue
		}
		if pairing, ok := InferPairing(name, names); ok {
			out = append(out, writeActionInventoryItem{
				Action:  name,
				Pairing: pairing,
				Reason:  "inferred create/cleanup pair exists but no curated verify stream/id field is certified for live execution yet",
			})
			continue
		}
		out = append(out, writeActionInventoryItem{
			Action: name,
			Reason: "not a safe standalone create action with a certified cleanup lifecycle",
		})
	}
	return out, nil
}
