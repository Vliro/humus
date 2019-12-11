package humus

import "strings"

//TODO: Allow proper auto-generation of facets. Is it needed though, is @facets poor performance?
type Facet struct {
	//facets []string
}

func (f Facet) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (f Facet) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteString("@facets")
	return nil
}

func (f Facet) priority() modifierType {
	return modifierFacet
}

func (f Facet) parenthesis() bool {
	return false
}
