package engine

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	// SurfaceProvenanceLegacyUnverified preserves v1 compatibility while the
	// provider-artifact migration sweep updates existing inventories.
	SurfaceProvenanceLegacyUnverified = "legacy_unverified"
	// SurfaceProvenanceComplete means every v2 endpoint has resolved provider
	// artifact evidence and an endpoint-local source citation.
	SurfaceProvenanceComplete = "complete"
	// SurfaceProvenanceInvalid means a v2 inventory has invalid or incomplete
	// provenance evidence.
	SurfaceProvenanceInvalid = "invalid"
)

// SurfaceProvenanceValidation is the shared v2 provenance result consumed by
// conformance, connectorgen, and certification. It deliberately reports
// evidence only; it neither classifies endpoints nor changes CoveredBy.
type SurfaceProvenanceValidation struct {
	LedgerVersion  int
	ArtifactCount  int
	EndpointCount  int
	CitedEndpoints int
	Status         string
	Issues         []SurfaceProvenanceIssue
}

// SurfaceProvenanceIssue identifies one invalid provenance field without
// changing an endpoint's executable coverage binding.
type SurfaceProvenanceIssue struct {
	Subject string
	Message string
}

func (i SurfaceProvenanceIssue) Error() string {
	if i.Subject == "" {
		return fmt.Sprintf("api_surface provenance: %s", i.Message)
	}
	return fmt.Sprintf("api_surface provenance %s: %s", i.Subject, i.Message)
}

// ValidateSurfaceProvenance validates the v2 provider-artifact evidence
// contract. Ledgers before v2 intentionally remain loadable and certifiable as
// legacy_unverified so the staged migration does not orphan existing bundles.
func ValidateSurfaceProvenance(surface *APISurface) SurfaceProvenanceValidation {
	result := SurfaceProvenanceValidation{Status: SurfaceProvenanceLegacyUnverified}
	if surface == nil {
		return result
	}

	result.LedgerVersion = surface.OperationLedgerVersion
	result.ArtifactCount = len(surface.Artifacts)
	result.EndpointCount = len(surface.Endpoints)
	switch surface.OperationLedgerVersion {
	case 0, 1:
		return result
	case 2:
	default:
		result.addIssue("operation_ledger_version", fmt.Sprintf("%d is unsupported; expected 1 or 2", surface.OperationLedgerVersion))
		return result
	}

	result.Status = SurfaceProvenanceComplete
	artifactCounts := make(map[string]int, len(surface.Artifacts))
	for _, artifact := range surface.Artifacts {
		artifactCounts[artifact.ID]++
	}
	for i, artifact := range surface.Artifacts {
		subject := fmt.Sprintf("artifact %d", i)
		if strings.TrimSpace(artifact.ID) == "" {
			result.addIssue(subject, "id is required")
		} else if artifactCounts[artifact.ID] > 1 {
			result.addIssue(subject, fmt.Sprintf("id %q is duplicated", artifact.ID))
		}
		if !isAbsoluteHTTPSURL(artifact.URL) {
			result.addIssue(subject, "url must be an absolute HTTPS URL")
		}
		if !isISO8601FullDate(artifact.RetrievedAt) {
			result.addIssue(subject, "retrieved_at must be an ISO-8601 full-date")
		}
		if artifact.SHA256 != "" && !isSHA256Hex(artifact.SHA256) {
			result.addIssue(subject, "sha256 must be a 64-character hexadecimal digest when present")
		}
	}

	for i, endpoint := range surface.Endpoints {
		subject := fmt.Sprintf("endpoint %d (%s %s)", i, endpoint.Method, endpoint.Path)
		if endpoint.Provenance == nil {
			result.addIssue(subject, "provenance is required")
		} else {
			if strings.TrimSpace(endpoint.Provenance.SourceURL) == "" {
				result.addIssue(subject, "provenance.source_url is required")
			} else {
				result.CitedEndpoints++
				if !isAbsoluteHTTPSURL(endpoint.Provenance.SourceURL) {
					result.addIssue(subject, "provenance.source_url must be an absolute HTTPS URL")
				}
			}

			artifactID := endpoint.Provenance.Artifact
			if strings.TrimSpace(artifactID) == "" {
				result.addIssue(subject, "provenance.artifact is required")
			} else if count := artifactCounts[artifactID]; count != 1 {
				result.addIssue(subject, fmt.Sprintf("provenance.artifact %q resolves to %d artifacts", artifactID, count))
			}
		}
		if endpoint.Operation != nil && strings.TrimSpace(endpoint.Operation.SourceURL) != "" {
			result.addIssue(subject, "operation.source_url is v1-only; use provenance.source_url")
		}
	}

	return result
}

func (r *SurfaceProvenanceValidation) addIssue(subject, message string) {
	r.Status = SurfaceProvenanceInvalid
	r.Issues = append(r.Issues, SurfaceProvenanceIssue{Subject: subject, Message: message})
}

func isAbsoluteHTTPSURL(raw string) bool {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return false
	}
	parsed, err := url.Parse(raw)
	return err == nil && parsed.IsAbs() && parsed.Scheme == "https" && parsed.Host != "" && parsed.User == nil
}

func isISO8601FullDate(raw string) bool {
	if raw == "" || raw != strings.TrimSpace(raw) {
		return false
	}
	parsed, err := time.Parse(time.DateOnly, raw)
	return err == nil && parsed.Format(time.DateOnly) == raw
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
