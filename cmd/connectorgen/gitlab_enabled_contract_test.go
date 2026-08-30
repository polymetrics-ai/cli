package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// TestGitLabEnabledContractReconcilesSourceLock is the GitLab pilot for the
// enabled-connector contract. It deliberately reads the immutable lock rather
// than inferring membership from an executable command: every locked operation
// must be accounted for by exactly one source-partition lane, while ETL and
// sync remain cited overlays on their selected read operations.
func TestGitLabEnabledContractReconcilesSourceLock(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	bundle, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab bundle: %v", err)
	}
	if bundle.EnabledContract == nil {
		t.Fatal("GitLab enabled_connector_contract.json was not loaded")
	}
	contract := bundle.EnabledContract
	if contract.SourceLock.Path != "sources/gitlab-operation-source-lock.json" || contract.SourceLock.SHA256 == "" || contract.SourceLock.Bytes <= 0 {
		t.Fatalf("GitLab source lock binding = %+v, want immutable cited lock", contract.SourceLock)
	}

	want := map[string]string{
		"direct_read":     "implemented",
		"direct_write":    "implemented",
		"binary_download": "deferred_foundation",
		"binary_upload":   "deferred_foundation",
		"etl":             "implemented",
		"reverse_etl":     "implemented",
		"sync_transport":  "implemented",
	}
	lanes := map[string]bool{}
	for _, lane := range contract.Lanes {
		if want[lane.Name] != lane.State {
			t.Fatalf("GitLab lane %q state = %q, want %q", lane.Name, lane.State, want[lane.Name])
		}
		lanes[lane.Name] = true
		if lane.Reason == "" || len(lane.Citations) == 0 || len(lane.Artifacts) == 0 {
			t.Fatalf("GitLab lane %q has incomplete evidence: %+v", lane.Name, lane)
		}
		for _, artifact := range lane.Artifacts {
			if _, err := fs.Stat(os.DirFS(filepath.Join(defsRoot, "gitlab")), artifact); err != nil {
				t.Fatalf("GitLab lane %q artifact %q: %v", lane.Name, artifact, err)
			}
		}
		if lane.Source.Expected == 0 {
			if lane.Source.Coverage != "not_applicable" {
				t.Fatalf("GitLab lane %q zero-source coverage = %q, want not_applicable", lane.Name, lane.Source.Coverage)
			}
		} else if lane.Source.Implemented == lane.Source.Expected {
			if lane.Source.Coverage != "complete" {
				t.Fatalf("GitLab lane %q complete source coverage = %q, want complete", lane.Name, lane.Source.Coverage)
			}
		} else if lane.Source.Coverage != "partial" {
			t.Fatalf("GitLab lane %q implemented=%d/%d coverage = %q, want partial", lane.Name, lane.Source.Implemented, lane.Source.Expected, lane.Source.Coverage)
		}
	}
	if len(lanes) != len(want) {
		t.Fatalf("GitLab contract lanes = %v, want exactly %v", lanes, want)
	}
	for lane := range want {
		if !lanes[lane] {
			t.Fatalf("GitLab enabled contract omits %q", lane)
		}
	}

	lockRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", contract.SourceLock.Path))
	if err != nil {
		t.Fatalf("read GitLab source lock: %v", err)
	}
	var lock struct {
		Rest struct {
			Operations []struct {
				ID     string `json:"id"`
				Method string `json:"method"`
			} `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatalf("decode GitLab source lock: %v", err)
	}
	operations := make([]connectors.EnabledContractSourceOperation, 0, len(lock.Rest.Operations))
	for _, operation := range lock.Rest.Operations {
		operations = append(operations, connectors.EnabledContractSourceOperation{ID: operation.ID, Method: operation.Method})
	}
	if err := contract.ReconcileSourceOperations(operations); err != nil {
		t.Fatalf("GitLab source-lock lane reconciliation: %v", err)
	}

	const binaryLockPath = "sources/gitlab-binary-operation-source-lock.json"
	if got := contract.SupplementalSourceLocks; len(got) != 1 || got[0].Path != binaryLockPath || got[0].Operations != 2 {
		t.Fatalf("GitLab supplemental source locks = %+v, want retained binary supplement", got)
	}
	binaryRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", binaryLockPath))
	if err != nil {
		t.Fatalf("read GitLab binary source lock: %v", err)
	}
	if _, err := parseSourceImportLock(binaryRaw, "gitlab"); err != nil {
		t.Fatalf("parse GitLab binary source lock: %v", err)
	}
	var binaryLock struct {
		Rest struct {
			SourceDocuments []struct {
				Artifact struct {
					SHA256 string `json:"sha256"`
					Bytes  int    `json:"bytes"`
				} `json:"artifact"`
				Operations []struct {
					ID     string `json:"id"`
					Method string `json:"method"`
				} `json:"operations"`
			} `json:"source_documents"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(binaryRaw, &binaryLock); err != nil {
		t.Fatalf("decode GitLab binary source lock: %v", err)
	}
	binaryOperations := make([]connectors.EnabledContractSourceOperation, 0, 2)
	for _, document := range binaryLock.Rest.SourceDocuments {
		artifactPath := filepath.Join(defsRoot, "gitlab", "sources", "artifacts", document.Artifact.SHA256+".artifact")
		artifact, err := os.ReadFile(artifactPath)
		if err != nil {
			t.Fatalf("read retained binary source artifact %q: %v", artifactPath, err)
		}
		if len(artifact) != document.Artifact.Bytes || fmtSHA256(artifact) != document.Artifact.SHA256 {
			t.Fatalf("retained binary source artifact %q does not match source lock", artifactPath)
		}
		for _, operation := range document.Operations {
			binaryOperations = append(binaryOperations, connectors.EnabledContractSourceOperation{ID: operation.ID, Method: operation.Method})
		}
	}
	if err := contract.ReconcileSupplementalSourceOperations(binaryLockPath, binaryOperations); err != nil {
		t.Fatalf("GitLab binary source-lock lane reconciliation: %v", err)
	}
	if got, want := slices.Collect(func(yield func(string) bool) {
		for _, operation := range binaryOperations {
			if !yield(operation.ID) {
				return
			}
		}
	}), []string{"gitlab.docs.generic_packages.upload_file", "gitlab.docs.repository_files.raw_download"}; !slices.Equal(got, want) {
		t.Fatalf("GitLab binary source operations = %v, want %v", got, want)
	}

	if findings := checkEnabledConnectorContract(os.DirFS(defsRoot), bundle); len(findings) != 0 {
		t.Fatalf("GitLab enabled contract validation findings = %+v", findings)
	}
}

// TestGitLabP5GeneratedNamedJSONValueFlagsRemainSourceBound proves that the
// narrow P5 closed-root action path retains the declared bare-string policy
// for an explicitly named JSON value. The allowance applies to the named
// source field only; it is not a generic JSON request-body flag.
func TestGitLabP5GeneratedNamedJSONValueFlagsRemainSourceBound(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	bundle, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab bundle: %v", err)
	}
	descriptorRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources", "gitlab-operation-descriptor.json"))
	if err != nil {
		t.Fatalf("read GitLab source descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
		t.Fatalf("decode GitLab source descriptor: %v", err)
	}
	writesRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "writes.json"))
	if err != nil {
		t.Fatalf("read GitLab writes: %v", err)
	}
	var writes orderedJSON
	if err := json.Unmarshal(writesRaw, &writes); err != nil {
		t.Fatalf("decode GitLab writes: %v", err)
	}

	wantBareString := map[string]string{
		"postApiV4ChatCompletions":                        "resource_id",
		"postApiV4FeaturesName":                           "value",
		"postApiV4GroupsIdMembers":                        "user_id",
		"postApiV4Mcp":                                    "id",
		"postApiV4ProjectsIdMembers":                      "user_id",
		"putApiV4AdminActiveContextCodeEnabledNamespaces": "namespace_id",
	}
	operations := map[string]sourceOperationDescriptor{}
	for _, operation := range descriptor.Operations {
		operations[operation.SourceID] = operation
	}
	for sourceID, field := range wantBareString {
		source, found := operations[sourceID]
		if !found {
			t.Fatalf("GitLab P5 source %q is missing", sourceID)
		}
		action := sourceProjectionActionForEndpoint(writes.root, source.Method, sourceProjectionDeclaredPath(source))
		if action == nil || !sourceProjectionGeneratedJSONBodyMutationAction(action, source) {
			t.Fatalf("GitLab P5 source %q action = %#v, want its exact generated closed-root action", sourceID, action)
		}
		contract, err := sourceContractForAction(source, action)
		if err != nil {
			t.Fatalf("GitLab P5 source %q contract: %v", sourceID, err)
		}
		if !contract.BareStringFields[field] {
			t.Fatalf("GitLab P5 source %q field %q did not retain declared string arm", sourceID, field)
		}
		var loadedAction *engine.WriteAction
		for index := range bundle.Writes {
			if bundle.Writes[index].Name == stringField(action, "name") {
				loadedAction = &bundle.Writes[index]
				break
			}
		}
		if loadedAction == nil {
			t.Fatalf("GitLab P5 source %q action %q was not loaded", sourceID, stringField(action, "name"))
		}
		for _, intent := range []string{"reverse_etl", "direct_write"} {
			var command engine.CLICommand
			for _, candidate := range bundle.CLISurface.Commands {
				if candidate.Intent == intent && candidate.Write == loadedAction.Name {
					command = candidate
					break
				}
			}
			if reason := sourceActionCoverageReason(*loadedAction, command, source); reason != "" {
				t.Fatalf("GitLab P5 source %q field %q %s binding = %s", sourceID, field, intent, reason)
			}
		}
	}
}

