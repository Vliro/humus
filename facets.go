package humus

import "strings"

//TODO: Allow proper auto-generation of facets. Is it needed though, is @facets poor performance?
type facet struct {
	active bool
	m      modifierList
}

func (f *facetCreator) Paginate(t PaginationType, value int) bool {
	return false
}

func (f *facetCreator) Variable(name string, value string, isAlias bool) bool {
	f.f.m = append(f.f.m, variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return true
}

/*
Applies a filter on the current facet.
*/
func (f *facetCreator) Filter(t FunctionType, variables ...interface{}) bool {
	var filter Filter
	filter.typ = t
	filter.variables = make([]graphVariable, len(variables))
	for k, v := range variables {
		val, typ := processInterface(v)
		filter.variables[k] = graphVariable{
			Value: val,
			Type:  typ,
		}
	}
	filter.mapVariables(f.q)
	filter.ignoreHeader = true
	f.f.m = append(f.f.m, &filter)
	return true
}

func (f *facetCreator) Sort(t OrderType, p Predicate) bool {
	return false
}

func (f *facetCreator) Aggregate(t AggregateType, v string, alias string) bool {
	f.m = append(f.m, aggregateValues{
		Type:     t,
		Alias:    alias,
		Variable: v,
	})
	return true
}

func (f *facetCreator) Count(p Predicate, alias string) bool {
	return false
}

func (f facet) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (f facet) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	if len(f.m) > 0 {
		f.m.sort()
		err := f.m.runFacet(root, meta, sb)
		if err != nil {
			return err
		}
	} else {
		sb.WriteString("@facets")
	}
	return nil
}

func (f facet) priority() modifierType {
	return modifierFacet
}

func (f facet) parenthesis() bool {
	return false
}
