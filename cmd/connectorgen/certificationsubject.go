package main

import (
	"bytes"
	"crypto/sha256"
	"debug/buildinfo"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	certificationSubjectSchemaVersion = 1
	certificationSubjectProofProtocol = "foundation-certification-proof-v1"
	certificationSubjectArtifactPath  = "internal/connectors/certifications/current-subject.json"
)

// certificationSubjectComponents is the complete immutable identity of the
// executable contract exercised by a live certification. Each value is a
// digest collected at the producer boundary; none is an operator assertion.
type certificationSubjectComponents struct {
	PMBinarySHA256          string
	PMBuildSHA256           string
	DeclarationsSHA256      string
	SourceProjectionSHA256  string
	CLICommandMappingSHA256 string
	RelevantConfigSHA256    string
	ProofProtocol           string
}

// certificationSubject is persisted in proof-bearing evidence and in the
// current-subject artifact. The fingerprint is derived from every component,
// then each component is retained so a collision or partial rewrite cannot
// turn an old proof into current evidence.
type certificationSubject struct {
	SchemaVersion           int    `json:"schema_version"`
	Fingerprint             string `json:"fingerprint"`
	PMBinarySHA256          string `json:"pm_binary_sha256"`
	PMBuildSHA256           string `json:"pm_build_sha256"`
	DeclarationsSHA256      string `json:"declarations_sha256"`
	SourceProjectionSHA256  string `json:"source_projection_sha256"`
	CLICommandMappingSHA256 string `json:"cli_command_mapping_sha256"`
	RelevantConfigSHA256    string `json:"relevant_config_sha256"`
	ProofProtocol           string `json:"proof_protocol"`
}

func newCertificationSubject(components certificationSubjectComponents) (certificationSubject, error) {
	for _, digest := range []string{
		components.PMBinarySHA256,
		components.PMBuildSHA256,
		components.DeclarationsSHA256,
		components.SourceProjectionSHA256,
		components.CLICommandMappingSHA256,
		components.RelevantConfigSHA256,
	} {
		if !evidenceSHA256.MatchString(digest) {
			return certificationSubject{}, fmt.Errorf("certification subject component must be a lowercase SHA-256 digest")
		}
	}
	if components.ProofProtocol == "" {
		return certificationSubject{}, fmt.Errorf("certification subject proof protocol is required")
	}
	subject := certificationSubject{
		SchemaVersion:           certificationSubjectSchemaVersion,
		PMBinarySHA256:          components.PMBinarySHA256,
		PMBuildSHA256:           components.PMBuildSHA256,
		DeclarationsSHA256:      components.DeclarationsSHA256,
		SourceProjectionSHA256:  components.SourceProjectionSHA256,
		CLICommandMappingSHA256: components.CLICommandMappingSHA256,
		RelevantConfigSHA256:    components.RelevantConfigSHA256,
		ProofProtocol:           components.ProofProtocol,
	}
	payload, err := json.Marshal(subject)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("encode certification subject: %w", err)
	}
	digest := sha256.Sum256(payload)
	subject.Fingerprint = hex.EncodeToString(digest[:])
	return subject, nil
}

func (subject certificationSubject) Components() certificationSubjectComponents {
	return certificationSubjectComponents{
		PMBinarySHA256:          subject.PMBinarySHA256,
		PMBuildSHA256:           subject.PMBuildSHA256,
		DeclarationsSHA256:      subject.DeclarationsSHA256,
		SourceProjectionSHA256:  subject.SourceProjectionSHA256,
		CLICommandMappingSHA256: subject.CLICommandMappingSHA256,
		RelevantConfigSHA256:    subject.RelevantConfigSHA256,
		ProofProtocol:           subject.ProofProtocol,
	}
}

func validateCertificationSubject(subject certificationSubject) error {
	if subject.SchemaVersion != certificationSubjectSchemaVersion {
		return fmt.Errorf("certification subject schema_version %d is unsupported", subject.SchemaVersion)
	}
	expected, err := newCertificationSubject(subject.Components())
	if err != nil {
		return err
	}
	if subject.Fingerprint != expected.Fingerprint {
		return fmt.Errorf("certification subject fingerprint does not match its components")
	}
	return nil
}

