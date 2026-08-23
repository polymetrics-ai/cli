package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/synccontract"
)

// DerivedSyncModes returns sync modes derived from the bundle-declared stream shape.
func DerivedSyncModes(s StreamSpec, sch *StreamSchema) []string {
	cursorField := effectiveCursorField(s, sch)
	return synccontract.SupportedPublicModeNames(synccontract.PublicModeCapabilities{
		HasPrimaryKey:          sch != nil && len(sch.PrimaryKey) > 0,
		HasCursor:              cursorField != "",
		HasIncrementalExecutor: hasIncrementalExecutor(s, cursorField),
	})
}

func effectiveCursorField(s StreamSpec, sch *StreamSchema) string {
	if sch != nil && sch.CursorField != "" {
		return sch.CursorField
	}
	if s.Incremental != nil {
		return s.Incremental.CursorField
	}
	return ""
}

func hasIncrementalExecutor(s StreamSpec, cursorField string) bool {
	if s.Incremental == nil || cursorField == "" {
		return false
	}
	return s.Incremental.CursorField == "" || s.Incremental.CursorField == cursorField
}

// Connector adapts a declarative Bundle (+ optional Tier-2 Hooks) to
// connectors.Connector and its optional interfaces. Every method is a thin
// wrapper over the package-level engine functions in read.go/write.go — no
// read/write/check logic is reimplemented here (per the wave0 handoff note).
type Connector struct {
	bundle Bundle
	hooks  Hooks
}

// New returns an engine-backed connectors.Connector for bundle b, dispatching
// to Tier-2 hooks h at their declared extension points (h may be nil when the
// bundle needs no hooks).
func New(b Bundle, h Hooks) *Connector {
	return &Connector{bundle: b, hooks: h}
}

func (c *Connector) Name() string { return c.bundle.Name }

// Metadata synthesizes connectors.Metadata from the bundle's metadata.json,
// matching the legacy hand-written Metadata() shape field-for-field (design
// §C, API-CONTRACT.md §1).
func (c *Connector) Metadata() connectors.Metadata {
	return synthesizeMetadata(c.bundle)
}

// Manifest synthesizes connectors.Manifest from the bundle: streams (with
// derived PK/cursor/sync-modes), write actions, and risk. This is what makes
// engine-backed connectors show up correctly in connectors.ManifestOf without
// any per-connector manifest.go (unlike legacy connectors such as stripe).
func (c *Connector) Manifest() connectors.Manifest {
	return synthesizeManifest(c.bundle)
}

// Definition synthesizes connectors.Definition from the bundle (design §C.1,
// the wave6 target shape already available today via DefinitionProvider).
func (c *Connector) Definition() connectors.Definition {
	return synthesizeDefinition(c.bundle)
}

func (c *Connector) CommandSurface() *connectors.CommandSurface {
	return synthesizeCommandSurface(c.bundle)
}

// RateLimitCoordination returns the safe declaration-level provenance used by
// connector inspection. It intentionally does not attempt a coordinator
// connection or expose any protected scope identity.
func (c *Connector) RateLimitCoordination() connectors.RateLimitCoordination {
	if c == nil || c.bundle.RateLimits == nil || c.bundle.RateLimits.State != connsdk.RateLimitStateDeclared || len(c.bundle.RateLimits.Policies) == 0 {
		return connectors.RateLimitCoordination{}
	}
	hasProcessLocal := false
	hasRequireShared := false
	hasCertificationOnlyRequireShared := false
	for _, policy := range c.bundle.RateLimits.Policies {
		if policy.Coordination == connsdk.RateLimitCoordinationRequireShared {
			if rateLimitPolicyIsCertificationOnly(policy) {
				hasCertificationOnlyRequireShared = true
				continue
			}
			hasRequireShared = true
		} else {
			hasProcessLocal = true
		}
	}
	if hasRequireShared && hasProcessLocal {
		return connectors.RateLimitCoordination{
			Mode:    connectors.RateLimitCoordinationMixed,
			Message: "Rate-limit coordination is policy-scoped: process-local policies protect this pm process only and are not shared across processes; require_shared policies refuse before sending when the optional coordinator is unavailable.",
		}
	}
	if hasRequireShared {
		return connectors.RateLimitCoordination{
			Mode:    connectors.RateLimitCoordinationRequireShared,
			Message: "Shared rate-limit coordination is required; the command refuses before sending a request when the coordinator is unavailable.",
		}
	}
	message := "Process-local rate-limit protection coordinates this pm process only; it is not shared across processes."
	if hasCertificationOnlyRequireShared {
		message += " Certification traffic requires shared rate-limit coordination and refuses before sending when the coordinator is unavailable."
	}
	return connectors.RateLimitCoordination{
		Mode:    connectors.RateLimitCoordinationProcessLocal,
		Message: message,
	}
}

// rateLimitPolicyIsCertificationOnly reports whether a policy's declared
// coordination requirement applies only to the certification runner. Connector
// inspection has no selected credential tier, so it reports ordinary traffic's
// process-local boundary separately while its message discloses this overlay.
func rateLimitPolicyIsCertificationOnly(policy connsdk.RateLimitPolicy) bool {
	if len(policy.Selector.Tiers) == 0 {
		return false
	}
	for _, tier := range policy.Selector.Tiers {
		if tier != "certification" {
			return false
		}
	}
	return true
}

// HasConfigurationConstraints reports whether this bundle declares
// configuration-time constraints. It is intentionally separate from the
// connector's optional interface presence so callers can distinguish a
// constraint-free bundle from one whose constraints can be evaluated.
func (c *Connector) HasConfigurationConstraints() bool {
	return c.bundle.Spec != nil && c.bundle.Spec.HasConfigurationConstraints()
}

// ValidateConfiguration validates supplied credential configuration against
// this bundle's declared configuration constraints.
func (c *Connector) ValidateConfiguration(config map[string]string) error {
	if c.bundle.Spec == nil {
		return nil
	}
	return c.bundle.Spec.ValidateConfiguration(config)
}

func (c *Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	return executeWithAuthCohort(ctx, cfg, func(admitted context.Context) error {
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, Check(admitted, c.bundle, cfg, c.hooks))
	})
}

