package gsd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	normalizedRegistrySchemaVersion  = "shepherd.gsd-unit-registry/v1"
	officialRegistryModulePath       = "dist/resources/extensions/gsd/unit-registry.js"
	officialRegistryDependencyPath   = "dist/resources/extensions/shared/browser-contract.js"
	officialRegistrySHA256           = "b53beccf9185793a835b940ace72a9dcb5bf80f1c31a39ff72bcc18e9a342fc1"
	officialRegistryDependencySHA256 = "37875f964dc0455173da8f7a1d94855c15953ecb4bb2304c33e1f8872dc00090"
	officialLoaderSHA256             = "3bd187abede338cfd442d7866722235783a80f5734184731d9e051508271f28b"
	officialPatchedHeadlessSHA256    = "07de6ec472df1b610f6aaea8d453dbd451b73a8b3da34da8502289e1c5b6384f"
	maxNormalizedRegistryBytes       = 256 * 1024
	maxRegistryStderrBytes           = 64 * 1024
)

var expectedOfficialUnitTypes = []string{
	"complete-milestone", "complete-slice", "discuss-milestone", "discuss-project",
	"discuss-requirements", "discuss-slice", "execute-task", "execute-task-simple",
	"gate-evaluate", "plan-milestone", "plan-slice", "reactive-execute", "reassess-roadmap",
	"refine-slice", "replan-slice", "research-decision", "research-milestone",
	"research-project", "research-slice", "rewrite-docs", "run-uat", "validate-milestone",
	"workflow-preferences",
}

var governedPhaseRoles = map[string]ModelRole{
	"research":         ModelRoleCoordinator,
	"planning":         ModelRoleCoordinator,
	"discuss":          ModelRoleCoordinator,
	"completion":       ModelRoleCoordinator,
	"validation":       ModelRoleCoordinator,
	"uat":              ModelRoleCoordinator,
	"execution":        ModelRoleImplementation,
	"execution_simple": ModelRoleImplementation,
}

type ModelRole string

const (
	ModelRoleCoordinator    ModelRole = "coordinator"
	ModelRoleImplementation ModelRole = "implementation"
)

type UnitMetadata struct {
	UnitType              string
	Kind                  string
	ScopeClass            string
	PhaseChain            []string
	PromptTemplates       []string
	AllowedGSDTools       []string
	RequiredWorkflowTools []string
	ForbiddenGSDTools     map[string]string
}

type UnitRegistry struct {
	Units map[string]UnitMetadata
}

type normalizedRegistryDocument struct {
	SchemaVersion    string                   `json:"schemaVersion"`
	Source           normalizedRegistrySource `json:"source"`
	Units            []normalizedUnitMetadata `json:"units"`
	ExcludedSidecars []string                 `json:"excludedSidecars"`
}

type normalizedRegistrySource struct {
	PackageName      string `json:"packageName"`
	PackageVersion   string `json:"packageVersion"`
	ModulePath       string `json:"modulePath"`
	ModuleSHA256     string `json:"moduleSHA256"`
	DependencyPath   string `json:"dependencyPath"`
	DependencySHA256 string `json:"dependencySHA256"`
	LoaderPath       string `json:"loaderPath"`
	LoaderSHA256     string `json:"loaderSHA256"`
	HeadlessSHA256   string `json:"headlessSHA256"`
	ComposerSHA256   string `json:"composerSHA256"`
}

type normalizedUnitMetadata struct {
	UnitType              string                    `json:"unitType"`
	Kind                  string                    `json:"kind"`
	ScopeClass            string                    `json:"scopeClass"`
	PhaseChain            []string                  `json:"phaseChain"`
	PromptTemplates       []string                  `json:"promptTemplates"`
	AllowedGSDTools       []string                  `json:"allowedGsdTools"`
	RequiredWorkflowTools []string                  `json:"requiredWorkflowTools"`
	ForbiddenGSDTools     []normalizedForbiddenTool `json:"forbiddenGsdTools"`
}

type normalizedForbiddenTool struct {
	Tool   string `json:"tool"`
	Reason string `json:"reason"`
}

type SidecarUnitPolicy struct {
	UnitType  string
	Supported bool
	Reason    string
}

type SidecarPolicy struct {
	Version string
	Units   map[string]SidecarUnitPolicy
}

func PinnedSidecarPolicy() SidecarPolicy {
	return SidecarPolicy{
		Version: "shepherd.gsd-sidecars/v1",
		Units: map[string]SidecarUnitPolicy{
			"quick-task":      {UnitType: "quick-task", Supported: false, Reason: "official GSD 1.11.0 declares no phase or tool contract"},
			"triage-captures": {UnitType: "triage-captures", Supported: false, Reason: "official GSD 1.11.0 declares no phase or tool contract"},
		},
	}
}

func (p SidecarPolicy) Lookup(unitType string) (SidecarUnitPolicy, bool) {
	entry, ok := p.Units[unitType]
	return entry, ok
}

