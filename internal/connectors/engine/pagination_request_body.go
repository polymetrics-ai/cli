package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

// paginationRequestPlacement is compiled from the existing pagination roles.
// Parameter names remain paginator state keys even when their wire destination
// is a typed body field. No second authored placement map is introduced.
type paginationRequestPlacement struct {
	roles    []paginationBodyRole
	schema   *Schema
	defaults map[string]any
	spec     PaginationSpec
}
type paginationBodyRole struct {
	key, field    string
	numeric, size bool
}

func hasPaginationBody(spec PaginationSpec) bool {
	return spec.BodyCursorField != "" || spec.BodyPageField != "" || spec.BodyOffsetField != "" || spec.BodyLimitField != ""
}

func compilePaginationRequestPlacement(spec PaginationSpec, raw json.RawMessage) (*paginationRequestPlacement, error) {
	p := &paginationRequestPlacement{spec: spec, defaults: map[string]any{}}
	if !hasPaginationBody(spec) {
		return p, nil
	}
	fail := func(err error) (*paginationRequestPlacement, error) {
		return nil, diagnosticAt("/rest/pagination", "pagination_body_invalid", "invalid typed body pagination placement", err)
	}
	switch spec.Type {
	case "cursor", "page_number", "offset_limit":
		// These strategies expose the existing typed cursor/page/offset roles.
	default:
		return fail(fmt.Errorf("pagination strategy has no admitted typed body navigation"))
	}
	var root map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return fail(fmt.Errorf("body pagination requires typed JSON object schema: %w", err))
	}
	if root["type"] != "object" {
		return fail(fmt.Errorf("body pagination requires a selected object schema"))
	}
	for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
		if _, ok := root[keyword]; ok {
			return fail(fmt.Errorf("body pagination requires a selected object variant"))
		}
	}
	props, _ := root["properties"].(map[string]any)
	add := func(key, field string, numeric, size bool) error {
		if field == "" {
			return nil
		}
		if key == "" {
			return fmt.Errorf("body role %q requires its paginator state parameter", field)
		}
		if err := safety.ValidateIdentifier(field, "body paging field"); err != nil {
			return err
		}
		if strings.ContainsAny(field, ".[]") {
			return fmt.Errorf("body paging field must be top-level")
		}
		for _, r := range p.roles {
			if r.field == field || r.key == key {
				return fmt.Errorf("pagination role has conflicting destinations")
			}
		}
		prop, ok := props[field].(map[string]any)
		if !ok {
			return fmt.Errorf("body paging field %q absent from typed request schema", field)
		}
		typ := prop["type"]
		if numeric {
			if typ != "integer" && typ != "number" {
				return fmt.Errorf("body paging field %q must be numeric", field)
			}
		} else if typ != "string" {
			return fmt.Errorf("body paging field %q must be string", field)
		}
		p.roles = append(p.roles, paginationBodyRole{key, field, numeric, size})
		return nil
	}
	if spec.BodyCursorField != "" && spec.Type != "cursor" {
		return fail(fmt.Errorf("body cursor requires cursor strategy"))
	}
	if spec.BodyPageField != "" && spec.Type != "page_number" {
		return fail(fmt.Errorf("body page requires page_number strategy"))
	}
	if spec.BodyOffsetField != "" && spec.Type != "offset_limit" {
		return fail(fmt.Errorf("body offset requires offset_limit strategy"))
	}
	sizeKey := spec.LimitParam
	if sizeKey == "" {
		sizeKey = spec.SizeParam
	}
	if spec.BodyLimitField != "" && spec.LimitParam != "" && spec.SizeParam != "" && spec.LimitParam != spec.SizeParam {
		return fail(fmt.Errorf("body limit has two conflicting size parameters"))
	}
	for _, r := range []paginationBodyRole{{spec.CursorParam, spec.BodyCursorField, false, false}, {spec.PageParam, spec.BodyPageField, true, false}, {spec.OffsetParam, spec.BodyOffsetField, true, false}, {sizeKey, spec.BodyLimitField, true, true}} {
		if err := add(r.key, r.field, r.numeric, r.size); err != nil {
			return fail(err)
		}
	}
	for name, node := range props {
		if prop, ok := node.(map[string]any); ok {
			if value, ok := prop["default"]; ok {
				p.defaults[name] = value
			}
		}
	}
	var err error
	p.schema, err = CompileSchema(raw)
	if err != nil {
		return fail(err)
	}
	for name, value := range p.defaults {
		if err := p.schema.node.properties[name].validate(value, "/body/"+name); err != nil {
			return fail(fmt.Errorf("invalid body default: %w", err))
		}
	}
	return p, nil
}

