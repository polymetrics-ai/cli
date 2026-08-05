package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

// InitWithProtectorChainForTest exercises Init's brand-new-vault
// auto-selection fallback chain with caller-supplied (fake) protectors, so
// tests can assert fallback behavior without depending on real OS keychain
// availability.
func InitWithProtectorChainForTest(projectDir string, chain ...KeyProtector) (*Vault, error) {
	dir := filepath.Join(projectDir, "vault")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault directory: %w", err)
	}
	return initWithProtectorChain(dir, chain...)
}

// ReadProtectorMarkerForTest exposes the on-disk protector marker so tests
// can assert a vault stays pinned to whichever protector first initialized
// it.
func ReadProtectorMarkerForTest(projectDir string) (string, bool, error) {
	return readProtectorMarker(filepath.Join(projectDir, "vault"))
}