type registryLoadOptions struct {
	expectedRegistrySHA256   string
	expectedDependencySHA256 string
	expectedLoaderSHA256     string
	expectedHeadlessSHA256   string
	expectedComposerSHA256   string
	expectedNodeSHA256       string
	timeout                  time.Duration
}

func defaultRegistryLoadOptions() registryLoadOptions {
	qualification := qualifiedHostRuntimes[runtime.GOOS+"/"+runtime.GOARCH]
	return registryLoadOptions{
		expectedRegistrySHA256:   officialRegistrySHA256,
		expectedDependencySHA256: officialRegistryDependencySHA256,
		expectedLoaderSHA256:     officialLoaderSHA256,
		expectedHeadlessSHA256:   officialPatchedHeadlessSHA256,
		expectedComposerSHA256:   officialPatchedComposerSHA256,
		expectedNodeSHA256:       qualification.NodeSHA256,
		timeout:                  5 * time.Second,
	}
}

// LoadPinnedUnitRegistry exports a normalized metadata document from the exact
// validated official runtime. It never evaluates project or candidate files.
func LoadPinnedUnitRegistry(ctx context.Context, command []string, gsdHome, expectedVersion string) (UnitRegistry, error) {
	return loadPinnedUnitRegistryWithOptions(ctx, command, gsdHome, expectedVersion, defaultRegistryLoadOptions())
}

func loadPinnedUnitRegistryWithOptions(ctx context.Context, command []string, gsdHome, expectedVersion string, options registryLoadOptions) (UnitRegistry, error) {
	if expectedVersion != "1.11.0" {
		return UnitRegistry{}, registryMismatch("unit registry compatibility is qualified only for GSD 1.11.0")
	}
	if ctx == nil {
		return UnitRegistry{}, registryMismatch("registry export requires a context")
	}
	if options.timeout <= 0 || options.timeout > 30*time.Second || !validSHA256(options.expectedRegistrySHA256) || !validSHA256(options.expectedDependencySHA256) || !validSHA256(options.expectedLoaderSHA256) || !validSHA256(options.expectedHeadlessSHA256) || !validSHA256(options.expectedComposerSHA256) || !validSHA256(options.expectedNodeSHA256) {
		return UnitRegistry{}, registryMismatch("registry loader policy is invalid")
	}
	runtime, err := resolvePinnedRegistryRuntime(command, expectedVersion, options)
	if err != nil {
		return UnitRegistry{}, err
	}
	if err := validateActiveRegistrySource(gsdHome, options.expectedRegistrySHA256); err != nil {
		return UnitRegistry{}, err
	}

	exportCtx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()
	moduleURL, dependencyURL, err := verifiedRegistryDataURLs(runtime.registryPath, runtime.dependencyPath, options)
	if err != nil {
		return UnitRegistry{}, err
	}
	cmd := exec.CommandContext(exportCtx, runtime.nodePath,
		"--input-type=module", "--eval", normalizedRegistryExporter,
		moduleURL, expectedVersion, options.expectedRegistrySHA256,
		dependencyURL, options.expectedDependencySHA256,
		"verified-bytes", options.expectedLoaderSHA256,
		options.expectedHeadlessSHA256, options.expectedComposerSHA256,
	)
	cmd.Dir = runtime.packageRoot
	cmd.Env = registryExporterEnvironment()
	configureProcessTree(cmd)
	var stdout, stderr boundedBuffer
	stdout.limit = maxNormalizedRegistryBytes + 1
	stderr.limit = maxRegistryStderrBytes
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if runErr := cmd.Run(); runErr != nil {
		if exportCtx.Err() != nil {
			return UnitRegistry{}, registryMismatch("official registry exporter was cancelled or timed out")
		}
		return UnitRegistry{}, registryMismatch("official registry exporter failed: " + boundedDiagnostic(stderr.String()))
	}
	if stdout.Truncated() || stderr.Truncated() {
		return UnitRegistry{}, registryMismatch("official registry exporter exceeded its output bound")
	}
	return decodeNormalizedUnitRegistryWithSource([]byte(stdout.String()), expectedVersion, options.expectedRegistrySHA256, options.expectedDependencySHA256, options.expectedLoaderSHA256, options.expectedHeadlessSHA256, options.expectedComposerSHA256)
}

// LoadPinnedContainerUnitRegistry runs the same static exporter inside the
// already-pinned image without mounting project, credential, or state paths.
func LoadPinnedContainerUnitRegistry(ctx context.Context, config ContainerConfig, expectedVersion string) (UnitRegistry, error) {
	if expectedVersion != "1.11.0" || ctx == nil || !validImmutableImageID(config.Image) {
		return UnitRegistry{}, registryMismatch("container registry export requires GSD 1.11.0 at an immutable image ID")
	}
	// No complete image digest has yet been human-qualified for Slice D. Keep
	// Podman assets intact, but fail closed rather than trusting a self-reported
	// version and a handful of files from an arbitrary immutable local image.
	return UnitRegistry{}, registryMismatch("container execution has no approved complete-image digest for GSD 1.11.0")

}

