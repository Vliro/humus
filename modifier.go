package humus

import (
	"errors"
	"strings"
)

type modifierSource int

const (
	modifierField modifierSource = iota
	modifierFunction
)

type modifierType uint8

const (
	modifierVariable modifierType = 1 << iota
	modifierAggregate
	modifierFilter
	modifierPagination
	modifierOrder
	modifierFacet
	modifierGroupBy
)

type modifier interface {
	canApply(mt modifierSource) bool
	//While io.Writer is more generic, the utility of
	//multiple different write methods is unbeatable here.
	apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error
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

//AggregateValues represents a modifier with a type(sum),
//an alias for changing json key as well as what variable or predicate it acts on.
type AggregateValues struct {
	Type     AggregateType
	Alias    string
	Variable string
}

func (a AggregateValues) canApply(mt modifierSource) bool {
	return true
}

func (a AggregateValues) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteByte(' ')
	if a.Variable == "" {
		return errors.New("missing predicate in aggregateValues")
	}
	if a.Alias != "" {
		sb.WriteString(a.Alias)
		sb.WriteString(" : ")
	}
	sb.WriteString(string(a.Type))
	sb.WriteByte('(')
	sb.WriteString("val(")
	sb.WriteString(string(a.Variable))
	//if f.SchemaField.Lang {
	//	f.writeLanguageTag(sb, q.language)
	//}
	sb.WriteByte(')')
	sb.WriteByte(')')
	sb.WriteByte(' ')
	return nil
}

func (a AggregateValues) priority() modifierType {
	return modifierAggregate
}

func (a AggregateValues) parenthesis() bool {
	return true
}

type groupBy Predicate

func (g groupBy) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (g groupBy) apply(root *GeneratedQuery, meta FieldMeta,mt modifierSource, sb *strings.Builder) error {
	panic("implement me")
}

func (g groupBy) priority() modifierType {
	return modifierGroupBy
}

func (g groupBy) parenthesis() bool {
	return false
}
