package mulbase

import (
	"bytes"
)

//Filter represents an object in the query that will be serialized as @filter (function...)
type Filter struct {
	Function *Function
}

func MakeFilter(f *Function) *Filter {
	fil := new(Filter)
	fil.Function = f
	return fil
}

func (f *Filter) create(q *GeneratedQuery, parent string, sb *bytes.Buffer) {
	if f != nil && f.Function != nil {
		sb.WriteString(tokenFilter)
		sb.WriteString(tokenLP)
		if f.Function != nil {
			f.Function.create(q, parent, sb)
		}
		sb.WriteString(tokenRP)
	}
}

func (f *Filter) check(q *GeneratedQuery) error {
	// check query
	err := f.Function.check(q)
	if err != nil {
		return err
	}
	return nil
}