type pinnedRegistryRuntime struct {
	nodePath       string
	packageRoot    string
	registryPath   string
	dependencyPath string
}

func resolvePinnedRegistryRuntime(command []string, expectedVersion string, options registryLoadOptions) (pinnedRegistryRuntime, error) {
	if len(command) != 2 || !filepath.IsAbs(command[1]) || filepath.Clean(command[1]) != command[1] {
		return pinnedRegistryRuntime{}, registryMismatch("GSD command must be node plus a clean absolute loader path")
	}
	nodePath, err := resolveQualifiedNode(command, options.expectedNodeSHA256)
	if err != nil {
		return pinnedRegistryRuntime{}, err
	}
	loader := command[1]
	if filepath.Base(loader) != "loader.js" || filepath.Base(filepath.Dir(loader)) != "dist" {
		return pinnedRegistryRuntime{}, registryMismatch("GSD command does not target the packaged loader")
	}
	packageRoot := filepath.Dir(filepath.Dir(loader))
	registryPath := filepath.Join(packageRoot, filepath.FromSlash(officialRegistryModulePath))
	dependencyPath := filepath.Join(packageRoot, filepath.FromSlash(officialRegistryDependencyPath))
	for _, path := range []string{
		packageRoot, filepath.Join(packageRoot, "dist"), loader,
		filepath.Join(packageRoot, "dist", "resources"),
		filepath.Join(packageRoot, "dist", "resources", "extensions"),
		filepath.Join(packageRoot, "dist", "resources", "extensions", "gsd"),
		filepath.Join(packageRoot, "dist", "resources", "extensions", "shared"),
		registryPath, dependencyPath, filepath.Join(packageRoot, "package.json"),
	} {
		if err := requireRuntimePathWithoutSymlink(path); err != nil {
			return pinnedRegistryRuntime{}, err
		}
	}
	for _, path := range []string{loader, registryPath, dependencyPath, filepath.Join(packageRoot, "package.json")} {
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return pinnedRegistryRuntime{}, registryMismatch(fmt.Sprintf("runtime file %s is missing or not regular", filepath.Base(path)))
		}
		within, relErr := filepath.Rel(packageRoot, path)
		if relErr != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
			return pinnedRegistryRuntime{}, registryMismatch("runtime file escapes the package root")
		}
	}
	if err := validatePackageIdentity(packageRoot, expectedVersion); err != nil {
		return pinnedRegistryRuntime{}, err
	}
	for path, expected := range map[string]string{loader: options.expectedLoaderSHA256, registryPath: options.expectedRegistrySHA256, dependencyPath: options.expectedDependencySHA256} {
		digest, _, err := boundedRuntimeFileSHA256(path, maxRuntimeSourceBytes)
		if err != nil || digest != expected {
			return pinnedRegistryRuntime{}, registryMismatch("official executable registry source drifted from the qualified GSD 1.11.0 runtime")
		}
	}
	return pinnedRegistryRuntime{nodePath: nodePath, packageRoot: packageRoot, registryPath: registryPath, dependencyPath: dependencyPath}, nil
}

func requireRuntimePathWithoutSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return registryMismatch(fmt.Sprintf("inspect runtime path %s: %v", filepath.Base(path), err))
	}
	if info.Mode()&os.ModeSymlink != 0 || !runtimePathOwnedByCurrentUser(info) {
		return registryMismatch(fmt.Sprintf("runtime path %s must be owned and must not be a symlink", filepath.Base(path)))
	}
	return nil
}

func validateActiveRegistrySource(gsdHome, expectedHash string) error {
	if strings.TrimSpace(gsdHome) == "" || !filepath.IsAbs(gsdHome) {
		return registryMismatch("absolute controlled GSD home is required")
	}
	path := filepath.Join(gsdHome, "agent", "extensions", "gsd", "unit-registry.js")
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return registryMismatch("controlled registry cache must be a regular file")
	}
	if !fileHasSHA256(path, expectedHash) {
		return registryMismatch("controlled registry cache differs from the pinned package")
	}
	return nil
}

func fileHasSHA256(path, expected string) bool {
	digest, _, err := boundedRuntimeFileSHA256(path, maxRuntimeSourceBytes)
	return err == nil && digest == expected
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && strings.ToLower(value) == value
}

