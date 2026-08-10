// Package certifications exposes the generated, proof-bearing connector
// certification status to pm's user-facing surfaces. It has no dependency on
// connector execution, so reading a status never reaches credentials or a
// provider.
package certifications

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

const (
	statusSchemaVersion    = 1
	statusGeneratedCommand = "go run ./cmd/connectorgen certification-matrix"
	certifiedLabel         = "CERTIFIED"
	uncertifiedLabel       = "COMMUNITY BUILD, UNCERTIFIED"
	uncertifiedWarning     = "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED."
)

//go:embed status.json
var embeddedStatusJSON []byte

// Status is the deliberately binary quality signal shown at connector
// inspection. Availability is independent: an uncertified connector remains
// reachable and receives the plain warning below.
type Status struct {
	Connector string `json:"connector"`
	Certified bool   `json:"certified"`
	Label     string `json:"label"`
	Warning   string `json:"warning,omitempty"`
}

type statusArtifact struct {
	SchemaVersion    int      `json:"schema_version"`
	GeneratedCommand string   `json:"generated_command"`
	Connectors       []Status `json:"connectors"`
}

var embeddedStatuses struct {
	once   sync.Once
	byName map[string]Status
	err    error
}

// StatusFor returns the generated status for a connector. A malformed or
// incomplete embedded projection is an explicit error rather than a silent
// green or an invented fallback status.
func StatusFor(connector string) (Status, error) {
	embeddedStatuses.once.Do(loadEmbeddedStatuses)
	if embeddedStatuses.err != nil {
		return Status{}, embeddedStatuses.err
	}
	status, ok := embeddedStatuses.byName[connector]
	if !ok {
		return Status{}, fmt.Errorf("generated certification status omits connector %q", connector)
	}
	return status, nil
}

// AllStatuses returns the stable generated set for catalog or website
// renderers. Callers receive a copy and cannot mutate the cached record.
func AllStatuses() ([]Status, error) {
	embeddedStatuses.once.Do(loadEmbeddedStatuses)
	if embeddedStatuses.err != nil {
		return nil, embeddedStatuses.err
	}
	statuses := make([]Status, 0, len(embeddedStatuses.byName))
	for _, status := range embeddedStatuses.byName {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Connector < statuses[j].Connector })
	return statuses, nil
}

func loadEmbeddedStatuses() {
	var artifact statusArtifact
	if err := decodeStatusJSON(embeddedStatusJSON, &artifact); err != nil {
		embeddedStatuses.err = fmt.Errorf("parse embedded certification status: %w", err)
		return
	}
	if artifact.SchemaVersion != statusSchemaVersion || artifact.GeneratedCommand != statusGeneratedCommand {
		embeddedStatuses.err = fmt.Errorf("embedded certification status has unsupported schema or generator")
		return
	}
	byName := make(map[string]Status, len(artifact.Connectors))
	for _, status := range artifact.Connectors {
		if strings.TrimSpace(status.Connector) == "" {
			embeddedStatuses.err = fmt.Errorf("embedded certification status contains an empty connector")
			return
		}
		if _, exists := byName[status.Connector]; exists {
			embeddedStatuses.err = fmt.Errorf("embedded certification status duplicates connector %q", status.Connector)
			return
		}
		if status.Certified {
			if status.Label != certifiedLabel || status.Warning != "" {
				embeddedStatuses.err = fmt.Errorf("embedded certified status %q is malformed", status.Connector)
				return
			}
		} else if status.Label != uncertifiedLabel || status.Warning != uncertifiedWarning {
			embeddedStatuses.err = fmt.Errorf("embedded uncertified status %q is malformed", status.Connector)
			return
		}
		byName[status.Connector] = status
	}
	if len(byName) == 0 {
		embeddedStatuses.err = fmt.Errorf("embedded certification status is empty")
		return
	}
	embeddedStatuses.byName = byName
}

func decodeStatusJSON(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err == io.EOF {
		return nil
	} else if err == nil {
		return fmt.Errorf("multiple JSON values")
	} else {
		return err
	}
}
