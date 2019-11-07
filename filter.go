package ngraph

import (
	"bytes"
)

type Filter struct {
	Function *Function
}

func MakeFilter(f *Function) *Filter {
	fil := new(Filter)
	fil.Function = f
	return fil
}

func (f *Filter) stringChan(q *Query, parent string, sb *bytes.Buffer) {
	if f != nil && f.Function != nil {
		sb.WriteString(tokenFilter)
		sb.WriteString(tokenLP)
		if f.Function != nil {
			f.Function.string(q, parent, sb)
		}
		sb.WriteString(tokenRP)
	}
}

func (f *Filter) check(q *Query) error {
	// check query
	err := f.Function.check(q)
	if err != nil {
		return err
	}
	return nil
}
