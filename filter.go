package mulbase

import (
	"bytes"
	"fmt"
)

//Filter represents an object in the query that will be serialized as @filter (function...)
//It just holds a function.
type Filter struct {
	Function *Function
}

func MakeFilter(f *Function) *Filter {
	return &Filter{Function:f}
}

func (f *Filter) create(q *GeneratedQuery, parent Predicate, sb *bytes.Buffer) {
	//No nil checks. Done during check.
	sb.WriteString(tokenFilter)
	sb.WriteString(tokenLP)
	if f.Function != nil {
		f.Function.create(q, parent, sb)
	}
	sb.WriteString(tokenRP)
}

func (f *Filter) check(q *GeneratedQuery) error {
	// check query
	if f == nil {
		return fmt.Errorf(fErrNil, "filter")
	}
	err := f.Function.check(q)
	return err
}