// TestGitLabSurrogateRegexSourceRowStaysMappedAndDeferred pins the current
// Atlas gap: the provider's JSON-Schema regex is preserved in the source
// descriptor, while both command lanes give a discoverable typed outcome
// until a captain-approved engine dialect extension exists.
func TestGitLabSurrogateRegexSourceRowStaysMappedAndDeferred(t *testing.T) {
	const (
		defsRoot = "../../internal/connectors/defs"
		sourceID = "postApiV4VulnerabilitiesVulnerabilityIdFlagsAiDetection"
	)
	bundle, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab bundle: %v", err)
	}
	descriptorRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources", "gitlab-operation-descriptor.json"))
	if err != nil {
		t.Fatalf("read GitLab source descriptor: %v", err)
	}
	var descriptor sourceImportDescriptorDocument
	if err := decodeSourceStrictJSON(descriptorRaw, &descriptor); err != nil {
		t.Fatalf("decode GitLab source descriptor: %v", err)
	}
	var source sourceOperationDescriptor
	for _, operation := range descriptor.Operations {
		if operation.SourceID == sourceID {
			source = operation
			break
		}
	}
	if source.SourceID == "" {
		t.Fatalf("GitLab source descriptor omits %q", sourceID)
	}
	matchedGap := false
	for _, gap := range source.Runtime.Gaps {
		if gap.Foundation == "cli-request-schema-foundation-r1" && gap.Location == "request body property origin" && strings.Contains(gap.Reason, "engine JSON-Schema regex compiler") {
			matchedGap = true
			break
		}
	}
	if !source.Runtime.MergeBlocked || !matchedGap {
		t.Fatalf("GitLab surrogate-regex source runtime = %+v, want exact retained schema gap", source.Runtime)
	}
	for _, intent := range []string{"reverse_etl", "direct_write"} {
		matched := false
		for _, command := range bundle.CLISurface.Commands {
			if command.SourceOperation != sourceID || command.Intent != intent {
				continue
			}
			matched = true
			if command.Availability != "partial" || !strings.Contains(command.Notes, "missing_foundation=cli-request-schema-foundation-r1") {
				t.Fatalf("GitLab surrogate-regex %s command = %+v, want source-cited partial outcome", intent, command)
			}
		}
		if !matched {
			t.Fatalf("GitLab surrogate-regex source has no %s command", intent)
		}
	}
}