func verifiedRegistryDataURLs(registryPath, dependencyPath string, options registryLoadOptions) (string, string, error) {
	registryRaw, _, err := readBoundedRuntimeFile(registryPath, maxRuntimeSourceBytes)
	if err != nil {
		return "", "", registryMismatch("read verified registry source")
	}
	dependencyRaw, _, err := readBoundedRuntimeFile(dependencyPath, maxRuntimeSourceBytes)
	if err != nil {
		return "", "", registryMismatch("read verified registry dependency")
	}
	registryHash := sha256.Sum256(registryRaw)
	dependencyHash := sha256.Sum256(dependencyRaw)
	if hex.EncodeToString(registryHash[:]) != options.expectedRegistrySHA256 || hex.EncodeToString(dependencyHash[:]) != options.expectedDependencySHA256 {
		return "", "", registryMismatch("verified registry bytes differ from the qualified source")
	}
	dependencyURL := "data:text/javascript;base64," + base64.StdEncoding.EncodeToString(dependencyRaw)
	const dependencyImport = `"../shared/browser-contract.js"`
	if bytes.Count(registryRaw, []byte(dependencyImport)) != 1 {
		return "", "", registryMismatch("verified registry dependency import has an unexpected shape")
	}
	rewritten := bytes.Replace(registryRaw, []byte(dependencyImport), []byte(strconv.Quote(dependencyURL)), 1)
	registryURL := "data:text/javascript;base64," + base64.StdEncoding.EncodeToString(rewritten)
	return registryURL, dependencyURL, nil
}

func registryExporterEnvironment() []string {
	allowed := map[string]struct{}{"PATH": {}, "TMPDIR": {}, "LANG": {}, "LC_ALL": {}, "NO_COLOR": {}}
	environment := make([]string, 0, len(allowed))
	for _, entry := range os.Environ() {
		name, _, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if _, keep := allowed[strings.ToUpper(name)]; keep {
			environment = append(environment, entry)
		}
	}
	return environment
}

func decodeNormalizedUnitRegistry(raw []byte) (UnitRegistry, error) {
	return decodeNormalizedUnitRegistryWithSource(raw, "1.11.0", officialRegistrySHA256, officialRegistryDependencySHA256, officialLoaderSHA256, officialPatchedHeadlessSHA256, officialPatchedComposerSHA256)
}

func decodeNormalizedUnitRegistryWithSource(raw []byte, expectedVersion, expectedRegistryHash, expectedDependencyHash, expectedLoaderHash, expectedHeadlessHash, expectedComposerHash string) (UnitRegistry, error) {
	if len(raw) == 0 || len(raw) > maxNormalizedRegistryBytes {
		return UnitRegistry{}, registryMismatch("normalized registry is empty or oversized")
	}
	if err := rejectDuplicateJSONFields(raw); err != nil {
		return UnitRegistry{}, registryMismatch("normalized registry has duplicate or malformed fields")
	}
	var document normalizedRegistryDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return UnitRegistry{}, registryMismatch(fmt.Sprintf("decode normalized registry: %v", err))
	}
	if err := requireJSONEOF(decoder); err != nil {
		return UnitRegistry{}, err
	}
	wantSource := normalizedRegistrySource{
		PackageName: "@opengsd/gsd-pi", PackageVersion: expectedVersion,
		ModulePath: officialRegistryModulePath, ModuleSHA256: expectedRegistryHash,
		DependencyPath: officialRegistryDependencyPath, DependencySHA256: expectedDependencyHash,
		LoaderPath: "dist/loader.js", LoaderSHA256: expectedLoaderHash,
		HeadlessSHA256: expectedHeadlessHash, ComposerSHA256: expectedComposerHash,
	}
	if document.SchemaVersion != normalizedRegistrySchemaVersion || !reflect.DeepEqual(document.Source, wantSource) {
		return UnitRegistry{}, registryMismatch("normalized registry schema or source identity differs")
	}
	if document.Units == nil || len(document.Units) != len(expectedOfficialUnitTypes) {
		return UnitRegistry{}, registryMismatch("normalized registry unit count differs from the qualified contract")
	}
	policy := PinnedSidecarPolicy()
	if document.ExcludedSidecars == nil || len(document.ExcludedSidecars) != len(policy.Units) {
		return UnitRegistry{}, registryMismatch("normalized registry sidecar exclusions differ")
	}
	for _, unitType := range document.ExcludedSidecars {
		if _, ok := policy.Lookup(unitType); !ok {
			return UnitRegistry{}, registryMismatch(fmt.Sprintf("unknown excluded sidecar %q", unitType))
		}
	}
	if err := requireUniqueStrings("excluded sidecars", document.ExcludedSidecars, len(policy.Units)); err != nil {
		return UnitRegistry{}, err
	}

	expected := make(map[string]struct{}, len(expectedOfficialUnitTypes))
	for _, unitType := range expectedOfficialUnitTypes {
		expected[unitType] = struct{}{}
	}
	units := make(map[string]UnitMetadata, len(document.Units))
	for _, normalized := range document.Units {
		if _, ok := expected[normalized.UnitType]; !ok {
			return UnitRegistry{}, registryMismatch(fmt.Sprintf("unknown official unit %q", normalized.UnitType))
		}
		if _, duplicate := units[normalized.UnitType]; duplicate {
			return UnitRegistry{}, registryMismatch(fmt.Sprintf("duplicate official unit %q", normalized.UnitType))
		}
		metadata, err := parseUnitMetadata(normalized)
		if err != nil {
			return UnitRegistry{}, err
		}
		units[metadata.UnitType] = metadata
	}
	for _, unitType := range expectedOfficialUnitTypes {
		if _, ok := units[unitType]; !ok {
			return UnitRegistry{}, registryMismatch(fmt.Sprintf("required official unit %q is missing", unitType))
		}
	}
	return UnitRegistry{Units: units}, nil
}

