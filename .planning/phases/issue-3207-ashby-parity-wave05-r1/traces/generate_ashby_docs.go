package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/ashby"
)

func main() {
	connector := ashby.New()
	definition, ok := connectors.DefinitionOf(connector)
	if !ok {
		panic("ashby does not expose a connector definition")
	}

	connectorDir := filepath.Join("docs", "connectors", "ashby")
	manual := "# pm connectors inspect ashby\n\n```text\n" + connectors.RenderConnectorManual(connector) + "\n```\n"
	mustWrite(filepath.Join(connectorDir, "MANUAL.md"), []byte(manual))
	mustWrite(filepath.Join(connectorDir, "SKILL.md"), []byte(connectors.RenderConnectorSkill(connector)))

	catalogPath := filepath.Join("docs", "connectors", "catalog", "all-connectors.json")
	var catalog []connectors.Definition
	mustDecode(catalogPath, &catalog)
	found := false
	for index := range catalog {
		if catalog[index].Name == definition.Name {
			catalog[index] = definition
			found = true
			break
		}
	}
	if !found {
		panic("ashby is missing from the connector catalog")
	}
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		panic(fmt.Errorf("encode connector catalog: %w", err))
	}
	mustWrite(catalogPath, append(data, '\n'))
}

func mustDecode(path string, target any) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read %s: %w", path, err))
	}
	if err := json.Unmarshal(data, target); err != nil {
		panic(fmt.Errorf("decode %s: %w", path, err))
	}
}

func mustWrite(path string, data []byte) {
	if err := os.WriteFile(path, data, 0o644); err != nil {
		panic(fmt.Errorf("write %s: %w", path, err))
	}
}