func certificationSubjectsEqual(left, right certificationSubject) bool {
	return left == right && validateCertificationSubject(left) == nil
}

// classifyEvidenceForCertificationSubject retains every nonmatching accepted
// record as historical while preventing it from driving live_tested. Legacy
// records intentionally have no subject and therefore remain historical.
func classifyEvidenceForCertificationSubject(evidence []acceptedEvidence, current certificationSubject) (live, historical []acceptedEvidence) {
	for _, item := range evidence {
		if certificationSubjectsEqual(item.Proof.CertificationSubject, current) {
			live = append(live, item)
			continue
		}
		historical = append(historical, item)
	}
	return live, historical
}

type currentCertificationSubjectArtifact struct {
	SchemaVersion    int                  `json:"schema_version"`
	GeneratedCommand string               `json:"generated_command"`
	Subject          certificationSubject `json:"subject"`
}

func runCertificationSubject(args []string, stdout, stderr io.Writer) int {
	root := "."
	binary := ""
	check := false
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--check":
			check = true
		case arg == "--pm":
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				logln(stderr, "connectorgen certification-subject: --pm requires a built pm binary")
				return 2
			}
			index++
			binary = args[index]
		case strings.HasPrefix(arg, "--"):
			logf(stderr, "connectorgen certification-subject: unknown flag %q\n", arg)
			return 2
		case root == ".":
			root = arg
		default:
			logf(stderr, "connectorgen certification-subject: unexpected extra argument %q\n", arg)
			return 2
		}
	}
	if binary == "" {
		logln(stderr, "connectorgen certification-subject: --pm is required")
		return 2
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		logf(stderr, "connectorgen certification-subject: resolve repository root: %v\n", err)
		return 1
	}
	subject, err := certificationSubjectForBinary(absRoot, binary)
	if err != nil {
		logf(stderr, "connectorgen certification-subject: %v\n", err)
		return 1
	}
	artifact := currentCertificationSubjectArtifact{
		SchemaVersion:    certificationSubjectSchemaVersion,
		GeneratedCommand: "go run ./cmd/connectorgen certification-subject --pm ./pm",
		Subject:          subject,
	}
	payload, err := marshalGeneratedJSON(artifact)
	if err != nil {
		logf(stderr, "connectorgen certification-subject: render subject: %v\n", err)
		return 1
	}
	path := filepath.Join(absRoot, filepath.FromSlash(certificationSubjectArtifactPath))
	if check {
		committed, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(committed, payload) {
			logf(stderr, "connectorgen certification-subject: current subject is stale; run `go run ./cmd/connectorgen certification-subject --pm ./pm`\n")
			return 1
		}
		logf(stdout, "connectorgen certification-subject: %s is current\n", certificationSubjectArtifactPath)
		return 0
	}
	if err := writeGeneratedArtifact(path, payload); err != nil {
		logf(stderr, "connectorgen certification-subject: write subject: %v\n", err)
		return 1
	}
	logf(stdout, "connectorgen certification-subject: wrote %s\n", certificationSubjectArtifactPath)
	return 0
}

func certificationSubjectForBinary(root, binary string) (certificationSubject, error) {
	binaryPath, err := filepath.Abs(binary)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("resolve pm binary: %w", err)
	}
	binaryBytes, err := os.ReadFile(binaryPath)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("read pm binary: %w", err)
	}
	if len(binaryBytes) == 0 {
		return certificationSubject{}, fmt.Errorf("pm binary is empty")
	}
	binaryDigest := sha256.Sum256(binaryBytes)
	buildDigest, err := certificationBuildDigest(binaryPath)
	if err != nil {
		return certificationSubject{}, err
	}
	declarationsDigest, err := digestCertificationFiles(root, isCertificationDeclarationFile)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("digest declarations: %w", err)
	}
	sourceDigest, err := digestCertificationFiles(root, isCertificationSourceProjectionFile)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("digest source projection: %w", err)
	}
	cliDigest, err := digestCertificationFiles(root, isCertificationCLICommandMappingFile)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("digest CLI command mapping: %w", err)
	}
	configDigest, err := digestCertificationFiles(root, isCertificationRelevantConfigFile)
	if err != nil {
		return certificationSubject{}, fmt.Errorf("digest relevant certification config: %w", err)
	}
	return newCertificationSubject(certificationSubjectComponents{
		PMBinarySHA256:          hex.EncodeToString(binaryDigest[:]),
		PMBuildSHA256:           buildDigest,
		DeclarationsSHA256:      declarationsDigest,
		SourceProjectionSHA256:  sourceDigest,
		CLICommandMappingSHA256: cliDigest,
		RelevantConfigSHA256:    configDigest,
		ProofProtocol:           certificationSubjectProofProtocol,
	})
}