func parseUnitMetadata(normalized normalizedUnitMetadata) (UnitMetadata, error) {
	if !safeUnitSlug(normalized.UnitType) || (normalized.Kind != "primary" && normalized.Kind != "variant") {
		return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s kind or identity is invalid", normalized.UnitType))
	}
	switch normalized.ScopeClass {
	case "standard", "section-close", "execute-task":
	default:
		return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s scope class is invalid", normalized.UnitType))
	}
	if normalized.PhaseChain == nil {
		return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s phaseChain is null", normalized.UnitType))
	}
	if len(normalized.PhaseChain) == 0 || len(normalized.PhaseChain) > 4 {
		return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s phaseChain is missing", normalized.UnitType))
	}
	var role ModelRole
	for _, phase := range normalized.PhaseChain {
		phaseRole, ok := governedPhaseRoles[phase]
		if !ok {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has unknown phase %q", normalized.UnitType, phase))
		}
		if role != "" && role != phaseRole {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s mixes coordination and implementation phases", normalized.UnitType))
		}
		role = phaseRole
	}
	if normalized.PromptTemplates == nil || normalized.AllowedGSDTools == nil || normalized.RequiredWorkflowTools == nil || normalized.ForbiddenGSDTools == nil {
		return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has null normalized metadata", normalized.UnitType))
	}
	if err := requireUniqueStrings(normalized.UnitType+" phase chain", normalized.PhaseChain, 4); err != nil {
		return UnitMetadata{}, err
	}
	if err := requireUniqueStrings(normalized.UnitType+" prompt templates", normalized.PromptTemplates, 8); err != nil {
		return UnitMetadata{}, err
	}
	if err := requireUniqueStrings(normalized.UnitType+" allowed tools", normalized.AllowedGSDTools, 64); err != nil {
		return UnitMetadata{}, err
	}
	if err := requireUniqueStrings(normalized.UnitType+" required tools", normalized.RequiredWorkflowTools, 64); err != nil {
		return UnitMetadata{}, err
	}
	allowed := make(map[string]struct{}, len(normalized.AllowedGSDTools))
	for _, tool := range normalized.AllowedGSDTools {
		if !safeToolName(tool) {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has unsafe allowed tool %q", normalized.UnitType, tool))
		}
		allowed[tool] = struct{}{}
	}
	for _, tool := range normalized.RequiredWorkflowTools {
		if !safeToolName(tool) {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has unsafe required tool %q", normalized.UnitType, tool))
		}
		if strings.HasPrefix(tool, "gsd_") {
			if _, ok := allowed[tool]; !ok {
				return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s requires unallowed workflow tool %q", normalized.UnitType, tool))
			}
		}
	}
	for _, template := range normalized.PromptTemplates {
		if !safeUnitSlug(template) {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has unsafe prompt template %q", normalized.UnitType, template))
		}
	}
	if len(normalized.ForbiddenGSDTools) > 32 {
		return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has too many forbidden tools", normalized.UnitType))
	}
	forbidden := make(map[string]string, len(normalized.ForbiddenGSDTools))
	for _, entry := range normalized.ForbiddenGSDTools {
		if !strings.HasPrefix(entry.Tool, "gsd_") || !safeToolName(entry.Tool) || strings.TrimSpace(entry.Reason) == "" || len(entry.Reason) > 512 {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s has invalid forbidden tool metadata", normalized.UnitType))
		}
		if _, ok := allowed[entry.Tool]; ok {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s both allows and forbids %q", normalized.UnitType, entry.Tool))
		}
		if _, duplicate := forbidden[entry.Tool]; duplicate {
			return UnitMetadata{}, registryMismatch(fmt.Sprintf("%s duplicates forbidden tool %q", normalized.UnitType, entry.Tool))
		}
		forbidden[entry.Tool] = entry.Reason
	}
	return UnitMetadata{
		UnitType: normalized.UnitType, Kind: normalized.Kind, ScopeClass: normalized.ScopeClass,
		PhaseChain:            append([]string(nil), normalized.PhaseChain...),
		PromptTemplates:       append([]string(nil), normalized.PromptTemplates...),
		AllowedGSDTools:       append([]string(nil), normalized.AllowedGSDTools...),
		RequiredWorkflowTools: append([]string(nil), normalized.RequiredWorkflowTools...),
		ForbiddenGSDTools:     forbidden,
	}, nil
}

func requireUniqueStrings(label string, values []string, limit int) error {
	if len(values) > limit {
		return registryMismatch(label + " exceeds its item bound")
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || len(value) > 256 {
			return registryMismatch(label + " contains an empty or oversized value")
		}
		if _, duplicate := seen[value]; duplicate {
			return registryMismatch(label + " contains a duplicate value")
		}
		seen[value] = struct{}{}
	}
	return nil
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return registryMismatch("normalized registry contains trailing JSON")
	}
	return nil
}

