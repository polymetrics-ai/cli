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
	SchemaVersion      int      `json:"schema_version"`
	GeneratedCommand   string   `json:"generated_command"`
	CertificationScope []string `json:"certification_scope"`
	Connectors         []Status `json:"connectors"`
}

var embeddedStatuses struct {
	once   sync.Once
	byName map[string]Status
	scope  map[string]bool
	err    error
}

// StatusFor returns the generated proof-bearing status for an allowlisted connector.
func StatusFor(connector string) (Status, error) {
	return statusFor(connector, false)
}

// StatusForRegistered returns the generated status for an allowlisted connector.
// For a registered connector outside the certification scope, it returns the
// community-build warning instead of inventing a certification claim.
func StatusForRegistered(connector string, registered bool) (Status, error) {
	return statusFor(connector, registered)
}

func statusFor(connector string, registered bool) (Status, error) {
	embeddedStatuses.once.Do(loadEmbeddedStatuses)
	if embeddedStatuses.err != nil {
		return Status{}, embeddedStatuses.err
	}
	status, ok := embeddedStatuses.byName[connector]
	if !ok {
		if !registered || embeddedStatuses.scope[connector] {
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
	byName, scope, err := loadStatusArtifact(embeddedStatusJSON)
	if err != nil {
		embeddedStatuses.err = err
		return
	}
	embeddedStatuses.byName = byName
	embeddedStatuses.scope = scope
}

func loadStatusArtifact(raw []byte) (map[string]Status, map[string]bool, error) {
	var artifact statusArtifact
	if err := decodeStatusJSON(raw, &artifact); err != nil {
		return nil, nil, fmt.Errorf("parse embedded certification status: %w", err)
	}
	if artifact.SchemaVersion != statusSchemaVersion || artifact.GeneratedCommand != statusGeneratedCommand {
		return nil, nil, fmt.Errorf("embedded certification status has unsupported schema or generator")
	}
	byName := make(map[string]Status, len(artifact.Connectors))
	for _, status := range artifact.Connectors {
		if strings.TrimSpace(status.Connector) == "" {
			return nil, nil, fmt.Errorf("embedded certification status contains an empty connector")
		}
		if _, exists := byName[status.Connector]; exists {
			return nil, nil, fmt.Errorf("embedded certification status duplicates connector %q", status.Connector)
		}
		if status.Certified {
			if status.Label != certifiedLabel || status.Warning != "" {
				return nil, nil, fmt.Errorf("embedded certified status %q is malformed", status.Connector)
			}
		} else if status.Label != uncertifiedLabel || status.Warning != uncertifiedWarning {
			return nil, nil, fmt.Errorf("embedded uncertified status %q is malformed", status.Connector)
		}
		byName[status.Connector] = status
	}
	if len(byName) == 0 {
		return nil, nil, fmt.Errorf("embedded certification status is empty")
	}
	scope := make(map[string]bool, len(artifact.CertificationScope))
	for _, connector := range artifact.CertificationScope {
		if strings.TrimSpace(connector) == "" {
			return nil, nil, fmt.Errorf("embedded certification status contains an empty scope connector")
		}
		if scope[connector] {
			return nil, nil, fmt.Errorf("embedded certification status duplicates scope connector %q", connector)
		}
		scope[connector] = true
		if _, exists := byName[connector]; !exists {
			return nil, nil, fmt.Errorf("generated certification status omits connector %q", connector)
		}
	}
	if len(scope) == 0 {
		return nil, nil, fmt.Errorf("embedded certification status scope is empty")
	}
	for _, status := range artifact.Connectors {
		if !scope[status.Connector] {
			return nil, nil, fmt.Errorf("embedded certification status scope omits connector %q", status.Connector)
		}
	}
	return byName, scope, nil
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
