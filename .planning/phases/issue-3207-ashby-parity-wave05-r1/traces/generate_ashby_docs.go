package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	updateCatalogMarkdown(
		filepath.Join("docs", "connectors", "catalog", "all-connectors.md"),
		len(definition.Streams),
		len(definition.WriteActions),
	)
}

func updateCatalogMarkdown(path string, streamCount, writeCount int) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("read %s: %w", path, err))
	}
	lines := strings.Split(string(data), "\n")
	found := false
	for index, line := range lines {
		if !strings.HasPrefix(line, "| `ashby` |") {
			continue
		}
		columns := strings.Split(line, "|")
		if len(columns) != 12 {
			panic(fmt.Errorf("unexpected Ashby catalog row: %s", line))
		}
		columns[7] = " " + strconv.Itoa(streamCount) + " "
		columns[8] = " " + strconv.Itoa(writeCount) + " "
		lines[index] = strings.Join(columns, "|")
		found = true
		break
	}
	if !found {
		panic("ashby is missing from the connector markdown catalog")
	}
	mustWrite(path, []byte(strings.Join(lines, "\n")))
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
