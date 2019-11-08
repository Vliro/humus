package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"mulbase/gen/graphql-go/common"
	"mulbase/gen/graphql-go/errors"
	"mulbase/gen/graphql-go/internal/exec"
	"mulbase/gen/graphql-go/internal/exec/resolvable"
	"mulbase/gen/graphql-go/internal/exec/selected"
	"mulbase/gen/graphql-go/internal/query"
	"mulbase/gen/graphql-go/internal/validation"
	"mulbase/gen/graphql-go/introspection"
	"mulbase/gen/graphql-go/log"
	"mulbase/gen/graphql-go/schema"
	"mulbase/gen/graphql-go/trace"
)

// ParseSchema parses a GraphQL Schema and attaches the given root resolver. It returns an error if
// the Go type signature of the resolvers does not match the Schema. If nil is passed as the
// resolver, then the Schema can not be executed, but it may be inspected (e.g. with ToJSON).
func ParseSchema(schemaString string, resolver interface{}, opts ...SchemaOpt) (*Schema, error) {
	s := &Schema{
		Schema:           schema.New(),
		maxParallelism:   10,
		tracer:           trace.OpenTracingTracer{},
		validationTracer: trace.NoopValidationTracer{},
		logger:           &log.DefaultLogger{},
	}
	for _, opt := range opts {
		opt(s)
	}

	if err := s.Schema.Parse(schemaString, s.useStringDescriptions); err != nil {
		return nil, err
	}
	if err := s.validateSchema(); err != nil {
		return nil, err
	}

	r, err := resolvable.ApplyResolver(s.Schema, resolver)
	if err != nil {
		return nil, err
	}
	s.res = r

	return s, nil
}

// MustParseSchema calls ParseSchema and panics on error.
func MustParseSchema(schemaString string, resolver interface{}, opts ...SchemaOpt) *Schema {
	s, err := ParseSchema(schemaString, resolver, opts...)
	if err != nil {
		panic(err)
	}
	return s
}

// Schema represents a GraphQL Schema with an optional resolver.
type Schema struct {
	Schema *schema.Schema
	res    *resolvable.Schema

	maxDepth              int
	maxParallelism        int
	tracer                trace.Tracer
	validationTracer      trace.ValidationTracer
	logger                log.Logger
	useStringDescriptions bool
	disableIntrospection  bool
}

// SchemaOpt is an option to pass to ParseSchema or MustParseSchema.
type SchemaOpt func(*Schema)

// UseStringDescriptions enables the usage of double quoted and triple quoted
// strings as descriptions as per the June 2018 spec
// https://facebook.github.io/graphql/June2018/. When this is not enabled,
// comments are parsed as descriptions instead.
func UseStringDescriptions() SchemaOpt {
	return func(s *Schema) {
		s.useStringDescriptions = true
	}
}

// UseFieldResolvers specifies whether to use struct field resolvers
func UseFieldResolvers() SchemaOpt {
	return func(s *Schema) {
		s.Schema.UseFieldResolvers = true
	}
}

// MaxDepth specifies the maximum field nesting depth in a query. The default is 0 which disables max depth checking.
func MaxDepth(n int) SchemaOpt {
	return func(s *Schema) {
		s.maxDepth = n
	}
}

// MaxParallelism specifies the maximum number of resolvers per request allowed to run in parallel. The default is 10.
func MaxParallelism(n int) SchemaOpt {
	return func(s *Schema) {
		s.maxParallelism = n
	}
}

// Tracer is used to trace queries and fields. It defaults to trace.OpenTracingTracer.
func Tracer(tracer trace.Tracer) SchemaOpt {
	return func(s *Schema) {
		s.tracer = tracer
	}
}

// ValidationTracer is used to trace validation errors. It defaults to trace.NoopValidationTracer.
func ValidationTracer(tracer trace.ValidationTracer) SchemaOpt {
	return func(s *Schema) {
		s.validationTracer = tracer
	}
}

// Logger is used to log panics during query execution. It defaults to exec.DefaultLogger.
func Logger(logger log.Logger) SchemaOpt {
	return func(s *Schema) {
		s.logger = logger
	}
}

// DisableIntrospection disables introspection queries.
func DisableIntrospection() SchemaOpt {
	return func(s *Schema) {
		s.disableIntrospection = true
	}
}

