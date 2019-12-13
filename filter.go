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
	ignoreHeader bool
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

func (f *Filter) parenthesis() bool {
	return false
}

func (f *Filter) stringify(q *GeneratedQuery, sb *strings.Builder) error {
	if f != nil {
		err := f.function.create(q, sb)
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
	if !f.ignoreHeader {
		sb.WriteString(tokenFilter)
		sb.WriteByte('(')
	}
	err := f.stringify(q, sb)
	if !f.ignoreHeader {
		sb.WriteByte(')')
	}
	return err
}

//Function related calls.

//Values is a wrapper to add a list of values to this filter.
func (f *Filter) Values(v ...interface{}) *Filter {
	f.function.values(v)
	return f
}