// Catalog derives connectors.Catalog from the bundle's streams and compiled
// schemas — no network call (matches the "static" shape of legacy Catalog()
// implementations that don't need Check to run first, e.g. stripe/searxng).
func (c *Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(c.bundle.Streams))
	for _, s := range c.bundle.Streams {
		streams = append(streams, legacyStreamOf(c.bundle, s))
	}
	return connectors.Catalog{Connector: c.bundle.Name, Streams: streams}, nil
}

func (c *Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, Read(admitted, c.bundle, req, c.hooks, emit))
	})
}

// ReadWithOutcome is the closed transport-source variant of Read. It retains
// the normal authentication boundary while making a known page budget stop
// distinguishable from provider exhaustion for a durable continuation.
func (c *Connector) ReadWithOutcome(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	return executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, ReadWithOutcome(admitted, c.bundle, req, c.hooks, emit))
	})
}

// RateLimitParkingScope reproduces the declaration-selected policy group for
// the failed stream. It accepts only connsdk's typed terminal 429 evidence;
// generic HTTP or string-shaped errors cannot create durable parking state.
func (c *Connector) RateLimitParkingScope(ctx context.Context, cfg connectors.RuntimeConfig, stream string, runErr error) (connectors.RateLimitScopeKey, error) {
	var rateErr *connsdk.RateLimitError
	if !errors.As(runErr, &rateErr) || rateErr == nil || !rateErr.HasReset || rateErr.ResetAt.IsZero() {
		return "", errors.New("rate-limit parking scope requires typed reset evidence")
	}
	spec, err := findStream(c.bundle, stream)
	if err != nil {
		return "", err
	}
	method := spec.Method
	if method == "" {
		method = "GET"
	}
	resolver := newRateLimitResolverWithContext(ctx, c.bundle, cfg)
	if resolver == nil {
		return "", errors.New("rate-limit parking scope has no declared policy")
	}
	matched, _, err := resolver.resolvePolicies(ctx, method, spec.Path, resolver.policies)
	if err != nil {
		return "", err
	}
	if len(matched) == 0 || matched[0].parkingScope == "" {
		return "", errors.New("rate-limit parking scope has no matching declared policy")
	}
	return matched[0].parkingScope, nil
}

