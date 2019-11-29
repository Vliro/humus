package mulbase

import "strings"

type facet struct {

}

func (f facet) canApply(mt modifierSource) bool {
	return true
}

func (f facet) apply(root *GeneratedQuery, meta FieldMeta, name string, sb *strings.Builder) (modifierType, error) {
	sb.WriteString("@facets")
	//TODO: Use a schema to get applicable facets.
	return 0, nil
}

func (f facet) priority() modifierType {
	return modifierFacet
}

func (f facet) parenthesis() bool {
	return false
}