func (p *paginationRequestPlacement) initial(body any, size, maxBytes int) (map[string]any, int, error) {
	initial := map[string]any{}
	if body != nil {
		m, ok := body.(map[string]any)
		if !ok {
			return nil, 0, fmt.Errorf("body pagination requires JSON object input")
		}
		for k, v := range m {
			initial[k] = copyRecordValue(v)
		}
	}
	for k, v := range p.defaults {
		if _, exists := initial[k]; !exists {
			initial[k] = copyRecordValue(v)
		}
	}
	rawInitial, err := json.Marshal(initial)
	if err != nil {
		return nil, 0, err
	}
	if len(rawInitial) > maxBytes {
		return nil, 0, fmt.Errorf("initial pagination body exceeds request byte bound")
	}
	for name, value := range initial {
		if node := p.schema.node.properties[name]; node != nil {
			if err := node.validate(value, "/body/"+name); err != nil {
				return nil, 0, err
			}
		}
	}
	for _, r := range p.roles {
		if !r.size {
			continue
		}
		if _, ok := initial[r.field]; !ok {
			initial[r.field] = json.Number(fmt.Sprint(size))
		}
		raw, err := json.Marshal(initial[r.field])
		if err != nil {
			return nil, 0, err
		}
		n, ok := new(big.Rat).SetString(string(raw))
		if !ok || !n.IsInt() || !n.Num().IsInt64() || n.Sign() <= 0 || n.Num().Int64() > int64(^uint(0)>>1) {
			return nil, 0, fmt.Errorf("body page size %q must be a positive bounded integer", r.field)
		}
		size = int(n.Num().Int64())
	}
	return initial, size, nil
}

