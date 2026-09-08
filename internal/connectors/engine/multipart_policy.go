package engine

import (
	"fmt"
	"math"
)

// validateMultipartEnvelope preserves legacy envelope declarations while making
// an explicit metadata budget an exact, independently bounded wire profile.
func validateMultipartEnvelope(spec *MultipartSpec) error {
	if spec == nil {
		return nil
	}
	for i, part := range spec.Parts {
		if part.FilenameEncoding != "" {
			if part.Type != "file" {
				return fmt.Errorf("multipart part %d filename_encoding requires a file part", i)
			}
			if part.FilenameEncoding != "identity" && part.FilenameEncoding != "url_percent_utf8" {
				return fmt.Errorf("multipart part %d filename_encoding is unsupported", i)
			}
		}
	}
	if spec.MaxMetadataBytes == nil {
		return nil
	}
	if *spec.MaxMetadataBytes <= 0 {
		return fmt.Errorf("multipart max_metadata_bytes must be positive")
	}
	total := *spec.MaxMetadataBytes
	names := make(map[string]bool, len(spec.Parts))
	for i, part := range spec.Parts {
		if names[part.Name] {
			return fmt.Errorf("multipart part %d duplicates a name", i)
		}
		names[part.Name] = true
		if part.Type != "file" {
			continue
		}
		if part.MaxBytes <= 0 || part.MaxBytes > math.MaxInt64-total {
			return fmt.Errorf("multipart part %d file bound is invalid or overflows the envelope", i)
		}
		total += part.MaxBytes
	}
	if spec.MaxBytes != total {
		return fmt.Errorf("multipart max_bytes must equal file bounds plus max_metadata_bytes")
	}
	return nil
}
