package mulbase

import (
	"context"
	"errors"
	errors2 "github.com/pkg/errors"
	"time"
)

var errInvalidType = errors.New("invalid query supplied")
var errInvalidLength = errors.New("invalid number of inputs")
var errParsing = errors.New("error parsing input")
var errTransaction = errors.New("invalid transaction")
var ErrUID = errors.New("missing UID")

var errMissingFunction = errors.New("missing function")
var errMissingVariables = errors.New("missing variables in function")
//Call this on a top level error.
func Error(err error) error {
	if err == nil {
		return nil
	}

	return errors2.Wrap(err,"mulbase")
}

var fErrNil = "nil check failed for %s"

//End of model.template
type dbError struct {
	//This line declares basic properties for a database node.
	Node
	//Regular fields
	Message   string    `json:"Error.message,omitempty"`
	ErrorType string    `json:"Error.errorType,omitempty"`
	Time      time.Time `json:"Error.time,omitempty"`
}

var ErrorFields FieldList = []Field{MakeField("Error.message", 0), MakeField("Error.errorType", 0), MakeField("Error.time", 0)}

//Generating constant field values.
const (
	ErrorMessageField   Predicate = "Error.message"
	ErrorErrorTypeField Predicate = "Error.errorType"
	ErrorTimeField      Predicate = "Error.time"
)

//SaveValues saves the node values that
//do not contain any references to other objects.
func (r *dbError) SaveValues(ctx context.Context, txn *Txn) error {
	mut := CreateMutation(r.Values(), QuerySet)
	err := txn.Query(ctx, mut)
	return err
}
func (r *dbError) GetType() []string {
	if r.Type == nil {
		r.SetType()
	}
	return r.Type
}

//Fields returns all Scalar fields for this value.
func (r *dbError) Fields() FieldList {
	return ErrorFields
}

//Sets the types. This DOES NOT include interfaces!
//as they are set in dgraph already.
func (r *dbError) SetType() {
	r.Type = []string{
		"Error",
	}
}

//Values returns all the scalar values for this node.
func (r *dbError) Values() DNode {
	var m ErrorScalars
	m.Message = r.Message
	m.ErrorType = r.ErrorType
	m.Time = r.Time
	r.SetType()
	m.Node = r.Node
	return &m
}

//Values returns all the scalar values for this node.
func (r *dbError) MapValues() Mapper {
	var m = make(map[string]interface{}, 3)
	m["Error.message"] = r.Message
	m["Error.errorType"] = r.ErrorType
	m["Error.time"] = r.Time
	if r.Uid != "" {
		m["uid"] = r.Uid
	}
	r.SetType()
	m["dgraph.type"] = r.Type
	return m
}

//ErrorScalars is simply to avoid a map[string]interface{}
//It is a mirror of the previous struct with all scalar values.
type ErrorScalars struct {
	Node
	Message   string    `json:"Error.message,omitempty"`
	ErrorType string    `json:"Error.errorType,omitempty"`
	Time      time.Time `json:"Error.time,omitempty"`
}

func (s *ErrorScalars) Values() DNode {
	return s
}

func (s *ErrorScalars) MapValues() Mapper {
	panic("ErrorScalars called, use the original one instead")
}

func (s *ErrorScalars) Fields() FieldList {
	return ErrorFields
}
