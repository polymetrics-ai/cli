package contractv1

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
)

//go:embed testdata/fixtures/*.json
var fixtureFS embed.FS

// AcceptedSyntheticFixtures returns defensive copies of the fixture values
// consumed by the CLI fake-broker foundation.
func AcceptedSyntheticFixtures() SyntheticFixtures {
	fixtures, err := loadSyntheticFixtures()
	if err != nil {
		panic(err)
	}
	return fixtures
}

func loadSyntheticFixtures() (SyntheticFixtures, error) {
	fixtures := SyntheticFixtures{
		Compatibility:            Compatibility{},
		Organization:             Organization{},
		Workspace:                Workspace{},
		Environment:              Environment{},
		BrokerProfile:            BrokerProfile{},
		ConnectorConnection:      ConnectorConnection{},
		ExecutionPlanRequest:     ExecutionPlanRequest{},
		ExecutionPlan:            ExecutionPlan{},
		IncompatibleVersionError: IncompatibleContractVersionErrorResponse{},
	}

	var err error
	if fixtures.Compatibility, err = decodeFixture[Compatibility]("compatibility-response.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.Organization, err = decodeFixture[Organization]("organization.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.Workspace, err = decodeFixture[Workspace]("workspace.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.Environment, err = decodeFixture[Environment]("environment.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.BrokerProfile, err = decodeFixture[BrokerProfile]("broker-profile.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.ConnectorConnection, err = decodeFixture[ConnectorConnection]("connector-connection.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.ExecutionPlanRequest, err = decodeFixture[ExecutionPlanRequest]("execution-plan-request.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.ExecutionPlan, err = decodeFixture[ExecutionPlan]("execution-plan-response.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if fixtures.IncompatibleVersionError, err = decodeFixture[IncompatibleContractVersionErrorResponse]("incompatible-version-error.json"); err != nil {
		return SyntheticFixtures{}, err
	}
	if err := fixtures.validate(); err != nil {
		return SyntheticFixtures{}, err
	}
	return cloneFixtures(fixtures), nil
}

func decodeFixture[T any](name string) (T, error) {
	var value T
	payload, err := fixtureFS.ReadFile("testdata/fixtures/" + name)
	if err != nil {
		return value, fmt.Errorf("read fixture %s: %w", name, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode fixture %s: %w", name, err)
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return value, fmt.Errorf("decode fixture %s: trailing JSON values", name)
		}
		return value, fmt.Errorf("decode fixture %s trailer: %w", name, err)
	}
	return value, nil
}

func (fixtures SyntheticFixtures) validate() error {
	if fixtures.Compatibility.CurrentVersion != ContractVersion1 ||
		fixtures.Compatibility.MinimumClientVersion != ContractVersion1 ||
		len(fixtures.Compatibility.SupportedVersions) != 1 ||
		fixtures.Compatibility.SupportedVersions[0] != ContractVersion1 {
		return ErrInvalidErrorResponse
	}
	if err := fixtures.Compatibility.IncompatibleVersionRefusal.ValidateIncompatibleVersion(); err != nil {
		return err
	}
	if !fixtures.Organization.OrganizationID.IsValid() ||
		!fixtures.Workspace.WorkspaceID.IsValid() || !fixtures.Workspace.OrganizationID.IsValid() ||
		!fixtures.Environment.EnvironmentID.IsValid() || !fixtures.Environment.WorkspaceID.IsValid() ||
		!fixtures.Environment.OrganizationID.IsValid() {
		return ErrInvalidIdentityBoundary
	}
	if err := fixtures.BrokerProfile.Validate(); err != nil {
		return err
	}
	if err := fixtures.ConnectorConnection.Validate(); err != nil {
		return err
	}
	if err := fixtures.ExecutionPlanRequest.Validate(); err != nil {
		return err
	}
	if err := fixtures.ExecutionPlan.Validate(); err != nil {
		return err
	}
	if err := fixtures.IncompatibleVersionError.Validate(); err != nil {
		return err
	}
	return nil
}

func cloneFixtures(fixtures SyntheticFixtures) SyntheticFixtures {
	fixtures.Compatibility.SupportedVersions = append([]ContractVersion(nil), fixtures.Compatibility.SupportedVersions...)
	fixtures.Compatibility.IncompatibleVersionRefusal.SupportedVersions = append([]ContractVersion(nil), fixtures.Compatibility.IncompatibleVersionRefusal.SupportedVersions...)
	fixtures.BrokerProfile.AllowedConnectorKinds = append([]string(nil), fixtures.BrokerProfile.AllowedConnectorKinds...)
	fixtures.IncompatibleVersionError.Error.SupportedVersions = append([]ContractVersion(nil), fixtures.IncompatibleVersionError.Error.SupportedVersions...)
	return fixtures
}