// Response represents a typical response of a GraphQL server. It may be encoded to JSON directly or
// it may be further processed to a custom response type, for example to include custom error data.
// Errors are intentionally serialized first based on the advice in https://github.com/facebook/graphql/commit/7b40390d48680b15cb93e02d46ac5eb249689876#diff-757cea6edf0288677a9eea4cfc801d87R107
type Response struct {
	Errors     []*errors.QueryError   `json:"errors,omitempty"`
	Data       json.RawMessage        `json:"data,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Validate validates the given query with the Schema.
func (s *Schema) Validate(queryString string) []*errors.QueryError {
	doc, qErr := query.Parse(queryString)
	if qErr != nil {
		return []*errors.QueryError{qErr}
	}

	return validation.Validate(s.Schema, doc, nil, s.maxDepth)
}

// Exec executes the given query with the Schema's resolver. It panics if the Schema was created
// without a resolver. If the context get cancelled, no further resolvers will be called and a
// the context error will be returned as soon as possible (not immediately).
func (s *Schema) Exec(ctx context.Context, queryString string, operationName string, variables map[string]interface{}) *Response {
	if s.res.Resolver == (reflect.Value{}) {
		panic("Schema created without resolver, can not exec")
	}
	return s.exec(ctx, queryString, operationName, variables, s.res)
}

func (s *Schema) exec(ctx context.Context, queryString string, operationName string, variables map[string]interface{}, res *resolvable.Schema) *Response {
	doc, qErr := query.Parse(queryString)
	if qErr != nil {
		return &Response{Errors: []*errors.QueryError{qErr}}
	}

	validationFinish := s.validationTracer.TraceValidation()
	errs := validation.Validate(s.Schema, doc, variables, s.maxDepth)
	validationFinish(errs)
	if len(errs) != 0 {
		return &Response{Errors: errs}
	}

	op, err := getOperation(doc, operationName)
	if err != nil {
		return &Response{Errors: []*errors.QueryError{errors.Errorf("%s", err)}}
	}

	// Subscriptions are not valid in Exec. Use Schema.Subscribe() instead.
	if op.Type == query.Subscription {
		return &Response{Errors: []*errors.QueryError{&errors.QueryError{ Message: "graphql-ws protocol header is missing" }}}
	}
	if op.Type == query.Mutation {
		if _, ok := s.Schema.EntryPoints["mutation"]; !ok {
			return &Response{Errors: []*errors.QueryError{{ Message: "no mutations are offered by the Schema" }}}
		}
	}

	// Fill in variables with the defaults from the operation
	if variables == nil {
		variables = make(map[string]interface{}, len(op.Vars))
	}
	for _, v := range op.Vars {
		if _, ok := variables[v.Name.Name]; !ok && v.Default != nil {
			variables[v.Name.Name] = v.Default.Value(nil)
		}
	}

	r := &exec.Request{
		Request: selected.Request{
			Doc:                  doc,
			Vars:                 variables,
			Schema:               s.Schema,
			DisableIntrospection: s.disableIntrospection,
		},
		Limiter: make(chan struct{}, s.maxParallelism),
		Tracer:  s.tracer,
		Logger:  s.logger,
	}
	varTypes := make(map[string]*introspection.Type)
	for _, v := range op.Vars {
		t, err := common.ResolveType(v.Type, s.Schema.Resolve)
		if err != nil {
			return &Response{Errors: []*errors.QueryError{err}}
		}
		varTypes[v.Name.Name] = introspection.WrapType(t)
	}
	traceCtx, finish := s.tracer.TraceQuery(ctx, queryString, operationName, variables, varTypes)
	data, errs := r.Execute(traceCtx, res, op)
	finish(errs)

	return &Response{
		Data:   data,
		Errors: errs,
	}
}

func (s *Schema) validateSchema() error {
	// https://graphql.github.io/graphql-spec/June2018/#sec-Root-Operation-Types
	// > The query root operation type must be provided and must be an Object type.
	if err := validateRootOp(s.Schema, "query", true); err != nil {
		return err
	}
	// > The mutation root operation type is optional; if it is not provided, the service does not support mutations.
	// > If it is provided, it must be an Object type.
	if err := validateRootOp(s.Schema, "mutation", false); err != nil {
		return err
	}
	// > Similarly, the subscription root operation type is also optional; if it is not provided, the service does not
	// > support subscriptions. If it is provided, it must be an Object type.
	if err := validateRootOp(s.Schema, "subscription", false); err != nil {
		return err
	}
	return nil
}

func validateRootOp(s *schema.Schema, name string, mandatory bool) error {
	t, ok := s.EntryPoints[name]
	if !ok {
		if mandatory {
			return fmt.Errorf("root operation %q must be defined", name)
		}
		return nil
	}
	if t.Kind() != "OBJECT" {
		return fmt.Errorf("root operation %q must be an OBJECT", name)
	}
	return nil
}

func getOperation(document *query.Document, operationName string) (*query.Operation, error) {
	if len(document.Operations) == 0 {
		return nil, fmt.Errorf("no operations in query document")
	}

	if operationName == "" {
		if len(document.Operations) > 1 {
			return nil, fmt.Errorf("more than one operation in query document and no operation name given")
		}
		for _, op := range document.Operations {
			return op, nil // return the one and only operation
		}
	}

	op := document.Operations.Get(operationName)
	if op == nil {
		return nil, fmt.Errorf("no operation with name %q", operationName)
	}
	return op, nil
}