// TestGitLabDeprecatedSourceOperationsRemainMappedWithoutAdmissionExclusion
// proves that a provider deprecation notice is retained as source evidence,
// never used as an admission-policy exclusion. A cited operation is either
// visible as a blocked API-surface outcome with its named foundation or as the
// paired direct-write/reverse-ETL command that now executes it.
func TestGitLabDeprecatedSourceOperationsRemainMappedWithoutAdmissionExclusion(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	lockRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources/gitlab-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read GitLab source lock: %v", err)
	}
	var lock struct {
		Rest struct {
			Operations []struct {
				Method      string `json:"method"`
				OperationID string `json:"operation_id"`
				Source      struct {
					Description string `json:"description"`
				} `json:"source_operation"`
			} `json:"operations"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatalf("decode GitLab source lock: %v", err)
	}
	allDocumentedDeprecations := map[string]string{}
	for _, operation := range lock.Rest.Operations {
		if strings.Contains(strings.ToLower(operation.Source.Description), "deprecated") {
			allDocumentedDeprecations[operation.OperationID] = operation.Method
		}
	}
	if len(allDocumentedDeprecations) == 0 {
		t.Fatal("GitLab source lock has no provider-documented deprecated operations to audit")
	}
	// These are the exact provider-deprecation rows that were previously
	// classified as justified exclusions in GitLab's source-backed API surface.
	// Other source descriptions can mention deprecation while already retaining
	// an actual implementable or blocked foundation disposition; they are not
	// admission-policy exclusions and therefore do not belong in this regression
	// set.
	deprecated := map[string]string{}
	for _, operationID := range []string{
		"postApiV4NamespacesIdGitlabSubscription",
		"putApiV4NamespacesIdGitlabSubscription",
		"postApiV4NamespacesIdMinutes",
		"postApiV4NamespacesIdSubscriptionAddOnPurchaseAddOnName",
		"putApiV4NamespacesIdSubscriptionAddOnPurchaseAddOnName",
		"putApiV4UserUserIdCreditCardValidation",
		"putApiV4ProjectsIdLabels",
		"deleteApiV4ProjectsIdLabels",
		"putApiV4ProjectsIdLabelsPromote",
		"getApiV4FeatureFlagsUnleashProjectIdFeatures",
		"postApiV4ProjectsIdMergeRequestsMergeRequestIidApprovals",
	} {
		method, ok := allDocumentedDeprecations[operationID]
		if !ok {
			t.Fatalf("GitLab source lock no longer retains deprecation evidence for %q", operationID)
		}
		deprecated[operationID] = method
	}

	bundle, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab bundle: %v", err)
	}
	found := map[string]bool{}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		if strings.Contains(endpoint.Operation.Notes, "classification=justified_excluded") || endpoint.Operation.Model == "deprecated" {
			t.Fatalf("GitLab API surface still has a deprecation admission-policy exclusion: %s %s %+v", endpoint.Method, endpoint.Path, endpoint.Operation)
		}
		operationID := gitLabSurfaceNoteValue(endpoint.Operation.Notes, "operationId")
		method, isDeprecated := deprecated[operationID]
		if !isDeprecated {
			continue
		}
		found[operationID] = true
		if !strings.Contains(endpoint.Operation.Notes, "classification=blocked_our_foundation") || !strings.Contains(endpoint.Operation.Notes, "source_deprecation=provider_documented") {
			t.Fatalf("deprecated source operation %s must retain a named foundation outcome and deprecation evidence: %+v", operationID, endpoint.Operation)
		}
		if !strings.Contains(endpoint.Operation.Reason, "deprecated") || !strings.Contains(endpoint.Operation.Reason, "contract") {
			t.Fatalf("deprecated source operation %s reason must retain deprecation and actual contract gap: %q", operationID, endpoint.Operation.Reason)
		}
		switch method {
		case "GET", "HEAD":
			if endpoint.Operation.Model != "direct_read" {
				t.Fatalf("deprecated GitLab read %s model = %q, want direct_read foundation gap", operationID, endpoint.Operation.Model)
			}
		case "DELETE":
			if endpoint.Operation.Model != "destructive_action" {
				t.Fatalf("deprecated GitLab deletion %s model = %q, want destructive_action foundation gap", operationID, endpoint.Operation.Model)
			}
		default:
			if endpoint.Operation.Model != "admin_reverse_etl" {
				t.Fatalf("deprecated GitLab mutation %s model = %q, want admin_reverse_etl foundation gap", operationID, endpoint.Operation.Model)
			}
		}
	}
	implemented := map[string]map[string]bool{}
	for _, command := range bundle.CLISurface.Commands {
		if _, isDeprecated := deprecated[command.SourceOperation]; !isDeprecated {
			continue
		}
		if command.Intent != "direct_write" && command.Intent != "reverse_etl" {
			continue
		}
		if command.Availability != "implemented" {
			t.Fatalf("deprecated GitLab source operation %s %s command = %+v, want implemented or a blocked API-surface outcome", command.SourceOperation, command.Intent, command)
		}
		if implemented[command.SourceOperation] == nil {
			implemented[command.SourceOperation] = map[string]bool{}
		}
		implemented[command.SourceOperation][command.Intent] = true
	}
	for operationID := range deprecated {
		if found[operationID] {
			continue
		}
		if implemented[operationID]["direct_write"] && implemented[operationID]["reverse_etl"] {
			continue
		}
		t.Fatalf("deprecated GitLab source operation %s is neither a blocked API-surface outcome nor paired executable command", operationID)
	}
}

func gitLabSurfaceNoteValue(notes, key string) string {
	prefix := key + "="
	for _, field := range strings.Split(notes, ";") {
		field = strings.TrimSpace(field)
		if value, ok := strings.CutPrefix(field, prefix); ok {
			return value
		}
	}
	return ""
}

func gitLabMissingOperationIDs(want map[string]string, found map[string]bool) []string {
	missing := make([]string, 0, len(want))
	for operationID := range want {
		if !found[operationID] {
			missing = append(missing, operationID)
		}
	}
	slices.Sort(missing)
	return missing
}

func TestEnabledConnectorContractBuildGateRejectsMissingDeclaredArtifact(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	missing := "gitlab/sources/artifacts/53244a720b8509536290e0058c946a246817c775c797df36f4c9aa1225fdf0a4.artifact"
	fsys := hiddenEnabledContractArtifactFS{FS: os.DirFS(defsRoot), hidden: missing}
	bundle, err := engine.Load(fsys, "gitlab")
	if err != nil {
		t.Fatalf("runtime bundle load must not require repository-only source evidence: %v", err)
	}
	findings := checkEnabledConnectorContract(fsys, bundle)
	for _, finding := range findings {
		if finding.Rule == ruleEnabledConnectorContract && strings.Contains(finding.Message, "artifact is unavailable") {
			return
		}
	}
	t.Fatalf("enabled-contract build gate findings = %+v, want missing declared artifact refusal", findings)
}

func TestEnabledConnectorContractBuildGateRejectsMalformedDeclaredArtifact(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	artifact := "gitlab/sources/artifacts/f59c93194c095d0e925a5751a08eb7a2176a26c6b5f38bda52f805154219d0f0.artifact"
	fsys := replacedEnabledContractArtifactFS{
		FS:       os.DirFS(defsRoot),
		replaced: artifact,
		raw:      []byte("malformed retained source evidence"),
	}
	bundle, err := engine.Load(fsys, "gitlab")
	if err != nil {
		t.Fatalf("runtime bundle load must not require repository-only source evidence: %v", err)
	}
	findings := checkEnabledConnectorContract(fsys, bundle)
	for _, finding := range findings {
		if finding.Rule == ruleEnabledConnectorContract && strings.Contains(finding.Message, "does not match its source lock identity") {
			return
		}
	}
	t.Fatalf("enabled-contract build gate findings = %+v, want malformed declared artifact refusal", findings)
}

type hiddenEnabledContractArtifactFS struct {
	fs.FS
	hidden string
}

func (h hiddenEnabledContractArtifactFS) Open(name string) (fs.File, error) {
	if name == h.hidden {
		return nil, fs.ErrNotExist
	}
	return h.FS.Open(name)
}

type replacedEnabledContractArtifactFS struct {
	fs.FS
	replaced string
	raw      []byte
}

func (r replacedEnabledContractArtifactFS) Open(name string) (fs.File, error) {
	if name == r.replaced {
		return fstest.MapFS{name: &fstest.MapFile{Data: r.raw}}.Open(name)
	}
	return r.FS.Open(name)
}

func fmtSHA256(raw []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}
