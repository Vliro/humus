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

type Operation func(m Mod)

type Mod interface {
	Paginate(t OrderType) bool
	Variable(name string, value string, isAlias bool) bool
	Filter(t FunctionType, variables ...interface{}) bool
	//ConnectedFilter()
	Sort(t OrderType, p Predicate) bool
	Aggregate(t AggregateType, v string) bool
	Count(p Predicate, alias string)
}

type modifierType uint8

const (
	modifierVariable modifierType = 1 << iota
	modifierAggregate
	//All above are field generating modifiers.
	modifierFilter
	modifierPagination
	modifierOrder
	modifierGroupBy
	modifierFacet
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

func (m modifierList) hasModifier(mt modifierType) bool {
	for _, v := range m {
		if v.priority() == mt {
			return true
		}
	}
	return false
}

func (m modifierList) Less(i, j int) bool {
	return m[i].priority() < m[j].priority()
}

func (m modifierList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

//aggregateValues represents a modifier with a type(sum),
//an alias for changing json key as well as what variable or predicate it acts on.
type aggregateValues struct {
	Type     AggregateType
	Alias    string
	Variable string
}

func (a aggregateValues) canApply(mt modifierSource) bool {
	return true
}

func (a aggregateValues) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	sb.WriteByte(' ')
	if a.Variable == "" {
		return errors.New("missing predicate in aggregateValues")
	}
	if a.Alias != "" {
		sb.WriteString(a.Alias)
		sb.WriteString(" : ")
	}
	sb.WriteString(string(a.Type))
	isCount := a.Type == "count"
	sb.WriteByte('(')
	if !isCount {
		sb.WriteString("val(")
	}
	sb.WriteString(a.Variable)
	if !isCount {
		sb.WriteByte(')')
	}
	sb.WriteByte(')')
	sb.WriteByte(' ')
	return nil
}

func (a aggregateValues) priority() modifierType {
	return modifierAggregate
}

func (a aggregateValues) parenthesis() bool {
	return true
}

type groupBy Predicate

func (g groupBy) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (g groupBy) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	if g == "" {
		return errors.New("missing predicate type in groupBy")
	}
	sb.WriteString("@groupby(")
	sb.WriteString(string(g))
	sb.WriteByte(')')
	return nil
}

func (g groupBy) priority() modifierType {
	return modifierGroupBy
}

func (g groupBy) parenthesis() bool {
	return false
}
