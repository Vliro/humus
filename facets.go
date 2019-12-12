package humus

import "strings"

//TODO: Allow proper auto-generation of facets. Is it needed though, is @facets poor performance?
type facet struct {
	active bool
	m      modifierList
}

func (f *facet) Paginate(t PaginationType, value int) bool {
	return false
}

func (f *facet) Variable(name string, value string, isAlias bool) bool {
	f.m = append(f.m, variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return true
}

func (f *facet) Filter(t FunctionType, variables ...interface{}) bool {
	return false
}

func (f *facet) Sort(t OrderType, p Predicate) bool {
	return false
}

func (f *facet) Aggregate(t AggregateType, v string, alias string) bool {
	f.m = append(f.m, aggregateValues{
		Type:     t,
		Alias:    alias,
		Variable: v,
	})
	return true
}

func (f *facet) Count(p Predicate, alias string) bool {
	return false
}

func (f facet) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (f facet) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteString("@facets")
	if len(f.m) > 0 {
		sb.WriteByte('(')
		err := f.m.runVariables(root, meta, mt, sb, true)
		if err != nil {
			return err
		}
		sb.WriteByte(')')
	}
	return nil
}

func (f facet) priority() modifierType {
	return modifierFacet
}

func (f facet) parenthesis() bool {
	return false
}