func (p *paginationRequestPlacement) compose(initial map[string]any, query url.Values, maxBytes int) (map[string]any, url.Values, error) {
	body := cloneAnyMap(initial)
	wire := cloneContinuationQuery(query)
	for _, r := range p.roles {
		if values, ok := query[r.key]; ok {
			if len(values) != 1 {
				return nil, nil, fmt.Errorf("ambiguous body paging field %q", r.field)
			}
			if !r.size {
				if r.numeric {
					body[r.field] = json.Number(values[0])
				} else {
					body[r.field] = values[0]
				}
			}
		}
		wire.Del(r.key)
	}
	if err := p.schema.Validate(body); err != nil {
		return nil, nil, fmt.Errorf("effective pagination body: %w", err)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	if len(raw) > maxBytes {
		return nil, nil, fmt.Errorf("effective pagination body exceeds request byte bound")
	}
	return body, wire, nil
}

func (p *paginationRequestPlacement) refuseQuery(query url.Values) error {
	for _, r := range p.roles {
		if _, ok := query[r.key]; ok {
			return fmt.Errorf("pagination parameter %q belongs to body field %q, not query", r.key, r.field)
		}
	}
	return nil
}

func bodyPagingOperationPlan(op OperationSpec) (*paginationRequestPlacement, error) {
	if op.REST == nil || op.REST.Pagination == nil || !hasPaginationBody(*op.REST.Pagination) {
		return nil, nil
	}
	if op.Kind != "rest_read" || strings.ToUpper(op.REST.Method) != http.MethodPost || operationDirectReadContentType(op) != "application/json" || op.REST.MaxBytes <= 0 {
		return nil, fmt.Errorf("body pagination requires bounded typed JSON rest_read POST")
	}
	return compilePaginationRequestPlacement(*op.REST.Pagination, op.REST.BodySchema)
}

// Reuse the existing bounded definition-bound continuation payload. Binding the
// effective initial body and query makes a changed filter/default/placement
// reject before constructing a runtime. The body itself is never in the token.
func bodyPagingIdentity(op OperationSpec, body map[string]any, query url.Values) StreamSpec {
	q := make(map[string]QueryParam, len(query))
	for k, v := range query {
		q[k] = QueryParam{Template: strings.Join(v, "\x00")}
	}
	return StreamSpec{Name: op.ID, Method: op.REST.Method, Path: op.REST.Path, Body: map[string]any{"input": body, "schema": op.REST.BodySchema}, Query: q, Pagination: op.REST.Pagination}
}

const bodyPagingCursorPrefix = "engine_body_v2:"

func decodeBodyPagingCursor(b Bundle, identity StreamSpec, cursor string) (*connsdk.NextPage, error) {
	if cursor == "" {
		return nil, nil
	}
	if !strings.HasPrefix(cursor, bodyPagingCursorPrefix) {
		return nil, fmt.Errorf("body pagination requires its definition-bound page cursor")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(cursor, bodyPagingCursorPrefix))
	if err != nil {
		return nil, fmt.Errorf("invalid body pagination cursor")
	}
	return readContinuationPage(b, identity, &connectors.ReadContinuation{Kind: "engine_pagination_v2", Token: raw})
}
func encodeBodyPagingCursor(b Bundle, identity StreamSpec, next *connsdk.NextPage) (string, error) {
	capsule, err := newReadContinuation(b, identity, next)
	if err != nil {
		return "", err
	}
	return bodyPagingCursorPrefix + base64.RawURLEncoding.EncodeToString(capsule.Token), nil
}

type bodyPagingRequest struct {
	plan     *paginationRequestPlacement
	initial  map[string]any
	size     int
	identity StreamSpec
	binding  Bundle
	resume   *connsdk.NextPage
}

func prepareBodyPagingRequest(b Bundle, op OperationSpec, body any, query url.Values, page int, cursor string, maxBytes int) (*bodyPagingRequest, error) {
	plan, err := bodyPagingOperationPlan(op)
	if err != nil || plan == nil {
		return nil, err
	}
	if err := plan.refuseQuery(query); err != nil {
		return nil, err
	}
	if err := refuseConflictingPagingInput(callerPagingParams(plan.spec, plan.spec.Type, query), directReadWalk{page: page, pageCursor: cursor}, plan.spec.Type); err != nil {
		return nil, err
	}
	size := plan.spec.PageSize
	if size <= 0 {
		size = defaultPageSize
	}
	initial, size, err := plan.initial(body, size, maxBytes)
	if err != nil {
		return nil, err
	}
	for _, role := range plan.roles {
		if !role.size {
			if _, ok := initial[role.field]; ok && (page > 0 || cursor != "") {
				return nil, fmt.Errorf("body field %q conflicts with PM page navigation", role.field)
			}
		}
	}
	identity := bodyPagingIdentity(op, initial, query)
	resume, err := decodeBodyPagingCursor(b, identity, cursor)
	if err != nil {
		return nil, err
	}
	if resume == nil && plan.spec.Type == "cursor" {
		if token, ok := initial[plan.spec.BodyCursorField].(string); ok && token != "" {
			resume = &connsdk.NextPage{Query: url.Values{plan.spec.CursorParam: []string{token}}}
		}
	}
	walkSpec := plan.spec
	walkSpec.PageSize = size
	paginator, err := newPaginator(walkSpec, size, "")
	if err != nil {
		return nil, err
	}
	mode := directReadPageMode{strategy: plan.spec.Type, pageable: true}
	if err := validateDirectReadPageRequest(mode, directReadWalk{page: page, pageCursor: cursor}); err != nil {
		return nil, err
	}
	next := paginator.Start()
	if resume != nil {
		if err := resumePaginator(paginator, resume); err != nil {
			return nil, err
		}
		next = resume
	} else {
		number := page
		if number <= 0 {
			number = 1
		}
		next = &connsdk.NextPage{Query: seekPageQuery(paginator, next, number)}
		for _, role := range plan.roles {
			if !role.size {
				if value, ok := initial[role.field]; ok {
					if next.Query == nil {
						next.Query = url.Values{}
					}
					next.Query.Set(role.key, fmt.Sprint(value))
				}
			}
		}
	}
	if _, _, err := plan.compose(initial, mergeQuery(declaredSizeQuery(plan.spec, size), next.Query), maxBytes); err != nil {
		return nil, err
	}
	return &bodyPagingRequest{plan: plan, initial: initial, size: size, identity: identity, binding: b, resume: resume}, nil
}

// prepareStreamBodyPagination consumes the detached input snapshot supplied by
// request_inputs preparation. Legacy streams without that contract retain their
// established body behavior; no endpoint-based operation inference is used.
func prepareStreamBodyPagination(b Bundle, stream StreamSpec, req connectors.ReadRequest) (StreamSpec, error) {
	spec := stream.Pagination
	if spec == nil {
		spec = b.HTTP.Pagination
	}
	if spec == nil || !hasPaginationBody(*spec) || stream.RequestInputs == nil {
		return stream, nil
	}
	raw, err := streamRequestBodySchema(stream)
	if err != nil {
		return stream, err
	}
	if methodOrDefault(stream.Method) != http.MethodPost || (stream.BodyType != "" && stream.BodyType != "json") || stream.GraphQL != nil {
		return stream, fmt.Errorf("typed body pagination requires JSON POST stream")
	}
	plan, err := compilePaginationRequestPlacement(*spec, raw)
	if err != nil {
		return stream, err
	}
	body := stream.preparedReadBody
	if !stream.preparedReadBodyPresent {
		body, err = resolveStreamBodyMap(stream.Body, requestVars(req.Config, nil, "", req.Query))
		if err != nil {
			return stream, err
		}
	}
	size := spec.PageSize
	if size <= 0 {
		size = defaultPageSize
	}
	initial, size, err := plan.initial(body, size, maxOperationDirectReadBytes)
	if err != nil {
		return stream, err
	}
	query, err := buildInitialQuery(stream, req)
	if err != nil {
		return stream, err
	}
	if err := plan.refuseQuery(query); err != nil {
		return stream, err
	}
	// Keep original source body and request_inputs declaration in identity while
	// binding effective initial values and the selected loaded schema bytes.
	identity := stream
	identity.Body = map[string]any{"declaration": stream.Body, "input": initial, "body_schema": raw, "input_schema": stream.inputPlan.raw}
	identity.Query = map[string]QueryParam{}
	for k, values := range query {
		identity.Query[k] = QueryParam{Template: strings.Join(values, "\x00")}
	}
	base, err := resolveStreamRoute(b, req.Config, stream)
	if err != nil {
		return stream, err
	}
	binding := b
	binding.HTTP.URL = base
	resume, err := readContinuationPage(binding, identity, req.Continuation)
	if err != nil {
		return stream, err
	}
	if resume == nil && spec.Type == "cursor" {
		if token, ok := initial[spec.BodyCursorField].(string); ok && token != "" {
			resume = &connsdk.NextPage{Query: url.Values{spec.CursorParam: []string{token}}}
		}
	}
	walk := *spec
	walk.PageSize = size
	paginator, err := newPaginator(walk, size, recordsPathOf(stream.Records))
	if err != nil {
		return stream, err
	}
	next := paginator.Start()
	if resume != nil {
		for _, role := range plan.roles {
			if !role.size {
				if _, exists := initial[role.field]; exists && req.Continuation != nil {
					return stream, fmt.Errorf("body field %q conflicts with source continuation", role.field)
				}
			}
		}
		if err := resumePaginator(paginator, resume); err != nil {
			return stream, err
		}
		next = resume
	}
	if _, _, err := plan.compose(initial, mergeQuery(declaredSizeQuery(*spec, size), next.Query), maxOperationDirectReadBytes); err != nil {
		return stream, err
	}
	stream.preparedBodyPagination = &bodyPagingRequest{plan: plan, initial: initial, size: size, identity: identity, binding: binding, resume: resume}
	return stream, nil
}
