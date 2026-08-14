// Package certifications exposes generated proof-bearing certification statuses
// and community-build warnings to pm's user-facing surfaces. It has no
// dependency on connector execution, so reading a status never reaches
// credentials or a provider.
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

	"polymetrics.ai/internal/connectors/bundleregistry"
)

const (
	statusSchemaVersion    = 1
	statusGeneratedCommand = "go run ./cmd/connectorgen certification-matrix --all"
	certifiedLabel         = "CERTIFIED"
	uncertifiedLabel       = "COMMUNITY BUILD, UNCERTIFIED"
	uncertifiedWarning     = "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED."
)

//go:embed status.json
var embeddedStatusJSON []byte

// Status is the proof-bearing quality signal for an allowlisted connector or
// the community-build warning shown for one outside that scope. Availability is
// independent in either case.
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

// StatusFor returns a proof-bearing status for an allowlisted connector. A
// connector outside that explicit certification scope has no pass or fail
// claim, so it retains the existing visible community-build warning rather
// than being omitted from inspection.
func StatusFor(connector string) (Status, error) {
	return statusFor(connector, func() bool {
		_, registered := bundleregistry.New().Get(connector)
		return registered
	})
}

// StatusForRegistered returns a status using the caller's connector-registration result.
func StatusForRegistered(connector string, registered bool) (Status, error) {
	return statusFor(connector, func() bool { return registered })
}

func statusFor(connector string, isRegistered func() bool) (Status, error) {
	embeddedStatuses.once.Do(loadEmbeddedStatuses)
	if embeddedStatuses.err != nil {
		return Status{}, embeddedStatuses.err
	}
	status, ok := embeddedStatuses.byName[connector]
	if !ok {
		if !isRegistered() {
			return Status{}, fmt.Errorf("generated certification status omits connector %q", connector)
		}
		return Status{
			Connector: connector,
			Label:     uncertifiedLabel,
			Warning:   uncertifiedWarning,
		}, nil
	}
	return status, nil
}

// AllStatuses returns the stable generated allowlist status set for catalog or
// website renderers. Callers receive a copy and cannot mutate the cached record.
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
