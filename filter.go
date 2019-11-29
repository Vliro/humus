package mulbase

import (
	"errors"
	"fmt"
	"strings"
)

//Filter represents an object in the query that will be serialized as @filter (function...)
//It just holds a function.
type Filter struct {
	Function *Function
}

func (f *Filter) canApply(mt modifierSource) bool {
	return true
}

func (f *Filter) apply(root *GeneratedQuery, meta FieldMeta, name string, sb *strings.Builder) (modifierType, error) {
	err := f.create(root, sb)
	return 0, err
}

func (f *Filter) priority() modifierType {
	return modifierFilter
}

func MakeFilter(f *Function) *Filter {
	return &Filter{Function:f}
}

func (f *Filter) parenthesis() bool {
	return false
}

func (f *Filter) create(q *GeneratedQuery, sb *strings.Builder) error {
	//No nil checks. Done during check.
	sb.WriteString(tokenFilter)
	sb.WriteString(tokenLP)
	if f.Function != nil {
		err := f.Function.create(q , sb)
		if err != nil {
			return err
		}
	} else {
		return errors.New("missing function in mulbase filter")
	}
	sb.WriteString(tokenRP)
	return nil
}

func (f *Filter) check(q *GeneratedQuery) error {
	// check query
	if f == nil {
		return fmt.Errorf(fErrNil, "filter")
	}
	return f.Function.check(q)
}