func (c *Connector) DirectRead(ctx context.Context, req connectors.DirectReadRequest) (connectors.DirectReadResult, error) {
	var result connectors.DirectReadResult
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = DirectRead(admitted, c.bundle, req, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

func (c *Connector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	var result connectors.DirectReadResult
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = OperationDirectRead(admitted, c.bundle, req, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

// ReadBackDeclarativeDestination performs the connector-owned provider read
// selected by a destination read-back declaration. Engine-backed connectors
// bind the opaque operation to an exact declared stream and reuse its bounded
// read executor; shared App code never learns a provider path or URL.
func (c *Connector) ReadBackDeclarativeDestination(ctx context.Context, req connectors.DeclarativeTypedDestinationReadBackRequest) ([]connectors.Record, error) {
	if req.MaxRecords < 1 {
		return nil, fmt.Errorf("declarative destination read-back requires a positive record bound")
	}
	stream, err := findStream(c.bundle, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("declarative destination read-back operation %q is not a declared stream", req.Operation)
	}
	queryParameter, declared := stream.Query[req.ReceiptLocator.QueryParameter]
	if !declared {
		return nil, fmt.Errorf("declarative destination read-back locator query parameter %q is not declared by operation %q", req.ReceiptLocator.QueryParameter, req.Operation)
	}
	if queryParameter.Template != "{{ query."+req.ReceiptLocator.QueryParameter+" }}" {
		return nil, fmt.Errorf("declarative destination read-back locator query parameter %q is not an exact declared query binding", req.ReceiptLocator.QueryParameter)
	}
	locators, err := connectors.ParseDeclarativeTypedDestinationReadBackReceipt(req.Receipt, req.ActionDefinitionSHA256, req.ReceiptLocator, req.MaxRecords)
	if err != nil {
		return nil, err
	}
	records := make([]connectors.Record, 0)
	for _, locator := range locators {
		err := c.Read(ctx, connectors.ReadRequest{
			Stream: req.Operation, Config: req.Runtime, Query: map[string]string{req.ReceiptLocator.QueryParameter: locator}, MaxPages: req.ReceiptLocator.MaxPages,
		}, func(record connectors.Record) error {
			if len(records) >= req.MaxRecords {
				return fmt.Errorf("declarative destination read-back exceeded max_records %d", req.MaxRecords)
			}
			copy := make(connectors.Record, len(record))
			for key, value := range record {
				copy[key] = value
			}
			records = append(records, copy)
			return nil
		})
		if err != nil {
			return records, err
		}
	}
	return records, nil
}

// PreflightOperationDirectRead proves a command's declared binding can reach
// this connector's bounded direct-read executor without resolving credentials
// or making a network request.
func (c *Connector) PreflightOperationDirectRead(operation, method, path string, maxBytes int, outputPolicy string) error {
	return PreflightOperationDirectRead(c.bundle, operation, method, path, maxBytes, outputPolicy)
}

func (c *Connector) PreflightOperationDirectReadBindings(operation string, pathFields, queryFields, bodyFields []string, rawBody bool) error {
	return PreflightOperationDirectReadBindings(c.bundle, operation, pathFields, queryFields, bodyFields, rawBody)
}

// OperationStatusCheck delegates one closed, response-less HEAD operation to
// the engine. It remains distinct from direct reads so a status declaration
// cannot gain JSON response behavior through the connector adapter.
func (c *Connector) OperationStatusCheck(ctx context.Context, req connectors.OperationStatusCheckRequest) (connectors.OperationStatusCheckResult, error) {
	var result connectors.OperationStatusCheckResult
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = OperationStatusCheck(admitted, c.bundle, req, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

// PreflightOperationStatusCheck proves a status_check command's fixed HEAD
// binding without resolving credentials or sending a request.
func (c *Connector) PreflightOperationStatusCheck(operation, method, path, outputPolicy string) error {
	return PreflightOperationStatusCheck(c.bundle, operation, method, path, outputPolicy)
}

// PreflightOperationStructuredJSONVariable lets commandrunner admit a JSON
// flag only after the fixed GraphQL operation's own closed variables schema
// has accepted that exact top-level variable. It resolves no credential and
// makes no request.
func (c *Connector) PreflightOperationStructuredJSONVariable(operation, variable string) error {
	op, err := findOperation(c.bundle, operation)
	if err != nil {
		return err
	}
	return ValidateGraphQLOperationStructuredJSONVariable(op, variable)
}

// PreflightOperationStructuredJSONBodyField admits one declared top-level
// structured input for a fixed REST or GraphQL write operation. It resolves no
// credentials and makes no request; the operation schema remains the only
// authority for the object or array accepted at that field.
func (c *Connector) PreflightOperationStructuredJSONBodyField(operation, field string) error {
	return PreflightOperationStructuredJSONBodyField(c.bundle, operation, field)
}

// PreflightOperationDirectWrite proves a command's declared binding can reach
// this connector's typed write executor without resolving credentials or
// making a network request.
func (c *Connector) PreflightOperationDirectWrite(operation, method, path, outputPolicy string, queryFields ...string) error {
	return PreflightOperationDirectWrite(c.bundle, operation, method, path, outputPolicy, queryFields...)
}

func (c *Connector) PreflightOperationDirectWriteBindings(operation string, pathFields, bodyFields []string) error {
	return PreflightOperationDirectWriteBindings(c.bundle, operation, pathFields, bodyFields)
}

func (c *Connector) MaterializeOperationDirectWriteBody(operation string, mappings map[string]any) (map[string]any, error) {
	return MaterializeOperationDirectWriteBodyMappings(c.bundle, operation, mappings)
}

func (c *Connector) ResolveOperationDirectWriteBodyValue(operation string, body map[string]any, path string) (any, bool, error) {
	return ResolveOperationDirectWriteBodyMappingValue(c.bundle, operation, body, path)
}

func (c *Connector) PreviewOperationDirectWrite(ctx context.Context, req connectors.OperationDirectWriteRequest) (connectors.WritePreview, error) {
	var result connectors.WritePreview
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = PreviewOperationDirectWrite(admitted, c.bundle, req, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

func (c *Connector) OperationDirectWrite(ctx context.Context, req connectors.OperationDirectWriteRequest) (connectors.OperationDirectWriteResult, error) {
	var result connectors.OperationDirectWriteResult
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = OperationDirectWrite(admitted, c.bundle, req, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

func (c *Connector) OperationDirectWriteMetadata(operation string) (connectors.OperationDirectWriteMetadata, error) {
	return OperationDirectWriteMetadata(c.bundle, operation)
}

// OperationBinaryDownload satisfies connectors.OperationBinaryDownloader by
// delegating to the package-level executor. The engine-local request type stays
// the executor's own contract; this adapter is the seam that lets a CLI command
// reach it without the connectors package depending on engine internals.
func (c *Connector) OperationBinaryDownload(ctx context.Context, req connectors.OperationBinaryDownloadRequest) (connectors.OperationBinaryDownloadResult, error) {
	var result BinaryDownloadResult
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = OperationBinaryDownload(admitted, c.bundle, BinaryDownloadRequest{
			Operation: req.Operation, Config: req.Config, PathParams: req.PathParams, Query: req.Query,
			Headers: req.Headers, HeaderValues: req.HeaderValues, MaxBytes: req.MaxBytes, DestRoot: req.DestRoot, FileName: req.FileName, RedactFields: req.RedactFields,
		}, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	publicResult := connectors.OperationBinaryDownloadResult{
		Connector: result.Connector,
		Operation: result.Operation,
		Method:    result.Method,
		Path:      result.Path,
		Record:    result.Record,
		Status:    result.Status,
		Headers:   result.Headers,
		Receipt:   result.Receipt,
	}
	return publicResult, err
}

func (c *Connector) PreflightOperationBinaryDownload(operation, method, path string) error {
	return PreflightOperationBinaryDownload(c.bundle, operation, method, path)
}

// InitialState satisfies connectors.StatefulReader by delegating to the
// package-level engine.InitialState.
func (c *Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	var result map[string]string
	err := executeWithAuthCohort(ctx, cfg, func(admitted context.Context) error {
		var err error
		result, err = InitialState(admitted, c.bundle, stream, cfg)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

// Write executes req against the bundle's writes.json actions. A bundle with
// no writes.json declared at all (c.bundle.Writes is nil) returns
// connectors.ErrUnsupportedOperation, matching every other read-only builtin
// connector (Sample, File) rather than surfacing write.go's
// action-not-found error, which is reserved for "writes.json exists but this
// action name is wrong."
func (c *Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if len(c.bundle.Writes) == 0 {
		return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
	}
	var result connectors.WriteResult
	err := executeWithAuthCohort(ctx, req.Config, func(admitted context.Context) error {
		var err error
		result, err = Write(admitted, c.bundle, req, records, c.hooks)
		return markDeclaredAuthenticationFailure(c.bundle.HTTP.ErrorMap, err)
	})
	return result, err
}

// AuthenticationFailureVerified classifies only errors matched by this
// connector's declared error map. A generic 401 without a matching declaration
// is deliberately insufficient.
func (c *Connector) AuthenticationFailureVerified(err error) bool {
	if connectors.IsVerifiedAuthenticationFailure(err) {
		return true
	}
	var engineErr *Error
	return errors.As(err, &engineErr) && (engineErr.Class == "auth_failed" || declaredAuthenticationRuleMatches(c.bundle.HTTP.ErrorMap, engineErr.Err))
}

func executeWithAuthCohort(ctx context.Context, cfg connectors.RuntimeConfig, operation func(context.Context) error) error {
	if cfg.AuthenticationAdmission == nil {
		return operation(ctx)
	}
	return cfg.AuthenticationAdmission.Execute(ctx, operation)
}

func markDeclaredAuthenticationFailure(rules []ErrorRule, err error) error {
	if err == nil {
		return nil
	}
	var engineErr *Error
	if errors.As(err, &engineErr) && (engineErr.Class == "auth_failed" || declaredAuthenticationRuleMatches(rules, engineErr.Err)) {
		return connectors.MarkVerifiedAuthenticationFailure(err)
	}
	return err
}

func declaredAuthenticationRuleMatches(rules []ErrorRule, err error) bool {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	for _, rule := range rules {
		if rule.Status != httpErr.Status || (rule.MatchBody != "" && !strings.Contains(httpErr.Body, rule.MatchBody)) {
			continue
		}
		return rule.Status == http.StatusUnauthorized || rule.Class == "auth_failed"
	}
	return false
}

// ValidateWrite satisfies connectors.WriteValidator.
func (c *Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if len(c.bundle.Writes) == 0 {
		return connectors.ErrUnsupportedOperation
	}
	return ValidateWrite(ctx, c.bundle, req, records)
}

// PreflightWriteAction exposes the declarative write promotion guard to the
// command runner. Keeping the inspection beside the raw bundle schema makes
// promotion enforcement part of the declarative runtime rather than a copied
// static-validator rule.
func (c *Connector) PreflightWriteAction(name string) error {
	action, err := findWriteAction(c.bundle, name)
	if err != nil {
		return err
	}
	shape, err := InspectRecordSchema(action.RecordSchema)
	if err != nil {
		return err
	}
	// A closed empty record is executable when the provider operation itself
	// declares that it consumes no record input. These actions are deliberate
	// triggers whose target is fully bound by connector configuration (for
	// example DELETE /repos/{configured owner}/{configured repo}); treating
	// them as hollow would make a provider-declared operation unreachable.
	if shape.AdmitsOnlyEmptyObject && writeActionConsumesNoRecord(action) {
		return nil
	}
	return ValidatePromotableRecordSchema(action.RecordSchema)
}

// PreflightBinaryUploadAction exposes the narrower public-upload contract to
// commandrunner. It reuses the existing declarative write action and returns
// only the source fields that the hand-rolled CLI may name; no caller-controlled
// URL, body, media type, or byte limit crosses this boundary.
func (c *Connector) PreflightBinaryUploadAction(name string) ([]connectors.BinaryUploadSource, error) {
	action, err := findWriteAction(c.bundle, name)
	if err != nil {
		return nil, err
	}
	if err := c.PreflightWriteAction(name); err != nil {
		return nil, err
	}
	return BinaryUploadSourcesForWriteAction(action)
}

// BinaryUploadSourcesForWriteAction returns only the declaration-owned file
// fields a public binary_upload command can bind. Both commandrunner and the
// static bundle gate use this function so no second shape of upload action can
// become executable in one path only.
func BinaryUploadSourcesForWriteAction(action WriteAction) ([]connectors.BinaryUploadSource, error) {
	var sources []connectors.BinaryUploadSource
	switch bodyTypeOf(action) {
	case "binary_upload":
		if action.BinaryUpload == nil {
			return nil, fmt.Errorf("binary_upload spec is required")
		}
		sources = append(sources, connectors.BinaryUploadSource{
			Field:             action.BinaryUpload.SourceField,
			MaxBytes:          action.BinaryUpload.MaxBytes,
			AllowedMediaTypes: append([]string(nil), action.BinaryUpload.AllowedMediaTypes...),
		})
	case "base64_upload":
		if action.Base64Upload == nil {
			return nil, fmt.Errorf("base64_upload spec is required")
		}
		sources = append(sources, connectors.BinaryUploadSource{
			Field:             action.Base64Upload.SourceField,
			MaxBytes:          action.Base64Upload.MaxDecodedBytes,
			AllowedMediaTypes: append([]string(nil), action.Base64Upload.AllowedMediaTypes...),
		})
	case "multipart":
		if action.Multipart == nil {
			return nil, fmt.Errorf("multipart spec is required")
		}
		for _, part := range action.Multipart.Parts {
			if part.Type != "file" || !part.Required {
				continue
			}
			sources = append(sources, connectors.BinaryUploadSource{
				Field:             part.Field,
				MaxBytes:          part.MaxBytes,
				AllowedMediaTypes: append([]string(nil), part.AllowedMediaTypes...),
			})
		}
	default:
		return nil, fmt.Errorf("body_type %q is not a binary upload", bodyTypeOf(action))
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("binary upload action requires at least one required file source")
	}
	for _, source := range sources {
		if strings.TrimSpace(source.Field) == "" || source.MaxBytes <= 0 {
			return nil, fmt.Errorf("binary upload source must declare its field and positive byte cap")
		}
		if len(source.AllowedMediaTypes) == 0 {
			return nil, fmt.Errorf("binary upload source %q must declare allowed_media_types", source.Field)
		}
	}
	return sources, nil
}

func writeActionConsumesNoRecord(action WriteAction) bool {
	if bodyTypeOf(action) != "none" || len(action.PathFields) != 0 || strings.Contains(action.Path, "record.") {
		return false
	}
	for _, param := range action.Query {
		if strings.Contains(param.Template, "record.") {
			return false
		}
	}
	return true
}

// PreflightWriteRecordField proves that one exact, declaration-owned write
// action exposes a top-level record field. It is intentionally a name check,
// not a value or request builder: callers cannot use it to add a field to a
// write action or to select another action at runtime.
func (c *Connector) PreflightWriteRecordField(actionName, field string) error {
	action, err := findWriteAction(c.bundle, actionName)
	if err != nil {
		return err
	}
	return ValidateRecordSchemaField(action.RecordSchema, field)
}

// PreflightWriteRecordFieldMapping proves a declaration-owned mapping covers
// the required top-level fields of one exact write action.
func (c *Connector) PreflightWriteRecordFieldMapping(actionName string, fields []string) error {
	action, err := findWriteAction(c.bundle, actionName)
	if err != nil {
		return err
	}
	return ValidateRecordSchemaFieldMapping(action.RecordSchema, fields)
}

func (c *Connector) DeclarativeTypedDestinationActionDigest(actionName string) (string, error) {
	return declarativeTypedDestinationActionDigest(c.bundle, actionName)
}

func (c *Connector) DeclarativeTypedDestinationIdempotencyHeader(actionName string) (string, error) {
	return declarativeTypedDestinationIdempotencyHeader(c.bundle, actionName)
}

// PreflightStructuredJSONRecordField makes the concrete write schema the
// authority for a commandrunner `json` flag. It intentionally accepts a field
// name rather than a raw body or arbitrary path, so the runner cannot grow a
// generic JSON request escape hatch around the declarative write contract.
func (c *Connector) PreflightStructuredJSONRecordField(actionName, field string) error {
	action, err := findWriteAction(c.bundle, actionName)
	if err != nil {
		return err
	}
	return ValidateStructuredJSONRecordField(action.RecordSchema, field)
}

// PreflightStructuredJSONRecordStringArm keeps bare command-line text tied to
// the same named, closed record field as its JSON flag declaration.
func (c *Connector) PreflightStructuredJSONRecordStringArm(actionName, field string) error {
	action, err := findWriteAction(c.bundle, actionName)
	if err != nil {
		return err
	}
	return ValidateStructuredJSONRecordStringArm(action.RecordSchema, field)
}

// DryRunWrite satisfies connectors.DryRunWriter.
func (c *Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if len(c.bundle.Writes) == 0 {
		return connectors.WritePreview{}, connectors.ErrUnsupportedOperation
	}
	return DryRunWrite(ctx, c.bundle, req, records, c.hooks)
}

// Base is embedded by Tier-3 native connectors (design §B.7 Tier 3, e.g.
// native/postgres) to serve identity/metadata/definition and operation
// direct-read preflight from their bundle without duplicating the synthesis
// logic Connector already has. Tier-3 connectors are NOT declaratively
// read/written by the engine — they implement Check/Catalog/Read/Write
// themselves.
type Base struct {
	bundle Bundle
}

// NewBase returns a Base backed by bundle b.
func NewBase(b Bundle) Base {
	return Base{bundle: b}
}

func (b Base) Name() string { return b.bundle.Name }

func (b Base) Metadata() connectors.Metadata {
	return synthesizeMetadata(b.bundle)
}

// BundleManifest gives a Tier-3 native the bundle's configuration, risk, and
// static-stream projection when that native elects to expose a Manifest.
// This stays explicitly named so adding a helper to Base cannot silently
// change every existing Tier-3 connector's public manifest projection.
func (b Base) BundleManifest() connectors.Manifest {
	return synthesizeManifest(b.bundle)
}

func (b Base) Definition() connectors.Definition {
	return synthesizeDefinition(b.bundle)
}

func (b Base) CommandSurface() *connectors.CommandSurface {
	return synthesizeCommandSurface(b.bundle)
}

// PreflightOperationDirectRead validates a native connector's declared
// operation direct-read binding without resolving credentials or network I/O.
func (b Base) PreflightOperationDirectRead(operation, method, path string, maxBytes int, outputPolicy string) error {
	return PreflightOperationDirectRead(b.bundle, operation, method, path, maxBytes, outputPolicy)
}

func (b Base) PreflightOperationDirectReadBindings(operation string, pathFields, queryFields, bodyFields []string, rawBody bool) error {
	return PreflightOperationDirectReadBindings(b.bundle, operation, pathFields, queryFields, bodyFields, rawBody)
}

// PreflightOperationDirectWrite validates a native connector's declared
// operation direct-write binding without resolving credentials or network I/O.
func (b Base) PreflightOperationDirectWrite(operation, method, path, outputPolicy string, queryFields ...string) error {
	return PreflightOperationDirectWrite(b.bundle, operation, method, path, outputPolicy, queryFields...)
}

func (b Base) PreflightOperationDirectWriteBindings(operation string, pathFields, bodyFields []string) error {
	return PreflightOperationDirectWriteBindings(b.bundle, operation, pathFields, bodyFields)
}

func (b Base) MaterializeOperationDirectWriteBody(operation string, mappings map[string]any) (map[string]any, error) {
	return MaterializeOperationDirectWriteBodyMappings(b.bundle, operation, mappings)
}

func (b Base) ResolveOperationDirectWriteBodyValue(operation string, body map[string]any, path string) (any, bool, error) {
	return ResolveOperationDirectWriteBodyMappingValue(b.bundle, operation, body, path)
}

// PreflightOperationStructuredJSONBodyField validates a native connector's
// declaration-owned structured input without resolving credentials or I/O.
func (b Base) PreflightOperationStructuredJSONBodyField(operation, field string) error {
	return PreflightOperationStructuredJSONBodyField(b.bundle, operation, field)
}

func (b Base) PreflightWriteRecordFieldMapping(actionName string, fields []string) error {
	action, err := findWriteAction(b.bundle, actionName)
	if err != nil {
		return err
	}
	return ValidateRecordSchemaFieldMapping(action.RecordSchema, fields)
}

func (b Base) DeclarativeTypedDestinationActionDigest(actionName string) (string, error) {
	return declarativeTypedDestinationActionDigest(b.bundle, actionName)
}

func (b Base) DeclarativeTypedDestinationIdempotencyHeader(actionName string) (string, error) {
	return declarativeTypedDestinationIdempotencyHeader(b.bundle, actionName)
}

func (b Base) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if len(b.bundle.Writes) == 0 {
		return connectors.ErrUnsupportedOperation
	}
	return ValidateWrite(ctx, b.bundle, req, records)
}

func declarativeTypedDestinationActionDigest(b Bundle, actionName string) (string, error) {
	action, err := findWriteAction(b, actionName)
	if err != nil {
		return "", err
	}
	actionJSON, err := json.Marshal(action)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(actionJSON))
	decoder.UseNumber()
	var canonicalAction any
	if err := decoder.Decode(&canonicalAction); err != nil {
		return "", err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return "", fmt.Errorf("action definition contains multiple JSON values")
		}
		return "", err
	}
	canonical, err := json.Marshal(struct {
		Connector     string            `json:"connector"`
		BaseURL       string            `json:"base_url"`
		UserAgent     string            `json:"user_agent"`
		Headers       map[string]string `json:"headers"`
		Auth          []AuthSpec        `json:"auth"`
		DefaultConfig map[string]string `json:"default_config"`
		Action        any               `json:"action"`
	}{
		Connector:     b.Name,
		BaseURL:       b.HTTP.URL,
		UserAgent:     b.HTTP.UserAgent,
		Headers:       b.HTTP.Headers,
		Auth:          b.HTTP.Auth,
		DefaultConfig: materializeConfigDefaults(b, connectors.RuntimeConfig{}).Config,
		Action:        canonicalAction,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func declarativeTypedDestinationIdempotencyHeader(b Bundle, actionName string) (string, error) {
	action, err := findWriteAction(b, actionName)
	if err != nil {
		return "", err
	}
	header := strings.TrimSpace(action.IdempotencyKeyHeader)
	if header == "" {
		return "", fmt.Errorf("action %q has no provider idempotency key header", actionName)
	}
	return header, nil
}

// OperationDirectReadMaxBytes returns the bounded response limit for a
// declared operation direct read.
func (b Base) OperationDirectReadMaxBytes(operation string, requested int) (int, error) {
	op, err := operationDirectReadSpec(b.bundle, operation)
	if err != nil {
		return 0, err
	}
	return clampOperationDirectReadMaxBytes(requested, op.REST.MaxBytes), nil
}

// HasConfigurationConstraints exposes bundle-declared configuration
// constraints to Tier-3 native connectors that embed Base.
func (b Base) HasConfigurationConstraints() bool {
	return b.bundle.Spec != nil && b.bundle.Spec.HasConfigurationConstraints()
}

// ValidateConfiguration validates supplied credential configuration against
// the embedded bundle's declared configuration constraints.
func (b Base) ValidateConfiguration(config map[string]string) error {
	if b.bundle.Spec == nil {
		return nil
	}
	return b.bundle.Spec.ValidateConfiguration(config)
}

// synthesizeMetadata is the single source of truth for bundle -> Metadata,
// shared by Connector.Metadata and Base.Metadata so the two views never
// drift from each other.
func synthesizeMetadata(b Bundle) connectors.Metadata {
	return connectors.Metadata{
		Name:            b.Metadata.Name,
		DisplayName:     b.Metadata.DisplayName,
		IntegrationType: b.Metadata.IntegrationType,
		Description:     b.Metadata.Description,
		Capabilities: connectors.Capabilities{
			Check:   b.Metadata.Capabilities.Check,
			Catalog: true,
			Read:    b.Metadata.Capabilities.Read,
			Write:   b.Metadata.Capabilities.Write,
			Query:   b.Metadata.Capabilities.Query,
		},
	}
}

// synthesizeManifest builds connectors.Manifest from the bundle: legacy
// Stream shape (Fields/PrimaryKey/CursorFields) plus the union of every
// stream's derived sync modes (design §B.6) for Manifest.SyncModes.
func synthesizeManifest(b Bundle) connectors.Manifest {
	streams := make([]connectors.Stream, 0, len(b.Streams))
	writeActions := make([]connectors.WriteActionSpec, 0, len(b.Writes))
	configFields := []connectors.ConfigField{}
	secretFields := []connectors.SecretField{}
	modeSet := map[string]bool{}
	var syncModes []string

	if b.Spec != nil {
		secretSet := map[string]bool{}
		for _, key := range b.Spec.SecretKeys() {
			secretSet[key] = true
		}
		requiredSet := map[string]bool{}
		for _, key := range b.Spec.RequiredKeys() {
			requiredSet[key] = true
		}
		for _, property := range b.Spec.Properties() {
			if secretSet[property] {
				secretFields = append(secretFields, connectors.SecretField{Name: property, Required: requiredSet[property]})
				continue
			}
			configFields = append(configFields, connectors.ConfigField{Name: property, Required: requiredSet[property]})
		}
	}

	for _, s := range b.Streams {
		streams = append(streams, legacyStreamOf(b, s))
		for _, mode := range DerivedSyncModes(s, b.Schemas[s.Name]) {
			if !modeSet[mode] {
				modeSet[mode] = true
				syncModes = append(syncModes, mode)
			}
		}
	}
	// Preserve the shared public mode ordering rather than first-seen-per-stream
	// order.
	syncModes = orderCanonicalModes(syncModes)

	for _, a := range b.Writes {
		confirm := confirmationKindForWriteAction(a)
		writeActions = append(writeActions, connectors.WriteActionSpec{
			Name:            a.Name,
			RequiredFields:  writeActionRequiredFields(a),
			OptionalFields:  writeActionOptionalFields(a),
			Method:          a.Method,
			Path:            a.Path,
			RedactFields:    append([]string(nil), a.RedactFields...),
			Risk:            a.Risk,
			Batchable:       cloneBoolPtr(a.Batchable),
			Confirm:         confirm,
			AllowsUnchanged: a.Kind == "delete" && a.Delete != nil && len(a.Delete.MissingOkStatus) > 0,
		})
	}

	return connectors.Manifest{
		Metadata:     synthesizeMetadata(b),
		ConfigFields: configFields,
		SecretFields: secretFields,
		Streams:      streams,
		WriteActions: writeActions,
		SyncModes:    syncModes,
		Risk: connectors.RiskSpec{
			Read:     b.Metadata.Risk.Read,
			Write:    b.Metadata.Risk.Write,
			Approval: b.Metadata.Risk.Approval,
		},
	}
}

func writeActionRequiredFields(action WriteAction) []string {
	fields := appendUniqueStrings(nil, action.PathFields...)
	var schema struct {
		Required []string `json:"required"`
	}
	if len(action.RecordSchema) > 0 && json.Unmarshal(action.RecordSchema, &schema) == nil {
		fields = appendUniqueStrings(fields, schema.Required...)
	}
	return fields
}

func writeActionOptionalFields(action WriteAction) []string {
	required := map[string]bool{}
	for _, field := range writeActionRequiredFields(action) {
		required[field] = true
	}
	var optional []string
	for _, field := range action.BodyFields {
		if !required[field] {
			optional = appendUniqueStrings(optional, field)
		}
	}
	return optional
}

func appendUniqueStrings(base []string, values ...string) []string {
	seen := map[string]bool{}
	for _, value := range base {
		seen[value] = true
	}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		base = append(base, value)
	}
	return base
}

// synthesizeDefinition builds connectors.Definition from the bundle (design
// §C.1). Spec is the compiled spec.json's original bytes when available; a
// bundle built ad hoc in tests without a Spec gets an empty JSON object
// rather than a nil/invalid RawMessage.
func synthesizeDefinition(b Bundle) connectors.Definition {
	streamSummaries := make([]connectors.StreamSummary, 0, len(b.Streams))
	for _, s := range b.Streams {
		sch := b.Schemas[s.Name]
		summary := connectors.StreamSummary{
			Name:      s.Name,
			SyncModes: DerivedSyncModes(s, sch),
		}
		if sch != nil {
			summary.PrimaryKey = sch.PrimaryKey
		}
		summary.CursorField = effectiveCursorField(s, sch)
		streamSummaries = append(streamSummaries, summary)
	}

	writeActions := make([]connectors.WriteActionInfo, 0, len(b.Writes))
	for _, a := range b.Writes {
		confirm := confirmationKindForWriteAction(a)
		writeActions = append(writeActions, connectors.WriteActionInfo{
			Name:             a.Name,
			Kind:             a.Kind,
			Method:           a.Method,
			Path:             a.Path,
			Risk:             a.Risk,
			Batchable:        cloneBoolPtr(a.Batchable),
			Confirm:          confirm,
			TransportBinding: a.TransportBinding.Clone(),
		})
	}

	var changefeed *connectors.ChangefeedDescriptor
	if b.Changefeed != nil {
		changefeed = b.Changefeed.Clone()
	}
	var pollingWatermark *connectors.PollingWatermarkDescriptor
	if b.PollingWatermark != nil {
		pollingWatermark = b.PollingWatermark.Clone()
	}
	var syncTransport *connectors.SyncTransportDescriptor
	if b.SyncTransport != nil {
		syncTransport = b.SyncTransport.Clone()
	}

	return connectors.Definition{
		Name:             b.Metadata.Name,
		DisplayName:      b.Metadata.DisplayName,
		Description:      b.Metadata.Description,
		IntegrationType:  b.Metadata.IntegrationType,
		DocsURL:          b.Metadata.DocsURL,
		ReleaseStage:     b.Metadata.ReleaseStage,
		Capabilities:     synthesizeMetadata(b).Capabilities,
		Changefeed:       changefeed,
		PollingWatermark: pollingWatermark,
		SyncTransport:    syncTransport,
		Spec:             specJSON(b),
		Streams:          streamSummaries,
		WriteActions:     writeActions,
		Risk: connectors.RiskSpec{
			Read:     b.Metadata.Risk.Read,
			Write:    b.Metadata.Risk.Write,
			Approval: b.Metadata.Risk.Approval,
		},
	}
}

func synthesizeCommandSurface(b Bundle) *connectors.CommandSurface {
	if b.CLISurface == nil {
		return nil
	}
	surface := b.CLISurface
	out := &connectors.CommandSurface{
		Tagline:     surface.Tagline,
		Usage:       surface.Usage,
		Groups:      make([]connectors.CommandSurfaceGroup, 0, len(surface.Groups)),
		GlobalFlags: make([]connectors.CommandSurfaceFlag, 0, len(surface.GlobalFlags)),
		Commands:    make([]connectors.CommandSurfaceCommand, 0, len(surface.Commands)),
		HelpTopics:  make([]connectors.CommandSurfaceHelpTopic, 0, len(surface.HelpTopics)),
	}
	if surface.SourceCLI != nil {
		out.SourceCLI = &connectors.CommandSurfaceSource{
			Name:      surface.SourceCLI.Name,
			Docs:      surface.SourceCLI.Docs,
			Reference: surface.SourceCLI.Reference,
			Source:    surface.SourceCLI.Source,
		}
	}
	for _, group := range surface.Groups {
		out.Groups = append(out.Groups, connectors.CommandSurfaceGroup{
			ID:       group.ID,
			Title:    group.Title,
			Commands: append([]string(nil), group.Commands...),
		})
	}
	for _, flag := range surface.GlobalFlags {
		out.GlobalFlags = append(out.GlobalFlags, commandSurfaceFlag(flag))
	}
	for _, cmd := range surface.Commands {
		flags := make([]connectors.CommandSurfaceFlag, 0, len(cmd.Flags))
		for _, flag := range cmd.Flags {
			flags = append(flags, commandSurfaceOperationFlag(b, cmd, flag))
		}
		out.Commands = append(out.Commands, connectors.CommandSurfaceCommand{
			Path:          cmd.Path,
			Summary:       cmd.Summary,
			Intent:        cmd.Intent,
			Availability:  cmd.Availability,
			Stream:        cmd.Stream,
			Write:         cmd.Write,
			Operation:     cmd.Operation,
			SourceCLIPath: cmd.SourceCLIPath,
			SourceURL:     cmd.SourceURL,
			Flags:         flags,
			Constraints:   commandSurfaceConstraints(cmd.Constraints),
			Examples:      append([]string(nil), cmd.Examples...),
			APISurface:    commandSurfaceEndpointRefs(cmd.APISurface),
			OutputPolicy:  cmd.OutputPolicy,
			RedactFields:  append([]string(nil), cmd.RedactFields...),
			Risk:          cmd.Risk,
			Approval:      cmd.Approval,
			Notes:         cmd.Notes,
		})
	}
	if commandSurfaceHasWriteIntent(out.Commands) {
		for _, flag := range connectors.ReverseETLApprovalFlags() {
			if !commandSurfaceHasGlobalFlag(out.GlobalFlags, flag.Name) {
				out.GlobalFlags = append(out.GlobalFlags, flag)
			}
		}
	}
	for _, topic := range surface.HelpTopics {
		out.HelpTopics = append(out.HelpTopics, connectors.CommandSurfaceHelpTopic{
			Name:    topic.Name,
			Summary: topic.Summary,
		})
	}
	return out
}

func commandSurfaceHasWriteIntent(commands []connectors.CommandSurfaceCommand) bool {
	for _, cmd := range commands {
		if cmd.Intent == "reverse_etl" || cmd.Intent == "binary_upload" || cmd.Intent == "direct_write" {
			return true
		}
	}
	return false
}

func commandSurfaceHasGlobalFlag(flags []connectors.CommandSurfaceFlag, name string) bool {
	for _, flag := range flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}

func commandSurfaceEndpointRefs(refs []CLISurfaceEndpointRef) []connectors.CommandSurfaceEndpointRef {
	out := make([]connectors.CommandSurfaceEndpointRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, connectors.CommandSurfaceEndpointRef{Method: ref.Method, Path: ref.Path})
	}
	return out
}

func commandSurfaceFlag(flag CLIFlag) connectors.CommandSurfaceFlag {
	return connectors.CommandSurfaceFlag{
		Name:            flag.Name,
		Type:            flag.Type,
		Summary:         flag.Summary,
		Values:          append([]string(nil), flag.Values...),
		MapsTo:          flag.MapsTo,
		Format:          flag.Format,
		AllowEmpty:      cloneBoolPtr(flag.AllowEmpty),
		Minimum:         cloneExactNumberPtr(flag.Minimum),
		Maximum:         cloneExactNumberPtr(flag.Maximum),
		Required:        flag.Required,
		Repeatable:      flag.Repeatable,
		EnvOnly:         flag.EnvOnly,
		AllowBareString: flag.AllowBareString,
		MaxItems:        flag.MaxItems,
		MinItems:        flag.MinItems,
		MaxBytes:        flag.MaxBytes,
	}
}

func commandSurfaceOperationFlag(b Bundle, cmd CLICommand, flag CLIFlag) connectors.CommandSurfaceFlag {
	projected := commandSurfaceFlag(flag)
	location, name, ok := strings.Cut(strings.TrimSpace(flag.MapsTo), ".")
	if !ok || (location != "path" && location != "query") || name == "" {
		return projected
	}
	declaredCLIBytes := projected.MaxBytes
	projected.MaxBytes = defaultOperationParameterMaxBytes
	if declaredCLIBytes > 0 && declaredCLIBytes < projected.MaxBytes {
		projected.MaxBytes = declaredCLIBytes
	}
	if strings.TrimSpace(cmd.Operation) == "" {
		return projected
	}
	op, err := findOperation(b, cmd.Operation)
	if err != nil || op.REST == nil {
		return projected
	}
	parameters, err := operationParametersForLocation(op, location)
	if err != nil {
		return projected
	}
	if parameter, declared := parameters[name]; declared {
		projected.MaxBytes = operationParameterByteCap(parameter)
	}
	if declaredCLIBytes > 0 && declaredCLIBytes < projected.MaxBytes {
		projected.MaxBytes = declaredCLIBytes
	}
	return projected
}

func commandSurfaceConstraints(constraints []CLIConstraint) []connectors.CommandSurfaceConstraint {
	out := make([]connectors.CommandSurfaceConstraint, 0, len(constraints))
	for _, constraint := range constraints {
		out = append(out, connectors.CommandSurfaceConstraint{
			Kind:          constraint.Kind,
			Fields:        append([]string(nil), constraint.Fields...),
			Left:          constraint.Left,
			Right:         constraint.Right,
			Op:            constraint.Op,
			ValueType:     constraint.ValueType,
			LeftFallback:  constraint.LeftFallback,
			RightFallback: constraint.RightFallback,
			Message:       constraint.Message,
		})
	}
	return out
}

func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func cloneExactNumberPtr(value *connectors.ExactNumber) *connectors.ExactNumber {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

// specJSON returns the bundle's spec.json VERBATIM (F5, REVIEW.md fix): a
// bundle loaded via Load/LoadAll always has RawSpec populated with the exact
// bytes read from disk, so Definition.Spec now serves types, enums,
// defaults, required, and descriptions byte-for-byte instead of a lossy
// reconstruction. A bundle with no RawSpec at all (an ad hoc test bundle
// that only ever set Spec directly, never went through Load) falls back to
// reconstructing a JSON object from the compiled *Schema's Properties()/
// SecretKeys() — every property flattened to {"type":"string"} (+x-secret) —
// preserving prior behavior for that construction path. A bundle with
// NEITHER RawSpec NOR Spec gets an empty JSON object.
func specJSON(b Bundle) []byte {
	if len(b.RawSpec) > 0 {
		return []byte(b.RawSpec)
	}
	if b.Spec == nil {
		return []byte("{}")
	}
	secrets := make(map[string]bool, len(b.Spec.SecretKeys()))
	for _, name := range b.Spec.SecretKeys() {
		secrets[name] = true
	}
	properties := make(map[string]any, len(b.Spec.Properties()))
	for _, name := range b.Spec.Properties() {
		prop := map[string]any{"type": "string"}
		if secrets[name] {
			prop["x-secret"] = true
		}
		properties[name] = prop
	}
	doc := map[string]any{"type": "object", "properties": properties}
	raw, err := json.Marshal(doc)
	if err != nil {
		return []byte("{}")
	}
	return raw
}

// legacyStreamOf builds the shared connectors.Stream shape for one StreamSpec.
// Bundle loading retains the original schema specifically so static catalogs
// use the same projection as provider-discovered schemas.
func legacyStreamOf(b Bundle, s StreamSpec) connectors.Stream {
	sch := b.Schemas[s.Name]
	stream := connectors.Stream{Name: s.Name}
	cursorField := effectiveCursorField(s, sch)
	if sch == nil {
		if cursorField != "" {
			stream.CursorFields = []string{cursorField}
		}
		return stream
	}
	if len(sch.Raw) > 0 {
		projected, err := connectors.StreamFromSchema(s.Name, "", sch.Raw)
		if err == nil {
			if cursorField != "" {
				projected.CursorFields = []string{cursorField}
			}
			return projected
		}
	}
	// Hand-assembled test bundles predating raw-schema retention still need a
	// useful catalog projection. Loaded bundles always take the path above.
	stream.PrimaryKey = sch.PrimaryKey
	if cursorField != "" {
		stream.CursorFields = []string{cursorField}
	}
	for _, name := range sch.Properties() {
		stream.Fields = append(stream.Fields, connectors.Field{Name: name})
	}
	return stream
}

var canonicalModeOrder = synccontract.PublicModeNames()

// orderCanonicalModes returns the subset of canonicalModeOrder present in
// modes, in canonical order.
func orderCanonicalModes(modes []string) []string {
	present := make(map[string]bool, len(modes))
	for _, m := range modes {
		present[m] = true
	}
	out := make([]string, 0, len(modes))
	for _, m := range canonicalModeOrder {
		if present[m] {
			out = append(out, m)
		}
	}
	return out
}