func rejectDuplicateJSONFields(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := walkJSONValue(decoder); err != nil {
		return err
	}
	if token, err := decoder.Token(); err != io.EOF || token != nil {
		if err != nil {
			return err
		}
		return errors.New("trailing JSON token")
	}
	return nil
}

func walkJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("JSON object key is not a string")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate JSON field %q", key)
			}
			seen[key] = struct{}{}
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("malformed JSON object")
		}
	case '[':
		for decoder.More() {
			if err := walkJSONValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("malformed JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func (r UnitRegistry) Lookup(unitType string) (UnitMetadata, bool) {
	metadata, ok := r.Units[unitType]
	return metadata, ok
}

func (r UnitRegistry) CommandForUnit(unitType string) (string, error) {
	if _, ok := r.Lookup(unitType); !ok {
		return "", registryMismatch(fmt.Sprintf("unsupported canonical unit type %q", unitType))
	}
	if unitType == "discuss-milestone" {
		return "discuss", nil
	}
	return unitType, nil
}

func (r UnitRegistry) ModelRoleForUnit(unitType string) (ModelRole, error) {
	metadata, ok := r.Lookup(unitType)
	if !ok {
		return "", registryMismatch(fmt.Sprintf("unsupported canonical unit type %q", unitType))
	}
	var role ModelRole
	for _, phase := range metadata.PhaseChain {
		phaseRole, ok := governedPhaseRoles[phase]
		if !ok {
			return "", registryMismatch(fmt.Sprintf("%s has unknown governed phase %q", unitType, phase))
		}
		if role != "" && role != phaseRole {
			return "", registryMismatch(fmt.Sprintf("%s has an ambiguous governed model role", unitType))
		}
		role = phaseRole
	}
	if role == "" {
		return "", registryMismatch(fmt.Sprintf("%s has no governed model phase", unitType))
	}
	return role, nil
}

func (r UnitRegistry) CanonicalCommands() map[string]struct{} {
	commands := make(map[string]struct{}, len(r.Units))
	for unitType := range r.Units {
		command, err := r.CommandForUnit(unitType)
		if err == nil {
			commands[command] = struct{}{}
		}
	}
	return commands
}

func (r UnitRegistry) IsCanonicalCommand(command string) bool {
	if command == "discuss" {
		_, ok := r.Lookup("discuss-milestone")
		return ok
	}
	_, ok := r.Lookup(command)
	return ok
}

func (r UnitRegistry) CanonicalUnitForInvocation(command, observedUnitType string) (string, bool, error) {
	if command == "next" {
		if observedUnitType == "" {
			return "", false, registryMismatch("next has no canonical observed unit")
		}
		if _, ok := r.Lookup(observedUnitType); !ok {
			return "", false, registryMismatch(fmt.Sprintf("next observed unsupported unit %q", observedUnitType))
		}
		return observedUnitType, true, nil
	}
	if command == "discuss" {
		if observedUnitType != "discuss-milestone" {
			return "", false, registryMismatch("discuss is not bound to discuss-milestone")
		}
		return "discuss-milestone", true, nil
	}
	if _, ok := r.Lookup(command); ok {
		if observedUnitType != command {
			return "", false, registryMismatch(fmt.Sprintf("%s is not the observed canonical next unit", command))
		}
		return command, true, nil
	}
	return "", false, nil
}

func (r UnitRegistry) ValidateObservedTool(unitType, observedTool string) error {
	tool, err := canonicalObservedToolName(observedTool)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(tool, "gsd_") {
		return nil
	}
	metadata, ok := r.Lookup(unitType)
	if !ok {
		return registryMismatch(fmt.Sprintf("unsupported unit %q for observed workflow tool", unitType))
	}
	if reason, forbidden := metadata.ForbiddenGSDTools[tool]; forbidden {
		return registryMismatch(fmt.Sprintf("%s called forbidden tool %s: %s", unitType, tool, reason))
	}
	for _, allowed := range metadata.AllowedGSDTools {
		if allowed == tool {
			return nil
		}
	}
	return registryMismatch(fmt.Sprintf("%s called disallowed workflow tool %s", unitType, tool))
}

func canonicalObservedToolName(tool string) (string, error) {
	parts := strings.Split(tool, "__")
	if len(parts) == 3 && parts[0] == "mcp" && parts[1] != "" && parts[2] != "" {
		if parts[1] != "gsd-workflow" && strings.HasPrefix(parts[2], "gsd_") {
			return "", registryMismatch(fmt.Sprintf("untrusted MCP namespace %q for workflow tool", parts[1]))
		}
		return parts[2], nil
	}
	return tool, nil
}

func (r UnitRegistry) GovernedPhases() (map[string]ModelRole, error) {
	phases := make(map[string]ModelRole)
	for _, metadata := range r.Units {
		for _, phase := range metadata.PhaseChain {
			role, ok := governedPhaseRoles[phase]
			if !ok {
				return nil, registryMismatch(fmt.Sprintf("unknown governed phase %q", phase))
			}
			if existing, exists := phases[phase]; exists && existing != role {
				return nil, registryMismatch(fmt.Sprintf("ambiguous role for governed phase %q", phase))
			}
			phases[phase] = role
		}
	}
	return phases, nil
}

func safeUnitSlug(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func safeToolName(value string) bool {
	if value == "" || len(value) > 128 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			continue
		}
		return false
	}
	return true
}