func certificationBuildDigest(binary string) (string, error) {
	info, err := buildinfo.ReadFile(binary)
	if err != nil {
		return "", fmt.Errorf("read pm build identity: %w", err)
	}
	type setting struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	settings := make([]setting, 0, len(info.Settings))
	for _, value := range info.Settings {
		settings = append(settings, setting{Key: value.Key, Value: value.Value})
	}
	sort.Slice(settings, func(i, j int) bool { return settings[i].Key < settings[j].Key })
	payload, err := json.Marshal(struct {
		GoVersion string    `json:"go_version"`
		Path      string    `json:"path"`
		MainPath  string    `json:"main_path"`
		MainVer   string    `json:"main_version"`
		MainSum   string    `json:"main_sum"`
		Settings  []setting `json:"settings"`
	}{
		GoVersion: info.GoVersion, Path: info.Path, MainPath: info.Main.Path,
		MainVer: info.Main.Version, MainSum: info.Main.Sum, Settings: settings,
	})
	if err != nil {
		return "", fmt.Errorf("encode pm build identity: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func digestCertificationFiles(root string, include func(string) bool) (string, error) {
	paths := make([]string, 0)
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if include(relative) {
			paths = append(paths, relative)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return "", fmt.Errorf("subject component has no inputs")
	}
	hash := sha256.New()
	for _, relative := range paths {
		contents, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			return "", err
		}
		contentDigest := sha256.Sum256(contents)
		if _, err := io.WriteString(hash, relative+"\x00"+hex.EncodeToString(contentDigest[:])+"\n"); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isConnectorDefinitionPath(path string) bool {
	return strings.HasPrefix(path, "internal/connectors/defs/") && strings.HasSuffix(path, ".json")
}

func isCertificationDeclarationFile(path string) bool {
	if !isConnectorDefinitionPath(path) || strings.Contains(path, "/sources/") {
		return false
	}
	base := filepath.Base(path)
	return base != "cli_surface.json" && base != "api_surface.json" && !strings.HasPrefix(base, "certification-") && base != "certification.json"
}

func isCertificationSourceProjectionFile(path string) bool {
	return isConnectorDefinitionPath(path) && strings.Contains(path, "/sources/")
}

func isCertificationCLICommandMappingFile(path string) bool {
	return isConnectorDefinitionPath(path) && (filepath.Base(path) == "cli_surface.json" || filepath.Base(path) == "api_surface.json")
}

func isCertificationRelevantConfigFile(path string) bool {
	if path == "go.mod" || path == "go.sum" {
		return true
	}
	if !isConnectorDefinitionPath(path) {
		return false
	}
	base := filepath.Base(path)
	return base == "certification.json" || base == "certification-observations.json" || base == "rate_limits.json" || base == "sync_transport.json" || base == "changefeed.json"
}

func loadCurrentCertificationSubject(root string) (certificationSubject, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(certificationSubjectArtifactPath)))
	if err != nil {
		return certificationSubject{}, fmt.Errorf("read current certification subject: %w", err)
	}
	var artifact currentCertificationSubjectArtifact
	if err := decodeStrictJSON(raw, &artifact); err != nil {
		return certificationSubject{}, fmt.Errorf("parse current certification subject: %w", err)
	}
	if artifact.SchemaVersion != certificationSubjectSchemaVersion || artifact.GeneratedCommand != "go run ./cmd/connectorgen certification-subject --pm ./pm" {
		return certificationSubject{}, fmt.Errorf("current certification subject artifact is invalid")
	}
	if err := validateCertificationSubject(artifact.Subject); err != nil {
		return certificationSubject{}, fmt.Errorf("current certification subject: %w", err)
	}
	return artifact.Subject, nil
}
