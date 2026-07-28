package pmbroker

import "fmt"

// ContractVersion is the negotiated PM Broker /v1 compatibility version.
type ContractVersion string

const (
	// ContractVersion1 is the only supported version in the initial PM Broker contract.
	ContractVersion1 ContractVersion = "1.0"
)

// ErrorCode is a client-safe PM Broker error code.
type ErrorCode string

const (
	// ErrorCodeIncompatibleContractVersion is returned with HTTP 426 for refused versions.
	ErrorCodeIncompatibleContractVersion ErrorCode = "incompatible_contract_version"
)

// IncompatibleContractVersionSafeError mirrors the PM Broker /v1 426 error payload.
type IncompatibleContractVersionSafeError struct {
	Code              ErrorCode         `json:"code"`
	Message           string            `json:"message"`
	CorrelationID     string            `json:"correlation_id"`
	SupportedVersions []ContractVersion `json:"supported_versions"`
}

// IncompatibleContractVersionError is a safe local seam for broker 426 refusals.
type IncompatibleContractVersionError struct {
	HTTPStatus int                                  `json:"http_status"`
	Error      IncompatibleContractVersionSafeError `json:"error"`

	// UnsafeRequestedVersion intentionally stays out of JSON and is not populated by
	// the constructor so callers cannot accidentally render untrusted version input.
	UnsafeRequestedVersion string `json:"-"`
}

type incompatibleContractVersionFailure struct {
	response IncompatibleContractVersionError
}

func (e incompatibleContractVersionFailure) Error() string {
	return fmt.Sprintf("pm broker contract version incompatible: supported=%v", e.response.Error.SupportedVersions)
}

// IncompatibleContractVersionResponse returns the safe 426 payload.
func (e incompatibleContractVersionFailure) IncompatibleContractVersionResponse() IncompatibleContractVersionError {
	return e.response
}

// NewIncompatibleContractVersionError returns exact PM Broker /v1 refusal semantics.
// TODO(pm-broker fake-client): replace direct fixture construction with the fake
// client contract package once the CLI fake-broker/contract-fixture lane lands.
func NewIncompatibleContractVersionError(_ string, correlationID string) IncompatibleContractVersionError {
	if correlationID == "" {
		correlationID = "corr_local_contract_version_refusal"
	}
	return IncompatibleContractVersionError{
		HTTPStatus: 426,
		Error: IncompatibleContractVersionSafeError{
			Code:              ErrorCodeIncompatibleContractVersion,
			Message:           "client PM Broker contract version is not compatible with this broker",
			CorrelationID:     correlationID,
			SupportedVersions: []ContractVersion{ContractVersion1},
		},
	}
}

// ValidateContractVersion refuses missing or incompatible contract versions.
func ValidateContractVersion(version string, correlationID string) error {
	if ContractVersion(version) == ContractVersion1 {
		return nil
	}
	response := NewIncompatibleContractVersionError(version, correlationID)
	return incompatibleContractVersionFailure{response: response}
}
