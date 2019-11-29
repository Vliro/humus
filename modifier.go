package mulbase

import (
	"strings"
)

type modifierSource int

const (
	modifierField modifierSource = iota
	modifierFunction
)

type modifierType uint8

const (
	modifierAggregate modifierType = 1 << iota
	modifierPagination
	modifierOrder
	modifierFilter
	modifierFacet
)

type modifier interface {
	canApply(mt modifierSource) bool
	//While io.Writer is more generic, the utility of
	//multiple different write methods is unbeatable here.
	apply(root *GeneratedQuery, meta FieldMeta, name string, sb *strings.Builder) (modifierType, error)
	priority() modifierType
	parenthesis() bool
}

type modifierList []modifier

func (m modifierList) Len() int {
	return len(m)
}

func (m modifierList) Less(i, j int) bool {
	return m[i].priority() < m[j].priority()
}

func (m modifierList) Swap(i, j int) {
	m[i],m[j] = m[j], m[i]
}

//An aggregate value i.e. sum as well as what alias to name it as.
type AggregateValues struct {
	Type  AggregateType
	Alias string
}

func (a AggregateValues) canApply(mt modifierSource) bool {
	return true
}

func (a AggregateValues) apply(root *GeneratedQuery, meta FieldMeta, name string, sb *strings.Builder) (modifierType, error) {
	sb.WriteString(a.Alias)
	sb.WriteString(" : ")
	sb.WriteString(string(a.Type))
	sb.WriteString(tokenLP)
	sb.WriteString(name)
	//if f.SchemaField.Lang {
	//	f.writeLanguageTag(sb, q.Language)
	//}
	sb.WriteString(tokenRP)
	return 0,nil
}

func (a AggregateValues) priority() modifierType {
	return modifierAggregate
}

func (a AggregateValues) parenthesis() bool {
	return true
}