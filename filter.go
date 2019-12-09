package humus

import (
	"errors"
	"strings"
)

type logicalOp uint8

const (
	logicalOr logicalOp = iota
	logicalAnd
	logicalNot
)

//Filter represents an object in the query that will be serialized as @filter (function...)
//It is just a wrapper over a function.
type Filter struct {
//	op logicalOp
	function
//	node *Filter
}

func (f *Filter) canApply(mt modifierSource) bool {
	return true
}

func (f *Filter) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	err := f.create(root, sb)
	return err
}

func (f *Filter) priority() modifierType {
	return modifierFilter
}
//MakeFilter returns a new filter with the function type typ.
func MakeFilter(typ FunctionType) *Filter {
	return &Filter{function: newFunction(typ)}
}

func (f *Filter) parenthesis() bool {
	return false
}

func (f *Filter) stringify(q *GeneratedQuery, sb *strings.Builder) error {
	if f != nil {
		err := f.function.create(q , sb)
		if err != nil {
			return err
		}
	} else {
		return errors.New("missing function in humus filter")
	}
	return nil
}

func (f *Filter) create(q *GeneratedQuery, sb *strings.Builder) error {
	//No nil checks. Done during check.
	sb.WriteString(tokenFilter)
	sb.WriteByte('(')
	err := f.stringify(q, sb)
	sb.WriteByte(')')
	return err
}

//Function related calls.

//Pred sets a predicate variable, for a has function.
func (f *Filter) Pred(pred Predicate) *Filter {
	f.function.pred(pred)
	return f
}
//PredValue sets a predicate alongside a value, useful for eq.
func (f *Filter) PredValue(pred Predicate, value interface{}) *Filter {
	f.function.predValue(pred, value)
	return f
}
//PredValues sets a predicate alongside a list
func (f *Filter) PredValues(pred Predicate, value ...interface{}) *Filter {
	f.function.predMultiple(pred, value)
	return f
}
//Value is a wrapper to add a value to this filter.
func (f *Filter) Value(v interface{}) *Filter {
	f.function.value(v)
	return f
}
//Values is a wrapper to add a list of values to this filter.
func (f *Filter) Values(v ...interface{}) *Filter {
	f.function.values(v)
	return f
}