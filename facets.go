package humus

import "strings"
//TODO: Allow proper auto-generation of facets. Is it needed though, is @facets poor performance?
type facet struct {
	//facets []string
}

func (f facet) canApply(mt modifierSource) bool {
	return true
}

func (f facet) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteString("@facets")
	//TODO: Use a schema to get applicable facets.
	return nil
}

func (f facet) priority() modifierType {
	return modifierFacet
}

func (f facet) parenthesis() bool {
	return false
}
