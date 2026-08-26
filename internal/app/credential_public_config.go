package app

import (
	"fmt"
	"path/filepath"
	"strings"

	statestore "polymetrics.ai/internal/state"
)

// PublicCredentialConfiguration returns a selected credential's public
// configuration with a caller overlay applied. It deliberately opens neither
// the vault nor App's coordination, approval, transport, or protected state:
// a command can use it for no-I/O declaration preflight before credential
// resolution. Secret material remains inaccessible on this path.
func PublicCredentialConfiguration(root, connectorName, credentialName string, overlay map[string]string) (map[string]string, error) {
	config := cloneStringMap(overlay)
	if strings.TrimSpace(credentialName) == "" {
		return config, nil
	}

	credential, found, err := publicCredentialConfiguration(root, credentialName)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("credential %q not found", credentialName)
	}
	if connectorName != "" && connectorName != credential.Connector {
		return nil, fmt.Errorf("credential %q is for connector %q, not %q", credentialName, credential.Connector, connectorName)
	}
	config = cloneStringMap(credential.Config)
	for key, value := range overlay {
		config[key] = value
	}
	return config, nil
}

// publicCredentialState intentionally decodes only the credential metadata
// projection. It must not grow fields from App's protected state: this reader
// exists solely to perform public configuration admission before App opens.
type publicCredentialState struct {
	Credentials []CredentialMeta `json:"credentials"`
}

func publicCredentialConfiguration(root, name string) (CredentialMeta, bool, error) {
	if session := activeCertificationEphemeralSession(root); session != nil {
		credential, found := session.publicCredential(name)
		return credential, found, nil
	}
	if root == "" {
		root = "."
	}
	store := statestore.JSONStore[publicCredentialState]{
		Path: filepath.Join(root, ".polymetrics", "state", "state.json"),
	}
	loaded, err := store.LoadReadOnly()
	if err != nil {
		return CredentialMeta{}, false, fmt.Errorf("read public credential configuration: %w", err)
	}
	for _, credential := range loaded.Credentials {
		if credential.Name == name || credential.ID == name {
			credential.Config = cloneStringMap(credential.Config)
			return credential, true, nil
		}
	}
	return CredentialMeta{}, false, nil
}