func (r UnitRegistry) UnitTypes() []string {
	unitTypes := make([]string, 0, len(r.Units))
	for unitType := range r.Units {
		unitTypes = append(unitTypes, unitType)
	}
	sort.Strings(unitTypes)
	return unitTypes
}

func registryMismatch(message string) error {
	return fmt.Errorf("%w: %s", ErrRuntimeContractMismatch, message)
}

func boundedDiagnostic(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 2048 {
		value = value[:2048]
	}
	if value == "" {
		return "no diagnostic"
	}
	return value
}

const normalizedRegistryExporter = `
import { createHash } from "node:crypto";
import { existsSync, lstatSync, readFileSync, realpathSync } from "node:fs";
import { delimiter, dirname, join, relative, resolve, sep } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

let [moduleURL, expectedVersion, expectedModuleHash, dependencyURL, expectedDependencyHash, loaderURL, expectedLoaderHash, expectedHeadlessHash, expectedComposerHash] = process.argv.slice(1);
const fail = (message) => { throw new Error("runtime_contract_mismatch: " + message); };
const hash = (raw) => createHash("sha256").update(raw).digest("hex");
const assertRegular = (path, label) => {
  const info = lstatSync(path);
  if (info.isSymbolicLink() || !info.isFile()) fail(label + " must be a regular non-symlink file");
};
const assertDirectory = (path, label) => {
  const info = lstatSync(path);
  if (info.isSymbolicLink() || !info.isDirectory()) fail(label + " must be a non-symlink directory");
};
const exactKeys = (value, allowed, label) => {
  if (!value || typeof value !== "object" || Array.isArray(value) || Object.getPrototypeOf(value) !== Object.prototype) fail(label + " must be a plain object");
  for (const key of Object.keys(value)) if (!allowed.includes(key)) fail(label + " has unexpected field " + key);
};
const strings = (value, label, limit) => {
  if (!Array.isArray(value) || value.length > limit) fail(label + " must be a bounded array");
  const seen = new Set();
  for (const item of value) {
    if (typeof item !== "string" || item.length === 0 || item.length > 256 || seen.has(item)) fail(label + " has invalid or duplicate values");
    seen.add(item);
  }
  return [...value];
};
const verifiedBytes = loaderURL === "verified-bytes";
if (moduleURL === "discover" && dependencyURL === "discover" && loaderURL === "discover") {
  const loaderCandidate = (process.env.PATH ?? "").split(delimiter).map((directory) => join(directory, "gsd")).find((path) => existsSync(path));
  if (!loaderCandidate) fail("container GSD loader is not discoverable");
  const loader = realpathSync(loaderCandidate);
  if (loader.split(sep).at(-1) !== "loader.js" || dirname(loader).split(sep).at(-1) !== "dist") fail("container GSD loader shape differs");
  const discoveredRoot = dirname(dirname(loader));
  moduleURL = pathToFileURL(resolve(discoveredRoot, "dist/resources/extensions/gsd/unit-registry.js")).href;
  dependencyURL = pathToFileURL(resolve(discoveredRoot, "dist/resources/extensions/shared/browser-contract.js")).href;
  loaderURL = pathToFileURL(loader).href;
}
if (verifiedBytes) {
  if (!moduleURL.startsWith("data:text/javascript;base64,") || !dependencyURL.startsWith("data:text/javascript;base64,")) fail("verified registry bytes are malformed");
} else {
  const modulePath = fileURLToPath(moduleURL);
  const dependencyPath = fileURLToPath(dependencyURL);
  const loaderPath = fileURLToPath(loaderURL);
  const packageRoot = resolve(dirname(modulePath), "../../../..");
  for (const [path, label] of [
    [packageRoot, "package root"], [resolve(packageRoot, "dist"), "dist"],
    [resolve(packageRoot, "dist/resources"), "resources"], [resolve(packageRoot, "dist/resources/extensions"), "extensions"],
    [dirname(modulePath), "registry directory"], [dirname(dependencyPath), "dependency directory"],
  ]) assertDirectory(path, label);
  for (const [path, label] of [[modulePath, "registry module"], [dependencyPath, "registry dependency"], [loaderPath, "runtime loader"]]) {
    assertRegular(path, label);
    const rel = relative(packageRoot, path);
    if (rel === ".." || rel.startsWith(".." + sep)) fail(label + " escapes package root");
  }
  const packageJSONPath = resolve(packageRoot, "package.json");
  assertRegular(packageJSONPath, "package metadata");
  const packageJSON = JSON.parse(readFileSync(packageJSONPath, "utf8"));
  if (packageJSON.name !== "@opengsd/gsd-pi" || packageJSON.version !== expectedVersion) fail("package identity differs");
  const moduleRaw = readFileSync(modulePath);
  const dependencyRaw = readFileSync(dependencyPath);
  const loaderRaw = readFileSync(loaderPath);
  const headlessPath = resolve(packageRoot, "dist/headless.js");
  const composerPath = resolve(packageRoot, "dist/resources/extensions/gsd/unit-context-composer.js");
  assertRegular(headlessPath, "patched headless runtime");
  assertRegular(composerPath, "patched prompt composer");
  if (hash(moduleRaw) !== expectedModuleHash || hash(dependencyRaw) !== expectedDependencyHash || hash(loaderRaw) !== expectedLoaderHash || hash(readFileSync(headlessPath)) !== expectedHeadlessHash || hash(readFileSync(composerPath)) !== expectedComposerHash) fail("registry executable source drift");
}
const imported = await import(moduleURL);
exactKeys(imported.UNIT_REGISTRY, Object.keys(imported.UNIT_REGISTRY), "UNIT_REGISTRY");
const units = [];
const excludedSidecars = [];
for (const [unitType, descriptor] of Object.entries(imported.UNIT_REGISTRY)) {
  if (!/^[a-z0-9-]+$/.test(unitType)) fail("unsafe unit type");
  exactKeys(descriptor, ["kind", "scopeClass", "phaseChain", "promptTemplate", "promptTemplates", "toolContract"], unitType);
  if (descriptor.kind !== "primary" && descriptor.kind !== "variant") fail(unitType + " has invalid kind");
  if (!["standard", "section-close", "execute-task"].includes(descriptor.scopeClass)) fail(unitType + " has invalid scope class");
  if (descriptor.phaseChain === null || descriptor.toolContract === null) {
    if (!(["quick-task", "triage-captures"].includes(unitType) && descriptor.phaseChain === null && descriptor.toolContract === null)) fail(unitType + " has partial null metadata");
    excludedSidecars.push(unitType);
    continue;
  }
  const phaseChain = strings(descriptor.phaseChain, unitType + " phaseChain", 4);
  if (phaseChain.length === 0) fail(unitType + " phaseChain is empty");
  let promptTemplates = [];
  if (descriptor.promptTemplate !== undefined && descriptor.promptTemplates !== undefined) fail(unitType + " declares conflicting prompt fields");
  if (descriptor.promptTemplate !== undefined) promptTemplates = strings([descriptor.promptTemplate], unitType + " promptTemplate", 8);
  if (descriptor.promptTemplates !== undefined) promptTemplates = strings(descriptor.promptTemplates, unitType + " promptTemplates", 8);
  exactKeys(descriptor.toolContract, ["allowedGsdTools", "requiredWorkflowTools", "forbiddenGsdTools"], unitType + " toolContract");
  const allowedGsdTools = strings(descriptor.toolContract.allowedGsdTools, unitType + " allowedGsdTools", 64);
  const requiredWorkflowTools = strings(descriptor.toolContract.requiredWorkflowTools, unitType + " requiredWorkflowTools", 64);
  const allowed = new Set(allowedGsdTools);
  for (const tool of requiredWorkflowTools) if (tool.startsWith("gsd_") && !allowed.has(tool)) fail(unitType + " requires an unallowed workflow tool");
  const forbiddenObject = descriptor.toolContract.forbiddenGsdTools ?? {};
  exactKeys(forbiddenObject, Object.keys(forbiddenObject), unitType + " forbiddenGsdTools");
  const forbiddenGsdTools = Object.entries(forbiddenObject).sort(([a], [b]) => a.localeCompare(b)).map(([tool, reason]) => {
    if (!tool.startsWith("gsd_") || typeof reason !== "string" || reason.trim().length === 0 || reason.length > 512 || allowed.has(tool)) fail(unitType + " has invalid forbidden metadata");
    return { tool, reason };
  });
  units.push({ unitType, kind: descriptor.kind, scopeClass: descriptor.scopeClass, phaseChain, promptTemplates, allowedGsdTools, requiredWorkflowTools, forbiddenGsdTools });
}
excludedSidecars.sort();
process.stdout.write(JSON.stringify({
  schemaVersion: "shepherd.gsd-unit-registry/v1",
  source: { packageName: "@opengsd/gsd-pi", packageVersion: expectedVersion, modulePath: "dist/resources/extensions/gsd/unit-registry.js", moduleSHA256: expectedModuleHash, dependencyPath: "dist/resources/extensions/shared/browser-contract.js", dependencySHA256: expectedDependencyHash, loaderPath: "dist/loader.js", loaderSHA256: expectedLoaderHash, headlessSHA256: expectedHeadlessHash, composerSHA256: expectedComposerHash },
  units,
  excludedSidecars,
}));
`
