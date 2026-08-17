package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors/certify"
)

const externalRuntimeObservationSchemaVersion = 1

// externalRuntimeObservation is intentionally secret-safe evidence written by
// the child itself. A snapshot is taken only after the child has resolved the
// credential into memory, so its process-list, argv, and temporary-state
// observations do not depend on a parent winning a scheduling race.
type externalRuntimeObservation struct {
	SchemaVersion      int                                  `json:"schema_version"`
	ProcessList        externalRuntimeObservationProcess    `json:"process_list"`
	Argv               externalRuntimeObservationArgv       `json:"argv"`
	ProjectArtifacts   externalRuntimeObservationLocation   `json:"project_artifacts"`
	TemporaryArtifacts []externalRuntimeObservationLocation `json:"temporary_artifacts"`
}

type externalRuntimeObservationProcess struct {
	Command            string `json:"command"`
	ContainsCredential bool   `json:"contains_credential"`
}

type externalRuntimeObservationArgv struct {
	Values             []string `json:"values"`
	ContainsCredential bool     `json:"contains_credential"`
}

type externalRuntimeObservationLocation struct {
	Path               string `json:"path"`
	FilesScanned       int    `json:"files_scanned"`
	ContainsCredential bool   `json:"contains_credential"`
}

func writeExternalRuntimeObservation(path, projectRoot string, input certify.RuntimeObservationInput) error {
	path, err := externalRuntimeObservationPath(projectRoot, path)
	if err != nil {
		return err
	}
	command, err := exec.Command("ps", "-p", strconv.Itoa(os.Getpid()), "-o", "command=").Output()
	if err != nil {
		return fmt.Errorf("inspect child process list: %w", err)
	}
	processCommand, processContainsCredential := redactRuntimeObservationValue(strings.TrimSpace(string(command)), input.SecretValues)
	argv := make([]string, 0, len(os.Args))
	argvContainsCredential := false
	for _, arg := range os.Args {
		safeArg, containsCredential := redactRuntimeObservationValue(arg, input.SecretValues)
		argv = append(argv, safeArg)
		argvContainsCredential = argvContainsCredential || containsCredential
	}

	projectArtifacts, err := inspectExternalRuntimeObservationLocation(projectRoot, input.SecretValues)
	if err != nil {
		return fmt.Errorf("inspect child project artifacts: %w", err)
	}
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate child executable: %w", err)
	}
	temporaryRoots := uniqueRuntimeObservationPaths(input.Workdir, filepath.Dir(executablePath))
	temporaryArtifacts := make([]externalRuntimeObservationLocation, 0, len(temporaryRoots))
	for _, root := range temporaryRoots {
		location, err := inspectExternalRuntimeObservationLocation(root, input.SecretValues)
		if err != nil {
			return fmt.Errorf("inspect child temporary artifacts: %w", err)
		}
		temporaryArtifacts = append(temporaryArtifacts, location)
	}

	observation := externalRuntimeObservation{
		SchemaVersion: externalRuntimeObservationSchemaVersion,
		ProcessList: externalRuntimeObservationProcess{
			Command:            processCommand,
			ContainsCredential: processContainsCredential,
		},
		Argv: externalRuntimeObservationArgv{
			Values:             argv,
			ContainsCredential: argvContainsCredential,
		},
		ProjectArtifacts:   projectArtifacts,
		TemporaryArtifacts: temporaryArtifacts,
	}
	raw, err := json.MarshalIndent(observation, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal child runtime observation: %w", err)
	}
	if len(certify.ScanForSecrets(string(raw), input.SecretValues)) != 0 {
		return errors.New("child runtime observation would expose credential material")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create child runtime observation directory: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write child runtime observation: %w", err)
	}
	return nil
}

func loadExternalRuntimeObservation(path string) (externalRuntimeObservation, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return externalRuntimeObservation{}, err
	}
	var observation externalRuntimeObservation
	if err := json.Unmarshal(raw, &observation); err != nil {
		return externalRuntimeObservation{}, err
	}
	return observation, nil
}

func externalRuntimeObservationPath(projectRoot, path string) (string, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	observationPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve runtime observation path: %w", err)
	}
	rel, err := filepath.Rel(root, observationPath)
	if err != nil {
		return "", fmt.Errorf("relativize runtime observation path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", errors.New("runtime observation path must stay inside the project root")
	}
	return observationPath, nil
}

func inspectExternalRuntimeObservationLocation(path string, secrets []string) (externalRuntimeObservationLocation, error) {
	safePath, pathContainsCredential := redactRuntimeObservationValue(path, secrets)
	location := externalRuntimeObservationLocation{
		Path:               safePath,
		ContainsCredential: pathContainsCredential,
	}
	err := filepath.WalkDir(path, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}
		location.FilesScanned++
		containsCredential, err := externalRuntimeObservationFileContainsCredential(filePath, secrets)
		if err != nil {
			return err
		}
		location.ContainsCredential = location.ContainsCredential || containsCredential
		return nil
	})
	if err != nil {
		return externalRuntimeObservationLocation{}, err
	}
	return location, nil
}

func externalRuntimeObservationFileContainsCredential(path string, secrets []string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	maxSecretLength := 0
	for _, secret := range secrets {
		if len(secret) > maxSecretLength {
			maxSecretLength = len(secret)
		}
	}
	const chunkBytes = 64 << 10
	buffer := make([]byte, chunkBytes)
	carry := make([]byte, 0, max(maxSecretLength-1, 0))
	for {
		read, readErr := file.Read(buffer)
		if read > 0 {
			window := append(append([]byte(nil), carry...), buffer[:read]...)
			if containsExternalRuntimeCredential(window, secrets) {
				return true, nil
			}
			carry = carry[:0]
			if overlap := maxSecretLength - 1; overlap > 0 && len(window) > overlap {
				carry = append(carry, window[len(window)-overlap:]...)
			} else if overlap > 0 {
				carry = append(carry, window...)
			}
		}
		if errors.Is(readErr, io.EOF) {
			return false, nil
		}
		if readErr != nil {
			return false, readErr
		}
	}
}

func containsExternalRuntimeCredential(payload []byte, secrets []string) bool {
	for _, secret := range secrets {
		if secret != "" && bytes.Contains(payload, []byte(secret)) {
			return true
		}
	}
	return false
}

func redactRuntimeObservationValue(value string, secrets []string) (string, bool) {
	containsCredential := false
	for _, secret := range secrets {
		if secret == "" || !strings.Contains(value, secret) {
			continue
		}
		containsCredential = true
		value = strings.ReplaceAll(value, secret, "<credential>")
	}
	return value, containsCredential
}

func uniqueRuntimeObservationPaths(paths ...string) []string {
	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		unique = append(unique, clean)
	}
	return unique
}
