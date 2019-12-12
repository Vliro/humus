package humus

import (
	"errors"
	"strings"
)

type modifierSource int

/*
Begin simple mods
*/

type mapElement struct {
	m modifierList
	g groupBy
	f facet
}

const (
	modifierField modifierSource = iota
	modifierFunction
)

type Operation func(m Mod)

type Mod interface {
	Paginate(t PaginationType, value int) bool
	Variable(name string, value string, isAlias bool) bool
	Filter(t FunctionType, variables ...interface{}) bool
	Sort(t OrderType, p Predicate) bool
	Aggregate(t AggregateType, v string, alias string) bool
	Count(p Predicate, alias string) bool
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

func (m modifierList) runNormal(q *GeneratedQuery, meta FieldMeta, where modifierSource, sb *strings.Builder) error {
	var curType modifierType
	for k, v := range m {
		newType := v.priority()
		if newType <= modifierAggregate {
			continue
		}
		if v.canApply(where) {
			if newType > curType && v.parenthesis() {
				sb.WriteByte('(')
			}
			err := v.apply(q, meta, where, sb)
			if k != len(m)-1 && v.parenthesis() {
				if p := m[k+1].priority(); p == newType {
					sb.WriteByte(',')
				} else if p != newType {
					sb.WriteByte(')')
				}
			}
			if err != nil {
				return err
			}
		}
		if k == len(m)-1 && v.parenthesis() {
			sb.WriteByte(')')
		}
		curType = newType
	}
	return nil
}

func (m modifierList) runTopLevel(q *GeneratedQuery, meta FieldMeta, where modifierSource, sb *strings.Builder) error {
	for _, v := range m {
		if v.priority() > modifierFilter && v.canApply(modifierFunction) {
			sb.WriteByte(',')
			err := v.apply(q, 0, modifierFunction, sb)
			if err != nil {
				return err
			}
		}
	}
	sb.WriteByte(')')
	for _, v := range m {
		if v.priority() == modifierFilter && v.canApply(modifierFunction) {
			err := v.apply(q, 0, modifierFunction, sb)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m modifierList) runVariables(q *GeneratedQuery, meta FieldMeta, where modifierSource, sb *strings.Builder, commaSeparated bool) error {
	for k, v := range m {
		if v.priority() > modifierAggregate {
			break
		}
		err := v.apply(q, meta, where, sb)
		if err != nil {
			return err
		}
		if commaSeparated && k < len(m)-1 {
			sb.WriteByte(',')
		}
	}
	return nil
}

func (m *modifierList) Paginate(t PaginationType, value int) bool {
	*m = append(*m, pagination{Type: t, Value: value})
	return true
}

func (m *modifierList) Variable(name string, value string, isAlias bool) bool {
	*m = append(*m, variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return true
}

func (m *modifierList) Filter(t FunctionType, variables ...interface{}) bool {
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
	*m = append(*m, &filter)
	return true
}

func (m *modifierList) Sort(t OrderType, p Predicate) bool {
	*m = append(*m, Ordering{
		Type:      t,
		Predicate: p,
	})
	return true
}

func (m *modifierList) Aggregate(t AggregateType, v string, alias string) bool {
	*m = append(*m, aggregateValues{
		Type:     t,
		Alias:    alias,
		Variable: v,
	})
	return true
}

func (m *modifierList) Count(p Predicate, alias string) bool {
	*m = append(*m, variable{
		name:  alias,
		value: "count(" + string(p) + ")",
		alias: true,
	})
	return true
}

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

type groupBy struct {
	m modifierList
	p Predicate
}

func (g *groupBy) Paginate(t PaginationType, value int) bool {
	return false
}

func (g *groupBy) Variable(name string, value string, isAlias bool) bool {
	g.m = append(g.m, variable{
		name:  name,
		value: value,
		alias: isAlias,
	})
	return true
}

func (g *groupBy) Filter(t FunctionType, variables ...interface{}) bool {
	return false
}

func (g *groupBy) Sort(t OrderType, p Predicate) bool {
	return false
}

func (g *groupBy) Aggregate(t AggregateType, v string, alias string) bool {
	g.m = append(g.m, aggregateValues{
		Type:     t,
		Alias:    alias,
		Variable: v,
	})
	return true
}

func (g *groupBy) Count(p Predicate, alias string) bool {
	return false
}

func (g groupBy) canApply(mt modifierSource) bool {
	return mt == modifierField
}

func (g groupBy) apply(root *GeneratedQuery, meta FieldMeta, mt modifierSource, sb *strings.Builder) error {
	if g.p == "" {
		return errors.New("missing predicate type in groupBy")
	}
	sb.WriteString("@groupby(")
	sb.WriteString(string(g.p))
	sb.WriteByte(')')
	sb.WriteByte('{')
	g.m.runVariables(root, 0, mt, sb, false)
	sb.WriteByte('}')
	return nil
}

func (g groupBy) priority() modifierType {
	return modifierGroupBy
}

func (g groupBy) parenthesis() bool {
	return false
}
